# 规则 Schema v1

> 状态：草案。规则随工具内嵌发布，一条规则一个 YAML 条目。运行时不读取任何部署源文件。

## 字段

| 字段 | 必填 | 说明 |
|---|---|---|
| `id` | 是 | 稳定 ID，如 `CP-06`，不复用 |
| `title` | 是 | 一句话描述检查内容 |
| `status` | 是 | `draft` / `active` / `deprecated` |
| `mode` | 是 | `judge`：判定 PASS/FAIL；`info`：仅输出当前值 |
| `scope` | 是 | `cluster` / `node` |
| `source` | 是 | 依据，仅用于追溯，如 `new-cluster-config-guide.md:§4`。运行时不读取 |
| `field` | 取值方式声明，见下文 | 集群级：`config.<key>` 读取快照 `config` 中的字段；`count_role` 按 `expect.role` 计数节点。节点级：`label.<key>` 节点标签值、`allocatable.<key>` 节点可分配资源量、`taints` 节点污点列表、`pods` 该节点的 Pod（配合 `check: all_ready`）。`equal_baseline` 不用 `field` |
| `node_label` | 否，仅节点级 | 形如 `key=value`，规则只作用于带该标签的节点。无匹配节点时输出一条 SKIP，不判定 |
| `collect` | 否 | 经 SSH 采集的取数方式，见下文。集群数据（`field`）与 `collect` 二选一 |
| `check` | `judge` 时必填 | 判断方式，见下文 |
| `expect` | `judge` 时必填 | 期望值，格式由 `check` 决定 |
| `severity` | `judge` 时必填 | `error`：不符合时输出 FAIL，结论为异常；`warn`：输出 WARN，结论不变。取值规则随规则逐步定义 |
| `impact` | 占位 | 不处理会带来什么后果，大白话。具体写法和必填规则待实际问题中细化 |
| `hint` | 占位 | 处理方式，可直接复制的命令或操作方向。具体写法和必填规则待实际问题中细化 |

## 规则包头部

| 字段 | 必填 | 说明 |
|---|---|---|
| `pack` | 是 | 规则包标识，如 `2.32` |
| `rules_rev` | 是 | 规则版本，整数 |
| `applies_to.kubelet` | 版本选包的包必填 | 适用的 kubelet 版本列表，精确匹配。选包依据，见 `output-contract.md`。仅 `--pack` 显式指定的包（如 `gpu`）留空，不做版本选包 |

## 状态

- `draft`：只做 schema 校验，不参与输出。
- `active`：参与输出。
- `deprecated`：不参与输出，保留文件与 ID。

## 判断方式（check）

| check | expect | 含义 |
|---|---|---|
| `equals` | `value` | 实际值等于 `value` |
| `count` | `value`，可选 `role`、`selectors`、`pod_prefix` | 对象数量等于 `value` |
| `count_eq_nodes` | `role`，可选 `selectors`、`pod_prefix` | 对象数量等于该角色节点数 |
| `control_plane_pods_ready` | `role`，可选 `components` | 控制面核心组件（kube-apiserver/kube-controller-manager/kube-scheduler）的 Pod 数各等于控制面节点数且全部 Ready；任一数量不符或未 Ready，判 FAIL |
| `all_ready` | `selectors` | 目标节点上匹配任一选择器的 Pod 全部 Ready；一个都没有视为 FAIL（未部署）。实际值为 `Ready 数/总数`。`selectors` 为逗号分隔的 `key=value` 列表（k8s 标签键值不含逗号，D-015） |
| `no_restarts` | `window` | 窗口内重启次数为 0，单位 `m` |
| `set` | — | 字段存在且非空。缺失即 FAIL，实际值显示为“（缺失）” |
| `equals_label` | `label` | 实际值等于同节点的另一标签值，期望值逐节点取自快照 |
| `equal_baseline` | `baseline: control-plane` | 非控制面节点的 kubelet 版本都等于控制面基线。基线缺失或控制面版本不一致、或工作节点版本为空时报 ERROR；版本不符的工作节点逐个列出，判 FAIL |

只增不改语义。新增判断方式需要 B 类确认。

## 取数据方式（collect，仅节点）

```yaml
collect:
  command: sysctl-inotify-instances   # 白名单命令 ID，见 ADR 0001
  as_root: false
```

- 命令 ID 必须在代码白名单中注册。规则不得写 shell 命令。
- 白名单命令的结果方式：`output`（默认，输出去除首尾空白后作为实际值）或 `exit_code`（退出码的十进制字符串作为实际值，非 0 不是采集错误）。

## 节点角色

- `control-plane`：带标签 `node-role.kubernetes.io/control-plane` 的节点。
- `worker`：其余节点。

## 示例

```yaml
id: NODE-02
title: inotify max_user_instances
status: active
mode: judge
scope: node
source: os/80-inotify.conf
collect:
  command: sysctl-inotify-instances
  as_root: false
check: equals
expect:
  value: "1280"
```

## 信息项示例

```yaml
id: SYS-04
title: metrics-server 副本数
status: active
mode: info
scope: cluster
source: 部署配置（仅追溯）
```

## 约束

- 无法自动判定的检查，必须用 `mode: info`，不得为凑数而判定。
- `source` 必填。无依据的规则不得进入 `active`。
- 规则变更视为 B 类，需确认。
