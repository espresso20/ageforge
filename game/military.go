package game

import (
	"fmt"
	"maps"
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
}

// ActiveExpedition represents an ongoing expedition
type ActiveExpedition struct {
	Key       string
	Name      string
	Soldiers  int
	TicksLeft int
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
		expeditions: []ExpeditionDef{
			{
				Name: "Scout Party", Key: "scout_party",
				Category: ExpeditionScouting,
				MinAge:   "primitive_age", MaxAge: "bronze_age",
				SoldiersNeeded: 0,
				DurationMin:    100, DurationMax: 160,
				DifficultyBase: 0.2,
				Cost:           map[string]float64{"food": 30, "wood": 30},
				Rewards:        map[string]float64{"food": 60, "wood": 60, "stone": 20},
				Description:    "A small band of foragers scouts nearby territory for resources.",
			},
			{
				Name: "Scout Nearby Ruins", Key: "scout_ruins",
				Category: ExpeditionScouting,
				MinAge:   "bronze_age", SoldiersNeeded: 0,
				DurationMin: 60, DurationMax: 100,
				DifficultyBase: 0.2,
				Cost:           map[string]float64{"food": 40, "wood": 30},
				Rewards:        map[string]float64{"food": 30, "wood": 20, "stone": 15},
				Description:    "Send scouts to explore nearby ruins for resources.",
			},
			{
				Name: "Raid Bandit Camp", Key: "raid_bandits",
				Category: ExpeditionMilitary,
				MinAge:   "bronze_age", SoldiersNeeded: 5,
				DurationMin: 60, DurationMax: 100,
				DifficultyBase: 0.4,
				Rewards:        map[string]float64{"gold": 30, "iron": 15, "food": 20},
				Description:    "Attack a bandit encampment and seize their loot.",
			},
			{
				Name: "Trade Escort", Key: "trade_escort",
				Category: ExpeditionMilitary,
				MinAge:   "iron_age", SoldiersNeeded: 3,
				DurationMin: 60, DurationMax: 100,
				DifficultyBase: 0.3,
				Rewards:        map[string]float64{"gold": 50, "knowledge": 10},
				Description:    "Escort merchants on a dangerous trade route.",
			},
			{
				Name: "Conquer Territory", Key: "conquer_territory",
				Category: ExpeditionMilitary,
				MinAge:   "iron_age", SoldiersNeeded: 10,
				DurationMin: 60, DurationMax: 100,
				DifficultyBase: 0.6,
				Rewards:        map[string]float64{"gold": 80, "iron": 40, "food": 50},
				Description:    "Conquer a neighboring territory for its resources.",
			},
			{
				Name: "Siege Enemy Castle", Key: "siege_castle",
				Category: ExpeditionMilitary,
				MinAge:   "medieval_age", SoldiersNeeded: 15,
				DurationMin: 60, DurationMax: 100,
				DifficultyBase: 0.7,
				Rewards:        map[string]float64{"gold": 150, "steel": 30, "faith": 20},
				Description:    "Lay siege to an enemy stronghold.",
			},
			{
				Name: "Naval Expedition", Key: "naval_expedition",
				Category: ExpeditionScouting,
				MinAge:   "renaissance_age", SoldiersNeeded: 0,
				DurationMin: 60, DurationMax: 100,
				DifficultyBase: 0.5,
				Cost:           map[string]float64{"food": 150, "wood": 100},
				Rewards:        map[string]float64{"gold": 200, "culture": 30, "knowledge": 40},
				Description:    "Explore distant lands by sea.",
			},
			{
				Name: "Colonial Campaign", Key: "colonial_campaign",
				Category: ExpeditionMilitary,
				MinAge:   "industrial_age", SoldiersNeeded: 20,
				DurationMin: 60, DurationMax: 100,
				DifficultyBase: 0.6,
				Rewards:        map[string]float64{"gold": 300, "oil": 50, "steel": 40},
				Description:    "Establish colonial presence in new territories.",
			},
			{
				Name: "World Domination", Key: "world_domination",
				Category: ExpeditionMilitary,
				MinAge:   "modern_age", SoldiersNeeded: 50,
				DurationMin: 60, DurationMax: 100,
				DifficultyBase: 0.8,
				Rewards:        map[string]float64{"gold": 1000, "electricity": 200, "knowledge": 500},
				Description:    "Launch a global military campaign for world domination.",
			},
			{
				Name: "Cyber Raid", Key: "cyber_raid",
				Category: ExpeditionMilitary,
				MinAge:   "information_age", SoldiersNeeded: 30,
				DurationMin: 60, DurationMax: 100,
				DifficultyBase: 0.6,
				Rewards:        map[string]float64{"data": 200, "crypto": 50, "gold": 500},
				Description:    "Hack into enemy networks and steal digital assets.",
			},
			{
				Name: "Neon Heist", Key: "neon_heist",
				Category: ExpeditionMilitary,
				MinAge:   "cyberpunk_age", SoldiersNeeded: 25,
				DurationMin: 60, DurationMax: 100,
				DifficultyBase: 0.55,
				Rewards:        map[string]float64{"crypto": 100, "data": 150, "gold": 800},
				Description:    "Pull off a daring heist in the neon-lit underworld.",
			},
			{
				Name: "Fusion Plant Assault", Key: "fusion_assault",
				Category: ExpeditionMilitary,
				MinAge:   "fusion_age", SoldiersNeeded: 35,
				DurationMin: 60, DurationMax: 100,
				DifficultyBase: 0.65,
				Rewards:        map[string]float64{"plasma": 120, "electricity": 500, "uranium": 50},
				Description:    "Capture a rival's fusion power facility.",
			},
			{
				Name: "Orbital Strike", Key: "orbital_strike",
				Category: ExpeditionMilitary,
				MinAge:   "space_age", SoldiersNeeded: 40,
				DurationMin: 60, DurationMax: 100,
				DifficultyBase: 0.7,
				Rewards:        map[string]float64{"titanium": 100, "plasma": 80, "knowledge": 300},
				Description:    "Deploy orbital weapons platform against hostile targets.",
			},
			{
				Name: "Warp Invasion", Key: "warp_invasion",
				Category: ExpeditionMilitary,
				MinAge:   "interstellar_age", SoldiersNeeded: 60,
				DurationMin: 65, DurationMax: 105,
				DifficultyBase: 0.75,
				Rewards:        map[string]float64{"dark_matter": 50, "titanium": 200, "gold": 2000},
				Description:    "Invade a neighboring star system through warp gates.",
			},
			{
				Name: "Galactic Conquest", Key: "galactic_conquest",
				Category: ExpeditionMilitary,
				MinAge:   "galactic_age", SoldiersNeeded: 80,
				DurationMin: 80, DurationMax: 130,
				DifficultyBase: 0.8,
				Rewards:        map[string]float64{"antimatter": 30, "dark_matter": 100, "gold": 5000},
				Description:    "Conquer an entire galactic sector.",
			},
			{
				Name: "Quantum Incursion", Key: "quantum_incursion",
				Category: ExpeditionMilitary,
				MinAge:   "quantum_age", SoldiersNeeded: 100,
				DurationMin: 90, DurationMax: 145,
				DifficultyBase: 0.85,
				Rewards:        map[string]float64{"quantum_flux": 20, "antimatter": 50, "knowledge": 5000},
				Description:    "Launch an incursion across quantum realities.",
			},
		},
	}
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
	def := mm.ExpeditionDefByKey(key)
	if def == nil {
		return fmt.Errorf("unknown expedition or campaign '%s'. Type expedition or campaign to see what you can send", key)
	}

	if existing := mm.activeByCat[def.Category]; existing != nil {
		return fmt.Errorf("%s is already under way (%s left). Wait for it to come back", categoryNoun(def.Category), approxTicks(existing.TicksLeft, BaseTickInterval))
	}

	if ageOrder[def.MinAge] > ageOrder[currentAge] {
		return fmt.Errorf("%s needs %s. Advance to send it", def.Name, laterAgeRef(mm.rules, currentAge, def.MinAge))
	}
	if def.MaxAge != "" && ageOrder[currentAge] > ageOrder[def.MaxAge] {
		return fmt.Errorf("%s ended with the %s. Type %s to see what you can send now", def.Name, mm.rules.Name(rules.KindAge, def.MaxAge), categoryCommand(def.Category))
	}

	// Roll a randomized active duration in [DurationMin, DurationMax] (inclusive),
	// stretched for the age (the ruleset's StretchTicks: the defs' ranges are
	// typed for the base curve, and from the Bronze Age on the ages run
	// config.PacingStretch times longer and so do expeditions, so an age holds
	// as many of them, and their encounters, as it did; the roll is drawn in
	// the typed range first, so launching takes the same draws in every
	// age). Every shipped def carries a valid
	// range; the guards below only keep a malformed def from handing rng.Intn an
	// arg <= 0 or pinning an expedition at 0 ticks (which would resolve it
	// instantly, forever).
	ticks := def.DurationMin
	if def.DurationMax > def.DurationMin {
		ticks = def.DurationMin + rng.Intn(def.DurationMax-def.DurationMin+1)
	}
	if ticks <= 0 {
		ticks = minExpeditionDurationTicks
	}
	ticks = mm.missionTicks(def.Category, mm.rules.StretchTicks(currentAge, ticks))

	mm.activeByCat[def.Category] = &ActiveExpedition{
		Key:       key,
		Name:      def.Name,
		Soldiers:  def.SoldiersNeeded,
		TicksLeft: ticks,
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
	Rewards  map[string]float64
	Message  string
	Key      string // ExpeditionDef.Key of the expedition that resolved
	Name     string // ExpeditionDef.Name — a verb-led order title as often as a noun phrase
	Category string // ExpeditionScouting or ExpeditionMilitary
	Success  bool   // true if the expedition succeeded
}

