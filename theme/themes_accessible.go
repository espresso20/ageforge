package theme

import "github.com/gdamore/tcell/v2"

// The accessible themes (theming.md §4) are Accessible: true — never milestone-
// gated, always unlocked. Their defining constraint: AgeForge uses green=gain /
// red=loss everywhere, which collapses under red-green deficiency. So these themes
// encode ± with BLUE (positive) vs ORANGE (negative) — the canonical deutan/protan-
// safe opposition — AND set GainGlyph/LossGlyph so the sign is carried by shape as
// well as hue (redundant encoding; never rely on color alone).
//
// All three clear the WCAG AA luminance floors AND the colorblind-distinguishability
// guard (Positive vs Negative stay separated under both deuteranopia and protanopia
// simulation), enforced by contrast_test.go.

// signedGlyphs are shared by the colorblind-safe themes: ▲ for gains, ▼ for losses.
const (
	gainGlyph = "▲"
	lossGlyph = "▼"
)

// Deuteranopia is tuned for red-green deficiency of the M-cone type. Positive is a
// bright blue, Negative a warm orange; Accent/Highlight/Label favor the blue/
// yellow/white spread that stays mutually distinct under deuteranopia simulation
// (avoiding an accent≈positive collision).
var Deuteranopia = Theme{
	Key:        "deuteranopia",
	Name:       "Deuteranopia-safe",
	Blurb:      "Red-green safe: blue gains, orange losses, ▲/▼ signs.",
	Accessible: true,
	Colors: [numRoles]tcell.Color{
		RoleBackground: tcell.NewRGBColor(0x10, 0x14, 0x1a),
		RoleText:       tcell.NewRGBColor(0xf5, 0xf7, 0xfa),
		RoleDim:        tcell.NewRGBColor(0x9a, 0xa4, 0xb2),
		RoleLabel:      tcell.NewRGBColor(0xe9, 0xd8, 0x5a), // warm yellow label, far from the blues
		RoleAccent:     tcell.NewRGBColor(0xc9, 0xa2, 0x27), // muted gold accent
		RoleHighlight:  tcell.NewRGBColor(0xf0, 0xc8, 0x4b),
		RolePositive:   tcell.NewRGBColor(0x3b, 0x9e, 0xff), // blue = gain
		RoleNegative:   tcell.NewRGBColor(0xff, 0x8c, 0x42), // orange = loss
		RoleSelection:  tcell.NewRGBColor(0x24, 0x3b, 0x55),
	},
	GainGlyph: gainGlyph,
	LossGlyph: lossGlyph,
}

// Protanopia is tuned for red-green deficiency of the L-cone type. The safe hues
// differ slightly from deuteranopia (the orange is pushed a touch warmer/lighter
// and positive a touch deeper) since L-cone loss darkens reds differently; shipping
// it separately serves protanopes better than a single "colorblind" catch-all.
var Protanopia = Theme{
	Key:        "protanopia",
	Name:       "Protanopia-safe",
	Blurb:      "Red-green safe (L-cone): blue gains, amber losses, ▲/▼ signs.",
	Accessible: true,
	Colors: [numRoles]tcell.Color{
		RoleBackground: tcell.NewRGBColor(0x10, 0x14, 0x1a),
		RoleText:       tcell.NewRGBColor(0xf5, 0xf7, 0xfa),
		RoleDim:        tcell.NewRGBColor(0x9a, 0xa4, 0xb2),
		RoleLabel:      tcell.NewRGBColor(0xe9, 0xd8, 0x5a),
		RoleAccent:     tcell.NewRGBColor(0xcf, 0xb0, 0x3a),
		RoleHighlight:  tcell.NewRGBColor(0xf2, 0xd0, 0x55),
		RolePositive:   tcell.NewRGBColor(0x4f, 0xa8, 0xff), // blue = gain
		RoleNegative:   tcell.NewRGBColor(0xff, 0xa6, 0x3d), // amber = loss (lighter than deutan's)
		RoleSelection:  tcell.NewRGBColor(0x24, 0x3b, 0x55),
	},
	GainGlyph: gainGlyph,
	LossGlyph: lossGlyph,
}

// HighContrast is maximum legibility: near-black background, near-white text, and
// saturated, unambiguous role colors comfortably above the AA floor. For low-vision
// players and high-glare terminals. It also keeps ± blue/orange + glyphs so it is
// colorblind-safe as well as high-contrast.
var HighContrast = Theme{
	Key:        "high_contrast",
	Name:       "High Contrast",
	Blurb:      "Maximum legibility: near-black on near-white, bold roles.",
	Accessible: true,
	Colors: [numRoles]tcell.Color{
		RoleBackground: tcell.NewRGBColor(0x00, 0x00, 0x00),
		RoleText:       tcell.NewRGBColor(0xff, 0xff, 0xff),
		RoleDim:        tcell.NewRGBColor(0xc8, 0xc8, 0xc8),
		RoleLabel:      tcell.NewRGBColor(0x6e, 0xd6, 0xff), // bright sky-blue label
		RoleAccent:     tcell.NewRGBColor(0xff, 0xe0, 0x33), // vivid yellow accent
		RoleHighlight:  tcell.NewRGBColor(0xff, 0xff, 0x66),
		RolePositive:   tcell.NewRGBColor(0x5a, 0xb6, 0xff), // blue = gain
		RoleNegative:   tcell.NewRGBColor(0xff, 0x9d, 0x3d), // orange = loss
		RoleSelection:  tcell.NewRGBColor(0x33, 0x33, 0x33),
	},
	GainGlyph: gainGlyph,
	LossGlyph: lossGlyph,
}

func init() {
	register(Deuteranopia)
	register(Protanopia)
	register(HighContrast)
}
