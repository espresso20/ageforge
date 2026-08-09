package ui

import (
	"testing"

	"github.com/espresso20/ageforge/game"
)

// oneSecondTick is a snapshot whose tick is exactly one second, so a tick count
// and a second count are the same number and the cases below read as durations.
var oneSecondTick = game.GameState{TickIntervalMs: 1000, SpeedMultiplier: 1.0}

func TestFormatTicks(t *testing.T) {
	cases := []struct {
		ticks int
		want  string
	}{
		{0, "~0s"},
		{-5, "~0s"}, // a countdown past its floor still prints something sane
		{1, "~1s"},
		{38, "~38s"},
		{59, "~59s"},
		{60, "~1m"}, // a zero second unit is dropped, not printed as "1m 0s"
		{284, "~4m 44s"},
		{3600, "~1h"},
		{4320, "~1h 12m"},
		{86400, "~1d"},
		{97200, "~1d 3h"},
	}
	for _, tc := range cases {
		if got := formatTicks(tc.ticks, oneSecondTick); got != tc.want {
			t.Errorf("formatTicks(%d) = %q, want %q", tc.ticks, got, tc.want)
		}
	}
}

// TestFormatTicksUsesSnapshotInterval pins the conversion to state.TickIntervalMs.
// That field ALREADY has the tick-speed bonus and the speed multiplier folded in
// (see GetState), so the helper must not apply SpeedMultiplier a second time —
// a doubled multiplier here would quietly halve every countdown in the game.
func TestFormatTicksUsesSnapshotInterval(t *testing.T) {
	fast := game.GameState{TickIntervalMs: 500, SpeedMultiplier: 4.0}
	if got := formatTicks(120, fast); got != "~1m" {
		t.Errorf("120 ticks at 500ms = %q, want ~1m (SpeedMultiplier must not be reapplied)", got)
	}
}

// TestFormatTicksZeroValueState guards the fallback: a zero-valued GameState
// (every test that builds one by hand) must not divide by a zero interval.
func TestFormatTicksZeroValueState(t *testing.T) {
	if got := formatTicks(30, game.GameState{}); got == "" || got == "~0s" {
		t.Errorf("zero-valued state fell back badly: %q", got)
	}
}

func TestFormatTickRange(t *testing.T) {
	if got := formatTickRange(60, 100, oneSecondTick); got != "~1m – 1m 40s" {
		t.Errorf("formatTickRange(60,100) = %q, want %q", got, "~1m – 1m 40s")
	}
	// A degenerate or inverted range collapses to one reading so callers do not
	// have to pre-check the bounds.
	if got := formatTickRange(90, 90, oneSecondTick); got != "~1m 30s" {
		t.Errorf("equal bounds = %q, want ~1m 30s", got)
	}
	if got := formatTickRange(90, 10, oneSecondTick); got != "~1m 30s" {
		t.Errorf("inverted bounds = %q, want ~1m 30s", got)
	}
}
