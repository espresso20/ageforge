package skyline

import (
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/uniseg"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/mapmodel/fixture"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
	"github.com/espresso20/ageforge/ui/mapstyle/capture"
)

// skyAges are the sky arc's ages, scene by scene (mapmodel/sky.go).
var skyAges = []string{"space_age", "interstellar_age", "galactic_age", "quantum_age", "transcendent_age"}

// skyModel builds a busy sky-age state: routes, a war, the harbinger.
func skyModel(age string, seed int64, grow int) *mapmodel.Model {
	return build(fixture.Options{Age: age, Seed: seed, Routes: 3, Wars: 1, Harbinger: true, Tick: tickAt(0.6)}, grow)
}

// TestSkyScenes: every sky age draws through the sky compose (a sky
// layout in the layout slot, which layoutFor never mistakes for its own)
// and the Fusion Age still draws on the ground. The Transcendent Age has
// no lots to lay out (TestSkyMandalaAlone).
func TestSkyScenes(t *testing.T) {
	for i, age := range skyAges {
		m := skyModel(age, 3, 0)
		if m.Sky() != mapmodel.SkyScene(i+1) {
			t.Fatalf("%s: scene %d", age, m.Sky())
		}
		v := newView()
		s := v.compose(mapstyle.Frame{Model: m, Anim: 5}, 160, 48)
		if m.Sky() == mapmodel.SkyMandala {
			if v.slay != nil || v.lay != nil {
				t.Errorf("%s: a layout of lots for the mandala", age)
			}
			continue
		}
		if v.slay == nil || v.slay.groundY >= 0 || s.groundY != skyGround(45) {
			t.Fatalf("%s: not the sky compose (layout key %d, baseline %d)", age, v.slay.groundY, s.groundY)
		}
		lay := v.slay
		_ = v.compose(mapstyle.Frame{Model: m, Anim: 6}, 160, 48)
		if v.slay != lay {
			t.Errorf("%s: the sky layout was rebuilt for the next frame", age)
		}
		if g := v.layoutFor(m, s.groundY); g == lay || g.groundY != s.groundY {
			t.Errorf("%s: layoutFor took the sky layout for a ground one", age)
		}
	}
	m := skyModel("fusion_age", 3, 0)
	v := newView()
	if s := v.compose(mapstyle.Frame{Model: m}, 160, 48); m.Sky() != mapmodel.SkyGround || v.lay.groundY != s.S-3 {
		t.Error("fusion_age left the ground")
	}
}

// TestSkyEverySize: every sky age at every size, the compact views at 40x15
// and 65x9 and the tiny sizes write every cell of their rect and nothing
// outside it.
func TestSkyEverySize(t *testing.T) {
	for i, age := range skyAges {
		for k, o := range []fixture.Options{
			{Age: age, Seed: int64(i + 2), Tick: tickAt(0.5)},
			{Age: age, Seed: int64(i + 7), Tick: tickAt(0.95), Harbinger: true, Wars: 1, Catastrophe: true, Routes: 4},
		} {
			m := build(o, k)
			v := newView()
			v.flows, v.inspect = k == 1, k == 1
			for _, sz := range [][2]int{{200, 60}, {160, 48}, {120, 40}, {100, 30}, {80, 24}, {60, 16}} {
				drawBoxed(t, v, m, sz[0], sz[1], 10+i, mapmodel.TierUnicode, false)
			}
			for _, sz := range [][2]int{{40, 15}, {65, 9}} {
				drawBoxed(t, v, m, sz[0], sz[1], 3, mapmodel.TierUnicode, true)
				drawBoxed(t, v, m, sz[0], sz[1], 4, mapmodel.TierASCII, true)
			}
		}
	}
	for _, age := range skyAges {
		m := build(fixture.Options{Age: age, Seed: 2, Harbinger: true, Catastrophe: true}, 1)
		v := newView()
		v.inspect, v.flows = true, true
		for w := 1; w <= 70; w += 3 {
			for h := 1; h <= 19; h += 2 {
				drawBoxed(t, v, m, w, h, 5, mapmodel.TierUnicode, false)
				drawBoxed(t, v, m, w, h, 5, mapmodel.TierASCII, true)
			}
		}
	}
}

