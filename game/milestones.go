package game

import (
	"github.com/user/ageforge/config"
)

// MilestoneManager tracks and checks milestones
type MilestoneManager struct {
	defs      []config.MilestoneDef
	completed map[string]bool
}

// NewMilestoneManager creates a new milestone manager
func NewMilestoneManager() *MilestoneManager {
	return &MilestoneManager{
		defs:      config.Milestones(),
		completed: make(map[string]bool),
	}
}

// CheckMilestones checks all milestones against current state.
// Returns list of newly completed milestones.
func (mm *MilestoneManager) CheckMilestones(
	tick int,
	age string,
	ageOrder map[string]int,
	resources *ResourceManager,
	buildings *BuildingManager,
	population int,
	techCount int,
	totalBuilt int,
	researchedTechs map[string]bool,
	soldierCount int,
	wonderCount int,
) []config.MilestoneDef {
	var completed []config.MilestoneDef

	for _, def := range mm.defs {
		if mm.completed[def.Key] {
			continue
		}

		if mm.checkMilestone(def, tick, age, ageOrder, resources, buildings, population, techCount, totalBuilt, researchedTechs, soldierCount, wonderCount) {
			mm.completed[def.Key] = true
			completed = append(completed, def)
		}
	}

	return completed
}

func (mm *MilestoneManager) checkMilestone(
	def config.MilestoneDef,
	tick int,
	age string,
	ageOrder map[string]int,
	resources *ResourceManager,
	buildings *BuildingManager,
	population int,
	techCount int,
	totalBuilt int,
	researchedTechs map[string]bool,
	soldierCount int,
	wonderCount int,
) bool {
	// Check min tick
	if def.MinTick > 0 && tick < def.MinTick {
		return false
	}

	// Check age
	if def.MinAge != "" {
		if ageOrder[age] < ageOrder[def.MinAge] {
			return false
		}
	}

	// Check resources
	for res, required := range def.MinResources {
		if resources.Get(res) < required {
			return false
		}
	}

	// Check buildings
	for bld, required := range def.MinBuildings {
		if buildings.GetCount(bld) < required {
			return false
		}
	}

	// Check population
	if def.MinPopulation > 0 && population < def.MinPopulation {
		return false
	}

	// Check tech count
	if def.MinTechCount > 0 && techCount < def.MinTechCount {
		return false
	}

	// Check specific techs
	for _, tech := range def.RequiredTechs {
		if !researchedTechs[tech] {
			return false
		}
	}

	// Special checks based on milestone key
	switch def.Key {
	case "master_builder":
		if totalBuilt < 20 {
			return false
		}
	case "war_machine":
		if soldierCount < 10 {
			return false
		}
	case "wonder_builder":
		if wonderCount < 1 {
			return false
		}
	case "scholars_haven":
		// This checks for 5+ scholars — we need villager data
		// For simplicity, use population >= 10 as proxy (already set in def)
	}

	return true
}

// IsCompleted checks if a milestone has been achieved
func (mm *MilestoneManager) IsCompleted(key string) bool {
	return mm.completed[key]
}

// CompletedCount returns how many milestones are completed
func (mm *MilestoneManager) CompletedCount() int {
	return len(mm.completed)
}

// GetCompleted returns all completed milestone keys
func (mm *MilestoneManager) GetCompleted() []string {
	var keys []string
	for k := range mm.completed {
		keys = append(keys, k)
	}
	return keys
}

// Snapshot returns milestone state for UI
func (mm *MilestoneManager) Snapshot() MilestoneState {
	milestones := make(map[string]MilestoneInfo)
	for _, def := range mm.defs {
		milestones[def.Key] = MilestoneInfo{
			Name:        def.Name,
			Description: def.Description,
			Completed:   mm.completed[def.Key],
		}
	}
	return MilestoneState{
		Milestones:     milestones,
		CompletedCount: len(mm.completed),
		TotalCount:     len(mm.defs),
	}
}

// LoadState restores milestone state from save
func (mm *MilestoneManager) LoadState(completed []string) {
	mm.completed = make(map[string]bool)
	for _, key := range completed {
		mm.completed[key] = true
	}
}
