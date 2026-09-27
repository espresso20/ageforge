package game

import (
	"slices"

	"github.com/espresso20/ageforge/config"
)

// injectDef adds (or replaces) a building def after construction, keeping the
// manager's sorted key order in step. Production never changes defs once the
// manager is built; tests that fabricate a synthetic building go through here
// so eachBuilt sees it.
func (bm *BuildingManager) injectDef(def config.BuildingDef) {
	if _, exists := bm.defs[def.Key]; !exists {
		i, _ := slices.BinarySearch(bm.order, def.Key)
		bm.order = slices.Insert(bm.order, i, def.Key)
	}
	bm.defs[def.Key] = def
}
