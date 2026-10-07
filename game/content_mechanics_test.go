package game

import (
	"math"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// TestRouteIncomeAndMoraleCeiling: the two mechanic numbers the first
// content batch added, each where the game uses it. Boatbuilding makes a
// route's run bring in 10% more of the listed import, and a harbor's share
// is added to that, not multiplied by it, so each delivers what it says.
// Priesthood lets morale rise 5 points higher.
func TestRouteIncomeAndMoraleCeiling(t *testing.T) {
	run := func(boats bool, harbor float64) (got, listed float64) {
		ge := newSeededEngine(9)
		for _, a := range []string{"stone_age", "bronze_age"} {
			if err := ge.EnterAgeForTest(a); err != nil {
				t.Fatal(err)
			}
		}
		learn(ge, "the_wheel")
		if boats {
			learn(ge, "boatbuilding")
		}
		ge.mu.Lock()
		defer ge.mu.Unlock()
		ge.recalculateRates()
		def := ge.Trade.routeDefs["local_barter"]
		ge.Buildings.counts[def.RequiredBld] = def.MinCount
		for res, r := range ge.Resources.resources {
			ge.Resources.UnlockResource(res)
			r.Storage, r.Amount = 1e9, 1e6
		}
		var res string
		for res, listed = range def.Import {
		}
		before := ge.Resources.Get(res)
		ge.Trade.activeRoutes["local_barter"] = &ActiveRoute{Key: "local_barter", TicksLeft: 1}
		ge.Trade.Tick(ge.Resources, ge.Buildings, nil, harbor)
		// The panels read the same number, harbors aside.
		for _, info := range ge.Trade.Snapshot(ge.age, ge.progress.GetAgeOrder(), ge.Buildings, nil).ActiveRoutes {
			want := listed
			if boats {
				want = float64(listed * 1.10)
			}
			if math.Abs(info.Import[res]-want) > 1e-9 {
				t.Errorf("the panel lists %v %s a run (boats %v), want %v", info.Import[res], res, boats, want)
			}
		}
		return ge.Resources.Get(res) - before, listed
	}
	for _, c := range []struct {
		boats  bool
		harbor float64
		want   float64
	}{{false, 0, 1}, {true, 0, 1.10}, {false, 0.05, 1.05}, {true, 0.05, 1.15}} {
		got, listed := run(c.boats, c.harbor)
		if want := float64(listed * c.want); math.Abs(got-want) > 1e-9 {
			t.Errorf("a run with Boatbuilding %v and a harbor bonus of %v brought in %v, want %v (x%v of the listed %v)", c.boats, c.harbor, got, want, c.want, listed)
		}
	}

	ge := newSeededEngine(9)
	before := ge.moraleCap()
	learn(ge, "priesthood")
	if got := ge.moraleCap(); math.Abs(got-before-0.05) > 1e-12 {
		t.Errorf("the morale ceiling is %v with Priesthood and %v without, want 5 points higher", got, before)
	}
	if got := ge.GetState().MoraleCap; math.Abs(got-before-0.05) > 1e-12 {
		t.Errorf("the snapshot's morale ceiling is %v, want %v", got, before+0.05)
	}
	if def := config.MechanicByKey()[config.MechanicMoraleCap]; def.EffectText(0.05) != "morale can rise 5 points higher" {
		t.Errorf("the effect reads %q", def.EffectText(0.05))
	}
}
