package config

import (
	"math"
	"testing"
)

// TestIncomeFactorAppliesThePoolInFull: the pacing model multiplies income by
// the whole all-production pool a typical player holds, however large, as
// the engine does, and by the techs' layer on top.
func TestIncomeFactorAppliesThePoolInFull(t *testing.T) {
	order := AgeOrder()
	pos := AgePositions(order)
	layer := func(age string) float64 {
		l := 1.0
		for _, tech := range Technologies() {
			if j, ok := pos[tech.Age]; !ok || j > pos[age] {
				continue
			}
			for _, e := range tech.Effects {
				if e.Kind == EffectAllOutput {
					l += e.Value
				}
			}
		}
		return l
	}
	// "zz_none" is a resource no tech names: the factor is the pool's
	// multiplier times the techs' all-production layer.
	big := 0
	for _, age := range order {
		held := ProductionAllHeld[age]
		want := float64((1 + held) * layer(age))
		if got := IncomeFactor(Technologies(), pos, age, "zz_none"); math.Abs(got-want) > 1e-12 {
			t.Errorf("%s: the model holds %+.2f and multiplies income by %v, want %v", age, held, got, want)
		}
		if held > 2 {
			big++
		}
	}
	if big == 0 {
		t.Error("no age holds more than +200%: the check that a large pool applies in full proved nothing")
	}
	for i := 1; i < len(order); i++ {
		if ProductionAllHeld[order[i]] < ProductionAllHeld[order[i-1]] {
			t.Errorf("%s holds less all production than %s", order[i], order[i-1])
		}
	}
}
