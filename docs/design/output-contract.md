# 输出契约 v1

> 状态：草案。输出字段属于契约，变更须按 `versioning.md` 处理。

## 退出码

| 码 | 含义 |
|---|---|
| 0 | 全部判定项 PASS（INFO 不影响退出码） |
| 1 | 存在 FAIL 或 ERROR |
| 2 | 工具错误（参数错误、无适用规则、内部错误） |

## 结果状态

| 状态 | 含义 |
|---|---|
| `PASS` | 判定项符合期望 |
| `FAIL` | 判定项不符合期望 |
| `ERROR` | 判定项无法取得实际值（连接失败、超时、命令未注册） |
| `INFO` | 信息项，仅输出当前值，不判定 |
| `SKIP` | 不适用（如无对应角色节点） |

## 规则包选择

- 以观测到的 kubelet 版本选择规则包。`v1.36.4` → `2.32`。
- 没有匹配的规则包 → 退出码 2，输出"无适用规则"，不做判定。
- 工具不读取部署源文件，不做基线哈希校验。

## JSON 结构

```json
{
  "schema_version": 1,
  "kready_version": "0.1.0",
  "pack": "2.32",
  "rules_rev": 1,
  "summary": { "pass": 0, "fail": 0, "error": 0, "info": 0, "skip": 0 },
  "results": [
    {
      "id": "NODE-02",
      "status": "FAIL",
      "target": "cn-192-168-3-52",
      "actual": "128",
      "expect": "1280"
    },
    {
      "id": "SYS-04",
      "status": "INFO",
      "target": "cluster",
      "actual": "3"
    }
  ]
}
```

- `target`：集群级为 `cluster`，节点级为节点名。
- `actual`：实际值，字符串。
- `expect`：仅 `judge` 项有，字符串。
- `ERROR` 结果增加 `error` 字段，内容为简短原因，不含凭据。

## 人类可读输出

- 第一行：结论，`正常` 或 `异常`。
- 异常时列出每条 FAIL 与 ERROR：`ID  目标  实际值  期望值`。
- 信息项单独一段列出：`ID  目标  当前值`。

## 兼容规则

- 新增可选字段：`schema_version` 不变。
- 删除、重命名字段，或改变状态与退出码含义：递增 `schema_version`，按 MAJOR 处理。
