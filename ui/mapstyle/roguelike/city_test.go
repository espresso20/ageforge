package roguelike

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/mapmodel/fixture"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
)

var earthAges = []string{"modern_age", "information_age", "digital_age", "cyberpunk_age", "fusion_age"}

func lookOf(t *testing.T, age string) mapmodel.CityLook {
	t.Helper()
	l, ok := mapmodel.CityLookAt(ageIndex(t, age))
	if !ok {
		t.Fatalf("%s has no Earth-arc look", age)
	}
	return l
}

// greenCells counts the cells of the greenery window whose land the city
// keeps green (forest and grass it has not built over, parks, gardens,
// yards and the reserve), over cityLand's measure (the land, but for the
// plan's streets and square, the wonder plots and the railway): a cell
// under a line laid over the city is not green, nor is anything in the
// megacity. A building hides its cell's land in every age alike.
func greenCells(s *scene, cl *cityLand) int {
	n := 0
	w := s.w
	for y := cl.win[1]; y <= cl.win[3]; y++ {
		for x := cl.win[0]; x <= cl.win[2]; x++ {
			i := y*w.W + x
			if !cl.measure[i] || !s.land(i) || cl.ov[i] != 0 {
				continue
			}
			switch u := cl.use[i]; {
			case u.green():
				n++
			case u == uNone && !cl.megacity() && (w.T[i] == mapmodel.TGrass || w.T[i] == mapmodel.TForest):
				n++
			}
		}
	}
	return n
}

// TestCityGreeneryFade holds the roguelike to the greenery rule
// (mapmodel.Greenery): over the settlement view round the square, the
// Information Age keeps half the green the Modern Age has (parks and rooftop
// gardens), the Digital Age only its walled reserve's 15%, and the megacity
// none at all, not a park, not a field drawn green, not a green canal.
func TestCityGreeneryFade(t *testing.T) {
	modern := lookOf(t, "modern_age")
	for _, seed := range []int64{2, 7, 11, 23} {
		for _, age := range earthAges {
			m := modelFor(t, fixture.Options{Age: age, Seed: seed})
			s := newScene(m)
			if s.city == nil {
				t.Fatalf("%s: no city", age)
			}
			base := greenCells(s, s.cityFor(modern))
			got := greenCells(s, s.city)
			want := mapmodel.Greenery(m.AgeIdx)
			if base < 200 {
				t.Fatalf("seed %d %s: only %d green cells to fade", seed, age, base)
			}
			ratio := float64(got) / float64(base)
			if math.Abs(ratio-want) > 0.03 {
				t.Errorf("seed %d %s keeps %.3f of the Modern green (%d of %d), want %.2f", seed, age, ratio, got, base, want)
			}
		}
	}
	// the megacity's palette has no green left in it: the farms are vats,
	// the water is toxic or cooling, and nothing is drawn in a park's ink
	for _, age := range []string{"cyberpunk_age", "fusion_age"} {
		m := modelFor(t, fixture.Options{Age: age, Seed: 7})
		v := newView()
		draw(v, m, 160, 48, 0, mapmodel.TierUnicode, false)
		for _, c := range []mapmodel.Class{mapmodel.CFlora, mapmodel.CWater, mapmodel.CGround} {
			if isGreen(v.pal.Fg[c]) {
				t.Errorf("%s: class %d is still green", age, c)
			}
		}
	}
}

// isGreen reports a colour that reads as greenery: green ahead of red and
// blue by a margin.
func isGreen(c tcell.Color) bool {
	r, g, b := c.RGB()
	return g > r+24 && g > b+12
}

