package mapview

import (
	"image"
	"image/color"
	"math"
	"sync"
	"sync/atomic"

	"github.com/espresso20/ageforge/world"
)

// The Orion arm is drawn from three things, all made by code. A band of faint
// dust, read from one noise texture made per save and dithered with a pattern
// that is pinned to the sky, so a pan slides it rigidly and never shimmers. A
// field of faint stars on world-anchored cells, two octaves of them cross-
// fading by a per-star threshold as the zoom moves. And the named objects,
// each by its form: a cluster is a scatter of bright points, a nebula is a
// noise cloud read from a small texture of its own (made once per object), a
// dark nebula the same cloud drawn as a silhouette.

const (
	dustN    = 256  // texels across the dust texture
	dustHalf = 9500 // display units the dust texture spans either side of Sol
	cloudN   = 128  // texels across a nebula's texture
	bandW    = 2600 // half-width of the arm's band, in display units
)

// A cloud is the texture of one nebula over the square -1..1: how much light
// (d), and where the dark filaments run through it (f). A dark nebula keeps
// its silhouette (s) in a texture of its own over a small window of the cloud,
// so it stays crisp at a fraction of the cloud's size; sb is the silhouette's
// bounding box in that texture's texels.
type cloud struct {
	d, f, s []uint8
	sb      image.Rectangle
}

type cloudKey struct {
	key  string
	form world.Form
}

// artCache holds what the near bands make once: the dust and the clouds.
type artCache struct {
	mu     sync.Mutex
	dust   []uint8
	dustK  int64
	clouds map[cloudKey]*cloud
	builds atomic.Int64
}

func (c *artCache) dustFor(seed int64) []uint8 {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.dust == nil || c.dustK != seed {
		c.dust, c.dustK = buildDust(uint32(seed)), seed
		c.builds.Add(1)
	}
	return c.dust
}

func (c *artCache) cloudFor(o *world.Object) *cloud {
	k := cloudKey{o.Key, o.Form}
	c.mu.Lock()
	defer c.mu.Unlock()
	if t := c.clouds[k]; t != nil {
		return t
	}
	if c.clouds == nil || len(c.clouds) > 64 {
		c.clouds = map[cloudKey]*cloud{}
	}
	t := buildCloud(o.Form, hashKey(o.Key))
	c.clouds[k] = t
	c.builds.Add(1)
	return t
}

// warm makes the near bands' textures, off the draw path.
func (c *artCache) warm(f *world.Frame, seed int64) {
	if f == nil || f.Kind != world.FrameRegion {
		return
	}
	c.dustFor(seed)
	for i := range f.Objects {
		if o := &f.Objects[i]; o.Kind == world.KindFeature && isCloud(o.Form) {
			c.cloudFor(o)
		}
	}
}

func isCloud(f world.Form) bool {
	return f == world.FormEmission || f == world.FormDark || f == world.FormRemnant || f == world.FormKnot
}

func fbm(x, y float64, seed uint32, oct int) float64 {
	sum, amp, tot := 0.0, 1.0, 0.0
	for i := 0; i < oct; i++ {
		sum += Noise(x, y, seed+uint32(i)*131) * amp
		tot += amp
		x, y, amp = x*2.03+17, y*2.03+9, amp*.5
	}
	return sum / tot
}

func clamp01(x float64) float64 { return math.Max(0, math.Min(1, x)) }

func buildDust(seed uint32) []uint8 {
	d := make([]uint8, dustN*dustN)
	for y := 0; y < dustN; y++ {
		for x := 0; x < dustN; x++ {
			n := fbm(float64(x)/22, float64(y)/22, seed, 4)
			d[y*dustN+x] = uint8(255 * clamp01((n-.4)*3.4))
		}
	}
	return d
}

// The Horsehead is two things. The nebula is a glow that hugs one edge (the
// bottom), brightest right on the rim. The silhouette is a head and neck
// rising out of that rim, hsize of the nebula's radius tall and drawn from a
// texture over a window hwin of the nebula's radius either side of (hx, hy).
const (
	hsize = .24  // the silhouette's height, in nebula radii
	hlen  = 2.5  // that height in the shape's own units
	hx    = -.14 // where the shape sits in the nebula, in radii from its centre
	hy    = .58
	hwin  = 1.6 * hsize / hlen // the window's half size, in nebula radii
)

