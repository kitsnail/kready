package snapshot_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kitsnail/kready/internal/snapshot"
)

// kubelet 版本首尾空白应被去除，同一版本不因空白被视为不同版本。
func TestLoadTrimsKubeletVersion(t *testing.T) {
	dir := t.TempDir()
	cluster := `{"nodes":[{"name":"w-1","kubelet":" v1.36.4 "}],"config":{}}`
	if err := os.WriteFile(filepath.Join(dir, "cluster.json"), []byte(cluster), 0o644); err != nil {
		t.Fatal(err)
	}
	s, err := snapshot.Load(dir)
	if err != nil {
		t.Fatalf("加载快照: %v", err)
	}
	if got := s.Cluster.Nodes[0].Kubelet; got != "v1.36.4" {
		t.Errorf("kubelet = %q，期望 v1.36.4", got)
	}
}
