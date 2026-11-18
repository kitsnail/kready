package snapshot

import (
	"os"
	"path/filepath"
	"testing"
)

// 保存的快照可被 Load 读回，且内容一致（复现判定的前提）。
func TestSaveRoundTrip(t *testing.T) {
	src, err := Load(filepath.Join("..", "..", "testdata", "positive", "2.32"))
	if err != nil {
		t.Fatalf("加载正例快照: %v", err)
	}
	dir := t.TempDir()
	if err := Save(dir, src); err != nil {
		t.Fatalf("保存快照: %v", err)
	}
	back, err := Load(dir)
	if err != nil {
		t.Fatalf("读回快照: %v", err)
	}
	if len(back.Cluster.Nodes) != len(src.Cluster.Nodes) {
		t.Errorf("节点数 %d，期望 %d", len(back.Cluster.Nodes), len(src.Cluster.Nodes))
	}
	if len(back.Nodes) != len(src.Nodes) {
		t.Errorf("节点采集数 %d，期望 %d", len(back.Nodes), len(src.Nodes))
	}
	for name, data := range src.Nodes {
		for k, v := range data {
			if back.Nodes[name][k] != v {
				t.Errorf("%s/%s = %q，期望 %q", name, k, back.Nodes[name][k], v)
			}
		}
	}
	if back.Cluster.Nodes[0].Kubelet != src.Cluster.Nodes[0].Kubelet {
		t.Errorf("kubelet 未保留")
	}
}

// 保存的文件权限为 0600。
func TestSavePermissions(t *testing.T) {
	src, err := Load(filepath.Join("..", "..", "testdata", "positive", "2.32"))
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := Save(dir, src); err != nil {
		t.Fatal(err)
	}
	fi, err := os.Stat(filepath.Join(dir, "cluster.json"))
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode().Perm() != 0o600 {
		t.Errorf("cluster.json 权限 = %v，期望 0600", fi.Mode().Perm())
	}
}
