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
	if !Blocking([]Result{{Status: Fail}}) {
		t.Error("确认不合规应阻断")
	}
	if !Blocking([]Result{{Status: Error, Source: SourceCollect}}) {
		t.Error("采集失败应阻断（未能确认）")
	}
}
