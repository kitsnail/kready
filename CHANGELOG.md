# CHANGELOG

## [Unreleased]

### Added
- 项目脚手架：Go 模块（go 1.27.4）、`cmd/kready` 入口、`Makefile`（fmt / lint / test / build / check）、CI 工作流。
- 版本与发布规范：`docs/design/versioning.md`。
- 决策分级与禁止事项：`AGENTS.md`。
- ADR 0001：节点级检查通过 SSH 只读采集（root 账号，节点清单由 kubectl 获取），见 `docs/decisions/0001-node-collection-via-ssh.md`。
- 设计草案 v1：规则 schema（`docs/design/rule-schema.md`）、输出契约（`docs/design/output-contract.md`）、快照格式（`docs/design/snapshot-format.md`）。规则不读取部署源文件，无法判定的项输出 INFO。

### Changed
- 设计文档移至 `docs/design/kready-design.md`。
- 规则草案移至 `docs/rules/2.32-draft.md`，待转换为 YAML 后删除。
- 输出契约：结论改为"正常 / 异常"，新增 WARN 状态（不改变结论），退出码 0 为正常、不含 FAIL 与 ERROR。
- 规则包 2.32 草案：SYS-04 改为 `mode: info`，不判定。
- 设计文档：第七节退出码与结论按新术语更新；第十一节拆分为"已决定"与"待讨论"。
- 问题说明字段（`severity`、`impact`、`hint`）保留为占位，具体取值与必填规则在实际问题中细化。
- 旧清单归档至 `docs/archive/checklist-v0.md`。
- `AGENTS.md` 修订：增加 A/B/C 决策分级，允许 A 类改动自动执行；禁止直接推送或合并 `main`。
