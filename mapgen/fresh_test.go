package mapgen

import (
	"reflect"
	"testing"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/mapmodel/fixture"
)

// grownModel is a town a step on from the one the player last saw, so some
// building types are new since the last visit.
func grownModel(age string) *mapmodel.Model {
	st := fixture.State(fixture.Options{Age: age, Seed: 7, Tick: noon})
	b := mapmodel.NewBuilder(testCatalog)
	visit := mapmodel.VisitOf(b.Build(&st, nil))
	next := fixture.Grow(st, 1)
	return b.Build(&next, visit)
}

// TestFreshLightsOnlyTheGroundOfWhatIsNew: with the news lit, the parcels of
// the types that grew wear the fresh tint behind their art, and nothing else
// of the picture changes; unlit, nothing wears it.
func TestFreshLightsOnlyTheGroundOfWhatIsNew(t *testing.T) {
	for _, age := range append(append([]string(nil), testAges...), "interstellar_age", "galactic_age") {
		m := grownModel(age)
		var gen Generator
		dark, lit := gen.DrawAt(m, Moment{Visit: -1}).Grid, gen.DrawAt(m, Moment{Visit: -1, Fresh: true}).Grid
		tinted := 0
		for i := range dark.Cells {
			a, b := dark.Cells[i], lit.Cells[i]
			if b.Bg == freshTint && a.Bg != freshTint {
				tinted++
			}
			a.Bg, b.Bg = Paint{}, Paint{}
			if a != b || dark.Hit[i] != lit.Hit[i] {
				t.Fatalf("%s: lighting the news changed more than the ground at (%d,%d)", age, i%dark.W, i/dark.W)
			}
			if dark.Cells[i].Bg == freshTint {
				t.Fatalf("%s: an unlit picture wears the tint at (%d,%d)", age, i%dark.W, i/dark.W)
			}
		}
		fresh := 0
		for _, b := range m.Buildings {
			if b.Delta > 0 {
				fresh++
			}
		}
		if fresh == 0 || tinted == 0 {
			t.Errorf("%s: %d types are new and %d cells are lit", age, fresh, tinted)
		}
	}
}

// TestFreshLightsAFinishedWonder: a wonder finished since the last visit has
// its plot lit.
func TestFreshLightsAFinishedWonder(t *testing.T) {
	m := grownModel("modern_age")
	m.Wonders = append([]mapmodel.Wonder(nil), m.Wonders...)
	for i := range m.Wonders {
		m.Wonders[i].Delta = m.Wonders[i].Built
	}
	var gen Generator
	f := gen.frame(m, 0)
	f.fresh = true
	f.draw()
	for _, ws := range f.wonders() {
		if ws.w.Delta && f.g.At(ws.plot.Rect.X, ws.plot.Rect.Y).Bg != freshTint {
			t.Errorf("the plot of %s is not lit", ws.w.Key)
		}
	}
}

// TestDistrictsAreOneLineEach: every district is a building line that has
// something standing, in the order of the lines, with its heart on a building
// of the line inside its box.
func TestDistrictsAreOneLineEach(t *testing.T) {
	for _, age := range append(append([]string(nil), testAges...), "interstellar_age", "galactic_age") {
		sc := (&Generator{}).Draw(model(state(age, 7, noon)), 0)
		if len(sc.Districts) < 6 {
			t.Errorf("%s: only %d districts", age, len(sc.Districts))
		}
		last := -1
		for _, d := range sc.Districts {
			if idx := lineOf[d.Line]; idx <= last {
				t.Errorf("%s: district %s is out of order", age, d.Line)
			} else {
				last = idx
			}
			th := sc.Grid.ThingAt(d.Heart.X, d.Heart.Y)
			if th.Kind != KindBuilding || th.Lineage != d.Line {
				t.Errorf("%s: the heart of %s is %v", age, d.Line, th)
			}
			if d.Heart.X < d.Box.X || d.Heart.X >= d.Box.X+d.Box.W || d.Heart.Y < d.Box.Y || d.Heart.Y >= d.Box.Y+d.Box.H {
				t.Errorf("%s: the heart of %s is outside its box", age, d.Line)
			}
		}
	}
}

// TestFocusDimsTheOtherLines: with a district in focus, its buildings are as
// they were and every other line's, and the ordinary houses, are dimmer; the
// glyphs do not change, and it is deterministic.
func TestFocusDimsTheOtherLines(t *testing.T) {
	for _, age := range append(append([]string(nil), testAges...), "galactic_age") {
		m := model(state(age, 7, noon))
		var gen Generator
		whole := gen.DrawAt(m, Moment{Visit: -1})
		line := whole.Districts[len(whole.Districts)/2].Line
		focus := gen.DrawAt(m, Moment{Visit: -1, District: line})
		again := (&Generator{}).DrawAt(m, Moment{Visit: -1, District: line})
		if !reflect.DeepEqual(focus.Grid.Cells, again.Grid.Cells) {
			t.Fatalf("%s: a district in focus is not deterministic", age)
		}
		dimmed, kept := 0, 0
		for i, h := range whole.Grid.Hit {
			th := whole.Grid.Things[h]
			a, b := whole.Grid.Cells[i], focus.Grid.Cells[i]
			if a.R != b.R || a.Fg != b.Fg {
				t.Fatalf("%s: focus changed a glyph at (%d,%d)", age, i%whole.Grid.W, i/whole.Grid.W)
			}
			switch {
			case th.Kind == KindBuilding && th.Lineage == line:
				kept++
				if a != b {
					t.Fatalf("%s: focus changed the district itself at (%d,%d)", age, i%whole.Grid.W, i/whole.Grid.W)
				}
			case th.Kind == KindBuilding || th.Kind == KindAmbient:
				if b.Dim <= a.Dim {
					t.Fatalf("%s: %s at (%d,%d) is not dimmed", age, th.Lineage, i%whole.Grid.W, i/whole.Grid.W)
				}
				dimmed++
			}
		}
		if dimmed == 0 || kept == 0 {
			t.Errorf("%s: %d cells dimmed, %d kept", age, dimmed, kept)
		}
	}
}
