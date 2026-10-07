package game

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strconv"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/rules"
)

// Civilizational catastrophes (Phase 9).
//
// A catastrophe strikes when a doom fated for its era comes (fate.go): at a
// random tick anywhere in the era, or at an advance the player reached first.
// Players cannot trigger one directly; the dev console's /catastrophe does, for
// testing. Either way it only sets pendingCatastrophe: nothing is destroyed
// until the player chooses Endure or Succumb. While a catastrophe is pending
// the game keeps running but AdvanceAge and DoPrestige refuse, so the choice
// cannot be skipped or overwritten.
//
// Rules in one place:
//   - No catastrophe before config.CatastropheGateEpoch (the Iron Era).
//   - At most one doom per era per run: each era's fate rolls once, when it
//     is entered, and resolves once.
//   - All randomness comes from the seeded ge.rng over stable pools.
//
// Every method here that is not exported expects the engine write lock to be
// held, except the read-only helpers used by GetState (catastropheOutlook,
// succumbResearchFactor, countCatastropheOutcomes), which only
// read state and are safe under the read lock.

const (
	// Epoch transition good-event chance by faith strength (faith.go).
	epochGoodChanceLowFaith  = 0.40 // strength < 25%
	epochGoodChanceBase      = 0.50 // 25%–75%
	epochGoodChanceHighFaith = 0.60 // strength > 75%

	// catastropheChanceOnBadRoll is the old chance that a bad epoch roll
	// escalated into a catastrophe. Transitions no longer bring catastrophes,
	// but the passage chance it made, (1-goodChance) × this (18% at low faith,
	// 15% mid, 12% high), is still the Last Passage's odds at prestige, and
	// FateStrikeScale times it is a fated doom's strike chance.
	catastropheChanceOnBadRoll = 0.30

	// Endure consequences.
	// Endure destroys 20% of destroyable buildings (floor) unbraced: every
	// building but wonders and storage (isDestroyable). See braceDestroyPct in
	// harbinger.go for the braced shares.
	endureResourceKeep     = 0.15 // resources drop to 15% of stored amounts
	endureWorkerLoss       = 0.25 // 25% of the worker pool is lost
	endureDebuffTicks      = 216  // Reconstruction Effort duration on the base curve (EndureDebuffTicksIn)
	endureDebuffProduction = -0.10
	endureMoraleHit        = -0.10

	// Succumb consequences.

	// SuccumbRuinCount is how many buildings Succumb turns into ruins.
	SuccumbRuinCount = 8
	// SuccumbResearchTimeFactor is Ancient Knowledge: what research time is
	// multiplied by for each distinct epoch succumbed in (legacy flag). It
	// multiplies: x0.8, x0.64, ... x0.26 with all six epochs, so research
	// never floors. (It used to be +25% research speed per epoch, taken off
	// the listed time, which brought every tech to one tick at four epochs.)
	SuccumbResearchTimeFactor = 0.8
)

// Endure consequences, exported for the catastrophe modal so its text is built
// from the numbers Endure applies rather than copies of them.
const (
	EndureWorkerLoss       = endureWorkerLoss
	EndureDebuffProduction = endureDebuffProduction
	EndureMoraleHit        = endureMoraleHit
)

// EndureDebuffTicksIn is how long Endure's Reconstruction Effort lasts in age
// on the core ruleset: endureDebuffTicks stretched like the age, so the
// debuff covers the same share of a longer age. Dooms strike from the Iron Era
// on, so in play it is always the stretched length.
func EndureDebuffTicksIn(age string) int {
	return rules.Core().StretchTicks(age, endureDebuffTicks)
}

// Catastrophe record outcomes (EpochEventRecord.Outcome).
const (
	CatastrophePending   = "pending"
	CatastropheEndured   = "endured"
	CatastropheSuccumbed = "succumbed"
)

// Civilization-log markers. countCatastropheOutcomes matches on these, so the
// history entries written by Endure/Succumb must keep them.
const (
	catHistEnduredMarker   = " — Endured "
	catHistSuccumbedMarker = " — Succumbed to "
)

// CatastropheTier is a coarse bucket for the catastrophe probability, meant for
// flavour text (a harbinger's warning) rather than exact odds.
type CatastropheTier string

