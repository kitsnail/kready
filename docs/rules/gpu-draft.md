# GPU 规则草案（待人批准）

> 状态：草案。依据 D-012 与 REQ-001、REQ-002。未经批准不得写入 `rules/gpu/rules.yaml`。
> 期望值须人工逐条批准（AGENTS.md 门禁）。

## 一、规则条目草案

以下 ID 为建议值，批准后分配，不复用。

### GPU-01（REQ-001 标准 1）device plugin Pod 全部 Ready

- scope：node（取 `nvidia.com/gpu.present=true` 的节点）
- check：`all_ready`（复用现有判定）
- 待实测：device plugin Pod 的标签选择器。官方 Helm 部署带 `app.kubernetes.io/name=nvidia-device-plugin` 标签（来源：NVIDIA k8s-device-plugin README 的 Helm 部署说明），GPU Operator 的标签可能不同，实施前在真实集群核对。
- severity：error

### GPU-02（REQ-001 标准 2）allocatable 与 GFD 卡数一致

- scope：node
- check：需新增判定类型（比较两个字段），属于 `internal/check` 的新判定，**B 类**。
- 期望：`status.allocatable["nvidia.com/gpu"] == label["nvidia.com/gpu.count"]`
- severity：error

### GPU-03（REQ-001 标准 3）GFD 卡数标签存在

- scope：node
- check：`set`（复用现有判定）
- 期望：标签 `nvidia.com/gpu.count` 存在且非空
- severity：error

### GPU-04（REQ-001 标准 4）污点列出（INFO）

- scope：node
- mode：`info`
- 输出：GPU 节点的污点列表，不判定。
- 注意：`mode: info` 是现有规则 schema 支持的值，但当前代码只采集节点标签与 kubelet 版本，污点字段需要扩展快照（属于 `internal/collect` 与 `snapshot` 的改动，**B 类**）。

### GPU-05（REQ-002 标准 1）驱动加载

- scope：node
- collect：`command: gpu-nvidia-smi-list`，`as_root: false`
- check：需新增判定类型（退出码为 0），**B 类**。
- 期望：退出码 0 判 PASS；非 0 判 FAIL。
- 待实测：驱动未加载时的确切退出码（NVIDIA 文档为 9）。判定只看 0 与非 0，具体值只影响 `actual` 文字。

### GPU-06（REQ-002 标准 2）物理卡数一致（**暂缓**）

- 依赖 `lspci -d 10de::0302` 对数据中心卡的覆盖，未经实机验证，**本轮不实施**。

## 二、白名单命令草案（B 类）

| ID | 命令 | 说明 |
|---|---|---|
| `gpu-lspci-nvidia-vga` | `lspci -d 10de::0300` | 统计 NVIDIA VGA 类设备。与 `0302` 之和待 GPU-06 实施时补充 |
| `gpu-nvidia-smi-list` | `nvidia-smi -L` | 驱动加载检查，退出码为判定依据 |

两条均为只读，不含写操作。

## 三、输出契约草案（B 类）

- 问题：`--pack gpu` 无规则时输出"正常"，运维可能误读为 GPU 已检查通过。
- 建议：规则包无 active 规则时，结论输出 `未运行任何规则`，退出码仍为 0（不是异常，也不是工具错误）。
- 影响：`docs/design/output-contract.md`（新增结论取值）；`internal/check` 的结论计算；`schema_version` 按 `versioning.md` 判断是否递增（新增取值属于兼容性变更，需人确认）。

## 四、待人决定的问题

1. 是否批准 GPU-01 至 GPU-05 的期望值与判定类型（GPU-02、GPU-05 需新判定，GPU-04 需扩展快照）。
2. 是否批准白名单两条命令。
3. 是否批准输出契约中新增"未运行任何规则"取值。
4. GPU-06 暂缓，是否同意。
