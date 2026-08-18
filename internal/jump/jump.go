// Package jump 实现 Jump Consistent Hash 算法。
// Jump Consistent Hash 是一种快速、零内存消耗的一致性哈希算法，
// 由 Google 的 John Lamping 和 Eric Veach 提出。
package jump

import "hash/crc32"

// Hash 使用 Jump Consistent Hash 算法将 key 映射到 [0, numBuckets) 范围的桶。
// 算法特点：O(ln(n)) 时间复杂度，零内存开销，完美的均匀分布。
func Hash(key uint64, numBuckets int) int {
	if numBuckets <= 0 {
		return 0
	}
	var b, j int64
	b = -1
	j = 0
	for j < int64(numBuckets) {
		b = j
		key = key*2862933555777941757 + 1
		j = int64(float64(b+1) * (float64(int64(1)<<31) / float64((key>>33)+1)))
	}
	return int(b)
}

// HashString 将字符串键通过 CRC32 转换后使用 Jump Hash 算法。
func HashString(key string, numBuckets int) int {
	h := crc32.ChecksumIEEE([]byte(key))
	return Hash(uint64(h), numBuckets)
}

// Distribution 统计一组键在各桶中的分布情况。
// 返回一个长度为 numBuckets 的切片，每个元素表示对应桶中的键数量。
func Distribution(keys []string, numBuckets int) []int {
	if numBuckets <= 0 {
		return nil
	}
	counts := make([]int, numBuckets)
	for _, key := range keys {
		bucket := HashString(key, numBuckets)
		counts[bucket]++
	}
	return counts
}

// Monotonic 验证 Jump Hash 的单调性：当桶数增加时，
// 键只会从旧桶移到新桶，不会在旧桶之间移动。
func Monotonic(key uint64, oldBuckets, newBuckets int) bool {
	if newBuckets <= oldBuckets {
		return true
	}
	oldBucket := Hash(key, oldBuckets)
	newBucket := Hash(key, newBuckets)
	// 键要么留在原桶，要么移动到新增的桶中
	return newBucket == oldBucket || newBucket >= oldBuckets
}
