package mapstyle

import (
	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
)

// The inspect cursor.
//
// The cursor is a highlighted cell: a solid fill, with the glyph of the
// cell beneath it drawn on top in the theme's black or white, whichever
// reads better, so it is readable over every kind of cell. The fill is a
// colour the map does not draw in the theme. On the dark themes that is an
// orange. It is never a colour typed here: the candidates are made from the
// theme's own roles (its warning colour pulled toward its negative, plain
// and lifted toward the brightest text, then the roles a theme without such
// a hue falls back on), and of those that stand out from the map's
// background the fill is the one furthest (theme.Distance) from every
// colour of the palette, glyphs and backgrounds alike.
//
// The cursor blinks slowly between the fill and a brighter shade of it (a
// darker one on a light theme). The blink
// follows the animation frame, which the motion setting holds still: with
// motion off the cursor holds the fill.

// CursorBlinkFrames is how many animation frames each half of the cursor's
// blink lasts. The maps animate at about eight frames a second.
const CursorBlinkFrames = 6

const (
	// cursorMinContrast is the least contrast a fill may have against the
	// map's background.
	cursorMinContrast = 4.5
)

// cursorBlinkStep is how far the blink's other half moves the fill toward
// the brightest colour: enough to see, not so far that it reads as another
// colour.
const cursorBlinkStep = 0.3

// cursorCandidates lists the colours a fill is chosen from, all made from
// the active theme's roles: first the warm ones (an orange where the theme
// has one), then the roles that stand in for it elsewhere.
func cursorCandidates() []tcell.Color {
	warn, neg, bright := theme.Color(theme.RoleWarning), theme.Color(theme.RoleNegative), theme.Color(theme.RoleBright)
	var out []tcell.Color
	for _, t := range []float64{0.5, 0.6, 0.75, 0.4, 0.25} {
		base := theme.Mix(warn, neg, t)
		for _, lift := range []float64{0, 0.2, 0.4} {
			out = append(out, theme.Mix(base, bright, lift))
		}
	}
	for _, r := range []theme.Role{theme.RoleHighlight, theme.RoleAccent, theme.RoleWarning, theme.RoleNegative, theme.RoleBright, theme.RoleText} {
		out = append(out, theme.Color(r))
	}
	return out
}

// Paints lists every colour the palette draws the map in: each glyph class
// and each background. The cursor's fill is chosen against them.
func (p *Palette) Paints() []tcell.Color {
	out := make([]tcell.Color, 0, int(mapmodel.NumClasses)+4)
	for c := mapmodel.Class(1); c < mapmodel.NumClasses; c++ {
		out = append(out, p.Fg[c])
	}
	return append(out, p.Backgrounds()...)
}

// Backgrounds lists the colours the palette puts under a cell.
func (p *Palette) Backgrounds() []tcell.Color {
	return []tcell.Color{p.Bg, p.WaterBg, p.FreshBg, p.FlowBg}
}

// FitCursor chooses the cursor's styles for the palette as it stands. A
// style that changes the palette's colours after NewPalette (a night city)
// calls it again.
func (p *Palette) FitCursor() {
	paints := p.Paints()
	apart := func(c tcell.Color) float64 {
		least := -1.0
		for _, m := range paints {
			if d := theme.Distance(c, m); least < 0 || d < least {
				least = d
			}
		}
		return least
	}
	fill, best, found := theme.Color(theme.RoleText), -1.0, false
	for _, c := range cursorCandidates() {
		stands := theme.ContrastRatio(c, p.Bg) >= cursorMinContrast
		// The first candidate that stands out is taken; a later one must
		// be clearly further from the map to replace it, so the warm ones
		// at the head of the list win a near tie.
		if d := apart(c); stands && (!found || d > best*1.05) {
			fill, best, found = c, d, true
		}
	}
	// The blink's other half is the same fill a step brighter (darker on a
	// light theme): further from the background still.
	alt := theme.Mix(fill, theme.BestOn(p.Bg), cursorBlinkStep)
	p.Cursor = tcell.StyleDefault.Foreground(theme.BestOn(fill)).Background(fill).Bold(true)
	p.CursorAlt = tcell.StyleDefault.Foreground(theme.BestOn(alt)).Background(alt).Bold(true)
}

// CursorStyle is the cursor's style at animation frame anim: the fill, and
// every other CursorBlinkFrames frames its other shade. Frame 0 is the
// fill, which is the frame the motion setting holds.
func (p *Palette) CursorStyle(anim int) tcell.Style {
	if anim > 0 && (anim/CursorBlinkFrames)%2 == 1 {
		return p.CursorAlt
	}
	return p.Cursor
}

// CursorMargins measures how far the cursor stands from the palette, for
// the tests that hold every theme, age and style to the same margins:
// glyph is the least distance (theme.Distance) from the fill to any colour
// the palette draws a glyph in; ground the least distance, and over the
// least contrast, from either half of the blink to any background under a
// cell; text the least contrast of the glyph drawn on either half.
func (p *Palette) CursorMargins() (glyph, ground, over, text float64) {
	glyph, ground, over, text = 9, 9, 99, 99
	for i, st := range []tcell.Style{p.Cursor, p.CursorAlt} {
		fg, fill, _ := st.Decompose()
		text = min(text, theme.ContrastRatio(fg, fill))
		if i == 0 {
			for c := mapmodel.Class(1); c < mapmodel.NumClasses; c++ {
				glyph = min(glyph, theme.Distance(fill, p.Fg[c]))
			}
		}
		for _, bg := range p.Backgrounds() {
			ground = min(ground, theme.Distance(fill, bg))
			over = min(over, theme.ContrastRatio(fill, bg))
		}
	}
	return glyph, ground, over, text
}

// The margins every palette's cursor keeps (CursorMargins), which the
// tests of each style hold it to.
const (
	// CursorFromGlyph: five times what is just noticeable side by side.
	// No glyph is the cursor's colour. CursorFromGlyphGrey is the margin
	// on a theme with no hue in its roles (Monochrome), where a palette of
	// greys has only lightness to tell colours apart by and the fill is
	// what marks the cursor (CursorGlyphMargin picks between them).
	CursorFromGlyph     = 0.05
	CursorFromGlyphGrey = 0.03
	// CursorFromGround and CursorOverGround: the highlight never melts
	// into what it covers.
	CursorFromGround = 0.25
	CursorOverGround = 3.0
	// CursorGlyphContrast: the glyph on the highlight is readable.
	CursorGlyphContrast = 4.5
)

// CursorGlyphMargin is the least distance the cursor's fill keeps from the
// colours glyphs are drawn in, on the active theme: CursorFromGlyph, or
// CursorFromGlyphGrey on a theme whose roles are all greys.
func CursorGlyphMargin() float64 {
	for r := theme.Role(0); int(r) < theme.NumRoles; r++ {
		if red, green, blue := theme.Color(r).RGB(); red != green || green != blue {
			return CursorFromGlyph
		}
	}
	return CursorFromGlyphGrey
}