// TestCitySteady: the same save lays the same city, frame after frame and
// view after view, and growing the town (inside the same wall ring) changes
// only the cells the new buildings take.
func TestCitySteady(t *testing.T) {
	for _, age := range earthAges {
		st := fixture.State(fixture.Options{Age: age, Seed: 5})
		a, b := newScene(testBuilder.Build(&st, nil)), newScene(testBuilder.Build(&st, nil))
		for i := range a.city.use {
			if a.city.use[i] != b.city.use[i] || a.city.aux[i] != b.city.aux[i] || a.city.ov[i] != b.city.ov[i] {
				t.Fatalf("%s: the same save laid cell %d two ways", age, i)
			}
		}
		g := st
		for step := 1; step <= 3; step++ {
			g = fixture.Grow(g, step)
			c := newScene(testBuilder.Build(&g, nil))
			if c.wallR != a.wallR {
				break // the ring stepped out: the city follows it
			}
			moved := 0
			for i := range a.city.use {
				if a.land(i) && c.land(i) && a.city.use[i] != c.city.use[i] {
					moved++
				}
			}
			if moved > 0 {
				t.Errorf("%s: growing the town %d steps changed %d cells of the city's land", age, step, moved)
			}
		}
	}
}

// TestCityLooksDistinct: every Earth-arc age draws its own signature
// structure and lists it in the legend, and inspecting it says what it is.
func TestCityLooksDistinct(t *testing.T) {
	for _, c := range []struct {
		age   string
		feats []mapmodel.CityFeature
	}{
		{"modern_age", []mapmodel.CityFeature{mapmodel.FeatSuburbs, mapmodel.FeatHighway, mapmodel.FeatGlassTower}},
		{"information_age", []mapmodel.CityFeature{mapmodel.FeatServerFarm, mapmodel.FeatDish, mapmodel.FeatFiber, mapmodel.FeatGarden}},
		{"digital_age", []mapmodel.CityFeature{mapmodel.FeatReserve, mapmodel.FeatLandfill, mapmodel.FeatDataGlow}},
		{"cyberpunk_age", []mapmodel.CityFeature{mapmodel.FeatMegablock, mapmodel.FeatCorpTower, mapmodel.FeatArcology,
			mapmodel.FeatNeonSign, mapmodel.FeatSkyRail}},
		{"fusion_age", []mapmodel.CityFeature{mapmodel.FeatReactor, mapmodel.FeatConduit, mapmodel.FeatMaglevLine,
			mapmodel.FeatTether, mapmodel.FeatLaunchTower}},
	} {
		m := modelFor(t, fixture.Options{Age: c.age, Seed: 7})
		v := newView()
		v.zoom = zRegion // the whole world, so every feature is somewhere in the frame
		draw(v, m, 200, 60, 0, mapmodel.TierUnicode, false)
		v2 := newView()
		draw(v2, m, 200, 60, 0, mapmodel.TierUnicode, false)
		s := v2.sceneFor(m)
		for _, f := range c.feats {
			if !v.seen[lgCity+lgID(f)].on && !v2.seen[lgCity+lgID(f)].on {
				t.Errorf("%s: the legend never lists the %s", c.age, f.Info().Name)
			}
			p, ok := cityCellOf(s, f)
			if !ok {
				t.Errorf("%s: no %s in the city", c.age, f.Info().Name)
				continue
			}
			v2.cur, v2.inspect = p, true
			in, _ := v2.Inspect(mapstyle.Frame{Model: m, Tier: mapmodel.TierUnicode})
			if in.Title == "" || f != mapmodel.FeatCorpTower && in.Title != f.Info().Title ||
				f == mapmodel.FeatCorpTower && !strings.HasSuffix(in.Title, " tower") {
				t.Errorf("%s: inspecting the %s at %v says %+v", c.age, f.Info().Name, p, in)
			}
		}
	}
}

// cityCellOf finds a cell that shows feature f: a line of its kind, or a
// cell of its land, where nothing drawn over it hides it.
func cityCellOf(s *scene, f mapmodel.CityFeature) (mapmodel.Pt, bool) {
	cl := s.city
	w := s.w
	for i := range cl.use {
		x, y := i%w.W, i/w.W
		if rdist(x, y, w.CX, w.CY) > 60 || y < 2 || moverOn(s, pt(x, y)) {
			continue
		}
		if ln, ok := cl.lineAt(s, i); ok {
			if ln.feat == f {
				return pt(x, y), true
			}
			continue
		}
		if s.cells[i].k != kNone || s.cells[i].rail {
			continue
		}
		if useFeature[cl.use[i]] == f && (f != mapmodel.FeatDataGlow || cl.use[i] == uDataHall) {
			return pt(x, y), true
		}
	}
	return mapmodel.Pt{}, false
}

