package skyline

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/mapmodel/fixture"
	"github.com/espresso20/ageforge/ui/mapstyle"
)

// mandalaState is a busy Transcendent Age (routes, a war, the harbinger)
// reached through the first eras eras only (the run's record of ages cut
// down to theirs), or through all seven.
func mandalaState(eras int) game.GameState {
	st := fixture.State(fixture.Options{Age: "transcendent_age", Seed: 4, Routes: 3, Wars: 1, Harbinger: true, Scale: 1.5,
		Tick: tickAt(0.6)})
	if eras < 7 {
		st.Stats.AgesReached = nil
		for _, e := range config.Epochs()[:eras] {
			st.Stats.AgesReached = append(st.Stats.AgesReached, e.Ages...)
		}
		st.Stats.AgesReached = append(st.Stats.AgesReached, st.Age)
	}
	return st
}

func mandalaModel(eras int) *mapmodel.Model {
	st := mandalaState(eras)
	return mapmodel.NewBuilder(catalog()).Build(&st, nil)
}

// TestSkyMandalaAlone: the Transcendent Age leaves the corporeal world.
// Its frame holds the void, a few stars and the mandala, and nothing else:
// no lots, no ground or mirror, no far civs or harbinger, no names, no
// traffic, no minimap strip of a panorama, and no layout of lots is ever
// built. The cursor has no building, civ, harbinger or tether to stand on.
func TestSkyMandalaAlone(t *testing.T) {
	st := mandalaState(7)
	for i := 1; i <= 3; i++ {
		st = fixture.Grow(st, i) // a build queue and new buildings: none of it shows
	}
	m := mapmodel.NewBuilder(catalog()).Build(&st, nil)
	if len(m.Skyline.Lots) == 0 || m.Harbinger == nil {
		t.Fatal("the fixture has no lots or no harbinger to leave out")
	}
	for _, sz := range [][2]int{{160, 48}, {200, 60}, {80, 24}} {
		W, H := sz[0], sz[1]
		for _, anim := range []int{0, 37, 211} {
			v := newView()
			s := v.compose(mapstyle.Frame{Model: m, Anim: anim, Tier: mapmodel.TierUnicode}, W, H)
			if v.slay != nil {
				t.Fatalf("%dx%d: a layout of lots was built", W, H)
			}
			if s.S != H-2 || s.groundY != s.S {
				t.Errorf("%dx%d: %d scene rows and a baseline at %d; want every row between the header and the status line", W, H, s.S, s.groundY)
			}
			mandala := 0
			for y := 0; y < s.S; y++ {
				for x := 0; x < W; x++ {
					switch c := v.fb.at(x, s.Y(y)); c.d {
					case dMandala:
						mandala++
					case dSky, dStar:
					default:
						t.Fatalf("%dx%d frame %d: %q at (%d,%d) is neither the void, a star nor the mandala (depth %d)", W, H, anim, c.ch, x, y, c.d)
					}
				}
			}
			if mandala < 100 {
				t.Errorf("%dx%d: only %d cells of mandala", W, H, mandala)
			}
			o := newOrb(s, mapmodel.SkyMandala)
			if n := len(skyTrafficFor(o.geo())); n != 0 {
				t.Errorf("%dx%d: %d vehicles in the Transcendent Age", W, H, n)
			}
		}
	}
	for _, tg := range targets(m, 160, 0) {
		switch tg.kind {
		case tCore, tMark, tRing:
		default:
			t.Errorf("a target of kind %d in the Transcendent Age", tg.kind)
		}
	}
}

