package game

import "testing"

// TestPrestige_CapsApplyImmediately: after a prestige, the storage and
// population cap upgrades showed only from the first tick, because the reset
// left derived caps at their base values until recalculateRates ran. They
// must be in the very first snapshot of the new run, and the starting
// resources must not be clamped to the base storage the bonus raises.
func TestPrestige_CapsApplyImmediately(t *testing.T) {
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
	up := run(map[string]int{"storage_bonus": 3, "population_cap": 2, "starting_food": 3})

	// The new run's Primitive Age runs at catch-up speed (Era Mastery), and
	// storage grows with k, the upgrade's share included.
	if got, want := up.Resources["food"].Storage-plain.Resources["food"].Storage, float64(60*up.Mastery.K); got != want {
		t.Errorf("storage_bonus 3: food storage +%v right after prestige, want +%v", got, want)
	}
	if got, want := up.Workers.MaxPop-plain.Workers.MaxPop, 4; got != want {
		t.Errorf("population_cap 2: max pop +%d right after prestige, want +%d", got, want)
	}
	if got, want := up.Resources["food"].Amount-plain.Resources["food"].Amount, 75.0; got != want {
		t.Errorf("starting_food 3: food +%v right after prestige, want +%v (clamped to base storage?)", got, want)
	}
}
