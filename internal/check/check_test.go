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

// GPU-05 采集数据缺失：报 ERROR（source=collect），结论为未能确认，不判 FAIL（REQ-002）。
func TestGPUCollectMissingIsUnconfirmed(t *testing.T) {
	p := loadPack(t, "gpu")
	s, err := snapshot.Load(filepath.Join(testdataRoot, "positive", "gpu"))
	if err != nil {
		t.Fatalf("加载正例快照: %v", err)
	}
	delete(s.Nodes, "gpu-1")

	results := check.Run(p, s)
	found := false
	for _, r := range results {
		if r.ID == "GPU-05" && r.Target == "gpu-1" && r.Status == check.Error {
			found = true
		}
	}
	if !found {
		t.Errorf("gpu-1 缺采集数据应对 GPU-05 报 ERROR；结果: %v", results)
	}
	if got := check.Conclusion(results); got != "未能确认" {
		t.Errorf("结论 = %q，期望 未能确认", got)
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

// GPU-01 多选择器（REQ-003，D-015）：两种部署方式的标签均纳入判定。
// 正例快照中 gpu-1 的 plugin Pod 用 Helm 标签、gpu-2 用 GPU Operator 标签，两节点都应 PASS。
func TestGPU01SelectorsMatchEitherDeployment(t *testing.T) {
	p := loadPack(t, "gpu")
	s, err := snapshot.Load(filepath.Join(testdataRoot, "positive", "gpu"))
	if err != nil {
		t.Fatalf("加载正例快照: %v", err)
	}
	pass := map[string]bool{}
	for _, r := range check.Run(p, s) {
		if r.ID != "GPU-01" {
			continue
		}
		if r.Status != check.Pass {
			t.Errorf("GPU-01 %s = %v（实际 %q），期望 PASS（两种部署标签均应匹配）", r.Target, r.Status, r.Actual)
		}
		pass[r.Target] = true
	}
	if !pass["gpu-1"] || !pass["gpu-2"] {
		t.Errorf("GPU-01 应在 gpu-1、gpu-2 上各输出一条结果，实际覆盖 %v", pass)
	}
}

// GPU-01 两种标签都无匹配时判 FAIL（未部署）（REQ-003 标准 2）。
func TestGPU01NoMatchingPodsIsFail(t *testing.T) {
	p := loadPack(t, "gpu")
	s, err := snapshot.Load(filepath.Join(testdataRoot, "positive", "gpu"))
	if err != nil {
		t.Fatalf("加载正例快照: %v", err)
	}
	for i := range s.Cluster.Pods {
		s.Cluster.Pods[i].Labels = map[string]string{"app": "unrelated"}
	}
	for _, r := range check.Run(p, s) {
		if r.ID != "GPU-01" {
			continue
		}
		if r.Status != check.Fail || r.Actual != "0/0" {
			t.Errorf("GPU-01 %s = %v（实际 %q），期望 FAIL 0/0（未部署）", r.Target, r.Status, r.Actual)
		}
	}
}

// all_ready 选择器列表非法（非 key=value）时报 ERROR（tool-gap）。
func TestAllReadyInvalidSelectorIsError(t *testing.T) {
	p := &rules.Pack{Pack: "test", Rules: []rules.Rule{{
		ID: "T-01", Title: "测试", Status: "active", Mode: "judge", Scope: "node",
		Field: "pods", Check: "all_ready",
		Expect: map[string]string{"selectors": "bad"},
	}}}
	s, err := snapshot.Load(filepath.Join(testdataRoot, "positive", "gpu"))
	if err != nil {
		t.Fatalf("加载正例快照: %v", err)
	}
	results := check.Run(p, s)
	if len(results) != 1 {
		t.Fatalf("应输出一条结果，得到 %d", len(results))
	}
	if r := results[0]; r.Status != check.Error || r.Source != check.SourceToolGap {
		t.Errorf("非法选择器应报 ERROR（tool-gap），得到 %+v", r)
	}
}

// REQ-004: CP-02 etcd 成员数为 3 判 PASS，非 3 判 FAIL。
func TestCP02EtcdMemberCount(t *testing.T) {
	p := &rules.Pack{Pack: "2.32", Rules: []rules.Rule{{
		ID: "CP-02", Title: "etcd 成员数", Status: "active", Mode: "judge", Scope: "cluster",
		Field: "pods", Check: "count",
		Expect: map[string]string{"value": "3", "selectors": "component=etcd,k8s-app=etcd", "pod_prefix": "etcd-"},
	}}}
	s := &snapshot.Snapshot{
		Cluster: snapshot.Cluster{
			Pods: []snapshot.Pod{
				{Namespace: "kube-system", Name: "etcd-cp-1", Labels: map[string]string{"component": "etcd"}},
				{Namespace: "kube-system", Name: "etcd-cp-2", Labels: map[string]string{"component": "etcd"}},
				{Namespace: "kube-system", Name: "etcd-cp-3", Labels: map[string]string{"component": "etcd"}},
			},
		},
	}
	results := check.Run(p, s)
	if len(results) != 1 || results[0].Status != check.Pass || results[0].Actual != "3" {
		t.Fatalf("3 个 etcd 成员应 PASS，得到: %+v", results)
	}

	// 负例：只有 2 个 etcd 成员 -> FAIL
	s.Cluster.Pods = s.Cluster.Pods[:2]
	results = check.Run(p, s)
	if len(results) != 1 || results[0].Status != check.Fail || results[0].Actual != "2" {
		t.Fatalf("2 个 etcd 成员应 FAIL，得到: %+v", results)
	}
}

// REQ-004: CP-03 apiserver / kcm / scheduler Pod 数各等于控制面节点数且全部 Ready。
func TestCP03ControlPlanePodsReady(t *testing.T) {
	p := &rules.Pack{Pack: "2.32", Rules: []rules.Rule{{
		ID: "CP-03", Title: "apiserver / kcm / scheduler Pod 数与就绪状态", Status: "active", Mode: "judge", Scope: "cluster",
		Field: "pods", Check: "control_plane_pods_ready",
		Expect: map[string]string{"role": "control-plane", "components": "kube-apiserver,kube-controller-manager,kube-scheduler"},
	}}}
	s := &snapshot.Snapshot{
		Cluster: snapshot.Cluster{
			Nodes: []snapshot.Node{
				{Name: "cp-1", Labels: map[string]string{"node-role.kubernetes.io/control-plane": ""}},
				{Name: "cp-2", Labels: map[string]string{"node-role.kubernetes.io/control-plane": ""}},
				{Name: "cp-3", Labels: map[string]string{"node-role.kubernetes.io/control-plane": ""}},
			},
			Pods: []snapshot.Pod{
				{Namespace: "kube-system", Name: "kube-apiserver-cp-1", Labels: map[string]string{"component": "kube-apiserver"}, Ready: true},
				{Namespace: "kube-system", Name: "kube-apiserver-cp-2", Labels: map[string]string{"component": "kube-apiserver"}, Ready: true},
				{Namespace: "kube-system", Name: "kube-apiserver-cp-3", Labels: map[string]string{"component": "kube-apiserver"}, Ready: true},
				{Namespace: "kube-system", Name: "kube-controller-manager-cp-1", Labels: map[string]string{"component": "kube-controller-manager"}, Ready: true},
				{Namespace: "kube-system", Name: "kube-controller-manager-cp-2", Labels: map[string]string{"component": "kube-controller-manager"}, Ready: true},
				{Namespace: "kube-system", Name: "kube-controller-manager-cp-3", Labels: map[string]string{"component": "kube-controller-manager"}, Ready: true},
				{Namespace: "kube-system", Name: "kube-scheduler-cp-1", Labels: map[string]string{"component": "kube-scheduler"}, Ready: true},
				{Namespace: "kube-system", Name: "kube-scheduler-cp-2", Labels: map[string]string{"component": "kube-scheduler"}, Ready: true},
				{Namespace: "kube-system", Name: "kube-scheduler-cp-3", Labels: map[string]string{"component": "kube-scheduler"}, Ready: true},
			},
		},
	}
	results := check.Run(p, s)
	if len(results) != 1 || results[0].Status != check.Pass {
		t.Fatalf("控制面核心组件全部 3/3 Ready 应 PASS，得到: %+v", results)
	}

	// 负例 1: kcm 未全部 Ready -> FAIL
	s.Cluster.Pods[5].Ready = false
	results = check.Run(p, s)
	if len(results) != 1 || results[0].Status != check.Fail {
		t.Fatalf("kcm 存在未 Ready 时应 FAIL，得到: %+v", results)
	}

	// 负例 2: 缺少 1 个 apiserver Pod -> FAIL
	s.Cluster.Pods[5].Ready = true
	s.Cluster.Pods = append(s.Cluster.Pods[:1], s.Cluster.Pods[2:]...)
	results = check.Run(p, s)
	if len(results) != 1 || results[0].Status != check.Fail {
		t.Fatalf("缺少 apiserver Pod 时应 FAIL，得到: %+v", results)
	}
}

// REQ-004: CP-04 kube-vip Pod 数等于控制面节点数。
func TestCP04KubeVipCountEqControlPlane(t *testing.T) {
	p := &rules.Pack{Pack: "2.32", Rules: []rules.Rule{{
		ID: "CP-04", Title: "kube-vip Pod 数", Status: "active", Mode: "judge", Scope: "cluster",
		Field: "pods", Check: "count_eq_nodes",
		Expect: map[string]string{"role": "control-plane", "selectors": "app.kubernetes.io/name=kube-vip,app=kube-vip", "pod_prefix": "kube-vip-"},
	}}}
	s := &snapshot.Snapshot{
		Cluster: snapshot.Cluster{
			Nodes: []snapshot.Node{
				{Name: "cp-1", Labels: map[string]string{"node-role.kubernetes.io/control-plane": ""}},
				{Name: "cp-2", Labels: map[string]string{"node-role.kubernetes.io/control-plane": ""}},
				{Name: "cp-3", Labels: map[string]string{"node-role.kubernetes.io/control-plane": ""}},
			},
			Pods: []snapshot.Pod{
				{Namespace: "kube-system", Name: "kube-vip-cp-1", Labels: map[string]string{"app.kubernetes.io/name": "kube-vip"}},
				{Namespace: "kube-system", Name: "kube-vip-cp-2", Labels: map[string]string{"app.kubernetes.io/name": "kube-vip"}},
				{Namespace: "kube-system", Name: "kube-vip-cp-3", Labels: map[string]string{"app.kubernetes.io/name": "kube-vip"}},
			},
		},
	}
	results := check.Run(p, s)
	if len(results) != 1 || results[0].Status != check.Pass || results[0].Actual != "3" {
		t.Fatalf("3 个 kube-vip Pod 等于 3 个控制面节点应 PASS，得到: %+v", results)
	}

	// 负例：只有 2 个 kube-vip Pod -> FAIL
	s.Cluster.Pods = s.Cluster.Pods[:2]
	results = check.Run(p, s)
	if len(results) != 1 || results[0].Status != check.Fail || results[0].Actual != "2" || results[0].Expect != "3" {
		t.Fatalf("2 个 kube-vip Pod 不等 3 应 FAIL，得到: %+v", results)
	}
}

// REQ-004: CP-05 nginx-proxy Pod 数等于工作节点数。
func TestCP05NginxProxyCountEqWorkers(t *testing.T) {
	p := &rules.Pack{Pack: "2.32", Rules: []rules.Rule{{
		ID: "CP-05", Title: "nginx-proxy Pod 数", Status: "active", Mode: "judge", Scope: "cluster",
		Field: "pods", Check: "count_eq_nodes",
		Expect: map[string]string{"role": "worker", "selectors": "k8s-app=nginx-proxy,app.kubernetes.io/name=nginx-proxy", "pod_prefix": "nginx-proxy-"},
	}}}
	s := &snapshot.Snapshot{
		Cluster: snapshot.Cluster{
			Nodes: []snapshot.Node{
				{Name: "cp-1", Labels: map[string]string{"node-role.kubernetes.io/control-plane": ""}},
				{Name: "cp-2", Labels: map[string]string{"node-role.kubernetes.io/control-plane": ""}},
				{Name: "w-1", Labels: map[string]string{}},
				{Name: "w-2", Labels: map[string]string{}},
			},
			Pods: []snapshot.Pod{
				{Namespace: "kube-system", Name: "nginx-proxy-w-1", Labels: map[string]string{"k8s-app": "nginx-proxy"}},
				{Namespace: "kube-system", Name: "nginx-proxy-w-2", Labels: map[string]string{"k8s-app": "nginx-proxy"}},
			},
		},
	}
	results := check.Run(p, s)
	if len(results) != 1 || results[0].Status != check.Pass || results[0].Actual != "2" {
		t.Fatalf("2 个 nginx-proxy Pod 等于 2 个工作节点应 PASS，得到: %+v", results)
	}

	// 负例：只有 1 个 nginx-proxy Pod -> FAIL
	s.Cluster.Pods = s.Cluster.Pods[:1]
	results = check.Run(p, s)
	if len(results) != 1 || results[0].Status != check.Fail || results[0].Actual != "1" || results[0].Expect != "2" {
		t.Fatalf("1 个 nginx-proxy Pod 不等 2 应 FAIL，得到: %+v", results)
	}
}
