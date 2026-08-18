package rebalance

import (
	"hash/crc32"
	"testing"
)

func hashStr(s string) uint32 {
	return crc32.ChecksumIEEE([]byte(s))
}

func TestPlan(t *testing.T) {
	oldNodes := []string{"A", "B", "C"}
	newNodes := []string{"A", "B", "C", "D"}
	keys := []string{"key1", "key2", "key3", "key4", "key5", "key6", "key7", "key8", "key9", "key10"}

	transfers := Plan(oldNodes, newNodes, keys, hashStr)

	if transfers == nil {
		t.Log("no transfers needed, which is possible but unlikely")
	}

	for _, tr := range transfers {
		if tr.From == tr.To {
			t.Errorf("transfer from and to are the same: %s -> %s for key %s", tr.From, tr.To, tr.Key)
		}
	}

	result := Plan(nil, newNodes, keys, hashStr)
	if result != nil {
		t.Error("expected nil for empty old nodes")
	}
}

func TestEstimateMovement(t *testing.T) {
	ratio := EstimateMovement(3, 4)
	if ratio <= 0 || ratio > 1 {
		t.Errorf("expected ratio in (0, 1], got %f", ratio)
	}

	ratio = EstimateMovement(5, 5)
	if ratio != 0 {
		t.Errorf("expected 0 for same count, got %f", ratio)
	}

	ratio = EstimateMovement(0, 5)
	if ratio != 0 {
		t.Errorf("expected 0 for invalid input, got %f", ratio)
	}

	ratio = EstimateMovement(10, 8)
	if ratio <= 0 {
		t.Errorf("expected positive ratio for node removal, got %f", ratio)
	}
}

func TestMinimalTransfer(t *testing.T) {
	old := map[string][]string{
		"node1": {"a", "b", "c"},
		"node2": {"d", "e"},
		"node3": {"f"},
	}
	new := map[string][]string{
		"node1": {"a", "b"},
		"node2": {"d", "e", "f"},
		"node3": {"c"},
	}

	transfers := MinimalTransfer(old, new)
	if len(transfers) != 2 {
		t.Errorf("expected 2 transfers, got %d", len(transfers))
	}

	transferMap := make(map[string]Transfer)
	for _, tr := range transfers {
		transferMap[tr.Key] = tr
	}

	if tr, ok := transferMap["c"]; ok {
		if tr.From != "node1" || tr.To != "node3" {
			t.Errorf("unexpected transfer for 'c': %+v", tr)
		}
	}

	if tr, ok := transferMap["f"]; ok {
		if tr.From != "node3" || tr.To != "node2" {
			t.Errorf("unexpected transfer for 'f': %+v", tr)
		}
	}
}
