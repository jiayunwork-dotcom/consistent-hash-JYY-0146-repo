package ketama

import (
	"fmt"
	"testing"
)

func TestGet(t *testing.T) {
	nodes := []string{"cache-1", "cache-2", "cache-3"}
	ring := NewKetama(nodes, 150)

	key := "session-abc"
	r1 := ring.Get(key)
	r2 := ring.Get(key)
	if r1 != r2 {
		t.Errorf("Get is not deterministic: %s != %s", r1, r2)
	}

	valid := false
	for _, n := range nodes {
		if r1 == n {
			valid = true
			break
		}
	}
	if !valid {
		t.Errorf("Get returned invalid node: %s", r1)
	}

	empty := NewKetama(nil, 150)
	if empty.Get("key") != "" {
		t.Error("expected empty string for empty ring")
	}
}

func TestGetWithFallback(t *testing.T) {
	nodes := []string{"n1", "n2", "n3", "n4", "n5"}
	ring := NewKetama(nodes, 100)

	result := ring.GetWithFallback("my-key", 3)
	if len(result) != 3 {
		t.Errorf("expected 3 nodes, got %d", len(result))
	}

	seen := make(map[string]bool)
	for _, n := range result {
		if seen[n] {
			t.Errorf("duplicate node: %s", n)
		}
		seen[n] = true
	}

	primary := ring.Get("my-key")
	if result[0] != primary {
		t.Errorf("first fallback node %s != primary %s", result[0], primary)
	}

	all := ring.GetWithFallback("key", 10)
	if len(all) != len(nodes) {
		t.Errorf("expected %d nodes (capped), got %d", len(nodes), len(all))
	}

	none := ring.GetWithFallback("key", 0)
	if none != nil {
		t.Error("expected nil for n=0")
	}
}

func TestDistribution(t *testing.T) {
	nodes := []string{"server-a", "server-b", "server-c"}
	ring := NewKetama(nodes, 150)

	counts := make(map[string]int)
	numKeys := 3000
	for i := 0; i < numKeys; i++ {
		key := fmt.Sprintf("object-%d", i)
		node := ring.Get(key)
		counts[node]++
	}

	for _, n := range nodes {
		if counts[n] == 0 {
			t.Errorf("node %s has zero keys", n)
		}
	}

	avg := numKeys / len(nodes)
	for n, c := range counts {
		if c < avg/3 {
			t.Errorf("node %s has too few keys: %d (avg: %d)", n, c, avg)
		}
	}
}

func TestLen(t *testing.T) {
	ring := NewKetama([]string{"a", "b", "c"}, 50)
	if ring.Len() != 3 {
		t.Errorf("expected Len() = 3, got %d", ring.Len())
	}
}
