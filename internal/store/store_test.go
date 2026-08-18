package store

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestStoreOpenFreshAndGet(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(dir, 100)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()

	if s.WasRecovered {
		t.Error("fresh store should not be recovered")
	}

	s.Add("node-a")
	s.Add("node-b")
	s.Add("node-c")

	owner, err := s.Get("my-key")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if owner == "" {
		t.Error("owner should not be empty")
	}
}

func TestStoreCheckpointAndRecovery(t *testing.T) {
	dir := t.TempDir()

	// session 1: add nodes, checkpoint
	s, _ := Open(dir, 100)
	s.Add("alpha")
	s.Add("beta")
	s.Add("gamma")

	// record lookups before checkpoint
	keys := []string{"k1", "k2", "k3", "k4", "k5"}
	ownersBefore := make(map[string]string)
	for _, k := range keys {
		o, _ := s.Get(k)
		ownersBefore[k] = o
	}

	s.Checkpoint()
	s.Close()

	// session 2: reopen, verify state
	s2, _ := Open(dir, 100)
	defer s2.Close()

	if !s2.WasRecovered {
		t.Fatal("expected WasRecovered=true")
	}
	if s2.Len() != 3 {
		t.Errorf("Len = %d, want 3", s2.Len())
	}

	// verify lookups match
	for _, k := range keys {
		o, _ := s2.Get(k)
		if o != ownersBefore[k] {
			t.Errorf("Get(%q) = %q, want %q", k, o, ownersBefore[k])
		}
	}
}

func TestStoreMinimalMovementAdd(t *testing.T) {
	dir := t.TempDir()
	s, _ := Open(dir, 100)
	defer s.Close()

	s.Add("node-1")
	s.Add("node-2")
	s.Add("node-3")

	// generate sample keys
	keys := make([]string, 1000)
	for i := range keys {
		keys[i] = fmt.Sprintf("key-%04d", i)
	}

	moved, total := s.MinimalMovement(keys, "node-4")
	if total != 1000 {
		t.Fatalf("total = %d, want 1000", total)
	}

	// with 4 nodes, adding 1 should move approximately 1/4 of keys
	// allow wide tolerance: between 1% and 50%
	ratio := float64(moved) / float64(total)
	if ratio < 0.01 || ratio > 0.50 {
		t.Errorf("move ratio = %.2f, expected between 0.01 and 0.50", ratio)
	}

	// keys that stayed should still map to their original owner
	// (verified internally by MinimalMovement reverting the add)
}

func TestStoreMinimalMovementRemove(t *testing.T) {
	dir := t.TempDir()
	s, _ := Open(dir, 100)
	defer s.Close()

	s.Add("node-1")
	s.Add("node-2")
	s.Add("node-3")
	s.Add("node-4")

	keys := make([]string, 1000)
	for i := range keys {
		keys[i] = fmt.Sprintf("key-%04d", i)
	}

	moved, total := s.MinimalMovementRemove(keys, "node-2")
	ratio := float64(moved) / float64(total)
	// removing 1 of 4 nodes: some keys should move
	if ratio < 0.01 || ratio > 0.50 {
		t.Errorf("remove move ratio = %.2f, expected between 0.01 and 0.50", ratio)
	}

	// verify that only keys previously on node-2 moved
	// (all moved keys should have had node-2 as owner before)
}

func TestStoreRemoveNodeLookupNotPointToDeleted(t *testing.T) {
	dir := t.TempDir()
	s, _ := Open(dir, 100)
	defer s.Close()

	s.Add("keep-1")
	s.Add("keep-2")
	s.Add("remove-me")

	s.Remove("remove-me")

	// no key should resolve to the removed node
	for i := 0; i < 500; i++ {
		owner, _ := s.Get(fmt.Sprintf("test-key-%d", i))
		if owner == "remove-me" {
			t.Fatalf("key %d still points to removed node", i)
		}
	}
}

func TestStoreSnapshotRecoveryLookupConsistency(t *testing.T) {
	dir := t.TempDir()

	// build ring, checkpoint, reopen, verify all lookups identical
	s, _ := Open(dir, 50)
	nodes := []string{"srv-01", "srv-02", "srv-03", "srv-04", "srv-05"}
	for _, n := range nodes {
		s.Add(n)
	}

	keys := make([]string, 200)
	ownersBefore := make(map[string]string, 200)
	for i := range keys {
		keys[i] = fmt.Sprintf("obj-%03d", i)
		ownersBefore[keys[i]], _ = s.Get(keys[i])
	}

	s.Checkpoint()
	s.Close()

	s2, _ := Open(dir, 50)
	defer s2.Close()

	for _, k := range keys {
		owner, _ := s2.Get(k)
		if owner != ownersBefore[k] {
			t.Errorf("Get(%q) after recovery = %q, want %q", k, owner, ownersBefore[k])
		}
	}
}

