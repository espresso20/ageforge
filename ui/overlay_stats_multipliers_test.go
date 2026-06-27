package ui

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/game"
)

func TestMultiplierSourceLabel(t *testing.T) {
	cases := map[string]string{
		"research":               "Research",
		"prestige":               "Prestige",
		"wonders":                "Wonders",
		"permanent":              "Permanent",
		"morale":                 "Morale",
		"event":                  "Event",
		"event:Peaceful Century": "Event: Peaceful Century",
		"mystery_source":         "Mystery_source", // capitalize fallback
	}
	for in, want := range cases {
		if got := multiplierSourceLabel(in); got != want {
			t.Errorf("multiplierSourceLabel(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestMultiplierTargetLabel(t *testing.T) {
	cases := map[string]string{
		"production_all": "All Production",
		"tick_speed":     "Tick Speed",
		"gather_rate":    "Gather Rate",
		"food_rate":      "Food Rate", // <res>_rate via formatBonusName
	}
	for in, want := range cases {
		if got := multiplierTargetLabel(in); got != want {
			t.Errorf("multiplierTargetLabel(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSummarizeBreakdown_AdditiveAndMul(t *testing.T) {
	// production_all: research +10%, wonders +5%, morale ×1.18.
	mods := []game.Modifier{
		{Source: "research", Target: "production_all", Op: game.OpAdd, Value: 0.10},
		{Source: "wonders", Target: "production_all", Op: game.OpAdd, Value: 0.05},
		{Source: "morale", Target: "production_all", Op: game.OpMul, Value: 1.18},
	}
	got := summarizeBreakdown(mods)
	want := "Research +10% · Wonders +5% · Morale ×1.18"
	if got != want {
		t.Errorf("summarizeBreakdown = %q, want %q", got, want)
	}
}

func TestSummarizeBreakdown_MergesSameSource(t *testing.T) {
	// Two research contributions on one target should merge into one entry.
	mods := []game.Modifier{
		{Source: "research", Target: "food_rate", Op: game.OpAdd, Value: 0.12},
		{Source: "research", Target: "food_rate", Op: game.OpAdd, Value: 0.08},
	}
	if got := summarizeBreakdown(mods); got != "Research +20%" {
		t.Errorf("summarizeBreakdown = %q, want %q", got, "Research +20%")
	}
}

func TestSummarizeBreakdown_DropsNoops(t *testing.T) {
	mods := []game.Modifier{
		{Source: "research", Target: "gather_rate", Op: game.OpAdd, Value: 0.0},
		{Source: "morale", Target: "gather_rate", Op: game.OpMul, Value: 1.0},
	}
	if got := summarizeBreakdown(mods); got != "" {
		t.Errorf("summarizeBreakdown of no-ops = %q, want empty", got)
	}
}

// TestRenderActiveMultipliers_ActiveEvent proves the bug fix: a timed event's
// production_all/tick_speed contributions now surface in the panel (they used
// to be invisible because the old hand-aggregation never read events).
func TestRenderActiveMultipliers_ActiveEvent(t *testing.T) {
	state := game.GameState{
		SpeedMultiplier: 1.0,
		Modifiers: []game.Modifier{
			{Source: "research", Target: "production_all", Op: game.OpAdd, Value: 0.10},
			{Source: "wonders", Target: "production_all", Op: game.OpAdd, Value: 0.05},
			{Source: "morale", Target: "production_all", Op: game.OpMul, Value: 1.18},
			{Source: "prestige", Target: "tick_speed", Op: game.OpAdd, Value: 0.05},
			{Source: "event:Peaceful Century", Target: "tick_speed", Op: game.OpAdd, Value: 0.01},
		},
	}
	out := renderActiveMultipliers(state)

	if !strings.Contains(out, "Event: Peaceful Century +1%") {
		t.Errorf("active event bonus missing from panel:\n%s", out)
	}
	if !strings.Contains(out, "All Production") || !strings.Contains(out, "Morale ×1.18") {
		t.Errorf("production_all line malformed:\n%s", out)
	}
	if !strings.Contains(out, "Tick Speed") || !strings.Contains(out, "Prestige +5%") {
		t.Errorf("tick_speed line malformed:\n%s", out)
	}
}

func TestRenderActiveMultipliers_SpeedMultiplierLine(t *testing.T) {
	state := game.GameState{
		SpeedMultiplier: 2.0,
		Modifiers:       nil,
	}
	out := renderActiveMultipliers(state)
	if !strings.Contains(out, "Game Speed") || !strings.Contains(out, "×2.00") {
		t.Errorf("speed multiplier line missing:\n%s", out)
	}
}

// TestRenderActiveMultipliers_Empty guards nil-safety on a fresh game: no
// modifiers, no speed multiplier → the "No active multipliers" line, no panic.
func TestRenderActiveMultipliers_Empty(t *testing.T) {
	state := game.GameState{SpeedMultiplier: 1.0, Modifiers: nil}
	out := renderActiveMultipliers(state)
	if !strings.Contains(out, "No active multipliers") {
		t.Errorf("empty state should report no multipliers, got:\n%s", out)
	}
}

// TestRenderActiveMultipliers_NoopTargetsHidden ensures a target whose net
// effect is ×1.0 (e.g. only no-op contributions) does not render a line.
func TestRenderActiveMultipliers_NoopTargetsHidden(t *testing.T) {
	state := game.GameState{
		SpeedMultiplier: 1.0,
		Modifiers: []game.Modifier{
			{Source: "morale", Target: "production_all", Op: game.OpMul, Value: 1.0},
		},
	}
	out := renderActiveMultipliers(state)
	if !strings.Contains(out, "No active multipliers") {
		t.Errorf("no-op target should be hidden, got:\n%s", out)
	}
}
