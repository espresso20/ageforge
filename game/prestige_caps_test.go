package game

import "testing"

// TestPrestige_RetiredPerksAreInert: the first shop's perks are retired (shop
// version 2). A save that still holds tiers of them (a hand edit; the refund
// zeroes them on load) gets nothing from them: no storage, housing or
// starting resources in the first snapshot of the new run.
func TestPrestige_RetiredPerksAreInert(t *testing.T) {
	isolateAccountDir(t)
	run := func(upgrades map[string]int) GameState {
		ge := newSeededEngine(1)
		ge.mu.Lock()
		ge.age = "modern_age"
		ge.Prestige.LoadState(0, 0, 0, upgrades)
		ge.completePrestige(prestigePlain)
		ge.mu.Unlock()
		return ge.GetState()
	}
	plain := run(nil)
	up := run(map[string]int{"storage_bonus": 3, "population_cap": 2, "starting_food": 3, "tick_speed": 5})

	if got := up.Resources["food"].Storage - plain.Resources["food"].Storage; got != 0 {
		t.Errorf("retired storage_bonus 3: food storage +%v after prestige, want +0", got)
	}
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