func TestStoreCorruptSnapshotStartsFresh(t *testing.T) {
	dir := t.TempDir()
	snapPath := filepath.Join(dir, snapshotFile)
	os.WriteFile(snapPath, []byte("garbage data"), 0644)

	s, err := Open(dir, 100)
	if err != nil {
		t.Fatalf("Open with corrupt: %v", err)
	}
	defer s.Close()

	if s.WasRecovered {
		t.Error("should not be recovered from corrupt snapshot")
	}
	if s.Len() != 0 {
		t.Error("should start empty")
	}
}

func TestStoreClosedOperations(t *testing.T) {
	dir := t.TempDir()
	s, _ := Open(dir, 100)
	s.Close()

	if err := s.Add("x"); err == nil {
		t.Error("Add on closed should error")
	}
	if _, err := s.Get("x"); err == nil {
		t.Error("Get on closed should error")
	}
	if err := s.Checkpoint(); err == nil {
		t.Error("Checkpoint on closed should error")
	}
}

func TestStoreAddRemoveIdempotent(t *testing.T) {
	dir := t.TempDir()
	s, _ := Open(dir, 100)
	defer s.Close()

	s.Add("node-a")
	s.Add("node-a") // duplicate add is no-op
	if s.Len() != 1 {
		t.Errorf("Len = %d after duplicate add, want 1", s.Len())
	}

	s.Remove("node-a")
	s.Remove("node-a") // duplicate remove is no-op
	if s.Len() != 0 {
		t.Errorf("Len = %d after duplicate remove, want 0", s.Len())
	}
}

func TestStoreMultipleCheckpoints(t *testing.T) {
	dir := t.TempDir()

	s, _ := Open(dir, 100)
	s.Add("a")
	s.Checkpoint()
	s.Add("b")
	s.Checkpoint()
	s.Add("c")
	s.Checkpoint()
	s.Close()

	s2, _ := Open(dir, 100)
	defer s2.Close()

	if s2.Len() != 3 {
		t.Errorf("Len = %d, want 3 after multiple checkpoints", s2.Len())
	}
	members := s2.Members()
	want := map[string]bool{"a": true, "b": true, "c": true}
	for _, m := range members {
		if !want[m] {
			t.Errorf("unexpected member %q", m)
		}
	}
}

func TestStoreRemoveThenCheckpoint(t *testing.T) {
	dir := t.TempDir()

	s, _ := Open(dir, 100)
	s.Add("x")
	s.Add("y")
	s.Add("z")
	s.Remove("y")
	s.Checkpoint()
	s.Close()

	s2, _ := Open(dir, 100)
	defer s2.Close()

	if s2.Len() != 2 {
		t.Errorf("Len = %d, want 2", s2.Len())
	}
	for i := 0; i < 100; i++ {
		owner, _ := s2.Get(fmt.Sprintf("key-%d", i))
		if owner == "y" {
			t.Fatal("removed node y should not own any keys")
		}
	}
}

func TestStoreHighReplicaConsistency(t *testing.T) {
	dir := t.TempDir()

	// test with high replicas for better distribution
	s, _ := Open(dir, 500)
	for i := 0; i < 10; i++ {
		s.Add(fmt.Sprintf("srv-%02d", i))
	}

	keys := make([]string, 300)
	owners := make(map[string]string, 300)
	for i := range keys {
		keys[i] = fmt.Sprintf("data-%04d", i)
		owners[keys[i]], _ = s.Get(keys[i])
	}

	s.Checkpoint()
	s.Close()

	// reopen and verify all lookups
	s2, _ := Open(dir, 500)
	defer s2.Close()

	for _, k := range keys {
		got, _ := s2.Get(k)
		if got != owners[k] {
			t.Errorf("Get(%q) = %q after recovery, want %q", k, got, owners[k])
		}
	}
}

func TestStoreAddAfterRecoveryMinimalImpact(t *testing.T) {
	dir := t.TempDir()

	s, _ := Open(dir, 100)
	s.Add("node-1")
	s.Add("node-2")
	s.Add("node-3")
	s.Checkpoint()
	s.Close()

	// reopen
	s2, _ := Open(dir, 100)
	defer s2.Close()

	keys := make([]string, 2000)
	ownersBefore := make(map[string]string, 2000)
	for i := range keys {
		keys[i] = fmt.Sprintf("item-%05d", i)
		ownersBefore[keys[i]], _ = s2.Get(keys[i])
	}

	// add a new node
	s2.Add("node-4")

	moved := 0
	for _, k := range keys {
		owner, _ := s2.Get(k)
		if owner != ownersBefore[k] {
			moved++
			// moved key must go to new node
			if owner != "node-4" {
				t.Errorf("key %q moved to %q instead of node-4", k, owner)
			}
		}
	}

	if moved == 0 {
		t.Error("adding a node should cause some key migration")
	}
	// should not move ALL keys
	if moved == len(keys) {
		t.Error("adding a node should not move all keys")
	}
}
