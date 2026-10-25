package styletest

import (
	"testing"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapmodel"
)

func TestFault(t *testing.T) {
	all := Rules{Query: Every, Width: Every, EmDash: true}
	beyond := Rules{Query: NonASCII, Width: NonASCII}
	for _, c := range []struct {
		name  string
		rules Rules
		r     rune
		tier  mapmodel.GlyphTier
		want  string
	}{
		{"plain ASCII", all, 'a', mapmodel.TierASCII, ""},
		{"plain box", all, '█', mapmodel.TierUnicode, ""},
		{"past ASCII in the ASCII tier", Rules{}, '█', mapmodel.TierASCII, "non-ASCII rune"},
		{"'?' in every tier", all, '?', mapmodel.TierASCII, "a '?' fallback"},
		{"'?' beyond ASCII only", beyond, '?', mapmodel.TierASCII, ""},
		{"'?' beyond ASCII only, in Unicode", beyond, '?', mapmodel.TierUnicode, "a '?' fallback"},
		{"'?' not checked", Rules{}, '?', mapmodel.TierNerd, ""},
		{"wide rune", beyond, '世', mapmodel.TierUnicode, "not one cell wide"},
		{"wide rune not checked", Rules{}, '世', mapmodel.TierUnicode, ""},
		{"control rune in every tier", all, '\x00', mapmodel.TierASCII, "not one cell wide"},
		{"control rune beyond ASCII only", beyond, '\x00', mapmodel.TierASCII, ""},
		{"em dash", all, '—', mapmodel.TierUnicode, "an em dash"},
		{"em dash not checked", beyond, '—', mapmodel.TierUnicode, ""},
	} {
		if got := c.rules.Fault(c.r, c.tier); got != c.want {
			t.Errorf("%s: fault %q, want %q", c.name, got, c.want)
		}
	}
}

func TestHiddenInk(t *testing.T) {
	red, blue := tcell.NewRGBColor(200, 0, 0), tcell.NewRGBColor(0, 0, 200)
	cell := func(r []rune, fg, bg tcell.Color) tcell.SimCell {
		return tcell.SimCell{Runes: r, Style: tcell.StyleDefault.Foreground(fg).Background(bg)}
	}
	for _, c := range []struct {
		name string
		cell tcell.SimCell
		want bool
	}{
		{"glyph in its own colour", cell([]rune{'x'}, red, red), true},
		{"glyph on another colour", cell([]rune{'x'}, red, blue), false},
		{"blank in its own colour", cell([]rune{' '}, red, red), false},
		{"unwritten cell", cell(nil, red, red), false},
	} {
		if got := HiddenInk(c.cell); got != c.want {
			t.Errorf("%s: %v, want %v", c.name, got, c.want)
		}
	}
}
