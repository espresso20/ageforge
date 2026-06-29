package game

import (
	"fmt"
	"math/rand"

	"github.com/espresso20/ageforge/config"
)

// Diplomacy tuning constants. Kept here (not in config) because they govern
// engine-side cadence rather than per-civ data.
const (
	// Opinion bounds.
	opinionMin = -100
	opinionMax = 100

	// Passive personality drift fires on this tick cadence (every 25 ticks ≈
	// 50s of real time) so drift is gradual and does not swamp embassy gains.
	driftInterval = 25

	// War starts only when opinion is below this AND a provocation threshold is
	// crossed. Both conditions are required — anger alone never starts a war.
	warOpinionThreshold = -75

	// Provocations needed to push a sub-threshold civ into war: one raided trade
	// route counts as 1, each embargo counts as 1 — two embargoes (or a raid +
	// embargo) trips it.
	warProvocationThreshold = 2

	// A war auto-ends after this many provocation-free ticks (the "wait them
	// out" peace path). Sending tribute ends it immediately.
	warCooldownTicks = 300

	// Worker-lending: a lent batch stays for this many ticks before returning,
	// unless the lending civ's opinion is above lendPermanentOpinion (then the
	// workers stay permanently).
	lendDurationTicks    = 200
	lendPermanentOpinion = 80
)

// DiplomacyManager handles NPC civilization discovery and diplomatic relations.
// The 6 original factions plus 5 new civilizations form an 11-civ roster spanning
// all epochs. Civs are discovered through age/epoch-gated first-contact events.
// Once discovered, opinion drifts according to each civ's personality and the
// player's actions; high-opinion peaceful civs lend workers; provoked low-opinion
// aggressive civs may declare war (event-driven raids, no combat).
//
// All mutation happens under the engine write lock (DiplomacyManager has no lock
// of its own); callers must hold ge.mu when invoking mutating methods.
type DiplomacyManager struct {
	factions map[string]*FactionState

	// lentBatches tracks worker loans in flight so they can be returned on time.
	lentBatches []LentWorkerBatch

	// pendingLends / pendingReturns / pendingWar / pendingRaids accumulate side
	// effects produced during Tick that the engine must apply to other systems
	// (the worker pool, the event log). They are drained via TakePending*.
	pendingLends   []LendRequest
	pendingReturns []int
	pendingRaids   []RaidRequest
}

// FactionState tracks the relationship with one NPC civilization.
// Opinion range: -100 (hostile) to +100 (beloved).
// Status transitions: neutral → friendly (opinion ≥ 25 after gift) → allied (opinion ≥ 50, 500 gold).
// rival and embargo both cause -5 opinion per 50 ticks.
type FactionState struct {
	Discovered bool
	Opinion    int    // -100 to 100
	Status     string // "neutral", "friendly", "allied", "rival", "embargo"
	TradeCount int

	// OpinionAccum is a transient sub-1.0/tick accumulator (e.g. embassy
	// trickle). Not persisted — it spills into Opinion as it crosses 1.0.
	OpinionAccum float64

	// War state (event-driven; no combat). War begins only when Opinion <
	// warOpinionThreshold AND Provocations >= warProvocationThreshold.
	AtWar               bool
	Provocations        int // count of crossed provocations (raids on their routes + embargoes)
	RaidedRoutes        int // times the player raided this civ's trade route
	Embargoes           int // times the player embargoed this civ
	LastProvocationTick int // tick of the most recent provocation (drives the wait-out timer)
}

// LentWorkerBatch records one outstanding worker loan from a civilization.
type LentWorkerBatch struct {
	FactionKey string `json:"faction_key"`
	Count      int    `json:"count"`
	ReturnTick int    `json:"return_tick"` // tick at which workers leave (ignored if Permanent)
	Permanent  bool   `json:"permanent"`   // true → workers never leave
}

// LendRequest is a queued instruction for the engine to add lent workers to the
// pool and log the accompanying flavour.
type LendRequest struct {
	FactionKey string
	Count      int
	Message    string
}

