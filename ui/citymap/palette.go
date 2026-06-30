package citymap

import (
	"image/color"
	"math"

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

	// Biome fills (P4 biome generator). Each is blended from theme roles toward a
	// biome-characteristic hue so the whole biome map still retints on a theme
	// switch. The elevation bands above are retained for the flourishes (smoke/
	// starfields key off peak/highland/shadow); the biome fills drive the terrain.
	bDeepWater    color.RGBA
	bShallowWater color.RGBA
	bSand         color.RGBA
	bGrass        color.RGBA
	bForest       color.RGBA
	bRock         color.RGBA
	bMountain     color.RGBA
	bSnow         color.RGBA

	// Structure / marker colors.
	palace   color.RGBA
	building color.RGBA
	shadow   color.RGBA

	// Road color: a desaturated path tone derived from RoleDim pulled toward
	// RoleText, so roads sit visibly above the terrain but below the buildings.
	road color.RGBA

	// Trade-lane color: a bright positive/highlight tint for the dashed connector
	// lines from the city to the civ-edge markers, so a lane reads as a live trade
	// route, distinct from the quieter solid roads. Pulled from RolePositive so it
	// retints with the theme.
	tradeLane color.RGBA
}

// rotateHue rotates a color around the HSL hue wheel by deg degrees, preserving
// (approximately) its saturation and lightness. Used to fan a small set of theme
// role bases into one distinct-but-in-family color per production lineage: there
// are more lineages than roles, so each lineage gets a role base plus a stable
// hue rotation. Saturation/lightness from the theme are preserved, so the result
// still retints when the theme changes.
func rotateHue(c color.RGBA, deg float64) color.RGBA {
	h, s, l := rgbToHSL(c)
	h += deg / 360.0
	h -= math.Floor(h) // wrap into [0,1)
	return hslToRGB(h, s, l)
}

// rgbToHSL converts an 8-bit RGBA to HSL with each component in [0,1].
func rgbToHSL(c color.RGBA) (h, s, l float64) {
	r := float64(c.R) / 255.0
	g := float64(c.G) / 255.0
	b := float64(c.B) / 255.0
	maxc := math.Max(r, math.Max(g, b))
	minc := math.Min(r, math.Min(g, b))
	l = (maxc + minc) / 2
	d := maxc - minc
	if d == 0 {
		return 0, 0, l // achromatic
	}
	if l > 0.5 {
		s = d / (2 - maxc - minc)
	} else {
		s = d / (maxc + minc)
	}
	switch maxc {
	case r:
		h = (g - b) / d
		if g < b {
			h += 6
		}
	case g:
		h = (b-r)/d + 2
	default:
		h = (r-g)/d + 4
	}
	h /= 6
	return h, s, l
}

// hslToRGB converts HSL (each in [0,1]) back to an opaque 8-bit RGBA.
func hslToRGB(h, s, l float64) color.RGBA {
	if s == 0 {
		v := uint8(l*255 + 0.5)
		return color.RGBA{R: v, G: v, B: v, A: 0xff}
	}
	var q float64
	if l < 0.5 {
		q = l * (1 + s)
	} else {
		q = l + s - l*s
	}
	p := 2*l - q
	conv := func(t float64) uint8 {
		if t < 0 {
			t += 1
		}
		if t > 1 {
			t -= 1
		}
		switch {
		case t < 1.0/6.0:
			t = p + (q-p)*6*t
		case t < 1.0/2.0:
			t = q
		case t < 2.0/3.0:
			t = p + (q-p)*(2.0/3.0-t)*6
		default:
			t = p
		}
		return uint8(t*255 + 0.5)
	}
	return color.RGBA{R: conv(h + 1.0/3.0), G: conv(h), B: conv(h - 1.0/3.0), A: 0xff}
}

// brighten lightens a color toward white by t in [0,1] (used for lit roofs).
func brighten(c color.RGBA, t float64) color.RGBA {
	return blend(c, color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}, t)
}

