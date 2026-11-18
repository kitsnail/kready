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
| `field` | 集群级 `judge` 必填 | 集群级取值方式：`config.<key>` 读取快照 `config` 中的字段；`kubelet_versions` 全部节点 kubelet 版本（去重，逗号连接）；`count_role` 按 `expect.role` 计数节点。节点级不用 `field`，用 `collect` |
| `collect` | 否 | 取数据方式，见下文。缺省表示集群级取值（由 `field` 声明） |
| `check` | `judge` 时必填 | 判断方式，见下文 |
| `expect` | `judge` 时必填 | 期望值，格式由 `check` 决定 |
| `severity` | `judge` 时必填 | `error`：不符合时输出 FAIL，结论为异常；`warn`：输出 WARN，结论不变。取值规则随规则逐步定义 |
| `impact` | 占位 | 不处理会带来什么后果，大白话。具体写法和必填规则待实际问题中细化 |
| `hint` | 占位 | 处理方式，可直接复制的命令或操作方向。具体写法和必填规则待实际问题中细化 |

## 状态

- `draft`：只做 schema 校验，不参与输出。
- `active`：参与输出。
- `deprecated`：不参与输出，保留文件与 ID。

## 判断方式（check）

| check | expect | 含义 |
|---|---|---|
| `equals` | `value` | 实际值等于 `value` |
| `count` | `value`，可选 `role` | 对象数量等于 `value` |
| `count_eq_nodes` | `role` | 对象数量等于该角色节点数 |
| `all_ready` | — | 全部对象 Ready |
| `no_restarts` | `window` | 窗口内重启次数为 0，单位 `m` |
| `set` | — | 字段存在且非空 |

只增不改语义。新增判断方式需要 B 类确认。

## 取数据方式（collect，仅节点）

```yaml
collect:
  command: sysctl-inotify-instances   # 白名单命令 ID，见 ADR 0001
  as_root: false
```

- 命令 ID 必须在代码白名单中注册。规则不得写 shell 命令。
- 命令输出去除首尾空白后，作为实际值。

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
