// Package collect 通过 SSH 只读采集节点数据（见 docs/decisions/0001-node-collection-via-ssh.md）。
//
// 安全边界：执行层只接收白名单命令 ID，不接收原始命令字符串。
package collect

import "fmt"

// Command 是一条白名单命令：固定参数与取结果方式。
type Command struct {
	// Argv 是固定参数列表，不含 shell 元字符转义，由执行层负责安全拼接。
	Argv []string
	// Result 声明采集结果来源：
	//   "output"    （默认）命令输出（去除首尾空白）
	//   "exit_code" 命令退出码（十进制字符串，非 0 不是错误）
	Result string
}

// whitelist 将命令 ID 映射为固定的只读命令。规则包中的 collect.command 必须出现在此表中。
// 修改本表属于 B 类改动，需人工确认。gpu-nvidia-smi-list 经 D-014 批准。
var whitelist = map[string]Command{
	"sysctl-inotify-instances":   {Argv: []string{"sysctl", "-n", "fs.inotify.max_user_instances"}},
	"sysctl-inotify-watches":     {Argv: []string{"sysctl", "-n", "fs.inotify.max_user_watches"}},
	"sshd-password-auth":         {Argv: []string{"sshd", "-T"}},
	"unattended-upgrades-active": {Argv: []string{"systemctl", "is-active", "unattended-upgrades"}},
	"gpu-nvidia-smi-list":        {Argv: []string{"nvidia-smi", "-L"}, Result: "exit_code"},
}

// Argv 返回白名单命令 ID 对应的固定参数列表。未在白名单中的 ID 返回错误。
func Argv(id string) ([]string, error) {
	cmd, err := lookupCommand(id)
	if err != nil {
		return nil, err
	}
	out := make([]string, len(cmd.Argv))
	copy(out, cmd.Argv)
	return out, nil
}

// TakesExitCode 报告命令 ID 是否以退出码作为采集结果。未在白名单中的 ID 恒为 false，
// 由 Argv 先行校验。
func TakesExitCode(id string) bool {
	return whitelist[id].Result == "exit_code"
}

// lookupCommand 取白名单条目，未注册的 ID 返回错误。
func lookupCommand(id string) (Command, error) {
	cmd, ok := whitelist[id]
	if !ok {
		return Command{}, fmt.Errorf("命令未在白名单中: %q", id)
	}
	return cmd, nil
}

// IDs 返回白名单中所有命令 ID（排序前的无序列表，用于测试与校验）。
func IDs() []string {
	ids := make([]string, 0, len(whitelist))
	for id := range whitelist {
		ids = append(ids, id)
	}
	return ids
}
