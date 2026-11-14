package theme

import "image/color"

// A Color is a colour as the game hands one round: a literal RGB, a role of
// the active theme that is looked up when a cell is painted (Ref), a ground
// that is no colour (Transparent), or nothing at all (the zero value,
// ColorDefault): the theme's own text or ground, or "what is already
// there", as whoever takes it says. Colors compare with ==.
//
// It is the UI kit's color too (kit.Color is this type), kept small: text
// holds one per glyph.
type Color struct {
	kind, role uint8
	r, g, b    uint8
}

const (
	colUnset uint8 = iota
	colRole
	colRGB
	colClear
)

// ColorDefault is no color.
var ColorDefault = Color{}

// Transparent is a ground that is no color: a cell with it shows what the
// window has drawn under the cells (a map image). It means nothing as ink.
var Transparent = Color{kind: colClear}

// RGB is a literal color.
func RGB(r, g, b uint8) Color { return Color{kind: colRGB, r: r, g: g, b: b} }

// NewRGBColor is RGB for parts held as wider integers.
func NewRGBColor(r, g, b int32) Color { return RGB(uint8(r), uint8(g), uint8(b)) }

// NewHexColor is the literal color 0xRRGGBB.
func NewHexColor(v int32) Color { return RGB(uint8(v>>16), uint8(v>>8), uint8(v)) }

// Ref is the active theme's color for role, whichever theme that is when
// the cell is painted: a theme switch is a repaint and nothing else.
func Ref(role Role) Color {
	if role < 0 || role >= numRoles {
		return ColorDefault
	}
	return Color{kind: colRole, role: uint8(role)}
}

// RefRole reports the role a Ref stands for; ok is false for anything else.
func RefRole(c Color) (Role, bool) { return Role(c.role), c.kind == colRole }

// IsSet reports whether c is a color (or Transparent) rather than nothing.
func (c Color) IsSet() bool { return c.kind != colUnset }

// RGB is a literal color's parts; all three are -1 for anything that is not
// a literal (a Ref included: it has no color until it is painted).
func (c Color) RGB() (r, g, b int32) {
	if c.kind != colRGB {
		return -1, -1, -1
	}
	return int32(c.r), int32(c.g), int32(c.b)
}

// Hex is a literal color as 0xRRGGBB, and -1 for anything else.
func (c Color) Hex() int32 {
	if c.kind != colRGB {
		return -1
	}
	return int32(c.r)<<16 | int32(c.g)<<8 | int32(c.b)
}

// RGBA resolves c against the active theme; ok is false for no color. A
// Transparent comes out with no alpha.
func (c Color) RGBA() (color.RGBA, bool) {
	switch c.kind {
	case colRole:
		return RGBA(Role(c.role)), true
	case colRGB:
		return color.RGBA{R: c.r, G: c.g, B: c.b, A: 255}, true
	case colClear:
		return color.RGBA{}, true
	}
	return color.RGBA{}, false
}
