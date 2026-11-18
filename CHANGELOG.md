# CHANGELOG

本文件遵循 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/) 格式。版本号遵循 SemVer，0.x 阶段的规则见 `docs/design/versioning.md`。

## [Unreleased]

### Changed
- CI 门禁（方案 B、C）：修改 `Makefile`、`.github/`、`.githooks/`、`rules/`、`testdata/` 的 PR 须在正文引用 `D-nnn`；修改 `internal/`、`cmd/`、`rules/` 的 PR 须同时更新 `CHANGELOG.md`。仅对 `pull_request` 事件生效。
- PR 模板（`.github/pull_request_template.md`）：要求写明范围、门禁文件与引用的 `D-nnn`、CHANGELOG 状态与 `make check` 结果。
- 仓库迁移至 `kitsnail/kready`，Go 模块路径同步改为 `github.com/kitsnail/kready`（`go.mod` 与全部 import）。旧仓库 `scclabs/kready` 已删除。
- 决策留痕：回溯补记 D-004 至 D-007，并将原 D-008 并入 D-001（见 `docs/product/decisions.md`）。
- 流程文档整理：`docs/design/versioning.md` 的发布权责与 AGENTS.md 一致（正式 tag 由 Agent 执行，人负责验收）；清理已完成的待办项。
- 新增 `docs/product/`：产品输入的目录结构、需求模板与开发流程关系（由人维护）。
- 仓库设置：开启「合并后自动删除源分支」（`delete_branch_on_merge`），与 AGENTS.md「分支」段一致。
- CHANGELOG 按 Keep a Changelog 重组：v0.1.0 之前的条目归入 `[0.1.0]`。

### Notes
- 文档整理提交 `2c67781` 未经 PR 直接落入 `main`，违反 AGENTS.md 禁止事项 6；经人工确认保留，内容已通过 `make check`。

## [0.1.0] - 2026-11-18

首个版本，0.x 开发与验证期。

### Added
- 检查切片（2.32）：规则包 `rules/2.32/rules.yaml`（8 条判定；NODE-05、SYS-04 为信息项）；`kready check --snapshot <目录>` 离线判定，输出正常/异常，异常时列出实际值与期望值。
- 实时采集：`kready check --live`。kubectl 读取节点清单并选包；按规则包声明的命令，通过 SSH（root，ssh-agent，校验 known_hosts）只读采集节点。`--snapshot` 与 `--live` 二选一。
- 采集白名单：`internal/collect` 将命令 ID 映射为固定只读命令（4 条，与规则包一致），拒绝未知 ID。
- 本地目录标准：统一放在 `KREADY_HOME`（默认 `$HOME/.kready`），子目录 `bin/ config/ logs/ reports/ snapshots/ tmp/`，目录 0700，文件 0600（`docs/design/kready-home.md`）。
- 默认写入：`check` 每次把报告写入 `reports/`；`--live` 另把采集快照写入 `snapshots/<时间戳>/`，可用 `--snapshot` 复现判定。
- 版本注入：`make build` 通过 `-ldflags` 注入 `main.version`（来源 `git describe`）；`kready version` 可显示版本；`.goreleaser.yaml` 仅定义构建。
- 输出：结论三态 `正常` / `异常` / `未能确认`；文本报告分为【必须处理】【未能检查】【参考信息】三段；ERROR 带 `source`（`collect` / `tool-gap`）；每条结果以中文标题开头，ID 放括号中；JSON 输出含 `title`。
- 文本报告第二行输出节点数（`节点数: N`），仅作参考，不做判定。
- 规则包选择：按控制面节点 kubelet 版本精确匹配 `applies_to.kubelet`；无匹配或控制面版本不一致输出"无适用规则"，退出码 2。
- 依赖：`golang.org/x/crypto`（SSH 采集，ADR 0001）、`gopkg.in/yaml.v3`（规则包解析）。
- 项目脚手架：Go 模块、`cmd/kready` 入口、`Makefile`（fmt / lint / test / build / check）、CI 工作流、release-please 工作流。
- 文档：设计 v1（规则 schema、输出契约、快照格式、本地目录标准）、ADR 0001、决策分级与禁止事项（`AGENTS.md`）、根目录 `README.md`。

### Changed
- **破坏性变更**：JSON `schema_version` 由 1 递增为 2，退出码 `1` 表示异常或未能确认；工具未覆盖项不再使退出码为 1。
- **破坏性变更**：JSON `schema_version` 由 2 递增为 3。`summary` 改名为 `rules`（按规则计数，取最严重状态）；按结果的计数改名为 `results_by_status`（按节点结果条数）。
- 规则包选择与 VER-01 语义（B 类，已确认）：选包以控制面节点 kubelet 版本为基线；VER-01 改为检查工作节点版本与基线一致（`equal_baseline`）。`rules_rev` 由 1 升至 2。此前 VER-01 的结论不再适用。
- 工作节点版本不一致：原为退出码 2，现为退出码 1（VER-01 FAIL），并列出不符合的节点。
- 控制面版本不一致或无控制面节点：仍为退出码 2。
- 【工具未覆盖】同一原因的项合并为一行，并列出各项名称。
- 去掉“（待补充）”空占位：规则未填写影响与处理时不显示。
- 规则包 2.32 草案：SYS-04 改为 `mode: info`，不判定。
- 设计文档：第七节退出码与结论按新术语更新；第十一节拆分为"已决定"与"待讨论"。
- 问题说明字段（`severity`、`impact`、`hint`）保留为占位，具体取值在实际问题中细化。
- 旧清单归档至 `docs/archive/checklist-v0.md`；规则草案位于 `docs/rules/2.32-draft.md`，待转换为 YAML 后删除。
- 治理规则：AGENTS.md 增加 A/B/C 决策分级，允许 A 类改动自动执行；禁止直接推送或合并 `main`。
- 结论口径：输出契约改为"正常 / 异常"，新增 WARN 状态（不改变结论）；退出码 0 为正常，不含 FAIL 与 ERROR。

### Fixed
- 文本报告：相同取值的节点合并为一行（节点名全部保留）。
- SSH 认证改为 agent 优先、默认私钥兜底；host key 仍严格校验。
- `sshd -T` 输出原样保存导致 NODE-04 实际值为整段配置。现固定提取 `passwordauthentication` 一行。
- 工作节点 kubelet 版本为空时，VER-01 原判为 FAIL（显示 `w-2=`，易误读）。现改为 ERROR，结论异常，退出码 1。
- 节点采集文件缺失时，原实现静默跳过该节点，结论可能误报"正常"。现以 `cluster.nodes` 为准，缺失节点报 ERROR。
- 信息项（INFO）缺值时，集群级原输出空的当前值。现与节点级一致，报 ERROR。

### Open
- NODE-05（内核自动升级服务状态）：判定方案未定，暂为信息项（见 `docs/rules/2.32-draft.md` 待确认 5）。
