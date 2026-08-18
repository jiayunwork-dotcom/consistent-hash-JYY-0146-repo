package weight

import (
	"fmt"
	"math"
	"testing"
)

func TestGet(t *testing.T) {
	nodes := []WeightedNode{
		{Name: "big", Weight: 3},
		{Name: "medium", Weight: 2},
		{Name: "small", Weight: 1},
	}
	ring := NewWeighted(nodes)

	key := "test-item"
	r1 := ring.Get(key)
	r2 := ring.Get(key)
	if r1 != r2 {
		t.Errorf("Get is not deterministic: %s != %s", r1, r2)
	}

	valid := false
	for _, n := range nodes {
		if r1 == n.Name {
			valid = true
			break
		}
	}
	if !valid {
		t.Errorf("Get returned invalid node: %s", r1)
	}

	empty := NewWeighted(nil)
	if empty.Get("key") != "" {
		t.Error("expected empty string for empty ring")
	}
}

func TestDistribution(t *testing.T) {
	nodes := []WeightedNode{
		{Name: "heavy", Weight: 4},
		{Name: "light", Weight: 1},
	}
	ring := NewWeighted(nodes)

	keys := make([]string, 10000)
	for i := range keys {
		keys[i] = fmt.Sprintf("item-%d", i)
	}

	dist := ring.Distribution(keys)

	heavyCount := dist["heavy"]
	lightCount := dist["light"]

	heavyRatio := float64(heavyCount) / float64(len(keys))
	expectedRatio := 4.0 / 5.0

	if math.Abs(heavyRatio-expectedRatio) > 0.15 {
		t.Errorf("heavy node ratio %.2f, expected ~%.2f (heavy=%d, light=%d)",
			heavyRatio, expectedRatio, heavyCount, lightCount)
	}
}

func TestTotalWeight(t *testing.T) {
	nodes := []WeightedNode{
		{Name: "a", Weight: 3},
		{Name: "b", Weight: 5},
		{Name: "c", Weight: 2},
	}
	ring := NewWeighted(nodes)

	if ring.TotalWeight() != 10 {
		t.Errorf("expected total weight 10, got %d", ring.TotalWeight())
	}
}

func TestLen(t *testing.T) {
	nodes := []WeightedNode{
		{Name: "x", Weight: 1},
		{Name: "y", Weight: 2},
	}
	ring := NewWeighted(nodes)
	if ring.Len() != 2 {
		t.Errorf("expected Len() = 2, got %d", ring.Len())
	}
}
