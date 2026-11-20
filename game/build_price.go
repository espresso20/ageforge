package game

import "github.com/espresso20/ageforge/detmath"

// maxBuildBatch bounds how many copies MaxBuildable counts: the same bound
// the build command's "max" uses.
const maxBuildBatch = 10000

// queuedOf is how many copies of key the snapshot's build queue holds.
func (s GameState) queuedOf(key string) int {
	n := 0
	for _, q := range s.BuildQueue {
		if q.Key == key {
			n++
		}
	}
	return n
}

// BuildBatch is the price of the next n copies of the building key, as the
// engine charges them: each copy is priced on the copies built and queued
// before it and floored on its own, so a batch equals the sum of n single
// builds. ok is false for a key the snapshot's rules do not know.
func (s GameState) BuildBatch(key string, n int) (cost map[string]float64, ok bool) {
	def, ok := s.Ruleset().Building(key)
	if !ok {
		return nil, false
	}
	bs := s.Buildings[key]
	first := bs.Count + s.queuedOf(key)
	cost = make(map[string]float64, len(def.BaseCost))
	for i := 0; i < n; i++ {
		exp := float64(first + i)
		for res, base := range def.BaseCost {
			cost[res] += applyCostMult(base*detmath.Pow(def.CostScale, exp), s.BuildCostFactor)
		}
	}
	return cost, true
}

// MaxBuildable is how many copies of the building key the held resources pay
// for now, within the building's count limit and at most maxBuildBatch: what
// the build command's "max" builds. 0 when not even one is affordable.
func (s GameState) MaxBuildable(key string) int {
	def, ok := s.Ruleset().Building(key)
	if !ok {
		return 0
	}
	bs := s.Buildings[key]
	first := bs.Count + s.queuedOf(key)
	limit := maxBuildBatch
	if def.MaxCount > 0 {
		limit = min(limit, def.MaxCount-first)
	}
	spent := make(map[string]float64, len(def.BaseCost))
	n := 0
	for ; n < limit; n++ {
		exp := float64(first + n)
		next := make(map[string]float64, len(def.BaseCost))
		for res, base := range def.BaseCost {
			next[res] = spent[res] + applyCostMult(base*detmath.Pow(def.CostScale, exp), s.BuildCostFactor)
			if next[res] > s.Resources[res].Amount {
				return n
			}
		}
		spent = next
	}
	return n
}

// CanGatherByHand reports whether the age the snapshot is in still allows
// the gather command: hand-gathering ends after the Medieval Age.
func (s GameState) CanGatherByHand() bool {
	return !pastMedievalForGather(s.Ruleset(), s.Age)
}
