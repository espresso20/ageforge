package game

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestGameStats_CloneSharesNothing covers that a copy of the stats (a save, a load, a UI
// snapshot) can be changed without touching the live ones.
func TestGameStats_CloneSharesNothing(t *testing.T) {
	gs := NewGameStats()
	gs.RecordGather("wood", 5)
	gs.RecordAge("stone_age", 40)
	gs.Defense = &DefenseTally{Resources: map[string]float64{"wood": 2}}

	c := gs.clone()
	c.TotalGathered["wood"] = 99
	c.AgesReached[0] = "changed"
	c.AgeLog[0].Tick = 99
	c.Defense.Resources["wood"] = 99
	if gs.TotalGathered["wood"] != 5 || gs.AgesReached[0] != "primitive_age" || gs.AgeLog[0].Tick != 0 || gs.Defense.Resources["wood"] != 2 {
		t.Errorf("changing the copy changed the live stats: %+v defense %+v", gs, gs.Defense)
	}

	snap := gs.Snapshot()
	snap.TotalGathered["wood"] = 99
	snap.AgesReached[0] = "changed"
	if gs.TotalGathered["wood"] != 5 || gs.AgesReached[0] != "primitive_age" {
		t.Error("changing the snapshot changed the live stats")
	}
}

// TestGameStats_CloneKeepsMapsAndListsNonNil covers that stats with no gathered map or ages
// still copy to a map you can write to and a list written as [], never null.
func TestGameStats_CloneKeepsMapsAndListsNonNil(t *testing.T) {
	c := (&GameStats{}).clone()
	if c.TotalGathered == nil {
		t.Fatal("TotalGathered is nil after a copy; RecordGather would panic")
	}
	c.RecordGather("wood", 1)
	b, _ := json.Marshal(c)
	for _, want := range []string{`"total_gathered":{"wood":1}`, `"ages_reached":[]`} {
		if !strings.Contains(string(b), want) {
			t.Errorf("saved stats %s lack %s", b, want)
		}
	}
	if strings.Contains(string(b), "age_log") {
		t.Errorf("an empty age log is written: %s", b)
	}
}

// TestSaveSnapshot_StatsNeverNull covers that a game whose stats lost their gathered map
// and ages still saves {} and [], as it always has.
func TestSaveSnapshot_StatsNeverNull(t *testing.T) {
	ge := newSeededEngine(1)
	ge.Stats.TotalGathered = nil
	ge.Stats.AgesReached = nil
	ge.mu.RLock()
	save := ge.buildSaveSnapshot()
	ge.mu.RUnlock()
	b, _ := json.Marshal(save.Stats)
	for _, want := range []string{`"total_gathered":{}`, `"ages_reached":[]`} {
		if !strings.Contains(string(b), want) {
			t.Errorf("saved stats %s lack %s", b, want)
		}
	}
}
