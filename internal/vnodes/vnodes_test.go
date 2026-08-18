package vnodes

import (
	"hash/crc32"
	"testing"
)

func simpleHash(data []byte) uint32 {
	return crc32.ChecksumIEEE(data)
}

func TestDistribute(t *testing.T) {
	nodes := []string{"node-A", "node-B", "node-C"}
	replicaFactor := 3

	vnodes := Distribute(nodes, replicaFactor, simpleHash)
	if len(vnodes) != len(nodes)*replicaFactor {
		t.Errorf("expected %d vnodes, got %d", len(nodes)*replicaFactor, len(vnodes))
	}

	for _, vn := range vnodes {
		found := false
		for _, n := range nodes {
			if vn.PhysicalNode == n {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("vnode %s has invalid physical node %s", vn.ID, vn.PhysicalNode)
		}
	}

	result := Distribute(nil, 3, simpleHash)
	if result != nil {
		t.Error("expected nil for empty nodes")
	}

	result = Distribute(nodes, 0, simpleHash)
	if result != nil {
		t.Error("expected nil for zero replica factor")
	}
}

func TestAssignedVNodes(t *testing.T) {
	nodes := []string{"alpha", "beta", "gamma"}
	vnodes := Distribute(nodes, 4, simpleHash)

	assigned := AssignedVNodes(vnodes, "beta")
	if len(assigned) != 4 {
		t.Errorf("expected 4 vnodes for beta, got %d", len(assigned))
	}

	for _, vn := range assigned {
		if vn.PhysicalNode != "beta" {
			t.Errorf("expected physical node beta, got %s", vn.PhysicalNode)
		}
	}

	assigned = AssignedVNodes(vnodes, "nonexistent")
	if len(assigned) != 0 {
		t.Errorf("expected 0 vnodes for nonexistent, got %d", len(assigned))
	}
}

func TestRemovePhysical(t *testing.T) {
	nodes := []string{"x", "y", "z"}
	vnodes := Distribute(nodes, 5, simpleHash)
	original := len(vnodes)

	remaining := RemovePhysical(vnodes, "y")
	if len(remaining) != original-5 {
		t.Errorf("expected %d vnodes after removal, got %d", original-5, len(remaining))
	}

	for _, vn := range remaining {
		if vn.PhysicalNode == "y" {
			t.Error("found removed physical node in remaining vnodes")
		}
	}
}

func TestCount(t *testing.T) {
	nodes := []string{"s1", "s2", "s3"}
	vnodes := Distribute(nodes, 7, simpleHash)

	count := Count(vnodes, "s1")
	if count != 7 {
		t.Errorf("expected count 7, got %d", count)
	}

	count = Count(vnodes, "missing")
	if count != 0 {
		t.Errorf("expected count 0, got %d", count)
	}
}

func TestPositions(t *testing.T) {
	nodes := []string{"a", "b"}
	vnodes := Distribute(nodes, 3, simpleHash)

	positions := Positions(vnodes)
	if len(positions) != 6 {
		t.Errorf("expected 6 positions, got %d", len(positions))
	}
}
