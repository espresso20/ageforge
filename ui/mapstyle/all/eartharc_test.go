package all

import (
	"math"
	"sort"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/mapmodel/fixture"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
	"github.com/espresso20/ageforge/ui/mapstyle/capture"
)

// eartharc_test.go holds the Earth arc (the Modern Age to the Fusion Age)
// to its brief across both styles: every age reads at a glance (the
// disparity guard), the green fades by the rule, and nothing of a later age
// shows early.

// disparityMin is the disparity guard's threshold: two neighbouring ages'
// frames must differ by at least this much (see signatureDistance). It sits
// well above what one age's frames differ by among themselves at different
// moments and seeds, and well below what the redesign's neighbours differ
// by; the old maps, where an age drew like the one before it with a few
// buildings more, scored under it (the PR records the numbers).
const disparityMin = 0.25

// signature is a frame's look: how often each glyph is drawn and how often
// each colour shows (glyph inks, and every cell's background), over the map
// area only (header, inspector, status line and legend left out).
type signature struct {
	glyph map[rune]float64
	color map[int]float64
}

// colorBucket quantises a colour to a 6x6x6 cube on a perceptual (square
// root) scale, so dark colours, which a night frame is mostly made of,
// count as distinct as they look.
func colorBucket(c tcell.Color) int {
	r, g, b := c.RGB()
	if r < 0 {
		return -1
	}
	q := func(v int32) int { return min(5, int(math.Sqrt(float64(v)/255)*6)) }
	return q(r)*36 + q(g)*6 + q(b)
}

func frameSignature(s tcell.SimulationScreen, x0, y0, x1, y1 int) signature {
	sg := signature{glyph: map[rune]float64{}, color: map[int]float64{}}
	s.Show()
	cells, w, _ := s.GetContents()
	ng, nc := 0.0, 0.0
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			c := cells[y*w+x]
			r := ' '
			if len(c.Runes) > 0 {
				r = c.Runes[0]
			}
			fg, bg, _ := c.Style.Decompose()
			sg.color[colorBucket(bg)]++
			nc++
			if r != ' ' {
				sg.glyph[r]++
				sg.color[colorBucket(fg)]++
				ng++
				nc++
			}
		}
	}
	for k := range sg.glyph {
		sg.glyph[k] /= math.Max(1, ng)
	}
	for k := range sg.color {
		sg.color[k] /= math.Max(1, nc)
	}
	return sg
}

// tvd is the total variation distance between two distributions: 0 the
// same, 1 nothing in common.
func tvd[K comparable](a, b map[K]float64) float64 {
	d := 0.0
	for k, v := range a {
		d += math.Abs(v - b[k])
	}
	for k, v := range b {
		if _, ok := a[k]; !ok {
			d += v
		}
	}
	return d / 2
}

// signatureDistance is how unalike two frames look: the mean of the total
// variation distances of their glyph and colour frequencies, from 0 (drawn
// alike) to 1 (nothing in common).
func signatureDistance(a, b signature) float64 {
	return (tvd(a.glyph, b.glyph) + tvd(a.color, b.color)) / 2
}

// earthFrame draws a style's frame of a synthetic age at 160x48 on the
// default theme and returns its map area's signature: the roguelike's
// settlement left of its legend, the skyline's scene rows, by day or
// night (tod).
func earthFrame(t *testing.T, style, age string, seed int64, tod float64, anim int) signature {
	t.Helper()
	tick := 2*mapmodel.DayTicks + int(tod*mapmodel.DayTicks) - mapmodel.DayTicks*3/8
	st := fixture.State(fixture.Options{Age: age, Seed: seed, Tick: tick, Routes: 3})
	m := mapmodel.NewBuilder(nil).Build(&st, nil)
	v, _ := Registry().New(style)
	v.SetOption(mapstyle.OptInspect, false)
	scr := capture.NewScreen(160, 48)
	defer scr.Fini()
	v.Draw(scr, mapstyle.Rect{W: 160, H: 48}, mapstyle.Frame{Model: m, Anim: anim, Tier: mapmodel.TierUnicode})
	if style == "roguelike" {
		return frameSignature(scr, 0, 1, 160-26, 48-4)
	}
	return frameSignature(scr, 0, 1, 160, 48-2)
}