// RaidRequest is a queued instruction for the engine to apply a war raid
// (resource loss) and log the accompanying flavour.
type RaidRequest struct {
	FactionKey string
	Resource   string
	Amount     float64
	Message    string
}

// NewDiplomacyManager creates a new diplomacy manager
func NewDiplomacyManager() *DiplomacyManager {
	return &DiplomacyManager{
		factions: make(map[string]*FactionState),
	}
}

// AddPassiveOpinion distributes a total opinion-per-tick amount across all
// discovered factions that are NOT rival/embargo/at-war (i.e.
// neutral/friendly/allied — coherent with embassy diplomacy raising standing
// with non-hostile powers). Fractional amounts accumulate per faction and spill
// into the integer Opinion once they cross 1.0. Opinion is capped at +100.
// Returns the number of factions that received opinion (0 if none eligible).
func (dm *DiplomacyManager) AddPassiveOpinion(totalPerTick float64) int {
	if totalPerTick <= 0 {
		return 0
	}
	// Count eligible factions first.
	eligible := 0
	for _, fs := range dm.factions {
		if dm.passiveEligible(fs) {
			eligible++
		}
	}
	if eligible == 0 {
		return 0
	}
	per := totalPerTick / float64(eligible)
	for _, fs := range dm.factions {
		if !dm.passiveEligible(fs) {
			continue
		}
		if fs.Opinion >= opinionMax {
			fs.OpinionAccum = 0
			continue
		}
		fs.OpinionAccum += per
		if fs.OpinionAccum >= 1.0 {
			whole := int(fs.OpinionAccum)
			fs.Opinion += whole
			fs.OpinionAccum -= float64(whole)
			if fs.Opinion > opinionMax {
				fs.Opinion = opinionMax
				fs.OpinionAccum = 0
			}
		}
	}
	return eligible
}

// passiveEligible reports whether a faction can receive embassy passive opinion:
// discovered, not hostile (rival/embargo), and not at war.
func (dm *DiplomacyManager) passiveEligible(fs *FactionState) bool {
	return fs.Discovered && fs.Status != "rival" && fs.Status != "embargo" && !fs.AtWar
}

// DiscoverFactions auto-discovers civilizations when reaching their MinAge.
// Returns the keys of newly-discovered civs so the caller can fire first-contact
// flavour events. New civs are seeded at neutral opinion.
func (dm *DiplomacyManager) DiscoverFactions(age string, ageOrder map[string]int) []string {
	var discovered []string
	for _, def := range config.BaseFactions() {
		if _, exists := dm.factions[def.Key]; exists {
			continue
		}
		if ageOrder[age] >= ageOrder[def.MinAge] {
			dm.factions[def.Key] = &FactionState{
				Discovered: true,
				Opinion:    0,
				Status:     "neutral",
			}
			discovered = append(discovered, def.Key)
		}
	}
	return discovered
}

// clampOpinion keeps a faction's integer opinion within [-100, 100].
func clampOpinion(fs *FactionState) {
	if fs.Opinion > opinionMax {
		fs.Opinion = opinionMax
	} else if fs.Opinion < opinionMin {
		fs.Opinion = opinionMin
	}
}

// applyPersonalityDrift nudges each discovered civ's opinion toward the
// direction implied by its personality and the player's recent trade activity.
// Called on the driftInterval cadence. Drift is small (±1) and clamped.
//   - aggressive:    -1 (trends hostile)
//   - peaceful:      +1 (trends friendly)
//   - mercantile:    +1 if the player traded recently this window, else 0
//   - isolationist:  pulls gently toward 0 (neutral)
//
// Civs at war never drift upward (they only worsen via raids/war logic).
func (dm *DiplomacyManager) applyPersonalityDrift(tradedRecently bool) {
	defs := config.FactionByKey()
	for key, fs := range dm.factions {
		if !fs.Discovered {
			continue
		}
		def, ok := defs[key]
		if !ok {
			continue
		}
		switch def.Personality {
		case "aggressive":
			fs.Opinion--
		case "peaceful":
			if !fs.AtWar {
				fs.Opinion++
			}
		case "mercantile":
			if tradedRecently && !fs.AtWar {
				fs.Opinion++
			} else if !tradedRecently && fs.Opinion > 0 {
				// Mercantile civs cool off when ignored.
				fs.Opinion--
			}
		case "isolationist":
			// Drift gently toward neutral from either direction.
			if fs.Opinion > 0 {
				fs.Opinion--
			} else if fs.Opinion < 0 {
				fs.Opinion++
			}
		}
		clampOpinion(fs)
	}
}

