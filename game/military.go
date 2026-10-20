package game

import (
	"fmt"
	"maps"
	"math"
	"math/rand"
	"sort"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/rules"
)

// Expedition categories. Scouting expeditions cost only resources (no soldiers)
// and are playable before the soldiers resource exists (pre-iron_age). Military
// expeditions cost soldiers (plus any Cost).
const (
	ExpeditionScouting = "scouting"
	ExpeditionMilitary = "military"
)

// failedLootShare is the share of an expedition's or campaign's loot that
// comes back when it fails.
const failedLootShare = 0.3

// minExpeditionDurationTicks is the floor LaunchExpedition falls back to if a def
// somehow carries a non-positive duration range. No shipped def hits this — it
// exists so a malformed def can never produce a 0-tick expedition.
const minExpeditionDurationTicks = 60

// ExpeditionDef defines an available expedition
type ExpeditionDef struct {
	Name           string
	Key            string
	Era            string
	Kind           string
	Category       string // ExpeditionScouting or ExpeditionMilitary
	MinAge         string
	MaxAge         string // empty = no upper bound; expedition is unavailable once past this age
	SoldiersNeeded int
	// DurationMin/DurationMax bound the randomized active duration rolled at launch,
	// inclusive. LaunchExpedition rolls a uniform value in [DurationMin, DurationMax];
	// every def must carry a valid range (DurationMin > 0, DurationMax > DurationMin).
	DurationMin    int
	DurationMax    int
	DifficultyBase float64            // 0.0 - 1.0, higher = harder
	Cost           map[string]float64 // resource cost to launch, in addition to soldiers
	Rewards        map[string]float64
	Description    string
	MinMinutes     float64
	MaxMinutes     float64
	BaseSuccess    float64
	Yield          float64
	CostMinutes    float64
}

// ActiveExpedition represents an ongoing expedition
type ActiveExpedition struct {
	Key            string
	Name           string
	Soldiers       int
	TicksLeft      int
	LengthMinutes  float64            `json:"length_minutes,omitempty"`
	SuccessChance  float64            `json:"success_chance,omitempty"`
	HaulMinutes    float64            `json:"haul_minutes,omitempty"`
	BaseRewards    map[string]float64 `json:"base_rewards,omitempty"`
	SuccessRewards map[string]float64 `json:"success_rewards,omitempty"`
	FailureRewards map[string]float64 `json:"failure_rewards,omitempty"`
	Cost           map[string]float64 `json:"cost,omitempty"`
}

// MilitaryManager handles military expeditions and defense ratings.
//
// One expedition per CATEGORY can be active at a time: a scouting expedition and
// a military expedition may run concurrently. activeByCat is keyed by
// ExpeditionScouting / ExpeditionMilitary. Soldiers are a real resource
// (config/resources.go): launching an expedition spends SoldiersNeeded of the
// soldiers resource plus any Cost, validated and deducted by the engine before
// LaunchExpedition is called. Soldiers are spent at launch (win or lose);
// success vs failure differs only in reward, not soldier loss.
type MilitaryManager struct {
	rules          *rules.Set
	expeditions    []ExpeditionDef
	activeByCat    map[string]*ActiveExpedition
	completedCount int
	totalLoot      map[string]float64
	defenseRating  float64
	// scoutTime is the techs' term on the time a scouting expedition takes
	// (config.MechanicExpeditionTicks). The engine sets it (SetScoutTime); 0
	// reads as 1. Campaigns keep their time.
	scoutTime float64
	// campaignTime and campaignPay are the techs' terms on the time a
	// military campaign takes and on what it brings back
	// (config.MechanicCampaignTicks, config.MechanicCampaignReward). The
	// engine sets them (SetCampaignTerms); 0 reads as 1.
	campaignTime, campaignPay float64
}

// SetScoutTime sets the techs' term on a scouting expedition's time.
func (mm *MilitaryManager) SetScoutTime(f float64) { mm.scoutTime = f }

// SetCampaignTerms sets the techs' terms on a military campaign's time and
// on what it brings back.
func (mm *MilitaryManager) SetCampaignTerms(time, pay float64) {
	mm.campaignTime, mm.campaignPay = time, pay
}

// missionTicks is ticks, a mission's time already stretched for its age,
// with the techs' cut for its category: scouting expeditions have one,
// campaigns another.
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

