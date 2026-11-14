package theme

import (
	"image/color"

	"github.com/gdamore/tcell/v2"
)

// rgba.go is the theme for code that paints without tcell (the window's UI
// kit, window/kit): roles and tag names as plain RGBA.

// RGBA is the active theme's color for role.
func RGBA(role Role) color.RGBA { return plainRGBA(Color(role)) }

func plainRGBA(c tcell.Color) color.RGBA {
	r, g, b := c.RGB()
	if r < 0 {
		return color.RGBA{A: 255}
	}
	return color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}
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