// moverOn reports a mover on p at frame 0.
func moverOn(s *scene, p mapmodel.Pt) bool {
	for i := range s.movers {
		if s.movers[i].at(0, p) {
			return true
		}
	}
	return false
}

// TestCityNoSpoilers walks the Atomic Age through the Space Age: a frame
// lists, and the cursor names, only the city features its age or an earlier
// one has built, and none that gave way before it. The owner's rule: the
// map never shows what an age the player has not reached will bring.
func TestCityNoSpoilers(t *testing.T) {
	for _, age := range []string{"atomic_age", "modern_age", "information_age", "digital_age", "cyberpunk_age", "fusion_age", "space_age"} {
		ai := ageIndex(t, age)
		m := modelFor(t, fixture.Options{Age: age, Seed: 3, Routes: 3})
		names := map[string]bool{}
		for _, z := range []int{zRegion, zSettlement, zDistrict} {
			v := newView()
			v.zoom = z
			draw(v, m, 200, 60, 7, mapmodel.TierUnicode, false)
			for f := mapmodel.CityFeature(1); f < mapmodel.NumCityFeatures; f++ {
				if v.seen[lgCity+lgID(f)].on && !f.Info().In(ai) {
					t.Errorf("%s zoom %d: the legend lists a %s", age, z, f.Info().Name)
				}
			}
			s := v.sceneFor(m)
			if s.city == nil {
				continue
			}
			for i := range s.city.use {
				if i%7 != 0 { // a spread of cells is plenty
					continue
				}
				v.cur = pt(i%s.w.W, i/s.w.W)
				if in, ok := v.describeCity(s, v.cur, s.cells[i].k); ok {
					names[in.Title] = true
				}
			}
		}
		for f := mapmodel.CityFeature(1); f < mapmodel.NumCityFeatures; f++ {
			if names[f.Info().Title] && !f.Info().In(ai) {
				t.Errorf("%s: the cursor names a %s", age, f.Info().Name)
			}
		}
	}
}

// TestCityTraffic: each Earth-arc age's signature mover is about and on
// screen at the default camera; the megacity's sky trains ride every
// elevated line (three to five, at least two trains a line); the tether
// carries its climbers; and the Cyberpunk Age is the busiest town in the
// game.
func TestCityTraffic(t *testing.T) {
	busiest, most := "", 0
	for _, age := range config.AgeOrder() {
		m := modelFor(t, fixture.Options{Age: age, Seed: 7, Routes: 3})
		s := newScene(m)
		if len(s.movers) > most {
			busiest, most = age, len(s.movers)
		}
		look, ok := mapmodel.CityLookAt(m.AgeIdx)
		if !ok {
			continue
		}
		count := map[mapmodel.Mover]int{}
		for _, mv := range s.movers {
			count[mv.k]++
		}
		if count[look.Mover] == 0 {
			t.Errorf("%s: no %s, its signature mover", age, look.Mover.Info().Name)
		}
		lines := 0
		for _, ln := range s.city.lines {
			if ln.feat == mapmodel.FeatSkyRail || ln.feat == mapmodel.FeatMaglevLine {
				lines++
			}
		}
		switch age {
		case "cyberpunk_age":
			if lines < 3 || lines > 5 || count[mapmodel.MoverSkyTrain] < 2*lines {
				t.Errorf("%s: %d sky rails carrying %d sky trains", age, lines, count[mapmodel.MoverSkyTrain])
			}
			if count[mapmodel.MoverCrowd] < 8 || count[mapmodel.MoverHovercar] < 6 || count[mapmodel.MoverDrone] < 8 {
				t.Errorf("%s: not busy enough: %d crowds, %d hovercars, %d drones", age,
					count[mapmodel.MoverCrowd], count[mapmodel.MoverHovercar], count[mapmodel.MoverDrone])
			}
		case "fusion_age":
			if lines < 3 || count[mapmodel.MoverMaglev] < lines || count[mapmodel.MoverClimber] < 2 || count[mapmodel.MoverSkyTrain] > 0 {
				t.Errorf("%s: %d maglev lines, %d maglevs, %d climbers, %d sky trains", age, lines,
					count[mapmodel.MoverMaglev], count[mapmodel.MoverClimber], count[mapmodel.MoverSkyTrain])
			}
		default:
			if lines != 0 {
				t.Errorf("%s: %d elevated lines before the megacity", age, lines)
			}
		}
	}
	if busiest != "cyberpunk_age" {
		t.Errorf("the busiest town is the %s (%d movers), not the Cyberpunk Age's", busiest, most)
	}
}

