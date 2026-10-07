package theme

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
)

// derive.go holds the pure color math the rest of the codebase leans on when it
// needs a color that isn't literally a role: extended-role defaults, legibility
// correction for fixed identity hues, and light/dark-aware shading. Keeping it
// here is what lets "no raw colors outside theme/" hold without every caller
// reinventing a blend.

var (
	black = tcell.NewRGBColor(0, 0, 0)
	white = tcell.NewRGBColor(0xff, 0xff, 0xff)
)

// deriveRoles fills every unset (zero / ColorDefault) extended role from the core
// nine. The defaults reproduce exactly what the UI drew before those roles
// existed, so a theme that predates them renders identically:
//
//	Surface       = Background        (modals/overlays painted the canvas color)
//	Border        = Accent            (borders were Accent)
//	SelectionText = Text              (selected rows drew Text)
//	Bright        = Text
//	Warning       = Highlight         ([yellow] carried warnings)
//	OnAccent      = black or white, whichever reads better on Accent
//	Chip          = Background 15% toward Text
//	OnNegative    = Text if it clears 3:1 on Negative, else black or white
//
// Core roles are never touched: a theme that leaves one of those unset is a bug
// the contrast test will catch.
func deriveRoles(t Theme) Theme {
	c := &t.Colors
	unset := func(r Role) bool { return c[r] == tcell.ColorDefault }
	if unset(RoleSurface) {
		c[RoleSurface] = c[RoleBackground]
	}
	if unset(RoleBorder) {
		c[RoleBorder] = c[RoleAccent]
	}
	if unset(RoleSelectionText) {
		c[RoleSelectionText] = c[RoleText]
	}
	if unset(RoleBright) {
		c[RoleBright] = c[RoleText]
	}
	if unset(RoleWarning) {
		c[RoleWarning] = c[RoleHighlight]
	}
	if unset(RoleOnAccent) {
		c[RoleOnAccent] = BestOn(c[RoleAccent])
	}
	if unset(RoleChip) {
		c[RoleChip] = Mix(c[RoleBackground], c[RoleText], 0.15)
	}
	if unset(RoleOnNegative) {
		if ContrastRatio(c[RoleText], c[RoleNegative]) >= 3.0 {
			c[RoleOnNegative] = c[RoleText]
		} else {
			c[RoleOnNegative] = BestOn(c[RoleNegative])
		}
	}
	return t
}

// BestOn returns black or white, whichever has the higher contrast against bg.
func BestOn(bg tcell.Color) tcell.Color {
	if ContrastRatio(black, bg) >= ContrastRatio(white, bg) {
		return black
	}
	return white
}

// Mix linearly interpolates a toward b by t in [0,1] in sRGB space.
func Mix(a, b tcell.Color, t float64) tcell.Color {
	if t <= 0 {
		return a
	}
	if t >= 1 {
		return b
	}
	ar, ag, ab := a.RGB()
	br, bg, bb := b.RGB()
	lerp := func(x, y int32) int32 { return x + int32(float64(y-x)*t+0.5*sign(y-x)) }
	return tcell.NewRGBColor(lerp(ar, br), lerp(ag, bg), lerp(ab, bb))
}

func sign(v int32) float64 {
	if v < 0 {
		return -1
	}
	return 1
}

// Legible nudges a fixed identity hue (an epoch color, a lineage tint, a civ
// marker) just far enough toward the theme's text pole that it clears min
// contrast against bg — lighter on a dark background, darker on a light one. A
// color that already clears min is returned untouched, so dark themes keep their
// original hues wherever those were already readable.
func Legible(c, bg tcell.Color, min float64) tcell.Color {
	if ContrastRatio(c, bg) >= min {
		return c
	}
	pole := white
	if RelativeLuminance(bg) >= lightLuminanceThreshold {
		pole = black
	}
	// Binary search the smallest blend toward the pole that clears min. 12
	// iterations is sub-1/4000 precision — far below one 8-bit step.
	lo, hi := 0.0, 1.0
	for i := 0; i < 12; i++ {
		mid := (lo + hi) / 2
		if ContrastRatio(Mix(c, pole, mid), bg) >= min {
			hi = mid
		} else {
			lo = mid
		}
	}
	return Mix(c, pole, hi)
}

// LegibleOnActive is Legible against the active theme's Background at the
// body-text floor (4.5:1).
func LegibleOnActive(c tcell.Color) tcell.Color {
	return Legible(c, Color(RoleBackground), 4.5)
}

// HexTag renders a literal color as a tview "[#rrggbb]" tag. Use it only for
// colors that are NOT the active theme's roles (e.g. the picker drawing another
// theme's swatches) — hex tags are early-bound and do not retint on a switch.
func HexTag(c tcell.Color) string {
	h := c.Hex()
	if h < 0 {
		// Invalid/default color: fall back to the active Dim so the tag is never
		// malformed and never a hard-coded grey.
		h = Color(RoleDim).Hex()
		if h < 0 {
			h = 0x808080
		}
	}
	return fmt.Sprintf("[#%06x]", h)
}

// HexTagFgBg renders a literal fg/bg pair as a tview "[#rrggbb:#rrggbb]" tag.
func HexTagFgBg(fg, bg tcell.Color) string {
	f := HexTag(fg)
	b := HexTag(bg)
	return f[:len(f)-1] + ":" + b[1:]
}

// Hue parses an identity colour written as "#rrggbb" (a tech lane's, in the
// lane table) and holds it to the contrast rule for art on bg: at least 3.0
// against the background it is drawn on. A string that is not a colour reads
// as the Accent role.
func Hue(hex string, bg tcell.Color) tcell.Color {
	var r, g, b int32
	if _, err := fmt.Sscanf(hex, "#%02x%02x%02x", &r, &g, &b); err != nil {
		return Legible(Color(RoleAccent), bg, 3)
	}
	return Legible(tcell.NewRGBColor(r, g, b), bg, 3)
}
