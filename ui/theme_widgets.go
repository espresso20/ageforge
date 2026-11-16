package ui

import (
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/window/kit"
)

// theme_widgets.go collects the few widget-styling idioms that need more than a
// single role: filled buttons and danger panels. A filled surface always pairs
// its fill with the matching "on" role, so the label can never end up drawn in a
// color that only reads on the canvas (dark ink on a dark-red fill under a light
// theme, or [red] text on a red panel under any theme).

// styleFilledButton paints a button with fill as its background and the role (or
// computed contrast color) that reads on that fill as its label. Accent pairs
// with OnAccent, Negative with OnNegative; any other fill gets black or white,
// whichever reads better (theme.BestOn). Buttons styled here are transient
// (rebuilt with their modal), so construction-time colors are fine.
func styleFilledButton(b *nativeButton, fill theme.Role) *nativeButton {
	b.SetBackgroundColor(theme.Of(fill))
	b.SetLabelColor(onFill(fill))
	return b
}

// onFill returns the label color for text drawn on a fill role.
func onFill(fill theme.Role) kit.Color {
	switch fill {
	case theme.RoleAccent:
		return theme.Of(theme.RoleOnAccent)
	case theme.RoleNegative:
		return theme.Of(theme.RoleOnNegative)
	case theme.RoleSelection:
		return theme.Of(theme.RoleSelectionText)
	case theme.RoleChip, theme.RoleSurface, theme.RoleBackground:
		return theme.Of(theme.RoleText)
	default:
		return theme.BestOn(theme.Of(fill))
	}
}

// styleDangerModal turns a dialog into one that asks about something that
// cannot be undone: its buttons are action buttons, the last, which confirms,
// the danger one and the others plain. The keyboard starts on the first, the
// way out.
func styleDangerModal(m *nativeDialog) *nativeDialog { return m.asActions(true) }

// dangerTag is the inline tag for text inside a danger panel. Every color tag in
// a Negative-filled panel must be this (optionally bold via dangerTagBold) — the
// canvas roles ([red], [yellow], [gray]) are not guaranteed to read on the fill,
// and [red] on the Negative fill is literally invisible.
var (
	dangerTag     = theme.Tag(theme.RoleOnNegative)
	dangerTagBold = "[" + theme.TagName(theme.RoleOnNegative) + "::b]"
)
