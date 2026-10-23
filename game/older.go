package game

import (
	"github.com/espresso20/ageforge/config"
)

// Older buildings. A building from an earlier age is an older building when
// its line has a newer rung for the town's current age. It keeps producing; it
// can be upgraded to that rung. Nothing is stored: whether a building is older
// is read from the line and the age every time, so there is no flag to
// forget, no offer to expire and no save field to keep in step.

// UpgradeTarget is the building an older building upgrades to: the current
// rung of its line, however many rungs up. ok is false for a building that is
// not older (the current rung, the last of a line that has paused, or one in
// no line at all: storage, a wonder, a monument, a knowledge building).
func (bm *BuildingManager) UpgradeTarget(key string) (config.BuildingDef, bool) {
	targetKey, ok := bm.UpgradeTargetKey(key)
	if !ok {
		return config.BuildingDef{}, false
	}
	// The manager's own definition: it is the one the prices are worked from.
	if def, has := bm.defs[targetKey]; has {
		return def, true
	}
	target, _ := bm.rules.Building(targetKey)
	return target, true
}

// UpgradeTargetKey is the key of UpgradeTarget's building, without copying the
// definition: what the snapshot of every building asks.
func (bm *BuildingManager) UpgradeTargetKey(key string) (string, bool) {
	if bm.age == nil {
		return "", false
	}
	targetKey, ok := bm.rules.UpgradeKeys(bm.age())[key]
	return targetKey, ok
}

// IsOlder reports whether a building is an older building.
func (bm *BuildingManager) IsOlder(key string) bool {
	_, ok := bm.UpgradeTarget(key)
	return ok
}

// OlderKeys is the older buildings the town holds copies of, sorted.
func (bm *BuildingManager) OlderKeys() []string {
	var out []string
	for _, key := range bm.order {
		if bm.counts[key] > 0 && bm.IsOlder(key) {
			out = append(out, key)
		}
	}
	return out
}

// CountFrom is how many copies the town holds of a building and of every
// later rung of its line. A requirement that names a building by key asks
// for this, so upgrading copies away never takes progress back: an upgraded
// copy is the building it became.
func (bm *BuildingManager) CountFrom(key string) int {
	n := 0
	for _, k := range bm.rules.RungKeysFrom(key) {
		n += bm.counts[k]
	}
	return n
}

// CountFromAny is how many copies the town holds across keys and every later
// rung of their lines, each building counted once.
func (bm *BuildingManager) CountFromAny(keys []string) int {
	seen := make(map[string]bool, len(keys))
	n := 0
	for _, key := range keys {
		for _, k := range bm.rules.RungKeysFrom(key) {
			if !seen[k] {
				seen[k] = true
				n += bm.counts[k]
			}
		}
	}
	return n
}
