package coins

import "math"

var (
	ice   = newRamp(0x0a1a30, 0x1a3a68, 0x3a70b0, 0x80b8e8, 0xe0f4ff)
	lilac = newRamp(0x2a2450, 0x6a60b0, 0xb0a8ec, 0xe0dcff, 0xffffff)
)

// interstellar is a tall six-sided crystal prism with hard facets and a star
// chart inside. Every copy is drawn alike.
func interstellar(l, lod int, _ uint32) *spec {
	lw := lineWidth(lod)
	pts := []vec{{4, -56}, {30, -38}, {30, 30}, {6, 54}, {-26, 34}, {-30, -32}}
	sp := &spec{seed: 201, edge: 7, rimW: 3, rimH: 0.9}
	sp.sd = func(x, y float64) float64 { return sdPoly(x, y, pts) }
	sp.ramp = [4]ramp{ice, lilac, cyan, ice}
	er := blueTi
	sp.edgeR = &er
	sp.edgeT = level(l, 0.15, 0.2, 0.3, 0.4, 0.5, 0.62)
	stars := []vec{{-14, -24}, {2, -30}, {16, -20}, {20, -4}, {-4, -6}, {-18, 4}, {8, 10}, {-8, 22}, {16, 26}, {2, 38}}
	links := [][2]int{{0, 1}, {1, 2}, {2, 3}, {1, 4}, {4, 5}, {4, 6}, {6, 7}, {6, 8}, {7, 9}, {3, 6}}
	nS := int(level(l, 2, 3, 5, 7, 8, 10))
	nL := int(level(l, 0, 0, 2, 5, 8, 10))
	tone := level(l, 0.45, 0.55, 0.65, 0.78, 0.9, 1)
	facets := level(l, 1, 2, 3, 4, 5, 6)
	sp.face = func(x, y float64) sample {
		for k := 0; k < nS; k++ {
			s := stars[k]
			if math.Hypot(x-s.x, y-s.y) < lw(2.2) || (l >= 5 && ((math.Abs(x-s.x) < 0.9 && math.Abs(y-s.y) < 5) || (math.Abs(y-s.y) < 0.9 && math.Abs(x-s.x) < 5))) {
				return sample{h: ground, z: 1, tone: tone, seal: true}
			}
		}
		for k := 0; k < nL; k++ {
			a, b := stars[links[k][0]], stars[links[k][1]]
			if sdSeg(x, y, a.x, a.y, b.x, b.y, lw(0.8)) < 0 {
				return sample{h: ground, z: 2, tone: tone * 0.8, seal: true}
			}
		}
		// hard planes: three long faces down the prism, a cap at each end
		h := ground
		f := facets / 6
		switch {
		case y < -34+math.Abs(x-2)*0.9:
			h += 0.16*(-y-34)*f*0.8 + 0.05*x*f
		case y > 34-math.Abs(x+10)*0.9:
			h -= 0.14 * (y - 20) * f * 0.5
		case x < -10:
			h += 0.1 * (x + 30) * f
		case x > 12:
			h -= 0.1 * (x - 12) * f * 1.1
		}
		if facets >= 3 && (math.Abs(x+10) < lw(0.7) || math.Abs(x-12) < lw(0.7)) {
			return sample{h: ground + 1.2, z: 1, tone: 0.9, seal: true}
		}
		return sample{h: h}
	}
	return sp
}

// interstellarLegend is the level 6 prism a little taller, with a chart of
// ball-shaped stars joined by lines, thin metal caps and bright arrises.
func interstellarLegend() *spec {
	sp := interstellar(6, 0, 0)
	pts := []vec{{4, -59}, {31, -41}, {31, 33}, {6, 58}, {-27, 37}, {-31, -35}}
	sp.sd = func(x, y float64) float64 { return sdPoly(x, y, pts) }
	sp.ramp = [4]ramp{ice, nickel, cyan, lilac}
	sp.edgeT = 0.52
	stars := []vec{{-12, -27}, {6, -32}, {20, -20}, {14, -6}, {-6, -10}, {-19, 2}, {4, 8}, {-10, 20}, {18, 22}, {2, 30}}
	links := [][2]int{{0, 1}, {1, 2}, {2, 3}, {1, 4}, {4, 5}, {4, 6}, {6, 7}, {6, 8}, {7, 9}, {3, 6}, {8, 9}}
	edge := func(x, y float64, a, b vec) float64 { return sdSeg(x, y, a.x, a.y, b.x, b.y, 0) }
	sp.face = func(x, y float64) sample {
		for k, s := range stars {
			if k == 6 {
				if math.Hypot(x-s.x, y-s.y) < 5 {
					return flat(3, 4)
				}
				continue
			}
			if math.Hypot(x-s.x, y-s.y) < 3.3 {
				return flat(2, 3)
			}
		}
		for _, l := range links {
			a, b := stars[l[0]], stars[l[1]]
			if sdSeg(x, y, a.x, a.y, b.x, b.y, 0.95) < 0 {
				return flat(2, 2)
			}
		}
		// the thin metal caps at both ends
		top := math.Min(edge(x, y, pts[5], pts[0]), edge(x, y, pts[0], pts[1]))
		bot := math.Min(edge(x, y, pts[2], pts[3]), edge(x, y, pts[3], pts[4]))
		cap := math.Min(top, bot)
		if cap < 6.4 {
			switch {
			case cap > 5:
				return flat(0, 0)
			case cap < 2.6:
				return flat(1, 4)
			}
			return flat(1, 3)
		}
		// three hard facets with a bright arris between
		if math.Abs(x+10) < 1.3 || math.Abs(x-12) < 1.3 {
			return flat(0, 4)
		}
		switch {
		case x < -10:
			return flat(0, 3)
		case x > 12:
			return flat(0, 1)
		}
		return flat(0, 2)
	}
	return sp
}
