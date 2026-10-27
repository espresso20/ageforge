package game

import (
	"fmt"
	"math"
	"math/rand"
	"regexp"
	"strconv"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/rules"
	"github.com/espresso20/ageforge/vibe"
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
	// CultureStrength is the current culture strength in [0,1], which sets
	// the tiers of good epoch event open at the next advance into a new era
	// (CultureTier, the best of them), and CultureDevotion and CultureKept
	// the two things it comes from (culture.go).
	CultureStrength float64
	CultureTier     CultureTier
	CultureDevotion float64
	CultureKept     float64
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
	return vibe.Errorf("game.catastrophe.catastropheBlockErr.01", "%s is upon you. Type 'catastrophe' to choose Endure or Succumb before %s", name, action)
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
	// The doom has a name now, warned of or not.
	ge.note(config.BadgeEvDoomNamed, epochKey)

	eventName := catName
	if source != catastropheRolled {
		eventName = catName + " (" + source + ")"
	}
	ge.addLog("warning", vibe.Sprintf("game.catastrophe.triggerCatastrophe.01", "☄ %s threatens the %s. Prepare yourself.", catName, ep.Name))
	ge.addLog("warning", vibe.Line("game.catastrophe.triggerCatastrophe.02", "  Type 'catastrophe' to choose Endure or Succumb. Advancing and prestige wait until you do."))

	ge.epochEventHistory = append(ge.epochEventHistory, EpochEventRecord{
		EpochKey: epochKey, EpochName: ep.Name,
		EventKey: ep.CatastropheKey, EventName: eventName, EventType: "catastrophe",
		Tick: ge.tick, Outcome: CatastrophePending, Source: source,
	})
	ge.publish(EventEpochEventFired, payload{
		"event_key":  ep.CatastropheKey,
		"event_name": catName,
		"event_type": "catastrophe",
		"epoch_key":  epochKey,
		"source":     source,
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
	culture := ge.cultureStrength()
	out.CultureStrength, out.CultureTier = culture, CultureTierAt(culture)
	out.CultureDevotion, out.CultureKept = CultureDevotionOf(ge.cultureMeasure), CultureKeptOf(ge.Resources.Get("culture"), ge.cultureMeasure)
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
	for _, key := range sortedKeys(destroyed) {
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
		return vibe.Errorf("game.catastrophe.Endure.01", "no pending catastrophe to endure")
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
	invited := ge.doomInvited(epochKey)
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

	ge.addLog("warning", vibe.Sprintf("game.catastrophe.Endure.02", "☄ ENDURE: %s. [comment]%s[-]", catName, catFlavor))
	ge.addLog("warning", vibe.Sprintf("game.catastrophe.Endure.03", "  Buildings destroyed: %d", destroyCount))
	for _, desc := range names {
		ge.addLogAbout("warning", TagBad, vibe.Sprintf("game.catastrophe.Endure.04", "  → %s lost", desc))
	}
	if brace > 0 {
		ge.addLogAbout("info", TagBad, vibe.Sprintf("game.catastrophe.Endure.05", "  Braced (level %d): %d%% of buildings lost instead of %d%%, %.0f%% of stock kept instead of %.0f%%.",
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
	ge.addLog("warning", vibe.Sprintf("game.catastrophe.Endure.06", "  All resources reduced to %.0f%% of stored amounts.", keep*100))
	ge.addLogAbout("warning", TagPeople|TagBad, vibe.Sprintf("game.catastrophe.Endure.07", "  %.0f%% of workers lost.", endureWorkerLoss*100))
	ge.addLog("info", vibe.Sprintf("game.catastrophe.Endure.08", "  Reconstruction: all production %.0f%% for %s. Morale %+.0f points.",
		endureDebuffProduction*100, approxTicks(debuff, ge.tickIntervalLocked()), endureMoraleHit*100))
	// Past the all-production cap the surplus absorbs the penalty: say so.
	ge.logCapped(config.Effect{Type: "production_all", Value: endureDebuffProduction})
	ge.addLog("success", vibe.Sprintf("game.catastrophe.Endure.09", "  ✦ You endured the %s. Its badge records it.", epName))
	// Cosmetic flavour — a wry beat after surviving the catastrophe.
	if q := config.PickLogFlavor(config.LogFlavorCatastropheSurvived, ge.quipRNG()); q != "" {
		ge.addLog("info", fmt.Sprintf("  [comment]%s[-]", q))
	}

	ge.catastropheHistory = append(ge.catastropheHistory,
		fmt.Sprintf("Tick %d"+catHistEnduredMarker+"%s (%s). %d buildings lost.", ge.tick, doomMark(epochKey), eraMark(epochKey), destroyCount))

	ge.applyMorale(endureMoraleHit)
	ge.report(Event{Kind: config.BadgeEvEndured, Subject: epochKey, Attrs: map[string]float64{
		"brace": float64(brace), "saved": float64(outcome.BuildingsSaved),
		"garrison": float64(outcome.Garrison), "invited": boolFact(invited),
	}})
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
		return vibe.Errorf("game.catastrophe.Succumb.01", "no pending catastrophe to succumb to")
	}
	epochKey := ge.pendingCatastrophe
	catName, _ := ge.rules.Catastrophe(epochKey)
	ep, _ := ge.rules.Era(epochKey)
	ge.setCatastropheOutcome(epochKey, CatastropheSuccumbed)
	// Reported while the run that fell is still here to be judged against.
	ge.report(Event{Kind: config.BadgeEvSuccumbed, Subject: epochKey,
		Attrs: map[string]float64{"invited": boolFact(ge.doomInvited(epochKey))}})

	newLegacy := !ge.legacyBonuses[epochKey]
	legacyNote := "Legacy bonus earned."
	if !newLegacy {
		legacyNote = "Legacy bonus already held."
	}
	ge.catastropheHistory = append(ge.catastropheHistory,
		fmt.Sprintf("Tick %d"+catHistSuccumbedMarker+"%s (%s). Civilization reset. %s", ge.tick, doomMark(epochKey), eraMark(epochKey), legacyNote))

	ge.Buildings.GenerateRuins(ge.gameRNG(), SuccumbRuinCount)
	droppedRuins := ge.Buildings.EnforceRuinCap()
	savedRuins := ge.Buildings.GetAllRuins()

	ge.legacyBonuses[epochKey] = true
	// Era Mastery: a fall ends the run as a prestige does, so every age the
	// run completed gains a level through the same commit, and the rebuild
	// starts on known ground. The commit starts the run's furthest age over,
	// so a later prestige counts only what the rebuilt run completes: no age
	// is counted twice for one run.
	ge.setPrestigeFloorLocked()
	masteryLine := masteryCommitLine(ge.rules, ge.Prestige.CommitRun())
	// The legacy kit remembers the fallen civilization's plan, research
	// order, civilizations met and shares before the reset.
	ge.captureLegacyLocked()
	// The ending's record, for whoever tells it (run_ending.go).
	end := RunEnding{Kind: RunEndFallen, Age: ge.age, Level: ge.Prestige.GetLevel(), Catastrophe: catName}
	runStarted := ge.Stats.GameStarted

	// Full reset, as at a prestige (resetRunLocked); morale starts lower.
	// Prestige, the legacy flags and both histories carry over untouched.
	ge.resetRunLocked(0.50)
	// The tick speed bonus and a pending Last Passage start over too.
	ge.tickSpeedBonus = 0
	ge.pendingLastPassage = false

	ge.beginNextRunLocked(savedRuins)

	ge.sayEnding(&end, EndLineVerdict, "event", vibe.Sprintf("game.catastrophe.Succumb.06", "☄ %s: civilization has fallen. A new dawn.", catName))
	if newLegacy {
		ge.sayEnding(&end, EndLineLegacy, "success", vibe.Sprintf("game.catastrophe.Succumb.07", "%s legacy bonus (permanent): %s.", ep.Name, legacyBonusText(ge.rules, epochKey)))
	} else {
		ge.sayEnding(&end, EndLineLegacy, "info", vibe.Sprintf("game.catastrophe.Succumb.08", "The %s legacy was already yours; no new legacy bonus.", ep.Name))
	}
	ge.sayEnding(&end, EndLineKnowledge, "success", vibe.Sprintf("game.catastrophe.Succumb.09", "Ancient Knowledge: research time %s for each epoch succumbed in, now %s (permanent).", ResearchFactorText(SuccumbResearchTimeFactor), ResearchFactorText(ge.succumbResearchFactor())))
	if masteryLine != "" {
		ge.sayEnding(&end, EndLineMastery, "info", masteryLine)
	}
	if line := ge.masteryEntryLine(ge.age, 1); line != "" {
		ge.sayEnding(&end, EndLineGround, "info", line)
	}
	if len(savedRuins) > 0 {
		ge.sayEnding(&end, EndLineRuins, "info", vibe.Sprintf("game.catastrophe.Succumb.10", "%s from fallen civilizations carry forward (max %d).", textfmt.Count(ge.Buildings.RuinTotal(), vibe.Line("game.catastrophe.Succumb.11", "ruin"), vibe.Line("game.catastrophe.Succumb.12", "ruins")), MaxRuins))
	}
	if droppedRuins > 0 {
		ge.sayEnding(&end, EndLineRuins, "info", vibe.Sprintf("game.catastrophe.Succumb.13", "%s crumbled to make room.", textfmt.Count(droppedRuins, vibe.Line("game.catastrophe.Succumb.14", "older, lower-value ruin"), vibe.Line("game.catastrophe.Succumb.15", "older, lower-value ruins"))))
	}
	// The legacy kit: shares, the first age's template slice, old friends.
	kitFrom := len(ge.log)
	ge.startRunLegacyLocked()
	ge.takeEndingLines(&end, kitFrom, EndLineKit)
	ge.addLog("info", vibe.Line("game.catastrophe.Succumb.16", "Type [cyan]help[-] to rebuild."))
	// The run that starts here is a new civilization.
	ge.noteCivilizationStartedLocked()

	// Roll for an Ancient Memory cache (only when this account has prestiged before;
	// a first-ever Succumb with no prestige history offers nothing — see the gate).
	ge.maybeOfferAncientMemory()

	ge.publishRunEnded(end, runStarted)
	return nil
}

// startReconstruction starts Endure's Reconstruction Effort, the timed
// all-production penalty, and returns how many ticks it lasts in the current
// age. Under the write lock.
func (ge *GameEngine) startReconstruction() int {
	debuff := ge.rules.StretchTicks(ge.age, endureDebuffTicks)
	ge.Events.InjectEvent(ActiveEvent{
		Key:       "endure_reconstruction",
		Src:       nameFromCode,
		Ref:       "endure_reconstruction",
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

// legacyBonusText lists an epoch's legacy bonus in set for a log line: "iron
// +10% production, faith +5% production", in resource-key order.
func legacyBonusText(set *rules.Set, epochKey string) string {
	bonuses := set.LegacyBonus(epochKey)
	var parts []string
	for _, res := range sortedKeys(bonuses) {
		parts = append(parts, vibe.Sprintf("game.catastrophe.legacyBonusText.01", "%s +%.0f%% production", resourceLabel(res), bonuses[res]*100))
	}
	if len(parts) == 0 {
		return "none"
	}
	return strings.Join(parts, ", ")
}

// recordNames are the names an epoch event record carries, drawn from the
// words of the ruleset in force: the era's, and the event's or the doom's. It
// reports false for a record whose names are not drawn from them (one written
// by hand, or whose event the ruleset does not have).
func (ge *GameEngine) recordNames(r EpochEventRecord) (epoch, event string, ok bool) {
	era, found := ge.rules.Era(r.EpochKey)
	if !found {
		return "", "", false
	}
	if r.EventType == "catastrophe" {
		event, _ = ge.rules.Catastrophe(r.EpochKey)
		if r.Source != catastropheRolled {
			event += " (" + r.Source + ")"
		}
		return era.Name, event, true
	}
	ev, found := ge.rules.EraEvent(r.EventKey)
	return era.Name, ev.Name, found
}

// redrawHistory writes the names in the epoch event history again from the
// words of the ruleset in force. The history holds the names it was told in,
// so this runs when the engine moves onto another ruleset and after a load.
func (ge *GameEngine) redrawHistory() {
	for i, r := range ge.epochEventHistory {
		if epoch, event, ok := ge.recordNames(r); ok {
			ge.epochEventHistory[i].EpochName, ge.epochEventHistory[i].EventName = epoch, event
		}
	}
}

// epochHistoryForSave is the history as a save holds it: a record keeps the
// keys its names are drawn from and not the names, so a save reads the same
// in any script (VibeMode).
func (ge *GameEngine) epochHistoryForSave() []EpochEventRecord {
	out := append([]EpochEventRecord(nil), ge.epochEventHistory...)
	for i, r := range out {
		if _, _, ok := ge.recordNames(r); ok {
			out[i].EpochName, out[i].EventName = "", ""
		}
	}
	return out
}

// The civilization log (catastropheHistory) holds its lines with the names
// they tell of as markers, {doom:ERA}, {era:ERA} and {passage}, and not as
// the names: a save then reads the same in any script (VibeMode). The
// markers are written in with the names of the ruleset in force when a
// state is read (historyLines).
var historyMarkRe = regexp.MustCompile(`\{(doom|era):([a-z_]+)\}|\{passage\}`)

func doomMark(epochKey string) string { return "{doom:" + epochKey + "}" }
func eraMark(epochKey string) string  { return "{era:" + epochKey + "}" }

const passageMark = "{passage}"

// historyLines is the civilization log as read now: every marker replaced by
// the name it stands for in the words of the ruleset in force.
func (ge *GameEngine) historyLines() []string {
	out := make([]string, len(ge.catastropheHistory))
	for i, line := range ge.catastropheHistory {
		if vibe.On() {
			if told, ok := ge.vibeHistoryLine(line); ok {
				out[i] = told
				continue
			}
		}
		out[i] = historyMarkRe.ReplaceAllStringFunc(line, func(m string) string {
			parts := historyMarkRe.FindStringSubmatch(m)
			switch parts[1] {
			case "doom":
				name, _ := ge.rules.Catastrophe(parts[2])
				return name
			case "era":
				return eraName(ge.rules, parts[2])
			}
			name, _ := ge.rules.LastPassage()
			return name
		})
	}
	return out
}

// The four lines the civilization log is written as, in plain English with
// their markers: a save holds them so, and they are the key the line is told
// again by when VibeMode is on (vibeHistoryLine).
var (
	historyEnduredRe   = regexp.MustCompile(`^Tick (\d+)` + regexp.QuoteMeta(catHistEnduredMarker) + `\{doom:([a-z_]+)\} \(\{era:([a-z_]+)\}\)\. (\d+) buildings lost\.$`)
	historySuccumbedRe = regexp.MustCompile(`^Tick (\d+)` + regexp.QuoteMeta(catHistSuccumbedMarker) + `\{doom:([a-z_]+)\} \(\{era:([a-z_]+)\}\)\. Civilization reset\. (Legacy bonus (?:earned|already held)\.)$`)
	historyPassageEnRe = regexp.MustCompile(`^Tick (\d+)` + regexp.QuoteMeta(catHistEnduredMarker) + `\{passage\} \(\{era:([a-z_]+)\}\)\. Prestige kept (\d+) of (\d+) points\.$`)
	historyPassageSuRe = regexp.MustCompile(`^Tick (\d+)` + regexp.QuoteMeta(catHistSuccumbedMarker) + `\{passage\} \(\{era:([a-z_]+)\}\)\. No prestige points\. Cosmic Legacy earned\.$`)
)

// vibeHistoryLine tells a line of the civilization log in the words of the
// script in force, from the plain English the save holds it in. It reports
// false for a line it does not know (one written by another version), which
// is then read as it was written.
func (ge *GameEngine) vibeHistoryLine(line string) (string, bool) {
	tick := func(s string) int { n, _ := strconv.Atoi(s); return n }
	doom := func(key string) string { name, _ := ge.rules.Catastrophe(key); return name }
	era := func(key string) string { return eraName(ge.rules, key) }
	passage := func() string { name, _ := ge.rules.LastPassage(); return name }
	switch {
	case historyEnduredRe.MatchString(line):
		m := historyEnduredRe.FindStringSubmatch(line)
		return vibe.Sprintf("game.catastrophe.historyLines.01", "Tick %d — Endured %s (%s). %d buildings lost.",
			tick(m[1]), doom(m[2]), era(m[3]), tick(m[4])), true
	case historySuccumbedRe.MatchString(line):
		m := historySuccumbedRe.FindStringSubmatch(line)
		note := vibe.Line("game.catastrophe.historyLines.03", "Legacy bonus earned.")
		if strings.Contains(m[4], "already held") {
			note = vibe.Line("game.catastrophe.historyLines.04", "Legacy bonus already held.")
		}
		return vibe.Sprintf("game.catastrophe.historyLines.02", "Tick %d — Succumbed to %s (%s). Civilization reset. %s",
			tick(m[1]), doom(m[2]), era(m[3]), note), true
	case historyPassageEnRe.MatchString(line):
		m := historyPassageEnRe.FindStringSubmatch(line)
		return vibe.Sprintf("game.catastrophe.historyLines.05", "Tick %d — Endured %s (%s). Prestige kept %d of %d points.",
			tick(m[1]), passage(), era(m[2]), tick(m[3]), tick(m[4])), true
	case historyPassageSuRe.MatchString(line):
		m := historyPassageSuRe.FindStringSubmatch(line)
		return vibe.Sprintf("game.catastrophe.historyLines.06", "Tick %d — Succumbed to %s (%s). No prestige points. Cosmic Legacy earned.",
			tick(m[1]), passage(), era(m[2])), true
	}
	return "", false
}
