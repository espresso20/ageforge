package mapstyle

import (
	"testing"

	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/window/kit"
)

// checkCursor holds one palette's cursor, both halves of its blink, to the
// margins every style's cursor keeps (CursorFromGlyph and the rest). where
// names the palette in a failure. It returns the fill's least distance from
// a glyph colour.
func checkCursor(t *testing.T, where string, p *Palette) float64 {
	t.Helper()
	for name, st := range map[string]kit.Style{"fill": p.Cursor, "blink": p.CursorAlt} {
		if _, _, attr := st.Decompose(); attr&kit.Bold == 0 {
			t.Errorf("%s: the cursor's %s is not bold", where, name)
		}
	}
	_, fill, _ := p.Cursor.Decompose()
	glyph, ground, over, text := p.CursorMargins()
	if glyph < CursorFromGlyph {
		t.Errorf("%s: the cursor's fill #%06x is %.3f from a colour the map draws a glyph in, want at least %.2f", where, fill.Hex(), glyph, CursorFromGlyph)
	}
	if ground < CursorFromGround || over < CursorOverGround {
		t.Errorf("%s: the cursor (fill #%06x) over the map's backgrounds: distance %.3f and contrast %.2f, want %.2f and %.1f", where, fill.Hex(), ground, over, CursorFromGround, CursorOverGround)
	}
	if text < CursorGlyphContrast {
		t.Errorf("%s: a glyph on the cursor (fill #%06x) has contrast %.2f, want %.1f", where, fill.Hex(), text, CursorGlyphContrast)
	}
	return glyph
}

// TestCursorStandsOut: in each era, the cursor is a highlighted cell whose
// fill is no colour the palette draws: clearly apart from every glyph colour
// and from every background under a cell, with the glyph on it readable. The
// blink's other half is a shade of the fill: it is held to the backgrounds
// and to the glyph on it. The fill is warm (an orange) and made from Forge's
// own roles, whatever theme is worn: the picture is drawn in Forge's colours.
func TestCursorStandsOut(t *testing.T) {
	forge, _ := theme.ByKey(theme.DefaultKey)
	least := 9.0
	var fill theme.Color
	for epoch := 0; epoch < 7; epoch++ {
		p := NewMapPalette(epoch)
		least = min(least, checkCursor(t, "forge", p))
		_, fill, _ = p.Cursor.Decompose()
		if _, alt, _ := p.CursorAlt.Decompose(); theme.Distance(alt, fill) < 0.03 {
			t.Errorf("the two halves of the blink, #%06x and #%06x, are too close to see it blink", fill.Hex(), alt.Hex())
		}
		made := false
		for _, c := range cursorCandidates(forge.Color) {
			made = made || c == fill
		}
		if !made {
			t.Errorf("the fill #%06x is not one of the colours made from Forge's roles", fill.Hex())
		}
	}
	r, g, b := fill.RGB()
	if warn := forge.Color(theme.RoleWarning); isWarm(warn) && !(r > g && g > b) {
		t.Errorf("the fill #%06x is not an orange, and Forge's warning colour #%06x could make one", fill.Hex(), warn.Hex())
	}
	t.Logf("fill #%06x, at least %.3f from every glyph colour", fill.Hex(), least)
}

// isWarm reports whether c is a yellow, an orange or a red with some
// colour to it: red at or above green, green well above blue.
func isWarm(c theme.Color) bool {
	r, g, b := c.RGB()
	return r >= g && g > b+40
}

// TestCursorBlinksSlowlyAndHoldsStill: the cursor's style follows the
// animation frame: the fill for CursorBlinkFrames frames, the lighter one
// for the next. Frame 0 is the fill, and frame 0 is every frame while the
// motion setting is off.
func TestCursorBlinksSlowlyAndHoldsStill(t *testing.T) {
	p := NewMapPalette(0)
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

// TestCursorOverAnyGround: over a ground the palette never drew, the cursor's
// fill (either half of its blink) stands clear of that ground, and the glyph
// on it is readable.
func TestCursorOverAnyGround(t *testing.T) {
	p := NewMapPalette(0)
	for level := 0; level <= 255; level += 15 {
		for _, tint := range [][3]int32{{0, 0, 0}, {40, 0, 0}, {0, 40, 0}, {0, 0, 40}} {
			ground := theme.NewRGBColor(int32(level)/2+tint[0], int32(level)/2+tint[1], int32(level)/2+tint[2])
			for _, anim := range []int{0, CursorBlinkFrames} {
				fg, fill, _ := p.CursorOver(ground, anim).Decompose()
				if d, cr := theme.Distance(fill, ground), theme.ContrastRatio(fill, ground); d < CursorFromGround || cr < CursorOverGround {
					t.Errorf("ground #%06x, frame %d: the cursor #%06x has distance %.3f and contrast %.2f", ground.Hex(), anim, fill.Hex(), d, cr)
				}
				if cr := theme.ContrastRatio(fg, fill); cr < CursorGlyphContrast {
					t.Errorf("ground #%06x, frame %d: the glyph on the cursor has contrast %.2f", ground.Hex(), anim, cr)
				}
			}
		}
	}
}
