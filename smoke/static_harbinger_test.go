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

// TestBraceTakesEffort: every thread's level-1 Brace, in every era and
// every age its harbinger can arrive in, takes a moderate economy a real
// share of the thread's shortest warning and is made inside it, in
// materials the age's own buildings make. An ordinary doom's takes a third
// of its warning; the final era's two take more (the Reality Tear's about
// ten and a half hours of income, the Last Passage's about twenty-one).
// Every one of them used to be a share of what the era's advances ask, which
// later ages make in seconds; the check is what keeps a requirement-sized
// price from coming back.
func TestBraceTakesEffort(t *testing.T) {
	for _, p := range StaticBraceRules() {
		t.Errorf("%s", p)
	}

	rows := HarbingerPrices()
	cosmic := config.EpochByKey()["cosmic_era"].Ages
	income, built := rules.Core().TypicalIncome, rules.Core().BuildingOutput
	passages, tears, ordinary := 0, 0, 0
	for _, r := range rows {
		if r.FinalEra != (r.Epoch == "cosmic_era") {
			t.Errorf("%s (target %q) in %s: final era is %v", r.Epoch, r.TargetEpoch, r.Age, r.FinalEra)
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
		switch {
		case r.TargetEpoch == "":
			passages++
			if hours < 20 || hours > 22 {
				t.Errorf("the Last Passage in %s: Brace level 1 is %.1f hours of income, want about 21 (a third of the age)", r.Age, hours)
			}
		case r.FinalEra:
			tears++
			// A price rounds up to two figures, which moves the hours by up to
			// 10% (1.01 becomes 1.1, as the Interstellar Age's dark matter does).
			if hours < WarningBraceMinHours || hours > 11.6 {
				t.Errorf("the Reality Tear in %s: Brace level 1 is %.1f hours of income, want about 10.5 (a sixth of the age)", r.Age, hours)
			}
		default:
			ordinary++
			// A third of the warning, give or take the same rounding.
			if share := slowest / r.WarningTicks; share < 1.0/3 || share > 1.1/3 {
				t.Errorf("a doom of %s in %s: Brace level 1 takes %.3f of the shortest warning (%.2f hours of income), want a third", r.Epoch, r.Age, share, hours)
			}
		}
	}
	if passages != len(cosmic) || tears != len(cosmic) || ordinary != len(rows)-2*len(cosmic) || ordinary == 0 {
		t.Fatalf("%d Last Passage rows, %d Reality Tear rows and %d ordinary rows of %d for %v", passages, tears, ordinary, len(rows), cosmic)
	}

	// The check fires on broken numbers, on either thread of the final era's
	// first age and on an ordinary doom's.
	type target struct {
		name, epoch, age string
		lastPassage      bool
		a, b             string // two of its materials, a the slowest
	}
	electric := config.EpochByKey()["electric_era"].Ages
	for _, tg := range []target{
		{"the Last Passage: ", "cosmic_era", cosmic[0], true, "titanium", "dark_matter"},
		{"the Reality Tear: ", "cosmic_era", cosmic[0], false, "titanium", "dark_matter"},
		{"an Electric Era doom: ", "electric_era", electric[1], false, "steel", "oil"},
	} {
		var row PriceRow
		alter := func(change func(r *PriceRow)) []BraceRuleProblem {
			broken := make([]PriceRow, len(rows))
			for i, r := range rows {
				r.BraceL1 = maps.Clone(r.BraceL1)
				if r.Epoch == tg.epoch && (r.TargetEpoch == "") == tg.lastPassage && r.Age == tg.age {
					row = r
					change(&r)
				}
				broken[i] = r
			}
			return braceRuleProblems(broken, income, built)
		}
		want := func(what string, got []BraceRuleProblem, rule string, resources ...string) {
			t.Helper()
			if len(got) != len(resources) {
				t.Errorf("%s%s: %d problem(s) %v, want %d", tg.name, what, len(got), got, len(resources))
				return
			}
			for i, p := range got {
				if p.Rule != rule || p.Resource != resources[i] || p.Age != tg.age || p.LastPassage != tg.lastPassage || p.String() == "" {
					t.Errorf("%s%s: problem %d is %+v, want rule %s on %q in %s", tg.name, what, i, p, rule, resources[i], tg.age)
				}
			}
		}
		if got := alter(func(*PriceRow) {}); len(got) != 0 || row.Age != tg.age || row.BraceL1[tg.a] <= 0 || row.BraceL1[tg.b] <= 0 {
			t.Fatalf("%sno row to break (%+v), or it is broken already: %v", tg.name, row, got)
		}
		least := braceMinTicks(row)
		// The era's old price, a share of its largest requirements: seconds of
		// income. The Cosmic Era's was 1.56T dark matter and 75.6B titanium
		// (dark matter its dearest part, under 7 ticks), the Electric Era's
		// 56.4M steel, 924K oil and 3.96M electricity (steel, 3 ticks of the
		// Electric Age's income; the electricity is under one).
		old, dearest := map[string]float64{"dark_matter": 1.56e12, "titanium": 7.56e10}, "dark_matter"
		if tg.epoch == "electric_era" {
			old, dearest = map[string]float64{"steel": 5.64e7, "oil": 9.24e5, "electricity": 3.96e6}, "steel"
		}
		want("the era's old price", alter(func(r *PriceRow) { r.BraceL1 = maps.Clone(old) }), BraceRuleEffort, dearest)
		// Nine tenths of the least it may take is still too little.
		want("just under the least", alter(func(r *PriceRow) {
			for res := range r.BraceL1 {
				r.BraceL1[res] = income(res, r.Age) * least * 0.8
			}
			r.BraceL1[tg.a] = income(tg.a, r.Age) * least * 0.9
		}), BraceRuleEffort, tg.a)
		// One dear resource is enough: the price takes as long as its slowest part.
		if got := alter(func(r *PriceRow) { r.BraceL1[tg.b] = 1 }); len(got) != 0 {
			t.Errorf("%sa price with one cheap resource and the rest dear: %v, want no problem", tg.name, got)
		}
		want("nothing asked", alter(func(r *PriceRow) {
			r.BraceL1 = map[string]float64{}
		}), BraceRuleEffort, "")
		// More than the warning makes of one resource.
		want("more than the warning makes", alter(func(r *PriceRow) {
			r.BraceL1[tg.a] = income(tg.a, r.Age) * r.WarningTicks * 1.01
		}), BraceRuleWarning, tg.a)
		// A resource no building makes by the arrival age: one nothing makes
		// at all, and crypto, which only a wonder trickles out.
		want("a resource nothing makes", alter(func(r *PriceRow) {
			r.BraceL1["antimatter_fuel_nobody_makes"] = 1
		}), BraceRuleMaterial, "antimatter_fuel_nobody_makes")
		want("a resource only a wonder makes", alter(func(r *PriceRow) {
			r.BraceL1["crypto"] = 1
		}), BraceRuleMaterial, "crypto")
	}
	// The two prices the rule dropped: the Modern Age's data, which it buys
	// at the market, and the Neon Era's crypto.
	for _, r := range rows {
		if _, ok := r.BraceL1["crypto"]; ok {
			t.Errorf("%s in %s: Brace asks crypto, which no building makes", r.Epoch, r.Age)
		}
		if _, ok := r.BraceL1["data"]; ok && r.Age == "modern_age" {
			t.Errorf("%s in %s: Brace asks data, which the Modern Age buys", r.Epoch, r.Age)
		}
		if _, ok := r.BraceL1["data"]; !ok && r.Age == "information_age" {
			t.Errorf("%s in %s: Brace asks no data, which the Information Age makes", r.Epoch, r.Age)
		}
	}
}

// TestHarbingerSecondLevelPrices: the price table carries each level's own
// price, and level 2 costs the same again as level 1 in every thread.
func TestHarbingerSecondLevelPrices(t *testing.T) {
	checked := 0
	for _, r := range HarbingerPrices() {
		for _, pair := range []struct {
			answer string
			l1, l2 map[string]float64
		}{{"Appease", r.AppeaseL1, r.AppeaseL2}, {"Brace", r.BraceL1, r.BraceL2}} {
			if len(pair.l1) == 0 || len(pair.l2) != len(pair.l1) {
				t.Errorf("%s (target %q) in %s: %s level 1 %v, level 2 %v", r.Epoch, r.TargetEpoch, r.Age, pair.answer, pair.l1, pair.l2)
			}
			for res, l1 := range pair.l1 {
				checked++
				if got := pair.l2[res]; got != l1 {
					t.Errorf("%s (target %q) in %s: %s level 2 asks %v %s, want level 1's %v again", r.Epoch, r.TargetEpoch, r.Age, pair.answer, got, res, l1)
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
