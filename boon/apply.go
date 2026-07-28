package boon

import (
	"math"

	"github.com/espresso20/ageforge/config"
)

// Applier is the narrow boundary a consumer implements. It is the ONLY way this
// package reaches the outside world, and every method maps to machinery the
// game already has. Keeping it this small is what lets the roll engine be tested
// in full isolation with a fake.
//
// A game-side implementation is expected to wire these as:
//   - InjectTimedEffects → build an ActiveEvent{Name, TicksLeft: ticks,
//     Effects: effects} and hand it to EventManager.InjectEvent. The engine
//     already reads "<res>_rate", "production_all", and "tick_speed" effect
//     Types out of active events, so no new effect handling is needed.
//   - GrantResource     → Resources.Add(resource, amount).
//   - GrantTempWorkers  → WorkerManager.AddLentWorkers(count) (the loan-return /
//     expiry after ticks is the caller's concern; this package only states
//     intent).
//   - DrainResource     → the resource-removal path war raids already use
//     (ResourceManager.Remove), floored at zero.
//   - LoseWorkers       → WorkerManager.KillWorker(count), which already clamps
//     to the live population and reconciles assignments.
//
// The last two exist because a malus is not "GrantResource with a minus sign":
// a drain is a FRACTION of what you currently hold and must floor at zero, and
// dead workers are a head-count, not a resource. Naming them separately keeps
// both the intent and the flooring obligation explicit at the boundary.
type Applier interface {
	// InjectTimedEffects injects a set of timed multiplier effects that expire
	// together after ticks. name labels the injected event for logs/UI. Values
	// may be negative — that is exactly how a timed malus lands.
	InjectTimedEffects(effects []config.Effect, ticks int, name string)
	// GrantResource immediately adds amount of resource to the stockpile.
	GrantResource(resource string, amount float64)
	// GrantTempWorkers lends count workers for ticks (loan window enforced by
	// the caller).
	GrantTempWorkers(count, ticks int)
	// DrainResource removes `fraction` (0..1) of the resource's CURRENT
	// stockpile. Implementations must floor the result at zero and must never
	// produce a negative amount. A fraction of current — not a flat lump — so a
	// setback scales with what the player actually has instead of wiping a small
	// early-game store outright.
	DrainResource(resource string, fraction float64)
	// LoseWorkers removes count workers. Implementations must clamp to the live
	// population so the pool can never go negative.
	LoseWorkers(count int)
}

// Apply dispatches a rolled Boon to the right Applier call(s) and returns its
// player-facing flavor line. ALL boon→effect translation lives here, not in the
// caller: percentage buffs become the mapped config.Effect and ride
// InjectTimedEffects; instant grants and worker loans take their own methods.
//
// Effect-type mapping (keyed on Type, which is what the engine reads):
//   - RateBuff on "food"  → Effect{Type: "food_rate",     Target: "food",           Value: Magnitude}
//   - AllProduction       → Effect{Type: "production_all", Target: "production_all", Value: Magnitude}
//   - TickSpeed           → Effect{Type: "tick_speed",     Target: "tick_speed",     Value: Magnitude}
//
// A NEGATIVE-polarity RateBuff/AllProduction rides the exact same two lines with
// a negative Value — the engine's additive pools already handle debuffs, so a
// malus needs no new effect plumbing. Only the two malus-only Kinds do:
//   - ResourceDrain → DrainResource(Resource, InstantAmount as a fraction),
//     plus an optional production_all dip when Magnitude/DurationTicks are set.
//   - WorkerLoss    → LoseWorkers(round(InstantAmount)).
//
// StorageCap and TradeIncome are reserved and have no Applier path yet, so Apply
// intentionally no-ops them (still returning the flavor). Wire them once the
// engine grows the matching modifiers.
func Apply(b Boon, a Applier) string {
	switch b.Kind {
	case RateBuff:
		a.InjectTimedEffects([]config.Effect{{
			Type:   b.Resource + "_rate",
			Target: b.Resource,
			Value:  b.Magnitude,
		}}, b.DurationTicks, b.Name)
	case AllProduction:
		a.InjectTimedEffects([]config.Effect{{
			Type:   "production_all",
			Target: "production_all",
			Value:  b.Magnitude,
		}}, b.DurationTicks, b.Name)
	case TickSpeed:
		a.InjectTimedEffects([]config.Effect{{
			Type:   "tick_speed",
			Target: "tick_speed",
			Value:  b.Magnitude,
		}}, b.DurationTicks, b.Name)
	case InstantResource:
		a.GrantResource(b.Resource, b.InstantAmount)
	case TempWorkers:
		a.GrantTempWorkers(int(math.Round(b.InstantAmount)), b.DurationTicks)
	case ResourceDrain:
		// Drain first, then the OPTIONAL production dip. A Def with no
		// magnitude/duration (Spoiled Supplies) is a pure drain and makes a
		// single call; one that carries both (Lost Scouts) makes two.
		if b.Resource != "" && b.InstantAmount > 0 {
			a.DrainResource(b.Resource, b.InstantAmount)
		}
		if b.Magnitude != 0 && b.DurationTicks > 0 {
			a.InjectTimedEffects([]config.Effect{{
				Type:   "production_all",
				Target: "production_all",
				Value:  b.Magnitude,
			}}, b.DurationTicks, b.Name)
		}
	case WorkerLoss:
		if n := int(math.Round(b.InstantAmount)); n > 0 {
			a.LoseWorkers(n)
		}
	case StorageCap, TradeIncome:
		// Reserved: no engine plumbing yet. No-op until an Applier path exists.
	}
	return b.Flavor
}
