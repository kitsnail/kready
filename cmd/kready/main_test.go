package main

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// testdata 位于仓库根，测试在 cmd/kready 下运行，需回到根目录。
const testdataRoot = "../../testdata"

func runArgs(t *testing.T, args ...string) (int, string, string) {
	t.Helper()
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
	for _, k := range []string{"schema_version", "kready_version", "pack", "rules_rev", "conclusion", "summary", "results"} {
		if _, ok := d[k]; !ok {
			t.Errorf("缺少字段 %s", k)
		}
	}
	if d["pack"] != "2.32" || d["conclusion"] != "正常" {
		t.Errorf("pack/conclusion 不符: %v / %v", d["pack"], d["conclusion"])
	}
	summary, _ := d["summary"].(map[string]any)
	for _, k := range []string{"pass", "fail", "warn", "error", "info", "skip"} {
		if _, ok := summary[k]; !ok {
			t.Errorf("summary 缺少 %s", k)
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
	if !strings.Contains(out, "FAIL  VER-01") {
		t.Errorf("应对 VER-01 报 FAIL，实际: %s", out)
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
