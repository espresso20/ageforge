package logroll

import (
	"bytes"
	"fmt"
	"go/parser"
	"go/token"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func same(a, b *image.RGBA) bool { return a.Rect == b.Rect && bytes.Equal(a.Pix, b.Pix) }

func inkRuns(img *image.RGBA) (tall int) {
	for x := 0; x < img.Rect.Dx(); x++ {
		n := 0
		for y := 0; y < img.Rect.Dy(); y++ {
			if c := img.RGBAAt(x, y); c == Ink {
				n++
			}
		}
		if n >= 10 {
			tall++
		}
	}
	return tall
}

func TestTallyStrokes(t *testing.T) {
	for n := 1; n <= 15; n++ {
		img := Tally(n, 9, 120, 15, Ink, Paper)
		if got, want := inkRuns(img)/2, n-n/5; got != want {
			t.Errorf("n=%d: %d upright strokes, want %d", n, got, want)
		}
	}
	if same(Tally(20, 1, 120, 15, Ink, Paper), Tally(15, 1, 120, 15, Ink, Paper)) == false {
		t.Error("more than 15 strokes must draw 15")
	}
	if same(Tally(8, 1, 80, 15, Ink, Paper), Tally(8, 2, 80, 15, Ink, Paper)) {
		t.Error("strokes of two seeds are identical")
	}
}

func TestDeterministicAndSeeded(t *testing.T) {
	l := Plan(279, 68)
	if !same(Desk(l, 7, 15, 3), Desk(l, 7, 15, 3)) {
		t.Error("the desk must repeat for a seed")
	}
	if same(Desk(l, 7, 15, 3), Desk(l, 7, 15, 4)) {
		t.Error("the desk must vary by seed")
	}
	if !same(SlateBare(300, 200, 5), SlateBare(300, 200, 5)) || same(SlateBare(300, 200, 5), SlateBare(300, 200, 6)) {
		t.Error("swipes must repeat for a seed and vary between seeds")
	}
	for o := 0; o < Offices; o++ {
		if same(Impression(o, 1, 21, 15, Inks[o], Paper), Impression(o, 2, 21, 15, Inks[o], Paper)) {
			t.Errorf("office %d: two impressions identical", o)
		}
	}
}

func TestNoTwoCardsAlike(t *testing.T) {
	var all []*image.RGBA
	for o := 0; o < Offices; o++ {
		all = append(all, Tile(o, true, .5, false, 1, 182, 45))
	}
	for i := range all {
		for j := i + 1; j < len(all); j++ {
			if same(all[i], all[j]) {
				t.Errorf("cards %d and %d are identical", i, j)
			}
			// the faces differ by design; the card shape must too
			d := 0
			for p := 0; p < len(all[i].Pix); p += 4 {
				if all[i].Pix[p] != all[j].Pix[p] || all[i].Pix[p+1] != all[j].Pix[p+1] {
					d++
				}
			}
			if d < 120 {
				t.Errorf("cards %d and %d differ in only %d pixels", i, j, d)
			}
		}
	}
	if same(Tile(2, true, 1, false, 1, 182, 45), Tile(2, false, 1, false, 1, 182, 45)) ||
		same(Tile(2, true, 1, false, 1, 182, 45), Tile(2, true, 1, true, 1, 182, 45)) {
		t.Error("an off card and a wax dot must show")
	}
}

func TestSlateSwipesVisible(t *testing.T) {
	img := SlateBare(400, 300, 11)
	n := 0
	for p := 0; p < len(img.Pix); p += 4 {
		if img.Pix[p] != Slate.R || img.Pix[p+1] != Slate.G {
			n++
		}
	}
	if n < 400*300/12 {
		t.Errorf("swipes cover only %d of %d pixels", n, 400*300)
	}
}

func TestLedgeTools(t *testing.T) {
	l := Plan(279, 68)
	img := Desk(l, 7, 15, 1)
	var chalk, rag int
	for y := l.Ledge.Y * 15; y < l.Ledge.Bottom()*15; y++ {
		for x := l.Ledge.X * 7; x < l.Ledge.Right()*7; x++ {
			switch img.RGBAAt(x, y) {
			case Chalk:
				chalk++
			case hx(0xbfb49a):
				rag++
			}
		}
	}
	if chalk < 56*7 || rag < 90*20 {
		t.Errorf("chalk %d px, rag %d px: too small", chalk, rag)
	}
}

func TestPartsSizes(t *testing.T) {
	l := Plan(279, 68)
	if g := Lever(l, 7, 15, 1).Rect; g.Dx() != l.Lever.W*7 || g.Dy() != l.Lever.H*15 {
		t.Errorf("lever picture %v", g)
	}
	if g := Rail(21, 900, []Mark{{.5, Trouble}, {.2, Seam}, {.9, Headline}}, .3, .6, 1).Rect; g.Dx() != 21 || g.Dy() != 900 {
		t.Errorf("rail picture %v", g)
	}
	for _, k := range []RowKind{RowBand, RowBanner, RowOpen} {
		if g := RowBack(k, 100, 15, 1).Rect; g.Dx() != 100 || g.Dy() != 15 {
			t.Errorf("row %d picture %v", k, g)
		}
	}
	if Bar(60, 15, 2, 1).Rect.Dx() != 60 || Arrow(1, 15, Chalk).Rect.Dx() != 12 || SeamRow(80, 15, 20, 30, 1).Rect.Dy() != 15 {
		t.Error("a small picture has the wrong size")
	}
}

func TestDeskSpeed(t *testing.T) {
	l := Plan(279, 71)
	t0 := time.Now()
	Desk(l, 7, 15, 1)
	if d := time.Since(t0); d > 60*time.Millisecond && os.Getenv("CI") == "" {
		t.Logf("desk took %v (target 60 ms)", d)
	}
}

func TestPureOfTheGame(t *testing.T) {
	fs, _ := filepath.Glob("*.go")
	for _, f := range fs {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		af, err := parser.ParseFile(token.NewFileSet(), f, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, im := range af.Imports {
			if strings.Contains(im.Path.Value, "ageforge") {
				t.Errorf("%s imports %s", f, im.Path.Value)
			}
		}
	}
}

// TestWritePictures writes the parts as PNG when LOGROLL_PNG names a folder,
// for a look with the eyes.
func TestWritePictures(t *testing.T) {
	dir := os.Getenv("LOGROLL_PNG")
	if dir == "" {
		t.Skip("LOGROLL_PNG not set")
	}
	save := func(name string, img *image.RGBA) {
		f, _ := os.Create(filepath.Join(dir, name))
		defer f.Close()
		png.Encode(f, img)
	}
	for _, s := range [][2]int{{279, 68}, {157, 40}} {
		l := Plan(s[0], s[1])
		save(fmt.Sprintf("desk-%dx%d.png", s[0], s[1]), Desk(l, 7, 15, 1))
	}
	l := Plan(279, 68)
	save("lever.png", Lever(l, 7, 15, 1))
	for o := 0; o < Offices; o++ {
		save(fmt.Sprintf("card-%d.png", o), Tile(o, o != 3, .6, o == 7, 1, 182, 45))
	}
}
