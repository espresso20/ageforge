package roguelike

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/mapmodel/fixture"
)

// legendLabels reads the legend's row labels (the text after each glyph)
// from a full view drawn by draw at w x h.
func legendLabels(s tcell.SimulationScreen, w, h int) []string {
	var out []string
	x0 := 2 + w - sideW + 2 // draw's offset, the sidebar, its border and margin
	for y := 2; y < h-4; y++ {
		var b strings.Builder
		for x := x0; x < 2+w; x++ {
			r, _, _, _ := s.GetContent(x, y)
			b.WriteRune(r)
		}
		row := strings.TrimSpace(b.String())
		if _, label, ok := strings.Cut(row, " "); ok {
			out = append(out, label)
		}
	}
	return out
}

// TestHarbingerSteady: the harbinger stands still on the map, and nothing
// that flashes (a war mark, chimney smoke) makes a legend row come and go.
// Adam found the harbinger row in the legend's "state" group blinking.
func TestHarbingerSteady(t *testing.T) {
	m := modelFor(t, fixture.Options{Age: "iron_age", Seed: 7, Harbinger: true, Wars: 1})
	v := newView()
	var first []string
	for anim := 0; anim < 56; anim++ {
		scr := draw(v, m, 160, 48, anim, mapmodel.TierUnicode, false)
		s := v.sceneFor(m)
		if !s.hasHb {
			t.Fatal("no harbinger placed")
		}
		x, y, ok := v.g.cellOf(s.harb)
		if !ok || s.seen(s.harb) < 2 {
			t.Fatal("the harbinger is out of sight at the default camera")
		}
		if r, _, _, _ := scr.GetContent(x+2, y+1); r != mapmodel.R(mapmodel.SymHarbinger, mapmodel.TierUnicode) {
			t.Fatalf("frame %d: the harbinger's cell holds %q", anim, r)
		}
		labels := legendLabels(scr, 160, 48)
		joined := strings.Join(labels, "|")
		for _, want := range []string{"harbinger", "war on this trail"} {
			if !strings.Contains(joined, want) {
				t.Fatalf("frame %d: legend has no %q row: %v", anim, want, labels)
			}
		}
		if anim == 0 {
			first = labels
		} else if joined != strings.Join(first, "|") {
			t.Fatalf("frame %d: the legend changed\nwas %v\nnow %v", anim, first, labels)
		}
	}
}
