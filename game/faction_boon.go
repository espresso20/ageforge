package game

import (
	"fmt"
	"math"

	"github.com/espresso20/ageforge/boon"
	"github.com/espresso20/ageforge/config"
)

// Faction-encounter → boon glue (Phase 2b of the faction redesign).
//
// Phase 1 fired a single hand-rolled specialty buff. This file makes faction
// encounters the FIRST consumer of the standalone boon engine (package boon):
// every encounter derives a boon.Profile from the faction's DATA (specialty +
// personality + strength) and your STANDING with them, rolls a varied boon off
// the shared weighted catalog, and applies it through a game-side Applier that
// wraps the exact machinery Phase 1 used (timed ActiveEvents, resource grants,
// worker loans).
//
// There are no per-faction tables here — a faction's character enters purely as
// biases on the shared catalog (which Kinds it favours, how big, how rare). All
// the tuning constants live grouped below so balancing is a data edit.

// --- Profile-derivation tuning (grouped so it is trivially tunable) ---------
const (
	// specialtyRateBuffWeight is the standing up-weight on RateBuff — an
	// encounter's signature gift is the faction sharing its craft, so the
	// specialty windfall is favoured for EVERY faction regardless of character.
	specialtyRateBuffWeight = 1.6

	// Personality weight leanings. Each is a multiplier on a catalog Kind's base
	// weight; we can only bias the existing kinds, never add entries.
	//
	// peaceful — gentle production/knowledge windfalls; no martial "spoils".
	peacefulRateBuffMult    = 1.25 // stacks on specialtyRateBuffWeight (folds in Enlightenment, a RateBuff)
	peacefulAllProdWeight   = 1.6
	peacefulTickSpeedWeight = 0.6
	// aggressive — war-tempo + spoils of contact; the gentle broad buff is rarer.
	aggressiveTickSpeedWeight = 1.8
	aggressiveInstantWeight   = 1.7
	aggressiveAllProdWeight   = 0.6
	// mercantile — lump-gold caravans + resource-surge windfalls.
	mercantileInstantWeight = 1.9
	mercantileRateBuffMult  = 1.2 // stacks on specialtyRateBuffWeight
	// isolationist — hoards/grand caches (rare InstantResource) + rarity lift so
	// the rare tier actually surfaces (also favours Enlightenment via RateBuff).
	isolationistInstantWeight = 1.6
	isolationistRarityScale   = 1.5

	// Strength → magnitude. Str1 ≈ ×0.98, Str5 ≈ ×1.30. Str is clamped to [1,5].
	strengthMagBase    = 0.90
	strengthMagPerUnit = 0.08

	// Standing → magnitude + rarity. Allied gifts are bigger AND reach the rare
	// tier; friendly are a touch bigger; hostile (rival/embargo) are meagre.
	alliedMagMult     = 1.40
	alliedRarityScale = 1.6
	friendlyMagMult   = 1.15
	neutralMagMult    = 1.00
	hostileMagMult    = 0.70
)

