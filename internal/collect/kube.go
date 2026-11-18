package collect

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// Node 是 kubectl 返回的节点最小信息。
type Node struct {
	Name    string
	IP      string
	Labels  map[string]string
	Kubelet string
	Ready   bool
}

// kubeNodeList 是 `kubectl get nodes -o json` 的解析结构（只取所需字段）。
type kubeNodeList struct {
	Items []struct {
		Metadata struct {
			Name   string            `json:"name"`
			Labels map[string]string `json:"labels"`
		} `json:"metadata"`
		Status struct {
			Addresses []struct {
				Type    string `json:"type"`
				Address string `json:"address"`
			} `json:"addresses"`
			NodeInfo struct {
				KubeletVersion string `json:"kubeletVersion"`
			} `json:"nodeInfo"`
			Conditions []struct {
				Type   string `json:"type"`
				Status string `json:"status"`
			} `json:"conditions"`
		} `json:"status"`
	} `json:"items"`
}

// Kubectl 执行 kubectl 只读命令的函数，便于测试替换。生产环境使用 execKubectl。
var Kubectl = execKubectl

func execKubectl(args ...string) ([]byte, error) {
	cmd := exec.Command("kubectl", args...)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("kubectl %s 失败: %w", strings.Join(args, " "), err)
	}
	return out, nil
}

// ListNodes 通过 kubectl 只读获取节点清单。只使用 get，不写入集群。
func ListNodes() ([]Node, error) {
	out, err := Kubectl("get", "nodes", "-o", "json")
	if err != nil {
		return nil, err
	}
	return ParseNodes(out)
}

// ParseNodes 将 `kubectl get nodes -o json` 输出解析为节点清单。
// 节点 IP 取 InternalIP；缺失 InternalIP 的节点会返回错误，因为无法发起采集。
func ParseNodes(data []byte) ([]Node, error) {
	var list kubeNodeList
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, fmt.Errorf("解析节点清单失败: %w", err)
	}
	var nodes []Node
	for _, it := range list.Items {
		n := Node{
			Name:    it.Metadata.Name,
			Labels:  it.Metadata.Labels,
			Kubelet: it.Status.NodeInfo.KubeletVersion,
		}
		for _, a := range it.Status.Addresses {
			if a.Type == "InternalIP" {
				n.IP = a.Address
			}
		}
		for _, c := range it.Status.Conditions {
			if c.Type == "Ready" && c.Status == "True" {
				n.Ready = true
			}
		}
		if n.IP == "" {
			return nil, fmt.Errorf("节点 %s 缺少 InternalIP，无法采集", n.Name)
		}
		nodes = append(nodes, n)
	}
	return nodes, nil
}
