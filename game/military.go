package game

import (
	"fmt"
	"math/rand"
)

// ExpeditionDef defines an available expedition
type ExpeditionDef struct {
	Name           string
	Key            string
	MinAge         string
	SoldiersNeeded int
	Duration       int     // ticks
	DifficultyBase float64 // 0.0 - 1.0, higher = harder
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

// MilitaryManager handles soldiers, defense, and expeditions
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

// LaunchExpedition starts an expedition, consuming soldiers
func (mm *MilitaryManager) LaunchExpedition(key string, soldierCount int, currentAge string, ageOrder map[string]int, militaryBonus float64) error {
	if mm.active != nil {
		return fmt.Errorf("expedition '%s' already in progress (%d ticks left)", mm.active.Name, mm.active.TicksLeft)
	}

	var def *ExpeditionDef
	for i := range mm.expeditions {
		if mm.expeditions[i].Key == key {
			def = &mm.expeditions[i]
			break
		}
	}
	if def == nil {
		return fmt.Errorf("unknown expedition: %s", key)
	}

	if ageOrder[def.MinAge] > ageOrder[currentAge] {
		return fmt.Errorf("%s requires %s age", def.Name, def.MinAge)
	}

	if soldierCount < def.SoldiersNeeded {
		return fmt.Errorf("%s needs %d soldiers (have: %d)", def.Name, def.SoldiersNeeded, soldierCount)
	}

	mm.active = &ActiveExpedition{
		Key:       key,
		Name:      def.Name,
		Soldiers:  def.SoldiersNeeded,
		TicksLeft: def.Duration,
	}
	return nil
}

// Tick processes expedition progress. Returns (rewards, message, soldiers_lost) if completed.
func (mm *MilitaryManager) Tick(militaryBonus, expeditionBonus float64) (rewards map[string]float64, message string, soldiersLost int) {
	if mm.active == nil {
		return nil, "", 0
	}

	mm.active.TicksLeft--
	if mm.active.TicksLeft > 0 {
		return nil, "", 0
	}

	// Expedition complete - calculate results
	var def *ExpeditionDef
	for i := range mm.expeditions {
		if mm.expeditions[i].Key == mm.active.Key {
			def = &mm.expeditions[i]
			break
		}
	}
	if def == nil {
		mm.active = nil
		return nil, "", 0
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

		// Small chance to lose soldiers even on success
		if rand.Float64() < difficulty*0.3 {
			soldiersLost = 1
			message += " (1 soldier lost)"
		}
	} else {
		// Partial rewards on failure
		for res, amount := range def.Rewards {
			partial := amount * 0.3
			rewards[res] = partial
			mm.totalLoot[res] += partial
		}
		soldiersLost = 1 + rand.Intn(2)
		if soldiersLost > mm.active.Soldiers {
			soldiersLost = mm.active.Soldiers
		}
		message = fmt.Sprintf("%s failed! Partial loot recovered. Lost %d soldier(s).", def.Name, soldiersLost)
	}

	mm.completedCount++
	mm.active = nil
	return rewards, message, soldiersLost
}

// GetAvailableExpeditions returns expeditions available for the current age
func (mm *MilitaryManager) GetAvailableExpeditions(currentAge string, ageOrder map[string]int) []ExpeditionDef {
	var available []ExpeditionDef
	for _, def := range mm.expeditions {
		if ageOrder[def.MinAge] <= ageOrder[currentAge] {
			available = append(available, def)
		}
	}
	return available
}

// CalculateDefense calculates defense rating from soldiers and bonuses
func (mm *MilitaryManager) CalculateDefense(soldierCount int, militaryBonus float64) float64 {
	base := float64(soldierCount) * 2.0
	return base * (1.0 + militaryBonus)
}

// Snapshot returns military state for UI
func (mm *MilitaryManager) Snapshot(currentAge string, ageOrder map[string]int, soldierCount int, militaryBonus, expeditionBonus float64) MilitaryState {
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
		expList = append(expList, ExpeditionInfo{
			Name:           def.Name,
			Key:            def.Key,
			SoldiersNeeded: def.SoldiersNeeded,
			Duration:       def.Duration,
			Difficulty:     def.DifficultyBase,
			Description:    def.Description,
			CanLaunch:      soldierCount >= def.SoldiersNeeded && mm.active == nil,
		})
	}

	loot := make(map[string]float64)
	for k, v := range mm.totalLoot {
		loot[k] = v
	}

	return MilitaryState{
		SoldierCount:    soldierCount,
		DefenseRating:   mm.CalculateDefense(soldierCount, militaryBonus),
		MilitaryBonus:   militaryBonus,
		ExpeditionBonus: expeditionBonus,
		ActiveExpedition: activeExp,
		Expeditions:     expList,
		CompletedCount:  mm.completedCount,
		TotalLoot:       loot,
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
