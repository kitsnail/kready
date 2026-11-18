package collect

import (
	"strings"
	"testing"
)

// 真实 sshd -T 输出的节选（约 95 行，此处保留目标行与若干干扰行）。
const sshdSample = `port 22
permitrootlogin yes
passwordauthentication yes
pubkeyauthentication yes
kbdinteractiveauthentication no
`

// 从整段 sshd -T 输出中只取 passwordauthentication，格式为 "PasswordAuthentication <值>"。
func TestSSHDPasswordAuthExtractsSingleLine(t *testing.T) {
	got, err := applyExtractor("sshd-password-auth", sshdSample)
	if err != nil {
		t.Fatal(err)
	}
	if got != "PasswordAuthentication yes" {
		t.Errorf("提取结果 = %q，期望 %q", got, "PasswordAuthentication yes")
	}
}

// 键名大小写不敏感（sshd -T 输出为小写，但不依赖此假设）。
func TestSSHDExtractIsCaseInsensitive(t *testing.T) {
	got, err := applyExtractor("sshd-password-auth", "PasswordAuthentication no\n")
	if err != nil {
		t.Fatal(err)
	}
	if got != "PasswordAuthentication no" {
		t.Errorf("提取结果 = %q", got)
	}
}

// 输出中缺少该键：报错，不得返回整段输出或空值。
func TestSSHDMissingKeyIsError(t *testing.T) {
	_, err := applyExtractor("sshd-password-auth", "port 22\npermitrootlogin yes\n")
	if err == nil {
		t.Error("缺少 passwordauthentication 应报错")
	}
}

// 没有提取规则的命令，输出原样保留。
func TestUnextractedCommandIsUnchanged(t *testing.T) {
	got, err := applyExtractor("sysctl-inotify-instances", "8192")
	if err != nil || got != "8192" {
		t.Errorf("= %q, %v；期望原样 8192", got, err)
	}
}

// 提取后的结果应与规则期望值一致（NODE-04 的 expect 为 "PasswordAuthentication no"）。
func TestSSHDExtractMatchesRuleExpectFormat(t *testing.T) {
	got, _ := applyExtractor("sshd-password-auth", "passwordauthentication no\n")
	if !strings.HasPrefix(got, "PasswordAuthentication ") {
		t.Errorf("格式应与规则期望一致，实际 %q", got)
	}
}
