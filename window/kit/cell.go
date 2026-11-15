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

// A Color is the theme's: late bound (a theme role), a literal, Clear, or
// nothing at all (the zero value), which leaves whatever is there. It is
// resolved against the active theme when a cell is painted.
type Color = theme.Color

// Role is the active theme's color for a role, whichever theme that is when
// the cell is painted.
func Role(r theme.Role) Color { return theme.Ref(r) }

// RGB is a literal color, for the hues that are not the theme's to choose.
func RGB(r, g, b uint8) Color { return theme.RGB(r, g, b) }

// Clear is a ground that is no color: a cell with it shows what the window
// draws under the cells (a map). It means nothing as a foreground.
var Clear = theme.Transparent

// Style is how text is painted. An unset Fg is the theme's text color; an
// unset Bg keeps the ground the cell already has.
type Style struct {
	Fg, Bg Color
	Attr   Attr
}

// Foreground is s with fg as its ink, and Background with bg as its ground.
func (s Style) Foreground(fg Color) Style { s.Fg = fg; return s }
func (s Style) Background(bg Color) Style { s.Bg = bg; return s }

// with is s with an attribute set or cleared.
func (s Style) with(a Attr, on bool) Style {
	if on {
		s.Attr |= a
	} else {
		s.Attr &^= a
	}
	return s
}

// Bold, Italic and Underline are s with that attribute on or off.
func (s Style) Bold(on bool) Style      { return s.with(Bold, on) }
func (s Style) Italic(on bool) Style    { return s.with(Italic, on) }
func (s Style) Underline(on bool) Style { return s.with(Underline, on) }

// Decompose is the style's parts.
func (s Style) Decompose() (fg, bg Color, attr Attr) { return s.Fg, s.Bg, s.Attr }

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
	// Pic is one more than the index, in the buffer's Pics, of the picture
	// whose pixels this cell shows; 0 for none. Whatever is painted on the
	// cell afterwards takes the cell back.
	Pic uint16
}

// A Pic is a picture the window draws in real pixels over a rectangle of
// cells, stretched to fill it. The grid only keeps its place: each cell of
// the rectangle that still shows the picture is marked with it (Cell.Pic).
// The image must not be written to once it has been painted; to change the
// picture, paint another image.
type Pic struct {
	Rect image.Rectangle // in cells
	Img  *image.RGBA
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
	// Pointer is the shape the mouse pointer takes where it is.
	Pointer Pointer
	// Pics are the pictures painted this frame (see Painter.Picture).
	Pics []Pic
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
	b.Pics = b.Pics[:0]
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
