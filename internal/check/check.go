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
	Error Status = "ERROR"
	Info  Status = "INFO"
)

// Result 是一条判定结果（一个目标一条）。
type Result struct {
	ID     string `json:"id"`
	Status Status `json:"status"`
	Target string `json:"target"`
	Actual string `json:"actual"`
	Expect string `json:"expect,omitempty"`
	Error  string `json:"error,omitempty"`
}

// Run 对快照执行所有 active 规则。结果顺序与规则包顺序一致，节点按名排序。
func Run(p *rules.Pack, s *snapshot.Snapshot) []Result {
	var out []Result
	for _, r := range p.Active() {
		switch r.Scope {
		case "cluster":
			out = append(out, runCluster(r, s))
		case "node":
			out = append(out, runNodes(r, s)...)
		default:
			out = append(out, Result{ID: r.ID, Status: Error, Target: r.Scope, Error: "未知 scope"})
		}
	}
	return out
}

func runCluster(r rules.Rule, s *snapshot.Snapshot) Result {
	actual := clusterValue(r, s)
	expect := r.Expect["value"]
	res := Result{ID: r.ID, Target: "cluster", Actual: actual, Expect: expect}
	// 缺值一律 ERROR，信息项也不例外，避免输出空的"当前值="造成误判。
	if actual == "" {
		res.Status, res.Error = Error, "快照中缺少该字段"
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
	case r.Field == "kubelet_versions":
		return kubeletVersions(s)
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
			res.Status, res.Error = Error, "采集失败或命令未注册: "+r.Collect.Command
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

// kubeletVersions 返回所有节点的 kubelet 版本，全部一致时为该版本，否则逗号连接。
func kubeletVersions(s *snapshot.Snapshot) string {
	seen := map[string]bool{}
	var vals []string
	for _, n := range s.Cluster.Nodes {
		if !seen[n.Kubelet] {
			seen[n.Kubelet] = true
			vals = append(vals, n.Kubelet)
		}
	}
	return strings.Join(vals, ",")
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

func sortedNodeNames(s *snapshot.Snapshot) []string {
	names := make([]string, 0, len(s.Nodes))
	for n := range s.Nodes {
		names = append(names, n)
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

// Conclusion 按输出契约给出结论：无 FAIL 与 ERROR 为正常，否则异常。
func Conclusion(rs []Result) string {
	for _, r := range rs {
		if r.Status == Fail || r.Status == Error {
			return "异常"
		}
	}
	return "正常"
}

// String 便于调试输出。
func (r Result) String() string {
	return fmt.Sprintf("%s %s %s actual=%q expect=%q", r.ID, r.Status, r.Target, r.Actual, r.Expect)
}
