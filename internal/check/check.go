// Package check 按规则对快照求值，输出结果。
package check

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/kitsnail/kready/internal/rules"
	"github.com/kitsnail/kready/internal/snapshot"
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
			if r.Check == "control_plane_pods_ready" {
				out = append(out, withTitle(runControlPlanePodsReady(r, s), r))
				continue
			}
			if r.Check == "count_eq_nodes" {
				out = append(out, withTitle(runCountEqNodes(r, s), r))
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
//	pods           按 expect.selectors / pod_prefix 计数 Pod 数
func clusterValue(r rules.Rule, s *snapshot.Snapshot) string {
	switch {
	case r.Field == "count_role":
		return strconv.Itoa(countNodes(s, r.Expect["role"]))
	case r.Field == "pods" && r.Check == "count":
		return strconv.Itoa(countMatchingPods(s, r))
	case strings.HasPrefix(r.Field, "config."):
		return s.Cluster.Config[strings.TrimPrefix(r.Field, "config.")]
	}
	return ""
}

func countMatchingPods(s *snapshot.Snapshot, r rules.Rule) int {
	var sels [][2]string
	if selsStr := r.Expect["selectors"]; selsStr != "" {
		sels, _ = parseSelectors(selsStr)
	}
	prefix := r.Expect["pod_prefix"]
	ns := r.Expect["namespace"]
	if ns == "" {
		ns = "kube-system"
	}
	count := 0
	for _, p := range s.Cluster.Pods {
		if p.Namespace != "" && p.Namespace != ns {
			continue
		}
		if matchPod(p, sels, prefix) {
			count++
		}
	}
	return count
}

func matchPod(p snapshot.Pod, sels [][2]string, prefix string) bool {
	if len(sels) > 0 && matchAnySelector(p.Labels, sels) {
		return true
	}
	if prefix != "" && strings.HasPrefix(p.Name, prefix) {
		return true
	}
	return false
}

func runCountEqNodes(r rules.Rule, s *snapshot.Snapshot) Result {
	res := Result{ID: r.ID, Target: "cluster"}
	role := r.Expect["role"]
	expectedCount := 0
	switch role {
	case "control-plane":
		expectedCount = countNodes(s, "control-plane")
	case "worker":
		expectedCount = countWorkerNodes(s)
	default:
		res.Status, res.Error, res.Source = Error, "未知节点角色: "+role, SourceToolGap
		return res
	}
	res.Expect = strconv.Itoa(expectedCount)
	actualCount := countMatchingPods(s, r)
	res.Actual = strconv.Itoa(actualCount)
	if actualCount == expectedCount {
		res.Status = Pass
	} else {
		res.Status = Fail
	}
	return res
}

func countWorkerNodes(s *snapshot.Snapshot) int {
	c := 0
	for _, n := range s.Cluster.Nodes {
		if !isControlPlane(n) {
			c++
		}
	}
	return c
}

func runControlPlanePodsReady(r rules.Rule, s *snapshot.Snapshot) Result {
	res := Result{ID: r.ID, Target: "cluster"}
	cpCount := countNodes(s, "control-plane")
	res.Expect = fmt.Sprintf("各 %d 个全部 Ready", cpCount)

	components := []string{"kube-apiserver", "kube-controller-manager", "kube-scheduler"}
	if compStr := r.Expect["components"]; compStr != "" {
		components = strings.Split(compStr, ",")
	}

	allOk := true
	var parts []string

	for _, comp := range components {
		total := 0
		ready := 0
		for _, p := range s.Cluster.Pods {
			if p.Namespace != "" && p.Namespace != "kube-system" {
				continue
			}
			if p.Labels["component"] == comp || strings.HasPrefix(p.Name, comp+"-") {
				total++
				if p.Ready {
					ready++
				}
			}
		}
		if total != cpCount || ready != cpCount {
			allOk = false
		}
		parts = append(parts, fmt.Sprintf("%s: %d/%d Ready", comp, ready, total))
	}

	if allOk && cpCount > 0 {
		res.Actual = fmt.Sprintf("各 %d 个全部 Ready", cpCount)
		res.Status = Pass
	} else {
		res.Actual = strings.Join(parts, ", ")
		res.Status = Fail
	}
	return res
}

func runNodes(r rules.Rule, s *snapshot.Snapshot) []Result {
	targets := nodeTargets(r, s)
	if len(targets) == 0 && r.NodeLabel != "" {
		// 规则声明了节点标签但无匹配节点：不适用，不判定（SKIP 不影响结论）。
		return []Result{{ID: r.ID, Status: Skip, Target: "cluster", Actual: "无匹配节点（node_label: " + r.NodeLabel + "）"}}
	}
	if r.Check == "all_ready" {
		return runPodsReady(r, s, targets)
	}
	var out []Result
	for _, n := range targets {
		if r.Collect != nil {
			// 采集数据：缺键为 ERROR（采集失败）；空输出按相等比较，保持既有语义。
			res := Result{ID: r.ID, Target: n, Expect: r.Expect["value"]}
			val, ok := s.Nodes[n][r.Collect.Command]
			if !ok {
				res.Status, res.Error, res.Source = Error, "采集失败: "+r.Collect.Command, SourceCollect
				out = append(out, res)
				continue
			}
			res.Actual = strings.TrimSpace(val)
			if r.Mode == "info" {
				res.Status = Info
			} else if res.Actual == res.Expect {
				res.Status = Pass
			} else {
				res.Status = Fail
			}
			out = append(out, res)
			continue
		}
		out = append(out, judgeNodeField(r, nodeByName(s, n), n))
	}
	return out
}

// judgeNodeField 判定取自集群数据（field 声明）的节点级规则。
// 集群数据缺失（字段为空）时无法判定，报 ERROR 而不是 FAIL；
// set 例外：缺失就是被测的失败本身。
func judgeNodeField(r rules.Rule, n snapshot.Node, name string) Result {
	res := Result{ID: r.ID, Target: name, Actual: nodeValue(r, n)}
	switch {
	case r.Mode == "info":
		res.Status = Info
	case r.Check == "set":
		res.Expect = "存在且非空"
		if res.Actual != "" {
			res.Status = Pass
		} else {
			res.Status, res.Actual = Fail, "（缺失）"
		}
	case r.Check == "equals_label":
		// equals_label：实际值等于同节点的另一标签值，期望值逐节点取自快照。
		expect := n.Labels[r.Expect["label"]]
		if expect == "" {
			res.Status, res.Error, res.Source = Error, "对比标签缺失: "+r.Expect["label"], SourceCollect
			return res
		}
		res.Expect = expect
		switch {
		case res.Actual == "":
			res.Status, res.Error, res.Source = Error, "实际值缺失: "+r.Field, SourceCollect
		case res.Actual == expect:
			res.Status = Pass
		default:
			res.Status = Fail
		}
	default:
		res.Expect = r.Expect["value"]
		switch {
		case res.Actual == "":
			res.Status, res.Error, res.Source = Error, "实际值缺失: "+r.Field, SourceCollect
		case res.Actual == res.Expect:
			res.Status = Pass
		default:
			res.Status = Fail
		}
	}
	return res
}

// nodeByName 在快照的节点列表中查找节点；名字来自 cluster.nodes，必定命中。
func nodeByName(s *snapshot.Snapshot, name string) snapshot.Node {
	for _, n := range s.Cluster.Nodes {
		if n.Name == name {
			return n
		}
	}
	return snapshot.Node{}
}

// nodeTargets 返回规则作用的目标节点（按名排序）。规则声明 node_label 时只取匹配节点。
func nodeTargets(r rules.Rule, s *snapshot.Snapshot) []string {
	var names []string
	for _, n := range sortedNodeNames(s) {
		if r.NodeLabel == "" || nodeHasLabel(nodeByName(s, n), r.NodeLabel) {
			names = append(names, n)
		}
	}
	return names
}

// nodeHasLabel 判断节点是否带 key=value 标签。
func nodeHasLabel(n snapshot.Node, kv string) bool {
	key, value, ok := strings.Cut(kv, "=")
	if !ok {
		return false
	}
	v, found := n.Labels[key]
	return found && v == value
}

// nodeValue 从集群数据取节点级实际值。取值方式由规则的 field 声明：
//
//	label.<key>        节点标签值
//	allocatable.<key>  节点可分配资源量
//	taints             节点污点列表（空为“无”，不视为缺失）
func nodeValue(r rules.Rule, n snapshot.Node) string {
	switch {
	case r.Field == "taints":
		return formatTaints(n)
	case strings.HasPrefix(r.Field, "label."):
		return n.Labels[strings.TrimPrefix(r.Field, "label.")]
	case strings.HasPrefix(r.Field, "allocatable."):
		return n.Allocatable[strings.TrimPrefix(r.Field, "allocatable.")]
	}
	return ""
}

// formatTaints 把污点列表格式化为 key=value:effect，逗号分隔；无污点为“无”。
func formatTaints(n snapshot.Node) string {
	if len(n.Taints) == 0 {
		return "无"
	}
	parts := make([]string, len(n.Taints))
	for i, t := range n.Taints {
		parts[i] = t.Key + "=" + t.Value + ":" + t.Effect
	}
	sort.Strings(parts)
	return strings.Join(parts, ",")
}

// runPodsReady 判定目标节点上匹配任一选择器的 Pod 全部 Ready。
// 实际值为“Ready 数/总数”；目标节点上一个 Pod 都没有视为 FAIL（未部署）。
func runPodsReady(r rules.Rule, s *snapshot.Snapshot, targets []string) []Result {
	sels, err := parseSelectors(r.Expect["selectors"])
	if err != nil {
		return []Result{{ID: r.ID, Status: Error, Target: "cluster", Error: err.Error(), Source: SourceToolGap}}
	}
	var out []Result
	for _, name := range targets {
		res := Result{ID: r.ID, Target: name, Expect: "全部 Ready"}
		total, ready := 0, 0
		for _, p := range s.Cluster.Pods {
			if p.Node != name || !matchAnySelector(p.Labels, sels) {
				continue
			}
			total++
			if p.Ready {
				ready++
			}
		}
		res.Actual = strconv.Itoa(ready) + "/" + strconv.Itoa(total)
		if total > 0 && ready == total {
			res.Status = Pass
		} else {
			res.Status = Fail
		}
		out = append(out, res)
	}
	return out
}

// parseSelectors 解析逗号分隔的 key=value 选择器列表，Pod 标签匹配任一即纳入判定。
// k8s 标签的键与值都不允许逗号，分隔符无歧义。
func parseSelectors(s string) ([][2]string, error) {
	if s == "" {
		return nil, fmt.Errorf("规则配置错误，选择器列表为空")
	}
	var out [][2]string
	for _, part := range strings.Split(s, ",") {
		key, value, ok := strings.Cut(part, "=")
		if !ok || key == "" {
			return nil, fmt.Errorf("规则配置错误，选择器须为 key=value: %s", part)
		}
		out = append(out, [2]string{key, value})
	}
	return out, nil
}

// matchAnySelector 判断标签是否匹配任一选择器。
func matchAnySelector(labels map[string]string, sels [][2]string) bool {
	for _, kv := range sels {
		if labels[kv[0]] == kv[1] {
			return true
		}
	}
	return false
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
// Summary 按结果条数计数（一个节点一条即一条结果）。
func Summary(rs []Result) map[Status]int {
	m := map[Status]int{}
	for _, r := range rs {
		m[r.Status]++
	}
	return m
}

// ruleSeverity 给每条结果排序严重程度，数值越大越严重。
// 与结论一致：FAIL 最严重；采集失败次之；工具未覆盖、信息、通过依次降低。
func ruleSeverity(r Result) int {
	switch {
	case r.Status == Fail:
		return 5
	case r.Status == Error && r.Source == SourceCollect:
		return 4
	case r.Status == Error:
		return 3
	case r.Status == Warn:
		return 2
	case r.Status == Info:
		return 1
	}
	return 0
}

// RuleStatus 按规则计数：同一规则的多条结果取最严重的状态，每条规则计一次。
// 用于决策：工程师看到的是“几件事需要处理”，而不是节点级的条数。
func RuleStatus(rs []Result) map[Status]int {
	worst := map[string]Result{}
	var order []string
	for _, r := range rs {
		cur, ok := worst[r.ID]
		if !ok {
			order = append(order, r.ID)
			worst[r.ID] = r
			continue
		}
		if ruleSeverity(r) > ruleSeverity(cur) {
			worst[r.ID] = r
		}
	}
	m := map[Status]int{}
	for _, id := range order {
		m[worst[id].Status]++
	}
	return m
}

// Conclusion 只描述集群，结论四种：
//   - 未运行任何规则：所选规则包无 active 规则（结果为空）。不是异常，也不是工具错误。
//   - 异常：存在 FAIL（确认不合规）。
//   - 未能确认：无 FAIL，但存在采集失败（SourceCollect 的 ERROR），无法确认集群状态。
//   - 正常：无 FAIL，且全部判定项已采集。
//
// 工具未覆盖（SourceToolGap）与 SKIP 不影响结论。
func Conclusion(rs []Result) string {
	if len(rs) == 0 {
		return "未运行任何规则"
	}
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

// Blocking 报告结论是否使退出码为非 0：异常与未能确认阻断；未运行任何规则与正常不阻断。
func Blocking(rs []Result) bool {
	c := Conclusion(rs)
	return c == "异常" || c == "未能确认"
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