// Rebind moves the manager onto set, whose clocks time its expeditions. The
// expeditions themselves are defined here, not in a ruleset.
func (mm *MilitaryManager) Rebind(set *rules.Set) { mm.rules = set }

// NewMilitaryManagerWith creates a military manager on set.
func NewMilitaryManagerWith(set *rules.Set) *MilitaryManager {
	return &MilitaryManager{
		rules:       set,
		totalLoot:   make(map[string]float64),
		activeByCat: make(map[string]*ActiveExpedition),
		expeditions: eraExpeditions(set),
	}
}

func eraExpeditions(set *rules.Set) []ExpeditionDef {
	ages := set.Ages()
	byEra := make(map[string][]string)
	for _, age := range ages {
		byEra[age.EpochKey] = append(byEra[age.EpochKey], age.Key)
	}
	var defs []ExpeditionDef
	for _, mission := range config.Missions() {
		bounds := byEra[mission.Era]
		category := ExpeditionScouting
		if mission.Kind == config.MissionCampaign {
			category = ExpeditionMilitary
		}
		defs = append(defs, ExpeditionDef{
			Name: mission.Name, Key: mission.Key, Era: mission.Era, Kind: mission.Kind,
			Category: category, MinAge: bounds[0], MaxAge: bounds[len(bounds)-1],
			DurationMin:    int(math.Round(mission.MinMinutes * 60 / config.TickSeconds)),
			DurationMax:    int(math.Round(mission.MaxMinutes * 60 / config.TickSeconds)),
			DifficultyBase: 1 - mission.BaseSuccess, Description: mission.Description,
			MinMinutes: mission.MinMinutes, MaxMinutes: mission.MaxMinutes,
			BaseSuccess: mission.BaseSuccess, Yield: mission.Yield, CostMinutes: mission.CostMinutes,
		})
	}
	return defs
}

// ExpeditionDefByKey returns the definition for an expedition key, or nil.
func (mm *MilitaryManager) ExpeditionDefByKey(key string) *ExpeditionDef {
	for i := range mm.expeditions {
		if mm.expeditions[i].Key == key {
			return &mm.expeditions[i]
		}
	}
	return nil
}

// LaunchExpedition validates age + per-category active status and sets up the
// active expedition for its category. Resource costs (soldiers + Cost) are
// validated and deducted by the engine BEFORE this is called — this method
// assumes the player can afford the launch and only enforces age range and the
// one-active-per-category rule (a scouting and a military expedition may run
// concurrently, but not two of the same category).
//
// The duration roll comes from rng (the engine's seeded run RNG).
func (mm *MilitaryManager) LaunchExpedition(rng *rand.Rand, key, currentAge string, ageOrder map[string]int) error {
	return mm.LaunchMission(rng, key, currentAge, ageOrder, 0, 0, nil, nil, 0)
}

// LaunchMission records the three launch rolls and the state-sized cost and haul.
func (mm *MilitaryManager) LaunchMission(rng *rand.Rand, key, currentAge string, ageOrder map[string]int, power, typical float64, cost, haul map[string]float64, soldiers float64) error {
	def := mm.ExpeditionDefByKey(key)
	if def == nil {
		return fmt.Errorf("unknown expedition or campaign '%s'. Type expedition or campaign to see what you can send", key)
	}

	if existing := mm.activeByCat[def.Category]; existing != nil {
		return fmt.Errorf("%s is already under way (%s left). Wait for it to come back", categoryNoun(def.Category), approxTicks(existing.TicksLeft, BaseTickInterval))
	}

	if def.Era != config.EpochForAge(currentAge) {
		return fmt.Errorf("%s is available in the %s", def.Name, mm.rules.Name(rules.KindAge, def.MinAge))
	}
	length := def.MinMinutes + float64(rng.Float64()*(def.MaxMinutes-def.MinMinutes))
	success := def.BaseSuccess + (float64(rng.Float64()*0.2) - 0.1)
	if def.Kind == config.MissionCampaign && typical > 0 {
		army := 0.10 * (power/typical - 1)
		success += math.Max(-0.15, math.Min(0.15, army))
	}
	success = math.Max(0.10, math.Min(0.95, success))
	haulBase := float64(length * def.Yield * 0.6)
	haulMinutes := float64(length*def.Yield*rng.Float64()) + haulBase
	baseTicks := int(math.Round(length * 60 / config.TickSeconds))
	ticks := mm.missionTicks(def.Category, baseTicks)
	if ticks < 1 {
		ticks = 1
	}
	if def.Kind == config.MissionCampaign {
		soldiers = math.Max(1, soldiers)
	}

	mm.activeByCat[def.Category] = &ActiveExpedition{
		Key:           key,
		Name:          def.Name,
		Soldiers:      int(soldiers),
		TicksLeft:     ticks,
		LengthMinutes: length, SuccessChance: success, HaulMinutes: haulMinutes,
		BaseRewards: maps.Clone(haul), Cost: maps.Clone(cost),
	}
	return nil
}

