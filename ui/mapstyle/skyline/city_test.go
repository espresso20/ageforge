package skyline

import (
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

var cityAges = []string{"modern_age", "information_age", "digital_age", "cyberpunk_age", "fusion_age"}

// TestCitySweep draws every Earth-arc age by day and by night, in every
// glyph tier and on every theme, full and compact: every rune one cell (and
// ASCII in the ASCII tier, with no '?' fallback elsewhere), no glyph in its
// own background colour.
func TestCitySweep(t *testing.T) {
	orig := theme.Active().Key
	t.Cleanup(func() { _ = theme.SetActive(orig) })
	var models []*mapmodel.Model
	for _, a := range cityAges {
		for _, tod := range []float64{0.45, 0.95} {
			models = append(models, build(fixture.Options{Age: a, Seed: 6, Routes: 4, Harbinger: true, Tick: tickAt(tod)}, 1))
		}
	}
	for _, m := range models {
		for _, tier := range []mapmodel.GlyphTier{mapmodel.TierASCII, mapmodel.TierUnicode, mapmodel.TierNerd} {
			for _, compact := range []bool{false, true} {
				w, h := 160, 45
				if compact {
					w, h = 40, 15
				}
				v := newView()
				v.inspect, v.legend = true, true
				scr := draw(v, m, w, h, 33, tier, compact)
				cells, _, _ := scr.GetContents()
				for _, c := range cells {
					if len(c.Runes) == 0 {
						continue
					}
					r := c.Runes[0]
					switch {
					case tier == mapmodel.TierASCII && r >= 0x80:
						t.Fatalf("%s ascii compact=%v: non-ASCII rune %q", m.Age, compact, r)
					case tier != mapmodel.TierASCII && uniseg.StringWidth(string(r)) != 1:
						t.Fatalf("%s %s compact=%v: rune %q is not one cell wide", m.Age, tier, compact, r)
					case tier != mapmodel.TierASCII && r == '?':
						t.Fatalf("%s %s compact=%v: a '?' fallback leaked into the frame", m.Age, tier, compact)
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
			for _, compact := range []bool{false, true} {
				w, h := 120, 36
				if compact {
					w, h = 40, 15
				}
				scr := draw(newView(), m, w, h, 7, mapmodel.TierUnicode, compact)
				cells, sw, _ := scr.GetContents()
				for i, c := range cells {
					fg, bg, _ := c.Style.Decompose()
					if len(c.Runes) > 0 && c.Runes[0] != ' ' && fg == bg {
						t.Errorf("theme %s, %s compact=%v: %q at (%d,%d) has fg == bg", th.Key, m.Age, compact, c.Runes[0], i%sw, i/sw)
						break
					}
				}
			}
		}
	}
}

// TestCityWorks: each Earth-arc age draws its own works and nothing of a
// later one: the freeway from the Modern Age to the Digital Age, the
// Information Age's cables, three to five megacity rails with sky trains
// on every one, and the tether only from the Fusion Age, where the cursor
// can inspect it.
func TestCityWorks(t *testing.T) {
	for _, age := range append([]string{"atomic_age"}, cityAges...) {
		m := build(fixture.Options{Age: age, Seed: 4, Routes: 3, Tick: tickAt(0.5)}, 0)
		v := newView()
		s := v.compose(mapstyle.Frame{Model: m, Anim: 5}, 160, 45)
		deck := 0
		for x := 0; x < s.W; x++ {
			if c := v.fb.at(x, s.Y(freewayY(s.groundY))); c != nil && c.d == dLane0+1 {
				deck++
			}
		}
		if want := hasFreeway(m.AgeIdx); want != (deck > s.W/4) {
			t.Errorf("%s: freeway deck on %d columns, want a freeway %v", age, deck, want)
		}
		trains, rails := 0, map[int]bool{}
		for _, vh := range trafficFor(m, 5, s.cam, s.W, s.groundY) {
			if vh.t == &tSkyTrain {
				trains++
				rails[vh.y] = true
			}
		}
		switch {
		case age == "cyberpunk_age":
			n := len(cityRailYs(s.groundY))
			if n < 3 || n > 5 || len(rails) != n || trains < 2*n {
				t.Errorf("%s: %d rails, sky trains on %d of them (%d trains)", age, n, len(rails), trains)
			}
		case trains > 0:
			t.Errorf("%s: %d sky trains outside the megacity", age, trains)
		}
		if (s.elevatorX() >= 0) != (m.AgeIdx >= m.Catalog.AgeIdx["fusion_age"]) {
			t.Errorf("%s: the tether stands at %d", age, s.elevatorX())
		}
	}
	// the tether is a Tab target
	m := build(fixture.Options{Age: "fusion_age", Seed: 4, Tick: tickAt(0.5)}, 0)
	v := newView()
	f := mapstyle.Frame{Model: m}
	v.SetOption(mapstyle.OptInspect, true)
	found := false
	for i := 0; i < len(m.Skyline.Lots)+30 && !found; i++ {
		v.HandleKey(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone), f)
		if in, ok := v.Inspect(f); ok && v.cur.kind == tTether {
			found = in.Title == mapmodel.FeatTether.Info().Title && in.Command != ""
			txt := capture.Text(draw(v, m, 160, 48, 1, mapmodel.TierUnicode, false))
			if !strings.Contains(txt, in.Title) {
				t.Errorf("the cursor on the tether does not show its name")
			}
		}
	}
	if !found {
		t.Error("Tab never reached the space elevator")
	}
}

// TestCityPerformance: the Earth arc's busiest skylines keep inside the
// frame budget (the map redraws every 125 ms).
func TestCityPerformance(t *testing.T) {
	if testing.Short() || raceEnabled {
		t.Skip("timing")
	}
	for _, age := range []string{"cyberpunk_age", "fusion_age", "digital_age"} {
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
	}
}
