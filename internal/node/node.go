// Package node provides helpers for normalizing, deduplicating and validating
// the set of physical node names used by the consistent-hash ring.
package node

import (
	"errors"
	"strings"
)

// ErrEmptyName is returned when a node name is empty after normalization.
var ErrEmptyName = errors.New("node: empty name")

// Normalize trims surrounding whitespace and lower-cases a node name so that
// " Node-A " and "node-a" are treated as the same node.
func Normalize(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

// Validate reports whether a single (already normalized) name is usable.
// Names cannot be empty or contain the '#' character (reserved for virtual
// node key construction).
func Validate(name string) error {
	if name == "" {
		return ErrEmptyName
	}
	if strings.ContainsRune(name, '#') {
		return ErrInvalidChar
	}
	return nil
}

// ErrInvalidChar is returned when a node name contains a reserved character.
var ErrInvalidChar = errors.New("node: name contains reserved character '#'")

// NormalizeAll normalizes each input name, drops empties and returns the
// surviving names with duplicates removed (order preserved, first wins).
func NormalizeAll(names []string) []string {
	seen := make(map[string]struct{}, len(names))
	out := make([]string, 0, len(names))
	for _, n := range names {
		nn := Normalize(n)
		if nn == "" {
			continue
		}
		if _, ok := seen[nn]; ok {
			continue
		}
		seen[nn] = struct{}{}
		out = append(out, nn)
	}
	return out
}

// Sort sorts a slice of node names in-place.
func Sort(names []string) {
	for i := 1; i < len(names); i++ {
		key := names[i]
		j := i - 1
		for j >= 0 && names[j] > key {
			names[j+1] = names[j]
			j--
		}
		names[j+1] = key
	}
}

// Contains reports whether names contains target.
func Contains(names []string, target string) bool {
	for _, n := range names {
		if n == target {
			return true
		}
	}
	return false
}

// Diff returns elements in a that are not in b.
func Diff(a, b []string) []string {
	set := make(map[string]struct{}, len(b))
	for _, x := range b {
		set[x] = struct{}{}
	}
	var result []string
	for _, x := range a {
		if _, ok := set[x]; !ok {
			result = append(result, x)
		}
	}
	return result
}
