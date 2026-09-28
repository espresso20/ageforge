package main

import (
	"math"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/uniseg"
)

// Canvas is one frame of the atlas as cells. Every layer writes glyphs with
// a z-order, so a river never paints over a town and a label never lands on
// a capital. Labels reserve their cells so later labels route around them.

type Cell struct {
	R    rune
	Fg   tcell.Color
	Bg   tcell.Color
	Attr tcell.AttrMask
	Z    int8
}

type Canvas struct {
	W, H int
	C    []Cell
	lab  []bool // label reservation
	mono bool   // monochrome proof: colours are dropped, glyphs remain
}

func newCanvas(w, h int, bg tcell.Color) *Canvas {
	c := &Canvas{W: w, H: h, C: make([]Cell, w*h), lab: make([]bool, w*h)}
	for i := range c.C {
		c.C[i] = Cell{R: ' ', Fg: bg, Bg: bg}
	}
	return c
}

func (c *Canvas) in(x, y int) bool { return x >= 0 && y >= 0 && x < c.W && y < c.H }

func (c *Canvas) At(x, y int) *Cell {
	if !c.in(x, y) {
		return nil
	}
	return &c.C[y*c.W+x]
}

// Put writes a glyph if z is at least the cell's current z.
func (c *Canvas) Put(x, y int, r rune, fg tcell.Color, z int8) bool {
	if !c.in(x, y) {
		return false
	}
	cell := &c.C[y*c.W+x]
	if z < cell.Z {
		return false
	}
	cell.R, cell.Fg, cell.Z = r, fg, z
	cell.Attr = 0
	return true
}

func (c *Canvas) PutA(x, y int, r rune, fg tcell.Color, attr tcell.AttrMask, z int8) {
	if c.Put(x, y, r, fg, z) {
		c.C[y*c.W+x].Attr = attr
	}
}

// Bg sets a cell's background (a wash), whatever its glyph.
func (c *Canvas) SetBg(x, y int, bg tcell.Color) {
	if c.in(x, y) {
		c.C[y*c.W+x].Bg = bg
	}
}

// Text writes a string at z, ignoring reservations (chrome).
func (c *Canvas) Text(x, y int, s string, fg tcell.Color, attr tcell.AttrMask, z int8) int {
	for _, r := range s {
		if c.Put(x, y, r, fg, z) {
			c.C[y*c.W+x].Attr = attr
		}
		x += uniseg.StringWidth(string(r))
	}
	return x
}

// Fill paints a box with a background and blanks (chrome panels).
func (c *Canvas) Fill(x0, y0, w, h int, bg tcell.Color, z int8) {
	for y := y0; y < y0+h; y++ {
		for x := x0; x < x0+w; x++ {
			if c.in(x, y) {
				cell := &c.C[y*c.W+x]
				if z >= cell.Z {
					*cell = Cell{R: ' ', Fg: bg, Bg: bg, Z: z}
				}
			}
		}
	}
}

// Label places s near (x, y) if a free spot exists among a few offsets;
// returns false if it could not be placed. halo keeps a blank cell on each
// side so text stays readable over busy terrain.
func (c *Canvas) Label(x, y int, s string, fg tcell.Color, attr tcell.AttrMask, z int8, offsets [][2]int, bounds [4]int) bool {
	n := uniseg.StringWidth(s)
	for _, o := range offsets {
		lx, ly := x+o[0], y+o[1]
		if o[0] == -999 { // centred
			lx, ly = x-n/2, y+o[1]
		}
		if o[0] == -998 { // left of the point
			lx, ly = x-n-1, y+o[1]
		}
		if lx-1 < bounds[0] || ly < bounds[1] || lx+n+1 > bounds[2] || ly >= bounds[3] {
			continue
		}
		ok := true
		for i := -1; i <= n; i++ {
			cx := lx + i
			if !c.in(cx, ly) || c.lab[ly*c.W+cx] || c.C[ly*c.W+cx].Z > z {
				ok = false
				break
			}
		}
		if !ok {
			continue
		}
		for i := -1; i <= n; i++ {
			c.lab[ly*c.W+lx+i] = true
		}
		// a thin halo: blank the cells just left and right of the word
		c.Put(lx-1, ly, ' ', fg, z)
		c.Put(lx+n, ly, ' ', fg, z)
		c.Text(lx, ly, s, fg, attr, z)
		return true
	}
	return false
}

// Reserve marks cells as taken for labels (marks, art).
func (c *Canvas) Reserve(x, y, w, h int) {
	for yy := y; yy < y+h; yy++ {
		for xx := x; xx < x+w; xx++ {
			if c.in(xx, yy) {
				c.lab[yy*c.W+xx] = true
			}
		}
	}
}

func (c *Canvas) Reserved(x, y int) bool { return c.in(x, y) && c.lab[y*c.W+x] }

// Line draws a polyline between screen points with a glyph chooser.
func (c *Canvas) Line(x0, y0, x1, y1 float64, pick func(i, x, y int, dx, dy float64) (rune, bool), fg tcell.Color, z int8) {
	dx, dy := x1-x0, y1-y0
	n := int(math.Max(math.Abs(dx), math.Abs(dy)*2)) + 1
	lastX, lastY := -1, -1
	for i := 0; i <= n; i++ {
		t := float64(i) / float64(n)
		x := int(math.Floor(x0 + dx*t))
		y := int(math.Floor(y0 + dy*t))
		if x == lastX && y == lastY {
			continue
		}
		lastX, lastY = x, y
		if r, ok := pick(i, x, y, dx, dy); ok {
			c.Put(x, y, r, fg, z)
		}
	}
}

// Stamp draws multi-line art centred on (x, y); spaces are transparent.
func (c *Canvas) Stamp(x, y int, art []string, fg tcell.Color, z int8) {
	h := len(art)
	w := 0
	for _, l := range art {
		if n := uniseg.StringWidth(l); n > w {
			w = n
		}
	}
	x0, y0 := x-w/2, y-h/2
	for j, l := range art {
		xx := x0
		for _, r := range l {
			if r != ' ' {
				c.Put(xx, y0+j, r, fg, z)
				c.lab[(y0+j)*c.W+xx] = c.in(xx, y0+j)
			}
			xx++
		}
	}
}

// dirGlyph picks a line glyph for a direction (cells are twice as tall as
// wide, so dy counts double).
func dirGlyph(dx, dy float64, h, v, d1, d2 rune) rune {
	a := math.Atan2(dy*2, dx)
	if a < 0 {
		a += math.Pi
	}
	switch {
	case a < math.Pi/8 || a >= 7*math.Pi/8:
		return h
	case a < 3*math.Pi/8:
		return d2 // down-right on screen: ╲
	case a < 5*math.Pi/8:
		return v
	}
	return d1
}

// marching-squares outline glyphs, bits TL=8 TR=4 BR=2 BL=1
var msRound = [16]rune{0, '╮', '╭', '─', '╰', '│', '│', '╯', '╯', '│', '│', '╰', '─', '╭', '╮', 0}
var msSquare = [16]rune{0, '┐', '┌', '─', '└', '│', '│', '┘', '┘', '│', '│', '└', '─', '┌', '┐', 0}
