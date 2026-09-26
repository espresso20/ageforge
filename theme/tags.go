package theme

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
)

// tags.go is the inline-tag half of the theme API. Every helper here emits a
// NAMED role tag ([accent], [onaccent:accent:b], …), never a hex literal, because
// named tags are late-bound: tview resolves them through tcell.ColorNames on every
// Draw, so text set once at construction still retints on a live theme switch.
// A hex tag computed at format time would freeze the color of whatever theme was
// active then. HexTag (derive.go) exists only for colors that are deliberately
// not the active theme's (the picker's swatches of other themes).

// TagName returns the inline-tag color name for role ("accent", "dim", …).
func TagName(role Role) string {
	if role < 0 || role >= numRoles {
		return "text"
	}
	return roleTagNames[role]
}

// Tag returns the late-bound inline color tag for role, e.g. "[accent]".
func Tag(role Role) string {
	return "[" + TagName(role) + "]"
}

// TagFgBg returns a late-bound fg+bg tag, e.g. "[onaccent:accent]".
func TagFgBg(fg, bg Role) string {
	return "[" + TagName(fg) + ":" + TagName(bg) + "]"
}

// TagFgBgAttr returns a late-bound fg+bg+attributes tag, e.g.
// "[onaccent:accent:b]". attrs uses tview's attribute letters ("b", "u", "i", …).
func TagFgBgAttr(fg, bg Role, attrs string) string {
	return "[" + TagName(fg) + ":" + TagName(bg) + ":" + attrs + "]"
}

// Reset closes any tag opened by the helpers above (fg, bg and attributes).
const Reset = "[-:-:-]"

// Keycap renders a keyboard key as a filled cap: OnAccent text on an Accent fill,
// bold. It replaces the old "[black:gold:b] key " literal, which left black text on
// whatever Accent a theme picked (unreadable on a dark accent).
func Keycap(key string) string {
	return TagFgBgAttr(RoleOnAccent, RoleAccent, "b") + " " + key + " " + Reset
}

// KeycapButton renders a key cap followed by its action label on a Chip fill —
// the footer-button idiom ("[Enter] Keep"). Replaces the "[white:#30363d:b]"
// label literal, whose fixed dark chip vanished under dark text on a light theme.
func KeycapButton(key, label string) string {
	return TagFgBgAttr(RoleOnAccent, RoleAccent, "b") + " " + key + " " +
		TagFgBgAttr(RoleText, RoleChip, "b") + " " + label + " " + Reset
}

// Selected renders a label as the active/selected chip in a menu column
// (OnAccent on Accent, not bold) — the sidebar's current-overlay marker.
func Selected(label string) string {
	return TagFgBg(RoleOnAccent, RoleAccent) + label + "[-:-]"
}

// LegibleTag returns an early-bound hex tag for a fixed identity hue, corrected
// for legibility against the active background (Legible at 4.5:1). Use it for
// data-driven colors that must keep their own hue (epoch colors) rather than map
// to a role. The text it tags should be rebuilt on refresh, as overlays are.
func LegibleTag(c tcell.Color) string {
	return HexTag(LegibleOnActive(c))
}

// NameTag returns a tag for a data-supplied color name (e.g. config/epochs.go's
// "magenta"). Theme-owned names (role names and legacy aliases) pass through as
// late-bound named tags; any other tcell color name is resolved to its literal
// hue and legibility-corrected via LegibleTag. Unknown names fall back to Text.
func NameTag(name string) string {
	n := strings.ToLower(strings.TrimSpace(name))
	if _, owned := TagNames()[n]; owned {
		return "[" + n + "]"
	}
	if c, ok := cssColors[n]; ok {
		return LegibleTag(c)
	}
	return Tag(RoleText)
}

// cssColors snapshots tcell's named colors at package init, BEFORE applyRemap
// overwrites the theme-owned keys, so NameTag resolves e.g. "blue" to the real
// CSS blue rather than whatever a later remap stored.
var cssColors = func() map[string]tcell.Color {
	m := make(map[string]tcell.Color, len(tcell.ColorNames))
	for k, v := range tcell.ColorNames {
		m[k] = v
	}
	return m
}()

// Sprintf-style convenience: Paint wraps s in role's tag and a fg reset.
func Paint(role Role, s string) string {
	return fmt.Sprintf("%s%s[-]", Tag(role), s)
}
