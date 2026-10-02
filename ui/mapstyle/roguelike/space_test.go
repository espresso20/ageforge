package roguelike

import (
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

// skyAges are the ages with a scene of their own, in order.
var skyAges = []string{"space_age", "interstellar_age", "galactic_age", "quantum_age", "transcendent_age"}

func skyStyleFor() *skyStyle { return withSky(newView()) }

func drawStyle(st mapstyle.Style, m *mapmodel.Model, w, h, anim int, tier mapmodel.GlyphTier, compact bool) tcell.SimulationScreen {
	s := capture.NewScreen(w+4, h+2)
	fillScreen(s)
	r := mapstyle.Rect{X: 2, Y: 1, W: w, H: h}
	f := mapstyle.Frame{Model: m, Anim: anim, Tier: tier}
	if compact {
		st.DrawCompact(s, r, f)
	} else {
		st.Draw(s, r, f)
	}
	s.Show()
	return s
}

// TestSkyEveryAgeEverySize: every sky scene, at both of its zooms and the
// mini view, writes every cell of its rect and nothing outside it, at every
// size down to a sliver.
func TestSkyEveryAgeEverySize(t *testing.T) {
	sizes := [][2]int{{200, 60}, {160, 48}, {120, 40}, {100, 30}, {80, 24}, {40, 15}, {65, 9}, {13, 5}, {3, 2}, {1, 1}}
	for ai, age := range skyAges {
		m := modelFor(t, opts(ai, age))
		if m.Sky() == mapmodel.SkyGround {
			t.Fatalf("%s is on the ground", age)
		}
		for _, z := range []int{zSettlement, zDistrict} {
			st := skyStyleFor()
			st.g.zoom, st.g.flows = z, ai%2 == 1
			for _, sz := range sizes {
				for _, compact := range []bool{false, true} {
					s := drawStyle(st, m, sz[0], sz[1], ai*7, mapmodel.TierUnicode, compact)
					checkRect(t, s, mapstyle.Rect{X: 2, Y: 1, W: sz[0], H: sz[1]},
						age+" zoom "+strconv.Itoa(z)+" "+strconv.Itoa(sz[0])+"x"+strconv.Itoa(sz[1]))
				}
			}
		}
	}
}

// TestSkyDispatch: the sky scenes take over from the Space Age at the
// settlement and district zooms; the region zoom stays the known world;
// ages before keep the town.
func TestSkyDispatch(t *testing.T) {
	ground := modelFor(t, fixture.Options{Age: "fusion_age", Seed: 3})
	space := modelFor(t, fixture.Options{Age: "space_age", Seed: 3})
	st := skyStyleFor()
	if st.onSky(ground) {
		t.Error("the Fusion Age draws the sky")
	}
	if !st.onSky(space) {
		t.Error("the Space Age draws the ground")
	}
	f := mapstyle.Frame{Model: space}
	st.HandleKey(tcell.NewEventKey(tcell.KeyPgUp, 0, tcell.ModNone), f)
	if st.g.zoom != zRegion || st.onSky(space) {
		t.Errorf("PgUp from the sky: zoom %d, sky %v", st.g.zoom, st.onSky(space))
	}
	txt := capture.Text(drawStyle(st, space, 160, 48, 0, mapmodel.TierUnicode, false))
	if !strings.Contains(txt, "REGION") {
		t.Error("the known world does not draw at the region zoom")
	}
	st.HandleKey(tcell.NewEventKey(tcell.KeyPgDn, 0, tcell.ModNone), f)
	if st.g.zoom != zSettlement || !st.onSky(space) {
		t.Errorf("PgDn from the known world: zoom %d", st.g.zoom)
	}
	st.HandleKey(tcell.NewEventKey(tcell.KeyPgDn, 0, tcell.ModNone), f)
	if st.g.zoom != zDistrict {
		t.Errorf("PgDn in the sky: zoom %d", st.g.zoom)
	}
	st.SetOption(mapstyle.OptWorld, true)
	if st.g.zoom != zRegion {
		t.Error("OptWorld does not open the known world")
	}
	if st.Name() != "roguelike" || !st.CompactShowsNews() {
		t.Error("the wrapped style lost its name or its news")
	}
	if _, ok := Entry().New().(*skyStyle); !ok {
		t.Error("the registry entry is not wrapped for the sky")
	}
}

// TestSkyDeterministicAndAlive: the same frame draws the same cells, a fresh
// view draws them too, and something moves between frames.
func TestSkyDeterministicAndAlive(t *testing.T) {
	for _, age := range skyAges {
		m, _ := grown(t, fixture.Options{Age: age, Seed: 9, Harbinger: true})
		for _, z := range []int{zSettlement, zDistrict} {
			st := skyStyleFor()
			st.g.zoom = z
			a := capture.Text(drawStyle(st, m, 160, 48, 400, mapmodel.TierUnicode, false))
			b := capture.Text(drawStyle(st, m, 160, 48, 400, mapmodel.TierUnicode, false))
			if a != b {
				t.Errorf("%s zoom %d: the same frame drew differently", age, z)
			}
			st2 := skyStyleFor()
			st2.g.zoom = z
			if c := capture.Text(drawStyle(st2, m, 160, 48, 400, mapmodel.TierUnicode, false)); c != a {
				t.Errorf("%s zoom %d: a fresh view drew differently", age, z)
			}
			moved := false
			for _, f := range []int{403, 431, 487, 640} {
				moved = moved || capture.Text(drawStyle(st, m, 160, 48, f, mapmodel.TierUnicode, false)) != a
			}
			if !moved {
				t.Errorf("%s zoom %d: nothing moved", age, z)
			}
		}
	}
}

// TestSkyThemeSweep: in every theme no glyph is drawn in its background
// colour, in any sky scene.
func TestSkyThemeSweep(t *testing.T) {
	orig := theme.Active().Key
	t.Cleanup(func() { _ = theme.SetActive(orig) })
	var models []*mapmodel.Model
	for i, age := range skyAges {
		m, _ := grown(t, fixture.Options{Age: age, Seed: 5, Harbinger: true, Wars: 1, Catastrophe: i%2 == 0})
		models = append(models, m)
	}
	for _, th := range theme.All() {
		if err := theme.SetActive(th.Key); err != nil {
			t.Fatal(err)
		}
		for _, m := range models {
			st := skyStyleFor()
			st.g.flows = true
			for _, compact := range []bool{false, true} {
				s := drawStyle(st, m, 140, 44, 2, mapmodel.TierUnicode, compact)
				cells, _, _ := s.GetContents()
				for i, c := range cells {
					if len(c.Runes) == 0 || c.Runes[0] == ' ' || c.Runes[0] == sentinel {
						continue
					}
					if fg, bg, _ := c.Style.Decompose(); fg == bg {
						t.Errorf("theme %s, %s: cell %d rune %q has fg == bg", th.Key, m.Age, i, c.Runes[0])
						break
					}
				}
			}
		}
	}
}

// TestSkyGlyphWidths: every sky scene draws one cell per rune in every
// glyph tier, ASCII only in the ASCII tier, and never the canvas's '?'
// fallback.
func TestSkyGlyphWidths(t *testing.T) {
	for _, age := range skyAges {
		m, _ := grown(t, fixture.Options{Age: age, Seed: 2, Harbinger: true, Wars: 1, Catastrophe: true, Ruins: true})
		for _, tier := range []mapmodel.GlyphTier{mapmodel.TierASCII, mapmodel.TierUnicode, mapmodel.TierNerd} {
			for _, z := range []int{zSettlement, zDistrict} {
				st := skyStyleFor()
				st.g.zoom, st.g.flows = z, true
				for _, compact := range []bool{false, true} {
					for _, anim := range []int{1, 77, 333} {
						s := drawStyle(st, m, 150, 46, anim, tier, compact)
						cells, sw, _ := s.GetContents()
						for i, c := range cells {
							if len(c.Runes) == 0 || c.Runes[0] == sentinel {
								continue
							}
							r := c.Runes[0]
							switch {
							case r == '?' && i/sw != 46:
								t.Fatalf("%s %s zoom %d frame %d: fallback '?' at row %d", age, tier, z, anim, i/sw)
							case tier == mapmodel.TierASCII && r >= 0x80:
								t.Fatalf("%s ascii zoom %d frame %d: non-ASCII rune %q", age, z, anim, r)
							case uniseg.StringWidth(string(r)) != 1:
								t.Fatalf("%s %s zoom %d: rune %q is not one cell", age, tier, z, r)
							}
						}
					}
				}
			}
		}
	}
}

// TestSkyInspect: Tab steps through every target, each with a title, and a
// building, a civ or a wonder gives a command from the model; the inspector
// line shows it; Home goes back to the hub.
func TestSkyInspect(t *testing.T) {
	for _, age := range skyAges {
		m, _ := grown(t, fixture.Options{Age: age, Seed: 7, Harbinger: true, Wars: 1, Ruins: true})
		words := commandWords(m)
		st := skyStyleFor()
		st.SetOption(mapstyle.OptInspect, true)
		f := mapstyle.Frame{Model: m, Tier: mapmodel.TierUnicode}
		drawStyle(st, m, 160, 48, 0, mapmodel.TierUnicode, false)
		s := st.sv.sceneFor(m)
		n := len(s.targets)
		if n < 10 {
			t.Fatalf("%s: only %d targets", age, n)
		}
		seen := map[mapmodel.Pt]bool{}
		for i := 0; i < n; i++ {
			st.HandleKey(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone), f)
			seen[st.sv.cur] = true
			in, ok := st.Inspect(f)
			if !ok || in.Title == "" {
				t.Errorf("%s target %v: no inspection", age, st.sv.cur)
				continue
			}
			switch c := s.at(st.sv.cur.X, st.sv.cur.Y); c.k {
			case skUnit, skCiv, skWonder, skMark, skHub, skCore:
				if in.Command == "" || !words[strings.Fields(in.Command)[0]] {
					t.Errorf("%s %s: command %q is not a model command", age, in.Title, in.Command)
				}
			}
		}
		if len(seen) != n {
			t.Errorf("%s: tab visited %d of %d targets", age, len(seen), n)
		}
		txt := capture.Text(drawStyle(st, m, 160, 48, 0, mapmodel.TierUnicode, false))
		if in, _ := st.Inspect(f); in.Command != "" && !strings.Contains(txt, "type: "+in.Command) {
			t.Errorf("%s: the inspector line is missing %q", age, in.Command)
		}
		st.HandleKey(tcell.NewEventKey(tcell.KeyHome, 0, tcell.ModNone), f)
		if st.sv.cur != s.hub {
			t.Errorf("%s: Home put the cursor at %v, not the hub %v", age, st.sv.cur, s.hub)
		}
		if in, _ := st.Inspect(f); in.Command != m.SquareCommand() {
			t.Errorf("%s: the hub's command is %q, want %q", age, in.Command, m.SquareCommand())
		}
		for _, r := range "hjklzxq" {
			if st.HandleKey(tcell.NewEventKey(tcell.KeyRune, r, 0), f) {
				t.Errorf("%s: %q was used; letters belong to the prompt", age, r)
			}
		}
		st.SetOption(mapstyle.OptInspect, false)
		if _, ok := st.Inspect(f); ok {
			t.Errorf("%s: Inspect with the cursor hidden reported ok", age)
		}
	}
}

