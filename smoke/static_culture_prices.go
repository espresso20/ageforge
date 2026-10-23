package smoke

import (
	"fmt"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/rules"
)

// The culture price check: what a festival and a Black Market deal cost the
// town an age's rules are written for, read against the culture the
// reference model's town makes there. The prices are minutes of the town's own
// culture income, so a price that keeps in step with income costs about the
// same number of minutes of it in every age. The check works out the price
// for a town making the age's typical culture income (the town the rule is
// tuned on, game.CulturePrices: the engine's own figure), and divides it by
// the reference model's culture income (smoke/reference.go), a second
// measure of the same town that the rule does not read.
//
// Every age that has a culture income must keep a festival between
// CulturePriceLowMinutes and CulturePriceHighMinutes of it, and a Black Market
// deal between twice each (the deal asks twice the share). The bounds come
// from the fixed curve of the ages the prices were tuned in and played. On
// the reference town the festival costs 552 minutes of culture income in the
// Classical Age, the cheapest, and 4,233 in the Colonial Age, the dearest of
// them; the ages after sit between 3,300 and 5,100. Low is three quarters of
// the cheapest and high a half more than the dearest, so the tuned ages pass
// as they are and the check has room for the reference town reading a little
// differently from the typical one, but not for a price that runs away: the
// stock scale left alone asked 1.3e11 minutes in the Quantum Age, and any
// ceiling more than half again the Colonial Age's fails.
const (
	CulturePriceLowMinutes  = 400.0
	CulturePriceHighMinutes = 6500.0
	// CultureDealTimes is how many times a festival's minutes a Black Market
	// deal asks: its share of the scale is twice the festival's.
	CultureDealTimes = 2.0
)

// The rules a CulturePriceProblem can break.
const (
	// CulturePriceHigh: a festival or a deal asks more minutes of culture
	// income than the band.
	CulturePriceHigh = "high"
	// CulturePriceLow: it asks fewer.
	CulturePriceLow = "low"
)

// CulturePriceProblem is one price outside its band.
type CulturePriceProblem struct {
	Age     string  `json:"age"`
	Rule    string  `json:"rule"`
	What    string  `json:"what"` // "festival" or "black market deal"
	Minutes float64 `json:"minutes"`
	Low     float64 `json:"low"`
	High    float64 `json:"high"`
}

func (p CulturePriceProblem) String() string {
	return fmt.Sprintf("%s: a %s costs %s minutes of the reference town's culture income, want between %s and %s", p.Age, p.What, num(p.Minutes), num(p.Low), num(p.High))
}

// CulturePriceRow is one age of the check.
type CulturePriceRow struct {
	Age string `json:"age"`
	// Income is the reference town's culture per tick; Festival and Deal are
	// the culture the two cost a town making the age's typical income, and
	// the Minutes what each comes to in minutes of Income.
	Income         float64 `json:"culture_per_tick"`
	Festival       float64 `json:"festival"`
	Deal           float64 `json:"black_market_deal"`
	FestivalMinute float64 `json:"festival_minutes"`
	DealMinute     float64 `json:"black_market_deal_minutes"`
}

// StaticCulturePrices checks every age with a culture income against the
// band, and returns what falls outside it with one row per such age.
func StaticCulturePrices() ([]CulturePriceProblem, []CulturePriceRow) {
	set := rules.Core()
	tb := newRefTables()
	income := make([]float64, len(tb.ages))
	for i := range tb.ages {
		income[i] = tb.state(tb.refTown(i), i, 1).income["culture"]
	}
	return culturePriceProblems(tb.ages, income, func(age string) (festival, deal float64) {
		return game.CulturePrices(set, age, set.TypicalIncome("culture", age))
	})
}

// culturePriceProblems is StaticCulturePrices over a reference income per age
// (per tick, by position) and the prices a rule asks in an age: the check's
// broken-number tests feed in other rules.
func culturePriceProblems(ages []config.AgeDef, income []float64, prices func(age string) (festival, deal float64)) ([]CulturePriceProblem, []CulturePriceRow) {
	var problems []CulturePriceProblem
	var rows []CulturePriceRow
	for i, age := range ages {
		if income[i] <= 0 {
			continue // before culture: the prices are their minimums
		}
		festival, deal := prices(age.Key)
		perMinute := config.TicksPerMinute * income[i]
		row := CulturePriceRow{Age: age.Key, Income: income[i], Festival: festival, Deal: deal,
			FestivalMinute: festival / perMinute, DealMinute: deal / perMinute}
		rows = append(rows, row)
		for _, c := range []struct {
			what    string
			minutes float64
			times   float64
		}{{"festival", row.FestivalMinute, 1}, {"black market deal", row.DealMinute, CultureDealTimes}} {
			low, high := CulturePriceLowMinutes*c.times, CulturePriceHighMinutes*c.times
			switch {
			case c.minutes > high:
				problems = append(problems, CulturePriceProblem{Age: age.Key, Rule: CulturePriceHigh, What: c.what, Minutes: c.minutes, Low: low, High: high})
			case c.minutes < low:
				problems = append(problems, CulturePriceProblem{Age: age.Key, Rule: CulturePriceLow, What: c.what, Minutes: c.minutes, Low: low, High: high})
			}
		}
	}
	return problems, rows
}

// writeCulturePrices renders the check for the static scenario.
func writeCulturePrices(sb *strings.Builder, problems []CulturePriceProblem, rows []CulturePriceRow) {
	fmt.Fprintf(sb, "A festival and a Black Market deal cost minutes of a town's own culture income. Each is priced for a town making the age's typical culture income and read against the culture the reference model's town makes in the age, so a price that keeps in step with income costs about the same minutes of it in every age. A festival must cost between %s and %s minutes of it, a deal twice as many, in every age that has a culture income. `go test ./smoke` fails on any row outside that.\n\n", num(CulturePriceLowMinutes), num(CulturePriceHighMinutes))
	if len(problems) == 0 {
		sb.WriteString("No problems.\n\n")
	} else {
		for _, p := range problems {
			fmt.Fprintf(sb, "- %s\n", p)
		}
		sb.WriteString("\n")
	}
	sb.WriteString("| age | reference culture per tick | festival | festival in minutes of income | deal | deal in minutes of income |\n|---|---|---|---|---|---|\n")
	for _, r := range rows {
		fmt.Fprintf(sb, "| %s | %s | %s | %s | %s | %s |\n", r.Age, num(r.Income), num(r.Festival), num(r.FestivalMinute), num(r.Deal), num(r.DealMinute))
	}
}
