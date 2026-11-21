package main

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// testdata 位于仓库根，测试在 cmd/kready 下运行，需回到根目录。
const testdataRoot = "../../testdata"

func runArgs(t *testing.T, args ...string) (int, string, string) {
	t.Helper()
	// 测试不得写真实家目录：每次运行使用独立的 KREADY_HOME，除非测试自行设置。
	if os.Getenv("KREADY_HOME") == "" {
		t.Setenv("KREADY_HOME", t.TempDir())
	}
	var out, errb bytes.Buffer
	code := run(args, &out, &errb)
	return code, out.String(), errb.String()
}

// 正例：文本输出首行为正常，退出码 0。
func TestCheckPositiveText(t *testing.T) {
	code, out, _ := runArgs(t, "check", "--snapshot", filepath.Join(testdataRoot, "positive", "2.32"))
	if code != exitOK {
		t.Fatalf("退出码 = %d，期望 %d", code, exitOK)
	}
	if !strings.HasPrefix(out, "正常\n") {
		t.Errorf("首行应为正常，实际输出: %q", out)
	}
}

// 负例：存在 FAIL，退出码 1。
func TestCheckNegativeExitCode(t *testing.T) {
	code, _, _ := runArgs(t, "check", "--snapshot", filepath.Join(testdataRoot, "negative", "2.32", "CP-01"))
	if code != exitBad {
		t.Fatalf("退出码 = %d，期望 %d", code, exitBad)
	}
}

// 正例 JSON：结构字段齐全，summary 含六个计数。
func TestCheckPositiveJSON(t *testing.T) {
	code, out, _ := runArgs(t, "check", "--snapshot", filepath.Join(testdataRoot, "positive", "2.32"), "--format", "json")
	if code != exitOK {
		t.Fatalf("退出码 = %d，期望 %d", code, exitOK)
	}
	var d map[string]any
	if err := json.Unmarshal([]byte(out), &d); err != nil {
		t.Fatalf("JSON 解析失败: %v\n%s", err, out)
	}
	for _, k := range []string{"schema_version", "kready_version", "pack", "rules_rev", "conclusion", "rules", "results_by_status", "results"} {
		if _, ok := d[k]; !ok {
			t.Errorf("缺少字段 %s", k)
		}
	}
	// 锁定输出契约版本：改动结论、退出码或字段语义时必须同步递增并更新此断言。
	if d["schema_version"] != float64(4) {
		t.Errorf("schema_version = %v，期望 4", d["schema_version"])
	}
	if d["pack"] != "2.32" || d["conclusion"] != "正常" {
		t.Errorf("pack/conclusion 不符: %v / %v", d["pack"], d["conclusion"])
	}
	for _, field := range []string{"rules", "results_by_status"} {
		m, _ := d[field].(map[string]any)
		for _, k := range []string{"pass", "fail", "warn", "error", "info", "skip"} {
			if _, ok := m[k]; !ok {
				t.Errorf("%s 缺少 %s", field, k)
			}
		}
	}
}

// 负例 JSON：FAIL 结果带 expect，结论异常。
func TestCheckNegativeJSONHasExpect(t *testing.T) {
	code, out, _ := runArgs(t, "check", "--snapshot", filepath.Join(testdataRoot, "negative", "2.32", "CP-01"), "--format", "json")
	if code != exitBad {
		t.Fatalf("退出码 = %d，期望 %d", code, exitBad)
	}
	if !strings.Contains(out, `"status": "FAIL"`) || !strings.Contains(out, `"expect": "3"`) {
		t.Errorf("FAIL 结果应包含 status 与 expect，实际: %s", out)
	}
}

// 无适用规则：退出码 2，不输出判定结论。
func TestNoApplicablePackExitsTwo(t *testing.T) {
	dir := t.TempDir()
	copyPositive(t, dir)
	rewriteKubelet(t, dir, "v1.35.8")
	code, out, errOut := runArgs(t, "check", "--snapshot", dir)
	if code != exitTool {
		t.Fatalf("退出码 = %d，期望 %d", code, exitTool)
	}
	if out != "" {
		t.Errorf("无适用规则时不应输出判定结果，实际: %q", out)
	}
	if !strings.Contains(errOut, "无适用规则") {
		t.Errorf("应提示无适用规则，实际: %q", errOut)
	}
}

