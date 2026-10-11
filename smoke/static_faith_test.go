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
	// Stonehenge, every town is given far more faith than a moderate set makes.
	for _, r := range rows {
		if r.Age == "iron_age" && r.Given < 3*r.ModerateSet {
			t.Errorf("iron_age: every town is given %v faith a tick against a moderate set's %v; the check means to walk Stonehenge's 0.6", r.Given, r.ModerateSet)
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
	// 3. The fill of the store faith is kept in: nobody leaves the bottom
	// band.
	store := rulesBroken(func(held float64, _ game.FaithSave, in faithIncome) float64 {
		return capped(held / MaxStorage(in.age, "faith"))
	})
	for _, age := range ages[3:] {
		if store[FaithRuleModerate][age] || !store[FaithRuleTwice][age] || !store[FaithRuleTop][age] {
			t.Errorf("the fill of the store in %s: problems moderate %v, twice %v, top %v; want the devoted towns stuck in the bottom band", age, store[FaithRuleModerate][age], store[FaithRuleTwice][age], store[FaithRuleTop][age])
		}
	}
	// 4. Full strength set too low or too high: the moderate town reaches
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

// TestFaithNeverFilledTheGeneralStore keeps the arithmetic that retired the
// old rule. The faith bands read faith as a share of its storage, but faith
// has no store of its own: it is kept in the general one, and no building
// stores faith alone. Against the least of that store anyone can hold, a
// devoted economy (three times the moderate one) saving through a doom's
// longest warning, and a moderate one that has saved every tick since the
// game began, both stayed under the middle band's 25% in every age; against
// a typical store they stayed under 7%, and under 1% from the Renaissance
// Age on. So every roll read the bottom band.
//
// On the walled stores (the storage rule) that holds in every age but one.
// The Iron Age's least store is 31.9K, the five Warehouses' and no more, and
// a devoted economy through the longest warning holds 59% of it and a
// moderate hoard 35%: the old measure could have reached the middle band
// there, and 11% of a typical store. The test says so, in that age only, so
// a store that moves it either way is seen. The rule it retired stays
// retired: faith has a store of its own.
func TestFaithNeverFilledTheGeneralStore(t *testing.T) {
	for _, d := range config.BaseBuildings() {
		for _, e := range d.Effects {
			if e.Type == "storage" && e.Target == "faith" {
				t.Errorf("%s stores faith alone: faith has a store of its own now, and this test's premise needs another look", d.Key)
			}
		}
	}
	least, typical := leastGeneralStore(), typicalGeneralStore()
	income := rules.Core().FlowIncome
	type ageOf struct{ Epoch, Age string }
	var all []ageOf
	for _, a := range config.AgeOrder() {
		all = append(all, ageOf{config.EpochForAge(a), a})
	}
	hoard := 0.0 // a moderate economy's faith since the game began, never spent
	t.Log("| era | age | least store | typical store | most store | moderate, shortest warning | of least | of typical | devoted, longest warning | of least | of typical | moderate hoard | of least | of typical |")
	for _, m := range all {
		ageMakes := income("faith", m.Age) * config.AgeTargetTicks(m.Age)
		hoard += ageMakes
		moderate := ageMakes * 0.2    // a doom's shortest warning: a fifth of the age
		devoted := 3 * ageMakes * 0.6 // three times the moderate economy, the longest warning: three fifths
		t.Logf("| %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s |", m.Epoch, m.Age,
			num(least[m.Age]), num(typical[m.Age]), num(MaxStorage(m.Age, "faith")),
			faithAmount(moderate), sharePct(moderate/least[m.Age]), sharePct(moderate/typical[m.Age]),
			faithAmount(devoted), sharePct(devoted/least[m.Age]), sharePct(devoted/typical[m.Age]),
			faithAmount(hoard), sharePct(hoard/least[m.Age]), sharePct(hoard/typical[m.Age]))
		if least[m.Age] > typical[m.Age] && m.Age != "transcendent_age" {
			// The last gate asks for more than five copies of each vault hold.
			t.Errorf("%s: the least store (%v) is over the typical one (%v)", m.Age, least[m.Age], typical[m.Age])
		}
		for _, c := range []struct {
			name  string
			faith float64
		}{{"a devoted economy through the longest warning", devoted}, {"a moderate economy's hoard since the start", hoard}} {
			reaches := m.Age == "iron_age" // the one age whose walled least store is that small
			if share := c.faith / least[m.Age]; (share >= game.FaithMidAt) != reaches {
				t.Errorf("%s: %s holds %v faith, %.3f of the least store (%v): the old measure reaching the middle band here is %v, want %v", m.Age, c.name, c.faith, share, least[m.Age], share >= game.FaithMidAt, reaches)
			}
			limit := 0.07
			if config.AgePositions(config.AgeOrder())[m.Age] >= config.AgePositions(config.AgeOrder())["renaissance_age"] {
				limit = 0.01
			}
			if reaches {
				limit = 0.12 // a devoted economy holds 0.1145 of the Iron Age's typical store
			}
			if share := c.faith / typical[m.Age]; share >= limit && m.Age != "transcendent_age" {
				t.Errorf("%s: %s holds %.4f of a typical store, want under %v", m.Age, c.name, share, limit)
			}
		}
	}
}
