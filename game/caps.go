package game

import (
	"math"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/pkg/textfmt"
)

// Bonus caps: where a pool of bonuses stops counting in full, and how the
// game says so.
//
// Bonuses of one kind add together into a pool ("+30% all production" and
// "+20% all production" make +50%), and the engine applies some pools only up
// to a limit:
//
//   - all production, and each resource's own production: a soft cap
//     (config.SoftCap, read through the ruleset). The pool applies in full
//     up to its knee, +200%, and a quarter of every point past it, so a
//     pool that has earned +320% applies +230%. The multiplier 1 + pool is
//     never under productionFloor (x0.1);
//   - worker output: no cap, the same floor;
//   - research speed: a tech always takes at least one tick, so +100% is the
//     most that counts;
//   - building costs: never under 10% of the listed price, never over it;
//   - game speed: a tick is never shorter than MinTickInterval.
//
// Military power and expedition rewards have no limit. Research bonuses on
// output are in no pool: they are a layer of their own, applied after these
// (tech_layer.go).
//
// The game is honest about every limit: poolApplied is the one place that
// says what of a pool counts, the engine applies the pools through it, and
// every list of bonuses (the Stats panel, milestone and wonder rewards, the
// rates breakdown, the log line a bonus is earned with) reads the same
// numbers through BonusPool. A bonus a hard limit holds back says "capped";
// one past a knee says it counts a quarter, and how much that is now.

// BonusPool is one pool of bonuses and what the engine applies of it.
type BonusPool struct {
	// Target is the pool's key: "production_all", "gold_rate",
	// "research_speed", "build_cost", "tick_speed", "gather_rate".
	Target string
	// Earned is every bonus in the pool added together (+4.05 is +405%).
	Earned float64
	// Applied is what counts (+2.5125 of a production pool that has earned
	// +4.05: +200% in full and a quarter of the rest).
	Applied float64
	// Limit is the cap, floor or knee holding the pool, as a bonus (+2.00,
	// -0.90), and Limited whether one is: Applied differs from Earned.
	Limit   float64
	Limited bool
	// Soft says the limit is a knee, not a wall: the pool still grows past
	// it, at a share of each point (config.SoftCap.PastKnee). False at a
	// hard cap or a floor.
	Soft bool
	// limitArg is the one input a limit depends on besides the pool's total:
	// the player speed multiplier, for game speed.
	limitArg float64
	// soft is the rule a production pool's knee follows, from the ruleset
	// of the engine (or snapshot) the pool was read off.
	soft config.SoftCap
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
	a, _, _ := poolApplied(p.Target, without+v, p.limitArg, p.soft)
	b, _, _ := poolApplied(p.Target, without, p.limitArg, p.soft)
	return a - b
}

// PoolOf is st's pool for target: the one the engine reported, or, for a
// pool nothing has joined yet, an empty one under the same rules.
func PoolOf(st GameState, target string) BonusPool {
	if p, ok := st.Pools[target]; ok {
		return p
	}
	return BonusPool{Target: target, soft: st.Ruleset().SoftCap()}
}

// poolLimitEps is the slack under which a bonus counts as applied in full.
const poolLimitEps = 1e-9

// poolBounds is the range a pool's multiplier, 1 + its total, is held to
// (an infinite end is no limit), and whether the pool has a knee instead of
// a top: a production pool follows the soft cap. arg is the player speed
// multiplier for "tick_speed" and unused otherwise. Every limit on a bonus
// pool lives here.
func poolBounds(target string, arg float64) (lo, hi float64, knee bool) {
	lo, hi = math.Inf(-1), math.Inf(1)
	switch {
	case target == "production_all":
		lo, knee = productionFloor, true
	case target == "gather_rate":
		lo = productionFloor
	case target == "research_speed":
		hi = 2 // +100%: a tech takes one tick at least
	case target == "build_cost":
		lo, hi = buildCostFloor, buildCostCap
	case target == "tick_speed":
		hi = float64(BaseTickInterval) / float64(MinTickInterval) / math.Max(arg, 1)
	case strings.HasSuffix(target, "_rate"):
		lo, knee = productionFloor, true
	}
	return lo, hi, knee
}