// SetStatus changes diplomatic status with a faction
func (dm *DiplomacyManager) SetStatus(factionKey, status string, gold float64) (float64, error) {
	defs := config.FactionByKey()
	def, ok := defs[factionKey]
	if !ok {
		return 0, fmt.Errorf("unknown faction: %s", factionKey)
	}

	fs, ok := dm.factions[factionKey]
	if !ok || !fs.Discovered {
		return 0, fmt.Errorf("%s has not been discovered yet", def.Name)
	}

	var cost float64
	switch status {
	case "allied":
		if fs.Opinion < 50 {
			return 0, fmt.Errorf("need opinion >= 50 to ally with %s (current: %d)", def.Name, fs.Opinion)
		}
		cost = 500
	case "rival":
		cost = 0
	case "embargo":
		cost = 0
	case "neutral":
		cost = 0
	default:
		return 0, fmt.Errorf("invalid diplomatic status: %s (valid: allied, rival, embargo, neutral)", status)
	}

	if gold < cost {
		return 0, fmt.Errorf("not enough gold (have: %.0f, need: %.0f)", gold, cost)
	}

	// Embargo is a provocation: track it and possibly trip a war if standing is
	// already deeply hostile.
	if status == "embargo" && fs.Status != "embargo" {
		dm.recordProvocation(fs, def, 0)
	}

	fs.Status = status
	return cost, nil
}

// recordProvocation increments a civ's provocation counters and starts a war if
// the war conditions are met. `kind` selects the counter: 0 = embargo, 1 =
// raided route. tick is the current tick (0 is fine — only used for the wait-out
// timer). Returns true if this provocation started a war.
func (dm *DiplomacyManager) recordProvocation(fs *FactionState, def config.FactionDef, kind int) bool {
	switch kind {
	case 1:
		fs.RaidedRoutes++
	default:
		fs.Embargoes++
	}
	fs.Provocations++
	if !fs.AtWar && fs.Opinion < warOpinionThreshold && fs.Provocations >= warProvocationThreshold {
		fs.AtWar = true
		return true
	}
	return false
}

// RaidTradeRoute records that the player raided this civ's trade route — a
// provocation that can trigger war if standing is already hostile. Returns
// (warStarted, error). Used by the diplomacy command.
func (dm *DiplomacyManager) RaidTradeRoute(factionKey string, tick int) (bool, error) {
	defs := config.FactionByKey()
	def, ok := defs[factionKey]
	if !ok {
		return false, fmt.Errorf("unknown civilization: %s", factionKey)
	}
	fs, ok := dm.factions[factionKey]
	if !ok || !fs.Discovered {
		return false, fmt.Errorf("%s has not been discovered yet", def.Name)
	}
	// Raiding tanks opinion immediately, then registers the provocation.
	fs.Opinion -= 20
	clampOpinion(fs)
	fs.LastProvocationTick = tick
	started := dm.recordProvocation(fs, def, 1)
	return started, nil
}

