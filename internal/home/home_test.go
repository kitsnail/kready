package home

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// 标准子目录与项目约定一致，不得增减。
func TestSubdirsStandard(t *testing.T) {
	want := []string{"bin", "config", "logs", "reports", "snapshots", "tmp"}
	if len(Subdirs) != len(want) {
		t.Fatalf("子目录 = %v，期望 %v", Subdirs, want)
	}
	for i := range want {
		if Subdirs[i] != want[i] {
			t.Errorf("子目录[%d] = %s，期望 %s", i, Subdirs[i], want[i])
		}
	}
}

// KREADY_HOME 优先于默认值。
func TestRootHonorsEnv(t *testing.T) {
	t.Setenv("KREADY_HOME", "/custom/kready")
	got, err := Root()
	if err != nil || got != "/custom/kready" {
		t.Errorf("Root() = %q, %v；期望 /custom/kready", got, err)
	}
}

// 未设置 KREADY_HOME 时默认 $HOME/.kready。
func TestRootDefaultsToHome(t *testing.T) {
	t.Setenv("KREADY_HOME", "")
	t.Setenv("HOME", "/root")
	got, err := Root()
	if err != nil || got != "/root/.kready" {
		t.Errorf("Root() = %q, %v；期望 /root/.kready", got, err)
	}
}

// Ensure 创建全部子目录，主目录与子目录权限 0700。
func TestEnsureCreatesStandardLayout(t *testing.T) {
	root := filepath.Join(t.TempDir(), ".kready")
	if err := Ensure(root); err != nil {
		t.Fatal(err)
	}
	for _, d := range Subdirs {
		fi, err := os.Stat(filepath.Join(root, d))
		if err != nil {
			t.Fatalf("缺少子目录 %s: %v", d, err)
		}
		if !fi.IsDir() || fi.Mode().Perm() != 0o700 {
			t.Errorf("%s 应为 0700 目录，实际 %v", d, fi.Mode())
		}
	}
}

// 已存在的目录权限不被改动（避免破坏用户已有设置）。
func TestEnsureDoesNotChangeExistingPerm(t *testing.T) {
	root := filepath.Join(t.TempDir(), ".kready")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := Ensure(root); err != nil {
		t.Fatal(err)
	}
	fi, _ := os.Stat(root)
	if fi.Mode().Perm() != 0o755 {
		t.Errorf("已存在目录权限被改动为 %v", fi.Mode().Perm())
	}
}

// CreateNew 不覆盖已有文件。
func TestCreateNewDoesNotOverwrite(t *testing.T) {
	p := filepath.Join(t.TempDir(), "r.txt")
	if err := CreateNew(p, []byte("first")); err != nil {
		t.Fatal(err)
	}
	if err := CreateNew(p, []byte("second")); err == nil {
		t.Fatal("已存在的文件应被拒绝覆盖")
	}
	b, _ := os.ReadFile(p)
	if string(b) != "first" {
		t.Errorf("原文件被覆盖: %q", b)
	}
}

// 同一秒内的多次运行得到不同文件名，旧报告不被覆盖。
func TestNextNameAvoidsCollision(t *testing.T) {
	dir := t.TempDir()
	base := "check-" + Stamp(time.Date(2026, 11, 18, 9, 50, 25, 0, time.UTC))
	first := NextName(dir, base, ".txt")
	if first != base+".txt" {
		t.Fatalf("首个文件名 = %s", first)
	}
	if err := CreateNew(filepath.Join(dir, first), []byte("x")); err != nil {
		t.Fatal(err)
	}
	second := NextName(dir, base, ".txt")
	if second != base+"-2.txt" {
		t.Errorf("冲突后文件名 = %s，期望 %s-2.txt", second, base)
	}
}

// 时间戳为 UTC、可排序的固定格式。
func TestStampFormat(t *testing.T) {
	got := Stamp(time.Date(2026, 11, 18, 9, 50, 25, 0, time.FixedZone("CST", 8*3600)))
	if got != "20261118T015025Z" {
		t.Errorf("Stamp = %s，期望 UTC 时间 20261118T015025Z", got)
	}
}
