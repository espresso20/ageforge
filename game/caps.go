package game

import (
	"math"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/pkg/textfmt"
)

// Bonus caps: where a pool of bonuses stops counting, and how the game says
// so.
//
// Bonuses of one kind add together into a pool ("+30% all production" and
// "+20% all production" make +50%), and the engine applies some pools only up
// to a limit:
//
//   - all production, and each resource's own production: the multiplier
//     1 + pool is held between productionFloor and productionCap (x0.1 to
//     x3, so +200% is the most that counts);
//   - worker output: no cap, the same floor;
//   - research speed: a tech always takes at least one tick, so +100% is the
//     most that counts;
//   - building costs: never under 10% of the listed price, never over it;
//   - game speed: a tick is never shorter than MinTickInterval.
//
// Military power and expedition rewards have no limit.
//
// These limits are a design decision that is still open (the tech tree
// redesign brings new stacking rules). Until then the game is honest about
// them: poolApplied is the one place that says what of a pool counts, the
// engine applies the pools through it, and every list of bonuses (the Stats
// and Research panels, milestone and wonder rewards, the log line a bonus
// is earned with) reads the same numbers through BonusPool and says
// "capped" where a limit is holding a bonus back.

// BonusPool is one pool of bonuses and what the engine applies of it.
type BonusPool struct {
	// Target is the pool's key: "production_all", "gold_rate",
	// "research_speed", "build_cost", "tick_speed", "gather_rate".
	Target string
	// Earned is every bonus in the pool added together (+4.05 is +405%).
	Earned float64
	// Applied is what counts (+2.00 when the x3 cap holds the pool).
	Applied float64
	// Limit is the cap or floor holding the pool, as a bonus (+2.00, -0.90),
	// and Limited whether one is: Applied differs from Earned.
	Limit   float64
	Limited bool
	// limitArg is the one input a limit depends on besides the pool's total:
	// the player speed multiplier, for game speed.
	limitArg float64
}

// Counts is how much of one bonus of v counts right now: all of it, part of
// it or none. With held set the bonus is already in the pool (a researched
// tech, a completed milestone), and what counts is what the pool would lose
// without it; otherwise it is what the pool would gain with it.
func (p BonusPool) Counts(v float64, held bool) float64 {
	without := p.Earned
	if held {
		without -= v
	}
	a, _, _ := poolApplied(p.Target, without+v, p.limitArg)
	b, _, _ := poolApplied(p.Target, without, p.limitArg)
	return a - b
}

// poolLimitEps is the slack under which a bonus counts as applied in full.
const poolLimitEps = 1e-9

// poolBounds is the range a pool's multiplier, 1 + its total, is held to
// (an infinite end is no limit). arg is the player speed multiplier for
// "tick_speed" and unused otherwise. Every limit on a bonus pool lives here.
func poolBounds(target string, arg float64) (lo, hi float64) {
	lo, hi = math.Inf(-1), math.Inf(1)
	switch {
	case target == "production_all":
		lo, hi = productionFloor, productionCap
	case target == "gather_rate":
		lo = productionFloor
	case target == "research_speed":
		hi = 2 // +100%: a tech takes one tick at least
	case target == "build_cost":
		lo, hi = buildCostFloor, buildCostCap
	case target == "tick_speed":
		hi = float64(BaseTickInterval) / float64(MinTickInterval) / math.Max(arg, 1)
	case strings.HasSuffix(target, "_rate"):
		lo, hi = productionFloor, productionCap
	}
	return lo, hi
}

// poolFactor is the multiplier the engine applies for a pool that has
// earned bonuses in all: 1 + earned, held to the pool's bounds.
func poolFactor(target string, earned float64) float64 {
	lo, hi := poolBounds(target, 1)
	return clamp(1.0+earned, lo, hi)
}

// poolApplied is what counts of a pool that has earned bonuses in all: the
// applied bonus, the limit it ran into (as a bonus) and whether it ran into
// one.
func poolApplied(target string, earned, arg float64) (applied, limit float64, limited bool) {
	lo, hi := poolBounds(target, arg)
	switch f := 1.0 + earned; {
	case f > hi+poolLimitEps:
		return hi - 1, hi - 1, true
	case f < lo-poolLimitEps:
		return lo - 1, lo - 1, true
	}
	return earned, 0, false
}

// hasPoolLimit reports whether target is a pool a limit can hold.
func hasPoolLimit(target string) bool {
	switch target {
	case "production_all", "gather_rate", "research_speed", "build_cost", "tick_speed":
		return true
	}
	return strings.HasSuffix(target, "_rate")
}

