package main

import (
	"encoding/json"
	"io"

	"github.com/kitsnail/kready/internal/check"
	"github.com/kitsnail/kready/internal/rules"
)

// schemaVersion 是 JSON 输出的契约版本，见 docs/design/output-contract.md。
// 版本 2：结论新增“未能确认”，退出码 1 含义扩展为异常或未能确认（MAJOR 变更）。
// 版本 3：summary 改名为 rules（按规则计数），按结果计数改名为 results_by_status。
const schemaVersion = 3

// report 是 JSON 输出的结构，字段与输出契约一致。
type report struct {
	SchemaVersion   int            `json:"schema_version"`
	KreadyVersion   string         `json:"kready_version"`
	Pack            string         `json:"pack"`
	RulesRev        int            `json:"rules_rev"`
	Conclusion      string         `json:"conclusion"`
	Rules           map[string]int `json:"rules"`
	ResultsByStatus map[string]int `json:"results_by_status"`
	Results         []check.Result `json:"results"`
}

// countsOf 把状态计数整理为固定六个键，未出现的状态为 0。
func countsOf(m map[check.Status]int) map[string]int {
	return map[string]int{
		"pass":  m[check.Pass],
		"fail":  m[check.Fail],
		"warn":  m[check.Warn],
		"error": m[check.Error],
		"info":  m[check.Info],
		"skip":  m[check.Skip],
	}
}

// writeJSON 按输出契约写出 JSON。
// rules：按规则计数，给决策用（“几件事需要处理”）。
// results_by_status：按节点结果条数计数，给脚本对账用。两组计数各六个键，始终输出。
func writeJSON(w io.Writer, pack *rules.Pack, results []check.Result) error {
	byRule := countsOf(check.RuleStatus(results))
	byResult := countsOf(check.Summary(results))
	out := report{
		SchemaVersion:   schemaVersion,
		KreadyVersion:   version,
		Pack:            pack.Pack,
		RulesRev:        pack.RulesRev,
		Conclusion:      check.Conclusion(results),
		Rules:           byRule,
		ResultsByStatus: byResult,
		Results:         results,
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}
