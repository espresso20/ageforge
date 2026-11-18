package cabinet

import (
	"image"
	"sort"
	"testing"

	"github.com/espresso20/ageforge/pkg/coins"
)

func stackOf(n int) []Copy {
	var st []Copy
	for k := 0; k < n; k++ {
		st = append(st, Copy{Level: coins.Levels - k*coins.Levels/n, Seed: uint32(40 + 3*k)})
	}
	return st
}

// fixed lists every plate and well of the tray, in pixels.
func fixed(l TrayLayout) []image.Rectangle {
	rs := append(l.Wells[:], l.Legend)
	for _, c := range append(l.Labels[:], l.Name, l.Count, l.Hint, l.Above, l.Below, l.Ticket) {
		rs = append(rs, l.pix(c))
	}
	return rs
}

func TestFanClearsEverything(t *testing.T) {
	mint := coins.NewMint(400)
	for _, s := range traySizes[:2] {
		l, tr := planTray(t, s), testTray(4)
		for w := 0; w < Wells; w++ {
			tr.Stacks[w] = stackOf(9)
			cells, img := Fan(l, tr, w, mint)
			r := l.pix(cells)
			if cells.Empty() || img.Rect.Dx() != r.Dx() || img.Rect.Dy() != r.Dy() || !cells.In(l.fan) {
				t.Fatalf("%v well %d: fan %v of %v", s, w, cells, l.fan)
			}
			for i, o := range fixed(l) {
				if r.Overlaps(o) {
					t.Errorf("%v well %d: the fan overlaps rectangle %d", s, w, i)
				}
			}
		}
	}
}

func TestFanRowIsOneCurve(t *testing.T) {
	for _, s := range traySizes[:2] {
		l := planTray(t, s)
		for w := 0; w < Wells; w++ {
			for _, n := range []int{4, FanMax} {
				if q := l.fanAt(w, n); q[0].Y == q[n/2].Y && n > 2 {
					t.Errorf("%v well %d, %d copies: the row is flat", s, w, n)
				}
				p, sign := l.fanAt(w, n), [2]int{}
				sort.Slice(p, func(a, b int) bool { return p[a].X < p[b].X })
				for k := 2; k < n; k++ {
					for a, d := range []int{p[k].X - 2*p[k-1].X + p[k-2].X, p[k].Y - 2*p[k-1].Y + p[k-2].Y} {
						if d != 0 && sign[a] != 0 && (d > 0) != (sign[a] > 0) {
							t.Errorf("%v well %d, %d copies: the row kinks at %d", s, w, n, k)
						}
						if d != 0 {
							sign[a] = d
						}
					}
				}
			}
		}
	}
}

func TestFanCopiesDiffer(t *testing.T) {
	mint := coins.NewMint(400)
	l, tr := planTray(t, traySizes[1]), testTray(4)
	tr.Stacks[2] = []Copy{{4, 1}, {4, 2}}
	_, a := Fan(l, tr, 2, mint)
	tr.Stacks[2] = []Copy{{4, 1}, {4, 1}}
	_, b := Fan(l, tr, 2, mint)
	if a.Rect != b.Rect || string(a.Pix) == string(b.Pix) {
		t.Error("two copies of one stack with different seeds look the same")
	}
	if c, _ := Fan(l, tr, 4, mint); !c.Empty() {
		t.Error("an empty stack fans")
	}
}

func TestLiftIsPlain(t *testing.T) {
	mint := coins.NewMint(400)
	l, tr := planTray(t, traySizes[1]), testTray(4)
	open := Open(l, tr, mint)
	for _, w := range []int{1, Wells} {
		cells, lift := Lift(l, tr, w, mint)
		r, n := l.pix(cells), 0
		for y := 0; y < r.Dy(); y++ {
			for x := 0; x < r.Dx(); x++ {
				if lift.RGBAAt(x, y) != open.RGBAAt(r.Min.X+x, r.Min.Y+y) {
					n++
				}
			}
		}
		if n*10 < r.Dx()*r.Dy() {
			t.Errorf("well %d: a lift changes %d of %d pixels", w, n, r.Dx()*r.Dy())
		}
	}
}
