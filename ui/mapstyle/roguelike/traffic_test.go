package roguelike

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/uniseg"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/mapmodel/fixture"
	"github.com/espresso20/ageforge/ui/mapstyle"
	"github.com/espresso20/ageforge/ui/mapstyle/capture"
)

func ageIndex(t testing.TB, key string) int {
	t.Helper()
	for i, k := range config.AgeOrder() {
		if k == key {
			return i
		}
	}
	t.Fatalf("no age %s", key)
	return -1
}

// TestTrafficNoSpoilers walks every age: a town plans, draws and lists in
// its legend only movers that its age or an earlier one introduced, and the
// railway comes with the Industrial Age. The owner's rule: the map never
// shows what an age the player has not reached will bring.
func TestTrafficNoSpoilers(t *testing.T) {
	industrial, cosmic := ageIndex(t, "industrial_age"), ageIndex(t, "interstellar_age")
	railed := false
	for ai, age := range config.AgeOrder() {
		for oi, o := range []fixture.Options{{Age: age, Seed: 2, Routes: 3}, {Age: age, Seed: 11, Scale: 2, Wars: 1}} {
			m := modelFor(t, o)
			// the settlement over a few frames, the district once, and the
			// region where its plate draws movers (the cosmic orbit)
			views := []struct{ zoom, anim int }{{zSettlement, 0}, {zSettlement, 1777}}
			if oi == 1 {
				views = []struct{ zoom, anim int }{{zDistrict, 333}}
			}
			if ai >= cosmic {
				views = append(views, struct{ zoom, anim int }{zRegion, 4001})
			}
			var s *scene
			for _, vw := range views {
				v := newView()
				v.zoom = vw.zoom
				draw(v, m, 160, 48, vw.anim, mapmodel.TierUnicode, false)
				for k := mapmodel.Mover(1); k < mapmodel.NumMovers; k++ {
					if v.seen[lgMover+lgID(k)].on && !mapmodel.Introduced(k, ai) {
						t.Errorf("%s zoom %d frame %d: the legend lists a %s", age, vw.zoom, vw.anim, k.Info().Name)
					}
				}
				s = v.sceneFor(m)
			}
			for _, mv := range s.movers {
				if !mapmodel.Introduced(mv.k, ai) || !mv.info.In(ai) {
					t.Errorf("%s: a %s is about", age, mv.info.Name)
				}
			}
			if ai < industrial && len(s.rail) > 0 {
				t.Errorf("%s: a railway before the Industrial Age", age)
			}
			railed = railed || len(s.rail) > 0
		}
	}
	if !railed {
		t.Error("no town ever laid a railway")
	}
}

// findShown finds a frame in [from, limit) where the mover's cell number j0
// is on screen and drawn (not on a building, not a boat under a bridge, no
// other mover, worker, idle hand or scout on it), and that cell.
func findShown(v *view, mv *mover, from, limit int, j0 int) (int, mapmodel.Pt, bool) {
	s := v.sc
	for f := from; f < limit; f++ {
		var hit mapmodel.Pt
		ok := false
		mv.cells(f, func(p mapmodel.Pt, j, _ int) {
			if j != j0 || ok {
				return
			}
			c := s.at(p.X, p.Y)
			if _, _, in := v.g.cellOf(p); !in || s.seen(p) < 2 || c.k == kTile || c.k == kWonder || c.k == kSite || c.k == kCentre {
				return
			}
			if mv.info.Way == mapmodel.WayWater && (c.k != kNone || c.rail) {
				return
			}
			hit, ok = p, true
		})
		if !ok || !v.shown(mv) {
			continue
		}
		alone := true
		for i := range s.movers {
			if o := &s.movers[i]; o != mv && o.at(f, hit) {
				alone = false
			}
		}
		for i, path := range s.walks {
			if path[walkAt(i, len(path), f)] == hit {
				alone = false
			}
		}
		if alone && (!s.hasHb || hit != s.harb) && !has(s.idle, hit) && !has(s.scout, hit) && !has(s.hazard, hit) {
			return f, hit, true
		}
	}
	return 0, mapmodel.Pt{}, false
}