// TestCityTiersAndThemes: every Earth-arc age draws one cell per rune in
// every glyph tier (nothing but ASCII in the ASCII tier, no '?' fallback),
// in every view including the mini map, and on every theme no glyph is
// drawn in its own background colour.
func TestCityTiersAndThemes(t *testing.T) {
	orig := theme.Active().Key
	t.Cleanup(func() { _ = theme.SetActive(orig) })
	var models []*mapmodel.Model
	for i, age := range earthAges {
		m, _ := grown(t, fixture.Options{Age: age, Seed: 4 + int64(i), Routes: 2, Harbinger: true, Wars: 1})
		models = append(models, m)
	}
	for _, m := range models {
		for _, tier := range []mapmodel.GlyphTier{mapmodel.TierASCII, mapmodel.TierUnicode, mapmodel.TierNerd} {
			for z := zRegion; z <= zDistrict; z++ {
				for _, compact := range []bool{false, true} {
					v := newView()
					v.zoom = z
					w, h := 160, 48
					if compact {
						w, h = 40, 15
					}
					s := draw(v, m, w, h, 11, tier, compact)
					cells, sw, _ := s.GetContents()
					for i, c := range cells {
						if len(c.Runes) == 0 || c.Runes[0] == sentinel || i/sw == h {
							continue
						}
						r := c.Runes[0]
						if r == '?' || tier == mapmodel.TierASCII && r >= 0x80 {
							t.Fatalf("%s %s zoom %d compact %v: rune %q at %d,%d", m.Age, tier, z, compact, r, i%sw, i/sw)
						}
					}
				}
			}
		}
	}
	for _, th := range theme.All() {
		if err := theme.SetActive(th.Key); err != nil {
			t.Fatal(err)
		}
		for _, m := range models {
			for _, z := range []int{zRegion, zSettlement} {
				v := newView()
				v.zoom = z
				s := draw(v, m, 140, 44, 3, mapmodel.TierUnicode, false)
				cells, sw, _ := s.GetContents()
				for i, c := range cells {
					if len(c.Runes) == 0 || c.Runes[0] == ' ' || c.Runes[0] == sentinel {
						continue
					}
					if fg, bg, _ := c.Style.Decompose(); fg == bg {
						t.Errorf("theme %s, %s zoom %d: %q at %d,%d has fg == bg", th.Key, m.Age, z, c.Runes[0], i%sw, i/sw)
						break
					}
				}
			}
		}
	}
}

// TestCityPerf holds the Earth arc to the map's frame budget (it redraws
// every 125 ms): laying a megacity scene and drawing its frames stay cheap.
func TestCityPerf(t *testing.T) {
	if testing.Short() || raceOn {
		t.Skip("timing")
	}
	for _, age := range []string{"cyberpunk_age", "fusion_age", "information_age"} {
		m, _ := grown(t, fixture.Options{Age: age, Seed: 7, Routes: 4, Harbinger: true, Scale: 2})
		start := time.Now()
		const scenes = 5
		for i := 0; i < scenes; i++ {
			newScene(m)
		}
		lay := time.Since(start) / scenes
		v := newView()
		scr := draw(v, m, 200, 60, 0, mapmodel.TierUnicode, false)
		const n = 30
		start = time.Now()
		for i := 0; i < n; i++ {
			v.Draw(scr, mapstyle.Rect{W: 200, H: 60}, mapstyle.Frame{Model: m, Anim: i, Tier: mapmodel.TierUnicode})
		}
		frame := time.Since(start) / n
		t.Logf("%s: scene %v, 200x60 frame %v", age, lay, frame)
		if lay > 80*time.Millisecond {
			t.Errorf("%s: laying the scene takes %v", age, lay)
		}
		if frame > 20*time.Millisecond {
			t.Errorf("%s: a 200x60 frame takes %v", age, frame)
		}
	}
}
