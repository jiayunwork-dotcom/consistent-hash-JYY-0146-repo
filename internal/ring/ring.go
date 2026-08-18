// Package ring implements a consistent-hash ring with virtual nodes.
//
// Each physical node is placed on the ring replicas times at positions given
// by hash.HashVirtual(node, i). A key is owned by the first virtual node found
// clockwise (i.e. >= hash(key)) from the key's position, wrapping around.
package ring

import (
	"errors"
	"sort"

	"consistent-hash/internal/hash"
	"consistent-hash/internal/node"
)

// ErrEmptyRing is returned by Get when the ring has no nodes.
var ErrEmptyRing = errors.New("ring: empty ring")

// Ring is a consistent-hash ring with virtual nodes.
type Ring struct {
	replicas int
	hash     func(string) uint32
	// keys holds sorted virtual-node hashes for binary search.
	keys []uint32
	// virtual maps a virtual-node hash to its owning physical node.
	virtual map[uint32]string
	// members is the set of physical nodes currently in the ring.
	members map[string]struct{}
}

// New returns an empty ring that places replicas virtual nodes per physical node.
func New(replicas int) *Ring {
	if replicas < 1 {
		replicas = 1
	}
	return &Ring{
		replicas: replicas,
		hash:     hash.HashString,
		virtual:  make(map[uint32]string),
		members:  make(map[string]struct{}),
	}
}

// Len returns the number of physical nodes in the ring.
func (r *Ring) Len() int { return len(r.members) }

// Add inserts a physical node (and its virtual nodes) into the ring.
// Empty/whitespace names are rejected. Re-adding an existing node is a no-op.
func (r *Ring) Add(name string) error {
	n := node.Normalize(name)
	if err := node.Validate(n); err != nil {
		return err
	}
	if _, ok := r.members[n]; ok {
		return nil
	}
	r.members[n] = struct{}{}
	for i := 0; i < r.replicas; i++ {
		vh := hash.HashVirtual(n, i)
		r.virtual[vh] = n
	}
	r.rebuildKeys()
	return nil
}

// Remove deletes a physical node and all of its virtual nodes from the ring.
func (r *Ring) Remove(name string) error {
	n := node.Normalize(name)
	if err := node.Validate(n); err != nil {
		return err
	}
	if _, ok := r.members[n]; !ok {
		return nil
	}
	delete(r.members, n)
	for i := 0; i < r.replicas; i++ {
		vh := hash.HashVirtual(n, i)
		delete(r.virtual, vh)
	}
	r.rebuildKeys()
	return nil
}

func (r *Ring) rebuildKeys() {
	r.keys = r.keys[:0]
	for vh := range r.virtual {
		r.keys = append(r.keys, vh)
	}
	sort.Slice(r.keys, func(i, j int) bool { return r.keys[i] < r.keys[j] })
}

// Get returns the physical node that owns key, chosen as the first virtual
// node clockwise of hash(key). It returns ErrEmptyRing if the ring is empty.
func (r *Ring) Get(key string) (string, error) {
	if len(r.keys) == 0 {
		return "", ErrEmptyRing
	}
	h := r.hash(key)
	idx := sort.Search(len(r.keys), func(i int) bool { return r.keys[i] >= h })
	if idx == len(r.keys) {
		idx = 0
	}
	return r.virtual[r.keys[idx]], nil
}
