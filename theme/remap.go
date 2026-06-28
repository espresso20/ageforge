package theme

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// remappedNames is the FIXED set of tcell color-name keys the theme system owns,
// paired with the role each is retinted to. This is the load-bearing table from
// theming.md §3.2.
//
// Discipline (theming.md §3.2): tcell.ColorNames is global, mutable, process-wide
// state — tcell.GetColor("gold") reads the SAME map as tview's [gold] tag parser.
// So overwriting these keys retints every named-color resolution in the process,
// not just inline text tags. We own exactly these keys, and applyRemap rewrites
// ALL of them on every switch so no name is ever left pointing at a stale value.
// Do NOT reach for tcell.ColorNames["gold"] expecting tcell's gold once a theme
// is active — it's the active Accent.
var remappedNames = []struct {
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

// applyRemap retints the fixed set of named tcell colors to the given theme's role
// colors and pushes chrome defaults into tview.Styles. After this runs, the next
// Draw cycle retints every existing named inline tag ([gold], [green], …) for free,
// because tview re-resolves named tags through tcell.ColorNames on every Draw.
//
// applyRemap does NOT call app.Draw — the caller (UI layer) owns the redraw. It is
// safe to call repeatedly; every call overwrites the full set.
func applyRemap(t Theme) {
	for _, m := range remappedNames {
		tcell.ColorNames[m.name] = t.Color(m.role)
	}

	// Chrome defaults for *future* widgets (existing ones are handled by Restyle).
	// tview.Styles is read inside NewBox() at construction time (theming.md §2).
	tview.Styles.PrimitiveBackgroundColor = t.Color(RoleBackground)
	tview.Styles.BorderColor = t.Color(RoleAccent)
	tview.Styles.TitleColor = t.Color(RoleAccent)
	tview.Styles.PrimaryTextColor = t.Color(RoleText)
	tview.Styles.SecondaryTextColor = t.Color(RoleDim)

	// ContrastBackgroundColor backs selection/active rows. Use the theme's
	// Selection role — it's chosen to read against Text and is what list selection
	// and focused fields lean on.
	tview.Styles.ContrastBackgroundColor = t.Color(RoleSelection)
}
