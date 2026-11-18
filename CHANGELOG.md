# CHANGELOG

## [Unreleased]

### Added
- 检查切片（2.32）：规则包 `rules/2.32/rules.yaml`（8 条判定；NODE-05、SYS-04 为信息项）；`kready check --snapshot <目录>` 离线读取快照，输出正常/异常，异常时列出实际值与期望值。
- 快照测试：`testdata/positive/2.32`（全部通过）与 `testdata/negative/2.32/<RULE-ID>/`（每条判定规则一个负例）。测试只使用快照。
- NODE-05（内核自动升级服务状态）：判定方案未定，暂为信息项，仅输出当前值（见 `docs/rules/2.32-draft.md` 待确认 5）。
- 规则包选择：按快照中节点的 kubelet 版本选包（规则包 `applies_to.kubelet` 精确匹配）。无匹配或节点版本不一致时输出"无适用规则"，退出码 2，不做判定。
- JSON 输出：`kready check --snapshot <目录> --format json`，结构见 `docs/design/output-contract.md`（`schema_version` 1）。
- 依赖：`gopkg.in/yaml.v3`。理由：规则包使用 YAML 格式，标准库无 YAML 解析器。
- 项目脚手架：Go 模块（go 1.27.4）、`cmd/kready` 入口、`Makefile`（fmt / lint / test / build / check）、CI 工作流。
- 版本与发布规范：`docs/design/versioning.md`。
- 决策分级与禁止事项：`AGENTS.md`。
- ADR 0001：节点级检查通过 SSH 只读采集（root 账号，节点清单由 kubectl 获取），见 `docs/decisions/0001-node-collection-via-ssh.md`。
- 设计草案 v1：规则 schema（`docs/design/rule-schema.md`）、输出契约（`docs/design/output-contract.md`）、快照格式（`docs/design/snapshot-format.md`）。规则不读取部署源文件，无法判定的项输出 INFO。

### Changed
- 规则包选择与 VER-01 语义（B 类，已确认）：选包以控制面节点 kubelet 版本为基线；VER-01 改为检查工作节点版本与基线一致（新检查方式 `equal_baseline`）。`rules_rev` 由 1 升至 2。
- 工作节点版本不一致：原为退出码 2（无适用规则），现为退出码 1（VER-01 FAIL），并列出不符合的节点。此前 VER-01 的结论不再适用。
- 控制面版本不一致或无控制面节点：仍为退出码 2。

### Fixed
- 工作节点 kubelet 版本为空时，VER-01 原判为 FAIL（显示 `w-2=`，易误读）。现改为 ERROR，结论异常，退出码 1，并列出缺失版本的节点。采集结果为空属于无法取得实际值，应报 ERROR。
- 节点采集文件缺失时，原实现静默跳过该节点，结论可能误报"正常"。现以 `cluster.nodes` 为准，缺失节点报 ERROR，结论为异常。
- 信息项（INFO）缺值时，集群级原输出空的当前值。现与节点级一致，报 ERROR。

### Changed
- 设计文档移至 `docs/design/kready-design.md`。
- 规则草案移至 `docs/rules/2.32-draft.md`，待转换为 YAML 后删除。
- 输出契约：结论改为"正常 / 异常"，新增 WARN 状态（不改变结论），退出码 0 为正常、不含 FAIL 与 ERROR。
- 规则包 2.32 草案：SYS-04 改为 `mode: info`，不判定。
- 设计文档：第七节退出码与结论按新术语更新；第十一节拆分为"已决定"与"待讨论"。
- 问题说明字段（`severity`、`impact`、`hint`）保留为占位，具体取值与必填规则在实际问题中细化。
- 旧清单归档至 `docs/archive/checklist-v0.md`。
- `AGENTS.md` 修订：增加 A/B/C 决策分级，允许 A 类改动自动执行；禁止直接推送或合并 `main`。
