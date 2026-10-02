package all

import (
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/mapmodel/fixture"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
	"github.com/espresso20/ageforge/ui/mapstyle/capture"
)

// sky_test.go checks the sky arc across both styles: each age looks like
// itself (the disparity guard), nothing names what the player has not
// reached, the rare visitor keeps out of every legend and mini map, and a
// frame stays inside the redraw budget.

// arcAges are the sky arc's ages with the last ground age before them.
var arcAges = []string{"fusion_age", "space_age", "interstellar_age", "galactic_age", "quantum_age", "transcendent_age"}

var skyBuilder = mapmodel.NewBuilder(nil)

func arcModel(age string) *mapmodel.Model {
	st := fixture.State(fixture.Options{Age: age, Seed: 11, Scale: 1.5, Routes: 2})
	return skyBuilder.Build(&st, nil)
}

// signature is what a frame's picture is made of (not the header, legend
// or status lines every age shares): how often each glyph is drawn, and how
// often each colour is (the ink of every drawn glyph and every background
// that is not the picture's own void, on an 8x8x8 grid, with the void
// itself counted at a quarter weight). Empty space is left out of the glyph
// count: every age has plenty of it, and it is what is drawn on it that
// makes an age look like itself.
type signature struct {
	glyph  map[rune]float64
	colour map[int]float64
}

func sigOf(scr tcell.SimulationScreen, x0, y0, x1, y1 int) signature {
	s := signature{glyph: map[rune]float64{}, colour: map[int]float64{}}
	bucket := func(c tcell.Color) int {
		r, g, b := c.RGB()
		if r < 0 {
			return -1
		}
		return int(r>>5)<<6 | int(g>>5)<<3 | int(b>>5)
	}
	bgs := map[tcell.Color]int{}
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			_, _, st, _ := scr.GetContent(x, y)
			_, bg, _ := st.Decompose()
			bgs[bg]++
		}
	}
	var void tcell.Color
	most := -1
	for c, n := range bgs {
		if n > most || n == most && c < void {
			void, most = c, n
		}
	}
	glyphs, inks := 0.0, 0.0
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			r, _, st, _ := scr.GetContent(x, y)
			fg, bg, _ := st.Decompose()
			if r != ' ' {
				s.glyph[r]++
				glyphs++
				s.colour[bucket(fg)]++
				inks++
			}
			if bg != void {
				s.colour[bucket(bg)]++
				inks++
			}
		}
	}
	s.colour[1<<12+bucket(void)] += inks / 4
	inks += inks / 4
	for k := range s.glyph {
		s.glyph[k] /= max(1, glyphs)
	}
	for k := range s.colour {
		s.colour[k] /= max(1, inks)
	}
	return s
}

