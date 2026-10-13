package config

import "testing"

// TestEveryKnowledgeBuildingClimbsAtTheKnowledgeRate: R1. Each class is held
// to its own rate: producers to ProducerRate (town_rule_test.go), knowledge
// buildings to KnowledgeRate.
func TestEveryKnowledgeBuildingClimbsAtTheKnowledgeRate(t *testing.T) {
	seen := 0
	for _, d := range BaseBuildings() {
		if !IsKnowledgeBuilding(d) {
			continue
		}
		seen++
		if d.CostScale != KnowledgeRate {
			t.Errorf("%s makes knowledge and climbs at %g a copy, the knowledge rate is %g", d.Key, d.CostScale, KnowledgeRate)
		}
	}
	if seen != len(KnowledgeGateWas()) {
		t.Errorf("%d knowledge buildings, KnowledgeGateWas lists %d", seen, len(KnowledgeGateWas()))
	}
}

// TestKnowledgeGatesAskTheSameSpend: R2. The gate out of an age asks for the
// count of each knowledge building whose total price at KnowledgeRate is
// nearest the total the old count cost at ProducerRate (a tie goes to the
// lower count). Prices climb by one factor in every part of a price, so the
// totals are compared in units of the first copy.
func TestKnowledgeGatesAskTheSameSpend(t *testing.T) {
	ages := Ages()
	gate := map[string]int{}
	for _, a := range ages {
		for k, n := range a.BuildingReqs {
			gate[k] = n
		}
	}
	for k, was := range KnowledgeGateWas() {
		want := KnowledgeGateCount(was)
		if gate[k] != want {
			t.Errorf("%s: the gate asks %d copies; %d copies at %g (the spend of %d at %g) is %d", k, gate[k], want, KnowledgeRate, was, ProducerRate, want)
		}
		// nearest, by brute force, ties lower
		spend := climbTotal(ProducerRate, was)
		best, gap := 0, 1e18
		for n := 1; n < 200; n++ {
			g := climbTotal(KnowledgeRate, n) - spend
			if g < 0 {
				g = -g
			}
			if g < gap {
				best, gap = n, g
			}
		}
		if best != want {
			t.Errorf("%s: KnowledgeGateCount(%d) = %d, brute force says %d", k, was, want, best)
		}
	}
}

func TestKnowledgeGateCountTieGoesLower(t *testing.T) {
	// 1 copy costs 1 at any rate: the count for a spend of 1 is 1.
	if got := KnowledgeGateCount(1); got != 1 {
		t.Errorf("KnowledgeGateCount(1) = %d", got)
	}
}

// TestKnowledgeBaseIsTheFirstTech: R3's base is what a town needs to hold the
// first age's cheapest tech before it owns any knowledge building: that
// tech's price, rounded up to two significant figures.
func TestKnowledgeBaseIsTheFirstTech(t *testing.T) {
	ages, techs := Ages(), Technologies()
	if got := KnowledgeBaseFor(ages, techs); got != KnowledgeBase {
		t.Errorf("KnowledgeBase is %g, the first age's cheapest tech rounds up to %g", KnowledgeBase, got)
	}
	cheapest := 0.0
	for _, tc := range techs {
		if tc.Age == ages[0].Key && (cheapest == 0 || tc.Cost < cheapest) {
			cheapest = tc.Cost
		}
	}
	if KnowledgeBase < cheapest {
		t.Errorf("the base %g cannot hold the first age's cheapest tech, %g", KnowledgeBase, cheapest)
	}
	for _, r := range BaseResources() {
		if r.Key == "knowledge" && r.BaseStorage != KnowledgeBase {
			t.Errorf("the knowledge resource starts with %g of storage, the knowledge base is %g", r.BaseStorage, KnowledgeBase)
		}
	}
}

// TestKnowledgeStoreFollowsTheRule: R3. What one copy of a knowledge building
// holds is the figure the rule gives: the gate's copies of the age's knowledge
// buildings, on top of what the earlier ages' required copies hold, hold the
// larger of the age's dearest tech plus a quarter and 1.2 times the second-
// cheapest tech of the next age, rounded up to two significant figures; buildings
// of one age hold equal amounts.
func TestKnowledgeStoreFollowsTheRule(t *testing.T) {
	defs, ages, techs := BaseBuildings(), Ages(), Technologies()
	rows := KnowledgeStoreRows(defs, ages, techs, KnowledgeBase)
	held := KnowledgeStoreHeld(defs, ages)
	seen := 0
	for i, r := range rows {
		inAge := 0
		for _, d := range defs {
			if d.RequiredAge != r.Age || !IsKnowledgeBuilding(d) {
				continue
			}
			inAge++
			seen++
			if d.KnowledgeStore != r.Figure {
				t.Errorf("%s holds %g knowledge a copy, the rule gives %g", d.Key, d.KnowledgeStore, r.Figure)
			}
		}
		if inAge == 0 {
			if r.Copies != 0 || r.Figure != 0 {
				t.Errorf("%s has no knowledge building and the rule gives it copies %d, figure %g", r.Age, r.Copies, r.Figure)
			}
			// No building of the age can raise the store: the earlier ages'
			// must already hold what the rule asks.
			if r.Need > r.Before {
				t.Errorf("%s has no knowledge building and its store, %g, is under the %g the rule asks", r.Age, r.Before, r.Need)
			}
			continue
		}
		if r.Figure <= 0 {
			t.Errorf("%s: the earlier ages already hold %g, the rule asks %g: it gives no figure to author", r.Age, r.Before, r.Need)
		}
		if held[r.Age] < r.Need {
			t.Errorf("%s: the gate's copies hold %g, under the %g the rule asks", r.Age, held[r.Age], r.Need)
		}
		if got := rows[i].Held; got != held[r.Age] {
			t.Errorf("%s: the rule's store is %g, the tables' is %g", r.Age, got, held[r.Age])
		}
	}
	if seen != len(KnowledgeGateWas()) {
		t.Errorf("%d knowledge buildings checked, %d exist", seen, len(KnowledgeGateWas()))
	}
}

// TestNothingButKnowledgeBuildingsRaiseTheKnowledgeStore: R3. The figure is
// on knowledge buildings alone.
func TestNothingButKnowledgeBuildingsRaiseTheKnowledgeStore(t *testing.T) {
	for _, d := range BaseBuildings() {
		if d.KnowledgeStore != 0 && !IsKnowledgeBuilding(d) {
			t.Errorf("%s holds %g knowledge and is not a knowledge building", d.Key, d.KnowledgeStore)
		}
	}
}
