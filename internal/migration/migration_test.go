package migration

import (
	"hash/crc32"
	"testing"
)

func makeHashFunc(nodes []string) func(string) string {
	return func(key string) string {
		if len(nodes) == 0 {
			return ""
		}
		h := crc32.ChecksumIEEE([]byte(key))
		idx := int(h) % len(nodes)
		return nodes[idx]
	}
}

func TestComputeMigration(t *testing.T) {
	oldNodes := []string{"A", "B", "C"}
	newNodes := []string{"A", "B", "C", "D"}
	keys := []string{"k1", "k2", "k3", "k4", "k5", "k6", "k7", "k8", "k9", "k10"}

	before := makeHashFunc(oldNodes)
	after := makeHashFunc(newNodes)

	plan := ComputeMigration(before, after, keys)

	for _, step := range plan.Steps {
		if step.Source == step.Dest {
			t.Errorf("step has same source and dest: %+v", step)
		}
	}

	for i := 1; i < len(plan.Steps); i++ {
		if plan.Steps[i].Key < plan.Steps[i-1].Key {
			t.Error("steps are not sorted by key")
		}
	}

	emptyPlan := ComputeMigration(nil, after, keys)
	if emptyPlan.AffectedKeys() != 0 {
		t.Error("expected 0 affected keys for nil before func")
	}
}

func TestAffectedKeys(t *testing.T) {
	before := func(key string) string { return "A" }
	after := func(key string) string {
		if key == "k1" || key == "k2" {
			return "B"
		}
		return "A"
	}
	keys := []string{"k1", "k2", "k3", "k4"}

	plan := ComputeMigration(before, after, keys)
	if plan.AffectedKeys() != 2 {
		t.Errorf("expected 2 affected keys, got %d", plan.AffectedKeys())
	}
}

func TestBySource(t *testing.T) {
	plan := &MigrationPlan{
		Steps: []Step{
			{Key: "a", Source: "node1", Dest: "node3"},
			{Key: "b", Source: "node1", Dest: "node2"},
			{Key: "c", Source: "node2", Dest: "node3"},
		},
	}

	bySource := plan.BySource()
	if len(bySource["node1"]) != 2 {
		t.Errorf("expected 2 steps from node1, got %d", len(bySource["node1"]))
	}
	if len(bySource["node2"]) != 1 {
		t.Errorf("expected 1 step from node2, got %d", len(bySource["node2"]))
	}
}

func TestReverse(t *testing.T) {
	plan := &MigrationPlan{
		Steps: []Step{
			{Key: "x", Source: "A", Dest: "B"},
			{Key: "y", Source: "C", Dest: "D"},
		},
	}

	reversed := plan.Reverse()
	if len(reversed.Steps) != 2 {
		t.Errorf("expected 2 steps in reversed plan, got %d", len(reversed.Steps))
	}

	if reversed.Steps[0].Source != "B" || reversed.Steps[0].Dest != "A" {
		t.Errorf("unexpected reversed step: %+v", reversed.Steps[0])
	}
}

func TestPercentage(t *testing.T) {
	plan := &MigrationPlan{
		Steps: make([]Step, 25),
	}

	pct := plan.Percentage(100)
	if pct != 25.0 {
		t.Errorf("expected 25%%, got %.1f%%", pct)
	}

	if plan.Percentage(0) != 0 {
		t.Error("expected 0 for zero total keys")
	}
}
