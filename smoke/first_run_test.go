package smoke

import (
	"strings"
	"testing"
	"time"

	"github.com/espresso20/ageforge/game"
)

// firstRunOf is a run whose first cycle reaches the Modern Age (prestiging
// there) after days of 1x play, or stops short of it when reached is false.
func firstRunOf(seed int64, days float64, reached bool) *RunResult {
	secs := days * 86400
	r := &RunResult{Seed: seed, Ages: []AgeSplit{
		{Cycle: 1, Age: "primitive_age", Seconds: 900},
		{Cycle: 1, Age: "stone_age", Seconds: 2700},
		{Cycle: 1, Age: "bronze_age", Seconds: secs - 3600},
	}}
	if reached {
		r.Ages = append(r.Ages, AgeSplit{Cycle: 1, Age: "modern_age", Prestiged: true},
			AgeSplit{Cycle: 2, Age: "primitive_age", Seconds: 600})
	}
	return r
}

// TestFirstRunBand: the first run to the Modern Age is graded on its median
// across seeds against FirstRunLow-FirstRunHigh, fails the set only under
// enforce, counts a seed that never got there as slower than any that did,
// and is left out when no seed got there (the fast tier).
func TestFirstRunBand(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Pacing = PacingEnforce
	cases := []struct {
		name    string
		runs    []*RunResult
		verdict string
	}{
		{"on target", []*RunResult{firstRunOf(1, 5.0, true), firstRunOf(2, 5.3, true), firstRunOf(3, 7.0, true)}, VerdictOK},
		{"fast", []*RunResult{firstRunOf(1, 4.0, true), firstRunOf(2, 4.5, true), firstRunOf(3, 5.5, true)}, VerdictFast},
		{"slow", []*RunResult{firstRunOf(1, 6.5, true), firstRunOf(2, 6.4, true), firstRunOf(3, 5.0, true)}, VerdictSlow},
		{"one never got there", []*RunResult{firstRunOf(1, 5.0, true), firstRunOf(2, 5.5, true), firstRunOf(3, 9, false)}, VerdictOK},
		{"most never got there", []*RunResult{firstRunOf(1, 5.0, true), firstRunOf(2, 9, false), firstRunOf(3, 9, false)}, VerdictSlow},
	}
	for _, c := range cases {
		sum := NewSummary("progression", cfg, time.Now(), c.runs)
		if sum.FirstRun == nil || sum.FirstRun.Verdict != c.verdict {
			t.Errorf("%s: first run %+v, want verdict %q", c.name, sum.FirstRun, c.verdict)
			continue
		}
		if sum.FirstRunFailed != (c.verdict != VerdictOK) {
			t.Errorf("%s: failed = %v under enforce", c.name, sum.FirstRunFailed)
		}
		var sb strings.Builder
		sum.writeFirstRun(&sb)
		if !strings.Contains(sb.String(), "First run to the Modern Age") {
			t.Errorf("%s: no first-run line in %q", c.name, sb.String())
		}
	}
	report := DefaultConfig()
	if sum := NewSummary("progression", report, time.Now(), cases[1].runs); sum.FirstRunFailed || sum.Failed {
		t.Errorf("report mode failed the set on a fast first run")
	}
	if sum := NewSummary("progression", cfg, time.Now(), []*RunResult{firstRunOf(1, 0.1, false)}); sum.FirstRun != nil || sum.FirstRunFailed {
		t.Errorf("no seed reached the Modern Age, yet the first run was graded: %+v", sum.FirstRun)
	}
	if got, ok := firstRunToModern(firstRunOf(4, 5.25, true), map[string]int{"primitive_age": 0, "stone_age": 1, "bronze_age": 2, game.PrestigeMinAge: 12}); !ok || got != 5.25*86400 {
		t.Errorf("firstRunToModern = %v, %v; want %v, true", got, ok, 5.25*86400)
	}
}

// TestQuietStretch: an age's longest stretch with nothing new to decide runs
// between its marks (entering, a building type built for the first time this
// run, a tech) or to its end, and says what came before it.
func TestQuietStretch(t *testing.T) {
	r := newRunner(DefaultConfig(), 1, game.NewGameEngine())
	st := game.GameState{Age: "iron_age", Buildings: map[string]game.BuildingState{}}
	r.enterAge(st)
	at := func(h float64) { r.sim = time.Duration(h * float64(time.Hour)) }

	at(1)
	st.Buildings = map[string]game.BuildingState{"forge": {Count: 1}}
	r.trackNovelty(st) // 1h quiet after entering
	at(5)
	st.Research.TotalResearched = 1
	r.trackNovelty(st) // 4h quiet after the forge
	at(6)
	st.Buildings = map[string]game.BuildingState{"forge": {Count: 3}}
	r.trackNovelty(st) // more copies are not new
	at(7)
	a := r.split(false)
	if a.QuietSecs != 4*3600 || a.QuietAfter != "new forge" {
		t.Errorf("quiet %v after %q, want 4h after \"new forge\"", time.Duration(a.QuietSecs)*time.Second, a.QuietAfter)
	}

	// The stretch to the age's end counts, and a new age starts afresh.
	at(20)
	if a := r.split(false); a.QuietSecs != 15*3600 || a.QuietAfter != "tech" {
		t.Errorf("quiet to the end %v after %q, want 15h after \"tech\"", time.Duration(a.QuietSecs)*time.Second, a.QuietAfter)
	}
	r.enterAge(game.GameState{Age: "classical_age"})
	at(21)
	if a := r.split(false); a.QuietSecs != 3600 || a.QuietAfter != "entering classical_age" {
		t.Errorf("new age: quiet %v after %q", time.Duration(a.QuietSecs)*time.Second, a.QuietAfter)
	}

	// Its median reaches the pacing table.
	runs := []*RunResult{{Seed: 1, Ages: []AgeSplit{{Cycle: 1, Age: "iron_age", Seconds: 3600, QuietSecs: 1800}}},
		{Seed: 2, Ages: []AgeSplit{{Cycle: 1, Age: "iron_age", Seconds: 3600, QuietSecs: 900}}},
		{Seed: 3, Ages: []AgeSplit{{Cycle: 1, Age: "iron_age", Seconds: 3600, QuietSecs: 2700}}}}
	sum := NewSummary("progression", DefaultConfig(), time.Now(), runs)
	if len(sum.Pacing) != 1 || sum.Pacing[0].QuietSecs != 1800 {
		t.Errorf("pacing rows %+v, want the 1800s median quiet", sum.Pacing)
	}
	var sb strings.Builder
	sum.writePacingTable(&sb)
	if !strings.Contains(sb.String(), "longest quiet") || !strings.Contains(sb.String(), "| 30m |") {
		t.Errorf("pacing table lacks the quiet column:\n%s", sb.String())
	}
}
