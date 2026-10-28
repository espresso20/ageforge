package mapstyle

import (
	"slices"
	"testing"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapgen"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/mapmodel/fixture"
	"github.com/espresso20/ageforge/ui/mapstyle/capture"
)

// visitAges are one age of each scale a visit comes to: the land, the city,
// orbit and the solar system. The star map's ages schedule none.
var visitAges = []string{"bronze_age", "modern_age", "space_age", "interstellar_age"}

// visitorAt finds a visit of the model whose visitor is of the given form
// ("visitor" or "alien") and returns the frame it is at its middle, and the
// cells it is drawn in.
func visitorAt(t *testing.T, m *mapmodel.Model, form string) (frame int, cells []mapgen.Pt) {
	t.Helper()
	late := m.AgeIdx >= m.Catalog.SpaceAge()
	var gen mapgen.Generator
	for from, tries := 0, 0; tries < 60; tries++ {
		sg := mapmodel.NextSighting(m.Seed, late, from)
		from = sg.Start + sg.Frames
		frame = sg.Start + sg.Frames/2
		g := gen.DrawAt(m, 0, frame).Grid
		for i := range g.Cells {
			if th := g.Things[g.Hit[i]]; th.Kind == mapgen.KindVisitor && th.Key == form {
				cells = append(cells, mapgen.Pt{X: i % g.W, Y: i / g.W})
			}
		}
		if len(cells) > 0 {
			return frame, cells
		}
	}
	t.Fatalf("%s: no %s among the first visits", m.Age, form)
	return 0, nil
}

// steer presses the arrows until the view's cursor is on p, p being whole
// steps of unit cells from where it is, Shift for the long way. It reports
// whether it got there.
func steer(v Style, f Frame, p mapgen.Pt, unit int) bool {
	for i := 0; i < 400; i++ {
		at, _ := v.Pointer()
		dx, dy := p.X-at.X, p.Y-at.Y
		if dx == 0 && dy == 0 {
			return true
		}
		key, along := tcell.KeyRight, dx
		switch {
		case dx < 0:
			key = tcell.KeyLeft
		case dx == 0 && dy < 0:
			key, along = tcell.KeyUp, dy
		case dx == 0:
			key, along = tcell.KeyDown, dy
		}
		mod := tcell.ModNone
		if abs(along) >= 8*unit {
			mod = tcell.ModShift
		}
		v.HandleKey(tcell.NewEventKey(key, 0, mod), f)
	}
	return false
}

func abs(v int) int { return max(v, -v) }

// TestVisitorInspectsAsTheVisitor: on every scale the cursor on the visitor
// names it with its own words and reports KindAlien, the visitor and the
// alien's craft each with a title and a line of their own.
func TestVisitorInspectsAsTheVisitor(t *testing.T) {
	fixture.WithTheme(t, "forge")
	titles := map[string]string{"visitor": "Visitor", "alien": "Unknown craft"}
	for _, age := range visitAges {
		m := modelFor(t, fixture.Options{Age: age, Seed: 3})
		for form, title := range titles {
			frame, cells := visitorAt(t, m, form)
			v := New().(*view)
			f := Frame{Model: m, Anim: frame, Clock: frame}
			v.scene(f)
			for _, p := range cells {
				v.cur = p
				in, ok := v.Inspect(f)
				if !ok || in.Kind != KindAlien || in.Title != title || len(in.Lines) != 1 || in.Lines[0] == "" {
					t.Fatalf("%s: the %s at %v inspects as %+v", age, form, p, in)
				}
			}
		}
	}
}

// TestVisitorIsReachableByTheKeys: a player with nothing but the arrow keys
// can put the cursor on the visitor, in the close-up and in the whole-town
// view, and what it reports is KindAlien, which is what earns the badge.
func TestVisitorIsReachableByTheKeys(t *testing.T) {
	fixture.WithTheme(t, "forge")
	for _, age := range visitAges {
		m := modelFor(t, fixture.Options{Age: age, Seed: 3})
		for _, form := range []string{"visitor", "alien"} {
			frame, cells := visitorAt(t, m, form)
			f := Frame{Model: m, Anim: frame, Clock: frame}
			for _, size := range []struct{ w, h int }{{160, 48}, {80, 24}} {
				v := New().(*view)
				scr := capture.NewScreen(size.w, size.h)
				v.Draw(scr, Rect{W: size.w, H: size.h}, f)
				scr.Fini()
				target, unit := cells[len(cells)/2], 1
				if v.overview {
					// every step is a block of the world: stop in the block the visitor is in
					unit = v.ov.Scale
					x, y, _ := v.ov.CellOf(v.cur)
					tx, ty, _ := v.ov.CellOf(target)
					target = mapgen.Pt{X: v.cur.X + (tx-x)*unit, Y: v.cur.Y + (ty-y)*unit}
				}
				if !steer(v, f, target, unit) {
					t.Fatalf("%s %dx%d: the arrows never reach the %s at %v", age, size.w, size.h, form, target)
				}
				if in, ok := v.Inspect(f); !ok || in.Kind != KindAlien {
					t.Errorf("%s %dx%d: the cursor on the %s inspects as %+v", age, size.w, size.h, form, in)
				}
			}
		}
	}
}

