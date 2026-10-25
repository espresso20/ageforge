//go:build smokebot

package smoke

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

func problemChecks(ps []problem) map[string]string {
	out := map[string]string{}
	for _, p := range ps {
		out[p.check] = p.msg
	}
	return out
}

// A building the next age asks for and that can no longer be built (its age
// has passed) is named at the first sweep. Stock has no cap, so a price is
// never "over the storage": this is the one way a required copy is stranded.
func TestRequiredBuildingUnbuildableIsNamed(t *testing.T) {
	defs := config.BuildingByKey()
	st := game.GameState{
		Age:            "stone_age",
		NextAge:        "bronze_age",
		NextAgeBldReqs: map[string]int{"stash": 5, "storage_pit": 5},
		Buildings: map[string]game.BuildingState{
			"stash":       {Name: "Stash", Unlocked: true, Count: 3},
			"storage_pit": {Name: "Storage Pit", Unlocked: true, Count: 1},
		},
	}
	got := problemChecks(requiredBuildingProblems(st, defs))
	msg, ok := got["required_building_unbuildable"]
	if !ok {
		t.Fatalf("a Stash short in the Stone Age was not flagged; got %v", got)
	}
	if !strings.Contains(msg, "needs 5 stash but only 3 exist") {
		t.Errorf("message %q does not say what is short", msg)
	}
	if strings.Contains(msg, "storage_pit") {
		t.Errorf("the Storage Pit is buildable in the Stone Age and must not be flagged: %q", msg)
	}
	// Enough standing: nothing to flag. The old "over the storage" problems
	// are gone with the stores.
	st.Buildings["stash"] = game.BuildingState{Name: "Stash", Unlocked: true, Count: 5}
	if ps := requiredBuildingProblems(st, defs); len(ps) != 0 {
		t.Errorf("five Stashes stand and the gate asks five, got %v", ps)
	}
}

// Blockers names what is short and what the next copy costs, and no longer
// speaks of a store.
func TestBlockersSayWhatIsShort(t *testing.T) {
	st := game.GameState{
		Age:            "bronze_age",
		NextAge:        "iron_age",
		NextAgeBldReqs: map[string]int{"warehouse": 5},
		Buildings: map[string]game.BuildingState{
			"warehouse": {Name: "Warehouse", Unlocked: true, Count: 2, NextCost: map[string]float64{"wood": 100}},
		},
	}
	line := Blockers(st)
	if !strings.Contains(line, "warehouse 2/5") {
		t.Errorf("blockers %q do not say the warehouses are 2 of 5", line)
	}
	if strings.Contains(line, "storage") && !strings.Contains(line, "warehouse") {
		t.Errorf("blockers %q speak of a store", line)
	}
}
