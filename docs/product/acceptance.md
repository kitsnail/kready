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
