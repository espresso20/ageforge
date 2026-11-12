package mapview

import (
	"image"
	"image/color"
	"math"
	"sync"
	"sync/atomic"

	"github.com/espresso20/ageforge/world"
)

// The System band is an orrery: dim dithered orbits, belts of scattered rocks,
// planets as lit pixel spheres, moons as dots. A planet's sprite depends on
// what it is, how big it draws and which way its sun is, never on where the
// camera is, so panning and easing the zoom reuse the same few sprites.

const lightBuckets = 16

// spriteSteps are the radii (art pixels) a planet is drawn at: a zoom between
// two of them snaps to the lower, so the cache holds a handful of sprites each.
var spriteSteps = [...]int{1, 2, 3, 4, 5, 6, 7, 9, 11, 14, 17}

func stepR(r float64) int {
	s := spriteSteps[0]
	for _, k := range spriteSteps {
		if float64(k) <= r+.5 {
			s = k
		}
	}
	return s
}

type bodyKey struct {
	key    string
	r, dir int // radius in art pixels; the light's direction, in 16ths of a turn
}

// A sprite is a body on a clear square, centred at (ext, ext).
type sprite struct {
	img *image.RGBA
	ext int
}

// bodyCache holds the planet sprites, made on first ask and kept until the
// palette changes.
type bodyCache struct {
	mu     sync.Mutex
	pal    world.Palette
	m      map[bodyKey]*sprite
	builds atomic.Int64
}

func (c *bodyCache) get(pal *world.Palette, o *world.Object, r, dir int) *sprite {
	k := bodyKey{o.Key, r, dir}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.m == nil || c.pal != *pal {
		c.m, c.pal = map[bodyKey]*sprite{}, *pal
	}
	if s := c.m[k]; s != nil {
		return s
	}
	s := buildBody(pal, o, r, dir)
	c.m[k] = s
	c.builds.Add(1)
	return s
}

// lift raises c toward white until it is as bright as the palette's mid tone,
// so a dark colour (Neptune's blue) still shows against the void.
func lift(pal *world.Palette, c color.RGBA) color.RGBA {
	mid := pal.Mix(world.ToneWhite, 128)
	white := pal.Tones[world.ToneWhite]
	for k := 0; luma(c) < luma(mid) && k < 255; k += 15 {
		c = mixRGB(c, white, k)
	}
	return c
}

func luma(c color.RGBA) int { return 3*int(c.R) + 6*int(c.G) + int(c.B) }

// mixRGB is a toward b by k of 255.
func mixRGB(a, b color.RGBA, k int) color.RGBA {
	m := func(x, y uint8) uint8 { return uint8((int(x)*(255-k) + int(y)*k + 127) / 255) }
	return color.RGBA{R: m(a.R, b.R), G: m(a.G, b.G), B: m(a.B, b.B), A: 255}
}

func scale(c color.RGBA, f float64) color.RGBA {
	m := func(x uint8) uint8 { return uint8(math.Max(0, math.Min(255, float64(x)*f))) }
	return color.RGBA{R: m(c.R), G: m(c.G), B: m(c.B), A: 255}
}

