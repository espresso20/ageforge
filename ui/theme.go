package ui

import (
	"github.com/espresso20/ageforge/theme"
)

// BarFillColor is the tview color tag for filled progress-bar segments. Formerly a
// fixed "#9370DB" literal; now role-derived (Accent) and emitted as a late-bound
// named tag ([accent]) so bars retint with the theme (the theming design §3.4).
func BarFillColor() string { return theme.Tag(theme.RoleAccent) }

// BarEmptyColor is the tview color tag for empty bar segments. Role-derived from
// Dim (was a fixed "#444444"; the theming design §3.4).
func BarEmptyColor() string { return theme.Tag(theme.RoleDim) }
