package game

import (
	"fmt"
	"math"
	"strings"

	"github.com/espresso20/ageforge/config"
)

// Faction-encounter engine (Phase 1 of the faction-system redesign).
//
// Expeditions are the primary way you MEET the world's civilizations. When a
// scouting expedition or military campaign resolves, we roll an "encounter" on
// the engine's seeded RNG (ge.rng). An encounter either:
//   - DISCOVERS an eligible-but-undiscovered faction (age is a floor; see
//     DiplomacyManager.DiscoverFaction), firing first contact, OR
//   - RE-ENCOUNTERS an already-discovered faction,
// and in both cases grants ONE timed boon — the proof buff. Phase 1 ships a
// single boon kind (a specialty-production multiplier); rollFactionBuff is the
// seam Phase 2 turns into a weighted table of many boon kinds.
//
// All rolls go through ge.rng, so a run's whole encounter/buff stream is
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

	// Proof-buff roll ranges. Magnitude is a fractional production multiplier on the
	// faction's specialty resource; duration is in ticks. Both rolled on ge.rng.
	factionBuffPctMin   = 0.08
	factionBuffPctMax   = 0.20
	factionBuffTicksMin = 3000
	factionBuffTicksMax = 6000
)

// factionEncounterBuff bundles the timed modifier to inject and its player-facing
// flavour. Structured (rather than injected inline) so Phase 2 can return other
// buff kinds from rollFactionBuff behind a weighted roll without touching the
// encounter-resolution flow.
type factionEncounterBuff struct {
	event   ActiveEvent // the timed modifier, injected into the event system
	logLine string      // flavour naming the faction + the boost
}

// rollExpeditionEncounter is called once per resolved expedition. It rolls the
// encounter chance for (category, success); on a hit it discovers an eligible
// faction (preferred) or re-encounters a discovered one, and applies the proof
// buff. Returns any log lines produced (first-contact + buff flavour); nil when
// no encounter fires or no faction is eligible yet. Must run under the write lock.
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

	if buff, ok := ge.rollFactionBuff(target); ok {
		ge.Events.InjectEvent(buff.event)
		messages = append(messages, buff.logLine)
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

// rollFactionBuff produces the proof buff for a faction encounter: "+X% of the
// faction's specialty resource for N ticks", both rolled on ge.rng. It returns
// (_, false) when the faction has no specialty to boost.
//
// The buff rides the EXISTING timed-modifier machinery: it is an ActiveEvent
// carrying a "<specialty>_rate" effect. EventManager.Modifiers() emits that as an
// OpAdd into the resolver's "<resource>_rate" additive pool, which recalculateRates
// turns into a ×(1 + Σ) multiplier on that resource's rate — the same pool research
// and allied-faction bonuses already use. Expiry is the event system's own
// TicksLeft countdown; no parallel bookkeeping.
//
// PHASE 2 SEAM: expand this into a weighted roll over several buff kinds (flat
// resource grants, tick-speed windows, temporary worker loans, ...), scaled by the
// faction's standing. Phase 1 always returns the specialty-production kind.
func (ge *GameEngine) rollFactionBuff(def config.FactionDef) (factionEncounterBuff, bool) {
	if def.Specialty == "" {
		return factionEncounterBuff{}, false
	}
	pct := factionBuffPctMin + ge.rng.Float64()*(factionBuffPctMax-factionBuffPctMin)
	ticks := factionBuffTicksMin + ge.rng.Intn(factionBuffTicksMax-factionBuffTicksMin+1)

	ev := ActiveEvent{
		Key:       factionBuffKey(def.Key),
		Name:      def.Name + " Boon",
		TicksLeft: ticks,
		Effects: []config.Effect{{
			Type:   def.Specialty + "_rate",
			Target: def.Specialty,
			Value:  pct,
		}},
	}
	line := fmt.Sprintf("[gold]✦ The %s share their craft[-] — [green]+%d%% %s[-] for %d ticks.",
		def.Name, int(math.Round(pct*100)), specialtyLabel(def.Specialty), ticks)

	return factionEncounterBuff{event: ev, logLine: line}, true
}

// factionBuffKey is the ActiveEvent key for a faction's encounter boon. Stable per
// faction so save/UI can identify it; re-encounters inject a fresh copy (boons
// stack and each expires on its own timer, matching how the engine's other
// InjectEvent side-effects behave).
func factionBuffKey(factionKey string) string {
	return "faction_boon_" + factionKey
}

// specialtyLabel prettifies a resource key for flavour text (quantum_flux →
// "quantum flux").
func specialtyLabel(resourceKey string) string {
	return strings.ReplaceAll(resourceKey, "_", " ")
}
