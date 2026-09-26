package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/theme"
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
func styleFilledButton(b *tview.Button, fill theme.Role) *tview.Button {
	b.SetBackgroundColor(theme.Color(fill))
	b.SetLabelColor(onFill(fill))
	return b
}

// onFill returns the label color for text drawn on a fill role.
func onFill(fill theme.Role) tcell.Color {
	switch fill {
	case theme.RoleAccent:
		return theme.Color(theme.RoleOnAccent)
	case theme.RoleNegative:
		return theme.Color(theme.RoleOnNegative)
	case theme.RoleSelection:
		return theme.Color(theme.RoleSelectionText)
	case theme.RoleChip, theme.RoleSurface, theme.RoleBackground:
		return theme.Color(theme.RoleText)
	default:
		return theme.BestOn(theme.Color(fill))
	}
}

// styleDangerModal turns a tview.Modal into a solid danger panel: Negative fill,
// OnNegative text. Buttons keep tview's defaults (Selection fill / Text), which
// the contrast test guarantees read on every theme.
func styleDangerModal(m *tview.Modal) *tview.Modal {
	m.SetBackgroundColor(theme.Color(theme.RoleNegative))
	m.SetTextColor(theme.Color(theme.RoleOnNegative))
	return m
}

// dangerTag is the inline tag for text inside a danger panel. Every color tag in
// a Negative-filled panel must be this (optionally bold via dangerTagBold) — the
// canvas roles ([red], [yellow], [gray]) are not guaranteed to read on the fill,
// and [red] on the Negative fill is literally invisible.
var (
	dangerTag     = theme.Tag(theme.RoleOnNegative)
	dangerTagBold = "[" + theme.TagName(theme.RoleOnNegative) + "::b]"
)
