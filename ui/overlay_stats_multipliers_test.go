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
	// Each fragment is wrapped in its own sign-colored tag; positives are green.
	want := "[green]Research +10%[-] [gray]·[-] [green]Wonders +5%[-] [gray]·[-] [green]Morale ×1.18[-]"
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
	if got := summarizeBreakdown(mods); got != "[green]Research +20%[-]" {
		t.Errorf("summarizeBreakdown = %q, want %q", got, "[green]Research +20%[-]")
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

// TestRenderActiveMultipliers_SignColors proves each contributing source is
// colored by its own sign: a positive source gets [green], a negative [red].
func TestRenderActiveMultipliers_SignColors(t *testing.T) {
	state := game.GameState{
		SpeedMultiplier: 1.0,
		Modifiers: []game.Modifier{
			{Source: "research", Target: "production_all", Op: game.OpAdd, Value: 0.20},
			{Source: "event:Famine", Target: "production_all", Op: game.OpAdd, Value: -0.05},
		},
	}
	out := renderActiveMultipliers(state)
	if !strings.Contains(out, "[green]Research +20%[-]") {
		t.Errorf("positive source should be green-tagged:\n%s", out)
	}
	if !strings.Contains(out, "[red]Event: Famine -5%[-]") {
		t.Errorf("negative source should be red-tagged:\n%s", out)
	}
	// Net +15% → green headline.
	if !strings.Contains(out, "[green]+15%[-]") {
		t.Errorf("net-positive headline should be green:\n%s", out)
	}
}

// TestRenderActiveMultipliers_OpposingSourcesStillRender guards the net-zero
// collapse fix: a target whose sources cancel to ×1.0 must STILL render, with
// both the positive and negative fragments visible.
func TestRenderActiveMultipliers_OpposingSourcesStillRender(t *testing.T) {
	state := game.GameState{
		SpeedMultiplier: 1.0,
		Modifiers: []game.Modifier{
			{Source: "research", Target: "gather_rate", Op: game.OpAdd, Value: 0.10},
			{Source: "event:Drought", Target: "gather_rate", Op: game.OpAdd, Value: -0.10},
		},
	}
	out := renderActiveMultipliers(state)
	if strings.Contains(out, "No active multipliers") {
		t.Fatalf("opposing sources netting to ~1.0 must still render a line:\n%s", out)
	}
	if !strings.Contains(out, "[green]Research +10%[-]") {
		t.Errorf("positive fragment missing from net-zero row:\n%s", out)
	}
	if !strings.Contains(out, "[red]Event: Drought -10%[-]") {
		t.Errorf("negative fragment missing from net-zero row:\n%s", out)
	}
	// Net ~0% → white headline (neither bonus nor penalty).
	if !strings.Contains(out, "[white]") {
		t.Errorf("net-zero headline should be white:\n%s", out)
	}
}

// TestRenderActiveMultipliers_CapacityTargetsExcluded guards the
// isPanelMultiplier filter: flat/capacity targets (population, "all", bare
// storage keys) are not rate multipliers and must not appear here as percents.
func TestRenderActiveMultipliers_CapacityTargetsExcluded(t *testing.T) {
	state := game.GameState{
		SpeedMultiplier: 1.0,
		Modifiers: []game.Modifier{
			{Source: "prestige", Target: "population", Op: game.OpAdd, Value: 2.0},
			{Source: "wonders", Target: "all", Op: game.OpAdd, Value: 20.0},
			{Source: "research", Target: "food", Op: game.OpAdd, Value: 0.50},
		},
	}
	out := renderActiveMultipliers(state)
	if strings.Contains(out, "Population") {
		t.Errorf("population (capacity) must not render in the multiplier panel:\n%s", out)
	}
	if strings.Contains(out, "All ") || strings.Contains(out, "+2000%") {
		t.Errorf("\"all\" (flat) must not render as a percentage:\n%s", out)
	}
	if !strings.Contains(out, "No active multipliers") {
		t.Errorf("only capacity/flat targets present — panel should be empty:\n%s", out)
	}
}

// TestStatsProvider_ActiveEventEffects proves the stats overlay now renders
// each active event's ongoing effects beneath it, color-coded by sign: a
// negative production effect gets a [red] "food" line, a positive one [green].
func TestStatsProvider_ActiveEventEffects(t *testing.T) {
	state := game.GameState{
		SpeedMultiplier: 1.0,
		ActiveEvents: []game.ActiveEventState{
			{
				Name:      "Famine",
				Key:       "famine",
				TicksLeft: 8,
				Effects: []game.EventEffectInfo{
					{Type: "production", Target: "food", Value: -3.0},
				},
			},
			{
				Name:      "Gold Rush",
				Key:       "gold_rush",
				TicksLeft: 12,
				Effects: []game.EventEffectInfo{
					{Type: "production", Target: "gold", Value: 1.0},
				},
			},
		},
	}
	out := statsProvider(state, 0)

	if !strings.Contains(out, "[red]    food -3.0/t[-]") {
		t.Errorf("negative production effect should render red-tagged:\n%s", out)
	}
	if !strings.Contains(out, "[green]    gold +1.0/t[-]") {
		t.Errorf("positive production effect should render green-tagged:\n%s", out)
	}
}

// TestIsPanelMultiplier covers the rate/flat classification directly.
func TestIsPanelMultiplier(t *testing.T) {
	rate := []string{
		"production_all", "gather_rate", "tick_speed", "military_power",
		"expedition_reward", "research_speed", "build_cost", "food_rate", "iron_rate",
	}
	for _, tgt := range rate {
		if !isPanelMultiplier(tgt) {
			t.Errorf("isPanelMultiplier(%q) = false, want true", tgt)
		}
	}
	flat := []string{"population", "all", "food", "culture", "data"}
	for _, tgt := range flat {
		if isPanelMultiplier(tgt) {
			t.Errorf("isPanelMultiplier(%q) = true, want false", tgt)
		}
	}
}
