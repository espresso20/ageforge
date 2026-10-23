package game

import "github.com/espresso20/ageforge/config"

// buildingExtras is what the engine says of one building for its snapshot:
// the scholars a copy comes with and the most copies that can be built (a
// knowledge building), the soldiers a copy comes with (a military building),
// and, so the panels can show a building its age has unlocked greyed out, the
// research a tech still locks it behind: LockTech and LockTechPath, for the
// buildings whose needsTech names one.
func (ge *GameEngine) buildingExtras(scholars, maxCount, soldiers int, needsTech string) buildingExtra {
	var x buildingExtra
	if scholars > 0 {
		x.scholars, x.scholarCap = scholars, maxCount
	}
	if soldiers > 0 {
		x.soldiers = ge.soldierPerCopy(config.BuildingDef{Soldiers: soldiers})
	}
	if needsTech != "" {
		x.lockTech = ge.techName(needsTech)
		x.lockTechPath = ge.lockTechPath(needsTech)
	}
	return x
}

// lockTechPath is the display names of the prerequisites of tech that are
// not researched yet: nearest first, each once, reaching down through every
// prerequisite that is itself still to come. Empty when tech's own
// prerequisites are all researched.
func (ge *GameEngine) lockTechPath(tech string) []string {
	var path []string
	seen := map[string]bool{tech: true}
	queue := []string{tech}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, p := range ge.Research.defs[cur].Prerequisites {
			if seen[p] || ge.Research.IsResearched(p) {
				continue
			}
			seen[p] = true
			path = append(path, ge.techName(p))
			queue = append(queue, p)
		}
	}
	return path
}
