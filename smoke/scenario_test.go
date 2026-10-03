package smoke

import (
	"bytes"
	"maps"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

func TestTargetsCoverEveryAge(t *testing.T) {
	ages := config.AgeOrder()
	for i, a := range ages {
		_, ok := Target(a)
		if last := i == len(ages)-1; ok == last {
			t.Errorf("%s: has target = %v; every age but the last needs one", a, ok)
		}
	}
	for a := range PacingTargets {
		if _, ok := config.AgeByKey()[a]; !ok {
			t.Errorf("pacing target for unknown age %q", a)
		}
	}
	if got := CumulativeTarget("modern_age"); got < 160*time.Hour || got > 175*time.Hour {
		t.Errorf("targets to the Modern Age sum to %s; the plan is about a week", got)
	}
	if got := CumulativeTarget("modern_age"); got < FirstRunHigh {
		t.Errorf("targets to the Modern Age sum to %s, under the first-run band's top %s", got, FirstRunHigh)
	}
	if FirstRunLow >= FirstRunHigh {
		t.Errorf("first-run band %s-%s is empty", FirstRunLow, FirstRunHigh)
	}
	if got := AgeTimeout("primitive_age"); got != TimeoutFloor {
		t.Errorf("primitive timeout %s, want the %s floor", got, TimeoutFloor)
	}
	if got := AgeTimeout("atomic_age"); got != 4*(31*time.Hour+12*time.Minute) {
		t.Errorf("atomic timeout %s, want 4 x 31h12m", got)
	}
	target := PacingTargets["stone_age"].Seconds()
	for _, c := range []struct {
		secs     float64
		finished bool
		want     string
	}{
		{target, true, VerdictOK}, {target * 0.49, true, VerdictFast}, {target * 2.01, true, VerdictSlow},
		{target * 0.5, true, VerdictOK}, {target * 2, true, VerdictOK}, {target * 0.1, false, VerdictNone},
		{target * 3, false, VerdictSlow},
	} {
		if got := Verdict("stone_age", c.secs, c.finished); got != c.want {
			t.Errorf("Verdict(stone, %.2fx, finished=%v) = %q, want %q", c.secs/target, c.finished, got, c.want)
		}
	}
}

func TestFirstDiff(t *testing.T) {
	type inner struct{ V []float64 }
	type s struct {
		M map[string]float64
		P *inner
		T []string
	}
	mk := func() s {
		return s{M: map[string]float64{"a": 1, "b": 0.1 + 0.2}, P: &inner{V: []float64{1, 2}}, T: nil}
	}
	a, b := mk(), mk()
	if d := firstDiff(a, b, nil); d != "" {
		t.Fatalf("equal values differ: %s", d)
	}
	b.P.V[1] = 2 + 1e-15
	if d := firstDiff(a, b, nil); !strings.HasPrefix(d, "P.V[1]") {
		t.Errorf("a one-ulp change should be found at P.V[1], got %q", d)
	}
	b = mk()
	b.T = []string{}
	if d := firstDiff(a, b, nil); d != "" {
		t.Errorf("nil and empty slices should match, got %q", d)
	}
	b.M["c"] = 3
	if d := firstDiff(a, b, func(p string) bool { return p == "M" }); d != "" {
		t.Errorf("skipped path still reported: %q", d)
	}

	ta, _ := jsonTree([]byte(`{"x":["b","a"],"y":{"z":["1","2"]},"n":1.5}`))
	tb, _ := jsonTree([]byte(`{"x":["a","b"],"y":{"z":["1","2"]},"n":1.5}`))
	if got := sortStringSets(ta, tb, ""); len(got) != 1 || got[0] != "x" {
		t.Errorf("order-only paths = %v, want [x]", got)
	}
	if d := firstDiff(ta, tb, nil); d != "" {
		t.Errorf("after sorting, trees differ: %s", d)
	}
}

// TestPrestigeFormulaMatchesDocs: the worked examples in
// site/docs/prestige.md. Depth points: 3^epoch per completed age, with no
// divisor, so the level and the run's milestones, techs and buildings
// change nothing.
func TestPrestigeFormulaMatchesDocs(t *testing.T) {
	for _, c := range []struct {
		age  string
		want int
	}{
		{"medieval_age", 9},     // 1+1+1 (Stone Era) + 3+3 (Iron, Classical)
		{"renaissance_age", 12}, // + 3 for the Medieval Age
		{"modern_age", 120},     // 3 + 9 + 27 + 81
		{"information_age", 201},
		{"cyberpunk_age", 363}, // a run through the Digital Age
		{"interstellar_age", 1092},
		{"transcendent_age", 3279},
	} {
		st := game.GameState{Age: c.age}
		st.Milestones.CompletedCount, st.Research.TotalResearched, st.Stats.TotalBuilt = 25, 31, 149
		st.Prestige.Level = 7
		if got := PrestigePoints(st); got != c.want {
			t.Errorf("%s: %d points, want %d", c.age, got, c.want)
		}
		if got := config.DepthPoints(c.age); got != c.want {
			t.Errorf("config.DepthPoints(%s) = %d, want %d", c.age, got, c.want)
		}
	}
}

// The scenarios below run in the plain test suite in a few seconds each, so
// a harness regression shows up in `go test ./...` and not only in CI's
// smoke job.

func testEnv() *Env {
	return &Env{Tier: TierFast, SeedBase: 1, Pacing: PacingReport, Parallel: 2, Base: DefaultConfig(), RepoRoot: ".."}
}

func runScenario(t *testing.T, name string) *Result {
	t.Helper()
	sess, err := RunScenarios(testEnv(), []string{name})
	if err != nil {
		t.Fatal(err)
	}
	r := sess.Scenarios[0]
	for _, f := range r.Failures {
		t.Errorf("%s: %s: %s\n%s\n%s", name, f.Check, f.Message, f.Repro, f.Detail)
	}
	return r
}

func TestAccountsScenario(t *testing.T) {
	runScenario(t, "accounts")
}

func TestOfflineScenario(t *testing.T) {
	if testing.Short() || raceEnabled {
		t.Skip("plays to the Stone Age")
	}
	runScenario(t, "offline")
}

func TestPrestigeHookedMechanics(t *testing.T) {
	if testing.Short() || raceEnabled {
		t.Skip("plays to the Stone Age twice")
	}
	t.Cleanup(game.SetDataDirForTest(t.TempDir()))
	res := &Result{}
	steps := prestigeHooked(testEnv(), res)
	for _, f := range res.Failures {
		t.Errorf("%s: %s\n%s", f.Check, f.Message, f.Detail)
	}
	if len(steps) < 20 {
		t.Errorf("only %d hooked steps ran:\n%s", len(steps), strings.Join(steps, "\n"))
	}
}

func TestFuzzShortRun(t *testing.T) {
	if testing.Short() || raceEnabled {
		t.Skip("plays a few thousand ticks")
	}
	t.Cleanup(game.SetDataDirForTest(t.TempDir()))
	t.Chdir(t.TempDir()) // `dump` writes under ./data/logs
	ge := game.NewGameEngine()
	gen := newFuzzer(7, ge)
	if len(gen.commands) < 40 {
		t.Fatalf("the command registry gave only %d commands: %v", len(gen.commands), gen.commands)
	}
	script := fuzzScript(gen.r, 60, 20, 20)
	f, done := fuzzExec(7, script, gen)
	if f != nil {
		t.Fatalf("%s at step %d: %s\n%s", f.check, f.step, f.msg, f.detail)
	}
	if len(done) != len(script) {
		t.Errorf("ran %d of %d steps", len(done), len(script))
	}
}

func TestDocsyncParsers(t *testing.T) {
	cmds, accepts := registryCommands()
	if len(cmds) < 30 || len(accepts["trade"]) == 0 || !accepts["trade"]["route"] {
		t.Errorf("registry read looks wrong: %d commands, trade accepts %v", len(cmds), accepts["trade"])
	}
	if subs := offeredSubcommands(); !slices.Contains(subs["trade"], "route start") || slices.Contains(subs["gather"], "food") {
		t.Errorf("offered subcommands look wrong: trade %v, gather %v", subs["trade"], subs["gather"])
	}
	claims, err := docClaims("..")
	if err != nil {
		t.Fatal(err)
	}
	hero := 0
	for _, c := range claims {
		if strings.HasPrefix(c.text, "hero stat") {
			hero++
		}
	}
	if hero < 3 {
		t.Errorf("found %d hero stats in site/index.html, want the ages, buildings and techs", hero)
	}
	if n, rows, err := lineageTable(".."); err != nil || n == 0 || rows == 0 {
		t.Errorf("lineage table: heading %d, rows %d, err %v", n, rows, err)
	}
	md := "| Key | Tab |\n|---|---|\n| `e` | Economy |\n\n| Shortcut | Command |\n|---|---|\n| `b` | `build` |\n| `h`, `?` | `help` |\n"
	want := map[string]string{"b": "build", "h": "help", "?": "help"}
	if got := documentedShortcuts(md); !maps.Equal(got, want) {
		t.Errorf("shortcuts table read as %v, want %v", got, want)
	}
}

func TestSessionReports(t *testing.T) {
	sess := &Session{Tier: TierFast, Pacing: PacingReport, Scenarios: []*Result{
		{Name: "fuzz", Status: StatusFail, Summary: "1 crash", Failures: []Finding{{Check: "fuzz_panic", Message: "boom", Repro: "go run ./cmd/smoke -scenario fuzz"}}},
		{Name: "docsync", Status: StatusPass, Warnings: []Finding{{Check: "w", Message: "careful"}}},
		{Name: "styles", Status: StatusSkip, Summary: "full tier only"},
	}, Failed: true}
	var sum, md, js bytes.Buffer
	for _, w := range []func() error{
		func() error { return sess.WriteSummary(&sum) },
		func() error { return sess.WriteMarkdown(&md) },
		func() error { return sess.WriteJSON(&js) },
	} {
		if err := w(); err != nil {
			t.Fatal(err)
		}
	}
	for _, want := range []string{"FAIL", "| fuzz | ✗ FAIL |", "fuzz / fuzz_panic", "boom", "careful", "- skip"} {
		if !strings.Contains(sum.String(), want) {
			t.Errorf("summary missing %q:\n%s", want, sum.String())
		}
	}
	if !strings.Contains(md.String(), "go run ./cmd/smoke -scenario fuzz") || !strings.Contains(js.String(), `"fuzz_panic"`) {
		t.Errorf("report or JSON missing the finding")
	}
}

// An age left by prestige is not a completed age: prestiging on entering the
// Modern Age spends 0 seconds there, which used to grade as a "fast" Modern
// Age and fail every -pacing enforce run with prestige cycles.
func TestPrestigedAgeIsNotGraded(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Pacing = PacingEnforce
	r := newRunner(cfg, 1, game.NewGameEngine())
	r.age = "modern_age"
	r.closeAgeByPrestige()
	if len(r.res.Anomalies) != 0 {
		t.Fatalf("closing an age by prestige raised %+v", r.res.Anomalies[0])
	}
	if a := r.res.Ages[0]; !a.Prestiged || a.Verdict != VerdictNone {
		t.Errorf("split = %+v, want prestiged and ungraded", a)
	}
	// Enforcement grades the median across seeds: one fast seed of three
	// passes, a fast median fails, and the prestige age never counts.
	atomic := PacingTargets["atomic_age"].Seconds()
	runs := []*RunResult{
		{Seed: 1, Ages: []AgeSplit{{Cycle: 1, Age: "modern_age", Prestiged: true}, {Cycle: 1, Age: "atomic_age", Seconds: atomic}, {Cycle: 1, Age: "electric_age", Seconds: 1}}},
		{Seed: 2, Ages: []AgeSplit{{Cycle: 1, Age: "modern_age", Prestiged: true}, {Cycle: 1, Age: "atomic_age", Seconds: 0.1 * atomic}, {Cycle: 1, Age: "electric_age", Seconds: 1}}},
		{Seed: 3, Ages: []AgeSplit{{Cycle: 1, Age: "modern_age", Prestiged: true}, {Cycle: 1, Age: "atomic_age", Seconds: atomic}, {Cycle: 2, Age: "electric_age", Seconds: 1}}},
	}
	sum := NewSummary("progression", cfg, time.Now(), runs)
	if len(sum.PacingFailures) != 1 || sum.PacingFailures[0].Age != "electric_age" || !sum.Failed {
		t.Errorf("pacing failures = %+v, want only the cycle-1 electric_age median", sum.PacingFailures)
	}
	var sb strings.Builder
	sum.writePacingTable(&sb)
	for _, p := range sum.Pacing {
		if p.Age == "modern_age" && (!p.Prestiged || p.Verdict != VerdictNone) {
			t.Errorf("modern row = %+v, want prestiged and ungraded", p)
		}
	}
	if !strings.Contains(sb.String(), "modern_age (left by prestige)") {
		t.Errorf("pacing table does not mark the prestige age:\n%s", sb.String())
	}
}