// buildBody is one planet's sprite: a sphere lit from direction dir (16ths of
// a turn, screen angle toward the star), in four steps and an ordered dither
// at the terminator, with a night side that stays above the void. The face is
// noise: bands on the gas giants and Venus, a ringed Saturn, seas and land and
// cloud on the home world, mottling on the rest.
func buildBody(pal *world.Palette, o *world.Object, r, dir int) *sprite {
	b := o.Body
	sp := b.Spec
	ringed := b.Rings && r >= 2
	ext := r + 1
	if ringed {
		ext = int(math.Ceil(float64(r)*2.3)) + 1
	}
	n := 2*ext + 1
	img := image.NewRGBA(image.Rect(0, 0, n, n))
	rf := float64(r) + .5
	sd := hashKey(o.Key)
	ang := float64(dir) * 2 * math.Pi / lightBuckets
	lx, ly, lz := math.Cos(ang)*.85, math.Sin(ang)*.85, .5

	base := lift(pal, pal.Of(sp.Base))
	if sp.Base.W == 0 {
		base = lift(pal, pal.Mix(o.Tone, 255))
	}
	var bands []color.RGBA
	for _, s := range sp.Bands {
		bands = append(bands, lift(pal, pal.Of(s)))
	}
	ocean, land := lift(pal, pal.Mix(world.ToneBlue, 255)), lift(pal, pal.Mix(world.ToneTeal, 255))
	white, red := lift(pal, pal.Mix(world.ToneWhite, 255)), lift(pal, pal.Mix(world.ToneRed, 255))
	blotch := lift(pal, pal.Of(sp.Blotch))
	ringC := lift(pal, pal.Mix(world.ToneSand, 255))
	const tilt, flat = -.32, .3 // the ring plane's turn, and how edge-on it is
	ct, st := math.Cos(tilt), math.Sin(tilt)
	steps := [...]int{56, 108, 158, 208, 255}

	for py := -ext; py <= ext; py++ {
		for px := -ext; px <= ext; px++ {
			var col color.RGBA
			have, front := false, false
			if ringed { // the ring: an ellipse in the ring plane, a gap in it
				u, v := float64(px)*ct+float64(py)*st, (-float64(px)*st+float64(py)*ct)/flat
				if e := math.Hypot(u, v) / (float64(r) * 2.25); e >= .62 && e <= 1 && !(e > .8 && e < .85) {
					s := fbm(e*30, 3, sd, 2)
					col, have, front = scale(ringC, .5+.6*s), true, v > 0
					col = mixRGB(pal.Bg, col, 190)
				}
			}
			nx, ny := float64(px)/rf, float64(py)/rf
			if d2 := nx*nx + ny*ny; d2 <= 1 && !(have && front) {
				nz := math.Sqrt(1 - d2)
				u := math.Atan2(nx, nz) / math.Pi
				mott := fbm(u*4+3, ny*4+3, sd, 3)
				var c color.RGBA
				switch {
				case sp.World:
					c = ocean
					if fbm(u*3.2+7, ny*3.2, sd, 4) > .53 {
						c = land
					}
					if math.Abs(ny) > .84 || fbm(u*5+20, ny*9, sd+9, 3) > .6 {
						c = mixRGB(c, white, 200)
					}
				case len(bands) > 0:
					t := ny*.5 + .5 + (fbm(u*2+1, ny*6, sd, 2)-.5)*.12
					c = scale(bands[min(len(bands)-1, max(0, int(t*float64(len(bands)))))], .88+.24*mott)
					if sp.Spot && math.Pow((u-.12)/.2, 2)+math.Pow((ny-.3)/.1, 2) < 1 {
						c = red
					}
				default:
					c = scale(base, .72+.55*mott)
					switch {
					case sp.Cap && math.Abs(ny) > .82:
						c = white
					case sp.Blotch.W > 0 && mott > .62:
						c = blotch
					}
				}
				lam := nx*lx + ny*ly + nz*lz
				k := int(math.Floor(math.Max(0, lam)*4 + (Bayer(px+ext, py+ext)-.5)*.9 + .5))
				col, have = mixRGB(pal.Bg, c, steps[min(4, max(0, k))]), true
			}
			if have {
				i := img.PixOffset(px+ext, py+ext)
				img.Pix[i], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3] = col.R, col.G, col.B, 255
			}
		}
	}
	return &sprite{img: img, ext: ext}
}

// stamp copies the sprite onto the canvas, centred on x, y.
func (p *pen) stamp(s *sprite, x, y int) {
	n := 2*s.ext + 1
	for sy := 0; sy < n; sy++ {
		yy := y + sy - s.ext
		if yy < 0 || yy >= p.h {
			continue
		}
		for sx := 0; sx < n; sx++ {
			xx := x + sx - s.ext
			if xx < 0 || xx >= p.w {
				continue
			}
			if i := s.img.PixOffset(sx, sy); s.img.Pix[i+3] != 0 {
				copy(p.img.Pix[p.img.PixOffset(xx, yy):], s.img.Pix[i:i+4])
			}
		}
	}
}

// growth is how much the bodies of the system swell as the view closes in, so
// a planet has a face to look at before the zoom reaches its true size.
func (p *pen) growth() float64 { return math.Max(1, math.Min(4.5, math.Pow(p.ppu/3, .45))) }

