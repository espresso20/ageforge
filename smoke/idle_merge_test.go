package smoke

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/espresso20/ageforge/game"
)

// TestIdleCheckInSpendsTheVisit: an idle player's visit is several rounds,
// not one decision. Over the first 12 hours at 3-hour check-ins the bot must
// leave the Primitive Age and play more than one round at some visit.
func TestIdleCheckInSpendsTheVisit(t *testing.T) {
	if testing.Short() {
		t.Skip("plays ~20k ticks")
	}
	t.Cleanup(game.SetDataDirForTest(t.TempDir()))
	cfg, err := ApplyStyle(DefaultConfig(), StyleIdle)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.CheckIn != IdleCheckIn || cfg.Horizon != IdleCheckIn {
		t.Fatalf("idle style: check-in %s, horizon %s, want %s for both", cfg.CheckIn, cfg.Horizon, IdleCheckIn)
	}
	cfg.MaxSim = 13 * time.Hour
	cfg.StopAge = "stone_age"
	var trace bytes.Buffer
	ge := game.NewGameEngine()
	ge.SeedRNG(1)
	r := newRunner(cfg, 1, ge)
	r.bot.Trace = &trace
	r.play()
	if len(r.res.Ages) == 0 || r.res.Ages[0].Age != "primitive_age" {
		t.Fatalf("idle bot never left the Primitive Age in %s (outcome %s)", cfg.MaxSim, r.res.Outcome)
	}
	if !strings.Contains(trace.String(), "checkin primitive_age") {
		t.Fatalf("no check-in lines in the trace:\n%s", trace.String())
	}
	multi := false
	for _, line := range strings.Split(trace.String(), "\n") {
		if strings.Contains(line, "leaves after") && !strings.Contains(line, "leaves after 1 round(s)") {
			multi = true
		}
	}
	if !multi {
		t.Errorf("every visit was a single round:\n%s", trace.String())
	}

	// Other styles never inherit a check-in interval.
	base := DefaultConfig()
	base.CheckIn = 8 * time.Hour
	if g, _ := ApplyStyle(base, StyleGreedy); g.CheckIn != 0 {
		t.Errorf("greedy style kept check-in %s", g.CheckIn)
	}
	if i, _ := ApplyStyle(base, StyleIdle); i.CheckIn != 8*time.Hour || i.DecideEvery != int(8*time.Hour/game.BaseTickInterval) {
		t.Errorf("idle style with an 8h check-in: %s, decide every %d ticks", i.CheckIn, i.DecideEvery)
	}
}

// TestMergeProgression: per-seed sessions pool into one graded session, the
// median is taken across all of them, and a seed reported twice is refused.
func TestMergeProgression(t *testing.T) {
	part := func(seed int64, bronzeSecs float64) *Session {
		cfg := DefaultConfig()
		cfg.Seeds = []int64{seed}
		runs := []*RunResult{{Seed: seed, Outcome: OutcomeDone, FinalAge: "iron_age", Ages: []AgeSplit{
			{Cycle: 1, Age: "primitive_age", Seconds: 900},
			{Cycle: 1, Age: "stone_age", Seconds: 2700},
			{Cycle: 1, Age: "bronze_age", Seconds: bronzeSecs},
		}}}
		res := &Result{Name: "progression", Status: StatusPass, Progression: []*Summary{NewSummary("progression", cfg, time.Now(), runs)}}
		s := &Session{Tier: TierDeep, Pacing: PacingReport, Started: time.Now(), Scenarios: []*Result{res}}
		// Round-trip through JSON, as the CI artifacts do.
		data, err := json.Marshal(s)
		if err != nil {
			t.Fatal(err)
		}
		var back Session
		if err := json.Unmarshal(data, &back); err != nil {
			t.Fatal(err)
		}
		return &back
	}
	e := &Env{Tier: TierDeep, Pacing: PacingEnforce}
	bronze := PacingTargets["bronze_age"].Seconds()
	onTarget, near, slow := bronze, 1.1*bronze, 7.4*bronze
	// One seed's Bronze Age is far too slow; the median of three is on target.
	sess, err := MergeProgression(e, []*Session{part(1, onTarget), part(2, slow), part(3, near)})
	if err != nil {
		t.Fatal(err)
	}
	if sess.Failed {
		t.Fatalf("median on target, want a pass: %+v", sess.Scenarios[1].Failures)
	}
	prog := sess.pacing()
	if prog == nil || len(prog.Runs) != 3 {
		t.Fatalf("want 3 pooled runs, got %+v", prog)
	}
	// Two of three slow: the median fails.
	if sess, _ = MergeProgression(e, []*Session{part(1, slow), part(2, slow), part(3, near)}); !sess.Failed {
		t.Errorf("median past the band, want a failure")
	}
	if _, err := MergeProgression(e, []*Session{part(1, onTarget), part(1, onTarget)}); err == nil {
		t.Errorf("a seed reported twice must be refused")
	}
}
