package ui

import (
	"slices"
	"testing"

	"github.com/espresso20/ageforge/game"
)

func TestAutoCompleter_OffersMissingCommands(t *testing.T) {
	defer game.SetDataDirForTest(t.TempDir())()
	comp := NewAutoCompleter(game.NewGameEngine())

	if got := comp("ac"); !slices.Contains(got, "acct") {
		t.Errorf("completing %q = %v, want the acct alias", "ac", got)
	}
	for _, line := range []string{"diplomacy ", "dip "} {
		got := comp(line)
		for _, want := range []string{line + "tribute", line + "raid"} {
			if !slices.Contains(got, want) {
				t.Errorf("completing %q = %v, want %q", line, got, want)
			}
		}
	}
	if got := comp("acct "); !slices.Contains(got, "acct switch") {
		t.Errorf("completing %q = %v, want the account subcommands", "acct ", got)
	}
}

// TestAutoCompleter_NoRepeatAfterLastArg: once a command's last argument is
// typed there is nothing left to offer. It used to re-offer the same words
// ("research list list", "assign hut all all").
func TestAutoCompleter_NoRepeatAfterLastArg(t *testing.T) {
	defer game.SetDataDirForTest(t.TempDir())()
	comp := NewAutoCompleter(game.NewGameEngine())
	for _, line := range []string{
		"research list ", "research cancel ", "research list l", "res cancel c",
		"expedition list ", "campaign list ",
		"assign hut all ", "unassign hut all ", "dismiss hut all ",
		"prestige confirm yes ", "prestige buy x ", "festival confirm yes ",
		"save list ", "load x ", "gather food ",
		"account switch x ",
	} {
		if got := comp(line); len(got) != 0 {
			t.Errorf("completing %q = %v, want nothing", line, got)
		}
	}
	// The first argument still completes.
	if got := comp("research "); !slices.Contains(got, "research list") {
		t.Errorf("completing %q = %v, want it to offer list", "research ", got)
	}
}

// TestEnterLine is the Enter rule table, without the UI.
func TestEnterLine(t *testing.T) {
	t.Cleanup(game.SetDataDirForTest(t.TempDir()))
	c := newCompleter(game.NewGameEngine(), nil)
	for _, tc := range []struct {
		typed, line string
		run         bool
	}{
		{"advance", "advance", true}, // whole: as typed
		{"adv", "advance", true},     // ghost makes it whole
		{"b", "b", true},             // an alias is whole (bare build lists)
		{"research list", "research list", true},
		{"research ", "research ", true}, // bare research opens the panel
		{"plan bu", "plan bu", true},     // plan build needs a building: as typed
		{"zzqx", "zzqx", true},           // garbage: as typed
		{"build nosuch", "build nosuch", true},
		{"plan cle", "plan clear", false}, // dangerous: filled in, not run
		{"prestige confirm y", "prestige confirm yes", false},
		{"research canc", "research cancel", false},
		{"prestige confirm yes", "prestige confirm yes", true}, // typed in full: runs
		{"festival conf", "festival confirm", true},            // the prompt, not the act
		{"build hut m", "build hut max", true},
		{"q", "quit", false}, // not a quit on a stray key
	} {
		line, run := c.enterLine(tc.typed)
		if line != tc.line || run != tc.run {
			t.Errorf("Enter on %q = (%q, run %v), want (%q, run %v)", tc.typed, line, run, tc.line, tc.run)
		}
	}
}
