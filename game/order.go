package game

import (
	"maps"
	"slices"
)

// sortedKeys returns m's keys in ascending order.
//
// Go randomises map iteration order on every range, so a loop over a map must
// not let that order reach anything a seed is supposed to pin down: a random
// draw, the order of log lines or queued effects, a tie-break, or a float sum
// into a shared total (float addition is not associative, so summing the same
// values in a different order can change the last bit, and a last-bit change
// in a rate can move a threshold crossing by a tick). Such loops range over
// sortedKeys(m) instead. Loops whose result does not depend on order (integer
// counts, per-key copies, lookups) can keep ranging over the map directly.
func sortedKeys[V any](m map[string]V) []string {
	keys := slices.AppendSeq(make([]string, 0, len(m)), maps.Keys(m))
	slices.Sort(keys)
	return keys
}

// cloneMap copies m so the copy shares nothing with it; an empty map reads as
// nil, so a field tagged omitempty leaves it out of a save.
func cloneMap[K comparable, V any](m map[K]V) map[K]V {
	if len(m) == 0 {
		return nil
	}
	return maps.Clone(m)
}

// cloneSlice is cloneMap for a slice.
func cloneSlice[E any](s []E) []E {
	if len(s) == 0 {
		return nil
	}
	return slices.Clone(s)
}
