// Package metrics 提供分布质量的统计度量功能。
// 用于评估一致性哈希算法的负载分布均匀程度。
package metrics

import (
	"math"
	"sort"
)

// StandardDeviation 计算计数列表的标准差。
// 标准差越小，表示分布越均匀。
func StandardDeviation(counts []int) float64 {
	if len(counts) == 0 {
		return 0
	}
	mean := average(counts)
	var sumSquares float64
	for _, c := range counts {
		diff := float64(c) - mean
		sumSquares += diff * diff
	}
	return math.Sqrt(sumSquares / float64(len(counts)))
}

// CoeffOfVariation 计算变异系数（标准差/均值）。
// 变异系数是一个无量纲数，用于比较不同规模数据的离散程度。
// 值越接近0表示分布越均匀。
func CoeffOfVariation(counts []int) float64 {
	if len(counts) == 0 {
		return 0
	}
	mean := average(counts)
	if mean == 0 {
		return 0
	}
	sd := StandardDeviation(counts)
	return sd / mean
}

// MaxMinRatio 计算最大值与最小值的比率。
// 比率越接近1表示分布越均匀。
func MaxMinRatio(counts []int) float64 {
	if len(counts) == 0 {
		return 0
	}
	minVal := counts[0]
	maxVal := counts[0]
	for _, c := range counts[1:] {
		if c < minVal {
			minVal = c
		}
		if c > maxVal {
			maxVal = c
		}
	}
	if minVal == 0 {
		if maxVal == 0 {
			return 1
		}
		return math.Inf(1)
	}
	return float64(maxVal) / float64(minVal)
}

// EntropyBalance 计算基于信息熵的均衡度。
// 返回值范围为 [0, 1]，1 表示完全均匀分布。
func EntropyBalance(counts []int) float64 {
	if len(counts) == 0 {
		return 0
	}
	total := 0
	for _, c := range counts {
		total += c
	}
	if total == 0 {
		return 0
	}

	var entropy float64
	for _, c := range counts {
		if c > 0 {
			p := float64(c) / float64(total)
			entropy -= p * math.Log2(p)
		}
	}

	// 最大熵（完全均匀分布时）
	maxEntropy := math.Log2(float64(len(counts)))
	if maxEntropy == 0 {
		return 1
	}
	return entropy / maxEntropy
}

// Percentile 计算计数列表的第 p 百分位值。
// p 的范围是 [0, 100]。
func Percentile(counts []int, p float64) float64 {
	if len(counts) == 0 {
		return 0
	}
	if p < 0 {
		p = 0
	}
	if p > 100 {
		p = 100
	}

	sorted := make([]int, len(counts))
	copy(sorted, counts)
	sort.Ints(sorted)

	// 使用线性插值计算百分位
	rank := (p / 100) * float64(len(sorted)-1)
	lower := int(math.Floor(rank))
	upper := int(math.Ceil(rank))

	if lower == upper {
		return float64(sorted[lower])
	}

	fraction := rank - float64(lower)
	return float64(sorted[lower]) + fraction*(float64(sorted[upper])-float64(sorted[lower]))
}

// Gini 计算基尼系数，衡量分布的不平等程度。
// 返回值范围 [0, 1]，0 表示完全平等，1 表示完全不平等。
func Gini(counts []int) float64 {
	n := len(counts)
	if n == 0 {
		return 0
	}

	sorted := make([]int, n)
	copy(sorted, counts)
	sort.Ints(sorted)

	var sumNumerator float64
	var sumDenominator float64
	for i, c := range sorted {
		sumNumerator += float64(2*(i+1)-n-1) * float64(c)
		sumDenominator += float64(c)
	}

	if sumDenominator == 0 {
		return 0
	}
	return sumNumerator / (float64(n) * sumDenominator)
}

// average 计算整数切片的平均值。
func average(counts []int) float64 {
	if len(counts) == 0 {
		return 0
	}
	sum := 0
	for _, c := range counts {
		sum += c
	}
	return float64(sum) / float64(len(counts))
}