// TestSkyDeterminismAndLife: the same frame draws the same cells twice, in
// the full view and the compact one, and every sky age moves.
func TestSkyDeterminismAndLife(t *testing.T) {
	for _, age := range skyAges {
		m := skyModel(age, 9, 1)
		a := screenCells(draw(newView(), m, 160, 45, 33, mapmodel.TierUnicode, false))
		b := screenCells(draw(newView(), m, 160, 45, 33, mapmodel.TierUnicode, false))
		if !sameCells(a, b) {
			t.Errorf("%s: the same frame drew differently twice", age)
		}
		if c := screenCells(draw(newView(), m, 160, 45, 61, mapmodel.TierUnicode, false)); sameCells(a, c) {
			t.Errorf("%s: two frames drew the same cells: nothing moves", age)
		}
		v := newView()
		x := screenCells(draw(v, m, 160, 45, 90, mapmodel.TierUnicode, false))
		y := screenCells(draw(v, m, 160, 45, 90, mapmodel.TierUnicode, false))
		if !sameCells(x, y) {
			t.Errorf("%s: a view redrew the same frame differently (stale cache)", age)
		}
		for _, sz := range [][2]int{{40, 15}, {65, 9}} {
			ca := screenCells(draw(newView(), m, sz[0], sz[1], 33, mapmodel.TierUnicode, true))
			cb := screenCells(draw(newView(), m, sz[0], sz[1], 33, mapmodel.TierUnicode, true))
			if !sameCells(ca, cb) {
				t.Errorf("%s %dx%d: the compact view is not deterministic", age, sz[0], sz[1])
			}
		}
	}
}

// TestSkyThemeSweep: in every theme no drawn glyph has its ink equal to its
// background, in the full and the compact views.
func TestSkyThemeSweep(t *testing.T) {
	orig := theme.Active().Key
	t.Cleanup(func() { _ = theme.SetActive(orig) })
	models := map[string]*mapmodel.Model{}
	for _, a := range skyAges {
		models[a] = build(fixture.Options{Age: a, Seed: 5, Tick: tickAt(0.3), Harbinger: true, Routes: 3, Catastrophe: a == "quantum_age"}, 1)
	}
	for _, th := range theme.All() {
		if err := theme.SetActive(th.Key); err != nil {
			t.Fatal(err)
		}
		for name, m := range models {
			for _, sz := range [][3]int{{120, 36, 0}, {40, 15, 1}, {65, 9, 1}} {
				v := newView()
				v.flows, v.inspect = true, true
				scr := draw(v, m, sz[0], sz[1], 7, mapmodel.TierUnicode, sz[2] == 1)
				cells, sw, _ := scr.GetContents()
				for i, c := range cells {
					fg, bg, _ := c.Style.Decompose()
					if len(c.Runes) > 0 && c.Runes[0] != ' ' && fg == bg {
						t.Errorf("theme %s, %s %dx%d: %q at (%d,%d) has fg == bg", th.Key, name, sz[0], sz[1], c.Runes[0], i%sw, i/sw)
						break
					}
				}
			}
		}
	}
}

// TestSkyGlyphs: every rune a sky scene draws is one cell wide, folds to
// ASCII in the ASCII tier (never to the '?' of an unknown rune) and none is
// an em dash, in every tier, full and compact.
func TestSkyGlyphs(t *testing.T) {
	for _, age := range skyAges {
		m := build(fixture.Options{Age: age, Seed: 6, Routes: 4, Harbinger: true, Catastrophe: true, Wars: 1, Tick: tickAt(0.9)}, 1)
		for _, tier := range []mapmodel.GlyphTier{mapmodel.TierUnicode, mapmodel.TierNerd, mapmodel.TierASCII} {
			for _, sz := range [][3]int{{160, 45, 0}, {40, 15, 1}, {65, 9, 1}} {
				for _, anim := range []int{21, 140} {
					v := newView()
					v.inspect, v.flows = true, true
					cells, _, _ := draw(v, m, sz[0], sz[1], anim, tier, sz[2] == 1).GetContents()
					for _, c := range cells {
						if len(c.Runes) == 0 {
							continue
						}
						r := c.Runes[0]
						switch {
						case tier == mapmodel.TierASCII && r >= 0x80:
							t.Fatalf("%s ascii %dx%d: non-ASCII rune %q", age, sz[0], sz[1], r)
						case r == '?':
							t.Fatalf("%s %s %dx%d: a '?' fallback in the frame", age, tier, sz[0], sz[1])
						case tier != mapmodel.TierASCII && uniseg.StringWidth(string(r)) != 1:
							t.Fatalf("%s %s: rune %q is not one cell wide", age, tier, r)
						case r == '—':
							t.Fatalf("%s: an em dash in the frame", age)
						}
					}
				}
			}
		}
	}
}

