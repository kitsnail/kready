package rules

// CollectIDs 返回规则包中所有生效规则声明的节点采集命令 ID（去重，保持出现顺序）。
// 节点采集只执行这些命令，且每个 ID 都必须在采集白名单中。
func (p *Pack) CollectIDs() []string {
	seen := map[string]bool{}
	var ids []string
	for _, r := range p.Active() {
		if r.Collect == nil || seen[r.Collect.Command] {
			continue
		}
		seen[r.Collect.Command] = true
		ids = append(ids, r.Collect.Command)
	}
	return ids
}
