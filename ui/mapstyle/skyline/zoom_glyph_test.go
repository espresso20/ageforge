package skyline

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/mapmodel/fixture"
)

// block is the four cells (upper left, upper right, lower left, lower right)
// a world cell becomes at the close level.
func block(c cell, tier mapmodel.GlyphTier) [4]cell {
	var out [4]cell
	for q := range out {
		out[q] = zoomCell(c, q, tier)
	}
	return out
}

func count(b [4]cell, r rune) (n int) {
	for _, c := range b {
		if c.ch == r {
			n++
		}
	}
	return n
}

// TestCloseFeaturesOnce: a feature glyph (a face, a heart, a letter, a
// flag) is drawn once in its block, never in all four cells, and the other
// cells show what is behind it. A shape (a head, a dot, a window, a tree) is
// redrawn from quadrants and so is never the glyph repeated.
func TestCloseFeaturesOnce(t *testing.T) {
	ink, behind := tcell.NewRGBColor(230, 230, 230), tcell.NewRGBColor(20, 30, 90)
	for _, tier := range []mapmodel.GlyphTier{mapmodel.TierUnicode, mapmodel.TierASCII, mapmodel.TierNerd} {
		for _, r := range "☻☺♥♡❤♠♣♦†✕×►◄▸^*\",.oOvAS7◇" {
			b := block(cell{ch: r, fg: ink, bg: behind, d: 20}, tier)
			if n := count(b, r); n != 1 {
				t.Errorf("tier %v: %q is drawn %d times in its block, want once", tier, r, n)
			}
			for q, c := range b {
				if c.ch == r {
					continue
				}
				if c.ch != ' ' || c.bg != behind || c.d != 20 {
					t.Errorf("tier %v: the cell beside %q (quadrant %d) is %q on %v depth %d, want what is behind it", tier, r, q, c.ch, c.bg, c.d)
				}
			}
		}
	}
	for _, r := range "●◘•·∙▪▫▲◠" {
		b := block(cell{ch: r, fg: ink, bg: behind, d: 20}, mapmodel.TierUnicode)
		if n := count(b, r); n != 0 {
			t.Errorf("%q is repeated %d times: a shape is redrawn at twice the size", r, n)
		}
		ink4 := 0
		for _, c := range b {
			if c.ch != ' ' {
				ink4++
				if _, ok := quadMask[c.ch]; !ok {
					t.Errorf("%q becomes %q, want a block glyph", r, c.ch)
				}
			}
		}
		if ink4 == 0 {
			t.Errorf("%q vanishes at the close level", r)
		}
		// the ASCII tier has no quadrants: the glyph is a feature there
		if a := block(cell{ch: r, fg: ink, bg: behind, d: 20}, mapmodel.TierASCII); r < 0x80 && count(a, r) > 1 {
			t.Errorf("ascii %q is repeated", r)
		}
	}
	// a head is a round blob two cells wide and two tall
	head := block(cell{ch: '●', fg: ink, bg: behind, d: 20}, mapmodel.TierUnicode)
	if head[0].ch != '▟' || head[1].ch != '▙' || head[2].ch != '▜' || head[3].ch != '▛' {
		t.Errorf("a head is %q%q/%q%q", head[0].ch, head[1].ch, head[2].ch, head[3].ch)
	}
	// strokes and surfaces thicken: the glyph is in all four cells
	for _, r := range "│─═▓░▒■" {
		if n := count(block(cell{ch: r, fg: ink, bg: behind, d: 20}, mapmodel.TierUnicode), r); n != 4 {
			t.Errorf("%q is in %d of its block's cells, want 4", r, n)
		}
	}
	// a lower eighth block keeps its height at twice the scale
	for in, want := range map[rune][4]rune{'▂': {' ', ' ', '▄', '▄'}, '▅': {'▂', '▂', '█', '█'}, '▇': {'▆', '▆', '█', '█'}} {
		b := block(cell{ch: in, fg: ink, bg: behind, d: 20}, mapmodel.TierUnicode)
		if got := [4]rune{b[0].ch, b[1].ch, b[2].ch, b[3].ch}; got != want {
			t.Errorf("%q becomes %q, want %q", in, string(got[:]), string(want[:]))
		}
	}
	// a diagonal runs across the block once, a triangle fills its half
	if b := block(cell{ch: '╲', fg: ink, bg: behind, d: 20}, mapmodel.TierUnicode); b[0].ch != '╲' || b[3].ch != '╲' || b[1].ch != ' ' || b[2].ch != ' ' {
		t.Errorf("a falling diagonal is %q%q/%q%q", b[0].ch, b[1].ch, b[2].ch, b[3].ch)
	}
	if b := block(cell{ch: '◢', fg: ink, bg: behind, d: 20}, mapmodel.TierUnicode); b[0].ch != ' ' || b[1].ch != '◢' || b[2].ch != '◢' || b[3].ch != '█' {
		t.Errorf("a lower right triangle is %q%q/%q%q", b[0].ch, b[1].ch, b[2].ch, b[3].ch)
	}
}

