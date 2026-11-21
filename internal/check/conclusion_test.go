package check

import "testing"

// 三种结论的口径：只描述集群，工具问题不进入结论。
func TestConclusionThreeStates(t *testing.T) {
	cases := []struct {
		name string
		rs   []Result
		want string
	}{
		{"全部通过", []Result{{Status: Pass}}, "正常"},
		{"确认不合规", []Result{{Status: Fail}}, "异常"},
		{"采集失败", []Result{{Status: Error, Source: SourceCollect}}, "未能确认"},
		{"工具未覆盖不影响结论", []Result{{Status: Error, Source: SourceToolGap}, {Status: Pass}}, "正常"},
		{"FAIL 优先于采集失败", []Result{{Status: Error, Source: SourceCollect}, {Status: Fail}}, "异常"},
		{"工具未覆盖与通过并存", []Result{{Status: Info}, {Status: Error, Source: SourceToolGap}}, "正常"},
		{"未运行任何规则", nil, "未运行任何规则"},
		{"SKIP 不影响结论", []Result{{Status: Skip}}, "正常"},
	}
	for _, c := range cases {
		if got := Conclusion(c.rs); got != c.want {
			t.Errorf("%s：结论 = %q，期望 %q", c.name, got, c.want)
		}
	}
}

// 退出码：异常与未能确认均阻断；正常与仅工具未覆盖不阻断。
func TestBlockingFollowsConclusion(t *testing.T) {
	if Blocking([]Result{{Status: Pass}}) {
		t.Error("全部通过不应阻断")
	}
	if Blocking([]Result{{Status: Error, Source: SourceToolGap}}) {
		t.Error("工具未覆盖不应阻断")
	}
	if Blocking(nil) {
		t.Error("未运行任何规则不应阻断")
	}
	if Blocking([]Result{{Status: Skip}}) {
		t.Error("SKIP 不应阻断")
	}
	if !Blocking([]Result{{Status: Fail}}) {
		t.Error("确认不合规应阻断")
	}
	if !Blocking([]Result{{Status: Error, Source: SourceCollect}}) {
		t.Error("采集失败应阻断（未能确认）")
	}
}

// 按规则计数：同一规则的多个节点结果只计一次，取最严重状态。
func TestRuleStatusCountsEachRuleOnce(t *testing.T) {
	rs := []Result{
		{ID: "NODE-02", Status: Fail, Target: "a"},
		{ID: "NODE-02", Status: Fail, Target: "b"},
		{ID: "NODE-02", Status: Pass, Target: "c"},
		{ID: "NODE-03", Status: Pass, Target: "a"},
		{ID: "NODE-03", Status: Info, Target: "b"},
		{ID: "CP-08", Status: Error, Source: SourceToolGap, Target: "cluster"},
		{ID: "VER-04", Status: Error, Source: SourceCollect, Target: "a"},
		{ID: "VER-04", Status: Error, Source: SourceToolGap, Target: "b"},
	}
	got := RuleStatus(rs)
	want := map[Status]int{Fail: 1, Pass: 0, Info: 1, Error: 2}
	for st, n := range want {
		if got[st] != n {
			t.Errorf("%s = %d，期望 %d（完整: %v）", st, got[st], n, got)
		}
	}
	// NODE-02 有 FAIL，不能被同规则的 PASS 冲掉；VER-04 的 collect 优先于 tool-gap。
	if got[Fail] != 1 {
		t.Errorf("同规则 FAIL 应保留，实际: %v", got)
	}
}
