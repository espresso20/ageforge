package config

import (
	"testing"

	"github.com/espresso20/ageforge/detmath"
)

// townWait is the wait to afford copy n of a building that makes the same
// amount in every copy and climbs at rate, up to a constant: the price
// rate^(n-1) over the n-1 copies already making. Defined for n >= 2.
func townWait(rate float64, n int) float64 {
	return detmath.Pow(rate, float64(n-1)) / float64(n-1)
}

// TestTownCopiesIsTheShortestWait: TownCopies is the whole number of copies
// at which the wait to afford the next copy is smallest, worked out from the
// rate producers climb at.
func TestTownCopiesIsTheShortestWait(t *testing.T) {
	best, bestWait := 0, 0.0
	for n := 2; n <= 60; n++ {
		if w := townWait(ProducerRate, n); best == 0 || w < bestWait {
			best, bestWait = n, w
		}
	}
	if best != TownCopies {
		t.Errorf("the wait to afford a copy at %g a copy is smallest at copy %d, TownCopies is %d", ProducerRate, best, TownCopies)
	}
	// The wait falls to TownCopies and rises after it.
	if townWait(ProducerRate, TownCopies+1) <= townWait(ProducerRate, TownCopies) {
		t.Errorf("copy %d is not a longer wait than copy %d", TownCopies+1, TownCopies)
	}
	if townWait(ProducerRate, TownCopies-1) <= townWait(ProducerRate, TownCopies) {
		t.Errorf("copy %d is not a longer wait than copy %d", TownCopies-1, TownCopies)
	}
}

// TestEveryProducerClimbsAtTheProducerRate: ProducerRate is the rate of every
// building that makes something, so the derivation above is about the
// buildings the gate rule asks TownCopies of.
func TestEveryProducerClimbsAtTheProducerRate(t *testing.T) {
	seen := 0
	for _, d := range BaseBuildings() {
		if d.Category == "wonder" || !makesSomething(d) {
			continue
		}
		seen++
		if d.CostScale != ProducerRate {
			t.Errorf("%s makes something and climbs at %g a copy, the producers' rate is %g", d.Key, d.CostScale, ProducerRate)
		}
	}
	if seen == 0 {
		t.Fatal("no producers found")
	}
}
