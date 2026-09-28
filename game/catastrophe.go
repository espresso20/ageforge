package game

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"

	"github.com/espresso20/ageforge/config"
)

// Civilizational catastrophes (Phase 9).
//
// A catastrophe strikes at an epoch transition (a bad epoch roll escalates).
// Players cannot trigger one directly; the dev console's /catastrophe does, for
// testing. Either way it only sets pendingCatastrophe: nothing is destroyed
// until the player chooses Endure or Succumb. While a catastrophe is pending the game keeps running but AdvanceAge
// and DoPrestige refuse, so the choice cannot be skipped or overwritten.
//
// Rules in one place:
//   - No catastrophe before config.CatastropheGateEpoch (the Iron Era).
//   - At most one random catastrophe per epoch per run: each epoch's
//     transition rolls once (epochEventFired).
//   - All randomness comes from the seeded ge.rng over stable pools.
//
// Every method here that is not exported expects the engine write lock to be
// held, except the read-only helpers used by GetState (catastropheOutlook,
// succumbResearchBonus, legacyModifiers, countCatastropheOutcomes), which only
// read state and are safe under the read lock.

const (
	// Epoch transition good-event chance by faith storage fill.
	epochGoodChanceLowFaith  = 0.40 // faith fill < 25%
	epochGoodChanceBase      = 0.50 // 25%–75%, or faith has no storage yet
	epochGoodChanceHighFaith = 0.60 // faith fill > 75%

	// catastropheChanceOnBadRoll is the chance a bad epoch roll escalates into a
	// catastrophe. Overall odds per transition are (1-goodChance) × this:
	// 18% at low faith, 15% at mid faith, 12% at high faith.
	catastropheChanceOnBadRoll = 0.30

	// Endure consequences.
	// Endure destroys 20% of non-wonder buildings (floor) unbraced; see
	// braceDestroyPct in harbinger.go for the braced shares.
	endureResourceKeep     = 0.15 // resources drop to 15% of stored amounts
	endureWorkerLoss       = 0.25 // 25% of the worker pool is lost
	endureDebuffTicks      = 216  // Reconstruction Effort duration
	endureDebuffProduction = -0.10
	endureMoraleHit        = -0.10

	// Succumb consequences.

	// SuccumbRuinCount is how many buildings Succumb turns into ruins.
	SuccumbRuinCount = 8
	// SuccumbResearchBonusPerEpoch is the permanent research_speed bonus per
	// distinct epoch succumbed (legacy flag), stacking across epochs.
	SuccumbResearchBonusPerEpoch = 0.25
)

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

