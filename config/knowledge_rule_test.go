package config

import (
	"math"
	"testing"
)

// scholarsPerCopy is the crew each knowledge building was staffed with, and
// is the scholars it comes with now (decision: a knowledge building comes with
// its scholars, the crew size it had).
var scholarsPerCopy = map[string]int{
	"story_circle": 2, "elders_hall": 2, "scriptorium": 3, "agora": 3, "library": 4,
	"monastery_library": 4, "university": 5, "natural_philosophy_hall": 5,
	"research_institute": 6, "academy": 6, "physics_laboratory": 7, "research_campus": 7,
	"think_tank": 8, "innovation_hub": 8, "ai_research_lab": 10, "neuro_research_center": 10,
	"theoretical_institute": 12, "deep_space_observatory": 12, "xenology_institute": 15,
	"cosmic_research_station": 15, "reality_academy": 20,
}

// TestKnowledgeBuildingsComeWithScholars: every knowledge building has the
// scholars it was staffed with, takes no workers, and no other building has
// scholars.
func TestKnowledgeBuildingsComeWithScholars(t *testing.T) {
	seen := 0
	for _, d := range BaseBuildings() {
		want, isK := scholarsPerCopy[d.Key]
		if IsKnowledgeBuilding(d) != isK {
			t.Errorf("%s: IsKnowledgeBuilding is %v, the table of scholars says %v", d.Key, IsKnowledgeBuilding(d), isK)
		}
		if d.Scholars != want {
			t.Errorf("%s comes with %d scholars, want %d", d.Key, d.Scholars, want)
		}
		if isK {
			seen++
			if d.WorkerCapacity != 0 || d.WorkerDomain != "" {
				t.Errorf("%s takes workers (%s x %d): its scholars come with it", d.Key, d.WorkerDomain, d.WorkerCapacity)
			}
		}
	}
	if seen != len(scholarsPerCopy) {
		t.Errorf("%d knowledge buildings, the table lists %d", seen, len(scholarsPerCopy))
	}
}

// TestKnowledgeCapIsTheGate: a knowledge building cannot be built past the
// count its gate asks for, and that count is the one the age tables author.
func TestKnowledgeCapIsTheGate(t *testing.T) {
	ages := Ages()
	for _, d := range BaseBuildings() {
		if !IsKnowledgeBuilding(d) {
			continue
		}
		cap := KnowledgeCap(d.Key)
		if d.MaxCount != cap || cap <= 0 {
			t.Errorf("%s: the most copies is %d, the cap is %d", d.Key, d.MaxCount, cap)
		}
		for i, a := range ages {
			if a.Key != d.RequiredAge || i+1 >= len(ages) {
				continue
			}
			if got := ages[i+1].BuildingReqs[d.Key]; got != cap {
				t.Errorf("%s: the gate out of %s asks for %d, the cap is %d", d.Key, a.Key, got, cap)
			}
		}
	}
	if KnowledgeCap("hut") != 0 {
		t.Error("a building that is not a knowledge building has a cap")
	}
}

// TestKnowledgeIsNotAResourceOrAPrice: nothing is priced in knowledge and no
// resource holds it.
func TestKnowledgeIsNotAResourceOrAPrice(t *testing.T) {
	if _, ok := ResourceByKey()["knowledge"]; ok {
		t.Error("knowledge is a resource")
	}
	for _, d := range BaseBuildings() {
		if _, ok := d.BaseCost["knowledge"]; ok {
			t.Errorf("%s is priced in knowledge", d.Key)
		}
	}
	for _, r := range BaseTradeRoutes() {
		if _, ok := r.Export["knowledge"]; ok {
			t.Errorf("route %s asks for knowledge", r.Key)
		}
	}
}

// TestRepricedBuildingsKeepTheirUnits: a building that was priced in
// knowledge has its other parts multiplied so that its price in the age's
// price units is what it was with the knowledge part; the age's price levels
// are the ones they were (they read the prices from before, repricedFrom).
func TestRepricedBuildingsKeepTheirUnits(t *testing.T) {
	defs := BaseBuildings()
	lv := priceLevels(defs)
	for _, d := range defs {
		want, ok := repricedUnits[d.Key]
		if !ok {
			continue
		}
		if got := priceUnits(d.BaseCost, lv[d.RequiredAge]); math.Abs(got/want-1) > 0.003 {
			t.Errorf("%s costs %.4f price units of %s, it cost %.4f with its knowledge part", d.Key, got, d.RequiredAge, want)
		}
	}
	for k := range repricedUnits {
		if _, ok := repricedFrom[k]; !ok {
			t.Errorf("%s has units recorded and no price from before", k)
		}
	}
}

// TestTeachingLadder: with every scholar teaching, the bonus is 10 percent in
// ages 1 to 3, 8 in 4 to 6, 6 in 7 to 9, 5 in 10 to 15 and 4 in 16 and later.
func TestTeachingLadder(t *testing.T) {
	for pos := 0; pos < len(Ages()); pos++ {
		place := pos + 1
		want := 0.04
		switch {
		case place <= 3:
			want = 0.10
		case place <= 6:
			want = 0.08
		case place <= 9:
			want = 0.06
		case place <= 15:
			want = 0.05
		}
		if got := TeachingBonus(pos); got != want {
			t.Errorf("age %d: teaching gives %v, want %v", place, got, want)
		}
	}
}