// bonusPoolsLocked is every pool a limit can hold, with what it has earned
// and what counts, by target. All production is always there. Caller holds
// the lock.
func (ge *GameEngine) bonusPoolsLocked(r *Resolver) map[string]BonusPool {
	out := make(map[string]BonusPool)
	targets := append(r.Targets(), "production_all")
	for _, target := range targets {
		if _, done := out[target]; done || !hasPoolLimit(target) {
			continue
		}
		out[target] = ge.bonusPoolLocked(r, target)
	}
	return out
}

// bonusPoolLocked is one pool as it stands. Caller holds the lock.
func (ge *GameEngine) bonusPoolLocked(r *Resolver, target string) BonusPool {
	p := BonusPool{Target: target, Earned: r.AddTotal(target), limitArg: ge.speedMultiplier}
	p.Applied, p.Limit, p.Limited = poolApplied(target, p.Earned, p.limitArg)
	return p
}

// effectPool is the pool an effect adds to and whether it adds to one: the
// target of a tech's or wonder's "bonus" and of a milestone's
// "permanent_bonus", the type of a timed event's effect ("production_all",
// "tick_speed", "<res>_rate").
func effectPool(e config.Effect) (string, bool) {
	switch e.Type {
	case "bonus", "permanent_bonus":
		return e.Target, hasPoolLimit(e.Target)
	case "production_all", "tick_speed":
		return e.Type, true
	}
	if strings.HasSuffix(e.Type, "_rate") {
		return e.Type, true
	}
	return "", false
}

// EffectPool is effectPool for the panels.
func EffectPool(e config.Effect) (string, bool) { return effectPool(e) }

// PoolLimitText names the limit a pool is held to, for a "capped" note:
// "+200%" for the x3 production cap, "one tick" for research.
func PoolLimitText(p BonusPool) string {
	if p.Target == "research_speed" {
		return "+100%, one tick a tech"
	}
	return textfmt.SignedPercent(p.Limit)
}

// CapNote is the short note beside one bonus (or penalty) of v in pool p: ""
// when all of it counts, "capped: no effect now" when none does, "capped:
// +5% of it counts now" in between. held says it is already in the pool.
func CapNote(p BonusPool, v float64, held bool) string {
	if !hasPoolLimit(p.Target) || v == 0 {
		return ""
	}
	counts := p.Counts(v, held)
	switch {
	case math.Abs(counts-v) <= 1e-6*math.Max(1, math.Abs(v)):
		return ""
	case math.Abs(counts) <= 1e-6:
		return "capped: no effect now"
	}
	return "capped: " + textfmt.SignedPercent(counts) + " of it counts now"
}

// PoolNote is the note beside a pool's total: "" when all of it counts,
// "capped at +200%: +405% earned" when a limit holds it.
func PoolNote(p BonusPool) string {
	if !p.Limited {
		return ""
	}
	if p.Target == "research_speed" {
		return "capped: a tech takes one tick at least, so only +100% counts"
	}
	return "capped at " + textfmt.SignedPercent(p.Limit) + ": " + textfmt.SignedPercent(p.Earned) + " earned"
}

// capNoteLocked is CapNote for a bonus the engine is about to add to, or has
// just added to, a pool: the note a log line carries. Caller holds the lock.
func (ge *GameEngine) capNoteLocked(e config.Effect, held bool) string {
	target, ok := effectPool(e)
	if !ok {
		return ""
	}
	return CapNote(ge.bonusPoolLocked(ge.buildResolver(), target), e.Value, held)
}

// capLinesLocked is one log line for every effect a cap keeps from counting
// in full: "  → +20% all production is capped: no effect now. All
// production counts up to +200%, and you have earned +405%." Empty when
// every effect counts. Caller holds the lock.
func (ge *GameEngine) capLinesLocked(effects []config.Effect, held bool) []string {
	var out []string
	r := ge.buildResolver()
	seen := map[string]bool{}
	for _, e := range effects {
		target, ok := effectPool(e)
		if !ok || seen[target] {
			continue
		}
		p := ge.bonusPoolLocked(r, target)
		note := CapNote(p, e.Value, held)
		if note == "" {
			continue
		}
		seen[target] = true
		earned := p.Earned
		if !held {
			earned += e.Value
		}
		name := EffectTargetName(target)
		line := "  → " + textfmt.SignedPercent(e.Value) + " " + name + " is " + note + ". "
		if target == "research_speed" {
			line += "A tech takes one tick at least, so research speed counts up to +100%, and you have earned " + textfmt.SignedPercent(earned) + "."
		} else {
			_, limit, _ := poolApplied(target, earned, ge.speedMultiplier)
			line += textfmt.Capitalize(name) + " counts up to " + textfmt.SignedPercent(limit) + ", and you have earned " + textfmt.SignedPercent(earned) + "."
		}
		out = append(out, line)
	}
	return out
}
