# AGENTS.md — kready

## 目标
kready：面向一线运维的 Kubernetes 集群交付就绪检查工具。一条命令，输出"正常/异常"，异常时列出实际值与期望值。
宗旨：简单、易懂、快捷、只读。

## 范围
- 只做能自动判定的检查（只读、确定性、期望值明确）。
- 不做：修复、写操作、测试类（fio、演练）、人工事实（网卡名、VIP、机房 DNS）。
- 当前只支持 kubespray 2.32.x，固定 Kubernetes v1.36.4。基线不匹配 → 输出"无适用规则"，不判定。

## 目录
- `rules/2.32/`：规则包（数据）。每条规则有稳定 ID，如 `CP-06`。
- `internal/`：检查类型实现（有限集合）。
- `cmd/kready/`：CLI 入口。
- `testdata/`：正例快照（全部通过）与负例快照（每条规则至少一个，报对应失败）。
- `docs/design/`：设计文档。`docs/rules/`：规则草案。`docs/product/`：产品文档，由人维护。
- `CHANGELOG.md`：每次变更。

## 命令
- `make fmt` 格式化
- `make lint` 静态检查
- `make test` 单元与样本测试
- `make check` 以上全部，提交前必须通过

## 决策分级
能否用脚本复算出唯一结果，决定由谁执行。

| 级别 | 范围 | Agent 行为 |
|---|---|---|
| A 自动 | 提交（符合 Conventional Commits）；生成 CHANGELOG 草稿；按提交计算版本号；打 `rc` 预发布 tag；打正式版本 tag（合并到 `main` 之后，版本与清单一致）；修复缺陷且不改判定结论；合并 Agent 自己开的、满足合并条件的 PR | 直接执行，`make check` 通过后完成并报告 |
| B 确认 | 新增或修改规则、期望值或检查逻辑；废弃规则；1.0 转正；新增依赖；修改门禁文件 | 准备材料，停下等人确认 |
| C 禁止 | 见下方禁止事项 | 拒绝并说明原因 |

## 合并条件
Agent 合并 PR 须同时满足：
- 由 Agent 自己开出；
- `make check` 通过，CI 通过；
- 无未解决的审查意见；
- 改动不涉及门禁文件（`Makefile`、lint 配置、CI 配置、`rules/`、`testdata/` 的期望值）。涉及门禁文件的 PR 须先经人工批准。

## 禁止事项（硬性）
1. 不得在代码中调用 create / update / patch / delete，仅允许 get / list / watch。
2. 不得擅自修改门禁：`Makefile`、lint 配置、CI 配置、`rules/`、`testdata/` 的期望值，须人工批准。
3. 不得删除或放宽测试来让构建通过。失败就如实报告。
4. 不得读取或输出 Secret 内容、私钥、token。
5. 不得在测试中连接真实集群或外网。测试只用 fake client 和快照。
6. 不得直接推送 `main`。只能推送功能分支并开 PR；仅在满足"合并条件"时合并。
7. 不得跳过 `make check`，不得在 `make check` 失败时提交或发布。
8. 不得为单次需求引入新依赖，新依赖需在 PR 中说明理由。
9. 规则 ID 不复用；废弃规则标记为废弃，不删除。
10. 不得手算版本号，版本由 release-please 计算。

## 工作方式
1. A 类改动：直接执行，完成后报告改了什么、运行了哪些命令、结果如何。
2. B 类改动：先输出方案（改动文件、影响的规则 ID、测试计划），等待确认后再实现。
3. 小步提交：一次只做一个可验证的切片。
4. 每次改动同步更新对应文档与 CHANGELOG。
5. 未运行的命令要明确说明。
6. 不确定时提问，不要猜。规则期望值以 `rules/` 与依据文档为准，不自行推断。
7. 依赖方向（D-010）：判定逻辑（`internal/check`）不得直接 import 数据采集（`internal/collect`）；采集结果经由快照或规则包的数据结构传入。
8. 先复用（D-010）：新增公共函数前，先用搜索（如 `grep -rn`）查找仓库内同功能实现，有则复用，不得另写一份。
9. 命名（D-010）：函数名说明它做什么；同一概念全仓库只用一个名字；一个函数只做一件事，参数尽量不超过 3 个。

## 提交与版本
- 提交信息遵循 Conventional Commits：`feat`、`fix`、`docs`、`chore`、`refactor`、`test`；规则改动用 scope `rules`。
- 版本规则见 `docs/design/versioning.md`。版本号由 release-please 计算，不手写。
- 正式版本 tag 在合并 Release PR 或对应发布提交到 `main` 后，由 Agent 执行，版本须与 `.release-please-manifest.json` 一致。

## 分支

- 命名：`<type>/<slug>`，type 取自提交类型（`feat`、`fix`、`docs`、`chore`、`refactor`、`test`）。
- 从最新 `main` 切出，一个分支只做一个可验证的切片。
- PR 合并后删除分支，远端与本地一并删除。仓库已开启「合并后自动删除源分支」。
- release-please 的分支（`release-please--*`）由工具管理，不手工创建或删除。

## 文档风格
简明，正向陈述，一个概念只定义一次。写"要做什么"，少写"不要做什么"（禁止项除外）。
