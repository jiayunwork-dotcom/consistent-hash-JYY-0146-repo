// Package bounded 实现有界负载的一致性哈希。
// 该算法确保每个节点的负载不超过平均负载乘以负载因子，
// 从而避免热点问题。参考 Google 论文 "Consistent Hashing with Bounded Loads"。
package bounded

import (
	"hash/crc32"
	"math"
	"sort"
	"sync"
)

// BoundedHash 是有界负载一致性哈希结构。
type BoundedHash struct {
	mu         sync.RWMutex
	nodes      []string
	loadFactor float64
	loads      map[string]int
	totalLoad  int
	circle     []point
}

type point struct {
	hash uint32
	node string
}

// NewBounded 创建一个新的有界负载一致性哈希。
// loadFactor 是负载因子，通常设为 1.25，表示允许的最大负载为平均负载的 1.25 倍。
func NewBounded(nodes []string, loadFactor float64) *BoundedHash {
	if loadFactor < 1.0 {
		loadFactor = 1.0
	}
	b := &BoundedHash{
		nodes:      make([]string, len(nodes)),
		loadFactor: loadFactor,
		loads:      make(map[string]int),
	}
	copy(b.nodes, nodes)
	b.buildCircle()
	return b
}

// Get 根据键获取对应的节点，同时保证不超过负载上限。
// 如果首选节点已达到负载上限，将按环上顺序查找下一个可用节点。
func (b *BoundedHash) Get(key string) string {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.circle) == 0 {
		return ""
	}

	h := crc32.ChecksumIEEE([]byte(key))
	idx := sort.Search(len(b.circle), func(i int) bool {
		return b.circle[i].hash >= h
	})
	if idx >= len(b.circle) {
		idx = 0
	}

	maxLoad := b.maxLoadValue()

	// 从首选节点开始，查找未超载的节点
	for i := 0; i < len(b.circle); i++ {
		pos := (idx + i) % len(b.circle)
		node := b.circle[pos].node
		if b.loads[node] < maxLoad {
			b.loads[node]++
			b.totalLoad++
			return node
		}
	}

	// 所有节点都已满载（理论上不应发生），返回首选节点
	node := b.circle[idx].node
	b.loads[node]++
	b.totalLoad++
	return node
}

// Load 返回指定节点的当前负载。
func (b *BoundedHash) Load(node string) int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.loads[node]
}

// MaxLoad 返回允许的最大负载值。
func (b *BoundedHash) MaxLoad() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.maxLoadValue()
}

// Reset 重置所有节点的负载计数。
func (b *BoundedHash) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.loads = make(map[string]int)
	b.totalLoad = 0
}

// TotalLoad 返回所有节点的总负载。
func (b *BoundedHash) TotalLoad() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.totalLoad
}

// maxLoadValue 计算当前允许的最大负载值。
func (b *BoundedHash) maxLoadValue() int {
	if len(b.nodes) == 0 {
		return 0
	}
	avgLoad := float64(b.totalLoad+1) / float64(len(b.nodes))
	maxLoad := math.Ceil(avgLoad * b.loadFactor)
	if maxLoad < 1 {
		maxLoad = 1
	}
	return int(maxLoad)
}

// buildCircle 构建哈希环。
func (b *BoundedHash) buildCircle() {
	replicas := 100
	b.circle = make([]point, 0, len(b.nodes)*replicas)
	for _, node := range b.nodes {
		for i := 0; i < replicas; i++ {
			key := node + "#" + intToStr(i)
			h := crc32.ChecksumIEEE([]byte(key))
			b.circle = append(b.circle, point{hash: h, node: node})
		}
	}
	sort.Slice(b.circle, func(i, j int) bool {
		return b.circle[i].hash < b.circle[j].hash
	})
}

func intToStr(n int) string {
	if n == 0 {
		return "0"
	}
	buf := make([]byte, 0, 10)
	for n > 0 {
		buf = append(buf, byte('0'+n%10))
		n /= 10
	}
	for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}
	return string(buf)
}