// horse is the silhouette in window units -1..1: how dark. A head taller than
// wide, a neck with a skull, a muzzle and an ear on top, its edges warped by
// noise, the neck running off the bottom to the rim.
func horse(u, v float64, seed uint32) float64 {
	u, v = u*1.6, v*1.6+.25 // the shape's own units
	wx := u + (fbm(u*3+11, v*3, seed, 3)-.5)*.2
	wy := v + (fbm(u*3, v*3+23, seed+5, 3)-.5)*.2 - .1
	capsule := func(ax, ay, bx, by, r float64) float64 {
		px, py, dx, dy := wx-ax, wy-ay, bx-ax, by-ay
		t := clamp01((px*dx + py*dy) / (dx*dx + dy*dy))
		return math.Hypot(px-t*dx, py-t*dy) - r
	}
	sdf := capsule(.34, 1.2, .08, .0, .3)                                                                                    // the neck
	for _, s := range [...][5]float64{{.05, -.4, -.28, -.2, .27}, {-.3, -.2, -.52, .12, .17}, {.14, -.66, .22, -.92, .07}} { // skull, muzzle, ear
		a, b := sdf, capsule(s[0], s[1], s[2], s[3], s[4])
		h := clamp01(.5 + .5*(b-a)/.12) // a smooth union: the joins are not creases
		sdf = b*(1-h) + a*h - .12*h*(1-h)
	}
	return clamp01(.5 - sdf/.07)
}

// glow is the Horsehead's bright cloud at u, v in nebula radii: a band along
// the bottom rim, broken by noise, fading up into the dark.
func glow(u, v float64, seed uint32) float64 {
	r2 := u*u + v*v
	rim := math.Exp(-math.Pow((math.Sqrt(r2)-.8)/.22, 2)) * clamp01((v+.25)/.9) // the lit edge, strongest at the bottom
	body := .35 * clamp01((v+.5)/1.2) * (1 - r2)
	return clamp01((rim + body) * (.7 + .6*fbm(u*6+3, v*6, seed+9, 3)))
}

func buildCloud(form world.Form, seed uint32) *cloud {
	c := &cloud{d: make([]uint8, cloudN*cloudN), f: make([]uint8, cloudN*cloudN)}
	for y := 0; y < cloudN; y++ {
		for x := 0; x < cloudN; x++ {
			u, v := (float64(x)+.5)/cloudN*2-1, (float64(y)+.5)/cloudN*2-1
			r := math.Hypot(u, v)
			n := fbm(u*3+5, v*3+5, seed, 5)
			ridge := 1 - clamp01(math.Abs(fbm(u*5+31, v*5+31, seed+7, 3)-.5)*16) // thin where the noise crosses its middle
			var d, f float64
			switch form {
			case world.FormDark:
				f = glow(u, v, seed)
			case world.FormRemnant:
				shell := math.Exp(-math.Pow((r-.6)/.2, 2))
				d = clamp01(shell*(.35+1.1*ridge+.4*n) + .6*math.Exp(-r*r*90))
			default: // emission, knot
				d = clamp01((1 - r*r) * (n*1.5 - .12))
				f = ridge
			}
			c.d[y*cloudN+x], c.f[y*cloudN+x] = uint8(255*d), uint8(255*f)
		}
	}
	if form == world.FormDark {
		c.s = make([]uint8, cloudN*cloudN)
		c.sb = image.Rect(cloudN, cloudN, 0, 0)
		for y := 0; y < cloudN; y++ {
			for x := 0; x < cloudN; x++ {
				d := horse((float64(x)+.5)/cloudN*2-1, (float64(y)+.5)/cloudN*2-1, seed)
				c.s[y*cloudN+x] = uint8(255 * d)
				if d > .5 {
					c.sb = c.sb.Union(image.Rect(x, y, x+1, y+1))
				}
			}
		}
	}
	return c
}

