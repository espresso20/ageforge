package coins

import "math"

// The Iron Age: a thick gold coin in a shallow cup, abstract relief only.

var (
	celtPale = newRamp(0x2e2a14, 0x5e5428, 0x9a8c48, 0xcabc7a, 0xf0e6b8) // levels 1-2
	celtGold = newRamp(0x3a2408, 0x7a5414, 0xc08a20, 0xe8b840, 0xfde488) // levels 3-5
	celtRed  = newRamp(0x3a1608, 0x7a3412, 0xb8621e, 0xe0943e, 0xf8cc80) // level 6, legendary
)

// celtCoin is one piece: outline, cup and relief. dots are pellets (x, y, r).
type celtCoin struct {
	r0, a1, a2, depth, p1, p2 float64
	rot, sx, sy               float64
	dots                      [][3]float64
	crescent, outer           float64 // ring half-widths, 0 none
	torc, twist               bool
	rt, tw, kr, gap, gapHalf  float64 // torc radius, half-width, knob radius, gap centre and half-width
	ra, aw, er                float64 // whorl arm radius, half-width, end pellet radius
	beads                     int
	br                        float64
}

// celtRing is the height of a ring of radius rt and half-width tw, 0 off it.
func celtRing(rad, rt, tw float64) float64 {
	d := math.Abs(rad - rt)
	return 1.6 * math.Sqrt(math.Max(0, 1-(d/tw)*(d/tw))) // 0 off the ring
}

// celtTri is a three-armed whorl through the origin, with a pellet at each far end.
func celtTri(x, y, ra, aw, er float64) float64 {
	h := 0.0
	for k := 0.0; k < 3; k++ {
		phi := -math.Pi/2 + 2*math.Pi/3*k
		cx, cy := ra*math.Cos(phi), ra*math.Sin(phi)
		ux, uy := x-cx, y-cy
		a, d := math.Mod(angle(ux, uy)-phi-math.Pi+4*math.Pi, 2*math.Pi), math.Abs(math.Hypot(ux, uy)-ra)
		if a <= 220*math.Pi/180 && d < aw {
			h = math.Max(h, 1.6*math.Sqrt(1-(d/aw)*(d/aw)))
		}
		h = math.Max(h, dome(x, y, cx+ra*math.Cos(phi+40*math.Pi/180), cy+ra*math.Sin(phi+40*math.Pi/180), er, 1.8))
	}
	return h
}

// relief is the height above the cup at a point of the design frame, 0 off it.
func (c *celtCoin) relief(dx, dy float64) float64 {
	h, rad, a := 0.0, math.Hypot(dx, dy), angle(dx, dy)
	for _, p := range c.dots {
		h = math.Max(h, dome(dx, dy, p[0], p[1], p[2], 1.8))
	}
	if c.crescent > 0 && a > -8*math.Pi/9 && a < -math.Pi/9 {
		h = math.Max(h, celtRing(rad, 26, c.crescent))
	}
	if c.torc {
		t := 0.0
		if math.Abs(wrap(a-c.gap)) >= c.gapHalf {
			t = celtRing(rad, c.rt, c.tw)
		}
		if c.twist && t > 0 && frac(a*30/4.5+(rad-c.rt)/(2*c.tw)) < 0.3 {
			t = math.Max(t-0.8, 0.2)
		}
		h = math.Max(h, t)
		for _, s := range []float64{-1, 1} { // knobs at the two ends of the gap
			k := c.gap + s*c.gapHalf
			h = math.Max(h, dome(dx, dy, c.rt*math.Cos(k), c.rt*math.Sin(k), c.kr, 2.2))
		}
	}
	if c.outer > 0 {
		h = math.Max(h, celtRing(rad, 41, c.outer))
	}
	if c.ra > 0 {
		h = math.Max(h, celtTri(dx, dy, c.ra, c.aw, c.er))
	}
	return h
}

// build sets the outline, cup and relief on sp.
func (c *celtCoin) build(sp *spec) *spec {
	rAt := func(a float64) float64 { return c.r0 + c.a1*math.Sin(3*a+c.p1) + c.a2*math.Sin(5*a+c.p2) }
	sp.sd = func(x, y float64) float64 { return math.Hypot(x, y) - rAt(angle(x, y)) }
	sp.face = func(x, y float64) sample {
		s := clamp(math.Hypot(x, y)/rAt(angle(x, y)), 0, 1)
		h := c.relief(rotate(x-c.sx, y-c.sy, -c.rot))
		for i := 0; i < c.beads; i++ { // beads follow the outline, not the design frame
			a := 2*math.Pi*float64(i)/float64(c.beads) + 0.1
			br := rAt(a) - 6.5
			h = math.Max(h, dome(x, y, br*math.Cos(a), br*math.Sin(a), c.br, 1.4))
		}
		if h > 0 {
			return sample{h: ground + c.depth*s*s + h, z: 1}
		}
		return sample{h: ground + c.depth*s*s}
	}
	return sp
}

