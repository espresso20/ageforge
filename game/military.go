package game

import (
	"fmt"
	"maps"
	"math/rand"

	"github.com/espresso20/ageforge/rules"
)

// Expedition categories. Every mission costs resources, but only a military
// grand mission costs soldiers and counts the army: that one is
// ExpeditionMilitary, and short, long and civil grand missions are
// ExpeditionScouting. The category picks the techs' terms on a mission and the
// odds of an encounter when it comes back.
const (
	ExpeditionScouting = "scouting"
	ExpeditionMilitary = "military"
)

// failedLootShare is the share of an expedition's or campaign's loot that
// comes back when it fails.
const failedLootShare = 0.3

// ActiveExpedition is a mission under way: a copy of the offer it was sent
// from, what it cost, and how long is left.
type ActiveExpedition struct {
	MissionOffer
	Soldiers       int                `json:"soldiers,omitempty"`
	LaunchTick     int                `json:"launch_tick,omitempty"`
	TicksLeft      int                `json:"ticks_left"`
	BaseRewards    map[string]float64 `json:"base_rewards,omitempty"`
	SuccessRewards map[string]float64 `json:"success_rewards,omitempty"`
	FailureRewards map[string]float64 `json:"failure_rewards,omitempty"`
	Cost           map[string]float64 `json:"cost,omitempty"`
}

// MilitaryManager handles the mission board, the missions under way and the
// defense rating.
//
// One mission per TIER can be under way at a time: a short, a long and a
// grand mission may run together (active is keyed by TierShort, TierLong and
// TierGrand). A military grand mission spends soldiers, a real resource
// (config/resources.go); the engine validates and deducts every cost before
// LaunchOffer is called. Soldiers are spent at launch; half come home on a
// success.
type MilitaryManager struct {
	rules          *rules.Set
	active         map[string]*ActiveExpedition
	board          MissionBoard
	completedCount int
	totalLoot      map[string]float64
	defenseRating  float64
	// scoutTime is the techs' term on the time a short, long or civil
	// mission takes (config.MechanicExpeditionTicks). The engine sets it
	// (SetScoutTime); 0 reads as 1. Military grand missions keep their time.
	scoutTime float64
	// campaignTime and campaignPay are the techs' terms on the time a
	// military grand mission takes and on what it brings back
	// (config.MechanicCampaignTicks, config.MechanicCampaignReward). The
	// engine sets them (SetCampaignTerms); 0 reads as 1.
	campaignTime, campaignPay float64
}

// SetScoutTime sets the techs' term on a short, long or civil mission's time.
func (mm *MilitaryManager) SetScoutTime(f float64) { mm.scoutTime = f }

// SetCampaignTerms sets the techs' terms on a military grand mission's time
// and on what it brings back.
func (mm *MilitaryManager) SetCampaignTerms(time, pay float64) {
	mm.campaignTime, mm.campaignPay = time, pay
}

// missionTicks is ticks with the techs' cut for a category.
func (mm *MilitaryManager) missionTicks(category string, ticks int) int {
	if category == ExpeditionScouting {
		return techTimeTicks(ticks, mm.scoutTime)
	}
	return techTimeTicks(ticks, mm.campaignTime)
}

// CampaignPay is amount of a campaign's loot with the techs' share on top:
// amount itself with none.
func (mm *MilitaryManager) CampaignPay(amount float64) float64 {
	if mm.campaignPay <= 1 {
		return amount
	}
	return float64(amount * mm.campaignPay)
}

// NewMilitaryManager creates a military manager on the core ruleset.
func NewMilitaryManager() *MilitaryManager { return NewMilitaryManagerWith(rules.Core()) }

// Rebind moves the manager onto set, whose clocks time its missions.
func (mm *MilitaryManager) Rebind(set *rules.Set) { mm.rules = set }

// NewMilitaryManagerWith creates a military manager on set.
func NewMilitaryManagerWith(set *rules.Set) *MilitaryManager {
	return &MilitaryManager{
		rules:     set,
		totalLoot: make(map[string]float64),
		active:    make(map[string]*ActiveExpedition),
	}
}

