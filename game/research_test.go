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

// treeFixtureTechs is a research order the tree's prerequisites no longer allow for Alchemy: the first four
// ages' techs but Philosophy and what stood on it.
var treeFixtureTechs = []string{
	"tool_making", "fire_mastery", "stoneworking", "animal_husbandry", "pottery", "primitive_writing",
	"bronze_working", "agriculture", "currency", "masonry", "military_tactics",
	"iron_smelting", "road_building", "mathematics", "siege_warfare",
}

// TestResearchInProgressOutOfOrderFinishes is the same promise without the
// clock: research restored mid-way, as a load restores it, with a
// prerequisite the tree now asks for missing, ticks to its end.
func TestResearchInProgressOutOfOrderFinishes(t *testing.T) {
	ge := NewGameEngine()
	ge.Research.LoadState(treeFixtureTechs, "alchemy", 0, false)
	if def, _ := ge.rules.Tech("alchemy"); def.PrereqsMet(ge.Research.IsResearched) {
		t.Fatal("Alchemy's prerequisites are met: nothing is out of order")
	}
	price := ge.Research.defs["alchemy"].Cost
	ge.researchPerTick = price / 1199.5
	for i := 0; i < 1199; i++ {
		ge.processResearch()
	}
	if ge.Research.currentTech != "alchemy" {
		t.Fatalf("one tick from the end: researching %q", ge.Research.currentTech)
	}
	ge.processResearch()
	if !ge.Research.IsResearched("alchemy") || ge.Research.currentTech != "" {
		t.Errorf("Alchemy researched %v, still researching %q", ge.Research.IsResearched("alchemy"), ge.Research.currentTech)
	}
	if !logHas(ge, "Research complete: Alchemy.") {
		t.Error("no line in the log for the finished research")
	}
}