// silhouette is the dark nebula's silhouette read bilinearly at u, v in the
// window's -1..1.
func (c *cloud) silhouette(u, v float64) float64 { return sample(c.s, u, v) }

// at is the texture read bilinearly at u, v in -1..1.
func (c *cloud) at(u, v float64) (d, f float64) { return sample(c.d, u, v), sample(c.f, u, v) }

// sample reads one texture bilinearly at u, v in -1..1.
func sample(a []uint8, u, v float64) float64 {
	x, y := (u+1)*.5*cloudN-.5, (v+1)*.5*cloudN-.5
	x0, y0 := int(math.Floor(x)), int(math.Floor(y))
	fx, fy := x-float64(x0), y-float64(y0)
	pick := func(xx, yy int) float64 {
		return float64(a[min(cloudN-1, max(0, yy))*cloudN+min(cloudN-1, max(0, xx))])
	}
	t := pick(x0, y0)*(1-fx) + pick(x0+1, y0)*fx
	b := pick(x0, y0+1)*(1-fx) + pick(x0+1, y0+1)*fx
	return (t + (b-t)*fy) / 255
}

// worldAt is the display position under canvas pixel x, y.
func (p *pen) worldAt(x, y float64) (float64, float64) {
	c := regionPos(p.f, p.c)
	return c.X + (x-p.cx)/p.ppu, c.Y + (y-p.cy)/p.ppu
}

// pinned is the offset that makes a Bayer pattern hold still on the sky.
func (p *pen) pinned() (int, int) {
	c := regionPos(p.f, p.c)
	return int(math.Round(c.X * p.ppu)), int(math.Round(c.Y * p.ppu))
}

// armBackdrop is the band of dust and the field of faint stars under the arm.
func (p *pen) armBackdrop(o *world.Object, art *artCache, seed int64) {
	dust := art.dustFor(seed)
	ang := math.Atan2(o.To.Y, o.To.X)
	dx, dy := math.Cos(ang), math.Sin(ang)
	ox, oy := p.pinned()
	glow := [...]color.RGBA{p.pal.Mix(world.ToneArm, 12), p.pal.Mix(world.ToneArm, 24), p.pal.Mix(world.ToneArm, 40), p.pal.Mix(world.ToneArm, 62)}
	fade := clamp01((.14 - p.ppu) * 14) // close in, the stars carry the field
	const k = float64(dustN) / (2 * dustHalf)
	cc, inv := regionPos(p.f, p.c), 1/p.ppu
	for yy := 0; yy < p.h && fade > 0; yy++ {
		wy := cc.Y + (float64(yy)-p.cy)*inv
		for xx := 0; xx < p.w; xx++ {
			wx := cc.X + (float64(xx)-p.cx)*inv
			u, v := wx*dx+wy*dy, (-wx*dy+wy*dx)/bandW
			if v*v >= 1 || math.Abs(u) >= 10000 {
				continue
			}
			b := (1 - v*v) * (1 - v*v)
			fx, fy := (wx+dustHalf)*k-.5, (wy+dustHalf)*k-.5
			tx, ty := int(math.Floor(fx)), int(math.Floor(fy))
			if tx < 0 || ty < 0 || tx >= dustN-1 || ty >= dustN-1 {
				continue
			}
			fx, fy = fx-float64(tx), fy-float64(ty)
			at := func(dx, dy int) float64 { return float64(dust[(ty+dy)*dustN+tx+dx]) }
			t0, t1 := at(0, 0)+(at(1, 0)-at(0, 0))*fx, at(0, 1)+(at(1, 1)-at(0, 1))*fx
			t := (t0 + (t1-t0)*fy) / 255
			lvl := (.06 + 1.5*t*t) * b * 2.6 * fade * (1 - (u*u/1e8)*(u*u/1e8))
			n := int(lvl + Bayer(xx+ox, yy+oy) - .5)
			if n > 0 {
				p.px(xx, yy, glow[min(3, n-1)])
			}
		}
	}
	p.armStars(dx, dy, seed)
}

var starTones = [...]world.Tone{world.ToneWhite, world.ToneCool, world.ToneBlue, world.ToneSun, world.ToneAmber, world.ToneStar}

