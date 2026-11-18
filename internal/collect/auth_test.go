package collect

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/crypto/ssh"
)

// 生成一个临时私钥（仅测试用），写入 home/.ssh/id_ed25519。
func writeTestKey(t *testing.T, home string) string {
	t.Helper()
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	block, err := ssh.MarshalPrivateKey(priv, "")
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(home, ".ssh")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "id_ed25519")
	if err := os.WriteFile(path, pem.EncodeToMemory(block), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// 没有 agent、只有默认私钥文件时，仍能得到认证方式（不依赖 ssh-agent）。
func TestAuthFromDefaultKeyWithoutAgent(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("SSH_AUTH_SOCK", "")
	writeTestKey(t, home)
	if got := authMethods(); len(got) == 0 {
		t.Fatal("有默认私钥时应返回认证方式，不应要求 ssh-agent")
	}
}

// 既无 agent 也无私钥：报错，且提示中不包含任何密钥内容。
func TestNoAuthGivesActionableError(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("SSH_AUTH_SOCK", "")
	if err := os.MkdirAll(filepath.Join(home, ".ssh"), 0o700); err != nil {
		t.Fatal(err)
	}
	// known_hosts 必须存在，否则会先因 known_hosts 报错，无法测到认证分支。
	if err := os.WriteFile(filepath.Join(home, ".ssh", "known_hosts"), []byte(""), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := NewSSHRunner()
	if err == nil {
		t.Fatal("无认证方式时应报错")
	}
	if strings.Contains(err.Error(), "PRIVATE") {
		t.Errorf("错误信息不得包含私钥内容: %v", err)
	}
}

// 私钥文件存在但无法解析（如损坏）时跳过，不输出其内容。
func TestUnparseableKeyIsSkippedSilently(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("SSH_AUTH_SOCK", "")
	dir := filepath.Join(home, ".ssh")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "id_ed25519"), []byte("not a key"), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := authMethods(); len(got) != 0 {
		t.Errorf("无法解析的私钥不应产生认证方式，实际 %d 个", len(got))
	}
}