// SendTribute is the player's peace action: pay gold + culture to a civ at war
// to end the war immediately and restore a small amount of opinion. Returns the
// (gold, culture) actually spent and an error if conditions aren't met.
func (dm *DiplomacyManager) SendTribute(factionKey string, gold, culture float64) (float64, float64, error) {
	defs := config.FactionByKey()
	def, ok := defs[factionKey]
	if !ok {
		return 0, 0, fmt.Errorf("unknown civilization: %s", factionKey)
	}
	fs, ok := dm.factions[factionKey]
	if !ok || !fs.Discovered {
		return 0, 0, fmt.Errorf("%s has not been discovered yet", def.Name)
	}
	if !fs.AtWar {
		return 0, 0, fmt.Errorf("%s is not at war with you", def.Name)
	}
	// Tribute cost scales with the civ's strength.
	goldCost := 300.0 * float64(def.Strength)
	cultureCost := 50.0 * float64(def.Strength)
	if gold < goldCost {
		return 0, 0, fmt.Errorf("not enough gold for tribute to %s (have: %.0f, need: %.0f)", def.Name, gold, goldCost)
	}
	if culture < cultureCost {
		return 0, 0, fmt.Errorf("not enough culture for tribute to %s (have: %.0f, need: %.0f)", def.Name, culture, cultureCost)
	}
	// Peace: end the war, reset provocations, nudge opinion up to a wary truce.
	dm.endWar(fs)
	fs.Opinion += 25
	if fs.Opinion > 0 {
		fs.Opinion = 0 // a truce is wary neutrality at best, never instant friendship
	}
	clampOpinion(fs)
	return goldCost, cultureCost, nil
}

// endWar clears war + provocation state for a civ and drops any hostile status
// back to neutral so the wait-out / tribute paths produce a clean truce.
func (dm *DiplomacyManager) endWar(fs *FactionState) {
	fs.AtWar = false
	fs.Provocations = 0
	fs.RaidedRoutes = 0
	fs.Embargoes = 0
	if fs.Status == "embargo" || fs.Status == "rival" {
		fs.Status = "neutral"
	}
}

// SendGift sends a gift to a faction, increasing opinion
func (dm *DiplomacyManager) SendGift(factionKey string, gold float64) (float64, error) {
	defs := config.FactionByKey()
	def, ok := defs[factionKey]
	if !ok {
		return 0, fmt.Errorf("unknown faction: %s", factionKey)
	}

	fs, ok := dm.factions[factionKey]
	if !ok || !fs.Discovered {
		return 0, fmt.Errorf("%s has not been discovered yet", def.Name)
	}

	cost := 200.0
	if gold < cost {
		return 0, fmt.Errorf("not enough gold to send gift (have: %.0f, need: %.0f)", gold, cost)
	}

	fs.Opinion += 15
	if fs.Opinion > opinionMax {
		fs.Opinion = opinionMax
	}

	// Auto-upgrade to friendly if opinion hits 25+
	if fs.Status == "neutral" && fs.Opinion >= 25 {
		fs.Status = "friendly"
	}

	return cost, nil
}

// GetTradeBonus returns the sum of bonuses from allied factions for a resource.
// Civs at war never grant a bonus regardless of stored status.
func (dm *DiplomacyManager) GetTradeBonus(resourceKey string) float64 {
	defs := config.FactionByKey()
	bonus := 0.0
	for key, fs := range dm.factions {
		if fs.Status != "allied" || fs.AtWar {
			continue
		}
		def, ok := defs[key]
		if !ok {
			continue
		}
		if def.Specialty == resourceKey {
			bonus += def.TradeBonus
		}
	}
	return bonus
}

// DisruptedResources returns the set of specialty resources currently under
// trade disruption: a resource is disrupted if any discovered civ that
// specialises in it is AtWar with the player OR has been put under embargo.
// The trade system uses this to block income on routes that import a disrupted
// resource (see TradeManager.Tick). Reuses the existing war/embargo state — no
// parallel disruption flag. Must be called under the engine write lock.
//
// The bool value is true when the cause is an active war (harsher framing in the
// log) and false when it is "only" an embargo; callers may ignore it.
func (dm *DiplomacyManager) DisruptedResources() map[string]bool {
	out := make(map[string]bool)
	defs := config.FactionByKey()
	for key, fs := range dm.factions {
		if !fs.Discovered {
			continue
		}
		if !fs.AtWar && fs.Status != "embargo" {
			continue
		}
		def, ok := defs[key]
		if !ok || def.Specialty == "" {
			continue
		}
		out[def.Specialty] = true
	}
	return out
}

