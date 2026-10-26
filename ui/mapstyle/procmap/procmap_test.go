package procmap

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapgen"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/mapmodel/fixture"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
	"github.com/espresso20/ageforge/ui/mapstyle/capture"
)

var cat = mapmodel.NewCatalog()

const (
	noon     = 225 + 1800*5
	midnight = 1125 + 1800*5
)

func model(age string, tick int) *mapmodel.Model {
	st := fixture.State(fixture.Options{Age: age, Seed: 4, Tick: tick})
	return mapmodel.NewBuilder(cat).Build(&st, nil)
}

// sizes are the three the map is drawn at: a big panel, a small one, and the
// mini map's space.
var sizes = []struct {
	name    string
	w, h    int
	compact bool
}{{"120x36", 120, 36, false}, {"80x24", 80, 24, false}, {"mini", 40, 9, true}}

func draw(v mapstyle.Style, m *mapmodel.Model, w, h int, compact bool, tier mapmodel.GlyphTier) tcell.SimulationScreen {
	scr := capture.NewScreen(w, h)
	f := mapstyle.Frame{Model: m, Tier: tier}
	if compact {
		v.DrawCompact(scr, mapstyle.Rect{W: w, H: h}, f)
	} else {
		v.Draw(scr, mapstyle.Rect{W: w, H: h}, f)
	}
	scr.Show()
	return scr
}

// TestEveryAgeDrawsAtEverySize: every age of the game draws in a big panel, a
// small one and the mini map, day and night, with something on the screen.
func TestEveryAgeDrawsAtEverySize(t *testing.T) {
	fixture.WithTheme(t, "forge")
	for _, age := range cat.Ages {
		for _, tick := range []int{noon, midnight} {
			m := model(age, tick)
			for _, sz := range sizes {
				scr := draw(New(), m, sz.w, sz.h, sz.compact, mapmodel.TierUnicode)
				if strings.TrimSpace(capture.Text(scr)) == "" {
					t.Errorf("%s %s tick %d: blank screen", age, sz.name, tick)
				}
				scr.Fini()
			}
		}
	}
}

// TestEveryThemeDraws: the map draws in every theme, light and dark, by day
// and by night, and a glyph never sinks into its ground.
func TestEveryThemeDraws(t *testing.T) {
	for th := range fixture.Themes(t) {
		for _, age := range []string{"bronze_age", "modern_age"} {
			for _, tick := range []int{noon, midnight} {
				scr := draw(New(), model(age, tick), 120, 36, false, mapmodel.TierUnicode)
				cells, w, h := scr.GetContents()
				for i, c := range cells[:w*h] {
					if len(c.Runes) == 0 || c.Runes[0] == ' ' {
						continue
					}
					fg, bg, _ := c.Style.Decompose()
					if cr := theme.ContrastRatio(fg, bg); cr < 1.5 {
						t.Fatalf("theme %s, %s tick %d: %q at (%d,%d) has contrast %.2f", th.Key, age, tick, c.Runes[0], i%w, i/w, cr)
					}
				}
				scr.Fini()
			}
		}
	}
}

// TestAsciiTierIsAscii: on the ASCII tier every cell of every age is plain
// ASCII, and none is the question mark a rune with no fold becomes.
func TestAsciiTierIsAscii(t *testing.T) {
	fixture.WithTheme(t, "forge")
	for _, age := range cat.Ages {
		for _, sz := range sizes {
			scr := draw(New(), model(age, noon), sz.w, sz.h, sz.compact, mapmodel.TierASCII)
			cells, w, h := scr.GetContents()
			for i, c := range cells[:w*h] {
				if len(c.Runes) > 0 && (c.Runes[0] > 0x7e || c.Runes[0] == '?') {
					t.Fatalf("%s %s: %q at (%d,%d) is not clean ASCII", age, sz.name, c.Runes[0], i%w, i/w)
				}
			}
			scr.Fini()
		}
	}
}

