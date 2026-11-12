package world

import (
	"reflect"
	"testing"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/mapmodel/fixture"
)

func web(t *testing.T, age string, seed int64) *Frame {
	t.Helper()
	w := Build(model(age, seed, fixture.Options{}), seed)
	if w.Root == nil || w.Root.Kind != FrameWeb {
		t.Fatalf("%s: root %v, want the web", age, w.Root)
	}
	return w.Root
}

func TestWebIsTheSameForTheSameSeed(t *testing.T) {
	a, b, c := web(t, AgeTranscendent, 7), web(t, AgeTranscendent, 7), web(t, AgeTranscendent, 8)
	if !reflect.DeepEqual(a.Objects, b.Objects) {
		t.Error("two builds of one seed differ")
	}
	if reflect.DeepEqual(a.Objects, c.Objects) {
		t.Error("two seeds wove the same web")
	}
}

func TestWebSize(t *testing.T) {
	f := web(t, AgeTranscendent, 7)
	nodes, strands := 0, 0
	for _, o := range f.Objects {
		switch o.Kind {
		case KindCluster:
			nodes++
		case KindStrand:
			strands++
		}
	}
	if nodes < 20 || nodes > 40 || strands < 40 || strands > 80 {
		t.Errorf("%d nodes, %d strands: want 20-40 and 40-80", nodes, strands)
	}
	if f.Find(FrameLocalGroup) == nil || f.Children[0].Anchor != "local_group" {
		t.Error("the Local Group is not hung on its node")
	}
	for i := range f.Objects {
		if o := &f.Objects[i]; o.Key == "local_group" && (o.Pos.X*o.Pos.X+o.Pos.Y*o.Pos.Y) > 30*30 {
			t.Errorf("the Local Group node is at %v, not near the centre", o.Pos)
		}
	}
}

func TestEveryLineHasAStrand(t *testing.T) {
	for _, seed := range []int64{1, 7, 99} {
		have := map[string]bool{}
		for _, o := range web(t, AgeTranscendent, seed).Objects {
			if o.Kind == KindStrand {
				have[o.Line] = true
			}
		}
		for _, l := range mapmodel.LineageOrder {
			if !have[l] {
				t.Errorf("seed %d: line %s has no strand", seed, l)
			}
		}
	}
}

// A line with more buildings, or a higher tier, never draws a dimmer strand.
func TestStrengthNeverFallsWithMore(t *testing.T) {
	prev := 0
	for tier := 0; tier <= 3; tier++ {
		prev = 0
		for n := 0; n <= 5000; n++ {
			s := Strength(n, tier)
			if s < prev {
				t.Fatalf("count %d tier %d: %d, below %d", n, tier, s, prev)
			}
			if tier > 0 && n > 0 && s < Strength(n, tier-1) {
				t.Fatalf("count %d: tier %d dimmer than tier %d", n, tier, tier-1)
			}
			prev = s
		}
	}
	if Strength(0, 3) != 0 || Strength(1, 0) < 70 || Strength(1<<30, 3) > 255 {
		t.Errorf("ends: %d %d %d", Strength(0, 3), Strength(1, 0), Strength(1<<30, 3))
	}
}

func TestStrandsGlowWithTheirLine(t *testing.T) {
	few := web(t, AgeTranscendent, 7)
	for _, o := range few.Objects {
		if o.Kind == KindStrand && (o.Count > 0) != (o.Weight > StrandUnbuilt) {
			t.Fatalf("strand %s: count %d weight %d", o.Key, o.Count, o.Weight)
		}
	}
}

func TestWebIsNotBelowQuantum(t *testing.T) {
	for _, age := range SpaceAges[:3] {
		if w := Build(model(age, 7, fixture.Options{}), 7); w.Root.Find(FrameWeb) != nil {
			t.Errorf("%s draws the web", age)
		}
	}
	if q := web(t, AgeQuantum, 7); !GateOf(AgeQuantum).Locked || GateOf(AgeQuantum).Open != FrameLocalGroup || q.Find(FrameLocalGroup) == nil {
		t.Error("at Quantum the Local Group should be open with the web locked above")
	}
}
