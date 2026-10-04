package game

import (
	"fmt"
	"math"
	"strings"

	"github.com/espresso20/ageforge/boon"
	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/pkg/textfmt"
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

// --- Malus (setback) tuning -------------------------------------------------
//
// Maluses invert the standing logic: a gift from a friend is bigger, but a
// setback from an ENEMY is worse. Strength still scales severity (a mightier
// civ's ill will bites harder) and the spread is deliberately narrower than the
// boon spread — a setback should sting, not decide the run.
const (
	// Strength → severity. Str1 ≈ ×0.90, Str5 ≈ ×1.10.
	malusStrengthBase    = 0.85
	malusStrengthPerUnit = 0.05

	// Standing → severity. Inverted relative to the boon multipliers above.
	malusAtWarSeverity    = 1.35
	malusHostileSeverity  = 1.15
	malusNeutralSeverity  = 1.00
	malusFriendlySeverity = 0.85
	malusAlliedSeverity   = 0.75
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
// routes at-war encounters to the MALUS table instead — see rollExpeditionEncounter
// and factionMalusProfile; this disable is belt-and-suspenders against a caller
// that forgets.)
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
	prof.MagnitudeScale *= strengthMagBase + float64(strengthMagPerUnit*float64(str))

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
//   - GrantTempWorkers   → WorkerManager.AddLentWorkers, plus a loan record
//     that sends them home when it runs out (returnBoonWorkers).
//
// name/key identify the granting faction so the injected event is stable per
// faction (same key as Phase 1: factionBuffKey) and legible in save/UI.
type boonApplier struct {
	ge   *GameEngine
	name string // faction display name, for the injected event's label
	key  string // faction key, for a stable ActiveEvent key
	// malus flips the injected ActiveEvent onto the factionMalusKey namespace.
	// Boons and setbacks must not share a key prefix: the capacity gate counts
	// live BOONS, and a setback taking up a boon slot would mean being punished
	// twice for the same encounter.
	malus bool
}

// eventKey is the ActiveEvent key this applier injects under — boon or malus
// namespace depending on polarity.
func (a boonApplier) eventKey() string {
	if a.malus {
		return factionMalusKey(a.key)
	}
	return factionBuffKey(a.key)
}

// InjectTimedEffects rides the existing timed-modifier machinery: an ActiveEvent
// carrying the boon's effects, expiring on its own TicksLeft countdown. Effect
// values may be negative — that is how a timed malus lands.
func (a boonApplier) InjectTimedEffects(effects []config.Effect, ticks int, name string) {
	a.ge.Events.InjectEvent(ActiveEvent{
		Key:       a.eventKey(),
		Name:      a.name + ": " + name,
		TicksLeft: ticks,
		Effects:   effects,
	})
}

// GrantResource immediately adds a lump grant to the stockpile.
func (a boonApplier) GrantResource(resource string, amount float64) {
	a.ge.Resources.Add(resource, amount)
}

// GrantTempWorkers lends workers into the pool and records the loan, so they
// go home when ticks run out (returnBoonWorkers).
func (a boonApplier) GrantTempWorkers(count, ticks int) {
	if count <= 0 {
		return
	}
	a.ge.Workers.AddLentWorkers(count)
	a.ge.Diplomacy.addBoonLoan(BoonWorkerLoan{FactionKey: a.key, Count: count, TicksLeft: ticks})
}

// BoonWorkerLoan is a crew of workers a faction boon lent (Extra Hands). Like
// the timed boons, it counts down in live ticks; offline catch-up does not
// run it down.
type BoonWorkerLoan struct {
	FactionKey string `json:"faction_key"`
	Count      int    `json:"count"`
	TicksLeft  int    `json:"ticks_left"`
}

// addBoonLoan records a boon crew on loan.
func (dm *DiplomacyManager) addBoonLoan(l BoonWorkerLoan) {
	dm.boonLoans = append(dm.boonLoans, l)
}

// tickBoonLoans counts every boon loan down one tick, drops the ones whose
// time is up and returns them, oldest first.
func (dm *DiplomacyManager) tickBoonLoans() []BoonWorkerLoan {
	var due []BoonWorkerLoan
	kept := dm.boonLoans[:0]
	for _, l := range dm.boonLoans {
		l.TicksLeft--
		if l.TicksLeft <= 0 {
			due = append(due, l)
			continue
		}
		kept = append(kept, l)
	}
	if len(kept) == 0 {
		kept = nil
	}
	dm.boonLoans = kept
	return due
}

// boonLoansForSave returns the boon crews on loan for serialization.
func (dm *DiplomacyManager) boonLoansForSave() []BoonWorkerLoan {
	return append([]BoonWorkerLoan(nil), dm.boonLoans...)
}

// loadBoonLoans restores the boon crews on loan from a save, replacing any
// held now. Their workers are already in the saved pool, so they are not
// added again.
func (dm *DiplomacyManager) loadBoonLoans(loans []BoonWorkerLoan) {
	dm.boonLoans = append([]BoonWorkerLoan(nil), loans...)
}

// returnBoonWorkers sends home the boon crews whose time is up. They leave
// through KillWorker, as the diplomacy loans do: idle workers go first, then
// the largest assignments give way. The line is routine unless the crew was
// staffing buildings. Runs every live tick, under the write lock.
func (ge *GameEngine) returnBoonWorkers() {
	for _, l := range ge.Diplomacy.tickBoonLoans() {
		idle := ge.Workers.IdleCount("")
		gone := ge.Workers.KillWorker(l.Count)
		if gone <= 0 {
			continue
		}
		line := fmt.Sprintf("%s from the %s went home.", textfmt.Count(gone, "worker", "workers"), CivName(l.FactionKey))
		switch staffed := gone - max(idle, 0); {
		case staffed == 1:
			ge.addLog("info", line+" One of them was staffing a building.")
		case staffed > 1:
			ge.addLog("info", fmt.Sprintf("%s %d of them were staffing buildings.", line, staffed))
		default:
			ge.addLog(LogRoutine, line)
		}
	}
}

// DrainResource removes a fraction of the resource's CURRENT stockpile, reusing
// the same removal path war raids take (ResourceManager.Remove, see the pending
// raids loop in doTick). The amount is derived from what is actually held and
// clamped to it, so the stockpile floors at 0 and can never go negative.
func (a boonApplier) DrainResource(resource string, fraction float64) {
	if resource == "" || fraction <= 0 {
		return
	}
	if fraction > 1 {
		fraction = 1
	}
	cur := a.ge.Resources.Get(resource)
	if cur <= 0 {
		return
	}
	lost := math.Min(cur*fraction, cur) // Min guards float rounding at fraction == 1
	a.ge.Resources.Remove(resource, lost)
}

// LoseWorkers removes workers via the existing starvation path. KillWorker
// already clamps the count to the live population and reconciles building
// assignments downward, so the pool cannot go negative and no building is left
// claiming workers that no longer exist.
func (a boonApplier) LoseWorkers(count int) {
	if count <= 0 {
		return
	}
	a.ge.Workers.KillWorker(count)
}

// applyFactionBoon rolls ONE boon for a faction encounter through the standalone
// boon engine and applies it, returning the faction-attributed, player-facing
// log line. It returns "" when nothing was granted — an empty specialty with no
// fallback, or an all-disabled (at-war) profile that yields the zero Boon.
//
// Uses the engine's seeded ge.rng so the whole encounter/boon stream is
// reproducible from the run's persisted seed. Runs under the write lock.
func (ge *GameEngine) applyFactionBoon(def config.FactionDef, state FactionState) string {
	return ge.applyRolledFactionBoon(def, ge.rollFactionBoon(def, state))
}

// rollFactionBoon rolls ONE boon for a faction encounter WITHOUT applying it, so
// the caller can inspect what came up first. That matters at boon capacity: the
// gate binds only on the kinds that occupy a slot, so the encounter path has to
// see the roll before deciding whether to grant it or bounce it (see
// rollExpeditionEncounter). Returns the zero Boon when nothing rolled.
func (ge *GameEngine) rollFactionBoon(def config.FactionDef, state FactionState) boon.Boon {
	prof := factionProfile(def, state, ge.age)
	prof.TickInterval = ge.tickIntervalLocked() // durations read at the current game speed
	return boon.RollBoon(prof, ge.rng)
}

// applyRolledFactionBoon applies an already-rolled boon and returns its
// faction-attributed, player-facing line ("" for the zero Boon).
func (ge *GameEngine) applyRolledFactionBoon(def config.FactionDef, b boon.Boon) string {
	if b == (boon.Boon{}) {
		return "" // no kind rolled (e.g. at-war): nothing to apply
	}
	// A timed buff joins a bonus pool: say so when a cap keeps it from
	// counting, before it is applied.
	note := ""
	switch b.Kind {
	case boon.RateBuff:
		note = ge.capNoteLocked(config.Effect{Type: b.Resource + "_rate", Target: b.Resource, Value: b.Magnitude}, false)
	case boon.AllProduction:
		note = ge.capNoteLocked(config.Effect{Type: "production_all", Value: b.Magnitude}, false)
	case boon.TickSpeed:
		note = ge.capNoteLocked(config.Effect{Type: "tick_speed", Value: b.Magnitude}, false)
	}
	stock := ge.Resources.Get(b.Resource)
	line := boon.Apply(b, boonApplier{ge: ge, name: def.Name, key: def.Key})
	if b.Kind == boon.InstantResource {
		// A lump a full store cut short says what fit.
		if took := ge.Resources.Get(b.Resource) - stock; took < b.InstantAmount-float64(1e-9*b.InstantAmount) {
			note = "storage was nearly full: only " + Amount(math.Max(took, 0), b.Resource) + " fit"
		}
	}
	if note != "" {
		line += " [yellow](" + note + ")[-]"
	}
	return fmt.Sprintf("[gold]✦ %s:[-] %s", def.Name, line)
}

// boonHoldsSlot reports whether a POSITIVE boon occupies one of the
// MaxConcurrentFactionBoons slots — i.e. whether applying it injects a timed
// ActiveEvent under factionBuffKeyPrefix. It must stay in lockstep with what
// boon.Apply injects for a positive polarity: the three timed kinds do, the
// instant lump and the work-gang do not (they are consumed on arrival and
// activeFactionBoonCount never sees them).
func boonHoldsSlot(b boon.Boon) bool {
	switch b.Kind {
	case boon.RateBuff, boon.AllProduction, boon.TickSpeed:
		return true
	default:
		return false
	}
}

// factionMalusProfile is factionProfile's negative twin: same seam (a faction's
// DATA biasing a shared table), opposite sign. Polarity Negative points RollBoon
// at MalusCatalog(); severity rises with the faction's strength and FALLS with
// your standing — a friend's bad news is gentler than an enemy's.
//
// Deliberately flat on weights: the malus table is small and every entry should
// stay reachable, so a faction's personality does not reshape it. What a faction
// IS shows up in the boons it gives; what it costs you shows up here as size.
func factionMalusProfile(def config.FactionDef, state FactionState, age string) boon.Profile {
	prof := boon.DefaultProfile()
	prof.Polarity = boon.Negative
	prof.Specialty = def.Specialty
	prof.Age = age

	str := def.Strength
	if str < 1 {
		str = 1
	} else if str > 5 {
		str = 5
	}
	prof.MagnitudeScale = malusStrengthBase + float64(malusStrengthPerUnit*float64(str))

	switch {
	case state.AtWar:
		prof.MagnitudeScale *= malusAtWarSeverity
	case state.Status == "rival", state.Status == "embargo":
		prof.MagnitudeScale *= malusHostileSeverity
	case state.Status == "allied":
		prof.MagnitudeScale *= malusAlliedSeverity
	case state.Status == "friendly":
		prof.MagnitudeScale *= malusFriendlySeverity
	default: // neutral or unknown
		prof.MagnitudeScale *= malusNeutralSeverity
	}

	return prof
}

// applyFactionMalus rolls ONE setback off the malus table and applies it,
// returning the faction-attributed, player-facing log line ("" if nothing
// rolled). Same engine, same seeded ge.rng, same Applier — only the table and
// the event namespace differ.
//
// Timed setbacks are capacity-bounded exactly like boons: at
// MaxConcurrentFactionMaluses the timed kinds are disabled for the roll and any
// production dip riding along on a ResourceDrain is stripped, so a long war
// cannot bury the active-events panel. The instant half (drain, worker loss)
// always lands — that is the part that should hurt.
//
// Runs under the write lock.
func (ge *GameEngine) applyFactionMalus(def config.FactionDef, state FactionState) string {
	prof := factionMalusProfile(def, state, ge.age)
	prof.TickInterval = ge.tickIntervalLocked()

	atCap := ge.activeFactionMalusCount() >= MaxConcurrentFactionMaluses
	if atCap {
		prof.Enabled = map[boon.Kind]bool{
			boon.RateBuff:      false,
			boon.AllProduction: false,
		}
	}

	b := boon.RollBoon(prof, ge.rng)
	if b == (boon.Boon{}) {
		return ""
	}
	if atCap {
		// Strip the optional dip a ResourceDrain may carry; the drain itself
		// stays, and its line drops the slowdown it no longer brings.
		b = b.WithoutDip()
	}

	line := boon.Apply(b, boonApplier{ge: ge, name: def.Name, key: def.Key, malus: true})
	return fmt.Sprintf("[red]✖ %s:[-] %s", def.Name, line)
}

// activeFactionBoonCount is the number of live faction BOON events — the value
// the capacity gate compares against MaxConcurrentFactionBoons. Only timed boon
// kinds inject an event, which is precisely what "holding a foreign favour"
// means; instant gifts are consumed on arrival and hold no slot.
func (ge *GameEngine) activeFactionBoonCount() int {
	return ge.countActiveWithPrefix(factionBuffKeyPrefix)
}

// activeFactionMalusCount is the same count for live faction SETBACKS.
func (ge *GameEngine) activeFactionMalusCount() int {
	return ge.countActiveWithPrefix(factionMalusKeyPrefix)
}

// countActiveWithPrefix counts active events whose Key carries the given prefix.
// Reads ge.Events.active directly (same package) — must run under the write lock,
// like every other encounter-path helper.
func (ge *GameEngine) countActiveWithPrefix(prefix string) int {
	n := 0
	for _, ae := range ge.Events.active {
		if strings.HasPrefix(ae.Key, prefix) {
			n++
		}
	}
	return n
}