// 控制面版本不一致：无法确定基线，退出码 2。
func TestControlPlaneMixedVersionsExitsTwo(t *testing.T) {
	dir := t.TempDir()
	copyPositive(t, dir)
	rewriteControlPlaneKubelet(t, dir, "cp-1", "v1.36.3")
	code, _, errOut := runArgs(t, "check", "--snapshot", dir)
	if code != exitTool {
		t.Fatalf("退出码 = %d，期望 %d", code, exitTool)
	}
	if !strings.Contains(errOut, "控制面") {
		t.Errorf("应提示控制面版本不一致，实际: %q", errOut)
	}
}

// 工作节点版本与基线不一致：VER-01 FAIL，退出码 1，不是无适用规则。
func TestWorkerMixedVersionFailsVER01(t *testing.T) {
	dir := t.TempDir()
	copyPositive(t, dir)
	rewriteWorkerKubelet(t, dir, "w-1", "v1.36.3")
	code, out, _ := runArgs(t, "check", "--snapshot", dir)
	if code != exitBad {
		t.Fatalf("退出码 = %d，期望 %d", code, exitBad)
	}
	// 新格式：标题在前，ID 在括号中；应标为“不合规”并给出标题与 ID。
	if !strings.Contains(out, "【必须处理】 工作节点 kubelet 版本与控制面一致（VER-01）") {
		t.Errorf("应在【必须处理】下列出 VER-01 并带标题，实际: %s", out)
	}
	// 规则未填写影响与处理时，不显示空占位，避免干扰判断。
	if strings.Contains(out, "（待补充）") {
		t.Errorf("不应输出空占位，实际: %s", out)
	}
}

// 未知输出格式：退出码 2。
func TestUnknownFormatExitsTwo(t *testing.T) {
	code, _, _ := runArgs(t, "check", "--snapshot", filepath.Join(testdataRoot, "positive", "2.32"), "--format", "xml")
	if code != exitTool {
		t.Fatalf("退出码 = %d，期望 %d", code, exitTool)
	}
}

// copyPositive 把正例快照复制到 dst，供测试改写。
func copyPositive(t *testing.T, dst string) {
	t.Helper()
	src := filepath.Join(testdataRoot, "positive", "2.32")
	if err := copyTree(src, dst); err != nil {
		t.Fatalf("复制正例: %v", err)
	}
}

// rewriteKubelet 把 cluster.json 中所有节点的 kubelet 版本设为 version。
func rewriteKubelet(t *testing.T, dir, version string) {
	t.Helper()
	editCluster(t, dir, func(nodes []map[string]any) {
		for _, n := range nodes {
			n["kubelet"] = version
		}
	})
}

// rewriteControlPlaneKubelet 只改指定控制面节点的 kubelet 版本。
func rewriteControlPlaneKubelet(t *testing.T, dir, name, version string) {
	t.Helper()
	editCluster(t, dir, func(nodes []map[string]any) {
		for _, n := range nodes {
			if n["name"] == name {
				n["kubelet"] = version
			}
		}
	})
}

// rewriteWorkerKubelet 只改指定工作节点的 kubelet 版本。
func rewriteWorkerKubelet(t *testing.T, dir, name, version string) {
	t.Helper()
	editCluster(t, dir, func(nodes []map[string]any) {
		for _, n := range nodes {
			if n["name"] == name {
				n["kubelet"] = version
			}
		}
	})
}

func editCluster(t *testing.T, dir string, fn func([]map[string]any)) {
	t.Helper()
	path := filepath.Join(dir, "cluster.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("读取 cluster.json: %v", err)
	}
	var c map[string]any
	if err := json.Unmarshal(data, &c); err != nil {
		t.Fatalf("解析 cluster.json: %v", err)
	}
	var nodes []map[string]any
	for _, raw := range c["nodes"].([]any) {
		nodes = append(nodes, raw.(map[string]any))
	}
	fn(nodes)
	out, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		t.Fatalf("编码 cluster.json: %v", err)
	}
	if err := os.WriteFile(path, out, 0o644); err != nil {
		t.Fatalf("写入 cluster.json: %v", err)
	}
}

