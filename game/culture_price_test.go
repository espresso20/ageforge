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
// the first with a culture income; before it the price is its minimum. They
// are the shares of the stock scale the prices are made of; the ages the
// prices were tuned in (to the Industrial Age) still cost exactly this, and
// the later ones are held to a ceiling (festivalMaxMinutes).
var oldFestivalPrices = map[string]float64{
	"classical_age": 84140, "medieval_age": 609140, "renaissance_age": 2759140, "colonial_age": 13009140,
	"industrial_age": 43009140, "victorian_age": 543.00914 * config.Million, "electric_age": 6.04300914 * config.Billion,
	"atomic_age": 36.04300914 * config.Billion, "modern_age": 156.04300915 * config.Billion, "information_age": 981.043009 * config.Billion,
	"digital_age": 3.731043009 * config.Trillion, "cyberpunk_age": 33.73104301 * config.Trillion, "fusion_age": 166.231043 * config.Trillion,
	"space_age": 816.231043 * config.Trillion, "interstellar_age": 24.316231045 * config.Quadrillion, "galactic_age": 221.81623105 * config.Quadrillion,
	"quantum_age": 17.22181623 * config.Quintillion, "transcendent_age": 17.22181623 * config.Quintillion,
}

// TestCulturePricesKeepTheReferenceTownsPrice: the festival and the Black
// Market ask minutes of the town's culture income, worked out so that the
// town making what the age's moderate economy makes pays what a store of the
// age's scale made them cost, up to the ceiling in minutes: the same figure
// in every age up to the Industrial Age, which are the ages the prices were
// tuned in, and the ceiling from the Victorian Age on, where the scale has
// outrun any culture income.
func TestCulturePricesKeepTheReferenceTownsPrice(t *testing.T) {
	set := rules.Core()
	pos := config.AgePositions(config.AgeOrder())
	for age, want := range oldFestivalPrices {
		typical := set.TypicalIncome("culture", age)
		if typical <= 0 {
			t.Errorf("%s: no typical culture income", age)
			continue
		}
		festival, deal := CulturePrices(set, age, typical)
		capF := float64(festivalMaxMinutes*config.TicksPerMinute) * typical
		capB := float64(blackMarketMaxMinutes*config.TicksPerMinute) * typical
		wantF, wantB := math.Min(want, capF), math.Min(2*want, capB)
		if math.Abs(festival-wantF) > 1e-9*wantF || math.Abs(deal-wantB) > 1e-9*wantB {
			t.Errorf("%s: a festival costs the reference town %v and a deal %v, want %v and %v", age, festival, deal, wantF, wantB)
		}
		// The ceiling holds the festival from the Victorian Age on, and never
		// before it.
		if held := want > capF; held != (pos[age] >= pos["victorian_age"]) {
			t.Errorf("%s: festival held to its ceiling is %v, want it so from the Victorian Age on only", age, held)
		}
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