// TestTrafficPerEra: every era's town carries that era's movers, drawn with
// their glyphs where the roster says, and inspecting one says what it is.
func TestTrafficPerEra(t *testing.T) {
	for _, c := range []struct {
		age  string
		want []mapmodel.Mover
	}{
		{"stone_age", []mapmodel.Mover{mapmodel.MoverHunter}},
		{"iron_age", []mapmodel.Mover{mapmodel.MoverOxCart, mapmodel.MoverRider, mapmodel.MoverRowboat}},
		{"industrial_age", []mapmodel.Mover{mapmodel.MoverWagon, mapmodel.MoverSailShip, mapmodel.MoverSteamTrain}},
		{"electric_age", []mapmodel.Mover{mapmodel.MoverTram, mapmodel.MoverEarlyCar, mapmodel.MoverSteamship}},
		{"digital_age", []mapmodel.Mover{mapmodel.MoverCar, mapmodel.MoverTruck, mapmodel.MoverPlane, mapmodel.MoverBoxShip, mapmodel.MoverTrain}},
		{"cyberpunk_age", []mapmodel.Mover{mapmodel.MoverMaglev, mapmodel.MoverDrone, mapmodel.MoverHovercar}},
		{"space_age", []mapmodel.Mover{mapmodel.MoverShuttle}},
		{"galactic_age", []mapmodel.Mover{mapmodel.MoverSatellite, mapmodel.MoverHabitat}},
	} {
		m := modelFor(t, fixture.Options{Age: c.age, Seed: 7})
		v := newView()
		draw(v, m, 160, 48, 0, mapmodel.TierUnicode, false)
		s := v.sceneFor(m)
		for _, k := range c.want {
			var mv *mover
			for i := range s.movers {
				if s.movers[i].k == k && mv == nil {
					mv = &s.movers[i]
				}
			}
			if mv == nil {
				t.Errorf("%s: no %s", c.age, k.Info().Name)
				continue
			}
			j := 0
			want := mapmodel.R(mv.info.Sym, mapmodel.TierUnicode)
			if mv.cars > 1 {
				j = 1 // a car behind the engine
			}
			// the first frames it shows on screen, until its glyph is there
			// (a caravan or a label drawn later may sit on it now and then)
			f, p, shown, from := 0, mapmodel.Pt{}, false, 0
			for try := 0; try < 12 && !shown; try++ {
				var ok bool
				if f, p, ok = findShown(v, mv, from, 6000, j); !ok {
					break
				}
				scr := draw(v, m, 160, 48, f, mapmodel.TierUnicode, false)
				cx, cy, _ := v.g.cellOf(p)
				r, _, _, _ := scr.GetContent(cx+2, cy+1)
				shown, from = r == want, f+1
			}
			if !shown {
				t.Errorf("%s: the %s never shows its glyph %q at the default camera", c.age, mv.info.Name, want)
				continue
			}
			v.cur = p
			in, ok := v.Inspect(mapstyle.Frame{Model: m, Anim: f, Tier: mapmodel.TierUnicode})
			if !ok || in.Title != mv.info.Title || len(in.Lines) != 1 || in.Lines[0] != mv.info.Line(mv.n) || in.Kind != mapstyle.KindMover {
				t.Errorf("%s: inspecting the %s says %+v", c.age, mv.info.Name, in)
			}
		}
	}
}

