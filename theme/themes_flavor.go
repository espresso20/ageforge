package theme

import "github.com/gdamore/tcell/v2"

// Flavor themes (theming.md §4) are cosmetic, curated palettes that unlock via
// milestones. They are all Accessible: false — the gating itself lands in Phase 3b;
// this file only DEFINES and REGISTERS them. Because they're not accessible, the
// colorblind-distinguishability guard does not apply to them, but every WCAG AA
// luminance floor in contrast_test.go DOES, and each theme below is tuned to clear
// those with margin (Text/Label/Positive/Negative/Highlight vs Background >= 4.5,
// Dim/Accent vs Background >= 3.0, Text-on-Selection >= 4.5).
//
// Flavor themes don't need colorblind separation, but the ▲/▼ glyphs are harmless
// and keep delta formatting consistent across every theme, so they're set anyway.
const (
	flavorGainGlyph = "▲"
	flavorLossGlyph = "▼"
)

// Parchment is the one LIGHT-background theme: dark ink-brown text on warm cream,
// with sepia/leather accents. It's the real contrast-guard exercise (theming.md §4,
// §8) — on a light bg the role colors must be DARK enough to stay legible, so the
// usual bright accent/positive/negative become deep, saturated versions: forest
// green gains, wax-red losses, umber labels, sepia accent/highlight.
var Parchment = Theme{
	Key:        "parchment",
	Name:       "Parchment",
	Blurb:      "Ink on warm parchment — the one light-background look.",
	Accessible: false,
	Colors: [numRoles]tcell.Color{
		RoleBackground: tcell.NewRGBColor(0xf2, 0xe8, 0xd0), // warm cream
		RoleText:       tcell.NewRGBColor(0x3a, 0x2c, 0x1a), // dark ink brown
		RoleDim:        tcell.NewRGBColor(0x6e, 0x5c, 0x40), // faded sepia
		RoleLabel:      tcell.NewRGBColor(0x6b, 0x4a, 0x12), // deep umber
		RoleAccent:     tcell.NewRGBColor(0x7e, 0x50, 0x10), // sepia / leather
		RoleHighlight:  tcell.NewRGBColor(0x7e, 0x50, 0x10), // wax-amber (dark on light)
		RolePositive:   tcell.NewRGBColor(0x2c, 0x6e, 0x2c), // forest green
		RoleNegative:   tcell.NewRGBColor(0xa3, 0x2a, 0x1f), // wax red
		RoleSelection:  tcell.NewRGBColor(0xd8, 0xc4, 0x9a), // darker cream backing
	},
	GainGlyph: flavorGainGlyph,
	LossGlyph: flavorLossGlyph,
}

// Bronze is a warm metallic look: copper/bronze accent, amber highlight, olive
// gains and burnt-orange losses on a dark warm brown-black background.
var Bronze = Theme{
	Key:        "bronze",
	Name:       "Bronze",
	Blurb:      "Burnished copper and amber over dark, warm metal.",
	Accessible: false,
	Colors: [numRoles]tcell.Color{
		RoleBackground: tcell.NewRGBColor(0x1c, 0x14, 0x0c), // dark warm brown-black
		RoleText:       tcell.NewRGBColor(0xf0, 0xe2, 0xc8), // warm parchment text
		RoleDim:        tcell.NewRGBColor(0x9a, 0x82, 0x60), // patina'd bronze
		RoleLabel:      tcell.NewRGBColor(0xd9, 0xa8, 0x6a), // copper
		RoleAccent:     tcell.NewRGBColor(0xcd, 0x7f, 0x32), // bronze
		RoleHighlight:  tcell.NewRGBColor(0xf2, 0xc1, 0x60), // amber
		RolePositive:   tcell.NewRGBColor(0x8f, 0xb5, 0x4a), // olive gain
		RoleNegative:   tcell.NewRGBColor(0xe0, 0x6a, 0x3c), // burnt-orange loss
		RoleSelection:  tcell.NewRGBColor(0x3a, 0x2a, 0x18),
	},
	GainGlyph: flavorGainGlyph,
	LossGlyph: flavorLossGlyph,
}

