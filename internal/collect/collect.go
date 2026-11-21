package collect

import (
	"context"
	"sort"

	"github.com/kitsnail/kready/internal/snapshot"
)

// controlPlaneLabel 与快照判定使用的标签一致，用于标记控制面节点。
const controlPlaneLabel = "node-role.kubernetes.io/control-plane"

// ClusterFromNodes 将 kubectl 节点清单转为快照的 cluster 部分。
// 节点按名排序，保证输出稳定。
func ClusterFromNodes(nodes []Node) snapshot.Cluster {
	sorted := make([]Node, len(nodes))
	copy(sorted, nodes)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Name < sorted[j].Name })
	c := snapshot.Cluster{Config: map[string]string{}}
	for _, n := range sorted {
		c.Nodes = append(c.Nodes, snapshot.Node{
			Name:        n.Name,
			Labels:      n.Labels,
			Ready:       n.Ready,
			Kubelet:     n.Kubelet,
			Taints:      n.Taints,
			Allocatable: n.Allocatable,
		})
	}
	return c
}

// PodsFromList 将 kubectl Pod 清单转为快照的 pods 部分，按命名空间与名称排序。
func PodsFromList(pods []snapshot.Pod) []snapshot.Pod {
	sorted := make([]snapshot.Pod, len(pods))
	copy(sorted, pods)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Namespace != sorted[j].Namespace {
			return sorted[i].Namespace < sorted[j].Namespace
		}
		return sorted[i].Name < sorted[j].Name
	})
	return sorted
}

// NodeData 在每个节点上执行 commandIDs 中的白名单命令，返回 节点名 → 命令ID → 输出。
// 单条命令失败只缺少该键（判定时报 ERROR），不影响其它节点与命令。
func NodeData(ctx context.Context, runner Runner, nodes []Node, commandIDs []string) map[string]map[string]string {
	out := map[string]map[string]string{}
	for _, n := range nodes {
		data := map[string]string{}
		for _, id := range commandIDs {
			val, err := runner.Run(ctx, n.IP, id)
			if err != nil {
				continue
			}
			data[id] = val
		}
		out[n.Name] = data
	}
	return out
}
