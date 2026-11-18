# CHANGELOG

## [Unreleased]

### Changed
- 流程文档整理：`docs/design/versioning.md` 的发布权责改为与 AGENTS.md 一致（正式 tag 由 Agent 执行，人负责验收）；清理已完成的待办项。
- 新增 `docs/product/README.md`：产品输入的目录结构、需求模板与与开发流程的关系（由人维护）。
- release-please 改用内置 `GITHUB_TOKEN`，移除长期 PAT（原 `RELEASE_PLEASE_TOKEN` 已过期，导致工作流 `Bad credentials` 静默失败）。配套打开组织设置 `can_approve_pull_request_reviews`，Actions 因而可创建 PR。`docs/design/versioning.md` 补充凭据要求。
- 分支生命周期：功能分支合并后即删除，仓库开启「合并后自动删除源分支」；清理已合并分支 5 个。规则见 `AGENTS.md` 的「分支」段。
- 项目 logo：`docs/assets/logo.svg`。实心蓝六边形配白色对勾与右上角绿色状态点，六边形环留缺口容纳状态点。无文字，缩放后仍可辨认。
- README 顶部居中显示 logo。
- 新增根目录 `README.md`：项目定位、适用范围、快速开始（附真实输出）、用法、退出码、开发命令与文档索引。
- release-please 工作流改用 `RELEASE_PLEASE_TOKEN`（个人访问令牌）创建 Release PR，绕开仓库中被禁用的 Actions 建 PR 权限。
- 构建注入版本：`make build` 通过 `-ldflags` 注入 `main.version`，来源为 `git describe`。`kready version` 可显示版本。新增 `.goreleaser.yaml`，仅定义构建，不启用发布。
- 治理规则：Agent 可合并自己开的、满足合并条件的 PR（门禁文件仍需人工批准）；可在合并到 `main` 后打正式版本 tag。见 `AGENTS.md`、`docs/design/versioning.md`。
- 版本计算改用 release-please（社区标准工具），替代自定义的 `scripts/next-version.sh`，该脚本已删除。
- 新增 `release-please` 工作流：只创建 Release PR，正式 tag 由人执行。
- 结论三态：`正常`、`异常`、`未能确认`。结论只描述集群。`未能确认` 表示采集失败，无法确认集群状态，不得写成正常。
- ERROR 结果增加 `source` 字段：`collect`（采集失败，影响结论）与 `tool-gap`（工具未覆盖，不影响结论）。
- 文本报告独立成块：【未能确认】（附处理方式）与【工具未覆盖】（注明“不代表集群有问题”）。
- 输出框架：文本报告分为【必须处理】【未能检查】【参考信息】三段。规则增加可选字段 `impact`、`hint`（当前为空，输出 `（待补充）`，待逐项研究后填写）。结论口径暂未变更。
- 输出可读性：每条结果以中文标题开头，ID 放括号中（`标题（ID）`）；标题取自规则包。文本报告按状态分段：不合规、无法判定、信息，各项竖排。JSON 结果新增可选字段 `title`。
- 文本报告第二行输出节点数（`节点数: N`），如实反映 kubectl 发现的节点数。节点数不作为判定项，不做期望检查。
- 本地目录标准：`docs/design/kready-home.md`。统一放在 `KREADY_HOME`（默认 `$HOME/.kready`），子目录 `bin/ config/ logs/ reports/ snapshots/ tmp/` 为项目标准。目录 0700，文件 0600。
- 默认写入：`check` 每次把报告写入 `reports/`（文本或 JSON，与终端一致）；`--live` 另把采集快照写入 `snapshots/<时间戳>/`，可用 `--snapshot` 复现判定。同一秒多次运行加序号，不覆盖旧文件。
- 失败策略：目录不可用时判定前报错（退出码 2）；报告或快照写入失败只警告，不改变判定结果。
- 实时采集：`kready check --live`。kubectl 只读取节点清单（`get nodes`）并选包；选包成功后按规则包声明的命令，通过 SSH（root，ssh-agent，校验 known_hosts）只读采集节点。`--snapshot` 与 `--live` 二选一。
- 采集白名单：`internal/collect` 将命令 ID 映射为固定只读命令（4 条，与规则包一致）。执行层只接收 ID，拒绝未知 ID。
- 依赖：`golang.org/x/crypto`。理由：标准库无 SSH 客户端，ADR 0001 要求 SSH 采集且校验 known_hosts。
- 原型验证记录：`docs/design/prototype-verification.md`，列出 10 项已验证行为（选包、基线、ERROR、JSON、退出码）。
- 检查切片（2.32）：规则包 `rules/2.32/rules.yaml`（8 条判定；NODE-05、SYS-04 为信息项）；`kready check --snapshot <目录>` 离线读取快照，输出正常/异常，异常时列出实际值与期望值。
- 快照测试：`testdata/positive/2.32`（全部通过）与 `testdata/negative/2.32/<RULE-ID>/`（每条判定规则一个负例）。测试只使用快照。
- NODE-05（内核自动升级服务状态）：判定方案未定，暂为信息项，仅输出当前值（见 `docs/rules/2.32-draft.md` 待确认 5）。
- 规则包选择：按快照中节点的 kubelet 版本选包（规则包 `applies_to.kubelet` 精确匹配）。无匹配或节点版本不一致时输出"无适用规则"，退出码 2，不做判定。
- JSON 输出：`kready check --snapshot <目录> --format json`，结构见 `docs/design/output-contract.md`（`schema_version` 2）。
- 依赖：`gopkg.in/yaml.v3`。理由：规则包使用 YAML 格式，标准库无 YAML 解析器。
- 项目脚手架：Go 模块（go 1.27.4）、`cmd/kready` 入口、`Makefile`（fmt / lint / test / build / check）、CI 工作流。
- 版本与发布规范：`docs/design/versioning.md`。
- 决策分级与禁止事项：`AGENTS.md`。
- ADR 0001：节点级检查通过 SSH 只读采集（root 账号，节点清单由 kubectl 获取），见 `docs/decisions/0001-node-collection-via-ssh.md`。
- 设计草案 v1：规则 schema（`docs/design/rule-schema.md`）、输出契约（`docs/design/output-contract.md`）、快照格式（`docs/design/snapshot-format.md`）。规则不读取部署源文件，无法判定的项输出 INFO。
- **破坏性变更**：JSON `schema_version` 由 2 递增为 3。`summary` 改名为 `rules`（按规则计数，每条规则计一次，取最严重状态，给决策用）；按结果的计数改名为 `results_by_status`（按节点结果条数，给脚本对账用）。
- 【工具未覆盖】同一原因的项合并为一行，并列出各项名称。
- 退出码：`1` 表示结论为异常或未能确认。工具未覆盖不再使退出码为 1。**破坏性变更**：JSON `schema_version` 由 1 递增为 2。
- 去掉“（待补充）”空占位：规则未填写影响与处理时不显示。
- 规则包选择与 VER-01 语义（B 类，已确认）：选包以控制面节点 kubelet 版本为基线；VER-01 改为检查工作节点版本与基线一致（新检查方式 `equal_baseline`）。`rules_rev` 由 1 升至 2。
- 工作节点版本不一致：原为退出码 2（无适用规则），现为退出码 1（VER-01 FAIL），并列出不符合的节点。此前 VER-01 的结论不再适用。
- 控制面版本不一致或无控制面节点：仍为退出码 2。

