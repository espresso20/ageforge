package cabinet

import (
	"image"
	"testing"

	"github.com/espresso20/ageforge/pkg/coins"
)

var traySizes = [][4]int{{160, 44, 12, 24}, {250, 62, 7, 15}, {120, 32, 7, 15}}

func testTray(slot int) Tray {
	t := Tray{Slot: slot, Money: "classical_age", Legend: true,
		Types: [Wells]string{"age", "wonder", "techs", "maximalist", "harbinger"}}
	for i := 0; i < 4; i++ {
		for k := 0; k <= i; k++ {
			t.Stacks[i] = append(t.Stacks[i], Copy{Level: 1 + (i+k)%coins.Levels, Seed: uint32(10*i + k + 1)})
		}
	}
	return t
}

func planTray(t *testing.T, s [4]int) TrayLayout {
	t.Helper()
	l, ok := PlanTray(s[0], s[1], s[2], s[3])
	if !ok {
		t.Fatalf("PlanTray%v refused", s)
	}
	return l
}

func TestPlanTrayShapes(t *testing.T) {
	for _, s := range traySizes {
		l := planTray(t, s)
		var rs []image.Rectangle
		for _, w := range l.Wells {
			rs = append(rs, w)
		}
		rs = append(rs, l.Legend)
		for _, c := range l.Labels {
			rs = append(rs, l.pix(c))
		}
		for _, c := range []image.Rectangle{l.Name, l.Count, l.Hint, l.Above, l.Below} {
			rs = append(rs, l.pix(c))
		}
		for i, a := range rs {
			for j := i + 1; j < len(rs); j++ {
				if a.Overlaps(rs[j]) {
					t.Errorf("%v: rectangles %d and %d overlap", s, i, j)
				}
			}
		}
		for i, w := range l.Wells {
			if got := l.At(w.Min.X/l.Cell.X+1, w.Min.Y/l.Cell.Y+1); got != i {
				t.Errorf("%v: At the middle of well %d gave %d", s, i, got)
			}
		}
		if got := l.At(l.Legend.Min.X/l.Cell.X+1, l.Legend.Min.Y/l.Cell.Y+1); got != Wells {
			t.Errorf("%v: At the legend gave %d", s, got)
		}
		if l.At(0, 0) != -1 {
			t.Errorf("%v: the frame answers as a well", s)
		}
	}
	if _, ok := PlanTray(40, 12, 7, 15); ok {
		t.Error("a 40x12 tray is accepted")
	}
}

func TestOpenAndLift(t *testing.T) {
	mint := coins.NewMint(400)
	s := traySizes[0]
	l := planTray(t, s)
	tr := testTray(4)
	img := Open(l, tr, mint)
	if img.Rect.Dx() != s[0]*s[2] || img.Rect.Dy() != s[1]*s[3] {
		t.Fatalf("Open is %v, want %dx%d", img.Rect, s[0]*s[2], s[1]*s[3])
	}
	diff := func(a *image.RGBA, ra image.Rectangle, b *image.RGBA, rb image.Rectangle) int {
		n := 0
		for y := 0; y < ra.Dy(); y++ {
			for x := 0; x < ra.Dx(); x++ {
				if a.RGBAAt(ra.Min.X+x, ra.Min.Y+y) != b.RGBAAt(rb.Min.X+x, rb.Min.Y+y) {
					n++
				}
			}
		}
		return n
	}
	if diff(img, l.Wells[2], img, l.Wells[3]) == 0 {
		t.Error("a held well looks like an empty one")
	}
	if diff(img, l.Wells[0], img, l.Wells[1]) == 0 {
		t.Error("two wells have the same lip")
	}
	cells, lift := Lift(l, tr, 2, mint)
	r := l.pix(cells)
	if lift.Rect.Dx() != r.Dx() || lift.Rect.Dy() != r.Dy() {
		t.Fatalf("Lift is %v, want %v", lift.Rect, r)
	}
	if diff(lift, lift.Rect, img, r) == 0 {
		t.Error("a lifted well matches the closed one")
	}
	if diff(lift, image.Rect(0, 0, 4, 4), img, image.Rectangle{Min: r.Min, Max: r.Min.Add(image.Pt(4, 4))}) != 0 {
		t.Error("the corner of a lift differs from Open")
	}
}

func TestSealsDifferBySeed(t *testing.T) {
	r := image.Rect(0, 0, 64, 64)
	pic := func(seed int) *image.RGBA {
		c := newCanvas(r, 1)
		c.fill(r, black)
		c.seal(32, 32, seed)
		return c.img
	}
	a, b, n := pic(3), pic(11), 0
	for y := 12; y < 52; y++ { // the seal's own box, with its shadow and drip
		for x := 12; x < 52; x++ {
			if a.RGBAAt(x, y) != b.RGBAAt(x, y) {
				n++
			}
		}
	}
	if n < 40*40/5 {
		t.Errorf("seals differ in %d of 1600 pixels", n)
	}
}