func withTheme(t *testing.T, key string) {
	t.Helper()
	orig := theme.Active().Key
	t.Cleanup(func() { _ = theme.SetActive(orig) })
	if err := theme.SetActive(key); err != nil {
		t.Fatal(err)
	}
}

// TestEarthArcDisparity is the disparity guard: in both styles, by day and
// by night, on two seeds, every Earth-arc age's frame differs from its
// neighbour's (the Atomic Age's before the Modern Age) by at least
// disparityMin, so no two neighbours look alike. The owner found the late
// ages too alike ("space kinda looks like cyber"); this keeps them apart.
func TestEarthArcDisparity(t *testing.T) {
	withTheme(t, "forge")
	ages := []string{"atomic_age", "modern_age", "information_age", "digital_age", "cyberpunk_age", "fusion_age"}
	for _, seed := range []int64{7, 11} {
		for _, style := range []string{"roguelike", "skyline"} {
			for _, tod := range []float64{0.45, 0.95} {
				if style == "roguelike" && tod > 0.5 {
					continue // the roguelike draws the same by night but for its satellites and headlights
				}
				sigs := make([]signature, len(ages))
				for i, a := range ages {
					sigs[i] = earthFrame(t, style, a, seed, tod, 600)
				}
				for i := 1; i < len(ages); i++ {
					d := signatureDistance(sigs[i-1], sigs[i])
					t.Logf("seed %d %s tod %.2f: %s -> %s: %.3f", seed, style, tod, ages[i-1], ages[i], d)
					if d < disparityMin {
						t.Errorf("seed %d %s (tod %.2f): the %s looks too like the %s (distance %.3f, want at least %.2f)",
							seed, style, tod, ages[i], ages[i-1], d, disparityMin)
					}
				}
			}
		}
	}
}

// TestEarthArcDisparityThreshold keeps the guard honest: one age's frame a
// few seconds and a few minutes of game time later (the traffic moved, the
// neon flickered, the clouds drifted) differs from itself by well under
// the threshold, so the guard measures the ages, not the animation.
func TestEarthArcDisparityThreshold(t *testing.T) {
	withTheme(t, "forge")
	worst := 0.0
	for _, style := range []string{"roguelike", "skyline"} {
		for _, age := range []string{"atomic_age", "modern_age", "information_age", "digital_age", "cyberpunk_age", "fusion_age"} {
			for _, tod := range []float64{0.45, 0.95} {
				a := earthFrame(t, style, age, 7, tod, 600)
				b := earthFrame(t, style, age, 7, tod+0.01, 680)
				d := signatureDistance(a, b)
				worst = math.Max(worst, d)
				if d > disparityMin*0.6 {
					t.Errorf("%s %s (tod %.2f): the same age a moment later differs by %.3f, too near the guard's %.2f",
						style, age, tod, d, disparityMin)
				}
			}
		}
	}
	t.Logf("the same age a moment later: at most %.3f (the guard asks %.2f of neighbours)", worst, disparityMin)
}

// greenShare is how much of a frame's map area reads as greenery: cells
// whose glyph or background is a natural green (dark-to-mid, green well
// ahead of red and blue), not neon and not a lime light.
func greenShare(s tcell.SimulationScreen, x0, y0, x1, y1 int) float64 {
	s.Show()
	cells, w, _ := s.GetContents()
	green := func(c tcell.Color) bool {
		r, g, b := c.RGB()
		if r < 0 {
			return false
		}
		return g > r+28 && g > b+30 && g < 215 // (a teal glass, blue near the green, is no greenery)
	}
	n, tot := 0, 0
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			c := cells[y*w+x]
			fg, bg, _ := c.Style.Decompose()
			ink := len(c.Runes) > 0 && c.Runes[0] != ' '
			if ink && green(fg) || green(bg) {
				n++
			}
			tot++
		}
	}
	return float64(n) / float64(max(1, tot))
}

