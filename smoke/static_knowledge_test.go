package smoke

import (
	"math"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// TestKnowledgeTechPricesFollowTheLadder: R4. Each age's techs are priced to
// total KnowledgeLadderShare of what the ordinary reference player's
// knowledge buildings make during his stay, climbing by KnowledgeLadderRatio
// from the cheapest to the dearest in the order the table has them, each to
// three significant figures. The ratio fits: the dearest tech and a quarter
// is inside the age's knowledge store, and the cheapest is inside the store
// a town arrives with.
func TestKnowledgeTechPricesFollowTheLadder(t *testing.T) {
	techs := config.Technologies()
	ladders := KnowledgeLadders(techs, KnowledgeMade(), func(string) float64 { return KnowledgeLadderRatio })
	byKey := map[string]config.TechDef{}
	for _, tc := range techs {
		byKey[tc.Key] = tc
	}
	defs := config.BaseBuildings()
	rows := config.KnowledgeStoreRows(defs, config.Ages(), techs, config.KnowledgeBase)
	held := config.KnowledgeStoreHeld(defs, config.Ages())
	ages := config.Ages()
	for i, l := range ladders {
		if _, stopped := KnowledgeLadderHeld[l.Age]; stopped {
			// The ladder at 1.3 puts the dearest tech over the store, and no
			// ratio fits: the dearest of n techs priced to a total is never
			// under total/n.
			dearest := l.Total / float64(len(l.Techs))
			if dearest*config.KnowledgeDearestMargin <= held[l.Age] {
				t.Errorf("%s is held out of the ladder and its flattest ladder (%g a tech) fits the store %g: price it", l.Age, dearest, held[l.Age])
			}
			continue
		}
		for j, k := range l.Techs {
			if got := byKey[k].Cost; got != l.Prices[j] {
				t.Errorf("%s: %s costs %g, the ladder gives %g", l.Age, k, got, l.Prices[j])
			}
		}
		if n := len(l.Prices); n > 0 {
			if l.Prices[n-1]*config.KnowledgeDearestMargin > held[l.Age] {
				t.Errorf("%s: the dearest tech, %g, and a quarter is over the knowledge store %g", l.Age, l.Prices[n-1], held[l.Age])
			}
			arrive := config.KnowledgeBase
			if i > 0 {
				arrive = held[ages[i-1].Key]
			}
			if l.Prices[0] > arrive {
				t.Errorf("%s: the cheapest tech, %g, is over the store a town arrives with, %g", l.Age, l.Prices[0], arrive)
			}
			if sum := sumOf(l.Prices); math.Abs(sum/l.Total-1) > 0.02 {
				t.Errorf("%s: the techs total %g, the ladder asks %g", l.Age, sum, l.Total)
			}
		}
		if rows[i].Held != held[l.Age] {
			t.Errorf("%s: store %g, rule %g", l.Age, held[l.Age], rows[i].Held)
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

// TestKnowledgePriceFitsTheKnowledgeStore: R5. Every non-wonder price part
// in knowledge fits the knowledge store of its age: the first copy inside the
// store a town arrives with, the last copy the reference town holds inside
// the store it leaves the age with.
func TestKnowledgePriceFitsTheKnowledgeStore(t *testing.T) {
	ages := config.Ages()
	defs := config.BaseBuildings()
	held := config.KnowledgeStoreHeld(defs, ages)
	rt := newRefTables()
	for _, d := range defs {
		part := d.BaseCost["knowledge"]
		if part <= 0 || d.Category == "wonder" || d.Category == "monument" {
			continue
		}
		i := rt.idx[d.RequiredAge]
		arrive := config.KnowledgeBase
		if i > 0 {
			arrive = held[ages[i-1].Key]
		}
		if part > arrive {
			t.Errorf("%s: its first copy costs %g knowledge, over the %g a town arrives in the %s with", d.Key, part, arrive, d.RequiredAge)
		}
		last := copyPrice(d, rt.refCount(d)-1)["knowledge"]
		if last > held[d.RequiredAge] {
			t.Errorf("%s: its last required copy costs %g knowledge, over the %g the %s holds", d.Key, last, held[d.RequiredAge], d.RequiredAge)
		}
	}
}