// armStars scatters faint stars over cells pinned to the sky. Two octaves of
// cell size share the zoom: one thins out as the other grows in, each star
// by its own threshold, so none pops.
func (p *pen) armStars(dx, dy float64, seed int64) {
	kf := math.Log2(7 / p.ppu)
	k0 := math.Floor(kf)
	frac := kf - k0
	cc := regionPos(p.f, p.c)
	cx, cy := cc.X, cc.Y
	x0, y0 := p.worldAt(0, 0)
	x1, y1 := p.worldAt(float64(p.w), float64(p.h))
	for oct, vis := range [2]float64{1 - frac, frac} {
		s := math.Exp2(k0 + float64(oct))
		sd := uint32(seed) + uint32(k0)*977 + uint32(oct)*31
		for gy := int(math.Floor(y0 / s)); gy <= int(math.Floor(y1/s)); gy++ {
			for gx := int(math.Floor(x0 / s)); gx <= int(math.Floor(x1/s)); gx++ {
				h := hash2(gx, gy, sd)
				wx, wy := (float64(gx)+float64(h&0xff)/256)*s, (float64(gy)+float64(h>>8&0xff)/256)*s
				v := (-wx*dy + wy*dx) / bandW
				dens := .22 + .5*math.Max(0, 1-v*v)
				if float64(h>>16&0xff)/256 > vis*dens*.55 {
					continue
				}
				bright := 70 + int(h>>24)*110/256
				if h>>24 > 230 {
					bright = 235
				}
				c := p.pal.Mix(starTones[int(h>>5)%len(starTones)], uint8(bright))
				p.px(int(math.Round(p.cx+(wx-cx)*p.ppu)), int(math.Round(p.cy+(wy-cy)*p.ppu)), c)
			}
		}
	}
}

// cluster is a scatter of bright points: mostly single pixels, a few brighter
// with a glint, in the cluster's colour and white and a cooler blue.
func (p *pen) cluster(o *world.Object, x, y, r float64) {
	n := min(90, max(10, int(r*2.6)))
	seed := hashKey(o.Key)
	warm := o.Tone == world.ToneAmber // the Hyades: old, orange giants among the members
	// thin the background field under the cluster so the members stand out
	if x+r > 0 && y+r > 0 && x-r < float64(p.w) && y-r < float64(p.h) && r < 400 {
		ox, oy := p.pinned()
		for yy := max(0, int(y-r*1.2)); yy <= min(p.h-1, int(y+r*1.2)); yy++ {
			for xx := max(0, int(x-r*1.2)); xx <= min(p.w-1, int(x+r*1.2)); xx++ {
				dx, dy := float64(xx)-x, float64(yy)-y
				if d := math.Sqrt(dx*dx+dy*dy) / (r * 1.2); d < 1 && 1.1*(1-d) > Bayer(xx+ox, yy+oy) {
					i := p.img.PixOffset(xx, yy)
					p.img.Pix[i], p.img.Pix[i+1], p.img.Pix[i+2] = p.img.Pix[i]/3, p.img.Pix[i+1]/3, p.img.Pix[i+2]/3
				}
			}
		}
	}
	type mem struct{ x, y float64 }
	var bright []mem
	for i := 0; i < n; i++ {
		h := hash2(i, 7, seed)
		a := float64(h&0xffff) / 65536 * 2 * math.Pi
		d := math.Min(r*math.Sqrt(-math.Log(1-.97*float64(h>>16)/65536))*.7, r)
		sx, sy := x+d*math.Cos(a), y+d*math.Sin(a)
		t := world.ToneCool // hot blue-white
		if warm && i%5 == 0 {
			t = world.ToneAmber
		} else if warm {
			t = world.ToneWhite
		} else if i%4 == 0 {
			t = world.ToneWhite
		}
		switch {
		case i < 7 && r >= 5: // the brightest: 2-3 px with a glint
			big := 1.0 + float64(hash2(i, 9, seed)%3)*.4
			p.disc(sx, sy, big, t, 255)
			for _, g := range [4][2]float64{{-3, 0}, {3, 0}, {0, -3}, {0, 3}} {
				p.px(int(math.Round(sx+g[0])), int(math.Round(sy+g[1])), p.pal.Mix(t, 170))
			}
			bright = append(bright, mem{sx, sy})
		case i%3 == 0:
			p.px(int(math.Round(sx)), int(math.Round(sy)), p.pal.Mix(t, 255))
		default:
			p.px(int(math.Round(sx)), int(math.Round(sy)), p.pal.Mix(t, uint8(150+hash2(i, 11, seed)>>8&0x4f)))
		}
	}
	if !warm { // the Pleiades' reflection haze round its brightest members
		for _, m := range bright {
			p.glow(m.x, m.y, math.Max(5, r*.35), world.ToneBlue, 90)
		}
	}
}

