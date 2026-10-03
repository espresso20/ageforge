package roguelike

import (
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/mapmodel/fixture"
	"github.com/espresso20/ageforge/ui/mapstyle"
)

// mandalaModel is the Transcendent Age reached through the first eras
// eras only (the run's record of ages cut down to theirs), or through all
// seven.
func mandalaModel(t *testing.T, eras int) *mapmodel.Model {
	t.Helper()
	st := fixture.State(fixture.Options{Age: "transcendent_age", Seed: 7, Scale: 1.5})
	if eras < 7 {
		st.Stats.AgesReached = nil
		for _, e := range config.Epochs()[:eras] {
			st.Stats.AgesReached = append(st.Stats.AgesReached, e.Ages...)
		}
		st.Stats.AgesReached = append(st.Stats.AgesReached, st.Age)
	}
	return testBuilder.Build(&st, nil)
}

// TestSkyMandalaSymmetric: on screen the mandala is its own mirror image
// across both axes through the core, glyph for glyph and colour for colour,
// at every frame of the breath; and it has one ring for each era reached,
// innermost first, each in that era's glyph.
func TestSkyMandalaSymmetric(t *testing.T) {
	for _, eras := range []int{7, 4, 1} {
		m := mandalaModel(t, eras)
		st := skyStyleFor()
		st.SetOption(mapstyle.OptInspect, false)
		for _, anim := range []int{0, 37, 300, 1001} {
			scr := drawStyle(st, m, 160, 48, anim, mapmodel.TierUnicode, false)
			s := st.sv.sceneFor(m)
			hx, hy, ok := st.sv.sg.cellOf(s.hub)
			if !ok {
				t.Fatalf("%d eras: the core is off the screen", eras)
			}
			hx, hy = hx+2, hy+1 // drawStyle's rect sits at (2, 1)
			ext := s.md.Extent + 1.5
			n := 0
			for dy := -int(ext); dy <= int(ext); dy++ {
				for dx := -int(2 * ext); dx <= int(2*ext); dx++ {
					if !inEllipse(dx, dy, 0, 0, mapstyle.MandalaAspect*ext, ext) {
						continue
					}
					r, _, sty, _ := scr.GetContent(hx+dx, hy+dy)
					for _, o := range [2][2]int{{-dx, dy}, {dx, -dy}} {
						or, _, osty, _ := scr.GetContent(hx+o[0], hy+o[1])
						if or != r || osty != sty {
							t.Errorf("%d eras, frame %d: %q at (%d,%d) from the core, %q at (%d,%d)", eras, anim, r, dx, dy, or, o[0], o[1])
						}
					}
					if r != ' ' {
						n++
					}
				}
			}
			if n < 40 {
				t.Errorf("%d eras, frame %d: only %d cells drawn in the mandala", eras, anim, n)
			}
			scr.Fini()
		}
		s := st.sv.sceneFor(m)
		if len(s.rings) != eras || s.md.Rings != eras {
			t.Errorf("%d eras reached: %d rings in the model, %d drawn", eras, len(s.rings), s.md.Rings)
		}
		for k, r := range s.rings {
			if r.Epoch != k {
				t.Errorf("%d eras: ring %d is era %d", eras, k, r.Epoch)
			}
		}
		// each ring in its own era's glyph, nothing else on it, and no ring
		// drawn over another
		want := map[int]int{}
		for _, mc := range s.md.Cells {
			if mc.Part == mapstyle.MdBead || mc.Part == mapstyle.MdLine {
				want[mc.Ring]++
			}
		}
		got := map[int]int{}
		for _, c := range s.cells {
			switch {
			case c.k == skRing:
				got[int(c.ref)]++
			case c.k == skMark && c.ref>>10 != mdPetal:
				k := int(c.ref >> 10)
				got[k]++
				if c.sym != mapmodel.EraSym(s.rings[k].Epoch) {
					t.Errorf("%d eras: ring %d carries %v, not its era's glyph", eras, k, c.sym)
				}
			}
		}
		for k := 0; k < eras; k++ {
			if got[k] != want[k] || got[k] == 0 {
				t.Errorf("%d eras: ring %d has %d cells on the plane, the layout %d", eras, k, got[k], want[k])
			}
		}
	}
}

// TestSkyMandalaStill: the breath changes brightness only: every frame
// draws the same glyph in every cell of the mandala.
func TestSkyMandalaStill(t *testing.T) {
	m := mandalaModel(t, 7)
	st := skyStyleFor()
	st.SetOption(mapstyle.OptInspect, false)
	first := map[[2]int]rune{}
	for anim := 0; anim < 200; anim += 7 {
		scr := drawStyle(st, m, 160, 48, anim, mapmodel.TierUnicode, false)
		s := st.sv.sceneFor(m)
		hx, hy, _ := st.sv.sg.cellOf(s.hub)
		hx, hy = hx+2, hy+1
		for _, mc := range s.md.Cells {
			r, _, _, _ := scr.GetContent(hx+mc.DX, hy+mc.DY)
			k := [2]int{mc.DX, mc.DY}
			if f, ok := first[k]; !ok {
				first[k] = r
			} else if f != r {
				t.Fatalf("frame %d: (%d,%d) changed from %q to %q", anim, mc.DX, mc.DY, f, r)
			}
		}
		scr.Fini()
	}
}