const (
	CatastropheTierNone   CatastropheTier = "none"
	CatastropheTierLow    CatastropheTier = "low"
	CatastropheTierMedium CatastropheTier = "medium"
	CatastropheTierHigh   CatastropheTier = "high"
)

// Tier thresholds on Probability. With the current odds the three faith bands
// map one-to-one: high faith 12% → low, mid 15% → medium, low faith 18% → high.
const (
	catastropheTierMediumAt = 0.14
	catastropheTierHighAt   = 0.17
)

// CatastropheOutlook is the catastrophe outlook as the player can know it. It
// is built from what the player has seen, never from the hidden fate: in an
// era whose doom has not been foretold it reads the same whether or not one
// is fated (fate.go).
//
// In an era: the doom foretold by the harbinger present, if one is; otherwise
// quiet (safe, for now) while the era can still bring one. In the final
// epoch, whose passage is prestige: the Last Passage's odds (last_passage.go).
type CatastropheOutlook struct {
	// Passage is PassageEpoch, or PassagePrestige in the final epoch.
	Passage string
	// NextEpochKey is the epoch the next transition enters; "" in the final
	// epoch, whose passage is prestige. For the engine's own readers: the UI
	// must not name it.
	NextEpochKey string
	// Possible: a catastrophe could still come. In an era: it is past the
	// Iron gate and its doom has neither struck nor been lifted (or a
	// harbinger is warning of one). In the final epoch: prestige can bring
	// the Last Passage and it is not already pending.
	Possible bool
	// Warned: a harbinger is present warning of this era's fated doom (in
	// the final epoch, its Reality Tear rather than the Last Passage).
	Warned bool
	// Probability is the chance in [0,1] as it can be known: while Warned,
	// what the harbinger's warning says (the claim, for a false prophet); in a
	// quiet era 0; in the final epoch always the Last Passage's chance at
	// prestige (1 when invited), whatever its fated doom says (that is on
	// GameState.Harbinger).
	Probability float64
	// Tier buckets Probability: none / low / medium / high.
	Tier CatastropheTier
	// FaithStrength is the current faith strength in [0,1] that drives the
	// odds, and FaithBand the band it falls in, the one every roll reads
	// (faith.go).
	FaithStrength float64
	FaithBand     FaithBand
	// FaithDevotion and FaithKept are the two things it comes from, for the
	// views that explain it: what the town's faith buildings have made this
	// run against a moderate set (1 is the moderate town), and the share of
	// the run's faith income still held (at most 1).
	FaithDevotion float64
	FaithKept     float64
}

func catastropheTierFor(p float64) CatastropheTier {
	switch {
	case p <= 0:
		return CatastropheTierNone
	case p < catastropheTierMediumAt:
		return CatastropheTierLow
	case p < catastropheTierHighAt:
		return CatastropheTierMedium
	default:
		return CatastropheTierHigh
	}
}

// gameRNG returns the seeded run RNG, seeding it on first use for engines that
// were not built by NewGameEngine. Call only under the write lock.
func (ge *GameEngine) gameRNG() *rand.Rand {
	if ge.rng == nil {
		ge.SeedRNG(newSeed())
	}
	return ge.rng
}

// epochGoodChance returns the chance that an epoch transition rolls a good
// event, by the town's faith strength. Read-only.
func (ge *GameEngine) epochGoodChance() float64 {
	return goodChanceFor(ge.faithStrength())
}

// EpochGoodChanceIn is the chance an epoch transition rolls a good event in
// faith band band: what the faith row prints beside the strength. Pure.
func EpochGoodChanceIn(band FaithBand) float64 {
	switch band {
	case FaithBandLow:
		return epochGoodChanceLowFaith
	case FaithBandHigh:
		return epochGoodChanceHighFaith
	}
	return epochGoodChanceBase
}

// goodChanceFor is the good-event chance at faith strength strength: the
// faith bands every catastrophe chance reads. Pure.
func goodChanceFor(strength float64) float64 {
	return EpochGoodChanceIn(FaithBandAt(strength))
}

// catastropheBlockErr is the error AdvanceAge and DoPrestige return while a
// catastrophe is pending. action is a gerund ("advancing", "prestiging").
func (ge *GameEngine) catastropheBlockErr(action string) error {
	name, _ := ge.rules.Catastrophe(ge.pendingCatastrophe)
	return fmt.Errorf("%s is upon you. Type 'catastrophe' to choose Endure or Succumb before %s", name, action)
}

