package ui

import (
	"fmt"
	"time"

	"github.com/espresso20/ageforge/game"
)

// Wall-clock rendering of tick counts.
//
// Ticks are the engine's unit and they are meaningless to a player: "216 ticks"
// answers no question anyone actually has. Every player-facing duration and
// countdown goes through the helpers here instead, so the whole game speaks in
// minutes and seconds and speaks it the same way.
//
// The conversion is deliberately anchored on state.TickIntervalMs, which GetState
// computes as BaseTickInterval / ((1 + tickSpeedBonus) * speedMultiplier) — the
// tick-speed bonus AND the speed multiplier are ALREADY folded into it. Do not
// divide by state.SpeedMultiplier again here; that would double-count the very
// thing the field exists to express. A consequence worth knowing: these readings
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
	return "~" + humanizeDuration(time.Duration(ticks)*tickInterval(state))
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
		humanizeDuration(time.Duration(minTicks)*iv),
		humanizeDuration(time.Duration(maxTicks)*iv))
}

// humanizeDuration renders d at two units of precision, largest first, dropping
// a zero second unit: "45s", "4m 44s", "5m", "1h 12m", "2h", "3d 4h".
//
// Two units is the ceiling on purpose. "1h 12m 07s" is false precision for a
// number that moves whenever tick speed does, and it makes a status line harder
// to scan than the thing it is describing.
func humanizeDuration(d time.Duration) string {
	if d <= 0 {
		return "0s"
	}
	total := int(d.Round(time.Second) / time.Second)
	switch {
	case total < 60:
		return fmt.Sprintf("%ds", total)
	case total < 3600:
		if s := total % 60; s != 0 {
			return fmt.Sprintf("%dm %ds", total/60, s)
		}
		return fmt.Sprintf("%dm", total/60)
	case total < 86400:
		if m := (total % 3600) / 60; m != 0 {
			return fmt.Sprintf("%dh %dm", total/3600, m)
		}
		return fmt.Sprintf("%dh", total/3600)
	default:
		if h := (total % 86400) / 3600; h != 0 {
			return fmt.Sprintf("%dd %dh", total/86400, h)
		}
		return fmt.Sprintf("%dd", total/86400)
	}
}