// copyTree 递归复制目录，仅用于测试夹具。
func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, p)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
}

// VER-01 FAIL 必须列出不符合基线的工作节点名，便于直接定位。
func TestVER01FailNamesNonconformingWorker(t *testing.T) {
	dir := t.TempDir()
	copyPositive(t, dir)
	rewriteWorkerKubelet(t, dir, "w-1", "v1.36.3")
	_, out, _ := runArgs(t, "check", "--snapshot", dir)
	if !strings.Contains(out, "w-1=v1.36.3") {
		t.Errorf("VER-01 应列出不符合基线的节点 w-1=v1.36.3，实际: %s", out)
	}
	if strings.Contains(out, "w-2=") {
		t.Errorf("符合基线的 w-2 不应被列出，实际: %s", out)
	}
}

// 仅有控制面节点：VER-01 通过（无工作节点可比较），退出码 0。
func TestNoWorkerNodesPassesVER01(t *testing.T) {
	dir := t.TempDir()
	copyPositive(t, dir)
	path := filepath.Join(dir, "cluster.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("读取 cluster.json: %v", err)
	}
	var c map[string]any
	if err := json.Unmarshal(data, &c); err != nil {
		t.Fatalf("解析 cluster.json: %v", err)
	}
	var cp []any
	for _, raw := range c["nodes"].([]any) {
		n := raw.(map[string]any)
		if labels, _ := n["labels"].(map[string]any); len(labels) > 0 {
			cp = append(cp, raw)
		}
	}
	c["nodes"] = cp
	var cpPods []any
	if pods, ok := c["pods"].([]any); ok {
		for _, raw := range pods {
			p := raw.(map[string]any)
			nodeName, _ := p["node"].(string)
			for _, cpNode := range cp {
				if cpNode.(map[string]any)["name"] == nodeName {
					cpPods = append(cpPods, raw)
					break
				}
			}
		}
		c["pods"] = cpPods
	}
	out, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("编码 cluster.json: %v", err)
	}
	if err := os.WriteFile(path, out, 0o644); err != nil {
		t.Fatalf("写入 cluster.json: %v", err)
	}
	code, out2, _ := runArgs(t, "check", "--snapshot", dir, "--format", "json")
	if code != exitOK {
		t.Fatalf("退出码 = %d，期望 %d；输出: %s", code, exitOK, out2)
	}
	if !strings.Contains(out2, "无工作节点") {
		t.Errorf("JSON 应标明无工作节点，实际: %s", out2)
	}
}

// 工作节点版本带空白但与基线一致：不得误判为混布。
func TestPaddedWorkerVersionMatchesBaseline(t *testing.T) {
	dir := t.TempDir()
	copyPositive(t, dir)
	editCluster(t, dir, func(nodes []map[string]any) {
		for _, n := range nodes {
			if n["name"] == "w-1" {
				n["kubelet"] = " v1.36.4 "
			}
		}
	})
	code, _, _ := runArgs(t, "check", "--snapshot", dir)
	if code != exitOK {
		t.Fatalf("退出码 = %d，期望 %d（空白不应导致 FAIL）", code, exitOK)
	}
}

// 工作节点 kubelet 版本为空：VER-01 报 ERROR，结论异常，退出码 1，且不判为版本不符。
func TestEmptyWorkerKubeletIsError(t *testing.T) {
	dir := t.TempDir()
	copyPositive(t, dir)
	rewriteWorkerKubelet(t, dir, "w-2", "")
	code, out, _ := runArgs(t, "check", "--snapshot", dir, "--format", "json")
	if code != exitBad {
		t.Fatalf("退出码 = %d，期望 %d", code, exitBad)
	}
	if !strings.Contains(out, `"status": "ERROR"`) || !strings.Contains(out, "w-2") {
		t.Errorf("应对 w-2 报 ERROR，实际: %s", out)
	}
	if strings.Contains(out, `"status": "FAIL"`) && strings.Contains(out, "w-2=") {
		t.Errorf("空版本不应判为 FAIL，实际: %s", out)
	}
}

