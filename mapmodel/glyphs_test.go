package mapmodel_test

import (
	"testing"

	"github.com/rivo/uniseg"

	"github.com/espresso20/ageforge/mapmodel"
)

func isPUA(r rune) bool {
	return r >= 0xE000 && r <= 0xF8FF || r >= 0xF0000 && r <= 0xFFFFD || r >= 0x100000 && r <= 0x10FFFD
}

// TestGlyphTiers: every symbol draws one cell in every tier. The ASCII tier
// is printable ASCII, the Unicode tier never uses a double-width (emoji
// presentation) code point, and every Nerd Font icon is a single-width
// Private Use Area glyph with a Unicode fallback.
func TestGlyphTiers(t *testing.T) {
	for i, g := range mapmodel.AllGlyphs() {
		if g.ASCII < 0x20 || g.ASCII > 0x7e {
			t.Errorf("symbol %d: ASCII glyph %q is not printable ASCII", i+1, g.ASCII)
		}
		if w := uniseg.StringWidth(string(g.Unicode)); w != 1 {
			t.Errorf("symbol %d: Unicode glyph %q (U+%04X) is %d cells wide", i+1, g.Unicode, g.Unicode, w)
		}
		if g.Nerd != 0 {
			if !isPUA(g.Nerd) {
				t.Errorf("symbol %d: Nerd glyph U+%04X is not a Private Use Area icon", i+1, g.Nerd)
			}
			if w := uniseg.StringWidth(string(g.Nerd)); w != 1 {
				t.Errorf("symbol %d: Nerd glyph U+%04X is %d cells wide", i+1, g.Nerd, w)
			}
			if g.Unicode == 0 || g.Unicode == ' ' && g.Nerd != 0 {
				t.Errorf("symbol %d: Nerd glyph U+%04X has no Unicode fallback", i+1, g.Nerd)
			}
		}
		for _, tier := range []mapmodel.GlyphTier{mapmodel.TierASCII, mapmodel.TierUnicode, mapmodel.TierNerd} {
			if r := g.In(tier); uniseg.StringWidth(string(r)) != 1 {
				t.Errorf("symbol %d in tier %s: %q is not one cell", i+1, tier, r)
			}
		}
	}
	if _, ok := mapmodel.ParseTier("nerd"); !ok {
		t.Error("ParseTier(nerd) failed")
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
	for s := mapmodel.Sym(1); s < mapmodel.SymBuild+1; s++ {
		n := mapmodel.R(s, mapmodel.TierNerd)
		if a := mapmodel.Fold(n, mapmodel.TierASCII); a > 0x7e {
			t.Errorf("a nerd glyph folds to %q", a)
		}
	}
}
