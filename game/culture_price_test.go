package game

import (
	"math"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/rules"
)

// oldFestivalPrices is what a festival cost the reference town (the culture
// income the age's moderate economy makes) while it was 5% of the culture
// store: the Black Market's was twice that (10%). From the Classical Age,
// the first with a culture income; before it the price is its minimum.
var oldFestivalPrices = map[string]float64{
	"classical_age": 84140, "medieval_age": 609140, "renaissance_age": 2759140, "colonial_age": 13009140,
	"industrial_age": 43009140, "victorian_age": 5.4300914e+08, "electric_age": 6.04300914e+09,
	"atomic_age": 3.604300914e+10, "modern_age": 1.5604300915e+11, "information_age": 9.81043009e+11,
	"digital_age": 3.731043009e+12, "cyberpunk_age": 3.373104301e+13, "fusion_age": 1.66231043e+14,
	"space_age": 8.16231043e+14, "interstellar_age": 2.4316231045e+16, "galactic_age": 2.2181623105e+17,
	"quantum_age": 1.722181623e+19, "transcendent_age": 1.722181623e+19,
}

// TestCulturePricesKeepTheReferenceTownsPrice: the festival and the Black
// Market ask minutes of the town's culture income, worked out so that the
// town making what the age's moderate economy makes pays what a store of the
// age's scale made them cost: the same figure in every age.
func TestCulturePricesKeepTheReferenceTownsPrice(t *testing.T) {
	set := rules.Core()
	for age, want := range oldFestivalPrices {
		typical := set.TypicalIncome("culture", age)
		if typical <= 0 {
			t.Errorf("%s: no typical culture income", age)
			continue
		}
		for _, c := range []struct {
			name  string
			share float64
			want  float64
		}{{"festival", festivalCostShare, want}, {"black market", blackMarketCostShare, 2 * want}} {
			got := set.StockMinutes("culture", age, c.share) * config.TicksPerMinute * typical
			if math.Abs(got-c.want) > 1e-9*c.want {
				t.Errorf("%s: a %s costs the reference town %v, it cost %v", age, c.name, got, c.want)
			}
		}
	}
}

// TestCulturePricesFollowTheTownsIncome: a town making twice the culture pays
// twice, one making none pays the minimum, and Era Mastery's speed is part of
// the income as the rates on screen show it.
func TestCulturePricesFollowTheTownsIncome(t *testing.T) {
	ge := NewGameEngine()
	ge.age = "industrial_age"
	typical := ge.rules.TypicalIncome("culture", ge.age)
	r := ge.Resources.resources["culture"]
	cost := func(rate float64) (festival, bm float64) {
		r.Rate = rate
		return ge.festivalCost(), ge.blackMarketCost()
	}
	f1, b1 := cost(typical)
	if want := oldFestivalPrices[ge.age]; math.Abs(f1-want) > 1e-9*want || math.Abs(b1-2*want) > 1e-9*want {
		t.Fatalf("at the typical income a festival costs %v and a deal %v, want %v and %v", f1, b1, want, 2*want)
	}
	f2, b2 := cost(2 * typical)
	if math.Abs(f2-2*f1) > 1e-9*f1 || math.Abs(b2-2*b1) > 1e-9*b1 {
		t.Errorf("at twice the income a festival costs %v and a deal %v, want %v and %v", f2, b2, 2*f1, 2*b1)
	}
	if f0, b0 := cost(0); f0 != festivalMinCost || b0 != blackMarketMinCost {
		t.Errorf("with no income a festival costs %v and a deal %v, want the minimums %v and %v", f0, b0, festivalMinCost, blackMarketMinCost)
	}
	// An age with no typical culture income yet asks the minimum.
	ge.age = "bronze_age"
	if f, b := cost(typical); f != festivalMinCost || b != blackMarketMinCost {
		t.Errorf("in the Bronze Age a festival costs %v and a deal %v, want the minimums", f, b)
	}
}