// categoryNoun is the player-facing noun phrase for one mission of a
// category, sentence-initial: "An expedition" (scouting) or "A campaign"
// (military).
func categoryNoun(category string) string {
	if category == ExpeditionMilitary {
		return "A campaign"
	}
	return "An expedition"
}

// categoryCommand is the command that sends missions of a category:
// expedition for scouting, campaign for military.
func categoryCommand(category string) string {
	if category == ExpeditionMilitary {
		return "campaign"
	}
	return "expedition"
}

// ActiveByCategory returns the active expedition for a category, or nil. Used by
// the engine/tests; the returned pointer is the live manager state.
func (mm *MilitaryManager) ActiveByCategory(category string) *ActiveExpedition {
	return mm.activeByCat[category]
}

// HasActive reports whether any expedition (any category) is currently running.
func (mm *MilitaryManager) HasActive() bool {
	for _, exp := range mm.activeByCat {
		if exp != nil {
			return true
		}
	}
	return false
}

// ExpeditionResult holds the rewards + player-facing message for one expedition
// that resolved this tick. Category and Success describe HOW it resolved so the
// engine can weight a faction encounter on resolution (see
// GameEngine.rollExpeditionEncounter).
//
// Key and Name identify WHICH expedition resolved. They exist because the varied
// flavour line is generated ENGINE-side (MilitaryManager has no rng; ge.rng does,
// and the prose stream must come off the same seeded source as everything else) —
// see GameEngine.expeditionFlavorLine. Message stays the mechanical line; flavour
// rides alongside it, never in place of it.
type ExpeditionResult struct {
	Rewards      map[string]float64
	Message      string
	Key          string // ExpeditionDef.Key of the expedition that resolved
	Name         string // ExpeditionDef.Name — a verb-led order title as often as a noun phrase
	Category     string // ExpeditionScouting or ExpeditionMilitary
	Success      bool   // true if the expedition succeeded
	SoldiersSent int
}

// Tick advances every active expedition (one per category) by one tick and
// returns a result for each that resolved this tick. Categories tick down and
// complete independently — a scouting and a military expedition resolve on their
// own schedules.
//
// Success probability per expedition: successRoll > config.MissionDifficulty
// (DifficultyBase less 0.3 × missionPower, 5% at least). missionPower is the
// player's military power on the current age's yardstick (rules.Set.MissionPower);
// expeditionBonus scales reward amounts.
// Soldiers are spent at launch (win or lose); success vs failure differs only in
// reward (full vs 30%), not in any extra soldier loss. The success roll comes
// from rng, one draw per resolving expedition, scouting before military.
func (mm *MilitaryManager) Tick(rng *rand.Rand, missionPower, expeditionBonus float64) []ExpeditionResult {
	// Iterate categories in a stable order so resolution logs/results are
	// deterministic regardless of map iteration order.
	cats := []string{ExpeditionScouting, ExpeditionMilitary}
	var results []ExpeditionResult
	for _, cat := range cats {
		if res, ok := mm.tickCategory(rng, cat, missionPower, expeditionBonus); ok {
			results = append(results, res)
		}
	}
	return results
}

