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
	made := KnowledgeMade()
	for i, age := range config.Ages() {
		made[i] /= config.LatePriceMultiplier(age.Key)
	}
	ladders := KnowledgeLadders(techs, made, func(string) float64 { return KnowledgeLadderRatio })
	byKey := map[string]config.TechDef{}
	for _, tc := range techs {
		byKey[tc.Key] = tc
	}
	for _, l := range ladders {
		for j, k := range l.Techs {
			got := byKey[k].Cost
			first := l.Total * (l.Ratio - 1) / (math.Pow(l.Ratio, float64(len(l.Techs))) - 1)
			want := first * math.Pow(l.Ratio, float64(j)) * config.LatePriceMultiplier(l.Age)
			// Authored prices are rounded to three significant figures and the late
			// price climb multiplies them, so a price sits within 1% of its ladder
			// rung. Rungs are 30% apart, so this still pins each tech to its rung.
			if math.Abs(got-want)/want > 0.01 {
				t.Errorf("%s: %s runtime cost %g, the climbed unrounded ladder gives %g", l.Age, k, got, want)
			}
		}
		if n := len(l.Prices); n > 0 {
			if sum := sumOf(l.Prices) * config.LatePriceMultiplier(l.Age); math.Abs(sum/(l.Total*config.LatePriceMultiplier(l.Age))-1) > 0.02 {
				t.Errorf("%s: the techs total %g, the ladder asks %g", l.Age, sum, l.Total*config.LatePriceMultiplier(l.Age))
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
