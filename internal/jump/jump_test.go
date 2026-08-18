package jump

import (
	"fmt"
	"testing"
)

func TestHash(t *testing.T) {
	numBuckets := 10
	for i := uint64(0); i < 100; i++ {
		bucket := Hash(i, numBuckets)
		if bucket < 0 || bucket >= numBuckets {
			t.Errorf("Hash(%d, %d) = %d, out of range", i, numBuckets, bucket)
		}
	}

	for i := uint64(0); i < 50; i++ {
		a := Hash(i, numBuckets)
		b := Hash(i, numBuckets)
		if a != b {
			t.Errorf("Hash is not deterministic for key %d", i)
		}
	}

	if Hash(42, 0) != 0 {
		t.Error("expected 0 for zero buckets")
	}
	if Hash(42, 1) != 0 {
		t.Error("expected 0 for single bucket")
	}
}

func TestHashString(t *testing.T) {
	numBuckets := 5
	keys := []string{"hello", "world", "foo", "bar", "consistent", "hash"}

	for _, key := range keys {
		bucket := HashString(key, numBuckets)
		if bucket < 0 || bucket >= numBuckets {
			t.Errorf("HashString(%q, %d) = %d, out of range", key, numBuckets, bucket)
		}
	}

	for _, key := range keys {
		a := HashString(key, numBuckets)
		b := HashString(key, numBuckets)
		if a != b {
			t.Errorf("HashString is not deterministic for key %q", key)
		}
	}
}

func TestDistribution(t *testing.T) {
	numBuckets := 4
	keys := make([]string, 1000)
	for i := range keys {
		keys[i] = fmt.Sprintf("key-%d", i)
	}

	dist := Distribution(keys, numBuckets)
	if len(dist) != numBuckets {
		t.Errorf("expected %d buckets, got %d", numBuckets, len(dist))
	}

	total := 0
	for _, c := range dist {
		total += c
	}
	if total != len(keys) {
		t.Errorf("expected total %d, got %d", len(keys), total)
	}

	for i, c := range dist {
		if c == 0 {
			t.Errorf("bucket %d has zero keys, distribution may be skewed", i)
		}
	}

	if Distribution(keys, 0) != nil {
		t.Error("expected nil for zero buckets")
	}
}

func TestMonotonic(t *testing.T) {
	for key := uint64(0); key < 200; key++ {
		for old := 1; old < 10; old++ {
			newBuckets := old + 1
			if !Monotonic(key, old, newBuckets) {
				t.Errorf("monotonicity violated for key=%d, old=%d, new=%d", key, old, newBuckets)
			}
		}
	}
}