// TestSkyInspect: in every sky age with lots Tab reaches every one, every
// target says what it is with a command the model knows, and the status
// line words a lot in the scene's terms (mapmodel.SkyPartOf). The
// Transcendent Age's cursor reads the mandala (TestSkyMandalaInspect).
func TestSkyInspect(t *testing.T) {
	for i, age := range skyAges {
		m := skyModel(age, 3, 1)
		sc := mapmodel.SkyScene(i + 1)
		if sc == mapmodel.SkyMandala {
			continue
		}
		words := commandWords(m)
		v := newView()
		f := mapstyle.Frame{Model: m}
		_ = draw(v, m, 160, 48, 1, mapmodel.TierUnicode, false)
		want := map[target]bool{}
		for _, l := range m.Skyline.Lots {
			if m.Catalog.Defs[l.Key] != nil {
				want[target{tLot, l.Key, l.Copy}] = true
			}
		}
		seen := map[target]bool{}
		kinds := map[tkind]int{}
		for k := 0; k < len(m.Skyline.Lots)+24; k++ {
			v.HandleKey(tabKey(), f)
			in, ok := v.Inspect(f)
			if !ok || in.Title == "" {
				t.Fatalf("%s: step %d: nothing under the cursor", age, k)
			}
			if in.Command == "" || !words[strings.Fields(in.Command)[0]] {
				t.Fatalf("%s: %q: command %q is not a model command", age, in.Title, in.Command)
			}
			seen[v.cur] = true
			kinds[v.cur.kind]++
		}
		for tg := range want {
			if !seen[tg] {
				t.Errorf("%s: Tab never reached %s copy %d", age, tg.key, tg.cp)
			}
		}
		if kinds[tCiv] == 0 || kinds[tHarbinger] == 0 {
			t.Errorf("%s: Tab did not reach the civs and the harbinger: %v", age, kinds)
		}
		// the status line names every lot in the scene's terms, with its
		// command (the details give way before the command does)
		order := make([]target, 0, len(want))
		for tg := range want {
			order = append(order, tg)
		}
		sort.Slice(order, func(i, j int) bool {
			if order[i].key != order[j].key {
				return order[i].key < order[j].key
			}
			return order[i].cp < order[j].cp
		})
		for _, tg := range order {
			def := m.Catalog.Defs[tg.key]
			part := mapmodel.SkyPartOf(sc, def.Lineage)
			if def.Wonder || part.Name == mapmodel.LineageNames[def.Lineage] || tg.cp > 0 {
				continue
			}
			v.cur = tg
			v.reveal = true
			scr := draw(v, m, 160, 48, 2, mapmodel.TierUnicode, false)
			if st := rowText(scr, 47); !strings.Contains(st, part.Name) || !strings.Contains(st, "type: ") {
				t.Errorf("%s: the status line for %s does not say %q with its command: %q", age, tg.key, part.Name, st)
			}
			if c := v.fb.at(0, 47); c == nil {
				t.Fatal("no status line")
			}
		}
		// the cursor lands on a lot: its bracket under the baseline
		for tg := range want {
			v.cur, v.reveal = tg, true
			break
		}
		scr := draw(v, m, 160, 48, 3, mapmodel.TierUnicode, false)
		if !strings.ContainsAny(capture.Text(scr), "└┘") {
			t.Errorf("%s: no cursor bracket under the inspected lot", age)
		}
	}
}

// lotCells counts the cells the lots own over the whole panorama.
func lotCells(m *mapmodel.Model) int {
	v := newView()
	v.follow = false
	n := 0
	for cam := 0; cam < m.Skyline.Width; cam += 160 {
		v.cam = cam
		v.compose(mapstyle.Frame{Model: m, Anim: 3}, 160, 48)
		for _, c := range v.fb.c {
			if c.d >= dWonder && c.d <= dRow2 {
				n++
			}
		}
	}
	return n
}

