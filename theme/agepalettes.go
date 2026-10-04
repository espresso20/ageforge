package theme

import "github.com/gdamore/tcell/v2"

// agepalettes.go is inert data moved out of ui/theme.go so the "no raw colors
// outside theme/" rule holds (the theming design §3.8). It is the old per-age chrome
// palette, kept for the Phase-4 epoch-adaptive theme (the theming design §9). Nothing
// reads it today.

// AgePalette defines the color theme for an age era.
//
// note: retained as data only. The age-palette globals used to be a competing
// color authority that mutated the chrome colors as the player advanced ages —
// touching chrome but never the inline [gold]/[cyan]/… body tags, which is why an
// age advance recolored borders but not text. That half-measure is subsumed by the
// theme package. ui.ApplyAgePalette is a no-op; this data is kept for Phase 4,
// which reframes it as the optional epoch-adaptive Adaptive theme (the theming design §9).
type AgePalette struct {
	Title    tcell.Color
	Accent   tcell.Color
	Resource tcell.Color
	Building tcell.Color
	Dim      tcell.Color
}

// AgePalettes maps age keys to their color palette. Inert until §9 (Phase 4).
var AgePalettes = map[string]AgePalette{
	// Primitive/Stone: earthy greens and browns
	"primitive_age": {tcell.ColorDarkGreen, tcell.ColorOlive, tcell.ColorTeal, tcell.ColorSaddleBrown, tcell.ColorDimGray},
	"stone_age":     {tcell.ColorDarkGreen, tcell.ColorOlive, tcell.ColorTeal, tcell.ColorSienna, tcell.ColorDimGray},
	// Bronze/Iron: warm bronze and metallic
	"bronze_age": {tcell.ColorGold, tcell.ColorDarkGoldenrod, tcell.ColorTeal, tcell.ColorOrangeRed, tcell.ColorGray},
	"iron_age":   {tcell.ColorSilver, tcell.ColorSteelBlue, tcell.ColorTeal, tcell.ColorOrangeRed, tcell.ColorGray},
	// Classical: marble white and royal blue
	"classical_age": {tcell.ColorWhite, tcell.ColorRoyalBlue, tcell.ColorCadetBlue, tcell.ColorCoral, tcell.ColorLightGray},
	// Medieval: dark purple and stone
	"medieval_age":    {tcell.ColorDarkMagenta, tcell.ColorMediumPurple, tcell.ColorDarkCyan, tcell.ColorFireBrick, tcell.ColorDimGray},
	"renaissance_age": {tcell.ColorGold, tcell.ColorMediumOrchid, tcell.ColorDarkCyan, tcell.ColorOrangeRed, tcell.ColorGray},
	"colonial_age":    {tcell.ColorNavajoWhite, tcell.ColorBurlyWood, tcell.ColorTeal, tcell.ColorSienna, tcell.ColorGray},
	// Industrial: dark grays and orange
	"industrial_age": {tcell.ColorDarkOrange, tcell.ColorOrange, tcell.ColorDarkSlateGray, tcell.ColorFireBrick, tcell.ColorDarkGray},
	"victorian_age":  {tcell.ColorRosyBrown, tcell.ColorDarkKhaki, tcell.ColorSlateGray, tcell.ColorBrown, tcell.ColorDimGray},
	"electric_age":   {tcell.ColorYellow, tcell.ColorGold, tcell.ColorTeal, tcell.ColorOrangeRed, tcell.ColorGray},
	// Modern: clean blue and white
	"atomic_age":      {tcell.ColorLimeGreen, tcell.ColorGreen, tcell.ColorDarkCyan, tcell.ColorRed, tcell.ColorDarkGray},
	"modern_age":      {tcell.ColorDodgerBlue, tcell.ColorSteelBlue, tcell.ColorTeal, tcell.ColorOrangeRed, tcell.ColorGray},
	"information_age": {tcell.ColorDeepSkyBlue, tcell.ColorCornflowerBlue, tcell.ColorMediumAquamarine, tcell.ColorOrangeRed, tcell.ColorLightGray},
	// Digital: blue/cyan tech
	"digital_age":   {tcell.ColorDarkCyan, tcell.ColorDodgerBlue, tcell.ColorMediumAquamarine, tcell.ColorDeepPink, tcell.ColorDarkSlateGray},
	"cyberpunk_age": {tcell.ColorHotPink, tcell.ColorDarkMagenta, tcell.ColorAqua, tcell.ColorLime, tcell.ColorDarkSlateGray},
	// Fusion/Space: blue and white
	"fusion_age":       {tcell.ColorAquaMarine, tcell.ColorDarkCyan, tcell.ColorTurquoise, tcell.ColorOrangeRed, tcell.ColorSlateGray},
	"space_age":        {tcell.ColorSteelBlue, tcell.ColorLightSkyBlue, tcell.ColorLightCyan, tcell.ColorOrangeRed, tcell.ColorSlateGray},
	"interstellar_age": {tcell.ColorMediumPurple, tcell.ColorSlateBlue, tcell.ColorLightBlue, tcell.ColorGold, tcell.ColorDimGray},
	// Cosmic: deep purple and gold
	"galactic_age":     {tcell.ColorBlueViolet, tcell.ColorMediumPurple, tcell.ColorLavender, tcell.ColorGold, tcell.ColorDimGray},
	"quantum_age":      {tcell.ColorMediumOrchid, tcell.ColorOrchid, tcell.ColorPlum, tcell.ColorGold, tcell.ColorDarkSlateGray},
	"transcendent_age": {tcell.ColorGold, tcell.ColorWhite, tcell.ColorLightGoldenrodYellow, tcell.ColorGold, tcell.ColorLightGray},
}