// unitCells is how many plane cells a model's buildings fill: units, or in
// the mandala the marks and the core's light.
func unitCells(s *skyScene) int {
	n := 0
	for _, c := range s.cells {
		if c.k == skUnit || c.k == skMark || c.k == skCore || c.soft && c.bg == mapmodel.InkAccent && s.sky == mapmodel.SkyMandala {
			n++
		}
	}
	return n
}

// TestSkyGrowsAndHolds: building more adds units to the scene, and every
// unit already drawn stays in its cell, as the same lineage.
func TestSkyGrowsAndHolds(t *testing.T) {
	for _, age := range skyAges {
		st := fixture.State(fixture.Options{Age: age, Seed: 4})
		m0 := testBuilder.Build(&st, nil)
		s0 := newSkyScene(m0, newSkyBase(m0, m0.Sky()))
		for step := 1; step <= 4; step++ {
			st = fixture.Grow(st, step)
		}
		m1 := testBuilder.Build(&st, nil)
		s1 := newSkyScene(m1, newSkyBase(m1, m1.Sky()))
		if a, b := unitCells(s0), unitCells(s1); b <= a {
			t.Errorf("%s: %d unit cells after building more, %d before", age, b, a)
		}
		if m0.Sky() == mapmodel.SkyMandala {
			continue // marks are types, laid evenly round a ring: more types respace them
		}
		moved := 0
		for i, c := range s0.cells {
			if c.k != skUnit {
				continue
			}
			d := s1.cells[i]
			if d.k != skUnit || m1.Town.Tiles[d.ref].Lineage != m0.Town.Tiles[c.ref].Lineage {
				moved++
			}
		}
		if moved > 0 {
			t.Errorf("%s: %d units moved or vanished when the town grew", age, moved)
		}
	}
}

