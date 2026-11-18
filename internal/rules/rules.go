// Package rules 加载内嵌的规则包（YAML）。
package rules

import (
	"errors"
	"fmt"

	data "github.com/scclabs/kready/rules"
	"gopkg.in/yaml.v3"
)

// Pack 是一份规则包。
type Pack struct {
	Pack      string    `yaml:"pack"`
	RulesRev  int       `yaml:"rules_rev"`
	AppliesTo AppliesTo `yaml:"applies_to"`
	Rules     []Rule    `yaml:"rules"`
}

// AppliesTo 声明规则包适用的组件版本（精确匹配）。
type AppliesTo struct {
	Kubelet []string `yaml:"kubelet"`
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
	// Impact、Hint 为占位，内容待逐项研究确定后填写。空值不输出，不编造内容。
	Impact string `yaml:"impact,omitempty"`
	Hint   string `yaml:"hint,omitempty"`
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

// Select 以控制面节点的 kubelet 版本为基线选择规则包。
// 控制面版本必须唯一且精确匹配某个规则包的 applies_to，否则返回 ErrNoPack，附原因。
func Select(controlPlane []string) (*Pack, error) {
	if len(controlPlane) == 0 {
		return nil, fmt.Errorf("%w: 未找到控制面节点，无法确定基线版本", ErrNoPack)
	}
	if len(controlPlane) > 1 {
		return nil, fmt.Errorf("%w: 控制面节点 kubelet 版本不一致 %v，请先统一", ErrNoPack, controlPlane)
	}
	version := controlPlane[0]
	var supported []string
	for _, pack := range Packs() {
		p, err := Load(pack)
		if err != nil {
			return nil, err
		}
		for _, v := range p.AppliesTo.Kubelet {
			if v == version {
				return p, nil
			}
			supported = append(supported, v)
		}
	}
	return nil, fmt.Errorf("%w: 控制面 kubelet 版本 %s 不在支持范围 %v 内", ErrNoPack, version, supported)
}

// ErrNoPack 表示无适用规则。
var ErrNoPack = errors.New("无适用规则")

// Packs 列出内嵌的规则包标识。当前仅 2.32。
func Packs() []string {
	return []string{"2.32"}
}
