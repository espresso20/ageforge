// Package styletest holds what the map styles' tests share: the checks every
// drawn cell must pass, the frame-time check and the mirror of a glyph. Only
// tests import it.
package styletest

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/uniseg"

	"github.com/espresso20/ageforge/mapmodel"
)

// Scope says in which glyph tiers a rune rule holds.
type Scope uint8

const (
	Off      Scope = iota // the rule is not checked
	NonASCII              // in the tiers beyond ASCII
	Every                 // in every tier
)

func (s Scope) holds(tier mapmodel.GlyphTier) bool {
	return s == Every || s == NonASCII && tier != mapmodel.TierASCII
}

// Rules are the checks the runes of a frame face. Nothing past ASCII may
// show in the ASCII tier, whatever the rules; the sweeps differ in how
// strict they are about the rest, so each one names its own.
type Rules struct {
	Query  Scope // no '?', the canvas's fallback for a rune the tier cannot show
	Width  Scope // every rune one cell wide
	EmDash bool  // no em dash
}

// Fault names the rule that r, drawn in tier, breaks, or is "" when it
// breaks none.
func (k Rules) Fault(r rune, tier mapmodel.GlyphTier) string {
	switch {
	case tier == mapmodel.TierASCII && r >= 0x80:
		return "non-ASCII rune"
	case r == '?' && k.Query.holds(tier):
		return "a '?' fallback"
	case k.Width.holds(tier) && uniseg.StringWidth(string(r)) != 1:
		return "not one cell wide"
	case k.EmDash && r == '—':
		return "an em dash"
	}
	return ""
}

// HiddenInk reports whether c is a drawn glyph (not a blank) whose ink is
// its own background colour, so no one can read it.
func HiddenInk(c tcell.SimCell) bool {
	if len(c.Runes) == 0 || c.Runes[0] == ' ' {
		return false
	}
	fg, bg, _ := c.Style.Decompose()
	return fg == bg
}