// TestCloseCoupleOnce is the owner's screenshot as a unit: two people side
// by side (a head over a body each) with a heart above them. At the close
// level each head is one shape, and the heart is one heart: not a 2x2 block
// of faces or of hearts.
func TestCloseCoupleOnce(t *testing.T) {
	ink, body, sky := tcell.NewRGBColor(240, 200, 170), tcell.NewRGBColor(180, 40, 40), tcell.NewRGBColor(20, 30, 90)
	src := &fb{}
	src.reset(8, 3+6) // six scene rows
	for y := 0; y < 6; y++ {
		for x := 0; x < 8; x++ {
			src.set(x, y+1, ' ', sky, sky, dSky)
		}
	}
	src.set(2, 2, '♥', body, sky, dRow0)
	src.set(1, 3, '☻', ink, sky, dRow0)
	src.set(1, 4, '▙', body, sky, dRow0)
	src.set(3, 3, '☻', ink, sky, dRow0)
	src.set(3, 4, '▟', body, sky, dRow0)
	dst := &fb{}
	dst.reset(16, 3+12)
	magnify(dst, src, 12, 0, 0, mapmodel.TierUnicode)
	var text []string
	for y := 1; y <= 12; y++ {
		var sb strings.Builder
		for x := 0; x < 16; x++ {
			sb.WriteRune(dst.at(x, y).ch)
		}
		text = append(text, sb.String())
	}
	all := strings.Join(text, "\n")
	if n := strings.Count(all, "♥"); n != 1 {
		t.Errorf("the heart is drawn %d times:\n%s", n, all)
	}
	if n := strings.Count(all, "☻"); n != 2 {
		t.Errorf("the couple has %d faces, want 2:\n%s", n, all)
	}
	assertNoClump(t, dst, mapmodel.TierUnicode, "a couple")
}

// assertNoClump fails on a 2x2 window of the same feature glyph anywhere in
// the frame: the signature of a glyph repeated by the close level.
func assertNoClump(t *testing.T, f *fb, tier mapmodel.GlyphTier, what string) {
	t.Helper()
	for y := 1; y+1 < f.h; y++ {
		for x := 0; x+1 < f.w; x++ {
			r := f.at(x, y).ch
			if r == ' ' || r == 0 || glyphKindOf(r, tier) != gOnce {
				continue
			}
			if f.at(x+1, y).ch == r && f.at(x, y+1).ch == r && f.at(x+1, y+1).ch == r {
				t.Errorf("%s: %q is drawn as a 2x2 block at column %d, row %d", what, r, x, y)
				return
			}
		}
	}
}

// TestCloseFramesHaveNoClumps draws real towns at the close level, with
// their people, animals, fires and traffic, and finds no feature glyph
// repeated into a 2x2 block (the four-eyed people).
func TestCloseFramesHaveNoClumps(t *testing.T) {
	for _, age := range []string{"primitive_age", "stone_age", "bronze_age", "iron_age", "medieval_age", "industrial_age", "modern_age", "cyberpunk_age"} {
		for _, tod := range []float64{0.45, 0.95} {
			for _, tier := range []mapmodel.GlyphTier{mapmodel.TierUnicode, mapmodel.TierASCII} {
				m := build(fixture.Options{Age: age, Seed: 3, Routes: 4, Tick: tickAt(tod)}, 1)
				for _, di := range []int{0, len(m.Skyline.Districts) - 1} {
					v := newView()
					v.zoom, v.sw = zClose, 160
					v.follow = false
					v.cam = m.Skyline.Districts[di].Centre - zClose.cols(160)/2
					for k := 0; k < 3; k++ {
						draw(v, m, 160, 45, 40+k*5, tier, false)
						assertNoClump(t, &v.fb, tier, age)
					}
				}
			}
		}
	}
}