// TestEarthArcGreeneryOnScreen checks the greenery fade where the player
// sees it, in both styles' frames (on two seeds, the settlement and the
// skyline by day): the Information Age shows well under the Modern Age's
// green, the Digital Age less again (its reserve), and the megacity none.
func TestEarthArcGreeneryOnScreen(t *testing.T) {
	withTheme(t, "forge")
	for _, style := range []string{"roguelike", "skyline"} {
		for _, seed := range []int64{7, 11} {
			share := map[string]float64{}
			for _, age := range []string{"modern_age", "information_age", "digital_age", "cyberpunk_age", "fusion_age"} {
				tick := 2*mapmodel.DayTicks + int(0.45*mapmodel.DayTicks) - mapmodel.DayTicks*3/8
				st := fixture.State(fixture.Options{Age: age, Seed: seed, Tick: tick})
				m := mapmodel.NewBuilder(nil).Build(&st, nil)
				v, _ := Registry().New(style)
				v.SetOption(mapstyle.OptChanges, false)
				scr := capture.NewScreen(160, 48)
				v.Draw(scr, mapstyle.Rect{W: 160, H: 48}, mapstyle.Frame{Model: m, Anim: 600, Tier: mapmodel.TierUnicode})
				if style == "roguelike" {
					share[age] = greenShare(scr, 0, 1, 160-26, 48-4)
				} else {
					share[age] = greenShare(scr, 0, 1, 160, 48-2)
				}
				scr.Fini()
			}
			t.Logf("%s seed %d: %v", style, seed, share)
			mod := share["modern_age"]
			switch {
			case mod < 0.01:
				t.Errorf("%s seed %d: the Modern Age shows almost no green (%.3f)", style, seed, mod)
			case share["information_age"] > mod*0.75 || share["information_age"] < mod*0.25:
				t.Errorf("%s seed %d: the Information Age shows %.3f green to the Modern Age's %.3f, want about half",
					style, seed, share["information_age"], mod)
			case share["digital_age"] > mod*0.4 || share["digital_age"] > share["information_age"]:
				t.Errorf("%s seed %d: the Digital Age shows %.3f green (Information %.3f, Modern %.3f)", style, seed,
					share["digital_age"], share["information_age"], mod)
			case share["digital_age"] <= 0:
				t.Errorf("%s seed %d: the Digital Age shows no reserve", style, seed)
			}
			for _, a := range []string{"cyberpunk_age", "fusion_age"} {
				if share[a] > mod*0.02 {
					t.Errorf("%s seed %d: the %s still shows %.4f green", style, seed, a, share[a])
				}
			}
		}
	}
}

// TestEarthArcNoSpoilers: a frame of each Earth-arc age, in both styles,
// at every zoom, with the cursor stepped through every target, names no age
// or era not yet reached, and no city feature of a later age (by its legend
// name or its inspect title).
func TestEarthArcNoSpoilers(t *testing.T) {
	for _, age := range []string{"atomic_age", "modern_age", "information_age", "digital_age", "cyberpunk_age", "fusion_age"} {
		st := fixture.State(fixture.Options{Age: age, Seed: 5, Harbinger: true, Routes: 2})
		txt := mapText(t, st)
		for _, word := range unreached(st) {
			if i := strings.Index(txt, word); i >= 0 {
				lo, hi := max(0, i-60), min(len(txt), i+len(word)+60)
				t.Errorf("%s: the map shows %q:\n...%s...", age, word, txt[lo:hi])
			}
		}
		ai := 0
		for i, k := range config.AgeOrder() {
			if k == age {
				ai = i
			}
		}
		var later []string
		for f := mapmodel.CityFeature(1); f < mapmodel.NumCityFeatures; f++ {
			if i := f.Info(); i.From > ai {
				later = append(later, i.Name, i.Title)
			}
		}
		sort.Strings(later)
		for _, word := range later {
			if strings.Contains(txt, word) && !earlierNamed(word, ai) {
				t.Errorf("%s: the map names %q, which a later age builds", age, word)
			}
		}
	}
}

// earlierNamed reports a later feature's name that an age's own feature
// shares (none do today; the check keeps a shared word from failing).
func earlierNamed(word string, age int) bool {
	for f := mapmodel.CityFeature(1); f < mapmodel.NumCityFeatures; f++ {
		if i := f.Info(); i.From <= age && (i.Name == word || i.Title == word) {
			return true
		}
	}
	return false
}
