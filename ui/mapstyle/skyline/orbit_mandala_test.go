package skyline

import (
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/mapmodel/fixture"
	"github.com/espresso20/ageforge/ui/mapstyle"
)

// mandalaModel is the Transcendent Age reached through the first eras
// eras only (the run's record of ages cut down to theirs), or all seven.
func mandalaModel(eras int) *mapmodel.Model {
	st := fixture.State(fixture.Options{Age: "transcendent_age", Seed: 4, Routes: 3, Harbinger: true, Tick: tickAt(0.6)})
	if eras < 7 {
		st.Stats.AgesReached = nil
		for _, e := range config.Epochs()[:eras] {
			st.Stats.AgesReached = append(st.Stats.AgesReached, e.Ages...)
		}
		st.Stats.AgesReached = append(st.Stats.AgesReached, st.Age)
	}
	return mapmodel.NewBuilder(nil).Build(&st, nil)
}

// TestSkyMandalaSymmetric: the skyline's mandala is its own mirror image
// across both axes through the core, glyph for glyph and colour for colour,
// wherever nothing nearer (a lot, a name, a mote) covers one side; the
// stars and the civs keep off it; and it has one ring for each era
// reached, each in its era's glyph.
func TestSkyMandalaSymmetric(t *testing.T) {
	for _, sz := range [][2]int{{160, 48}, {200, 60}, {80, 24}} {
		for _, eras := range []int{7, 4, 1} {
			m := mandalaModel(eras)
			W, H := sz[0], sz[1]
			for _, anim := range []int{0, 37, 300} {
				v := newView()
				v.noTraffic = true
				s := v.compose(mapstyle.Frame{Model: m, Anim: anim, Tier: mapmodel.TierUnicode}, W, H)
				o := &orb{scene: s}
				cx, cy, ry, rx := o.mandalaGeom()
				g := mapstyle.LayMandala(eras, ry, rx)
				ext := g.Extent + 1.5
				glyphs := map[rune]bool{}
				for y := 0; y < s.groundY; y++ {
					for x := 0; x < W; x++ {
						dx, dy := float64(x-cx)/mapstyle.MandalaAspect, float64(y-cy)
						if dx*dx+dy*dy >= ext*ext {
							continue
						}
						c := v.fb.at(x, s.Y(y))
						switch {
						case c.d == dStar:
							t.Errorf("%dx%d %d eras: a star inside the mandala at (%d,%d)", W, H, eras, x, y)
						case c.d == dRidgeTown || c.d == dHarbinger:
							t.Errorf("%dx%d %d eras: a far object inside the mandala at (%d,%d)", W, H, eras, x, y)
						}
						if c.d != dMandala || c.ch == ' ' {
							continue
						}
						glyphs[c.ch] = true
						for i, q := range [2][2]int{{2*cx - x, y}, {x, 2*cy - y}} {
							mc := v.fb.at(q[0], s.Y(q[1]))
							if mc == nil || mc.d < dMandala || q[1] >= s.groundY {
								continue // covered by something nearer, or under the mirror line
							}
							if mc.d != dMandala || mdMirror(mc.ch, i == 0) != c.ch || mc.fg != c.fg || mc.bg != c.bg {
								t.Errorf("%dx%d %d eras frame %d: %q at (%d,%d), its mirror (%d,%d) %q", W, H, eras, anim, c.ch, x, y, q[0], q[1], mc.ch)
							}
						}
					}
				}
				rings := 0
				for e := 0; e < 7; e++ {
					if glyphs[mapmodel.R(mapmodel.EraSym(e), mapmodel.TierUnicode)] {
						rings++
						if e >= eras {
							t.Errorf("%dx%d: era %d's ring drawn, but only %d eras reached", W, H, e, eras)
						}
					}
				}
				if rings != g.Rings || g.Rings > eras {
					t.Errorf("%dx%d %d eras: %d rings drawn, %d laid out", W, H, eras, rings, g.Rings)
				}
				if W >= 160 && g.Rings != eras {
					t.Errorf("%dx%d: %d rings of %d eras", W, H, g.Rings, eras)
				}
			}
		}
	}
}

// TestSkyMandalaPixelBands: where the rings stand too close for bands of
// glyphs they are drawn in half-block pixels, and no cell ever holds two
// rings, nor a ring and a cell of the layout's own.
func TestSkyMandalaPixelBands(t *testing.T) {
	for _, sz := range [][2]float64{{12, 24}, {16, 31}, {5, 10}, {4, 19}} {
		for n := 1; n <= 7; n++ {
			g := mapstyle.LayMandala(n, sz[0], sz[1])
			if g.Lines {
				continue
			}
			lit := 0
			for dy := -int(sz[0]) - 1; dy <= int(sz[0])+1; dy++ {
				for dx := -int(sz[1]) - 1; dx <= int(sz[1])+1; dx++ {
					up, lo := pixelRing(g, dx, dy, 0), pixelRing(g, dx, dy, 1)
					if up >= 0 && lo >= 0 && up != lo {
						t.Errorf("%v n=%d: (%d,%d) holds rings %d and %d", sz, n, dx, dy, up, lo)
					}
					if up >= 0 || lo >= 0 {
						lit++
					}
				}
			}
			if lit == 0 {
				t.Errorf("%v n=%d: no pixel bands", sz, n)
			}
		}
	}
}

// mdMirror is r seen in a mirror: across the vertical axis (lr) or the
// horizontal one. The petal points, the core's rays and the half-block
// pixels turn with the mirror; every other glyph is its own mirror image.
func mdMirror(r rune, lr bool) rune {
	pairs := map[rune]rune{'╱': '╲', '╲': '╱'}
	if lr {
		for a, b := range map[rune]rune{'◄': '►', '◤': '◥', '◣': '◢'} {
			pairs[a], pairs[b] = b, a
		}
	} else {
		for a, b := range map[rune]rune{'▲': '▼', '◤': '◣', '◥': '◢', '▀': '▄'} {
			pairs[a], pairs[b] = b, a
		}
	}
	if m, ok := pairs[r]; ok {
		return m
	}
	return r
}
