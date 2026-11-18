# 产品决策记录（decisions）

> 维护者：人（产品经理）。内容由人决定；Agent 仅按下方留痕规则补记条目，不改写已有条目。
> 留痕规则：B 类决策（规则、期望值、废弃、依赖、门禁文件）经人批准后，以及 REQ 未确认前的检查逻辑改动（D-011）经人批准后，Agent 在实施的同一 PR 中补记一条 `D-nnn`，并在 CHANGELOG 中引用。门禁文件的 PR 描述须引用一条 `D-nnn`，否则不合并。依据写明 PR 号、提交号或原始指令，不编造日期。
> 作用：记录范围、取舍、废弃与发布相关的产品决策，保证可追溯。
> 与 `docs/decisions/`（技术 ADR）分工：本文件只记产品层决策。

## 条目模板

```markdown
### D-000 <决策标题>（YYYY-MM-DD）

- 决策：<决定了什么>
- 原因：<为什么>
- 影响：<影响哪些 REQ、规则 ID 或发布>
- 决策人：<人>
```

## 决策列表

### D-001 发布版本从 0.x 起步，不以日历发布（2026-11-18）

- 决策：kready 0.x 阶段按需求与缺陷修复发布，不按日历。
- 原因：规则变更与缺陷修复驱动发布，避免空版本。
- 影响：`docs/design/versioning.md` 第七节；发布权责（release-please 计算版本，Agent 打正式 tag，人验收）见 `AGENTS.md`「提交与版本」与 `docs/design/versioning.md` 第七节（原 D-008 并入此条）。
- 决策人：kitsnail

### D-002 0.x 阶段 feat 与 fix 同样升 PATCH（2026-11-18）

- 决策：`bump-minor-pre-major` 与 `bump-patch-for-minor-pre-major` 启用，破坏性变更升 MINOR，feat / fix 升 PATCH。
- 原因：0.x 阶段保持版本号稳定，1.0 前不引入频繁的 MINOR。
- 影响：release-please 配置；`docs/design/versioning.md` 第六节。
- 决策人：kitsnail

### D-003 本阶段范围限定为软件工程流程，暂缓业务规则（2026-11-19）

- 决策：当前阶段只处理软件工程流程与生命周期管理。2.32 规则包的业务内容暂缓，包括：
  - `docs/rules/2.32-draft.md` 中未转为 YAML 的规则（VER-02、VER-03、VER-05、CP-02~07、NODE-01、SYS-01~03）；
  - 草案待确认项 POL-01、POL-03、OS 内核参数的处理；
  - 草案文件的删除，以及规则、期望值、检查逻辑的任何变更。
- 原因：业务规则由人确认，需要单独的产品输入；流程问题可以先独立解决，不依赖业务规则的定稿。
- 影响：`rules/`、`docs/rules/` 暂不改动；恢复业务规则工作时，需新增 REQ 并在 `backlog.md` 排期。
- 决策人：kitsnail

### D-004 暂缓 NODE-05 判定、ERROR 文案与真实采集（2026-11-18）

- 决策：NODE-05 保持信息项、不判定；ERROR 文案保持占位；真实采集暂缓，原型阶段以快照验证为主。
- 原因：原型阶段规则够用即可，判定方案未定前不做真实采集。
- 影响：NODE-05（`rules/2.32/rules.yaml`、`docs/rules/2.32-draft.md` 待确认第 5 条）；`docs/design/prototype-verification.md`。
- 依据：提交 `6726f3a`。
- 决策人：kitsnail

### D-005 规则包选择与 VER-01 语义（2026-11-18）

- 决策：规则包以控制面节点 kubelet 版本为基线选包；VER-01 改为检查工作节点版本与基线一致（`equal_baseline`），`rules_rev` 由 1 升至 2。
- 原因：控制面版本决定适用的规则包，工作节点需与之一致，不一致即为异常。
- 影响：VER-01；`rules/2.32/rules.yaml`；输出退出码语义（工作节点不一致由退出码 2 改为 1）。
- 依据：CHANGELOG `[0.1.0]` 的 Changed 条目（标注"B 类，已确认"）；提交 `f12c241`。
- 决策人：kitsnail

### D-006 节点级检查通过 SSH 只读采集（2026-11-18）

