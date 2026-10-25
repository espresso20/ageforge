package game

import (
	"math"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/rules"
)

// referencePrices is what a festival and a Black Market deal cost the
// reference town, the one making the age's typical culture income: their share
// of the age's stock scale (which is that many minutes of the typical income),
// held to the ceiling in minutes of it (festivalMaxMinutes,
// blackMarketMaxMinutes), and never under the minimum.
func referencePrices(age string, typical float64) (festival, deal float64) {
	capF := float64(festivalMaxMinutes*config.TicksPerMinute) * typical
	capB := float64(blackMarketMaxMinutes*config.TicksPerMinute) * typical
	scale := config.StockScale(age)
	festival = max(festivalMinCost, min(festivalCostShare*scale, capF))
	deal = max(blackMarketMinCost, min(blackMarketCostShare*scale, capB))
	return festival, deal
}

// TestCulturePricesAreHoursOfTheTownsIncome: the festival and the Black Market
// ask minutes of the town's culture income, worked out so that the town making
// what the age's moderate economy makes pays its share of the age's stock
// scale, up to the ceiling in minutes of that income, in every age with a
// culture income.
func TestCulturePricesAreHoursOfTheTownsIncome(t *testing.T) {
	set := rules.Core()
	checked := 0
	for _, age := range config.AgeOrder() {
		typical := set.TypicalIncome("culture", age)
		if typical <= 0 {
			continue
		}
		checked++
		festival, deal := CulturePrices(set, age, typical)
		wantF, wantB := referencePrices(age, typical)
		if math.Abs(festival-wantF) > 1e-9*wantF || math.Abs(deal-wantB) > 1e-9*wantB {
			t.Errorf("%s: a festival costs the reference town %v and a deal %v, want %v and %v", age, festival, deal, wantF, wantB)
		}
		// Never more than hours of the town's own income (the minimum aside).
		if capF := float64(festivalMaxMinutes*config.TicksPerMinute) * typical; festival > max(festivalMinCost, capF) {
			t.Errorf("%s: a festival costs %v, over its ceiling of %v", age, festival, capF)
		}
		if capB := float64(blackMarketMaxMinutes*config.TicksPerMinute) * typical; deal > max(blackMarketMinCost, capB) {
			t.Errorf("%s: a deal costs %v, over its ceiling of %v", age, deal, capB)
		}
	}
	if checked == 0 {
		t.Fatal("no age has a typical culture income")
	}
}

// TestCulturePricesFollowTheTownsIncome: a town making twice the culture pays
// twice, one making none pays the minimum, and Era Mastery's speed is part of
// the income as the rates on screen show it. In the Quantum Age, where the
// ceiling holds, a town making the typical income pays 100 hours of it for a
// festival and 200 for a deal.
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
	if wantF, wantB := referencePrices(ge.age, typical); math.Abs(f1-wantF) > 1e-9*wantF || math.Abs(b1-wantB) > 1e-9*wantB {
		t.Fatalf("at the typical income a festival costs %v and a deal %v, want %v and %v", f1, b1, wantF, wantB)
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
	// Where the ceiling holds the minutes are the ceiling's, and the price
	// still follows the town.
	ge.age = "quantum_age"
	typical = ge.rules.TypicalIncome("culture", ge.age)
	f1, b1 = cost(typical)
	wantF, wantB := float64(festivalMaxMinutes*config.TicksPerMinute)*typical, float64(blackMarketMaxMinutes*config.TicksPerMinute)*typical
	if math.Abs(f1-wantF) > 1e-9*wantF || math.Abs(b1-wantB) > 1e-9*wantB {
		t.Fatalf("in the Quantum Age at the typical income a festival costs %v and a deal %v, want %v and %v", f1, b1, wantF, wantB)
	}
	if f2, b2 = cost(2 * typical); math.Abs(f2-2*f1) > 1e-9*f1 || math.Abs(b2-2*b1) > 1e-9*b1 {
		t.Errorf("in the Quantum Age at twice the income a festival costs %v and a deal %v, want %v and %v", f2, b2, 2*f1, 2*b1)
	}
}
