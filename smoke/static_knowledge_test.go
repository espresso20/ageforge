package smoke

import (
	"math"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// TestKnowledgeTechPricesFollowTheLadder: R4. Each age's techs are priced to
// total config.ResearchShare of what the ordinary reference player's
// knowledge buildings make during his stay (their scholars, all researching),
// climbing by KnowledgeLadderRatio from the cheapest to the dearest in the
// order the table has them, each to three significant figures.
func TestKnowledgeTechPricesFollowTheLadder(t *testing.T) {
	techs := config.Technologies()
	ladders := KnowledgeLadders(techs, KnowledgeMade(), func(string) float64 { return KnowledgeLadderRatio })
	byKey := map[string]config.TechDef{}
	for _, tc := range techs {
		byKey[tc.Key] = tc
	}
	for _, l := range ladders {
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

// TestKnowledgeLadderKeepsResearchShare checks every age, including the
// last, against the research-share rule and the Research Covenant.
func TestKnowledgeLadderKeepsResearchShare(t *testing.T) {
	made := KnowledgeMade()
	for i, a := range config.Ages() {
		sum := 0.0
		for _, tc := range config.Technologies() {
			if tc.Age == a.Key {
				sum += tc.Cost
			}
		}
		share := sum / made[i]
		low, high := researchBand(a.Key)
		if share < low || share > high {
			t.Errorf("%s: property 10 reads %.3f, outside %g to %g, and the age is not the named exception", a.Key, share, low, high)
		}
	}
	for _, r := range StaticResearch() {
		if p := r.Problems(); len(p) > 0 {
			t.Errorf("the Research Covenant breaks in %s: %v", r.Age, p)
		}
	}
}
