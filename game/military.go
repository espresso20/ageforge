package game

import (
	"fmt"
	"math/rand"
	"sort"
)

// ExpeditionDef defines an available expedition
type ExpeditionDef struct {
	Name           string
	Key            string
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
// Only one expedition can be active at a time (active pointer). Soldiers are a
// real resource (config/resources.go): launching an expedition spends
// SoldiersNeeded of the soldiers resource plus any Cost, validated and deducted
// by the engine before LaunchExpedition is called. Soldiers are spent at launch
// (win or lose); success vs failure differs only in reward, not soldier loss.
type MilitaryManager struct {
	expeditions    []ExpeditionDef
	active         *ActiveExpedition
	completedCount int
	totalLoot      map[string]float64
	defenseRating  float64
}

// NewMilitaryManager creates a military manager
func NewMilitaryManager() *MilitaryManager {
	return &MilitaryManager{
		totalLoot: make(map[string]float64),
		expeditions: []ExpeditionDef{
			{
				Name: "Scout Party", Key: "scout_party",
				MinAge: "primitive_age", MaxAge: "bronze_age",
				SoldiersNeeded: 0, Duration: 20,
				DifficultyBase: 0.2,
				Cost:           map[string]float64{"food": 30, "wood": 30},
				Rewards:        map[string]float64{"food": 60, "wood": 60, "stone": 20},
				Description:    "A small band of foragers scouts nearby territory for resources.",
			},
			{
				Name: "Scout Nearby Ruins", Key: "scout_ruins",
				MinAge: "bronze_age", SoldiersNeeded: 2, Duration: 10,
				DifficultyBase: 0.2,
				Rewards:        map[string]float64{"food": 30, "wood": 20, "stone": 15},
				Description:    "Send scouts to explore nearby ruins for resources.",
			},
			{
				Name: "Raid Bandit Camp", Key: "raid_bandits",
				MinAge: "bronze_age", SoldiersNeeded: 5, Duration: 15,
				DifficultyBase: 0.4,
				Rewards:        map[string]float64{"gold": 30, "iron": 15, "food": 20},
				Description:    "Attack a bandit encampment and seize their loot.",
			},
			{
				Name: "Trade Escort", Key: "trade_escort",
				MinAge: "iron_age", SoldiersNeeded: 3, Duration: 12,
				DifficultyBase: 0.3,
				Rewards:        map[string]float64{"gold": 50, "knowledge": 10},
				Description:    "Escort merchants on a dangerous trade route.",
			},
			{
				Name: "Conquer Territory", Key: "conquer_territory",
				MinAge: "iron_age", SoldiersNeeded: 10, Duration: 25,
				DifficultyBase: 0.6,
				Rewards:        map[string]float64{"gold": 80, "iron": 40, "food": 50},
				Description:    "Conquer a neighboring territory for its resources.",
			},
			{
				Name: "Siege Enemy Castle", Key: "siege_castle",
				MinAge: "medieval_age", SoldiersNeeded: 15, Duration: 30,
				DifficultyBase: 0.7,
				Rewards:        map[string]float64{"gold": 150, "steel": 30, "faith": 20},
				Description:    "Lay siege to an enemy stronghold.",
			},
			{
				Name: "Naval Expedition", Key: "naval_expedition",
				MinAge: "renaissance_age", SoldiersNeeded: 10, Duration: 35,
				DifficultyBase: 0.5,
				Rewards:        map[string]float64{"gold": 200, "culture": 30, "knowledge": 40},
				Description:    "Explore distant lands by sea.",
			},
			{
				Name: "Colonial Campaign", Key: "colonial_campaign",
				MinAge: "industrial_age", SoldiersNeeded: 20, Duration: 40,
				DifficultyBase: 0.6,
				Rewards:        map[string]float64{"gold": 300, "oil": 50, "steel": 40},
				Description:    "Establish colonial presence in new territories.",
			},
			{
				Name: "World Domination", Key: "world_domination",
				MinAge: "modern_age", SoldiersNeeded: 50, Duration: 60,
				DifficultyBase: 0.8,
				Rewards:        map[string]float64{"gold": 1000, "electricity": 200, "knowledge": 500},
				Description:    "Launch a global military campaign for world domination.",
			},
			{
				Name: "Cyber Raid", Key: "cyber_raid",
				MinAge: "information_age", SoldiersNeeded: 30, Duration: 45,
				DifficultyBase: 0.6,
				Rewards:        map[string]float64{"data": 200, "crypto": 50, "gold": 500},
				Description:    "Hack into enemy networks and steal digital assets.",
			},
			{
				Name: "Neon Heist", Key: "neon_heist",
				MinAge: "cyberpunk_age", SoldiersNeeded: 25, Duration: 35,
				DifficultyBase: 0.55,
				Rewards:        map[string]float64{"crypto": 100, "data": 150, "gold": 800},
				Description:    "Pull off a daring heist in the neon-lit underworld.",
			},
			{
				Name: "Fusion Plant Assault", Key: "fusion_assault",
				MinAge: "fusion_age", SoldiersNeeded: 35, Duration: 40,
				DifficultyBase: 0.65,
				Rewards:        map[string]float64{"plasma": 120, "electricity": 500, "uranium": 50},
				Description:    "Capture a rival's fusion power facility.",
			},
			{
				Name: "Orbital Strike", Key: "orbital_strike",
				MinAge: "space_age", SoldiersNeeded: 40, Duration: 50,
				DifficultyBase: 0.7,
				Rewards:        map[string]float64{"titanium": 100, "plasma": 80, "knowledge": 300},
				Description:    "Deploy orbital weapons platform against hostile targets.",
			},
			{
				Name: "Warp Invasion", Key: "warp_invasion",
				MinAge: "interstellar_age", SoldiersNeeded: 60, Duration: 65,
				DifficultyBase: 0.75,
				Rewards:        map[string]float64{"dark_matter": 50, "titanium": 200, "gold": 2000},
				Description:    "Invade a neighboring star system through warp gates.",
			},
			{
				Name: "Galactic Conquest", Key: "galactic_conquest",
				MinAge: "galactic_age", SoldiersNeeded: 80, Duration: 80,
				DifficultyBase: 0.8,
				Rewards:        map[string]float64{"antimatter": 30, "dark_matter": 100, "gold": 5000},
				Description:    "Conquer an entire galactic sector.",
			},
			{
				Name: "Quantum Incursion", Key: "quantum_incursion",
				MinAge: "quantum_age", SoldiersNeeded: 100, Duration: 90,
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

// LaunchExpedition validates age + active status and sets up the active
// expedition. Resource costs (soldiers + Cost) are validated and deducted by
// the engine BEFORE this is called — this method assumes the player can afford
// the launch and only enforces age range and the single-active-expedition rule.
func (mm *MilitaryManager) LaunchExpedition(key, currentAge string, ageOrder map[string]int) error {
	if mm.active != nil {
		return fmt.Errorf("expedition '%s' already in progress (%d ticks left)", mm.active.Name, mm.active.TicksLeft)
	}

	def := mm.ExpeditionDefByKey(key)
	if def == nil {
		return fmt.Errorf("unknown expedition: %s", key)
	}

	if ageOrder[def.MinAge] > ageOrder[currentAge] {
		return fmt.Errorf("%s requires %s age", def.Name, def.MinAge)
	}
	if def.MaxAge != "" && ageOrder[currentAge] > ageOrder[def.MaxAge] {
		return fmt.Errorf("%s is no longer available past the %s age", def.Name, def.MaxAge)
	}

	mm.active = &ActiveExpedition{
		Key:       key,
		Name:      def.Name,
		Soldiers:  def.SoldiersNeeded,
		TicksLeft: def.Duration,
	}
	return nil
}

// Tick advances the active expedition by one tick. Returns non-empty rewards
// and a message only on the tick when the expedition resolves.
//
// Success probability: successRoll > (DifficultyBase - militaryBonus×0.3).
// militaryBonus reduces effective difficulty; expeditionBonus scales reward amounts.
// Soldiers are spent at launch (win or lose); success vs failure differs only in
// reward (full vs 30%), not in any extra soldier loss.
func (mm *MilitaryManager) Tick(militaryBonus, expeditionBonus float64) (rewards map[string]float64, message string) {
	if mm.active == nil {
		return nil, ""
	}

	mm.active.TicksLeft--
	if mm.active.TicksLeft > 0 {
		return nil, ""
	}

	// Expedition complete - calculate results
	def := mm.ExpeditionDefByKey(mm.active.Key)
	if def == nil {
		mm.active = nil
		return nil, ""
	}

	// Success calculation: military bonus reduces difficulty
	difficulty := def.DifficultyBase - (militaryBonus * 0.3)
	if difficulty < 0.05 {
		difficulty = 0.05
	}

	successRoll := rand.Float64()
	success := successRoll > difficulty

	rewards = make(map[string]float64)
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
	mm.active = nil
	return rewards, message
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

// launchability reports whether an expedition can be launched right now and, if
// not, a short player-facing reason. It mirrors the engine's LaunchExpedition
// validation order: single-active rule, then soldiers, then each Cost resource.
// soldierCount is the current soldiers amount; resources is the live resource
// map (nil → Cost treated as unaffordable).
func (mm *MilitaryManager) launchability(def ExpeditionDef, soldierCount int, resources map[string]float64) (bool, string) {
	if mm.active != nil {
		return false, "expedition already in progress"
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
	var activeExp *ExpeditionSnapshot
	if mm.active != nil {
		activeExp = &ExpeditionSnapshot{
			Name:      mm.active.Name,
			Soldiers:  mm.active.Soldiers,
			TicksLeft: mm.active.TicksLeft,
		}
	}

	available := mm.GetAvailableExpeditions(currentAge, ageOrder)
	var expList []ExpeditionInfo
	for _, def := range available {
		canLaunch, reason := mm.launchability(def, soldierCount, currentResources)
		expList = append(expList, ExpeditionInfo{
			Name:              def.Name,
			Key:               def.Key,
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
		MilitaryBonus:    militaryBonus,
		ExpeditionBonus:  expeditionBonus,
		ActiveExpedition: activeExp,
		Expeditions:      expList,
		CompletedCount:   mm.completedCount,
		TotalLoot:        loot,
	}
}

// LoadState restores military state from save
func (mm *MilitaryManager) LoadState(active *ActiveExpedition, completedCount int, totalLoot map[string]float64) {
	mm.active = active
	mm.completedCount = completedCount
	if totalLoot != nil {
		mm.totalLoot = totalLoot
	}
}

// GetActiveForSave returns active expedition for saving
func (mm *MilitaryManager) GetActiveForSave() *ActiveExpedition {
	if mm.active == nil {
		return nil
	}
	copy := *mm.active
	return &copy
}
