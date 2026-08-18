package ring

import (
	"testing"

	"consistent-hash/internal/node"
)

func TestRingGetEmpty(t *testing.T) {
	r := New(100)
	if _, err := r.Get("anything"); err != ErrEmptyRing {
		t.Errorf("err = %v, want ErrEmptyRing", err)
	}
}

func TestRingAddInvalidName(t *testing.T) {
	r := New(100)
	if err := r.Add("   "); err != node.ErrEmptyName {
		t.Errorf("err = %v, want ErrEmptyName", err)
	}
}

func TestRingAddGetConsistent(t *testing.T) {
	r := New(100)
	for _, n := range []string{"a", "b", "c"} {
		if err := r.Add(n); err != nil {
			t.Fatalf("Add(%q): %v", n, err)
		}
	}
	first, err := r.Get("user:42")
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 20; i++ {
		if got, _ := r.Get("user:42"); got != first {
			t.Errorf("Get not stable: %q vs %q", got, first)
		}
	}
}

func TestRingRemove(t *testing.T) {
	r := New(50)
	_ = r.Add("a")
	_ = r.Add("b")
	before, err := r.Get("k")
	if err != nil {
		t.Fatal(err)
	}
	_ = r.Remove("a")
	if r.Len() != 1 {
		t.Fatalf("Len = %d, want 1", r.Len())
	}
	after, err := r.Get("k")
	if err != nil {
		t.Fatal(err)
	}
	// After removing a node the result must still be a surviving member.
	if after != "b" {
		t.Errorf("Get = %q, want %q (surviving node)", after, "b")
	}
	if before != "a" && before != "b" {
		t.Errorf("unexpected owner before removal: %q", before)
	}
}

func TestRingGetDistributes(t *testing.T) {
	r := New(200)
	for _, n := range []string{"n1", "n2", "n3", "n4"} {
		_ = r.Add(n)
	}
	seen := make(map[string]int)
	for i := 0; i < 4000; i++ {
		owner, err := r.Get("key-" + itoa(i))
		if err != nil {
			t.Fatal(err)
		}
		seen[owner]++
	}
	// With enough virtual nodes every node should receive traffic.
	if len(seen) != 4 {
		t.Errorf("only %d nodes received keys: %v", len(seen), seen)
	}
}

func TestRingAddIsIdempotent(t *testing.T) {
	r := New(10)
	_ = r.Add("x")
	_ = r.Add("x")
	if r.Len() != 1 {
		t.Errorf("Len = %d, want 1 (idempotent Add)", r.Len())
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

func TestRingRemoveMissingNoError(t *testing.T) {
	r := New(10)
	if err := r.Remove("ghost"); err != nil {
		t.Errorf("Remove of missing node returned %v, want nil", err)
	}
}

func TestRingMembers(t *testing.T) {
	r := New(10)
	r.Add("c")
	r.Add("a")
	r.Add("b")

	members := r.Members()
	if len(members) != 3 {
		t.Fatalf("Members len = %d, want 3", len(members))
	}
	// order is not guaranteed, just verify all present
	has := map[string]bool{}
	for _, m := range members {
		has[m] = true
	}
	for _, want := range []string{"a", "b", "c"} {
		if !has[want] {
			t.Errorf("Members missing %q", want)
		}
	}
}

func TestRingReplicas(t *testing.T) {
	r := New(42)
	if r.Replicas() != 42 {
		t.Errorf("Replicas = %d, want 42", r.Replicas())
	}
}

func TestRingMinimalMovementProperty(t *testing.T) {
	r := New(100)
	r.Add("n1")
	r.Add("n2")
	r.Add("n3")

	keys := make([]string, 500)
	ownersBefore := make(map[string]string, 500)
	for i := range keys {
		keys[i] = "obj-" + itoa(i)
		ownersBefore[keys[i]], _ = r.Get(keys[i])
	}

	// add a node
	r.Add("n4")
	moved := 0
	for _, k := range keys {
		owner, _ := r.Get(k)
		if owner != ownersBefore[k] {
			// moved key must now be on the new node
			if owner != "n4" {
				t.Errorf("key %q moved from %q to %q (not the new node n4)", k, ownersBefore[k], owner)
			}
			moved++
		}
	}
	// at least some keys should move
	if moved == 0 {
		t.Error("expected at least some keys to move to new node")
	}
}

func TestRingRemoveNodeNoOrphanLookup(t *testing.T) {
	r := New(100)
	r.Add("keep-1")
	r.Add("keep-2")
	r.Add("gone")

	r.Remove("gone")

	for i := 0; i < 200; i++ {
		owner, _ := r.Get("k-" + itoa(i))
		if owner == "gone" {
			t.Fatalf("key %d still points to removed node", i)
		}
	}
}
