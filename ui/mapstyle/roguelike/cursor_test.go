package roguelike

import (
	"testing"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/mapmodel/fixture"
	"github.com/espresso20/ageforge/rules"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
)

// cursorCells returns the cells of s drawn in the cursor's fill (either
// half of its blink), with the rune and style of each.
func cursorCells(s tcell.SimulationScreen, p *mapstyle.Palette) (cells []tcell.SimCell, at [][2]int) {
	_, fill, _ := p.Cursor.Decompose()
	_, alt, _ := p.CursorAlt.Decompose()
	all, w, h := s.GetContents()
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c := all[y*w+x]
			if _, bg, _ := c.Style.Decompose(); bg == fill || bg == alt {
				cells = append(cells, c)
				at = append(at, [2]int{x, y})
			}
		}
	}
	return cells, at
}

// TestCursorIsAHighlightedCell: in every theme and in every age's palette
// (a night city has colours of its own), the roguelike's cursor is one
// highlighted cell in the cursor's fill, which no other cell of the map is
// drawn in; the glyph beneath it is kept and stays readable; and the fill
// keeps the margins every cursor keeps from the map's own colours.
func TestCursorIsAHighlightedCell(t *testing.T) {
	orig := theme.Active().Key
	t.Cleanup(func() { _ = theme.SetActive(orig) })
	ages := rules.Core().AgeKeys()
	for _, th := range theme.All() {
		if err := theme.SetActive(th.Key); err != nil {
			t.Fatal(err)
		}
		for i, age := range ages {
			m, _ := grown(t, fixture.Options{Age: age, Seed: 5})
			if m.Sky() != mapmodel.SkyGround {
				continue // the sky ages are held below, on their own view
			}
			v := newView()
			v.zoom, v.inspect = zSettlement, false
			// Without the cursor, then with it: the only cells that change
			// are the cursor's.
			plain := draw(v, m, 120, 40, 0, mapmodel.TierUnicode, false)
			if cells, _ := cursorCells(plain, v.pal.Palette); len(cells) != 0 {
				t.Errorf("%s, %s: %d cells are drawn in the cursor's fill with no cursor out", th.Key, age, len(cells))
			}
			before, w, _ := plain.GetContents()
			v.inspect = true
			lit := draw(v, m, 120, 40, 0, mapmodel.TierUnicode, false)
			cells, at := cursorCells(lit, v.pal.Palette)
			if len(cells) != v.g.cellW || len(cells) == 0 {
				t.Errorf("%s, %s: the cursor highlights %d cells, want the %d of one tile", th.Key, age, len(cells), v.g.cellW)
				continue
			}
			for k, c := range cells {
				under := before[at[k][1]*w+at[k][0]]
				if len(c.Runes) == 0 || len(under.Runes) == 0 || c.Runes[0] != under.Runes[0] {
					t.Errorf("%s, %s: the cursor hides the glyph beneath it", th.Key, age)
				}
				// The cell beneath it looked nothing like it.
				_, ubg, _ := under.Style.Decompose()
				_, fill, _ := c.Style.Decompose()
				if d, cr := theme.Distance(fill, ubg), theme.ContrastRatio(fill, ubg); d < mapstyle.CursorFromGround || cr < mapstyle.CursorOverGround {
					t.Errorf("%s, %s: the cursor #%06x over the cell beneath it #%06x: distance %.3f, contrast %.2f", th.Key, age, fill.Hex(), ubg.Hex(), d, cr)
				}
			}
			// The fill against the palette this age really draws with.
			glyph, ground, over, text := v.pal.CursorMargins()
			if glyph < mapstyle.CursorGlyphMargin() || ground < mapstyle.CursorFromGround || over < mapstyle.CursorOverGround || text < mapstyle.CursorGlyphContrast {
				_, fill, _ := v.pal.Cursor.Decompose()
				t.Errorf("%s, %s (age %d): the cursor's fill #%06x stands %.3f from the nearest glyph colour, %.3f and contrast %.2f from the nearest background, glyph contrast %.2f; want %.2f, %.2f, %.1f, %.1f",
					th.Key, age, i, fill.Hex(), glyph, ground, over, text,
					mapstyle.CursorGlyphMargin(), mapstyle.CursorFromGround, mapstyle.CursorOverGround, mapstyle.CursorGlyphContrast)
			}
		}
	}
}

// TestCursorBlinksWithTheMap: the cursor blinks slowly while the map moves
// (the fill, then its brighter shade, each for mapstyle.CursorBlinkFrames
// frames) and holds the fill while the motion setting holds the frame at 0.
func TestCursorBlinksWithTheMap(t *testing.T) {
	m, _ := grown(t, fixture.Options{Age: "stone_age", Seed: 5})
	v := newView()
	v.zoom, v.inspect = zSettlement, true
	bgAt := func(anim int) tcell.Color {
		s := draw(v, m, 120, 40, anim, mapmodel.TierUnicode, false)
		cells, _ := cursorCells(s, v.pal.Palette)
		if len(cells) == 0 {
			t.Fatalf("frame %d: no cursor", anim)
		}
		_, bg, _ := cells[0].Style.Decompose()
		return bg
	}
	_, fill, _ := mapstyle.NewPalette(0).Cursor.Decompose()
	_, alt, _ := mapstyle.NewPalette(0).CursorAlt.Decompose()
	if got := bgAt(0); got != fill {
		t.Errorf("with motion off the cursor is #%06x, want the fill #%06x", got.Hex(), fill.Hex())
	}
	if got := bgAt(mapstyle.CursorBlinkFrames); got != alt {
		t.Errorf("half a blink on the cursor is #%06x, want the brighter shade #%06x", got.Hex(), alt.Hex())
	}
	if got := bgAt(2 * mapstyle.CursorBlinkFrames); got != fill {
		t.Errorf("a whole blink on the cursor is #%06x, want the fill #%06x again", got.Hex(), fill.Hex())
	}
}

// TestMiniMapMarksTheCursor: the compact view marks the tile another
// view's cursor stands on with the same highlighted cell, and nothing when
// that cursor is not out.
func TestMiniMapMarksTheCursor(t *testing.T) {
	m, _ := grown(t, fixture.Options{Age: "stone_age", Seed: 5})
	panel, mini := withSky(newView()), withSky(newView())
	panel.g.zoom, panel.g.inspect = zSettlement, true
	draw(panel.g, m, 120, 40, 0, mapmodel.TierUnicode, false) // places the cursor
	src, dst := mapstyle.Pointing(panel), mapstyle.Pointing(mini)

	dst.Point(src.Pointer())
	s := drawMini(mini, m, 58, 9)
	if cells, _ := cursorCells(s, mini.g.pal.Palette); len(cells) != 1 {
		t.Errorf("the mini map marks %d cells for the Map panel's cursor, want 1", len(cells))
	}
	panel.g.inspect = false
	dst.Point(src.Pointer())
	s = drawMini(mini, m, 58, 9)
	if cells, _ := cursorCells(s, mini.g.pal.Palette); len(cells) != 0 {
		t.Errorf("the mini map marks %d cells with no cursor out, want none", len(cells))
	}
}

// drawMini draws a style's compact view at w by h.
func drawMini(st *skyStyle, m *mapmodel.Model, w, h int) tcell.SimulationScreen {
	s := tcell.NewSimulationScreen("UTF-8")
	_ = s.Init()
	s.SetSize(w, h)
	st.DrawCompact(s, mapstyle.Rect{X: 0, Y: 0, W: w, H: h}, mapstyle.Frame{Model: m, Tier: mapmodel.TierUnicode})
	s.Show()
	return s
}