// darken deepens a color toward black by t in [0,1] (used for shaded walls).
func darken(c color.RGBA, t float64) color.RGBA {
	return blend(c, color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0xff}, t)
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
	positive := rgba(theme.Color(theme.RolePositive))

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

	// ---- Biome fills (P4) -------------------------------------------------------
	// Each biome is a theme role pulled toward a biome-characteristic hue, so the
	// whole biome map retints with the theme but still reads as land/water/forest/
	// rock/snow/sand. The hues are muted anchors blended at modest strength so light
	// themes don't get a cartoon palette.
	green := color.RGBA{R: 0x35, G: 0x7d, B: 0x3a, A: 0xff}      // foliage anchor
	sandAnchor := color.RGBA{R: 0xcf, G: 0xb8, B: 0x82, A: 0xff} // warm light sand

	// Water reuses the elevation water tones (already blue-anchored + dim-deepened).
	bDeepWater := deep
	bShallowWater := shallow
	// Sand/beach: a warm light land tone — background lifted toward text, then warmed.
	bSand := blend(blend(bg, text, 0.30), sandAnchor, 0.40)
	// Grassland: a lighter land tone with a gentle green cast.
	bGrass := blend(blend(land, text, 0.12), green, 0.18)
	// Forest: background grounded toward dim, then pushed green — the darkest land.
	bForest := blend(blend(bg, dim, 0.30), green, 0.34)
	// Hills / rock: midland pulled toward a dim gray.
	bRock := blend(mid, dim, 0.30)
	// Mountain: high ground deepened toward dim gray (bare stone).
	bMountain := blend(high, dim, 0.40)
	// Snow / peak: pushed toward the bright text role with a touch of accent sparkle.
	bSnow := blend(blend(high, text, 0.62), accent, 0.08)

	// Road: a desaturated path tone. Start from RoleDim and pull a third of the
	// way toward RoleText so paths read clearly above the land but stay quieter
	// than the lit building roofs that sit on top of them.
	road := blend(dim, text, 0.33)

	// Trade lane: the positive role brightened a touch so the dashed lanes pop above
	// the terrain and read as live routes, clearly distinct from the muted roads.
	tradeLane := brighten(positive, 0.12)

	p := terrainPalette{
		deepWater:    deep,
		shallowWater: shallow,
		lowland:      low,
		midland:      mid,
		highland:     high,
		peak:         peak,

		bDeepWater:    bDeepWater,
		bShallowWater: bShallowWater,
		bSand:         bSand,
		bGrass:        bGrass,
		bForest:       bForest,
		bRock:         bRock,
		bMountain:     bMountain,
		bSnow:         bSnow,

		palace:    accent,
		building:  highlight,
		shadow:    blend(dim, bg, 0.35),
		road:      road,
		tradeLane: tradeLane,
	}

	// Faint per-age feel on the terrain fills only (markers keep pure role color so
	// they stay legible across ages). Both the legacy elevation bands and the biome
	// fills are shifted, so the per-age progression rides on top of the biome map.
	if ageShift != 0 {
		p.deepWater = shiftHue(p.deepWater, ageShift)
		p.shallowWater = shiftHue(p.shallowWater, ageShift)
		p.lowland = shiftHue(p.lowland, ageShift)
		p.midland = shiftHue(p.midland, ageShift)
		p.highland = shiftHue(p.highland, ageShift)
		p.peak = shiftHue(p.peak, ageShift)

		p.bDeepWater = shiftHue(p.bDeepWater, ageShift)
		p.bShallowWater = shiftHue(p.bShallowWater, ageShift)
		p.bSand = shiftHue(p.bSand, ageShift)
		p.bGrass = shiftHue(p.bGrass, ageShift)
		p.bForest = shiftHue(p.bForest, ageShift)
		p.bRock = shiftHue(p.bRock, ageShift)
		p.bMountain = shiftHue(p.bMountain, ageShift)
		p.bSnow = shiftHue(p.bSnow, ageShift)
	}
	return p
}

