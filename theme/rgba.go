package theme

import (
	"image/color"

	"github.com/gdamore/tcell/v2"
)

// rgba.go is the theme for code that paints without tcell (the window's UI
// kit, window/kit): roles and tag names as plain RGBA.

// RGBA is the active theme's color for role.
func RGBA(role Role) color.RGBA { return plainRGBA(Of(role)) }

func plainRGBA(c Color) color.RGBA {
	if c.kind != colRGB {
		return color.RGBA{A: 255}
	}
	return color.RGBA{R: c.r, G: c.g, B: c.b, A: 255}
}

// FromRGBA is a plain color as tcell's, for the bridge that copies the kit's
// cells onto a tcell screen (ui/native.go) while there still is one.
func FromRGBA(c color.RGBA) tcell.Color {
	return tcell.NewRGBColor(int32(c.R), int32(c.G), int32(c.B))
}

// Tcell is c as tcell's color, for the same bridge and for tests that read a
// tcell screen back: a Ref as the color it is now, Transparent as
// TcellTransparent, and no color as tcell's default.
func Tcell(c Color) tcell.Color {
	switch c.kind {
	case colClear:
		return TcellTransparent
	case colUnset:
		return tcell.ColorDefault
	}
	rgba, _ := c.RGBA()
	return FromRGBA(rgba)
}

// fromTcell is a tcell color as the theme's: a literal, or no color.
func fromTcell(c tcell.Color) Color {
	r, g, b := c.RGB()
	if r < 0 {
		return ColorDefault
	}
	return NewRGBColor(r, g, b)
}

// tagRoles is TagNames, made once.
var tagRoles = TagNames()

// TagRole is the role an inline-tag color name stands for ("accent", or a
// legacy alias like "gold"); ok is false for a name the theme does not own.
func TagRole(name string) (Role, bool) {
	r, ok := tagRoles[name]
	return r, ok
}

// CSSRGBA is the literal color of a CSS color name ("blue", "orange"), as it
// was before any theme retinted the names it owns; ok is false for a name
// that is not a color.
func CSSRGBA(name string) (color.RGBA, bool) {
	c, ok := cssColors[name]
	if !ok {
		return color.RGBA{}, false
	}
	return plainRGBA(c), true
}