var (
	emissionTones = [...]struct {
		t world.Tone
		w uint8
	}{{world.ToneRed, 60}, {world.ToneRed, 130}, {world.TonePink, 150}, {world.TonePink, 215}, {world.TonePink, 255}, {world.ToneWhite, 210}}
	remnantTones = [...]struct {
		t world.Tone
		w uint8
	}{{world.ToneBlue, 60}, {world.ToneBlue, 130}, {world.ToneCyan, 150}, {world.TonePink, 190}, {world.ToneCyan, 255}, {world.ToneWhite, 230}}
)

// nebula draws a cloud form into the square r art pixels either side of x, y.
func (p *pen) nebula(o *world.Object, art *artCache, x, y, r float64) {
	if x+r < 0 || y+r < 0 || x-r > float64(p.w) || y-r > float64(p.h) || r > 1e6 {
		return
	}
	cl := art.cloudFor(o)
	ox, oy := p.pinned()
	tones := &emissionTones
	if o.Form == world.FormRemnant {
		tones = &remnantTones
	}
	var ramp [6]color.RGBA
	scale := int(o.Weight)
	for i, t := range tones {
		ramp[i] = p.pal.Mix(t.t, uint8(int(t.w)*scale/255))
	}
	dark := p.pal.Mix(world.ToneNone, 0)
	dark = color.RGBA{R: dark.R / 2, G: dark.G / 2, B: dark.B / 2, A: 255} // below the void: the silhouette
	for yy := max(0, int(y-r)); yy <= min(p.h-1, int(y+r)); yy++ {
		for xx := max(0, int(x-r)); xx <= min(p.w-1, int(x+r)); xx++ {
			u, v := (float64(xx)-x)/r, (float64(yy)-y)/r
			if u*u+v*v > 1 {
				continue
			}
			d, f := cl.at(u, v)
			th := Bayer(xx+ox, yy+oy)
			if o.Form == world.FormDark {
				var sil float64
				if wu, wv := (u-hx)/hwin, (v-hy)/hwin; math.Abs(wu) <= 1 && math.Abs(wv) <= 1 {
					sil = cl.silhouette(wu, wv)
				}
				switch i := int(f*5.2 + th - .55); {
				case sil > .3+.6*th: // the silhouette takes the stars and the dust with it
					p.px(xx, yy, dark)
				case i > 0: // the bright cloud behind it
					p.px(xx, yy, ramp[min(5, i)])
				}
				continue
			}
			d *= 1 - .6*f*f // dark filaments cut through
			if i := int(d*5.2 + th - .55); i > 0 {
				p.px(xx, yy, ramp[min(5, i)])
			}
		}
	}
}

// region draws one of the arm's features by its form.
func (p *pen) region(o *world.Object, x, y, r float64, art *artCache, seed int64) {
	switch o.Form {
	case world.FormBand:
		p.armBackdrop(o, art, seed)
	case world.FormCluster:
		p.cluster(o, x, y, math.Max(r, 4))
	case world.FormDark:
		p.nebula(o, art, x, y, math.Max(r, 3))
	case world.FormEmission, world.FormRemnant:
		p.nebula(o, art, x, y, math.Max(r, 3))
	case world.FormKnot:
		p.nebula(o, art, x, y, math.Max(r, 2))
	case world.FormTerritory:
		p.ring(0, math.Sqrt(o.Size*regionSqueeze), o.Tone, 110, false)
	}
}
