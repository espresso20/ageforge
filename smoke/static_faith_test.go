package smoke

import (
	"math"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/rules"
)

// TestFaithStrengthBars: walked through the whole run with their hoards
// carried from age to age, a moderate town reads the bottom faith band at
// every moment a roll can fall, a town with twice the faith buildings the
// middle band, and one with three and a half times the top band, by the rule
// the rolls read. A town with no faith buildings reads nothing, whatever its
// wonders make. The check fails in either direction, and on each of the
// measures this one replaced.
func TestFaithStrengthBars(t *testing.T) {
	problems, rows := StaticFaithStrength()
	for _, p := range problems {
		t.Errorf("%s", p)
	}
	ages := config.AgeOrder()
	if len(rows) != len(ages) {
		t.Fatalf("%d rows for %d ages", len(rows), len(ages))
	}
	near := func(s FaithSpan, want float64) bool {
		return math.Abs(s.Low-want) < 1e-9 && math.Abs(s.High-want) < 1e-9
	}
	for i, r := range rows {
		if r.Age != ages[i] || r.Epoch != config.EpochForAge(r.Age) || r.ModerateSet <= 0 {
			t.Errorf("row %d: %+v, want %s with a moderate set that makes faith", i, r, ages[i])
		}
		if !near(r.None, 0) || !near(r.Moderate, 1/game.FaithFullSets) || !near(r.Twice, FaithTwiceSets/game.FaithFullSets) ||
			!near(r.Thrice, FaithThriceSets/game.FaithFullSets) || !near(r.Top, FaithTopSets/game.FaithFullSets) {
			t.Errorf("%s: no faith buildings %s, moderate %s, twice %s, three times %s, %gx %s; want each town's sets over %g at every moment", r.Age, r.None, r.Moderate, r.Twice, r.Thrice, FaithTopSets, r.Top, game.FaithFullSets)
		}
	}
	// The wonders the check is about are in the model: the age after
	// Stonehenge, every town is given at least the faith Stonehenge makes.
	// (This was "three times a moderate set's", which held at 0.72 against
	// 0.15; the moderate set is eight copies now, 0.288, and the same
	// Stonehenge is 2.5 times it. What the line means is that the wonder is
	// walked, so it says that.)
	stonehenge := 0.0
	for _, e := range config.BuildingByKey()["stonehenge"].Effects {
		if e.Type == "production" && e.Target == "faith" {
			stonehenge = e.Value
		}
	}
	if stonehenge <= 0 {
		t.Fatal("Stonehenge makes no faith: the check below has nothing to walk")
	}
	for _, r := range rows {
		if r.Age == "iron_age" && r.Given < stonehenge {
			t.Errorf("iron_age: every town is given %v faith a tick against a moderate set's %v; the check means to walk Stonehenge's %v", r.Given, r.ModerateSet, stonehenge)
		}
	}
	var sb strings.Builder
	writeFaithStrength(&sb, problems, rows)
	if out := sb.String(); !strings.Contains(out, "No problems.") || strings.Count(out, "\n| ") != len(ages)+1 ||
		!strings.Contains(out, "22% (bottom)") || !strings.Contains(out, "44% (middle)") || !strings.Contains(out, "78% (top)") {
		t.Errorf("the report section:\n%s", out)
	}

	// The check fires on the measures this one replaced, in the direction
	// each of them failed.
	incomes := faithIncomes(rules.Core())
	rulesBroken := func(strengthOf func(held float64, m game.FaithSave, in faithIncome) float64) map[string]map[string]bool {
		got := map[string]map[string]bool{}
		problems, _ := faithStrengthProblems(incomes, strengthOf)
		for _, p := range problems {
			if p.String() == "" {
				t.Errorf("a problem with nothing to say: %+v", p)
			}
			if got[p.Rule] == nil {
				got[p.Rule] = map[string]bool{}
			}
			got[p.Rule][p.Age] = true
		}
		return got
	}
	capped := func(v float64) float64 { return math.Min(v, 1) }

	// 1. Faith held against three fifths of what one age makes: a hoard
	// carried in from the ages before puts every saver in the top band, and a
	// wonder's faith counts like any other.
	oneAge := rulesBroken(func(held float64, _ game.FaithSave, in faithIncome) float64 {
		return capped(held / ((in.set + in.given) * in.ticks * 0.6))
	})
	for _, age := range ages[3:] { // from the Iron Age, where hoards and Stonehenge arrive
		if !oneAge[FaithRuleModerate][age] {
			t.Errorf("faith against one age's income: no moderate town leaves the bottom band in %s; it read the top band there", age)
		}
	}
	// 2. All the faith held against the moderate town's whole hoard: no
	// wonder buys a band, but the wonders' faith drowns the faith buildings'
	// while it is most of the hoard, and the top band takes ten times the
	// moderate set in the Iron Era.
	allFaith := rulesBroken(func(held float64, m game.FaithSave, _ faithIncome) float64 {
		return capped(held / (game.FaithFullSets * (m.Moderate + m.Other)))
	})
	if len(allFaith[FaithRuleModerate]) != 0 {
		t.Errorf("all faith against the moderate hoard: moderate towns out of the bottom band in %v", allFaith[FaithRuleModerate])
	}
	for _, age := range []string{"iron_age", "medieval_age", "colonial_age", "electric_age"} {
		if !allFaith[FaithRuleTop][age] {
			t.Errorf("all faith against the moderate hoard: %gx the faith buildings reach the top band in %s; the wonders' faith should hold them under it", FaithTopSets, age)
		}
	}
	if allFaith[FaithRuleTop]["interstellar_age"] {
		t.Error("all faith against the moderate hoard: by the Interstellar Age the wonders' faith is nothing and the top band should be in reach")
	}
	// 3. Full strength set too low or too high: the moderate town reaches
	// the middle band, or the devoted one misses the top.
	sets := func(full float64) func(float64, game.FaithSave, faithIncome) float64 {
		return func(held float64, m game.FaithSave, _ faithIncome) float64 {
			return capped(game.FaithStrengthOf(held, m) * game.FaithFullSets / full)
		}
	}
	if low := rulesBroken(sets(3.9)); len(low[FaithRuleModerate]) != len(ages) {
		t.Errorf("full strength at 3.9 sets: a moderate town leaves the bottom band in %d ages, want all %d", len(low[FaithRuleModerate]), len(ages))
	}
	if high := rulesBroken(sets(5)); len(high[FaithRuleTop]) != len(ages) || len(high[FaithRuleModerate]) != 0 {
		t.Errorf("full strength at 5 sets: the top band is out of reach in %d ages (want all %d), moderate problems %d", len(high[FaithRuleTop]), len(ages), len(high[FaithRuleModerate]))
	}
}
