package mapview

import (
	"image/color"
	"math"
	"sync"
	"sync/atomic"

	"github.com/espresso20/ageforge/world"
)

// A galaxy is drawn from a face-on texture made once per look and seed: four
// channels (light, bulge, dust, pink knots) over the disc, which spans -1..1
// on both axes. A frame only maps screen pixels back into it (undoing the
// tilt and squash), reads it with mip levels, and turns the light into steps of
// a palette ramp with an ordered dither that is anchored to the galaxy, so
// panning never shimmers. The texture is not a picture of a palette: colour
// comes at draw time, so a theme change does not rebuild it.

const (
	chL = iota // arm, disc and bar light
	chC        // bulge: where the warm colour takes over
	chD        // dust
	chK        // star-forming knots
	nCh
)

type galTex struct {
	t    int
	mips [][]uint8 // level l is (t>>l) square, nCh bytes a texel
	seed uint32
}

type texKey struct {
	shape  world.GalaxyShape
	look   world.GalaxyLook
	seed   int64
	shaped bool
	t      int
}

// texCache holds the textures; builds counts how many it has made.
type texCache struct {
	mu     sync.Mutex
	m      map[texKey]*galTex
	builds atomic.Int64
}

// A geom maps screen offsets to the face-on disc: the home galaxy's rows are
// drawn 0.8 as tall as its columns, on top of its tilt and its flattening.
type geom struct{ cos, sin, q, ys float64 }

func geomOf(o *world.Object) geom {
	switch {
	case o.Shape != nil:
		a := float64(o.Shape.Tilt) * 2 * math.Pi / world.Turn
		return geom{math.Cos(a), math.Sin(a), float64(max(o.Shape.Flat, 8)) / 100, 0.8}
	case o.Galaxy != nil:
		a := float64(o.Galaxy.Phi) * 2 * math.Pi / world.Turn
		return geom{math.Cos(a), math.Sin(a), float64(max(o.Galaxy.Q, 16)) / 256, 1}
	}
	return geom{1, 0, 1, 1}
}

// face is the disc point (-1..1) under screen offset dx, dy from the centre of
// a galaxy rpx art pixels across its radius.
func (g geom) face(dx, dy, rpx float64) (u, v float64) {
	fx, fy := dx/rpx, dy/(rpx*g.ys)
	return fx*g.cos + fy*g.sin, (-fx*g.sin + fy*g.cos) / g.q
}

func texKeyOf(o *world.Object) (texKey, bool) {
	switch {
	case o.Shape != nil:
		return texKey{shape: *o.Shape, seed: o.Seed, shaped: true, t: 512}, true
	case o.Galaxy != nil:
		l := *o.Galaxy
		l.R, l.Q, l.Phi = 0, 0, 0
		t := 96
		if o.Size >= 100 {
			t = 256
		}
		return texKey{look: l, seed: int64(hashKey(o.Key)), t: t}, true
	}
	return texKey{}, false
}

func hashKey(s string) uint32 {
	h := uint32(2166136261)
	for i := 0; i < len(s); i++ {
		h = (h ^ uint32(s[i])) * 16777619
	}
	return h
}

