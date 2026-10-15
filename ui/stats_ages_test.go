package ui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/espresso20/ageforge/game"
)

// TestStatsShowsAgesThisRun: the Stats panel lists the ages of this run with
// the time spent in each. The last age is the current one and is marked so far.
func TestStatsShowsAgesThisRun(t *testing.T) {
	_, eng := mapTestDashboard(t, true)
	if err := eng.StartNewNamedGame("run"); err != nil {
		t.Fatal(err)
	}
	st := eng.GetState()
	st.TickIntervalMs = 2000
	st.Tick = 450
	st.Stats.AgeLog = []game.AgeEntry{
		{Age: "primitive_age", Tick: 0, At: time.Now()},
		{Age: "stone_age", Tick: 300, At: time.Now()},
	}

	text := statsProvider(st, 0)
	if !strings.Contains(text, "Ages this run") {
		t.Fatalf("stats panel has no Ages this run block:\n%s", text)
	}
	// 300 ticks at 2s is 10 minutes; the current age has run 150 ticks, 5 minutes.
	for _, want := range []string{
		fmt.Sprintf("   %-14s ~10m\n", game.AgeName("primitive_age")),
		fmt.Sprintf("   %-14s ~5m so far\n", game.AgeName("stone_age")),
	} {
		if !strings.Contains(text, want) {
			t.Errorf("stats panel missing line %q:\n%s", want, text)
		}
	}
}
