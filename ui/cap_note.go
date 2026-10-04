package ui

import (
	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// Capped bonuses (game/caps.go). Some bonus pools count only up to a limit:
// all production and each resource's own production up to +200%, research
// speed up to +100%. Wherever a panel lists a bonus, these helpers put a
// short "capped" note beside one a limit is holding back, read off the same
// numbers the engine applies (GameState.Pools).

// capNote is the note for one bonus, "" when a limit is not in its way.
//
// A bonus the player does not hold yet (a tech to research, a wonder to
// build, a milestone to finish, a festival to pay for) says what it would
// add: "capped: no effect now", or "capped: +5% of it counts now".
//
// A bonus the player holds (held) says where its pool stands, since no one
// bonus of a pool is the one that goes uncounted: "capped at +200%: +405%
// earned".
func capNote(state game.GameState, eff config.Effect, held bool) string {
	target, ok := game.EffectPool(eff)
	if !ok {
		return ""
	}
	pool, ok := state.Pools[target]
	if !ok {
		pool = game.BonusPool{Target: target}
	}
	if held {
		return game.PoolNote(pool)
	}
	return game.CapNote(pool, eff.Value, false)
}

// capTag is capNote as a tag to append to a line: " [yellow](capped: no
// effect now)[-]", or "". after is the color the line carries on in
// ("gray" inside a dim line, "-" to reset).
func capTag(state game.GameState, eff config.Effect, held bool, after string) string {
	note := capNote(state, eff, held)
	if note == "" {
		return ""
	}
	return " [yellow](" + note + ")[" + after + "]"
}

// poolTag is the note beside a pool's total: " [yellow]capped at +200%:
// +405% earned[-]", or "" when all of the pool counts.
func poolTag(state game.GameState, target string) string {
	pool, ok := state.Pools[target]
	if !ok {
		return ""
	}
	note := game.PoolNote(pool)
	if note == "" {
		return ""
	}
	return " [yellow]" + note + "[-]"
}
