package replica

import "testing"

func TestReplicas(t *testing.T) {
	allNodes := []string{"node1", "node2", "node3", "node4", "node5"}
	strategy := Strategy{Factor: 3, RackAware: false}

	replicas := strategy.Replicas("node1", allNodes, "my-key")
	if len(replicas) != 3 {
		t.Errorf("expected 3 replicas, got %d", len(replicas))
	}

	if replicas[0] != "node1" {
		t.Errorf("first replica should be primary, got %s", replicas[0])
	}

	seen := make(map[string]bool)
	for _, r := range replicas {
		if seen[r] {
			t.Errorf("duplicate replica: %s", r)
		}
		seen[r] = true
	}

	replicas2 := strategy.Replicas("node1", allNodes, "my-key")
	for i, r := range replicas {
		if r != replicas2[i] {
			t.Errorf("not deterministic at index %d: %s != %s", i, r, replicas2[i])
		}
	}

	s1 := Strategy{Factor: 1}
	r1 := s1.Replicas("node1", allNodes, "key")
	if len(r1) != 1 || r1[0] != "node1" {
		t.Errorf("expected [node1], got %v", r1)
	}

	s0 := Strategy{Factor: 0}
	if s0.Replicas("node1", allNodes, "key") != nil {
		t.Error("expected nil for Factor=0")
	}
}

func TestPreferenceList(t *testing.T) {
	nodes := []string{"A", "B", "C", "D", "E"}
	ring := &mockRing{nodes: nodes}

	prefs := PreferenceList(ring, "test-key", 3)
	if len(prefs) > 3 {
		t.Errorf("expected at most 3 nodes, got %d", len(prefs))
	}

	seen := make(map[string]bool)
	for _, p := range prefs {
		if seen[p] {
			t.Errorf("duplicate in preference list: %s", p)
		}
		seen[p] = true
	}

	if PreferenceList(ring, "key", 0) != nil {
		t.Error("expected nil for n=0")
	}
}

func TestReplicasExceedNodes(t *testing.T) {
	allNodes := []string{"A", "B"}
	strategy := Strategy{Factor: 5}

	replicas := strategy.Replicas("A", allNodes, "key")
	if len(replicas) > len(allNodes) {
		t.Errorf("replicas should not exceed available nodes: got %d", len(replicas))
	}
}

type mockRing struct {
	nodes []string
}

func (m *mockRing) Get(key string) string {
	if len(m.nodes) == 0 {
		return ""
	}
	h := 0
	for _, c := range key {
		h = h*31 + int(c)
	}
	if h < 0 {
		h = -h
	}
	return m.nodes[h%len(m.nodes)]
}
