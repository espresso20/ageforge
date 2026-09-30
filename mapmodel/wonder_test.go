package mapmodel_test

import (
	"testing"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel/fixture"
)

// TestWonderHiddenUntilStarted: this age's wonder stays off the maps until
// the player starts it: some of its cost banked, or its construction
// queued. Adam saw the Sacred Grove's plot on the map before he had banked
// a thing.
func TestWonderHiddenUntilStarted(t *testing.T) {
	for _, age := range []string{"primitive_age", "iron_age", "victorian_age"} {
		st := fixture.State(fixture.Options{Age: age, Seed: 7})
		key, name := st.CurrentAgeWonderKey, st.CurrentAgeWonderName
		if key == "" {
			t.Fatalf("%s: the fixture has no current wonder", age)
		}
		shown := func(st game.GameState) bool {
			m := build(t, st, nil)
			inWonders, onPlot := false, false
			for _, w := range m.Wonders {
				inWonders = inWonders || w.Key == key
			}
			for _, w := range m.Town.Wonders {
				onPlot = onPlot || w.Key == key
			}
			if inWonders != onPlot {
				t.Errorf("%s: %s in the wonders %v but on its plot %v", age, key, inWonders, onPlot)
			}
			return inWonders
		}
		bs := st.Buildings[key]
		bs.WonderBank = nil
		st.Buildings[key] = bs
		if shown(st) {
			t.Errorf("%s: %s shows with nothing banked or queued", age, name)
		}
		queued := st
		queued.BuildQueue = append([]game.BuildQueueSnapshot{{Name: name, TicksLeft: 50, TotalTicks: 100}}, st.BuildQueue...)
		if !shown(queued) {
			t.Errorf("%s: %s is under construction but not shown", age, name)
		}
		bs.WonderBank = map[string]float64{"gold": 1}
		st.Buildings[key] = bs
		if !shown(st) {
			t.Errorf("%s: %s has a bank started but is not shown", age, name)
		}
	}
}
