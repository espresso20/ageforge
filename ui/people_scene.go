package ui

import (
	"image"
	"image/color"
	"image/draw"

	"github.com/espresso20/ageforge/pkg/peopledash"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/window/kit"
)

// people_scene.go is what the People panel draws a tab on: one picture of the
// tab's shapes and the words that sit over it. The picture runs from row y0
// to the row before the hint row; the frame above and below it is cells.

// peoplePalette is the panel's colours on the active theme: the group colours
// are the fixed ones, held clear of the tile.
func peoplePalette() peopledash.Palette {
	bg, d := theme.Of(theme.RoleBackground), peopledash.Dark
	tile := theme.Mix(bg, theme.Of(theme.RoleText), .1)
	leg := func(c color.RGBA) color.RGBA {
		return rgbOf(theme.Legible(theme.RGB(c.R, c.G, c.B), tile, 3))
	}
	return peopledash.Palette{
		Ground: rgbOf(bg), Tile: rgbOf(tile), Rule: rgbOf(theme.Mix(bg, theme.Of(theme.RoleText), .22)), Ink: rgbOf(bg),
		Text: rgbOf(theme.Legible(theme.Of(theme.RoleText), tile, 4.5)), Dim: rgbOf(theme.Legible(theme.Of(theme.RoleDim), tile, 3)),
		Work: leg(d.Work), Waiting: leg(d.Waiting), Soldiers: leg(d.Soldiers), Scholars: leg(d.Scholars), Homes: leg(d.Homes),
		Green: leg(d.Green), Amber: leg(d.Amber), Alarm: leg(d.Alarm), Red: leg(d.Red),
	}
}

// pWord is a word on the scene, at a cell.
type pWord struct {
	x, y int
	s    string
	col  color.RGBA
	bold bool
}

// pScene is a tab's picture and words. Cell rows are the panel's; the picture
// starts at row y0.
type pScene struct {
	w, y0, end int // the panel's cells across; the picture's first row and the row after its last
	cw, ch     int // pixels to a cell
	cv         *peopledash.Canvas
	pal        peopledash.Palette
	words      []pWord
}

func newPScene(w, y0, end int) *pScene {
	cw, ch := cellPixels()
	pal := peoplePalette()
	return &pScene{w: w, y0: y0, end: end, cw: cw, ch: ch, pal: pal, cv: peopledash.New((w-2)*cw, (end-y0)*ch, pal)}
}

// px is the pixel column of a cell: the picture starts inside the frame's edge.
func (s *pScene) px(c int) int { return (c - 1) * s.cw }

// py is the pixel row of a cell row.
func (s *pScene) py(r int) int { return (r - s.y0) * s.ch }

// rect paints whole cells; px paints pixels with the cell grid's origin.
func (s *pScene) rect(c, r, w, h int, col color.RGBA) {
	s.cv.Rect(s.px(c), s.py(r), w*s.cw, h*s.ch, col)
}

// tile is a flat panel of the given colour with a 6 px band of band along its top.
func (s *pScene) tile(c, r, w, h int, ground, band color.RGBA) {
	s.rect(c, r, w, h, ground)
	s.cv.Rect(s.px(c), s.py(r), w*s.cw, 6, band)
}

// tx puts a word at a cell; txR ends it at the cell before c+w.
func (s *pScene) tx(c, r int, t string, col color.RGBA, bold bool) {
	s.words = append(s.words, pWord{c, r, t, col, bold})
}

func (s *pScene) txR(c, r int, t string, col color.RGBA, bold bool) {
	s.tx(c-runeLen(t), r, t, col, bold)
}

// numType is the big type a numeral of the given scale is set in: bold ones
// in full strokes, thin ones in lighter.
func numType(bold bool, scale int) (n, wt int) {
	switch {
	case !bold:
		return 3, 2
	case scale >= 5:
		return 5, 5
	}
	return 3, 3
}

// figure is t as a numeral's picture, cut to its ink.
func figure(t string, col color.RGBA, bold bool, scale int) *image.RGBA {
	n, wt := numType(bold, scale)
	col.A = 255
	return bigNum(t, n, wt, col)
}

