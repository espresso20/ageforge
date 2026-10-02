package skyline

import (
	"math"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
)

// orbit_deep.go is the Interstellar Age: deep space with a distant sun. No
// planet below: a violet-black void where the first colour in space shows,
// thin nebula wisps in violet and teal drifting far off; the home system
// shrunk to an orrery in the upper left; the colony worlds as small planets
// in the far layer, each with a blinking beacon, more of them as the
// civilization holds more lineages; the lots in formation on a faint dotted
// line as arks, small worlds and frigates; and standing over the present,
// the warp gate, its ring built as far as the engineers have got and
// shimmering once the Warp Nexus closes it.

// deepBackdrop draws everything behind the lots.
func (o *orb) deepBackdrop() {
	o.nebula()
	o.stars(0, o.S, 0.018, 0.015, 557)
	o.colonies()
	o.orrery()
	o.gate()
	o.formation()
}

// ------------------------------------------------------------ textures

// texture is a precomputed field kept in the sprite cache: each cell holds
// two bytes (in its fg and bg slots). Fields depend on the seed and the
// size only, so they survive model rebuilds.
func (v *view) texture(name string, seed int64, w, h int, gen func(x, y int) (uint8, uint8)) *sprite {
	if v.sprites == nil || len(v.sprites) > 6000 {
		v.sprites = map[spriteKey]*sprite{}
	}
	k := spriteKey{"\x00tex:" + name, uint64(seed), w<<12 | h}
	if t := v.sprites[k]; t != nil {
		return t
	}
	t := newSprite(w, h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			a, b := gen(x, y)
			t.c[y*w+x] = scell{fg: slot(a), bg: slot(b)}
		}
	}
	v.sprites[k] = t
	return t
}

// texAt reads a texture cell, wrapping x.
func texAt(t *sprite, x, y int) (uint8, uint8) {
	if y < 0 || y >= t.h {
		return 0, 0
	}
	x %= t.w
	if x < 0 {
		x += t.w
	}
	c := t.c[y*t.w+x]
	return uint8(c.fg), uint8(c.bg)
}

// pnoise is value noise that repeats every period lattice cells along x,
// so a texture of period × scale columns wraps without a seam.
func pnoise(seed int64, x, y float64, period int) float64 {
	x0, y0 := math.Floor(x), math.Floor(y)
	sx, sy := x-x0, y-y0
	sx = float64(sx*sx) * (3 - float64(2*sx))
	sy = float64(sy*sy) * (3 - float64(2*sy))
	ix := int64(x0) % int64(period)
	if ix < 0 {
		ix += int64(period)
	}
	ix1 := (ix + 1) % int64(period)
	iy := int64(y0)
	lerp := func(a, b, t float64) float64 { return a + float64((b-a)*t) }
	top := lerp(mapmodel.HashF(seed, ix, iy), mapmodel.HashF(seed, ix1, iy), sx)
	bot := lerp(mapmodel.HashF(seed, ix, iy+1), mapmodel.HashF(seed, ix1, iy+1), sx)
	return lerp(top, bot, sy)
}

// turns is the angle of (dx, dy) in turns, 0..1, counterclockwise from +x
// with y down (a polynomial atan2: the map styles keep off package math's
// transcendental functions).
func turns(dx, dy float64) float64 {
	ax, ay := math.Abs(dx), math.Abs(dy)
	if ax == 0 && ay == 0 {
		return 0
	}
	a := math.Min(ax, ay) / math.Max(ax, ay)
	s := float64(a * a)
	r := float64(float64(float64(float64(-0.0464964749*s)+0.15931422)*s)-0.327622764)*s*a + a
	if ay > ax {
		r = 1.57079637 - r
	}
	if dx < 0 {
		r = 3.14159274 - r
	}
	if dy > 0 {
		r = -r
	}
	t := r / (2 * math.Pi)
	if t < 0 {
		t++
	}
	return t
}

// ------------------------------------------------------------ nebula

const nebulaW = 480 // the nebula texture's period in columns

