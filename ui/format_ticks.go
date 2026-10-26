package ui

import (
	"fmt"
	"time"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
)

// Wall-clock rendering of tick counts.
//
// Ticks are the engine's unit and they are meaningless to a player: "216 ticks"
// answers no question anyone actually has. Every player-facing duration and
// countdown goes through the helpers here instead, so the whole game speaks in
// minutes and seconds and speaks it the same way.
//
// The conversion is deliberately anchored on state.TickIntervalMs, which GetState
// computes as BaseTickInterval / ((1 + tickSpeedBonus) * speedMultiplier): the
// tick-speed bonus (and the dev console's /speed override) are ALREADY folded
// into it. Do not scale it again here; that would double-count the very thing
// the field exists to express. A consequence worth knowing: these readings
// move as the player's tick speed does, which is correct — a research boost
// really does shorten the wait.
//
// Raw tick counts survive in exactly one place: the `dumplog` debug dump, where
// they are printed ALONGSIDE the wall-clock reading because that file is read by
// us, not by players.

// tickInterval is the wall-clock length of one tick for this snapshot. Falls
// back to the engine's base interval if the snapshot has no sane value (a
// zero-valued GameState in a test, mostly), so callers never divide by nothing.
func tickInterval(state game.GameState) time.Duration {
	if state.TickIntervalMs <= 0 {
		return game.BaseTickInterval
	}
	return time.Duration(state.TickIntervalMs) * time.Millisecond
}

// formatTicks renders a tick count as an approximate wall-clock duration:
// "~38s", "~4m 44s", "~1h 12m". The tilde is part of the format and is honest —
// the reading drifts as tick speed changes.
//
// Zero, negative and sub-second counts all render "~0s" rather than an empty
// string or a negative time, so a countdown that has just hit its floor still
// prints something sane.
func formatTicks(ticks int, state game.GameState) string {
	return "~" + textfmt.Duration(time.Duration(ticks)*tickInterval(state))
}

// formatTickRange renders a [min, max] tick range as one approximate span:
// "~2m – 3m 20s". Used for expedition durations, which are rolled per launch and
// so can only be previewed as a range. One tilde covers both ends.
//
// A degenerate or inverted range collapses to a single reading, so callers do
// not have to pre-check whether the two bounds happen to be equal.
func formatTickRange(minTicks, maxTicks int, state game.GameState) string {
	if maxTicks <= minTicks {
		return formatTicks(minTicks, state)
	}
	iv := tickInterval(state)
	return fmt.Sprintf("~%s – %s",
		textfmt.Duration(time.Duration(minTicks)*iv),
		textfmt.Duration(time.Duration(maxTicks)*iv))
}
