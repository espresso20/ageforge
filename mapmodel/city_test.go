package mapmodel

import (
	"regexp"
	"strings"
	"testing"

	"github.com/rivo/uniseg"

	"github.com/espresso20/ageforge/config"
)

func ageOf(t *testing.T, key string) int {
	t.Helper()
	for i, k := range config.AgeOrder() {
		if k == key {
			return i
		}
	}
	t.Fatalf("no age %s", key)
	return -1
}

// TestCityLooks: the Earth arc is the Modern Age to the Fusion Age, oldest
// first, and every age has a palette, a signature structure, a signature
// mover that is about in that age, and an ambient effect.
func TestCityLooks(t *testing.T) {
	want := []string{"modern_age", "information_age", "digital_age", "cyberpunk_age", "fusion_age"}
	looks := CityLooks()
	if len(looks) != len(want) {
		t.Fatalf("%d looks, want %d", len(looks), len(want))
	}
	for i, l := range looks {
		switch {
		case l.Key != want[i] || l.Age != ageOf(t, want[i]):
			t.Errorf("look %d is %s (age %d), want %s", i, l.Key, l.Age, want[i])
		case len(l.Palette) < 4:
			t.Errorf("%s: a palette of %d colours", l.Key, len(l.Palette))
		case l.Name == "" || l.Structure == "" || l.Ambient == "":
			t.Errorf("%s: no name, structure or ambient effect", l.Key)
		case !l.Mover.Info().In(l.Age):
			t.Errorf("%s: its signature mover, the %s, is not about", l.Key, l.Mover.Info().Name)
		case l.Smog < 0 || l.Smog > 1:
			t.Errorf("%s: smog %v", l.Key, l.Smog)
		}
		if got, ok := CityLookAt(l.Age); !ok || got.Key != l.Key {
			t.Errorf("CityLookAt(%d) = %s, want %s", l.Age, got.Key, l.Key)
		}
		for _, s := range []string{l.Name, l.Structure, l.Ambient} {
			if strings.ContainsAny(s, "—!") {
				t.Errorf("%s: %q breaks the house style", l.Key, s)
			}
		}
	}
	if _, ok := CityLookAt(ageOf(t, "atomic_age")); ok {
		t.Error("the Atomic Age has an Earth-arc look")
	}
	if _, ok := CityLookAt(ageOf(t, "space_age")); ok {
		t.Error("the Space Age has an Earth-arc look (it is the Sky arc's)")
	}
	// the signature movers differ, so every age brings its own
	seen := map[Mover]string{}
	for _, l := range looks {
		if o, dup := seen[l.Mover]; dup {
			t.Errorf("%s and %s share a signature mover", o, l.Key)
		}
		seen[l.Mover] = l.Key
	}
}

// TestGreeneryFade is the greenery rule: all of the green up to the Modern
// Age, about half in the Information Age, the reserve's 15% in the Digital
// Age and none from the Cyberpunk Age on, never growing back.
func TestGreeneryFade(t *testing.T) {
	for _, c := range []struct {
		age  string
		want float64
	}{
		{"primitive_age", 1}, {"atomic_age", 1}, {"modern_age", 1}, {"information_age", 0.5},
		{"digital_age", 0.15}, {"cyberpunk_age", 0}, {"fusion_age", 0}, {"space_age", 0}, {"transcendent_age", 0},
	} {
		if got := Greenery(ageOf(t, c.age)); got != c.want {
			t.Errorf("%s keeps %v of the green, want %v", c.age, got, c.want)
		}
	}
	last := 2.0
	for a := range config.AgeOrder() {
		g := Greenery(a)
		if g > last {
			t.Errorf("age %d grows the green back (%v after %v)", a, g, last)
		}
		last = g
	}
	if SmogLevel(ageOf(t, "modern_age")) != 0 || SmogLevel(ageOf(t, "cyberpunk_age")) <= SmogLevel(ageOf(t, "digital_age")) ||
		SmogLevel(ageOf(t, "fusion_age")) >= SmogLevel(ageOf(t, "cyberpunk_age")) {
		t.Error("the smog should rise to the megacity's and thin once fusion cleans the air")
	}
}