// factionProfile derives a boon.Profile from a faction's DATA and your standing
// with it — the seam that lets a faction's identity bias the shared boon catalog
// without the boon engine ever learning what a faction is.
//
//   - Specialty/Age target and bound the roll.
//   - Personality leans the per-Kind weights (favouring the kinds that fit the
//     faction's character, dampening the ones that clearly do not).
//   - Strength scales magnitude (a mightier civ gifts harder).
//   - Standing scales magnitude and, when allied, lifts RarityScale so the rare
//     tier unlocks.
//
// An at-war faction raids rather than gifts, so its profile DISABLES every kind:
// RollBoon then returns the zero Boon and no positive reward lands. (The caller
// also skips the roll for at-war factions; this is belt-and-suspenders.)
func factionProfile(def config.FactionDef, state FactionState, age string) boon.Profile {
	prof := boon.DefaultProfile()
	prof.Specialty = def.Specialty
	prof.Age = age

	// War: no positive boon. Disable the whole catalogue.
	if state.AtWar {
		prof.Enabled = map[boon.Kind]bool{
			boon.RateBuff:        false,
			boon.AllProduction:   false,
			boon.TickSpeed:       false,
			boon.InstantResource: false,
			boon.TempWorkers:     false,
		}
		return prof
	}

	// Specialty windfall is always favoured.
	prof.WeightMult = map[boon.Kind]float64{boon.RateBuff: specialtyRateBuffWeight}

	// Personality → weight leanings.
	switch def.Personality {
	case "peaceful":
		prof.WeightMult[boon.RateBuff] *= peacefulRateBuffMult
		prof.WeightMult[boon.AllProduction] = peacefulAllProdWeight
		prof.WeightMult[boon.TickSpeed] = peacefulTickSpeedWeight
	case "aggressive":
		prof.WeightMult[boon.TickSpeed] = aggressiveTickSpeedWeight
		prof.WeightMult[boon.InstantResource] = aggressiveInstantWeight
		prof.WeightMult[boon.AllProduction] = aggressiveAllProdWeight
	case "mercantile":
		prof.WeightMult[boon.InstantResource] = mercantileInstantWeight
		prof.WeightMult[boon.RateBuff] *= mercantileRateBuffMult
	case "isolationist":
		prof.WeightMult[boon.InstantResource] = isolationistInstantWeight
		prof.RarityScale = math.Max(prof.RarityScale, isolationistRarityScale)
	}

	// Strength → magnitude (clamped 1..5).
	str := def.Strength
	if str < 1 {
		str = 1
	} else if str > 5 {
		str = 5
	}
	prof.MagnitudeScale *= strengthMagBase + strengthMagPerUnit*float64(str)

	// Standing → magnitude (+ rarity when allied).
	switch state.Status {
	case "allied":
		prof.MagnitudeScale *= alliedMagMult
		prof.RarityScale = math.Max(prof.RarityScale, alliedRarityScale)
	case "friendly":
		prof.MagnitudeScale *= friendlyMagMult
	case "rival", "embargo":
		prof.MagnitudeScale *= hostileMagMult
	default: // neutral (or any unknown status)
		prof.MagnitudeScale *= neutralMagMult
	}

	return prof
}

// boonApplier adapts a GameEngine to boon.Applier — the narrow boundary through
// which the boon engine reaches the game. Each method maps to the exact
// machinery Phase 1's inline buff used:
//   - InjectTimedEffects → EventManager.InjectEvent(ActiveEvent{...}), the timed
//     "<res>_rate" / "production_all" / "tick_speed" modifier path.
//   - GrantResource      → ResourceManager.Add.
//   - GrantTempWorkers   → WorkerManager.AddLentWorkers.
//
// name/key identify the granting faction so the injected event is stable per
// faction (same key as Phase 1: factionBuffKey) and legible in save/UI.
type boonApplier struct {
	ge   *GameEngine
	name string // faction display name, for the injected event's label
	key  string // faction key, for a stable ActiveEvent key
}

// InjectTimedEffects rides the existing timed-modifier machinery: an ActiveEvent
// carrying the boon's effects, expiring on its own TicksLeft countdown.
func (a boonApplier) InjectTimedEffects(effects []config.Effect, ticks int, name string) {
	a.ge.Events.InjectEvent(ActiveEvent{
		Key:       factionBuffKey(a.key),
		Name:      a.name + ": " + name,
		TicksLeft: ticks,
		Effects:   effects,
	})
}

// GrantResource immediately adds a lump grant to the stockpile.
func (a boonApplier) GrantResource(resource string, amount float64) {
	a.ge.Resources.Add(resource, amount)
}

// GrantTempWorkers lends workers into the pool. The loan window (ticks) is not
// yet enforced by a return timer here — AddLentWorkers matches how allied civs
// already lend labour; wiring the timed return is a later concern.
func (a boonApplier) GrantTempWorkers(count, ticks int) {
	a.ge.Workers.AddLentWorkers(count)
}

// applyFactionBoon rolls ONE boon for a faction encounter through the standalone
// boon engine and applies it, returning the faction-attributed, player-facing
// log line. It returns "" when nothing was granted — an empty specialty with no
// fallback, or an all-disabled (at-war) profile that yields the zero Boon.
//
// Uses the engine's seeded ge.rng so the whole encounter/boon stream is
// reproducible from the run's persisted seed. Runs under the write lock.
func (ge *GameEngine) applyFactionBoon(def config.FactionDef, state FactionState) string {
	prof := factionProfile(def, state, ge.age)
	b := boon.RollBoon(prof, ge.rng)
	if b == (boon.Boon{}) {
		return "" // no kind rolled (e.g. at-war): nothing to apply
	}
	line := boon.Apply(b, boonApplier{ge: ge, name: def.Name, key: def.Key})
	return fmt.Sprintf("[gold]✦ %s:[-] %s", def.Name, line)
}