// lineageRoleBase maps each production lineage to a starting theme role and a
// stable hue rotation (degrees). Two different lineages may share a role base but
// never the same (base, rotation) pair, so every district reads as a distinct
// color while still being a rotation of a live theme color — switch the theme and
// every district retints together. The rotation amounts are spread around the
// wheel and chosen per lineage key, not by iteration order, so adding a lineage
// later can't shuffle the existing ones.
//
// The 13 production lineages plus the harbor/trade economic lineages are covered;
// anything unknown falls through to a deterministic hash rotation of RoleHighlight
// (see lineageColor) so a new lineage still gets a stable distinct color.
var lineageRoleBase = map[string]struct {
	role   theme.Role
	rotate float64
}{
	// Core economy — fanned around RoleHighlight / RolePositive / RoleLabel.
	"food":                  {theme.RolePositive, 0},    // greens stay greenish
	"organic_extraction":    {theme.RolePositive, 35},   // leaf → amber
	"geological_extraction": {theme.RoleLabel, 200},     // stone/teal
	"metallurgy":            {theme.RoleHighlight, 25},  // forge orange
	"energy":                {theme.RoleHighlight, 320}, // hot magenta-red
	"engineering":           {theme.RoleLabel, 250},     // industrial blue
	"knowledge":             {theme.RoleHighlight, 180}, // cyan scholarship
	"faith":                 {theme.RoleAccent, 280},    // violet devotion
	"culture_arts":          {theme.RoleHighlight, 300}, // magenta culture
	"military":              {theme.RoleNegative, 0},    // martial red
	"harbor":                {theme.RoleLabel, 190},     // sea blue-green
	"trade":                 {theme.RolePositive, 80},   // mercantile yellow-green
	"hacker":                {theme.RoleHighlight, 140}, // neon green
	"astronaut":             {theme.RoleLabel, 220},     // deep space blue
	"housing":               {theme.RoleText, 30},       // warm neutral dwellings
}

// lineageColor returns a stable, distinct color for a building given its lineage
// key and category. Special categories override the lineage:
//
//	wonder    → RoleAccent (the gold/brand highlight) — the showpieces.
//	monument  → RoleAccent brightened — kin to wonders, a touch lighter.
//	storage   → RoleDim — quiet utilitarian sheds.
//	diplomacy → RoleNegative rotated — embassy district, distinct from military.
//
// Everything else resolves through lineageRoleBase (role + hue rotation). An
// unmapped lineage gets a deterministic hash rotation of RoleHighlight so it is
// still stable and visually distinct rather than defaulting to one shared color.
func lineageColor(lineageKey, category string) color.RGBA {
	switch category {
	case "wonder":
		return rgba(theme.Color(theme.RoleAccent))
	case "monument":
		return brighten(rgba(theme.Color(theme.RoleAccent)), 0.20)
	case "storage":
		return rgba(theme.Color(theme.RoleDim))
	case "diplomacy":
		return rotateHue(rgba(theme.Color(theme.RoleNegative)), 40)
	}
	if spec, ok := lineageRoleBase[lineageKey]; ok {
		base := rgba(theme.Color(spec.role))
		if spec.rotate == 0 {
			return base
		}
		return rotateHue(base, spec.rotate)
	}
	// Unknown lineage: deterministic hue from the key so it stays stable and
	// distinct across renders. FNV-1a over the bytes → a rotation in [0,360).
	var hsh uint32 = 2166136261
	for i := 0; i < len(lineageKey); i++ {
		hsh ^= uint32(lineageKey[i])
		hsh *= 16777619
	}
	deg := float64(hsh % 360)
	return rotateHue(rgba(theme.Color(theme.RoleHighlight)), deg)
}
