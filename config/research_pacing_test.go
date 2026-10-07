package config

import (
	"math"
	"testing"
)

// TestResearchTimeByKind: a tech's research time is its kind's share of its
// age's research cap (the age target ÷ 8): spine 40%, keystone 50%, optional
// 50%, capstone 80%. Nothing typed on a tech changes it.
func TestResearchTimeByKind(t *testing.T) {
	techs := Technologies()
	kinds := TechKinds(techs, BaseBuildings())
	for _, tech := range techs {
		limit := AgeTargetTicks(tech.Age) / ResearchTimeDivisor
		want := max(int(math.Round(float64(limit*ResearchTimeShare(kinds[tech.Key])))), 1)
		if tech.ResearchTicks != want {
			t.Errorf("%s (%s, %s) researches in %d ticks, want %d: %.0f%% of the age's cap of %.0f",
				tech.Key, kinds[tech.Key], tech.Age, tech.ResearchTicks, want, 100*ResearchTimeShare(kinds[tech.Key]), limit)
		}
	}
	shares := map[TechKind]float64{TechSpine: 0.4, TechKeystone: 0.5, TechOptional: 0.5, TechCapstone: 0.8, "": 0.5}
	for kind, want := range shares {
		if got := ResearchTimeShare(kind); got != want {
			t.Errorf("a %q tech takes %v of the cap, want %v", kind, got, want)
		}
	}
	// On a fixture: a typed time is not read, and each kind gets its share.
	age := "classical_age"
	limit := AgeTargetTicks(age) / ResearchTimeDivisor
	fixture := []TechDef{
		{Key: "s", Age: age, ResearchTicks: 7},
		{Key: "k", Age: age, ResearchTicks: 999999},
		{Key: "o", Age: age},
		{Key: "c", Age: age},
		{Key: "elsewhere", Age: "no_such_age", ResearchTicks: 123},
	}
	out := normalizeResearchTicks(fixture, map[string]TechKind{"s": TechSpine, "k": TechKeystone, "o": TechOptional, "c": TechCapstone})
	for i, share := range []float64{0.4, 0.5, 0.5, 0.8} {
		if want := int(math.Round(float64(limit * share))); out[i].ResearchTicks != want {
			t.Errorf("%s researches in %d ticks, want %d", out[i].Key, out[i].ResearchTicks, want)
		}
	}
	if out[4].ResearchTicks != 123 {
		t.Errorf("a tech of an age with no target was re-timed to %d ticks", out[4].ResearchTicks)
	}
}

// TestResearchCostWeights: the budget's weights by kind.
func TestResearchCostWeights(t *testing.T) {
	weights := map[TechKind]float64{TechSpine: 0.6, TechKeystone: 0.8, TechOptional: 1.0, TechCapstone: 1.6, "": 1.0}
	for kind, want := range weights {
		if got := ResearchCostWeight(kind); got != want {
			t.Errorf("a %q tech weighs %v, want %v", kind, got, want)
		}
	}
}