// LaunchOffer puts a mission under way in its tier's slot. The offer's
// SuccessChance is the chance the mission is sent with (the engine has
// already counted the army). Resource costs are validated and deducted by
// the engine BEFORE this is called; this method only enforces the
// one-mission-per-tier rule.
func (mm *MilitaryManager) LaunchOffer(offer MissionOffer, tick, soldiers int, cost, baseHaul map[string]float64) error {
	if existing := mm.active[offer.Tier]; existing != nil {
		return fmt.Errorf("%s is already under way (%s left). Wait for it to come back", tierNoun(offer.Tier), approxTicks(existing.TicksLeft, BaseTickInterval))
	}
	ticks := mm.missionTicks(offer.Category(), minutesToTicks(offer.LengthMinutes))
	if ticks < 1 {
		ticks = 1
	}
	mm.active[offer.Tier] = &ActiveExpedition{
		MissionOffer: offer,
		Soldiers:     soldiers,
		LaunchTick:   tick,
		TicksLeft:    ticks,
		BaseRewards:  maps.Clone(baseHaul),
		Cost:         maps.Clone(cost),
	}
	return nil
}

// tierNoun is the player-facing noun phrase for one mission of a tier,
// sentence-initial.
func tierNoun(tier string) string {
	switch tier {
	case TierLong:
		return "A long mission"
	case TierGrand:
		return "A grand mission"
	}
	return "A short mission"
}

// ActiveByTier returns the mission under way in a tier's slot, or nil. The
// returned pointer is the live manager state.
func (mm *MilitaryManager) ActiveByTier(tier string) *ActiveExpedition {
	return mm.active[tier]
}

// HasActive reports whether any mission is currently under way.
func (mm *MilitaryManager) HasActive() bool {
	for _, tier := range boardTiers {
		if mm.active[tier] != nil {
			return true
		}
	}
	return false
}

// ExpeditionResult holds the rewards + player-facing message for one mission
// that resolved this tick. Category and Success describe HOW it resolved so the
// engine can weight a faction encounter on resolution (see
// GameEngine.rollExpeditionEncounter).
//
// Name identifies WHICH mission resolved. The varied flavour line is generated
// ENGINE-side (MilitaryManager has no rng of its own for prose; ge.rng does,
// and the prose stream must come off the same seeded source as everything
// else) — see GameEngine.expeditionFlavorLine. Message stays the mechanical
// line; flavour rides alongside it, never in place of it.
type ExpeditionResult struct {
	Rewards      map[string]float64
	Message      string
	Name         string // the offer's title
	Category     string // ExpeditionScouting or ExpeditionMilitary
	Success      bool   // true if the mission succeeded
	SoldiersSent int
	// Offer is the offer the mission was sent from: its tier, age and
	// Legendary mark are what the badges read.
	Offer MissionOffer
}

// Tick advances every mission under way (one per tier) by one tick and
// returns a result for each that resolved this tick, in tier order. Tiers
// tick down and resolve independently.
//
// The success roll is the mission's own SuccessChance (rolled when the offer
// was made; the army was counted when it was sent). expeditionBonus is not
// used here: the loot was sized at launch. Soldiers are spent at launch; a
// failure differs from a success only in the loot (30%). The roll comes from
// rng, one draw per resolving mission, short before long before grand.
func (mm *MilitaryManager) Tick(rng *rand.Rand, missionPower, expeditionBonus float64) []ExpeditionResult {
	var results []ExpeditionResult
	for _, tier := range boardTiers {
		if res, ok := mm.tickTier(rng, tier); ok {
			results = append(results, res)
		}
	}
	return results
}

// tickTier advances one tier's mission. ok is true only on the tick it
// resolves.
func (mm *MilitaryManager) tickTier(rng *rand.Rand, tier string) (ExpeditionResult, bool) {
	active := mm.active[tier]
	if active == nil {
		return ExpeditionResult{}, false
	}
	active.TicksLeft--
	if active.TicksLeft > 0 {
		return ExpeditionResult{}, false
	}

	success := rng.Float64() < active.SuccessChance
	rewards := make(map[string]float64)
	var message string
	if success {
		for res, amount := range active.SuccessRewards {
			rewards[res] = amount
			mm.totalLoot[res] += amount
		}
		message = fmt.Sprintf("%s succeeded. Loot: %s.", active.Title, amountsText(rewards))
	} else {
		for res, amount := range active.FailureRewards {
			rewards[res] = amount
			mm.totalLoot[res] += amount
		}
		message = fmt.Sprintf("%s failed. You kept %.0f%% of the loot: %s.", active.Title, failedLootShare*100, amountsText(rewards))
	}

	mm.completedCount++
	mm.active[tier] = nil
	return ExpeditionResult{
		Rewards:      rewards,
		Message:      message,
		Name:         active.Title,
		Category:     active.Category(),
		Success:      success,
		SoldiersSent: active.Soldiers,
		Offer:        active.MissionOffer,
	}, true
}

