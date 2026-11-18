# 产品决策记录（decisions）

> 维护者：人（产品经理）。Agent 只读取，不修改本文件。
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
- 影响：`docs/design/versioning.md` 第七节。
- 决策人：待确认

### D-002 0.x 阶段 feat 与 fix 同样升 PATCH（2026-11-18）

- 决策：`bump-minor-pre-major` 与 `bump-patch-for-minor-pre-major` 启用，破坏性变更升 MINOR，feat / fix 升 PATCH。
- 原因：0.x 阶段保持版本号稳定，1.0 前不引入频繁的 MINOR。
- 影响：release-please 配置；`docs/design/versioning.md` 第六节。
- 决策人：待确认