// How a catastrophe came about; it only changes the log line and event name.
const (
	catastropheRolled  = ""        // a fated doom's strike roll hit
	catastropheForced  = "forced"  // dev console /catastrophe
	catastropheInvited = "invited" // an invited doom (the Harbinger's Invite)
)

// triggerCatastrophe makes epochKey's catastrophe pending: log line, history
// record, and a bus event so the dashboard toast fires. Callers make sure no
// catastrophe is pending (one is never overwritten) and the era is past the
// Iron gate. Must be called under the write lock; the bus
// handlers it reaches must not take the engine lock (see CLAUDE.md).
func (ge *GameEngine) triggerCatastrophe(epochKey, source string) {
	ep, _ := ge.rules.Era(epochKey)
	catName, _ := ge.rules.Catastrophe(epochKey)
	ge.pendingCatastrophe = epochKey

	eventName := catName
	if source != catastropheRolled {
		eventName = catName + " (" + source + ")"
	}
	ge.addLog("warning", fmt.Sprintf("☄ %s threatens the %s. Prepare yourself.", catName, ep.Name))
	ge.addLog("warning", "  Type 'catastrophe' to choose Endure or Succumb. Advancing and prestige wait until you do.")

	ge.epochEventHistory = append(ge.epochEventHistory, EpochEventRecord{
		EpochKey: epochKey, EpochName: ep.Name,
		EventKey: ep.CatastropheKey, EventName: eventName, EventType: "catastrophe",
		Tick: ge.tick, Outcome: CatastrophePending,
	})
	ge.Bus.Publish(EventData{
		Type: EventEpochEventFired,
		Payload: map[string]interface{}{
			"event_key":  ep.CatastropheKey,
			"event_name": catName,
			"event_type": "catastrophe",
			"epoch_key":  epochKey,
			"source":     source,
		},
	})
}

// setCatastropheOutcome stamps the most recent pending catastrophe record for
// epochKey with outcome.
func (ge *GameEngine) setCatastropheOutcome(epochKey, outcome string) {
	for i := len(ge.epochEventHistory) - 1; i >= 0; i-- {
		r := &ge.epochEventHistory[i]
		if r.EpochKey == epochKey && r.EventType == "catastrophe" && r.Outcome == CatastrophePending {
			r.Outcome = outcome
			return
		}
	}
}

// forceCatastrophe makes the current epoch's catastrophe pending right now.
// It is a testing tool behind the dev console (/catastrophe), not a player
// action. It still respects the Iron-epoch gate and never overwrites a pending
// catastrophe, but it ignores the era's fate (a fated doom still comes later,
// once this one is answered).
func (ge *GameEngine) forceCatastrophe() error {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	if ge.pendingCatastrophe != "" || ge.pendingLastPassage {
		return fmt.Errorf("a catastrophe is already pending")
	}
	if !ge.rules.CatastropheAllowed(ge.currentEpoch) {
		return fmt.Errorf("catastrophes cannot strike before the %s", eraName(ge.rules, config.CatastropheGateEpoch))
	}
	ge.triggerCatastrophe(ge.currentEpoch, catastropheForced)
	return nil
}

// Invite of the Last Passage (armed by the Cosmic Era thread's Invite; see
// harbinger.go). An era's doom keeps its invite on its fate (FateSave.Invited).
//
// catastropheInvited, when set, makes the next prestige from the final epoch
// bring the Last Passage instead of rolling (rollLastPassage). It is consumed
// there, persisted, and cleared by Succumb and prestige.

// inviteCatastrophe arms the Last Passage's invite. Must be called under the
// write lock.
func (ge *GameEngine) inviteCatastrophe() { ge.catastropheInvited = true }

// CatastropheOutlook reports the catastrophe outlook as the player can know
// it (see the type). Takes the read lock; use catastropheOutlook from code
// that already holds a lock.
func (ge *GameEngine) CatastropheOutlook() CatastropheOutlook {
	ge.mu.RLock()
	defer ge.mu.RUnlock()
	return ge.catastropheOutlook()
}

