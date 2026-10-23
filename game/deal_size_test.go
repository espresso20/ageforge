package game

import (
	"math"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// TestDealSizeFollowsTheTownsIncome: what a good stock of a resource comes to
// in a deal is minutes of the town's own income of it. A town making the
// age's typical income gets the age's stock scale (what a store of that size
// held, and the lots were capped at shares of); one making twice as much, twice
// that; a town with none, a quarter of it.
func TestDealSizeFollowsTheTownsIncome(t *testing.T) {
	ge := NewGameEngine()
	ge.age = "industrial_age"
	ge.Resources.UnlockResource("gold")
	typical := ge.rules.TypicalIncome("gold", ge.age)
	if typical <= 0 {
		t.Fatal("no typical gold income in the Industrial Age")
	}
	scale := config.StockScale(ge.age)
	r := ge.Resources.resources["gold"]
	for _, c := range []struct {
		rate, want float64
	}{{typical, scale}, {2 * typical, 2 * scale}, {0, config.EventIncomeFloor * scale}, {100 * typical, config.EventIncomeCeil * scale}} {
		r.Rate = c.rate
		if got := ge.newDealEnv().size["gold"]; math.Abs(got-c.want) > 1e-9*c.want {
			t.Errorf("at a gold income of %v the deal size is %v, want %v", c.rate, got, c.want)
		}
	}
}

// TestFullestIsTheMostHeldAgainstTheTown: a want or favor deal asks for what
// the town holds most of, counted against a good stock of each: the same
// amount is more of the resource whose stock is small.
func TestFullestIsTheMostHeldAgainstTheTown(t *testing.T) {
	env := dealEnv{
		amount: map[string]float64{"a": 100, "b": 100, "c": 100, "d": 5},
		size:   map[string]float64{"a": 1000, "b": 200, "c": 400, "d": 0},
	}
	got := env.fullest([]string{"a", "b", "c", "d"})
	want := []string{"b", "c", "a"}
	if len(got) != len(want) {
		t.Fatalf("fullest = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("fullest = %v, want %v", got, want)
		}
	}
}