// 离线检查默认把报告写入 KREADY_HOME/reports/，内容与终端输出一致。
func TestCheckSavesReportToKreadyHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("KREADY_HOME", home)
	code, out, _ := runArgs(t, "check", "--snapshot", filepath.Join(testdataRoot, "positive", "2.32"))
	if code != exitOK {
		t.Fatalf("退出码 = %d，期望 %d", code, exitOK)
	}
	entries, err := os.ReadDir(filepath.Join(home, "reports"))
	if err != nil || len(entries) != 1 {
		t.Fatalf("reports/ 应有 1 个报告，实际 %v（%v）", entries, err)
	}
	saved, err := os.ReadFile(filepath.Join(home, "reports", entries[0].Name()))
	if err != nil {
		t.Fatal(err)
	}
	if string(saved) != out {
		t.Errorf("保存的报告与终端输出不一致\n保存: %q\n终端: %q", saved, out)
	}
	fi, _ := os.Stat(filepath.Join(home, "reports", entries[0].Name()))
	if fi.Mode().Perm() != 0o600 {
		t.Errorf("报告权限 = %v，期望 0600", fi.Mode().Perm())
	}
}

// 多次运行不覆盖旧报告：每次都新增一个文件。
func TestRepeatedRunsDoNotOverwriteReports(t *testing.T) {
	home := t.TempDir()
	t.Setenv("KREADY_HOME", home)
	for i := 0; i < 3; i++ {
		runArgs(t, "check", "--snapshot", filepath.Join(testdataRoot, "positive", "2.32"))
	}
	entries, _ := os.ReadDir(filepath.Join(home, "reports"))
	if len(entries) != 3 {
		t.Errorf("3 次运行应得到 3 份报告，实际 %d", len(entries))
	}
}

// 目录无法创建（KREADY_HOME/reports 被普通文件占用）：判定前报错，退出码 2，不输出判定。
// 不做判定，避免用户误以为结果已存档。
func TestUnusableHomeFailsBeforeCheck(t *testing.T) {
	home := t.TempDir()
	t.Setenv("KREADY_HOME", home)
	if err := os.WriteFile(filepath.Join(home, "reports"), []byte("blocker"), 0o600); err != nil {
		t.Fatal(err)
	}
	code, out, errOut := runArgs(t, "check", "--snapshot", filepath.Join(testdataRoot, "positive", "2.32"))
	if code != exitTool {
		t.Fatalf("退出码 = %d，期望 %d", code, exitTool)
	}
	if out != "" {
		t.Errorf("目录不可用时不应输出判定，实际: %q", out)
	}
	if errOut == "" {
		t.Error("应说明目录不可用的原因")
	}
}

// 报告写入失败（目录存在但无法新建文件）只影响保存，不改变判定结果。
// 直接测试 saveReport：目标目录不存在时返回错误，调用方据此只警告。
func TestSaveReportErrorIsReturned(t *testing.T) {
	st := &store{root: filepath.Join(t.TempDir(), "missing")}
	if _, err := st.saveReport(time.Now(), ".txt", []byte("x")); err == nil {
		t.Error("目标目录不存在时 saveReport 应返回错误，由调用方降级为警告")
	}
}

// 文本报告第二行如实输出当前节点数，不与期望值比较。
func TestTextReportShowsActualNodeCount(t *testing.T) {
	code, out, _ := runArgs(t, "check", "--snapshot", filepath.Join(testdataRoot, "positive", "2.32"))
	if code != exitOK {
		t.Fatalf("退出码 = %d", code)
	}
	lines := strings.Split(out, "\n")
	if len(lines) < 2 || lines[1] != "节点数: 5" {
		t.Errorf("第二行应为 节点数: 5，实际 %q", lines[1:2])
	}
}