// catastropheOutlook is the lock-free body of CatastropheOutlook. Read-only:
// safe under either lock, and from GetState. It must never read whether a
// doom is fated (the fate's Fated flag or strike tick): only the harbinger
// present and what has already resolved in the open.
func (ge *GameEngine) catastropheOutlook() CatastropheOutlook {
	held, strength := ge.Resources.Get("faith"), ge.faithStrength()
	out := CatastropheOutlook{Passage: PassageEpoch, Tier: CatastropheTierNone, FaithStrength: strength, FaithBand: FaithBandAt(strength),
		FaithDevotion: FaithDevotionOf(ge.faithMeasure), FaithKept: FaithKeptOf(held, ge.faithMeasure)}
	if next, ok := ge.rules.NextEra(ge.currentEpoch); ok {
		out.NextEpochKey = next.Key
	}
	if ge.rules.IsFinalEra(ge.currentEpoch) {
		// The final epoch's passage is prestige: the Last Passage, at its
		// own thread's Appease. Its fated doom shows only once foretold.
		out.Passage = PassagePrestige
		out.Warned = ge.fateThread() != nil
		if !ge.lastPassageApplies() || ge.pendingLastPassage {
			return out
		}
		out.Possible = true
		out.Probability = (1 - ge.epochGoodChance()) * catastropheChanceOnBadRoll * appeaseMultiplierOf(ge.lastPassageThread())
		if ge.catastropheInvited {
			out.Probability = 1
		}
		out.Tier = catastropheTierFor(out.Probability)
		return out
	}
	if h := ge.harbinger; h != nil && h.TargetEpoch != "" && h.EpochKey == ge.currentEpoch {
		out.Possible, out.Warned = true, true
		out.Tier, out.Probability = ge.harbingerDisplay()
		return out
	}
	out.Possible = ge.rules.FateAllowed(ge.currentEpoch) && ge.pendingCatastrophe == "" && !ge.fateSettled()
	return out
}

// releaseWorkersFrom sends workers assigned to destroyed buildings back to the
// idle pool, the same way SellBuilding does: each building keeps at most
// WorkerCapacity × remaining count assigned. Keys are processed in sorted order.
func (ge *GameEngine) releaseWorkersFrom(destroyed map[string]int) {
	keys := make([]string, 0, len(destroyed))
	for k := range destroyed {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		def := ge.Buildings.defs[key]
		capacity := def.WorkerCapacity * ge.Buildings.GetCount(key)
		assigned := ge.Workers.GetAssignedCount("worker", key)
		if assigned > capacity {
			ge.Workers.Unassign("worker", key, assigned-capacity)
		}
	}
}

