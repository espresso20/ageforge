package coins

import "math"

var (
	nickel   = newRamp(0x3a4250, 0x7a8496, 0xbcc6d6, 0xe6ecf4, 0xffffff)
	graphite = newRamp(0x2a2f38, 0x3c424d, 0x505866, 0x6a7384, 0x8a93a6)
	aluDull  = newRamp(0x1c2128, 0x3a424e, 0x5e6876, 0x828d9d, 0xa8b3c3)
	paleGold = newRamp(0x4a4230, 0x8a7c54, 0xd4c48c, 0xf2e8c0, 0xfffbea)
)

// modern is a two-metal coin: an outer ring and a centre of another metal,
// with the join between them showing even at level 1. Every copy is drawn alike.
func modern(l, _ int, _ uint32) *spec {
	sp := &spec{seed: 171, edge: 6, rimW: 5, rimH: level(l, 1, 1.3, 1.6, 1.8, 1.9, 1.9)}
	sp.sd = func(x, y float64) float64 { return sdCircle(x, y, 0, 0, 57) }
	ni := nickel
	if l <= 2 {
		ni = aluDull
	}
	sp.ramp = [4]ramp{ni, gilt, nickel, graphite}
	if l >= 3 {
		sp.field = 1
	}
	if l >= 4 {
		sp.field = 2
		sp.fieldTone = 0.3
		sp.band = [2]float64{-36, -28}
	}
	if l >= 5 {
		er := gilt
		sp.edgeR = &er
		sp.edgeT = 0.45
	}
	sp.face = func(x, y float64) sample {
		r := math.Hypot(x, y)
		a := angle(x, y)
		z := 0
		h := ground
		if l == 1 { // a dull ring, a centre slug of a different grey, the join between them
			if math.Abs(r-36) < 1.3 {
				return sample{h: ground - 1.4, z: 3, tone: 0.12, seal: true}
			}
			if r < 36 {
				return sample{h: ground + 0.5, z: 3, tone: 0.52, seal: true}
			}
			return sample{h: h}
		}
		if r < 38 && l == 6 {
			z = 1
		}
		if math.Abs(r-39) < 1.1 {
			h += 1.1
			z = 0
		}
		if l >= 3 && r > 41 && r < 47 && math.Sin(a*80) > 0.2 {
			z = 3
		}
		if l >= 5 && r < 30 {
			cx, cy := math.Mod(x+80, 6), math.Mod(y+80, 6)
			if math.Abs(cx-3) < 1 && math.Abs(cy-3) < 1 && math.Max(math.Abs(x), math.Abs(y)) < 24 {
				return sample{h: ground + 0.9, z: z}
			}
		}
		if r < 12 {
			h += 1.2
		}
		return sample{h: h, z: z}
	}
	return sp
}

// modernLegend is the level 6 coin in pale gold round a heptagon of nickel,
// with ticks in the outer ring and a latent-image panel.
func modernLegend() *spec {
	sp := modern(6, 0, 0)
	sp.edge += 1.5
	sp.ramp = [4]ramp{paleGold, nickel, paleGold, graphite}
	sp.field = 1
	sp.reed = 90
	sp.edgeSeg = 12
	er := paleGold
	sp.edgeR = &er
	sp.edgeT = 0.4
	hept := ngon(7, 37, -math.Pi/2)
	sp.face = func(x, y float64) sample {
		r := math.Hypot(x, y)
		hd := sdPoly(x, y, hept)
		if hd > 0 { // ticks where lettering would be
			u, v := around(x, y, 48, 0)
			if math.Abs(v) < 1.3 && u > 44 && u < 50.5 {
				return sample{h: ground, tone: 0.2, seal: true}
			}
			if hd < 1.7 {
				return sample{h: ground - 0.8, z: 3, tone: 0.1, seal: true}
			}
			return sample{h: ground}
		}
		if r < 12 { // the centre plug of the outer metal
			return sample{h: ground + 1.8 + 1.2*bevel(r-12, 4)}
		}
		if r < 13.8 {
			return sample{h: ground - 0.4, z: 3, tone: 0.1, seal: true}
		}
		h := ground + 1.0
		py := y + 22.5
		if pd := sdRoundBox(x, py, 0, 0, 14, 7.5, 2); pd < 0 { // the latent image panel
			if pd > -1.1 {
				return sample{h: ground - 0.4, z: 3, tone: 0.15, seal: true}
			}
			if frac((x+py)/9) < 0.5 {
				return sample{h: h, z: 1, tone: 0.95, seal: true}
			}
			return sample{h: h, z: 1, tone: 0.5, seal: true}
		}
		w := (x - y) / 1.414
		if (w > -24 && w < -18) || (w > -12 && w < -10) {
			return sample{h: h, z: 1, tone: 0.97, seal: true}
		}
		return sample{h: h, z: 1, tone: 0.74, seal: true}
	}
	return sp
}
