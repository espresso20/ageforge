package boon

import (
	"math"

	"github.com/espresso20/ageforge/config"
)

// Applier is what a consumer implements to carry out a boon. It is the only way
// this package reaches the outside world, which lets the roll engine be tested
// with a fake.
//
// The game implements it as boonApplier in game/faction_boon.go:
//   - InjectTimedEffects builds an ActiveEvent{Name, TicksLeft: ticks, Effects:
//     effects} and passes it to EventManager.InjectEvent. The engine already
//     reads "<res>_rate", "production_all" and "tick_speed" effect Types from
//     active events.
//   - GrantResource calls ResourceManager.Add.
//   - GrantTempWorkers calls WorkerManager.AddLentWorkers. Sending the workers
//     home after ticks is the caller's job; this package only states the intent.
//   - DrainResource uses the removal path war raids use (ResourceManager.Remove),
//     so the stock cannot go below zero.
//   - LoseWorkers calls WorkerManager.KillWorker, which clamps to the live
//     population and reconciles assignments.
//
// DrainResource and LoseWorkers are separate methods because a malus is not
// "GrantResource with a minus sign": a drain is a fraction of the current stock
// and must not go below zero, and dead workers are a head-count, not a resource.
type Applier interface {
	// InjectTimedEffects injects timed effects that expire together after ticks.
	// name labels the injected event in logs and the UI. Values may be negative,
	// which is how a timed malus lands.
	InjectTimedEffects(effects []config.Effect, ticks int, name string)
	// GrantResource immediately adds amount of resource to the stockpile.
	GrantResource(resource string, amount float64)
	// GrantTempWorkers lends count workers for ticks. The caller enforces the
	// loan length.
	GrantTempWorkers(count, ticks int)
	// DrainResource removes fraction (0 to 1) of the resource's current stock.
	// Implementations must not let the stock go below zero. A fraction rather
	// than a flat amount makes a setback scale with what the player holds, so it
	// cannot wipe out a small early-game store.
	DrainResource(resource string, fraction float64)
	// LoseWorkers removes count workers. Implementations must clamp to the live
	// population so the pool cannot go negative.
	LoseWorkers(count int)
}

// Apply dispatches a rolled Boon to the right Applier calls and returns its
// player-facing flavor line. All boon→effect translation lives here, not in the
// caller: percentage buffs become the mapped config.Effect and go through
// InjectTimedEffects. Instant grants and worker loans use their own methods.
//
// Effect-type mapping (keyed on Type, which is what the engine reads):
//   - RateBuff on "food"  → Effect{Type: "food_rate",     Target: "food",           Value: Magnitude}
//   - AllProduction       → Effect{Type: "production_all", Target: "production_all", Value: Magnitude}
//   - TickSpeed           → Effect{Type: "tick_speed",     Target: "tick_speed",     Value: Magnitude}
//
// A negative-polarity RateBuff or AllProduction takes the same path with a
// negative Value: the engine's additive pools already handle debuffs. Only the
// two malus-only Kinds need their own calls:
//   - ResourceDrain → DrainResource(Resource, InstantAmount as a fraction),
//     plus an optional production_all dip when Magnitude and DurationTicks are set.
//   - WorkerLoss    → LoseWorkers(round(InstantAmount)).
//
// StorageCap and TradeIncome have no Applier path, so Apply does nothing for them
// and still returns the flavor line.
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
		// Drain first, then the optional production dip. A Def with no magnitude
		// or duration (Spoiled Supplies) is a pure drain and makes one call. One
		// that carries both (Lost Scouts) makes two.
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
		// Reserved: nothing to call.
	}
	return b.Flavor
}
