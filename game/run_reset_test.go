package game

import "testing"

// stageSoldiers leaves ge with an army the way a run in progress would: some
// soldiers posted at a military building, posts the fallen left, a waiting
// notice already said, and the posted-army count the last rates pass saw.
func stageSoldiers(t *testing.T, ge *GameEngine) {
	t.Helper()
	keys := ge.Buildings.soldierKeys
	if len(keys) == 0 {
		t.Fatal("the rules have no military building to post soldiers at")
	}
	key := keys[0]
	ge.soldiersPosted = map[string]int{key: 3}
	ge.soldiersVacant = map[string]int{key: 1}
	ge.soldierWaitLogged = map[string]bool{key: true}
	ge.soldiersSeen = 4
}

// A new run starts without an army. A prestige and a Succumb both end the run,
// so both clear the soldiers, the posts the fallen left, the "posts waiting on
// food" notices already said and the posted-army count: left behind, a notice
// said in the old run is never said in the new one.
func TestNewRun_ClearsSoldiers(t *testing.T) {
	ends := map[string]func(t *testing.T, ge *GameEngine){
		"succumb": func(t *testing.T, ge *GameEngine) {
			ge.pendingCatastrophe = ge.currentEpoch
			if err := ge.Succumb(); err != nil {
				t.Fatalf("Succumb: %v", err)
			}
		},
		"prestige": func(t *testing.T, ge *GameEngine) {
			ge.mu.Lock()
			defer ge.mu.Unlock()
			ge.completePrestige(prestigePlain)
		},
	}
	for name, end := range ends {
		t.Run(name, func(t *testing.T) {
			ge := catEngine(t, "iron_age", 6)
			stageSoldiers(t, ge)
			end(t, ge)
			if len(ge.soldiersPosted) != 0 {
				t.Errorf("soldiersPosted = %v after the run ended, want none", ge.soldiersPosted)
			}
			if len(ge.soldiersVacant) != 0 {
				t.Errorf("soldiersVacant = %v after the run ended, want none", ge.soldiersVacant)
			}
			if len(ge.soldierWaitLogged) != 0 {
				t.Errorf("soldierWaitLogged = %v after the run ended, want none", ge.soldierWaitLogged)
			}
			if ge.soldiersSeen != 0 {
				t.Errorf("soldiersSeen = %d after the run ended, want 0", ge.soldiersSeen)
			}
		})
	}
}
