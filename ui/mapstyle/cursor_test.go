package mapstyle

import (
	"testing"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/theme"
)

// checkCursor holds one palette's cursor, both halves of its blink, to the
// margins every style's cursor keeps (CursorFromGlyph and the rest). where
// names the palette in a failure. It returns the fill's least distance from
// a glyph colour.
func checkCursor(t *testing.T, where string, p *Palette) float64 {
	t.Helper()
	for name, st := range map[string]tcell.Style{"fill": p.Cursor, "blink": p.CursorAlt} {
		if _, _, attr := st.Decompose(); attr&tcell.AttrBold == 0 {
			t.Errorf("%s: the cursor's %s is not bold", where, name)
		}
	}
	_, fill, _ := p.Cursor.Decompose()
	glyph, ground, over, text := p.CursorMargins()
	if glyph < CursorGlyphMargin() {
		t.Errorf("%s: the cursor's fill #%06x is %.3f from a colour the map draws a glyph in, want at least %.2f", where, fill.Hex(), glyph, CursorGlyphMargin())
	}
	if ground < CursorFromGround || over < CursorOverGround {
		t.Errorf("%s: the cursor (fill #%06x) over the map's backgrounds: distance %.3f and contrast %.2f, want %.2f and %.1f", where, fill.Hex(), ground, over, CursorFromGround, CursorOverGround)
	}
	if text < CursorGlyphContrast {
		t.Errorf("%s: a glyph on the cursor (fill #%06x) has contrast %.2f, want %.1f", where, fill.Hex(), text, CursorGlyphContrast)
	}
	return glyph
}

// TestCursorStandsOutInEveryTheme: in each of the themes and each era, the
// cursor is a highlighted cell whose fill is no colour the palette draws:
// clearly apart from every glyph colour and from every background under a
// cell, with the glyph on it readable. The blink's other half is a shade of
// the fill: it is held to the backgrounds and to the glyph on it. On the
// dark themes the fill is warm (an orange); it is always made from the
// theme's own roles.
func TestCursorStandsOutInEveryTheme(t *testing.T) {
	defer func() { _ = theme.SetActive(theme.DefaultKey) }()
	themes := theme.All()
	if len(themes) != 16 {
		t.Fatalf("%d themes, want the 16 the game ships", len(themes))
	}
	for _, th := range themes {
		if err := theme.SetActive(th.Key); err != nil {
			t.Fatal(err)
		}
		least := 9.0
		var fill tcell.Color
		for epoch := 0; epoch < 7; epoch++ {
			p := NewPalette(epoch)
			least = min(least, checkCursor(t, th.Key, p))
			_, fill, _ = p.Cursor.Decompose()
			if _, alt, _ := p.CursorAlt.Decompose(); theme.Distance(alt, fill) < 0.03 {
				t.Errorf("%s: the two halves of the blink, #%06x and #%06x, are too close to see it blink", th.Key, fill.Hex(), alt.Hex())
			}
			// The fill is one of the candidates: made from the theme's roles.
			made := false
			for _, c := range cursorCandidates() {
				made = made || c == fill
			}
			if !made {
				t.Errorf("%s: the fill #%06x is not one of the colours made from the theme's roles", th.Key, fill.Hex())
			}
		}
		r, g, b := fill.RGB()
		// A warm fill on a dark theme that has colour to make one from:
		// red leads, blue trails. (Monochrome has only greys; the two
		// neon themes' warning and negative are pinks.)
		if warn := theme.Color(theme.RoleWarning); !theme.IsLight() && isWarm(warn) && !(r > g && g > b) {
			t.Errorf("%s: the fill #%06x is not an orange, and the theme's warning colour #%06x could make one", th.Key, fill.Hex(), warn.Hex())
		}
		t.Logf("%-20s fill #%06x, at least %.3f from every glyph colour", th.Key, fill.Hex(), least)
	}
}

// isWarm reports whether c is a yellow, an orange or a red with some
// colour to it: red at or above green, green well above blue.
func isWarm(c tcell.Color) bool {
	r, g, b := c.RGB()
	return r >= g && g > b+40
}

// TestCursorBlinksSlowlyAndHoldsStill: the cursor's style follows the
// animation frame: the fill for CursorBlinkFrames frames, the lighter one
// for the next. Frame 0 is the fill, and frame 0 is every frame while the
// motion setting is off.
func TestCursorBlinksSlowlyAndHoldsStill(t *testing.T) {
	p := NewPalette(0)
	if p.CursorStyle(0) != p.Cursor {
		t.Error("with motion off (frame 0) the cursor is not the fill")
	}
	for anim := 0; anim < 4*CursorBlinkFrames; anim++ {
		want := p.Cursor
		if (anim/CursorBlinkFrames)%2 == 1 {
			want = p.CursorAlt
		}
		if got := p.CursorStyle(anim); got != want {
			t.Errorf("frame %d: the cursor is %v, want %v", anim, got, want)
		}
	}
	// Slow: each half lasts most of a second at eight frames a second.
	if CursorBlinkFrames < 5 || CursorBlinkFrames > 12 {
		t.Errorf("each half of the blink lasts %d frames, want between 5 and 12 (about a second at 8 a second)", CursorBlinkFrames)
	}
}