// Tick advances every active expedition (one per category) by one tick and
// returns a result for each that resolved this tick. Categories tick down and
// complete independently — a scouting and a military expedition resolve on their
// own schedules.
//
// Success probability per expedition: successRoll > config.MissionDifficulty
// (DifficultyBase less 0.3 × missionPower, 5% at least). missionPower is the
// player's military power on the current age's yardstick (config.MissionPower);
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
	difficulty := config.MissionDifficulty(def.DifficultyBase, missionPower)

	successRoll := rng.Float64()
	success := successRoll > difficulty

	rewards := make(map[string]float64)
	var message string
	if success {
		// Apply expedition reward bonus
		rewardMult := 1.0 + expeditionBonus
		for res, amount := range def.Rewards {
			rewards[res] = amount * rewardMult
			if category == ExpeditionMilitary {
				rewards[res] = mm.CampaignPay(rewards[res])
			}
			mm.totalLoot[res] += rewards[res]
		}
		message = fmt.Sprintf("%s succeeded. Loot: %s.", def.Name, amountsText(rewards))
	} else {
		// Partial rewards on failure
		for res, amount := range def.Rewards {
			partial := float64(amount * failedLootShare)
			if category == ExpeditionMilitary {
				partial = mm.CampaignPay(partial)
			}
			rewards[res] = partial
			mm.totalLoot[res] += partial
		}
		message = fmt.Sprintf("%s failed. You kept %.0f%% of the loot: %s.", def.Name, failedLootShare*100, amountsText(rewards))
	}

	mm.completedCount++
	mm.activeByCat[category] = nil
	return ExpeditionResult{
		Rewards:  rewards,
		Message:  message,
		Key:      def.Key,
		Name:     def.Name,
		Category: category,
		Success:  success,
	}, true
}

