package smoke

import (
	"fmt"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/rules"
)

// The Black Market payout check: a win pays the same worth whichever resource
// it is taken in. For every age from the Colonial Age, where the Black Market
// opens, and every resource the smugglers deal in that the age makes, the
// payout of a win on a deal for a town making the age's typical incomes
// (game.BlackMarketPayout, the engine's own figure) is read in minutes of the
// income the reference model's town makes of that resource
// (smoke/reference.go, a measure the rule does not read), and must come to
// between BMPayoutLowMinutes and BMPayoutHighMinutes of it.
//
// The bounds come from the fixed curve. A deal that goes through pays the
// deal's own minutes of culture income (12,000 where it asks 200 hours) in
// minutes of income of the resource; against the reference model's incomes
// that reads from 5,064 minutes (dark matter, in the Interstellar Age) to
// 10,046 (gold, in the Transcendent Age), because the model's incomes part
// from the typical ones a little, by resource. Low is two thirds of the
// lowest and high half again the highest, so the curve passes as it is and a
// payout that parts from it by more than that does not. The fixed gold rates
// the payout was paid at before put 109 of these 115 payouts outside it, from
// next to nothing (steel, electricity) to over a million minutes (coal, iron),
// for the same deal, by item.
const (
	BMPayoutLowMinutes  = 3300.0
	BMPayoutHighMinutes = 15_000.0
)

// The rules a BMPayoutProblem can break.
const (
	// BMPayoutHigh: a win pays more minutes of the resource's income than
	// the band.
	BMPayoutHigh = "high"
	// BMPayoutLow: it pays fewer.
	BMPayoutLow = "low"
)

// BMPayoutProblem is one payout outside its band.
type BMPayoutProblem struct {
	Age      string  `json:"age"`
	Resource string  `json:"resource"`
	Rule     string  `json:"rule"`
	Minutes  float64 `json:"minutes"`
}

func (p BMPayoutProblem) String() string {
	return fmt.Sprintf("%s: a win on %s pays %s minutes of the reference town's income of it, want between %s and %s", p.Age, p.Resource, num(p.Minutes), num(BMPayoutLowMinutes), num(BMPayoutHighMinutes))
}

// BMPayoutRow is one age and resource of the check.
type BMPayoutRow struct {
	Age      string `json:"age"`
	Resource string `json:"resource"`
	// Income is the reference town's income of the resource per tick, Payout
	// what a win pays a town making the age's typical incomes, and Minutes
	// that in minutes of Income.
	Income  float64 `json:"income_per_tick"`
	Payout  float64 `json:"payout"`
	Minutes float64 `json:"minutes"`
}

// StaticBlackMarketPayouts checks every payout of the Black Market against the
// band, and returns what falls outside it with one row per age and resource.
func StaticBlackMarketPayouts() ([]BMPayoutProblem, []BMPayoutRow) {
	set := rules.Core()
	tb := newRefTables()
	income := make([]map[string]float64, len(tb.ages))
	for i := range tb.ages {
		income[i] = tb.state(tb.refTown(i), i, 1).income
	}
	return bmPayoutProblems(tb.ages, income, func(age, res string) float64 {
		typC := set.TypicalIncome("culture", age)
		_, deal := game.CulturePrices(set, age, typC)
		return game.BlackMarketPayout(set, age, res, deal, typC, set.TypicalIncome(res, age))
	})
}

// bmPayoutProblems is StaticBlackMarketPayouts over a reference income per
// resource for each age (per tick, by position) and the payout a rule gives
// in an age on a resource: the check's broken-number tests feed in other
// rules. A payout of nothing (the age makes none of the resource) or a
// resource the reference model makes none of has no row.
func bmPayoutProblems(ages []config.AgeDef, income []map[string]float64, payout func(age, res string) float64) ([]BMPayoutProblem, []BMPayoutRow) {
	var problems []BMPayoutProblem
	var rows []BMPayoutRow
	pos := config.AgePositions(config.AgeOrder())
	for i, age := range ages {
		if pos[age.Key] < pos["colonial_age"] {
			continue // before the Black Market opens
		}
		for _, res := range game.BlackMarketResources() {
			made, pay := income[i][res], payout(age.Key, res)
			if made <= 0 || pay <= 0 {
				continue
			}
			row := BMPayoutRow{Age: age.Key, Resource: res, Income: made, Payout: pay, Minutes: pay / (config.TicksPerMinute * made)}
			rows = append(rows, row)
			switch {
			case row.Minutes > BMPayoutHighMinutes:
				problems = append(problems, BMPayoutProblem{Age: age.Key, Resource: res, Rule: BMPayoutHigh, Minutes: row.Minutes})
			case row.Minutes < BMPayoutLowMinutes:
				problems = append(problems, BMPayoutProblem{Age: age.Key, Resource: res, Rule: BMPayoutLow, Minutes: row.Minutes})
			}
		}
	}
	return problems, rows
}

// writeBMPayouts renders the check for the static scenario.
func writeBMPayouts(sb *strings.Builder, problems []BMPayoutProblem, rows []BMPayoutRow) {
	fmt.Fprintf(sb, "A deal that goes through on the Black Market pays the same worth whichever resource it is taken in: the deal's minutes of culture income, as minutes of the town's income of the resource. For every age from the Colonial Age and every resource the smugglers deal in that the age makes, a payout for a town making the age's typical incomes is read in minutes of the reference model's income of that resource, and must come to between %s and %s of them. `go test ./smoke` fails on any row outside that.\n\n", num(BMPayoutLowMinutes), num(BMPayoutHighMinutes))
	if len(problems) == 0 {
		sb.WriteString("No problems.\n\n")
	} else {
		for _, p := range problems {
			fmt.Fprintf(sb, "- %s\n", p)
		}
		sb.WriteString("\n")
	}
	sb.WriteString("| age | resource | reference income per tick | a deal pays | in minutes of that income |\n|---|---|---|---|---|\n")
	for _, r := range rows {
		fmt.Fprintf(sb, "| %s | %s | %s | %s | %s |\n", r.Age, r.Resource, num(r.Income), num(r.Payout), num(r.Minutes))
	}
}