// TestSkyVisitors: in the Space and Interstellar scenes the rare visitor
// shows during a visit, inspects as KindAlien, never takes a legend row and
// never shows in the mini view; from the Galactic Age there are no visits,
// and the alien ships are ordinary traffic with a row of their own.
func TestSkyVisitors(t *testing.T) {
	ufo := mapmodel.R(mapmodel.SymUFO, mapmodel.TierUnicode)
	for _, age := range []string{"space_age", "interstellar_age"} {
		m := modelFor(t, fixture.Options{Age: age, Seed: 7})
		sg := mapmodel.NextSighting(m.Seed, true, 0)
		st := skyStyleFor()
		st.SetOption(mapstyle.OptInspect, true)
		found := false
		for f := sg.Start + 4; f < sg.Start+sg.Frames-4 && !found; f += 5 {
			scr := drawStyle(st, m, 160, 48, f, mapmodel.TierUnicode, false)
			p, _, ok := st.sv.skyVisit()
			if !ok {
				t.Fatalf("%s frame %d: no visit inside one", age, f)
			}
			cx, cy, in := st.sv.sg.cellOf(p)
			if !in {
				continue
			}
			if r, _, _, _ := scr.GetContent(cx+2, cy+1); r != ufo {
				continue
			}
			found = true
			st.sv.cur = p
			if ins, _ := st.Inspect(mapstyle.Frame{Model: m, Anim: f}); ins.Kind != mapstyle.KindAlien {
				t.Errorf("%s: inspecting the saucer says %+v", age, ins)
			}
			for id, e := range st.sv.seen {
				if e.on && e.r == ufo {
					t.Errorf("%s: legend row %d uses the saucer's glyph", age, id)
				}
			}
			if screenHas(drawStyle(skyStyleFor(), m, 40, 15, f, mapmodel.TierUnicode, true), []rune{ufo}) {
				t.Errorf("%s: the saucer shows in the mini view", age)
			}
		}
		if !found {
			t.Errorf("%s: the visit at frame %d never showed", age, sg.Start)
		}
		if screenHas(drawStyle(st, m, 160, 48, sg.Start-1, mapmodel.TierUnicode, false), []rune{ufo}) {
			t.Errorf("%s: a saucer before its visit", age)
		}
	}
	m := modelFor(t, fixture.Options{Age: "galactic_age", Seed: 7, Scale: 2})
	if _, ok := m.SightingAt(mapmodel.NextSighting(m.Seed, true, 0).Start + 10); ok {
		t.Error("galactic_age: the model still schedules visits")
	}
	st := skyStyleFor()
	seen := false
	for f := 0; f < 4000 && !seen; f += 9 {
		drawStyle(st, m, 160, 48, f, mapmodel.TierUnicode, false)
		seen = st.sv.seen[slMover+skyLgID(smAlien)].on
	}
	if !seen {
		t.Error("galactic_age: no alien ship ever took its legend row")
	}
}