// TestCityRoster: every city feature has a key, names, a line in the house
// style and the ages it stands in; the Modern Age is the first to build any
// and every Earth-arc age stands something of its own.
func TestCityRoster(t *testing.T) {
	ages := config.AgeOrder()
	retired := regexp.MustCompile(`(?i)\b(colour|centre|harbour|favour|villager|neighbour)`)
	keys := map[string]bool{}
	for f := CityFeature(1); f < NumCityFeatures; f++ {
		i := f.Info()
		switch {
		case i.Key == "" || i.Name == "" || i.Title == "" || len(i.Lines) == 0:
			t.Errorf("feature %d is missing a key, name, title or line: %+v", f, i)
		case keys[i.Key]:
			t.Errorf("feature key %q twice", i.Key)
		case i.From < ageOf(t, "modern_age") || i.From >= len(ages):
			t.Errorf("%s: brought by age %d, outside the Earth arc", i.Key, i.From)
		case i.Until >= 0 && i.Until < i.From:
			t.Errorf("%s: gives way (%d) before it stands (%d)", i.Key, i.Until, i.From)
		}
		keys[i.Key] = true
		for _, s := range append([]string{i.Name, i.Title}, i.Lines...) {
			if strings.ContainsAny(s, "—!") || retired.MatchString(s) {
				t.Errorf("%s: %q breaks the house style", i.Key, s)
			}
		}
		if !Built(f, i.From) || Built(f, i.From-1) {
			t.Errorf("%s: Built disagrees with From %d", i.Key, i.From)
		}
	}
	if n := len(FeaturesAt(ageOf(t, "atomic_age"))); n != 0 {
		t.Errorf("the Atomic Age stands %d city features", n)
	}
	for _, l := range CityLooks() {
		own := 0
		for _, f := range FeaturesAt(l.Age) {
			if f.Info().From == l.Age {
				own++
			}
		}
		if own == 0 {
			t.Errorf("%s brings no city feature of its own", l.Key)
		}
	}
}

// TestCityGlyphs: the city block's symbols draw one cell in every tier (a
// printable ASCII mark, a single-width Unicode symbol that is not an emoji,
// a Nerd Font icon from the Private Use Area), G finds them past the core
// table, AllGlyphs lists them, and every art rune the Earth arc draws folds
// to ASCII.
func TestCityGlyphs(t *testing.T) {
	all := map[Glyph]bool{}
	for _, g := range AllGlyphs() {
		all[g] = true
	}
	for s := symCityBase; s < symCityEnd; s++ {
		g := G(s)
		if g != cityGlyphs[s-symCityBase] {
			t.Errorf("G(%d) = %+v, not the city block's", s, g)
		}
		if !all[g] {
			t.Errorf("AllGlyphs leaves out symbol %d", s)
		}
		if g.ASCII < 0x21 || g.ASCII > 0x7e {
			t.Errorf("symbol %d: ASCII %q is not a printable mark", s, g.ASCII)
		}
		if uniseg.StringWidth(string(g.Unicode)) != 1 || g.Unicode < 0x80 {
			t.Errorf("symbol %d: Unicode %q is not one cell or is plain ASCII", s, g.Unicode)
		}
		if g.Nerd < 0xE000 || g.Nerd > 0xF8FF || uniseg.StringWidth(string(g.Nerd)) != 1 {
			t.Errorf("symbol %d: Nerd glyph U+%04X is not a single-width Private Use Area icon", s, g.Nerd)
		}
		for _, tier := range []GlyphTier{TierASCII, TierUnicode, TierNerd} {
			if r := R(s, tier); uniseg.StringWidth(string(r)) != 1 {
				t.Errorf("symbol %d in tier %s: %q is not one cell", s, tier, r)
			}
		}
		if a := Fold(g.Unicode, TierASCII); a < 0x20 || a > 0x7e {
			t.Errorf("symbol %d: %q folds to %q", s, g.Unicode, a)
		}
	}
	if G(symCityEnd) != glyphTable[SymNone] || G(symCityBase-1) != glyphTable[SymNone] {
		t.Error("a symbol outside every block is not the empty one")
	}
	for r := range cityFold {
		if a := Fold(r, TierASCII); a < 0x20 || a > 0x7e || a == '?' {
			t.Errorf("%q folds to %q", r, a)
		}
		if uniseg.StringWidth(string(r)) != 1 {
			t.Errorf("%q is not one cell wide", r)
		}
	}
}
