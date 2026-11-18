package cabinet

import (
	"bytes"
	"image"
	"testing"

	"github.com/espresso20/ageforge/pkg/coins"
)

var sizes = [][4]int{{140, 40, 7, 15}, {160, 44, 12, 24}, {250, 62, 7, 15}}

func plan(t *testing.T, s [4]int) Layout {
	t.Helper()
	l, ok := Plan(s[0], s[1], s[2], s[3])
	if !ok {
		t.Fatalf("Plan%v refused", s)
	}
	return l
}

func holdings() []Drawer {
	ds := make([]Drawer, Slots)
	for i := range ds {
		ds[i] = Drawer{Money: "classical_age", Reached: i%4 != 3, Best: 1 + i%coins.Levels, Total: 6, Held: i % 7}
	}
	for i := 22; i < Slots; i++ {
		ds[i] = Drawer{Reached: true}
	}
	return ds
}

func TestPlanAccepts(t *testing.T) {
	for _, s := range sizes {
		l := plan(t, s)
		pic := image.Rect(0, 0, s[0]*s[2], s[1]*s[3])
		for i, f := range l.Fronts {
			if !f.In(pic) {
				t.Errorf("%v: front %d outside the picture", s, i)
			}
			for j := i + 1; j < Slots; j++ {
				if f.Overlaps(l.Fronts[j]) {
					t.Errorf("%v: fronts %d and %d overlap", s, i, j)
				}
			}
			p := l.Plates[i]
			if p.Dy() != 1 || !l.pix(p).In(f) {
				t.Errorf("%v: plate %d is not one row inside its front", s, i)
			}
			if got := l.At(p.Min.X, p.Min.Y); got != i {
				t.Errorf("%v: At the plate of %d gave %d", s, i, got)
			}
		}
		if l.At(0, 0) != -1 {
			t.Errorf("%v: the cornice answers as a drawer", s)
		}
	}
	if _, ok := Plan(60, 20, 7, 15); ok {
		t.Error("a 60x20 panel is accepted")
	}
}

func TestFrontSizeAndRepeat(t *testing.T) {
	for _, s := range sizes {
		l := plan(t, s)
		img := Front(l, holdings(), coins.NewMint(64))
		if got := img.Rect.Size(); got != image.Pt(s[0]*s[2], s[1]*s[3]) {
			t.Fatalf("%v: picture is %v", s, got)
		}
	}
	l := plan(t, sizes[1])
	same := Drawer{Money: "classical_age", Reached: true, Best: 3, Total: 6, Held: 3}
	ds := make([]Drawer, Slots)
	for i := range ds {
		ds[i] = same
	}
	img := Front(l, ds, coins.NewMint(64))
	differ := 0
	for i := 0; i < 7; i++ {
		a, b := l.Fronts[i], l.Fronts[i+1]
		diff := 0
		for y := 0; y < a.Dy(); y++ {
			for x := 0; x < a.Dx(); x++ {
				if img.RGBAAt(a.Min.X+x, a.Min.Y+y) != img.RGBAAt(b.Min.X+x, b.Min.Y+y) {
					diff++
				}
			}
		}
		if diff*50 >= a.Dx()*a.Dy() {
			differ++
		}
	}
	if differ < 6 {
		t.Errorf("only %d of 7 neighbouring fronts differ by 2%% or more", differ)
	}
}

func TestPlatesAreFlat(t *testing.T) {
	for _, s := range sizes {
		l := plan(t, s)
		ds := holdings()
		img := Front(l, ds, coins.NewMint(64))
		flat := func(r image.Rectangle, dim bool) {
			r = l.pix(r)
			for y := r.Min.Y; y < r.Max.Y; y++ {
				for x := r.Min.X; x < r.Max.X; x++ {
					if img.RGBAAt(x, y) != PlateFace(dim) {
						t.Fatalf("%v: plate pixel %d,%d is not the face colour", s, x, y)
					}
				}
			}
		}
		for i, p := range l.Plates {
			flat(p, ds[i].Money != "" && !ds[i].Reached)
		}
		flat(l.Title, false)
		flat(l.Hint, false)
	}
}

func TestPatchAndDeterminism(t *testing.T) {
	for _, s := range sizes {
		l := plan(t, s)
		ds, mint := holdings(), coins.NewMint(64)
		for _, i := range []int{0, 4, 12, 20, 24} {
			cells, img := Patch(l, ds, i, mint)
			px := l.pix(cells)
			if img.Rect.Size() != px.Size() {
				t.Errorf("%v: patch %d is %v for cells %v", s, i, img.Rect.Size(), cells)
			}
			if !l.Fronts[i].In(px) {
				t.Errorf("%v: patch %d does not hold its front", s, i)
			}
			for j, f := range l.Fronts {
				if j != i && f.Overlaps(px) {
					t.Errorf("%v: patch %d touches front %d", s, i, j)
				}
			}
			_, again := Patch(l, ds, i, mint)
			if !bytes.Equal(img.Pix, again.Pix) {
				t.Errorf("%v: patch %d differs between runs", s, i)
			}
		}
		if !bytes.Equal(Front(l, ds, mint).Pix, Front(l, ds, mint).Pix) {
			t.Errorf("%v: the front differs between runs", s)
		}
	}
}
