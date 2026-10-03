package mapstyle

import (
	"testing"
)

// mandalaSizes are layouts the styles ask for: the roguelike's full plane,
// the skyline at 80x24, 160x48 and 200x60, and the mini maps.
var mandalaSizes = [][2]float64{{20, 60}, {5, 10}, {12, 24}, {16, 31}, {6, 19}, {4, 19}, {2, 10}, {0, 3}}

// TestMandalaSymmetric: every layout is its own mirror image across the
// vertical axis and across the horizontal one, part for part and ring for
// ring, and a bead's or petal's number mirrors with it.
func TestMandalaSymmetric(t *testing.T) {
	for n := 0; n <= 7; n++ {
		for _, sz := range mandalaSizes {
			g := LayMandala(n, sz[0], sz[1])
			at := map[[2]int]MandalaCell{}
			for _, c := range g.Cells {
				at[[2]int{c.DX, c.DY}] = c
			}
			count := func(c MandalaCell) int {
				switch c.Part {
				case MdBead:
					return g.Beads[c.Ring]
				case MdPetal:
					return g.Petals
				}
				return 1
			}
			for _, c := range g.Cells {
				b := count(c)
				for _, m := range []struct {
					dx, dy int
					idx    int
					axis   string
				}{{-c.DX, c.DY, (b - c.Idx) % b, "left to right"}, {c.DX, -c.DY, ((b/2-c.Idx)%b + b) % b, "top to bottom"}} {
					o, ok := at[[2]int{m.dx, m.dy}]
					switch {
					case !ok:
						t.Errorf("n=%d %v: %v at (%d,%d) has no mirror %s", n, sz, c.Part, c.DX, c.DY, m.axis)
					case o.Part != c.Part || o.Ring != c.Ring || o.Tip != c.Tip:
						t.Errorf("n=%d %v: (%d,%d) is %v ring %d, its mirror %s %v ring %d", n, sz, c.DX, c.DY, c.Part, c.Ring, m.axis, o.Part, o.Ring)
					case (c.Part == MdBead || c.Part == MdPetal) && o.Idx != m.idx:
						t.Errorf("n=%d %v: %v %d at (%d,%d) mirrors %s to number %d, want %d", n, sz, c.Part, c.Idx, c.DX, c.DY, m.axis, o.Idx, m.idx)
					}
				}
			}
		}
	}
}

// TestMandalaRings: one ring per era asked for when there is room (all
// seven on the roguelike's plane and the skyline at 160x48), never more,
// the newest kept when there is not; the radii grow outward inside the
// space; and the core is the one centre cell.
func TestMandalaRings(t *testing.T) {
	for n := 0; n <= 7; n++ {
		for _, sz := range mandalaSizes {
			g := LayMandala(n, sz[0], sz[1])
			if g.Rings > n || len(g.Radii) != g.Rings || len(g.Beads) != g.Rings {
				t.Fatalf("n=%d %v: %d rings, %d radii, %d bead counts", n, sz, g.Rings, len(g.Radii), len(g.Beads))
			}
			for k := 1; k < g.Rings; k++ {
				if g.Radii[k] <= g.Radii[k-1] {
					t.Errorf("n=%d %v: ring %d at %.2f inside ring %d at %.2f", n, sz, k, g.Radii[k], k-1, g.Radii[k-1])
				}
			}
			if g.Extent > sz[0] || 2*g.Extent > sz[1] {
				t.Errorf("n=%d %v: reaches %.2f rows, past the space", n, sz, g.Extent)
			}
			drawn := map[int]bool{}
			cores := 0
			for _, c := range g.Cells {
				if c.Part == MdBead {
					drawn[c.Ring] = true
				}
				if c.Part == MdCore {
					cores++
					if c.DX != 0 || c.DY != 0 {
						t.Errorf("n=%d %v: the core is off centre at (%d,%d)", n, sz, c.DX, c.DY)
					}
				}
			}
			if cores != 1 {
				t.Errorf("n=%d %v: %d cores", n, sz, cores)
			}
			if len(drawn) != g.Rings {
				t.Errorf("n=%d %v: beads on %d rings of %d", n, sz, len(drawn), g.Rings)
			}
		}
	}
	for _, sz := range [][2]float64{{20, 60}, {12, 24}} {
		if g := LayMandala(7, sz[0], sz[1]); g.Rings != 7 {
			t.Errorf("%v: %d rings of 7", sz, g.Rings)
		}
	}
	if g := LayMandala(7, 20, 60); !g.Lines || g.Petals != 8 {
		t.Errorf("the roguelike's plane: lines %v, %d petals; want lines and the crown", g.Lines, g.Petals)
	}
}

