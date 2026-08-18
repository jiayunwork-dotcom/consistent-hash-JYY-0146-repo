// Package replica 提供数据复制策略的实现。
// 支持简单复制和机架感知复制策略。
package replica

import (
	"hash/crc32"
	"sort"
)

// Strategy 定义复制策略的配置。
type Strategy struct {
	// Factor 是复制因子，表示每个键需要存储的副本数
	Factor int
	// RackAware 表示是否启用机架感知复制
	RackAware bool
}

// Replicas 根据策略返回指定键的副本节点列表。
// primary 是主节点，allNodes 是所有可用节点，key 用于确定性地选择副本。
func (s Strategy) Replicas(primary string, allNodes []string, key string) []string {
	if s.Factor <= 0 || len(allNodes) == 0 {
		return nil
	}

	result := []string{primary}
	if s.Factor == 1 {
		return result
	}

	// 创建候选节点列表（排除主节点）
	candidates := make([]string, 0, len(allNodes)-1)
	for _, node := range allNodes {
		if node != primary {
			candidates = append(candidates, node)
		}
	}

	if len(candidates) == 0 {
		return result
	}

	// 使用键的哈希来确定性地排列候选节点
	type scored struct {
		node  string
		score uint32
	}
	scores := make([]scored, len(candidates))
	for i, node := range candidates {
		h := crc32.ChecksumIEEE([]byte(key + "|" + node))
		scores[i] = scored{node: node, score: h}
	}
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score < scores[j].score
	})

	// 选择前 Factor-1 个副本
	needed := s.Factor - 1
	if needed > len(scores) {
		needed = len(scores)
	}
	for i := 0; i < needed; i++ {
		result = append(result, scores[i].node)
	}
	return result
}

// PreferenceList 返回一个键的偏好列表，包含 n 个节点。
// ring 是实现了 Get 方法的哈希环接口。
// 通过在键后追加不同后缀来获取多个不同的节点。
func PreferenceList(ring interface{ Get(string) string }, key string, n int) []string {
	if n <= 0 {
		return nil
	}

	seen := make(map[string]bool)
	var result []string

	// 首先尝试原始键
	primary := ring.Get(key)
	if primary != "" {
		seen[primary] = true
		result = append(result, primary)
	}

	// 通过追加后缀获取更多节点
	suffix := 0
	maxAttempts := n * 10 // 防止无限循环
	for len(result) < n && suffix < maxAttempts {
		candidate := ring.Get(key + "#" + intToStr(suffix))
		if candidate != "" && !seen[candidate] {
			seen[candidate] = true
			result = append(result, candidate)
		}
		suffix++
	}
	return result
}

// intToStr 整数转字符串的辅助函数。
func intToStr(n int) string {
	if n == 0 {
		return "0"
	}
	buf := make([]byte, 0, 10)
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		buf = append(buf, byte('0'+n%10))
		n /= 10
	}
	if neg {
		buf = append(buf, '-')
	}
	for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}
	return string(buf)
}