// TestResearchCostsShareTheAgesBudget: an age's techs cost its budget
// together (to the rounding of each price), each in proportion to its kind's
// weight, and the budget is the age's share of the knowledge it makes in its
// target time.
func TestResearchCostsShareTheAgesBudget(t *testing.T) {
	techs := Technologies()
	kinds := TechKinds(techs, BaseBuildings())
	total := map[string]float64{}
	weight := map[string]float64{}
	count := map[string]int{}
	for _, tech := range techs {
		total[tech.Age] += tech.Cost
		weight[tech.Age] += ResearchCostWeight(kinds[tech.Key])
		count[tech.Age]++
	}
	for _, age := range AgeOrder() {
		hours := AgeTargets[age].Hours()
		if KnowledgePerHour[age] <= 0 || hours <= 0 {
			t.Errorf("%s has no knowledge income or no target to size its research from", age)
			continue
		}
		share := ResearchBudgetShare
		switch age {
		case "primitive_age":
			share = 0.5
		case "transcendent_age":
			share = 0.3
		}
		budget := float64(KnowledgePerHour[age] * hours * share)
		if got := ResearchBudget(age); math.Abs(got/budget-1) > 1e-12 {
			t.Errorf("%s: budget %v, want %v: %.0f%% of %v knowledge an hour for %v hours", age, got, budget, 100*share, KnowledgePerHour[age], hours)
		}
		if count[age] == 0 {
			t.Errorf("%s has no techs", age)
			continue
		}
		// Each price is rounded to three figures and a whole number: half
		// a percent and half a knowledge each at most.
		if math.Abs(total[age]-budget) > 0.005*budget+0.5*float64(count[age]) {
			t.Errorf("%s: its %d techs cost %v together, want its budget of %v", age, count[age], total[age], budget)
		}
	}
	for _, tech := range techs {
		want := float64(ResearchBudget(tech.Age) * ResearchCostWeight(kinds[tech.Key]) / weight[tech.Age])
		if math.Abs(tech.Cost-want) > 0.005*want+0.5 {
			t.Errorf("%s (%s) costs %v, want %v: its weight's share of the %s budget", tech.Key, kinds[tech.Key], tech.Cost, want, tech.Age)
		}
		if tech.Cost < 1 || tech.Cost != math.Round(tech.Cost) {
			t.Errorf("%s costs %v: a price is a whole number of knowledge, 1 at least", tech.Key, tech.Cost)
		}
	}
	// What a run must research is the cheap part of every age: no keystone
	// or spine tech costs more than an optional tech of its age. A capstone
	// is the dear one: ResearchCostCapstone times an optional tech's price.
	optional := map[string]float64{}
	for _, tech := range techs {
		if kinds[tech.Key] == TechOptional {
			optional[tech.Age] = tech.Cost
		}
	}
	capstones := 0
	for _, tech := range techs {
		o, ok := optional[tech.Age]
		switch {
		case !ok || kinds[tech.Key] == TechOptional:
		case kinds[tech.Key] == TechCapstone:
			capstones++
			if want := float64(o * ResearchCostCapstone); math.Abs(tech.Cost-want) > 0.01*want {
				t.Errorf("%s (capstone) costs %v, want %v: %g times an optional tech of its age (%v)", tech.Key, tech.Cost, want, ResearchCostCapstone, o)
			}
		case tech.Cost >= o:
			t.Errorf("%s (%s) costs %v, no less than an optional tech of its age (%v)", tech.Key, kinds[tech.Key], tech.Cost, o)
		}
	}
	if capstones == 0 {
		t.Error("no capstone was priced: the tree has two from the Medieval Age on")
	}
}

// TestCapstonesAreTheLongOnes: a capstone takes ResearchTimeCapstone of its
// age's research cap, longer than any other tech of the age, and still
// inside the cap.
func TestCapstonesAreTheLongOnes(t *testing.T) {
	techs := Technologies()
	kinds := TechKinds(techs, BaseBuildings())
	longest := map[string]int{}
	for _, tech := range techs {
		if kinds[tech.Key] != TechCapstone {
			longest[tech.Age] = max(longest[tech.Age], tech.ResearchTicks)
		}
	}
	seen := 0
	for _, tech := range techs {
		if kinds[tech.Key] != TechCapstone {
			continue
		}
		seen++
		limit := ResearchCapTicks(tech.Age)
		if want := int(math.Round(float64(limit * ResearchTimeCapstone))); tech.ResearchTicks != want {
			t.Errorf("%s takes %d ticks, want %d: %g of the %s research cap (%g ticks)", tech.Key, tech.ResearchTicks, want, ResearchTimeCapstone, tech.Age, limit)
		}
		if tech.ResearchTicks <= longest[tech.Age] || float64(tech.ResearchTicks) > limit {
			t.Errorf("%s takes %d ticks: the longest other tech of its age takes %d and the cap is %g", tech.Key, tech.ResearchTicks, longest[tech.Age], limit)
		}
	}
	if seen != 2 {
		t.Errorf("%d capstones, want 2: Scholasticism and Guilds", seen)
	}
}

