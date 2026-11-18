package main

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/scclabs/kready/internal/check"
	"github.com/scclabs/kready/internal/home"
	"github.com/scclabs/kready/internal/rules"
	"github.com/scclabs/kready/internal/snapshot"
)

// store 持有本次运行使用的 KREADY_HOME 目录，并保证目录结构已就绪。
type store struct {
	root string
}

// openStore 解析并创建 KREADY_HOME 目录结构。失败时在判定前报错，避免运行后才发现无处保存。
func openStore() (*store, error) {
	root, err := home.Root()
	if err != nil {
		return nil, err
	}
	if err := home.Ensure(root); err != nil {
		return nil, err
	}
	return &store{root: root}, nil
}

// saveReport 把报告写入 reports/，返回文件路径。文件名含时间戳，不覆盖旧报告。
func (s *store) saveReport(now time.Time, ext string, content []byte) (string, error) {
	dir := filepath.Join(s.root, home.DirReports)
	name := home.NextName(dir, "check-"+home.Stamp(now), ext)
	path := filepath.Join(dir, name)
	if err := home.CreateNew(path, content); err != nil {
		return "", fmt.Errorf("保存报告失败: %w", err)
	}
	return path, nil
}

// saveSnapshot 把 --live 采集结果写入 snapshots/，目录结构与离线快照一致（可被 --snapshot 复用）。
func (s *store) saveSnapshot(now time.Time, snap *snapshot.Snapshot) (string, error) {
	dir := filepath.Join(s.root, home.DirSnapshots, home.Stamp(now))
	if err := home.Ensure(dir); err != nil {
		return "", err
	}
	if err := snapshot.Save(dir, snap); err != nil {
		return "", err
	}
	return dir, nil
}

// renderText 把文本报告渲染为字节，与终端输出一致。
func renderText(nodeCount int, results []check.Result) []byte {
	var buf bytes.Buffer
	writeReport(&buf, nodeCount, results)
	return buf.Bytes()
}

// renderJSON 把 JSON 报告渲染为字节。
func renderJSON(pack *rules.Pack, results []check.Result) ([]byte, error) {
	var buf bytes.Buffer
	if err := writeJSON(&buf, pack, results); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// tee 同时写入终端与报告内容，保证两者一致。
func tee(w io.Writer, b []byte) error {
	_, err := w.Write(b)
	return err
}
