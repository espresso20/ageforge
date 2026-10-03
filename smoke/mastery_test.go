package smoke

import (
	"testing"
	"time"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// TestVerdictK: an age is graded against its target ÷ k; on known ground the
// Primitive and Stone Ages get no verdict of their own.
func TestVerdictK(t *testing.T) {
	iron, _ := Target("iron_age")
	if v := VerdictK("iron_age", iron.Seconds()/4, true, 4); v != VerdictOK {
		t.Errorf("iron at target ÷ 4 with k = 4: %q, want ok", v)
	}
	if v := VerdictK("iron_age", iron.Seconds(), true, 4); v != VerdictSlow {
		t.Errorf("iron at the full target with k = 4: %q, want slow (4x its target ÷ k)", v)
	}
	if v := VerdictK("iron_age", iron.Seconds(), true, 1); v != VerdictOK {
		t.Errorf("the frontier: %q", v)
	}
	if v := VerdictK("stone_age", 1e6, true, 4); v != VerdictNone {
		t.Errorf("stone on known ground: %q, want no verdict", v)
	}
}

// TestPresets: the veteran and returning presets' mastery and records.
func TestPresets(t *testing.T) {
	m, rec, ok := presetMastery(PresetVeteran)
	if !ok || rec != "interstellar_age" || m["space_age"] != config.MasteryCap || m["primitive_age"] != config.MasteryCap || m["interstellar_age"] != 0 {
		t.Errorf("veteran: %v, record %q", m, rec)
	}
	m, rec, ok = presetMastery(PresetReturning)
	if !ok || rec != "cyberpunk_age" || m["information_age"] != 2 || m["digital_age"] != 1 || m["cyberpunk_age"] != 0 || m["stone_age"] != 2 {
		t.Errorf("returning: %v, record %q", m, rec)
	}
	if ValidPreset("nope") || !ValidPreset("") {
		t.Error("ValidPreset")
	}
}

// TestLaterRunGrading: cycle 2 is compared with cycle 1 on cycle 1's ages,
// and graded on its median speed-up and depth.
func TestLaterRunGrading(t *testing.T) {
	order := map[string]int{}
	for i, a := range config.AgeOrder() {
		order[a] = i
	}
	run := func(seed int64, c1, c2 float64, c2End string) *RunResult {
		r := &RunResult{Seed: seed, FinalAge: "primitive_age"}
		r.Ages = []AgeSplit{
			{Cycle: 1, Age: "primitive_age", Seconds: c1 / 2},
			{Cycle: 1, Age: "stone_age", Seconds: c1 / 2},
			{Cycle: 1, Age: "bronze_age", Prestiged: true},
			{Cycle: 2, Age: "primitive_age", Seconds: c2 / 2},
			{Cycle: 2, Age: "stone_age", Seconds: c2 / 2},
			{Cycle: 2, Age: "bronze_age", Seconds: 100},
			{Cycle: 2, Age: c2End, Prestiged: true},
		}
		r.Cycles = []CycleSplit{{Cycle: 1, Seconds: c1, FinalAge: "bronze_age"}, {Cycle: 2, Seconds: c1, FinalAge: c2End}}
		return r
	}
	l := newLaterRun([]*RunResult{run(1, 1000, 400, "iron_age"), run(2, 1000, 500, "iron_age"), run(3, 1000, 600, "bronze_age")}, order)
	if l == nil || l.MedianSpeedup != 2 || l.MedianDepth != 1 || l.Failed {
		t.Fatalf("later run: %+v", l)
	}
	l = newLaterRun([]*RunResult{run(1, 1000, 700, "iron_age")}, order)
	if !l.Failed {
		t.Errorf("a 1.43x speed-up passed: %+v", l)
	}
}

// TestEarlyKnownGround: the Primitive and Stone Ages are graded together on
// known ground only.
func TestEarlyKnownGround(t *testing.T) {
	run := func(k, secs float64) *RunResult {
		return &RunResult{Ages: []AgeSplit{{Cycle: 1, Age: "primitive_age", Seconds: secs / 2, K: k}, {Cycle: 1, Age: "stone_age", Seconds: secs / 2, K: k}}}
	}
	if e := newEarly([]*RunResult{run(0, 3600)}); e != nil {
		t.Errorf("a new player's early ages were graded: %+v", e)
	}
	if e := newEarly([]*RunResult{run(4, 1800)}); e == nil || e.Verdict != VerdictOK {
		t.Errorf("30 minutes at k = 4: %+v", e)
	}
	if e := newEarly([]*RunResult{run(4, 2*VeteranEarlyMax.Seconds())}); e == nil || e.Verdict != VerdictSlow {
		t.Errorf("two hours at k = 4: %+v", e)
	}
	if lo, hi, ok := firstRunBand(Config{Preset: PresetVeteran}); !ok || lo != VeteranRunLow || hi != VeteranRunHigh || VeteranRunHigh != 36*time.Hour {
		t.Errorf("veteran band %v-%v", lo, hi)
	}
}

// TestMasteryCarryProblems: a prestige must raise exactly the ages below the
// run's furthest, capped, and keep the record.
func TestMasteryCarryProblems(t *testing.T) {
	before := game.GameState{Age: "bronze_age", Mastery: game.MasteryState{Record: "iron_age", RunFurthest: "bronze_age",
		Ages: map[string]int{"primitive_age": 10, "stone_age": 1, "iron_age": 3}}}
	good := game.GameState{Age: "primitive_age", Mastery: game.MasteryState{Record: "iron_age",
		Ages: map[string]int{"primitive_age": 10, "stone_age": 2, "iron_age": 3}}}
	if p := masteryCarryProblems(before, good); len(p) != 0 {
		t.Errorf("a correct commit flagged: %v", p)
	}
	bad := game.GameState{Age: "primitive_age", Mastery: game.MasteryState{Record: "stone_age",
		Ages: map[string]int{"primitive_age": 10, "stone_age": 1, "bronze_age": 1, "iron_age": 3}}}
	if p := masteryCarryProblems(before, bad); len(p) != 3 {
		t.Errorf("want 3 problems (stone not raised, bronze raised, record moved back), got %v", p)
	}
}

// TestVeteranPresetRun: a veteran-preset run plays its known ages at k > 1,
// cleanly, and its Primitive and Stone Ages come in under the limit.
func TestVeteranPresetRun(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Preset, cfg.StopAge, cfg.MaxSim = PresetVeteran, "bronze_age", 20*time.Hour
	r := Run(cfg, 1)
	if r.Failed() || r.Outcome != OutcomeDone {
		t.Fatalf("veteran run: %s at %s, anomalies %v", r.Outcome, r.FinalAge, r.Anomalies)
	}
	if len(r.Ages) < 2 || r.Ages[0].K != config.MasteryK(config.MasteryCap) {
		t.Fatalf("the Primitive Age ran at %v, want %v: %+v", r.Ages[0].K, config.MasteryK(config.MasteryCap), r.Ages)
	}
	if e := newEarly([]*RunResult{r}); e == nil || e.Verdict != VerdictOK {
		t.Errorf("Primitive and Stone on known ground: %+v", e)
	}
}
