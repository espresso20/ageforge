package smoke

import (
	"math"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// knowledgeLadderHeld names an age the ladder cannot price, with the reason;
// its techs keep the prices they had before the knowledge store.
var knowledgeLadderHeld = map[string]string{
	"transcendent_age": "no knowledge building: the store stays at the Quantum Age's, and the age's four techs, priced to 0.8 of what the town makes, cost more than it holds at any ratio",
}

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
		if _, stopped := knowledgeLadderHeld[l.Age]; stopped {
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

// TestKnowledgeHeldAgePricesAreRecorded: the held age keeps its prices from
// before the knowledge store (their sum is recorded), so a later change shows.
func TestKnowledgeHeldAgePricesAreRecorded(t *testing.T) {
	sum, n := 0.0, 0
	for _, tc := range config.Technologies() {
		if tc.Age == "transcendent_age" {
			sum += tc.Cost
			n++
		}
	}
	if n != 4 || math.Abs(sum-21.54e9) > 1e6 {
		t.Errorf("the Transcendent Age's %d techs cost %g, held at the 21.54e9 they cost before the knowledge store", n, sum)
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
