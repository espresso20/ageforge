package game

import (
	"math"
	"testing"
)

func TestBuildingManager_UnlockAndCount(t *testing.T) {
	bm := NewBuildingManager()

	if bm.IsUnlocked("hut") {
		t.Error("hut should not be unlocked initially")
	}

	bm.UnlockBuilding("hut")
	if !bm.IsUnlocked("hut") {
		t.Error("hut should be unlocked after UnlockBuilding")
	}
	if got := bm.GetCount("hut"); got != 0 {
		t.Errorf("hut count = %v, want 0", got)
	}
}

func TestBuildingManager_CostScaling(t *testing.T) {
	bm := NewBuildingManager()
	bm.UnlockBuilding("hut")

	cost0 := bm.GetCost("hut")
	if cost0 == nil {
		t.Fatal("GetCost returned nil for unlocked hut")
	}
	baseWood := cost0["wood"]
	if baseWood <= 0 {
		t.Fatalf("hut base wood cost = %v, expected > 0", baseWood)
	}

	// Simulate building one hut
	bm.counts["hut"] = 1

	cost1 := bm.GetCost("hut")
	scaledWood := cost1["wood"]

	def := bm.defs["hut"]
	expected := math.Floor(def.BaseCost["wood"] * math.Pow(def.CostScale, 1))
	if scaledWood != expected {
		t.Errorf("scaled cost = %v, want %v (base=%v, scale=%v)",
			scaledWood, expected, def.BaseCost["wood"], def.CostScale)
	}

	// Cost should always increase
	if scaledWood <= baseWood {
		t.Errorf("cost should increase: base=%v, scaled=%v", baseWood, scaledWood)
	}
}

func TestBuildingManager_PopCapacity(t *testing.T) {
	bm := NewBuildingManager()
	bm.UnlockBuilding("hut")

	if got := bm.GetPopCapacity(); got != 0 {
		t.Errorf("pop cap with no buildings = %v, want 0", got)
	}

	bm.counts["hut"] = 3
	cap := bm.GetPopCapacity()
	if cap <= 0 {
		t.Errorf("pop cap with 3 huts = %v, want > 0", cap)
	}
}

func TestBuildingManager_GetAll(t *testing.T) {
	bm := NewBuildingManager()
	bm.UnlockBuilding("hut")
	bm.counts["hut"] = 5

	all := bm.GetAll()
	if all["hut"] != 5 {
		t.Errorf("GetAll[hut] = %v, want 5", all["hut"])
	}
}

func TestBuildingManager_LoadCounts(t *testing.T) {
	bm := NewBuildingManager()
	bm.LoadCounts(map[string]int{"hut": 3, "farm": 2})

	if bm.GetCount("hut") != 3 {
		t.Errorf("loaded hut count = %v, want 3", bm.GetCount("hut"))
	}
	if bm.GetCount("farm") != 2 {
		t.Errorf("loaded farm count = %v, want 2", bm.GetCount("farm"))
	}
}
