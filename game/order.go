package game

import "sort"

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
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
