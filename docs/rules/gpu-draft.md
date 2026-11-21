# GPU 规则草案

> 状态：GPU-01 至 GPU-04 已批准实施（D-014，2026-11-21），见 `rules/gpu/rules.yaml`；GPU-05 随 REQ-002 实施；GPU-06 暂缓。
> 期望值已人工逐条批准（AGENTS.md 门禁）。

## 一、规则条目

### GPU-01（REQ-001 标准 1）device plugin Pod 全部 Ready — 已实施

- scope：node（取 `nvidia.com/gpu.present=true` 的节点）
- check：`all_ready`，选择器 `app.kubernetes.io/name=nvidia-device-plugin`
- 待实测：选择器取 NVIDIA k8s-device-plugin Helm 默认标签；GPU Operator 部署的标签可能不同，实施前须在真实集群核对（见 `rules/gpu/rules.yaml` GPU-01 source）。

### GPU-02（REQ-001 标准 2）allocatable 与 GFD 卡数一致 — 已实施

- scope：node
- check：`equals_label`（新判定类型，D-014 批准）
- 期望：`status.allocatable["nvidia.com/gpu"] == label["nvidia.com/gpu.count"]`

### GPU-03（REQ-001 标准 3）GFD 卡数标签存在 — 已实施

- scope：node
- check：`set`
- 期望：标签 `nvidia.com/gpu.count` 存在且非空

### GPU-04（REQ-001 标准 4）污点列出（INFO）— 已实施

- scope：node
- mode：`info`
- 输出：GPU 节点的污点列表（`key=value:effect`，逗号分隔；无污点为“无”），不判定。
- 快照已扩展节点 `taints`、`allocatable` 字段与 Pod `node`、`labels` 字段（见 `docs/design/snapshot-format.md`）。

### GPU-05（REQ-002 标准 1）驱动加载 — 随 REQ-002 实施

- scope：node
- collect：`command: gpu-nvidia-smi-list`，`as_root: false`
- check：`equals`，期望值为命令退出码 `0`（白名单新增 exit_code 取结果方式）。
- 待实测：驱动未加载时的确切退出码（NVIDIA 文档为 9）。判定只看 0 与非 0，具体值只影响 `actual` 文字。

### GPU-06（REQ-002 标准 2）物理卡数一致（**暂缓**）

- 依赖 `lspci -d 10de::0302` 对数据中心卡的覆盖，未经实机验证，本轮不实施。

## 二、白名单命令

| ID | 命令 | 说明 |
|---|---|---|
| `gpu-lspci-nvidia-vga` | `lspci -d 10de::0300` | **暂缓**：随 GPU-06 一并实施时再申请批准，本轮不申请 |
| `gpu-nvidia-smi-list` | `nvidia-smi -L` | 驱动加载检查，退出码为判定依据（D-014 批准，随 REQ-002 实施） |

两条均为只读，不含写操作。

## 三、输出契约（D-014 批准）

- 规则包无 active 规则时，结论输出 `未运行任何规则`，退出码仍为 0（不是异常，也不是工具错误）。
- 规则声明 `node_label` 但无匹配节点时输出 SKIP，不影响结论。
- 已写入 `docs/design/output-contract.md`，`schema_version` 递增至 4。

## 四、已决事项（D-014，2026-11-21）

1. GPU-01 至 GPU-05 的期望值与判定类型已批准（GPU-02 `equals_label`、GPU-05 退出码判定为新判定；GPU-04 快照扩展已批准）。
2. 白名单命令 `gpu-nvidia-smi-list`（`nvidia-smi -L`）已批准；`gpu-lspci-nvidia-vga` 随 GPU-06 暂缓。
3. 输出契约新增“未运行任何规则”取值已批准。
4. GPU-06 暂缓已同意。