// TestSkyGrowth: the scene is still the civilization: a grown state draws
// more of it in every sky age with lots (the Transcendent Age grows its
// core's glow: TestSkyMandalaGrowth), and the Space Age's cities light up
// more under a district with more built.
func TestSkyGrowth(t *testing.T) {
	for _, age := range skyAges {
		if age == "transcendent_age" {
			continue
		}
		o := fixture.Options{Age: age, Seed: 12, Tick: tickAt(0.5)}
		base := build(o, 0)
		grown := build(o, 4)
		if a, b := lotCells(base), lotCells(grown); b <= a {
			t.Errorf("%s: %d lot cells grown, %d before: building more did not draw more", age, b, a)
		}
	}
	m := build(fixture.Options{Age: "space_age", Seed: 12, Tick: tickAt(0.5)}, 0)
	v := newView()
	s := v.compose(mapstyle.Frame{Model: m, Anim: 3}, 160, 48)
	o := newOrb(s, mapmodel.SkyOrbit)
	thin := *m
	thin.Skyline.Lots = m.Skyline.Lots[:len(m.Skyline.Lots)/3]
	o2 := newOrb(&scene{m: &thin, W: s.W, cam: s.cam, p: s.p, v: v}, mapmodel.SkyOrbit)
	sum := func(xs []float64) (t float64) {
		for _, x := range xs {
			t += x
		}
		return t
	}
	if a, b := sum(o2.cityDensity()), sum(o.cityDensity()); b <= a {
		t.Errorf("more buildings did not light more cities: %.2f against %.2f", b, a)
	}
}

// sceneOfDepth names the scene each signature depth belongs to.
var sceneOfDepth = map[uint8]mapmodel.SkyScene{
	dTruss: mapmodel.SkyOrbit, dPlanet: mapmodel.SkyOrbit, dMoon: mapmodel.SkyOrbit, dBelt: mapmodel.SkyOrbit,
	dTether:    mapmodel.SkyOrbit,
	dFormation: mapmodel.SkyDeep, dGate: mapmodel.SkyDeep, dNebula: mapmodel.SkyDeep, dOrrery: mapmodel.SkyDeep,
	dOrrFront: mapmodel.SkyDeep, dOrrSun: mapmodel.SkyDeep, dOrrBack: mapmodel.SkyDeep,
	dColony:   mapmodel.SkyDeep,
	dDockRing: mapmodel.SkyGalaxy, dHub: mapmodel.SkyGalaxy, dSpiral: mapmodel.SkyGalaxy,
	dGalNebula: mapmodel.SkyGalaxy, dSystem: mapmodel.SkyGalaxy,
	dEchoLine: mapmodel.SkyQuantum, dEcho: mapmodel.SkyQuantum, dFractal: mapmodel.SkyQuantum,
	dHaloQ: mapmodel.SkyQuantum, dFringe: mapmodel.SkyQuantum,
	dMandala: mapmodel.SkyMandala,
}

// TestSkyNoSpoilers: no sky age draws anything of a later scene: none of
// their signature structures (each has a depth of its own) and none of
// their traffic, and no vehicle shows before the age that introduces its
// mover or after the one that retires it. Each age does draw its own.
func TestSkyNoSpoilers(t *testing.T) {
	for i, age := range append([]string{"fusion_age"}, skyAges...) {
		sc := mapmodel.SkyScene(i)
		m := build(fixture.Options{Age: age, Seed: 4, Routes: 6, Wars: 1, Harbinger: true, Scale: 1.5, Tick: tickAt(0.5)}, 1)
		own := map[uint8]bool{}
		v := newView()
		v.follow = false
		for cam := 0; cam < m.Skyline.Width; cam += 120 {
			for _, anim := range []int{0, 37, 211} {
				v.cam = cam
				v.compose(mapstyle.Frame{Model: m, Anim: anim}, 160, 48)
				for _, c := range v.fb.c {
					if o, ok := sceneOfDepth[c.d]; ok {
						if o > sc && sc != mapmodel.SkyGround {
							t.Fatalf("%s: a cell of a later scene's structure (depth %d, scene %d)", age, c.d, o)
						}
						if o == sc {
							own[c.d] = true
						}
					}
				}
			}
		}
		if sc == mapmodel.SkyGround {
			continue
		}
		if sc == mapmodel.SkyMandala { // the mandala alone, and nothing moving: TestSkyMandalaAlone
			if !own[dMandala] {
				t.Errorf("%s: no mandala", age)
			}
			continue
		}
		if len(own) < 3 {
			t.Errorf("%s: it drew only %d of its own signature structures", age, len(own))
		}
		// the traffic
		s := newView().compose(mapstyle.Frame{Model: m}, 160, 48)
		o := newOrb(s, sc)
		movers := map[mapmodel.Mover]bool{}
		for cam := 0; cam < m.Skyline.Width; cam += 120 {
			for anim := 0; anim < 2000; anim += 37 {
				g := o.geo()
				g.cam, g.anim = cam, anim
				for _, vh := range skyTrafficFor(g) {
					k := vh.t.mover
					if k == mapmodel.MoverNone {
						continue
					}
					movers[k] = true
					if !mapmodel.Introduced(k, m.AgeIdx) || !k.Info().In(m.AgeIdx) {
						t.Fatalf("%s: a %s out of its ages", age, k.Info().Key)
					}
				}
			}
		}
		sig := map[mapmodel.SkyScene]mapmodel.Mover{mapmodel.SkyOrbit: mapmodel.MoverClimber, mapmodel.SkyDeep: mapmodel.MoverGenShip,
			mapmodel.SkyGalaxy: mapmodel.MoverStarship, mapmodel.SkyQuantum: mapmodel.MoverPhaseShip}
		if !movers[sig[sc]] {
			t.Errorf("%s: no %s, its signature mover", age, sig[sc].Info().Key)
		}
		if sc == mapmodel.SkyGalaxy && !movers[mapmodel.MoverAlienShip] {
			t.Errorf("%s: no alien ships in the traffic", age)
		}
	}
}

