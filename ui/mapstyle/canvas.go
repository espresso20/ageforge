package mapstyle

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/uniseg"

	"github.com/espresso20/ageforge/mapmodel"
)

// Canvas writes into one rectangle of a screen: it clips to the rectangle,
// folds every rune to the frame's glyph tier and never writes a wide rune
// (a double-width cell would shift the rest of the row). Coordinates are
// relative to the rectangle.
type Canvas struct {
	Scr  tcell.Screen
	R    Rect
	Tier mapmodel.GlyphTier

	wide map[rune]bool // width cache for non-ASCII runes
}

// NewCanvas clears r to st and returns a canvas on it.
func NewCanvas(scr tcell.Screen, r Rect, tier mapmodel.GlyphTier, st tcell.Style) *Canvas {
	c := &Canvas{Scr: scr, R: r, Tier: tier}
	c.Fill(0, 0, r.W, r.H, ' ', st)
	return c
}

// In reports whether (x, y) is inside the canvas.
func (c *Canvas) In(x, y int) bool { return x >= 0 && y >= 0 && x < c.R.W && y < c.R.H }

// Put writes one cell.
func (c *Canvas) Put(x, y int, r rune, st tcell.Style) {
	if !c.In(x, y) {
		return
	}
	r = mapmodel.Fold(r, c.Tier)
	if r == 0 || r < ' ' {
		r = ' '
	} else if r >= 0x80 && c.isWide(r) {
		r = '?'
	}
	c.Scr.SetContent(c.R.X+x, c.R.Y+y, r, nil, st)
}

func (c *Canvas) isWide(r rune) bool {
	if r < 0x2190 { // Latin, Greek and the like: always one cell
		return false
	}
	if c.wide == nil {
		c.wide = map[rune]bool{}
	}
	w, ok := c.wide[r]
	if !ok {
		w = uniseg.StringWidth(string(r)) != 1
		c.wide[r] = w
	}
	return w
}

// Sym writes a model symbol in the canvas's tier.
func (c *Canvas) Sym(x, y int, s mapmodel.Sym, st tcell.Style) {
	c.Put(x, y, mapmodel.R(s, c.Tier), st)
}

// Text writes s from (x, y), at most max cells (max <= 0: to the edge), and
// returns the column after it.
func (c *Canvas) Text(x, y, max int, s string, st tcell.Style) int {
	if max <= 0 {
		max = c.R.W - x
	}
	n := 0
	for _, r := range s {
		if n >= max {
			break
		}
		c.Put(x+n, y, r, st)
		n++
	}
	return x + n
}

// Fill paints a rectangle with one rune.
func (c *Canvas) Fill(x, y, w, h int, r rune, st tcell.Style) {
	for yy := y; yy < y+h; yy++ {
		for xx := x; xx < x+w; xx++ {
			c.Put(xx, yy, r, st)
		}
	}
}

// Get reads back a cell (relative coordinates).
func (c *Canvas) Get(x, y int) (rune, tcell.Style) {
	if !c.In(x, y) {
		return ' ', tcell.StyleDefault
	}
	r, _, st, _ := c.Scr.GetContent(c.R.X+x, c.R.Y+y)
	return r, st
}

// TextLen is the number of cells s takes.
func TextLen(s string) int {
	n := 0
	for range s {
		n++
	}
	return n
}
