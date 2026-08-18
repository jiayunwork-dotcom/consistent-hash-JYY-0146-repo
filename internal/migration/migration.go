// Package migration 提供哈希环状态变更时的键迁移计划功能。
// 当哈希环的节点发生变化时，可以使用此包计算需要迁移的键及其源目标节点。
package migration

import "sort"

// Step 表示一个键的迁移步骤。
type Step struct {
	// Key 是需要迁移的键
	Key string
	// Source 是键的当前所属节点
	Source string
	// Dest 是键的目标节点
	Dest string
}

// MigrationPlan 包含完整的迁移计划。
type MigrationPlan struct {
	// Steps 是所有迁移步骤的有序列表
	Steps []Step
}

// ComputeMigration 通过比较变更前后的键归属函数来计算迁移计划。
// before 和 after 分别是变更前后的键到节点的映射函数，
// keys 是需要检查的所有键。
func ComputeMigration(before, after func(string) string, keys []string) *MigrationPlan {
	plan := &MigrationPlan{}
	if before == nil || after == nil {
		return plan
	}

	for _, key := range keys {
		src := before(key)
		dst := after(key)
		if src != dst && src != "" && dst != "" {
			plan.Steps = append(plan.Steps, Step{
				Key:    key,
				Source: src,
				Dest:   dst,
			})
		}
	}

	// 按键排序以获得确定性的输出
	sort.Slice(plan.Steps, func(i, j int) bool {
		return plan.Steps[i].Key < plan.Steps[j].Key
	})
	return plan
}

// AffectedKeys 返回需要迁移的键的数量。
func (p *MigrationPlan) AffectedKeys() int {
	return len(p.Steps)
}

// BySource 按源节点分组返回迁移步骤。
func (p *MigrationPlan) BySource() map[string][]Step {
	result := make(map[string][]Step)
	for _, step := range p.Steps {
		result[step.Source] = append(result[step.Source], step)
	}
	return result
}

// ByDest 按目标节点分组返回迁移步骤。
func (p *MigrationPlan) ByDest() map[string][]Step {
	result := make(map[string][]Step)
	for _, step := range p.Steps {
		result[step.Dest] = append(result[step.Dest], step)
	}
	return result
}

// Percentage 返回需要迁移的键占总键数的百分比。
func (p *MigrationPlan) Percentage(totalKeys int) float64 {
	if totalKeys <= 0 {
		return 0
	}
	return float64(len(p.Steps)) / float64(totalKeys) * 100
}

// Reverse 生成反向迁移计划（用于回滚）。
func (p *MigrationPlan) Reverse() *MigrationPlan {
	reversed := &MigrationPlan{
		Steps: make([]Step, len(p.Steps)),
	}
	for i, step := range p.Steps {
		reversed.Steps[i] = Step{
			Key:    step.Key,
			Source: step.Dest,
			Dest:   step.Source,
		}
	}
	return reversed
}
