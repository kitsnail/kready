# 快照格式 v1

> 状态：草案。用于单元与样本测试。测试不连接真实集群或节点。

## 目录

```
testdata/
├── positive/<pack>/              # 全部判定项应 PASS
│   ├── cluster.json
│   └── nodes/<node>.json
└── negative/<pack>/<RULE-ID>/    # 应报对应 FAIL
    ├── cluster.json
    └── nodes/<node>.json
```

## cluster.json

仅保存检查所需字段。

```json
{
  "nodes": [
    {
      "name": "mn-1",
      "labels": { "node-role.kubernetes.io/control-plane": "" },
      "ready": true,
      "kubelet": "v1.36.4",
      "taints": [ { "key": "node-role.kubernetes.io/control-plane", "value": "", "effect": "NoSchedule" } ],
      "allocatable": { "nvidia.com/gpu": "8" }
    }
  ],
  "pods": [
    { "namespace": "kube-system", "name": "kube-vip-mn-1", "node": "mn-1", "labels": { "app.kubernetes.io/name": "kube-vip" }, "ready": true, "restarts": 0 }
  ],
  "config": {
    "etcd_version": "3.6.14",
    "kube_proxy_mode": "nftables",
    "etcd_quota_backend_bytes": "8589934592",
    "metrics_server_replicas": "3"
  }
}
```

- `nodes`：所有节点；`kubelet` 为节点 kubelet 版本。节点角色由 `labels` 推导（`node-role.kubernetes.io/control-plane` 存在即为控制面）。
- `taints`（可选）：节点污点列表，每项含 `key`、`value`、`effect`。实时模式取自 `kubectl get nodes`。
- `allocatable`（可选）：节点可分配资源量（字符串值），实时模式取自 `kubectl get nodes`。
- `pods`：全部命名空间的 Pod。`node` 为所在节点名，`labels` 为 Pod 标签，均用于按节点与选择器筛选（如 GPU-01）。
- `config`：集群配置项的字符串值，规则通过 `field: config.<key>` 引用。
- 规则通过 `field` 声明取值方式，见 `rule-schema.md`。新增 `config` 键须同时更新本文件与规则。

## nodes/<node>.json

键为白名单命令 ID（与规则 `collect.command` 一致），值按命令的结果方式（见 `rule-schema.md` 白名单）：

- `output`（默认）：命令输出，去除首尾空白；
- `exit_code`：命令退出码的十进制字符串（如 `"0"`、`"9"`），非 0 不是采集错误。

```json
{
  "sysctl-inotify-instances": "1280",
  "sshd-password-auth": "PasswordAuthentication no",
  "gpu-nvidia-smi-list": "0"
}
```

缺少某个键，表示该命令采集失败，对应判定项为 ERROR。

## 脱敏

- 保留 IP 与节点名，检查依赖这些字段。
- 不含 Secret、token、私钥、证书内容。
- 快照仅用于本仓库测试，不得外发。

## 覆盖要求

- 每条 `active` 且 `mode: judge` 的规则，至少 1 个正例覆盖，并有 1 个负例 `negative/<pack>/<RULE-ID>/`。
- 负例只改变被测字段，其它字段与正例一致。
- `mode: info` 规则只需正例，验证输出为 INFO。
- 快照变更视为 B 类，需确认。
