package snapshot

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Save 把快照写入 dir，结构与 Load 读取的离线快照一致：cluster.json 与 nodes/<name>.json。
// 文件权限 0600；dir 须已存在。
func Save(dir string, s *Snapshot) error {
	if err := writeJSON(filepath.Join(dir, "cluster.json"), s.Cluster); err != nil {
		return err
	}
	nodesDir := filepath.Join(dir, "nodes")
	if err := os.MkdirAll(nodesDir, 0o700); err != nil {
		return fmt.Errorf("创建 nodes 目录失败: %w", err)
	}
	for name, data := range s.Nodes {
		if err := writeJSON(filepath.Join(nodesDir, name+".json"), data); err != nil {
			return err
		}
	}
	return nil
}

func writeJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("编码 %s 失败: %w", path, err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("写入 %s 失败: %w", path, err)
	}
	return nil
}
