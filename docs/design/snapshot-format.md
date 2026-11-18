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
    { "name": "mn-1", "labels": { "node-role.kubernetes.io/control-plane": "" }, "ready": true, "kubelet": "v1.36.4" }
  ],
  "pods": [
    { "namespace": "kube-system", "name": "kube-vip-mn-1", "ready": true, "restarts": 0 }
  ],
  "kcm": { "memory_request": "512Mi", "memory_limit": "1Gi", "startup_probe": true }
}
```

- 节点角色由 `labels` 推导，规则见 `rule-schema.md`。
- 字段名与判断方式对应。新增字段须同时更新本文件。

## nodes/<node>.json

键为白名单命令 ID，值为该命令去除首尾空白后的输出。

```json
{
  "sysctl-inotify-instances": "1280",
  "sshd-password-auth": "PasswordAuthentication no"
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
