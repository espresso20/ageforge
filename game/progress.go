package game

import (
	"maps"
	"slices"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/rules"
)

// ProgressManager handles age progression
type ProgressManager struct {
	rules      *rules.Set
	ages       []config.AgeDef
	ageIndex   map[string]int
	ageWonders map[string]string // ageKey -> wonder building key (or "")
}

// NewProgressManager creates a new progress manager on the core ruleset.
func NewProgressManager() *ProgressManager { return NewProgressManagerWith(rules.Core()) }

// NewProgressManagerWith creates a new progress manager on set's ages.
func NewProgressManagerWith(set *rules.Set) *ProgressManager {
	pm := &ProgressManager{}
	pm.Rebind(set)
	return pm
}

// Rebind moves the manager onto set: it takes set's ages, their order and
// each age's wonder. The manager holds no state of its own.
func (pm *ProgressManager) Rebind(set *rules.Set) {
	pm.rules = set
	pm.ages = slices.Clone(set.Ages())
	pm.ageIndex = make(map[string]int, len(pm.ages))
	pm.ageWonders = make(map[string]string)
	for i, a := range pm.ages {
		pm.ageIndex[a.Key] = i
		if w := set.Wonder(a.Key); w != "" {
			pm.ageWonders[a.Key] = w
		}
	}
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
	// Require current age's wonder to be built before advancing
	if wonderKey := pm.ageWonders[currentKey]; wonderKey != "" {
		if buildings.GetCount(wonderKey) < 1 {
			return ""
		}
	}
	return nextKey
}

// WonderForAge returns the wonder building key for the given age, or "" if none.
func (pm *ProgressManager) WonderForAge(ageKey string) string {
	return pm.ageWonders[ageKey]
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

// GetRequirementsForNext returns copies of the requirements for the next age
// (they end up in GameState, and the originals are the engine's age table).
func (pm *ProgressManager) GetRequirementsForNext(currentKey string) (map[string]float64, map[string]int) {
	nextKey := pm.GetNextAge(currentKey)
	if nextKey == "" {
		return nil, nil
	}
	nextAge := pm.ages[pm.ageIndex[nextKey]]
	return maps.Clone(nextAge.ResourceReqs), maps.Clone(nextAge.BuildingReqs)
}
