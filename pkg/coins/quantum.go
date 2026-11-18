package coins

import "math"

var (
	indigo = newRamp(0x06051a, 0x0e0c30, 0x181650, 0x26246e, 0x38368e)
	violet = newRamp(0x0a0618, 0x1c0f44, 0x3a1f8a, 0x7a58d8, 0xb8a0f8)
)

// quantum is a rounded square with a lattice of nodes, some filled and some
// empty rings. Every copy is drawn alike.
func quantum(l, lod int, _ uint32) *spec {
	lw := lineWidth(lod)
	sp := &spec{seed: 31, edge: 6, rimW: 5, rimH: 1.4}
	sp.sd = func(x, y float64) float64 { return sdRoundBox(x, y, 0, 0, 54, 54, 3) }
	sp.ramp = [4]ramp{indigo, violet, lilac, indigo}
	er := violet
	sp.edgeR = &er
	sp.edgeT = level(l, 0.14, 0.2, 0.28, 0.36, 0.48, 0.6)
	m := int(level(l, 2, 3, 3, 4, 4, 5))
	fill := level(l, 0.25, 0.3, 0.4, 0.5, 0.55, 0.6)
	node := func(k int) float64 { return -30 + 60*float64(k)/float64(m-1) }
	set := func(i, j int) bool { return hash(i+40, j+40, 5) < fill }
	if m == 2 {
		set = func(i, j int) bool { return i == j }
	}
	nodeT := level(l, 0.55, 0.6, 0.6, 0.75, 0.88, 0.97)
	ringT := level(l, 0.36, 0.4, 0.42, 0.5, 0.58, 0.66)
	lineT := level(l, 0, 0, 0.3, 0.38, 0.46, 0.5)
	sp.face = func(x, y float64) sample {
		d := -sdRoundBox(x, y, 0, 0, 54, 54, 3)
		if l >= 2 && d > 7 && d < 7+lw(1.8) {
			return sample{h: ground, z: 1, tone: level(l, 0, 0.3, 0.4, 0.5, 0.62, 0.8), seal: true}
		}
		if l >= 4 { // corner brackets
			for _, sx := range []float64{-1, 1} {
				for _, sy := range []float64{-1, 1} {
					cx, cy := sx*40, sy*40
					if (math.Abs(y-cy) < lw(1.8) && (x-cx)*sx >= -lw(1.8) && (x-cx)*sx < 11) ||
						(math.Abs(x-cx) < lw(1.8) && (y-cy)*sy >= -lw(1.8) && (y-cy)*sy < 11) {
						return sample{h: ground, z: 2, tone: 0.85, seal: true}
					}
				}
			}
		}
		if l == 6 { // links between filled nodes
			for j := 0; j < m; j++ {
				for i := 0; i < m; i++ {
					if !set(i, j) {
						continue
					}
					if i+1 < m && set(i+1, j) && polyline(x, y, []vec{{node(i), node(j)}, {node(i + 1), node(j)}}, lw(1.2)) < 0 {
						return sample{h: ground, z: 2, tone: 0.8, seal: true}
					}
					if j+1 < m && set(i, j+1) && polyline(x, y, []vec{{node(i), node(j)}, {node(i), node(j + 1)}}, lw(1.2)) < 0 {
						return sample{h: ground, z: 2, tone: 0.8, seal: true}
					}
				}
			}
		}
		for j := 0; j < m; j++ {
			for i := 0; i < m; i++ {
				cx, cy := node(i), node(j)
				r := math.Hypot(x-cx, y-cy)
				if set(i, j) {
					if r < 4.8 {
						z := 2
						if l <= 2 {
							z = 1
						}
						return sample{h: ground, z: z, tone: nodeT, seal: true}
					}
					if l >= 5 && math.Abs(r-7.4) < lw(0.9) {
						return sample{h: ground, z: 1, tone: 0.5, seal: true}
					}
				} else if math.Abs(r-5.4) < lw(1.3) {
					return sample{h: ground, z: 1, tone: ringT, seal: true}
				}
			}
		}
		if lineT > 0 {
			on := func(v float64, k int) bool { return math.Abs(v-node(k)) < lw(0.8) }
			if math.Abs(x) <= 30.5 && math.Abs(y) <= 30.5 {
				for k := 0; k < m; k++ {
					if l == 3 && k != m/2 {
						continue
					}
					if on(x, k) || on(y, k) {
						return sample{h: ground, z: 1, tone: lineT, seal: true}
					}
				}
			}
		}
		return sample{h: ground}
	}
	return sp
}

// quantumLegend is the level 6 square with key teeth cut into its lower edge,
// a five by five lattice mirrored on both axes and the diagonal, corner blocks
// and a bar along the bottom.
func quantumLegend() *spec {
	sp := quantum(6, 0, 0)
	sp.edge = 3
	sp.edgeT = 0.27
	sp.field = 2
	sp.fieldTone = stepTone(2)
	sp.band = [2]float64{900, 901}
	depth := []float64{5.4, 8.6, 2.2, 5.4, 2.2, 8.6, 5.4} // square teeth, 9 px apart
	sp.sd = func(x, y float64) float64 {
		d := sdRoundBox(x, y, 0, 0, 54, 54, 3)
		for i, dp := range depth {
			d = sdSub(d, sdBox(x, y, -27+9*float64(i), 54, 2.8, dp))
		}
		return d
	}
	pat := func(i, j int) int { // 2 centre, 1 filled, 0 ring
		a, b := min(max(i, -i), max(j, -j)), max(max(i, -i), max(j, -j))
		switch [2]int{a, b} {
		case [2]int{0, 0}:
			return 2
		case [2]int{0, 1}, [2]int{1, 2}, [2]int{2, 2}:
			return 1
		}
		return 0
	}
	const pitch = 13.5
	sp.face = func(x, y float64) sample {
		ax, ay := math.Abs(x), math.Abs(y)
		if math.Abs(ax-44) < 6 && math.Abs(ay-44) < 6 {
			return sample{h: ground + 1.6, z: 2, tone: stepTone(3), seal: true}
		}
		if y > 42 && ax < 50.5 { // a bar below a dark groove
			if y < 43.4 {
				return flat(0, 0)
			}
			return sample{h: ground + 1.2, z: 2, tone: stepTone(3), seal: true}
		}
		mx := math.Max(ax, ay)
		if math.Abs(mx-44) < 1.7 {
			return flat(2, 2)
		}
		if math.Abs(mx-38.4) < 1.4 {
			return flat(1, 3)
		}
		if mx < 36 {
			i, j := int(math.Floor(x/pitch+0.5)), int(math.Floor(y/pitch+0.5))
			px, py := x-float64(i)*pitch, y-float64(j)*pitch
			mm := math.Max(math.Abs(px), math.Abs(py))
			switch pat(i, j) {
			case 2: // both states at once
				if ax < 7.6 && ay < 7.6 {
					if mm > 6.2 {
						return flat(2, 3)
					}
					if px+py < 0 {
						return flat(2, 4)
					}
					return flat(3, 0)
				}
			case 1:
				if mm < 4.1 {
					return flat(2, 3)
				}
			case 0:
				if mm < 4.1 && mm > 2.5 {
					return flat(1, 3)
				}
			}
		}
		return sample{h: ground}
	}
	return sp
}
