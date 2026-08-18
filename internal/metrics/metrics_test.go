package metrics

import (
	"math"
	"testing"
)

func TestStandardDeviation(t *testing.T) {
	uniform := []int{100, 100, 100, 100}
	sd := StandardDeviation(uniform)
	if sd != 0 {
		t.Errorf("expected SD=0 for uniform distribution, got %f", sd)
	}

	skewed := []int{10, 20, 30, 40}
	sd = StandardDeviation(skewed)
	if sd <= 0 {
		t.Errorf("expected positive SD for skewed distribution, got %f", sd)
	}

	if StandardDeviation(nil) != 0 {
		t.Error("expected 0 for nil slice")
	}
}

func TestCoeffOfVariation(t *testing.T) {
	uniform := []int{50, 50, 50}
	cv := CoeffOfVariation(uniform)
	if cv != 0 {
		t.Errorf("expected CV=0 for uniform, got %f", cv)
	}

	skewed := []int{10, 50, 90}
	cv = CoeffOfVariation(skewed)
	if cv <= 0 {
		t.Errorf("expected positive CV, got %f", cv)
	}

	zeros := []int{0, 0, 0}
	if CoeffOfVariation(zeros) != 0 {
		t.Error("expected 0 for all-zero counts")
	}
}

func TestMaxMinRatio(t *testing.T) {
	uniform := []int{10, 10, 10}
	ratio := MaxMinRatio(uniform)
	if ratio != 1.0 {
		t.Errorf("expected ratio=1 for uniform, got %f", ratio)
	}

	skewed := []int{5, 15}
	ratio = MaxMinRatio(skewed)
	if ratio != 3.0 {
		t.Errorf("expected ratio=3, got %f", ratio)
	}

	withZero := []int{0, 10}
	ratio = MaxMinRatio(withZero)
	if !math.IsInf(ratio, 1) {
		t.Errorf("expected +Inf for min=0, got %f", ratio)
	}

	allZero := []int{0, 0}
	if MaxMinRatio(allZero) != 1 {
		t.Errorf("expected 1 for all zeros, got %f", MaxMinRatio(allZero))
	}
}

func TestEntropyBalance(t *testing.T) {
	uniform := []int{25, 25, 25, 25}
	balance := EntropyBalance(uniform)
	if math.Abs(balance-1.0) > 0.001 {
		t.Errorf("expected balance~1 for uniform, got %f", balance)
	}

	skewed := []int{100, 0, 0, 0}
	balance = EntropyBalance(skewed)
	if balance != 0 {
		t.Errorf("expected balance=0 for single-node distribution, got %f", balance)
	}

	if EntropyBalance(nil) != 0 {
		t.Error("expected 0 for nil")
	}
	if EntropyBalance([]int{0, 0}) != 0 {
		t.Error("expected 0 for all zeros")
	}
}

func TestPercentile(t *testing.T) {
	counts := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	p50 := Percentile(counts, 50)
	if p50 != 5.5 {
		t.Errorf("expected P50=5.5, got %f", p50)
	}

	p0 := Percentile(counts, 0)
	if p0 != 1 {
		t.Errorf("expected P0=1, got %f", p0)
	}

	p100 := Percentile(counts, 100)
	if p100 != 10 {
		t.Errorf("expected P100=10, got %f", p100)
	}

	if Percentile(nil, 50) != 0 {
		t.Error("expected 0 for nil slice")
	}
}

func TestGini(t *testing.T) {
	uniform := []int{10, 10, 10, 10}
	g := Gini(uniform)
	if math.Abs(g) > 0.001 {
		t.Errorf("expected Gini~0 for uniform, got %f", g)
	}

	skewed := []int{1, 1, 1, 100}
	g = Gini(skewed)
	if g <= 0 {
		t.Errorf("expected positive Gini for skewed distribution, got %f", g)
	}

	if Gini(nil) != 0 {
		t.Error("expected 0 for nil")
	}
}
