// Command kready 检查 Kubernetes 集群交付就绪状态。
//
// 当前为切片阶段：仅支持读取本地快照执行检查（kready check --snapshot <dir>）。
package main

import (
	"flag"
	"fmt"
	"github.com/scclabs/kready/internal/collect"
	"io"
	"os"
	"time"

	"github.com/scclabs/kready/internal/check"
	"github.com/scclabs/kready/internal/rules"
	"github.com/scclabs/kready/internal/snapshot"
)

// version 在发布时通过 -ldflags 注入。
var version = "dev"

// 退出码，见 docs/design/output-contract.md。
const (
	exitOK   = 0
	exitBad  = 1
	exitTool = 2
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "用法: kready check --snapshot <目录> | kready version")
		return exitTool
	}
	switch args[0] {
	case "version":
		fmt.Fprintln(stdout, "kready", version)
		return exitOK
	case "check":
		return runCheck(args[1:], stdout, stderr)
	}
	fmt.Fprintf(stderr, "未知命令: %s\n", args[0])
	return exitTool
}

func runCheck(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("check", flag.ContinueOnError)
	fs.SetOutput(stderr)
	dir := fs.String("snapshot", "", "快照目录（离线，只读）")
	live := fs.Bool("live", false, "实时采集：kubectl 取节点清单，SSH（root，ssh-agent）只读采集节点")
	format := fs.String("format", "text", "输出格式：text 或 json")
	if err := fs.Parse(args); err != nil {
		return exitTool
	}
	if (*dir == "") == !*live {
		fmt.Fprintln(stderr, "请指定 --snapshot <目录> 或 --live（二选一）")
		return exitTool
	}
	if *format != "text" && *format != "json" {
		fmt.Fprintf(stderr, "未知输出格式: %s（支持 text、json）\n", *format)
		return exitTool
	}
	st, err := openStore()
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitTool
	}
	now := time.Now()
	var snap *snapshot.Snapshot
	var nodes []collect.Node
	if *live {
		s, ns, err := liveSnapshot()
		if err != nil {
			fmt.Fprintln(stderr, err)
			return exitTool
		}
		snap, nodes = s, ns
	} else {
		s, err := snapshot.Load(*dir)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return exitTool
		}
		snap = s
	}
	pack, err := rules.Select(check.ControlPlaneKubelet(snap))
	if err != nil {
		// 无适用规则：不判定，退出码 2，见 output-contract.md。
		fmt.Fprintln(stderr, err)
		return exitTool
	}
	if *live {
		// 实时模式：按所选规则包补采节点命令（选包后才知道需要哪些命令）。
		s, err := collectNodes(snap, nodes, pack)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return exitTool
		}
		snap = s
	}
	results := check.Run(pack, snap)

	// 实时采集的快照默认保存，便于复现判定（snapshots/<时间戳>/）。
	if *live {
		if dir, err := st.saveSnapshot(now, snap); err != nil {
			fmt.Fprintf(stderr, "警告：快照保存失败: %v\n", err)
		} else {
			fmt.Fprintf(stderr, "快照已保存: %s\n", dir)
		}
	}

	// 报告：先输出到终端，再保存到 reports/（默认行为）。保存失败只警告，不改变判定结果。
	var content []byte
	ext := ".txt"
	if *format == "json" {
		ext = ".json"
		b, err := renderJSON(pack, results)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return exitTool
		}
		content = b
	} else {
		content = renderText(len(snap.Cluster.Nodes), results)
	}
	if err := tee(stdout, content); err != nil {
		fmt.Fprintln(stderr, err)
		return exitTool
	}
	if path, err := st.saveReport(now, ext, content); err != nil {
		fmt.Fprintf(stderr, "警告：报告保存失败: %v\n", err)
	} else {
		fmt.Fprintf(stderr, "报告已保存: %s\n", path)
	}

	for _, r := range results {
		if r.Status == check.Fail || r.Status == check.Error {
			return exitBad
		}
	}
	return exitOK
}

// writeReport 输出人类可读结论：第一行为结论，异常时列出 FAIL 与 ERROR，最后列出 INFO。
// label 返回结果的显示名：有标题时为"标题（ID）"，否则只显示 ID。
// 标题在前便于直接阅读，ID 保留以便对照规则包与文档。
func label(r check.Result) string {
	if r.Title == "" {
		return r.ID
	}
	return r.Title + "（" + r.ID + "）"
}

func writeReport(w io.Writer, nodeCount int, results []check.Result) {
	fmt.Fprintln(w, check.Conclusion(results))
	// 如实反映 kubectl 发现的节点数，不与期望值比较（节点数不作为判定项）。
	fmt.Fprintf(w, "节点数: %d\n", nodeCount)
	writeGroups(w, results, check.Fail, "【必须处理】")
	writeGroups(w, results, check.Error, "【未能检查】")
	writeGroups(w, results, check.Info, "【参考信息】")
}

// writeGroups 按规则分组输出同一状态的结果：每条规则一个标题，下面逐行列出各目标。
// 同一规则只出现一次标题，避免按节点重复。
func writeGroups(w io.Writer, results []check.Result, status check.Status, section string) {
	var order []string
	groups := map[string][]check.Result{}
	for _, r := range results {
		if r.Status != status {
			continue
		}
		if _, ok := groups[r.ID]; !ok {
			order = append(order, r.ID)
		}
		groups[r.ID] = append(groups[r.ID], r)
	}
	for _, id := range order {
		rs := groups[id]
		fmt.Fprintf(w, "\n%s %s\n", section, label(rs[0]))
		for _, r := range rs {
			switch status {
			case check.Fail:
				fmt.Fprintf(w, "    %s  实际: %s  期望: %s\n", r.Target, r.Actual, r.Expect)
			case check.Error:
				fmt.Fprintf(w, "    %s  原因: %s\n", r.Target, r.Error)
			case check.Info:
				fmt.Fprintf(w, "    %s  当前: %s\n", r.Target, r.Actual)
			}
		}
		// 影响与处理：规则未填写时显示占位，不编造内容（待逐项研究后补充）。
		if status == check.Fail {
			fmt.Fprintf(w, "    影响: %s\n    处理: %s\n", orPending(rs[0].Impact), orPending(rs[0].Hint))
		}
	}
}

// orPending 为空时返回占位文字，避免输出空的影响或处理。
func orPending(v string) string {
	if v == "" {
		return "（待补充）"
	}
	return v
}