// tv is the total variation distance of two distributions: 0 alike, 1
// nothing in common.
func tv[K comparable](a, b map[K]float64) float64 {
	d := 0.0
	for k, v := range a {
		d += abs(v - b[k])
	}
	for k, v := range b {
		if _, ok := a[k]; !ok {
			d += v
		}
	}
	return d / 2
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

// distance is how unlike two frames are: the mean of the glyph and the
// colour distances.
func distance(a, b signature) float64 { return (tv(a.glyph, b.glyph) + tv(a.colour, b.colour)) / 2 }

// disparityFloor is how unlike two next-door ages of the sky arc must look,
// on the scale of distance (0 the same picture, 1 nothing in common). The
// same age a few frames later sits under 0.02 (things move, stars twinkle).
// When the guard was written next-door ages measured from 0.47 (the
// roguelike's Interstellar to Galactic) to 0.74 (Fusion to Space), so 0.30
// leaves room to tune an age without letting two neighbours drift into
// looking like one.
const disparityFloor = 0.30

// frameSig draws a model in a style at 160x48 and signs the picture.
func frameSig(t *testing.T, style string, m *mapmodel.Model, anim int) signature {
	t.Helper()
	v, _ := Registry().New(style)
	scr := capture.NewScreen(160, 48)
	defer scr.Fini()
	v.Draw(scr, mapstyle.Rect{W: 160, H: 48}, mapstyle.Frame{Model: m, Anim: anim, Tier: mapmodel.TierUnicode})
	scr.Show()
	return sigOf(scr, 0, 1, 132, 43)
}

// TestSkyDisparity: in both styles every age of the sky arc, and the Fusion
// Age before it, looks unlike its neighbours by at least disparityFloor,
// while the same age a moment later looks like itself.
func TestSkyDisparity(t *testing.T) {
	orig := theme.Active().Key
	t.Cleanup(func() { _ = theme.SetActive(orig) })
	if err := theme.SetActive("forge"); err != nil {
		t.Fatal(err)
	}
	models := make([]*mapmodel.Model, len(arcAges))
	for i, a := range arcAges {
		models[i] = arcModel(a)
	}
	for _, style := range Registry().Names() {
		sigs := make([]signature, len(arcAges))
		for i, m := range models {
			sigs[i] = frameSig(t, style, m, 400)
			if self := distance(sigs[i], frameSig(t, style, m, 403)); self > disparityFloor/3 {
				t.Errorf("%s %s: a frame later it is already %.2f away from itself", style, arcAges[i], self)
			}
		}
		for i := 1; i < len(arcAges); i++ {
			d := distance(sigs[i-1], sigs[i])
			t.Logf("%s: %s to %s %.2f (glyphs %.2f, colours %.2f)", style, arcAges[i-1], arcAges[i], d,
				tv(sigs[i-1].glyph, sigs[i].glyph), tv(sigs[i-1].colour, sigs[i].colour))
			if d < disparityFloor {
				t.Errorf("%s: %s and %s look too alike: %.2f, under %.2f", style, arcAges[i-1], arcAges[i], d, disparityFloor)
			}
		}
	}
}

// TestSkyNoSpoilers: the sky ages name no age, era or civ the player has
// not reached or met, in either style, on screen, on the inspect line or in
// the mini map, and name the next age once it is within reach.
func TestSkyNoSpoilers(t *testing.T) {
	for _, age := range arcAges[1:] {
		for _, ready := range []bool{false, true} {
			st := fixture.State(fixture.Options{Age: age, Seed: 5, Harbinger: true, Wars: 1, Routes: 2})
			met := 0
			for k, f := range st.Diplomacy.Factions {
				if f.Discovered && met >= 3 { // forget all but three of the civs
					f.Discovered, f.AtWar = false, false
				}
				if f.Discovered {
					met++
				}
				st.Diplomacy.Factions[k] = f
			}
			st.AgeReady = ready
			txt := mapText(t, st)
			for _, word := range unreached(st) {
				if i := strings.Index(txt, word); i >= 0 {
					lo, hi := max(0, i-60), min(len(txt), i+len(word)+60)
					t.Errorf("%s: the map shows %q:\n...%s...", age, word, txt[lo:hi])
				}
			}
			if ready && st.NextAgeName != "" && !strings.Contains(txt, st.NextAgeName) {
				t.Errorf("%s: ready to advance, but the map never names %q", age, st.NextAgeName)
			}
		}
	}
}

// TestSkyVisitorRules: in the Space and Interstellar Ages the rare visitor
// never shows in either style's mini map; from the Galactic Age the model
// schedules none, and the aliens are ordinary traffic.
func TestSkyVisitorRules(t *testing.T) {
	ufo := mapmodel.R(mapmodel.SymUFO, mapmodel.TierUnicode)
	for _, age := range []string{"space_age", "interstellar_age", "galactic_age"} {
		m := arcModel(age)
		sg := mapmodel.NextSighting(m.Seed, true, 0)
		_, visiting := m.SightingAt(sg.Start + sg.Frames/2)
		if age == "galactic_age" {
			if visiting {
				t.Error("galactic_age: the model schedules a visit")
			}
			continue
		}
		if !visiting {
			t.Fatalf("%s: no visit inside the schedule's", age)
		}
		for _, style := range Registry().Names() {
			for f := sg.Start; f < sg.Start+sg.Frames; f += 11 {
				v, _ := Registry().New(style)
				scr := capture.NewScreen(40, 15)
				v.DrawCompact(scr, mapstyle.Rect{W: 40, H: 15}, mapstyle.Frame{Model: m, Anim: f})
				if strings.ContainsRune(capture.Text(scr), ufo) {
					t.Errorf("%s %s frame %d: the visitor shows in the mini map", style, age, f)
					scr.Fini()
					break
				}
				scr.Fini()
			}
		}
	}
}

// TestSkyFrameBudget: in both styles a sky age's 200x60 frame takes well
// under the map's 125 ms redraw (the budget is a fifth of it, generous for
// a busy CI runner; actuals are logged).
func TestSkyFrameBudget(t *testing.T) {
	if testing.Short() || raceEnabled {
		t.Skip("timing")
	}
	const budget = 25 * time.Millisecond
	for _, style := range Registry().Names() {
		for _, age := range arcAges[1:] {
			m := arcModel(age)
			v, _ := Registry().New(style)
			scr := capture.NewScreen(200, 60)
			r := mapstyle.Rect{W: 200, H: 60}
			v.Draw(scr, r, mapstyle.Frame{Model: m})
			const n = 15
			start := time.Now()
			for i := 0; i < n; i++ {
				v.Draw(scr, r, mapstyle.Frame{Model: m, Anim: i, Tier: mapmodel.TierUnicode})
			}
			per := time.Since(start) / n
			t.Logf("%s %s: %v a frame", style, age, per)
			if per > budget {
				t.Errorf("%s %s: a frame takes %v, over %v", style, age, per, budget)
			}
			scr.Fini()
		}
	}
}

// TestSkyStatesSorted keeps the arc's age list in game order.
func TestSkyStatesSorted(t *testing.T) {
	cat := skyBuilder.Catalog()
	idx := make([]int, len(arcAges))
	for i, a := range arcAges {
		idx[i] = cat.AgeIdx[a]
	}
	if !sort.IntsAreSorted(idx) || cat.SkySceneAt(idx[0]) != mapmodel.SkyGround {
		t.Errorf("arcAges out of order: %v", idx)
	}
}
