package game

import "sort"

// Op defines how a modifier combines with its peers on the SAME target.
//
// The split exists because not every bonus stacks the same way. Most game
// bonuses are additive percentage points that get pooled and applied once
// (e.g. +10% research here, +5% there → ×1.15 overall). A few are genuinely
// independent multipliers that must be folded in on their own (e.g. a morale
// factor of ×1.18, or a catastrophe penalty of ×0.5) and should not dilute
// into the additive pool.
type Op int

const (
	// OpAdd contributes additive percentage points. All OpAdd values on a
	// target are summed, then applied once as ×(1 + Σ).
	OpAdd Op = iota
	// OpMul contributes an independent multiplier, multiplied in directly.
	OpMul
)

// Modifier is one contribution to a target's total multiplier.
//
// Source is a stable id used purely for attribution in a Breakdown — it never
// affects the math. Examples: "research:masonry", "wonder:colossus",
// "prestige:passive", "morale", "event:peaceful_century",
// "milestone_chain:settlement".
//
// Target names the multiplier bucket this contributes to, e.g.
// "production_all", "food_rate", "tick_speed", "gather_rate".
type Modifier struct {
	Source string
	Target string
	Op     Op
	// Value is the contribution. For OpAdd, 0.10 means +10%. For OpMul,
	// 1.18 means ×1.18 (and 0.50 means ×0.50).
	Value float64
}

// Resolver aggregates modifiers by target and is the single source of truth for
// both Total and Breakdown. Consumers that need a number call Total; consumers
// that need to explain that number call Breakdown — both read the same stored
// contributions, so they can never disagree.
type Resolver struct {
	mods map[string][]Modifier
}

// NewResolver returns an empty Resolver ready to accept modifiers.
func NewResolver() *Resolver {
	return &Resolver{mods: make(map[string][]Modifier)}
}

// Add appends a single modifier under its target.
//
// Add is intentionally dumb: it stores everything, including no-ops (an OpAdd
// of 0.0 or an OpMul of 1.0). This keeps Breakdown faithful to exactly what was
// contributed. Total is unaffected because +0 and ×1 don't change the result;
// callers that want a tidy Breakdown can filter no-ops themselves.
func (r *Resolver) Add(m Modifier) {
	r.mods[m.Target] = append(r.mods[m.Target], m)
}

// AddAll appends each modifier in ms via Add.
func (r *Resolver) AddAll(ms []Modifier) {
	for _, m := range ms {
		r.Add(m)
	}
}

// Total returns the combined multiplier for target.
//
// Combination rule (fixed ordering): (1 + Σ OpAdd.Value) × Π OpMul.Value.
// The additive pool is summed and turned into a single factor FIRST, then every
// OpMul factor is multiplied onto it. An empty or unknown target returns 1.0.
func (r *Resolver) Total(target string) float64 {
	addSum := 0.0
	mulProd := 1.0
	for _, m := range r.mods[target] {
		switch m.Op {
		case OpMul:
			mulProd *= m.Value
		default: // OpAdd
			addSum += m.Value
		}
	}
	return (1 + addSum) * mulProd
}

// AddTotal returns the additive bonus pool for target: the sum of every OpAdd
// Value, ignoring OpMul factors entirely. This is the raw Σ that Total turns
// into a (1 + Σ) factor — engine code that historically hand-summed these
// additive contributions reads it here instead, so the pool has a single
// source of truth. An empty or unknown target returns 0.
func (r *Resolver) AddTotal(target string) float64 {
	sum := 0.0
	for _, m := range r.mods[target] {
		if m.Op == OpAdd {
			sum += m.Value
		}
	}
	return sum
}

// Breakdown returns the modifiers contributing to target, in insertion order.
//
// The returned slice is a copy, so mutating it (or its backing array) cannot
// corrupt the resolver's internal state. An empty or unknown target yields an
// empty slice.
func (r *Resolver) Breakdown(target string) []Modifier {
	src := r.mods[target]
	out := make([]Modifier, len(src))
	copy(out, src)
	return out
}

// Targets returns the sorted, de-duplicated list of targets that currently have
// at least one modifier. Handy for UI that wants to enumerate every multiplier
// bucket. (Map keys are inherently unique, so this is sorted-only; the "deduped"
// guarantee is structural.)
func (r *Resolver) Targets() []string {
	out := make([]string, 0, len(r.mods))
	for target := range r.mods {
		out = append(out, target)
	}
	sort.Strings(out)
	return out
}
