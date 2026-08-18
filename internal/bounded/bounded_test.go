package bounded

import (
	"fmt"
	"testing"
)

func TestGet(t *testing.T) {
	nodes := []string{"node-1", "node-2", "node-3"}
	b := NewBounded(nodes, 1.25)

	result := b.Get("key-1")
	if result == "" {
		t.Error("expected a node, got empty string")
	}

	valid := false
	for _, n := range nodes {
		if result == n {
			valid = true
			break
		}
	}
	if !valid {
		t.Errorf("Get returned invalid node: %s", result)
	}

	empty := NewBounded(nil, 1.25)
	if empty.Get("key") != "" {
		t.Error("expected empty string for empty hash")
	}
}

func TestBoundedLoad(t *testing.T) {
	nodes := []string{"a", "b", "c"}
	b := NewBounded(nodes, 1.5)

	numKeys := 300
	for i := 0; i < numKeys; i++ {
		key := fmt.Sprintf("request-%d", i)
		b.Get(key)
	}

	if b.TotalLoad() != numKeys {
		t.Errorf("expected total load %d, got %d", numKeys, b.TotalLoad())
	}

	maxAllowed := b.MaxLoad()
	for _, node := range nodes {
		load := b.Load(node)
		if load > maxAllowed {
			t.Errorf("node %s load %d exceeds max %d", node, load, maxAllowed)
		}
	}

	avgLoad := numKeys / len(nodes)
	for _, node := range nodes {
		load := b.Load(node)
		if load < avgLoad/3 {
			t.Errorf("node %s has unexpectedly low load: %d (avg: %d)", node, load, avgLoad)
		}
	}
}

func TestReset(t *testing.T) {
	nodes := []string{"x", "y"}
	b := NewBounded(nodes, 1.25)

	for i := 0; i < 50; i++ {
		b.Get(fmt.Sprintf("k%d", i))
	}
	if b.TotalLoad() != 50 {
		t.Errorf("expected 50, got %d", b.TotalLoad())
	}

	b.Reset()
	if b.TotalLoad() != 0 {
		t.Errorf("expected 0 after reset, got %d", b.TotalLoad())
	}
	for _, n := range nodes {
		if b.Load(n) != 0 {
			t.Errorf("expected load 0 for %s after reset", n)
		}
	}
}

func TestMaxLoad(t *testing.T) {
	nodes := []string{"s1", "s2", "s3", "s4"}
	b := NewBounded(nodes, 2.0)

	if b.MaxLoad() < 1 {
		t.Errorf("expected MaxLoad >= 1, got %d", b.MaxLoad())
	}
}