// body draws a planet of the system at x, y, r art pixels: its sprite, lit
// from the star at the frame's centre, and its moons once there is room.
func (p *pen) body(o *world.Object, x, y, r float64) {
	if o.Body == nil {
		p.planet(x, y, math.Max(1, r), o.Tone)
		return
	}
	sx, sy := p.at(world.Vec{})
	dir := (int(math.Round(math.Atan2(sy-y, sx-x)/(2*math.Pi)*lightBuckets)) + lightBuckets) % lightBuckets
	xi, yi := int(math.Round(x)), int(math.Round(y))
	p.stamp(p.bod.get(p.pal, o, stepR(r), dir), xi, yi)
	if r < 3 || o.Body.Moons == 0 {
		return
	}
	col, k0 := p.pal.Mix(world.ToneWhite, 200), hashKey(o.Key)
	rm := r + 2
	if o.Body.Rings {
		rm = r*2.25 + 2
	}
	for k := 0; k < min(o.Body.Moons, 6); k++ {
		a := float64(hash2(k, 3, k0)) / (1 << 32) * 2 * math.Pi
		d := rm + 1.7*float64(k)
		p.px(int(math.Round(x+d*math.Cos(a))), int(math.Round(y+d*math.Sin(a))), col)
	}
}

// guides draws what lies under the bodies: each orbit as a dim dithered
// circle and the belts as scattered rocks.
func (p *pen) guides() {
	col := p.pal.Mix(world.ToneStar, 70)
	for i := range p.f.Objects {
		o := &p.f.Objects[i]
		switch {
		case o.Kind == world.KindBelt:
			p.belt(o)
		case (o.Kind == world.KindPlanet || o.Kind == world.KindComet) && o.Body != nil && o.Body.Orbit > 0:
			p.orbit(squeeze(p.f, o.Body.Orbit)*p.ppu, col)
		}
	}
}

// orbit is a circle of radius R art pixels about the star, every other pixel
// by the Bayer pattern pinned to the sky: the arcs are walked by column where
// they are shallow and by row where they are steep, so a huge one costs only
// the canvas.
func (p *pen) orbit(R float64, col color.RGBA) {
	if R < 3 || R > 1e7 {
		return
	}
	cx, cy := p.at(world.Vec{})
	ox, oy := p.pinned()
	plot := func(x, y float64) {
		if xi, yi := int(math.Round(x)), int(math.Round(y)); Bayer(xi+ox, yi+oy) < .5 {
			p.px(xi, yi, col)
		}
	}
	const q = .7072
	for xi := max(0, int(math.Ceil(cx-R*q))); xi <= min(p.w-1, int(math.Floor(cx+R*q))); xi++ {
		dx := float64(xi) - cx
		s := math.Sqrt(R*R - dx*dx)
		plot(float64(xi), cy-s)
		plot(float64(xi), cy+s)
	}
	for yi := max(0, int(math.Ceil(cy-R*q))); yi <= min(p.h-1, int(math.Floor(cy+R*q))); yi++ {
		dy := float64(yi) - cy
		s := math.Sqrt(R*R - dy*dy)
		plot(cx-s, float64(yi))
		plot(cx+s, float64(yi))
	}
}

// belt is a band of one pixel rocks. Each rock has a place in the belt (an
// angle and a spread across it) from its number alone, so the belt holds still
// as the camera moves and only spreads as the zoom closes in.
func (p *pen) belt(o *world.Object) {
	R := squeeze(p.f, o.Size) * p.ppu
	if R < 3 || R > 1e7 {
		return
	}
	cx, cy := p.at(world.Vec{})
	if cx+R < 0 || cy+R < 0 || cx-R > float64(p.w) || cy-R > float64(p.h) {
		return
	}
	hw := math.Max((squeeze(p.f, o.Size*1.13)-squeeze(p.f, o.Size*.87))/2*p.ppu, 1.5)
	seed := hashKey(o.Key)
	rocks := 900
	if o.Tone != world.ToneRock {
		rocks = 1300
	}
	for i := 0; i < rocks; i++ {
		h := hash2(i, 1, seed)
		a := float64(h&0xffff) / 65536 * 2 * math.Pi
		spread := float64(h>>16&0xff)/256 + float64(hash2(i, 2, seed)&0xff)/256 - 1 // -1..1, thick in the middle
		d := R + spread*hw
		p.px(int(math.Round(cx+d*math.Cos(a))), int(math.Round(cy+d*math.Sin(a))), p.pal.Mix(o.Tone, uint8(70+int(h>>24)*110/256)))
	}
}

// comet is a dot with a short tail streaming away from the star.
func (p *pen) comet(o *world.Object, x, y float64) {
	sx, sy := p.at(world.Vec{})
	d := math.Hypot(x-sx, y-sy)
	if d < 1 {
		d = 1
	}
	for i := 0; i < 5; i++ {
		p.px(int(math.Round(x+(x-sx)/d*float64(i))), int(math.Round(y+(y-sy)/d*float64(i))), p.pal.Mix(o.Tone, uint8(max(30, int(o.Weight)-i*45))))
	}
}
