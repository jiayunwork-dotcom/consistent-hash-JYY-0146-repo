// Package rebalance 提供节点增删后的重平衡策略。
// 当哈希环中的节点发生变化时，需要计算哪些键需要迁移到新的节点。
package rebalance

import (
	"math"
	"sort"
)

// Transfer 表示一个键从旧节点到新节点的迁移记录。
type Transfer struct {
	// Key 是需要迁移的键
	Key string
	// From 是键的原始所属节点
	From string
	// To 是键的新所属节点
	To string
}

// Plan 计算节点变更后需要迁移的键列表。
// oldNodes 和 newNodes 分别是变更前后的节点集合，
// keys 是所有需要检查的键，hashFn 是哈希函数。
func Plan(oldNodes, newNodes []string, keys []string, hashFn func(string) uint32) []Transfer {
	if len(oldNodes) == 0 || len(newNodes) == 0 || len(keys) == 0 {
		return nil
	}

	var transfers []Transfer
	for _, key := range keys {
		oldOwner := findOwner(oldNodes, key, hashFn)
		newOwner := findOwner(newNodes, key, hashFn)
		if oldOwner != newOwner {
			transfers = append(transfers, Transfer{
				Key:  key,
				From: oldOwner,
				To:   newOwner,
			})
		}
	}
	return transfers
}

// findOwner 使用哈希函数确定键属于哪个节点（简单取模方式）。
func findOwner(nodes []string, key string, hashFn func(string) uint32) string {
	if len(nodes) == 0 {
		return ""
	}
	h := hashFn(key)
	idx := int(h) % len(nodes)
	if idx < 0 {
		idx += len(nodes)
	}
	sorted := make([]string, len(nodes))
	copy(sorted, nodes)
	sort.Strings(sorted)
	return sorted[idx]
}

// EstimateMovement 估算节点数量变化后需要迁移数据的比例。
// 理论上，从 oldCount 变为 newCount 时，约 1/newCount 的数据需要迁移（添加节点时），
// 约 1/oldCount 的数据需要迁移（移除节点时）。
func EstimateMovement(oldCount, newCount int) float64 {
	if oldCount <= 0 || newCount <= 0 {
		return 0
	}
	if oldCount == newCount {
		return 0
	}
	// 理想一致性哈希中的迁移比例估算
	diff := math.Abs(float64(newCount - oldCount))
	maxNodes := math.Max(float64(oldCount), float64(newCount))
	return diff / maxNodes
}

// MinimalTransfer 通过对比新旧映射关系，计算最小化的迁移列表。
// old 和 new 分别是节点到其所拥有的键列表的映射。
func MinimalTransfer(old, new map[string][]string) []Transfer {
	// 构建键到旧归属节点的反向索引
	keyToOld := make(map[string]string)
	for node, keys := range old {
		for _, k := range keys {
			keyToOld[k] = node
		}
	}

	// 构建键到新归属节点的反向索引
	keyToNew := make(map[string]string)
	for node, keys := range new {
		for _, k := range keys {
			keyToNew[k] = node
		}
	}

	var transfers []Transfer
	for key, oldNode := range keyToOld {
		newNode, exists := keyToNew[key]
		if !exists {
			continue
		}
		if oldNode != newNode {
			transfers = append(transfers, Transfer{
				Key:  key,
				From: oldNode,
				To:   newNode,
			})
		}
	}

	sort.Slice(transfers, func(i, j int) bool {
		return transfers[i].Key < transfers[j].Key
	})
	return transfers
}