// TestSkyVisitors: in the Space and Interstellar Ages the visitor's saucer
// still draws during a visit, with the cursor's mark over it, and never in
// the compact view; in the Galactic Age the saucers are traffic.
func TestSkyVisitors(t *testing.T) {
	dome := mapmodel.R(mapmodel.SymUFO, mapmodel.TierUnicode)
	for _, age := range []string{"space_age", "interstellar_age"} {
		m := build(fixture.Options{Age: age, Seed: 4, Tick: tickAt(0.95)}, 0)
		sg := flyby(m.Seed)
		shown := false
		for anim := sg.Start + sg.Frames/2 - 6; anim <= sg.Start+sg.Frames/2+6; anim++ {
			_, _, x, y, r, hidden, ok := domeAt(m, anim, mapmodel.TierUnicode)
			if !ok {
				t.Fatalf("%s frame %d: no visit mid-visit", age, anim)
			}
			if r == dome {
				shown = true
			} else if !hidden && x >= 0 && x < 160 && y >= 1 {
				t.Errorf("%s frame %d: %q at the dome and nothing in front of it", age, anim, r)
			}
		}
		if !shown {
			t.Errorf("%s: the saucer never showed mid-visit", age)
		}
		for f := sg.Start; f < sg.Start+sg.Frames; f += 7 {
			for _, sz := range [][2]int{{40, 15}, {65, 9}} {
				if strings.ContainsRune(capture.Text(draw(newView(), m, sz[0], sz[1], f, mapmodel.TierUnicode, true)), dome) {
					t.Fatalf("%s frame %d: the visitor shows in the compact view", age, f)
				}
			}
		}
		// outside a visit, no saucer anywhere
		if txt := capture.Text(draw(newView(), m, 160, 45, sg.Start-1, mapmodel.TierUnicode, false)); strings.ContainsRune(txt, dome) {
			t.Errorf("%s: a saucer with no visit showing", age)
		}
	}
	m := build(fixture.Options{Age: "galactic_age", Seed: 4, Tick: tickAt(0.5)}, 0)
	seen := false
	for anim := 0; anim < 400 && !seen; anim += 13 {
		seen = strings.ContainsRune(capture.Text(draw(newView(), m, 160, 48, anim, mapmodel.TierUnicode, false)), dome)
	}
	if !seen {
		t.Error("galactic_age: no alien saucers in the traffic")
	}
}

// TestSkyPerformance: a 200x60 frame of every sky age stays well inside the
// redraw budget.
func TestSkyPerformance(t *testing.T) {
	if testing.Short() || raceEnabled {
		t.Skip("timing")
	}
	for _, age := range skyAges {
		m := build(fixture.Options{Age: age, Seed: 7, Routes: 5, Tick: tickAt(0.9), Scale: 2}, 1)
		v := newView()
		scr := capture.NewScreen(200, 60)
		f := mapstyle.Frame{Model: m}
		v.Draw(scr, mapstyle.Rect{W: 200, H: 60}, f)
		const n = 30
		start := time.Now()
		for i := 0; i < n; i++ {
			f.Anim = i
			v.Draw(scr, mapstyle.Rect{W: 200, H: 60}, f)
		}
		per := time.Since(start) / n
		t.Logf("200x60 %s frame: %v", age, per)
		if per > 25*time.Millisecond {
			t.Errorf("%s: a frame takes %v", age, per)
		}
		scr.Fini()
	}
}

