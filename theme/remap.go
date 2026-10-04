package theme

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// roleTagNames is the semantic tag vocabulary: each role has a lowercase name
// that works as a tview inline color tag ([accent]…[-], [onaccent:accent]…).
// These are the names new code should write (via Tag/TagFgBg in tags.go).
var roleTagNames = [numRoles]string{
	RoleBackground:    "bg",
	RoleText:          "text",
	RoleDim:           "dim",
	RoleLabel:         "label",
	RoleAccent:        "accent",
	RoleHighlight:     "highlight",
	RolePositive:      "positive",
	RoleNegative:      "negative",
	RoleSelection:     "selection",
	RoleSurface:       "surface",
	RoleBorder:        "border",
	RoleSelectionText: "seltext",
	RoleBright:        "bright",
	RoleWarning:       "warning",
	RoleOnAccent:      "onaccent",
	RoleChip:          "chip",
	RoleOnNegative:    "onnegative",
}

// legacyAliases are the pre-role color names the UI was written with (~900
// inline tags). They are remapped onto roles so every existing [gold]/[green]
// tag retints with no edit. They are part of the theme-owned vocabulary: the
// raw-color guard (ui/theme_guard_test.go) accepts them and rejects every other
// name or hex literal outside theme/.
//
// Discipline (the theming design §3.2): tcell.ColorNames is global, mutable, process-wide
// state — tcell.GetColor("gold") reads the SAME map as tview's [gold] tag parser.
// So overwriting these keys retints every named-color resolution in the process,
// not just inline text tags. We own exactly these keys, and applyRemap rewrites
// ALL of them on every switch so no name is ever left pointing at a stale value.
// Do NOT reach for tcell.ColorNames["gold"] expecting tcell's gold once a theme
// is active — it's the active Accent.
var legacyAliases = []struct {
	name string
	role Role
}{
	{"gold", RoleAccent},
	{"gray", RoleDim},
	{"cyan", RoleLabel},
	{"green", RolePositive},
	{"red", RoleNegative},
	{"yellow", RoleHighlight},
	{"white", RoleText},
}

// remappedNames is every tcell color name the theme owns (role names + legacy
// aliases). Exposed to tests via TagNames.
func remappedNames() []struct {
	name string
	role Role
} {
	out := make([]struct {
		name string
		role Role
	}, 0, len(legacyAliases)+int(numRoles))
	out = append(out, legacyAliases...)
	for r := Role(0); r < numRoles; r++ {
		out = append(out, struct {
			name string
			role Role
		}{roleTagNames[r], r})
	}
	return out
}

// TagNames returns every inline-tag color name the theme owns, mapped to its
// role. The raw-color guard uses it as the allow-list.
func TagNames() map[string]Role {
	m := make(map[string]Role)
	for _, n := range remappedNames() {
		m[n.name] = n.role
	}
	return m
}

// applyRemap retints the owned tcell color names to the given theme's role
// colors and points tview.Styles at late-bound role sentinels. After this runs,
// the next Draw retints every existing inline tag (tview re-resolves named tags
// through tcell.ColorNames on every Draw) and every widget's chrome (WrapScreen
// resolves the sentinels on every SetContent).
//
// applyRemap does NOT call app.Draw — the caller (UI layer) owns the redraw. It is
// safe to call repeatedly; every call overwrites the full set.
func applyRemap(t Theme) {
	for _, m := range remappedNames() {
		tcell.ColorNames[m.name] = t.Color(m.role)
	}

	// Chrome defaults, read by tview at widget construction. They are Refs, not
	// concrete colors, so a widget built under one theme still draws the next one
	// correctly — the old "existing widgets keep the construction-time canvas"
	// failure mode (the reason a light theme could not work) is gone. See
	// screen.go. Every field is set, including the ones tview uses for inverse
	// states (button activation, list selection, dropdowns), so no tview default
	// color leaks through.
	tview.Styles.PrimitiveBackgroundColor = Ref(RoleBackground)
	tview.Styles.ContrastBackgroundColor = Ref(RoleSelection)
	tview.Styles.MoreContrastBackgroundColor = Ref(RoleChip)
	tview.Styles.BorderColor = Ref(RoleBorder)
	tview.Styles.TitleColor = Ref(RoleAccent)
	tview.Styles.GraphicsColor = Ref(RoleBorder)
	tview.Styles.PrimaryTextColor = Ref(RoleText)
	tview.Styles.SecondaryTextColor = Ref(RoleDim)
	tview.Styles.TertiaryTextColor = Ref(RoleLabel)
	tview.Styles.InverseTextColor = Ref(RoleBackground) // drawn on a PrimaryTextColor (Text) fill
	tview.Styles.ContrastSecondaryTextColor = Ref(RoleDim)
}
