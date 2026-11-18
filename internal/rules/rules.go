// Package rules 加载内嵌的规则包（YAML）。
package rules

import (
	"fmt"

	data "github.com/scclabs/kready/rules"
	"gopkg.in/yaml.v3"
)

// Pack 是一份规则包。
type Pack struct {
	Pack     string `yaml:"pack"`
	RulesRev int    `yaml:"rules_rev"`
	Rules    []Rule `yaml:"rules"`
}

// Rule 是一条规则，字段见 docs/design/rule-schema.md。
type Rule struct {
	ID      string            `yaml:"id"`
	Title   string            `yaml:"title"`
	Status  string            `yaml:"status"`
	Mode    string            `yaml:"mode"`
	Scope   string            `yaml:"scope"`
	Source  string            `yaml:"source"`
	Collect *Collect          `yaml:"collect,omitempty"`
	Field   string            `yaml:"field,omitempty"`
	Check   string            `yaml:"check,omitempty"`
	Expect  map[string]string `yaml:"expect,omitempty"`
}

// Collect 声明节点级取数方式，命令 ID 必须在采集层白名单中。
type Collect struct {
	Command string `yaml:"command"`
	AsRoot  bool   `yaml:"as_root"`
}

// Load 读取指定规则包（如 "2.32"）。
func Load(pack string) (*Pack, error) {
	raw, err := data.Data.ReadFile(pack + "/rules.yaml")
	if err != nil {
		return nil, fmt.Errorf("无适用规则: %s", pack)
	}
	var p Pack
	if err := yaml.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("规则包解析失败: %w", err)
	}
	if p.Pack != pack {
		return nil, fmt.Errorf("规则包标识不一致: 期望 %s，实际 %s", pack, p.Pack)
	}
	return &p, nil
}

// Active 返回参与输出的规则（status 为 active）。
func (p *Pack) Active() []Rule {
	var out []Rule
	for _, r := range p.Rules {
		if r.Status == "active" {
			out = append(out, r)
		}
	}
	return out
}
