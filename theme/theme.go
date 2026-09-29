// Package theme is AgeForge's single source of truth for UI color.
//
// It is a leaf package: it depends only on gdamore/tcell/v2 and go-colorful (plus
// rivo/tview in remap.go, solely to push chrome defaults into tview.Styles). It
// must NOT import game/ or ui/ — themes are pure presentation and an import cycle
// here would be a design smell. See design-and-architecture/theming.md §3.
//
// The model is a set of semantic *roles* (not literal color names): Positive is
// "the gains color," whatever hue the active theme picks. A theme fully specifies
// its surface — canvas, panel surface, border, selection, text tiers, accent and
// the good/warn/bad semantics — so a light theme paints correctly on a dark
// terminal and vice versa. Colors reach the screen three ways, all owned here:
//
//   - inline tview tags ([accent], or the legacy aliases [gold]/[green]/…) are
//     retinted through the tcell.ColorNames remap (remap.go);
//   - widget chrome uses Ref(role) sentinels that WrapScreen resolves to the
//     active theme on every SetContent (screen.go), so no widget can go stale;
//   - code that needs a concrete color calls Color(role) / the derive helpers.
package theme

import "github.com/gdamore/tcell/v2"

// Role enumerates the semantic color slots a theme must fill. Order is fixed and
// load-bearing: themes declare their palette as a [numRoles]tcell.Color indexed by
// these constants (see theming.md §3.1).
//
// The first nine are the original Phase-1 roles. The rest arrived with the
// light-theme overhaul so a theme can describe its whole surface. A theme may
// leave any of the extended roles unset; register() derives them from the core
// nine (deriveRoles), which is how the pre-existing dark themes kept their look.
type Role int

const (
	RoleBackground Role = iota // canvas / primitive background
	RoleText                   // primary readable text
	RoleDim                    // secondary / hints / disabled
	RoleLabel                  // field labels, values
	RoleAccent                 // titles, brand highlights
	RoleHighlight              // numbers, attention, "look here"
	RolePositive               // gains, success, +deltas (semantic "good")
	RoleNegative               // losses, errors, -deltas (semantic "bad")
	RoleSelection              // selected list-row background

	RoleSurface       // panel / modal / overlay background (derived: Background)
	RoleBorder        // box borders and rules (derived: Accent)
	RoleSelectionText // text drawn on Selection (derived: Text)
	RoleBright        // emphasized text inside body copy (derived: Text)
	RoleWarning       // caution, semantic "warn" (derived: Highlight)
	RoleOnAccent      // text drawn ON an Accent fill: keycaps, primary buttons (derived: black/white)
	RoleChip          // chip / keycap-label background (derived: Background 15% toward Text)
	RoleOnNegative    // text drawn ON a Negative fill: danger modals (derived: Text, else black/white)
	numRoles
)

// NumRoles is the number of roles, exported for tests and iteration.
const NumRoles = int(numRoles)

// String renders a Role for diagnostics and test failure messages.
func (r Role) String() string {
	switch r {
	case RoleBackground:
		return "Background"
	case RoleText:
		return "Text"
	case RoleDim:
		return "Dim"
	case RoleLabel:
		return "Label"
	case RoleAccent:
		return "Accent"
	case RoleHighlight:
		return "Highlight"
	case RolePositive:
		return "Positive"
	case RoleNegative:
		return "Negative"
	case RoleSelection:
		return "Selection"
	case RoleSurface:
		return "Surface"
	case RoleBorder:
		return "Border"
	case RoleSelectionText:
		return "SelectionText"
	case RoleBright:
		return "Bright"
	case RoleWarning:
		return "Warning"
	case RoleOnAccent:
		return "OnAccent"
	case RoleChip:
		return "Chip"
	case RoleOnNegative:
		return "OnNegative"
	default:
		return "Role(?)"
	}
}

// Theme is a complete, code-defined palette plus picker metadata. Colors carry
// true RGB via tcell.NewRGBColor so themes are not at the mercy of a terminal's
// 16-color palette on truecolor terminals (theming.md §3.1).
type Theme struct {
	Key        string // "forge", "deuteranopia", ... — stable identifier
	Name       string // "Forge" — shown in the picker
	Blurb      string // one-line flavor for the picker detail pane
	Accessible bool   // true => never milestone-gated, always unlocked
	Standard   bool   // true => always unlocked, listed under Standard (Forge, Daylight)

	Colors [numRoles]tcell.Color

	// Signed sentinels for the ± distinction in accessible themes (theming.md §4):
	// the sign is encoded by shape as well as hue so colorblind players never rely
	// on color alone. Non-accessible themes may leave these empty.
	GainGlyph string // e.g. "▲" / "+"
	LossGlyph string // e.g. "▼" / "-"

	// Duotone marks a one-ink-on-one-paper theme (Monochrome, Parchment):
	// pictorial surfaces such as the skyline map fold every scene color
	// onto the ramp from Background to Text instead of drawing full color.
	Duotone bool

	// Milestone-gated unlock condition (theming.md §5). A gated (flavor) theme
	// declares EXACTLY ONE of these — the milestone key or chain key whose
	// completion unlocks it account-wide. The mapping lives here, in the registry,
	// not scattered through engine/milestone code: theme stays a leaf package, so
	// these are plain strings (no game import), and the UI reverse-maps a completed
	// key back to a theme via UnlockedBy.
	//
	// Always-available themes (Accessible, or the Standard Forge/Daylight) leave
	// BOTH empty — they're never gated, so there is nothing to unlock. The
	// registry-consistency test (unlock_test.go) enforces the XOR for gated themes
	// and the empty-pair for the always-available set.
	UnlockMilestone string // milestone key (config/milestones.go) — XOR with UnlockChain
	UnlockChain     string // milestone-chain key — XOR with UnlockMilestone

	// UnlockHint is the human-readable unlock condition shown for a LOCKED theme in
	// the picker detail pane and `theme list` (e.g. "Reach the Cyberpunk Age").
	// Required for gated themes; empty for always-available ones.
	UnlockHint string
}