// Cyberpunk is neon on near-black: hot magenta accent, neon-cyan highlight, neon
// green/pink for ±, over a near-black violet background (riffs on the cyberpunk_age
// palette, theming.md §4).
var Cyberpunk = Theme{
	Key:        "cyberpunk",
	Name:       "Cyberpunk",
	Blurb:      "Hot magenta and neon cyan over near-black violet.",
	Accessible: false,
	Colors: [numRoles]tcell.Color{
		RoleBackground: tcell.NewRGBColor(0x0a, 0x06, 0x12), // near-black violet
		RoleText:       tcell.NewRGBColor(0xe6, 0xe6, 0xf5), // cool white
		RoleDim:        tcell.NewRGBColor(0x7a, 0x6e, 0x9e), // muted violet-grey
		RoleLabel:      tcell.NewRGBColor(0x3d, 0xe1, 0xe8), // cyan
		RoleAccent:     tcell.NewRGBColor(0xff, 0x2e, 0xc0), // hot magenta
		RoleHighlight:  tcell.NewRGBColor(0x4d, 0xf0, 0xff), // neon cyan
		RolePositive:   tcell.NewRGBColor(0x39, 0xff, 0x9e), // neon green gain
		RoleNegative:   tcell.NewRGBColor(0xff, 0x49, 0x6b), // neon pink-red loss
		RoleSelection:  tcell.NewRGBColor(0x2a, 0x12, 0x3a),
	},
	GainGlyph: flavorGainGlyph,
	LossGlyph: flavorLossGlyph,
}

// Monochrome is a stylistic greyscale terminal: a single hue's shades. Accent and
// Highlight are bright greys/white; the ± distinction rides on LIGHTNESS (a lighter
// grey gains, a mid grey loses) rather than hue. It is NOT an accessibility theme —
// just a clean mono look — so it carries Accessible: false and skips the colorblind
// guard, but still clears every luminance floor.
var Monochrome = Theme{
	Key:        "monochrome",
	Name:       "Monochrome",
	Blurb:      "Greyscale terminal — meaning carried by lightness, not hue.",
	Accessible: false,
	Colors: [numRoles]tcell.Color{
		RoleBackground: tcell.NewRGBColor(0x12, 0x12, 0x12), // near-black grey
		RoleText:       tcell.NewRGBColor(0xf5, 0xf5, 0xf5), // near-white
		RoleDim:        tcell.NewRGBColor(0x8a, 0x8a, 0x8a), // mid grey
		RoleLabel:      tcell.NewRGBColor(0xc8, 0xc8, 0xc8), // light grey
		RoleAccent:     tcell.NewRGBColor(0xe8, 0xe8, 0xe8), // bright grey
		RoleHighlight:  tcell.NewRGBColor(0xff, 0xff, 0xff), // white
		RolePositive:   tcell.NewRGBColor(0xd8, 0xd8, 0xd8), // lighter grey = gain
		RoleNegative:   tcell.NewRGBColor(0x9a, 0x9a, 0x9a), // mid grey = loss
		RoleSelection:  tcell.NewRGBColor(0x33, 0x33, 0x33),
	},
	GainGlyph: flavorGainGlyph,
	LossGlyph: flavorLossGlyph,
}

// Cosmic is deep-space: a dark indigo/violet background, starlight text, and
// nebula-pink / cyan accents with aurora-green gains (riffs on galactic_age,
// theming.md §4).
var Cosmic = Theme{
	Key:        "cosmic",
	Name:       "Cosmic",
	Blurb:      "Deep indigo with starlight and nebula accents.",
	Accessible: false,
	Colors: [numRoles]tcell.Color{
		RoleBackground: tcell.NewRGBColor(0x0c, 0x0a, 0x1f), // deep indigo
		RoleText:       tcell.NewRGBColor(0xe8, 0xe6, 0xf8), // starlight
		RoleDim:        tcell.NewRGBColor(0x7e, 0x78, 0xa6), // dust violet
		RoleLabel:      tcell.NewRGBColor(0x6e, 0xd6, 0xe8), // cyan
		RoleAccent:     tcell.NewRGBColor(0xc8, 0x6e, 0xff), // nebula violet
		RoleHighlight:  tcell.NewRGBColor(0xff, 0x8a, 0xd8), // nebula pink
		RolePositive:   tcell.NewRGBColor(0x5a, 0xe0, 0xa8), // aurora green
		RoleNegative:   tcell.NewRGBColor(0xff, 0x6e, 0x8a), // nebula red
		RoleSelection:  tcell.NewRGBColor(0x24, 0x20, 0x4a),
	},
	GainGlyph: flavorGainGlyph,
	LossGlyph: flavorLossGlyph,
}

func init() {
	register(Parchment)
	register(Bronze)
	register(Cyberpunk)
	register(Monochrome)
	register(Cosmic)
}