// TestSkyCaptures writes review captures of the sky scenes from the real
// smoke-bot states when SKYLINE_SKY_CAPTURES names a directory:
// .txt, .html (animated) and .png per shot.
func TestSkyCaptures(t *testing.T) {
	dir := os.Getenv("SKYLINE_SKY_CAPTURES")
	if dir == "" {
		t.Skip("set SKYLINE_SKY_CAPTURES to a directory to write sky captures")
	}
	states := filepath.Join("..", "..", "..", "map_captures", "states")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	type shot struct {
		name, state, theme string
		w, h               int
		compact            bool
		tier               mapmodel.GlyphTier
		frames             int
		anim0              int
		inspect            int // Tab presses
	}
	var shots []shot
	for _, a := range []string{"fusion_age", "space_age", "interstellar_age", "galactic_age", "quantum_age", "transcendent_age"} {
		n := strings.TrimSuffix(a, "_age")
		shots = append(shots, shot{name: n, state: a, w: 160, h: 48, frames: 8, anim0: 40},
			shot{name: n + "_mini", state: a, w: 40, h: 15, compact: true, frames: 4, anim0: 40},
			shot{name: n + "_strip", state: a, w: 65, h: 9, compact: true, frames: 4, anim0: 40})
	}
	shots = append(shots,
		shot{name: "space_light", state: "space_age", theme: "daylight", w: 160, h: 48, frames: 2, anim0: 40},
		shot{name: "galactic_light", state: "galactic_age", theme: "daylight", w: 160, h: 48, frames: 2, anim0: 40},
		shot{name: "transcendent_light", state: "transcendent_age", theme: "daylight", w: 160, h: 48, frames: 2, anim0: 40},
		shot{name: "space_mono", state: "space_age", theme: "monochrome", w: 160, h: 48, frames: 2, anim0: 40},
		shot{name: "quantum_mono", state: "quantum_age", theme: "monochrome", w: 160, h: 48, frames: 2, anim0: 40},
		shot{name: "interstellar_parchment", state: "interstellar_age", theme: "parchment", w: 160, h: 48, frames: 2, anim0: 40},
		shot{name: "space_ascii", state: "space_age", tier: mapmodel.TierASCII, w: 160, h: 48, frames: 2, anim0: 40},
		shot{name: "galactic_inspect", state: "galactic_age", w: 160, h: 48, frames: 2, anim0: 40, inspect: 3},
		shot{name: "space_80x24", state: "space_age", w: 80, h: 24, frames: 2, anim0: 40},
		shot{name: "galactic_200x60", state: "galactic_age", w: 200, h: 60, frames: 2, anim0: 40},
	)
	only := os.Getenv("SKYLINE_SKY_ONLY")
	orig := theme.Active().Key
	t.Cleanup(func() { _ = theme.SetActive(orig) })
	b := mapmodel.NewBuilder(nil)
	for _, sh := range shots {
		if only != "" && !strings.Contains(sh.name, only) {
			continue
		}
		st, err := capture.LoadState(filepath.Join(states, sh.state+".json.gz"))
		if err != nil {
			t.Logf("skip %s: %v", sh.name, err)
			continue
		}
		th := sh.theme
		if th == "" {
			th = "forge"
		}
		if err := theme.SetActive(th); err != nil {
			t.Fatal(err)
		}
		m := b.Build(&st, nil)
		v := newView()
		f := mapstyle.Frame{Model: m, Tier: sh.tier}
		for i := 0; i < sh.inspect; i++ {
			f.Anim = sh.anim0
			_ = draw(v, m, sh.w, sh.h, sh.anim0, sh.tier, sh.compact)
			v.HandleKey(tabKey(), f)
		}
		var frames []string
		var txt strings.Builder
		for k := 0; k < max(1, sh.frames); k++ {
			scr := draw(v, m, sh.w, sh.h, sh.anim0+k*3, sh.tier, sh.compact)
			frames = append(frames, capture.Frame(scr))
			if k == 0 {
				txt.WriteString(capture.Text(scr))
				if err := capture.PNG(filepath.Join(dir, sh.name+".png"), scr); err != nil {
					t.Fatal(err)
				}
				_ = os.WriteFile(filepath.Join(dir, sh.name+".cells"), []byte(cellDump(scr)), 0o644)
			}
			scr.Fini()
		}
		bg := capture.Hex(theme.Color(theme.RoleBackground))
		fg := capture.Hex(theme.Color(theme.RoleText))
		title := "skyline · " + sh.state + " · " + sh.name + " · theme " + th
		_ = os.WriteFile(filepath.Join(dir, sh.name+".txt"), []byte(title+"\n"+txt.String()), 0o644)
		_ = os.WriteFile(filepath.Join(dir, sh.name+".html"), []byte(capture.HTML(title, bg, fg, capture.Fonts, frames, 6)), 0o644)
	}
}

