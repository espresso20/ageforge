package coins

import "math"

// medieval is a thin hammered silver penny with a long cross, clipped at the
// edge at the lower levels.
func medieval(l, _ int, seed uint32) *spec {
	sp := medievalBlank(l)
	handStruck(sp, l)
	if seed != 0 {
		vary(sp, kindMedieval, l, seed)
	}
	return sp
}

func medievalBlank(l int) *spec {
	w := wear(l)
	sp := &spec{seed: 121, edge: 2, rimW: 4, rimH: 1.1 * (1 - 0.4*w), field: 1}
	nClip := int(level(l, 4, 3, 3, 2, 1, 1))
	angs := []float64{0.6, 2.5, 4.3, 5.6}
	sp.sd = func(x, y float64) float64 {
		d := sdCircle(x, y, 0, 0, 55+1.2*math.Sin(4*angle(x, y)+1))
		for k := 0; k < nClip; k++ {
			depth := 48.0 + 2*float64(k)
			if l == 6 {
				depth = 52
			}
			d = math.Max(d, x*math.Cos(angs[k])+y*math.Sin(angs[k])-depth)
		}
		return d
	}
	sp.ramp = [4]ramp{silver, silver, silver, silver}
	rel := 1 - 0.5*w
	sp.face = func(x, y float64) sample {
		h := ground
		bx, by := math.Abs(x), math.Abs(y)
		if (bx < 2.6 && by < 46) || (by < 2.6 && bx < 46) {
			h += 1.9 * rel
		}
		for _, q := range [][2]float64{{1, 1}, {1, -1}, {-1, 1}, {-1, -1}} { // pellets in the quarters
			for k := 0; k < 3; k++ {
				px, py := q[0]*(14+float64(k)*7), q[1]*(14+float64(k)*7)
				if k == 1 {
					px, py = q[0]*14, q[1]*21
				}
				if k == 2 {
					px, py = q[0]*21, q[1]*14
				}
				if math.Hypot(x-px, y-py) < 2.6 {
					h += 1.6 * rel
				}
			}
		}
		return sample{h: h}
	}
	return sp
}

// medievalLegend is the level 6 penny with the finest work of its age: a long
// forked cross, a lobed frame, pellets in the quarters and rings of tick marks.
func medievalLegend() *spec {
	sp := medievalBlank(6)
	sp.edge += 1.5
	sp.sd = func(x, y float64) float64 { return sdCircle(x, y, 0, 0, 55.5+0.4*math.Sin(5*angle(x, y))) }
	sp.rimW = 3
	sp.rimH = 1.5
	sp.face = func(x, y float64) sample {
		r := math.Hypot(x, y)
		a := angle(x, y)
		ax, ay := math.Abs(x), math.Abs(y)
		h := ground
		if (ax < 2.3 && ay < 46) || (ay < 2.3 && ax < 46) { // long cross, forked at the four ends
			return sample{h: ground + 2.4}
		}
		fu, fv := around(x, y, 4, 0)
		if sdSeg(fu, fv, 45, 0, 51.5, -4.6, 1.15) < 0 || sdSeg(fu, fv, 45, 0, 51.5, 4.6, 1.15) < 0 {
			return sample{h: ground + 2.2}
		}
		if math.Abs(r-(17+2.4*math.Abs(math.Sin(6*a)))) < 1.1 { // many-lobed frame round the centre
			return sample{h: ground + 1.9}
		}
		if r < 3.4 {
			return sample{h: ground + 2.4*bevel(r-3.4, 3.4)}
		}
		for _, q := range [][2]float64{{1, 1}, {1, -1}, {-1, 1}, {-1, -1}} { // pellets in the quarters
			for _, o := range [][2]float64{{12, 12}, {21, 12}, {12, 21}} {
				if d := math.Hypot(x-q[0]*o[0], y-q[1]*o[1]) - 2.6; d < 0 {
					h = math.Max(h, ground+2.1*bevel(d, 2.6))
				}
			}
			if d := math.Hypot(x-q[0]*29, y-q[1]*29) - 2.2; d < 0 {
				h = math.Max(h, ground+2.0*bevel(d, 2.2))
			}
		}
		if h > ground {
			return sample{h: h}
		}
		if ax > 4 && ay > 4 { // two rings of tick marks where lettering would be
			if r > 39.2 && r < 43.6 && math.Sin(a*44) > 0.45 {
				return sample{h: ground + 1.1}
			}
			if r > 44.8 && r < 49.2 && math.Sin(a*56) > 0.45 {
				return sample{h: ground + 1.1}
			}
			if math.Abs(r-38.2) < 0.55 || math.Abs(r-44.2) < 0.5 || math.Abs(r-50) < 0.5 {
				return sample{h: ground + 0.9}
			}
		}
		return sample{h: h}
	}
	return sp
}
