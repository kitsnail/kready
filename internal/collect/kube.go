package collect

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/kitsnail/kready/internal/snapshot"
)

// Node 是 kubectl 返回的节点最小信息。
type Node struct {
	Name        string
	IP          string
	Labels      map[string]string
	Kubelet     string
	Ready       bool
	Taints      []snapshot.Taint
	Allocatable map[string]string
}

// kubeNodeList 是 `kubectl get nodes -o json` 的解析结构（只取所需字段）。
type kubeNodeList struct {
	Items []struct {
		Metadata struct {
			Name   string            `json:"name"`
			Labels map[string]string `json:"labels"`
		} `json:"metadata"`
		Spec struct {
			Taints []snapshot.Taint `json:"taints"`
		} `json:"spec"`
		Status struct {
			Addresses []struct {
				Type    string `json:"type"`
				Address string `json:"address"`
			} `json:"addresses"`
			NodeInfo struct {
				KubeletVersion string `json:"kubeletVersion"`
			} `json:"nodeInfo"`
			Allocatable map[string]string `json:"allocatable"`
			Conditions  []struct {
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
			Name:        it.Metadata.Name,
			Labels:      it.Metadata.Labels,
			Kubelet:     it.Status.NodeInfo.KubeletVersion,
			Taints:      it.Spec.Taints,
			Allocatable: it.Status.Allocatable,
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

// kubePodList 是 `kubectl get pods -A -o json` 的解析结构（只取所需字段）。
type kubePodList struct {
	Items []struct {
		Metadata struct {
			Namespace string            `json:"namespace"`
			Name      string            `json:"name"`
			Labels    map[string]string `json:"labels"`
		} `json:"metadata"`
		Spec struct {
			NodeName string `json:"nodeName"`
		} `json:"spec"`
		Status struct {
			Conditions []struct {
				Type   string `json:"type"`
				Status string `json:"status"`
			} `json:"conditions"`
			ContainerStatuses []struct {
				RestartCount int `json:"restartCount"`
			} `json:"containerStatuses"`
		} `json:"status"`
	} `json:"items"`
}

// ListPods 通过 kubectl 只读获取全部命名空间的 Pod 清单。只使用 get，不写入集群。
func ListPods() ([]snapshot.Pod, error) {
	out, err := Kubectl("get", "pods", "-A", "-o", "json")
	if err != nil {
		return nil, err
	}
	return ParsePods(out)
}

// ParsePods 将 `kubectl get pods -A -o json` 输出解析为 Pod 清单。
// Ready 取 Ready 条件；重启数为各容器重启数之和。
func ParsePods(data []byte) ([]snapshot.Pod, error) {
	var list kubePodList
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, fmt.Errorf("解析 Pod 清单失败: %w", err)
	}
	pods := make([]snapshot.Pod, 0, len(list.Items))
	for _, it := range list.Items {
		p := snapshot.Pod{
			Namespace: it.Metadata.Namespace,
			Name:      it.Metadata.Name,
			Node:      it.Spec.NodeName,
			Labels:    it.Metadata.Labels,
		}
		for _, c := range it.Status.Conditions {
			if c.Type == "Ready" && c.Status == "True" {
				p.Ready = true
			}
		}
		for _, cs := range it.Status.ContainerStatuses {
			p.Restarts += cs.RestartCount
		}
		pods = append(pods, p)
	}
	return pods, nil
}
