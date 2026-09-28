package main

import (
	"github.com/gdamore/tcell/v2"
)

// Canvas is a grid of cells: one glyph, a foreground, a background and a
// priority each. One resolution for everything, so art and labels share the
// grid: a label never sits on a finer or coarser raster than the tree it
// names.
type Canvas struct {
	W, H  int
	cells []cell
}

type cell struct {
	ch  rune
	fg  tcell.Color
	bg  tcell.Color
	pri int8
	att tcell.AttrMask
}

func NewCanvas(w, h int, bg tcell.Color) *Canvas {
	c := &Canvas{W: w, H: h, cells: make([]cell, w*h)}
	for i := range c.cells {
		c.cells[i] = cell{ch: ' ', bg: bg, pri: -1}
	}
	return c
}

// Glyph puts ch at (x, y) unless something of higher priority is there.
func (c *Canvas) Glyph(x, y int, ch rune, fg tcell.Color, pri int8) {
	c.GlyphA(x, y, ch, fg, pri, 0)
}

func (c *Canvas) GlyphA(x, y int, ch rune, fg tcell.Color, pri int8, att tcell.AttrMask) {
	if x < 0 || y < 0 || x >= c.W || y >= c.H {
		return
	}
	ce := &c.cells[y*c.W+x]
	if pri >= ce.pri {
		ce.ch, ce.fg, ce.pri, ce.att = ch, fg, pri, att
	}
}

// Text writes a string; it returns the column after the last rune.
func (c *Canvas) Text(x, y int, s string, fg tcell.Color, pri int8, att tcell.AttrMask) int {
	for _, r := range s {
		c.GlyphA(x, y, r, fg, pri, att)
		x++
	}
	return x
}

// Free reports whether a run of n cells holds nothing at or above pri.
func (c *Canvas) Free(x, y, n int, pri int8) bool {
	if y < 0 || y >= c.H || x < 0 || x+n > c.W {
		return false
	}
	for i := 0; i < n; i++ {
		if c.cells[y*c.W+x+i].pri >= pri {
			return false
		}
	}
	return true
}

func (c *Canvas) PriAt(x, y int) int8 {
	if x < 0 || y < 0 || x >= c.W || y >= c.H {
		return 127
	}
	return c.cells[y*c.W+x].pri
}

func (c *Canvas) Bg(x, y int, col tcell.Color) {
	if x < 0 || y < 0 || x >= c.W || y >= c.H {
		return
	}
	c.cells[y*c.W+x].bg = col
}

func (c *Canvas) BgAt(x, y int) tcell.Color {
	if x < 0 || y < 0 || x >= c.W || y >= c.H {
		return tcell.ColorDefault
	}
	return c.cells[y*c.W+x].bg
}

// Blit writes the canvas to a screen at (ox, oy).
func (c *Canvas) Blit(s tcell.Screen, ox, oy int) {
	for y := 0; y < c.H; y++ {
		for x := 0; x < c.W; x++ {
			ce := c.cells[y*c.W+x]
			st := tcell.StyleDefault.Background(ce.bg).Foreground(ce.fg).Attributes(ce.att)
			s.SetContent(ox+x, oy+y, ce.ch, nil, st)
		}
	}
}

func round(f float64) int {
	if f < 0 {
		return int(f - 0.5)
	}
	return int(f + 0.5)
}

func abs(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}
