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

func TestEveryPieceDraws(t *testing.T) {
	for _, age := range coins.Ages() {
		for level := 1; level <= coins.Legendary; level++ {
			for _, size := range []int{48, 128} {
				im := coins.Render(age, "age", level, 0, size)
				if im == nil || im.Rect.Dx() != size || im.Rect.Dy() != size {
					t.Fatalf("%s level %d size %d: wrong picture", age, level, size)
				}
				if p := painted(im); p < size*size/4 {
					t.Errorf("%s level %d size %d: only %d painted pixels", age, level, size, p)
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

func TestSeedsAndLevelsDiffer(t *testing.T) {
	for _, age := range coins.Ages() {
		var ims []*image.RGBA
		for seed := uint32(1); seed <= 6; seed++ {
			ims = append(ims, coins.Render(age, "age", 5, seed, 128))
		}
		for i := range ims {
			for j := i + 1; j < len(ims); j++ {
				if d := differ(ims[i], ims[j]); d < 0.05 {
					t.Errorf("%s seeds %d and %d differ in only %.1f%%", age, i+1, j+1, d*100)
				}
			}
		}
		for level := 1; level < coins.Legendary; level++ {
			a := coins.Render(age, "age", level, 0, 48)
			b := coins.Render(age, "age", level+1, 0, 48)
			if d := differ(a, b); d < 0.05 {
				t.Errorf("%s levels %d and %d differ in only %.1f%%", age, level, level+1, d*100)
			}
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
