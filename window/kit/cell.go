// Package kit is the game's own UI kit: widgets that lay themselves out on a
// grid of text cells and paint into a Buffer of cells. The window paints the
// cells in pixels and the terminal shows the same cells, so a screen built on
// the kit is written once.
//
// The kit knows nothing about the game, and it imports no terminal library:
// colors are theme roles resolved to plain RGBA when a cell is painted, so a
// theme switch is a repaint and nothing else.
//
// One goroutine owns a widget tree (the UI's). Nothing here locks.
package kit

import (
	"image"
	"image/color"
	"strings"

	"github.com/espresso20/ageforge/theme"
)

// Attr is how a cell's glyph is drawn.
type Attr uint8

const (
	Bold Attr = 1 << iota
	Italic
	Dim
	Underline
	Reverse
	Strike
	Blink
)

// A Color is late bound: a theme role, a literal, or nothing at all (the zero
// value), which leaves whatever is there. It is resolved against the active
// theme when a cell is painted.
type Color struct {
	kind, role uint8 // role is a theme.Role, kept small: text holds a Color per glyph
	r, g, b    uint8
}

const (
	colUnset uint8 = iota
	colRole
	colRGB
	colClear
)

// Role is the active theme's color for a role, whichever theme that is when
// the cell is painted.
func Role(r theme.Role) Color { return Color{kind: colRole, role: uint8(r)} }

// RGB is a literal color, for the hues that are not the theme's to choose.
func RGB(r, g, b uint8) Color { return Color{kind: colRGB, r: r, g: g, b: b} }

// Clear is a ground that is no color: a cell with it shows what the window
// draws under the cells (a map). It means nothing as a foreground.
var Clear = Color{kind: colClear}

// IsSet reports whether c is a color (or Clear) rather than nothing.
func (c Color) IsSet() bool { return c.kind != colUnset }

// rgba resolves c; ok is false for the zero Color.
func (c Color) rgba() (color.RGBA, bool) {
	switch c.kind {
	case colRole:
		return theme.RGBA(theme.Role(c.role)), true
	case colRGB:
		return color.RGBA{c.r, c.g, c.b, 255}, true
	case colClear:
		return color.RGBA{}, true
	}
	return color.RGBA{}, false
}

// Style is how text is painted. An unset Fg is the theme's text color; an
// unset Bg keeps the ground the cell already has.
type Style struct {
	Fg, Bg Color
	Attr   Attr
}

// Fg is a style with only a foreground.
func Fg(r theme.Role) Style { return Style{Fg: Role(r)} }

// A Cell is one square of the grid, in resolved colors.
type Cell struct {
	R      rune   // the glyph; ' ' for none
	Comb   string // what combines with R (marks, joiners), usually empty
	Fg, Bg color.RGBA
	Attr   Attr
	Wide   bool // R covers this cell and the next
	Cont   bool // the right half of a wide glyph: nothing is drawn here
}

// Clear reports whether the cell has no ground (see the Clear color).
func (c Cell) Clear() bool { return c.Bg.A == 0 }

// A Buffer is the grid a tree paints into.
type Buffer struct {
	W, H  int
	Cells []Cell
	// Cursor is where the text cursor shows, when CursorOn.
	Cursor   image.Point
	CursorOn bool
}

// NewBuffer is a blank buffer of w by h cells.
func NewBuffer(w, h int) *Buffer {
	b := &Buffer{}
	b.Resize(w, h)
	return b
}

// Resize makes the buffer w by h and blanks it.
func (b *Buffer) Resize(w, h int) {
	w, h = max(w, 0), max(h, 0)
	b.W, b.H = w, h
	if cap(b.Cells) < w*h {
		b.Cells = make([]Cell, w*h)
	}
	b.Cells = b.Cells[:w*h]
	b.Reset()
}

// Reset blanks every cell to the theme's text on its ground, and hides the
// cursor.
func (b *Buffer) Reset() {
	blank := Cell{R: ' ', Fg: theme.RGBA(theme.RoleText), Bg: theme.RGBA(theme.RoleBackground)}
	for i := range b.Cells {
		b.Cells[i] = blank
	}
	b.CursorOn = false
}

// At is the cell at x, y; a blank outside the buffer.
func (b *Buffer) At(x, y int) Cell {
	if x < 0 || y < 0 || x >= b.W || y >= b.H {
		return Cell{R: ' '}
	}
	return b.Cells[y*b.W+x]
}

// Line is row y as text, trailing blanks dropped.
func (b *Buffer) Line(y int) string {
	var s strings.Builder
	for x := 0; x < b.W; x++ {
		c := b.At(x, y)
		if c.Cont {
			continue
		}
		s.WriteRune(c.R)
		s.WriteString(c.Comb)
	}
	return strings.TrimRight(s.String(), " ")
}

// Lines is every row as text: what a test reads a screen by.
func (b *Buffer) Lines() []string {
	out := make([]string, b.H)
	for y := range out {
		out[y] = b.Line(y)
	}
	return out
}
