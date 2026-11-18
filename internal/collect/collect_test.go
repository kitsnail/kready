package collect

import (
	"context"
	"errors"
	"sort"
	"strings"
	"testing"

	"github.com/scclabs/kready/internal/rules"
)

// fakeRunner 只返回预设输出，不发起任何 SSH 连接。
type fakeRunner struct {
	out map[string]string // key: ip + "|" + commandID
}

func (f fakeRunner) Run(_ context.Context, ip, id string) (string, error) {
	v, ok := f.out[ip+"|"+id]
	if !ok {
		return "", errors.New("fake: 无预设输出")
	}
	return v, nil
}

// 白名单拒绝未知 ID（ADR 0001 要求的单元测试）。
func TestArgvRejectsUnknownID(t *testing.T) {
	for _, id := range []string{"rm-rf", "", "sysctl-inotify-instances; rm -rf /", "../etc/passwd"} {
		if _, err := Argv(id); err == nil {
			t.Errorf("未知命令 %q 应被拒绝", id)
		}
	}
}

// 白名单返回副本，调用方修改不影响白名单。
func TestArgvReturnsCopy(t *testing.T) {
	a, err := Argv("sysctl-inotify-instances")
	if err != nil {
		t.Fatal(err)
	}
	a[0] = "rm"
	b, _ := Argv("sysctl-inotify-instances")
	if b[0] != "sysctl" {
		t.Errorf("白名单被外部修改: %v", b)
	}
}

// 规则包声明的采集命令必须全部在白名单中，否则运行时必然报错。
func TestRulePackCommandsAreWhitelisted(t *testing.T) {
	p, err := rules.Load("2.32")
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range p.CollectIDs() {
		if _, err := Argv(id); err != nil {
			t.Errorf("规则命令 %s 不在白名单中: %v", id, err)
		}
	}
	if len(p.CollectIDs()) == 0 {
		t.Error("规则包应声明至少一条采集命令")
	}
}

// 白名单参数逐个单引号包裹，单引号被转义，不会被 shell 解释为额外命令。
func TestShellJoinQuotesArguments(t *testing.T) {
	got := shellJoin([]string{"sysctl", "-n", "a'b; rm -rf /"})
	want := `'sysctl' '-n' 'a'\''b; rm -rf /'`
	if got != want {
		t.Errorf("shellJoin = %s，期望 %s", got, want)
	}
}

const nodeListJSON = `{"items":[
 {"metadata":{"name":"cp-1","labels":{"node-role.kubernetes.io/control-plane":""}},
  "status":{"addresses":[{"type":"InternalIP","address":"10.0.0.1"},{"type":"Hostname","address":"cp-1"}],
   "nodeInfo":{"kubeletVersion":"v1.36.4"},"conditions":[{"type":"Ready","status":"True"}]}},
 {"metadata":{"name":"w-1","labels":{}},
  "status":{"addresses":[{"type":"InternalIP","address":"10.0.0.2"}],
   "nodeInfo":{"kubeletVersion":"v1.36.4"},"conditions":[{"type":"Ready","status":"False"}]}}
]}`

// 解析节点清单：取 InternalIP、版本、Ready、控制面标签。
func TestParseNodes(t *testing.T) {
	nodes, err := ParseNodes([]byte(nodeListJSON))
	if err != nil {
		t.Fatal(err)
	}
	if len(nodes) != 2 {
		t.Fatalf("节点数 = %d，期望 2", len(nodes))
	}
	cp := nodes[0]
	if cp.IP != "10.0.0.1" || cp.Kubelet != "v1.36.4" || !cp.Ready {
		t.Errorf("cp-1 解析不符: %+v", cp)
	}
	if _, ok := cp.Labels["node-role.kubernetes.io/control-plane"]; !ok {
		t.Errorf("cp-1 应带控制面标签")
	}
	if nodes[1].Ready {
		t.Errorf("w-1 Ready=False，解析应为未就绪")
	}
}

// 缺少 InternalIP 的节点无法采集，应报错而不是静默跳过。
func TestParseNodesRequiresInternalIP(t *testing.T) {
	data := `{"items":[{"metadata":{"name":"x"},"status":{"addresses":[{"type":"Hostname","address":"x"}]}}]}`
	if _, err := ParseNodes([]byte(data)); err == nil {
		t.Error("缺少 InternalIP 应报错")
	}
}

// ListNodes 只调用 get，且参数固定。
func TestListNodesOnlyGets(t *testing.T) {
	var called []string
	orig := Kubectl
	Kubectl = func(args ...string) ([]byte, error) {
		called = append(called, strings.Join(args, " "))
		return []byte(nodeListJSON), nil
	}
	defer func() { Kubectl = orig }()
	if _, err := ListNodes(); err != nil {
		t.Fatal(err)
	}
	if len(called) != 1 || !strings.HasPrefix(called[0], "get nodes") {
		t.Errorf("应只调用 get nodes，实际: %v", called)
	}
	for _, c := range called {
		for _, bad := range []string{"apply", "delete", "patch", "create", "edit", "replace"} {
			if strings.Contains(c, bad) {
				t.Errorf("出现写操作 %q: %s", bad, c)
			}
		}
	}
}

// 节点数据组装：每个节点取回白名单命令输出；单条失败只缺该键。
func TestNodeDataPartialFailure(t *testing.T) {
	nodes := []Node{{Name: "cp-1", IP: "10.0.0.1"}, {Name: "w-1", IP: "10.0.0.2"}}
	r := fakeRunner{out: map[string]string{
		"10.0.0.1|sysctl-inotify-instances": "1280",
		"10.0.0.1|sysctl-inotify-watches":   "655360",
		"10.0.0.2|sysctl-inotify-instances": "1280",
	}}
	data := NodeData(context.Background(), r, nodes, []string{"sysctl-inotify-instances", "sysctl-inotify-watches"})
	if data["cp-1"]["sysctl-inotify-watches"] != "655360" {
		t.Errorf("cp-1 watches 应为 655360: %v", data["cp-1"])
	}
	if _, ok := data["w-1"]["sysctl-inotify-watches"]; ok {
		t.Errorf("w-1 采集失败的命令应缺键，不能填值: %v", data["w-1"])
	}
	if data["w-1"]["sysctl-inotify-instances"] != "1280" {
		t.Errorf("w-1 成功的命令应保留: %v", data["w-1"])
	}
}

// 组装的 cluster 按节点名排序，结果稳定。
func TestClusterFromNodesSorted(t *testing.T) {
	nodes := []Node{{Name: "w-2"}, {Name: "cp-1"}, {Name: "w-1"}}
	c := ClusterFromNodes(nodes)
	names := []string{c.Nodes[0].Name, c.Nodes[1].Name, c.Nodes[2].Name}
	if !sort.StringsAreSorted(names) {
		t.Errorf("节点未排序: %v", names)
	}
}
