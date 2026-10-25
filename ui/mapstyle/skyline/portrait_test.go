package skyline

import (
	"testing"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/mapmodel/fixture"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
	"github.com/espresso20/ageforge/ui/mapstyle/capture"
	"github.com/espresso20/ageforge/ui/mapstyle/styletest"
)

// TestPortraitEveryWonder: every wonder of the game is drawn alone, whatever
// age the town is in, in every glyph tier and several themes, in a panel as
// small as the game's smallest and as large as a big terminal's: every cell
// written, the wonder's own cells in it, nothing outside ASCII in the plain
// tier and no glyph in its own colour.
func TestPortraitEveryWonder(t *testing.T) {
	themes := theme.All()
	c := catalog()
	keys := 0
	m := build(fixture.Options{Age: "fusion_age", Seed: 3, Tick: tickAt(0.5)}, 0)
	for a := range c.ByAge {
		key := wonderKeyOf(a)
		if key == "" {
			continue
		}
		keys++
		for si, sz := range [][2]int{{66, 13}, {100, 30}, {40, 9}, {160, 42}} {
			for ti, tier := range []mapmodel.GlyphTier{mapmodel.TierASCII, mapmodel.TierUnicode, mapmodel.TierNerd} {
				fixture.WithTheme(t, themes[(a+si+ti)%len(themes)].Key)
				v := newView()
				scr := capture.NewScreen(sz[0], sz[1])
				if !v.Portrait(scr, mapstyle.Rect{W: sz[0], H: sz[1]}, mapstyle.Frame{Model: m, Anim: 3, Tier: tier}, key) {
					t.Fatalf("%s at %dx%d: no portrait", key, sz[0], sz[1])
				}
				scr.Show()
				cells, w, h := scr.GetContents()
				glyphs := 0
				for i, cell := range cells {
					if len(cell.Runes) == 0 {
						t.Fatalf("%s at %dx%d: cell %d,%d was never written", key, sz[0], sz[1], i%w, i/w)
					}
					r := cell.Runes[0]
					if f := queryRules.Fault(r, tier); f != "" {
						t.Fatalf("%s %dx%d %v: rune %q: %s", key, sz[0], sz[1], tier, r, f)
					}
					if styletest.HiddenInk(cell) {
						t.Fatalf("%s %dx%d %v: %q has fg == bg", key, sz[0], sz[1], tier, r)
					}
					if r != ' ' {
						glyphs++
					}
				}
				if w != sz[0] || h != sz[1] {
					t.Fatalf("the screen is %dx%d", w, h)
				}
				if sz[1] >= 9 && glyphs < 6 {
					t.Errorf("%s at %dx%d %v: %d glyphs: the wonder is not there", key, sz[0], sz[1], tier, glyphs)
				}
			}
		}
	}
	if keys != 22 {
		t.Fatalf("%d wonders drawn", keys)
	}
}

// TestPortraitIsPlainAndStable: the same wonder twice is the same picture,
// two wonders are two pictures, the view's own camera and zoom are not
// touched, and what is not a wonder, or has no room, is not drawn.
func TestPortraitIsPlainAndStable(t *testing.T) {
	m := build(fixture.Options{Age: "modern_age", Seed: 3, Tick: tickAt(0.5)}, 0)
	v := newView()
	v.cam, v.zoom, v.follow = 77, zClose, false
	text := func(key string) string {
		scr := capture.NewScreen(80, 20)
		if !v.Portrait(scr, mapstyle.Rect{W: 80, H: 20}, mapstyle.Frame{Model: m, Anim: 1, Tier: mapmodel.TierUnicode}, key) {
			t.Fatalf("%s: no portrait", key)
		}
		return capture.Frame(scr)
	}
	a, b := text("stonehenge"), text("eiffel_tower")
	if a == b {
		t.Errorf("two wonders drew the same picture")
	}
	if again := text("stonehenge"); again != a {
		t.Errorf("the same wonder drew differently twice")
	}
	if v.cam != 77 || v.zoom != zClose || v.follow {
		t.Errorf("the portrait moved the view: cam %d zoom %s follow %v", v.cam, v.zoom, v.follow)
	}
	scr := capture.NewScreen(80, 20)
	for _, key := range []string{"hut", "no_such_thing", ""} {
		if v.Portrait(scr, mapstyle.Rect{W: 80, H: 20}, mapstyle.Frame{Model: m}, key) {
			t.Errorf("a portrait of %q", key)
		}
	}
	if v.Portrait(scr, mapstyle.Rect{W: 4, H: 3}, mapstyle.Frame{Model: m}, "stonehenge") {
		t.Errorf("a portrait with no room")
	}
	if v.Portrait(scr, mapstyle.Rect{W: 80, H: 20}, mapstyle.Frame{}, "stonehenge") {
		t.Errorf("a portrait without a model")
	}
}

// TestPortraitIgnoresTheTownsHour: the portrait is lit at its own dusk, the
// lamps of the landmarks on and the sky still blue, whatever hour (and
// weather) the town is in, and it is the picture of the wonder's own age
// whatever age the town has reached.
func TestPortraitIgnoresTheTownsHour(t *testing.T) {
	picture := func(age string, tod float64, key string) string {
		m := build(fixture.Options{Age: age, Seed: 3, Tick: tickAt(tod)}, 0)
		scr := capture.NewScreen(80, 24)
		if !newView().Portrait(scr, mapstyle.Rect{W: 80, H: 24}, mapstyle.Frame{Model: m, Tier: mapmodel.TierUnicode}, key) {
			t.Fatalf("no portrait of %s", key)
		}
		return capture.Frame(scr)
	}
	noon, night := picture("fusion_age", 0.5, "grand_lighthouse"), picture("fusion_age", 0.95, "grand_lighthouse")
	if noon != night {
		t.Errorf("the portrait changed with the town's hour")
	}
	if early := picture("medieval_age", 0.5, "grand_lighthouse"); early != noon {
		t.Errorf("the portrait of a wonder changed with the age the town is in")
	}
}