// nebula is the Interstellar's thin wisps of violet and teal, far off and
// drifting: a ridged noise field, thresholded to filaments.
func (o *orb) nebula() {
	seed := int64(o.m.Seed%7919) + 5
	S := o.S
	t := o.v.texture("neb", seed, nebulaW, S, func(x, y int) (uint8, uint8) {
		fx, fy := float64(x), float64(y)
		n1 := pnoise(seed, fx/24, fy/7, nebulaW/24)
		r := 1 - math.Abs(2*n1-1)
		n2 := pnoise(seed+1, fx/12, fy/3.5, nebulaW/12)
		r *= 0.62 + 0.38*n2
		v := smoothstep(0.64, 0.97, r)
		c := pnoise(seed+2, fx/60, fy/14, nebulaW/60)
		return uint8(v * 255), uint8(c * 255)
	})
	void := o.c(mapmodel.InkVoid, iBack, 0)
	vio := o.c(mapmodel.InkCloud, iBack, 0)
	teal := o.c(mapmodel.InkCloud2, iBack, 0)
	if o.light {
		vio, teal = o.c(mapmodel.InkCloud, iEmit, 2), o.c(mapmodel.InkCloud2, iEmit, 2)
	}
	off := int(float64(o.cam)*0.08 + float64(o.anim)*0.03)
	for y := 0; y < S; y++ {
		for x := 0; x < o.W; x++ {
			a, b := texAt(t, x+off, y)
			if a < 18 {
				continue
			}
			v := float64(a) / 255
			col := vio
			if b > 128 {
				col = teal
			}
			bg := theme.Mix(void, col, float64(v*0.34))
			if v > 0.82 && (x+off+y)%2 == 0 {
				o.fb.set(x, o.Y(y), '░', theme.Mix(void, col, 0.62), bg, dNebula)
			} else {
				o.fb.set(x, o.Y(y), ' ', bg, bg, dNebula)
			}
		}
	}
}

// ------------------------------------------------------------ the orrery

// pixels draws a shape in half-block pixels over cells [x0, x1] x [y0, y1]
// at depth d: px gives a pixel's colour (column x, half row py) and
// whether the shape covers it. A half-covered cell keeps what is behind
// its other half.
func (o *orb) pixels(x0, x1, y0, y1 int, d uint8, px func(x, py int) (tcell.Color, bool)) {
	for y := max(0, y0); y <= y1 && y < o.S; y++ {
		for x := max(0, x0); x <= x1 && x < o.W; x++ {
			ct, ot := px(x, 2*y)
			cb, ob := px(x, 2*y+1)
			switch {
			case ot && ob:
				if ct == cb {
					o.fb.fill(x, o.Y(y), ct, d)
				} else {
					o.fb.set(x, o.Y(y), '▀', ct, cb, d)
				}
			case ot:
				o.fb.fg(x, o.Y(y), '▀', ct, d)
			case ob:
				o.fb.fg(x, o.Y(y), '▄', cb, d)
			}
		}
	}
}