// TestSkyMandalaSymmetric: the skyline's mandala stands in the middle of
// the pane and fills it, and is its own mirror image across both axes
// through the core, glyph for glyph and colour for colour, at every frame
// of the wave; the stars keep off it; it has one ring for each era
// reached, each in its era's glyph; and at 160x48 and larger it is the
// full mandala: bands of glyphs, the crown's pointed petals, the star.
func TestSkyMandalaSymmetric(t *testing.T) {
	for _, sz := range [][2]int{{160, 48}, {200, 60}, {120, 36}, {80, 24}} {
		for _, eras := range []int{7, 4, 1} {
			m := mandalaModel(eras)
			W, H := sz[0], sz[1]
			for _, anim := range []int{0, 37, 300} {
				v := newView()
				s := v.compose(mapstyle.Frame{Model: m, Anim: anim, Tier: mapmodel.TierUnicode}, W, H)
				l := v.md
				if l == nil || l.cx != W/2 || l.cy != (s.S-1)/2 {
					t.Fatalf("%dx%d: the mandala is not in the middle of the pane: %+v", W, H, l)
				}
				g, cx, cy := l.g, l.cx, l.cy
				ext := g.Extent + 1.5
				glyphs := map[rune]bool{}
				for y := 0; y < s.S; y++ {
					for x := 0; x < W; x++ {
						dx, dy := float64(x-cx)/mapstyle.MandalaAspect, float64(y-cy)
						if dx*dx+dy*dy >= ext*ext {
							continue
						}
						c := v.fb.at(x, s.Y(y))
						if c.d != dMandala {
							t.Fatalf("%dx%d %d eras: %q inside the mandala at (%d,%d) is not the mandala's (depth %d)", W, H, eras, c.ch, x, y, c.d)
						}
						if c.ch != ' ' {
							glyphs[c.ch] = true
						}
						for i, q := range [2][2]int{{2*cx - x, y}, {x, 2*cy - y}} {
							if q[1] < 0 || q[1] >= s.S {
								continue
							}
							mc := v.fb.at(q[0], s.Y(q[1]))
							if mc == nil || mc.d != dMandala || mdMirror(mc.ch, i == 0) != c.ch || mc.bg != c.bg || c.ch != ' ' && mc.fg != c.fg {
								t.Errorf("%dx%d %d eras frame %d: %q at (%d,%d) is not mirrored at (%d,%d)", W, H, eras, anim, c.ch, x, y, q[0], q[1])
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
				if rings != eras || g.Rings != eras {
					t.Errorf("%dx%d %d eras: %d rings drawn, %d laid out", W, H, eras, rings, g.Rings)
				}
				if W >= 160 {
					if !g.Lines || g.Petals != 8 || !glyphs['▲'] || !glyphs['┃'] {
						t.Errorf("%dx%d %d eras: not the full mandala (bands %v, %d petals)", W, H, eras, g.Lines, g.Petals)
					}
					if eras == 7 && g.Extent < float64(s.S)/2-2 {
						t.Errorf("%dx%d: the mandala reaches %.1f rows from its core in a pane of %d", W, H, g.Extent, s.S)
					}
				}
			}
		}
	}
}

// TestSkyMandalaInspect: the cursor still reads the player's history. Tab
// starts on the core and steps through the crown and every ring outward,
// reaching every building type still standing; every target says what it
// is with a command the model knows, and is marked on the mandala; the
// status line shows it; the arrows step round a ring (and past its end to
// its start) and from ring to ring; a ring with nothing left standing is
// still there to read.
func TestSkyMandalaInspect(t *testing.T) {
	m := mandalaModel(7)
	words := commandWords(m)
	v := newView()
	f := mapstyle.Frame{Model: m}
	_ = draw(v, m, 160, 48, 1, mapmodel.TierUnicode, false)
	want := map[string]bool{}
	for _, mk := range m.Crown() {
		want[mk.Key] = true
	}
	rings := m.Mandala()
	for _, r := range rings {
		for _, mk := range r.Marks {
			want[mk.Key] = true
		}
	}
	if len(want) < 40 || len(m.Crown()) == 0 {
		t.Fatalf("the fixture has %d building types and %d in the crown", len(want), len(m.Crown()))
	}
	v.HandleKey(tabKey(), f)
	if in, ok := v.Inspect(f); !ok || v.cur.kind != tCore || in.Title != "The core" || in.Command != m.SquareCommand() {
		t.Fatalf("the first Tab did not put the cursor on the core: %+v (%+v)", v.cur, in)
	}
	seen := map[string]bool{}
	for k := 0; k < len(want)+4; k++ {
		in, ok := v.Inspect(f)
		if !ok || in.Title == "" || len(in.Lines) == 0 {
			t.Fatalf("step %d: nothing under the cursor", k)
		}
		if in.Command == "" || !words[strings.Fields(in.Command)[0]] {
			t.Fatalf("%q: command %q is not a model command", in.Title, in.Command)
		}
		for _, l := range append([]string{in.Title}, in.Lines...) {
			if strings.ContainsRune(l, '—') {
				t.Errorf("%q: an em dash in %q", in.Title, l)
			}
		}
		if v.cur.kind == tMark {
			seen[v.cur.key] = true
		}
		// the cursor marks one cell, on the mandala
		scr := draw(v, m, 160, 48, 2+k, mapmodel.TierUnicode, false)
		marks := 0
		for y := 1; y < 47; y++ {
			for x := 0; x < 160; x++ {
				if c := v.fb.at(x, y); c.d == dTop {
					marks++
				}
			}
		}
		if marks != 1 {
			t.Fatalf("%q: %d cells marked by the cursor, want 1", in.Title, marks)
		}
		if st := rowText(scr, 47); !strings.Contains(st, in.Title) || !strings.Contains(st, "type: "+in.Command) {
			t.Fatalf("the status line for %q does not show it with its command: %q", in.Title, st)
		}
		v.HandleKey(tabKey(), f)
	}
	for k := range want {
		if !seen[k] {
			t.Errorf("Tab never reached %s", k)
		}
	}
	// the arrows: up from the core to the crown and on outward, down again
	key := func(k tcell.Key) { v.HandleKey(tcell.NewEventKey(k, 0, tcell.ModNone), f) }
	row := func() int {
		ts := v.targetsFor(m, 160, v.cam)
		return ts[indexOf(ts, v.cur)].row
	}
	v.cur = target{kind: tCore}
	for want := 1; want <= len(rings)+1; want++ {
		key(tcell.KeyUp)
		if row() != want {
			t.Fatalf("up from row %d reached row %d", want-1, row())
		}
	}
	key(tcell.KeyUp) // nothing past the outermost ring
	if row() != len(rings)+1 {
		t.Errorf("up from the outermost ring left it for row %d", row())
	}
	// round the outermost ring and back to where it started
	start, n := v.cur, len(rings[len(rings)-1].Marks)
	round := map[target]bool{}
	for i := 0; i < n; i++ {
		round[v.cur] = true
		key(tcell.KeyRight)
		if row() != len(rings)+1 {
			t.Fatalf("right left the ring for row %d", row())
		}
	}
	if len(round) != n || v.cur != start {
		t.Errorf("right stepped through %d of the ring's %d types and ended on %+v, not back on %+v", len(round), n, v.cur, start)
	}
	key(tcell.KeyLeft)
	if !round[v.cur] || v.cur == start {
		t.Errorf("left from the ring's first did not wrap to its last: %+v", v.cur)
	}
	for i := 0; i <= len(rings); i++ {
		key(tcell.KeyDown)
	}
	if v.cur.kind != tCore {
		t.Errorf("down from the outermost ring ended on %+v, not the core", v.cur)
	}
	// a building on a ring says which era's ring it shines on
	v.cur = target{kind: tMark, key: rings[0].Marks[0].Key}
	if in, _ := v.Inspect(f); len(in.Lines) == 0 || !strings.Contains(in.Lines[0], m.Catalog.EpochName[rings[0].Epoch]) {
		t.Errorf("a building of the first ring does not name its era: %+v", in)
	}
	v.cur = target{kind: tMark, key: m.Crown()[0].Key}
	if in, _ := v.Inspect(f); len(in.Lines) == 0 || !strings.Contains(in.Lines[0], "crown") {
		t.Errorf("a building of this age does not say it is in the crown: %+v", in)
	}

	// an era reached with nothing of it left standing is still a ring to read
	st := mandalaState(7)
	gone := map[string]bool{}
	for _, a := range config.Epochs()[0].Ages {
		gone[a] = true
	}
	for k, b := range st.Buildings {
		if gone[b.AgeKey] {
			delete(st.Buildings, k)
		}
	}
	bare := mapmodel.NewBuilder(catalog()).Build(&st, nil)
	if r := bare.Mandala(); len(r) != 7 || len(r[0].Marks) != 0 {
		t.Fatalf("the bare fixture: %d rings, %d marks on the first", len(r), len(r[0].Marks))
	}
	v = newView()
	fb := mapstyle.Frame{Model: bare}
	_ = draw(v, bare, 160, 48, 1, mapmodel.TierUnicode, false)
	found := false
	for k := 0; k < 400 && !found; k++ {
		v.HandleKey(tabKey(), fb)
		if v.cur.kind == tRing {
			in, ok := v.Inspect(fb)
			found = ok && strings.Contains(in.Title, bare.Catalog.EpochName[0]) && in.Command == mapmodel.CmdStatus
		}
	}
	if !found {
		t.Error("Tab never reached the ring of the era with nothing left standing")
	}
}

// TestSkyMandalaGrowth: the mandala is still the civilization. What is
// built in the Transcendent Age spreads the core's gold glow, and the
// crown's petals turn from white to gold with the first of it.
func TestSkyMandalaGrowth(t *testing.T) {
	glow := func(m *mapmodel.Model) (n int, petal tcell.Color) {
		v := newView()
		s := v.compose(mapstyle.Frame{Model: m, Anim: 30, Tier: mapmodel.TierUnicode}, 160, 48)
		o := newOrb(s, mapmodel.SkyMandala)
		void := o.c(mapmodel.InkVoid, iBack, 0)
		for y := 0; y < s.S; y++ {
			for x := 0; x < s.W; x++ {
				c := v.fb.at(x, s.Y(y))
				if c.d == dMandala && c.bg != void {
					n++
				}
				if c.d == dMandala && c.ch == '▲' {
					petal = c.fg
				}
			}
		}
		return n, petal
	}
	st := fixture.State(fixture.Options{Age: "transcendent_age", Seed: 12, Tick: tickAt(0.5)})
	b := mapmodel.NewBuilder(catalog())
	base := b.Build(&st, nil)
	grown := st
	for i := 1; i <= 4; i++ {
		grown = fixture.Grow(grown, i)
	}
	big := b.Build(&grown, nil)
	n0, p0 := glow(base)
	n1, _ := glow(big)
	if n1 <= n0 {
		t.Errorf("%d cells of glow after building more in the age, %d before", n1, n0)
	}
	for k, bs := range st.Buildings {
		if bs.AgeKey == "transcendent_age" {
			delete(st.Buildings, k)
		}
	}
	none := b.Build(&st, nil)
	if len(none.Crown()) != 0 {
		t.Fatal("the fixture still has a crown")
	}
	nn, pn := glow(none)
	if nn != 1 {
		t.Errorf("%d cells of glow with nothing built in the age, want the core alone", nn)
	}
	if pn == p0 {
		t.Error("the petals look the same with and without a crown")
	}
}

// TestSkyMandalaCompact: the mini map shows a small mandala and nothing
// else, in the middle of its space and symmetric about its core, with at
// least the newest era's ring.
func TestSkyMandalaCompact(t *testing.T) {
	m := mandalaModel(7)
	for _, sz := range [][2]int{{40, 15}, {65, 9}, {38, 9}} {
		W, H := sz[0], sz[1]
		for _, anim := range []int{0, 37} {
			v := newView()
			v.composeCompact(mapstyle.Frame{Model: m, Anim: anim, Tier: mapmodel.TierUnicode}, W, H)
			if v.sclay != nil {
				t.Fatalf("%dx%d: a layout of lots was built for the mini map", W, H)
			}
			l := v.md
			if l == nil || l.g.Rings < 1 || l.cx != W/2 || l.cy != (H-2)/2 {
				t.Fatalf("%dx%d: no mandala in the middle of the mini map: %+v", W, H, l)
			}
			newest := mapmodel.R(mapmodel.EraSym(6), mapmodel.TierUnicode)
			seen := false
			for y := 0; y < H-1; y++ {
				for x := 0; x < W; x++ {
					c := v.fb.at(x, y+1)
					switch c.d {
					case dMandala:
						seen = seen || c.ch == newest
						for i, q := range [2][2]int{{2*l.cx - x, y}, {x, 2*l.cy - y}} {
							if q[1] < 0 || q[1] >= H-1 || q[0] < 0 || q[0] >= W {
								continue
							}
							if mc := v.fb.at(q[0], q[1]+1); mc.d != dMandala || mdMirror(mc.ch, i == 0) != c.ch || mc.bg != c.bg {
								t.Errorf("%dx%d frame %d: %q at (%d,%d) is not mirrored at (%d,%d)", W, H, anim, c.ch, x, y, q[0], q[1])
							}
						}
					case dSky, dStar:
					default:
						t.Fatalf("%dx%d: %q at (%d,%d) is neither the void, a star nor the mandala (depth %d)", W, H, c.ch, x, y, c.d)
					}
				}
			}
			if !seen {
				t.Errorf("%dx%d: the newest era's ring is not on the mini map", W, H)
			}
		}
	}
}

// TestSkyMandalaPixelBands: where the rings stand too close for bands of
// glyphs they are drawn in half-block pixels, and no cell ever holds two
// rings, nor a ring and a cell of the layout's own.
func TestSkyMandalaPixelBands(t *testing.T) {
	for _, sz := range [][2]float64{{12, 24}, {16, 59}, {10, 39}, {5, 10}, {4, 19}} {
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
