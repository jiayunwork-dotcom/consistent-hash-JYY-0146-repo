package balance

import (
	"fmt"
	"testing"

	"consistent-hash/internal/ring"
)

func genKeys(n int) []string {
	keys := make([]string, n)
	for i := range keys {
		keys[i] = fmt.Sprintf("key-%05d", i)
	}
	return keys
}

func TestMeasureDistribution(t *testing.T) {
	r := ring.New(100)
	r.Add("a")
	r.Add("b")
	r.Add("c")
	r.Add("d")

	keys := genKeys(4000)
	dist, err := Measure(r, keys)
	if err != nil {
		t.Fatalf("Measure: %v", err)
	}

	if dist.Total != 4000 {
		t.Errorf("Total = %d, want 4000", dist.Total)
	}
	if len(dist.NodeCounts) != 4 {
		t.Errorf("NodeCounts has %d entries, want 4", len(dist.NodeCounts))
	}

	// each node should have roughly 1000 keys (within 3x tolerance)
	for node, count := range dist.NodeCounts {
		if count < 200 || count > 2000 {
			t.Errorf("node %q has %d keys (expected roughly 1000)", node, count)
		}
	}
}

func TestMaxMinRatio(t *testing.T) {
	r := ring.New(200)
	r.Add("n1")
	r.Add("n2")

	keys := genKeys(2000)
	dist, _ := Measure(r, keys)

	ratio := dist.MaxMinRatio()
	// with 200 replicas, should be reasonably balanced (ratio < 3)
	if ratio > 3.0 {
		t.Errorf("MaxMinRatio = %.2f, too unbalanced", ratio)
	}
}

func TestStdDev(t *testing.T) {
	r := ring.New(200)
	r.Add("a")
	r.Add("b")
	r.Add("c")

	keys := genKeys(3000)
	dist, _ := Measure(r, keys)

	sd := dist.StdDev()
	mean := float64(dist.Total) / float64(len(dist.NodeCounts))
	// stddev should be less than mean (reasonable balance)
	if sd > mean {
		t.Errorf("StdDev = %.1f exceeds mean = %.1f", sd, mean)
	}
}

func TestPlanAddNodeAllMovesToNew(t *testing.T) {
	r := ring.New(100)
	r.Add("a")
	r.Add("b")
	r.Add("c")

	keys := genKeys(500)
	plan, err := PlanAddNode(r, keys, "d")
	if err != nil {
		t.Fatalf("PlanAddNode: %v", err)
	}

	if len(plan.Moves) == 0 {
		t.Fatal("expected some moves")
	}

	// all moves should go TO the new node
	for _, m := range plan.Moves {
		if m.To != "d" {
			t.Errorf("Move %q: To = %q, want d", m.Key, m.To)
			break
		}
	}
}

func TestPlanRemoveNodeAllMovesFromRemoved(t *testing.T) {
	r := ring.New(100)
	r.Add("a")
	r.Add("b")
	r.Add("c")

	keys := genKeys(500)
	plan, err := PlanRemoveNode(r, keys, "b")
	if err != nil {
		t.Fatalf("PlanRemoveNode: %v", err)
	}

	if len(plan.Moves) == 0 {
		t.Fatal("expected some moves")
	}

	// all moves should come FROM the removed node
	for _, m := range plan.Moves {
		if m.From != "b" {
			t.Errorf("Move %q: From = %q, want b", m.Key, m.From)
			break
		}
	}
}

func TestSortedNodes(t *testing.T) {
	r := ring.New(100)
	r.Add("x")
	r.Add("y")

	keys := genKeys(100)
	dist, _ := Measure(r, keys)

	sorted := dist.SortedNodes()
	if len(sorted) != 2 {
		t.Fatalf("SortedNodes len = %d, want 2", len(sorted))
	}
	if sorted[0].Count < sorted[1].Count {
		t.Error("first should have >= count than second")
	}
}

func TestMeasureEmptyRing(t *testing.T) {
	r := ring.New(100)
	_, err := Measure(r, genKeys(10))
	if err == nil {
		t.Error("expected error for empty ring")
	}
}