// TestRailInspect: the railway says what runs on it.
func TestRailInspect(t *testing.T) {
	for _, c := range []struct{ age, title, line string }{
		{"industrial_age", "Railway", "steam trains run through "},
		{"digital_age", "Railway", "freight trains run through "},
		{"cyberpunk_age", "Maglev line", "maglevs run through "},
	} {
		m := modelFor(t, fixture.Options{Age: c.age, Seed: 7})
		v := newView()
		draw(v, m, 160, 48, 0, mapmodel.TierUnicode, false)
		s := v.sceneFor(m)
		if len(s.rail) == 0 {
			t.Fatalf("%s: no railway", c.age)
		}
		for _, p := range s.rail {
			if _, busy := v.moverAt(p); busy || s.seen(p) < 2 {
				continue
			}
			v.cur = p
			in, _ := v.Inspect(mapstyle.Frame{Model: m, Anim: 0})
			if in.Title != c.title || len(in.Lines) == 0 || in.Lines[0] != c.line+s.w.Name {
				t.Errorf("%s: the railway inspects as %+v", c.age, in)
			}
			break
		}
	}
}

// TestTrafficNeverCoversBuildings: with the traffic on, every building,
// wonder, civ and the town square draws exactly as it does without it.
func TestTrafficNeverCoversBuildings(t *testing.T) {
	for _, age := range []string{"iron_age", "industrial_age", "digital_age", "galactic_age"} {
		m := modelFor(t, fixture.Options{Age: age, Seed: 4, Scale: 2})
		for _, z := range []int{zSettlement, zDistrict} {
			for _, anim := range []int{10, 2222} {
				v := newView()
				v.zoom = z
				v.SetOption(mapstyle.OptInspect, false)
				a := capture.Text(draw(v, m, 160, 48, anim, mapmodel.TierUnicode, false))
				s := v.sceneFor(m)
				saved := s.movers
				s.movers = nil
				b := draw(v, m, 160, 48, anim, mapmodel.TierUnicode, false)
				s.movers = saved
				rows := strings.Split(a, "\n")
				for y := 0; y < v.g.h; y++ {
					for x := 0; x < v.g.w; x++ {
						p := mapmodel.Pt{X: v.g.vx + x/v.g.cellW, Y: v.g.vy + y}
						if k := s.at(p.X, p.Y).k; k != kTile && k != kWonder && k != kSite && k != kCentre {
							continue
						}
						want, _, _, _ := b.GetContent(v.g.x+x+2, v.g.y+y+1)
						if got := []rune(rows[v.g.y+y+1]); v.g.x+x+2 < len(got) && got[v.g.x+x+2] != want {
							t.Fatalf("%s zoom %d frame %d: traffic drew %q over %q at %v", age, z, anim, got[v.g.x+x+2], want, p)
						}
					}
				}
			}
		}
	}
}