// GetAvailableExpeditions returns expeditions available for the current age,
// respecting both MinAge and MaxAge bounds.
func (mm *MilitaryManager) GetAvailableExpeditions(currentAge string, ageOrder map[string]int) []ExpeditionDef {
	var available []ExpeditionDef
	for _, def := range mm.expeditions {
		if ageOrder[def.MinAge] > ageOrder[currentAge] {
			continue
		}
		if def.MaxAge != "" && ageOrder[currentAge] > ageOrder[def.MaxAge] {
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
	activeScout := snapshotActive(mm.activeByCat[ExpeditionScouting])
	activeMilitary := snapshotActive(mm.activeByCat[ExpeditionMilitary])

	available := mm.GetAvailableExpeditions(currentAge, ageOrder)
	var expList []ExpeditionInfo
	for _, def := range available {
		canLaunch, reason := mm.launchability(def, soldierCount, currentResources)
		expList = append(expList, ExpeditionInfo{
			Name:              def.Name,
			Key:               def.Key,
			Category:          def.Category,
			SoldiersNeeded:    def.SoldiersNeeded,
			DurationMin:       mm.missionTicks(def.Category, mm.rules.StretchTicks(currentAge, def.DurationMin)),
			DurationMax:       mm.missionTicks(def.Category, mm.rules.StretchTicks(currentAge, def.DurationMax)),
			Difficulty:        def.DifficultyBase,
			Cost:              maps.Clone(def.Cost), // def is the manager's table
			Description:       def.Description,
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
func snapshotActive(active *ActiveExpedition) *ExpeditionSnapshot {
	if active == nil {
		return nil
	}
	return &ExpeditionSnapshot{
		Name:      active.Name,
		Soldiers:  active.Soldiers,
		TicksLeft: active.TicksLeft,
	}
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
