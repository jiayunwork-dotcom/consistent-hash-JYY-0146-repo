package hash

import "testing"

func TestHashStringDeterministic(t *testing.T) {
	if HashString("hello") != HashString("hello") {
		t.Error("HashString is not deterministic")
	}
	if HashString("a") != 0 && HashString("a") == HashString("b") {
		t.Error("distinct inputs collided")
	}
}

func TestHashStringNonZeroForNonEmpty(t *testing.T) {
	// FNV-1a of a non-empty string is effectively never zero.
	if HashString("consistent-hash") == 0 {
		t.Errorf("unexpected zero hash for non-empty input")
	}
}

func TestHashVirtualDistinct(t *testing.T) {
	base := HashVirtual("node-a", 0)
	for i := 1; i < 10; i++ {
		if h := HashVirtual("node-a", i); h == base {
			t.Errorf("virtual node %d collided with virtual node 0", i)
		}
	}
}

func TestHashVirtualStable(t *testing.T) {
	if HashVirtual("node-b", 7) != HashVirtual("node-b", 7) {
		t.Error("HashVirtual is not deterministic across calls")
	}
}

func TestHashBytes(t *testing.T) {
	h1 := HashBytes([]byte("test"))
	h2 := HashBytes([]byte("test"))
	if h1 != h2 {
		t.Error("HashBytes not deterministic")
	}
	if HashBytes([]byte("x")) == HashBytes([]byte("y")) {
		t.Error("distinct inputs collided")
	}
}

func TestVirtualPositions(t *testing.T) {
	pos := VirtualPositions("srv-01", 10)
	if len(pos) != 10 {
		t.Fatalf("len = %d, want 10", len(pos))
	}
	// all should be distinct
	seen := map[uint32]bool{}
	for _, p := range pos {
		if seen[p] {
			t.Error("duplicate position found")
		}
		seen[p] = true
	}
}

func TestSpread(t *testing.T) {
	pos := VirtualPositions("test-node", 100)
	variance := Spread(pos)
	// just verify it runs without error and is non-negative
	if variance < 0 {
		t.Errorf("Spread = %f, should be non-negative", variance)
	}
}