// orrery is the home system receding: a small bright sun in its corona,
// three planets on flattened orbits passing in front of it and behind.
func (o *orb) orrery() {
	gy := o.groundY
	cx := float64(o.W)*0.14 - float64(o.cam)*0.004
	cpy := float64(2*max(4, gy/4)) + 1
	sun := o.c(mapmodel.InkAccent, iEmit, 0)
	hot := theme.Mix(sun, o.c(mapmodel.InkStarBright, iEmit, 0), 0.5)
	corona := o.c(mapmodel.InkAccent3, iEmit, 0)
	orbit := theme.Mix(o.c(mapmodel.InkFrameDim, iEmit, 0), o.c(mapmodel.InkStar, iEmit, 1), 0.35)
	void := o.c(mapmodel.InkVoid, iBack, 0)
	R := math.Max(2.5, float64(gy)/8)
	halo := R * 2.1
	// the corona: a soft glow fading into the void
	o.pixels(int(cx-halo)-1, int(cx+halo)+1, int((cpy-halo)/2)-1, int((cpy+halo)/2)+1, dOrrery,
		func(x, py int) (tcell.Color, bool) {
			dx, dy := float64(x)+0.5-cx, float64(py)+0.5-cpy
			d := math.Sqrt(float64(dx*dx)+float64(dy*dy)) / halo
			if d > 1 {
				return 0, false
			}
			return theme.Mix(corona, void, math.Min(1, float64(d*d)*1.1+0.12)), true
		})
	o.pixels(int(cx-R)-1, int(cx+R)+1, int((cpy-R)/2)-1, int((cpy+R)/2)+1, dOrrSun,
		func(x, py int) (tcell.Color, bool) {
			dx, dy := float64(x)+0.5-cx, float64(py)+0.5-cpy
			d := math.Sqrt(float64(dx*dx)+float64(dy*dy)) / R
			if d > 1 {
				return 0, false
			}
			if d < 0.55 {
				return hot, true
			}
			return sun, true
		})
	cy := cpy / 2
	planets := [3]mapmodel.SkyInk{mapmodel.InkSurface, mapmodel.InkRock, mapmodel.InkSurface2}
	for k := 0; k < 3; k++ {
		rx := R*2.6 + float64(k)*R*1.9
		ry := rx * 0.24
		steps := int(rx * 3)
		for i := 0; i < steps; i += 2 {
			t := float64(i) / float64(steps)
			x := int(math.Round(cx + rx*mapmodel.Cos(t)))
			y := int(math.Round(cy + ry*mapmodel.Sin(t)))
			if c := o.fb.at(x, o.Y(y)); c != nil && c.d >= dOrrery {
				o.fb.fg(x, o.Y(y), '·', orbit, dOrrery)
			}
		}
		a := float64(o.anim)/float64(320+k*230) + float64(k)*0.37
		px := int(math.Round(cx + rx*mapmodel.Cos(a)))
		py := int(math.Round(cy + ry*mapmodel.Sin(a)))
		d := dOrrFront // the near side: in front of the sun
		if mapmodel.Sin(a) < 0 {
			d = dOrrBack // the far side: behind it
		}
		o.fb.fg(px, o.Y(py), '●', o.c(planets[k], iEmit, 0), d)
		if k == 2 { // the outer world keeps a moon
			b := a * 5
			o.fb.fg(px+int(math.Round(2*mapmodel.Cos(b))), o.Y(py), '·', o.c(mapmodel.InkMoon, iLit, 1), d)
		}
	}
}

// ------------------------------------------------------------ colony worlds

// colonies are the colony worlds in the far layer: one for every couple of
// lineages held, each a small disc lit from the far sun, its night side's
// beacon blinking.
func (o *orb) colonies() {
	n := clampInt(2+len(o.m.SkyLineages())/2, 2, 10)
	gy := o.groundY
	span := o.W + int(float64(max(0, o.m.Skyline.Width-o.W))*0.1)
	shift := int(float64(o.cam) * 0.1)
	beacon := o.c(mapmodel.InkLight, iEmit, 1)
	void := o.c(mapmodel.InkVoid, iBack, 0)
	inks := [3]mapmodel.SkyInk{mapmodel.InkSurface, mapmodel.InkSurface2, mapmodel.InkRock}
	for i := 0; i < n; i++ {
		x := int(hash(i, 61, int(o.m.Seed%977))%uint64(max(1, span))) - shift
		y := 2 + int(hash(i, 67)%uint64(max(1, gy*6/10)))
		rr := 2 + int(hash(i, 71)%3)
		if x < -12 || x > o.W+12 {
			continue
		}
		base := theme.SpaceColor(o.pal[inks[i%3]])
		litC := o.resolve(theme.Shade(base, 1.25), iLit, 1)
		mid := o.resolve(base, iLit, 2)
		dark := o.resolve(theme.Shade(base, 0.4), iLit, 2)
		rim := o.c(mapmodel.InkGlow, iEmit, 2)
		R := float64(2 * rr)
		fx, cpy := float64(x)+0.5, float64(2*y)+1
		seed := int64(i*977) + 3
		o.pixels(x-2*rr-1, x+2*rr+1, y-rr-1, y+rr+1, dColony, func(xx, py int) (tcell.Color, bool) {
			dx := (float64(xx) + 0.5 - fx) / R
			dy := (float64(py) + 0.5 - cpy) / R
			d := float64(dx*dx) + float64(dy*dy)
			if d > 1 {
				return 0, false
			}
			lt := -dx*0.8 - dy*0.3 // lit from the upper left, where the home sun is
			c := mid
			switch {
			case d > 0.8 && lt > 0.2:
				c = rim
			case lt > 0.35:
				c = litC
			case lt < -0.25:
				c = dark
			}
			if mapmodel.Noise(seed, (dx+1)*2.4, (dy+1)*2.4) > 0.64 {
				c = theme.Mix(c, void, 0.3)
			}
			return c, true
		})
		if (o.anim/3+i*5)%9 < 3 {
			o.fb.fg(x+rr, o.Y(y-rr-1), '•', beacon, dColony)
		} else {
			o.fb.fg(x+rr, o.Y(y-rr-1), '·', theme.Mix(beacon, void, 0.5), dColony)
		}
	}
}

