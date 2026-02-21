package game

import (
	"testing"
)

// fullAgeOrder returns a complete age order map for tests
func fullAgeOrder() map[string]int {
	pm := NewProgressManager()
	return pm.GetAgeOrder()
}

func TestMilestoneManager_CheckFirstShelter(t *testing.T) {
	mm := NewMilestoneManager()
	rm := NewResourceManager()
	bm := NewBuildingManager()
	bm.UnlockBuilding("hut")

	ageOrder := fullAgeOrder()

	// No hut — should not complete
	completed := mm.CheckMilestones(1, "primitive_age", ageOrder, rm, bm, 0, 0, 0, nil, 0, 0)
	if len(completed) != 0 {
		t.Errorf("expected 0 completions with no hut, got %d", len(completed))
	}

	// Build a hut
	bm.counts["hut"] = 1
	completed = mm.CheckMilestones(2, "primitive_age", ageOrder, rm, bm, 0, 0, 0, nil, 0, 0)

	found := false
	for _, ms := range completed {
		if ms.Key == "first_shelter" {
			found = true
		}
	}
	if !found {
		t.Error("first_shelter should complete when hut count >= 1")
	}

	// Should not trigger again
	completed = mm.CheckMilestones(3, "primitive_age", ageOrder, rm, bm, 0, 0, 0, nil, 0, 0)
	for _, ms := range completed {
		if ms.Key == "first_shelter" {
			t.Error("first_shelter should not trigger twice")
		}
	}
}

func TestMilestoneManager_PopulationMilestone(t *testing.T) {
	mm := NewMilestoneManager()
	rm := NewResourceManager()
	bm := NewBuildingManager()
	ageOrder := fullAgeOrder()

	// small_village requires pop 5
	completed := mm.CheckMilestones(1, "primitive_age", ageOrder, rm, bm, 4, 0, 0, nil, 0, 0)
	for _, ms := range completed {
		if ms.Key == "small_village" {
			t.Error("small_village should not trigger at pop 4")
		}
	}

	completed = mm.CheckMilestones(2, "primitive_age", ageOrder, rm, bm, 5, 0, 0, nil, 0, 0)
	found := false
	for _, ms := range completed {
		if ms.Key == "small_village" {
			found = true
		}
	}
	if !found {
		t.Error("small_village should trigger at pop 5")
	}
}

func TestMilestoneManager_AgeMilestone(t *testing.T) {
	mm := NewMilestoneManager()
	rm := NewResourceManager()
	bm := NewBuildingManager()
	ageOrder := fullAgeOrder()

	// bronze_pioneer requires bronze_age
	completed := mm.CheckMilestones(1, "stone_age", ageOrder, rm, bm, 0, 0, 0, nil, 0, 0)
	for _, ms := range completed {
		if ms.Key == "bronze_pioneer" {
			t.Error("bronze_pioneer should not trigger in stone_age")
		}
	}

	completed = mm.CheckMilestones(2, "bronze_age", ageOrder, rm, bm, 0, 0, 0, nil, 0, 0)
	found := false
	for _, ms := range completed {
		if ms.Key == "bronze_pioneer" {
			found = true
		}
	}
	if !found {
		t.Error("bronze_pioneer should trigger in bronze_age")
	}
}

func TestMilestoneManager_ChainCompletion(t *testing.T) {
	mm := NewMilestoneManager()

	// Manually mark all military chain milestones as completed
	// Military chain only has "war_machine"
	mm.completed["war_machine"] = true

	chains := mm.CheckChains()
	found := false
	for _, c := range chains {
		if c.Key == "military_chain" {
			found = true
			if c.Title != "The Conquerors" {
				t.Errorf("military chain title = %v, want The Conquerors", c.Title)
			}
		}
	}
	if !found {
		t.Error("military_chain should complete when war_machine is done")
	}

	// Should not trigger again
	chains = mm.CheckChains()
	for _, c := range chains {
		if c.Key == "military_chain" {
			t.Error("military_chain should not trigger twice")
		}
	}
}

func TestMilestoneManager_TitleRecalculation(t *testing.T) {
	mm := NewMilestoneManager()

	// No milestones — no title
	mm.recalculateTitle()
	if mm.currentTitle != "" {
		t.Errorf("title with 0 milestones = %v, want empty", mm.currentTitle)
	}

	// 3 milestones = "Aspiring"
	mm.completed["first_shelter"] = true
	mm.completed["small_village"] = true
	mm.completed["knowledge_seeker"] = true
	mm.recalculateTitle()
	if mm.currentTitle != "Aspiring" {
		t.Errorf("title with 3 milestones = %v, want Aspiring", mm.currentTitle)
	}

	// Complete a chain — chain title overrides
	mm.completed["war_machine"] = true
	mm.chainsCompleted["military_chain"] = true
	mm.recalculateTitle()
	if mm.currentTitle != "The Conquerors" {
		t.Errorf("title with chain = %v, want The Conquerors", mm.currentTitle)
	}
}

func TestMilestoneManager_Snapshot(t *testing.T) {
	mm := NewMilestoneManager()
	mm.completed["first_shelter"] = true

	params := MilestoneSnapshotParams{
		Age:      "primitive_age",
		AgeOrder: fullAgeOrder(),
	}
	snap := mm.Snapshot(params)

	if snap.CompletedCount != 1 {
		t.Errorf("snapshot completed = %v, want 1", snap.CompletedCount)
	}
	if !snap.Milestones["first_shelter"].Completed {
		t.Error("first_shelter should be completed in snapshot")
	}
	if snap.Milestones["first_shelter"].RewardText == "" {
		t.Error("completed milestone should have reward text")
	}
}

func TestMilestoneManager_HiddenVisibility(t *testing.T) {
	mm := NewMilestoneManager()

	params := MilestoneSnapshotParams{
		Age:      "primitive_age",
		AgeOrder: fullAgeOrder(),
	}
	snap := mm.Snapshot(params)

	// metropolis is hidden, should not be visible at primitive age with no progress
	if snap.Milestones["metropolis"].Visible {
		t.Error("metropolis should be hidden at primitive_age with no progress")
	}

	// first_shelter is NOT hidden, should be visible
	if !snap.Milestones["first_shelter"].Visible {
		t.Error("first_shelter should be visible (not hidden)")
	}
}

func TestMilestoneManager_SaveLoadRoundTrip(t *testing.T) {
	mm := NewMilestoneManager()
	mm.completed["first_shelter"] = true
	mm.completed["war_machine"] = true
	mm.chainsCompleted["military_chain"] = true
	mm.currentTitle = "The Conquerors"

	// Save
	completed := mm.GetCompleted()
	chains := mm.GetChainsCompleted()
	title := mm.GetCurrentTitle()

	// Load into fresh
	mm2 := NewMilestoneManager()
	mm2.LoadState(completed, chains, title)

	if !mm2.IsCompleted("first_shelter") {
		t.Error("loaded manager should have first_shelter completed")
	}
	if !mm2.IsCompleted("war_machine") {
		t.Error("loaded manager should have war_machine completed")
	}
	if mm2.currentTitle != "The Conquerors" {
		t.Errorf("loaded title = %v, want The Conquerors", mm2.currentTitle)
	}
}
