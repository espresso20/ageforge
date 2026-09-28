package main

import (
	"github.com/espresso20/ageforge/theme"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/uniseg"
)

// Colour classes. Every cell names a class, never a colour: the palette maps
// classes to the active theme's roles, so a theme switch is free and a light
// theme needs no re-keying. Meaning never rides on colour alone; each class
// that signals something also has a glyph that does.
type cls uint8

const (
	cText cls = iota
	cDim
	cFaint // quieter than Dim: idle wires, empty gauge cells
	cLabel
	cAccent
	cBorder
	cWire
	cHi
	cPos
	cNeg
	cWarn
	cBright
	numCls
)

type bgc uint8

const (
	bgNone bgc = iota
	bgSurface
	bgChip
	bgSel
	bgAlarm
	numBg
)

type cell struct {
	r    rune
	fg   cls
	bg   bgc
	bold bool
}

type canvas struct {
	w, h  int
	cells []cell
}

func newCanvas(w, h int) *canvas {
	c := &canvas{w: w, h: h, cells: make([]cell, w*h)}
	for i := range c.cells {
		c.cells[i] = cell{r: ' '}
	}
	return c
}

func (c *canvas) in(x, y int) bool { return x >= 0 && y >= 0 && x < c.w && y < c.h }

func (c *canvas) at(x, y int) *cell {
	if !c.in(x, y) {
		return nil
	}
	return &c.cells[y*c.w+x]
}

func (c *canvas) set(x, y int, r rune, fg cls) {
	if p := c.at(x, y); p != nil {
		p.r, p.fg = r, fg
	}
}

func (c *canvas) setB(x, y int, r rune, fg cls, bold bool) {
	if p := c.at(x, y); p != nil {
		p.r, p.fg, p.bold = r, fg, bold
	}
}

func (c *canvas) bgRect(x, y, w, h int, b bgc) {
	for yy := y; yy < y+h; yy++ {
		for xx := x; xx < x+w; xx++ {
			if p := c.at(xx, yy); p != nil {
				p.bg = b
			}
		}
	}
}

// text writes s at x,y clipped to max cells, returning cells used.
func (c *canvas) text(x, y int, s string, fg cls, max int) int {
	return c.textB(x, y, s, fg, max, false)
}

func (c *canvas) textB(x, y int, s string, fg cls, max int, bold bool) int {
	n := 0
	for _, r := range s {
		if n >= max {
			break
		}
		c.setB(x+n, y, r, fg, bold)
		n++
	}
	return n
}

// width is the display width of s in cells (one per rune for our glyphs).
func width(s string) int { return uniseg.StringWidth(s) }

// clip cuts s to at most n cells, marking the cut with an ellipsis.
func clip(s string, n int) string {
	if n <= 0 {
		return ""
	}
	rs := []rune(s)
	if len(rs) <= n {
		return s
	}
	if n == 1 {
		return "…"
	}
	return string(rs[:n-1]) + "…"
}

// palette resolves classes against a theme, with the epoch's tint mixed into
// the chrome (accent and borders) only.
type palette struct {
	fg [numCls]tcell.Color
	bg [numBg]tcell.Color
}

func newPalette(t theme.Theme, s style, mono bool) palette {
	var p palette
	bgc0 := t.Color(theme.RoleBackground)
	col := func(r theme.Role) tcell.Color { return t.Color(r) }
	tint := func(c tcell.Color) tcell.Color {
		if ap, ok := theme.AgePalettes[s.tintAge]; ok && s.tintAmount > 0 {
			c = theme.Mix(c, ap.Accent, s.tintAmount)
		}
		return theme.Legible(c, bgc0, 3.0)
	}
	p.fg[cText] = col(theme.RoleText)
	p.fg[cDim] = col(theme.RoleDim)
	p.fg[cFaint] = theme.Legible(theme.Mix(col(theme.RoleDim), bgc0, 0.35), bgc0, 1.6)
	p.fg[cLabel] = col(theme.RoleLabel)
	p.fg[cAccent] = tint(col(theme.RoleAccent))
	p.fg[cBorder] = tint(col(theme.RoleBorder))
	p.fg[cWire] = theme.Legible(theme.Mix(p.fg[cBorder], col(theme.RoleDim), 0.5), bgc0, 2.6)
	p.fg[cHi] = col(theme.RoleHighlight)
	p.fg[cPos] = col(theme.RolePositive)
	p.fg[cNeg] = col(theme.RoleNegative)
	p.fg[cWarn] = col(theme.RoleWarning)
	p.fg[cBright] = col(theme.RoleBright)
	p.bg[bgNone] = bgc0
	p.bg[bgSurface] = col(theme.RoleSurface)
	p.bg[bgChip] = col(theme.RoleChip)
	p.bg[bgSel] = col(theme.RoleSelection)
	p.bg[bgAlarm] = theme.Mix(bgc0, col(theme.RoleNegative), 0.22)
	if mono {
		for i := range p.fg {
			p.fg[i] = col(theme.RoleText)
		}
		for i := range p.bg {
			p.bg[i] = bgc0
		}
		p.bg[bgSel] = col(theme.RoleSelection)
	}
	return p
}

func (p palette) style(c cell) tcell.Style {
	st := tcell.StyleDefault.Foreground(p.fg[c.fg]).Background(p.bg[c.bg])
	if c.bold {
		st = st.Bold(true)
	}
	return st
}

// blit copies the canvas onto a tcell screen.
func (c *canvas) blit(s tcell.Screen, p palette) {
	for y := 0; y < c.h; y++ {
		for x := 0; x < c.w; x++ {
			cl := c.cells[y*c.w+x]
			s.SetContent(x, y, cl.r, nil, p.style(cl))
		}
	}
}