// TestMandalaRingsApart: no cell belongs to two parts or two rings, and
// where the rings have dotted lines a clear cell parts every ring from the
// next (they never touch, let alone cross).
func TestMandalaRingsApart(t *testing.T) {
	for n := 0; n <= 7; n++ {
		for _, sz := range mandalaSizes {
			g := LayMandala(n, sz[0], sz[1])
			at := map[[2]int]MandalaCell{}
			for _, c := range g.Cells {
				k := [2]int{c.DX, c.DY}
				if o, ok := at[k]; ok {
					t.Errorf("n=%d %v: (%d,%d) is both %v ring %d and %v ring %d", n, sz, c.DX, c.DY, o.Part, o.Ring, c.Part, c.Ring)
				}
				at[k] = c
			}
			if !g.Lines {
				continue
			}
			for _, c := range g.Cells {
				if c.Part != MdLine && c.Part != MdBead {
					continue
				}
				for dy := -1; dy <= 1; dy++ {
					for dx := -1; dx <= 1; dx++ {
						o, ok := at[[2]int{c.DX + dx, c.DY + dy}]
						if ok && (o.Part == MdLine || o.Part == MdBead) && o.Ring != c.Ring {
							t.Errorf("n=%d %v: ring %d at (%d,%d) touches ring %d", n, sz, c.Ring, c.DX, c.DY, o.Ring)
						}
						if ok && o.Part == MdPetal || ok && o.Part == MdCore {
							t.Errorf("n=%d %v: ring %d at (%d,%d) touches the crown", n, sz, c.Ring, c.DX, c.DY)
						}
					}
				}
			}
		}
	}
}

// TestMandalaBreath: the wave of light is a clear crest that rolls outward
// (each ring lit a little after the one inside it, fully at its peak), comes
// round every MandalaPulse frames, leaves every ring dark for part of the
// wave, and changes smoothly: no ring jumps between two frames, not even
// when the wave starts over.
func TestMandalaBreath(t *testing.T) {
	for _, n := range []int{1, 4, 7} {
		for f := -MandalaPulse; f < 3*MandalaPulse; f++ {
			for k := -1; k < n; k++ {
				b := Breath(f, k, n)
				if d := Breath(f+1, k, n) - b; d > 0.35 || d < -0.35 {
					t.Fatalf("n=%d frame %d ring %d: the light jumps by %.3f", n, f, k, d)
				}
				if b != Breath(f+MandalaPulse, k, n) {
					t.Fatalf("n=%d frame %d ring %d: the wave does not come round", n, f, k)
				}
			}
		}
		prev := -1
		for k := -1; k < n; k++ {
			peak, at, dark := -1.0, 0, false
			for f := 0; f < MandalaPulse; f++ {
				b := Breath(f, k, n)
				if b > peak {
					peak, at = b, f
				}
				dark = dark || b == 0
			}
			if peak < 0.9 || !dark {
				t.Errorf("n=%d ring %d: the crest peaks at %.2f, dark for part of the wave %v", n, k, peak, dark)
			}
			if at <= prev {
				t.Errorf("n=%d ring %d peaks at frame %d, no later than the ring inside it (%d)", n, k, at, prev)
			}
			prev = at
		}
	}
}