func tabKey() *tcell.EventKey { return tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone) }

// cellDump writes a screen's cells for an outside renderer: "W H", then a
// line per row of tab-separated "rune fg bg" (decimal rune, hex colours).
func cellDump(scr tcell.SimulationScreen) string {
	cells, w, h := scr.GetContents()
	var b strings.Builder
	b.WriteString(strconv.Itoa(w) + " " + strconv.Itoa(h) + "\n")
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c := cells[y*w+x]
			r := ' '
			if len(c.Runes) > 0 {
				r = c.Runes[0]
			}
			fg, bg, _ := c.Style.Decompose()
			if x > 0 {
				b.WriteByte('\t')
			}
			b.WriteString(strconv.Itoa(int(r)) + " " + capture.Hex(fg) + " " + capture.Hex(bg))
		}
		b.WriteByte('\n')
	}
	return b.String()
}

// TestSkyTetherCarriesOn: the space elevator the Earth arc raises in the
// Fusion Age is the one the Space Age carries up to the station: it stands
// on the same world column, it is drawn once, from the planet's limb up,
// and the cursor can still inspect it. From the Interstellar Age, out in
// deep space, there is none, and no Tab target for it.
func TestSkyTetherCarriesOn(t *testing.T) {
	fus := skyModel("fusion_age", 6, 0)
	spc := skyModel("space_age", 6, 0)
	vf := newView()
	sf := vf.compose(mapstyle.Frame{Model: fus}, 160, 48)
	wf := sf.elevatorX() + sf.cam
	if sf.elevatorX() < 0 {
		t.Fatal("the Fusion Age has no tether")
	}
	if w := tetherX(spc); w != wf {
		t.Errorf("the Space Age's tether stands at world column %d, the Fusion Age's at %d", w, wf)
	}
	vs := newView()
	vs.follow, vs.cam = false, max(0, wf-80)
	ss := vs.compose(mapstyle.Frame{Model: spc}, 160, 48)
	x := wf - ss.cam
	if x < 0 || x >= ss.W {
		t.Fatalf("the tether is off the view at column %d", x)
	}
	// columns carrying a long run of the tether's cable
	cables := 0
	for cx := 0; cx < ss.W; cx++ {
		run, best := 0, 0
		for y := 0; y < ss.S; y++ {
			if c := vs.fb.at(cx, ss.Y(y)); c != nil && c.d == dTether && c.ch == '│' {
				run++
				best = max(best, run)
			} else {
				run = 0
			}
		}
		if best >= 6 {
			cables++
			if cx != x {
				t.Errorf("a tether's cable at column %d, not the elevator's %d", cx, x)
			}
		}
	}
	if cables != 1 {
		t.Errorf("%d tether cables in the Space Age, want 1", cables)
	}
	reached := false
	for _, tg := range targets(spc, 160, ss.cam) {
		reached = reached || tg.kind == tTether && tg.x == wf
	}
	if !reached {
		t.Error("the Space Age's tether is not a Tab target")
	}
	for _, age := range []string{"interstellar_age", "galactic_age", "transcendent_age"} {
		m := skyModel(age, 6, 0)
		if tetherX(m) >= 0 {
			t.Errorf("%s: a tether out in deep space", age)
		}
		for _, tg := range targets(m, 160, 0) {
			if tg.kind == tTether {
				t.Errorf("%s: a Tab target for a tether that is not drawn", age)
			}
		}
	}
}
