package game

import (
	"testing"
)

func TestResearchManager_CantAfford(t *testing.T) {
	rm := NewResearchManager()
	ageOrder := map[string]int{"primitive_age": 0}

	err := rm.StartResearch("tool_making", "primitive_age", ageOrder, 0)
	if err == nil {
		t.Error("StartResearch should fail with zero knowledge")
	}
}

func TestResearchManager_AgeGating(t *testing.T) {
	rm := NewResearchManager()
	ageOrder := map[string]int{"primitive_age": 0, "stone_age": 1}

	// stoneworking requires stone_age — should fail in primitive_age regardless of funds
	err := rm.StartResearch("stoneworking", "primitive_age", ageOrder, 999999)
	if err == nil {
		t.Error("StartResearch should fail when age requirement not met")
	}
}

func TestResearchManager_Prerequisites(t *testing.T) {
	rm := NewResearchManager()
	ageOrder := map[string]int{"primitive_age": 0}

	// fire_mastery requires tool_making — should fail without it
	err := rm.StartResearch("fire_mastery", "primitive_age", ageOrder, 999999)
	if err == nil {
		t.Error("StartResearch should fail when prerequisites not met")
	}
}
