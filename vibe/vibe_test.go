package vibe

import (
	"errors"
	"fmt"
	"sort"
	"testing"
)

func TestLineFollowsTheSwitch(t *testing.T) {
	defer UseScriptForTest(map[string]string{"x.one": "vibe one"})()
	Set(false)
	defer Set(false)
	if On() || Line("x.one", "plain one") != "plain one" {
		t.Error("with VibeMode off, Line is not the plain text")
	}
	Set(true)
	if !On() {
		t.Fatal("Set(true) did not turn VibeMode on")
	}
	if got := Line("x.one", "plain one"); got != "vibe one" {
		t.Errorf("with VibeMode on, a row the script has reads %q", got)
	}
	if got := Line("x.two", "plain two"); got != "plain two" {
		t.Errorf("with VibeMode on, a row the script lacks reads %q, want the plain text", got)
	}
	Set(false)
	if got := Line("x.one", "plain one"); got != "plain one" {
		t.Errorf("turned off again, Line reads %q", got)
	}
	// Text does not care about the switch: it is the script's line.
	if s, ok := Text("x.one"); !ok || s != "vibe one" {
		t.Errorf("Text: %q %v", s, ok)
	}
	if _, ok := Text("x.two"); ok {
		t.Error("Text found a row the script lacks")
	}
}

// TestSprintfAndErrorfFollowTheSwitch: the formats of the log and the
// refusals read the script's line while VibeMode is on and the script has one,
// are put together like fmt does, and are the plain format in every other
// case: switched off, no row, never an id.
func TestSprintfAndErrorfFollowTheSwitch(t *testing.T) {
	defer UseScriptForTest(map[string]string{
		"x.built": "%s is up, you got %d now",
		"x.cause": "nope, because %w",
	})()
	Set(false)
	defer Set(false)
	plainBuilt, plainCause := "Built %s (you have %d).", "Refused: %w"
	boom := errors.New("boom")

	if got, want := Sprintf("x.built", plainBuilt, "Hut", 3), fmt.Sprintf(plainBuilt, "Hut", 3); got != want {
		t.Errorf("off: Sprintf reads %q, want %q", got, want)
	}
	if got, want := Errorf("x.cause", plainCause, boom).Error(), "Refused: boom"; got != want {
		t.Errorf("off: Errorf reads %q, want %q", got, want)
	}

	Set(true)
	if got, want := Sprintf("x.built", plainBuilt, "Hut", 3), "Hut is up, you got 3 now"; got != want {
		t.Errorf("on: Sprintf reads %q, want %q", got, want)
	}
	err := Errorf("x.cause", plainCause, boom)
	if got, want := err.Error(), "nope, because boom"; got != want {
		t.Errorf("on: Errorf reads %q, want %q", got, want)
	}
	if !errors.Is(err, boom) {
		t.Error("on: Errorf lost the error it wraps")
	}
	// A row the script lacks reads plain, with the arguments put in, and the
	// id never shows.
	if got, want := Sprintf("x.nothing", plainBuilt, "Hut", 3), "Built Hut (you have 3)."; got != want {
		t.Errorf("on, no row: Sprintf reads %q, want %q", got, want)
	}
	if got := Errorf("x.nothing", "plain").Error(); got != "plain" {
		t.Errorf("on, no row: Errorf reads %q", got)
	}
}

func TestUseScriptForTestRestores(t *testing.T) {
	before := IDs()
	restore := UseScriptForTest(map[string]string{"b": "2", "a": "1"})
	if got := IDs(); len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Errorf("the loaned script's ids are %v", got)
	}
	restore()
	if got := IDs(); len(got) != len(before) {
		t.Errorf("after restore the script has %d rows, had %d", len(got), len(before))
	}
}

// TestTheShippedScript: the sheets in the binary read cleanly, and the rows
// come out in order.
func TestTheShippedScript(t *testing.T) {
	if p := Problems(); len(p) > 0 {
		t.Errorf("the shipped script has problems: %q", p)
	}
	ids := IDs()
	if len(ids) == 0 {
		t.Fatal("the shipped script has no rows")
	}
	if !sort.StringsAreSorted(ids) {
		t.Error("IDs are not in order")
	}
	for _, id := range ids {
		if s, ok := Text(id); !ok || s == "" {
			t.Errorf("%s is listed and has no text", id)
		}
	}
}
