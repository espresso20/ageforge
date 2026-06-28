package game

import (
	"fmt"
	"math/rand"
	"sort"
)

// Expedition categories. Scouting expeditions cost only resources (no soldiers)
// and are playable before the soldiers resource exists (pre-iron_age). Military
// expeditions cost soldiers (plus any Cost).
const (
	ExpeditionScouting = "scouting"
	ExpeditionMilitary = "military"
)

// ExpeditionDef defines an available expedition
type ExpeditionDef struct {
	Name           string
	Key            string
	Category       string // ExpeditionScouting or ExpeditionMilitary
	MinAge         string
	MaxAge         string // empty = no upper bound; expedition is unavailable once past this age
	SoldiersNeeded int
	Duration       int                // ticks
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
	expeditions    []ExpeditionDef
	activeByCat    map[string]*ActiveExpedition
	completedCount int
	totalLoot      map[string]float64
	defenseRating  float64
}

// NewMilitaryManager creates a military manager
func NewMilitaryManager() *MilitaryManager {
	return &MilitaryManager{
		totalLoot:   make(map[string]float64),
		activeByCat: make(map[string]*ActiveExpedition),
		expeditions: []ExpeditionDef{
			{
				Name: "Scout Party", Key: "scout_party",
				Category: ExpeditionScouting,
				MinAge:   "primitive_age", MaxAge: "bronze_age",
				SoldiersNeeded: 0, Duration: 20,
				DifficultyBase: 0.2,
				Cost:           map[string]float64{"food": 30, "wood": 30},
				Rewards:        map[string]float64{"food": 60, "wood": 60, "stone": 20},
				Description:    "A small band of foragers scouts nearby territory for resources.",
			},
			{
				Name: "Scout Nearby Ruins", Key: "scout_ruins",
				Category: ExpeditionScouting,
				MinAge:   "bronze_age", SoldiersNeeded: 0, Duration: 10,
				DifficultyBase: 0.2,
				Cost:           map[string]float64{"food": 40, "wood": 30},
				Rewards:        map[string]float64{"food": 30, "wood": 20, "stone": 15},
				Description:    "Send scouts to explore nearby ruins for resources.",
			},
			{
				Name: "Raid Bandit Camp", Key: "raid_bandits",
				Category: ExpeditionMilitary,
				MinAge:   "bronze_age", SoldiersNeeded: 5, Duration: 15,
				DifficultyBase: 0.4,
				Rewards:        map[string]float64{"gold": 30, "iron": 15, "food": 20},
				Description:    "Attack a bandit encampment and seize their loot.",
			},
			{
				Name: "Trade Escort", Key: "trade_escort",
				Category: ExpeditionMilitary,
				MinAge:   "iron_age", SoldiersNeeded: 3, Duration: 12,
				DifficultyBase: 0.3,
				Rewards:        map[string]float64{"gold": 50, "knowledge": 10},
				Description:    "Escort merchants on a dangerous trade route.",
			},
			{
				Name: "Conquer Territory", Key: "conquer_territory",
				Category: ExpeditionMilitary,
				MinAge:   "iron_age", SoldiersNeeded: 10, Duration: 25,
				DifficultyBase: 0.6,
				Rewards:        map[string]float64{"gold": 80, "iron": 40, "food": 50},
				Description:    "Conquer a neighboring territory for its resources.",
			},
			{
				Name: "Siege Enemy Castle", Key: "siege_castle",
				Category: ExpeditionMilitary,
				MinAge:   "medieval_age", SoldiersNeeded: 15, Duration: 30,
				DifficultyBase: 0.7,
				Rewards:        map[string]float64{"gold": 150, "steel": 30, "faith": 20},
				Description:    "Lay siege to an enemy stronghold.",
			},
			{
				Name: "Naval Expedition", Key: "naval_expedition",
				Category: ExpeditionScouting,
				MinAge:   "renaissance_age", SoldiersNeeded: 0, Duration: 35,
				DifficultyBase: 0.5,
				Cost:           map[string]float64{"food": 150, "wood": 100},
				Rewards:        map[string]float64{"gold": 200, "culture": 30, "knowledge": 40},
				Description:    "Explore distant lands by sea.",
			},
			{
				Name: "Colonial Campaign", Key: "colonial_campaign",
				Category: ExpeditionMilitary,
				MinAge:   "industrial_age", SoldiersNeeded: 20, Duration: 40,
				DifficultyBase: 0.6,
				Rewards:        map[string]float64{"gold": 300, "oil": 50, "steel": 40},
				Description:    "Establish colonial presence in new territories.",
			},
			{
				Name: "World Domination", Key: "world_domination",
				Category: ExpeditionMilitary,
				MinAge:   "modern_age", SoldiersNeeded: 50, Duration: 60,
				DifficultyBase: 0.8,
				Rewards:        map[string]float64{"gold": 1000, "electricity": 200, "knowledge": 500},
				Description:    "Launch a global military campaign for world domination.",
			},
			{
				Name: "Cyber Raid", Key: "cyber_raid",
				Category: ExpeditionMilitary,
				MinAge:   "information_age", SoldiersNeeded: 30, Duration: 45,
				DifficultyBase: 0.6,
				Rewards:        map[string]float64{"data": 200, "crypto": 50, "gold": 500},
				Description:    "Hack into enemy networks and steal digital assets.",
			},
			{
				Name: "Neon Heist", Key: "neon_heist",
				Category: ExpeditionMilitary,
				MinAge:   "cyberpunk_age", SoldiersNeeded: 25, Duration: 35,
				DifficultyBase: 0.55,
				Rewards:        map[string]float64{"crypto": 100, "data": 150, "gold": 800},
				Description:    "Pull off a daring heist in the neon-lit underworld.",
			},
			{
				Name: "Fusion Plant Assault", Key: "fusion_assault",
				Category: ExpeditionMilitary,
				MinAge:   "fusion_age", SoldiersNeeded: 35, Duration: 40,
				DifficultyBase: 0.65,
				Rewards:        map[string]float64{"plasma": 120, "electricity": 500, "uranium": 50},
				Description:    "Capture a rival's fusion power facility.",
			},
			{
				Name: "Orbital Strike", Key: "orbital_strike",
				Category: ExpeditionMilitary,
				MinAge:   "space_age", SoldiersNeeded: 40, Duration: 50,
				DifficultyBase: 0.7,
				Rewards:        map[string]float64{"titanium": 100, "plasma": 80, "knowledge": 300},
				Description:    "Deploy orbital weapons platform against hostile targets.",
			},
			{
				Name: "Warp Invasion", Key: "warp_invasion",
				Category: ExpeditionMilitary,
				MinAge:   "interstellar_age", SoldiersNeeded: 60, Duration: 65,
				DifficultyBase: 0.75,
				Rewards:        map[string]float64{"dark_matter": 50, "titanium": 200, "gold": 2000},
				Description:    "Invade a neighboring star system through warp gates.",
			},
			{
				Name: "Galactic Conquest", Key: "galactic_conquest",
				Category: ExpeditionMilitary,
				MinAge:   "galactic_age", SoldiersNeeded: 80, Duration: 80,
				DifficultyBase: 0.8,
				Rewards:        map[string]float64{"antimatter": 30, "dark_matter": 100, "gold": 5000},
				Description:    "Conquer an entire galactic sector.",
			},
			{
				Name: "Quantum Incursion", Key: "quantum_incursion",
				Category: ExpeditionMilitary,
				MinAge:   "quantum_age", SoldiersNeeded: 100, Duration: 90,
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
func (mm *MilitaryManager) LaunchExpedition(key, currentAge string, ageOrder map[string]int) error {
	def := mm.ExpeditionDefByKey(key)
	if def == nil {
		return fmt.Errorf("unknown expedition: %s", key)
	}

	if existing := mm.activeByCat[def.Category]; existing != nil {
		return fmt.Errorf("a %s expedition is already in progress (%d ticks left)", categoryLabel(def.Category), existing.TicksLeft)
	}

	if ageOrder[def.MinAge] > ageOrder[currentAge] {
		return fmt.Errorf("%s requires %s age", def.Name, def.MinAge)
	}
	if def.MaxAge != "" && ageOrder[currentAge] > ageOrder[def.MaxAge] {
		return fmt.Errorf("%s is no longer available past the %s age", def.Name, def.MaxAge)
	}

	mm.activeByCat[def.Category] = &ActiveExpedition{
		Key:       key,
		Name:      def.Name,
		Soldiers:  def.SoldiersNeeded,
		TicksLeft: def.Duration,
	}
	return nil
}

// categoryLabel returns a short player-facing word for an expedition category.
func categoryLabel(category string) string {
	switch category {
	case ExpeditionScouting:
		return "scouting"
	case ExpeditionMilitary:
		return "military"
	default:
		return category
	}
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
// that resolved this tick.
type ExpeditionResult struct {
	Rewards map[string]float64
	Message string
}

// Tick advances every active expedition (one per category) by one tick and
// returns a result for each that resolved this tick. Categories tick down and
// complete independently — a scouting and a military expedition resolve on their
// own schedules.
//
// Success probability per expedition: successRoll > (DifficultyBase - militaryBonus×0.3).
// militaryBonus reduces effective difficulty; expeditionBonus scales reward amounts.
// Soldiers are spent at launch (win or lose); success vs failure differs only in
// reward (full vs 30%), not in any extra soldier loss.
func (mm *MilitaryManager) Tick(militaryBonus, expeditionBonus float64) []ExpeditionResult {
	// Iterate categories in a stable order so resolution logs/results are
	// deterministic regardless of map iteration order.
	cats := []string{ExpeditionScouting, ExpeditionMilitary}
	var results []ExpeditionResult
	for _, cat := range cats {
		if res, ok := mm.tickCategory(cat, militaryBonus, expeditionBonus); ok {
			results = append(results, res)
		}
	}
	return results
}

// tickCategory advances one category's active expedition. ok is true only on the
// tick the expedition resolves (carrying its rewards + message).
func (mm *MilitaryManager) tickCategory(category string, militaryBonus, expeditionBonus float64) (ExpeditionResult, bool) {
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

	// Success calculation: military bonus reduces difficulty
	difficulty := def.DifficultyBase - (militaryBonus * 0.3)
	if difficulty < 0.05 {
		difficulty = 0.05
	}

	successRoll := rand.Float64()
	success := successRoll > difficulty

	rewards := make(map[string]float64)
	var message string
	if success {
		// Apply expedition reward bonus
		rewardMult := 1.0 + expeditionBonus
		for res, amount := range def.Rewards {
			rewards[res] = amount * rewardMult
			mm.totalLoot[res] += rewards[res]
		}
		message = fmt.Sprintf("%s succeeded! Gained loot.", def.Name)
	} else {
		// Partial rewards on failure
		for res, amount := range def.Rewards {
			partial := amount * 0.3
			rewards[res] = partial
			mm.totalLoot[res] += partial
		}
		message = fmt.Sprintf("%s failed! Partial loot recovered.", def.Name)
	}

	mm.completedCount++
	mm.activeByCat[category] = nil
	return ExpeditionResult{Rewards: rewards, Message: message}, true
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
		return false, fmt.Sprintf("a %s expedition is already in progress", categoryLabel(def.Category))
	}
	if soldierCount < def.SoldiersNeeded {
		return false, fmt.Sprintf("need %d soldiers", def.SoldiersNeeded)
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
			return false, fmt.Sprintf("need %.0f %s", amount, res)
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
			Duration:          def.Duration,
			Difficulty:        def.DifficultyBase,
			Cost:              def.Cost,
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
		SoldierCount:     soldierCount,
		SoldierCap:       soldierCap,
		SoldierRate:      soldierRate,
		DefenseRating:    mm.CalculateDefense(soldierCount, militaryBonus),
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
