package theme

import "github.com/gdamore/tcell/v2"

// Forge is the default theme: AgeForge's current dark + gold look, formalized as
// roles (theming.md §4). Accessible is false — it relies on green/red for ± — but
// it is the shipped default and still clears the WCAG AA luminance floors against
// its dark background (enforced by contrast_test.go).
//
// Role → color rationale:
//   - Accent  gold (#FFD700)  — titles, borders, brand
//   - Dim     #8b949e         — the existing dim-hint gray
//   - Label   cyan (#39C5CF)  — field labels/values (slightly deepened from pure
//                               cyan so it clears AA on the dark bg with margin)
//   - Positive green (#3FB950) — gains
//   - Negative red   (#F85149) — losses
//   - Highlight yellow (#F2CC60)— numbers / attention
//   - Text    #ffffff         — primary text
//   - Background #0d1117       — near-black canvas (GitHub-dark-ish)
//   - Selection #21304a        — dark slate selected-row backing. (Gold backing
//     was the obvious "brand" pick, but white-on-gold is ~1.4:1 — unreadable; the
//     selection backs light text, so it must stay dark. The gold brand lives in
//     borders/titles via Accent.)
var Forge = Theme{
	Key:        "forge",
	Name:       "Forge",
	Blurb:      "The classic dark-and-gold AgeForge look.",
	Accessible: false,
	Colors: [numRoles]tcell.Color{
		RoleBackground: tcell.NewRGBColor(0x0d, 0x11, 0x17),
		RoleText:       tcell.NewRGBColor(0xff, 0xff, 0xff),
		RoleDim:        tcell.NewRGBColor(0x8b, 0x94, 0x9e),
		RoleLabel:      tcell.NewRGBColor(0x39, 0xc5, 0xcf),
		RoleAccent:     tcell.NewRGBColor(0xff, 0xd7, 0x00),
		RoleHighlight:  tcell.NewRGBColor(0xf2, 0xcc, 0x60),
		RolePositive:   tcell.NewRGBColor(0x3f, 0xb9, 0x50),
		RoleNegative:   tcell.NewRGBColor(0xf8, 0x51, 0x49),
		RoleSelection:  tcell.NewRGBColor(0x21, 0x30, 0x4a),
	},
	// Non-accessible: sign carried by color alone, glyphs left empty.
	GainGlyph: "",
	LossGlyph: "",
}

func init() { register(Forge) }
