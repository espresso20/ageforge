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

// TestBuildingManager_BuildBatchCost_NoBuildingsQueued checks that BuildBatchCost
// with no prior buildings/queue returns the correct cumulative total for N units.
// Hut (post cost-curve rebalance): BaseCost={"wood":14}, CostScale=1.13.
// For 5 huts from scratch: cost_i = floor(14 * 1.13^i), i=0..4
// = 14 + 15 + 17 + 20 + 22 = 88.
// (Assertion below derives the expected value from the live def, so it stays
// correct even if these numbers are re-tuned again.)
func TestBuildingManager_BuildBatchCost_NoBuildingsQueued(t *testing.T) {
	bm := NewBuildingManager()
	bm.UnlockBuilding("hut")

	def := bm.defs["hut"]
	base := def.BaseCost["wood"]
	scale := def.CostScale

	expected := 0.0
	for i := 0; i < 5; i++ {
		expected += math.Floor(base * math.Pow(scale, float64(i)))
	}

	cost, ok := bm.BuildBatchCost("hut", 5, nil)
	if !ok {
		t.Fatal("BuildBatchCost returned false for known building")
	}
	if cost["wood"] != expected {
		t.Errorf("batch cost wood = %v, want %v", cost["wood"], expected)
	}
}

// TestBuildingManager_BuildBatchCost_WithBuiltAndQueued verifies that
// already-built and already-queued counts shift the exponent correctly.
// With 2 built and 1 queued, the next unit should cost floor(15 * 1.12^3).
func TestBuildingManager_BuildBatchCost_WithBuiltAndQueued(t *testing.T) {
	bm := NewBuildingManager()
	bm.UnlockBuilding("hut")
	bm.counts["hut"] = 2

	queue := []BuildQueueItem{
		{BuildingKey: "hut", TicksLeft: 5, TotalTicks: 8},
	}

	def := bm.defs["hut"]
	expected := math.Floor(def.BaseCost["wood"] * math.Pow(def.CostScale, 3))

	cost, ok := bm.BuildBatchCost("hut", 1, queue)
	if !ok {
		t.Fatal("BuildBatchCost returned false for known building")
	}
	if cost["wood"] != expected {
		t.Errorf("queue-aware cost wood = %v, want %v (scale^3)", cost["wood"], expected)
	}
}

// TestBuildingManager_GetQueueCount verifies queue counting.
func TestBuildingManager_GetQueueCount(t *testing.T) {
	bm := NewBuildingManager()
	queue := []BuildQueueItem{
		{BuildingKey: "hut"},
		{BuildingKey: "hut"},
		{BuildingKey: "longhouse"},
	}
	if got := bm.GetQueueCount("hut", queue); got != 2 {
		t.Errorf("GetQueueCount(hut) = %d, want 2", got)
	}
	if got := bm.GetQueueCount("longhouse", queue); got != 1 {
		t.Errorf("GetQueueCount(longhouse) = %d, want 1", got)
	}
	if got := bm.GetQueueCount("hut", nil); got != 0 {
		t.Errorf("GetQueueCount with nil queue = %d, want 0", got)
	}
}
