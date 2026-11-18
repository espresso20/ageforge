package coins

import "math"

var (
	chipBody  = newRamp(0x05060c, 0x0b0f1c, 0x131b30, 0x1e2a48, 0x30426c)
	chipGlass = newRamp(0x02030a, 0x0a1024, 0x1a2a58, 0x4a6ab0, 0xc0d8ff)
	chipDim   = newRamp(0x04202a, 0x08333f, 0x0c4a58, 0x0f6a7a, 0x1b8a9e)
	cyan      = newRamp(0x06303a, 0x0f6a7a, 0x30b0c8, 0x80e0f0, 0xd0fcff)
	magenta   = newRamp(0x3a0830, 0x7a1060, 0xc01a90, 0xf060d0, 0xffc0f0)
	deadPad   = newRamp(0x1a1a22, 0x26262f, 0x34343f, 0x46465a, 0x5e5e72)
)

var chipOutline = []vec{{-40, -40}, {52, -40}, {56, -36}, {56, 24}, {40, 40}, {-52, 40}, {-56, 36}, {-56, -24}}

var chipTraces = [][]vec{
	{{-19, -3}, {6, -3}, {18, -15}, {46, -15}},
	{{-19, 5}, {26, 5}, {36, 15}, {46, 15}},
	{{-19, 13}, {0, 13}, {8, 21}, {30, 21}},
	{{-22, -5}, {-22, -18}, {-8, -18}},
	{{-22, -18}, {-22, -30}, {14, -30}},
	{{-30, 17}, {-30, 28}, {20, 28}},
	{{18, -15}, {26, -26}, {42, -26}},
	{{36, 15}, {36, 26}, {48, 26}},
	{{-19, 9}, {10, 9}},
}

// cyberpunk is a credit chip: a notched card with lit traces, a contact pad and
// a holographic mark. Every copy is drawn alike.
func cyberpunk(l, lod int, _ uint32) *spec {
	lw := lineWidth(lod)
	sp := &spec{seed: 11, edge: 5, rimW: 5, rimH: 1.6}
	sp.sd = func(x, y float64) float64 { return sdPoly(x, y, chipOutline) }
	body := chipBody
	if l >= 5 {
		body = chipGlass
		sp.field = 2
		sp.fieldTone = 0.12
		sp.band = [2]float64{-34, -26}
	}
	pad := deadPad
	if l <= 2 {
		pad = chipDim
	}
	if l == 6 {
		pad = gilt
	}
	sp.ramp = [4]ramp{body, cyan, magenta, pad}
	if l >= 3 {
		er := cyan
		sp.edgeR = &er
		sp.edgeT = level(l, 0, 0, 0.3, 0.4, 0.55, 0.7)
	}
	nT := int(level(l, 1, 2, 4, 6, 8, 9))
	tw := lw(level(l, 3.8, 3.2, 2.6, 2.2, 1.8, 1.6))
	tt := level(l, 0.5, 0.58, 0.5, 0.66, 0.82, 0.97)
	neon := level(l, 0, 0, 0.45, 0.6, 0.8, 1)
	const pcx, pcy = -30.0, 6.0
	sp.face = func(x, y float64) sample {
		d := -sdPoly(x, y, chipOutline)
		if neon > 0 && d > 7 && d < 7+lw(level(l, 0, 0, 1.8, 2, 2.4, 3)) {
			return sample{h: ground, z: 1, tone: neon, seal: true}
		}
		if l == 6 && d > 11.5 && d < 13 {
			return sample{h: ground, z: 2, tone: 0.8, seal: true}
		}
		best := 1e9
		for k := 0; k < nT; k++ {
			best = math.Min(best, polyline(x, y, chipTraces[k], tw/2))
			p := chipTraces[k][len(chipTraces[k])-1] // a node at the last point
			best = math.Min(best, sdCircle(x, y, p.x, p.y, tw*1.15))
		}
		if best < 0 {
			return sample{h: ground, z: 1, tone: tt, seal: true}
		}
		if pd := sdBox(x, y, pcx, pcy, 11, 11); pd < 0 {
			pz, pt := 1, 0.0
			if l <= 2 || l == 6 {
				pz = 3
			}
			h := ground + 2.4*bevel(pd, 3)
			if l >= 5 && (math.Abs(x-pcx) < 0.7 || math.Abs(y-pcy) < 0.7) {
				h--
			}
			if l <= 2 {
				pt = 0.55
			} else if l < 6 {
				pt = level(l, 0, 0, 0.5, 0.6, 0.72, 0)
			}
			return sample{h: h, z: pz, tone: pt}
		}
		if l >= 4 { // the holographic mark
			q := math.Abs(x+36) + math.Abs(y+20)
			if math.Abs(q-10)-lw(1.1) < 0 {
				return sample{h: ground, z: 2, tone: level(l, 0, 0, 0, 0.75, 0.85, 0.97), seal: true}
			}
			if l >= 5 && q < 6 {
				if l == 6 && int(math.Floor(y))&1 == 0 {
					return sample{h: ground}
				}
				return sample{h: ground, z: 2, tone: 0.9, seal: true}
			}
		}
		return sample{h: ground}
	}
	return sp
}

// cyberpunkLegend is the level 6 chip with a hatched holographic panel, a
// divided contact pad, a cut window onto an inner layer, dots at the joints of
// the traces and a second layer of finer traces beneath the first.
func cyberpunkLegend() *spec {
	sp := cyberpunk(6, 0, 0)
	sp.edge += 1.5
	orig := sp.face
	fine := [][]vec{
		{{-15, 0}, {2, 0}, {12, -10}, {40, -10}},
		{{-15, 9}, {22, 9}, {30, 18}, {50, 18}},
		{{-26, -34}, {20, -34}},
		{{-24, 33}, {10, 33}, {16, 38}},
		{{22, -22}, {30, -22}, {36, -30}, {50, -30}},
	}
	sp.face = func(x, y float64) sample {
		if math.Abs(x+36)+math.Abs(y+20) < 9.5 {
			if int(math.Floor((x+y+200)/1.6))&1 == 0 {
				return sample{h: ground, z: 1, tone: 0.95, seal: true}
			}
			return sample{h: ground, z: 2, tone: 0.9, seal: true}
		}
		if sdBox(x, y, -30, 6, 11, 11) < 0 {
			if math.Abs(math.Mod(x+30+11.01, 7.34)-7.34) < 0.45 || math.Abs(math.Mod(y-6+11.01, 7.34)-7.34) < 0.45 {
				return sample{h: ground - 0.8, z: 3, tone: 0.12, seal: true}
			}
			return orig(x, y)
		}
		if e := sdBox(x, y, 29, 31, 9, 3.6); e < 0 {
			if e > -0.9 {
				return sample{h: ground, z: 1, tone: 0.9, seal: true}
			}
			if int(math.Floor(x/1.8))%2 == 0 {
				return sample{h: ground - 1, z: 2, tone: 0.45, seal: true}
			}
			return sample{h: ground - 1, z: 3, tone: 0.1, seal: true}
		}
		for _, tr := range chipTraces {
			for _, p := range tr {
				if math.Hypot(x-p.x, y-p.y) < 1.9 {
					return sample{h: ground + 0.8, z: 1, tone: 1, seal: true}
				}
			}
		}
		st := orig(x, y)
		if st.h == ground && st.tone == 0 && st.z == 0 && !st.seal {
			for _, tr := range fine {
				if polyline(x, y, tr, 0.55) < 0 {
					return sample{h: ground, z: 1, tone: 0.42, seal: true}
				}
			}
		}
		return st
	}
	return sp
}
