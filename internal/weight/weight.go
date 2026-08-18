// Package weight 实现加权一致性哈希。
// 通过为不同节点分配不同的权重，实现按比例分配负载的功能。
package weight

import (
	"hash/crc32"
	"sort"
)

// WeightedNode 表示一个带权重的节点。
type WeightedNode struct {
	// Name 是节点名称
	Name string
	// Weight 是节点权重，权重越大分配的负载越多
	Weight int
}

// WeightedRing 是加权一致性哈希环。
type WeightedRing struct {
	nodes   []WeightedNode
	circle  []entry
	totalWt int
}

// entry 是哈希环上的一个条目。
type entry struct {
	hash uint32
	node string
}

// NewWeighted 创建一个新的加权一致性哈希环。
// 每个节点根据权重获得相应比例的虚拟节点。
func NewWeighted(nodes []WeightedNode) *WeightedRing {
	r := &WeightedRing{
		nodes: make([]WeightedNode, len(nodes)),
	}
	copy(r.nodes, nodes)
	r.build()
	return r
}

// Get 根据键获取对应的节点名称。
func (r *WeightedRing) Get(key string) string {
	if len(r.circle) == 0 {
		return ""
	}
	h := crc32.ChecksumIEEE([]byte(key))
	idx := sort.Search(len(r.circle), func(i int) bool {
		return r.circle[i].hash >= h
	})
	if idx >= len(r.circle) {
		idx = 0
	}
	return r.circle[idx].node
}

// Distribution 统计一组键在各节点上的分布情况。
func (r *WeightedRing) Distribution(keys []string) map[string]int {
	dist := make(map[string]int)
	for _, key := range keys {
		node := r.Get(key)
		if node != "" {
			dist[node]++
		}
	}
	return dist
}

// TotalWeight 返回所有节点的总权重。
func (r *WeightedRing) TotalWeight() int {
	return r.totalWt
}

// Len 返回节点数量。
func (r *WeightedRing) Len() int {
	return len(r.nodes)
}

// build 根据权重构建哈希环。
func (r *WeightedRing) build() {
	r.totalWt = 0
	for _, n := range r.nodes {
		r.totalWt += n.Weight
	}
	if r.totalWt == 0 {
		return
	}

	// 基础副本数为 50，按权重比例分配虚拟节点
	baseReplicas := 50
	r.circle = make([]entry, 0)
	for _, node := range r.nodes {
		replicas := baseReplicas * node.Weight
		for i := 0; i < replicas; i++ {
			key := node.Name + "#" + itoa(i)
			h := crc32.ChecksumIEEE([]byte(key))
			r.circle = append(r.circle, entry{hash: h, node: node.Name})
		}
	}

	sort.Slice(r.circle, func(i, j int) bool {
		return r.circle[i].hash < r.circle[j].hash
	})
}

// itoa 简单的整数转字符串。
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	buf := make([]byte, 0, 10)
	for n > 0 {
		buf = append(buf, byte('0'+n%10))
		n /= 10
	}
	// 反转
	for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}
	return string(buf)
}
