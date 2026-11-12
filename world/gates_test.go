package world

import (
	"testing"

	"github.com/espresso20/ageforge/mapmodel/fixture"
)

func TestGateOfEachAge(t *testing.T) {
	for _, c := range []struct {
		age    string
		open   FrameKind
		locked bool
		top    FrameKind
	}{
		{"stone_age", FramePlanet, false, FramePlanet},
		{AgeSpace, FrameSystem, true, FrameNeighbourhood},
		{AgeInterstellar, FrameNeighbourhood, true, FrameRegion},
		{AgeGalactic, FrameGalaxy, true, FrameLocalGroup},
		{AgeQuantum, FrameLocalGroup, false, FrameLocalGroup},
		{AgeTranscendent, FrameLocalGroup, false, FrameLocalGroup},
	} {
		g := GateOf(c.age)
		if g.Open != c.open || g.Locked != c.locked || g.Top() != c.top {
			t.Errorf("%s: gate %+v top %v, want open %v locked %v top %v", c.age, g, g.Top(), c.open, c.locked, c.top)
		}
	}
}

// The world holds nothing outside the locked frame, and the lock names the
// age that opens it.
func TestBuildStopsAtTheLockedFrame(t *testing.T) {
	for _, age := range SpaceAges {
		w := Build(model(age, 7, fixture.Options{}), 7)
		if w.Root == nil || w.Root.Kind != GateOf(age).Top() {
			t.Errorf("%s: root %v, want the gate's top %v", age, w.Root, GateOf(age).Top())
		}
		if g := GateOf(age); g.Locked && OpensAt(g.Top()) == "" {
			t.Errorf("%s: the locked frame says no age opens it", age)
		}
	}
	if OpensAt(FrameNeighbourhood) != AgeInterstellar {
		t.Errorf("the neighbourhood opens in %q", OpensAt(FrameNeighbourhood))
	}
}