// CatastropheOutlook describes the catastrophe odds at the NEXT passage, given
// the current faith fill and the rules above. The passage is the next epoch
// transition, or in the final epoch (which has none) prestige itself, where the
// Last Passage can strike (last_passage.go).
type CatastropheOutlook struct {
	// Passage is PassageEpoch, or PassagePrestige in the final epoch.
	Passage string
	// NextEpochKey is the epoch the next transition enters; "" in the final
	// epoch, whose passage is prestige.
	NextEpochKey string
	// Possible is false when the next passage cannot bring a catastrophe at
	// all: it is before the Iron-epoch gate, its transition roll already
	// happened this run, or the Last Passage is already pending.
	Possible bool
	// Probability is the chance in [0,1] that the next passage produces a
	// catastrophe. 0 when !Possible; 1 when a catastrophe has been invited.
	Probability float64
	// Tier buckets Probability: none / low / medium / high.
	Tier CatastropheTier
	// FaithFill is the current faith fill (amount / storage) in [0,1] that
	// drives the odds; 0 when faith has no storage yet.
	FaithFill float64
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

// faithFill returns faith amount / storage clamped to [0,1], and whether faith
// has any storage at all. Read-only.
func (ge *GameEngine) faithFill() (float64, bool) {
	storage := ge.Resources.GetStorage("faith")
	if storage <= 0 {
		return 0, false
	}
	fill := ge.Resources.Get("faith") / storage
	if fill < 0 {
		fill = 0
	} else if fill > 1 {
		fill = 1
	}
	return fill, true
}

// epochGoodChance returns the chance that an epoch transition rolls a good
// event, gated by faith fill. Read-only.
func (ge *GameEngine) epochGoodChance() float64 {
	fill, ok := ge.faithFill()
	switch {
	case !ok:
		return epochGoodChanceBase
	case fill < 0.25:
		return epochGoodChanceLowFaith
	case fill > 0.75:
		return epochGoodChanceHighFaith
	}
	return epochGoodChanceBase
}

// catastropheCanStrike reports whether a catastrophe may be triggered in
// epochKey right now: past the Iron-epoch gate and nothing pending (a pending
// catastrophe is never overwritten). Read-only.
func (ge *GameEngine) catastropheCanStrike(epochKey string) bool {
	return ge.pendingCatastrophe == "" && config.CatastropheAllowed(epochKey)
}

// catastropheBlockErr is the error AdvanceAge and DoPrestige return while a
// catastrophe is pending. action is a gerund ("advancing", "prestiging").
func (ge *GameEngine) catastropheBlockErr(action string) error {
	name, _ := config.CatastropheInfo(ge.pendingCatastrophe)
	return fmt.Errorf("%s is upon you — type 'catastrophe' to choose Endure or Succumb before %s", name, action)
}

// How a catastrophe came about; it only changes the log line and event name.
const (
	catastropheRolled  = ""        // bad transition roll escalated
	catastropheForced  = "forced"  // dev console /catastrophe
	catastropheInvited = "invited" // honoured invite (the Harbinger's Invite)
)

// triggerCatastrophe makes epochKey's catastrophe pending: log line, history
// record, and a bus event so the dashboard toast fires. Callers must check
// catastropheCanStrike first. Must be called under the write lock; the bus
// handlers it reaches must not take the engine lock (see CLAUDE.md).
func (ge *GameEngine) triggerCatastrophe(epochKey, source string) {
	ep := config.EpochByKey()[epochKey]
	catName, _ := config.CatastropheInfo(epochKey)
	ge.pendingCatastrophe = epochKey

	eventName := catName
	if source != catastropheRolled {
		eventName = catName + " (" + source + ")"
	}
	ge.addLog("warning", fmt.Sprintf("☄ %s threatens the %s — prepare yourself.", catName, ep.Name))
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
// catastrophe, but it ignores the once-per-transition roll.
func (ge *GameEngine) forceCatastrophe() error {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	if ge.pendingCatastrophe != "" || ge.pendingLastPassage {
		return fmt.Errorf("a catastrophe is already pending")
	}
	if !config.CatastropheAllowed(ge.currentEpoch) {
		gate := config.EpochByKey()[config.CatastropheGateEpoch]
		return fmt.Errorf("catastrophes cannot strike before the %s", gate.Name)
	}
	ge.triggerCatastrophe(ge.currentEpoch, catastropheForced)
	return nil
}

// Invite (armed by the Harbinger's Invite action; see harbinger.go).
//
// catastropheInvited, when set, makes the next epoch transition into an epoch
// allowed by the Iron gate produce a catastrophe instead of rolling. It is
// consumed when honoured, kept while the target is gated or another
// catastrophe is pending, persisted, and cleared by Succumb and prestige.

// inviteCatastrophe arms the invite. Must be called under the write lock.
func (ge *GameEngine) inviteCatastrophe() { ge.catastropheInvited = true }

// honourInvite triggers the invited catastrophe for epochKey if the invite is
// armed and the catastrophe can strike there. Reports whether it did. Called
// by rollEpochEvent under the write lock.
func (ge *GameEngine) honourInvite(epochKey string) bool {
	if !ge.catastropheInvited || !ge.catastropheCanStrike(epochKey) {
		return false
	}
	ge.catastropheInvited = false
	ge.triggerCatastrophe(epochKey, catastropheInvited)
	return true
}

// CatastropheOutlook reports the catastrophe odds at the next passage: the next
// epoch transition, or prestige in the final epoch. Takes the read lock; use catastropheOutlook from code that already holds a lock.
func (ge *GameEngine) CatastropheOutlook() CatastropheOutlook {
	ge.mu.RLock()
	defer ge.mu.RUnlock()
	return ge.catastropheOutlook()
}

// catastropheOutlook is the lock-free body of CatastropheOutlook. Read-only:
// safe under either lock, and from GetState.
func (ge *GameEngine) catastropheOutlook() CatastropheOutlook {
	fill, _ := ge.faithFill()
	out := CatastropheOutlook{Passage: PassageEpoch, Tier: CatastropheTierNone, FaithFill: fill}
	next, ok := config.NextEpoch(ge.currentEpoch)
	switch {
	case ok:
		out.NextEpochKey = next.Key
		if !config.CatastropheAllowed(next.Key) || ge.epochEventFired[next.Key] {
			return out
		}
	case config.IsFinalEpoch(ge.currentEpoch):
		// The final epoch's passage is prestige: the Last Passage.
		out.Passage = PassagePrestige
		if !ge.lastPassageApplies() || ge.pendingLastPassage {
			return out
		}
	default:
		return out
	}
	out.Possible = true
	out.Probability = (1 - ge.epochGoodChance()) * catastropheChanceOnBadRoll * ge.harbingerAppeaseMultiplier()
	if ge.catastropheInvited {
		out.Probability = 1
	}
	out.Tier = catastropheTierFor(out.Probability)
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
//   - 20% of destroyable (non-wonder) buildings destroyed, at least 1 if any;
//     workers assigned to them return to the idle pool (15% / 10% when the
//     harbinger was braced at level 1 / 2)
//   - all unlocked resources drop to 15% of their stored amounts (30% / 45%
//     braced)
//   - 25% of the worker pool lost, spread proportionally over every building
//   - Reconstruction Effort: production -10% for 216 ticks
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
	// A Brace bought from the harbinger softens the blow.
	brace := ge.pendingBraceLevel
	if brace < 0 || brace > HarbingerMaxBrace {
		brace = 0
	}
	ge.pendingBraceLevel = 0
	destroyPct, keep := braceDestroyPct[brace], braceKeepFrac[brace]
	ge.survivedEpochs[epochKey] = true
	ge.setCatastropheOutcome(epochKey, CatastropheEndured)

	catName, catFlavor := config.CatastropheInfo(epochKey)
	epName := config.EpochByKey()[epochKey].Name

	destroyable := ge.Buildings.DestroyableCount()
	destroyCount := destroyable * destroyPct / 100
	if destroyCount < 1 && destroyable > 0 {
		destroyCount = 1
	}
	destroyed, names := ge.Buildings.DestroyRandom(ge.gameRNG(), destroyCount)
	ge.releaseWorkersFrom(destroyed)

	for key, r := range ge.Resources.resources {
		if r != nil && ge.Resources.IsUnlocked(key) {
			r.Amount *= keep
		}
	}

	ge.Workers.RemovePct(endureWorkerLoss)

	ge.Events.InjectEvent(ActiveEvent{
		Key:       "endure_reconstruction",
		Name:      "Reconstruction Effort",
		TicksLeft: endureDebuffTicks,
		Effects: []config.Effect{
			{Type: "production_all", Value: endureDebuffProduction},
		},
	})

	ge.addLog("warning", fmt.Sprintf("☄ ENDURE: %s — %s", catName, catFlavor))
	ge.addLog("warning", fmt.Sprintf("  Buildings destroyed: %d", destroyCount))
	for _, desc := range names {
		ge.addLog("warning", fmt.Sprintf("  → %s lost", desc))
	}
	if brace > 0 {
		ge.addLog("info", fmt.Sprintf("  Braced (level %d): %d%% of buildings lost instead of 20%%, %.0f%% of stock kept instead of 15%%.", brace, destroyPct, keep*100))
	}
	ge.addLog("warning", fmt.Sprintf("  All resources reduced to %.0f%% of stored amounts.", keep*100))
	ge.addLog("warning", "  25% of workers lost.")
	ge.addLog("info", fmt.Sprintf("  Timed: production -10%% for %d ticks (reconstruction period).", endureDebuffTicks))
	ge.addLog("success", fmt.Sprintf("  ✦ Survived marker earned for %s badge.", epName))
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
//   - up to 8 random non-wonder buildings become ruins (50% output, no workers),
//     carried into the next run; the ruin total is capped at MaxRuins
//   - the epoch's legacy flag is set: its per-resource legacy bonus and
//     +25% research speed (per distinct epoch succumbed, stacking) are permanent
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
	catName, _ := config.CatastropheInfo(epochKey)
	ep := config.EpochByKey()[epochKey]
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
	savedLegacy := copyBoolMap(ge.legacyBonuses)
	savedCatHistory := append([]string(nil), ge.catastropheHistory...)
	savedEpochHistory := append([]EpochEventRecord(nil), ge.epochEventHistory...)
	savedPrestige := ge.Prestige

	// Full reset — Bus intentionally kept so dashboard subscriptions survive.
	ge.tick = 0
	ge.age = "primitive_age"
	ge.Resources = NewResourceManager()
	ge.Buildings = NewBuildingManager()
	ge.Workers = NewWorkerManager()
	ge.Research = NewResearchManager()
	ge.Military = NewMilitaryManager()
	ge.Events = NewEventManager()
	ge.Milestones = NewMilestoneManager()
	ge.Trade = NewTradeManager()
	ge.Diplomacy = NewDiplomacyManager()
	ge.Stats = NewGameStats()
	ge.permanentBonuses = make(map[string]float64)
	ge.buildQueue = nil
	ge.plan = nil
	ge.log = nil
	ge.speedMultiplier = 1.0
	ge.tickSpeedBonus = 0
	ge.ageReady = false
	ge.starvationTicks = 0
	ge.currentEpoch = config.EpochForAge("primitive_age")
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

	// Restore persistent cross-run state
	ge.Prestige = savedPrestige
	ge.Buildings.LoadRuins(savedRuins)
	ge.legacyBonuses = savedLegacy
	ge.catastropheHistory = savedCatHistory
	ge.epochEventHistory = savedEpochHistory

	// Per-resource legacy rate bonuses live in permanentBonuses; the research
	// bonus is derived from the legacy flags (see legacyModifiers).
	ge.reapplyLegacyBonuses()

	ge.applyAgeUnlocks("primitive_age")
	ge.Resources.Add("food", 15)
	ge.Resources.Add("wood", 12)
	for res, amount := range ge.Prestige.GetStartingResources() {
		ge.Resources.Add(res, amount)
	}

	ge.recalculateTickSpeed()

	ge.addLog("event", fmt.Sprintf("☄ %s — civilization has fallen. A new dawn.", catName))
	if newLegacy {
		ge.addLog("success", fmt.Sprintf("Legacy Bonus: %s production permanently boosted.", ep.Name))
	} else {
		ge.addLog("info", fmt.Sprintf("The %s legacy was already yours; no new legacy bonus.", ep.Name))
	}
	ge.addLog("success", fmt.Sprintf("Ancient Knowledge: research speed +%.0f%% (permanent, +25%% per epoch succumbed).", ge.succumbResearchBonus()*100))
	if len(savedRuins) > 0 {
		ge.addLog("info", fmt.Sprintf("%d ruin(s) from fallen civilizations carry forward (max %d).", ge.Buildings.RuinTotal(), MaxRuins))
	}
	if droppedRuins > 0 {
		ge.addLog("info", fmt.Sprintf("%d older, lower-value ruin(s) crumbled to make room.", droppedRuins))
	}
	ge.addLog("info", "Type [cyan]help[-] to rebuild.")

	// Roll for an Ancient Memory cache (only when this account has prestiged before;
	// a first-ever Succumb with no prestige history offers nothing — see the gate).
	ge.maybeOfferAncientMemory()

	return nil
}

// reapplyLegacyBonuses restores the per-resource Succumb legacy rate bonuses
// into permanentBonuses after a reset. Must be called whenever permanentBonuses
// is cleared (prestige or succumb resets) so cross-run bonuses are not lost.
// The research bonus is not stored here; it is derived (legacyModifiers).
func (ge *GameEngine) reapplyLegacyBonuses() {
	for _, epochKey := range sortedKeys(ge.legacyBonuses) {
		if !ge.legacyBonuses[epochKey] {
			continue
		}
		for res, mult := range config.LegacyBonusForEpoch(epochKey) {
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

// succumbResearchBonus is the permanent research_speed bonus from Succumb:
// +25% per distinct epoch succumbed. Derived from the legacy flags, so it
// survives save/load, Succumb and DoPrestige without being stored. Read-only.
func (ge *GameEngine) succumbResearchBonus() float64 {
	return float64(float64(ge.legacyEpochCount()) * SuccumbResearchBonusPerEpoch)
}

// legacyModifiers emits the derived Succumb research bonus into the resolver
// (Source "legacy"), so it shows in the Active Multipliers panel. Read-only.
func (ge *GameEngine) legacyModifiers() []Modifier {
	b := ge.succumbResearchBonus()
	if b == 0 {
		return nil
	}
	return []Modifier{{Source: "legacy", Target: "research_speed", Op: OpAdd, Value: b}}
}

// combinedResearchSpeed sums research_speed from techs, permanent bonuses
// (milestones), prestige upgrades and the Succumb legacy bonus. Read-only.
func (ge *GameEngine) combinedResearchSpeed() float64 {
	return ge.Research.GetBonus("research_speed") +
		ge.permanentBonuses["research_speed"] +
		ge.Prestige.GetBonuses()["research_speed"] +
		ge.succumbResearchBonus()
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
		for _, ms := range config.Milestones() {
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
