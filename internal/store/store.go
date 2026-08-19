// Package store provides a persistent consistent-hash ring backed by snapshot.
//
// The Store wraps a Ring and persists membership changes via atomic snapshots.
// On Open, the last snapshot is loaded to rebuild the ring. Checkpoint writes
// the current ring state. The store verifies key stability: after recovery,
// lookups return the same node as before the restart for all non-removed nodes.
//
// Minimal-movement guarantee: adding or removing a node moves only keys that
// must change ownership. This is inherent to consistent hashing but tested
// explicitly.
package store

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"consistent-hash/internal/ring"
	"consistent-hash/internal/snapshot"
)

const snapshotFile = "ring.snap"

var (
	ErrClosed = errors.New("store: closed")
)

// Store is a persistent consistent-hash ring.
type Store struct {
	dir      string
	snapPath string
	ring     *ring.Ring
	closed   bool

	// WasRecovered indicates state was loaded from a snapshot.
	WasRecovered bool
}

// Open opens or creates a persistent ring in dir with the given replicas count.
// If a valid snapshot exists, the ring is rebuilt from it.
func Open(dir string, replicas int) (*Store, error) {
	if replicas < 1 {
		return nil, fmt.Errorf("store: replicas must be >= 1")
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("store: mkdir: %w", err)
	}

	snapPath := filepath.Join(dir, snapshotFile)
	s := &Store{
		dir:      dir,
		snapPath: snapPath,
	}

	// try loading snapshot
	state, err := snapshot.Load(snapPath)
	if err == nil {
		r := ring.New(state.Replicas)
		for _, n := range state.Nodes {
			r.Add(n)
		}
		s.ring = r
		s.WasRecovered = true
		return s, nil
	}

	// fresh ring
	s.ring = ring.New(replicas)
	return s, nil
}

// Add inserts a physical node into the ring.
func (s *Store) Add(name string) error {
	if s.closed {
		return ErrClosed
	}
	return s.ring.Add(name)
}

// Remove deletes a physical node from the ring.
func (s *Store) Remove(name string) error {
	if s.closed {
		return ErrClosed
	}
	return s.ring.Remove(name)
}

// Get looks up the owner of key.
func (s *Store) Get(key string) (string, error) {
	if s.closed {
		return "", ErrClosed
	}
	return s.ring.Get(key)
}

// Members returns the current physical node list.
func (s *Store) Members() []string {
	if s.closed {
		return nil
	}
	m := s.ring.Members()
	sort.Strings(m)
	return m
}

// Len returns the number of physical nodes.
func (s *Store) Len() int {
	if s.closed {
		return 0
	}
	return s.ring.Len()
}

// Checkpoint writes the current ring state to disk atomically.
func (s *Store) Checkpoint() error {
	if s.closed {
		return ErrClosed
	}
	members := s.ring.Members()
	sort.Strings(members)
	state := &snapshot.RingState{
		Replicas: s.ring.Replicas(),
		Nodes:    members,
	}
	return snapshot.Save(s.snapPath, state)
}

// Close marks the store as closed.
func (s *Store) Close() error {
	s.closed = true
	return nil
}

// MinimalMovement computes how many keys from a sample would change ownership
// if a node is added or removed. Returns (moved, total) where moved is the
// count of keys that changed owner out of total sample keys.
func (s *Store) MinimalMovement(keys []string, addNode string) (moved, total int) {
	if s.closed {
		return 0, 0
	}
	total = len(keys)
	ownersBefore := make(map[string]string, total)
	for _, k := range keys {
		owner, _ := s.ring.Get(k)
		ownersBefore[k] = owner
	}

	// temporarily add the node
	s.ring.Add(addNode)
	for _, k := range keys {
		owner, _ := s.ring.Get(k)
		if owner != ownersBefore[k] {
			moved++
		}
	}
	// revert
	s.ring.Remove(addNode)
	return moved, total
}

// MinimalMovementRemove computes how many keys change ownership when removing
// a node.
func (s *Store) MinimalMovementRemove(keys []string, removeNode string) (moved, total int) {
	if s.closed {
		return 0, 0
	}
	total = len(keys)
	ownersBefore := make(map[string]string, total)
	for _, k := range keys {
		owner, _ := s.ring.Get(k)
		ownersBefore[k] = owner
	}

	s.ring.Remove(removeNode)
	for _, k := range keys {
		owner, _ := s.ring.Get(k)
		if owner != ownersBefore[k] {
			moved++
		}
	}
	// restore
	s.ring.Add(removeNode)
	return moved, total
}
