package collect

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/knownhosts"
)

// 默认超时。每个节点连接与每条命令都受其约束，见 ADR 0001。
const (
	DialTimeout    = 10 * time.Second
	CommandTimeout = 15 * time.Second
	SSHUser        = "root"
	SSHPort        = "22"
)

// Runner 在单个节点上执行白名单命令并返回输出。生产环境为 SSHRunner，测试可替换。
type Runner interface {
	Run(ctx context.Context, ip, commandID string) (string, error)
}

// SSHRunner 通过 ssh-agent 与 known_hosts 登录节点并执行白名单命令。
// 不关闭主机密钥校验；不读取私钥内容。
type SSHRunner struct {
	client   map[string]*ssh.Client
	hostKeys ssh.HostKeyCallback
	agent    agent.ExtendedAgent
}

// NewSSHRunner 初始化 known_hosts 校验与 ssh-agent。两者缺失时直接报错，不降级为不校验。
func NewSSHRunner() (*SSHRunner, error) {
	hk, err := knownHostsCallback()
	if err != nil {
		return nil, err
	}
	sock := os.Getenv("SSH_AUTH_SOCK")
	if sock == "" {
		return nil, fmt.Errorf("未找到 ssh-agent（SSH_AUTH_SOCK 未设置），请先加载密钥")
	}
	conn, err := net.Dial("unix", sock)
	if err != nil {
		return nil, fmt.Errorf("连接 ssh-agent 失败: %w", err)
	}
	return &SSHRunner{
		client:   map[string]*ssh.Client{},
		hostKeys: hk,
		agent:    agent.NewClient(conn),
	}, nil
}

func knownHostsCallback() (ssh.HostKeyCallback, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("无法确定家目录: %w", err)
	}
	path := filepath.Join(home, ".ssh", "known_hosts")
	cb, err := knownhosts.New(path)
	if err != nil {
		return nil, fmt.Errorf("读取 known_hosts 失败（%s）: %w", path, err)
	}
	return cb, nil
}

// connect 返回节点的 SSH 连接，同一节点复用连接。
func (r *SSHRunner) connect(ip string) (*ssh.Client, error) {
	if c, ok := r.client[ip]; ok {
		return c, nil
	}
	cfg := &ssh.ClientConfig{
		User:            SSHUser,
		Auth:            []ssh.AuthMethod{ssh.PublicKeysCallback(r.agent.Signers)},
		HostKeyCallback: r.hostKeys,
		Timeout:         DialTimeout,
	}
	c, err := ssh.Dial("tcp", net.JoinHostPort(ip, SSHPort), cfg)
	if err != nil {
		return nil, fmt.Errorf("SSH 连接失败: %w", err)
	}
	r.client[ip] = c
	return c, nil
}

// Run 在节点上执行白名单命令。只接收命令 ID，参数来自白名单固定列表。
func (r *SSHRunner) Run(ctx context.Context, ip, commandID string) (string, error) {
	argv, err := Argv(commandID)
	if err != nil {
		return "", err
	}
	c, err := r.connect(ip)
	if err != nil {
		return "", err
	}
	sess, err := c.NewSession()
	if err != nil {
		return "", fmt.Errorf("创建会话失败: %w", err)
	}
	defer sess.Close()

	ctx, cancel := context.WithTimeout(ctx, CommandTimeout)
	defer cancel()

	type result struct {
		out []byte
		err error
	}
	done := make(chan result, 1)
	go func() {
		out, err := sess.Output(shellJoin(argv))
		done <- result{out, err}
	}()
	select {
	case res := <-done:
		if res.err != nil {
			return "", fmt.Errorf("命令 %s 执行失败: %w", commandID, res.err)
		}
		return strings.TrimSpace(string(res.out)), nil
	case <-ctx.Done():
		return "", fmt.Errorf("命令 %s 超时", commandID)
	}
}

// Close 关闭所有已建立的连接。
func (r *SSHRunner) Close() {
	for _, c := range r.client {
		c.Close()
	}
}

// shellJoin 将白名单参数逐个用单引号包裹后拼接。参数均为白名单固定值，不含外部输入。
func shellJoin(argv []string) string {
	quoted := make([]string, len(argv))
	for i, a := range argv {
		quoted[i] = "'" + strings.ReplaceAll(a, "'", `'\''`) + "'"
	}
	return strings.Join(quoted, " ")
}
