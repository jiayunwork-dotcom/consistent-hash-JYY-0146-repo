// Package ketama 实现 Ketama 兼容的一致性哈希算法。
// Ketama 是 memcached 客户端广泛使用的一致性哈希方案。
package ketama

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"sort"
)

// Ring 是 Ketama 一致性哈希环。
type Ring struct {
	nodes    []string
	replicas int
	circle   []point
}

// point 表示哈希环上的一个点。
type point struct {
	hash uint32
	node string
}

// NewKetama 创建一个新的 Ketama 哈希环。
// nodes 是节点列表，replicas 是每个节点的虚拟节点数量。
func NewKetama(nodes []string, replicas int) *Ring {
	r := &Ring{
		nodes:    make([]string, len(nodes)),
		replicas: replicas,
	}
	copy(r.nodes, nodes)
	r.build()
	return r
}

// Get 根据键获取对应的节点。
func (r *Ring) Get(key string) string {
	if len(r.circle) == 0 {
		return ""
	}
	h := ketamaHash(key)
	idx := sort.Search(len(r.circle), func(i int) bool {
		return r.circle[i].hash >= h
	})
	if idx >= len(r.circle) {
		idx = 0
	}
	return r.circle[idx].node
}

// GetWithFallback 返回键对应的主节点和 n-1 个备用节点。
// 返回的节点列表按照环上的顺序排列，且不重复。
func (r *Ring) GetWithFallback(key string, n int) []string {
	if len(r.circle) == 0 || n <= 0 {
		return nil
	}
	if n > len(r.nodes) {
		n = len(r.nodes)
	}

	h := ketamaHash(key)
	idx := sort.Search(len(r.circle), func(i int) bool {
		return r.circle[i].hash >= h
	})
	if idx >= len(r.circle) {
		idx = 0
	}

	result := make([]string, 0, n)
	seen := make(map[string]bool)
	for i := 0; i < len(r.circle) && len(result) < n; i++ {
		pos := (idx + i) % len(r.circle)
		node := r.circle[pos].node
		if !seen[node] {
			seen[node] = true
			result = append(result, node)
		}
	}
	return result
}

// Len 返回环上的节点数量。
func (r *Ring) Len() int {
	return len(r.nodes)
}

// build 构建哈希环。
func (r *Ring) build() {
	r.circle = make([]point, 0, len(r.nodes)*r.replicas)
	for _, node := range r.nodes {
		for i := 0; i < r.replicas; i++ {
			vkey := fmt.Sprintf("%s-%d", node, i)
			h := ketamaHash(vkey)
			r.circle = append(r.circle, point{hash: h, node: node})
		}
	}
	sort.Slice(r.circle, func(i, j int) bool {
		return r.circle[i].hash < r.circle[j].hash
	})
}

// ketamaHash 使用 MD5 计算 Ketama 兼容的哈希值。
func ketamaHash(key string) uint32 {
	digest := md5.Sum([]byte(key))
	return binary.LittleEndian.Uint32(digest[0:4])
}