// Endure executes the Endure consequences for the pending catastrophe:
//   - 20% of destroyable buildings (neither wonders nor storage) destroyed,
//     at least 1 if any;
//     workers assigned to them return to the idle pool (15% / 10% when the
//     harbinger was braced at level 1 / 2)
//   - all unlocked resources drop to 15% of their stored amounts (30% / 45%
//     braced)
//   - 25% of the worker pool lost, spread proportionally over every building
//   - Reconstruction Effort: production -10% for 216 ticks (stretched for the age)
//   - morale -0.10
//   - Survived marker for the epoch and a civilization-log entry
//
// Research, wonders, age and prestige are untouched.
func (ge *GameEngine) Endure() error {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	if ge.pendingCatastrophe == "" && ge.pendingLastPassage {
		return ge.resolveLastPassage(lastPassageEndured)
	}
	if ge.pendingCatastrophe == "" {
		return fmt.Errorf("no pending catastrophe to endure")
	}
	epochKey := ge.pendingCatastrophe
	ge.pendingCatastrophe = ""
	// The stock the next advance counted is mostly gone: the next tick (or
	// the advance command) checks the requirements afresh.
	ge.ageReady = false
	// A Brace bought from the harbinger softens the blow.
	brace := ge.pendingBraceLevel
	if brace < 0 || brace > HarbingerMaxBrace {
		brace = 0
	}
	ge.pendingBraceLevel = 0
	// Then the garrison, measured before the blow lands (soldiers are stock
	// too, and fall with the rest).
	outcome := ge.endurePreview(brace, ge.age)
	keep := outcome.KeepFrac
	ge.survivedEpochs[epochKey] = true
	ge.setCatastropheOutcome(epochKey, CatastropheEndured)

	catName, catFlavor := ge.rules.Catastrophe(epochKey)
	epName := eraName(ge.rules, epochKey)

	destroyCount := outcome.DestroyCount
	destroyed, names := ge.Buildings.DestroyRandom(ge.gameRNG(), destroyCount)
	ge.releaseWorkersFrom(destroyed)

	// Stock the garrison kept: the difference between the braced keep and the
	// garrison's keep, per resource (sorted, so the tally sums in one order).
	var keptStock map[string]float64
	if keep != outcome.BracedKeepFrac {
		keptStock = make(map[string]float64)
		for _, key := range sortedKeys(ge.Resources.resources) {
			if r := ge.Resources.resources[key]; r != nil && ge.Resources.IsUnlocked(key) {
				keptStock[key] = float64(r.Amount*keep) - float64(r.Amount*outcome.BracedKeepFrac)
			}
		}
	}
	for key, r := range ge.Resources.resources {
		if r != nil && ge.Resources.IsUnlocked(key) {
			r.Amount *= keep
		}
	}

	ge.Workers.RemovePct(endureWorkerLoss)

	debuff := ge.startReconstruction()

	ge.addLog("warning", fmt.Sprintf("☄ ENDURE: %s. %s", catName, catFlavor))
	ge.addLog("warning", fmt.Sprintf("  Buildings destroyed: %d", destroyCount))
	for _, desc := range names {
		ge.addLog("warning", fmt.Sprintf("  → %s lost", desc))
	}
	if brace > 0 {
		ge.addLog("info", fmt.Sprintf("  Braced (level %d): %d%% of buildings lost instead of %d%%, %.0f%% of stock kept instead of %.0f%%.",
			brace, braceDestroyPct[brace], braceDestroyPct[0], outcome.BracedKeepFrac*100, braceKeepFrac[0]*100))
	}
	if line := endureGarrisonLine(outcome); line != "" {
		ge.addLog("success", line)
		t := ge.defenseTally()
		t.Buildings += outcome.BuildingsSaved
		for _, key := range sortedKeys(keptStock) {
			ge.recordSavedResource(key, keptStock[key])
		}
	}
	ge.addLog("warning", fmt.Sprintf("  All resources reduced to %.0f%% of stored amounts.", keep*100))
	ge.addLog("warning", fmt.Sprintf("  %.0f%% of workers lost.", endureWorkerLoss*100))
	ge.addLog("info", fmt.Sprintf("  Reconstruction: all production %.0f%% for %s. Morale %+.0f points.",
		endureDebuffProduction*100, approxTicks(debuff, ge.tickIntervalLocked()), endureMoraleHit*100))
	// Past the all-production cap the surplus absorbs the penalty: say so.
	ge.logCapped(config.Effect{Type: "production_all", Value: endureDebuffProduction})
	ge.addLog("success", fmt.Sprintf("  ✦ You endured the %s. Its badge records it.", epName))
	// Cosmetic flavour — a wry beat after surviving the catastrophe.
	if q := config.PickLogFlavor(config.LogFlavorCatastropheSurvived, ge.quipRNG()); q != "" {
		ge.addLog("info", fmt.Sprintf("  [gray]%s[-]", q))
	}

	ge.catastropheHistory = append(ge.catastropheHistory,
		fmt.Sprintf("Tick %d"+catHistEnduredMarker+"%s (%s). %d buildings lost.", ge.tick, catName, epName, destroyCount))

	ge.applyMorale(endureMoraleHit)
	return nil
}

