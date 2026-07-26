package game

import (
	"github.com/espresso20/ageforge/config"
)

// Faction-encounter engine (Phase 1 discovery + Phase 2b boon wiring).
//
// Expeditions are the primary way you MEET the world's civilizations. When a
// scouting expedition or military campaign resolves, we roll an "encounter" on
// the engine's seeded RNG (ge.rng). An encounter either:
//   - DISCOVERS an eligible-but-undiscovered faction (age is a floor; see
//     DiplomacyManager.DiscoverFaction), firing first contact, OR
//   - RE-ENCOUNTERS an already-discovered faction,
// and in both cases grants ONE boon. Phase 1 fired a single hand-rolled
// specialty buff; Phase 2b routes the reward through the standalone boon engine
// (see faction_boon.go / package boon) — a VARIED boon drawn off a shared
// weighted catalog, biased by the faction's character and your standing.
//
// All rolls go through ge.rng, so a run's whole encounter/boon stream is
// reproducible from its persisted seed. Every entry point runs under the engine
// write lock (called from doTick's processExpeditions), so touching ge.rng /
// ge.Events / ge.Diplomacy here is lock-safe — no GetState / lock-acquiring calls.

const (
	// Encounter probabilities per resolved expedition. Scouting is the eyes of the
	// empire, so it meets far more than a war party does; success beats failure but
	// even a botched run turns up SOMEONE now and then. Tuned, not sacred — Phase 2
	// may scale these by standing.
	encounterChanceScoutSuccess    = 0.35
	encounterChanceScoutFail       = 0.12
	encounterChanceMilitarySuccess = 0.15
	encounterChanceMilitaryFail    = 0.05
)

// rollExpeditionEncounter is called once per resolved expedition. It rolls the
// encounter chance for (category, success); on a hit it discovers an eligible
// faction (preferred) or re-encounters a discovered one, and rolls+applies one
// boon off the boon engine (see applyFactionBoon). Returns any log lines produced
// (first-contact + boon flavour); nil when no encounter fires or no faction is
// eligible yet. Must run under the write lock.
func (ge *GameEngine) rollExpeditionEncounter(category string, success bool) []string {
	// Self-heal a missing RNG service rather than ever panicking on a nil source.
	if ge.rng == nil {
		ge.SeedRNG(newSeed())
	}
	if ge.rng.Float64() >= encounterChance(category, success) {
		return nil
	}

	ageOrder := ge.progress.GetAgeOrder()
	undiscovered := ge.eligibleFactions(ageOrder, false)
	discovered := ge.eligibleFactions(ageOrder, true)

	var target config.FactionDef
	var messages []string
	switch {
	case len(undiscovered) > 0:
		// Prefer meeting someone new: a discovery encounter.
		target = ge.pickWeightedFaction(undiscovered)
		if msg, ok := ge.Diplomacy.DiscoverFaction(target.Key); ok {
			messages = append(messages, msg)
		}
	case len(discovered) > 0:
		// Everyone reachable is already known — re-encounter one of them.
		target = ge.pickWeightedFaction(discovered)
	default:
		// No faction is even eligible at this age yet: nobody out there to meet.
		return nil
	}

	// Roll a boon off the shared boon engine, shaped by the faction's character
	// and your standing with it. An at-war faction raids rather than gifts, so it
	// grants no positive boon — skip the roll entirely for it.
	state, _ := ge.Diplomacy.StateOf(target.Key)
	if !state.AtWar {
		if line := ge.applyFactionBoon(target, state); line != "" {
			messages = append(messages, line)
		}
	}
	return messages
}

// encounterChance returns the per-resolution encounter probability for a category
// and outcome. Unknown categories fall back to the military odds (the more
// conservative pair).
func encounterChance(category string, success bool) float64 {
	switch category {
	case ExpeditionScouting:
		if success {
			return encounterChanceScoutSuccess
		}
		return encounterChanceScoutFail
	default: // ExpeditionMilitary and any unknown category
		if success {
			return encounterChanceMilitarySuccess
		}
		return encounterChanceMilitaryFail
	}
}

// eligibleFactions returns the roster factions whose MinAge floor is met at the
// current age, filtered to either the discovered or the undiscovered set. Iterates
// config.BaseFactions() (a stable slice) so weighted picks over the result are
// deterministic given ge.rng's state.
func (ge *GameEngine) eligibleFactions(ageOrder map[string]int, discovered bool) []config.FactionDef {
	cur := ageOrder[ge.age]
	var out []config.FactionDef
	for _, def := range config.BaseFactions() {
		if cur < ageOrder[def.MinAge] {
			continue // age floor not reached
		}
		if ge.Diplomacy.IsDiscovered(def.Key) != discovered {
			continue
		}
		out = append(out, def)
	}
	return out
}

// pickWeightedFaction chooses one faction from a non-empty slice, weighted by
// Strength (a stronger, more storied civ is more likely to be the one you run
// into). Deterministic given ge.rng's state and the slice order. Strength is
// clamped to a floor of 1 so every faction has a non-zero weight.
func (ge *GameEngine) pickWeightedFaction(defs []config.FactionDef) config.FactionDef {
	total := 0
	for _, d := range defs {
		total += factionWeight(d)
	}
	if total <= 0 {
		return defs[ge.rng.Intn(len(defs))]
	}
	roll := ge.rng.Intn(total)
	cum := 0
	for _, d := range defs {
		cum += factionWeight(d)
		if roll < cum {
			return d
		}
	}
	return defs[len(defs)-1]
}

// factionWeight is a faction's encounter weight (Strength, floored at 1).
func factionWeight(d config.FactionDef) int {
	if d.Strength < 1 {
		return 1
	}
	return d.Strength
}

// factionBuffKey is the ActiveEvent key for a faction's encounter boon. Stable per
// faction so save/UI can identify it; re-encounters inject a fresh copy (boons
// stack and each expires on its own timer, matching how the engine's other
// InjectEvent side-effects behave). Used by boonApplier when the rolled boon is a
// timed-effect kind (RateBuff/AllProduction/TickSpeed).
func factionBuffKey(factionKey string) string {
	return "faction_boon_" + factionKey
}
