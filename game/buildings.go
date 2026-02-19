package game

import (
	"math"

	"github.com/user/ageforge/config"
)

// BuildingManager manages all buildings
type BuildingManager struct {
	counts   map[string]int
	defs     map[string]config.BuildingDef
	unlocked map[string]bool
}

// NewBuildingManager creates a building manager
func NewBuildingManager() *BuildingManager {
	return &BuildingManager{
		counts:   make(map[string]int),
		defs:     config.BuildingByKey(),
		unlocked: make(map[string]bool),
	}
}

// UnlockBuilding makes a building available
func (bm *BuildingManager) UnlockBuilding(key string) {
	bm.unlocked[key] = true
}

// IsUnlocked returns whether a building type is available
func (bm *BuildingManager) IsUnlocked(key string) bool {
	return bm.unlocked[key]
}

// GetCount returns how many of a building exist
func (bm *BuildingManager) GetCount(key string) int {
	return bm.counts[key]
}

// GetCost calculates the cost for the next building of this type (with scaling)
func (bm *BuildingManager) GetCost(key string) map[string]float64 {
	def, ok := bm.defs[key]
	if !ok {
		return nil
	}
	count := bm.counts[key]
	cost := make(map[string]float64)
	for res, base := range def.BaseCost {
		cost[res] = math.Floor(base * math.Pow(def.CostScale, float64(count)))
	}
	return cost
}

// Build constructs a building. Returns false if can't afford or not unlocked.
func (bm *BuildingManager) Build(key string, resources *ResourceManager) bool {
	if !bm.unlocked[key] {
		return false
	}
	def, ok := bm.defs[key]
	if !ok {
		return false
	}
	if def.MaxCount > 0 && bm.counts[key] >= def.MaxCount {
		return false
	}
	cost := bm.GetCost(key)
	if !resources.Pay(cost) {
		return false
	}
	bm.counts[key]++
	return true
}

// GetEffects returns the total effects from all built buildings
func (bm *BuildingManager) GetEffects() []config.Effect {
	var effects []config.Effect
	for key, count := range bm.counts {
		if count == 0 {
			continue
		}
		def := bm.defs[key]
		for _, eff := range def.Effects {
			scaled := config.Effect{
				Type:   eff.Type,
				Target: eff.Target,
				Value:  eff.Value * float64(count),
			}
			effects = append(effects, scaled)
		}
	}
	return effects
}

// GetPopCapacity returns total population capacity from housing buildings
func (bm *BuildingManager) GetPopCapacity() int {
	cap := 0
	for key, count := range bm.counts {
		def := bm.defs[key]
		for _, eff := range def.Effects {
			if eff.Type == "capacity" && eff.Target == "population" {
				cap += int(eff.Value) * count
			}
		}
	}
	return cap
}

// GetStorageBonuses returns per-resource storage bonuses from buildings
// "all" key means it applies to every resource
func (bm *BuildingManager) GetStorageBonuses() map[string]float64 {
	bonuses := make(map[string]float64)
	for key, count := range bm.counts {
		def := bm.defs[key]
		for _, eff := range def.Effects {
			if eff.Type == "storage" {
				bonuses[eff.Target] += eff.Value * float64(count)
			}
		}
	}
	return bonuses
}

// GetAll returns building counts (for save)
func (bm *BuildingManager) GetAll() map[string]int {
	out := make(map[string]int)
	for key, count := range bm.counts {
		out[key] = count
	}
	return out
}

// LoadCounts restores building counts from save data
func (bm *BuildingManager) LoadCounts(counts map[string]int) {
	for key, count := range counts {
		bm.counts[key] = count
	}
}

// Snapshot returns building states for UI
func (bm *BuildingManager) Snapshot(resources *ResourceManager) map[string]BuildingState {
	out := make(map[string]BuildingState)
	for key, def := range bm.defs {
		cost := bm.GetCost(key)
		out[key] = BuildingState{
			Count:       bm.counts[key],
			Name:        def.Name,
			Category:    def.Category,
			Description: def.Description,
			Unlocked:    bm.unlocked[key],
			NextCost:    cost,
			CanBuild:    bm.unlocked[key] && resources.CanAfford(cost),
		}
	}
	return out
}
