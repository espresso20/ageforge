package ui

import (
	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/theme"
)

// Color theme — bridge over the theme package (design-and-architecture/theming.md
// §3.3). These were standalone tcell.Color globals; they are now thin accessors
// over theme.Color(role) so every call site keeps compiling but pulls the ACTIVE
// theme's color and tracks live theme switches. Under Forge (the default) they
// resolve to the game's current palette.
//
// Role mapping (a deliberate flattening of the old ad-hoc palette onto the ~9-role
// model): the prior globals carried five distinct accent-ish hues (gold title,
// DodgerBlue accent, teal resource, orange-red building, purple worker). The role
// model collapses those to Accent/Label/Highlight — so under non-Forge themes those
// chrome titles tint coherently instead of being one-off colors. See §3.1/§3.3.
func ColorBg() tcell.Color       { return theme.Color(theme.RoleBackground) }
func ColorFg() tcell.Color       { return theme.Color(theme.RoleText) }
func ColorTitle() tcell.Color    { return theme.Color(theme.RoleAccent) }
func ColorAccent() tcell.Color   { return theme.Color(theme.RoleLabel) }
func ColorSuccess() tcell.Color  { return theme.Color(theme.RolePositive) }
func ColorWarning() tcell.Color  { return theme.Color(theme.RoleHighlight) }
func ColorError() tcell.Color    { return theme.Color(theme.RoleNegative) }
func ColorDim() tcell.Color      { return theme.Color(theme.RoleDim) }
func ColorResource() tcell.Color { return theme.Color(theme.RoleLabel) }
func ColorBuilding() tcell.Color { return theme.Color(theme.RoleHighlight) }
func ColorVillager() tcell.Color { return theme.Color(theme.RoleAccent) }
func ColorAge() tcell.Color      { return theme.Color(theme.RoleAccent) }

// BarFillColor is the tview color tag for filled progress-bar segments. Formerly a
// fixed "#9370DB" literal; now role-derived (Accent) and emitted as a live hex tag
// so bars retint with the theme. theming.md §3.4.
func BarFillColor() string { return theme.Tag(theme.RoleAccent) }

// BarEmptyColor is the tview color tag for empty bar segments. Role-derived from
// Dim (was a fixed "#444444"). theming.md §3.4.
func BarEmptyColor() string { return theme.Tag(theme.RoleDim) }

// AgePalette defines the color theme for an age era.
//
// note: retained as data only. The age-palette globals used to be a competing
// color authority that mutated the chrome colors as the player advanced ages —
// touching chrome but never the inline [gold]/[cyan]/… body tags, which is why an
// age advance recolored borders but not text. That half-measure is subsumed by the
// theme package. ApplyAgePalette is now a no-op; this data is kept for Phase 4,
// which reframes it as the optional epoch-adaptive Adaptive theme (theming.md §9).
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

// ApplyAgePalette is intentionally a no-op (theming.md §3.3, §9).
//
// It used to mutate the chrome color globals on every age advance, fighting the
// theme as a second color authority. The theme package is now the single source of
// truth, so this does nothing. Callers (the dashboard age-advance path) are left in
// place; Phase 4 reframes the age-palette idea as the optional epoch-adaptive
// "Adaptive" theme that retints role colors through the theme package instead.
func ApplyAgePalette(ageKey string) {
	_ = ageKey
}

// ASCII art for splash screen
const SplashArt = `
███████   █████   ███████           ███████  █████   ██████   █████    ███████
█     █  █        █                 █       █     █  █     █  █        █
█     █  █        █                 █       █     █  █     █  █        █
███████  █  ████  █████    █████    █████   █     █  ██████   █  ████  █████
█     █  █     █  █                 █       █     █  █  █     █     █  █
█     █  █     █  █                 █       █     █  █   █    █     █  █
█     █   █████   ███████           █        █████   █    █    █████   ███████
`

const SplashTagline = "Forge the Ultimate Empire Through the Ages"
