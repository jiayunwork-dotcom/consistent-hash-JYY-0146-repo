// Package hash provides deterministic string hashing for the consistent-hash
// ring. It uses FNV-1a so behaviour is stable across platforms.
package hash

import "hash/fnv"

// HashString returns a deterministic 32-bit FNV-1a hash of s.
func HashString(s string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return h.Sum32()
}

// HashVirtual returns the hash for the i-th virtual node of a physical node.
// It combines "node#i" into a single string before hashing so that distinct
// virtual indices map to distinct (and well distributed) ring positions.
func HashVirtual(node string, i int) uint32 {
	return HashString(node + "#" + itoa(i))
}

// itoa is a tiny decimal formatter to avoid importing strconv in this leaf
// package (it keeps the dependency surface minimal and obvious).
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

// HashBytes returns a deterministic 32-bit FNV-1a hash of b.
func HashBytes(b []byte) uint32 {
	h := fnv.New32a()
	_, _ = h.Write(b)
	return h.Sum32()
}

// VirtualPositions returns all ring positions for a node with the given
// replica count. This is useful for debugging ring distribution.
func VirtualPositions(node string, replicas int) []uint32 {
	positions := make([]uint32, replicas)
	for i := 0; i < replicas; i++ {
		positions[i] = HashVirtual(node, i)
	}
	return positions
}

// Spread measures how uniformly positions are distributed across the uint32
// space. It returns the standard deviation of gaps between adjacent positions
// (sorted). A lower value indicates better spread.
func Spread(positions []uint32) float64 {
	if len(positions) < 2 {
		return 0
	}
	sorted := make([]uint32, len(positions))
	copy(sorted, positions)
	sortUint32(sorted)

	gaps := make([]float64, len(sorted))
	for i := 1; i < len(sorted); i++ {
		gaps[i-1] = float64(sorted[i] - sorted[i-1])
	}
	// wrap-around gap
	gaps[len(sorted)-1] = float64(^uint32(0)-sorted[len(sorted)-1]) + float64(sorted[0]) + 1

	mean := float64(^uint32(0)) / float64(len(sorted))
	var sumSq float64
	for _, g := range gaps {
		diff := g - mean
		sumSq += diff * diff
	}
	return sumSq / float64(len(gaps))
}

func sortUint32(s []uint32) {
	// simple insertion sort (adequate for small arrays in tests)
	for i := 1; i < len(s); i++ {
		key := s[i]
		j := i - 1
		for j >= 0 && s[j] > key {
			s[j+1] = s[j]
			j--
		}
		s[j+1] = key
	}
}
