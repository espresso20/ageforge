package game

import (
	"fmt"
	"math/rand"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/pkg/textfmt"
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

	// Allying needs at least AllyOpinion opinion and costs AllyCost gold. A
	// gift costs GiftCost gold and adds GiftOpinion opinion. Exported so the
	// UI quotes the same numbers.
	AllyOpinion = 50
	AllyCost    = 500.0
	GiftCost    = 200.0
	GiftOpinion = 15
)

// errUnknownCiv is the refusal for a civ key the player has not met: one not
// on the roster, or one not discovered yet. Both read alike, and the typo
// suggestion looks only at civilizations already met, so a refusal never
// names or confirms a civilization the player has not found (spoilers.go).
func (dm *DiplomacyManager) errUnknownCiv(key string) error {
	msg := fmt.Sprintf("You have not met a civilization called '%s'.", key)
	met := map[string]bool{}
	for k, fs := range dm.factions {
		if fs.Discovered {
			met[k] = true
		}
	}
	if s := closestKey(key, met); s != "" && s != key {
		msg += fmt.Sprintf(" Did you mean '%s'?", s)
	}
	return fmt.Errorf("%s Scouting expeditions make first contact; type diplomacy deals to see the ones you have met.", msg)
}

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

	// notices are log lines raised outside Tick (an embargo that starts a
	// war); the next Tick returns them first. Transient, not saved.
	notices []string

	// factionList / factionDefs are the static civ roster, built once at
	// construction so the per-tick paths (Tick, GetTradeBonus via the resolver,
	// DisruptedResources) don't rebuild config tables every call. Read-only.
	factionList []config.FactionDef
	factionDefs map[string]config.FactionDef
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

	// Trade deals (deals.go): the current offers, how many times they have
	// been rolled, ticks of live play since the last roll, and the age they
	// were priced for.
	Deals     []FactionDeal
	DealRound int
	DealTicks int
	DealsFor  string
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
	list := config.BaseFactions()
	defs := make(map[string]config.FactionDef, len(list))
	for _, def := range list {
		defs[def.Key] = def
	}
	return &DiplomacyManager{
		factions:    make(map[string]*FactionState),
		factionList: list,
		factionDefs: defs,
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

// ageFallbackGap is how many ages past a faction's MinAge the player must reach
// before that faction is auto-discovered without ever meeting it on an expedition.
// It is the anti-softlock net: expeditions are the intended discovery TRIGGER (see
// GameEngine.rollExpeditionEncounter), and MinAge is only a FLOOR — but a player
// who never runs expeditions still meets every civ eventually, just much later.
const ageFallbackGap = 2

// DiscoverFactions is the generous age FALLBACK for discovery, not the primary
// path. It auto-discovers any faction the player has long been eligible to meet
// (age >= MinAge) but hasn't: the current age is at least ageFallbackGap ages past
// the faction's MinAge. The primary discovery path is expedition encounters, which
// call DiscoverFaction directly. Reaching a faction's MinAge no longer discovers it
// on its own — age is a floor. Returns newly-discovered keys so the caller can fire
// first-contact flavour. New civs are seeded at neutral opinion.
func (dm *DiplomacyManager) DiscoverFactions(age string, ageOrder map[string]int) []string {
	var discovered []string
	cur := ageOrder[age]
	for _, def := range dm.factionList {
		if _, exists := dm.factions[def.Key]; exists {
			continue
		}
		if cur >= ageOrder[def.MinAge]+ageFallbackGap {
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

// DiscoverFaction discovers a single faction by key — the entry point the encounter
// engine calls when an expedition turns one up. It seeds neutral/0 state and returns
// that civ's first-contact flavour line. Returns ("", false) when the key is unknown
// or the faction is already discovered. Must be called under the engine write lock.
func (dm *DiplomacyManager) DiscoverFaction(key string) (string, bool) {
	def, ok := dm.factionDefs[key]
	if !ok {
		return "", false
	}
	if fs, exists := dm.factions[key]; exists && fs.Discovered {
		return "", false
	}
	dm.factions[key] = &FactionState{
		Discovered: true,
		Opinion:    0,
		Status:     "neutral",
	}
	return firstContactMessage(def), true
}

// IsDiscovered reports whether a faction has been discovered. Used by the encounter
// engine to split eligible factions into discovery vs re-encounter candidates.
func (dm *DiplomacyManager) IsDiscovered(key string) bool {
	fs, ok := dm.factions[key]
	return ok && fs.Discovered
}

// StateOf returns a COPY of a faction's diplomatic state. The bool is false when
// the faction has never been seeded (undiscovered and untouched); in that case
// the returned value is a safe neutral default. Used by the encounter engine to
// shape a boon profile by standing (see factionProfile) without exposing the
// internal pointer.
func (dm *DiplomacyManager) StateOf(key string) (FactionState, bool) {
	if fs, ok := dm.factions[key]; ok && fs != nil {
		return *fs, true
	}
	return FactionState{Status: "neutral"}, false
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
	defs := dm.factionDefs
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
	defs := dm.factionDefs
	def, ok := defs[factionKey]
	if !ok {
		return 0, dm.errUnknownCiv(factionKey)
	}

	fs, ok := dm.factions[factionKey]
	if !ok || !fs.Discovered {
		return 0, dm.errUnknownCiv(factionKey)
	}

	var cost float64
	switch status {
	case "allied":
		if fs.Opinion < AllyOpinion {
			return 0, fmt.Errorf("The %s need opinion %d before they will ally (now %d).", def.Name, AllyOpinion, fs.Opinion)
		}
		cost = AllyCost
	case "rival":
		cost = 0
	case "embargo":
		cost = 0
	case "neutral":
		cost = 0
	default:
		return 0, fmt.Errorf("Unknown diplomatic status '%s'. Use allied, rival, embargo or neutral.", status)
	}

	if gold < cost {
		return 0, fmt.Errorf("Not enough gold to ally with the %s: need %s, have %s.", def.Name, textfmt.Number(cost), textfmt.Number(gold))
	}

	// Embargo is a provocation: track it and possibly trip a war if opinion is
	// already deeply hostile. A war it starts is announced on the next tick.
	if status == "embargo" && fs.Status != "embargo" {
		if dm.recordProvocation(fs, def, 0) {
			dm.notices = append(dm.notices, fmt.Sprintf("The %s declared war on you.", def.Name))
		}
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
	defs := dm.factionDefs
	def, ok := defs[factionKey]
	if !ok {
		return false, dm.errUnknownCiv(factionKey)
	}
	fs, ok := dm.factions[factionKey]
	if !ok || !fs.Discovered {
		return false, dm.errUnknownCiv(factionKey)
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
	defs := dm.factionDefs
	def, ok := defs[factionKey]
	if !ok {
		return 0, 0, dm.errUnknownCiv(factionKey)
	}
	fs, ok := dm.factions[factionKey]
	if !ok || !fs.Discovered {
		return 0, 0, dm.errUnknownCiv(factionKey)
	}
	if !fs.AtWar {
		return 0, 0, fmt.Errorf("The %s are not at war with you.", def.Name)
	}
	// Tribute cost scales with the civ's strength.
	goldCost := 300.0 * float64(def.Strength)
	cultureCost := 50.0 * float64(def.Strength)
	if gold < goldCost {
		return 0, 0, fmt.Errorf("Not enough gold for tribute to the %s: need %s, have %s.", def.Name, textfmt.Number(goldCost), textfmt.Number(gold))
	}
	if culture < cultureCost {
		return 0, 0, fmt.Errorf("Not enough culture for tribute to the %s: need %s, have %s.", def.Name, textfmt.Number(cultureCost), textfmt.Number(culture))
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
	defs := dm.factionDefs
	def, ok := defs[factionKey]
	if !ok {
		return 0, dm.errUnknownCiv(factionKey)
	}

	fs, ok := dm.factions[factionKey]
	if !ok || !fs.Discovered {
		return 0, dm.errUnknownCiv(factionKey)
	}

	cost := GiftCost
	if gold < cost {
		return 0, fmt.Errorf("Not enough gold for a gift to the %s: need %s, have %s.", def.Name, textfmt.Number(cost), textfmt.Number(gold))
	}

	fs.Opinion += GiftOpinion
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
	// Roster order, not map order: the float sum must be the same every run.
	bonus := 0.0
	for _, def := range dm.factionList {
		fs, ok := dm.factions[def.Key]
		if !ok || fs.Status != "allied" || fs.AtWar {
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
	defs := dm.factionDefs
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
//
// rng supplies every random draw (worker lending). Loops that draw from it or
// queue side effects walk the factions in roster order (factionList), never
// map order, so a seeded rng gives the same outcome on every run.
func (dm *DiplomacyManager) Tick(rng *rand.Rand, age string, ageOrder map[string]int, tick int, tradedRecently bool) []string {
	messages := dm.notices
	dm.notices = nil

	// Discover new civs and announce first contact.
	discovered := dm.DiscoverFactions(age, ageOrder)
	defs := dm.factionDefs
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
	messages = append(messages, dm.processLending(rng, tick)...)

	// War: raids + auto-end (wait-them-out) timer.
	messages = append(messages, dm.processWar(tick)...)

	return messages
}

// processLending returns any lent batches whose ReturnTick has passed (queuing
// the worker removal), then rolls a small chance for high-opinion peaceful civs
// to lend new workers. Returns log messages for both directions.
func (dm *DiplomacyManager) processLending(rng *rand.Rand, tick int) []string {
	var messages []string
	defs := dm.factionDefs

	// Return expired (non-permanent) batches.
	kept := dm.lentBatches[:0]
	for _, b := range dm.lentBatches {
		if !b.Permanent && tick >= b.ReturnTick {
			dm.pendingReturns = append(dm.pendingReturns, b.Count)
			name := b.FactionKey
			if d, ok := defs[b.FactionKey]; ok {
				name = d.Name
			}
			messages = append(messages, fmt.Sprintf("%s returned home to the %s.", textfmt.Count(b.Count, "lent worker", "lent workers"), name))
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
	for _, def := range dm.factionList {
		key := def.Key
		fs, ok := dm.factions[key]
		if !ok || !fs.Discovered || fs.AtWar {
			continue
		}
		if def.Personality != "peaceful" || fs.Opinion < 40 {
			continue
		}
		// Don't stack loans from the same civ.
		if dm.hasLentBatch(key) {
			continue
		}
		// ~12% chance per eligible window.
		if rng.Float64() > 0.12 {
			continue
		}
		count := 3 + rng.Intn(4) // 3..6 workers
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
	for _, def := range dm.factionList {
		key := def.Key
		fs, ok := dm.factions[key]
		if !ok || !fs.AtWar {
			continue
		}
		// Wait-them-out: peace after a provocation-free cooldown.
		if tick-fs.LastProvocationTick >= warCooldownTicks {
			dm.endWar(fs)
			messages = append(messages, fmt.Sprintf("The war with the %s has burned out. An uneasy peace settles.", def.Name))
			continue
		}
		// Raid every 40 ticks while at war. Severity scales with Strength.
		if tick%40 == 0 {
			res := def.Specialty
			if res == "" {
				res = "gold"
			}
			amount := 50.0 * float64(def.Strength)
			// The raid is QUEUED only, not announced here. The engine logs it when
			// it drains the queue, where it can append a generated flavour line off
			// ge.rng — see GameEngine.processDiplomacy / raidLogLine. Returning it
			// from here as well would double-log every raid.
			dm.pendingRaids = append(dm.pendingRaids, RaidRequest{
				FactionKey: key, Resource: res, Amount: amount,
				Message: raidMessage(def, amount, res),
			})
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

	for _, def := range dm.factionList {
		fs, exists := dm.factions[def.Key]
		info := FactionInfo{
			Name:        def.Name,
			Specialty:   def.Specialty,
			TradeBonus:  def.TradeBonus,
			Personality: def.Personality,
			Backstory:   def.Backstory,
			Strength:    def.Strength,
		}
		if exists && fs.Discovered {
			info.Discovered = true
			info.Opinion = fs.Opinion
			info.Status = fs.Status
			info.TradeCount = fs.TradeCount
			info.AtWar = fs.AtWar
			info.Deals = dealInfos(fs, age)
			info.DealsBlocked = dealBlocked(*fs)
			info.DealRefreshIn = max(dealRefreshTicks-fs.DealTicks, 0)
		} else if ageOrder[age] >= ageOrder[def.MinAge] {
			// Eligible (age floor met) but not yet met: discovery is triggered by
			// running expeditions, with a late age fallback (see DiscoverFactions).
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
			Deals:               append([]FactionDeal(nil), v.Deals...),
			DealRound:           v.DealRound,
			DealTicks:           v.DealTicks,
			DealsFor:            v.DealsFor,
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
	// Trade deals (deals.go); absent from saves before them, which roll a
	// fresh set on the first tick.
	Deals     []FactionDeal `json:"deals,omitempty"`
	DealRound int           `json:"deal_round,omitempty"`
	DealTicks int           `json:"deal_ticks,omitempty"`
	DealsFor  string        `json:"deals_for,omitempty"`
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
			Deals:               append([]FactionDeal(nil), fs.Deals...),
			DealRound:           fs.DealRound,
			DealTicks:           fs.DealTicks,
			DealsFor:            fs.DealsFor,
		}
	}
	return out
}

// GetLentBatchesForSave returns outstanding worker loans for serialization.
func (dm *DiplomacyManager) GetLentBatchesForSave() []LentWorkerBatch {
	return append([]LentWorkerBatch(nil), dm.lentBatches...)
}

// firstContactMessage builds the line announced when a civ is first
// discovered. It introduces the name, personality and backstory, then says
// where to find their offers.
func firstContactMessage(def config.FactionDef) string {
	return fmt.Sprintf("[gold]✦ First contact: %s[-] [gray](%s)[-]. %s Type diplomacy to see their offers.",
		def.Name, def.Personality, def.Backstory)
}

// lendMessage builds the line for a worker loan and says whether the workers
// go home. The backstory belongs to first contact only, so it is not
// repeated here. The duration is at base speed (the manager has no clock).
func lendMessage(def config.FactionDef, count int, permanent bool) string {
	tail := fmt.Sprintf("(they return in %s)", DurationText(lendDurationTicks, BaseTickInterval))
	if permanent {
		tail = "(they stay for good)"
	}
	return fmt.Sprintf("[green]+%s on loan from the %s[-] %s.",
		textfmt.Count(count, "worker", "workers"), def.Name, tail)
}

// raidMessage builds the MECHANICAL line for a war raid resource loss: who, what,
// how much. It carries no flavour of its own — the engine appends a generated
// sentence at the point it logs the raid (see GameEngine.raidLogLine), because
// that is where the seeded rng lives.
//
// It used to append def.Backstory, the civ's entire two-sentence introduction, to
// EVERY raid. Raids fire every raidInterval ticks for the whole war, so a single
// war reprinted the same paragraph a few hundred times. A backstory is a
// first-contact line (see firstContactMessage) and nowhere else.
//
// The verb is PAST tense on purpose: the roster mixes singular and plural civ
// names ("Merchant Guild" vs "Void Reavers"), and the old present-tense "The %s
// raid you" agreed with only half of them.
func raidMessage(def config.FactionDef, amount float64, resource string) string {
	return fmt.Sprintf("[red]⚔ The %s raided you: you lost %s %s.[-]",
		def.Name, amountText(amount), resourceLabel(resource))
}

// raidMissedMessage is the line for a war raid that took nothing: the stock
// held less than the raid would have carried off, and a raid takes all of its
// amount or none of it.
func raidMissedMessage(def config.FactionDef, resource string) string {
	return fmt.Sprintf("[yellow]⚔ The %s raided you but found too little %s to carry off. You lost nothing.[-]",
		def.Name, resourceLabel(resource))
}

// raidMessageDefended is raidMessage for a raid the garrison blunted: what the
// player still lost, and what the army kept (guard is the share it blunted).
func raidMessageDefended(def config.FactionDef, lost, kept float64, resource string, guard float64) string {
	return fmt.Sprintf("[red]⚔ The %s raided you: you lost %s %s.[-] [green]Your garrison kept %s %s from them (about %.0f%% of the raid).[-]",
		def.Name, amountText(lost), resourceLabel(resource), amountText(kept), resourceLabel(resource), guard*100)
}
