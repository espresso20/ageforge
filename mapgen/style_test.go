package mapgen

import (
	"go/parser"
	"go/token"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/rivo/uniseg"

	"github.com/espresso20/ageforge/mapmodel"
)

// TestStylesAreWellFormed: the rows are in age order and name real ages, every
// family has art for every size, the art fits a parcel, and every glyph takes
// one cell.
func TestStylesAreWellFormed(t *testing.T) {
	last := -1
	for _, st := range styles {
		age, ok := testCatalog.AgeIdx[st.From]
		if !ok || age <= last {
			t.Fatalf("row %q is not a later age than the row before it", st.From)
		}
		last = age
		maxRows := lotH
		if st.Scale == Orbit {
			maxRows = stationH
		}
		for fam := range st.Shapes {
			for tier, variants := range st.Shapes[fam] {
				if len(variants) == 0 {
					t.Errorf("%s: family %d has no art for tier %d", st.From, fam, tier)
				}
				for _, a := range variants {
					w, h := a.size()
					if w > lotW || h > maxRows {
						t.Errorf("%s: art %v is %dx%d, bigger than a parcel", st.From, a, w, h)
					}
					checkGlyphs(t, st.From, strings.Join(a, ""))
				}
			}
		}
		if pw, ph := st.Pad.size(); pw > lotW || ph > portH-2 {
			t.Errorf("%s: a launch pad of %dx%d does not fit the spaceport", st.From, pw, ph)
		}
		if st.Pads > 0 && len(st.Pad) == 0 {
			t.Errorf("%s: %d launch pads and no pad art", st.From, st.Pads)
		}
		checkGlyphs(t, st.From, strings.Join(st.Pad, ""))
		for _, look := range st.Land {
			checkGlyphs(t, st.From, look.Glyphs)
		}
		for _, r := range st.Roads {
			checkGlyphs(t, st.From, r.Joins)
		}
		for _, s := range st.Scatter {
			checkGlyphs(t, st.From, s.Glyphs)
		}
		for _, a := range st.Agents {
			checkGlyphs(t, st.From, a.Glyphs)
		}
		checkGlyphs(t, st.From, st.Sky.Glyphs+st.Yard.Glyphs+st.Plaza.Glyphs+st.Rubble.Glyphs+st.Windows)
	}
	for key, tpl := range templates {
		if len(tpl.Tone) != 0 && len(tpl.Tone) != len(tpl.Art) {
			t.Errorf("template %s: %d tone rows for %d art rows", key, len(tpl.Tone), len(tpl.Art))
		}
		for i, row := range tpl.Tone {
			if len(row) < utf8.RuneCountInString(tpl.Art[i]) {
				t.Errorf("template %s: tone row %d is shorter than its art", key, i)
			}
			for _, letter := range []byte(row) {
				if _, ok := defaultTones[letter]; !ok && letter != ' ' {
					t.Errorf("template %s: tone %q is no material", key, letter)
				}
			}
		}
		w, h := tpl.size()
		if w > plotW || h > plotH {
			t.Errorf("template %s is %dx%d, bigger than a plot", key, w, h)
		}
		checkGlyphs(t, key, strings.Join(tpl.Art, ""))
	}
}

func checkGlyphs(t *testing.T, where, glyphs string) {
	t.Helper()
	for _, r := range glyphs {
		if r != ' ' && uniseg.StringWidth(string(r)) != 1 {
			t.Errorf("%s: %q is not one cell wide", where, r)
		}
	}
	if !utf8.ValidString(glyphs) {
		t.Errorf("%s: invalid text", where)
	}
}

// TestLineHuesFollowTheLines: the line hues are one per building line, in
// order, and the last hue is the last line's.
func TestLineHuesFollowTheLines(t *testing.T) {
	if numLines != len(mapmodel.LineageOrder) {
		t.Fatalf("numLines is %d, there are %d lines", numLines, len(mapmodel.LineageOrder))
	}
	for i, line := range mapmodel.LineageOrder {
		if lineHue(line) != HLine+Hue(i) {
			t.Errorf("line %s: hue %d, want %d", line, lineHue(line), HLine+Hue(i))
		}
	}
	if lineHue(mapmodel.LineageOrder[numLines-1]) != NumHues-1 {
		t.Errorf("the last line hue is not the last hue")
	}
}

// TestEveryLineHasItsLook: each building line has a silhouette and a sign, and
// the sign is one cell wide.
func TestEveryLineHasItsLook(t *testing.T) {
	for _, line := range mapmodel.LineageOrder {
		sign, signed := lineSigns[line]
		_, shaped := familyOf[line]
		if !signed || !shaped {
			t.Errorf("line %s: sign %v, family %v", line, signed, shaped)
		}
		checkGlyphs(t, line, string(sign))
	}
}

// TestEveryAgeHasARow: the first row serves the oldest age, so no age is left
// without a style.
func TestEveryAgeHasARow(t *testing.T) {
	ageIdx := func(k string) int { return testCatalog.AgeIdx[k] }
	if ageIdx(styles[0].From) != 0 {
		t.Fatalf("the first row starts at %q, not the first age", styles[0].From)
	}
	for age := range testCatalog.Ages {
		if styleFor(ageIdx, age) == nil {
			t.Fatalf("age %d has no style", age)
		}
	}
}

// TestPackageIsPure: the engine imports nothing of the screen or the ui.
func TestPackageIsPure(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	banned := []string{"github.com/gdamore/tcell", "github.com/rivo/tview", "github.com/espresso20/ageforge/ui", "github.com/espresso20/ageforge/theme"}
	for _, path := range files {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, imp := range f.Imports {
			p, _ := strconv.Unquote(imp.Path.Value)
			if slices.ContainsFunc(banned, func(b string) bool { return strings.HasPrefix(p, b) }) {
				t.Errorf("%s imports %s", path, p)
			}
			if p == "math" || p == "time" || p == "math/rand" {
				t.Errorf("%s imports %s", path, p)
			}
		}
	}
}