// TestLayoutUnchangedByTraffic: the railway and the movers join the town
// without moving anything: a scene laid with them has the same cells, walls,
// streets, walkers, targets and fog as one laid without, and the rail only
// marks cells that carry no building.
func TestLayoutUnchangedByTraffic(t *testing.T) {
	for _, age := range []string{"iron_age", "industrial_age", "digital_age", "galactic_age"} {
		for _, seed := range []int64{1, 9} {
			m := modelFor(t, fixture.Options{Age: age, Seed: seed, Wars: 1, Routes: 2, Harbinger: true, Catastrophe: true})
			if len(m.Factions) >= 127 {
				t.Fatalf("%d civs do not fit a cell's civ byte", len(m.Factions))
			}
			a, b := layScene(m, false), newScene(m)
			for i := range a.cells {
				ca, cb := a.cells[i], b.cells[i]
				if cb.rail && (cb.k == kTile || cb.k == kWonder || cb.k == kSite || cb.k == kCentre) {
					t.Errorf("%s/%d: the railway runs over a building at %d", age, seed, i)
				}
				cb.rail = false
				if ca != cb {
					t.Fatalf("%s/%d: cell %d,%d changed: %+v -> %+v", age, seed, i%a.w.W, i/a.w.W, ca, cb)
				}
				if a.vis[i] != b.vis[i] {
					t.Fatalf("%s/%d: the fog moved at %d", age, seed, i)
				}
			}
			same := func(what string, x, y []mapmodel.Pt) {
				if len(x) != len(y) {
					t.Errorf("%s/%d: %s changed", age, seed, what)
					return
				}
				for i := range x {
					if x[i] != y[i] {
						t.Errorf("%s/%d: %s changed", age, seed, what)
						return
					}
				}
			}
			same("idle hands", a.idle, b.idle)
			same("targets", a.targets, b.targets)
			same("hazard ring", a.hazard, b.hazard)
			same("raiders", a.raid, b.raid)
			if len(a.walks) != len(b.walks) {
				t.Errorf("%s/%d: walkers changed", age, seed)
			}
			for i := range a.walks {
				same("a walk", a.walks[i], b.walks[i])
			}
			if a.harb != b.harb || a.wallR != b.wallR || a.wallC != b.wallC || a.x0 != b.x0 || a.y1 != b.y1 {
				t.Errorf("%s/%d: the harbinger, the wall ring or the bounds moved", age, seed)
			}
			n := 0
			for i := range b.cells {
				if b.cells[i].rail {
					n++
				}
			}
			if n != len(b.rail) {
				t.Errorf("%s/%d: %d rail cells for a line of %d", age, seed, n, len(b.rail))
			}
		}
	}
}

// TestRailSteady: the same save always lays the same railway: as the town
// grows, and through every age from the Industrial on.
func TestRailSteady(t *testing.T) {
	for _, seed := range []int64{1, 4, 13} {
		st := fixture.State(fixture.Options{Age: "industrial_age", Seed: seed})
		var first []mapmodel.Pt
		for g := 0; g < 6; g++ {
			s := newScene(testBuilder.Build(&st, nil))
			if g == 0 {
				first = s.rail
				if len(first) == 0 {
					t.Fatalf("seed %d: no railway", seed)
				}
			} else if len(s.rail) != len(first) || s.rail[0] != first[0] || s.rail[len(s.rail)-1] != first[len(first)-1] {
				t.Fatalf("seed %d grow %d: the railway moved", seed, g)
			}
			st = fixture.Grow(st, 1)
		}
		for i, age := range config.AgeOrder()[ageIndex(t, "industrial_age"):] {
			if i%3 != 0 {
				continue // every third age is plenty: the line reads only the seed
			}
			s := newScene(modelFor(t, fixture.Options{Age: age, Seed: seed, Scale: 2}))
			if len(s.rail) != len(first) || s.rail[0] != first[0] {
				t.Errorf("seed %d %s: the railway moved", seed, age)
			}
		}
	}
}

// TestTrafficLegendSteady: the traffic's legend rows hold still while the
// movers come and go: a frame lists what is about, not what happens to be
// on screen this instant.
func TestTrafficLegendSteady(t *testing.T) {
	for _, age := range []string{"iron_age", "digital_age", "space_age"} {
		m := modelFor(t, fixture.Options{Age: age, Seed: 7})
		v := newView()
		var first string
		for anim := 0; anim < 400; anim += 19 {
			labels := strings.Join(legendLabels(draw(v, m, 200, 60, anim, mapmodel.TierUnicode, false), 200, 60), "|")
			if anim == 0 {
				first = labels
				listed := false
				for _, k := range mapmodel.MoversAt(m.AgeIdx) {
					listed = listed || k != mapmodel.MoverWalker && strings.Contains(labels, k.Info().Name)
				}
				if !listed {
					t.Errorf("%s: no traffic in the legend: %s", age, labels)
				}
			} else if labels != first {
				t.Fatalf("%s frame %d: the legend changed\nwas %s\nnow %s", age, anim, first, labels)
			}
		}
	}
}

