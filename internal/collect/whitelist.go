// Package collect 通过 SSH 只读采集节点数据（见 docs/decisions/0001-node-collection-via-ssh.md）。
//
// 安全边界：执行层只接收白名单命令 ID，不接收原始命令字符串。
package collect

import "fmt"

// whitelist 将命令 ID 映射为固定的只读命令。规则包中的 collect.command 必须出现在此表中。
// 修改本表属于 B 类改动，需人工确认。
var whitelist = map[string][]string{
	"sysctl-inotify-instances":   {"sysctl", "-n", "fs.inotify.max_user_instances"},
	"sysctl-inotify-watches":     {"sysctl", "-n", "fs.inotify.max_user_watches"},
	"sshd-password-auth":         {"sshd", "-T"},
	"unattended-upgrades-active": {"systemctl", "is-active", "unattended-upgrades"},
}

// Argv 返回白名单命令 ID 对应的固定参数列表。未在白名单中的 ID 返回错误。
func Argv(id string) ([]string, error) {
	argv, ok := whitelist[id]
	if !ok {
		return nil, fmt.Errorf("命令未在白名单中: %q", id)
	}
	out := make([]string, len(argv))
	copy(out, argv)
	return out, nil
}

// IDs 返回白名单中所有命令 ID（排序前的无序列表，用于测试与校验）。
func IDs() []string {
	ids := make([]string, 0, len(whitelist))
	for id := range whitelist {
		ids = append(ids, id)
	}
	return ids
}
