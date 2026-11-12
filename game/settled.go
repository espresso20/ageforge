package game

import "slices"

// settled.go keeps the sky places the player has settled at Quantum. The map
// asks for a place to be settled through SettlePlaces (a command, like any
// other the ui sends); the set rides in the save and in the snapshot.

// SettlePlaces records the given place ids as settled. It reports whether the
// set changed.
func (ge *GameEngine) SettlePlaces(ids []string) bool {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	changed := false
	for _, id := range ids {
		if id == "" || ge.settled[id] {
			continue
		}
		if ge.settled == nil {
			ge.settled = make(map[string]bool)
		}
		ge.settled[id] = true
		changed = true
	}
	return changed
}

func (ge *GameEngine) settledForSave() []string {
	out := make([]string, 0, len(ge.settled))
	for id := range ge.settled {
		out = append(out, id)
	}
	slices.Sort(out)
	if len(out) == 0 {
		return nil
	}
	return out
}

func (ge *GameEngine) loadSettled(ids []string) {
	ge.settled = nil
	for _, id := range ids {
		if ge.settled == nil {
			ge.settled = make(map[string]bool)
		}
		ge.settled[id] = true
	}
}
