package partition

import (
	"hash/crc32"
	"testing"
)

func hashNode(s string) uint32 {
	return crc32.ChecksumIEEE([]byte(s))
}

func TestPartition(t *testing.T) {
	nodes := []string{"node-a", "node-b", "node-c"}
	ranges := Partition(nodes, hashNode)

	if len(ranges) != len(nodes) {
		t.Errorf("expected %d ranges, got %d", len(nodes), len(ranges))
	}

	for _, r := range ranges {
		found := false
		for _, n := range nodes {
			if r.Owner == n {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("invalid owner: %s", r.Owner)
		}
	}

	if Partition(nil, hashNode) != nil {
		t.Error("expected nil for empty nodes")
	}
}

func TestFindOwner(t *testing.T) {
	ranges := []Range{
		{Start: 0, End: 99, Owner: "A"},
		{Start: 100, End: 199, Owner: "B"},
		{Start: 200, End: 299, Owner: "C"},
	}

	tests := []struct {
		token    uint32
		expected string
	}{
		{0, "A"},
		{50, "A"},
		{99, "A"},
		{100, "B"},
		{150, "B"},
		{200, "C"},
		{299, "C"},
		{300, ""},
	}

	for _, tt := range tests {
		got := FindOwner(ranges, tt.token)
		if got != tt.expected {
			t.Errorf("FindOwner(ranges, %d) = %q, want %q", tt.token, got, tt.expected)
		}
	}
}

func TestEvenPartition(t *testing.T) {
	nodes := []string{"s1", "s2", "s3", "s4"}
	ranges := EvenPartition(nodes)

	if len(ranges) != 4 {
		t.Errorf("expected 4 ranges, got %d", len(ranges))
	}

	if ranges[0].Start != 0 {
		t.Errorf("expected first range to start at 0, got %d", ranges[0].Start)
	}

	if ranges[len(ranges)-1].End != ^uint32(0) {
		t.Errorf("expected last range to end at MaxUint32, got %d", ranges[len(ranges)-1].End)
	}

	for i := 1; i < len(ranges); i++ {
		if ranges[i].Start != ranges[i-1].End+1 {
			t.Errorf("gap between range %d and %d", i-1, i)
		}
	}
}

func TestMerge(t *testing.T) {
	ranges := []Range{
		{Start: 0, End: 50, Owner: "A"},
		{Start: 51, End: 100, Owner: "A"},
		{Start: 101, End: 200, Owner: "B"},
		{Start: 201, End: 250, Owner: "B"},
	}

	merged := Merge(ranges)
	if len(merged) != 2 {
		t.Errorf("expected 2 merged ranges, got %d", len(merged))
	}

	if merged[0].Owner != "A" || merged[0].Start != 0 || merged[0].End != 100 {
		t.Errorf("unexpected first merged range: %+v", merged[0])
	}
}

func TestSize(t *testing.T) {
	r := Range{Start: 0, End: 99}
	if Size(r) != 100 {
		t.Errorf("expected size 100, got %d", Size(r))
	}

	r2 := Range{Start: 50, End: 50}
	if Size(r2) != 1 {
		t.Errorf("expected size 1, got %d", Size(r2))
	}
}
