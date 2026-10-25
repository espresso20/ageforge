package config

import "testing"

// Idempotency: Ages() rebuilds from literals each call, so two calls must yield
// identical BuildingReqs with no compounding of the multiplier.
func TestNormalizeAgeRequirements_Idempotent(t *testing.T) {
	first := AgeByKey()
	second := AgeByKey()

	for key, a1 := range first {
		a2 := second[key]
		for bld, c1 := range a1.BuildingReqs {
			if c2 := a2.BuildingReqs[bld]; c1 != c2 {
				t.Errorf("age %q building %q differs across Ages() calls: %d vs %d",
					key, bld, c1, c2)
			}
		}
	}
}
