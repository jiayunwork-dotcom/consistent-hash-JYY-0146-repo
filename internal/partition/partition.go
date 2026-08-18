// Package partition 提供基于令牌范围的分区管理功能。
// 类似于 Cassandra 的分区策略，将哈希空间划分为连续的范围并分配给节点。
package partition

import (
	"math"
	"sort"
)

// Range 表示哈希空间中的一个连续范围。
type Range struct {
	// Start 是范围的起始位置（含）
	Start uint32
	// End 是范围的结束位置（含）
	End uint32
	// Owner 是负责该范围的节点名称
	Owner string
}

// Partition 将哈希空间均匀划分给各节点。
// 使用 hashFn 为每个节点计算一个令牌位置，然后按位置排序分配范围。
func Partition(nodes []string, hashFn func(string) uint32) []Range {
	if len(nodes) == 0 {
		return nil
	}

	type token struct {
		pos  uint32
		node string
	}

	tokens := make([]token, len(nodes))
	for i, node := range nodes {
		tokens[i] = token{pos: hashFn(node), node: node}
	}

	sort.Slice(tokens, func(i, j int) bool {
		return tokens[i].pos < tokens[j].pos
	})

	ranges := make([]Range, len(tokens))
	for i := 0; i < len(tokens); i++ {
		var start uint32
		if i == 0 {
			// 第一个范围从上一个令牌的下一个位置开始（环绕）
			start = tokens[len(tokens)-1].pos + 1
		} else {
			start = tokens[i-1].pos + 1
		}
		ranges[i] = Range{
			Start: start,
			End:   tokens[i].pos,
			Owner: tokens[i].node,
		}
	}
	return ranges
}

// Merge 合并相邻的、属于同一所有者的范围。
func Merge(ranges []Range) []Range {
	if len(ranges) <= 1 {
		return ranges
	}

	// 先按 Start 排序
	sorted := make([]Range, len(ranges))
	copy(sorted, ranges)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Start < sorted[j].Start
	})

	merged := []Range{sorted[0]}
	for i := 1; i < len(sorted); i++ {
		last := &merged[len(merged)-1]
		curr := sorted[i]
		if curr.Owner == last.Owner && curr.Start <= last.End+1 {
			// 合并
			if curr.End > last.End {
				last.End = curr.End
			}
		} else {
			merged = append(merged, curr)
		}
	}
	return merged
}

// FindOwner 在范围列表中查找给定令牌的所有者。
func FindOwner(ranges []Range, token uint32) string {
	for _, r := range ranges {
		if r.Start <= r.End {
			// 正常范围
			if token >= r.Start && token <= r.End {
				return r.Owner
			}
		} else {
			// 环绕范围（start > end 意味着跨越了 uint32 的最大值）
			if token >= r.Start || token <= r.End {
				return r.Owner
			}
		}
	}
	return ""
}

// EvenPartition 将哈希空间完全均匀地划分给各节点。
func EvenPartition(nodes []string) []Range {
	if len(nodes) == 0 {
		return nil
	}

	rangeSize := math.MaxUint32 / uint32(len(nodes))
	ranges := make([]Range, len(nodes))
	for i, node := range nodes {
		start := uint32(i) * rangeSize
		end := start + rangeSize - 1
		if i == len(nodes)-1 {
			end = math.MaxUint32
		}
		ranges[i] = Range{
			Start: start,
			End:   end,
			Owner: node,
		}
	}
	return ranges
}

// Size 返回范围覆盖的哈希空间大小。
func Size(r Range) uint64 {
	if r.Start <= r.End {
		return uint64(r.End-r.Start) + 1
	}
	// 环绕范围
	return uint64(math.MaxUint32-r.Start) + uint64(r.End) + 2
}
