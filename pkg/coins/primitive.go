package coins

import "math"

// The Primitive Age: the money cowrie, seen from underneath. Levels 1 to 6 are
// the shell's belly with its slit and teeth, worn to whole; the legendary is a
// ring cowrie's back.

var (
	cowChalk = newRamp(0x3a3630, 0x6e685e, 0xa49c8e, 0xcdc6b8, 0xeee8dc) // level 1
	cowIvory = newRamp(0x3a2e22, 0x7a6450, 0xb8a080, 0xe0ccaa, 0xfaf0dc) // levels 2-3
	cowPink  = newRamp(0x3e2428, 0x7e5258, 0xbc8a8a, 0xe6bcb4, 0xfff0e8) // levels 4-5, legendary body
	cowGold  = newRamp(0x3e2a18, 0x80602e, 0xc49a52, 0xeed08c, 0xfff4d4) // level 6
	cowMouth = newRamp(0x120a08, 0x2a1a14, 0x4a3226, 0x6e5040, 0x98745c) // the slit
	cowSpot  = newRamp(0x2e1a1c, 0x5e3a3e, 0x8e6264, 0xb48886, 0xd4aaa4) // spots, levels 5-6
)

// cowrie is one shell in its own frame: u across, v along, the front end at negative v.
type cowrie struct {
	lie, a, b, tp float64 // tilt in radians, half-width, half-length, taper
	hx, hy, hr    float64 // hole centre and radius; hr 0 for none
	H, sc, w0, tb float64 // dome height, slit curve, slit half-width, lip width
	pitch, tph    float64 // teeth pitch and phase; pitch 0 for none
	cell          float64 // spot noise cell
	ns            int     // spot noise seed
	spots         bool
}

// frame turns a point of the piece into the shell's frame.
func (c *cowrie) frame(x, y float64) (float64, float64) { return rotate(x, y, -c.lie) }

// aw is the half-width at v; the back end (positive v) is the wider.
func (c *cowrie) aw(v float64) float64 { return c.a * (1 - c.tp*clamp(v/c.b, -1, 1)) }

// q is 1 on the outline and below 1 inside.
func (c *cowrie) q(u, v float64) float64 { return math.Hypot(u/c.aw(v), v/c.b) }

// dome is the height of the flattened belly at (u, v).
func (c *cowrie) dome(u, v float64) float64 {
	q := c.q(u, v)
	return ground + c.H*math.Sqrt(math.Max(0, 1-q*q))
}

// outline is the signed outline, negative inside, with the hole cut out.
func (c *cowrie) outline(u, v float64) float64 {
	sd := (c.q(u, v) - 1) * c.aw(v)
	if c.hr > 0 {
		sd = sdSub(sd, sdCircle(u, v, c.hx, c.hy, c.hr))
	}
	return sd
}

// slit is the slit's centre line at v and its half-width there.
func (c *cowrie) slit(v float64) (float64, float64) {
	return c.sc * math.Sin(v/c.b*math.Pi/2), c.w0 * (0.4 + 0.6*math.Sqrt(math.Max(0, 1-(v/c.b)*(v/c.b))))
}

// face is the belly: the slit, the teeth on its lips, and the spots outside them.
func (c *cowrie) face(x, y float64) sample {
	u, v := c.frame(x, y)
	h := c.dome(u, v)
	mid, ws := c.slit(v)
	d := math.Abs(u - mid)
	switch {
	case d < ws && math.Abs(v) < c.b-3:
		return sample{h: ground - 1.5, z: 1}
	case d >= ws && d < ws+c.tb && math.Abs(v) < c.b-8 && c.pitch > 0:
		if frac((v+c.tph)/c.pitch) < 0.5 {
			h++
		}
	case d >= ws+c.tb && c.spots && valueNoise(u, v, c.cell, c.ns) > 0.78:
		return sample{h: h, z: 3}
	}
	return sample{h: h}
}

// primitive is the Primitive Age's cowrie at level l (1 to 6), lod 0 or 1. Seed 0 is the plain table.
func primitive(l, lod int, seed uint32) *spec {
	v, small := 0.0, float64(lod)
	if seed != 0 {
		v = 1
	}
	pitch := level(l, 0, 10, 8, 7, 6, 5)
	if pitch > 0 {
		pitch = math.Max(pitch, 8*small)
	}
	c := &cowrie{
		lie: (20 + v*rs(seed, 1)*15) * math.Pi / 180,
		a:   38 + 2*small + v*rs(seed, 2)*3,
		b:   56 + 2*small + v*rs(seed, 3)*3,
		tp:  -0.12 + v*rs(seed, 4)*0.06,
		H:   level(l, 3.0, 3.5, 4.0, 4.5, 5.0, 5.0),
		sc:  3 + v*rs(seed, 5)*3,
		w0:  math.Max(level(l, 5.0, 4.5, 4.0, 3.5, 3.0, 3.0), 3.5*small),
		tb:  6 + small,
		tph: v * rnd(seed, 6) * pitch, pitch: pitch,
		cell: 4.5 + 2.5*small, ns: 5 + int(v*rnd(seed, 7)*1000),
		spots: l >= 5,
	}
	switch {
	case l == 1: // a worn hole beside the slit
		c.hx, c.hy, c.hr = 8, 6, 8+small
	case l <= 3: // a pierced stringing hole at the front end
		c.hx, c.hy, c.hr = 0, 11-c.b, 4+small
	}
	c.hx += v * rs(seed, 8) * 4
	c.hy += v * rs(seed, 9) * 4
	body := [6]ramp{cowChalk, cowIvory, cowIvory, cowPink, cowPink, cowGold}[l-1]
	sp := &spec{seed: 400, edge: 4, field: int(level(l, 0, 0, 1, 1, 1, 1)),
		ramp: [4]ramp{body, cowMouth, body, cowSpot}, bright: level(l, 0, 0, 0, 0, 0, 0.06)}
	sp.sd = func(x, y float64) float64 { return c.outline(c.frame(x, y)) }
	sp.face = c.face
	return sp
}

// primitiveLegend is the ring cowrie's back: a pale pink dome with a gold ring and no slit. It is still.
func primitiveLegend() *spec {
	c := &cowrie{lie: 20 * math.Pi / 180, a: 44, b: 60, tp: -0.10, H: 7}
	sp := &spec{seed: 400, field: 1, edge: 5, bright: 0.08, ramp: [4]ramp{cowPink, gilt, cowPink, cowSpot}}
	sp.sd = func(x, y float64) float64 { return c.outline(c.frame(x, y)) }
	sp.face = func(x, y float64) sample {
		u, v := c.frame(x, y)
		h := c.dome(u, v)
		q2 := math.Hypot(u/(c.aw(v)*0.95), (v-4)/c.b)
		switch {
		case math.Abs(q2-0.55) < 0.07:
			return sample{h: h + 0.3, z: 1}
		case q2 < 0.48 && valueNoise(u, v, 6, 9) > 0.78:
			return sample{h: h, z: 3}
		}
		return sample{h: h}
	}
	return sp
}
