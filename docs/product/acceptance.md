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

## REQ-001 GPU 节点自洽检查：device plugin、可分配 GPU 数、节点标签与污点

- [ ] 标准 1：对带 `nvidia.com/gpu.present=true` 标签的节点，属于 device plugin 的 Pod 全部 Ready 时判正常；否则判异常，列出 Ready 数与期望数。
- [ ] 标准 2：节点 `status.allocatable["nvidia.com/gpu"]` 与 GFD 标签 `nvidia.com/gpu.count` 相等时判正常；不等时判异常，列出两个实际值。
- [ ] 标准 3：带 GPU 资源的节点缺少 GFD 标签 `nvidia.com/gpu.count` 时判异常；存在时判正常。
- [ ] 标准 4：列出 GPU 节点上的全部污点，输出为 INFO，不判定。
- [ ] 标准 5：结果按 `nvidia.com/gpu.product` 机型分组输出。
- [ ] 标准 6：所有取数只使用 `get` / `list`，不产生写操作。

验收人：待定　日期：待定

## REQ-002 GPU 节点存在性检查：驱动加载与 GPU 数量

- [ ] 标准 1：对 GPU 节点执行 `nvidia-smi -L`，退出码为 0 时判正常；非 0 时判异常，列出退出码。
- [ ] 标准 2：节点 `lspci` 统计的 NVIDIA 设备数（类代码 `0300` 与 `0302` 之和）与 GFD 标签 `nvidia.com/gpu.count` 不一致，或与 allocatable `nvidia.com/gpu` 不一致时判异常，列出各实际值。
- [ ] 标准 3：结果按 `nvidia.com/gpu.product` 机型分组输出。
- [ ] 标准 4：所有取数只使用只读命令（`get` / `list` 与白名单只读命令），不产生写操作。

验收人：待定　日期：待定

## REQ-003 GPU 版本检查：驱动版本与 container toolkit 版本（本轮不做）

- [ ] 标准 1：（待补充）驱动版本的期望值与来源，需人确认标准文档后填写。
- [ ] 标准 2：（待补充）container toolkit 版本的期望值与来源，需人确认标准文档后填写。

验收人：待定　日期：待定

## 验收记录

（暂无）
