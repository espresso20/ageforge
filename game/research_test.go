package game

import (
	"testing"
)

func TestResearchManager_NoPriceToAfford(t *testing.T) {
	rm := NewResearchManager()
	ageOrder := map[string]int{"primitive_age": 0}

	if err := rm.StartResearch("tool_making", "primitive_age", ageOrder); err != nil {
		t.Errorf("a tech with its age and prerequisites has no price to afford: %v", err)
	}
}

func TestResearchManager_AgeGating(t *testing.T) {
	rm := NewResearchManager()
	ageOrder := map[string]int{"primitive_age": 0, "stone_age": 1}

	// stoneworking requires stone_age — should fail in primitive_age regardless of funds
	err := rm.StartResearch("stoneworking", "primitive_age", ageOrder)
	if err == nil {
		t.Error("StartResearch should fail when age requirement not met")
	}
}

func TestResearchManager_Prerequisites(t *testing.T) {
	rm := NewResearchManager()
	ageOrder := map[string]int{"primitive_age": 0, "stone_age": 1}

	// pottery requires fire_mastery — should fail without it
	err := rm.StartResearch("pottery", "stone_age", ageOrder)
	if err == nil {
		t.Error("StartResearch should fail when prerequisites not met")
	}
}
