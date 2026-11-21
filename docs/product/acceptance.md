# 验收标准（acceptance）

> 维护者：人（产品经理）。Agent 仅按 `docs/product/README.md`「字段级权限」写入（D-011）。
> 作用：每个 REQ 的验收条目。Agent 按此实现与验证；人按此确认是否验收。

## 编写要求

- 每条标准可以自动判定，或可由人在真实环境中复核。
- 一条标准只描述一个结果，不写实现方式。
- 涉及规则判定的，写明规则 ID 与期望结论（正常 / 异常 / 未能确认）。

## 模板

```markdown
## REQ-000 <标题>

- [ ] 标准 1：<可判定的结果>
- [ ] 标准 2：<可判定的结果>

验收人：<人>　日期：<YYYY-MM-DD>
```

## 验收记录

## REQ-001 GPU 节点基础检查

- [x] 标准 1：GPU-01：标签 `nvidia.com/gpu.present=true` 的节点上，device plugin Pod 全部 Ready，判「正常」；任一未 Ready，判「异常」（error）。
- [x] 标准 2：GPU-02：`status.allocatable["nvidia.com/gpu"]` 等于标签 `nvidia.com/gpu.count`，判「正常」；不等，判「异常」（error）。
- [x] 标准 3：GPU-03：GPU 节点标签 `nvidia.com/gpu.count` 存在且非空，判「正常」；缺失或为空，判「异常」（error）。
- [x] 标准 4：GPU-04：GPU 节点的污点列表以信息项（INFO）输出，不影响结论。
- [x] 标准 5：`--pack gpu` 无 active 规则时，结论为「未运行任何规则」，退出码 0。
- [x] 标准 6：GPU-01～GPU-03 各有负例快照，报出对应规则的 FAIL（`testdata/negative/gpu/`）。

验收人：kitsnail　日期：2026-11-21

## REQ-002 GPU 驱动加载检查

- [x] 标准 1：GPU-05：GPU 节点上 `nvidia-smi -L` 退出码 0，判「正常」；非 0，判「异常」（error）。
- [x] 标准 2：白名单新增 `gpu-nvidia-smi-list`（`nvidia-smi -L`，只读），不引入其他命令。
- [x] 标准 3：GPU-05 有负例快照，报出 FAIL。

验收人：kitsnail　日期：2026-11-21

## REQ-003 GPU-01 适配 GPU Operator 部署

- [x] 标准 1：GPU Operator 标签（`app=nvidia-device-plugin-daemonset`）与 Helm 标签（`app.kubernetes.io/name=nvidia-device-plugin`）的 Pod 均被 GPU-01 纳入判定。
- [x] 标准 2：两种标签都无匹配时判「异常」（FAIL，未部署）。
- [x] 标准 3：匹配到的 Pod 任一未 Ready 判「异常」（FAIL）（既有语义不变）。

验收人：kitsnail　日期：2026-11-21

## REQ-004 控制面规模与组件检查（CP-02～CP-05）

- [ ] 标准 1：CP-02：etcd 成员数为 3，判「正常」；不为 3，判「异常」（FAIL，列出实际成员数）。
- [ ] 标准 2：CP-03：kube-apiserver、kube-controller-manager、kube-scheduler 的 Pod 数各等于控制面节点数且全部 Ready；任一数量不符或未 Ready，判「异常」。
- [ ] 标准 3：CP-04：kube-vip Pod 数（含镜像 Pod）等于控制面节点数；不符判「异常」。
- [ ] 标准 4：CP-05：nginx-proxy Pod 数（含镜像 Pod）等于工作节点数；不符判「异常」。
- [ ] 标准 5：CP-02～05 各有负例快照，报出对应规则的 FAIL（`testdata/negative/`）。

## REQ-005 控制面资源与 etcd 配额检查（CP-06～CP-08）

- [ ] 标准 1：CP-06：kube-controller-manager 内存 request 与 limit 均已设置；任一缺失判「异常」。
- [ ] 标准 2：CP-07：kube-controller-manager startupProbe 已配置；缺失判「异常」。
- [ ] 标准 3：CP-08：etcd 后端配额为 `8589934592`（8 GiB）；不符判「异常」（列出实际值）。
- [ ] 标准 4：CP-06～08 各有负例快照，报出对应规则的 FAIL。

## REQ-006 系统组件健康检查（SYS-01～SYS-03）

- [ ] 标准 1：SYS-01：kube-system 全部 Pod Ready 且无 CrashLoopBackOff；任一不满足判「异常」（列出问题 Pod）。
- [ ] 标准 2：SYS-02：kube-system Pod 近 15 分钟容器重启 0 次；任一重启次数非 0 判「异常」（列出 Pod 与次数）。
- [ ] 标准 3：SYS-03：Cilium、kube-proxy、nodelocaldns 三个 DaemonSet 在每个节点各有一个 Ready Pod；缺失或未 Ready 判「异常」。
- [ ] 标准 4：SYS-01～03 各有负例快照，报出对应规则的 FAIL。

## REQ-007 NTP 结果判定（POL-03a/POL-03b）

- [ ] 标准 1：POL-03a：每节点 NTP 同步状态 = `yes`；任一节点非 `yes` 判「异常」（列出节点与实际状态）。
- [ ] 标准 2：POL-03b：节点间时钟偏差（epoch 毫秒 max−min）≤ `1000`；超限判「异常」（列出偏差值）。
- [ ] 标准 3：NTP 服务器地址与配置不在检查范围（无对应规则字段）。
- [ ] 标准 4：POL-03a/b 各有负例快照，报出对应规则的 FAIL。
- [ ] 标准 5：白名单仅新增 `timedatectl-ntp-sync`（`timedatectl show --property=NTPSynchronization --value`）与 `date-epoch-ms`（`date +%s%3N`）两条只读命令。

## REQ-008 OS 内核参数检查（OS-01～OS-03）

- [ ] 标准 1：OS-01：节点 `br_netfilter` 模块已加载；未加载判「异常」。
- [ ] 标准 2：OS-02：节点 `net.ipv4.ip_forward = 1`；不符判「异常」（列出实际值）。
- [ ] 标准 3：OS-03：节点 swap 已关闭；未关闭判「异常」。
- [ ] 标准 4：OS-01～03 各有负例快照，报出对应规则的 FAIL。
- [ ] 标准 5：白名单新增命令均为只读（具体命令清单在 B 类方案中列出并经人批准）。
