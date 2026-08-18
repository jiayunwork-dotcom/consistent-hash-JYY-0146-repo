// Package balance measures the distribution quality of a consistent-hash ring.
//
// It provides statistics about how evenly keys are distributed among nodes,
// and computes migration plans when nodes are added or removed.
package balance

import (
	"fmt"
	"math"
	"sort"

	"consistent-hash/internal/ring"
)

// Distribution holds per-node key ownership counts.
type Distribution struct {
	NodeCounts map[string]int
	Total      int
}

// Measure distributes a set of keys across the ring and returns ownership stats.
func Measure(r *ring.Ring, keys []string) (*Distribution, error) {
	if r.Len() == 0 {
		return nil, fmt.Errorf("balance: ring is empty")
	}
	counts := make(map[string]int)
	for _, k := range keys {
		owner, err := r.Get(k)
		if err != nil {
			return nil, err
		}
		counts[owner]++
	}
	return &Distribution{NodeCounts: counts, Total: len(keys)}, nil
}

// StdDev returns the standard deviation of key counts across nodes.
func (d *Distribution) StdDev() float64 {
	if len(d.NodeCounts) == 0 {
		return 0
	}
	mean := float64(d.Total) / float64(len(d.NodeCounts))
	var sumSq float64
	for _, c := range d.NodeCounts {
		diff := float64(c) - mean
		sumSq += diff * diff
	}
	return math.Sqrt(sumSq / float64(len(d.NodeCounts)))
}

// MaxMinRatio returns max_count / min_count. A perfectly balanced ring has
// ratio 1.0. Returns 0 if the ring has no nodes.
func (d *Distribution) MaxMinRatio() float64 {
	if len(d.NodeCounts) == 0 {
		return 0
	}
	minC, maxC := math.MaxInt, 0
	for _, c := range d.NodeCounts {
		if c < minC {
			minC = c
		}
		if c > maxC {
			maxC = c
		}
	}
	if minC == 0 {
		return math.Inf(1)
	}
	return float64(maxC) / float64(minC)
}

// MigrationPlan describes which keys move from which node to which node.
type MigrationPlan struct {
	Moves []Move
}

// Move represents a single key ownership change.
type Move struct {
	Key  string
	From string
	To   string
}

// PlanAddNode computes which keys would migrate if a node is added.
func PlanAddNode(r *ring.Ring, keys []string, newNode string) (*MigrationPlan, error) {
	if r.Len() == 0 {
		return nil, fmt.Errorf("balance: ring is empty")
	}

	// record current owners
	before := make(map[string]string, len(keys))
	for _, k := range keys {
		owner, _ := r.Get(k)
		before[k] = owner
	}

	// add node
	if err := r.Add(newNode); err != nil {
		return nil, err
	}

	var moves []Move
	for _, k := range keys {
		owner, _ := r.Get(k)
		if owner != before[k] {
			moves = append(moves, Move{Key: k, From: before[k], To: owner})
		}
	}

	// revert
	r.Remove(newNode)

	return &MigrationPlan{Moves: moves}, nil
}

// PlanRemoveNode computes which keys would migrate if a node is removed.
func PlanRemoveNode(r *ring.Ring, keys []string, removeNode string) (*MigrationPlan, error) {
	if r.Len() == 0 {
		return nil, fmt.Errorf("balance: ring is empty")
	}

	before := make(map[string]string, len(keys))
	for _, k := range keys {
		owner, _ := r.Get(k)
		before[k] = owner
	}

	r.Remove(removeNode)

	var moves []Move
	for _, k := range keys {
		owner, _ := r.Get(k)
		if owner != before[k] {
			moves = append(moves, Move{Key: k, From: before[k], To: owner})
		}
	}

	// restore
	r.Add(removeNode)

	return &MigrationPlan{Moves: moves}, nil
}

// SortedNodes returns the nodes sorted by their key count (descending).
func (d *Distribution) SortedNodes() []NodeStat {
	stats := make([]NodeStat, 0, len(d.NodeCounts))
	for n, c := range d.NodeCounts {
		stats = append(stats, NodeStat{Node: n, Count: c})
	}
	sort.Slice(stats, func(i, j int) bool { return stats[i].Count > stats[j].Count })
	return stats
}

// NodeStat holds a node's key count.
type NodeStat struct {
	Node  string
	Count int
}