// 节点数只陈述事实：工作节点被移除后，节点数如实为 3，且不产生针对节点数的 FAIL/ERROR。
// editCluster 无法删除节点，这里直接改写 cluster.json（与 TestNoWorkerNodesPassesVER01 一致）。
func TestNodeCountIsNotJudged(t *testing.T) {
	dir := t.TempDir()
	copyPositive(t, dir)
	path := filepath.Join(dir, "cluster.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var c map[string]any
	if err := json.Unmarshal(data, &c); err != nil {
		t.Fatal(err)
	}
	var kept []any
	for _, raw := range c["nodes"].([]any) {
		n := raw.(map[string]any)
		if labels, ok := n["labels"].(map[string]any); ok {
			if _, cp := labels["node-role.kubernetes.io/control-plane"]; cp {
				kept = append(kept, raw)
			}
		}
	}
	c["nodes"] = kept
	var keptPods []any
	if pods, ok := c["pods"].([]any); ok {
		for _, raw := range pods {
			p := raw.(map[string]any)
			nodeName, _ := p["node"].(string)
			for _, cpNode := range kept {
				if cpNode.(map[string]any)["name"] == nodeName {
					keptPods = append(keptPods, raw)
					break
				}
			}
		}
		c["pods"] = keptPods
	}
	out, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, out, 0o644); err != nil {
		t.Fatal(err)
	}

	code, out2, _ := runArgs(t, "check", "--snapshot", dir)
	if !strings.Contains(out2, "节点数: 3") {
		t.Errorf("应如实输出节点数 3，实际: %q", out2)
	}
	// 节点数不是判定项：不得有任何 FAIL/ERROR 行引用“节点数”。
	for _, line := range strings.Split(out2, "\n") {
		if (strings.HasPrefix(line, "FAIL") || strings.HasPrefix(line, "ERROR")) && strings.Contains(line, "节点数") {
			t.Errorf("节点数不应被判定: %q", line)
		}
	}
	// 控制面节点合规，移除工作节点不改变结论：仍为正常、退出码 0。
	if code != exitOK {
		t.Errorf("退出码 = %d，期望 %d（节点数不影响判定）", code, exitOK)
	}
	if !strings.HasPrefix(out2, "正常\n") {
		t.Errorf("结论应为正常，实际: %q", out2)
	}
}

// 相同取值的节点合并为一行，且每个节点名都保留（不丢失“哪台是哪个值”）。
func TestSameValueNodesAreMerged(t *testing.T) {
	dir := t.TempDir()
	copyPositive(t, dir)
	code, out, _ := runArgs(t, "check", "--snapshot", dir)
	if code != exitOK {
		t.Fatalf("退出码 = %d", code)
	}
	// 正例快照中 NODE-05 在所有节点取值相同：应只出现一行，包含全部节点名。
	count := strings.Count(out, "当前: active")
	if count > 1 {
		t.Errorf("相同取值应合并为一行，实际出现 %d 次: %s", count, out)
	}
	for _, name := range []string{"cp-1", "cp-2", "cp-3", "w-1", "w-2"} {
		if !strings.Contains(out, name) {
			t.Errorf("合并后应保留节点名 %s，实际: %s", name, out)
		}
	}
}

// 工具问题与集群问题在输出中分开：工具未覆盖单独成块并注明“不代表集群有问题”。
func TestToolGapIsSeparatedFromCollectFailure(t *testing.T) {
	// 实时模式下 config.* 无数据来源：删除快照中的 config 字段，模拟工具未覆盖。
	dir := t.TempDir()
	copyPositive(t, dir)
	path := filepath.Join(dir, "cluster.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var c map[string]any
	if err := json.Unmarshal(data, &c); err != nil {
		t.Fatal(err)
	}
	delete(c, "config")
	out2, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, out2, 0o644); err != nil {
		t.Fatal(err)
	}

	code, out, _ := runArgs(t, "check", "--snapshot", dir)
	if code != exitOK {
		t.Fatalf("退出码 = %d，工具未覆盖不应阻断", code)
	}
	if !strings.HasPrefix(out, "正常\n") {
		t.Errorf("工具未覆盖不影响结论，应为正常，实际: %q", out)
	}
	if !strings.Contains(out, "【工具未覆盖】") || !strings.Contains(out, "不代表集群有问题") {
		t.Errorf("工具未覆盖应单独成块并注明边界，实际: %s", out)
	}
	if strings.Contains(out, "【未能确认】") {
		t.Errorf("无采集失败时不应出现【未能确认】，实际: %s", out)
	}
}

