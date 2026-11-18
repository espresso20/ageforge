package coins

import "math"

var (
	silver = newRamp(0x20242e, 0x4a5262, 0x8a94a8, 0xc3cad8, 0xf4f7fc)
	gilt   = newRamp(0x4a3410, 0x8a6218, 0xd0a030, 0xf0d060, 0xfff2a0)
)

// classical is a silver piece: a temple front inside a ring of laurel.
func classical(l, _ int, seed uint32) *spec {
	sp := classicalBlank(l)
	handStruck(sp, l)
	if seed != 0 {
		vary(sp, kindClassical, l, seed)
	}
	return sp
}

func classicalBlank(l int) *spec {
	w := wear(l)
	sp := &spec{seed: 111, edge: 4, rimW: 5, rimH: 1.6 * (1 - 0.4*w), field: 1}
	sp.sd = func(x, y float64) float64 {
		return sdCircle(x, y, 0, 0, 54+1.8*math.Sin(3*angle(x, y)+1))
	}
	sp.ramp = [4]ramp{silver, gilt, silver, silver}
	rel := 1 - 0.5*w
	z := 0
	if l == 6 {
		z = 1
	}
	ped := []vec{{-27, -4}, {27, -4}, {0, -22}}
	sp.face = func(x, y float64) sample {
		h := ground
		d := sdPoly(x, y, ped)
		if d < 0 {
			h += 2.2 * rel * bevel(d, 2)
			if sdPoly(x, y, []vec{{-17, -6.5}, {17, -6.5}, {0, -16.5}}) < 0 {
				h -= 1.2 * rel // recessed tympanum
			}
		}
		for _, b := range [][4]float64{{0, -1.2, 29, 2.2}, {0, 20, 28, 1.7}, {0, 23.4, 31, 1.7}, {0, 26.8, 34, 1.7}} {
			if e := sdBox(x, y, b[0], b[1], b[2], b[3]); e < 0 {
				h += 2.3 * rel * bevel(e, 1.4)
			}
		}
		for _, cx := range []float64{-21, -7, 7, 21} {
			if e := sdBox(x, y, cx, 9.2, 3.1, 8.8); e < 0 {
				h += 2.3 * rel * bevel(e, 2)
			}
		}
		u, v := around(x, y, 18, 0) // laurel: leaves in a ring
		if e := sdTrueEllipse(u-43, v, 5.6, 2.3, 0.6); e < 0 {
			h += 1.7 * rel * bevel(e, 2)
		}
		zz := 0
		if h > ground+0.5 {
			zz = z
		}
		return sample{h: h, z: zz}
	}
	return sp
}

// classicalLegend is the level 6 piece with the finest work of its age: a
// dotted border, a full wreath of paired leaves and berries, a fluted temple.
func classicalLegend() *spec {
	sp := classicalBlank(6)
	sp.edge += 1.5
	sp.rimW = 2.6
	sp.rimH = 1.6
	sp.face = func(x, y float64) sample {
		r := math.Hypot(x, y)
		h := ground
		if r > 49.4 && r < 53 { // fine dotted border
			u, v := around(x, y, 84, 0)
			if math.Hypot(u-51.2, v) < 1.0 {
				return sample{h: ground + 2.6}
			}
		}
		const nL = 24 // leaves alternate inside and outside, berries between
		idx := sector(x, y, nL, 0)
		u, v := around(x, y, nL, 0)
		var el float64
		if idx%2 == 0 {
			el = sdTrueEllipse(u-39.6, v, 6.4, 2.9, 0.55)
		} else {
			el = sdTrueEllipse(u-45.2, v, 6.4, 2.9, -0.55)
		}
		if el < 0 {
			hh := 2.1 * bevel(el, 2)
			if math.Abs(v) < 0.35 { // midrib
				hh += 0.5
			}
			return sample{h: ground + hh, z: 1}
		}
		bu, bv := around(x, y, 12, math.Pi/12)
		if d := math.Hypot(bu-42.4, bv) - 2.1; d < 0 {
			return sample{h: ground + 2.6*bevel(d, 2.1), z: 1}
		}
		const k = 0.84 // the temple, a little smaller to sit inside the wreath
		tx, ty := x/k, (y+0.5)/k
		if d := sdPoly(tx, ty, []vec{{-32, -4}, {32, -4}, {0, -25}}); d < 0 {
			h += 3.2 * bevel(d, 2)
			if sdPoly(tx, ty, []vec{{-21, -6.5}, {21, -6.5}, {0, -18.5}}) < 0 {
				h -= 1.8
				for _, fx := range []float64{-12, -6, 0, 6, 12} {
					h += 2.2 * dome(tx, ty, fx, -8.6, 1.5, 1)
				}
			}
		}
		for _, b := range [][4]float64{{0, -1.2, 32, 2.2}, {0, 21, 30, 1.7}, {0, 24.4, 33, 1.7}, {0, 27.8, 36, 1.7}} {
			if e := sdBox(tx, ty, b[0], b[1], b[2], b[3]); e < 0 {
				h += 3.0 * bevel(e, 1.4)
			}
		}
		for _, cx := range []float64{-25, -15, -5, 5, 15, 25} {
			if e := sdBox(tx, ty, cx, 10, 3.1, 9.6); e < 0 {
				h += 3.2 * bevel(e, 2)
				if math.Abs(math.Abs(tx-cx)-1.0) < 0.35 {
					h -= 0.7
				}
			}
			if e := sdBox(tx, ty, cx, 1.6, 4.2, 1.0); e < 0 {
				h += 3.4 * bevel(e, 1)
			}
		}
		zz := 0
		if h > ground+0.5 {
			zz = 1
		}
		return sample{h: h, z: zz}
	}
	return sp
}
