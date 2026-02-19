package game

import "github.com/user/ageforge/config"

// ProgressManager handles age progression
type ProgressManager struct {
	ages     []config.AgeDef
	ageIndex map[string]int
}

// NewProgressManager creates a new progress manager
func NewProgressManager() *ProgressManager {
	ages := config.Ages()
	idx := make(map[string]int)
	for i, a := range ages {
		idx[a.Key] = i
	}
	return &ProgressManager{ages: ages, ageIndex: idx}
}

// GetAgeName returns the display name for an age key
func (pm *ProgressManager) GetAgeName(key string) string {
	if i, ok := pm.ageIndex[key]; ok {
		return pm.ages[i].Name
	}
	return key
}

// GetNextAge returns the next age key, or "" if at max
func (pm *ProgressManager) GetNextAge(currentKey string) string {
	i, ok := pm.ageIndex[currentKey]
	if !ok || i >= len(pm.ages)-1 {
		return ""
	}
	return pm.ages[i+1].Key
}

// CheckAdvancement checks if requirements are met for the next age
func (pm *ProgressManager) CheckAdvancement(currentKey string, resources *ResourceManager, buildings *BuildingManager) string {
	nextKey := pm.GetNextAge(currentKey)
	if nextKey == "" {
		return ""
	}
	nextAge := pm.ages[pm.ageIndex[nextKey]]

	for res, amount := range nextAge.ResourceReqs {
		if resources.Get(res) < amount {
			return ""
		}
	}
	for bld, count := range nextAge.BuildingReqs {
		if buildings.GetCount(bld) < count {
			return ""
		}
	}
	return nextKey
}

// GetUnlocks returns what an age unlocks
func (pm *ProgressManager) GetUnlocks(ageKey string) config.AgeDef {
	if i, ok := pm.ageIndex[ageKey]; ok {
		return pm.ages[i]
	}
	return config.AgeDef{}
}

// GetAgeOrder returns a map of age key -> order index
func (pm *ProgressManager) GetAgeOrder() map[string]int {
	out := make(map[string]int)
	for k, v := range pm.ageIndex {
		out[k] = v
	}
	return out
}

// GetRequirementsForNext returns the requirements for the next age
func (pm *ProgressManager) GetRequirementsForNext(currentKey string) (map[string]float64, map[string]int) {
	nextKey := pm.GetNextAge(currentKey)
	if nextKey == "" {
		return nil, nil
	}
	nextAge := pm.ages[pm.ageIndex[nextKey]]
	return nextAge.ResourceReqs, nextAge.BuildingReqs
}