// setNum sets t with the top left of its digits at pixel x, y and returns
// the pixel after it.
func (s *pScene) setNum(x, y int, t string, col color.RGBA, bold bool, scale int) int {
	img := figure(t, col, bold, scale)
	draw.Draw(s.cv.Img, img.Rect.Add(image.Pt(x, y)), img, image.Point{}, draw.Over)
	return x + img.Rect.Dx()
}

// numW is the pixel width of a numeral.
func (s *pScene) numW(t string, bold bool, scale int) int {
	return figure(t, color.RGBA{}, bold, scale).Rect.Dx()
}

// num paints a numeral at a cell and returns the first free cell after it.
func (s *pScene) num(c, r int, t string, col color.RGBA, bold bool, scale int) int {
	return (s.setNum(s.px(c), s.py(r)+2, t, col, bold, scale)+s.cw+1)/s.cw + 2
}

// numH is the rows a numeral takes.
func (s *pScene) numH(bold bool, scale int) int {
	return (figure("0", color.RGBA{}, bold, scale).Rect.Dy() + 2 + s.ch - 1) / s.ch
}

// chipAt paints a one-row chip of word on col at a cell; the next free cell comes back.
func (s *pScene) chipAt(c, r int, word string, col color.RGBA) int {
	w := runeLen(word) + 2
	s.rect(c, r, w, 1, col)
	s.tx(c+1, r, word, s.pal.Ink, true)
	return c + w
}

// bar is a track one cell tall (a little less) with frac of it filled.
func (s *pScene) bar(c, r, w int, frac float64, col color.RGBA) {
	s.cv.Bar(s.px(c), s.py(r)+3, w*s.cw, s.ch-6, frac, col, s.pal.Ground)
}

// blocks lays n blocks of col across at most room cells, solid up to fill.
func (s *pScene) blocks(c, r, h, n, fill int, col, ground color.RGBA, room int) {
	pitch := min(14, max(1, room*s.cw/max(n, 1)))
	bw := max(1, pitch-2)
	s.cv.Blocks(s.px(c), s.py(r)+2, h*s.ch-4, n, fill, col, ground, bw, pitch)
}

// paint puts the picture, then the words, on the painter at org. Without
// pictures only the words are printed.
func (s *pScene) paint(p2 *kit.Painter, org image.Point, plain bool) {
	if !plain {
		p2.Picture(image.Rect(1, s.y0, s.w-1, s.end).Add(org), s.cv.Img)
	}
	for _, w := range s.words {
		st := kit.Style{}.Foreground(theme.RGB(w.col.R, w.col.G, w.col.B)).Bold(w.bold)
		if plain {
			p2.Print(org.X+w.x, org.Y+w.y, runeLen(w.s), kit.Escape(w.s), st)
			continue
		}
		p2.Label(org.X+w.x, org.Y+w.y, kit.Escape(w.s), st)
	}
}

// frame outlines a block of cells 2 px thick, inside its edge.
func (s *pScene) frame(c, r, w, h int, col color.RGBA) {
	x, y, pw, ph := s.px(c), s.py(r), w*s.cw, h*s.ch
	s.cv.Rect(x, y, pw, 2, col)
	s.cv.Rect(x, y+ph-2, pw, 2, col)
	s.cv.Rect(x, y, 2, ph, col)
	s.cv.Rect(x+pw-2, y, 2, ph, col)
}

// lift writes the grid's text over the picture: a tab that draws its words on
// the cell grid keeps them above the shapes the scene put under them.
func (s *pScene) lift(p2 *kit.Painter, g *xgrid, org image.Point, pal *expPalette) {
	for y := s.y0; y < s.end; y++ {
		for x := 1; x < s.w-1; x++ {
			if c := g.c[y*g.w+x]; c.r != ' ' && c.r != 0 {
				p2.Label(org.X+x, org.Y+y, kit.Escape(string(c.r)), pal.style(c))
			}
		}
	}
}

// pane is a tile of the scene with a thin line along its top, at cells.
func (s *pScene) pane(c, r, w, h int, col color.RGBA) {
	s.rect(c, r, w, h, s.pal.Tile)
	s.cv.Rect(s.px(c), s.py(r), w*s.cw, 2, col)
}
