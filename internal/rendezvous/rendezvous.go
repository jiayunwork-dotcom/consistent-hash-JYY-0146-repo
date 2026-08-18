// Package rendezvous 实现 Rendezvous（最高随机权重，HRW）哈希算法。
// Rendezvous Hash 通过为每个节点计算与键相关的权重来选择目标节点。
package rendezvous

import (
	"hash/fnv"
	"sort"
)

// RendezvousHash 是 Rendezvous 哈希的主要结构体。
type RendezvousHash struct {
	nodes []string
}

// New 创建一个新的 RendezvousHash 实例。
func New(nodes []string) *RendezvousHash {
	h := &RendezvousHash{
		nodes: make([]string, len(nodes)),
	}
	copy(h.nodes, nodes)
	return h
}

// Get 返回给定键对应的节点。使用最高随机权重算法选择。
func (h *RendezvousHash) Get(key string) string {
	if len(h.nodes) == 0 {
		return ""
	}
	var maxHash uint64
	var maxNode string
	for _, node := range h.nodes {
		w := weight(key, node)
		if w > maxHash || maxNode == "" {
			maxHash = w
			maxNode = node
		}
	}
	return maxNode
}

// GetN 返回给定键对应的前 N 个节点（按权重降序排列）。
// 这对于复制策略很有用：可以将数据复制到权重最高的 N 个节点。
func (h *RendezvousHash) GetN(key string, n int) []string {
	if len(h.nodes) == 0 || n <= 0 {
		return nil
	}
	if n > len(h.nodes) {
		n = len(h.nodes)
	}

	type nodeWeight struct {
		node   string
		weight uint64
	}

	weights := make([]nodeWeight, len(h.nodes))
	for i, node := range h.nodes {
		weights[i] = nodeWeight{node: node, weight: weight(key, node)}
	}

	sort.Slice(weights, func(i, j int) bool {
		return weights[i].weight > weights[j].weight
	})

	result := make([]string, n)
	for i := 0; i < n; i++ {
		result[i] = weights[i].node
	}
	return result
}

// Add 添加一个节点到哈希环中。
func (h *RendezvousHash) Add(node string) {
	for _, n := range h.nodes {
		if n == node {
			return
		}
	}
	h.nodes = append(h.nodes, node)
}

// Remove 从哈希环中移除一个节点。
func (h *RendezvousHash) Remove(node string) {
	for i, n := range h.nodes {
		if n == node {
			h.nodes = append(h.nodes[:i], h.nodes[i+1:]...)
			return
		}
	}
}

// Len 返回当前节点数量。
func (h *RendezvousHash) Len() int {
	return len(h.nodes)
}

// weight 计算键和节点组合的哈希权重。
func weight(key, node string) uint64 {
	hasher := fnv.New64a()
	hasher.Write([]byte(key))
	hasher.Write([]byte(node))
	return hasher.Sum64()
}
