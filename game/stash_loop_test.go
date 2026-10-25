package game

import (
	"testing"

	"github.com/espresso20/ageforge/config"
)

// The Primitive Age gate, played on the real engine through the same commands
// a player types: a new game, stock topped up to a good stock of the age each
// step, every building and the wonder bought exactly as the gate asks. No
// purchase may be refused for anything but stock (nothing holds a copy back
// for its price against a store), and the advance must go through.
func TestStoneAgeIsReachableBuyingExactlyTheGate(t *testing.T) {
	restore := SetDataDirForTest(t.TempDir())
	defer restore()
	ge := NewGameEngine()
	ge.SeedRNG(1)
	age := config.Ages()[1]
	wonder := ge.progress.WonderForAge(ge.age)
	good := ge.rules.StockScale(ge.age)

	fill := func() {
		ge.mu.Lock()
		for res := range ge.Resources.GetAll() {
			if have := ge.Resources.Get(res); have < good {
				ge.Resources.Add(res, good-have)
			}
		}
		ge.mu.Unlock()
	}
	owned := func(k string) int { return ge.Buildings.GetCount(k) + queued(ge, k) }
	order := []string{"stash", "hut", "story_circle", "gathering_camp", "wood_camp", "shrine"}
	for step := 0; step < 4000; step++ {
		fill()
		for _, k := range order {
			for owned(k) < age.BuildingReqs[k] {
				if err := ge.BuildBuilding(k); err != nil {
					break // waiting for the one in the queue, or for stock
				}
			}
		}
		if wonder != "" && ge.Buildings.GetCount(wonder)+queued(ge, wonder) == 0 {
			for _, res := range []string{"wood", "food"} {
				_, _ = ge.BankWonderMax(wonder, res)
			}
			_ = ge.BuildBuilding(wonder)
		}
		ge.StepTicks(5)
		if err := ge.AdvanceAge(); err == nil {
			if got := ge.GetState().Age; got != age.Key {
				t.Fatalf("advanced to %s, want %s", got, age.Key)
			}
			return
		}
	}
	st := ge.GetState()
	for k, n := range age.BuildingReqs {
		t.Logf("%s: have %d of %d", k, st.Buildings[k].Count, n)
	}
	t.Fatalf("never reached the Stone Age; advance: %v", ge.AdvanceAge())
}
