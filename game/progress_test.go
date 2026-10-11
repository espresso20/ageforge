package game

import (
	"testing"
)

func TestProgressManager_AgeOrder(t *testing.T) {
	pm := NewProgressManager()
	order := pm.GetAgeOrder()

	if order["primitive_age"] != 0 {
		t.Errorf("primitive_age order = %v, want 0", order["primitive_age"])
	}
	if order["stone_age"] <= order["primitive_age"] {
		t.Error("stone_age should come after primitive_age")
	}
}

func TestProgressManager_GetNextAge(t *testing.T) {
	pm := NewProgressManager()

	next := pm.GetNextAge("primitive_age")
	if next != "stone_age" {
		t.Errorf("next after primitive = %v, want stone_age", next)
	}

	// Last age should return ""
	next = pm.GetNextAge("transcendent_age")
	if next != "" {
		t.Errorf("next after transcendent = %v, want empty", next)
	}
}

func TestProgressManager_GetAgeName(t *testing.T) {
	pm := NewProgressManager()

	name := pm.GetAgeName("primitive_age")
	if name == "" || name == "primitive_age" {
		t.Errorf("GetAgeName should return display name, got %q", name)
	}
}

func TestProgressManager_CheckAdvancement(t *testing.T) {
	pm := NewProgressManager()
	bm := NewBuildingManager()

	// Should not advance with nothing
	next := pm.CheckAdvancement("primitive_age", bm)
	if next != "" {
		t.Errorf("should not advance with no buildings, got %v", next)
	}
}

func TestProgressManager_GetRequirements(t *testing.T) {
	pm := NewProgressManager()

	if bldReqs := pm.GetRequirementsForNext("primitive_age"); len(bldReqs) == 0 {
		t.Error("next age should have some requirements")
	}
}

// A gate asks for buildings and the wonder, and no amount of any resource:
// with empty stores, the buildings and the wonder in place open it, and the
// check never reads a store.
func TestProgressManager_GateAsksForNoResources(t *testing.T) {
	pm := NewProgressManager()
	bm := NewBuildingManager()
	for key, n := range pm.GetUnlocks(pm.GetNextAge("primitive_age")).BuildingReqs {
		bm.counts[key] = n
	}
	if next := pm.CheckAdvancement("primitive_age", bm); next != "" {
		t.Fatalf("the gate opened without the wonder: %q", next)
	}
	bm.counts[pm.WonderForAge("primitive_age")] = 1
	if next := pm.CheckAdvancement("primitive_age", bm); next != "stone_age" {
		t.Errorf("buildings and wonder in place, empty stores: gate leads to %q, want stone_age", next)
	}
}
