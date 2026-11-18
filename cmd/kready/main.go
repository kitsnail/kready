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
	if *format == "json" {
		if err := writeJSON(stdout, pack, results); err != nil {
			fmt.Fprintln(stderr, err)
			return exitTool
		}
	} else {
		writeReport(stdout, results)
	}
	for _, r := range results {
		if r.Status == check.Fail || r.Status == check.Error {
			return exitBad
		}
	}
	return exitOK
}

// writeReport 输出人类可读结论：第一行为结论，异常时列出 FAIL 与 ERROR，最后列出 INFO。
func writeReport(w io.Writer, results []check.Result) {
	fmt.Fprintln(w, check.Conclusion(results))
	for _, r := range results {
		switch r.Status {
		case check.Fail:
			fmt.Fprintf(w, "FAIL  %s  %s  实际=%s  期望=%s\n", r.ID, r.Target, r.Actual, r.Expect)
		case check.Error:
			fmt.Fprintf(w, "ERROR %s  %s  %s\n", r.ID, r.Target, r.Error)
		}
	}
	for _, r := range results {
		if r.Status == check.Info {
			fmt.Fprintf(w, "INFO  %s  %s  当前值=%s\n", r.ID, r.Target, r.Actual)
		}
	}
}