- 决策：节点级检查经 SSH 只读采集，使用 root 账号，校验 known_hosts；节点清单由 kubectl 获取。
- 原因：规则中的节点检查（inotify、sshd 配置、内核自动升级等）需读取节点本机状态，Kubernetes API 无法提供。
- 影响：`internal/collect`；`golang.org/x/crypto` 依赖（见 CHANGELOG `[0.1.0]`）。
- 依据：`docs/decisions/0001-node-collection-via-ssh.md`（状态：已接受，2026-11-18）。
- 决策人：kitsnail

### D-007 本地目录标准 KREADY_HOME（2026-11-18）

- 决策：所有部署、配置、日志、报告与临时数据统一放在 `KREADY_HOME`（默认 `$HOME/.kready`），子目录 `bin/ config/ logs/ reports/ snapshots/ tmp/`，目录 0700，文件 0600。
- 原因：作为未来项目的标准，便于统一管理与清理。
- 影响：`internal/home`；`docs/design/kready-home.md`。
- 依据：`docs/design/kready-home.md`，文档状态写明"已确认（用户决定，作为未来项目的标准）"。
- 决策人：kitsnail

### D-009 CI 门禁：门禁文件须引用 D-nnn，产品改动须更新 CHANGELOG（2026-11-19）

- 决策：修改 `Makefile`、`.github/`、`.githooks/`、`rules/`、`testdata/` 的 PR，正文须引用一条已批准的 `D-nnn`，否则 CI 失败；修改 `internal/`、`cmd/`、`rules/` 的 PR，须同时更新 `CHANGELOG.md`，否则 CI 失败。两项检查仅对 `pull_request` 事件生效。
- 原因：把"门禁须经人工批准"与"每次产品改动同步 CHANGELOG"从人工自觉变成 CI 机制。
- 影响：`.github/workflows/ci.yml`；`CHANGELOG.md` 的维护规则（`docs/design/versioning.md` 第八节）。
- 依据：人批准方案 B、C 的实施（本次会话，2026-11-19）；PR 模板 PR #11 已合并。
- 决策人：kitsnail

### D-010 Agent 编码约束：依赖方向、先复用、命名（2026-11-19）

- 决策：在 AGENTS.md「工作方式」新增 3 条：依赖方向（`internal/check` 不直接 import `internal/collect`）；先复用（新增公共函数前先搜索仓库，有则复用）；命名（函数名说明行为，同一概念全仓库一个名字，单函数只做一件事、参数尽量不超过 3 个）。
- 原因：AI 在不同文件重复实现同一公共功能、函数名各异，造成组织混乱与冗余。
- 影响：`AGENTS.md`（工作方式第 7–9 条）；本决策不改门禁文件、不新增依赖。依赖方向的自动检查（脚本或 lint）暂缓，待后续单独 B 类决策。
- 依据：人确认方案（回复"1 同意，2. 后续再做"，本次会话，2026-11-19）。
- 决策人：kitsnail

### D-011 REQ 确认后的执行授权与独立审查（2026-11-19）

- 决策：
  1. 人在 `backlog.md` 将 REQ 置为「待开始」、且 `acceptance.md` 有对应验收标准后，Agent 在该 REQ 范围内自主完成检查逻辑（`internal/`、`cmd/`）的改动，不再逐次等待确认。
  2. `rules/`、`testdata/` 的期望值仍须人工批准；新增依赖、废弃规则、1.0 转正、其他门禁文件仍为 B 类。
  3. 合并前须有独立审查结论 approve。审查者为未参与实现的子 Agent，只读取 diff、对应 REQ 与验收标准。
  4. 每个 REQ 的自动推进上限为 10 轮；超出、验收标准有歧义、同一阻塞连续 3 轮未变化等情况，停下报告。
  5. Agent 可起草 REQ（标注「草稿」，经人确认后生效），并在 `docs/product/README.md` 的字段级权限范围内回填状态与关联。
- 原因：规则类之外的 B 类逐次确认是主要瓶颈；Agent 自审无效；轮数无上限会导致失控。
- 影响：`AGENTS.md`（决策分级、合并条件、工作方式第 10–12 条）；`docs/product/README.md`（字段级权限）；`CHANGELOG.md`。
- 修订：修订 D-003 中"检查逻辑任何变更暂缓"的部分；D-003 其余条目（规则与期望值暂缓）不变。
- 依据：人回复"Q1 和 Q2 按推荐来，已确认 D-011"（本次会话，2026-11-19）。
- 决策人：kitsnail
