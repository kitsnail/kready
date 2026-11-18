// Package home 定义 kready 的本地目录标准（见 docs/design/kready-home.md）。
//
// 所有部署、配置、日志、报告与临时数据统一放在 KREADY_HOME 下（默认 $HOME/.kready）。
package home

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// 子目录名为项目标准，不得随意更改。
const (
	DirBin       = "bin"
	DirConfig    = "config"
	DirLogs      = "logs"
	DirReports   = "reports"
	DirSnapshots = "snapshots"
	DirTmp       = "tmp"
)

// Subdirs 是标准的全部子目录。
var Subdirs = []string{DirBin, DirConfig, DirLogs, DirReports, DirSnapshots, DirTmp}

// Root 返回 kready 主目录：KREADY_HOME 优先，否则 $HOME/.kready。
func Root() (string, error) {
	if v := os.Getenv("KREADY_HOME"); v != "" {
		return v, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("无法确定家目录: %w", err)
	}
	return filepath.Join(home, ".kready"), nil
}

// Ensure 创建主目录与全部子目录，新建目录权限 0700。已存在的目录不改权限。
func Ensure(root string) error {
	if err := os.MkdirAll(root, 0o700); err != nil {
		return fmt.Errorf("创建 %s 失败: %w", root, err)
	}
	for _, d := range Subdirs {
		p := filepath.Join(root, d)
		if err := os.MkdirAll(p, 0o700); err != nil {
			return fmt.Errorf("创建 %s 失败: %w", p, err)
		}
	}
	return nil
}

// Stamp 返回用于文件名的时间戳（UTC，可排序，精确到秒）。
func Stamp(t time.Time) string {
	return t.UTC().Format("20060102T150405Z")
}

// CreateNew 以 0600 新建文件并写入内容。文件已存在时报错，绝不覆盖。
// 同一秒内的多次运行通过 Name 的序号后缀避让，见 NextName。
func CreateNew(path string, data []byte) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

// NextName 在 dir 中为 base（不含扩展名）+ ext 找一个未被占用的文件名。
// 已存在时依次追加 -2、-3……，保证不覆盖旧报告。
func NextName(dir, base, ext string) string {
	name := base + ext
	for i := 2; ; i++ {
		if _, err := os.Stat(filepath.Join(dir, name)); os.IsNotExist(err) {
			return name
		}
		name = fmt.Sprintf("%s-%d%s", base, i, ext)
	}
}