// Tick processes diplomacy each game tick. It discovers new civs (returning
// first-contact messages), applies personality drift and natural decay, runs
// the worker-lending lifecycle, drives war raids, and auto-ends stale wars.
//
// tradedRecently reflects whether the player completed a trade cycle in the
// recent window (drives mercantile drift). Side effects on other systems (worker
// pool, resource losses) are queued and drained by the engine via TakePending*.
func (dm *DiplomacyManager) Tick(age string, ageOrder map[string]int, tick int, tradedRecently bool) []string {
	var messages []string

	// Discover new civs and announce first contact.
	discovered := dm.DiscoverFactions(age, ageOrder)
	defs := config.FactionByKey()
	for _, key := range discovered {
		def := defs[key]
		messages = append(messages, firstContactMessage(def))
	}

	// Personality drift on the slow cadence.
	if tick%driftInterval == 0 {
		dm.applyPersonalityDrift(tradedRecently)
	}

	// Status-driven decay + natural drift toward 0.
	for _, fs := range dm.factions {
		if !fs.Discovered {
			continue
		}
		// Rival/embargo: -5 per 50 ticks.
		if tick%50 == 0 && (fs.Status == "rival" || fs.Status == "embargo") {
			fs.Opinion -= 5
			clampOpinion(fs)
		}
		// Natural drift toward 0 every 100 ticks (does not override war hostility).
		if tick%100 == 0 && !fs.AtWar {
			if fs.Opinion > 0 {
				fs.Opinion--
			} else if fs.Opinion < 0 {
				fs.Opinion++
			}
		}
	}

	// Worker-lending lifecycle: return due batches, then maybe lend new ones.
	messages = append(messages, dm.processLending(tick)...)

	// War: raids + auto-end (wait-them-out) timer.
	messages = append(messages, dm.processWar(tick)...)

	return messages
}

// processLending returns any lent batches whose ReturnTick has passed (queuing
// the worker removal), then rolls a small chance for high-opinion peaceful civs
// to lend new workers. Returns log messages for both directions.
func (dm *DiplomacyManager) processLending(tick int) []string {
	var messages []string
	defs := config.FactionByKey()

	// Return expired (non-permanent) batches.
	kept := dm.lentBatches[:0]
	for _, b := range dm.lentBatches {
		if !b.Permanent && tick >= b.ReturnTick {
			dm.pendingReturns = append(dm.pendingReturns, b.Count)
			name := b.FactionKey
			if d, ok := defs[b.FactionKey]; ok {
				name = d.Name
			}
			messages = append(messages, fmt.Sprintf("%d lent workers returned home to the %s.", b.Count, name))
			continue
		}
		kept = append(kept, b)
	}
	dm.lentBatches = kept

	// Roll a new lend: only on a slow cadence, and only for peaceful, friendly+
	// civs with healthy opinion that aren't at war.
	if tick%driftInterval != 0 {
		return messages
	}
	for key, fs := range dm.factions {
		if !fs.Discovered || fs.AtWar {
			continue
		}
		def, ok := defs[key]
		if !ok || def.Personality != "peaceful" || fs.Opinion < 40 {
			continue
		}
		// Don't stack loans from the same civ.
		if dm.hasLentBatch(key) {
			continue
		}
		// ~12% chance per eligible window.
		if rand.Float64() > 0.12 {
			continue
		}
		count := 3 + rand.Intn(4) // 3..6 workers
		permanent := fs.Opinion > lendPermanentOpinion
		batch := LentWorkerBatch{
			FactionKey: key,
			Count:      count,
			ReturnTick: tick + lendDurationTicks,
			Permanent:  permanent,
		}
		dm.lentBatches = append(dm.lentBatches, batch)
		msg := lendMessage(def, count, permanent)
		dm.pendingLends = append(dm.pendingLends, LendRequest{FactionKey: key, Count: count, Message: msg})
		messages = append(messages, msg)
	}
	return messages
}