// CalculateDefense calculates defense rating from soldiers and bonuses
func (mm *MilitaryManager) CalculateDefense(soldierCount int, militaryBonus float64) float64 {
	base := float64(soldierCount) * 2.0
	return base * (1.0 + militaryBonus)
}

// Snapshot returns military state for UI.
//
// soldierCount is the current soldiers resource amount; soldierCap and
// soldierRate are that resource's storage cap and per-tick production rate.
// The engine fills the board's offers (it knows the costs and the stores) and
// the automatic dispatch view.
func (mm *MilitaryManager) Snapshot(soldierCount, soldierCap int, soldierRate, militaryBonus, expeditionBonus float64) MilitaryState {
	var activeShort, activeLong, activeGrand *ExpeditionSnapshot
	snap := func(tier string) *ExpeditionSnapshot {
		return snapshotActive(mm.active[tier], expeditionBonus)
	}
	activeShort, activeLong, activeGrand = snap(TierShort), snap(TierLong), snap(TierGrand)

	// ActiveScout and ActiveMilitary are the two slots the older panels read:
	// the short (else the long) mission, and the grand one.
	activeScout := activeShort
	if activeScout == nil {
		activeScout = activeLong
	}

	loot := make(map[string]float64)
	for k, v := range mm.totalLoot {
		loot[k] = v
	}

	return MilitaryState{
		SoldierCount:    soldierCount,
		SoldierCap:      soldierCap,
		SoldierRate:     soldierRate,
		DefenseRating:   mm.CalculateDefense(soldierCount, militaryBonus),
		MilitaryBonus:   militaryBonus,
		ExpeditionBonus: expeditionBonus,
		ActiveShort:     activeShort,
		ActiveLong:      activeLong,
		ActiveGrand:     activeGrand,
		ActiveScout:     activeScout,
		ActiveMilitary:  activeGrand,
		CompletedCount:  mm.completedCount,
		TotalLoot:       loot,
	}
}

// snapshotActive converts a live ActiveExpedition into a UI snapshot, or nil.
func snapshotActive(active *ActiveExpedition, expeditionBonus float64) *ExpeditionSnapshot {
	if active == nil {
		return nil
	}
	success := maps.Clone(active.SuccessRewards)
	failure := maps.Clone(active.FailureRewards)
	if success == nil {
		success = scaleLoot(active.BaseRewards, 1+expeditionBonus)
	}
	if failure == nil {
		failure = scaleLoot(active.BaseRewards, failedLootShare)
	}
	return &ExpeditionSnapshot{
		Name:          active.Title,
		Tier:          active.Tier,
		Grand:         active.Grand,
		Rare:          active.Rare,
		Soldiers:      active.Soldiers,
		TicksLeft:     active.TicksLeft,
		LengthMinutes: active.LengthMinutes,
		SuccessChance: active.SuccessChance,
		HaulMinutes:   active.HaulMinutes,
		SuccessHaul:   success,
		FailureHaul:   failure,
	}
}

func scaleLoot(in map[string]float64, scale float64) map[string]float64 {
	out := make(map[string]float64, len(in))
	for key, value := range in {
		out[key] = value * scale
	}
	return out
}

// LoadState restores military state from save: the missions under way, the
// board and its counters, and the totals.
func (mm *MilitaryManager) LoadState(save MilitarySave) {
	mm.active = make(map[string]*ActiveExpedition)
	for _, entry := range save.Active {
		if entry.Mission == nil {
			continue
		}
		mission := *entry.Mission
		mm.active[entry.Slot] = &mission
	}
	mm.board = MissionBoard{}
	if save.Board != nil {
		mm.board = save.Board.clone()
	}
	mm.completedCount = save.CompletedCount
	if save.TotalLoot != nil {
		mm.totalLoot = save.TotalLoot
	}
}

// ActiveForSave returns deep copies of the missions under way, as a list
// keyed by slot (the tier), in tier order.
func (mm *MilitaryManager) ActiveForSave() []ActiveSlotSave {
	var out []ActiveSlotSave
	for _, tier := range boardTiers {
		if active := mm.active[tier]; active != nil {
			c := *active
			c.BaseRewards, c.SuccessRewards = maps.Clone(active.BaseRewards), maps.Clone(active.SuccessRewards)
			c.FailureRewards, c.Cost = maps.Clone(active.FailureRewards), maps.Clone(active.Cost)
			out = append(out, ActiveSlotSave{Slot: tier, Mission: &c})
		}
	}
	return out
}

// BoardForSave returns a deep copy of the board, or nil before one is made.
func (mm *MilitaryManager) BoardForSave() *MissionBoard {
	if mm.board.Age == "" {
		return nil
	}
	c := mm.board.clone()
	return &c
}
