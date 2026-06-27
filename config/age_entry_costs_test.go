package config

import "testing"

// TestAgeEntryCosts_StoneAge verifies AgeEntryCosts returns the minimum base
// cost of each resource among the non-wonder buildings unlocked in stone_age.
func TestAgeEntryCosts_StoneAge(t *testing.T) {
	got := AgeEntryCosts("stone_age")
	if len(got) == 0 {
		t.Fatal("AgeEntryCosts(\"stone_age\") returned empty map")
	}

	// Independently compute the expected minimums from BaseBuildings().
	want := make(map[string]float64)
	for _, b := range BaseBuildings() {
		if b.RequiredAge != "stone_age" || b.Category == "wonder" {
			continue
		}
		for res, amt := range b.BaseCost {
			if amt <= 0 {
				continue
			}
			if cur, ok := want[res]; !ok || amt < cur {
				want[res] = amt
			}
		}
	}

	if len(got) != len(want) {
		t.Errorf("resource-count mismatch: got %d (%v), want %d (%v)", len(got), got, len(want), want)
	}
	for res, w := range want {
		g, ok := got[res]
		if !ok {
			t.Errorf("missing resource %q in result", res)
			continue
		}
		if g != w {
			t.Errorf("resource %q: got %.2f, want min %.2f", res, g, w)
		}
		if g <= 0 {
			t.Errorf("resource %q has non-positive cost %.2f", res, g)
		}
	}
}

// TestAgeEntryCosts_ExcludesWonders asserts no wonder-only resource leaks into
// the result and that a known cheap building drives the minimum.
func TestAgeEntryCosts_ExcludesWonders(t *testing.T) {
	got := AgeEntryCosts("stone_age")
	// stone_age wood floor comes from a non-wonder building; ensure it is sane
	// and not dominated by a (much costlier) wonder's cost.
	wood, ok := got["wood"]
	if !ok {
		t.Fatal("expected wood entry cost for stone_age")
	}
	for _, b := range BaseBuildings() {
		if b.RequiredAge == "stone_age" && b.Category == "wonder" {
			if wc, has := b.BaseCost["wood"]; has && wc < wood {
				t.Errorf("wonder %q wood cost %.2f is below reported min %.2f — wonder not excluded", b.Key, wc, wood)
			}
		}
	}
}
