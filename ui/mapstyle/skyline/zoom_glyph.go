package skyline

import (
	"strings"

	"github.com/espresso20/ageforge/mapmodel"
)

// zoom_glyph.go is how the close level (zoom.go) scales one world cell into
// its 2x2 block of screen cells, by what the glyph is.
//
// Repeating a glyph in all four cells is right for a surface or a stroke (a
// shade, a line, a solid block: the surface grows, the stroke thickens) and
// wrong for a feature. A face with its two eyes shows four faces, a heart
// four hearts, a dot a 2x2 clump of dots. So a cell becomes:
//
//   - a block glyph: exactly, each quadrant one solid cell
//   - a shade, a line or a bar: the same glyph in all four cells
//   - a lower eighth block: the same height at twice the scale
//   - a diagonal or a corner triangle: the one long diagonal across the block
//   - a round or small shape (a head, a dot, a window, a tree): the shape
//     redrawn at twice the size from the block's own quadrants (the Unicode
//     tiers: the ASCII tier has no quadrants)
//   - anything else (a face, a heart, a letter, a flag, a spark): the glyph
//     once, in the upper left of the block for what hangs or flies and in
//     the lower left for what stands; the other three cells show what is
//     behind it
//
// The tables below are the whole list: a glyph that is in none of them is a
// feature drawn once.

type glyphKind uint8

const (
	gOnce   glyphKind = iota // a feature: drawn once
	gRepeat                  // a surface or a stroke: drawn in all four cells
	gEighth                  // a lower eighth block
	gSlope                   // a diagonal
	gShape                   // a shape redrawn from quadrants
	gBlock                   // a block glyph: exact
)

// repeatGlyphs are the surfaces and strokes that scale by repeating, besides
// the box drawing range (not its diagonals).
const repeatGlyphs = "░▒▓■▬|-_=#[]()"

// onceUpper are the features that hang or fly: drawn in the upper left of
// their block.
const onceUpper = "\"'^*°►◄▸◂✦♥♡❤"

// slopes are what a diagonal or a corner triangle is in each cell of its
// block (upper left, upper right, lower left, lower right): empty (' '),
// solid ('█') or the glyph itself.
var slopes = map[rune][4]rune{
	'╱': {' ', '╱', '╱', ' '}, '╲': {'╲', ' ', ' ', '╲'},
	'/': {' ', '/', '/', ' '}, '\\': {'\\', ' ', ' ', '\\'},
	'◢': {' ', '◢', '◢', '█'}, '◣': {'◣', ' ', '█', '◣'},
	'◤': {'█', '◤', '◤', ' '}, '◥': {'◥', '█', ' ', '◥'},
}

// shapes are the shapes that are redrawn at twice the size, each a 4x4
// bitmap (the block's 2x2 cells of 2x2 quadrants) cut into the four cells'
// quadrant masks.
var shapes = func() map[rune][4]uint8 {
	disc := bitmap(".XX.", "XXXX", "XXXX", ".XX.")
	dot := bitmap("....", ".XX.", ".XX.", "....")
	return map[rune][4]uint8{
		'●': disc, '◘': disc,
		'•': dot, '∙': dot, '▪': dot,
		'·': bitmap("....", ".X..", "....", "...."),
		'▫': bitmap("XXXX", "X..X", "X..X", "XXXX"),
		'▲': bitmap(".XX.", ".XX.", "XXXX", "XXXX"),
		'◠': bitmap(".XX.", "XXXX", "....", "...."),
	}
}()

// bitmap cuts four rows of four pixels ('X' set) into the quadrant masks of
// the block's cells: upper left, upper right, lower left, lower right.
func bitmap(rows ...string) [4]uint8 {
	var out [4]uint8
	for q := range out {
		r0, c0 := q/2*2, q%2*2
		for k := 0; k < 4; k++ {
			if rows[r0+k/2][c0+k%2] == 'X' {
				out[q] |= 1 << k
			}
		}
	}
	return out
}

// glyphKindOf is how ch scales (see the table above).
func glyphKindOf(ch rune, tier mapmodel.GlyphTier) glyphKind {
	if _, ok := quadMask[ch]; ok {
		return gBlock
	}
	if _, ok := slopes[ch]; ok {
		return gSlope
	}
	if _, ok := shapes[ch]; ok && tier != mapmodel.TierASCII {
		return gShape
	}
	switch {
	case ch >= '▁' && ch <= '▇':
		return gEighth
	case ch >= 0x2500 && ch <= 0x257F, ch >= 0x2580 && ch <= 0x259F, strings.ContainsRune(repeatGlyphs, ch):
		return gRepeat
	}
	return gOnce
}

// zoomCell is the cell of the 2x2 block that quadrant q (0 upper left … 3
// lower right) of c becomes at the close level.
func zoomCell(c cell, q int, tier mapmodel.GlyphTier) cell {
	empty := cell{ch: ' ', fg: c.fg, bg: c.bg, d: c.d}
	switch glyphKindOf(c.ch, tier) {
	case gBlock:
		m := quadMask[c.ch]
		switch m {
		case 0:
			return empty
		case 15:
			return c
		}
		col := c.bg
		if m&(1<<q) != 0 {
			col = c.fg
		}
		return cell{ch: '█', fg: col, bg: col, d: c.d}
	case gRepeat:
		return c
	case gEighth:
		// a cell filled n eighths from its foot is a block 2n eighths of a
		// cell high: the lower cell fills first, the upper one has the rest
		n := 2 * int(c.ch-0x2580)
		f := min(8, n)
		if q < 2 {
			f = max(0, n-8)
		}
		switch f {
		case 0:
			return empty
		case 8:
			return cell{ch: '█', fg: c.fg, bg: c.fg, d: c.d}
		}
		cc := c
		cc.ch = rune(0x2580 + f)
		return cc
	case gSlope:
		switch g := slopes[c.ch][q]; g {
		case ' ':
			return empty
		case '█':
			return cell{ch: '█', fg: c.fg, bg: c.fg, d: c.d}
		default:
			cc := c
			cc.ch = g
			return cc
		}
	case gShape:
		switch m := shapes[c.ch][q]; m {
		case 0:
			return empty
		case 15:
			return cell{ch: '█', fg: c.fg, bg: c.fg, d: c.d}
		default:
			return cell{ch: quadRunes[m], fg: c.fg, bg: c.bg, d: c.d}
		}
	}
	if q == onceQuadrant(c.ch) {
		return c
	}
	return empty
}

// onceQuadrant is the cell of its block a feature glyph is drawn in: the
// upper left for what hangs or flies, the lower left for what stands.
func onceQuadrant(ch rune) int {
	if strings.ContainsRune(onceUpper, ch) {
		return 0
	}
	return 2
}
