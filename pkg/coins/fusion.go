package coins

import "math"

var (
	gun       = newRamp(0x1c2028, 0x2e343f, 0x434b5a, 0x5d6879, 0x8592a8)
	orange    = newRamp(0x3a0e04, 0x8a2a08, 0xd85a14, 0xff9a3a, 0xffe0a0)
	whiteGold = newRamp(0x3a3220, 0x7a6a40, 0xc8b678, 0xeee0b0, 0xfffbe8)
)

// fusion is a sealed ring with plasma in its channel, held by clamps that
// multiply with the level. Every copy is drawn alike.
func fusion(l, lod int, _ uint32) *spec {
	lw := lineWidth(lod)
	sp := &spec{seed: 191, edge: 7, rimW: 5, rimH: 1.6}
	sp.sd = func(x, y float64) float64 { return sdSub(sdCircle(x, y, 0, 0, 57), sdCircle(x, y, 0, 0, 22)) }
	sp.ramp = [4]ramp{gun, orange, whiteGold, graphite}
	er := orange
	sp.edgeR = &er
	sp.edgeT = level(l, 0.1, 0.15, 0.25, 0.35, 0.5, 0.62)
	nClamp := int(level(l, 0, 4, 4, 6, 8, 12))
	gaps := int(level(l, 6, 4, 2, 1, 0, 0))
	pt := level(l, 0.3, 0.4, 0.5, 0.65, 0.82, 0.95)
	sp.face = func(x, y float64) sample {
		r := math.Hypot(x, y)
		if math.Abs(r-39.5) < 7.2 {
			if nClamp > 0 {
				u, v := around(x, y, nClamp, 0.2)
				if math.Abs(v) < lw(2.2) && u > 0 {
					return sample{h: ground + 2.2, z: 3}
				}
			}
			sec := int(math.Floor((angle(x, y) + math.Pi) / (2 * math.Pi) * 12))
			for g := 0; g < gaps; g++ {
				if sec == (g*5+2)%12 { // a dark gap in the plasma
					return sample{h: ground - 1.2, z: 3, tone: 0.15, seal: true}
				}
			}
			if l >= 3 && math.Abs(r-39.5) < lw(1.6) {
				return sample{h: ground, z: 2, tone: 0.95, seal: true}
			}
			return sample{h: ground - 0.8, z: 1, tone: pt, seal: true}
		}
		return sample{h: ground}
	}
	return sp
}

// fusionLegend is the level 6 ring on a dark mirror with twelve crisp
// clamps, banded plasma and a ring of ports round the inner wall.
func fusionLegend() *spec {
	sp := fusion(6, 0, 0)
	sp.edge += 1.5
	sp.edgeT = 0.27
	sp.field = 2
	sp.fieldTone = stepTone(1)
	sp.band = [2]float64{900, 901}
	sp.face = func(x, y float64) sample {
		r := math.Hypot(x, y)
		u, v := around(x, y, 12, 0)
		if u > 28 && u < 48.6 && math.Abs(v) < 3.3 {
			k := 2
			switch {
			case u < 30.2 || u > 46.4:
				k = 1
			case v < -0.6:
				k = 3
			}
			return sample{h: ground + 2, z: 3, tone: stepTone(k), seal: true}
		}
		d := math.Abs(r - 38)
		if d < 2.2 {
			return flat(2, 4)
		}
		if d < 5.6 {
			return flat(1, 3)
		}
		if d < 8 {
			return flat(3, 0)
		}
		pu, pv := around(x, y, 24, 0)
		if dd := math.Hypot(pu-26.4, pv); dd < 2.9 && r < 31 {
			if dd < 1.8 {
				return flat(3, 0)
			}
			return flat(0, 3)
		}
		if math.Abs(r-52.8) < 1.3 {
			return flat(1, 2)
		}
		return sample{h: ground}
	}
	return sp
}