// tickCategory advances one category's active expedition. ok is true only on the
// tick the expedition resolves (carrying its rewards + message).
func (mm *MilitaryManager) tickCategory(rng *rand.Rand, category string, missionPower, expeditionBonus float64) (ExpeditionResult, bool) {
	active := mm.activeByCat[category]
	if active == nil {
		return ExpeditionResult{}, false
	}

	active.TicksLeft--
	if active.TicksLeft > 0 {
		return ExpeditionResult{}, false
	}

	// Expedition complete - calculate results
	def := mm.ExpeditionDefByKey(active.Key)
	if def == nil {
		mm.activeByCat[category] = nil
		return ExpeditionResult{}, false
	}

	// Success calculation: military power, on the age's yardstick, reduces
	// difficulty (config.MissionDifficulty).
	successRoll := rng.Float64()
	successChance := active.SuccessChance
	if successChance == 0 {
		successChance = def.BaseSuccess
	}
	success := successRoll < successChance

	rewards := make(map[string]float64)
	var message string
	if success {
		// Apply expedition reward bonus
		rewardMult := 1.0 + expeditionBonus
		loot := active.SuccessRewards
		for res, amount := range loot {
			rewards[res] = amount
			if loot == nil {
				rewards[res] = amount * rewardMult
				if category == ExpeditionMilitary {
					rewards[res] = mm.CampaignPay(rewards[res])
				}
			}
			mm.totalLoot[res] += rewards[res]
		}
		message = fmt.Sprintf("%s succeeded. Loot: %s.", def.Name, amountsText(rewards))
	} else {
		// Partial rewards on failure
		loot := active.FailureRewards
		for res, amount := range loot {
			partial := amount
			if loot == nil {
				partial = amount * failedLootShare
				if category == ExpeditionMilitary {
					partial = mm.CampaignPay(partial)
				}
			}
			rewards[res] = partial
			mm.totalLoot[res] += partial
		}
		message = fmt.Sprintf("%s failed. You kept %.0f%% of the loot: %s.", def.Name, failedLootShare*100, amountsText(rewards))
	}

	mm.completedCount++
	mm.activeByCat[category] = nil
	return ExpeditionResult{
		Rewards:      rewards,
		Message:      message,
		Key:          def.Key,
		Name:         def.Name,
		Category:     category,
		Success:      success,
		SoldiersSent: active.Soldiers,
	}, true
}

// GetAvailableExpeditions returns expeditions available for the current age,
// respecting both MinAge and MaxAge bounds.
func (mm *MilitaryManager) GetAvailableExpeditions(currentAge string, ageOrder map[string]int) []ExpeditionDef {
	var available []ExpeditionDef
	for _, def := range mm.expeditions {
		if def.Era != config.EpochForAge(currentAge) {
			continue
		}
		available = append(available, def)
	}
	return available
}

// GetAvailableExpeditionsByCategory is GetAvailableExpeditions filtered to one
// Category (ExpeditionScouting or ExpeditionMilitary).
func (mm *MilitaryManager) GetAvailableExpeditionsByCategory(category, currentAge string, ageOrder map[string]int) []ExpeditionDef {
	var filtered []ExpeditionDef
	for _, def := range mm.GetAvailableExpeditions(currentAge, ageOrder) {
		if def.Category == category {
			filtered = append(filtered, def)
		}
	}
	return filtered
}

