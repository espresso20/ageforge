package smoke

import (
	"maps"
	"math"
	"slices"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/rules"
)

// TestHarbingerPricesAreRead: the price table the static checks run on has a
// row for every age a harbinger can arrive in. Stock has no cap, so no price
// is held against a store; the warning each is priced on holds it
// (TestAppeasePricesFitTheWarning, the Brace rules).
func TestHarbingerPricesAreRead(t *testing.T) {
	if len(HarbingerPrices()) == 0 {
		t.Fatal("no harbinger prices read")
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
		if age := config.LatePriceTargetTicks(r.Age); r.WarningTicks <= 0.6*age || r.WarningTicks > age {
			t.Errorf("the Last Passage in %s: warning %v ticks of an age of %v, want two thirds of it", r.Age, r.WarningTicks, age)
		}
	}

	// The check fires on broken numbers. The old era price (3.1B faith, 48B
	// culture in every age, on the late price climb) is more than the
	// Interstellar Age makes; a doom's
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
		c := config.LatePriceMultiplier(cosmic[0])
		r.AppeaseL1 = map[string]float64{"faith": float64(3.1 * config.Billion * c), "culture": float64(48 * config.Billion * c)}
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
	capped := map[string]int{} // arrival age -> materials on their ceiling
	var allCapped []string     // the ages whose every material is on it
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
		// In hours of the age before the late price climb, which lengthens
		// the age and the Brace price alike.
		hours := slowest * config.TickSeconds / 3600 / config.LatePriceMultiplier(r.Age)
		switch {
		case r.TargetEpoch == "":
			passages++
			// A third of the age is 20.8 hours; rounding the price up to two
			// figures can add a tenth.
			// A material on its ceiling asks less than that.
			onCeiling := false
			for res, c := range r.BraceL1 {
				if limit, ok := r.BraceCeiling[res]; ok && c >= float64(limit*braceAtCeiling) {
					onCeiling = true
				}
			}
			if hours > 23 || hours < 20 && !onCeiling {
				t.Errorf("the Last Passage in %s: Brace level 1 is %.1f hours of income, want about 21 (a third of the age), or less on its ceiling", r.Age, hours)
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
			// A third of the warning, give or take the same rounding, unless
			// the ceiling holds every material under it: three fifths of the
			// age's stock scale.
			onCeiling, all := 0, true
			for res, c := range r.BraceL1 {
				limit, ok := r.BraceCeiling[res]
				if !ok || limit <= 0 || r.StockScale <= 0 || r.BraceCeilingHours[res] <= 0 {
					t.Errorf("a doom of %s in %s: no ceiling for %s (%v of a stock scale of %v, %v hours)", r.Epoch, r.Age, res, limit, r.StockScale, r.BraceCeilingHours[res])
				}
				if c > limit {
					t.Errorf("a doom of %s in %s: Brace asks %v %s, over its ceiling %v", r.Epoch, r.Age, c, res, limit)
				}
				if c >= braceAtCeiling*limit {
					onCeiling++
				} else {
					all = false
				}
			}
			capped[r.Age] = onCeiling
			share := slowest / r.WarningTicks
			if share > 1.1/3 || share < 1.0/3 && !all {
				t.Errorf("a doom of %s in %s: Brace level 1 takes %.3f of the shortest warning (%.2f hours of income) with %d of %d materials on their ceiling, want a third", r.Epoch, r.Age, share, hours, onCeiling, len(r.BraceL1))
			}
			if all {
				allCapped = append(allCapped, r.Age)
			}
		}
	}
	if passages != len(cosmic) || tears != len(cosmic) || ordinary != len(rows)-2*len(cosmic) || ordinary == 0 {
		t.Fatalf("%d Last Passage rows, %d Reality Tear rows and %d ordinary rows of %d for %v", passages, tears, ordinary, len(rows), cosmic)
	}
	// The ceiling bites. Before the storage wall the price without it was one
	// and a half to three times a moderate store from the Colonial Age on.
	// Since the wall a store is sized by what an age's buildings cost
	// (config/storage_rule.go), not by what the age makes, and the ceiling
	// holds some material in every age a doom can come in. The moderate
	// economy is eight copies and a moderate store five, so a third of the
	// warning outgrows the ceiling of more of them than it did against five
	// copies: it holds every material in three, the Colonial, Industrial and
	// Modern Ages, where the price is then quicker to make than a third of
	// the warning. The maker rule gave stone a maker in the Iron, Classical
	// and Medieval Ages, so its Brace joins gold and iron on the ceiling there
	// (all three of the Iron Era's materials, in all three ages), and the
	// Electric Age's electricity and steel moved. A price counts as on its
	// ceiling only when it equals the row's ceiling (the rounded-down price a
	// capped material is given), so the Classical Age, whose stone is a third
	// of the warning and under it, is not. The Modern Age left the
	// all-capped list when the Mainframe Hall gave it a maker of data: its
	// data Brace is under the ceiling. Every new copy costs NewCopyPriceFactor
	// more and the stores were sized to the dearest first copy, so they grew
	// with it while what the age makes did not: a third of the warning is
	// under the ceiling of more materials, and the Iron, Medieval and Colonial
	// Ages are off the all-capped list (the Atomic, Electric and Modern Ages
	// have none on it). At the factor 1.43 the Iron Age's stone and the Modern
	// Age's materials are under their ceilings too. The counts are read from
	// the price table.
	wantCapped := map[string]int{
		"iron_age": 2, "classical_age": 2, "medieval_age": 1,
		"renaissance_age": 1, "colonial_age": 1, "industrial_age": 2,
		"victorian_age": 1, "electric_age": 1, "atomic_age": 0,
		"modern_age": 0, "information_age": 2, "digital_age": 2,
		"cyberpunk_age": 1, "fusion_age": 1, "space_age": 1,
	}
	if !maps.Equal(capped, wantCapped) {
		t.Errorf("materials on their ceiling by age: %v; want %v", capped, wantCapped)
	}
	// (The Iron Age's three prices equal the ceiling by chance: a third of the
	// warning, rounded up, is the same figure. The ceiling did not cut them,
	// and the test in package game that holds Appease against Brace counts
	// only the Industrial Age.)
	if want := []string{"industrial_age"}; !slices.Equal(allCapped, want) {
		t.Errorf("every material is on its ceiling in %v, want %v", allCapped, want)
	}
	// The final era's two threads have the ceiling a doom has, and the age's
	// stock scale holds them with room to spare. The Last Passage's takes
	// 0.60 of it in the age it arrives in (titanium on its ceiling; it was 0.55
	// of dark matter, 17Q of a scale of 31Q, until the Interstellar Age had a
	// maker of titanium and the price, which follows the moderate economy and
	// not the scale, outgrew it: 3.55 for the Last Passage, 1.75 for the
	// Reality Tear). Before the storage wall it took 0.97.
	for _, b := range BraceStoreRows(rows, income) {
		if b.Epoch != "cosmic_era" {
			if b.ScaleShare > 0.6+1e-9 {
				t.Errorf("a doom of %s in %s: the largest part is %.2f of the stock scale, over the ceiling", b.Epoch, b.Age, b.ScaleShare)
			}
			continue
		}
		if b.ScaleShare <= 0 || b.ScaleShare > 0.6+1e-9 {
			t.Errorf("%s (Last Passage %v) in %s: the largest part is %.2f of the stock scale, want it held under the ceiling", b.Epoch, b.LastPassage, b.Age, b.ScaleShare)
		}
		if b.LastPassage && b.Age == cosmic[0] && math.Abs(b.ScaleShare-0.60) > 0.02 {
			t.Errorf("the Last Passage in %s takes %.2f of the stock scale, want 0.60 (titanium on its ceiling)", b.Age, b.ScaleShare)
		}
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
		{"an Electric Era doom: ", "electric_era", electric[1], false, "oil", "steel"},
	} {
		var row PriceRow
		builtBy := built
		alter := func(change func(r *PriceRow)) []BraceRuleProblem {
			broken := make([]PriceRow, len(rows))
			for i, r := range rows {
				r.BraceL1 = maps.Clone(r.BraceL1)
				if r.Epoch == tg.epoch && (r.TargetEpoch == "") == tg.lastPassage && r.Age == tg.age {
					// The ceiling is tested on its own below: these are the
					// rules about the warning, with it out of the way.
					r.BraceCeiling = nil
					row = r
					change(&r)
				}
				broken[i] = r
			}
			return braceRuleProblems(broken, income, builtBy)
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
		old, dearest := map[string]float64{"dark_matter": 1.56 * config.Trillion, "titanium": 75.6 * config.Billion}, "dark_matter"
		if tg.epoch == "electric_era" {
			old, dearest = map[string]float64{"steel": 56.4 * config.Million, "oil": 924_000, "electricity": 3.96 * config.Million}, "steel"
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
		// (The Last Passage's real titanium sits on its ceiling, under the least;
		// give the slowest part the least it may take.)
		if got := alter(func(r *PriceRow) { r.BraceL1[tg.a] = income(tg.a, r.Age) * least * (1 + 1e-9); r.BraceL1[tg.b] = 1 }); len(got) != 0 {
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
		// at all, and one only a wonder trickles out. Crypto was that until the
		// Black Market Hub made it, so the second case takes the hub's output
		// away: with only the wonder making it, the check must flag it.
		want("a resource nothing makes", alter(func(r *PriceRow) {
			r.BraceL1["antimatter_fuel_nobody_makes"] = 1
		}), BraceRuleMaterial, "antimatter_fuel_nobody_makes")
		builtBy = func(res, age string) float64 {
			if res == "crypto" {
				return 0
			}
			return built(res, age)
		}
		want("a resource only a wonder makes", alter(func(r *PriceRow) {
			r.BraceL1["crypto"] = 1
		}), BraceRuleMaterial, "crypto")
		builtBy = built
	}
	// The ceiling, on an ordinary doom's row as it stands (the Industrial
	// Age's, where every material sits on it).
	ceiling := func(change func(r *PriceRow)) []BraceRuleProblem {
		broken := make([]PriceRow, len(rows))
		for i, r := range rows {
			r.BraceL1, r.BraceCeiling = maps.Clone(r.BraceL1), maps.Clone(r.BraceCeiling)
			if r.Epoch == "steel_era" && r.Age == "industrial_age" {
				change(&r)
			}
			broken[i] = r
		}
		return braceRuleProblems(broken, income, built)
	}
	if got := ceiling(func(*PriceRow) {}); len(got) != 0 {
		t.Errorf("the Industrial Age's row as it stands: %v", got)
	}
	// A price over the ceiling breaks the store rule.
	if got := ceiling(func(r *PriceRow) { r.BraceL1["gold"] = r.BraceCeiling["gold"] * 1.01 }); len(got) != 1 || got[0].Rule != BraceRuleStore || got[0].Resource != "gold" || got[0].String() == "" {
		t.Errorf("gold a hundredth over its ceiling: %v, want the store rule on gold", got)
	}
	// Without its ceiling the same price is too little for the warning: the
	// ceiling is what excuses it, and nothing else does.
	if got := ceiling(func(r *PriceRow) { r.BraceCeiling = nil }); len(got) != 1 || got[0].Rule != BraceRuleEffort {
		t.Errorf("the Industrial Age's price with no ceiling: %v, want the effort rule", got)
	}
	// A price well under both the ceiling and the warning's share is still
	// caught.
	if got := ceiling(func(r *PriceRow) {
		for res := range r.BraceL1 {
			r.BraceL1[res] = r.BraceCeiling[res] / 2
		}
	}); len(got) != 1 || got[0].Rule != BraceRuleEffort {
		t.Errorf("half the ceiling of every material: %v, want the effort rule", got)
	}

	// The two prices the rule dropped: the Modern Age's data, which it bought
	// at the market until the Mainframe Hall gave the age a maker (so the
	// Digital Era's Brace asks it in the Modern Age now, as in the
	// Information Age), and the Neon Era's crypto.
	for _, r := range rows {
		if _, ok := r.BraceL1["crypto"]; ok {
			t.Errorf("%s in %s: Brace asks crypto, which no building makes", r.Epoch, r.Age)
		}
		if _, ok := r.BraceL1["data"]; !ok && r.Epoch == "digital_era" && (r.Age == "modern_age" || r.Age == "information_age") {
			t.Errorf("%s in %s: Brace asks no data, which the age makes", r.Epoch, r.Age)
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
		{10000, "10K"}, {80000, "80K"}, {1.2 * config.Million, "1.2M"}, {4.43 * config.Million, "4.43M"},
		{75.6 * config.Billion, "75.6B"}, {1.56 * config.Trillion, "1.56T"}, {999.96 * config.Trillion, "1Q"},
		{8.9 * config.Quadrillion, "8.9Q"}, {11 * config.Quadrillion, "11Q"}, {56.09 * config.Quadrillion, "56.1Q"}, {5.556 * config.Quintillion, "5.56Qi"},
		{-2.5 * config.Quadrillion, "-2.5Q"},
	} {
		if got := num(c.v); got != c.want {
			t.Errorf("num(%v) = %q, want %q", c.v, got, c.want)
		}
		if got := num(c.v); strings.ContainsAny(got, "e+") {
			t.Errorf("num(%v) = %q: an exponent", c.v, got)
		}
	}
}
