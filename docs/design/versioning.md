# 版本与发布规范

> 状态：生效（草案转正前，以 AGENTS.md 为准）。
> 适用：kready 工具、规则包、输出契约。

## 一、版本维度

| 维度 | 格式 | 消费者 | 说明 |
|---|---|---|---|
| kready 工具 | SemVer `vMAJOR.MINOR.PATCH` | 运维、CI | 由 git tag 驱动 |
| 规则包 | `pack` + `rules_rev` | 报告、审计 | `pack` 为 kubespray 小版本（如 `2.32`），`rules_rev` 为整数 |
| 输出契约 | `schema_version` | Agent、脚本 | 含 JSON 字段与退出码 |
| 适用对象 | kubespray `2.32.x`，K8s `v1.36.4` | 识别逻辑 | 由部署策略决定，不随工具版本变化 |

kready 版本不跟随 kubespray 版本。新增 kubespray 版本通过新增规则包支持。

## 二、kready 版本（SemVer）

- **0.x 阶段**：开发与验证期。小版本可含破坏性变更，但必须写入 CHANGELOG。
- **1.0.0 门槛**：2.32 规则包在真实集群验证通过，误报率达到约定标准，CLI 与 JSON 契约冻结。
- **1.0 之后**：
  - PATCH：缺陷修复，不改变判定结论；
  - MINOR：新增规则包或检查类型，向后兼容；
  - MAJOR：CLI、退出码或 JSON 结构的破坏性变更。
- **版本注入**：由构建参数注入，不硬编码。发布版本须可追溯到 tag。

## 三、规则包版本

- `pack`：适用性判断的键，如 `2.32`。
- `rules_rev`：任何影响判定结论的改动都递增，包括新增规则、修改期望值、修改检查逻辑。
- 一个 kready 二进制可内嵌多个规则包，规则版本不等同于工具版本。
- 报告输出：`kready 版本 + pack + rules_rev + 基线哈希`。

## 四、输出契约版本

- 只增加可选字段：`schema_version` 不变。
- 删除或重命名字段、改变退出码含义：递增 `schema_version`，并判为 MAJOR 变更。
- 退出码：0 正常，1 异常，2 工具错误。

## 五、提交规范

- Conventional Commits：`feat`、`fix`、`docs`、`chore`、`refactor`、`test`。
- 规则改动使用 scope `rules`，例如 `feat(rules): add CP-09`。
- 破坏性变更在脚注中写 `BREAKING CHANGE:`。

## 六、版本计算（A 类，由脚本完成）

自上一个 tag 起，按提交类型计算下一版本：

| 提交 | 0.x 阶段 | 1.0 之后 |
|---|---|---|
| `fix` | PATCH | PATCH |
| `feat` | PATCH | MINOR |
| `BREAKING CHANGE` | MINOR | MAJOR |

版本由 release-please 计算，不手写，不自定义脚本。

- 工具：release-please（清单模式），配置见 `release-please-config.json`，当前版本见 `.release-please-manifest.json`。
- 流程：推送到 `main` 后，工作流创建或更新 Release PR。合并 Release PR 后，由人确认并打正式 tag。
- 0.x 规则：`bump-minor-pre-major` 为真，即破坏性变更升 MINOR，`feat`、`fix` 升 PATCH。**这是项目选择，不是 SemVer 或社区的强制规定**；SemVer 只规定 0.x 阶段"可以随时变化"。
- 1.0 之后：破坏性变更升 MAJOR，`feat` 升 MINOR，`fix` 升 PATCH。
- 首个版本：`v0.1.0`，由清单文件设定。

## 七、分支与发布

- `main` 始终满足 `make check`。
- 功能分支合入 `main` 由人执行，合并后删除分支。
- `rc` 标签（如 `v0.3.0-rc.1`）：Agent 可在 `make check` 通过后打，用于真实集群验证。
- 正式标签（如 `v0.3.0`）：由人执行，并确认 CHANGELOG 与规则变更说明。
- 发布不按日历，有规则变更或缺陷修复时发布。

## 八、CHANGELOG

- 格式：Keep a Changelog，分 Added / Changed / Deprecated / Removed / Fixed。
- 规则变更列出规则 ID，并说明对历史结论的影响。例如：`CP-06 期望值由 X 改为 Y，此前的 PASS 结论不再适用`。
- 废弃规则：标记 Deprecated，至少保留一个 MINOR 版本后才能移除。

## 九、待实施（B 类，需人工批准）

以下改动涉及门禁文件，尚未实施：

1. `Makefile`：通过构建参数注入版本号。
2. `scripts/next-version`：按提交计算版本。
3. `commit-msg` 钩子：拒绝不符合 Conventional Commits 的提交。
4. CI：校验版本号、tag 与 CHANGELOG 一致。
