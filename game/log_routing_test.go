package game

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// A finished building is a routine line (the Buildings panel shows the
// count) and a finished wonder stays notable. A completion's flavor quip
// takes its line's category, so the main log never shows a quip alone.
func TestBuildCompletionLogCategory(t *testing.T) {
	defs := config.BuildingByKey()
	if got := buildDoneLog(defs["hut"]); got != LogRoutine {
		t.Errorf("a finished hut logs as %q, want %q", got, LogRoutine)
	}
	wonders := 0
	for _, def := range defs {
		if def.Category != "wonder" {
			continue
		}
		wonders++
		if got := buildDoneLog(def); got != "success" {
			t.Errorf("a finished %s logs as %q, want success", def.Name, got)
		}
	}
	if wonders == 0 {
		t.Fatal("no wonders in config")
	}

	ge := NewGameEngine()
	ge.SeedRNG(1)
	ge.log = nil
	for i := 0; i < 30; i++ {
		ge.finishBuild(BuildQueueItem{BuildingKey: "hut"})
	}
	quips := 0
	for i, e := range ge.log {
		switch {
		case strings.Contains(e.Message, "Hut built"):
			if e.Type != LogRoutine {
				t.Errorf("%q logs as %q", e.Message, e.Type)
			}
		case strings.HasPrefix(e.Message, "  [gray]"):
			quips++
			if e.Type != LogRoutine || i == 0 || !strings.Contains(ge.log[i-1].Message, "Hut built") {
				t.Errorf("quip %q logs as %q after %q", e.Message, e.Type, ge.log[max(i-1, 0)].Message)
			}
		}
	}
	if quips == 0 {
		t.Error("30 completions drew no quip; the quip path went untested")
	}
}