// poolFactor is the multiplier the engine applies for a pool that has
// earned bonuses in all: 1 + what the soft cap leaves of earned for a
// production pool, 1 + earned for any other, held to the pool's bounds.
func poolFactor(target string, earned float64, soft config.SoftCap) float64 {
	lo, hi, knee := poolBounds(target, 1)
	if knee {
		earned = soft.Applied(earned)
	}
	return clamp(1.0+earned, lo, hi)
}

// poolApplied is what counts of a pool that has earned bonuses in all: the
// applied bonus, the limit it ran into (as a bonus) and whether it ran into
// one. Past a knee the limit is the knee and the applied bonus goes on
// growing.
func poolApplied(target string, earned, arg float64, soft config.SoftCap) (applied, limit float64, limited bool) {
	lo, hi, knee := poolBounds(target, arg)
	switch f := 1.0 + earned; {
	case knee && earned > soft.Knee+poolLimitEps:
		return soft.Applied(earned), soft.Knee, true
	case f > hi+poolLimitEps:
		return hi - 1, hi - 1, true
	case f < lo-poolLimitEps:
		return lo - 1, lo - 1, true
	}
	return earned, 0, false
}

// poolPastKnee reports whether a pool that has earned bonuses in all is
// past its knee.
func poolPastKnee(target string, earned float64, soft config.SoftCap) bool {
	_, _, knee := poolBounds(target, 1)
	return knee && earned > soft.Knee+poolLimitEps
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
	soft := ge.rules.SoftCap()
	p := BonusPool{Target: target, Earned: r.AddTotal(target), limitArg: ge.speedMultiplier, soft: soft}
	p.Applied, p.Limit, p.Limited = poolApplied(target, p.Earned, p.limitArg, soft)
	p.Soft = p.Limited && poolPastKnee(target, p.Earned, soft)
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

// CapNote is the short note beside one bonus (or penalty) of v in pool p: ""
// when all of it counts. Past a production pool's knee it is "counts a
// quarter past +200%: +2.5% now". At a hard limit it is "capped: no effect
// now" when none of it counts and "capped: +5% of it counts now" in between.
// held says it is already in the pool.
func CapNote(p BonusPool, v float64, held bool) string {
	note, _ := capNote(p, v, held)
	return note
}

// CapPhrase is CapNote as the rest of a sentence about the bonus: "counts a
// quarter past +200%: +2.5% now", "is capped: no effect now". "" when all of
// it counts.
func CapPhrase(p BonusPool, v float64, held bool) string {
	note, knee := capNote(p, v, held)
	if note == "" || knee {
		return note
	}
	return "is " + note
}

// capNote is CapNote, and whether a knee (not a hard limit) is what holds
// the bonus back.
func capNote(p BonusPool, v float64, held bool) (note string, knee bool) {
	if !hasPoolLimit(p.Target) || v == 0 {
		return "", false
	}
	counts := p.Counts(v, held)
	if math.Abs(counts-v) <= 1e-6*math.Max(1, math.Abs(v)) {
		return "", false
	}
	without := p.Earned
	if held {
		without -= v
	}
	if poolPastKnee(p.Target, math.Max(without, without+v), p.soft) && !floored(p.Target, math.Min(without, without+v)) {
		return kneeText(p.soft) + ": " + pointsText(counts) + " now", true
	}
	if math.Abs(counts) <= 1e-6 {
		return "capped: no effect now", false
	}
	return "capped: " + textfmt.SignedPercent(counts) + " of it counts now", false
}

// floored reports whether a pool that has earned bonuses in all sits under
// its floor.
func floored(target string, earned float64) bool {
	lo, _, _ := poolBounds(target, 1)
	return 1.0+earned < lo-poolLimitEps
}

// kneeText is the soft cap in a few words: "counts a quarter past +200%".
func kneeText(soft config.SoftCap) string {
	return "counts " + shareText(soft.PastKnee) + " past " + textfmt.SignedPercent(soft.Knee)
}

// shareText names a share the way a person would: "a quarter", "half", "a
// third", or the percentage.
func shareText(share float64) string {
	switch {
	case math.Abs(share-0.25) < 1e-9:
		return "a quarter"
	case math.Abs(share-0.5) < 1e-9:
		return "half"
	case math.Abs(share-1.0/3) < 1e-9:
		return "a third"
	}
	return textfmt.Percent(share)
}

// pointsText is a bonus as percentage points to three figures, signed:
// "+2.5%", "+1.25%", "+251%". A quarter of a small bonus is not a whole
// number of points. The value is rounded to a hundredth of a point first,
// so the last bit of a float sum never decides between "+68.7%" and
// "+68.8%".
func pointsText(frac float64) string {
	return textfmt.RateValue(math.Round(frac*1e4)/100) + "%"
}

// PoolNote is the note beside a pool's total: "" when all of it counts,
// "counts a quarter past +200%: +405% earned" for a production pool past
// its knee, "capped at -90%: -120% earned" when a hard limit holds it.
func PoolNote(p BonusPool) string {
	if !p.Limited {
		return ""
	}
	if p.Target == "research_speed" {
		return "capped: a tech takes one tick at least, so only +100% counts"
	}
	if p.Soft {
		return kneeText(p.soft) + ": " + textfmt.SignedPercent(p.Earned) + " earned"
	}
	return "capped at " + textfmt.SignedPercent(p.Limit) + ": " + textfmt.SignedPercent(p.Earned) + " earned"
}

// PoolLine is a pool that applies less (or more) than it has earned, as one
// line for the rates breakdown: "All production: +405% earned, +251%
// counts. Past +200% a bonus counts a quarter." "" when all of it counts.
func PoolLine(p BonusPool) string {
	if !p.Limited {
		return ""
	}
	line := textfmt.Capitalize(EffectTargetName(p.Target)) + ": " + textfmt.SignedPercent(p.Earned) + " earned, " + pointsText(p.Applied) + " counts."
	if p.Soft {
		return line + " Past " + textfmt.SignedPercent(p.soft.Knee) + " a bonus counts " + shareText(p.soft.PastKnee) + "."
	}
	return line
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

// capLinesLocked is one log line for every effect a limit keeps from
// counting in full. Past a production pool's knee: "  → +20% all production
// counts a quarter past +200%: +5% now. You have earned +405% all
// production, and +251% of it counts." At a hard limit: "  → +20% research
// speed is capped: no effect now. ..." Empty when every effect counts.
// Caller holds the lock.
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
		phrase := CapPhrase(p, e.Value, held)
		if phrase == "" {
			continue
		}
		seen[target] = true
		earned := p.Earned
		if !held {
			earned += e.Value
		}
		name := EffectTargetName(target)
		line := "  → " + textfmt.SignedPercent(e.Value) + " " + name + " " + phrase + ". "
		applied, limit, _ := poolApplied(target, earned, ge.speedMultiplier, p.soft)
		switch {
		case target == "research_speed":
			line += "A tech takes one tick at least, so research speed counts up to +100%, and you have earned " + textfmt.SignedPercent(earned) + "."
		case poolPastKnee(target, earned, p.soft):
			line += "You have earned " + textfmt.SignedPercent(earned) + " " + name + ", and " + pointsText(applied) + " of it counts."
		default:
			line += textfmt.Capitalize(name) + " counts up to " + textfmt.SignedPercent(limit) + ", and you have earned " + textfmt.SignedPercent(earned) + "."
		}
		out = append(out, line)
	}
	return out
}

// grantLocked adds a one-off grant of amount of res to the store (an event,
// a milestone reward) and returns what the store took: the storage cap cuts
// off the rest. Caller holds the lock.
func (ge *GameEngine) grantLocked(res string, amount float64) (took float64) {
	before := ge.Resources.Get(res)
	return ge.Resources.Add(res, amount) - before
}

// clippedLine is the log line for grants a full store cut short: "  →
// Storage was nearly full: only 30 food (of 250) fit." promised and took are
// by resource; "" when everything fit.
func clippedLine(promised, took map[string]float64) string {
	var parts []string
	for _, res := range sortedKeys(promised) {
		if want := promised[res]; want > 0 && took[res] < want-float64(1e-9*want) {
			parts = append(parts, Amount(math.Max(took[res], 0), res)+" (of "+textfmt.Number(want)+")")
		}
	}
	if len(parts) == 0 {
		return ""
	}
	return "  → Storage was nearly full: only " + joinAnd(parts) + " fit."
}
