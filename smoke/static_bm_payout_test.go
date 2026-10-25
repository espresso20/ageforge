package smoke

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/rules"
)

// TestBlackMarketPayoutsStayInBand: in every age from the Colonial Age, a win
// on each resource the smugglers deal in that the age makes pays between
// BMPayoutLowMinutes and BMPayoutHighMinutes of the reference town's income
// of that resource, by the engine's own payout: the same worth whichever
// resource it is taken in.
func TestBlackMarketPayoutsStayInBand(t *testing.T) {
	problems, rows := StaticBlackMarketPayouts()
	for _, p := range problems {
		t.Errorf("%s (%s)", p, p.Rule)
	}
	// Every dealt resource is paid in some age: gold and iron from the
	// Colonial Age, quantum flux from the Space Age.
	first := map[string]string{}
	for _, r := range rows {
		if _, ok := first[r.Resource]; !ok {
			first[r.Resource] = r.Age
		}
		if r.Income <= 0 || r.Payout <= 0 || r.Minutes < BMPayoutLowMinutes || r.Minutes > BMPayoutHighMinutes {
			t.Errorf("%+v: want a payout of %v to %v minutes of the reference income", r, BMPayoutLowMinutes, BMPayoutHighMinutes)
		}
	}
	if len(first) != len(game.BlackMarketResources()) || first["gold"] != "colonial_age" || first["iron"] != "colonial_age" || first["quantum_flux"] != "space_age" {
		t.Errorf("the first age each resource pays in: %v", first)
	}
	pos := config.AgePositions(config.AgeOrder())
	for _, r := range rows {
		if pos[r.Age] < pos["colonial_age"] {
			t.Errorf("a row in %s, before the Black Market opens", r.Age)
		}
	}
	var sb strings.Builder
	writeBMPayouts(&sb, problems, rows)
	if out := sb.String(); !strings.Contains(out, "No problems.") || strings.Count(out, "\n| ") != len(rows)+1 {
		t.Errorf("the report section:\n%s", out)
	}
}

// TestBlackMarketPayoutCheckCatchesTheFixedRates keeps the guard honest: the
// payout as it was (the deal as gold at a fixed rate for each resource) must
// be flagged high for some resources and low for others, in the late ages as
// in the middle ones, and a payout a rule that pays everything alike but too
// much or too little must be flagged too.
func TestBlackMarketPayoutCheckCatchesTheFixedRates(t *testing.T) {
	set := rules.Core()
	tb := newRefTables()
	income := make([]map[string]float64, len(tb.ages))
	for i := range tb.ages {
		income[i] = tb.state(tb.refTown(i), i, 1).income
	}
	deal := func(age string) float64 {
		_, d := game.CulturePrices(set, age, set.TypicalIncome("culture", age))
		return d
	}
	flagged := func(payout func(age, res string) float64) map[string]string {
		out := map[string]string{}
		problems, _ := bmPayoutProblems(tb.ages, income, payout)
		for _, p := range problems {
			if p.String() == "" {
				t.Errorf("a problem with nothing to say: %+v", p)
			}
			out[p.Age+"/"+p.Resource] = p.Rule
		}
		return out
	}
	// The fixed rates: the deal, times 2.5, as gold at the smugglers' rate.
	rates := map[string]float64{"gold": 1, "iron": 2, "coal": 0.08, "steel": 5, "oil": 3, "electricity": 0.5, "data": 5, "crypto": 20, "dark_matter": 50, "quantum_flux": 100}
	old := flagged(func(age, res string) float64 { return 2.5 * deal(age) / rates[res] })
	for _, c := range []struct{ key, rule string }{
		{"colonial_age/coal", BMPayoutHigh}, {"colonial_age/gold", BMPayoutLow},
		{"quantum_age/coal", BMPayoutHigh}, {"quantum_age/steel", BMPayoutLow}, {"quantum_age/electricity", BMPayoutLow},
		{"transcendent_age/coal", BMPayoutHigh},
	} {
		if old[c.key] != c.rule {
			t.Errorf("the fixed rates, %s: flagged %q, want %q", c.key, old[c.key], c.rule)
		}
	}
	// A rule that pays the same in every resource but ten times too much, or
	// the earlier 0.3 times the deal, or a tenth of it, fails everywhere.
	for _, c := range []struct {
		times float64
		rule  string
	}{{10, BMPayoutHigh}, {0.3, BMPayoutLow}, {0.1, BMPayoutLow}} {
		got := flagged(func(age, res string) float64 {
			typC, typR := set.TypicalIncome("culture", age), set.TypicalIncome(res, age)
			return c.times * game.BlackMarketPayout(set, age, res, deal(age), typC, typR)
		})
		if got["quantum_age/coal"] != c.rule || got["colonial_age/gold"] != c.rule {
			t.Errorf("a payout %vx the rule's: Quantum coal %q, Colonial gold %q, want both %q", c.times, got["quantum_age/coal"], got["colonial_age/gold"], c.rule)
		}
	}
	// The engine's own payout is flagged nowhere.
	if got := flagged(func(age, res string) float64 {
		typC := set.TypicalIncome("culture", age)
		return game.BlackMarketPayout(set, age, res, deal(age), typC, set.TypicalIncome(res, age))
	}); len(got) != 0 {
		t.Errorf("the engine's own payout is flagged: %v", got)
	}
}
