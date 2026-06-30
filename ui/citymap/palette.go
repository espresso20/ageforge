package citymap

import (
	"image/color"

	"github.com/espresso20/ageforge/theme"
	"github.com/gdamore/tcell/v2"
)

// palette.go derives every map color from the ACTIVE THEME at render time, so a
// theme switch retints the whole map. Nothing here is hard-coded except the
// notion of "water leans blue" — and even that is blended against theme roles so
// it stays in-family with the palette.

// rgba converts a tcell.Color (true-RGB in this codebase) to an image/color.RGBA.
func rgba(c tcell.Color) color.RGBA {
	r, g, b := c.RGB()
	return color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 0xff}
}

// blend linearly mixes a toward b by t in [0,1].
func blend(a, b color.RGBA, t float64) color.RGBA {
	if t <= 0 {
		return a
	}
	if t >= 1 {
		return b
	}
	lerp := func(x, y uint8) uint8 {
		return uint8(float64(x) + (float64(y)-float64(x))*t)
	}
	return color.RGBA{
		R: lerp(a.R, b.R),
		G: lerp(a.G, b.G),
		B: lerp(a.B, b.B),
		A: 0xff,
	}
}

// shiftHue nudges a color's channel balance to give each age a faint, distinct
// feel on top of the shared theme tint. Kept deliberately subtle in P1 (rich
// per-age styling is P2): warm ages lean red/green, cool/late ages lean blue.
// amt is a small signed fraction; positive warms, negative cools.
func shiftHue(c color.RGBA, amt float64) color.RGBA {
	adj := func(v uint8, d float64) uint8 {
		f := float64(v) + d
		if f < 0 {
			f = 0
		}
		if f > 255 {
			f = 255
		}
		return uint8(f)
	}
	// Warm: + red, - blue. Cool: - red, + blue. Green nudged half-strength.
	return color.RGBA{
		R: adj(c.R, amt*26),
		G: adj(c.G, amt*10),
		B: adj(c.B, -amt*26),
		A: 0xff,
	}
}

// terrainPalette is the set of elevation-band colors plus marker colors for one
// render. Built fresh from the active theme + age each frame, so it always
// reflects the live palette.
type terrainPalette struct {
	deepWater    color.RGBA
	shallowWater color.RGBA
	lowland      color.RGBA
	midland      color.RGBA
	highland     color.RGBA
	peak         color.RGBA

	// Structure / marker colors.
	palace   color.RGBA
	building color.RGBA
	shadow   color.RGBA
}

// buildPalette derives all band colors from the active theme palette, then
// applies a faint per-age hue shift. Role mapping:
//
//	RoleBackground → the canvas land base (everything blends from here).
//	RoleDim        → darker land + water depth + shadow.
//	RoleText       → lighter land + highland brightening.
//	RoleAccent     → peak highlight + the palace marker.
//	RoleHighlight  → building markers.
//
// Water is RoleBackground/RoleDim blended toward a theme-neutral blue so it
// stays cohesive with the palette while still reading as water.
func buildPalette(ageShift float64) terrainPalette {
	bg := rgba(theme.Color(theme.RoleBackground))
	dim := rgba(theme.Color(theme.RoleDim))
	text := rgba(theme.Color(theme.RoleText))
	accent := rgba(theme.Color(theme.RoleAccent))
	highlight := rgba(theme.Color(theme.RoleHighlight))

	// A blue anchor for water, kept muted so light themes don't get a cartoon sea.
	blue := color.RGBA{R: 0x20, G: 0x4a, B: 0x86, A: 0xff}

	// Land bands: walk from a shade darker than background up toward text.
	land := blend(bg, dim, 0.25) // base ground, slightly grounded toward dim
	low := blend(land, dim, 0.18)
	mid := blend(land, text, 0.16)
	high := blend(land, text, 0.34)
	peak := blend(high, accent, 0.30)

	// Water bands: background pulled toward blue, deepened with dim.
	shallow := blend(blend(bg, blue, 0.45), dim, 0.10)
	deep := blend(blend(bg, blue, 0.62), dim, 0.30)

	p := terrainPalette{
		deepWater:    deep,
		shallowWater: shallow,
		lowland:      low,
		midland:      mid,
		highland:     high,
		peak:         peak,
		palace:       accent,
		building:     highlight,
		shadow:       blend(dim, bg, 0.35),
	}

	// Faint per-age feel on the terrain bands only (markers keep pure role color
	// so they stay legible across ages).
	if ageShift != 0 {
		p.deepWater = shiftHue(p.deepWater, ageShift)
		p.shallowWater = shiftHue(p.shallowWater, ageShift)
		p.lowland = shiftHue(p.lowland, ageShift)
		p.midland = shiftHue(p.midland, ageShift)
		p.highland = shiftHue(p.highland, ageShift)
		p.peak = shiftHue(p.peak, ageShift)
	}
	return p
}
