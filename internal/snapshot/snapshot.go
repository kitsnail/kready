// Package snapshot 读取 docs/design/snapshot-format.md 定义的快照。
// 测试与离线检查只使用快照，不连接集群或节点。
package snapshot

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Cluster 是 cluster.json 的内容。
type Cluster struct {
	Nodes  []Node            `json:"nodes"`
	Pods   []Pod             `json:"pods"`
	Config map[string]string `json:"config"`
}

// Node 是集群中的一个节点。
type Node struct {
	Name    string            `json:"name"`
	Labels  map[string]string `json:"labels"`
	Ready   bool              `json:"ready"`
	Kubelet string            `json:"kubelet"`
}

// Pod 是 Pod 的最小字段。
type Pod struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Ready     bool   `json:"ready"`
	Restarts  int    `json:"restarts"`
}

// Snapshot 是一个完整快照：集群 + 各节点采集结果。
type Snapshot struct {
	Cluster Cluster
	// Nodes 键为节点名，值为白名单命令 ID → 输出。
	Nodes map[string]map[string]string
}

// Load 读取快照目录：cluster.json 与 nodes/<node>.json。
func Load(dir string) (*Snapshot, error) {
	var s Snapshot
	if err := readJSON(filepath.Join(dir, "cluster.json"), &s.Cluster); err != nil {
		return nil, err
	}
	s.Nodes = map[string]map[string]string{}
	entries, err := os.ReadDir(filepath.Join(dir, "nodes"))
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		name := e.Name()[:len(e.Name())-len(".json")]
		var m map[string]string
		if err := readJSON(filepath.Join(dir, "nodes", e.Name()), &m); err != nil {
			return nil, err
		}
		s.Nodes[name] = m
	}
	return &s, nil
}

func readJSON(path string, v any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("读取快照失败: %w", err)
	}
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("快照格式错误 %s: %w", path, err)
	}
	return nil
}
