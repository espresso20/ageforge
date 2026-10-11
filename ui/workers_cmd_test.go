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

// The workers command: bare opens the panel; auto-recruit shows and sets
// recruiting. The roster command shows, sets and clears shares, under its
// name and under its old one (workers share), which still runs it. Changes reply as routine
// lines (the Workers panel shows the result), a food share of 0 as a
// warning, and bad input with the usage.
func TestWorkersCommand(t *testing.T) {
	for _, roster := range []string{"roster", "workers share"} {
		testRosterCommand(t, roster)
	}
}

// testRosterCommand runs the roster command under one of its names.
func testRosterCommand(t *testing.T, roster string) {
	defer game.SetDataDirForTest(t.TempDir())()
	ge := sharesTestEngine(t)

	if res := HandleCommand("workers", ge); res.OverlayName != "workers" {
		t.Errorf("bare workers = %+v, want the Workers panel", res)
	}
	res := HandleCommand(roster+" knowledge 40", ge)
	if res.Type != game.LogRoutine || !strings.Contains(res.Message, "Knowledge 40%") {
		t.Errorf(roster+" knowledge 40 = %+v", res)
	}
	if got := ge.WorkerShares()["knowledge"]; got != 40 {
		t.Errorf("knowledge share = %v, want 40", got)
	}
	if res := HandleCommand(roster+" lumber 12.5%", ge); res.Type == "error" || ge.WorkerShares()["lumber"] != 12.5 {
		t.Errorf("a percent with a %% sign: %+v, shares %v", res, ge.WorkerShares())
	}
	res = HandleCommand(roster, ge)
	if res.Type != "info" || !strings.Contains(res.Message, "Knowledge: 40% (set)") || !strings.Contains(res.Message, "(auto)") {
		t.Errorf(roster+" listing = %+v", res)
	}
	if res := HandleCommand(roster+" knowledge", ge); res.Type != "info" || !strings.HasPrefix(res.Message, "Knowledge: 40% (set)") {
		t.Errorf("one domain's share = %+v", res)
	}
	if res := HandleCommand(roster+" knowledge auto", ge); res.Type != game.LogRoutine {
		t.Errorf(roster+" knowledge auto = %+v", res)
	}
	if _, set := ge.WorkerShares()["knowledge"]; set {
		t.Error("knowledge still has a share after auto")
	}
	if res := HandleCommand(roster+" food 0", ge); res.Type != "warning" {
		t.Errorf("a food share of 0 = %+v, want a warning", res)
	}
	if res := HandleCommand(roster+" auto", ge); res.Type != game.LogRoutine || ge.WorkerShares() != nil {
		t.Errorf(roster+" auto = %+v, shares %v", res, ge.WorkerShares())
	}

	for _, line := range []string{
		roster + " knowledge 101", roster + " knowledge -1", roster + " knowledge NaN",
		roster + " knowledge Inf", roster + " wizards 10", roster + " auto 10",
		roster + " knowledge 10 20", "workers auto-recruit maybe", "workers dance",
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
	// The roster's old name still runs and still completes once typed, but
	// it is not offered.
	if got := comp("workers "); !slices.Equal(got, []string{"workers auto-recruit"}) {
		t.Errorf("completing %q = %v, want auto-recruit alone", "workers ", got)
	}
	if got := comp("ros"); !slices.Contains(got, "roster") {
		t.Errorf("completing %q = %v, want roster", "ros", got)
	}
	got := comp("roster ")
	if len(got) < 3 || got[0] != "roster food" {
		t.Errorf("completing %q = %v, want the domains with buildings first", "roster ", got)
	}
	for _, want := range []string{"roster knowledge", "roster astronaut", "roster auto"} {
		if !slices.Contains(got, want) {
			t.Errorf("completing %q = %v, want %q", "roster ", got, want)
		}
	}
	if got := comp("roster knowledge "); !slices.Contains(got, "roster knowledge auto") {
		t.Errorf("completing %q = %v, want auto", "roster knowledge ", got)
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
	for _, line := range []string{"workers", "roster", "roster knowledge 40", "roster knowledge 40%", "roster auto",
		"workers share", "workers share knowledge 40", "workers share knowledge 40%", "workers share auto", "workers auto-recruit off"} {
		if !c.complete(line) {
			t.Errorf("%q is not a whole command", line)
		}
	}
	if c.complete("workers share wizards 10") || c.complete("roster wizards 10") {
		t.Error("an unknown domain counts as a whole command")
	}
}

// The Help panel lists the roster and the workers commands under Workers
// (the roster's old name in one line, not as a command of its own), and the
// Workers panel shows the roster and what auto-recruit is doing.
func TestWorkersHelpAndPanel(t *testing.T) {
	defer game.SetDataDirForTest(t.TempDir())()
	help := helpProvider(game.GameState{}, 0)
	for _, want := range []string{"roster", "roster auto", "workers auto-recruit", "Put every domain back on auto", "workers share still works"} {
		if !strings.Contains(help, want) {
			t.Errorf("the Help panel has no %q", want)
		}
	}
	if strings.Contains(help, "workers share <") || strings.Count(help, "workers share") != 1 {
		t.Error("the Help panel lists the roster's old name as a command")
	}
	ge := sharesTestEngine(t)
	ge.StepTicks(10)
	if res := HandleCommand("roster knowledge 50", ge); res.Type == "error" {
		t.Fatal(res.Message)
	}
	panel := workersProvider(ge.GetState(), 0)
	if strings.Contains(panel, "workers share") || strings.Contains(panel, "Shares") {
		t.Errorf("the Workers panel still uses the roster's old name:\n%s", panel)
	}
	for _, want := range []string{"Roster", "Auto-recruit:", "Knowledge", "set", "auto", "roster <domain> <percent|auto>"} {
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
