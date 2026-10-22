package game

// withLockHints fills LockTech and LockTechPath on every building that its
// age has unlocked and a tech still locks (the buildings NeedsTech names),
// so the panels can show them greyed out with the research they need.
func (ge *GameEngine) withLockHints(bs map[string]BuildingState) map[string]BuildingState {
	for key, b := range bs {
		if !ge.Buildings.AgeUnlocked(key) || !ge.Buildings.TechLocked(key) {
			continue
		}
		tech := ge.Buildings.defs[key].RequiredTech
		b.LockTech = ge.techName(tech)
		b.LockTechPath = ge.lockTechPath(tech)
		bs[key] = b
	}
	return bs
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
