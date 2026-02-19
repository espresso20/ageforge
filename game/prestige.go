package game

import (
	"fmt"
	"math"

	"github.com/user/ageforge/config"
)

// PrestigeManager manages prestige level, points, and upgrades
type PrestigeManager struct {
	level      int
	totalEarned int
	available  int
	upgrades   map[string]int // key -> tier purchased
}

// NewPrestigeManager creates a new prestige manager
func NewPrestigeManager() *PrestigeManager {
	return &PrestigeManager{
		upgrades: make(map[string]int),
	}
}

// CalculatePoints computes prestige points for the current run
func (pm *PrestigeManager) CalculatePoints(age string, ageOrder map[string]int, milestonesCompleted, techsResearched, totalBuilt int) int {
	// Base: 1 point per age beyond Primitive
	ageIdx, ok := ageOrder[age]
	if !ok {
		return 0
	}
	base := float64(ageIdx) // primitive=0, stone=1, ..., medieval=4

	// Bonus: +1 per 10 milestones, +1 per 15 techs, +1 per 50 buildings
	bonus := float64(milestonesCompleted/10) + float64(techsResearched/15) + float64(totalBuilt/50)

	total := base + bonus

	// Diminishing returns: divide by sqrt(totalPrestigeLevel + 1)
	divisor := math.Sqrt(float64(pm.level + 1))
	points := int(total / divisor)

	if points < 1 && base >= 4 {
		points = 1 // minimum 1 point if you've reached Medieval+
	}
	return points
}

// CanPrestige returns true if the player can prestige (Medieval Age or later)
func (pm *PrestigeManager) CanPrestige(age string, ageOrder map[string]int) bool {
	idx, ok := ageOrder[age]
	if !ok {
		return false
	}
	// Medieval Age is order 4
	return idx >= 4
}

// Prestige increments level and adds points
func (pm *PrestigeManager) Prestige(points int) {
	pm.level++
	pm.totalEarned += points
	pm.available += points
}

// BuyUpgrade purchases the next tier of an upgrade. Returns error if can't afford or maxed.
func (pm *PrestigeManager) BuyUpgrade(key string) error {
	defs := config.PrestigeUpgradeByKey()
	def, ok := defs[key]
	if !ok {
		return fmt.Errorf("unknown prestige upgrade: %s", key)
	}

	currentTier := pm.upgrades[key]
	if currentTier >= def.MaxTier {
		return fmt.Errorf("%s is already at max tier (%d)", def.Name, def.MaxTier)
	}

	cost := def.Costs[currentTier]
	if pm.available < cost {
		return fmt.Errorf("need %d prestige points for %s tier %d (have %d)", cost, def.Name, currentTier+1, pm.available)
	}

	pm.available -= cost
	pm.upgrades[key] = currentTier + 1
	return nil
}

// GetBonuses returns all prestige bonuses (passive + upgrades) as a bonus map
func (pm *PrestigeManager) GetBonuses() map[string]float64 {
	bonuses := make(map[string]float64)

	// Passive bonus: +2% production_all per prestige level
	if pm.level > 0 {
		bonuses["production_all"] = float64(pm.level) * 0.02
	}

	// Upgrade bonuses (rate and flat bonuses, not starting resources)
	defs := config.PrestigeUpgradeByKey()
	for key, tier := range pm.upgrades {
		if tier <= 0 {
			continue
		}
		def, ok := defs[key]
		if !ok {
			continue
		}
		if def.EffectType == "rate_bonus" || def.EffectType == "flat_bonus" {
			bonuses[def.EffectKey] += def.PerTier * float64(tier)
		}
	}

	return bonuses
}

// GetStartingResources returns bonus starting resources from prestige upgrades
func (pm *PrestigeManager) GetStartingResources() map[string]float64 {
	resources := make(map[string]float64)
	defs := config.PrestigeUpgradeByKey()
	for key, tier := range pm.upgrades {
		if tier <= 0 {
			continue
		}
		def, ok := defs[key]
		if !ok {
			continue
		}
		if def.EffectType == "starting_resource" {
			resources[def.EffectKey] += def.PerTier * float64(tier)
		}
	}
	return resources
}

// Snapshot returns a PrestigeState for UI consumption
func (pm *PrestigeManager) Snapshot() PrestigeState {
	defs := config.PrestigeUpgradeByKey()
	upgrades := make(map[string]PrestigeUpgradeState)

	for _, def := range config.PrestigeUpgrades() {
		tier := pm.upgrades[def.Key]
		nextCost := 0
		if tier < def.MaxTier {
			nextCost = def.Costs[tier]
		}
		upgrades[def.Key] = PrestigeUpgradeState{
			Name:        def.Name,
			Description: def.Description,
			Tier:        tier,
			MaxTier:     def.MaxTier,
			NextCost:    nextCost,
			Effect:      formatPrestigeEffect(def, tier),
		}
	}
	_ = defs // used via config.PrestigeUpgrades()

	passiveBonus := float64(pm.level) * 0.02

	return PrestigeState{
		Level:        pm.level,
		TotalEarned:  pm.totalEarned,
		Available:    pm.available,
		Upgrades:     upgrades,
		PassiveBonus: passiveBonus,
	}
}

// LoadState restores prestige state from save data
func (pm *PrestigeManager) LoadState(level, totalEarned, available int, upgrades map[string]int) {
	pm.level = level
	pm.totalEarned = totalEarned
	pm.available = available
	if upgrades != nil {
		pm.upgrades = upgrades
	}
}

// GetLevel returns the current prestige level
func (pm *PrestigeManager) GetLevel() int {
	return pm.level
}

func formatPrestigeEffect(def config.PrestigeUpgradeDef, tier int) string {
	if tier == 0 {
		return "Not purchased"
	}
	switch def.EffectType {
	case "rate_bonus":
		return fmt.Sprintf("+%.0f%% %s", def.PerTier*float64(tier)*100, def.EffectKey)
	case "flat_bonus":
		return fmt.Sprintf("+%.0f %s", def.PerTier*float64(tier), def.EffectKey)
	case "starting_resource":
		return fmt.Sprintf("+%.0f starting %s", def.PerTier*float64(tier), def.EffectKey)
	}
	return ""
}
