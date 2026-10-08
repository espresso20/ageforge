package smoke

import (
	"math"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/rules"
)

// TestCultureStrengthBars: walked through the whole run at the pacing
// targets, a moderate town stays on Minor good epoch events at every moment
// a roll can fall, a town with twice the culture buildings has Major events
// open, and one with three and a half times has the Legendary, by the rule
// the roll reads. A town with no culture buildings reads nothing, whatever
// its wonders make. The check fails in either direction, and on the rule it
// replaced.
func TestCultureStrengthBars(t *testing.T) {
	problems, rows := StaticCultureStrength()
	for _, p := range problems {
		t.Errorf("%s", p)
	}
	// Culture comes with the Classical Age: a row for it and every age after.
	ages := config.AgeOrder()
	pos := config.AgePositions(ages)
	if want := len(ages) - pos["classical_age"]; len(rows) != want || rows[0].Age != "classical_age" {
		t.Fatalf("%d rows from %s, want %d from classical_age", len(rows), rows[0].Age, want)
	}
	near := func(s FaithSpan, want float64) bool {
		return math.Abs(s.Low-want) < 1e-9 && math.Abs(s.High-want) < 1e-9
	}
	transitions := map[string]bool{} // the ages an epoch roll falls on leaving
	for _, ep := range config.Epochs() {
		if !config.IsFinalEpoch(ep.Key) {
			transitions[ep.Ages[len(ep.Ages)-1]] = true
		}
	}
	for _, r := range rows {
		if r.Epoch != config.EpochForAge(r.Age) || r.ModerateSet <= 0 {
			t.Errorf("%+v: want a moderate set that makes culture", r)
		}
		if !near(r.None, 0) || !near(r.Moderate, 1/game.CultureFullSets) || !near(r.Twice, CultureTwiceSets/game.CultureFullSets) ||
			!near(r.Thrice, CultureThriceSets/game.CultureFullSets) || !near(r.Top, CultureTopSets/game.CultureFullSets) {
			t.Errorf("%s: none %v, moderate %v, twice %v, three times %v, %gx %v; want each town's sets over %g at every moment", r.Age, r.None, r.Moderate, r.Twice, r.Thrice, CultureTopSets, r.Top, game.CultureFullSets)
		}
		// The rule this replaced: where an epoch roll falls, no town that
		// made its culture reached the 40% Major events asked, a moderate one
		// stayed under 8% and under 1% from the Modern Age on; and from the
		// Medieval Age (where the market starts selling culture) to the
		// Digital Age, under ten minutes of a moderate economy's gold bought
		// it instead.
		if !transitions[r.Age] {
			continue
		}
		if r.OldFillTop >= game.CultureMajorAbove || r.OldFill >= 0.08 || pos[r.Age] >= pos["modern_age"] && r.OldFill >= 0.01 {
			t.Errorf("%s: the old rule read %.3f for a moderate town and %.3f for %g times the buildings; the store was never filled by making culture", r.Age, r.OldFill, r.OldFillTop, CultureTopSets)
		}
		if pos[r.Age] <= pos["digital_age"] && (r.OldGoldTicks <= 0 || r.OldGoldTicks > 400) {
			t.Errorf("%s: %v ticks of gold bought the old rule's 40%%, want a few hundred at most", r.Age, r.OldGoldTicks)
		}
	}
	var sb strings.Builder
	writeCultureStrength(&sb, problems, rows)
	if out := sb.String(); !strings.Contains(out, "No problems.") || strings.Count(out, "\n| ") != len(rows)+1 ||
		!strings.Contains(out, "22% (Minor)") || !strings.Contains(out, "44% (Major)") || !strings.Contains(out, "78% (Legendary)") {
		t.Errorf("the report section:\n%s", out)
	}

	// The check fires in both directions: on a rule that lets a moderate
	// town into Major events (full strength at two sets), and on one that
	// keeps a devoted town out (full strength at ten).
	incomes := flowIncomes(rules.Core(), "culture")
	broken := func(fullSets float64) map[string]int {
		got := map[string]int{}
		ps, _ := cultureStrengthProblems(incomes, func(held float64, m game.CultureSave) float64 {
			return math.Min(game.CultureDevotionOf(m)*game.CultureKeptOf(held, m)/fullSets, 1)
		})
		for _, p := range ps {
			if p.String() == "" {
				t.Errorf("a problem with nothing to say: %+v", p)
			}
			got[p.Rule]++
		}
		return got
	}
	if got := broken(2); got[CultureRuleModerate] == 0 || got[CultureRuleTwice] == 0 {
		t.Errorf("full strength at two sets: problems %v, want the moderate and twice rules broken", got)
	}
	if got := broken(10); got[CultureRuleTwice] == 0 || got[CultureRuleTop] == 0 || got[CultureRuleModerate] != 0 {
		t.Errorf("full strength at ten sets: problems %v, want the twice and top rules broken", got)
	}
	// The old rule itself, read as a strength: nothing but Minor events in
	// any age for any town.
	least := leastGeneralStore()
	old, _ := cultureStrengthProblems(incomes, func(held float64, m game.CultureSave) float64 {
		return held / least["transcendent_age"]
	})
	rulesHit := map[string]bool{}
	for _, p := range old {
		rulesHit[p.Rule] = true
	}
	if !rulesHit[CultureRuleTwice] || !rulesHit[CultureRuleTop] || rulesHit[CultureRuleModerate] {
		t.Errorf("the fill of a late store read as strength breaks %v, want the twice and top rules (no town reached a tier)", rulesHit)
	}
}
