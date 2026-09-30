package mapmodel

import "testing"

// TestWaterReadsNaturally: the sea and the lakes have clean shores and the
// rivers run whole. Adam found the roguelike's water a blocky blob with
// notches. Over many seeds: deep water never touches land (every shore has
// its ring of shallows), no water pokes a one-tile notch into the land and
// no land a one-tile spur into the water, and every river runs unbroken to
// the sea or a lake.
func TestWaterReadsNaturally(t *testing.T) {
	d4 := [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
	still := func(t Terrain) bool { return t == TDeep || t == TShallow }
	for s := int64(1); s <= 24; s++ {
		seed := s * 7919
		w := NewWorld(seed, 11)
		deepShore, notches, spurs := 0, 0, 0
		for y := 1; y < w.H-1; y++ {
			for x := 1; x < w.W-1; x++ {
				tt := w.At(x, y)
				land, wet := 0, 0
				for _, d := range d4 {
					switch n := w.At(x+d[0], y+d[1]); {
					case n.Land():
						land++
					case still(n):
						wet++
					}
				}
				switch {
				case tt == TDeep && land > 0:
					deepShore++
				case still(tt) && land >= 3:
					notches++
				case tt.Land() && wet >= 3:
					spurs++
				}
			}
		}
		if deepShore > 0 || notches > 0 || spurs > 0 {
			t.Errorf("seed %d: %d deep tiles on a shore, %d one-tile notches, %d one-tile spurs", seed, deepShore, notches, spurs)
		}
		// Every river is one band that reaches still water.
		seen := make([]bool, w.W*w.H)
		for i, tt := range w.T {
			if tt != TRiver || seen[i] {
				continue
			}
			stack, size, mouth := []int{i}, 0, false
			seen[i] = true
			for len(stack) > 0 {
				j := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				size++
				x, y := j%w.W, j/w.W
				for dy := -1; dy <= 1; dy++ {
					for dx := -1; dx <= 1; dx++ {
						nx, ny := x+dx, y+dy
						if !w.In(nx, ny) {
							continue
						}
						k := ny*w.W + nx
						switch {
						case still(w.T[k]):
							mouth = true
						case w.T[k] == TRiver && !seen[k]:
							seen[k] = true
							stack = append(stack, k)
						}
					}
				}
			}
			if !mouth {
				t.Errorf("seed %d: a river of %d tiles near %d,%d never reaches the sea or a lake", seed, size, i%w.W, i/w.W)
			}
			if size < 8 {
				t.Errorf("seed %d: a scrap of river, %d tiles, near %d,%d", seed, size, i%w.W, i/w.W)
			}
		}
	}
}

// TestRiverBandIsUnbroken: a river drawn along any course (painted, then
// stitched, as rivers does) is one band joined side to side, whatever its
// angle; a thin brush used to skip rows or touch only at corners.
func TestRiverBandIsUnbroken(t *testing.T) {
	for _, c := range [][2]fpt{{{10, 10.5}, {60, 10.5}}, {{10, 10}, {60, 30}}, {{30, 5}, {33, 40}}, {{60, 10.5}, {10, 30.5}}} {
		w := &World{W: 80, H: 50, T: make([]Terrain, 80*50)}
		for i := range w.T {
			w.T[i] = TGrass
		}
		var course []fpt
		for k := 0; k <= 40; k++ {
			f := float64(k) / 40
			course = append(course, fpt{c[0].X + (c[1].X-c[0].X)*f, c[0].Y + (c[1].Y-c[0].Y)*f})
		}
		w.paintRiver(course)
		w.stitchRivers()
		// Walk the course: every point's tile must be river, and the band
		// must be one 4-connected piece.
		for _, p := range course {
			if w.At(int(p.X+0.5), int(p.Y+0.5)) != TRiver && w.At(int(p.X), int(p.Y)) != TRiver {
				t.Errorf("course %v: no river under %v", c, p)
			}
		}
		start := -1
		n := 0
		for i, tt := range w.T {
			if tt == TRiver {
				n++
				if start < 0 {
					start = i
				}
			}
		}
		seen := map[int]bool{start: true}
		stack := []int{start}
		for len(stack) > 0 {
			j := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			x, y := j%w.W, j/w.W
			for _, d := range [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
				if nx, ny := x+d[0], y+d[1]; w.In(nx, ny) && w.T[ny*w.W+nx] == TRiver && !seen[ny*w.W+nx] {
					seen[ny*w.W+nx] = true
					stack = append(stack, ny*w.W+nx)
				}
			}
		}
		if len(seen) != n {
			t.Errorf("course %v: the band is in pieces (%d of %d tiles connected)", c, len(seen), n)
		}
	}
}
