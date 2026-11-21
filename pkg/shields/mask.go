package shields

import "math"

// A mask is a small picture of tones, not colours: a shape is drawn in tones
// and a palette turns them into pixels. Tone 0 is empty.
const (
	tMain   = 1 // metal; the bevel picks its light and shade
	tLight  = 2
	tShade  = 3
	tDeep   = 4
	tGem    = 5
	tGemHi  = 6
	tRib    = 7 // ribbon
	tRibLo  = 8
	tRibHi  = 9
	tNoBevl = 10 // main metal that keeps its flat tone
)

type mask struct {
	w, h int
	t    []uint8
}

func newMask(w, h int) *mask { return &mask{w, h, make([]uint8, w*h)} }

func (m *mask) get(x, y int) uint8 {
	if x < 0 || y < 0 || x >= m.w || y >= m.h {
		return 0
	}
	return m.t[y*m.w+x]
}

func (m *mask) set(x, y int, t uint8) {
	if x >= 0 && y >= 0 && x < m.w && y < m.h {
		m.t[y*m.w+x] = t
	}
}

func (m *mask) rect(x, y, w, h int, t uint8) {
	for j := y; j < y+h; j++ {
		for i := x; i < x+w; i++ {
			m.set(i, j, t)
		}
	}
}

// disc fills the pixels whose centres lie within r of (cx, cy).
func (m *mask) disc(cx, cy, r float64, t uint8) {
	for y := 0; y < m.h; y++ {
		for x := 0; x < m.w; x++ {
			dx, dy := float64(x)+.5-cx, float64(y)+.5-cy
			if dx*dx+dy*dy <= r*r {
				m.set(x, y, t)
			}
		}
	}
}

// poly fills the polygon whose corners are the pairs in p, by pixel centres.
func (m *mask) poly(t uint8, p ...float64) {
	n := len(p) / 2
	for y := 0; y < m.h; y++ {
		fy := float64(y) + .5
		var xs []float64
		for i := 0; i < n; i++ {
			x0, y0, x1, y1 := p[2*i], p[2*i+1], p[2*((i+1)%n)], p[2*((i+1)%n)+1]
			if (y0 <= fy) != (y1 <= fy) {
				xs = append(xs, x0+(fy-y0)/(y1-y0)*(x1-x0))
			}
		}
		for i := 0; i < len(xs); i++ {
			for j := i + 1; j < len(xs); j++ {
				if xs[j] < xs[i] {
					xs[i], xs[j] = xs[j], xs[i]
				}
			}
		}
		for i := 0; i+1 < len(xs); i += 2 {
			for x := int(math.Ceil(xs[i] - .5)); float64(x)+.5 < xs[i+1]; x++ {
				m.set(x, y, t)
			}
		}
	}
}

// line is a one pixel line (Bresenham).
func (m *mask) line(x0, y0, x1, y1 int, t uint8) {
	dx, dy := abs(x1-x0), -abs(y1-y0)
	sx, sy := 1, 1
	if x0 > x1 {
		sx = -1
	}
	if y0 > y1 {
		sy = -1
	}
	for e := dx + dy; ; {
		m.set(x0, y0, t)
		if x0 == x1 && y0 == y1 {
			return
		}
		if e2 := 2 * e; e2 >= dy {
			e += dy
			x0 += sx
		} else if e2 <= dx {
			e += dx
			y0 += sy
		}
	}
}

// bevel turns main metal into light where the shape's upper left edge is, and
// shade where its lower right is.
func (m *mask) bevel() {
	out := make([]uint8, len(m.t))
	copy(out, m.t)
	for y := 0; y < m.h; y++ {
		for x := 0; x < m.w; x++ {
			if m.get(x, y) != tMain {
				continue
			}
			switch {
			case m.get(x, y-1) == 0 || m.get(x-1, y) == 0:
				out[y*m.w+x] = tLight
			case m.get(x, y+1) == 0 || m.get(x+1, y) == 0:
				out[y*m.w+x] = tShade
			}
		}
	}
	m.t = out
}

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}
