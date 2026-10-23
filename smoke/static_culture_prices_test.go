package smoke

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/rules"
)

// TestCulturePricesStayInBand: in every age that has a culture income the
// festival costs between CulturePriceLowMinutes and CulturePriceHighMinutes of
// the reference town's culture income, and a Black Market deal twice as many,
// by the engine's own prices. Culture comes with the Classical Age.
func TestCulturePricesStayInBand(t *testing.T) {
	problems, rows := StaticCulturePrices()
	for _, p := range problems {
		t.Errorf("%s (%s)", p, p.Rule)
	}
	ages := config.AgeOrder()
	pos := config.AgePositions(ages)
	if want := len(ages) - pos["classical_age"]; len(rows) != want || rows[0].Age != "classical_age" {
		t.Fatalf("%d rows from %s, want %d from classical_age", len(rows), rows[0].Age, want)
	}
	for _, r := range rows {
		if r.Income <= 0 || r.Festival <= 0 || r.FestivalMinute < CulturePriceLowMinutes || r.FestivalMinute > CulturePriceHighMinutes ||
			r.DealMinute < CultureDealTimes*CulturePriceLowMinutes || r.DealMinute > CultureDealTimes*CulturePriceHighMinutes {
			t.Errorf("%+v: want a festival in %v to %v minutes of the reference income and a deal in twice that", r, CulturePriceLowMinutes, CulturePriceHighMinutes)
		}
	}
	var sb strings.Builder
	writeCulturePrices(&sb, problems, rows)
	if out := sb.String(); !strings.Contains(out, "No problems.") || strings.Count(out, "\n| ") != len(rows)+1 {
		t.Errorf("the report section:\n%s", out)
	}
}

// TestCulturePricesCheckCatchesARunaway keeps the guard honest: the price
// as it was (a share of the stock scale, whatever the culture income) must be
// flagged from the Victorian Age on and not before it, a ceiling half again
// as high as the band must fail, and a price that asks next to nothing must
// fail the other way.
func TestCulturePricesCheckCatchesARunaway(t *testing.T) {
	set := rules.Core()
	tb := newRefTables()
	income := make([]float64, len(tb.ages))
	for i := range tb.ages {
		income[i] = tb.state(tb.refTown(i), i, 1).income["culture"]
	}
	flagged := func(prices func(age string) (festival, deal float64)) map[string]string {
		out := map[string]string{}
		problems, _ := culturePriceProblems(tb.ages, income, prices)
		for _, p := range problems {
			if p.String() == "" {
				t.Errorf("a problem with nothing to say: %+v", p)
			}
			out[p.Age+"/"+p.What] = p.Rule
		}
		return out
	}
	pos := config.AgePositions(config.AgeOrder())
	scaleOnly := flagged(func(age string) (float64, float64) {
		return 0.05 * config.StockScale(age), 0.10 * config.StockScale(age)
	})
	for age, at := range pos {
		if set.TypicalIncome("culture", age) <= 0 {
			continue
		}
		got := scaleOnly[age+"/festival"]
		if late := at >= pos["victorian_age"]; late != (got == CulturePriceHigh) {
			t.Errorf("the stock scale alone, %s: festival flagged %q, want %v from the Victorian Age on", age, got, late)
		}
	}
	if scaleOnly["quantum_age/black market deal"] != CulturePriceHigh {
		t.Errorf("the stock scale alone, Quantum Age deal: flagged %q, want %q", scaleOnly["quantum_age/black market deal"], CulturePriceHigh)
	}
	// A ceiling of twice the Colonial Age's price: the late ages go over.
	high := flagged(func(age string) (float64, float64) {
		perMinute := config.TicksPerMinute * set.TypicalIncome("culture", age)
		return 12_000 * perMinute, 24_000 * perMinute
	})
	if high["quantum_age/festival"] != CulturePriceHigh || high["quantum_age/black market deal"] != CulturePriceHigh {
		t.Errorf("a ceiling of 12,000 minutes: Quantum Age flagged %q and %q, want both %q", high["quantum_age/festival"], high["quantum_age/black market deal"], CulturePriceHigh)
	}
	// The minimums alone, as for a town with no income: far under the band.
	low := flagged(func(age string) (float64, float64) { return 2000, 5000 })
	if low["classical_age/festival"] != CulturePriceLow {
		t.Errorf("the minimums alone, Classical Age: festival flagged %q, want %q", low["classical_age/festival"], CulturePriceLow)
	}
	if low["quantum_age/festival"] != CulturePriceLow || low["quantum_age/black market deal"] != CulturePriceLow {
		t.Errorf("the minimums alone, Quantum Age: flagged %q and %q, want both %q", low["quantum_age/festival"], low["quantum_age/black market deal"], CulturePriceLow)
	}
	if got := flagged(func(age string) (float64, float64) {
		return game.CulturePrices(set, age, set.TypicalIncome("culture", age))
	}); len(got) != 0 {
		t.Errorf("the engine's own prices are flagged: %v", got)
	}
}
