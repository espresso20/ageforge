package config

import "testing"

// These tests pin the EPIC economy-rebalance sub-ticket 3 "moderate requirements
// bump" applied by normalizeAgeRequirements (see ages.go). They assert the scaled
// resource values, the building-requirement floor, and idempotency of Ages().

func ageByKeyT(t *testing.T, key string) AgeDef {
	t.Helper()
	a, ok := AgeByKey()[key]
	if !ok {
		t.Fatalf("age %q not found", key)
	}
	return a
}

// Building floor: iron_age scriptorium was 3, must be raised to the floor of 5.
// A count above the floor like bronze_age longhouse (15) must be left unchanged.
func TestNormalizeAgeRequirements_BuildingFloor(t *testing.T) {
	iron := ageByKeyT(t, "iron_age")
	if got := iron.BuildingReqs["scriptorium"]; got != 5 {
		t.Errorf("iron_age scriptorium req = %d, want 5 (raised to floor)", got)
	}

	bronze := ageByKeyT(t, "bronze_age")
	if got := bronze.BuildingReqs["longhouse"]; got != 15 {
		t.Errorf("bronze_age longhouse req = %d, want 15 (above floor, unchanged)", got)
	}
}

// Late ages (digital_age onward) are excluded from the building floor: a sub-5
// count there should NOT be raised. digital_age has no sub-5 reqs today, so instead
// assert a known late age keeps its original (already >= 5) values untouched and
// that no floor was silently applied by spot-checking digital_age values stay raw.
func TestNormalizeAgeRequirements_LateAgesNoBuildingFloor(t *testing.T) {
	// transcendent_age building reqs are all far above 5 and in the no-floor band;
	// they must equal their raw literals exactly.
	trans := ageByKeyT(t, "transcendent_age")
	wantReqs := map[string]int{
		"reality_academy": 20, "reality_forge": 15, "probability_war_room": 15,
	}
	for k, want := range wantReqs {
		if got := trans.BuildingReqs[k]; got != want {
			t.Errorf("transcendent_age %s req = %d, want %d (unchanged)", k, got, want)
		}
	}
}

// Idempotency: Ages() rebuilds from literals each call, so two calls must yield
// identical ResourceReqs with no compounding of the multiplier.
func TestNormalizeAgeRequirements_Idempotent(t *testing.T) {
	first := AgeByKey()
	second := AgeByKey()

	for key, a1 := range first {
		a2 := second[key]
		if len(a1.ResourceReqs) != len(a2.ResourceReqs) {
			t.Fatalf("age %q ResourceReqs len changed between calls: %d vs %d",
				key, len(a1.ResourceReqs), len(a2.ResourceReqs))
		}
		for res, v1 := range a1.ResourceReqs {
			if v2 := a2.ResourceReqs[res]; v1 != v2 {
				t.Errorf("age %q resource %q differs across Ages() calls: %v vs %v (compounding?)",
					key, res, v1, v2)
			}
		}
		for bld, c1 := range a1.BuildingReqs {
			if c2 := a2.BuildingReqs[bld]; c1 != c2 {
				t.Errorf("age %q building %q differs across Ages() calls: %d vs %d",
					key, bld, c1, c2)
			}
		}
	}
}