// hasLentBatch reports whether the given civ already has an outstanding loan.
func (dm *DiplomacyManager) hasLentBatch(factionKey string) bool {
	for _, b := range dm.lentBatches {
		if b.FactionKey == factionKey {
			return true
		}
	}
	return false
}

// processWar fires periodic raids from civs at war (scaled to their Strength)
// and auto-ends wars that have gone warCooldownTicks without a fresh
// provocation. Returns log messages; resource losses are queued for the engine.
func (dm *DiplomacyManager) processWar(tick int) []string {
	var messages []string
	defs := config.FactionByKey()
	for key, fs := range dm.factions {
		if !fs.AtWar {
			continue
		}
		def, ok := defs[key]
		if !ok {
			continue
		}
		// Wait-them-out: peace after a provocation-free cooldown.
		if tick-fs.LastProvocationTick >= warCooldownTicks {
			dm.endWar(fs)
			messages = append(messages, fmt.Sprintf("The war with the %s has burned out — an uneasy peace settles.", def.Name))
			continue
		}
		// Raid every 40 ticks while at war. Severity scales with Strength.
		if tick%40 == 0 {
			res := def.Specialty
			if res == "" {
				res = "gold"
			}
			amount := 50.0 * float64(def.Strength)
			msg := raidMessage(def, amount, res)
			dm.pendingRaids = append(dm.pendingRaids, RaidRequest{
				FactionKey: key, Resource: res, Amount: amount, Message: msg,
			})
			messages = append(messages, msg)
		}
	}
	return messages
}

// TakePendingLends drains queued worker-lend requests for the engine to apply.
func (dm *DiplomacyManager) TakePendingLends() []LendRequest {
	out := dm.pendingLends
	dm.pendingLends = nil
	return out
}

// TakePendingReturns drains queued lent-worker return counts (workers to remove
// from the pool) for the engine to apply.
func (dm *DiplomacyManager) TakePendingReturns() []int {
	out := dm.pendingReturns
	dm.pendingReturns = nil
	return out
}

// TakePendingRaids drains queued war-raid resource losses for the engine.
func (dm *DiplomacyManager) TakePendingRaids() []RaidRequest {
	out := dm.pendingRaids
	dm.pendingRaids = nil
	return out
}

// RecordTrade is called by TradeManager on each completed trade cycle to
// increment opinion with all discovered factions (+1 per cycle). This
// provides a passive path to friendly status without spending gold on gifts.
// Civs at war don't warm to you through trade.
func (dm *DiplomacyManager) RecordTrade() {
	for _, fs := range dm.factions {
		if !fs.Discovered || fs.AtWar {
			continue
		}
		fs.TradeCount++
		fs.Opinion++
		if fs.Opinion > opinionMax {
			fs.Opinion = opinionMax
		}
	}
}

// Snapshot returns diplomacy state for UI
func (dm *DiplomacyManager) Snapshot(age string, ageOrder map[string]int) DiplomacyState {
	factions := make(map[string]FactionInfo)

	for _, def := range config.BaseFactions() {
		fs, exists := dm.factions[def.Key]
		info := FactionInfo{
			Name:        def.Name,
			Specialty:   def.Specialty,
			TradeBonus:  def.TradeBonus,
			Personality: def.Personality,
			Backstory:   def.Backstory,
		}
		if exists && fs.Discovered {
			info.Discovered = true
			info.Opinion = fs.Opinion
			info.Status = fs.Status
			info.TradeCount = fs.TradeCount
			info.AtWar = fs.AtWar
		} else if ageOrder[age] >= ageOrder[def.MinAge] {
			// Should be discovered but isn't yet (will be next tick).
			info.Discovered = false
		}
		// Annotate lent-worker status for the overlay.
		for _, b := range dm.lentBatches {
			if b.FactionKey != def.Key {
				continue
			}
			info.LentWorkers += b.Count
			info.LentPerm = info.LentPerm || b.Permanent
			if !b.Permanent {
				info.LentReturn = b.ReturnTick // raw tick; overlay shows presence, not countdown math
			}
		}
		factions[def.Key] = info
	}

	return DiplomacyState{
		Factions: factions,
	}
}

