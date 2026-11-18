package coins

import "math"

var (
	titanium = newRamp(0x2a2c34, 0x50525e, 0x80848f, 0xb4b8c4, 0xe8ecf4)
	blueTi   = newRamp(0x0c1a3a, 0x183a78, 0x2a64b8, 0x60a0e8, 0xb0d8ff)
	steel    = newRamp(0x0e1626, 0x1c2a42, 0x30466a, 0x4a6890, 0x6c8cb8)
)

// space is a machined titanium token with eight scallops in its rim, grooves
// and a domed centre. Every copy is drawn alike.
func space(l, lod int, _ uint32) *spec {
	lw := lineWidth(lod)
	sp := &spec{seed: 51, edge: 5, rimW: 6, rimH: level(l, 1, 1.3, 1.6, 1.8, 2, 2)}
	sp.sd = func(x, y float64) float64 {
		u, v := around(x, y, 8, math.Pi/8)
		return sdSub(sdCircle(x, y, 0, 0, 55), sdCircle(u, v, 55, 0, 5.2))
	}
	ti := titanium
	if l <= 2 {
		ti = steel
	}
	sp.ramp = [4]ramp{ti, blueTi, orange, graphite}
	switch {
	case l == 3:
		sp.field = 1
	case l > 3:
		sp.field = 2
		sp.fieldTone = 0.3
		sp.band = [2]float64{-36, -26}
	}
	if l >= 5 {
		er := blueTi
		sp.edgeR = &er
		sp.edgeT = 0.5
	}
	grooves := []float64{28}
	if l >= 3 {
		grooves = []float64{36, 28, 20}
	}
	sp.face = func(x, y float64) sample {
		r := math.Hypot(x, y)
		if l >= 4 && r > 40.5 && r < 48 {
			z, t := 1, 0.5
			if l == 6 {
				a := angle(x, y)
				if int(math.Floor((a+math.Pi)/(math.Pi/4)))&1 == 1 {
					z, t = 2, 0.58
				}
				if sa := math.Mod(a+math.Pi, math.Pi/4); sa < 0.03 || sa > math.Pi/4-0.03 {
					return sample{h: ground}
				}
			}
			return sample{h: ground, z: z, tone: t, seal: true}
		}
		if l <= 2 && r > 38.5 && r < 47 { // a lit arc about 9 px wide at 128
			if a := angle(x, y); a > -2.2 && a < level(l, -1.0, 0.4) {
				return sample{h: ground, z: 1, tone: 0.7, seal: true}
			}
		}
		if l >= 2 {
			for _, g := range grooves {
				if math.Abs(r-g) < lw(1.3) {
					return sample{h: ground - 1.1, z: 3, tone: 0.3, seal: true}
				}
			}
		}
		if l >= 3 && r < 13 {
			return sample{h: ground + 3.6*math.Sqrt(math.Max(0, 1-(r/13)*(r/13)))}
		}
		if l >= 5 {
			bx, by := 28*math.Cos(-0.9), 28*math.Sin(-0.9)
			if br := math.Hypot(x-bx, y-by); br < 4 {
				return sample{h: ground + 3*math.Sqrt(math.Max(0, 1-(br/4)*(br/4)))}
			}
		}
		return sample{h: ground}
	}
	return sp
}

// spaceLegend is the level 6 token with square notches round the rim, a
// porthole at the centre, small bodies inlaid in the grooves and bolt heads.
func spaceLegend() *spec {
	sp := space(6, 0, 0)
	sp.edge += 1.5
	sp.field = 2
	sp.fieldTone = stepTone(1)
	sp.band = [2]float64{-36, -26}
	sp.sd = func(x, y float64) float64 {
		u, v := around(x, y, 8, math.Pi/8) // eight notches, 18 px wide and 12 deep at 208
		return sdSub(sdCircle(x, y, 0, 0, 55), sdBox(u, v, 56, 0, 7.4, 5.5))
	}
	grooves := []float64{34.5, 26.5, 18.8}
	bodyAng := []float64{-1.05, 2.62, 0.52}
	bz := []int{1, 2, 0}
	bl := []int{3, 3, 4}
	bd := []int{1, 1, 2}
	sp.face = func(x, y float64) sample {
		r := math.Hypot(x, y)
		if r < 8.6 { // the porthole
			if math.Hypot(x+3.2, y+3.2) < 1.9 {
				return flat(0, 4)
			}
			return flat(3, 0)
		}
		if r < 11.6 {
			return sample{h: ground + 1.5, tone: stepTone(3), seal: true}
		}
		for i, g := range grooves {
			if s, ok := sphere(x, y, g*math.Cos(bodyAng[i]), g*math.Sin(bodyAng[i]), 5.6, bz[i], bl[i], bd[i]); ok {
				return s
			}
		}
		for _, g := range grooves {
			if math.Abs(r-g) < 1.5 {
				return flat(3, 0)
			}
		}
		if r > 40 && r < 44.4 { // two hard anodized bands
			return flat(1, 2)
		}
		if r >= 44.4 && r < 48.6 {
			return flat(2, 3)
		}
		bu, bv := around(x, y, 16, math.Pi/16) // bolt heads between the notches
		if dd := math.Hypot(bu-51.7, bv); dd < 2.9 && r < 54.6 {
			if dd < 2.1 {
				return sample{h: ground + 1.6, tone: stepTone(3), seal: true}
			}
			return flat(3, 0)
		}
		return sample{h: ground}
	}
	return sp
}
