package ui

import (
	"maps"
	"slices"
)

// sortedKeysOf returns m's keys in order, so lists built from game maps come
// out the same on every call.
func sortedKeysOf[V any](m map[string]V) []string {
	return slices.Sorted(maps.Keys(m))
}

// keysWhere is the keys of m that keep accepts, sorted.
func keysWhere[V any](m map[string]V, keep func(key string, v V) bool) []string {
	var keys []string
	for k, v := range m {
		if keep(k, v) {
			keys = append(keys, k)
		}
	}
	slices.Sort(keys)
	return keys
}
