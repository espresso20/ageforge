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