// TestEveryGlyphFolds: every glyph the engine can draw has a plain ASCII
// shape, so no terminal gets a question mark.
func TestEveryGlyphFolds(t *testing.T) {
	for _, r := range mapgen.Runes() {
		if f := mapmodel.Fold(r, mapmodel.TierASCII); f > 0x7e || f == '?' || f < ' ' {
			t.Errorf("%q folds to %q", r, f)
		}
	}
}

// TestSmallPanelsShowTheWholeTown: a small panel and the mini map show the
// overview, and the town's own signs are on it.
func TestSmallPanelsShowTheWholeTown(t *testing.T) {
	fixture.WithTheme(t, "forge")
	m := model("modern_age", noon)
	for _, sz := range sizes[1:] {
		v := New().(*view)
		scr := draw(v, m, sz.w, sz.h, sz.compact, mapmodel.TierUnicode)
		text := capture.Text(scr)
		if !sz.compact && !v.overview {
			t.Errorf("%s: a small panel shows the close-up", sz.name)
		}
		signs := 0
		for _, l := range m.Lineages {
			if strings.ContainsRune(text, mapgen.Sign(l.Key)) {
				signs++
			}
		}
		if signs < 4 {
			t.Errorf("%s: only %d building signs on the overview:\n%s", sz.name, signs, text)
		}
		scr.Fini()
	}
	v := New().(*view)
	draw(v, m, 120, 36, false, mapmodel.TierUnicode).Fini()
	if v.overview {
		t.Error("a big panel shows the overview")
	}
}

// TestCursorWalksTheTown: Tab lands on a building and the inspector names it
// with the game's own words; the arrows move the cursor.
func TestCursorWalksTheTown(t *testing.T) {
	fixture.WithTheme(t, "forge")
	m := model("modern_age", noon)
	v := New()
	f := mapstyle.Frame{Model: m}
	draw(v, m, 120, 36, false, mapmodel.TierUnicode).Fini()
	start, _ := v.Pointer()
	tab := tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone)
	var in mapstyle.Inspection
	for i := 0; i < 60 && !strings.Contains(in.Title, "×"); i++ {
		if !v.HandleKey(tab, f) {
			t.Fatal("Tab was not taken")
		}
		in, _ = v.Inspect(f)
		if i == 0 && in.Command == "" {
			t.Fatalf("after Tab the inspector has nothing to type: %+v", in)
		}
	}
	if !strings.Contains(in.Title, "×") || in.Command == "" {
		t.Fatalf("Tab never reached a building; the inspector says %+v", in)
	}
	if at, _ := v.Pointer(); at == start {
		t.Fatal("Tab did not move the cursor")
	}
	before, _ := v.Pointer()
	v.HandleKey(tcell.NewEventKey(tcell.KeyRight, 0, tcell.ModShift), f)
	if after, _ := v.Pointer(); after.X != before.X+8 {
		t.Fatalf("Shift-Right moved from %v to %v", before, after)
	}
	v.HandleKey(tcell.NewEventKey(tcell.KeyHome, 0, tcell.ModNone), f)
	sq, _ := v.Inspect(f)
	if !strings.Contains(sq.Title, m.Town.World.Name) {
		t.Fatalf("Home is on %q, not the square of %s", sq.Title, m.Town.World.Name)
	}
}

// TestMiniMapMarksThePanelsCursor: the mini map highlights the cell the panel's
// cursor stands in.
func TestMiniMapMarksThePanelsCursor(t *testing.T) {
	fixture.WithTheme(t, "forge")
	m := model("modern_age", noon)
	panel := New()
	draw(panel, m, 120, 36, false, mapmodel.TierUnicode).Fini()
	mini := New()
	plain := draw(mini, m, 40, 9, true, mapmodel.TierUnicode)
	mini.Point(panel.Pointer())
	marked := draw(mini, m, 40, 9, true, mapmodel.TierUnicode)
	a, _, _ := plain.GetContents()
	b, _, _ := marked.GetContents()
	changed := 0
	for i := range a {
		if a[i].Style != b[i].Style {
			changed++
		}
	}
	if changed != 1 {
		t.Errorf("the mark changed %d cells, want 1", changed)
	}
}
