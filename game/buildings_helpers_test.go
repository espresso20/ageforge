package game

import "github.com/espresso20/ageforge/config"

// injectDef adds (or replaces) a building def after construction, keeping the
// manager's sorted key order in step. Production never changes defs once the
// manager is built; tests that fabricate a synthetic building go through here
// so eachBuilt sees it.
func (bm *BuildingManager) injectDef(def config.BuildingDef) {
	bm.defs[def.Key] = def
	bm.rebuildOrder()
}
