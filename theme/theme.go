// Package theme is AgeForge's single source of truth for UI color.
//
// It is a leaf package: it depends only on gdamore/tcell/v2 (plus rivo/tview in
// remap.go, solely to push chrome defaults into tview.Styles). It must NOT import
// game/ or ui/ — themes are pure presentation and an import cycle here would be a
// design smell. See design-and-architecture/theming.md §3.
//
// The model is ~9 semantic *roles* (not literal color names): Positive is "the
// gains color," whatever hue the active theme picks. Inline tview tags like
// [green] are retinted onto these roles via the name-remap in remap.go.
package theme

import "github.com/gdamore/tcell/v2"

// Role enumerates the semantic color slots a theme must fill. Order is fixed and
// load-bearing: themes declare their palette as a [numRoles]tcell.Color indexed by
// these constants (see theming.md §3.1).
type Role int

const (
	RoleBackground Role = iota // canvas / primitive background
	RoleText                   // primary readable text
	RoleDim                    // secondary / hints / disabled
	RoleLabel                  // field labels, values
	RoleAccent                 // titles, borders, brand highlights
	RoleHighlight              // numbers, attention, "look here"
	RolePositive               // gains, success, +deltas
	RoleNegative               // losses, errors, -deltas
	RoleSelection              // selected list-row background
	numRoles
)

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

	Colors [numRoles]tcell.Color

	// Signed sentinels for the ± distinction in accessible themes (theming.md §4):
	// the sign is encoded by shape as well as hue so colorblind players never rely
	// on color alone. Non-accessible themes may leave these empty.
	GainGlyph string // e.g. "▲" / "+"
	LossGlyph string // e.g. "▼" / "-"

	// Milestone-gated unlock condition (theming.md §5). A gated (flavor) theme
	// declares EXACTLY ONE of these — the milestone key or chain key whose
	// completion unlocks it account-wide. The mapping lives here, in the registry,
	// not scattered through engine/milestone code: theme stays a leaf package, so
	// these are plain strings (no game import), and the UI reverse-maps a completed
	// key back to a theme via UnlockedBy.
	//
	// Always-available themes (Accessible, or the default Forge) leave BOTH empty —
	// they're never gated, so there is nothing to unlock. The registry-consistency
	// test (unlock_test.go) enforces the XOR for gated themes and the empty-pair for
	// the always-available set.
	UnlockMilestone string // milestone key (config/milestones.go) — XOR with UnlockChain
	UnlockChain     string // milestone-chain key — XOR with UnlockMilestone

	// UnlockHint is the human-readable unlock condition shown for a LOCKED theme in
	// the picker detail pane and `theme list` (e.g. "Reach the Cyberpunk Age").
	// Required for gated themes; empty for always-available ones.
	UnlockHint string
}

// Gated reports whether the theme is milestone-gated (declares an unlock
// condition). Always-available themes (Accessible / Forge) are not gated.
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

// registry holds every built-in theme, keyed by Key. Populated by the
// themes_*.go init() functions via register().
var registry = map[string]Theme{}

// registryOrder preserves insertion order so All() is deterministic (Forge first).
var registryOrder []string

// register adds a built-in theme. Called from themes_*.go init(). Duplicate keys
// panic at init — a programming error, caught before main() runs.
func register(t Theme) {
	if _, dup := registry[t.Key]; dup {
		panic("theme: duplicate theme key " + t.Key)
	}
	registry[t.Key] = t
	registryOrder = append(registryOrder, t.Key)
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
