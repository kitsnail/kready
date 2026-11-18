package collect

import (
	"context"

	"github.com/scclabs/kready/internal/snapshot"
)

// Live 通过 kubectl 与 SSH 采集，组装成与离线快照相同结构的 Snapshot。
// commandIDs 由所选规则包提供（见 rules.Pack.CollectIDs），采集层不自行决定命令。
// 集群级 config 当前无来源，留空；依赖 config 的规则会报 ERROR（缺少字段），不会凭空填值。
func Live(ctx context.Context, runner Runner, nodes []Node, commandIDs []string) *snapshot.Snapshot {
	return &snapshot.Snapshot{
		Cluster: ClusterFromNodes(nodes),
		Nodes:   NodeData(ctx, runner, nodes, commandIDs),
	}
}