// TestVisitorChangesThePictureAtEverySize: a visit is on the screen in a big
// panel, a small one and the mini map, in the theme's colours, and with none
// out the picture is the one without it. In plain ASCII it stays clean.
func TestVisitorChangesThePictureAtEverySize(t *testing.T) {
	fixture.WithTheme(t, "forge")
	for _, age := range visitAges {
		m := modelFor(t, fixture.Options{Age: age, Seed: 3})
		for _, form := range []string{"visitor", "alien"} {
			frame, spots := visitorAt(t, m, form)
			for _, sz := range sizes {
				// the cursor stands on the visitor, so the close-up has it in view
				look := func(clock int) tcell.SimulationScreen {
					v := New().(*view)
					f := Frame{Model: m, Anim: 5, Clock: clock}
					v.scene(f)
					v.cur, v.recentre = spots[0], true
					return drawFrame(v, f, sz.w, sz.h, sz.compact)
				}
				with, without := look(frame), look(0)
				a, _, _ := with.GetContents()
				b, _, _ := without.GetContents()
				if slices.EqualFunc(a, b, func(x, y tcell.SimCell) bool { return x.Style == y.Style && slices.Equal(x.Runes, y.Runes) }) {
					t.Errorf("%s %s %s: a visit does not change the picture", age, form, sz.name)
				}
				with.Fini()
				without.Fini()
			}
			scr := drawFrame(New(), Frame{Model: m, Anim: 5, Clock: frame, Tier: mapmodel.TierASCII}, 120, 36, false)
			cells, w, h := scr.GetContents()
			for i, c := range cells[:w*h] {
				if len(c.Runes) > 0 && (c.Runes[0] > 0x7e || c.Runes[0] == '?') {
					t.Fatalf("%s %s: %q at (%d,%d) is not clean ASCII", age, form, c.Runes[0], i%w, i/w)
				}
			}
			scr.Fini()
		}
	}
}

// drawFrame draws a view of a frame into a w by h screen.
func drawFrame(v Style, f Frame, w, h int, compact bool) tcell.SimulationScreen {
	scr := capture.NewScreen(w, h)
	if compact {
		v.DrawCompact(scr, Rect{W: w, H: h}, f)
	} else {
		v.Draw(scr, Rect{W: w, H: h}, f)
	}
	scr.Show()
	return scr
}

// TestVisitorNeverCoversTheCursor: with the cursor on the visitor the cell
// keeps the visitor's glyph and wears the cursor's highlight over it.
func TestVisitorNeverCoversTheCursor(t *testing.T) {
	fixture.WithTheme(t, "forge")
	m := modelFor(t, fixture.Options{Age: "modern_age", Seed: 3})
	frame, cells := visitorAt(t, m, "alien")
	f := Frame{Model: m, Anim: frame, Clock: frame}
	target := cells[len(cells)/2]
	draw := func(cursor bool) (rune, tcell.Style) {
		v := New().(*view)
		v.scene(f)
		v.cur, v.recentre, v.inspect = target, true, cursor
		scr := drawFrame(v, f, 120, 36, false)
		defer scr.Fini()
		x, y := target.X-v.cam.X, 1+target.Y-v.cam.Y
		cell, _, _ := scr.GetContents()
		c := cell[y*120+x]
		return c.Runes[0], c.Style
	}
	bare, bareStyle := draw(false)
	marked, markedStyle := draw(true)
	if bare != marked {
		t.Errorf("the cursor changed the visitor's glyph from %q to %q", bare, marked)
	}
	if bareStyle == markedStyle {
		t.Error("the cursor is not shown on the visitor")
	}
}

// TestVisitorHoldsStillWithMotionOff: with the animation held at frame 0 and
// the clock running, the visitor is on the picture where the middle of its
// visit puts it, for the whole visit.
func TestVisitorHoldsStillWithMotionOff(t *testing.T) {
	fixture.WithTheme(t, "forge")
	m := modelFor(t, fixture.Options{Age: "space_age", Seed: 3})
	frame, _ := visitorAt(t, m, "visitor")
	sg, _ := m.SightingAt(frame)
	var first []tcell.SimCell
	for clock := sg.Start; clock < sg.Start+sg.Frames; clock += 9 {
		scr := drawFrame(New(), Frame{Model: m, Anim: 0, Clock: clock}, 120, 36, false)
		cells, _, _ := scr.GetContents()
		if first == nil {
			first = slices.Clone(cells)
		}
		for i := range cells {
			if cells[i].Style != first[i].Style || !slices.Equal(cells[i].Runes, first[i].Runes) {
				t.Fatalf("clock %d: the picture moved with motion off", clock)
			}
		}
		scr.Fini()
	}
}
