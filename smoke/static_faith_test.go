package smoke

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/rules"
)

// TestFaithStandingReachable: in every age a moderate faith economy that
// saves through a doom's shortest warning reads as the middle faith band and
// a devoted one as the top, by the rule the rolls use. The check fails on a
// measure no economy can fill, which is what the bands were read against
// until now.
func TestFaithStandingReachable(t *testing.T) {
	problems, rows := StaticFaithStanding()
	for _, p := range problems {
		t.Errorf("%s", p)
	}
	ages := config.AgeOrder()
	if len(rows) != len(ages) {
		t.Fatalf("%d rows for %d ages", len(rows), len(ages))
	}
	for i, r := range rows {
		if r.Age != ages[i] || r.Epoch != config.EpochForAge(r.Age) || r.Full <= 0 || r.Moderate <= 0 {
			t.Errorf("row %d: %+v, want %s with a measure and a faith income", i, r, ages[i])
		}
		// A doom's shortest warning is a third of its longest, which is what
		// full standing is measured on.
		if r.ModerateStanding < 0.33 || r.ModerateStanding > 0.34 || r.ModerateBand != string(game.FaithBandMid) {
			t.Errorf("%s: a moderate economy reaches %.3f (%s), want a third of full standing, the middle band", r.Age, r.ModerateStanding, r.ModerateBand)
		}
		if r.DevotedStanding < 0.99 || r.DevotedBand != string(game.FaithBandHigh) {
			t.Errorf("%s: a devoted economy reaches %.3f (%s), want full standing, the top band", r.Age, r.DevotedStanding, r.DevotedBand)
		}
	}
	var sb strings.Builder
	writeFaithStanding(&sb, problems, rows)
	if out := sb.String(); !strings.Contains(out, "No problems.") || strings.Count(out, "\n| ") != len(ages)+1 {
		t.Errorf("the report section:\n%s", out)
	}

	// The check fires on broken measures.
	measures := game.FaithMeasures()
	income := rules.Core().FlowIncome
	alter := func(full func(m game.FaithMeasure) float64) []FaithStandingProblem {
		broken := make([]game.FaithMeasure, len(measures))
		for i, m := range measures {
			m.Full = full(m)
			broken[i] = m
		}
		got, _ := faithStandingProblems(broken, income, FaithDevotedFactor)
		return got
	}
	want := func(name string, got []FaithStandingProblem, rule string) {
		t.Helper()
		if len(got) != len(ages) {
			t.Errorf("%s: %d problem(s), want one for each of the %d ages: %v", name, len(got), len(ages), got)
			return
		}
		for i, p := range got {
			if p.Rule != rule || p.Age != ages[i] || p.String() == "" {
				t.Errorf("%s: problem %d is %+v, want rule %s in %s", name, i, p, rule, ages[i])
			}
		}
	}
	// The old measure: the store faith is kept in, here at the least a
	// player can hold (the storage the age gates have forced by then).
	least := leastGeneralStore()
	want("the least general store", alter(func(m game.FaithMeasure) float64 { return least[m.Age] }), FaithRuleMiddle)
	want("the most general store", alter(func(m game.FaithMeasure) float64 { return MaxStorage(m.Age, "faith") }), FaithRuleMiddle)
	// What the whole age makes: a shortest warning is a fifth of it.
	want("the whole age's faith", alter(func(m game.FaithMeasure) float64 {
		return income("faith", m.Age) * config.AgeTargetTicks(m.Age)
	}), FaithRuleMiddle)
	want("no measure", alter(func(game.FaithMeasure) float64 { return 0 }), FaithRuleMeasure)
	// A devoted economy only twice the moderate one reaches two thirds of
	// full standing, which is still the middle band.
	twice, _ := faithStandingProblems(measures, income, 2)
	want("a devoted economy of twice the moderate", twice, FaithRuleTop)
}

// leastGeneralStore is, per age, the least general storage a player can be
// holding on entering it: the base store plus the most any age gate so far
// has forced them to hold at once (ladderForced; storage is never lost).
// Faith is kept in the general store.
func leastGeneralStore() map[string]float64 {
	ages, defs := config.Ages(), config.BuildingByKey()
	base := 0.0
	for _, r := range config.BaseResources() {
		if r.Key == "faith" {
			base = r.BaseStorage
		}
	}
	out := map[string]float64{}
	least := base
	for i, a := range ages {
		if i > 0 {
			if forced, _ := ladderForced(ages[i-1], a, defs); forced+base > least {
				least = forced + base
			}
		}
		out[a.Key] = least
	}
	return out
}

// typicalGeneralStore is, per age, the general storage of a moderate builder:
// the base store and config.FlowCopies copies of every storage building up
// to that age.
func typicalGeneralStore() map[string]float64 {
	idx := map[string]int{}
	for i, a := range config.AgeOrder() {
		idx[a] = i
	}
	out := map[string]float64{}
	for i, a := range config.AgeOrder() {
		total := 0.0
		for _, r := range config.BaseResources() {
			if r.Key == "faith" {
				total = r.BaseStorage
			}
		}
		for _, d := range config.BaseBuildings() {
			if j, ok := idx[d.RequiredAge]; !ok || j > i || d.Category != "storage" {
				continue
			}
			for _, e := range d.Effects {
				if e.Type == "storage" && (e.Target == "all" || e.Target == "faith") {
					total += e.Value * config.FlowCopies
				}
			}
		}
		out[a] = total
	}
	return out
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
	hoard := 0.0 // a moderate economy's faith since the game began, never spent
	t.Log("| era | age | least store | typical store | most store | moderate, shortest warning | of least | of typical | devoted, longest warning | of least | of typical | moderate hoard | of least | of typical |")
	for _, m := range game.FaithMeasures() {
		ageMakes := income("faith", m.Age) * config.AgeTargetTicks(m.Age)
		hoard += ageMakes
		moderate := income("faith", m.Age) * m.WarningTicks
		devoted := FaithDevotedFactor * ageMakes * 0.6 // the longest warning: three fifths of the age
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
			if share := c.faith / least[m.Age]; share >= game.FaithMidAt {
				t.Errorf("%s: %s holds %v faith, %.3f of the least store (%v): the old measure could reach the middle band here", m.Age, c.name, c.faith, share, least[m.Age])
			}
			limit := 0.07
			if config.AgePositions(config.AgeOrder())[m.Age] >= config.AgePositions(config.AgeOrder())["renaissance_age"] {
				limit = 0.01
			}
			if share := c.faith / typical[m.Age]; share >= limit && m.Age != "transcendent_age" {
				t.Errorf("%s: %s holds %.4f of a typical store, want under %v", m.Age, c.name, share, limit)
			}
		}
	}
}
