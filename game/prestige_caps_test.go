package game

import "testing"

// TestPrestige_RetiredPerksAreInert: the first shop's perks are retired. A
// save that still holds tiers of them (an old build wrote them; the shop is
// gone and a load ignores them) gets nothing from them: no storage, housing
// or starting resources in the first snapshot of the new run, and the next
// save it writes holds none.
func TestPrestige_RetiredPerksAreInert(t *testing.T) {
	isolateAccountDir(t)
	run := func(name string, upgrades map[string]int) GameState {
		src := newSeededEngine(1)
		src.mu.Lock()
		src.age = "modern_age"
		src.mu.Unlock()
		writeShopSave(t, src, name, func(gs *GameSave) { gs.Prestige.Upgrades = upgrades })
		if upgrades != nil && readSaveFromDisk(t, name).Prestige.Upgrades["storage_bonus"] != 3 {
			t.Fatal("setup: the save does not hold the retired tiers")
		}
		ge := NewGameEngine()
		if err := ge.LoadGame(name); err != nil {
			t.Fatal(err)
		}
		if ge.cheaterBadge {
			t.Fatal("the old signed save failed its signature check")
		}
		if err := ge.SaveGame(name + "-rewritten"); err != nil {
			t.Fatal(err)
		}
		if got := readSaveFromDisk(t, name+"-rewritten").Prestige.Upgrades; got != nil {
			t.Errorf("the rewritten save holds the shop's tiers: %v", got)
		}
		ge.mu.Lock()
		ge.completePrestige(prestigePlain)
		ge.mu.Unlock()
		return ge.GetState()
	}
	plain := run("plain", nil)
	up := run("perks", map[string]int{"storage_bonus": 3, "population_cap": 2, "starting_food": 3, "tick_speed": 5})

	if got := up.Workers.MaxPop - plain.Workers.MaxPop; got != 0 {
		t.Errorf("retired population_cap 2: max pop +%d after prestige, want +0", got)
	}
	if got := up.Resources["food"].Amount - plain.Resources["food"].Amount; got != 0 {
		t.Errorf("retired starting_food 3: food +%v after prestige, want +0", got)
	}
	if got := up.TickSpeedBonus - plain.TickSpeedBonus; got != 0 {
		t.Errorf("retired tick_speed 5: tick speed +%v after prestige, want +0", got)
	}
}
