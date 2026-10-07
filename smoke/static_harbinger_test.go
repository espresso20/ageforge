package smoke

import (
	"maps"
	"strings"
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
		// Longer than any doom's lead (three fifths of the age at most) and
		// no longer than the age: two thirds of it.
		if age := config.AgeTargetTicks(r.Age); r.WarningTicks <= 0.6*age || r.WarningTicks > age {
			t.Errorf("the Last Passage in %s: warning %v ticks of an age of %v, want two thirds of it", r.Age, r.WarningTicks, age)
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

// TestWarningBraceTakesEffort: every Brace priced on its warning, the Last
// Passage's and the Reality Tear's in every age of the Cosmic Era the thread
// can begin in, takes a moderate economy at least ten hours of income at
// level 1 and is made inside the thread's shortest warning. Both used to be
// the era's price, a share of the largest age requirement, which the
// Interstellar Age makes in under 7 ticks; the check is what keeps a
// requirement-sized price from coming back.
func TestWarningBraceTakesEffort(t *testing.T) {
	for _, p := range StaticBraceRules() {
		t.Errorf("%s", p)
	}

	rows := HarbingerPrices()
	cosmic := config.EpochByKey()["cosmic_era"].Ages
	income := rules.Core().TypicalIncome
	passages, tears := 0, 0
	for _, r := range rows {
		if r.BraceOnWarning != (r.Epoch == "cosmic_era") {
			t.Errorf("%s (target %q) in %s: Brace priced on the warning is %v; only the Cosmic Era's threads are", r.Epoch, r.TargetEpoch, r.Age, r.BraceOnWarning)
		}
		if !r.BraceOnWarning {
			continue
		}
		if len(r.BraceL1) == 0 || r.WarningTicks <= 0 {
			t.Errorf("%s (target %q) in %s: no Brace price (%v) or no warning (%v ticks) to check", r.Epoch, r.TargetEpoch, r.Age, r.BraceL1, r.WarningTicks)
		}
		// What the report quotes: the hours a moderate economy needs.
		slowest := 0.0
		for res, c := range r.BraceL1 {
			if rate := income(res, r.Age); rate > 0 && c/rate > slowest {
				slowest = c / rate
			}
		}
		hours := slowest * config.TickSeconds / 3600
		if r.TargetEpoch == "" {
			passages++
			if hours < 20 || hours > 22 {
				t.Errorf("the Last Passage in %s: Brace level 1 is %.1f hours of income, want about 21 (a third of the age)", r.Age, hours)
			}
			continue
		}
		tears++
		// A price rounds to two figures, which moves the hours by up to 5%.
		if hours < WarningBraceMinHours || hours > 11.2 {
			t.Errorf("the Reality Tear in %s: Brace level 1 is %.1f hours of income, want about 10.5 (a sixth of the age)", r.Age, hours)
		}
	}
	if passages != len(cosmic) || tears != len(cosmic) {
		t.Fatalf("%d Last Passage rows and %d Reality Tear rows for %v", passages, tears, cosmic)
	}

	// The check fires on broken numbers, on either thread of the era's first
	// age, and on no ordinary doom.
	for _, lastPassage := range []bool{true, false} {
		name := "the Reality Tear: "
		if lastPassage {
			name = "the Last Passage: "
		}
		alter := func(change func(r *PriceRow)) []BraceRuleProblem {
			broken := make([]PriceRow, len(rows))
			for i, r := range rows {
				r.BraceL1 = maps.Clone(r.BraceL1)
				if r.Epoch == "cosmic_era" && (r.TargetEpoch == "") == lastPassage && r.Age == cosmic[0] {
					change(&r)
				}
				broken[i] = r
			}
			return braceRuleProblems(broken, income)
		}
		want := func(what string, got []BraceRuleProblem, rule string, resources ...string) {
			t.Helper()
			if len(got) != len(resources) {
				t.Errorf("%s%s: %d problem(s) %v, want %d", name, what, len(got), got, len(resources))
				return
			}
			for i, p := range got {
				if p.Rule != rule || p.Resource != resources[i] || p.Age != cosmic[0] || p.LastPassage != lastPassage || p.String() == "" {
					t.Errorf("%s%s: problem %d is %+v, want rule %s on %q in %s", name, what, i, p, rule, resources[i], cosmic[0])
				}
			}
		}
		// The era's price: 1.56T dark matter and 75.6B titanium, seconds of
		// income. Dark matter is its dearest part, at under 7 ticks.
		want("the era's price", alter(func(r *PriceRow) {
			r.BraceL1 = map[string]float64{"dark_matter": 1.56e12, "titanium": 7.56e10}
		}), BraceRuleEffort, "dark_matter")
		// Nine hours of income is still too little: titanium at nine, dark
		// matter at eight.
		want("nine hours of income", alter(func(r *PriceRow) {
			r.BraceL1["dark_matter"] = income("dark_matter", r.Age) * 8 * 3600 / config.TickSeconds
			r.BraceL1["titanium"] = income("titanium", r.Age) * 9 * 3600 / config.TickSeconds
		}), BraceRuleEffort, "titanium")
		// One dear resource is enough: the price takes as long as its slowest part.
		if got := alter(func(r *PriceRow) { r.BraceL1["dark_matter"] = 1 }); len(got) != 0 {
			t.Errorf("%sa price with one cheap resource and one dear one: %v, want no problem", name, got)
		}
		want("nothing asked", alter(func(r *PriceRow) {
			r.BraceL1 = map[string]float64{}
		}), BraceRuleEffort, "")
		// More than the warning makes of one resource.
		want("more than the warning makes", alter(func(r *PriceRow) {
			r.BraceL1["titanium"] = income("titanium", r.Age) * r.WarningTicks * 1.01
		}), BraceRuleWarning, "titanium")
		// A resource nothing makes in the arrival age can never be gathered there.
		want("a resource nothing makes yet", alter(func(r *PriceRow) {
			r.BraceL1["antimatter_fuel_nobody_makes"] = 1
		}), BraceRuleWarning, "antimatter_fuel_nobody_makes")
	}
	// An ordinary doom's row is left alone, however cheap its Brace: against
	// an income no price is ten hours of, only the Cosmic Era's rows break.
	for _, p := range braceRuleProblems(rows, func(string, string) float64 { return 1e30 }) {
		if p.Epoch != "cosmic_era" {
			t.Errorf("an ordinary doom's Brace is held to the warning rule: %s", p)
		}
	}
}

// TestHarbingerSecondLevelPrices: the price table carries each level's own
// price. Level 2 costs double level 1 in every era but the final one, where
// both threads charge the same again.
func TestHarbingerSecondLevelPrices(t *testing.T) {
	checked := 0
	for _, r := range HarbingerPrices() {
		want := 2.0
		if r.Epoch == "cosmic_era" {
			want = 1
		}
		for _, pair := range []struct {
			answer string
			l1, l2 map[string]float64
		}{{"Appease", r.AppeaseL1, r.AppeaseL2}, {"Brace", r.BraceL1, r.BraceL2}} {
			if len(pair.l1) == 0 || len(pair.l2) != len(pair.l1) {
				t.Errorf("%s (target %q) in %s: %s level 1 %v, level 2 %v", r.Epoch, r.TargetEpoch, r.Age, pair.answer, pair.l1, pair.l2)
			}
			for res, l1 := range pair.l1 {
				checked++
				// An era's Brace rounds each level up to a whole unit.
				if got := pair.l2[res]; got < want*l1-1 || got > want*l1+1 {
					t.Errorf("%s (target %q) in %s: %s level 2 asks %v %s, want %v times level 1's %v", r.Epoch, r.TargetEpoch, r.Age, pair.answer, got, res, want, l1)
				}
			}
		}
	}
	if checked == 0 {
		t.Fatal("no price checked")
	}
}

// TestNumReadsAsTheGameDoes: the report prints large numbers the way the
// game does, quadrillions included. It stopped at T and printed the Last
// Passage's Brace as 8.9e+03T.
func TestNumReadsAsTheGameDoes(t *testing.T) {
	for _, c := range []struct {
		v    float64
		want string
	}{
		{0, "0"}, {950, "950"}, {8050, "8050"}, {9999, "9999"},
		{10000, "10K"}, {80000, "80K"}, {1.2e6, "1.2M"}, {4.43e6, "4.43M"},
		{7.56e10, "75.6B"}, {1.56e12, "1.56T"}, {9.9996e14, "1Q"},
		{8.9e15, "8.9Q"}, {1.1e16, "11Q"}, {5.609e16, "56.1Q"}, {5.556e18, "5560Q"},
		{-2.5e15, "-2.5Q"},
	} {
		if got := num(c.v); got != c.want {
			t.Errorf("num(%v) = %q, want %q", c.v, got, c.want)
		}
		if got := num(c.v); strings.ContainsAny(got, "e+") {
			t.Errorf("num(%v) = %q: an exponent", c.v, got)
		}
	}
}
