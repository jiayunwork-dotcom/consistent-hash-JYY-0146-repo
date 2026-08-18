package rendezvous

import "testing"

func TestGet(t *testing.T) {
	nodes := []string{"server1", "server2", "server3", "server4"}
	h := New(nodes)

	key := "test-key"
	result1 := h.Get(key)
	result2 := h.Get(key)
	if result1 != result2 {
		t.Errorf("Get is not deterministic: %s != %s", result1, result2)
	}

	found := false
	for _, n := range nodes {
		if result1 == n {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Get returned invalid node: %s", result1)
	}

	empty := New(nil)
	if empty.Get("key") != "" {
		t.Error("expected empty string for empty hash")
	}
}

func TestGetN(t *testing.T) {
	nodes := []string{"a", "b", "c", "d", "e"}
	h := New(nodes)

	result := h.GetN("my-key", 3)
	if len(result) != 3 {
		t.Errorf("expected 3 nodes, got %d", len(result))
	}

	seen := make(map[string]bool)
	for _, n := range result {
		if seen[n] {
			t.Errorf("duplicate node in result: %s", n)
		}
		seen[n] = true
	}

	all := h.GetN("key", 10)
	if len(all) != len(nodes) {
		t.Errorf("expected %d nodes (capped), got %d", len(nodes), len(all))
	}

	none := h.GetN("key", 0)
	if none != nil {
		t.Error("expected nil for n=0")
	}
}

func TestAddRemove(t *testing.T) {
	h := New([]string{"node1", "node2"})

	h.Add("node3")
	if h.Len() != 3 {
		t.Errorf("expected 3 nodes after add, got %d", h.Len())
	}

	h.Add("node3")
	if h.Len() != 3 {
		t.Errorf("expected 3 nodes after duplicate add, got %d", h.Len())
	}

	h.Remove("node2")
	if h.Len() != 2 {
		t.Errorf("expected 2 nodes after remove, got %d", h.Len())
	}

	h.Remove("nonexistent")
	if h.Len() != 2 {
		t.Errorf("expected 2 nodes after removing nonexistent, got %d", h.Len())
	}
}

func TestMinimalDisruption(t *testing.T) {
	nodes := []string{"s1", "s2", "s3", "s4"}
	h := New(nodes)

	keys := make([]string, 100)
	for i := range keys {
		keys[i] = string(rune('a'+i/26)) + string(rune('a'+i%26))
	}

	original := make(map[string]string)
	for _, k := range keys {
		original[k] = h.Get(k)
	}

	h.Add("s5")

	changed := 0
	for _, k := range keys {
		if h.Get(k) != original[k] {
			changed++
		}
	}

	maxExpected := len(keys) * 40 / 100
	if changed > maxExpected {
		t.Errorf("too many keys changed: %d out of %d", changed, len(keys))
	}
}
