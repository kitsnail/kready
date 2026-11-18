// Package check 按规则对快照求值，输出结果。
package check

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/scclabs/kready/internal/rules"
	"github.com/scclabs/kready/internal/snapshot"
)

// Status 是结果状态，见 docs/design/output-contract.md。
type Status string

const (
	Pass  Status = "PASS"
	Fail  Status = "FAIL"
	Warn  Status = "WARN"
	Error Status = "ERROR"
	Info  Status = "INFO"
	Skip  Status = "SKIP"
)

// Source 说明 ERROR 的性质，决定它是否影响结论。
//   - SourceCluster：集群本身的结果（PASS/FAIL/INFO），不适用于 ERROR。
//   - SourceToolGap：工具在当前模式下没有数据来源（如实时模式不采集 config.*），不是集群问题，不影响结论。
//   - SourceCollect：工具尝试采集或判定时失败（如 SSH 失败、节点数据缺失），无法确认集群状态，影响结论为"未能确认"。
type Source string

const (
	SourceCluster Source = "cluster"
	SourceToolGap Source = "tool-gap"
	SourceCollect Source = "collect"
)

// Result 是一条判定结果（一个目标一条）。
type Result struct {
	ID     string `json:"id"`
	Title  string `json:"title,omitempty"`
	Status Status `json:"status"`
	Source Source `json:"source,omitempty"`
	Target string `json:"target"`
	Actual string `json:"actual"`
	Expect string `json:"expect,omitempty"`
	Error  string `json:"error,omitempty"`
	Impact string `json:"impact,omitempty"`
	Hint   string `json:"hint,omitempty"`
}

// Run 对快照执行所有 active 规则。结果顺序与规则包顺序一致，节点按名排序。
func Run(p *rules.Pack, s *snapshot.Snapshot) []Result {
	var out []Result
	for _, r := range p.Active() {
		switch r.Scope {
		case "cluster":
			if r.Check == "equal_baseline" {
				out = append(out, withTitle(runBaseline(r, s), r))
				continue
			}
			out = append(out, withTitle(runCluster(r, s), r))
		case "node":
			for _, res := range runNodes(r, s) {
				out = append(out, withTitle(res, r))
			}
		default:
			out = append(out, Result{ID: r.ID, Status: Error, Target: r.Scope, Error: "未知 scope"})
		}
	}
	return out
}

// withTitle 把规则的标题、影响与处理写入结果。影响与处理为空时保持为空，不编造内容。
func withTitle(res Result, r rules.Rule) Result {
	res.Title = r.Title
	res.Impact = r.Impact
	res.Hint = r.Hint
	return res
}

func runCluster(r rules.Rule, s *snapshot.Snapshot) Result {
	actual := clusterValue(r, s)
	expect := r.Expect["value"]
	res := Result{ID: r.ID, Target: "cluster", Actual: actual, Expect: expect}
	// 缺值一律 ERROR，信息项也不例外，避免输出空的"当前值="造成误判。
	if actual == "" {
		res.Status, res.Error, res.Source = Error, "本模式不采集", SourceToolGap
		return res
	}
	switch {
	case r.Mode == "info":
		res.Status = Info
	case actual == expect:
		res.Status = Pass
	default:
		res.Status = Fail
	}
	return res
}

// clusterValue 从快照取出集群级实际值。取数方式由规则的 field 声明，代码只支持有限种类：
//
//	config.<key>   读取 cluster.json 的 config[key]
//	kubelet_versions 全部节点 kubelet 版本（去重）
//	count_role     按 expect.role 计数节点数
func clusterValue(r rules.Rule, s *snapshot.Snapshot) string {
	switch {
	case r.Field == "count_role":
		return strconv.Itoa(countNodes(s, r.Expect["role"]))
	case strings.HasPrefix(r.Field, "config."):
		return s.Cluster.Config[strings.TrimPrefix(r.Field, "config.")]
	}
	return ""
}

func runNodes(r rules.Rule, s *snapshot.Snapshot) []Result {
	var out []Result
	for _, n := range sortedNodeNames(s) {
		res := Result{ID: r.ID, Target: n, Expect: r.Expect["value"]}
		val, ok := s.Nodes[n][r.Collect.Command]
		if !ok {
			res.Status, res.Error, res.Source = Error, "采集失败: "+r.Collect.Command, SourceCollect
			out = append(out, res)
			continue
		}
		res.Actual = strings.TrimSpace(val)
		switch {
		case r.Mode == "info":
			res.Status = Info
		case res.Actual == res.Expect:
			res.Status = Pass
		default:
			res.Status = Fail
		}
		out = append(out, res)
	}
	return out
}

func countNodes(s *snapshot.Snapshot, role string) int {
	key := "node-role.kubernetes.io/" + role
	c := 0
	for _, n := range s.Cluster.Nodes {
		if _, ok := n.Labels[key]; ok {
			c++
		}
	}
	return c
}

