package game

import (
	"fmt"

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
	// even a botched run turns up SOMEONE now and then.
	//
	// CALIBRATION (measured pass, see game/boon_tuning_test.go). These are rolled
	// per RESOLVED expedition, and the cheapest scouting run in the game resolves
	// in 10 ticks — so at the original 0.35/0.12 a player who simply re-launched
	// it met a foreign civilization every ~25 ticks (~400 per 10k). That is a
	// firehose against an 11-faction roster, and it drowned the boon inventory:
	// encounters arrived ~15x faster than boon slots could free, so 97% of them
	// bounced off a full court. Cut ~4x, a continuous explorer now meets someone
	// roughly every ~95 ticks (~105 per 10k) — frequent enough to feel alive,
	// slow enough that the reward actually lands. Tuned, not sacred.
	encounterChanceScoutSuccess    = 0.09
	encounterChanceScoutFail       = 0.03
	encounterChanceMilitarySuccess = 0.04
	encounterChanceMilitaryFail    = 0.01

	// --- Boon capacity ------------------------------------------------------
	// maxConcurrentFactionBoons is how many faction boons you may HOLD at once.
	// Before it existed, encounters injected without limit and a soak measured
	// 238 concurrent boons stacking a rate pool to x20 — the reward loop had no
	// shape, just accumulation. Capacity gives it one: boons are a scarce slot
	// you spend by holding, not a counter you grow. Only TIMED boons occupy a
	// slot; an instant gift is consumed on arrival.
	//
	// 3 -> 5 in the measured tuning pass. Capacity sets both how OFTEN an
	// encounter can pay out (throughput is slots/duration) and how BIG the stacked
	// buff feels at any moment. At 3 the live-boon count pinned at ~2.9 and the
	// combined uplift sat around x1.25 — under the x1.2-1.6 band we want while
	// boons are running. 5 lifts the working average to ~3.8 boons and the uplift
	// into the band, while the worst realistic same-pool stack (5 allied str-5
	// Enlightenments at ~0.455 each = ~2.3) still clears productionCap's x3.0
	// clamp without touching it. Do not raise this further without re-checking
	// that sum: past ~6 the clamp starts doing the balancing instead of the data.
	maxConcurrentFactionBoons = 5
	// maxConcurrentFactionMaluses bounds live setbacks the same way, so a long
	// war cannot bury the active-events panel. See applyFactionMalus: at this
	// cap the timed malus kinds are disabled and only instant harm lands.
	maxConcurrentFactionMaluses = 3

	// --- Malus triggers (all tunable; set a chance to 0 to switch one off) ----
	// atCapacityMalusChance is the odds that an encounter you have no room for
	// turns sour instead of merely returning empty-handed. Low by design: the
	// default at-capacity outcome is nothing, not punishment.
	atCapacityMalusChance = 0.25
	// malusChanceOnExpeditionFailure is the odds a FAILED expedition's encounter
	// yields a setback rather than a gift. A botched run that still stumbles into
	// a foreign civ brings back trouble, not tribute.
	malusChanceOnExpeditionFailure = 1.0
	// malusChanceAtWar is the odds an encounter with a civ you are AT WAR with
	// yields a setback. War used to be a silent skip — the encounter fired and
	// nothing happened. Now it bites.
	malusChanceAtWar = 1.0
)

// atCapacityFlavors are the in-world lines for an encounter that arrives while
// your court already holds the maximum number of foreign favours. Player-facing
// production text: dry, in-world, no winking at the mechanic.
var atCapacityFlavors = []string{
	"Your stores are already thick with foreign gifts; the envoys are thanked, fed, and sent home with their crates unopened.",
	"The ledger of outstanding favours is full. There is nothing they can offer that you are not already owed.",
	"Your court can carry no more obligations this season — the party returns with courtesies and little else.",
}

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

	// Resolve the outcome. Three things can turn an encounter sour or empty, in
	// priority order:
	//
	//   1. WAR — an at-war civ raids rather than gifts. This used to be a silent
	//      skip; it now rolls the malus table (the deferred "negative table").
	//   2. FAILURE — a botched expedition that still met someone brings back
	//      trouble, not tribute.
	//   3. CAPACITY — you can only hold maxConcurrentFactionBoons favours at
	//      once. Over that, no positive boon lands; the party mostly returns
	//      empty-handed, and occasionally worse.
	//
	// Otherwise: a successful expedition, under capacity, gets a boon as before.
	state, _ := ge.Diplomacy.StateOf(target.Key)
	switch {
	case state.AtWar:
		if ge.rng.Float64() < malusChanceAtWar {
			if line := ge.applyFactionMalus(target, state); line != "" {
				messages = append(messages, line)
			}
		}
	case !success:
		if ge.rng.Float64() < malusChanceOnExpeditionFailure {
			if line := ge.applyFactionMalus(target, state); line != "" {
				messages = append(messages, line)
			}
		}
	case ge.activeFactionBoonCount() >= maxConcurrentFactionBoons:
		if ge.rng.Float64() < atCapacityMalusChance {
			if line := ge.applyFactionMalus(target, state); line != "" {
				messages = append(messages, line)
			}
		} else {
			flavor := atCapacityFlavors[ge.rng.Intn(len(atCapacityFlavors))]
			messages = append(messages, fmt.Sprintf("[gray]✦ %s:[-] %s", target.Name, flavor))
		}
	default:
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

// Key namespaces for faction-encounter outcomes. Boons and setbacks are kept
// apart on purpose: activeFactionBoonCount / activeFactionMalusCount count by
// prefix, so a setback must never be mistaken for a held favour.
const (
	factionBuffKeyPrefix  = "faction_boon_"
	factionMalusKeyPrefix = "faction_malus_"
)

// factionBuffKey is the ActiveEvent key for a faction's encounter boon. Stable per
// faction so save/UI can identify it; re-encounters inject a fresh copy (InjectEvent
// appends rather than replaces, so boons stack and each expires on its own timer).
// What bounds the stack is no longer expiry alone but maxConcurrentFactionBoons,
// enforced in rollExpeditionEncounter. Used by boonApplier when the rolled boon is a
// timed-effect kind (RateBuff/AllProduction/TickSpeed).
func factionBuffKey(factionKey string) string {
	return factionBuffKeyPrefix + factionKey
}

// factionMalusKey is the ActiveEvent key for a faction's encounter SETBACK — the
// negative namespace, bounded by maxConcurrentFactionMaluses.
func factionMalusKey(factionKey string) string {
	return factionMalusKeyPrefix + factionKey
}
