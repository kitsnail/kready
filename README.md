<div align="center">
  <img src="docs/assets/logo.svg" alt="kready" width="96">
</div>

# kready

面向一线运维的 Kubernetes 集群交付就绪检查工具。一条命令，输出**正常**或**异常**；异常时列出实际值与期望值。

- **只读**：不做任何写操作，不修改集群与节点。
- **简单**：无需配置文件，直接运行。
- **确定**：只做能自动判定的检查，期望值以规则包为准。

## 适用范围

| 项目 | 当前支持 |
|---|---|
| 部署工具 | kubespray 2.32.x |
| Kubernetes | v1.36.4（kubelet 版本精确匹配） |

基线不匹配时输出“无适用规则”，退出码为 `2`，不做判定。

## 快速开始

需要 Go 1.27.4。

```bash
make build                                    # 生成 bin/kready
./bin/kready check --snapshot testdata/positive/2.32
```

示例输出：

```text
正常
节点数: 5

【参考信息】 内核自动升级服务当前状态（NODE-05）
    cp-1、cp-2、cp-3、w-1、w-2  当前: inactive

【参考信息】 metrics-server 副本数（SYS-04）
    cluster  当前: 3
```

异常示例（控制面节点数不足）：

```text
异常
节点数: 5

【必须处理】 控制面节点数（CP-01）
    cluster
        实际: 2  期望: 3

【参考信息】 内核自动升级服务当前状态（NODE-05）
    cp-1、cp-2、cp-3、w-1、w-2  当前: inactive

【参考信息】 metrics-server 副本数（SYS-04）
    cluster  当前: 3
```

## 用法

```bash
kready check --snapshot <目录>        # 离线：读取快照，判定
kready check --live                   # 在线：kubectl 取节点清单，SSH 只读采集节点
kready check --snapshot <目录> --format json   # JSON 输出，便于脚本处理
kready version
```

- `--snapshot` 与 `--live` 二选一。
- `--live` 依赖 `kubectl`（可访问集群）与 ssh-agent（root 账号，校验 known_hosts）。

## 退出码

| 退出码 | 含义 |
|---|---|
| `0` | 正常 |
| `1` | 异常，或未能确认（采集失败） |
| `2` | 工具自身出错，如快照无法读取、无适用规则 |

可直接用于流水线门禁：返回 `0` 即通过。

## 输出与留存

- 报告写入 `$KREADY_HOME/reports/`（默认 `~/.kready`），文本或 JSON 与终端一致。
- `--live` 的采集快照写入 `$KREADY_HOME/snapshots/<时间戳>/`，可用 `--snapshot` 复现判定。
- 目录不可用时，判定前即报错并返回 `2`。

## 规则

- 规则包位于 `rules/2.32/rules.yaml`，每条规则有稳定 ID（如 `CP-01`）。
- 规则状态为 `active` 才参与判定；`mode: info` 仅输出当前值，不判定。
- 规则 schema、输出契约与快照格式见 `docs/design/`。

## 开发

```bash
make fmt      # 格式化
make lint     # 静态检查
make test     # 单元与样本测试
make check    # 以上全部，提交前必须通过
```

- 测试只使用快照与 fake client，不连接真实集群或外网。
- `testdata/positive/2.32` 为正例（全部通过）；`testdata/negative/2.32/<规则ID>/` 为负例（每条判定规则一个）。
- 提交信息遵循 Conventional Commits，规则改动使用 scope `rules`。
- 变更记录见 [CHANGELOG.md](CHANGELOG.md)。

## 文档

- [设计总览](docs/design/kready-design.md)
- [输出契约](docs/design/output-contract.md)
- [规则 schema](docs/design/rule-schema.md)
- [快照格式](docs/design/snapshot-format.md)
- [版本规范](docs/design/versioning.md)
- [ADR 0001：节点通过 SSH 只读采集](docs/decisions/0001-node-collection-via-ssh.md)
