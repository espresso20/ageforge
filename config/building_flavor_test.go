package config

import (
	"strings"
	"testing"
)

// TestBuildingFlavorKeysAreValid ensures every key in the buildingFlavor map
// corresponds to a real building. A typo'd key would silently render no flavor
// and rot unnoticed, so we fail loudly here instead.
func TestBuildingFlavorKeysAreValid(t *testing.T) {
	defs := BaseBuildings()
	valid := make(map[string]bool, len(defs))
	for _, d := range defs {
		valid[d.Key] = true
	}
	for key := range buildingFlavor {
		if !valid[key] {
			t.Errorf("buildingFlavor has key %q with no matching building", key)
		}
	}
}

// TestBuildingFlavorStringsAreSane verifies every flavor line is plain, safe
// text: non-empty, no tview color tags (which would corrupt rendering), and no
// stray printf directives (the strings are rendered through Fprintf with a %s).
func TestBuildingFlavorStringsAreSane(t *testing.T) {
	for key, flavor := range buildingFlavor {
		if strings.TrimSpace(flavor) == "" {
			t.Errorf("building %q has an empty/blank flavor string", key)
		}
		// tview color/style tags use square brackets; a literal bracket in flavor
		// would either be swallowed or mis-parsed by the dynamic-color renderer.
		if strings.ContainsAny(flavor, "[]") {
			t.Errorf("building %q flavor contains a square bracket (tview tag risk): %q", key, flavor)
		}
		// Flavor is interpolated via fmt.Fprintf(..., "%s", flavor); a stray %
		// directive would be harmless there but signals a copy-paste mistake.
		if strings.Contains(flavor, "%") {
			t.Errorf("building %q flavor contains a %% directive: %q", key, flavor)
		}
	}
}

// TestBuildingFlavorAppliesToDefs confirms applyBuildingFlavor actually stamps
// the Flavor field onto the assembled BuildingDefs (the chokepoint wiring), and
// that the count of flavored buildings matches the map size.
func TestBuildingFlavorAppliesToDefs(t *testing.T) {
	defs := BaseBuildings()
	got := 0
	for _, d := range defs {
		if d.Flavor != "" {
			got++
			// Spot-check the field carries exactly the mapped string.
			if want, ok := buildingFlavor[d.Key]; ok && d.Flavor != want {
				t.Errorf("building %q flavor = %q, want %q", d.Key, d.Flavor, want)
			}
		}
	}
	if got != len(buildingFlavor) {
		t.Errorf("flavored buildings = %d, want %d (map size); some keys may be unmatched", got, len(buildingFlavor))
	}
}

// TestAllAgesHaveQuips verifies every age carries a non-empty, sane one-liner for
// the age-advance splash. The splash renders Quip via Fprintf with %s, so apply
// the same bracket/directive guards as the building flavor.
func TestAllAgesHaveQuips(t *testing.T) {
	ages := Ages()
	if len(ages) != 22 {
		t.Fatalf("expected 22 ages, got %d", len(ages))
	}
	for _, a := range ages {
		if strings.TrimSpace(a.Quip) == "" {
			t.Errorf("age %q (%s) has an empty Quip", a.Key, a.Name)
		}
		if strings.ContainsAny(a.Quip, "[]") {
			t.Errorf("age %q Quip contains a square bracket (tview tag risk): %q", a.Key, a.Quip)
		}
		if strings.Contains(a.Quip, "%") {
			t.Errorf("age %q Quip contains a %% directive: %q", a.Key, a.Quip)
		}
	}
}
