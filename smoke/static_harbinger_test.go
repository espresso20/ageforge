package smoke

import (
	"maps"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/rules"
)

// TestHarbingerPricesFitStorage: every Appease and Brace price, both levels,
// fits the most storage buildable in the first age of its era. Appease grows
// with the pacing targets while faith and culture storage stay as typed, so a
// longer curve is what would break this.
func TestHarbingerPricesFitStorage(t *testing.T) {
	rows := HarbingerPrices()
	if len(rows) == 0 {
		t.Fatal("no harbinger prices read")
	}
	for _, p := range StaticHarbingerPrices() {
		t.Errorf("%s %s costs %v %s, over the %v buildable in %s", p.Epoch, p.Answer, p.Price, p.Resource, p.MaxStorage, p.Age)
	}
	// The check sees a price that cannot fit: faith storage in the Iron Era's
	// first age is finite, so no row may claim an uncapped faith store there.
	if m := MaxStorage(config.EpochByKey()["iron_era"].Ages[0], "faith"); m <= 0 || m > 1e300 {
		t.Errorf("iron-age faith storage %v: the check would never fire", m)
	}
}

// TestAppeasePricesFitTheWarning: every level-1 Appease price is one a
// moderate economy makes inside the thread's shortest warning, and the Last
// Passage's, in every age of the Cosmic Era its thread can begin in, costs
// more than a doom's foretold in the same age. The Last Passage used to keep
// an era price that no single age made (1.8 times the Interstellar Age's
// culture); the check is what holds its new one to the warning.
func TestAppeasePricesFitTheWarning(t *testing.T) {
	for _, p := range StaticAppeaseRules() {
		t.Errorf("%s", p)
	}

	// The check covers the Last Passage in every age it can arrive in, each
	// next to a doom's row for the same age, with a warning to hold it to.
	rows := HarbingerPrices()
	doom := map[string]bool{}
	var passages []PriceRow
	for _, r := range rows {
		if r.WarningTicks <= 0 || len(r.AppeaseL1) == 0 {
			t.Errorf("%s in %s: no warning (%v ticks) or no Appease price (%v) to check", r.Epoch, r.Age, r.WarningTicks, r.AppeaseL1)
		}
		if r.TargetEpoch == "" {
			passages = append(passages, r)
		} else {
			doom[r.Age] = true
		}
	}
	cosmic := config.EpochByKey()["cosmic_era"].Ages
	if len(passages) != len(cosmic) {
		t.Fatalf("%d Last Passage rows, want one for each of %v", len(passages), cosmic)
	}
	for i, r := range passages {
		if r.Age != cosmic[i] || !doom[r.Age] {
			t.Errorf("Last Passage row %d is for %s (a doom's row there: %v), want %s", i, r.Age, doom[r.Age], cosmic[i])
		}
		if r.WarningTicks != config.AgeTargetTicks(r.Age) {
			t.Errorf("the Last Passage in %s: warning %v ticks, want the whole age (%v)", r.Age, r.WarningTicks, config.AgeTargetTicks(r.Age))
		}
	}

	// The check fires on broken numbers. The old era price (3.1B faith, 48B
	// culture in every age) is more than the Interstellar Age makes; a doom's
	// own price is not above a doom's; a price with the culture left out is
	// caught too.
	income := rules.Core().FlowIncome
	alter := func(change func(r *PriceRow)) []AppeaseRuleProblem {
		broken := make([]PriceRow, len(rows))
		for i, r := range rows {
			r.AppeaseL1 = maps.Clone(r.AppeaseL1)
			if r.TargetEpoch == "" && r.Age == cosmic[0] {
				change(&r)
			}
			broken[i] = r
		}
		return appeaseRuleProblems(broken, income)
	}
	want := func(name string, got []AppeaseRuleProblem, rule string, resources ...string) {
		t.Helper()
		if len(got) != len(resources) {
			t.Errorf("%s: %d problem(s) %v, want %d", name, len(got), got, len(resources))
			return
		}
		for i, p := range got {
			if p.Rule != rule || p.Resource != resources[i] || !p.LastPassage || p.Age != cosmic[0] || p.String() == "" {
				t.Errorf("%s: problem %d is %+v, want rule %s on %s for the Last Passage in %s", name, i, p, rule, resources[i], cosmic[0])
			}
		}
	}
	want("the old era price", alter(func(r *PriceRow) {
		r.AppeaseL1 = map[string]float64{"faith": 3.1e9, "culture": 48e9}
	}), AppeaseRuleWarning, "culture", "faith")
	var ordinary map[string]float64
	for _, r := range rows {
		if r.TargetEpoch != "" && r.Age == cosmic[0] {
			ordinary = r.AppeaseL1
		}
	}
	want("a doom's price", alter(func(r *PriceRow) {
		r.AppeaseL1 = maps.Clone(ordinary)
	}), AppeaseRuleOrdinary, "culture", "faith")
	want("no culture asked", alter(func(r *PriceRow) {
		delete(r.AppeaseL1, "culture")
	}), AppeaseRuleOrdinary, "culture")
}