// Succumb executes the Succumb consequences for the pending catastrophe:
//   - up to 8 random destroyable buildings (neither wonders nor storage)
//     become ruins (50% output, no workers),
//     carried into the next run; the ruin total is capped at MaxRuins
//   - the epoch's legacy flag is set: its per-resource legacy bonus and
//     Ancient Knowledge (research time x0.8 per distinct epoch succumbed in)
//     are permanent
//   - Era Mastery: every age the run completed gains a level, as at a
//     prestige (CommitRun), so the rebuild runs on known ground
//   - full reset to the Primitive Age: buildings, resources, workers, research,
//     milestones, events. Prestige level, points and upgrades are kept; no
//     prestige points are earned
//   - civilization-log entry
func (ge *GameEngine) Succumb() error {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	if ge.pendingCatastrophe == "" && ge.pendingLastPassage {
		return ge.resolveLastPassage(lastPassageSuccumbed)
	}
	if ge.pendingCatastrophe == "" {
		return fmt.Errorf("no pending catastrophe to succumb to")
	}
	epochKey := ge.pendingCatastrophe
	catName, _ := ge.rules.Catastrophe(epochKey)
	ep, _ := ge.rules.Era(epochKey)
	ge.setCatastropheOutcome(epochKey, CatastropheSuccumbed)

	newLegacy := !ge.legacyBonuses[epochKey]
	legacyNote := "Legacy bonus earned."
	if !newLegacy {
		legacyNote = "Legacy bonus already held."
	}
	ge.catastropheHistory = append(ge.catastropheHistory,
		fmt.Sprintf("Tick %d"+catHistSuccumbedMarker+"%s (%s). Civilization reset. %s", ge.tick, catName, ep.Name, legacyNote))

	ge.Buildings.GenerateRuins(ge.gameRNG(), SuccumbRuinCount)
	droppedRuins := ge.Buildings.EnforceRuinCap()
	savedRuins := ge.Buildings.GetAllRuins()

	ge.legacyBonuses[epochKey] = true
	// Era Mastery: a fall ends the run as a prestige does, so every age the
	// run completed gains a level through the same commit, and the rebuild
	// starts on known ground. The commit starts the run's furthest age over,
	// so a later prestige counts only what the rebuilt run completes: no age
	// is counted twice for one run.
	ge.Prestige.NoteAgeEntered(ge.age)
	masteryLine := masteryCommitLine(ge.rules, ge.Prestige.CommitRun())
	// The legacy kit remembers the fallen civilization's plan, research
	// order, civilizations met and shares before the reset.
	ge.captureLegacyLocked()
	savedLegacy := copyBoolMap(ge.legacyBonuses)
	savedCatHistory := append([]string(nil), ge.catastropheHistory...)
	savedEpochHistory := append([]EpochEventRecord(nil), ge.epochEventHistory...)
	savedPrestige := ge.Prestige

	// Full reset — Bus intentionally kept so dashboard subscriptions survive.
	ge.tick = 0
	ge.age = "primitive_age"
	ge.Resources = NewResourceManagerWith(ge.rules)
	ge.Buildings = ge.newBuildingManager()
	ge.Workers = NewWorkerManagerWith(ge.rules)
	ge.Research = NewResearchManagerWith(ge.rules)
	ge.Military = NewMilitaryManagerWith(ge.rules)
	ge.Events = NewEventManagerWith(ge.rules)
	ge.Milestones = NewMilestoneManagerWith(ge.rules)
	ge.Trade = NewTradeManagerWith(ge.rules)
	ge.Diplomacy = NewDiplomacyManagerWith(ge.rules)
	ge.Stats = NewGameStats()
	ge.permanentBonuses = make(map[string]float64)
	ge.buildQueue = nil
	ge.plan = nil
	ge.log = nil
	ge.speedMultiplier = 1.0
	ge.tickSpeedBonus = 0
	ge.ageReady = false
	ge.starvationTicks = 0
	// The festival and black market cooldowns are tick numbers, and the tick
	// counter just went back to 0: left as they were, a cooldown with 780
	// ticks to run at tick 5,001 had 5,781 to run in the new run. The
	// automatic expedition's countdown starts over too, as at a prestige.
	ge.festivalReadyTick, ge.blackMarketReadyTick = 0, 0
	ge.grantedFeatures = nil
	ge.autoExpeditionTicksLeft = 0
	ge.autoExpeditionStarved = false
	ge.currentEpoch = ge.rules.EraOf("primitive_age")
	ge.epochEventFired = make(map[string]bool)
	ge.clearHarbingerRun()
	ge.awakeningsFired = make(map[string]bool)
	ge.survivedEpochs = make(map[string]bool)
	ge.pendingCatastrophe = ""
	ge.pendingLastPassage = false
	ge.morale = 0.50
	ge.lowMoraleWarned = false
	// Fresh run after the fall: eligible to roll a new Ancient Memory.
	ge.ancientMemoryUsed = false
	ge.pendingMemoryTech = ""
	ge.startRunShares()

	// Restore persistent cross-run state
	ge.Prestige = savedPrestige
	ge.Buildings.LoadRuins(savedRuins)
	ge.legacyBonuses = savedLegacy
	ge.catastropheHistory = savedCatHistory
	ge.epochEventHistory = savedEpochHistory

	// Per-resource legacy rate bonuses live in permanentBonuses; the research
	// bonus is derived from the legacy flags (see succumbResearchFactor).
	ge.reapplyLegacyBonuses()

	ge.applyAgeUnlocks("primitive_age")

	// Size the stores before the starting stock lands, as a prestige does:
	// storage follows Era Mastery and the bonuses kept across the fall, and
	// stock added to a store still at its base size was cut off at the base
	// (50 food and 50 wood). It also puts the real caps and rates in the new
	// run's first snapshot instead of leaving them to the first tick.
	ge.recalculateRates()

	ge.Resources.Add("food", 15)
	ge.Resources.Add("wood", 12)
	for res, amount := range ge.Prestige.GetStartingResources() {
		ge.Resources.Add(res, amount)
	}

	ge.recalculateTickSpeed()

	ge.addLog("event", fmt.Sprintf("☄ %s: civilization has fallen. A new dawn.", catName))
	if newLegacy {
		ge.addLog("success", fmt.Sprintf("%s legacy bonus (permanent): %s.", ep.Name, legacyBonusText(ge.rules, epochKey)))
	} else {
		ge.addLog("info", fmt.Sprintf("The %s legacy was already yours; no new legacy bonus.", ep.Name))
	}
	ge.addLog("success", fmt.Sprintf("Ancient Knowledge: research time %s for each epoch succumbed in, now %s (permanent).", ResearchFactorText(SuccumbResearchTimeFactor), ResearchFactorText(ge.succumbResearchFactor())))
	if masteryLine != "" {
		ge.addLog("info", masteryLine)
	}
	if line := ge.masteryEntryLine(ge.age, 1); line != "" {
		ge.addLog("info", line)
	}
	if len(savedRuins) > 0 {
		ge.addLog("info", fmt.Sprintf("%s from fallen civilizations carry forward (max %d).", textfmt.Count(ge.Buildings.RuinTotal(), "ruin", "ruins"), MaxRuins))
	}
	if droppedRuins > 0 {
		ge.addLog("info", fmt.Sprintf("%s crumbled to make room.", textfmt.Count(droppedRuins, "older, lower-value ruin", "older, lower-value ruins")))
	}
	// The legacy kit: shares, the first age's template slice, old friends.
	ge.startRunLegacyLocked()
	ge.addLog("info", "Type [cyan]help[-] to rebuild.")

	// Roll for an Ancient Memory cache (only when this account has prestiged before;
	// a first-ever Succumb with no prestige history offers nothing — see the gate).
	ge.maybeOfferAncientMemory()

	return nil
}

