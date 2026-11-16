package ui

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/game"
)

// TestStatsShowsAgesThisRun: the Stats panel lists the ages of this run with
// the time spent in each. The last age is the current one and is marked so far.
func TestStatsShowsAgesThisRun(t *testing.T) {
	_, eng := mapTestDashboard(t, true)
	must(t, eng.StartNewNamedGame("run"))
	st := eng.GetState()
	st.TickIntervalMs = 2000
	st.Tick = 450
	st.Stats.AgeLog = []game.AgeEntry{
		{Age: "primitive_age", Tick: 0},
		{Age: "stone_age", Tick: 300},
	}

	pg := buildStats(st, partRun, 120, 0, false, false, 0)
	if !strings.Contains(statsPageText(st, partRun, 120), "Ages this run") {
		t.Fatalf("stats panel has no Ages this run block")
	}
	// 300 ticks at 2s is 10 minutes; the current age has run 150 ticks, 5 minutes.
	want := []string{
		game.AgeName("primitive_age") + ": ~10m",
		game.AgeName("stone_age") + ": ~5m so far",
	}
	if len(pg.blocks) != len(want) {
		t.Fatalf("blocks = %+v", pg.blocks)
	}
	for i, w := range want {
		if pg.blocks[i].readout != w {
			t.Errorf("block %d reads %q, want %q", i, pg.blocks[i].readout, w)
		}
	}
}
