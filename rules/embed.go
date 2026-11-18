// Package rules 嵌入随二进制发布的规则包数据（YAML）。
// 规则数据位于同目录下的版本子目录，只有一份来源。
package rules

import "embed"

// Data 是所有规则包的只读文件系统，路径形如 "2.32/rules.yaml"。
//
//go:embed 2.32/rules.yaml
var Data embed.FS
