package maglev

import (
	"fmt"
	"testing"
)

func TestNewTableAndLookup(t *testing.T) {
	backends := []string{"backend-1", "backend-2", "backend-3"}
	table := NewTable(backends, 65537)

	key := "request-123"
	result1 := table.Lookup(key)
	result2 := table.Lookup(key)
	if result1 != result2 {
		t.Errorf("Lookup is not deterministic: %s != %s", result1, result2)
	}

	found := false
	for _, b := range backends {
		if result1 == b {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Lookup returned invalid backend: %s", result1)
	}

	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		k := fmt.Sprintf("key-%d", i)
		seen[table.Lookup(k)] = true
	}
	if len(seen) < 2 {
		t.Error("expected keys to be distributed across multiple backends")
	}
}

func TestLookupIndex(t *testing.T) {
	backends := []string{"a", "b", "c", "d"}
	table := NewTable(backends, 101)

	for i := 0; i < 50; i++ {
		key := fmt.Sprintf("item-%d", i)
		idx := table.LookupIndex(key)
		if idx < 0 || idx >= len(backends) {
			t.Errorf("LookupIndex(%q) = %d, out of range [0, %d)", key, idx, len(backends))
		}
	}

	empty := NewTable(nil, 101)
	if empty.Lookup("key") != "" {
		t.Error("expected empty string for nil backends")
	}
	if empty.LookupIndex("key") != -1 {
		t.Error("expected -1 for nil backends")
	}
}

func TestDistribution(t *testing.T) {
	backends := []string{"srv1", "srv2", "srv3", "srv4", "srv5"}
	table := NewTable(backends, 1009)

	counts := make(map[string]int)
	numKeys := 5000
	for i := 0; i < numKeys; i++ {
		key := fmt.Sprintf("user-%d", i)
		backend := table.Lookup(key)
		counts[backend]++
	}

	for _, b := range backends {
		if counts[b] == 0 {
			t.Errorf("backend %s got zero keys", b)
		}
	}

	threshold := numKeys / (len(backends) * 2)
	for b, c := range counts {
		if c < threshold {
			t.Errorf("backend %s got too few keys: %d (threshold: %d)", b, c, threshold)
		}
	}
}

func TestSize(t *testing.T) {
	table := NewTable([]string{"x", "y"}, 997)
	if table.Size() != 997 {
		t.Errorf("expected size 997, got %d", table.Size())
	}
}