// 采集失败：结论为未能确认，退出码 1，且不被写成正常。
func TestCollectFailureGivesUnconfirmedAndExitOne(t *testing.T) {
	dir := t.TempDir()
	copyPositive(t, dir)
	// 删除 cp-2 的采集文件，模拟节点采集失败。
	if err := os.Remove(filepath.Join(dir, "nodes", "cp-2.json")); err != nil {
		t.Fatal(err)
	}
	code, out, _ := runArgs(t, "check", "--snapshot", dir)
	if code != exitBad {
		t.Errorf("退出码 = %d，采集失败应为 %d", code, exitBad)
	}
	if !strings.HasPrefix(out, "未能确认\n") {
		t.Errorf("结论应为未能确认，实际: %q", out)
	}
	if !strings.Contains(out, "【未能确认】") || !strings.Contains(out, "请检查节点连通性") {
		t.Errorf("应单独成块并说明处理方式，实际: %s", out)
	}
}

// 同一原因的工具未覆盖项合并为一行，不逐条占块。
func TestToolGapSameReasonMergedToOneLine(t *testing.T) {
	dir := t.TempDir()
	copyPositive(t, dir)
	path := filepath.Join(dir, "cluster.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var c map[string]any
	if err := json.Unmarshal(data, &c); err != nil {
		t.Fatal(err)
	}
	delete(c, "config")
	out2, _ := json.MarshalIndent(c, "", "  ")
	if err := os.WriteFile(path, out2, 0o644); err != nil {
		t.Fatal(err)
	}
	_, out, _ := runArgs(t, "check", "--snapshot", dir)
	if n := strings.Count(out, "本模式不采集"); n != 1 {
		t.Errorf("同一原因应合并为一行，实际出现 %d 次: %s", n, out)
	}
	for _, name := range []string{"etcd 版本（VER-04）", "kube-proxy 模式（VER-06）", "etcd 后端配额（CP-08）", "metrics-server 副本数（SYS-04）"} {
		if !strings.Contains(out, name) {
			t.Errorf("合并后应保留 %s，实际: %s", name, out)
		}
	}
}

// --pack gpu：GPU 正例快照全部通过，输出正常，退出码 0（REQ-001）。
func TestCheckPackGPU(t *testing.T) {
	code, out, _ := runArgs(t, "check", "--snapshot", filepath.Join(testdataRoot, "positive", "gpu"), "--pack", "gpu")
	if code != exitOK {
		t.Fatalf("退出码 = %d，期望 %d", code, exitOK)
	}
	if !strings.HasPrefix(out, "正常\n") {
		t.Fatalf("首行应为正常，实际输出: %q", out)
	}
}

// --pack gpu 负例：GPU-02 FAIL（可分配数与标签不一致），退出码 1。
func TestCheckPackGPUNegative(t *testing.T) {
	code, out, _ := runArgs(t, "check", "--snapshot", filepath.Join(testdataRoot, "negative", "gpu", "GPU-02"), "--pack", "gpu")
	if code != exitBad {
		t.Fatalf("退出码 = %d，期望 %d", code, exitBad)
	}
	if !strings.Contains(out, "GPU-02") {
		t.Errorf("应列出 GPU-02 的失败，实际: %s", out)
	}
}

// --pack gpu 无 GPU 节点：所有规则 SKIP，不判定，结论正常，退出码 0（REQ-001）。
func TestCheckPackGPUNoGPUNodes(t *testing.T) {
	code, out, _ := runArgs(t, "check", "--snapshot", filepath.Join(testdataRoot, "positive", "2.32"), "--pack", "gpu")
	if code != exitOK {
		t.Fatalf("退出码 = %d，期望 %d", code, exitOK)
	}
	if !strings.HasPrefix(out, "正常\n") {
		t.Fatalf("无 GPU 节点应 SKIP 且结论正常，实际: %q", out)
	}
}

// --pack 指定未知规则包：无适用规则，退出码 2，不做判定。
func TestCheckPackUnknown(t *testing.T) {
	code, out, errb := runArgs(t, "check", "--snapshot", filepath.Join(testdataRoot, "positive", "2.32"), "--pack", "nope")
	if code != 2 {
		t.Fatalf("退出码 = %d，期望 2", code)
	}
	if out != "" {
		t.Fatalf("不应输出判定结论，实际: %q", out)
	}
	if !strings.Contains(errb, "无适用规则") {
		t.Fatalf("错误信息应包含「无适用规则」，实际: %q", errb)
	}
}
