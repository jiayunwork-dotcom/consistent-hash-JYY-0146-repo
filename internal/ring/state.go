package ring

// Members returns all physical node names currently in the ring.
func (r *Ring) Members() []string {
	names := make([]string, 0, len(r.members))
	for n := range r.members {
		names = append(names, n)
	}
	return names
}

// Replicas returns the configured number of virtual nodes per physical node.
func (r *Ring) Replicas() int {
	return r.replicas
}
