package game

import (
	"testing"
)

func TestResearchManager_StartResearch(t *testing.T) {
	rm := NewResearchManager()
	ageOrder := map[string]int{"primitive_age": 0, "stone_age": 1}

	// tool_making: primitive_age, cost 25, no prereqs
	err := rm.StartResearch("tool_making", "primitive_age", ageOrder, 50)
	if err != nil {
		t.Errorf("StartResearch failed: %v", err)
	}
}

func TestResearchManager_CantAfford(t *testing.T) {
	rm := NewResearchManager()
	ageOrder := map[string]int{"primitive_age": 0}

	// tool_making costs 25, only have 10
	err := rm.StartResearch("tool_making", "primitive_age", ageOrder, 10)
	if err == nil {
		t.Error("StartResearch should fail with insufficient knowledge")
	}
}

func TestResearchManager_AgeGating(t *testing.T) {
	rm := NewResearchManager()
	ageOrder := map[string]int{"primitive_age": 0, "stone_age": 1}

	// stoneworking requires stone_age
	err := rm.StartResearch("stoneworking", "primitive_age", ageOrder, 1000)
	if err == nil {
		t.Error("StartResearch should fail when age requirement not met")
	}
}

func TestResearchManager_Prerequisites(t *testing.T) {
	rm := NewResearchManager()
	ageOrder := map[string]int{"primitive_age": 0}

	// fire_mastery requires tool_making
	err := rm.StartResearch("fire_mastery", "primitive_age", ageOrder, 1000)
	if err == nil {
		t.Error("StartResearch should fail when prerequisites not met")
	}
}

func TestResearchManager_TickCompletion(t *testing.T) {
	rm := NewResearchManager()
	ageOrder := map[string]int{"primitive_age": 0}

	rm.StartResearch("tool_making", "primitive_age", ageOrder, 50)

	// Tick until complete
	var completed string
	for i := 0; i < 100; i++ {
		completed = rm.Tick()
		if completed != "" {
			break
		}
	}

	if completed != "tool_making" {
		t.Errorf("expected tool_making to complete, got %q", completed)
	}
	if !rm.IsResearched("tool_making") {
		t.Error("tool_making should be marked as researched")
	}
	if rm.ResearchedCount() != 1 {
		t.Errorf("ResearchedCount = %v, want 1", rm.ResearchedCount())
	}
}

func TestResearchManager_BonusAccumulation(t *testing.T) {
	rm := NewResearchManager()
	ageOrder := map[string]int{"primitive_age": 0}

	rm.StartResearch("tool_making", "primitive_age", ageOrder, 50)
	for rm.Tick() == "" {
	}

	// tool_making gives gather_rate +0.15
	bonus := rm.GetBonus("gather_rate")
	if bonus != 0.15 {
		t.Errorf("gather_rate bonus = %v, want 0.15", bonus)
	}
}

func TestResearchManager_CancelResearch(t *testing.T) {
	rm := NewResearchManager()
	ageOrder := map[string]int{"primitive_age": 0}

	rm.StartResearch("tool_making", "primitive_age", ageOrder, 50)
	key, ok := rm.CancelResearch()
	if !ok || key != "tool_making" {
		t.Errorf("CancelResearch = (%v, %v), want (tool_making, true)", key, ok)
	}

	// Cancel with nothing in progress
	_, ok = rm.CancelResearch()
	if ok {
		t.Error("CancelResearch should fail with nothing in progress")
	}
}

func TestResearchManager_CantResearchTwice(t *testing.T) {
	rm := NewResearchManager()
	ageOrder := map[string]int{"primitive_age": 0}

	rm.StartResearch("tool_making", "primitive_age", ageOrder, 50)
	for rm.Tick() == "" {
	}

	err := rm.StartResearch("tool_making", "primitive_age", ageOrder, 50)
	if err == nil {
		t.Error("StartResearch should fail for already-researched tech")
	}
}

func TestResearchManager_SaveLoadRoundTrip(t *testing.T) {
	rm := NewResearchManager()
	ageOrder := map[string]int{"primitive_age": 0}

	rm.StartResearch("tool_making", "primitive_age", ageOrder, 50)
	for rm.Tick() == "" {
	}

	researched := rm.GetResearched()
	rm2 := NewResearchManager()
	rm2.LoadState(researched, "", 0, 0)

	if !rm2.IsResearched("tool_making") {
		t.Error("loaded research manager should have tool_making")
	}
	if rm2.GetBonus("gather_rate") != 0.15 {
		t.Errorf("loaded bonus = %v, want 0.15", rm2.GetBonus("gather_rate"))
	}
}
