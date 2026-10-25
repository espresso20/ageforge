package config

import (
	"strings"
	"testing"
)

// TestBuildingFlavorStringsAreNotBlank verifies every flavor line is real
// text. A key with no matching building is caught by
// TestBuildingFlavorAppliesToDefs, and the style rules (no tview tags, no
// printf directives) by TestConfigPlayerTextStyle.
func TestBuildingFlavorStringsAreNotBlank(t *testing.T) {
	for key, flavor := range buildingFlavor {
		if strings.TrimSpace(flavor) == "" {
			t.Errorf("building %q has an empty/blank flavor string", key)
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

// TestAllAgesHaveQuips verifies every age carries a non-empty one-liner for the
// age-advance splash (TestConfigPlayerTextStyle skips an empty one).
func TestAllAgesHaveQuips(t *testing.T) {
	for _, a := range Ages() {
		if strings.TrimSpace(a.Quip) == "" {
			t.Errorf("age %q (%s) has an empty Quip", a.Key, a.Name)
		}
	}
}
