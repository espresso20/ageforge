package ui

import (
	"testing"

	"github.com/espresso20/ageforge/game"
)

func TestResourceRowsFor_ExcludesEndedResources(t *testing.T) {
	st := game.GameState{
		Age: "space_age",
		Resources: map[string]game.ResourceState{
			"food":   {Name: "Food", Amount: 1000, Unlocked: true},
			"wood":   {Name: "Wood", Amount: 500, Unlocked: true},
			"crypto": {Name: "Crypto", Amount: 200, Unlocked: true},
			"plasma": {Name: "Plasma", Amount: 100, Unlocked: true},
		},
	}

	rows := resourceRowsFor(st, nil, false)
	names := make(map[string]bool)
	for _, r := range rows {
		names[r.name] = true
	}

	if !names["Food"] {
		t.Errorf("expected Food in rows, but was missing")
	}
	if !names["Plasma"] {
		t.Errorf("expected Plasma in rows, but was missing")
	}
	if names["Wood"] {
		t.Errorf("wood has ended in space_age and should be excluded from rows")
	}
	if names["Crypto"] {
		t.Errorf("crypto has ended in space_age and should be excluded from rows")
	}
}

func TestUnlockedResourceKeys_ExcludesEndedResources(t *testing.T) {
	st := game.GameState{
		Age: "space_age",
		Resources: map[string]game.ResourceState{
			"food":   {Name: "Food", Amount: 1000, Unlocked: true},
			"wood":   {Name: "Wood", Amount: 500, Unlocked: true},
			"crypto": {Name: "Crypto", Amount: 200, Unlocked: true},
			"plasma": {Name: "Plasma", Amount: 100, Unlocked: true},
		},
	}

	keys := unlockedResourceKeys(st)
	has := func(target string) bool {
		for _, k := range keys {
			if k == target {
				return true
			}
		}
		return false
	}

	if !has("food") || !has("plasma") {
		t.Errorf("expected food and plasma in unlockedResourceKeys, got %v", keys)
	}
	if has("wood") || has("crypto") {
		t.Errorf("expected wood and crypto to be excluded in space_age, got %v", keys)
	}
}
