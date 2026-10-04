package ui

import (
	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/theme"
)

// Color theme — bridge over the theme package (the theming design §3.3).
// These were standalone tcell.Color globals; they are now thin accessors
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
func ColorWarning() tcell.Color  { return theme.Color(theme.RoleWarning) }
func ColorError() tcell.Color    { return theme.Color(theme.RoleNegative) }
func ColorDim() tcell.Color      { return theme.Color(theme.RoleDim) }
func ColorResource() tcell.Color { return theme.Color(theme.RoleLabel) }
func ColorBuilding() tcell.Color { return theme.Color(theme.RoleHighlight) }
func ColorVillager() tcell.Color { return theme.Color(theme.RoleAccent) }
func ColorAge() tcell.Color      { return theme.Color(theme.RoleAccent) }

// BarFillColor is the tview color tag for filled progress-bar segments. Formerly a
// fixed "#9370DB" literal; now role-derived (Accent) and emitted as a late-bound
// named tag ([accent]) so bars retint with the theme. theming.md §3.4.
func BarFillColor() string { return theme.Tag(theme.RoleAccent) }

// BarEmptyColor is the tview color tag for empty bar segments. Role-derived from
// Dim (was a fixed "#444444"). theming.md §3.4.
func BarEmptyColor() string { return theme.Tag(theme.RoleDim) }

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