// LoadState restores diplomacy state from save.
func (dm *DiplomacyManager) LoadState(factions map[string]FactionStateSave, lent []LentWorkerBatch) {
	for k, v := range factions {
		dm.factions[k] = &FactionState{
			Discovered:          v.Discovered,
			Opinion:             v.Opinion,
			Status:              v.Status,
			TradeCount:          v.TradeCount,
			AtWar:               v.AtWar,
			Provocations:        v.Provocations,
			RaidedRoutes:        v.RaidedRoutes,
			Embargoes:           v.Embargoes,
			LastProvocationTick: v.LastProvocationTick,
		}
	}
	if lent != nil {
		dm.lentBatches = append([]LentWorkerBatch(nil), lent...)
	}
}

// LentWorkerTotal returns the total number of workers currently on loan across
// all civs. Used by the engine to reconcile the pool after a load.
func (dm *DiplomacyManager) LentWorkerTotal() int {
	total := 0
	for _, b := range dm.lentBatches {
		total += b.Count
	}
	return total
}

// FactionStateSave is the serializable form of FactionState
type FactionStateSave struct {
	Discovered          bool   `json:"discovered"`
	Opinion             int    `json:"opinion"`
	Status              string `json:"status"`
	TradeCount          int    `json:"trade_count"`
	AtWar               bool   `json:"at_war"`
	Provocations        int    `json:"provocations"`
	RaidedRoutes        int    `json:"raided_routes"`
	Embargoes           int    `json:"embargoes"`
	LastProvocationTick int    `json:"last_provocation_tick"`
}

// GetFactionsForSave returns faction states for serialization
func (dm *DiplomacyManager) GetFactionsForSave() map[string]FactionStateSave {
	out := make(map[string]FactionStateSave, len(dm.factions))
	for k, fs := range dm.factions {
		out[k] = FactionStateSave{
			Discovered:          fs.Discovered,
			Opinion:             fs.Opinion,
			Status:              fs.Status,
			TradeCount:          fs.TradeCount,
			AtWar:               fs.AtWar,
			Provocations:        fs.Provocations,
			RaidedRoutes:        fs.RaidedRoutes,
			Embargoes:           fs.Embargoes,
			LastProvocationTick: fs.LastProvocationTick,
		}
	}
	return out
}

// GetLentBatchesForSave returns outstanding worker loans for serialization.
func (dm *DiplomacyManager) GetLentBatchesForSave() []LentWorkerBatch {
	return append([]LentWorkerBatch(nil), dm.lentBatches...)
}

// firstContactMessage builds the flavour line announced when a civ is first
// discovered. It introduces the name, personality, and backstory.
func firstContactMessage(def config.FactionDef) string {
	return fmt.Sprintf("[gold]✦ First contact: %s[-] [gray](%s)[-] — %s",
		def.Name, def.Personality, def.Backstory)
}

// lendMessage builds the flavour line for a worker loan, including a backstory
// snippet and whether the loan is permanent.
func lendMessage(def config.FactionDef, count int, permanent bool) string {
	tail := fmt.Sprintf("(returning in %d ticks)", lendDurationTicks)
	if permanent {
		tail = "(they choose to stay — permanent!)"
	}
	return fmt.Sprintf("[green]+%d workers from the %s[-] — %s %s",
		count, def.Name, def.Backstory, tail)
}

// raidMessage builds the flavour line for a war raid resource loss.
func raidMessage(def config.FactionDef, amount float64, resource string) string {
	return fmt.Sprintf("[red]⚔ The %s raid you — lost %.0f %s.[-] %s",
		def.Name, amount, resource, def.Backstory)
}
