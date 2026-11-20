package peopledash

import (
	"strings"
	"testing"
)

func TestBlockCountRoundsUp(t *testing.T) {
	for n, want := range map[int]int{0: 0, 1: 1, 5: 1, 6: 2, 168: 34, 553: 111} {
		if got := BlockCount(n); got != want {
			t.Errorf("BlockCount(%d) = %d, want %d", n, got, want)
		}
	}
}

func TestMinusIsTwoPixelsThick(t *testing.T) {
	for _, b := range []bool{true, false} {
		rows := 0
		for _, r := range glyph('-', b) {
			if strings.Contains(r, "1") {
				rows++
			}
		}
		if rows != 2 {
			t.Errorf("minus (bold=%v) is %d rows thick, want 2", b, rows)
		}
	}
}

func TestGlyphsAreRectangular(t *testing.T) {
	for name, set := range map[string]map[rune]string{"bold": bold, "thin": thin} {
		for r, g := range set {
			rows := strings.Split(g, "/")
			for _, row := range rows {
				if len(row) != len(rows[0]) {
					t.Errorf("%s glyph %q has uneven rows", name, r)
				}
			}
		}
	}
}

func TestNumWidth(t *testing.T) {
	if got := NumW("42", true, 5); got != 7*5*2+5 {
		t.Errorf("NumW = %d", got)
	}
	if NumW("", false, 4) != 0 {
		t.Error("empty numeral has width")
	}
}

func TestTrendNeedsTwoSamples(t *testing.T) {
	c := New(60, 30)
	c.Trend(0, 0, 60, 30, []float64{1}, 0, 2, Work, Tile)
	for _, p := range c.Img.Pix {
		if p != 0 {
			t.Fatal("a single sample drew something")
		}
	}
	c.Trend(0, 0, 60, 30, []float64{1, 2, 1.5}, 0, 2, Work, Tile)
	if c.Img.Pix[0] == 0 && c.Img.Pix[len(c.Img.Pix)-1] == 0 {
		t.Fatal("two or more samples drew nothing")
	}
}

func TestMoraleZones(t *testing.T) {
	if MoraleZone(.39) != Red || MoraleZone(.45) != Amber || MoraleZone(.6) != Green {
		t.Error("zone edges moved")
	}
}