// sortedNodeNames 以集群声明的节点为准（cluster.nodes），按名排序。
// 采集文件缺失的节点仍会列出，由 runNodes 报 ERROR，避免漏判后输出"正常"。
func sortedNodeNames(s *snapshot.Snapshot) []string {
	seen := map[string]bool{}
	var names []string
	for _, n := range s.Cluster.Nodes {
		if !seen[n.Name] {
			seen[n.Name] = true
			names = append(names, n.Name)
		}
	}
	sort.Strings(names)
	return names
}

// Summary 统计各状态数量。
func Summary(rs []Result) map[Status]int {
	m := map[Status]int{}
	for _, r := range rs {
		m[r.Status]++
	}
	return m
}

// Conclusion 只描述集群，结论三种：
//   - 异常：存在 FAIL（确认不合规）。
//   - 未能确认：无 FAIL，但存在采集失败（SourceCollect 的 ERROR），无法确认集群状态。
//   - 正常：无 FAIL，且全部判定项已采集。
//
// 工具未覆盖（SourceToolGap）不影响结论。
func Conclusion(rs []Result) string {
	hasCollect := false
	for _, r := range rs {
		if r.Status == Fail {
			return "异常"
		}
		if r.Status == Error && r.Source == SourceCollect {
			hasCollect = true
		}
	}
	if hasCollect {
		return "未能确认"
	}
	return "正常"
}

// Blocking 报告结论是否使退出码为非 0：异常与未能确认都视为未通过。
func Blocking(rs []Result) bool {
	c := Conclusion(rs)
	return c != "正常"
}

// String 便于调试输出。
func (r Result) String() string {
	return fmt.Sprintf("%s %s %s actual=%q expect=%q", r.ID, r.Status, r.Target, r.Actual, r.Expect)
}

const controlPlaneLabel = "node-role.kubernetes.io/control-plane"

func isControlPlane(n snapshot.Node) bool {
	_, ok := n.Labels[controlPlaneLabel]
	return ok
}

// ControlPlaneKubelet 返回控制面节点的 kubelet 版本（去重，排序）。选包以此为基线。
func ControlPlaneKubelet(s *snapshot.Snapshot) []string {
	return distinctKubelet(s, true)
}

// WorkerKubelet 返回非控制面节点的 kubelet 版本（去重，排序）。
func WorkerKubelet(s *snapshot.Snapshot) []string {
	return distinctKubelet(s, false)
}

func distinctKubelet(s *snapshot.Snapshot, controlPlane bool) []string {
	seen := map[string]bool{}
	var vals []string
	for _, n := range s.Cluster.Nodes {
		if isControlPlane(n) != controlPlane {
			continue
		}
		if !seen[n.Kubelet] {
			seen[n.Kubelet] = true
			vals = append(vals, n.Kubelet)
		}
	}
	sort.Strings(vals)
	return vals
}

// runBaseline 判定工作节点 kubelet 版本与控制面基线一致。
// 基线取控制面版本；无工作节点时视为通过。
func runBaseline(r rules.Rule, s *snapshot.Snapshot) Result {
	res := Result{ID: r.ID, Target: "cluster"}
	if r.Expect["baseline"] != "control-plane" {
		res.Status, res.Error, res.Source = Error, "规则配置错误，不支持的基线: "+r.Expect["baseline"], SourceToolGap
		return res
	}
	cp := ControlPlaneKubelet(s)
	if len(cp) != 1 {
		res.Status, res.Error, res.Source = Error, "控制面节点版本缺失或不一致，无法确定基线", SourceCollect
		return res
	}
	res.Expect = cp[0]
	workers := WorkerKubelet(s)
	if len(workers) == 0 {
		res.Actual, res.Status = "无工作节点", Pass
		return res
	}
	// 工作节点缺少 kubelet 版本说明采集不完整，报 ERROR，不判为版本不符。
	var missing []string
	for _, n := range s.Cluster.Nodes {
		if !isControlPlane(n) && n.Kubelet == "" {
			missing = append(missing, n.Name)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		res.Status, res.Error, res.Source = Error, "工作节点缺少 kubelet 版本: "+strings.Join(missing, ","), SourceCollect
		return res
	}
	// 列出与基线不符的工作节点，便于直接定位。
	var bad []string
	for _, n := range s.Cluster.Nodes {
		if !isControlPlane(n) && n.Kubelet != cp[0] {
			bad = append(bad, n.Name+"="+n.Kubelet)
		}
	}
	if len(bad) == 0 {
		res.Actual, res.Status = cp[0], Pass
		return res
	}
	sort.Strings(bad)
	res.Actual = strings.Join(bad, ",")
	res.Status = Fail
	return res
}
