package main

import (
	"encoding/json"
	"io"

	"github.com/scclabs/kready/internal/check"
	"github.com/scclabs/kready/internal/rules"
)

// schemaVersion 是 JSON 输出的契约版本，见 docs/design/output-contract.md。
// 版本 2：结论新增“未能确认”，退出码 1 含义扩展为异常或未能确认（MAJOR 变更）。
const schemaVersion = 2

// report 是 JSON 输出的结构，字段与输出契约一致。
type report struct {
	SchemaVersion int            `json:"schema_version"`
	KreadyVersion string         `json:"kready_version"`
	Pack          string         `json:"pack"`
	RulesRev      int            `json:"rules_rev"`
	Conclusion    string         `json:"conclusion"`
	Summary       map[string]int `json:"summary"`
	Results       []check.Result `json:"results"`
}

// writeJSON 按输出契约写出 JSON。summary 六个计数始终输出，未出现的状态为 0。
func writeJSON(w io.Writer, pack *rules.Pack, results []check.Result) error {
	counts := check.Summary(results)
	summary := map[string]int{
		"pass":  int(counts[check.Pass]),
		"fail":  int(counts[check.Fail]),
		"warn":  int(counts[check.Warn]),
		"error": int(counts[check.Error]),
		"info":  int(counts[check.Info]),
		"skip":  int(counts[check.Skip]),
	}
	out := report{
		SchemaVersion: schemaVersion,
		KreadyVersion: version,
		Pack:          pack.Pack,
		RulesRev:      pack.RulesRev,
		Conclusion:    check.Conclusion(results),
		Summary:       summary,
		Results:       results,
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}