### Fixed
- 文本报告：相同取值的节点合并为一行（节点名全部保留），避免同一个值在多个节点上重复多行。
- SSH 认证改为 agent 优先、默认私钥兜底；节点已配置免密登录，不再要求 ssh-agent。host key 仍严格校验。
- `sshd -T` 输出（约 95 行）原样保存导致 NODE-04 实际值为整段配置。现固定提取 `passwordauthentication` 一行，报告可读。
- 工作节点 kubelet 版本为空时，VER-01 原判为 FAIL（显示 `w-2=`，易误读）。现改为 ERROR，结论异常，退出码 1，并列出缺失版本的节点。采集结果为空属于无法取得实际值，应报 ERROR。
- 节点采集文件缺失时，原实现静默跳过该节点，结论可能误报"正常"。现以 `cluster.nodes` 为准，缺失节点报 ERROR，结论为异常。
- 信息项（INFO）缺值时，集群级原输出空的当前值。现与节点级一致，报 ERROR。

- 设计文档移至 `docs/design/kready-design.md`。
- 规则草案移至 `docs/rules/2.32-draft.md`，待转换为 YAML 后删除。
- 输出契约：结论改为"正常 / 异常"，新增 WARN 状态（不改变结论），退出码 0 为正常、不含 FAIL 与 ERROR。
- 规则包 2.32 草案：SYS-04 改为 `mode: info`，不判定。
- 设计文档：第七节退出码与结论按新术语更新；第十一节拆分为"已决定"与"待讨论"。
- 问题说明字段（`severity`、`impact`、`hint`）保留为占位，具体取值与必填规则在实际问题中细化。
- 旧清单归档至 `docs/archive/checklist-v0.md`。
- `AGENTS.md` 修订：增加 A/B/C 决策分级，允许 A 类改动自动执行；禁止直接推送或合并 `main`。