// TestSkyNoLaterAge: a scene shows nothing from a later scene: no traffic
// before its age, and none of a later scene's signature cells.
func TestSkyNoLaterAge(t *testing.T) {
	for ai, age := range skyAges {
		m := modelFor(t, fixture.Options{Age: age, Seed: 6, Scale: 2})
		s := newSkyScene(m, newSkyBase(m, m.Sky()))
		for _, mv := range s.movers {
			if !mapmodel.Introduced(skyMoverKinds[mv.kind], m.AgeIdx) {
				t.Errorf("%s: a %s before its age", age, mv.info.Key)
			}
		}
		sigs := map[skyKind]string{skField: "the gate's field", skCloud: "probability clouds", skEcho: "time echoes",
			skRing: "the mandala", skSpiral: "the galaxy", skColony: "colony worlds", skPlanet: "the homeworld"}
		home := map[skyKind]mapmodel.SkyScene{skField: mapmodel.SkyDeep, skColony: mapmodel.SkyDeep,
			skSpiral: mapmodel.SkyGalaxy, skCloud: mapmodel.SkyQuantum, skEcho: mapmodel.SkyQuantum,
			skRing: mapmodel.SkyMandala, skPlanet: mapmodel.SkyOrbit}
		for _, c := range s.cells {
			if h, ok := home[c.k]; ok && h > m.Sky() {
				t.Errorf("%s (scene %d) shows %s", age, ai, sigs[c.k])
				delete(home, c.k)
			}
		}
	}
}

