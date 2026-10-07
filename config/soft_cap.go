package config

// The soft cap on production bonus pools.
//
// Bonuses of one kind add together into a pool: "+30% all production" and
// "+20% all production" make +50%. A production pool (all production, and
// each resource's own production) applies in full up to a knee. Past the
// knee every further point still counts, at a fraction of its value. So a
// pool that has earned +320% applies +200% and a quarter of the other
// +120%: +230%.
//
// Until this rule the pools stopped dead at +200% (x3), and a player with
// every milestone and wonder had earned +661% by the last age: two thirds
// of those rewards did nothing. Under the soft cap each of them moves the
// rate.
//
// Research is not in these pools: a tech's bonus sits in a layer of its own,
// applied after them, and counts in full (tech_effects.go).

// SoftCap is the rule for a bonus pool with a knee: Knee is the most of the
// pool that applies in full, as a bonus (2 is +200%), and PastKnee the share
// of each point past it that still counts.
type SoftCap struct {
	Knee     float64
	PastKnee float64
}

// ProductionKnee and ProductionPastKnee are the one place the rule's two
// numbers live: a production pool applies in full up to +200%, and a quarter
// of everything past it. The game and the panels read them through the
// ruleset (rules.Set.SoftCap); the pacing model reads ProductionSoftCap.
const (
	ProductionKnee     = 2.0
	ProductionPastKnee = 0.25
)

// ProductionSoftCap is the soft cap every production pool follows.
func ProductionSoftCap() SoftCap {
	return SoftCap{Knee: ProductionKnee, PastKnee: ProductionPastKnee}
}

// Applied is what a pool that has earned bonuses in all applies: all of it
// up to the knee, and PastKnee of the rest. It is continuous at the knee and
// never decreases. A pool under the knee comes back untouched, bit for bit.
func (c SoftCap) Applied(earned float64) float64 {
	if earned <= c.Knee {
		return earned
	}
	return c.Knee + float64((earned-c.Knee)*c.PastKnee)
}

// Counts is how much a bonus of v adds to a pool that has earned bonuses in
// all without it: v under the knee, PastKnee of v past it, and part of each
// when v crosses the knee.
func (c SoftCap) Counts(earned, v float64) float64 {
	return c.Applied(earned+v) - c.Applied(earned)
}