// celtTable is the plain table for level l at lod, before any seed variation.
func celtTable(l, lod int) *celtCoin {
	big := lod == 1
	pr := level(l, 4.5, 4.5, 4, 3.4)
	c := &celtCoin{
		r0: level(l, 52, 53, 54, 55, 56, 57), a1: level(l, 3.5, 2.5, 1.8, 1.2, 0.8, 0.3),
		a2: level(l, 2, 1.2, 0.8, 0.5, 0.3, 0), depth: level(l, 3, 4, 5, 5, 5.5, 5.5),
		rt: 30, tw: 3.2, kr: 5, gapHalf: 28 * math.Pi / 180, torc: l >= 3, twist: l >= 5 && !big,
		beads: int(level(l, 0, 0, 0, 12, 16, 20)), br: 2,
	}
	if big {
		c.tw, c.kr, c.br, c.beads, pr = 4.2, 6, 3, int(level(l, 0, 0, 0, 8, 10, 12)), level(l, 5.5, 5.5, 5, 4.5)
	}
	switch l {
	case 1:
		c.dots = [][3]float64{{-8, -6, pr}, {6, -9, pr}, {0, 7, pr}}
	case 2:
		c.dots = [][3]float64{{-9, 4, pr}, {9, 4, pr}, {0, 14, pr}}
		c.crescent = lineWidth(lod)(3)
		if big {
			c.crescent = 3.6
		}
	case 3:
		c.dots = [][3]float64{{-9, -4, pr}, {9, -4, pr}, {0, 8, pr}}
	case 4:
		c.dots = [][3]float64{{0, 0, pr}}
		for i := 0; i < 6; i++ {
			if a := (30 + 60*float64(i)) * math.Pi / 180; !big || i%2 == 0 {
				c.dots = append(c.dots, [3]float64{14 * math.Cos(a), 14 * math.Sin(a), pr})
			}
		}
	case 5:
		c.ra, c.aw = 9, 2
	case 6:
		c.ra, c.aw, c.outer = 11, 2.4, 2.4
	}
	if big && l >= 5 {
		c.aw = 3
	}
	if big && l == 6 {
		c.outer = 3.2
	}
	c.er = c.aw + 1.2
	if big {
		c.er = 4.5
	}
	return c
}

// iron is the Iron Age cup coin at level l (1 to 6), lod 0 or 1. Seed 0 is the plain table.
func iron(l, lod int, seed uint32) *spec {
	c := celtTable(l, lod)
	v := float64(min(seed, 1)) // 0 for the plain copy, 1 for every other seed
	c.p1, c.p2, c.a1 = 1+v*rs(seed, 1)*3.1, 2+v*rs(seed, 2)*3.1, math.Max(0, c.a1+v*rs(seed, 3))
	c.rot, c.sx, c.sy = v*rs(seed, 4)*25*math.Pi/180, v*rs(seed, 5)*4, v*rs(seed, 6)*4
	for i := range c.dots {
		c.dots[i][0] += v * rs(seed, 10+i) * 1.5
		c.dots[i][1] += v * rs(seed, 20+i) * 1.5
	}
	c.gap = math.Pi/2 + v*rs(seed, 30)*15*math.Pi/180
	b := [6]ramp{celtPale, celtPale, celtGold, celtGold, celtGold, celtRed}[l-1]
	sp := c.build(&spec{seed: 401, field: 3, edge: level(l, 5.5, 5, 5, 4.5, 4.5, 4), ramp: [4]ramp{b, b, b, b}, bright: 0.06 * float64(l/6)})
	handStruck(sp, l)
	return sp
}

// ironLegend is the gilt relief on red gold: a large whorl, a torc gapped at the top, beads.
func ironLegend() *spec {
	c := &celtCoin{r0: 58, depth: 6, rt: 42, tw: 3, kr: 5.5, torc: true, twist: true, gap: -math.Pi / 2, gapHalf: 22 * math.Pi / 180, ra: 20, aw: 3.2, er: 5, beads: 24, br: 2.2}
	return c.build(&spec{seed: 401, field: 3, edge: 5, bright: 0.08, ramp: [4]ramp{celtRed, gilt, celtRed, gilt}})
}