// startReconstruction starts Endure's Reconstruction Effort, the timed
// all-production penalty, and returns how many ticks it lasts in the current
// age. Under the write lock.
func (ge *GameEngine) startReconstruction() int {
	debuff := ge.rules.StretchTicks(ge.age, endureDebuffTicks)
	ge.Events.InjectEvent(ActiveEvent{
		Key:       "endure_reconstruction",
		Name:      "Reconstruction Effort",
		TicksLeft: debuff,
		Effects: []config.Effect{
			{Type: "production_all", Value: endureDebuffProduction},
		},
	})
	return debuff
}

// reapplyLegacyBonuses restores the per-resource Succumb legacy rate bonuses
// into permanentBonuses after a reset. Must be called whenever permanentBonuses
// is cleared (prestige or succumb resets) so cross-run bonuses are not lost.
// The research bonus is not stored here; it is derived (succumbResearchFactor).
func (ge *GameEngine) reapplyLegacyBonuses() {
	for _, epochKey := range sortedKeys(ge.legacyBonuses) {
		if !ge.legacyBonuses[epochKey] {
			continue
		}
		for res, mult := range ge.rules.LegacyBonus(epochKey) {
			ge.permanentBonuses[res+"_rate"] += mult
		}
	}
}

// legacyEpochCount returns how many distinct epochs carry a Succumb legacy flag.
// Read-only.
func (ge *GameEngine) legacyEpochCount() int {
	n := 0
	for _, active := range ge.legacyBonuses {
		if active {
			n++
		}
	}
	return n
}

// AncientKnowledgeFactor is what research time is multiplied by after
// succumbing in epochs distinct epochs: SuccumbResearchTimeFactor for each,
// 1 with none. Multiplied out one at a time, so it is the same on every CPU.
func AncientKnowledgeFactor(epochs int) float64 {
	f := 1.0
	for i := 0; i < epochs; i++ {
		f = float64(f * SuccumbResearchTimeFactor)
	}
	return f
}

