package check_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kitsnail/kready/internal/check"
	"github.com/kitsnail/kready/internal/rules"
	"github.com/kitsnail/kready/internal/snapshot"
)

// testdata 位于仓库根，测试在 internal/check 下运行，需回到根目录。
const testdataRoot = "../../testdata"

// allPacks 是参与样本测试的全部规则包。新增规则包时在此登记。
func allPacks() []string {
	return []string{"2.32", "gpu"}
}

func loadPack(t *testing.T, name string) *rules.Pack {
	t.Helper()
	p, err := rules.Load(name)
	if err != nil {
		t.Fatalf("加载规则包 %s: %v", name, err)
	}
	return p
}

// 正例：全部 active 规则通过，结论为正常。
func TestPositiveSnapshotIsNormal(t *testing.T) {
	for _, name := range allPacks() {
		t.Run(name, func(t *testing.T) {
			p := loadPack(t, name)
			s, err := snapshot.Load(filepath.Join(testdataRoot, "positive", name))
			if err != nil {
				t.Fatalf("加载正例快照: %v", err)
			}
			results := check.Run(p, s)
			if got := check.Conclusion(results); got != "正常" {
				t.Errorf("结论 = %q，期望 正常；结果: %v", got, results)
			}
			for _, r := range results {
				if r.Status == check.Fail || r.Status == check.Error {
					t.Errorf("正例不应出现 %s: %v", r.Status, r)
				}
			}
		})
	}
}

// 负例：每条 active judge 规则的负例快照必须报出该规则的 FAIL，结论为异常。
func TestNegativeSnapshotsReportTheirRule(t *testing.T) {
	for _, name := range allPacks() {
		t.Run(name, func(t *testing.T) {
			p := loadPack(t, name)
			for _, rule := range p.Active() {
				if rule.Mode != "judge" {
					continue
				}
				dir := filepath.Join(testdataRoot, "negative", name, rule.ID)
				if _, err := os.Stat(dir); err != nil {
					t.Errorf("规则 %s 缺少负例快照: %v", rule.ID, err)
					continue
				}
				t.Run(rule.ID, func(t *testing.T) {
					s, err := snapshot.Load(dir)
					if err != nil {
						t.Fatalf("加载负例: %v", err)
					}
					results := check.Run(p, s)
					if got := check.Conclusion(results); got != "异常" {
						t.Errorf("结论 = %q，期望 异常", got)
					}
					found := false
					for _, r := range results {
						if r.ID == rule.ID && r.Status == check.Fail {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("负例未报 %s 的 FAIL；结果: %v", rule.ID, results)
					}
				})
			}
		})
	}
}

// 规则 ID 不重复（AGENTS.md：规则 ID 不复用）。
func TestRuleIDsUnique(t *testing.T) {
	seen := map[string]bool{}
	for _, name := range allPacks() {
		for _, r := range loadPack(t, name).Rules {
			if seen[r.ID] {
				t.Errorf("规则 ID 重复: %s", r.ID)
			}
			seen[r.ID] = true
		}
	}
}

// 规则包无 active 规则：结果为空，结论「未运行任何规则」，退出码不阻断（REQ-001 标准 5）。
func TestNoActiveRulesRunsNothing(t *testing.T) {
	p := loadPack(t, "gpu")
	// 只留 draft 规则，模拟无 active 规则的规则包。
	for i := range p.Rules {
		p.Rules[i].Status = "draft"
	}
	results := check.Run(p, &snapshot.Snapshot{})
	if len(results) != 0 {
		t.Fatalf("无 active 规则时结果应为空，实际: %v", results)
	}
	if got := check.Conclusion(results); got != "未运行任何规则" {
		t.Errorf("结论 = %q，期望 未运行任何规则", got)
	}
	if check.Blocking(results) {
		t.Error("未运行任何规则不应阻断")
	}
}

// 声明了 node_label 的规则在无匹配节点时输出 SKIP，不影响结论（REQ-001）。
func TestNodeLabelNoMatchIsSkip(t *testing.T) {
	p := loadPack(t, "gpu")
	// 复用 2.32 正例快照：其中没有 nvidia.com/gpu.present 节点。
	s, err := snapshot.Load(filepath.Join(testdataRoot, "positive", "2.32"))
	if err != nil {
		t.Fatalf("加载快照: %v", err)
	}
	results := check.Run(p, s)
	if len(results) == 0 {
		t.Fatal("无匹配节点时应输出 SKIP，而不是空结果")
	}
	for _, r := range results {
		if r.Status != check.Skip {
			t.Errorf("无 GPU 节点时所有规则应为 SKIP，得到 %v", r)
		}
	}
	if got := check.Conclusion(results); got != "正常" {
		t.Errorf("结论 = %q，期望 正常（SKIP 不影响结论）", got)
	}
}

// 信息项缺值必须报 ERROR，不得输出空的当前值。
func TestInfoMissingValueIsError(t *testing.T) {
	p := loadPack(t, "2.32")
	s, err := snapshot.Load(filepath.Join(testdataRoot, "positive", "2.32"))
	if err != nil {
		t.Fatalf("加载正例快照: %v", err)
	}
	delete(s.Cluster.Config, "metrics_server_replicas")
	delete(s.Nodes["cp-1"], "unattended-upgrades-active")

	results := check.Run(p, s)
	var clusterErr, nodeErr bool
	for _, r := range results {
		if r.ID == "SYS-04" && r.Status == check.Error {
			clusterErr = true
		}
		if r.ID == "NODE-05" && r.Target == "cp-1" && r.Status == check.Error {
			nodeErr = true
		}
	}
	if !clusterErr {
		t.Errorf("集群级信息项缺值应为 ERROR；结果: %v", results)
	}
	if !nodeErr {
		t.Errorf("节点级信息项缺值应为 ERROR；结果: %v", results)
	}
}

// 节点采集文件缺失必须报 ERROR，不得因漏判输出"正常"。
func TestMissingNodeFileIsError(t *testing.T) {
	p := loadPack(t, "2.32")
	s, err := snapshot.Load(filepath.Join(testdataRoot, "positive", "2.32"))
	if err != nil {
		t.Fatalf("加载正例快照: %v", err)
	}
	delete(s.Nodes, "cp-2")

	results := check.Run(p, s)
	// 采集失败不是确认的不合规，但也不能输出“正常”：结论为“未能确认”。
	if got := check.Conclusion(results); got != "未能确认" {
		t.Errorf("结论 = %q，期望 未能确认（cp-2 缺采集数据，不得判为正常）", got)
	}
	found := false
	for _, r := range results {
		if r.ID == "NODE-02" && r.Target == "cp-2" && r.Status == check.Error {
			found = true
		}
	}
	if !found {
		t.Errorf("缺失节点 cp-2 应对 NODE-02 报 ERROR；结果: %v", results)
	}
}
