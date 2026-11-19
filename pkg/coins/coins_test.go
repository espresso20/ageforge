package coins_test

import (
	"image"
	"sync"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/pkg/coins"
)

// differ is the share of pixels at which two same-size pictures differ.
func differ(a, b *image.RGBA) float64 {
	n := 0
	for i := 0; i < len(a.Pix); i += 4 {
		if a.Pix[i] != b.Pix[i] || a.Pix[i+1] != b.Pix[i+1] || a.Pix[i+2] != b.Pix[i+2] || a.Pix[i+3] != b.Pix[i+3] {
			n++
		}
	}
	return float64(n) / float64(len(a.Pix)/4)
}

func painted(im *image.RGBA) int {
	n := 0
	for i := 3; i < len(im.Pix); i += 4 {
		if im.Pix[i] != 0 {
			n++
		}
	}
	return n
}

// span is the shorter side of the box around a picture's painted pixels.
func span(im *image.RGBA) int {
	b := im.Bounds()
	x0, y0, x1, y1 := b.Max.X, b.Max.Y, b.Min.X-1, b.Min.Y-1
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if im.RGBAAt(x, y).A != 0 {
				x0, y0, x1, y1 = min(x0, x), min(y0, y), max(x1, x), max(y1, y)
			}
		}
	}
	return max(0, min(x1-x0, y1-y0)+1)
}

func TestEveryPieceDraws(t *testing.T) {
	for _, age := range coins.Ages() {
		for level := 1; level <= coins.Legendary; level++ {
			for _, size := range []int{48, 128} {
				im := coins.Render(age, "age", level, 0, size)
				if im == nil || im.Rect.Dx() != size || im.Rect.Dy() != size {
					t.Fatalf("%s level %d size %d: wrong picture", age, level, size)
				}
				// Some money is an open form that paints little of the square, so
				// the check is on the room a piece takes up, with a low floor on
				// its pixels.
				if w, p := span(im), painted(im); w < size/3 || p < size*size/100 {
					t.Errorf("%s level %d size %d: spans %d px with %d painted pixels", age, level, size, w, p)
				}
			}
		}
	}
}

func TestSameArgumentsSamePicture(t *testing.T) {
	for _, age := range coins.Ages() {
		for _, level := range []int{3, coins.Legendary} {
			a := coins.Render(age, "age", level, 7, 128)
			b := coins.Render(age, "age", level, 7, 128)
			if differ(a, b) != 0 {
				t.Errorf("%s level %d: two renders differ", age, level)
			}
		}
	}
}

// varied lists the ages whose copies differ from one another; the others draw
// every copy alike.
var varied = map[string]bool{"classical_age": true, "medieval_age": true, "victorian_age": true, "iron_age": true}

func TestSeedsAndLevelsDiffer(t *testing.T) {
	for _, age := range coins.Ages() {
		var ims []*image.RGBA
		for seed := uint32(1); seed <= 6; seed++ {
			ims = append(ims, coins.Render(age, "age", 5, seed, 128))
		}
		for i := range ims {
			for j := i + 1; j < len(ims); j++ {
				d := differ(ims[i], ims[j])
				if varied[age] && d < 0.05 {
					t.Errorf("%s seeds %d and %d differ in only %.1f%%", age, i+1, j+1, d*100)
				}
				if !varied[age] && d != 0 {
					t.Errorf("%s seeds %d and %d differ, but the age draws every copy alike", age, i+1, j+1)
				}
			}
		}
		// Neighbouring levels are compared at full detail. The small size drops
		// fine marks, so there only the two ends of the range are compared.
		for level := 1; level < coins.Legendary; level++ {
			a := coins.Render(age, "age", level, 0, 128)
			b := coins.Render(age, "age", level+1, 0, 128)
			if d := differ(a, b); d < 0.01 {
				t.Errorf("%s levels %d and %d differ in only %.1f%%", age, level, level+1, d*100)
			}
		}
		lo, hi := coins.Render(age, "age", 1, 0, 48), coins.Render(age, "age", coins.Levels, 0, 48)
		if d := differ(lo, hi); d < 0.10 {
			t.Errorf("%s levels 1 and %d differ in only %.1f%% at the small size", age, coins.Levels, d*100)
		}
	}
}

func TestOutOfRangeIsNil(t *testing.T) {
	if coins.Has("no_such_age") || coins.Render("no_such_age", "age", 1, 0, 48) != nil {
		t.Error("unknown age drew")
	}
	age := coins.Ages()[0]
	for _, level := range []int{0, coins.Legendary + 1} {
		if coins.Render(age, "age", level, 0, 48) != nil {
			t.Errorf("level %d drew", level)
		}
	}
}

func TestMint(t *testing.T) {
	m := coins.NewMint(2)
	age := coins.Ages()[0]
	a := m.Piece(age, "age", 1, 0, 48)
	if a == nil || m.Piece(age, "age", 1, 0, 48) != a {
		t.Fatal("second call did not return the same image")
	}
	m.Piece(age, "age", 2, 0, 48)
	m.Piece(age, "age", 3, 0, 48) // pushes level 1 out
	if m.Piece(age, "age", 1, 0, 48) == a {
		t.Error("the limit did not drop the oldest piece")
	}
	if m.Piece("no_such_age", "age", 1, 0, 48) != nil {
		t.Error("unknown age drew")
	}
	var wg sync.WaitGroup
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 6; i++ {
				m.Piece(age, "age", 1+i%4, uint32(g%3), 48)
			}
		}()
	}
	wg.Wait()
}

func TestMatchesConfig(t *testing.T) {
	real := config.AgeByKey()
	for _, age := range coins.Ages() {
		if _, ok := real[age]; !ok {
			t.Errorf("%q is not an age key", age)
		}
	}
	if coins.Levels != config.CoinLevels || coins.Legendary != config.CoinLegendary {
		t.Error("levels differ from config")
	}
}
