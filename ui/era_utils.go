package ui

// hashKey returns a deterministic uint64 hash for a string (FNV-1a).
func hashKey(key string) uint64 {
	h := uint64(14695981039346656037)
	for _, c := range []byte(key) {
		h ^= uint64(c)
		h *= 1099511628211
	}
	return h
}
