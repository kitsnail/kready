package collect

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
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

// SSHRunner 通过 SSH 密钥认证与 known_hosts 登录节点并执行白名单命令。
// 认证顺序：ssh-agent（若 SSH_AUTH_SOCK 可用）→ 默认私钥文件。不关闭主机密钥校验。
// 私钥仅用于签名，内容不打印、不写入报告或日志。
type SSHRunner struct {
	client   map[string]*ssh.Client
	hostKeys ssh.HostKeyCallback
	auth     []ssh.AuthMethod
}

// NewSSHRunner 初始化 known_hosts 校验与认证方式。known_hosts 缺失时直接报错，不降级为不校验。
// 认证方式至少要有一种（agent 或默认私钥），否则报错。
func NewSSHRunner() (*SSHRunner, error) {
	hk, err := knownHostsCallback()
	if err != nil {
		return nil, err
	}
	auth := authMethods()
	if len(auth) == 0 {
		return nil, fmt.Errorf("未找到可用的 SSH 密钥：请启动 ssh-agent 或在 ~/.ssh 中放置默认私钥")
	}
	return &SSHRunner{
		client:   map[string]*ssh.Client{},
		hostKeys: hk,
		auth:     auth,
	}, nil
}

// defaultKeyFiles 是按顺序尝试的默认私钥文件名（相对 ~/.ssh）。
var defaultKeyFiles = []string{"id_ed25519", "id_ecdsa", "id_rsa"}

// authMethods 返回可用的认证方式：先 agent，再默认私钥。无法加载的项直接跳过。
func authMethods() []ssh.AuthMethod {
	var methods []ssh.AuthMethod
	if sock := os.Getenv("SSH_AUTH_SOCK"); sock != "" {
		if conn, err := net.Dial("unix", sock); err == nil {
			methods = append(methods, ssh.PublicKeysCallback(agent.NewClient(conn).Signers))
		}
	}
	if home, err := os.UserHomeDir(); err == nil {
		var signers []ssh.Signer
		for _, name := range defaultKeyFiles {
			data, err := os.ReadFile(filepath.Join(home, ".ssh", name))
			if err != nil {
				continue
			}
			signer, err := ssh.ParsePrivateKey(data)
			if err != nil {
				// 有口令或格式不支持的私钥跳过，不输出其内容。
				continue
			}
			signers = append(signers, signer)
		}
		if len(signers) > 0 {
			methods = append(methods, ssh.PublicKeys(signers...))
		}
	}
	return methods
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
		Auth:            r.auth,
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
		// 退出码模式（Result: exit_code）：非 0 退出码是判定数据，不是错误。
		if TakesExitCode(commandID) {
			return exitCodeOf(res.err)
		}
		if res.err != nil {
			return "", fmt.Errorf("命令 %s 执行失败: %w", commandID, res.err)
		}
		return applyExtractor(commandID, strings.TrimSpace(string(res.out)))
	case <-ctx.Done():
		return "", fmt.Errorf("命令 %s 超时", commandID)
	}
}

// exitCoder 是携带远端命令退出码的错误，*ssh.ExitError 满足该接口。
type exitCoder interface{ ExitStatus() int }

// exitCodeOf 把命令执行结果转为退出码字符串（"0"、"9"…）。
// 仅远端命令退出（*ssh.ExitError）转为值；连接等其它失败仍是错误，由调用方记为 ERROR。
func exitCodeOf(err error) (string, error) {
	if err == nil {
		return "0", nil
	}
	var ee exitCoder
	if errors.As(err, &ee) {
		return strconv.Itoa(ee.ExitStatus()), nil
	}
	return "", err
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