// TestAddingATechLeavesTheAgesTotalAlone is the budget rule's promise: add
// a tech to an age and every tech already in it gets a little cheaper, while
// the age's total stays where it was. On the real table, with one more
// optional tech in each age in turn, and on a fixture with every kind.
func TestAddingATechLeavesTheAgesTotalAlone(t *testing.T) {
	sum := func(techs []TechDef, age string) float64 {
		s := 0.0
		for _, tech := range techs {
			if tech.Age == age {
				s += tech.Cost
			}
		}
		return s
	}
	base := Technologies()
	kinds := TechKinds(base, BaseBuildings())
	for _, age := range AgeOrder() {
		// Copied: the normalizer writes prices into the slice it is given.
		more := append(append([]TechDef(nil), base...), TechDef{Key: "zz_new", Age: age})
		more = normalizeResearchCosts(more, kinds)
		before, after := sum(base, age), sum(more, age)
		// The total stays put, to the rounding of each price (three
		// figures and a whole number: half a percent and half a
		// knowledge each at most, before and after).
		n := 0
		for _, tech := range more {
			if tech.Age == age {
				n++
			}
		}
		if math.Abs(after-before) > 0.01*before+float64(n) {
			t.Errorf("%s: one more tech moved the age's total from %v to %v", age, before, after)
		}
		for i, tech := range base {
			switch {
			case tech.Age == age && more[i].Cost >= tech.Cost:
				t.Errorf("%s: one more tech in its age left %s at %v (was %v); it should be cheaper", age, tech.Key, more[i].Cost, tech.Cost)
			case tech.Age != age && more[i].Cost != tech.Cost:
				t.Errorf("%s: one more tech there moved %s (%s) from %v to %v", age, tech.Key, tech.Age, tech.Cost, more[i].Cost)
			}
		}
		if got := more[len(more)-1].Cost; got <= 0 {
			t.Errorf("%s: the new tech has no price (%v)", age, got)
		}
	}

	// A fixture age with every kind: the shares are exact before rounding.
	age := "iron_age"
	fixKinds := map[string]TechKind{"s": TechSpine, "k": TechKeystone, "o": TechOptional, "c": TechCapstone, "n": TechOptional}
	four := normalizeResearchCosts([]TechDef{{Key: "s", Age: age, Cost: 1}, {Key: "k", Age: age, Cost: 2}, {Key: "o", Age: age}, {Key: "c", Age: age}}, fixKinds)
	five := normalizeResearchCosts([]TechDef{{Key: "s", Age: age}, {Key: "k", Age: age}, {Key: "o", Age: age}, {Key: "c", Age: age}, {Key: "n", Age: age}}, fixKinds)
	budget := ResearchBudget(age)
	for i, w := range []float64{0.6, 0.8, 1.0, 1.6} {
		if want := float64(budget * w / 4.0); math.Abs(four[i].Cost/want-1) > 0.005 {
			t.Errorf("fixture of four: %s costs %v, want %v", four[i].Key, four[i].Cost, want)
		}
		if want := float64(budget * w / 5.0); math.Abs(five[i].Cost/want-1) > 0.005 {
			t.Errorf("fixture of five: %s costs %v, want %v", five[i].Key, five[i].Cost, want)
		}
		if five[i].Cost >= four[i].Cost {
			t.Errorf("fixture: %s costs %v with a fifth tech in its age, no less than the %v before", four[i].Key, five[i].Cost, four[i].Cost)
		}
	}
	if a, b := sum(four, age), sum(five, age); math.Abs(a/budget-1) > 0.005 || math.Abs(b/budget-1) > 0.005 {
		t.Errorf("fixture: totals %v and %v, want the budget of %v both times", a, b, budget)
	}
	// A tech of an age with no budget keeps the price it came with.
	if out := normalizeResearchCosts([]TechDef{{Key: "x", Age: "no_such_age", Cost: 42}}, nil); out[0].Cost != 42 {
		t.Errorf("a tech of an age with no budget was re-priced to %v", out[0].Cost)
	}
}

// TestNoGateAsksForKnowledge: knowledge left the age gates when the wonders
// got their keystones. An age's knowledge goes to research, and the one tech
// the advance needs is the wonder's.
func TestNoGateAsksForKnowledge(t *testing.T) {
	for _, a := range Ages() {
		if v, ok := a.ResourceReqs["knowledge"]; ok {
			t.Errorf("the gate into the %s asks for %v knowledge", a.Name, v)
		}
	}
}

// TestKnowledgePerHourCoversEveryAge: research budgets are sized from the
// table, so every age needs a row, and no row names an age that is gone.
func TestKnowledgePerHourCoversEveryAge(t *testing.T) {
	ages := map[string]bool{}
	for _, a := range AgeOrder() {
		ages[a] = true
		if KnowledgePerHour[a] <= 0 {
			t.Errorf("%s has no knowledge income in KnowledgePerHour", a)
		}
	}
	for a := range KnowledgePerHour {
		if !ages[a] {
			t.Errorf("KnowledgePerHour lists %s, which is not an age", a)
		}
	}
	for a := range researchBudgetShares {
		if !ages[a] {
			t.Errorf("researchBudgetShares lists %s, which is not an age", a)
		}
	}
}
