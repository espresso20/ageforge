package game

import (
	"fmt"

	"github.com/espresso20/ageforge/boon"
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
	// per RESOLVED expedition, so the cadence they produce depends entirely on how
	// long an expedition actually takes. The first tuning pass cut them ~4x from
	// 0.35/0.12/0.16/0.04 on the belief that the cheapest scouting run resolves in
	// 10 ticks and a continuous explorer therefore met someone every ~25 ticks.
	// That number came from a legacy fixed Duration field on ExpeditionDef, which
	// LaunchExpedition IGNORED whenever a def carried a [DurationMin, DurationMax]
	// range — the real shortest scouting run is 60-100 ticks. The field has since
	// been deleted; DurationMin/Max are the only source of truth. Corrected, the cut
	// odds produced one meeting per ~750 ticks (13 per 10k): a roster of foreign
	// civilizations you essentially never met.
	//
	// Half that cut is given back here. A continuous explorer now meets someone
	// roughly every ~375 ticks (~27 per 10k) — frequent enough to feel alive, and
	// paired with the restored catalogue durations (see boon/catalog.go) it puts
	// offered load at ~3.7 of the 5 boon slots, so most encounters land. Tuned,
	// not sacred: the two levers multiply, so move one and re-measure both.
	encounterChanceScoutSuccess    = 0.18
	encounterChanceScoutFail       = 0.06
	encounterChanceMilitarySuccess = 0.08
	encounterChanceMilitaryFail    = 0.02

	// --- Boon capacity ------------------------------------------------------
	// maxConcurrentFactionBoons is how many faction boons you may HOLD at once.
	// Before it existed, encounters injected without limit and a soak measured
	// 238 concurrent boons stacking a rate pool to x20 — the reward loop had no
	// shape, just accumulation. Capacity gives it one: boons are a scarce slot
	// you spend by holding, not a counter you grow. Only TIMED boons occupy a
	// slot; an instant gift is consumed on arrival.
	//
	// 3 -> 5 in the first measured tuning pass, and HELD at 5 through the second
	// (the one that fixed the harness's cadence bug). Capacity sets both how OFTEN
	// an encounter can pay out — throughput is slots/duration — and how BIG the
	// stacked buff feels at any moment. With the corrected cadence, offered load is
	// ~3.7 slot-equivalents against these 5 slots: ~3 boons live at a time, a
	// median combined uplift of x1.24-1.32 (inside the x1.2-1.6 band), and full
	// capacity occupied only 12-18% of the time.
	//
	// The reason to stop at 5 is the stacking arithmetic, not the throughput: the
	// worst realistic same-pool stack (5 allied str-5 Enlightenments at ~0.455
	// each = ~2.3) still clears productionCap's x3.0 clamp without touching it. Do
	// not raise this without re-checking that sum — past ~6 the clamp starts doing
	// the balancing instead of the data. Prefer moving durations or encounter odds.
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
	// nothing happened. Now it bites, but not every time: at a certain setback the
	// measured at-war scenario ran 52% setbacks, double the top of the 10-25% band
	// we want, because a war makes a large share of your encounters at-war
	// encounters. At 0.35 a war still reads as dangerous without turning the whole
	// exploration loop into punishment. A war encounter that does NOT bite yields
	// nothing at all — there is no gift from a civ you are fighting.
	malusChanceAtWar = 0.35
)

// atCapacityFlavors are the in-world lines for an encounter that arrives while
// your court already holds the maximum number of foreign favours. Player-facing
// production text: dry, in-world, no winking at the mechanic.
var atCapacityFlavors = []string{
	"Your stores are already thick with foreign gifts; the envoys are thanked, fed, and sent home with their crates unopened.",
	"The ledger of outstanding favours is full. There is nothing they can offer that you are not already owed.",
	"Your court can carry no more obligations this season — the party returns with courtesies and little else.",
}

// atWarNoHarmFlavors are the lines for meeting a civilization you are AT WAR with
// and getting away with it. There is never a gift from an enemy, but war does not
// bite on every contact either (see malusChanceAtWar), and the encounter must not
// be SILENT: before the malus table existed, a war encounter fired and produced
// nothing at all, which read to the player — and to the measurement harness — as
// if no encounter had happened. A standoff is an outcome and gets a line.
// Player-facing production text: dry, in-world, no winking at the mechanic.
var atWarNoHarmFlavors = []string{
	"Your scouts and theirs see each other across the valley. Both parties withdraw without a word.",
	"An enemy column shadows the expedition most of a day, then breaks off. Nothing is gained and nothing is lost.",
	"Contact with the enemy, brief and inconclusive. Your people come home with nothing but the sighting.",
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
	//   1. WAR — an at-war civ raids rather than gifts. It rolls the malus table at
	//      malusChanceAtWar; short of that the contact is a standoff, reported but
	//      empty. There is no positive boon from an enemy either way.
	//   2. FAILURE — a botched expedition that still met someone brings back
	//      trouble, not tribute.
	//   3. CAPACITY — you can only HOLD maxConcurrentFactionBoons timed favours at
	//      once. Over that a timed boon is refused: the party mostly returns
	//      empty-handed, and occasionally worse. An instant gift (a lump of goods,
	//      a work-gang) occupies no slot, so it still lands.
	//
	// Otherwise: a successful expedition, under capacity, gets a boon as before.
	state, _ := ge.Diplomacy.StateOf(target.Key)
	switch {
	case state.AtWar:
		if ge.rng.Float64() < malusChanceAtWar {
			if line := ge.applyFactionMalus(target, state); line != "" {
				messages = append(messages, line)
			}
		} else {
			// No harm done — but no gift either, and not silence.
			flavor := atWarNoHarmFlavors[ge.rng.Intn(len(atWarNoHarmFlavors))]
			messages = append(messages, fmt.Sprintf("[gray]✖ %s:[-] %s", target.Name, flavor))
		}
	case !success:
		if ge.rng.Float64() < malusChanceOnExpeditionFailure {
			if line := ge.applyFactionMalus(target, state); line != "" {
				messages = append(messages, line)
			}
		}
	case ge.activeFactionBoonCount() >= maxConcurrentFactionBoons:
		// Capacity is a limit on what you can HOLD, so it can only bind on the
		// boons that are held — the timed kinds, which are exactly the ones
		// activeFactionBoonCount counts. Roll the gift first: an instant lump or a
		// work-gang is consumed on arrival, takes no slot, and lands normally even
		// with a full court. Only a TIMED roll is refused, and only that refusal
		// can sour into a setback.
		b := ge.rollFactionBoon(target, state)
		switch {
		case b != (boon.Boon{}) && !boonHoldsSlot(b):
			if line := ge.applyRolledFactionBoon(target, b); line != "" {
				messages = append(messages, line)
			}
		case ge.rng.Float64() < atCapacityMalusChance:
			if line := ge.applyFactionMalus(target, state); line != "" {
				messages = append(messages, line)
			}
		default:
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