// Gated reports whether the theme is milestone-gated (declares an unlock
// condition). Always-available themes (Accessible / Standard) are not gated.
func (t Theme) Gated() bool {
	return t.UnlockMilestone != "" || t.UnlockChain != ""
}

// UnlockKey returns the single milestone-or-chain key that unlocks a gated theme
// (whichever of UnlockMilestone/UnlockChain is set), and "" for an un-gated theme.
func (t Theme) UnlockKey() string {
	if t.UnlockMilestone != "" {
		return t.UnlockMilestone
	}
	return t.UnlockChain
}

// Color returns the theme's color for a role. Out-of-range roles return
// tcell.ColorDefault rather than panicking.
func (t Theme) Color(role Role) tcell.Color {
	if role < 0 || role >= numRoles {
		return tcell.ColorDefault
	}
	return t.Colors[role]
}

// Group is the picker section a theme is listed under.
type Group int

const (
	GroupStandard      Group = iota // always available, not an accessibility theme (Forge, Daylight)
	GroupAccessibility              // Accessible: colorblind-safe / high-contrast, always unlocked
	GroupUnlockable                 // milestone-gated flavor themes
)

// Groups lists the picker sections in display order.
var Groups = []Group{GroupStandard, GroupAccessibility, GroupUnlockable}

// String is the picker section heading for a group.
func (g Group) String() string {
	switch g {
	case GroupAccessibility:
		return "Accessibility"
	case GroupUnlockable:
		return "Unlockable"
	default:
		return "Standard"
	}
}

// Group reports which picker section the theme belongs to.
func (t Theme) Group() Group {
	switch {
	case t.Accessible:
		return GroupAccessibility
	case t.Standard:
		return GroupStandard
	default:
		return GroupUnlockable
	}
}

// AlwaysAvailable reports whether the theme is unlocked for every account from
// the start (Accessible or Standard). Everything else is milestone-gated; the
// registry test (unlock_test.go) enforces that split.
func (t Theme) AlwaysAvailable() bool {
	return t.Accessible || t.Standard
}

// lightLuminanceThreshold splits light from dark backgrounds. 0.25 relative
// luminance sits a little above perceptual mid-grey (L* ≈ 57): a cream or
// off-white canvas is light, anything from charcoal to slate is dark.
const lightLuminanceThreshold = 0.25

// IsLight reports whether the theme has a light background. Derived from the
// Background role's luminance rather than stored, so it can never drift from the
// palette. Map derivations and the picker's Light/Dark tag key off this.
func (t Theme) IsLight() bool {
	return RelativeLuminance(t.Color(RoleBackground)) >= lightLuminanceThreshold
}

// Variant is "Light" or "Dark" — the picker tag for IsLight.
func (t Theme) Variant() string {
	if t.IsLight() {
		return "Light"
	}
	return "Dark"
}

// registry holds every built-in theme, keyed by Key. Populated by the
// themes_*.go init() functions via register().
var registry = map[string]Theme{}

// registryOrder preserves insertion order so All() is deterministic (Forge first).
var registryOrder []string

// define finalizes a theme literal: unset extended roles are filled from the
// core nine (deriveRoles). The exported theme vars are built with it so
// theme.Forge.Color(RoleSurface) is as complete as the registry copy.
func define(t Theme) Theme { return deriveRoles(t) }

// register adds built-in themes. Called from package-level var initialization in
// themes_*.go (`var _ = register(...)`), which Go completes before any init()
// runs — so palette.go's init and unlock.go's index build always see the full
// registry, independent of file-name order. Duplicate keys panic — a programming
// error, caught before main() runs.
func register(ts ...Theme) bool {
	for _, t := range ts {
		if _, dup := registry[t.Key]; dup {
			panic("theme: duplicate theme key " + t.Key)
		}
		t = deriveRoles(t)
		registry[t.Key] = t
		registryOrder = append(registryOrder, t.Key)
	}
	return true
}

// ByKey looks up a registered theme. ok is false for unknown keys.
func ByKey(key string) (Theme, bool) {
	t, ok := registry[key]
	return t, ok
}

// All returns every registered theme with the default (Forge) first, then the rest
// in registration order. We force the default to the front explicitly rather than
// lean on init() filename ordering — that ordering is real but fragile, and the
// picker wants the default at the top deterministically. The returned slice is a
// fresh copy; callers may sort/filter it freely.
func All() []Theme {
	out := make([]Theme, 0, len(registryOrder))
	if t, ok := registry[DefaultKey]; ok {
		out = append(out, t)
	}
	for _, k := range registryOrder {
		if k == DefaultKey {
			continue
		}
		out = append(out, registry[k])
	}
	return out
}
