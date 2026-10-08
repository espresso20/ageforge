package mapstyle

import (
	"testing"

	"github.com/espresso20/ageforge/mapmodel"
)

// TestVisitFrame: the rare visitor comes and goes on the world's clock.
// While the map moves that is the animation frame. With motion off the
// animation frame holds at 0, the clock runs on, and a visitor that is
// here stands at the middle of its visit; with none here there is no
// frame to draw one at.
func TestVisitFrame(t *testing.T) {
	m := &mapmodel.Model{Seed: 11}
	sg := mapmodel.NextSighting(m.Seed, false, 0)
	mid := sg.Start + sg.Frames/2

	// A frame built without a clock (a capture, a test) is as it always was.
	if got := (Frame{Model: m, Anim: sg.Start + 3}).VisitFrame(); got != sg.Start+3 {
		t.Errorf("no clock: the visit frame is %d, want the animation frame %d", got, sg.Start+3)
	}
	// Motion on: the clock and the animation are one.
	if got := (Frame{Model: m, Anim: sg.Start + 3, Clock: sg.Start + 3}).VisitFrame(); got != sg.Start+3 {
		t.Errorf("motion on: the visit frame is %d, want %d", got, sg.Start+3)
	}
	// Motion off, all through the visit: the visitor is here, and holds.
	for clock := sg.Start; clock < sg.Start+sg.Frames; clock++ {
		if got := (Frame{Model: m, Clock: clock}).VisitFrame(); got != mid {
			t.Fatalf("motion off at clock %d: the visit frame is %d, want the middle of the visit, %d", clock, got, mid)
		}
		if _, ok := m.SightingAt(mid); !ok {
			t.Fatal("the middle of a visit is not in it")
		}
	}
	// Motion off, before and after: nobody is here.
	for _, clock := range []int{sg.Start - 1, sg.Start + sg.Frames} {
		if got := (Frame{Model: m, Clock: clock}).VisitFrame(); got != -1 {
			t.Errorf("motion off at clock %d, outside the visit: the visit frame is %d, want -1", clock, got)
		}
	}
	if got := (Frame{Clock: 5}).VisitFrame(); got != -1 {
		t.Errorf("no model: the visit frame is %d, want -1", got)
	}
}