// ResearchFactorText words a research time factor for player text: "×0.8",
// "×0.64", "×0.26" (two decimals at most).
func ResearchFactorText(f float64) string {
	return "×" + strconv.FormatFloat(math.Round(f*100)/100, 'f', -1, 64)
}

// succumbResearchFactor is Ancient Knowledge as it stands: what a research
// started now has its time multiplied by, after the research speed pool has
// taken its share (ResearchTicks). Derived from the legacy flags, so it
// survives save/load, Succumb and DoPrestige without being stored. It is no
// part of the research speed pool: a pool is added up and taken off the
// listed time, and this multiplies what is left. Read-only.
func (ge *GameEngine) succumbResearchFactor() float64 {
	return AncientKnowledgeFactor(ge.legacyEpochCount())
}

// combinedResearchSpeed is the research speed pool: every research_speed
// bonus the resolver holds (techs, milestones, wonders). It is the pool the
// Stats panel lists, so what the panel shows is what a research started now
// gets. Ancient Knowledge is not in it (succumbResearchFactor). Read-only.
func (ge *GameEngine) combinedResearchSpeed() float64 {
	return ge.buildResolver().AddTotal("research_speed")
}

// countCatastropheOutcomes counts Endure and Succumb entries in the
// civilization log. Pure.
func countCatastropheOutcomes(history []string) (endured, succumbed int) {
	for _, h := range history {
		switch {
		case strings.Contains(h, catHistEnduredMarker):
			endured++
		case strings.Contains(h, catHistSuccumbedMarker):
			succumbed++
		}
	}
	return endured, succumbed
}

// restoreCatastropheState finishes loading catastrophe state from a save and
// migrates saves written before the overhaul. Call from LoadGame after the
// epoch, legacy and milestone fields are restored, under the write lock.
func (ge *GameEngine) restoreCatastropheState(save *GameSave) {
	// save.CatastropheFired (written by early builds of this change) is ignored:
	// the once-per-epoch rule is the transition roll's own epochEventFired.

	// Outcomes on catastrophe records from older saves. The newest record per
	// epoch is the one that matters; older duplicates can only be succumbs from
	// earlier runs (the history survives Succumb).
	seen := make(map[string]bool)
	for i := len(ge.epochEventHistory) - 1; i >= 0; i-- {
		r := &ge.epochEventHistory[i]
		if r.EventType != "catastrophe" {
			continue
		}
		latest := !seen[r.EpochKey]
		seen[r.EpochKey] = true
		if r.Outcome != "" {
			continue
		}
		switch {
		case latest && r.EpochKey == ge.pendingCatastrophe:
			r.Outcome = CatastrophePending
		case latest && ge.survivedEpochs[r.EpochKey]:
			r.Outcome = CatastropheEndured
		case ge.legacyBonuses[r.EpochKey]:
			r.Outcome = CatastropheSuccumbed
		}
	}

	// The Succumb research bonus used to be written into permanentBonuses (and
	// wiped by the next reset). It is now derived from the legacy flags, so strip
	// the stored copy: rebuild research_speed from this run's completed
	// milestones, its only other writer.
	if !save.SuccumbResearchDerived {
		total := 0.0
		for _, ms := range ge.rules.Milestones() {
			if !ge.Milestones.IsCompleted(ms.Key) {
				continue
			}
			for _, eff := range ms.Rewards {
				if eff.Type == "permanent_bonus" && eff.Target == "research_speed" {
					total += eff.Value
				}
			}
		}
		if total == 0 {
			delete(ge.permanentBonuses, "research_speed")
		} else {
			ge.permanentBonuses["research_speed"] = total
		}
	}
}

// legacyBonusText lists an epoch's legacy bonus in set for a log line: "iron
// +10% production, faith +5% production", in resource-key order.
func legacyBonusText(set *rules.Set, epochKey string) string {
	bonuses := set.LegacyBonus(epochKey)
	var parts []string
	for _, res := range sortedKeys(bonuses) {
		parts = append(parts, fmt.Sprintf("%s +%.0f%% production", resourceLabel(res), bonuses[res]*100))
	}
	if len(parts) == 0 {
		return "none"
	}
	return strings.Join(parts, ", ")
}
