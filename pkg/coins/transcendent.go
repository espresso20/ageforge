package coins

import "math"

// ring is the distance to a ring of radius r, w wide.
func ring(x, y, r, w float64) float64 { return math.Abs(math.Hypot(x, y)-r) - w }

// transcendent is a thin halo round a bright mote, gaining rays and partial
// inner arcs with level, all in one white gold. Every copy is drawn alike.
func transcendent(l, lod int, _ uint32) *spec {
	lw := lineWidth(lod)
	sp := &spec{seed: 221, edge: 3}
	mote := level(l, 4, 4.6, 5.2, 5.8, 6.4, 7)
	nRay := int(level(l, 0, 0, 0, 4, 8, 8))
	rayL := level(l, 0, 0, 0, 12, 17, 22)
	gaps := level(l, 3, 2, 1, 0, 0, 0)
	sp.sd = func(x, y float64) float64 {
		a := angle(x, y)
		hal := ring(x, y, 52, lw(level(l, 1.3, 1.5, 1.7, 1.9, 2.1, 2.3)))
		if gaps > 0 {
			sec := int(math.Floor((a+math.Pi)/(2*math.Pi)*8)) % 8
			if sec%int(8/gaps) == 0 || (gaps == 3 && sec == 5) {
				hal += 60
			}
		}
		d := math.Min(hal, sdCircle(x, y, 0, 0, mote))
		for k := 0; k < nRay; k++ {
			lx, ly := rotate(x, y, float64(k)*2*math.Pi/float64(nRay))
			rl := rayL
			if k%2 == 1 {
				rl *= 0.6
			}
			d = math.Min(d, sdBox(lx, ly, rl/2+mote*0.5, 0, rl/2, lw(0.8)))
		}
		if l >= 5 { // partial arcs, never a full ring
			for _, a0 := range []float64{0.5, 3.64} {
				if math.Mod(a-a0+4*math.Pi, 2*math.Pi) < level(l, 0, 0, 0, 0, 1.3, 1.9) {
					d = math.Min(d, ring(x, y, 36, lw(1)))
				}
			}
		}
		return d
	}
	sp.ramp = [4]ramp{whiteGold, whiteGold, whiteGold, whiteGold}
	er := whiteGold
	sp.edgeR = &er
	sp.edgeT = level(l, 0.2, 0.3, 0.4, 0.5, 0.6, 0.75)
	br := level(l, 0.4, 0.5, 0.6, 0.75, 0.88, 1)
	sp.face = func(x, y float64) sample {
		r := math.Hypot(x, y)
		t := br
		if r < mote+1 {
			t = math.Min(1, br+0.1)
		} else if r > 30 {
			t = br * 0.9
		}
		return sample{h: ground, tone: t, seal: true}
	}
	return sp
}

// transcendentLegend is three thin rings and a mote joined by eight wedge
// rays, drawn in flat white gold.
func transcendentLegend() *spec {
	sp := transcendent(6, 0, 0)
	sp.edge = 1.6
	sp.rimW = 0
	sp.edgeT = 0.27
	wedge := []vec{{13.5, -4.4}, {29, -0.7}, {29, 0.7}, {13.5, 4.4}}
	sp.sd = func(x, y float64) float64 {
		d := math.Min(math.Min(ring(x, y, 55, 1.9), ring(x, y, 45.5, 1.4)), math.Min(ring(x, y, 36.5, 1.2), sdCircle(x, y, 0, 0, 9)))
		for k := 0; k < 8; k++ {
			lx, ly := rotate(x, y, float64(k)*math.Pi/4)
			d = math.Min(d, sdPoly(lx, ly, wedge))
		}
		return d
	}
	sp.face = func(x, y float64) sample {
		r := math.Hypot(x, y)
		if r < 4.8 {
			return flat(0, 4)
		}
		if r < 10 {
			return flat(0, 2)
		}
		return flat(0, 3)
	}
	return sp
}
