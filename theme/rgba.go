package theme

import (
	"image/color"
)

// rgba.go is the theme as plain RGBA, for the UI kit (window/kit), which
// paints with it: roles and tag names.

// RGBA is the active theme's color for role.
func RGBA(role Role) color.RGBA { return plainRGBA(Of(role)) }

func plainRGBA(c Color) color.RGBA {
	if c.kind != colRGB {
		return color.RGBA{A: 255}
	}
	return color.RGBA{R: c.r, G: c.g, B: c.b, A: 255}
}

// tagRoles is TagNames, made once.
var tagRoles = TagNames()

// TagRole is the role an inline-tag color name stands for ("accent", or a
// legacy alias like "gold"); ok is false for a name the theme does not own.
func TagRole(name string) (Role, bool) {
	r, ok := tagRoles[name]
	return r, ok
}

// CSSNames is every CSS color name an inline tag may carry.
func CSSNames() []string {
	out := make([]string, 0, len(cssNames))
	for k := range cssNames {
		out = append(out, k)
	}
	return out
}

// CSSRGBA is the literal color of a CSS color name ("blue", "orange"), as it
// was before any theme retinted the names it owns; ok is false for a name
// that is not a color.
func CSSRGBA(name string) (color.RGBA, bool) {
	v, ok := cssNames[name]
	if !ok {
		return color.RGBA{}, false
	}
	return plainRGBA(NewHexColor(v)), true
}
