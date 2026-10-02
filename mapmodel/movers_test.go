package mapmodel

import (
	"regexp"
	"strings"
	"testing"

	"github.com/rivo/uniseg"

	"github.com/espresso20/ageforge/config"
)

// TestMoverRoster: every mover has a key, names, an inspect line, a pace
// and an age that introduces it (on or before the last age it is seen in),
// and every era brings at least one new mover. Player text follows the
// house style: no em dashes, no exclamation marks, American spelling.
func TestMoverRoster(t *testing.T) {
	ages := config.AgeOrder()
	keys := map[string]bool{}
	retired := regexp.MustCompile(`(?i)\b(colour|centre|harbour|favour|villager)`)
	for k := Mover(1); k < NumMovers; k++ {
		i := k.Info()
		switch {
		case i.Key == "" || i.Name == "" || i.Title == "" || len(i.Lines) == 0:
			t.Errorf("mover %d is missing a key, name, title or line: %+v", k, i)
		case keys[i.Key]:
			t.Errorf("mover key %q twice", i.Key)
		case i.From < 0 || i.From >= len(ages):
			t.Errorf("%s: no age introduces it (From %d)", i.Key, i.From)
		case i.Until >= 0 && i.Until < i.From:
			t.Errorf("%s: retires (%d) before it arrives (%d)", i.Key, i.Until, i.From)
		case i.Pace < 1 || i.Pace > 12:
			t.Errorf("%s: pace %d", i.Key, i.Pace)
		case i.Sym == SymNone:
			t.Errorf("%s: no symbol", i.Key)
		}
		keys[i.Key] = true
		for _, s := range append([]string{i.Name, i.Title}, i.Lines...) {
			if strings.ContainsAny(s, "—!") || retired.MatchString(s) {
				t.Errorf("%s: %q breaks the house style", i.Key, s)
			}
		}
		if !Introduced(k, i.From) || i.From > 0 && Introduced(k, i.From-1) {
			t.Errorf("%s: Introduced disagrees with From %d", i.Key, i.From)
		}
	}
	// people stroll; nothing moves faster than a cell a frame
	if p := MoverWalker.Info().Pace; p != 5 {
		t.Errorf("walkers take %d frames a cell, want 5 (the roguelike's walkFrames)", p)
	}
	cat := NewCatalog()
	newIn := map[int]int{}
	for k := Mover(1); k < NumMovers; k++ {
		newIn[cat.EpochOfAge(k.Info().From)]++
	}
	for e := range cat.Epochs {
		if newIn[e] == 0 {
			t.Errorf("the %s brings no new movers", cat.EpochName[e])
		}
	}
	// spot checks on the eras the movers belong to
	for _, c := range []struct {
		k   Mover
		age string
	}{
		{MoverHunter, "primitive_age"}, {MoverOxCart, "iron_age"}, {MoverRowboat, "iron_age"},
		{MoverSteamTrain, "industrial_age"}, {MoverTram, "victorian_age"}, {MoverCar, "modern_age"},
		{MoverMaglev, "cyberpunk_age"}, {MoverShuttle, "space_age"}, {MoverHabitat, "interstellar_age"},
	} {
		if got := ages[c.k.Info().From]; got != c.age {
			t.Errorf("%s arrives in the %s, want the %s", c.k.Info().Key, got, c.age)
		}
	}
	if Introduced(MoverNone, len(ages)-1) || MoverNone.Info().In(0) || Introduced(NumMovers, 0) {
		t.Error("the empty mover is about")
	}
	if len(MoversAt(0)) == 0 || len(MoversAt(len(ages)-1)) == 0 {
		t.Error("the first or last age has no movers")
	}
}

// TestMoverGlyphs: every mover, and the visitor, draws one cell in every
// tier: printable ASCII, a single-width Unicode symbol that is not an emoji,
// and a Nerd Font icon from the Private Use Area. The saucer's rims fold to
// ASCII too.
func TestMoverGlyphs(t *testing.T) {
	syms := []Sym{SymUFO, SymAlien}
	for k := Mover(1); k < NumMovers; k++ {
		syms = append(syms, k.Info().Sym)
	}
	for _, s := range syms {
		g := G(s)
		if g.ASCII < 0x21 || g.ASCII > 0x7e {
			t.Errorf("sym %d: ASCII %q is not a printable mark", s, g.ASCII)
		}
		if uniseg.StringWidth(string(g.Unicode)) != 1 || g.Unicode < 0x80 {
			t.Errorf("sym %d: Unicode %q is not one cell or is plain ASCII", s, g.Unicode)
		}
		if g.Nerd < 0xE000 || g.Nerd > 0xF8FF {
			t.Errorf("sym %d: Nerd glyph U+%04X is not a Private Use Area icon", s, g.Nerd)
		}
		if a := Fold(g.Unicode, TierASCII); a < 0x20 || a > 0x7e {
			t.Errorf("sym %d: %q folds to %q", s, g.Unicode, a)
		}
	}
	for _, r := range "◄►▬▰╪━═░·˚" {
		if a := Fold(r, TierASCII); a < 0x20 || a > 0x7e {
			t.Errorf("%q folds to %q", r, a)
		}
	}
	// movers that share an age differ in every tier, so the legend tells
	// them apart
	for age := range config.AgeOrder() {
		at := MoversAt(age)
		for i, a := range at {
			for _, b := range at[i+1:] {
				ga, gb := G(a.Info().Sym), G(b.Info().Sym)
				if ga.ASCII == gb.ASCII || ga.Unicode == gb.Unicode {
					t.Errorf("age %d: %s and %s look alike", age, a.Info().Key, b.Info().Key)
				}
			}
		}
	}
}
