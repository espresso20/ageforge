package mapmodel_test

import (
	"testing"

	"github.com/rivo/uniseg"

	"github.com/espresso20/ageforge/mapmodel"
)

// TestGlyphTiers: every symbol draws one cell in every tier. The ASCII tier
// is printable ASCII, the Unicode tier never uses a double-width (emoji
// presentation) code point.
func TestGlyphTiers(t *testing.T) {
	for i, g := range mapmodel.AllGlyphs() {
		if g.ASCII < 0x20 || g.ASCII > 0x7e {
			t.Errorf("symbol %d: ASCII glyph %q is not printable ASCII", i+1, g.ASCII)
		}
		if w := uniseg.StringWidth(string(g.Unicode)); w != 1 {
			t.Errorf("symbol %d: Unicode glyph %q (U+%04X) is %d cells wide", i+1, g.Unicode, g.Unicode, w)
		}
		for _, tier := range []mapmodel.GlyphTier{mapmodel.TierASCII, mapmodel.TierUnicode} {
			if r := g.In(tier); uniseg.StringWidth(string(r)) != 1 {
				t.Errorf("symbol %d in tier %s: %q is not one cell", i+1, tier, r)
			}
		}
	}
	if _, ok := mapmodel.ParseTier("nerd"); ok {
		t.Error("ParseTier accepted nerd, which is gone")
	}
	if _, ok := mapmodel.ParseTier("emoji"); ok {
		t.Error("ParseTier accepted an unknown tier")
	}
}

// TestFold: the ASCII tier folds every art rune the maps draw to ASCII and
// leaves the other tiers alone.
func TestFold(t *testing.T) {
	for _, r := range "█▓▒░▀▄▌▐─│┼╔═╗║╚╝▲▼◆●○◘■□▪·•°☺☻♣♠♦≈∩⌂§Ω♫◊†¤λ↑Ψ×▸★☼☾╱╲╳⁵⁺" {
		if a := mapmodel.Fold(r, mapmodel.TierASCII); a < 0x20 || a > 0x7e {
			t.Errorf("Fold(%q) = %q, not ASCII", r, a)
		}
		if u := mapmodel.Fold(r, mapmodel.TierUnicode); u != r {
			t.Errorf("unicode tier changed %q to %q", r, u)
		}
	}
}