// ------------------------------------------------------------ the warp gate

// gateFrac is how much of the gate's ring is built: it grows with the
// engineering lineage's copies and closes when the Warp Nexus is built.
func gateFrac(m *mapmodel.Model) (float64, bool) {
	for _, w := range m.Wonders {
		if w.Key == "warp_nexus" && w.Built {
			return 1, true
		}
	}
	n := 0
	if l := m.Lineage(mapmodel.LinEngineer); l != nil {
		n = l.Count
	}
	return math.Min(0.92, 0.12+float64(n)/float64(n+40)*0.8), false
}

// gateGeom is the gate's centre column (screen), centre in half rows, and
// radius in half rows.
func (o *orb) gateGeom() (float64, float64, float64) {
	gy := o.groundY
	R := float64(clampInt(gy*84/100, 8, 28))
	cpy := float64(2*gy) - R - 1
	return float64(o.present.Centre-o.cam) + 0.5, cpy, R
}

// gate draws the warp gate over the present district: the built arc of its
// ring plated, the rest scaffold, and inside a closed gate the field's
// teal and violet swirl.
func (o *orb) gate() {
	frac, done := gateFrac(o.m)
	cx, cpy, R := o.gateGeom()
	if cx+2*R < -2 || cx-2*R > float64(o.W)+2 {
		return
	}
	plate := theme.SpaceColor(o.pal[mapmodel.InkFrame])
	lit := o.resolve(plate, iLit, 1)
	shade := o.resolve(theme.Shade(plate, 0.62), iLit, 1)
	scaf := o.c(mapmodel.InkFrameDim, iLit, 1)
	light := o.c(mapmodel.InkAccent2, iEmit, 1)
	void := o.c(mapmodel.InkVoid, iBack, 0)
	teal := o.c(mapmodel.InkAccent2, iEmit, 2)
	vio := o.c(mapmodel.InkCloud, iEmit, 2)
	thick := math.Max(2, R*0.12)
	rIn := R - thick
	x0, x1 := int(cx-R)-1, int(cx+R)+1
	y0, y1 := int((cpy-R)/2)-1, int((cpy+R)/2)+1
	pix := func(x, py int) (tcell.Color, int) {
		dx := float64(x) + 0.5 - cx
		dy := float64(py) + 0.5 - cpy
		d := math.Sqrt(float64(dx*dx) + float64(dy*dy))
		if d > R {
			return 0, 0
		}
		if d >= rIn {
			a := turns(dx, dy) // 0.75 is the bottom
			from := math.Abs(a - 0.75)
			if from > 0.5 {
				from = 1 - from
			}
			if from <= frac/2 {
				if dx > R*0.35 || dy < -R*0.55 {
					return shade, 1
				}
				return lit, 1
			}
			if (x+py)%3 == 0 {
				return scaf, 2
			}
			return 0, 0
		}
		if done && d < rIn-0.5 {
			a := turns(dx, dy)
			v := mapmodel.Sin(float64(a*3) + d/R*1.3 - float64(o.anim)/90)
			return theme.Mix(teal, vio, 0.5+0.5*v), 3
		}
		return 0, 0
	}
	for y := max(0, y0); y <= y1 && y < o.groundY; y++ {
		for x := max(0, x0); x <= x1 && x < o.W; x++ {
			ct, kt := pix(x, 2*y)
			cb, kb := pix(x, 2*y+1)
			if kt == 0 && kb == 0 {
				continue
			}
			if kt == 3 && kb == 3 {
				// the field: a faint swirl, sparked here and there
				bg := theme.Mix(void, ct, 0.3)
				ch, fg := ' ', bg
				dx, dy := float64(x)+0.5-cx, float64(2*y)+1-cpy
				sw := mapmodel.Sin(float64(turns(dx, dy)*4) + math.Sqrt(float64(dx*dx)+float64(dy*dy))/R*2.2 - float64(o.anim)/60)
				switch {
				case hash(x, y, o.anim/4)%19 == 0:
					ch, fg = '·', theme.Mix(ct, o.c(mapmodel.InkStarBright, iEmit, 0), 0.5)
				case sw > 0.55:
					ch, fg = '▒', theme.Mix(void, ct, 0.85)
				case sw > 0.1:
					ch, fg = '░', theme.Mix(void, ct, 0.75)
				}
				o.fb.set(x, o.Y(y), ch, fg, bg, dGate)
				continue
			}
			if kt == 2 || kb == 2 {
				if kt != 1 && kb != 1 {
					o.fb.fg(x, o.Y(y), '·', scaf, dGate)
					continue
				}
			}
			show := o.fb.showAt(x, o.Y(y))
			switch {
			case kt == 1 && kb == 1:
				if ct == cb {
					o.fb.fill(x, o.Y(y), ct, dGate)
				} else {
					o.fb.set(x, o.Y(y), '▀', ct, cb, dGate)
				}
			case kt == 1:
				bg := show
				if kb == 3 {
					bg = theme.Mix(void, cb, 0.22)
				}
				o.fb.set(x, o.Y(y), '▀', ct, bg, dGate)
			case kb == 1:
				bg := show
				if kt == 3 {
					bg = theme.Mix(void, ct, 0.22)
				}
				o.fb.set(x, o.Y(y), '▄', cb, bg, dGate)
			}
		}
	}
	// the gate's lights: eight round the built arc, chasing when it is closed
	for i := 0; i < 8; i++ {
		a := float64(i) / 8
		from := math.Abs(a - 0.75)
		if from > 0.5 {
			from = 1 - from
		}
		if from > frac/2 {
			continue
		}
		rr := R - thick/2
		x := int(cx + rr*mapmodel.Cos(a))
		y := int((cpy - rr*mapmodel.Sin(a)) / 2)
		on := !done || (o.anim/3+i)%8 < 4
		if on && y < o.groundY {
			o.fb.fg(x, o.Y(y), '•', light, dGate)
		}
	}
}

// formation is the faint dotted line the fleet holds formation on.
func (o *orb) formation() {
	gy := o.groundY
	c := o.c(mapmodel.InkFrameDim, iLit, 0)
	mark := o.c(mapmodel.InkAccent2, iEmit, 1)
	for x := 0; x < o.W; x++ {
		wx := x + o.cam
		if wx < o.x0 || wx > o.x1 {
			continue
		}
		switch {
		case wx%12 == 0:
			o.fb.fg(x, o.Y(gy), '+', mark, dFormation)
		case wx%2 == 0:
			o.fb.fg(x, o.Y(gy), '·', c, dFormation)
		}
	}
}
