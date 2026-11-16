package ui

import (
	"image"

	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/window/kit"
)

// rows.go is how a clickable row is lit, for every list on a menu page and
// the menu's own entries. A row under the pointer is a lit bar across its
// width with a chevron at its right end; the picked row (the one the
// keyboard is on) is lit the same, a little stronger, with a bar of the
// accent at its left edge. The two are told apart by the bar and the chevron
// as much as by the light, so they differ in a theme whose accent is
// close to its ground.

const (
	rowHoverMix  = .16
	rowPickedMix = .26
)

// lightRow lights the row whose cells are row: what is written on it keeps
// its colors, and the cells whose ground is the row's own take the lit
// ground. chevron is the column of the hover's chevron; the left bar is the
// row's first column, which the row leaves blank. plain draws both without
// glyphs a plain glyph set lacks.
func lightRow(p *kit.Painter, row image.Rectangle, picked, hover bool, chevron int, plain bool) {
	if row.Empty() || (!picked && !hover) {
		return
	}
	base := p.At(row.Min.X, row.Min.Y).Bg
	accent := rgbOf(theme.Of(theme.RoleAccent))
	mix := rowHoverMix
	if picked {
		mix = rowPickedMix
	}
	lit := toward(base, accent, mix)
	p.Reground(row, kit.RGB(base.R, base.G, base.B), kit.RGB(lit.R, lit.G, lit.B))
	y := row.Min.Y
	if picked {
		bar := kit.Style{Fg: kit.RGB(accent.R, accent.G, accent.B), Bg: kit.RGB(lit.R, lit.G, lit.B)}
		glyph := '▎'
		if plain {
			glyph, bar.Bg = ' ', bar.Fg
		}
		p.Put(row.Min.X, y, glyph, bar)
	}
	if hover && chevron > row.Min.X && chevron < row.Max.X {
		glyph := '›'
		if plain {
			glyph = '>'
		}
		p.Put(chevron, y, glyph, kit.Style{Fg: theme.Of(theme.RoleBright), Bg: kit.RGB(lit.R, lit.G, lit.B)}.Bold(true))
	}
}
