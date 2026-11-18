package theme

import (
	"github.com/espresso20/ageforge/rowid"
)

// Light themes (the theming design §4). Both are always available: Daylight sits in the
// Standard group next to Forge, High Contrast Light in Accessibility next to its
// dark sibling. They exist because a dark theme on a light terminal is only
// half-fixed by painting an explicit background — plenty of players simply want
// dark ink on a light page.
//
// On a light canvas every foreground role has to be DARK enough to read, so the
// familiar Forge hues become their deep counterparts: gold accent → deep amber,
// cyan labels → teal, bright green/red → forest/brick. Extended roles are set
// explicitly where the derivation would be wrong for a light page (Surface is a
// true white panel over an off-white canvas; Bright is near-black ink).

// Daylight is the clean light default: near-white canvas, white panels, charcoal
// ink, and deep-amber brand accents that echo Forge's gold.
var Daylight = define(Theme{
	Key:        "daylight",
	Name:       rowid.E("theme.daylight.name", "Daylight"),
	Blurb:      rowid.E("theme.daylight.blurb", "Clean light theme — charcoal ink on a bright page, amber accents."),
	Accessible: false,
	Standard:   true,
	Colors: [numRoles]Color{
		RoleBackground: NewRGBColor(0xf6, 0xf7, 0xf9), // off-white canvas
		RoleText:       NewRGBColor(0x1f, 0x23, 0x28), // charcoal ink
		RoleDim:        NewRGBColor(0x57, 0x60, 0x6a), // slate secondary
		RoleLabel:      NewRGBColor(0x0a, 0x6e, 0x7a), // deep teal (Forge's cyan, darkened)
		RoleAccent:     NewRGBColor(0x9a, 0x67, 0x00), // deep amber (Forge's gold, darkened)
		RoleHighlight:  NewRGBColor(0xa8, 0x52, 0x00), // burnt orange numbers
		RolePositive:   NewRGBColor(0x1a, 0x7f, 0x37), // forest green
		RoleNegative:   NewRGBColor(0xcf, 0x22, 0x2e), // brick red
		RoleSelection:  NewRGBColor(0xd6, 0xe4, 0xfa), // pale blue row
		RoleComment:    NewRGBColor(0xc2, 0x41, 0x0c), // flavor sentences only: deep orange, darkened for the light canvas
		RoleEffect:     NewRGBColor(0x0b, 0x5c, 0xad), // what a tech does: deep blue, darkened for the light canvas

		RoleSurface: NewRGBColor(0xff, 0xff, 0xff), // white panels over the canvas
		RoleBright:  NewRGBColor(0x0b, 0x0f, 0x14), // near-black emphasis
		RoleChip:    NewRGBColor(0xdd, 0xe1, 0xe6), // light grey keycap label
	},
	GainGlyph: "",
	LossGlyph: "",
})

// HighContrastLight is the light counterpart to High Contrast: pure white page,
// black ink, and every role at AAA (7:1) where it carries text. It keeps the
// colorblind-safe blue-gain / orange-loss encoding and the ▲/▼ glyphs, so it is
// accessible on both axes — contrast and color vision.
var HighContrastLight = define(Theme{
	Key:        "high_contrast_light",
	Name:       rowid.E("theme.high_contrast_light.name", "High Contrast Light"),
	Blurb:      rowid.E("theme.high_contrast_light.blurb", "Maximum legibility on a white page: black ink, bold roles."),
	Accessible: true,
	Colors: [numRoles]Color{
		RoleBackground: NewRGBColor(0xff, 0xff, 0xff),
		RoleText:       NewRGBColor(0x00, 0x00, 0x00),
		RoleDim:        NewRGBColor(0x40, 0x40, 0x40),
		RoleLabel:      NewRGBColor(0x00, 0x55, 0x6a), // dark teal label
		RoleAccent:     NewRGBColor(0x6e, 0x45, 0x00), // dark bronze accent
		RoleHighlight:  NewRGBColor(0x5b, 0x2d, 0x90), // deep violet numbers
		RolePositive:   NewRGBColor(0x00, 0x48, 0xa8), // blue = gain
		RoleNegative:   NewRGBColor(0x93, 0x38, 0x00), // dark orange = loss
		RoleSelection:  NewRGBColor(0xb8, 0xd4, 0xff),
		RoleComment:    NewRGBColor(0xb8, 0x43, 0x00), // flavor sentences only: dark orange, 5.5:1 on white
		RoleEffect:     NewRGBColor(0x1d, 0x4f, 0xb8), // what a tech does: deep blue on white

		RoleBorder: NewRGBColor(0x00, 0x00, 0x00), // hard black borders
	},
	GainGlyph: gainGlyph,
	LossGlyph: lossGlyph,
})

var _ = register(Daylight, HighContrastLight)
