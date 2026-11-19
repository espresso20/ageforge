package coins

import "math"

// The Bronze Age: metal cast as ingots and weighed. Level 1 is a bun ingot;
// levels 2 to 6 are the four-lugged oxhide, cleaner as the level rises.

var (
	cuBrown = newRamp(0x221408, 0x4a2c16, 0x6e4426, 0x94643e, 0xc09068) // dull copper, levels 1-3
	cuRed   = newRamp(0x2a0e06, 0x5a2410, 0x8e4020, 0xc0663a, 0xe8a078) // bright copper, level 4
	cuTin   = newRamp(0x2a1a06, 0x5c3c12, 0x946426, 0xc8944a, 0xf0c888) // tin bronze, levels 5-6
	cuGreen = newRamp(0x10241c, 0x1e4636, 0x2e6a52, 0x4e9478, 0x8cc4a8) // green corrosion crust
)

// ingot is one cast slab, centred at (cx, cy) and tilted by lie.
type ingot struct {
	cx, cy, lie    float64
	bun            bool
	hx, hy, bx, by float64 // half sizes; bx bows the top and bottom, by the sides
	q, rg, lip, dh float64 // bun wave phase, roughness, lip height, base height offset
	nr, mark       int     // roughness noise seed; mark strokes, 0 to 3
	mx, my, mr     float64 // mark shift and turn
}

// local turns a point of the piece into the ingot's frame.
func (g *ingot) local(x, y float64) (float64, float64) {
	return rotate(x-g.cx, y-g.cy, -g.lie)
}

// outline is the signed outline in the ingot's frame, negative inside.
func (g *ingot) outline(u, v float64) float64 {
	if g.bun {
		return sdEllipse(u, v, 0, 0, 52, 40, 0) - 2.5*math.Sin(4*angle(u, v)+g.q)
	}
	ex := math.Abs(u) - (g.hx - g.by*(1-math.Min(1, (v/g.hy)*(v/g.hy))))
	ey := math.Abs(v) - (g.hy - g.bx*(1-math.Min(1, (u/g.hx)*(u/g.hx))))
	return math.Max(ex, ey)
}

// body is the height of the slab before marks and crust.
func (g *ingot) body(u, v float64) float64 {
	h := ground
	if g.bun {
		h += 3.0 * math.Sqrt(math.Max(0, 1-(u/52)*(u/52)-(v/40)*(v/40)))
	} else {
		h += 2.0 * math.Sqrt(math.Max(0, 1-(u/g.hx)*(u/g.hx))) * math.Sqrt(math.Max(0, 1-(v/g.hy)*(v/g.hy)))
	}
	h += g.dh + g.rg*2*(fbm(u, v, 6, g.nr)-0.5)
	return h + g.lip*(1-clamp(-g.outline(u, v)/4, 0, 1))
}

// markAt is the distance to the nearest mark stroke minus w, so negative inside
// a stroke of half-width w. With no marks it is far away.
func (g *ingot) markAt(u, v, w float64) float64 {
	pu, pv := rotate(u-g.mx, v-g.my, g.mr)
	strokes := [][]vec{{{-10, -8}, {10, 8}}, {{-10, 8}, {10, -8}}, {{0, -13}, {0, 13}}}
	d := 1e9
	for _, s := range strokes[:g.mark] {
		d = math.Min(d, polyline(pu, pv, s, w))
	}
	return d
}

// bronze is the Bronze Age ingot at level l (1 to 6), lod 0 or 1. Seed 0 is the plain table.
func bronze(l, lod int, seed uint32) *spec {
	lw := lineWidth(lod)
	k := float64(min(seed, 1)) // 0 for the plain copy, 1 for every other seed
	g := &ingot{
		lie: (-8 + k*rs(seed, 1)*12) * math.Pi / 180, bun: l == 1,
		hx: level(l, 0, 54, 55, 56, 56, 56) + k*rs(seed, 3)*2,
		hy: level(l, 0, 40, 38, 36, 35, 34) + k*rs(seed, 4)*2,
		bx: level(l, 0, 5, 8, 10, 11, 12) + k*rs(seed, 5)*2,
		by: level(l, 0, 4, 7, 9, 10, 11) + k*rs(seed, 6)*2,
		q:  1.0 + k*rs(seed, 10)*3.0,
		rg: level(l, 1.4, 1.2, 1.0, 0.7, 0.35, 0.15), lip: level(l, 0, 0, 0, 0, 0.9, 1.2),
		nr: 11 + int(k*rnd(seed, 2)*1000), mark: int(level(l, 0, 0, 0, 1, 2, 3)),
		mx: k * rs(seed, 7) * 5, my: k * rs(seed, 8) * 4, mr: k * rs(seed, 9) * 20 * math.Pi / 180,
	}
	crust, nc := l <= 4, 7+int(k*rnd(seed, 2)*1000)
	t := level(l, 0.38, 0.48, 0.60, 0.72, 0, 0)
	bodyRamp := [6]ramp{cuBrown, cuBrown, cuBrown, cuRed, cuTin, cuTin}[l-1]
	sp := &spec{seed: 402, field: 3, edge: level(l, 4, 4, 4.5, 4.5, 5, 5),
		ramp: [4]ramp{bodyRamp, cuGreen, bodyRamp, bodyRamp}, bright: level(l, 0, 0, 0, 0, 0, 0.06)}
	sp.sd = func(x, y float64) float64 { return g.outline(g.local(x, y)) }
	sp.face = func(x, y float64) sample {
		u, v := g.local(x, y)
		h := g.body(u, v)
		if crust && fbm(u, v, 14, nc) > t {
			return sample{h: h + 0.6, z: 1}
		}
		if g.markAt(u, v, lw(1.6)) < 0 {
			h -= 1.2
		}
		return sample{h: h}
	}
	return sp
}

// bronzeLegend is two oxhide ingots crossed: red copper under tin bronze, a gilt mark on the upper one.
func bronzeLegend() *spec {
	lw := lineWidth(0)
	up := &ingot{cx: -2, cy: -4, lie: -14 * math.Pi / 180, hx: 54, hy: 33, bx: 12, by: 11, rg: 0.15, lip: 1.2, nr: 11, mark: 3}
	lo := &ingot{cx: 4, cy: 6, lie: 62 * math.Pi / 180, hx: 50, hy: 30, bx: 10, by: 9, rg: 0.35, nr: 11, dh: -1}
	sp := &spec{seed: 402, field: 3, edge: 5, bright: 0.08, ramp: [4]ramp{cuTin, gilt, cuRed, cuTin}}
	sp.sd = func(x, y float64) float64 { return math.Min(up.outline(up.local(x, y)), lo.outline(lo.local(x, y))) }
	sp.face = func(x, y float64) sample {
		if u, v := up.local(x, y); up.outline(u, v) < 0 {
			h := up.body(u, v)
			if up.markAt(u, v, lw(2.0)) < 0 {
				return sample{h: h + 0.4, z: 1}
			}
			return sample{h: h}
		}
		u, v := lo.local(x, y)
		return sample{h: lo.body(u, v), z: 2}
	}
	return sp
}
