package main

import (
	"context"
	"fmt"

	"github.com/kitsnail/kready/internal/collect"
	"github.com/kitsnail/kready/internal/rules"
	"github.com/kitsnail/kready/internal/snapshot"
)

// liveSnapshot 只通过 kubectl 取节点清单，用于选包。此时不发起 SSH 连接。
// 返回的 nodes 用于后续采集，保证选包与采集基于同一份清单。
func liveSnapshot() (*snapshot.Snapshot, []collect.Node, error) {
	nodes, err := collect.ListNodes()
	if err != nil {
		return nil, nil, err
	}
	return &snapshot.Snapshot{
		Cluster: collect.ClusterFromNodes(nodes),
		Nodes:   map[string]map[string]string{},
	}, nodes, nil
}

// collectNodes 按所选规则包的命令 ID，通过 SSH 只读采集节点，结果写入快照的 Nodes。
// 只在选包成功后调用，无适用规则时不发起任何 SSH 连接。
func collectNodes(snap *snapshot.Snapshot, nodes []collect.Node, pack *rules.Pack) (*snapshot.Snapshot, error) {
	runner, err := collect.NewSSHRunner()
	if err != nil {
		return nil, fmt.Errorf("SSH 初始化失败: %w", err)
	}
	defer runner.Close()
	snap.Nodes = collect.NodeData(context.Background(), runner, nodes, pack.CollectIDs())
	return snap, nil
}
