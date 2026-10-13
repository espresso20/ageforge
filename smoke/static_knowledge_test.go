package smoke

import (
	"math"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// TestKnowledgeTechPricesFollowTheLadder: R4. Each age's techs are priced to
// total KnowledgeLadderShare of what the ordinary reference player's
// knowledge buildings make during his stay (their scholars, all researching),
// climbing by KnowledgeLadderRatio from the cheapest to the dearest in the
// order the table has them, each to three significant figures. The last age is
// the one named exception (KnowledgeLadderHeld).
func TestKnowledgeTechPricesFollowTheLadder(t *testing.T) {
	techs := config.Technologies()
	ladders := KnowledgeLadders(techs, KnowledgeMade(), func(string) float64 { return KnowledgeLadderRatio })
	byKey := map[string]config.TechDef{}
	for _, tc := range techs {
		byKey[tc.Key] = tc
	}
	for _, l := range ladders {
		if _, stopped := KnowledgeLadderHeld[l.Age]; stopped {
			continue
		}
		for j, k := range l.Techs {
			if got := byKey[k].Cost; got != l.Prices[j] {
				t.Errorf("%s: %s costs %g, the ladder gives %g", l.Age, k, got, l.Prices[j])
			}
		}
		if n := len(l.Prices); n > 0 {
			if sum := sumOf(l.Prices); math.Abs(sum/l.Total-1) > 0.02 {
				t.Errorf("%s: the techs total %g, the ladder asks %g", l.Age, sum, l.Total)
			}
		}
	}
}

func sumOf(v []float64) float64 {
	s := 0.0
	for _, x := range v {
		s += x
	}
	return s
}

// TestKnowledgeLadderHasOneNamedException: the single exception is the last
// age, its techs keep the held prices, property 10 reads what is recorded for
// it and holds in every other age, and the Research Covenant, which holds in
// every age, is checked for it by name.
func TestKnowledgeLadderHasOneNamedException(t *testing.T) {
	ages := config.Ages()
	last := ages[len(ages)-1].Key
	if len(KnowledgeLadderHeld) != 1 || KnowledgeLadderHeld[last].Reason == "" {
		t.Fatalf("the ladder's exceptions are %v, want the last age (%s) alone, with a reason", KnowledgeLadderHeld, last)
	}
	made := KnowledgeMade()
	for i, a := range ages {
		sum := 0.0
		for _, tc := range config.Technologies() {
			if tc.Age == a.Key {
				sum += tc.Cost
			}
		}
		share := sum / made[i]
		if h, held := KnowledgeLadderHeld[a.Key]; held {
			if math.Abs(sum-h.TechSum) > 1e6 {
				t.Errorf("%s: its techs cost %g, held at %g", a.Key, sum, h.TechSum)
			}
			if share >= propResearchLow && share <= propResearchHigh || math.Abs(share-h.Property10) > 0.0005 {
				t.Errorf("%s: property 10 reads %.4f, the exception records %.4f (outside %g to %g)", a.Key, share, h.Property10, propResearchLow, propResearchHigh)
			}
			continue
		}
		if share < propResearchLow || share > propResearchHigh {
			t.Errorf("%s: property 10 reads %.3f, outside %g to %g, and the age is not the named exception", a.Key, share, propResearchLow, propResearchHigh)
		}
	}
	for _, r := range StaticResearch() {
		if p := r.Problems(); len(p) > 0 {
			t.Errorf("the Research Covenant breaks in %s: %v", r.Age, p)
		}
	}
	for n := range KnownPropertyReasons {
		if _, listed := KnownPropertyFailures[n]; !listed {
			t.Errorf("property %d has a recorded reason and is not a known failure", n)
		}
	}
	for _, p := range StaticProperties().List {
		if p.N == 10 && p.Holds {
			t.Error("property 10 holds: the named exception no longer fails it, take it off the list")
		}
	}
}
