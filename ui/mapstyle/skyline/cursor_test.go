package skyline

import (
	"testing"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/mapmodel/fixture"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
)

// filled counts the cells of s whose background is c.
func filled(s tcell.SimulationScreen, c tcell.Color) int {
	n := 0
	cells, _, _ := s.GetContents()
	for _, cell := range cells {
		if _, bg, _ := cell.Style.Decompose(); bg == c {
			n++
		}
	}
	return n
}

// TestCursorIsHighlighted: in every theme the skyline's cursor is drawn as
// highlighted cells in the cursor's own fill (a bar under the lot it is on
// and the lot's name on a label), where no cell of the scene is drawn in
// that fill without it; the glyphs on it are readable, and the fill stands
// clear of the backgrounds the theme's palette puts under a cell. It blinks
// to its brighter shade while the map moves and holds the fill at frame 0.
//
// The skyline also paints its scenery from fixed tables (materials, skies,
// lights), which no theme changes: the fill is held to the theme's palette
// and to what a scene really shows, not to every paint in those tables.
func TestCursorIsHighlighted(t *testing.T) {
	orig := theme.Active().Key
	t.Cleanup(func() { _ = theme.SetActive(orig) })
	for _, th := range theme.All() {
		if err := theme.SetActive(th.Key); err != nil {
			t.Fatal(err)
		}
		for _, age := range []string{"stone_age", "bronze_age", "medieval_age", "industrial_age"} {
			m := build(fixture.Options{Age: age, Seed: 3, Tick: tickAt(0.5)}, 0)
			v := newView()
			plain := draw(v, m, 120, 40, 0, mapmodel.TierUnicode, false)
			fg, fill, attr := v.mp.Cursor.Decompose()
			_, alt, _ := v.mp.CursorAlt.Decompose()
			if n := filled(plain, fill) + filled(plain, alt); n != 0 {
				t.Errorf("%s, %s: %d cells of the scene are drawn in the cursor's fill with no cursor out", th.Key, age, n)
			}
			v.SetOption(mapstyle.OptInspect, true)
			if n := filled(draw(v, m, 120, 40, 0, mapmodel.TierUnicode, false), fill); n < 3 {
				t.Errorf("%s, %s: the cursor highlights %d cells, want a bar and a label", th.Key, age, n)
			}
			if n := filled(draw(v, m, 120, 40, mapstyle.CursorBlinkFrames, mapmodel.TierUnicode, false), alt); n < 3 {
				t.Errorf("%s, %s: half a blink on, %d cells are in the brighter shade, want the bar and the label", th.Key, age, n)
			}
			if c := theme.ContrastRatio(fg, fill); c < mapstyle.CursorGlyphContrast || attr&tcell.AttrBold == 0 {
				t.Errorf("%s, %s: a glyph on the cursor has contrast %.2f (bold %v), want %.1f and bold", th.Key, age, c, attr&tcell.AttrBold != 0, mapstyle.CursorGlyphContrast)
			}
			if _, ground, over, _ := v.mp.CursorMargins(); ground < mapstyle.CursorFromGround || over < mapstyle.CursorOverGround {
				t.Errorf("%s, %s: the cursor over the palette's backgrounds: distance %.3f, contrast %.2f", th.Key, age, ground, over)
			}
		}
	}
}
