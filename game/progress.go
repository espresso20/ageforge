package game

import (
	"maps"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/rules"
)

// ProgressManager handles age progression
type ProgressManager struct {
	rules      *rules.Set
	ages       []config.AgeDef
	ageIndex   map[string]int
	ageWonders map[string]string // ageKey -> wonder building key (or "")
	// town is what the measured gate reads of the town beyond its buildings
	// (gate_v2.go); the engine fills it in. Today's gate never reads it.
	town gateTown
}

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
	pm.ages = set.Ages()
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

// CheckAdvancement checks if requirements are met for the next age: the
// buildings its gate asks for and the current age's wonder. No gate asks for
// an amount of a resource, and advancing spends none. Under the measured gate
// (AGEFORGE_GATES=v2) the buildings are replaced by the town's population,
// the income of the age's main materials and its key techs (gate_v2.go).
func (pm *ProgressManager) CheckAdvancement(currentKey string, buildings *BuildingManager) string {
	nextKey := pm.GetNextAge(currentKey)
	if nextKey == "" {
		return ""
	}
	if m, ok := pm.measureFor(currentKey); ok {
		return pm.checkMeasured(currentKey, nextKey, m, buildings)
	}
	nextAge := pm.ages[pm.ageIndex[nextKey]]

	for bld, count := range nextAge.BuildingReqs {
		// An upgraded copy counts as the building it became, and a gate that
		// asks for an older rung counts the later ones of its line.
		if buildings.CountFrom(bld) < count {
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
	return maps.Clone(pm.ageIndex) // never nil: Rebind makes it
}

// AgeIndex returns the order index for an age key without copying the age map.
func (pm *ProgressManager) AgeIndex(key string) (int, bool) {
	i, ok := pm.ageIndex[key]
	return i, ok
}

// ageOrder returns the read-only age index for internal callers that need the
// full map. External callers must use GetAgeOrder, which protects this map.
func (pm *ProgressManager) ageOrder() map[string]int {
	return pm.ageIndex
}

// GetRequirementsForNext returns a copy of the building counts the gate into
// the next age asks for (it ends up in GameState, and the original is the
// engine's age table).
func (pm *ProgressManager) GetRequirementsForNext(currentKey string) map[string]int {
	nextKey := pm.GetNextAge(currentKey)
	if nextKey == "" {
		return nil
	}
	if _, ok := pm.measureFor(currentKey); ok {
		return nil // the measured gate asks for no building counts
	}
	return maps.Clone(pm.ages[pm.ageIndex[nextKey]].BuildingReqs)
}
