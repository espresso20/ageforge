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
		"storage":                "Storage",        // capitalize fallback: the storage buildings' boost
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
		"production_all": "All production",
		"tick_speed":     "Game speed",
		"gather_rate":    "Worker output",
		"food_rate":      "Food production", // <res>_rate via game.EffectTargetName
	}
	for in, want := range cases {
		if got := multiplierTargetLabel(in); got != want {
			t.Errorf("multiplierTargetLabel(%q) = %q, want %q", in, got, want)
		}
	}
}

// srcTexts is a row's sources as the page writes them, with a sign marker.
func srcTexts(r multRow) []string {
	var out []string
	for _, s := range r.sources {
		sign := "+"
		if s.neg {
			sign = "-"
		}
		out = append(out, sign+" "+s.text())
	}
	return out
}

func multRowOf(t *testing.T, rows []multRow, target string) multRow {
	t.Helper()
	for _, r := range rows {
		if r.target == target {
			return r
		}
	}
	t.Fatalf("no %q row in %+v", target, rows)
	return multRow{}
}

func TestSummarizeBreakdown_AdditiveAndMul(t *testing.T) {
	// production_all: research +10%, wonders +5%, morale x1.18.
	mods := []game.Modifier{
		{Source: "research", Target: "production_all", Op: game.OpAdd, Value: 0.10},
		{Source: "wonders", Target: "production_all", Op: game.OpAdd, Value: 0.05},
		{Source: "morale", Target: "production_all", Op: game.OpMul, Value: 1.18},
	}
	got := summarizeBreakdown(mods)
	want := []multSrc{{"Research", "+10%", false}, {"Wonders", "+5%", false}, {"Morale", "×1.18", false}}
	if len(got) != len(want) {
		t.Fatalf("summarizeBreakdown = %+v, want %+v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("source %d = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestSummarizeBreakdown_MergesSameSource(t *testing.T) {
	// Two research contributions on one target merge into one entry.
	mods := []game.Modifier{
		{Source: "research", Target: "food_rate", Op: game.OpAdd, Value: 0.12},
		{Source: "research", Target: "food_rate", Op: game.OpAdd, Value: 0.08},
	}
	got := summarizeBreakdown(mods)
	if len(got) != 1 || got[0].text() != "Research +20%" || got[0].neg {
		t.Errorf("summarizeBreakdown = %+v, want one Research +20%%", got)
	}
}

func TestSummarizeBreakdown_DropsNoops(t *testing.T) {
	mods := []game.Modifier{
		{Source: "research", Target: "gather_rate", Op: game.OpAdd, Value: 0.0},
		{Source: "morale", Target: "gather_rate", Op: game.OpMul, Value: 1.0},
	}
	if got := summarizeBreakdown(mods); len(got) != 0 {
		t.Errorf("summarizeBreakdown of no-ops = %+v, want none", got)
	}
}

// TestActiveMultipliers_ActiveEvent: a timed event's production_all and
// tick_speed contributions surface in the table.
func TestActiveMultipliers_ActiveEvent(t *testing.T) {
	state := game.GameState{
		Modifiers: []game.Modifier{
			{Source: "research", Target: "production_all", Op: game.OpAdd, Value: 0.10},
			{Source: "wonders", Target: "production_all", Op: game.OpAdd, Value: 0.05},
			{Source: "morale", Target: "production_all", Op: game.OpMul, Value: 1.18},
			{Source: "prestige", Target: "tick_speed", Op: game.OpAdd, Value: 0.05},
			{Source: "event:Peaceful Century", Target: "tick_speed", Op: game.OpAdd, Value: 0.01},
		},
	}
	rows := activeMultipliers(state)
	speed := multRowOf(t, rows, "Game speed")
	if got := strings.Join(srcTexts(speed), "|"); !strings.Contains(got, "+ Event: Peaceful Century +1%") || !strings.Contains(got, "+ Prestige +5%") {
		t.Errorf("tick_speed sources: %s", got)
	}
	all := multRowOf(t, rows, "All production")
	if got := strings.Join(srcTexts(all), "|"); !strings.Contains(got, "Morale ×1.18") {
		t.Errorf("production_all sources: %s", got)
	}
}

// TestActiveMultipliers_Empty: a fresh game has no rows and does not panic.
func TestActiveMultipliers_Empty(t *testing.T) {
	if rows := activeMultipliers(game.GameState{Modifiers: nil}); len(rows) != 0 {
		t.Errorf("empty state has rows: %+v", rows)
	}
}

// TestActiveMultipliers_NoopTargetsHidden: a target whose net effect is x1.0
// from no-op contributions has no row.
func TestActiveMultipliers_NoopTargetsHidden(t *testing.T) {
	state := game.GameState{Modifiers: []game.Modifier{{Source: "morale", Target: "production_all", Op: game.OpMul, Value: 1.0}}}
	if rows := activeMultipliers(state); len(rows) != 0 {
		t.Errorf("no-op target has a row: %+v", rows)
	}
}

// TestActiveMultipliers_SignColors: each source carries its own sign and the
// row carries the net.
func TestActiveMultipliers_SignColors(t *testing.T) {
	state := game.GameState{
		Modifiers: []game.Modifier{
			{Source: "research", Target: "production_all", Op: game.OpAdd, Value: 0.20},
			{Source: "event:Famine", Target: "production_all", Op: game.OpAdd, Value: -0.05},
		},
	}
	r := multRowOf(t, activeMultipliers(state), "All production")
	if got := srcTexts(r); len(got) != 2 || got[0] != "+ Research +20%" || got[1] != "- Event: Famine -5%" {
		t.Errorf("sources = %q", got)
	}
	if r.pct < 14.9 || r.pct > 15.1 {
		t.Errorf("net = %v%%, want +15%%", r.pct)
	}
}

// TestActiveMultipliers_OpposingSourcesStillRender: a target whose sources
// cancel to x1.0 still has a row, with both fragments.
func TestActiveMultipliers_OpposingSourcesStillRender(t *testing.T) {
	state := game.GameState{
		Modifiers: []game.Modifier{
			{Source: "research", Target: "gather_rate", Op: game.OpAdd, Value: 0.10},
			{Source: "event:Drought", Target: "gather_rate", Op: game.OpAdd, Value: -0.10},
		},
	}
	rows := activeMultipliers(state)
	if len(rows) != 1 {
		t.Fatalf("opposing sources netting to ~1.0 must still have a row: %+v", rows)
	}
	if got := srcTexts(rows[0]); len(got) != 2 || got[0] != "+ Research +10%" || got[1] != "- Event: Drought -10%" {
		t.Errorf("sources = %q", got)
	}
	if rows[0].pct > 0.5 || rows[0].pct < -0.5 {
		t.Errorf("net = %v%%, want about 0", rows[0].pct)
	}
}

// TestActiveMultipliers_CapacityTargetsExcluded: flat and capacity targets
// (population, "all", bare storage keys) are not rate multipliers and have
// no row.
func TestActiveMultipliers_CapacityTargetsExcluded(t *testing.T) {
	state := game.GameState{
		Modifiers: []game.Modifier{
			{Source: "prestige", Target: "population", Op: game.OpAdd, Value: 2.0},
			{Source: "wonders", Target: "all", Op: game.OpAdd, Value: 20.0},
			{Source: "research", Target: "food", Op: game.OpAdd, Value: 0.50},
		},
	}
	if rows := activeMultipliers(state); len(rows) != 0 {
		t.Errorf("capacity or flat targets have rows: %+v", rows)
	}
}

// TestPrestigeRows_ActiveEventEffects: each active event's ongoing effects
// sit beneath it, a negative production effect bad and a positive one good.
func TestPrestigeRows_ActiveEventEffects(t *testing.T) {
	state := game.GameState{
		ActiveEvents: []game.ActiveEventState{
			{Name: "Famine", Key: "famine", TicksLeft: 8, Effects: []game.EventEffectInfo{{Type: "production", Target: "food", Value: -3.0}}},
			{Name: "Gold Rush", Key: "gold_rush", TicksLeft: 12, Effects: []game.EventEffectInfo{{Type: "production", Target: "gold", Value: 1.0}}},
		},
	}
	rows := prestigeRows(state)
	find := func(text string, k xk) {
		t.Helper()
		for _, r := range rows {
			if strings.TrimSpace(r.value) == text && r.k == k {
				return
			}
		}
		t.Errorf("no %q row of kind %v in %+v", text, k, rows)
	}
	find("-3 food/tick", xBad)
	find("+1 gold/tick", xGood)
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