// TestTrafficTiers: the traffic and the visitor draw one cell per rune in
// every glyph tier, and nothing but ASCII in the ASCII tier.
func TestTrafficTiers(t *testing.T) {
	for _, age := range []string{"stone_age", "industrial_age", "digital_age", "cyberpunk_age", "galactic_age"} {
		m := modelFor(t, fixture.Options{Age: age, Seed: 5, Scale: 2})
		visit := mapmodel.NextSighting(m.Seed, m.AgeIdx >= m.Catalog.SpaceAge(), 0)
		frames := []int{97, visit.Start + visit.Frames/2}
		for _, tier := range []mapmodel.GlyphTier{mapmodel.TierASCII, mapmodel.TierUnicode, mapmodel.TierNerd} {
			for _, anim := range frames {
				v := newView()
				s := draw(v, m, 150, 46, anim, tier, false)
				cells, sw, _ := s.GetContents()
				for i, c := range cells {
					if len(c.Runes) == 0 || c.Runes[0] == sentinel || i/sw == 46 {
						continue
					}
					r := c.Runes[0]
					if r == '?' || tier == mapmodel.TierASCII && r >= 0x80 || uniseg.StringWidth(string(r)) != 1 {
						t.Fatalf("%s %s frame %d: rune %q at row %d", age, tier, anim, r, i/sw)
					}
				}
			}
		}
	}
}

// visitorRunes are the runes the visitor draws with in a tier.
func visitorRunes(tier mapmodel.GlyphTier) []rune {
	return []rune{mapmodel.R(mapmodel.SymUFO, tier), mapmodel.R(mapmodel.SymAlien, tier)}
}

func screenHas(s tcell.SimulationScreen, rs []rune) bool {
	cells, _, _ := s.GetContents()
	for _, c := range cells {
		for _, r := range rs {
			if len(c.Runes) > 0 && c.Runes[0] == r {
				return true
			}
		}
	}
	return false
}

// TestVisitor: from the Space Age the visitor shows when the schedule says,
// as a saucer flying over, a saucer hovering or a small figure walking, and
// at no other time; inspecting it reports mapstyle.KindAlien and a fun line;
// it never takes a legend row and never shows in the mini view. Before the
// Space Age it is a rare joke, drawn the same way. (From the Galactic Age
// the aliens are ordinary traffic and the model schedules no visits.)
func TestVisitor(t *testing.T) {
	for _, c := range []struct {
		age  string
		late bool
	}{{"space_age", true}, {"interstellar_age", true}, {"primitive_age", false}, {"medieval_age", false}} {
		m := modelFor(t, fixture.Options{Age: c.age, Seed: 7})
		if (m.AgeIdx >= m.Catalog.SpaceAge()) != c.late {
			t.Fatalf("%s: wrong side of the Space Age", c.age)
		}
		seen := map[mapmodel.SightingKind]bool{}
		for a := 0; len(seen) < 3; {
			sg := mapmodel.NextSighting(m.Seed, c.late, a)
			a = sg.Start + sg.Frames
			if seen[sg.Kind] {
				continue
			}
			seen[sg.Kind] = true
			v := newView()
			before := draw(v, m, 160, 48, sg.Start-1, mapmodel.TierUnicode, false)
			if screenHas(before, visitorRunes(mapmodel.TierUnicode)) {
				t.Errorf("%s: the visitor shows before its visit", c.age)
			}
			found := false
			for f := sg.Start + sg.Frames/2; f < sg.Start+sg.Frames && !found; f += 3 {
				scr := draw(v, m, 160, 48, f, mapmodel.TierUnicode, false)
				vi, ok := v.visitNow()
				if !ok {
					t.Fatalf("%s frame %d: no visit inside one", c.age, f)
				}
				cx, cy, in := v.g.cellOf(vi.at[1])
				if !in {
					continue
				}
				want := mapmodel.R(mapmodel.SymAlien, mapmodel.TierUnicode)
				if vi.saucer {
					want = mapmodel.R(mapmodel.SymUFO, mapmodel.TierUnicode)
				}
				if r, _, _, _ := scr.GetContent(cx+2, cy+1); r != want {
					continue // under a label this frame
				}
				found = true
				v.cur = vi.at[1]
				ins, _ := v.Inspect(mapstyle.Frame{Model: m, Anim: f})
				lines := walkerLines
				if vi.saucer {
					lines = saucerLines
				}
				if ins.Kind != mapstyle.KindAlien || len(ins.Lines) != 1 || !strings.Contains(strings.Join(lines, "|"), ins.Lines[0]) {
					t.Errorf("%s: inspecting the visitor says %+v", c.age, ins)
				}
				lg := strings.ToLower(strings.Join(legendLabels(scr, 160, 48), "|"))
				for _, w := range []string{"craft", "visitor", "alien", "ufo", "saucer"} {
					if strings.Contains(lg, w) {
						t.Errorf("%s: the legend mentions the visitor: %s", c.age, lg)
					}
				}
				for id := lgID(0); id < numLg; id++ {
					for _, r := range visitorRunes(mapmodel.TierUnicode) {
						if v.seen[id].on && v.seen[id].r == r {
							t.Errorf("%s: a legend row uses the visitor's glyph", c.age)
						}
					}
				}
				mini := draw(newView(), m, 40, 15, f, mapmodel.TierUnicode, true)
				if screenHas(mini, visitorRunes(mapmodel.TierUnicode)) {
					t.Errorf("%s: the visitor shows in the mini view", c.age)
				}
			}
			if !found {
				t.Errorf("%s: a %d visit at frame %d never showed", c.age, sg.Kind, sg.Start)
			}
			after := draw(v, m, 160, 48, sg.Start+sg.Frames, mapmodel.TierUnicode, false)
			if screenHas(after, visitorRunes(mapmodel.TierUnicode)) {
				t.Errorf("%s: the visitor outstays its visit", c.age)
			}
		}
	}
}

