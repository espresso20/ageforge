package ui

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// TestBuildingsProviderRendersFlavor verifies the buildings overlay emits a
// building's cosmetic Flavor line (when present) and does not panic on a
// snapshot that includes flavored buildings. It uses a real primitive-age
// building so the assertion tracks the actual content.
func TestBuildingsProviderRendersFlavor(t *testing.T) {
	defs := config.BuildingByKey()
	hutDef, ok := defs["hut"]
	if !ok || hutDef.Flavor == "" {
		t.Fatal("expected 'hut' to exist with a non-empty Flavor")
	}

	state := game.GameState{
		Age: "primitive_age",
		Buildings: map[string]game.BuildingState{
			"hut": {
				Name:        hutDef.Name,
				Category:    hutDef.Category,
				Description: hutDef.Description,
				Flavor:      hutDef.Flavor,
				Unlocked:    true,
				AgeKey:      "primitive_age",
				Count:       1,
			},
		},
	}

	out := buildingsProvider(state, 80)
	if out == "" {
		t.Fatal("buildingsProvider returned empty output")
	}
	if !strings.Contains(out, hutDef.Flavor) {
		t.Errorf("buildings overlay missing hut flavor line %q\n--- output ---\n%s", hutDef.Flavor, out)
	}
	// The functional description must still be present — flavor is additive.
	if !strings.Contains(out, hutDef.Description) {
		t.Errorf("buildings overlay dropped functional description %q", hutDef.Description)
	}
}

// TestBuildingsProviderNoFlavorIsSafe verifies a building with an empty Flavor
// renders without emitting a stray flavor line or panicking.
func TestBuildingsProviderNoFlavorIsSafe(t *testing.T) {
	state := game.GameState{
		Age: "primitive_age",
		Buildings: map[string]game.BuildingState{
			"plain_thing": {
				Name:        "Plain Thing",
				Category:    "production",
				Description: "Does a thing.",
				Flavor:      "", // no flavor
				Unlocked:    true,
				AgeKey:      "primitive_age",
				Count:       0,
			},
		},
	}
	out := buildingsProvider(state, 80)
	if !strings.Contains(out, "Does a thing.") {
		t.Errorf("expected functional description in output, got:\n%s", out)
	}
}

// TestAgeSplashIncludesQuip verifies the assembled age-advance splash text
// contains the age's Quip line beneath the formal description, for every age.
func TestAgeSplashIncludesQuip(t *testing.T) {
	for _, a := range config.Ages() {
		text := buildAgeSplashText(a.Key, game.AgeAdvanceSummary{}, false, game.EpochEventRecord{})
		if text == "" {
			t.Fatalf("age %q produced empty splash text", a.Key)
		}
		if !strings.Contains(text, a.Quip) {
			t.Errorf("age %q splash missing quip %q\n--- splash ---\n%s", a.Key, a.Quip, text)
		}
		// The formal description must remain present too.
		if !strings.Contains(text, a.Description) {
			t.Errorf("age %q splash missing formal description %q", a.Key, a.Description)
		}
	}
}
