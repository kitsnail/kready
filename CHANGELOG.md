# CHANGELOG

## [Unreleased]

### Added
- 项目脚手架：Go 模块（go 1.27.4）、`cmd/kready` 入口、`Makefile`（fmt / lint / test / build / check）、CI 工作流。
- 版本与发布规范：`docs/design/versioning.md`。
- 决策分级与禁止事项：`AGENTS.md`。
- ADR 0001：节点级检查通过 SSH 只读采集（root 账号，节点清单由 kubectl 获取），见 `docs/decisions/0001-node-collection-via-ssh.md`。

### Changed
- 设计文档移至 `docs/design/kready-design.md`。
- 规则草案移至 `docs/rules/2.32-draft.md`，待转换为 YAML 后删除。
- 旧清单归档至 `docs/archive/checklist-v0.md`。
- `AGENTS.md` 修订：增加 A/B/C 决策分级，允许 A 类改动自动执行；禁止直接推送或合并 `main`。