// TestVisitorComesWithMotionOff: with the motion setting off the animation
// frame holds at 0 and the clock runs on, so the visitor still arrives,
// stands where the middle of its visit puts it, inspects as
// mapstyle.KindAlien, and leaves when the visit ends.
func TestVisitorComesWithMotionOff(t *testing.T) {
	m := modelFor(t, fixture.Options{Age: "space_age", Seed: 7})
	for a, kinds := 0, 0; kinds < 3; kinds++ {
		sg := mapmodel.NextSighting(m.Seed, true, a)
		a = sg.Start + sg.Frames
		still := func(clock int) mapstyle.Frame {
			return mapstyle.Frame{Model: m, Clock: clock, Tier: mapmodel.TierUnicode}
		}
		v := newView()
		at := func(clock int) (visit, bool) {
			s := capture.NewScreen(164, 50)
			fillScreen(s)
			v.Draw(s, mapstyle.Rect{X: 2, Y: 1, W: 160, H: 48}, still(clock))
			return v.visitNow()
		}
		first, ok := at(sg.Start + 1)
		if !ok {
			t.Fatalf("with motion off no visitor shows inside a visit (kind %d)", sg.Kind)
		}
		last, ok := at(sg.Start + sg.Frames - 1)
		if !ok || last.at != first.at {
			t.Errorf("with motion off the visitor moved during its visit: %v, then %v", first.at, last.at)
		}
		v.cur = first.at[1]
		if ins, _ := v.Inspect(still(sg.Start + 1)); ins.Kind != mapstyle.KindAlien {
			t.Errorf("with motion off the visitor does not inspect as one: %+v", ins)
		}
		for _, clock := range []int{sg.Start - 1, sg.Start + sg.Frames} {
			if _, ok := at(clock); ok {
				t.Errorf("with motion off a visitor shows at clock %d, outside its visit", clock)
			}
		}
	}
}
