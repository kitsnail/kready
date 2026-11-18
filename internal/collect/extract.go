package collect

import (
	"fmt"
	"strings"
)

// extractors 为需要做固定单行提取的白名单命令定义解析规则。
// 命令输出过长时（如 sshd -T 约 95 行）只保留判定所需的一行，格式与规则期望值一致。
// 没有对应条目的命令，输出原样保存。修改本表属于 B 类改动，需人工确认。
var extractors = map[string]func(string) (string, error){
	"sshd-password-auth": extractSSHDKey("passwordauthentication", "PasswordAuthentication"),
}

// extractSSHDKey 从 sshd -T 输出中取出键为 key（小写，忽略大小写匹配）的一行，
// 返回 label + " " + 值。找不到时报错，由调用方记为 ERROR。
func extractSSHDKey(key, label string) func(string) (string, error) {
	return func(out string) (string, error) {
		for _, line := range strings.Split(out, "\n") {
			fields := strings.Fields(line)
			if len(fields) >= 2 && strings.EqualFold(fields[0], key) {
				return label + " " + fields[1], nil
			}
		}
		return "", fmt.Errorf("输出中未找到 %s", key)
	}
}

// applyExtractor 对白名单命令的原始输出做固定提取。没有提取规则时原样返回。
func applyExtractor(commandID, raw string) (string, error) {
	fn, ok := extractors[commandID]
	if !ok {
		return raw, nil
	}
	return fn(raw)
}
