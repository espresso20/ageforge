package game

import (
	"slices"

	"github.com/espresso20/ageforge/config"
)

// buildingExtras is what the engine says of one building for its snapshot:
// the scholars a copy comes with and the copies that research at full rate (a
// knowledge building), the soldiers a copy comes with (a military building),
// and, so the panels can show a building its age has unlocked greyed out, the
// research a tech still locks it behind: LockTech, LockTechPath, and LockTechNext,
// for the buildings whose needsTech names one.
func (ge *GameEngine) buildingExtras(scholars, fullRate, soldiers int, needsTech string) buildingExtra {
	var x buildingExtra
	if scholars > 0 {
		x.scholars, x.fullRate = scholars, fullRate
	}
	if soldiers > 0 {
		x.soldiers = ge.soldierPerCopy(config.BuildingDef{Soldiers: soldiers})
	}
	if needsTech != "" {
		x.lockTech = ge.techName(needsTech)
		x.lockTechPath = ge.lockTechPath(needsTech)
		x.lockTechNext = ge.lockTechNext(needsTech)
	}
	return x
}

// lockTechPath is the display names of the prerequisites of tech that are
// not researched yet: nearest first, each once, reaching down through every
// prerequisite that is itself still to come. Empty when tech's own
// prerequisites are all researched.
func (ge *GameEngine) lockTechPath(tech string) []string {
	// The techs met, in the order they were found: the first is tech itself,
	// and the ones after it are the path. A path is a handful of techs, so a
	// scan of this beats a set, and the buffer stays on the stack.
	var buf [24]string
	found := append(buf[:0], tech)
	for next := 0; next < len(found); next++ {
		for _, p := range ge.Research.defs[found[next]].Prerequisites {
			if slices.Contains(found, p) || ge.Research.IsResearched(p) {
				continue
			}
			found = append(found, p)
		}
	}
	if len(found) == 1 {
		return nil
	}
	path := make([]string, len(found)-1)
	for i, key := range found[1:] {
		path[i] = ge.techName(key)
	}
	return path
}

// lockTechNext is the display names of the techs on the path from tech whose
// own prerequisites are all researched: the techs a player can research right
// now to progress toward tech. Returns them in the order they appear in the
// path, up to 3; empty when no techs on the path are immediately researchable.
func (ge *GameEngine) lockTechNext(tech string) []string {
	path := ge.lockTechPath(tech)
	if len(path) == 0 {
		return nil
	}
	// Walk the path (which has display names). We need to check prerequisites,
	// so translate back to keys. Build a map of display name -> key for the ones
	// on the path.
	pathKeys := []string{tech}
	for next := 0; next < len(pathKeys); next++ {
		for _, p := range ge.Research.defs[pathKeys[next]].Prerequisites {
			if slices.Contains(pathKeys, p) || ge.Research.IsResearched(p) {
				continue
			}
			pathKeys = append(pathKeys, p)
		}
	}
	// Now collect the techs whose prerequisites are all researched.
	var next []string
	for _, key := range pathKeys[1:] {
		// A tech can be researched if all its prerequisites are researched
		canResearch := true
		for _, p := range ge.Research.defs[key].Prerequisites {
			if !ge.Research.IsResearched(p) {
				canResearch = false
				break
			}
		}
		if canResearch {
			next = append(next, ge.techName(key))
		}
	}
	return next
}
