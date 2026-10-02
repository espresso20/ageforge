package ui

import (
	"slices"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/game"
)

// sharesTestEngine is a Primitive Age game with housing, two camps and a
// story circle, its workers already recruited by the shares routine.
func sharesTestEngine(t *testing.T) *game.GameEngine {
	t.Helper()
	ge := newCaseTestEngine(t, "")
	for _, b := range []string{"wood_camp", "story_circle", "hut"} {
		if msg := game.DevExecCommand("/build "+b, ge); msg != "built "+b {
			t.Fatalf("dev /build %s: %q", b, msg)
		}
	}
	return ge
}

// The workers command: bare opens the panel; share shows, sets and clears
// shares; auto-recruit shows and sets recruiting. Changes reply as routine
// lines (the Workers panel shows the result), a food share of 0 as a
// warning, and bad input with the usage.
func TestWorkersCommand(t *testing.T) {
	defer game.SetDataDirForTest(t.TempDir())()
	ge := sharesTestEngine(t)

	if res := HandleCommand("workers", ge); res.OverlayName != "workers" {
		t.Errorf("bare workers = %+v, want the Workers panel", res)
	}
	res := HandleCommand("workers share knowledge 40", ge)
	if res.Type != game.LogRoutine || !strings.Contains(res.Message, "Knowledge 40%") {
		t.Errorf("workers share knowledge 40 = %+v", res)
	}
	if got := ge.WorkerShares()["knowledge"]; got != 40 {
		t.Errorf("knowledge share = %v, want 40", got)
	}
	if res := HandleCommand("workers share lumber 12.5%", ge); res.Type == "error" || ge.WorkerShares()["lumber"] != 12.5 {
		t.Errorf("a percent with a %% sign: %+v, shares %v", res, ge.WorkerShares())
	}
	res = HandleCommand("workers share", ge)
	if res.Type != "info" || !strings.Contains(res.Message, "Knowledge: 40% (set)") || !strings.Contains(res.Message, "(auto)") {
		t.Errorf("workers share listing = %+v", res)
	}
	if res := HandleCommand("workers share knowledge", ge); res.Type != "info" || !strings.HasPrefix(res.Message, "Knowledge: 40% (set)") {
		t.Errorf("one domain's share = %+v", res)
	}
	if res := HandleCommand("workers share knowledge auto", ge); res.Type != game.LogRoutine {
		t.Errorf("workers share knowledge auto = %+v", res)
	}
	if _, set := ge.WorkerShares()["knowledge"]; set {
		t.Error("knowledge still has a share after auto")
	}
	if res := HandleCommand("workers share food 0", ge); res.Type != "warning" {
		t.Errorf("a food share of 0 = %+v, want a warning", res)
	}
	if res := HandleCommand("workers share auto", ge); res.Type != game.LogRoutine || ge.WorkerShares() != nil {
		t.Errorf("workers share auto = %+v, shares %v", res, ge.WorkerShares())
	}

	for _, line := range []string{
		"workers share knowledge 101", "workers share knowledge -1", "workers share knowledge NaN",
		"workers share knowledge Inf", "workers share wizards 10", "workers share auto 10",
		"workers share knowledge 10 20", "workers auto-recruit maybe", "workers dance",
	} {
		if res := HandleCommand(line, ge); res.Type != "error" {
			t.Errorf("%q = %+v, want a refusal", line, res)
		}
	}
	if ge.WorkerShares() != nil {
		t.Errorf("refused commands changed the shares: %v", ge.WorkerShares())
	}

	if res := HandleCommand("workers auto-recruit off", ge); res.Type != game.LogRoutine || ge.AutoRecruit() {
		t.Errorf("auto-recruit off = %+v, on %v", res, ge.AutoRecruit())
	}
	if res := HandleCommand("workers autorecruit", ge); res.Type != "info" || !strings.Contains(res.Message, "off") {
		t.Errorf("auto-recruit status = %+v", res)
	}
	if res := HandleCommand("workers auto-recruit on", ge); res.Type != game.LogRoutine || !ge.AutoRecruit() {
		t.Errorf("auto-recruit on = %+v, on %v", res, ge.AutoRecruit())
	}
}

// Completion walks the workers command: its subcommands, the domains (the
// ones with buildings first) and the auto words.
func TestWorkersCompletion(t *testing.T) {
	defer game.SetDataDirForTest(t.TempDir())()
	ge := sharesTestEngine(t)
	comp := NewAutoCompleter(ge)
	got := comp("workers ")
	for _, want := range []string{"workers share", "workers auto-recruit"} {
		if !slices.Contains(got, want) {
			t.Errorf("completing %q = %v, want %q", "workers ", got, want)
		}
	}
	got = comp("workers share ")
	if len(got) < 3 || got[0] != "workers share food" {
		t.Errorf("completing %q = %v, want the domains with buildings first", "workers share ", got)
	}
	for _, want := range []string{"workers share knowledge", "workers share astronaut", "workers share auto"} {
		if !slices.Contains(got, want) {
			t.Errorf("completing %q = %v, want %q", "workers share ", got, want)
		}
	}
	if got := comp("workers share kn"); !slices.Equal(got, []string{"workers share knowledge"}) {
		t.Errorf("completing %q = %v", "workers share kn", got)
	}
	if got := comp("workers share knowledge "); !slices.Contains(got, "workers share knowledge auto") {
		t.Errorf("completing %q = %v, want auto", "workers share knowledge ", got)
	}
	if got := comp("workers auto-recruit "); !slices.Equal(got, []string{"workers auto-recruit off", "workers auto-recruit on"}) {
		t.Errorf("completing %q = %v", "workers auto-recruit ", got)
	}

	// Enter runs whole commands as typed, a percent with its sign included.
	c := newCompleter(ge, nil)
	for _, line := range []string{"workers", "workers share", "workers share knowledge 40", "workers share knowledge 40%", "workers share auto", "workers auto-recruit off"} {
		if !c.complete(line) {
			t.Errorf("%q is not a whole command", line)
		}
	}
	if c.complete("workers share wizards 10") {
		t.Error("an unknown domain counts as a whole command")
	}
}

// The Help panel lists the workers commands under Workers, and the Workers
// panel shows the shares and what auto-recruit is doing.
func TestWorkersHelpAndPanel(t *testing.T) {
	defer game.SetDataDirForTest(t.TempDir())()
	help := helpProvider(game.GameState{}, 0)
	for _, want := range []string{"workers share", "workers auto-recruit", "Put every domain back on auto"} {
		if !strings.Contains(help, want) {
			t.Errorf("the Help panel has no %q", want)
		}
	}
	ge := sharesTestEngine(t)
	ge.StepTicks(10)
	if res := HandleCommand("workers share knowledge 50", ge); res.Type == "error" {
		t.Fatal(res.Message)
	}
	panel := workersProvider(ge.GetState(), 0)
	for _, want := range []string{"Shares", "Auto-recruit:", "Knowledge", "set", "auto", "workers share <domain> <percent|auto>"} {
		if !strings.Contains(panel, want) {
			t.Errorf("the Workers panel has no %q:\n%s", want, panel)
		}
	}
	// A fresh game shows the shares too, and says workers come on their own.
	fresh := game.NewGameEngine()
	panel = workersProvider(fresh.GetState(), 0)
	if !strings.Contains(panel, "come on their own") || !strings.Contains(panel, "Auto-recruit:") {
		t.Errorf("a fresh game's Workers panel:\n%s", panel)
	}
}