// TestSkyPerformance: a sky frame and a scene build stay well inside the
// map's 125 ms redraw.
func TestSkyPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("timing")
	}
	for _, age := range skyAges {
		m, _ := grown(t, fixture.Options{Age: age, Seed: 7, Harbinger: true, Scale: 2})
		st := skyStyleFor()
		scr := capture.NewScreen(200, 60)
		r := mapstyle.Rect{W: 200, H: 60}
		st.Draw(scr, r, mapstyle.Frame{Model: m})
		const n = 20
		start := time.Now()
		for i := 0; i < n; i++ {
			st.Draw(scr, r, mapstyle.Frame{Model: m, Anim: i, Tier: mapmodel.TierUnicode})
		}
		per := time.Since(start) / n
		start = time.Now()
		for i := 0; i < 5; i++ {
			newSkyScene(m, newSkyBase(m, m.Sky()))
		}
		build := time.Since(start) / 5
		t.Logf("%s: a 200x60 frame %v, a scene build %v", age, per, build)
		if per > 25*time.Millisecond {
			t.Errorf("%s: a frame takes %v", age, per)
		}
		if build > 60*time.Millisecond {
			t.Errorf("%s: a scene build takes %v", age, build)
		}
	}
}

func BenchmarkSkyDraw200x60(b *testing.B) {
	for _, age := range skyAges {
		b.Run(age, func(b *testing.B) {
			m, _ := grown(b, fixture.Options{Age: age, Seed: 7, Harbinger: true})
			st := skyStyleFor()
			scr := capture.NewScreen(200, 60)
			r := mapstyle.Rect{W: 200, H: 60}
			for i := 0; i < b.N; i++ {
				st.Draw(scr, r, mapstyle.Frame{Model: m, Anim: i, Tier: mapmodel.TierUnicode})
			}
		})
	}
}

func BenchmarkSkyScene(b *testing.B) {
	for _, age := range skyAges {
		b.Run(age, func(b *testing.B) {
			m, _ := grown(b, fixture.Options{Age: age, Seed: 7, Harbinger: true})
			for i := 0; i < b.N; i++ {
				newSkyScene(m, newSkyBase(m, m.Sky()))
			}
		})
	}
}

// TestSkyGateRises: the warp gate is the Interstellar Age's construction
// site: engineering raises at most three quarters of its ring, the Warp
// Nexus's banked share fills the keystone, and only the finished wonder
// closes the ring and opens its field.
func TestSkyGateRises(t *testing.T) {
	gateState := func(s *skyScene) (raised, scaffold, field int) {
		for _, p := range s.b.gate {
			switch s.at(p.X, p.Y).k {
			case skFrame, skUnit:
				raised++
			case skSlot:
				scaffold++
			}
		}
		for _, c := range s.cells {
			if c.k == skField {
				field++
			}
		}
		return
	}
	st := fixture.State(fixture.Options{Age: "interstellar_age", Seed: 3, Scale: 4})
	m := testBuilder.Build(&st, nil)
	s := newSkyScene(m, newSkyBase(m, m.Sky()))
	raised, scaffold, field := gateState(s)
	if scaffold == 0 || field > 0 {
		t.Errorf("without the Warp Nexus the gate is finished: %d raised, %d scaffold, %d field", raised, scaffold, field)
	}
	if raised < len(s.b.gate)/2 {
		t.Errorf("a big engineering lineage raised only %d of %d gate cells", raised, len(s.b.gate))
	}
	bs := st.Buildings["warp_nexus"]
	bs.Count = 1
	st.Buildings["warp_nexus"] = bs
	m = testBuilder.Build(&st, nil)
	s = newSkyScene(m, newSkyBase(m, m.Sky()))
	if raised, scaffold, field = gateState(s); scaffold > 0 || field == 0 || raised != len(s.b.gate) {
		t.Errorf("with the Warp Nexus: %d raised, %d scaffold, %d field", raised, scaffold, field)
	}
}