// get is the texture of o, made on first ask.
func (c *texCache) get(o *world.Object) *galTex {
	k, ok := texKeyOf(o)
	if !ok {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if t := c.m[k]; t != nil {
		return t
	}
	if c.m == nil || len(c.m) > 48 { // a new save's sky replaces the old
		c.m = map[texKey]*galTex{}
	}
	t := buildTex(k)
	c.m[k] = t
	c.builds.Add(1)
	return t
}

// warm makes the textures a sky will need, off the draw path.
func (c *texCache) warm(f *world.Frame) {
	for i := range f.Objects {
		if o := &f.Objects[i]; o.Kind == world.KindGalaxy {
			c.get(o)
		}
	}
	for _, ch := range f.Children {
		c.warm(ch)
	}
}

// ---- the model: arms as a table of angle by radius ----

const armBins = 256

type model struct {
	ang      [][armBins]float64 // arm k's angle at radius (i+.5)/armBins
	w        []float64          // arm weight: faint arms are fainter
	spin     float64
	core     float64
	bar      float64 // half length of the bar, 0 for none
	barAng   float64
	rIn      float64
	dust     float64
	irr, ell bool
	seed     uint32
}

func (m *model) addArm(w float64, f func(u float64) float64) {
	var a [armBins]float64
	for i := range a {
		a[i] = f((float64(i) + .5) / armBins)
	}
	m.ang, m.w = append(m.ang, a), append(m.w, w)
}

func modelOf(k texKey) *model {
	m := &model{seed: uint32(k.seed) ^ uint32(k.seed>>32)*0x9E3779B1, dust: .7}
	const rad = 2 * math.Pi / world.Turn
	if k.shaped {
		s := k.shape
		m.core, m.bar, m.barAng, m.spin = float64(s.Core)/world.GalRX, float64(s.Bar)/world.GalRX, float64(s.Rot)*rad, float64(s.Spin)
		rIn := float64(max(s.Core*3/2, s.Bar))
		m.rIn = rIn / world.GalRX
		for a := 0; a < s.Arms+s.Faint; a++ {
			w := 1.0
			if a >= s.Arms {
				w = .4
			}
			m.addArm(w, func(u float64) float64 {
				rf := int(math.Max(0, (u*world.GalRX-rIn)/(world.GalRX-rIn)) * 1000)
				return float64(s.ArmAngle(k.seed, a, rf)) * rad
			})
		}
		return m
	}
	l := k.look
	m.core, m.bar, m.dust, m.spin, m.irr = .05+float64(l.Core)/256*.2, float64(l.Bar)/256*.5, float64(l.Dust)/256, 1, l.Irr
	if l.Irr {
		m.barAng = .5
		return m
	}
	if l.Arms == 0 {
		m.ell = true
		return m
	}
	lo := 30.0
	if l.Bar > 0 {
		lo = 97
	}
	m.rIn = lo / 256
	for a := 0; a < l.Arms; a++ {
		off := float64(a) * 2 * math.Pi / float64(l.Arms)
		if l.Bar > 0 {
			off += 2 * math.Pi / 7
		}
		m.addArm(1, func(u float64) float64 {
			t := 255.99 * math.Sqrt(math.Max(0, (u*256-lo)/(256-lo)))
			return off + t*float64(l.Spin)*2*math.Pi/65536
		})
	}
	return m
}

func smooth(a, b, x float64) float64 {
	t := math.Max(0, math.Min(1, (x-a)/(b-a)))
	return t * t * (3 - 2*t)
}

func wrap(a float64) float64 {
	a = math.Mod(a+math.Pi, 2*math.Pi)
	if a < 0 {
		a += 2 * math.Pi
	}
	return a - math.Pi
}

// texel is the four channels (0..1) at disc point u, v.
func (m *model) texel(u, v float64) (l, c, d, k float64) {
	r := math.Hypot(u, v)
	if r >= 1 {
		return
	}
	edge := 1 - smooth(.7, 1, r)
	switch {
	case m.irr:
		return m.irregular(u, v, r)
	case m.ell:
		n := Noise(u*6, v*6, m.seed)
		l = math.Exp(-3.4*math.Pow(r, .8)) * (.8 + .4*n)
		return math.Min(1, l*edge), math.Min(1, math.Exp(-6*r)*1.4), 0, 0
	}
	bulge := (.6*math.Exp(-r/(m.core*.5)) + .5*math.Exp(-r/(m.core*1.7))) * (.8 + .4*Noise(u*45, v*45, m.seed^3))
	disc := .17 * math.Exp(-2.4*r) * (.7 + .6*Noise(u*5+9, v*5, m.seed))
	var arm, dust, knot float64
	th := math.Atan2(v, u)
	bin := min(armBins-1, int(r*armBins))
	sig := .03 + .05*r
	var wide float64
	fin := smooth(m.rIn*.7, m.rIn*1.15, r)
	for i := range m.ang {
		gap := wrap(th - m.ang[i][bin])
		d := gap * r
		if math.Abs(d) > 10*sig {
			continue
		}
		a := math.Exp(-.5 * (d / sig) * (d / sig))
		wide += m.w[i] * math.Exp(-.5*(d/(3.2*sig))*(d/(3.2*sig))) * fin
		mod := .4 + 1.2*Noise(m.ang[i][bin]*4, r*16, m.seed+uint32(i))
		arm += m.w[i] * a * mod * fin
		in := (-gap*m.spin*r - 1.3*sig) / (.55 * sig) // the inner edge of the arm
		dust += m.w[i] * math.Exp(-.5*in*in) * (.5 + .9*Noise(m.ang[i][bin]*7, r*30, m.seed^0x55)) * fin
		if a > .45 && r > .16 && m.w[i] == 1 {
			if n := Noise(u*64+float64(i)*11, v*64, m.seed^0x77); n > .78 {
				knot = math.Max(knot, (n-.78)/.22*a)
			}
		}
	}
	var bar float64
	if m.bar > 0 {
		s, cs := math.Sincos(m.barAng)
		along, across := u*cs+v*s, -u*s+v*cs
		bar = math.Exp(-.5*math.Pow(across/.05, 2)) * smooth(m.bar*1.1, m.bar*.7, math.Abs(along)) * .8
	}
	l = (bulge + disc + .8*arm + .3*wide*(.5+Noise(u*7, v*7, m.seed^9)) + .7*bar) * edge
	l = 1 - math.Exp(-1.6*l)
	l *= 1 - .8*math.Min(1, dust*m.dust)*(1-.8*bulge)
	return math.Min(1, l), warm(r, bulge), math.Min(1, dust), math.Min(1, knot*edge)
}

// warm is the bulge's colour channel: it has compact support, exactly zero
// from 30% of the radius out, so no dither step can put warm pixels in the arms.
func warm(r, bulge float64) float64 {
	c := math.Pow(1-smooth(0, .4, r), 2.5) * math.Min(1, bulge*1.25)
	return math.Max(0, c-.1) / .9
}

// irregular is a dwarf: warped, clumpy, with a few knots and maybe a bar.
func (m *model) irregular(u, v, r float64) (l, c, d, k float64) {
	wu := u + .3*(Noise(u*2.5, v*2.5, m.seed)-.5)
	wv := v + .3*(Noise(u*2.5+7, v*2.5+3, m.seed)-.5)
	env := 1 - smooth(.3, 1, math.Hypot(wu*.9, wv))
	n := .5*Noise(u*4, v*4, m.seed+1) + .3*Noise(u*9, v*9, m.seed+2) + .2*Noise(u*19, v*19, m.seed+3)
	l = .9 * smooth(.28, 1.05, n*1.05+env*.5-.22) * env * (.45 + .8*Noise(u*6+3, v*6, m.seed+5))
	if m.bar > 0 {
		s, cs := math.Sincos(m.barAng)
		along, across := u*cs+v*s, -u*s+v*cs
		l += .55 * math.Exp(-.5*math.Pow(across/.1, 2)) * smooth(.55, .3, math.Abs(along)) * (.6 + .6*n)
	}
	if kn := Noise(u*30, v*30, m.seed+4); kn > .8 && env > .4 && n > .5 {
		k = (kn - .8) / .2
	}
	return math.Min(1, l), 0, 0, k
}

func buildTex(k texKey) *galTex {
	m := modelOf(k)
	t := k.t
	base := make([]uint8, t*t*nCh)
	for y := 0; y < t; y++ {
		v := (float64(y)+.5)/float64(t)*2 - 1
		for x := 0; x < t; x++ {
			u := (float64(x)+.5)/float64(t)*2 - 1
			l, c, d, kn := m.texel(u, v)
			i := (y*t + x) * nCh
			base[i], base[i+1], base[i+2], base[i+3] = uint8(l*255+.5), uint8(c*255+.5), uint8(d*255+.5), uint8(kn*255+.5)
		}
	}
	g := &galTex{t: t, mips: [][]uint8{base}, seed: m.seed}
	for s := t / 2; s >= 16; s /= 2 { // box-filtered levels, for when the galaxy draws small
		prev, p := g.mips[len(g.mips)-1], s*2
		next := make([]uint8, s*s*nCh)
		for y := 0; y < s; y++ {
			for x := 0; x < s; x++ {
				for ch := 0; ch < nCh; ch++ {
					sum := int(prev[((2*y)*p+2*x)*nCh+ch]) + int(prev[((2*y)*p+2*x+1)*nCh+ch]) +
						int(prev[((2*y+1)*p+2*x)*nCh+ch]) + int(prev[((2*y+1)*p+2*x+1)*nCh+ch])
					next[(y*s+x)*nCh+ch] = uint8((sum + 2) / 4)
				}
			}
		}
		g.mips = append(g.mips, next)
	}
	return g
}

// at is channel ch of mip level lvl at disc point u, v, read bilinearly.
func (g *galTex) at(lvl int, u, v float64) (out [nCh]float64) {
	s := g.t >> lvl
	fx, fy := (u+1)/2*float64(s)-.5, (v+1)/2*float64(s)-.5
	x0, y0 := int(math.Floor(fx)), int(math.Floor(fy))
	tx, ty := fx-float64(x0), fy-float64(y0)
	m := g.mips[lvl]
	px := func(x, y, ch int) float64 {
		if x < 0 || y < 0 || x >= s || y >= s {
			return 0
		}
		return float64(m[(y*s+x)*nCh+ch])
	}
	for ch := range out {
		a := px(x0, y0, ch)*(1-tx) + px(x0+1, y0, ch)*tx
		b := px(x0, y0+1, ch)*(1-tx) + px(x0+1, y0+1, ch)*tx
		out[ch] = (a + (b-a)*ty) / 255
	}
	return out
}

// level is the mip to read when a texel covers tpp art pixels.
func (g *galTex) level(rpx float64) int {
	tpp, lvl := float64(g.t)/(2*rpx), 0
	for tpp >= 1.5 && lvl < len(g.mips)-1 {
		tpp, lvl = tpp/2, lvl+1
	}
	return lvl
}

const rampN = 6

// coreRamp is the bulge's four steps, orange through yellow to a pale
// yellow-white at the centre.
func coreRamp(pal *world.Palette) []color.RGBA {
	mix := func(a, b color.RGBA, t int) color.RGBA {
		f := func(x, y uint8) uint8 { return uint8((int(x)*(100-t) + int(y)*t) / 100) }
		return color.RGBA{f(a.R, b.R), f(a.G, b.G), f(a.B, b.B), 255}
	}
	or, ye, wh := pal.Mix(world.ToneCore, 255), pal.Mix(world.ToneSun, 255), pal.Mix(world.ToneWhite, 255)
	return []color.RGBA{or, ye, mix(ye, wh, 60), mix(ye, wh, 90)}
}

// galaxy draws o, rpx art pixels across its radius, centred at x, y.
func (p *pen) galaxy(o *world.Object, x, y, rpx float64) {
	g := p.tex.get(o)
	if rpx < 1.5 || g == nil {
		p.disc(x, y, 1, o.Tone, o.Weight)
		return
	}
	gm := geomOf(o)
	hx, hy := rpx*math.Hypot(gm.cos, gm.sin*gm.q), rpx*gm.ys*math.Hypot(gm.sin, gm.cos*gm.q)
	if x+hx < 0 || y+hy < 0 || x-hx > float64(p.w) || y-hy > float64(p.h) {
		return
	}
	xi, yi := int(math.Round(x)), int(math.Round(y))
	body, knot := Ramp(p.pal, o.Tone, rampN), Ramp(p.pal, world.TonePink, rampN)
	core := coreRamp(p.pal)
	star := p.pal.Mix(world.ToneWhite, 200)
	wgt, lvl := float64(o.Weight)/255, g.level(rpx)
	speck := rpx > 50
	for yy := max(0, yi-int(hy)-1); yy <= min(p.h-1, yi+int(hy)+1); yy++ {
		for xx := max(0, xi-int(hx)-1); xx <= min(p.w-1, xi+int(hx)+1); xx++ {
			wx, wy := xx-xi, yy-yi // dither and stars follow the galaxy, not the screen
			u, v := gm.face(float64(wx), float64(wy), rpx)
			if u*u+v*v >= 1 {
				continue
			}
			ch := g.at(lvl, u, v)
			b := Bayer(wx, wy)
			var col = body
			i := int(ch[chL]*wgt*(rampN-1) + b)
			j := int(ch[chC]*wgt*float64(len(core)) + b) // the bulge: a dithered step ramp, no edge
			if j > 0 && i > 0 {
				p.px(xx, yy, core[min(j, len(core))-1])
				continue
			}
			if k := int(ch[chK]*wgt*3 + b); k > 0 && ch[chL] > .08 {
				col, i = knot, min(rampN-1, k+2)
			} else if speck && ch[chL] > .05 && float64(hash2(wx, wy, g.seed)&0xffff)/65536 < .012*ch[chL] {
				p.px(xx, yy, star)
				continue
			}
			if i > 0 {
				p.px(xx, yy, col[min(i, rampN-1)])
			}
		}
	}
}