// launchability reports whether an expedition can be launched right now and, if
// not, a short player-facing reason. It mirrors the engine's LaunchExpedition
// validation order: single-active rule, then soldiers, then each Cost resource.
// soldierCount is the current soldiers amount; resources is the live resource
// map (nil → Cost treated as unaffordable).
func (mm *MilitaryManager) launchability(def ExpeditionDef, soldierCount int, resources map[string]float64) (bool, string) {
	if mm.activeByCat[def.Category] != nil {
		return false, fmt.Sprintf("%s is already under way: wait for it to come back", categoryNoun(def.Category))
	}
	if soldierCount < def.SoldiersNeeded {
		return false, fmt.Sprintf("needs %d soldiers (you have %d): military buildings train them", def.SoldiersNeeded, soldierCount)
	}
	// Check Cost resources in a stable (sorted) order so the surfaced reason is
	// deterministic regardless of map iteration order.
	keys := make([]string, 0, len(def.Cost))
	for res := range def.Cost {
		keys = append(keys, res)
	}
	sort.Strings(keys)
	for _, res := range keys {
		amount := def.Cost[res]
		if resources[res] < amount {
			return false, fmt.Sprintf("needs %s %s (you have %s)", amountText(amount), resourceLabel(res), amountText(resources[res]))
		}
	}
	return true, ""
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
// currentResources is the live resource map (from the engine, which owns
// ResourceManager) used to decide whether each expedition's Cost is affordable —
// CanLaunch requires soldiers, not-active, AND every Cost resource covered.
// currentResources may be nil, in which case Cost affordability is treated as
// unmet (defensive; the engine always passes a real map).
func (mm *MilitaryManager) Snapshot(currentAge string, ageOrder map[string]int, soldierCount, soldierCap int, soldierRate float64, currentResources map[string]float64, militaryBonus, expeditionBonus float64) MilitaryState {
	activeScout := snapshotActive(mm.activeByCat[ExpeditionScouting], expeditionBonus, false, mm.campaignPay)
	activeMilitary := snapshotActive(mm.activeByCat[ExpeditionMilitary], expeditionBonus, true, mm.campaignPay)

	available := mm.GetAvailableExpeditions(currentAge, ageOrder)
	var expList []ExpeditionInfo
	for _, def := range available {
		canLaunch, reason := mm.launchability(def, soldierCount, currentResources)
		expList = append(expList, ExpeditionInfo{
			Name:              def.Name,
			Key:               def.Key,
			Description:       def.Description,
			Kind:              def.Kind,
			MinMinutes:        def.MinMinutes,
			MaxMinutes:        def.MaxMinutes,
			BaseSuccess:       def.BaseSuccess,
			Category:          def.Category,
			DurationMin:       mm.missionTicks(def.Category, def.DurationMin),
			DurationMax:       mm.missionTicks(def.Category, def.DurationMax),
			Difficulty:        def.DifficultyBase,
			Cost:              maps.Clone(def.Cost), // def is the manager's table
			CanLaunch:         canLaunch,
			LaunchBlockReason: reason,
		})
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
		ActiveScout:     activeScout,
		ActiveMilitary:  activeMilitary,
		Expeditions:     expList,
		CompletedCount:  mm.completedCount,
		TotalLoot:       loot,
	}
}

// snapshotActive converts a live ActiveExpedition into a UI snapshot, or nil.
func snapshotActive(active *ActiveExpedition, expeditionBonus float64, campaign bool, campaignPay float64) *ExpeditionSnapshot {
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
	if campaign && active.SuccessRewards == nil {
		if campaignPay <= 1 {
			campaignPay = 1
		}
		for res := range success {
			success[res] *= campaignPay
		}
		for res := range failure {
			failure[res] *= campaignPay
		}
	}
	return &ExpeditionSnapshot{
		Name:          active.Name,
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

// LoadState restores military state from save. scout and military are the
// per-category active expeditions (either may be nil). Legacy single-active
// saves are migrated by the caller (see save.go) before reaching this point.
func (mm *MilitaryManager) LoadState(scout, military *ActiveExpedition, completedCount int, totalLoot map[string]float64) {
	if mm.activeByCat == nil {
		mm.activeByCat = make(map[string]*ActiveExpedition)
	}
	mm.activeByCat[ExpeditionScouting] = scout
	mm.activeByCat[ExpeditionMilitary] = military
	mm.completedCount = completedCount
	if totalLoot != nil {
		mm.totalLoot = totalLoot
	}
}

// GetActiveForSave returns deep copies of the active scouting and military
// expeditions for saving. Either may be nil.
func (mm *MilitaryManager) GetActiveForSave() (scout, military *ActiveExpedition) {
	return copyActive(mm.activeByCat[ExpeditionScouting]), copyActive(mm.activeByCat[ExpeditionMilitary])
}

// copyActive returns a deep copy of an ActiveExpedition, or nil.
func copyActive(active *ActiveExpedition) *ActiveExpedition {
	if active == nil {
		return nil
	}
	c := *active
	return &c
}

// migrateActives resolves the per-category active expeditions from a MilitarySave,
// migrating pre-Phase-2a saves. If the new ActiveScout/ActiveMilitary fields are
// present they win. Otherwise a legacy single-active ActiveExpedition is routed to
// the scout or military slot by its def's Category (unknown keys default to the
// military slot, matching pre-2a semantics where expeditions were "military").
func (mm *MilitaryManager) migrateActives(save MilitarySave) (scout, military *ActiveExpedition) {
	scout = save.ActiveScout
	military = save.ActiveMilitary
	if scout != nil || military != nil {
		return scout, military
	}
	if save.ActiveExpedition == nil {
		return nil, nil
	}
	legacy := save.ActiveExpedition
	if def := mm.ExpeditionDefByKey(legacy.Key); def != nil && def.Category == ExpeditionScouting {
		return legacy, nil
	}
	return nil, legacy
}
