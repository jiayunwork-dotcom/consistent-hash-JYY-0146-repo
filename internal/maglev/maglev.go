// Package maglev 实现 Maglev 一致性哈希算法。
// Maglev 是 Google 开发的负载均衡算法，具有查找速度快和负载均衡性好的特点。
package maglev

import (
	"hash/crc32"
	"hash/fnv"
)

// Table 是 Maglev 哈希查找表。
type Table struct {
	backends  []string
	lookup    []int
	tableSize int
}

// NewTable 创建一个新的 Maglev 哈希表。
// backends 是后端节点列表，tableSize 应该是一个素数以获得最佳分布。
func NewTable(backends []string, tableSize int) *Table {
	if len(backends) == 0 || tableSize <= 0 {
		return &Table{
			backends:  backends,
			lookup:    nil,
			tableSize: tableSize,
		}
	}

	t := &Table{
		backends:  make([]string, len(backends)),
		lookup:    make([]int, tableSize),
		tableSize: tableSize,
	}
	copy(t.backends, backends)
	t.populate()
	return t
}

// Lookup 根据键查找对应的后端节点名称。
func (t *Table) Lookup(key string) string {
	if len(t.backends) == 0 || len(t.lookup) == 0 {
		return ""
	}
	idx := t.LookupIndex(key)
	if idx < 0 {
		return ""
	}
	return t.backends[idx]
}

// LookupIndex 根据键查找对应的后端索引。
func (t *Table) LookupIndex(key string) int {
	if len(t.lookup) == 0 {
		return -1
	}
	h := crc32.ChecksumIEEE([]byte(key))
	pos := int(h) % t.tableSize
	if pos < 0 {
		pos += t.tableSize
	}
	return t.lookup[pos]
}

// Size 返回查找表的大小。
func (t *Table) Size() int {
	return t.tableSize
}

// populate 填充 Maglev 查找表。
func (t *Table) populate() {
	n := len(t.backends)
	m := t.tableSize

	// 初始化查找表为 -1（未分配）
	for i := range t.lookup {
		t.lookup[i] = -1
	}

	// 为每个后端计算 offset 和 skip
	permutation := make([][]int, n)
	for i, backend := range t.backends {
		offset := hashOffset(backend, m)
		skip := hashSkip(backend, m)
		perm := make([]int, m)
		for j := 0; j < m; j++ {
			perm[j] = (offset + j*skip) % m
		}
		permutation[i] = perm
	}

	// 填充查找表
	next := make([]int, n)
	filled := 0
	for filled < m {
		for i := 0; i < n; i++ {
			c := next[i]
			for t.lookup[permutation[i][c]] >= 0 {
				c++
				if c >= m {
					break
				}
			}
			if c >= m {
				continue
			}
			t.lookup[permutation[i][c]] = i
			next[i] = c + 1
			filled++
			if filled >= m {
				break
			}
		}
	}
}

// hashOffset 计算后端的偏移量。
func hashOffset(backend string, tableSize int) int {
	h := fnv.New32a()
	h.Write([]byte(backend))
	return int(h.Sum32()) % tableSize
}

// hashSkip 计算后端的跳步值。
func hashSkip(backend string, tableSize int) int {
	h := fnv.New32()
	h.Write([]byte(backend))
	skip := int(h.Sum32())%(tableSize-1) + 1
	return skip
}
