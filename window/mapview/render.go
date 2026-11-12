package mapview

import (
	"image"
	"image/color"
	"math"

	"github.com/espresso20/ageforge/world"
)

// Display squeezes. The world keeps true numbers; these are drawing choices
// that spread what the real distances bunch up. The region squeezes distance
// with a square root like the terminal map did (the Orion arm's near features
// would sit on Sol); the system squeezes with a logarithm, the orrery trick,
// so Mercury to Neptune keep a few pixels apart in one view. Both are radial
// and monotone, so every place that turns a point into a pixel (drawing,
// picking, labels, panning, the handover to the close-up) goes through them.
const (
	regionSqueeze = 8500 // ly: where the region's squeeze leaves distance alone
	sysEdge       = .39  // AU: Mercury, where the system's curve turns from a line to a log
	sysInner      = .4   // display length of the line, against the log's 1 per e-fold
	sysScale      = 7.87 // display units per e-fold, so Pluto sits near 40 as it did unsqueezed
)

// squeeze is the display length of the true length r in frame f.
func squeeze(f *world.Frame, r float64) float64 {
	switch f.Kind {
	case world.FrameRegion:
		return math.Sqrt(r * regionSqueeze)
	case world.FrameSystem:
		if r <= sysEdge {
			return sysScale * sysInner * r / sysEdge
		}
		return sysScale * (sysInner + math.Log(r/sysEdge))
	}
	return r
}

// unsqueeze inverts squeeze.
func unsqueeze(f *world.Frame, d float64) float64 {
	switch f.Kind {
	case world.FrameRegion:
		return d * d / regionSqueeze
	case world.FrameSystem:
		if d <= sysScale*sysInner {
			return d / (sysScale * sysInner) * sysEdge
		}
		return sysEdge * math.Exp(d/sysScale-sysInner)
	}
	return d
}

// regionPos is the display position of a point in a frame.
func regionPos(f *world.Frame, p world.Vec) world.Vec {
	if f.Kind != world.FrameRegion && f.Kind != world.FrameSystem {
		return p
	}
	r := math.Hypot(p.X, p.Y)
	if r == 0 {
		return p
	}
	s := squeeze(f, r) / r
	return world.Vec{X: p.X * s, Y: p.Y * s}
}

// A pen draws one frame onto a canvas.
type pen struct {
	img    *image.RGBA
	pal    *world.Palette
	f      *world.Frame
	c      world.Vec // view centre, in the frame's units
	ppu    float64   // art pixels per frame unit
	cx, cy float64   // canvas centre
	w, h   int
	tex    *texCache
	art    *artCache
	bod    *bodyCache
	seed   int64
}

// paint draws chain frame i, with the view centred on master-frame offset off.
func (v *View) paint(img *image.RGBA, i int, off world.Vec) {
	fill(img, v.pal.Bg)
	p := v.penFor(img, i, off)
	p.frame()
}

// penFor is a pen for chain frame i on img, centred on master-frame offset off.
func (v *View) penFor(img *image.RGBA, i int, off world.Vec) pen {
	b := img.Bounds()
	p := pen{img: img, pal: &v.pal, f: v.ch[i].f, c: v.ch.centreIn(v.m, i, off),
		ppu: float64(b.Dx()) * v.ch[i].mx / v.width(), w: b.Dx(), h: b.Dy(), tex: &v.tex, art: &v.art, bod: &v.bod}
	if v.w != nil {
		p.seed = v.w.Seed
	}
	p.cx, p.cy = float64(p.w)/2, float64(p.h)/2
	return p
}

// radius is how far out an object draws, in art pixels. Size is a true radius
// in the galaxy frames and the close-up, but an art size in the system (a
// planet's 68 is not 68 AU): there a body is a dot whose radius grows with it.
func (p *pen) radius(o *world.Object) float64 {
	if o.Kind == world.KindStar {
		if p.f.Kind == world.FrameSystem {
			return starRadius(o) * math.Min(p.growth(), 2.5)
		}
		return starRadius(o)
	}
	if p.f.Kind == world.FrameSystem && o.Kind != world.KindBelt {
		return (1 + o.Size/24) * p.growth()
	}
	return o.Size * p.ppu
}

func fill(img *image.RGBA, c color.RGBA) {
	for i := 0; i < len(img.Pix); i += 4 {
		img.Pix[i], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3] = c.R, c.G, c.B, 255
	}
}

// at is where a frame point lands on the canvas.
func (p *pen) at(v world.Vec) (x, y float64) {
	v = regionPos(p.f, v)
	c := regionPos(p.f, p.c)
	return p.cx + (v.X-c.X)*p.ppu, p.cy + (v.Y-c.Y)*p.ppu
}

func (p *pen) frame() {
	if p.f.Kind == world.FramePlanet { // the home world itself, which the frame is centred on
		x, y := p.at(world.Vec{})
		p.disc(x, y, p.ppu, world.ToneBlue, 255)
		p.disc(x, y, p.ppu*0.6, world.ToneTeal, 160)
	}
	if p.f.Kind == world.FrameNeighbourhood { // the faint distance rings, a dotted circle every RingLy
		for ly := float64(world.RingLy); ly < p.f.Radius; ly += world.RingLy {
			p.ring(0, ly, world.ToneNone, 45, false)
		}
	}
	system := p.f.Kind == world.FrameSystem
	if system {
		p.guides()
	}
	for i := range p.f.Objects {
		if o := &p.f.Objects[i]; !system || o.Kind != world.KindBelt {
			p.object(o)
		}
	}
}

func (p *pen) object(o *world.Object) {
	x, y := p.at(o.Pos)
	r := p.radius(o)
	switch o.Kind {
	case world.KindGalaxy:
		p.galaxy(o, x, y, r)
	case world.KindFeature:
		if len(o.Path) <= 1 { // the home arm's path is drawn by the galaxy's own light
			p.region(o, x, y, r, p.art, p.seed)
		}
	case world.KindLane:
		tx, ty := p.at(o.To)
		p.line(x, y, tx, ty, o.Tone, o.Weight)
	case world.KindBelt:
		p.ring(0, o.Size, o.Tone, o.Weight, true)
	case world.KindPlanet, world.KindMoon:
		if p.f.Kind == world.FrameSystem {
			p.body(o, x, y, r)
			break
		}
		if o.Body != nil && o.Body.Orbit > 0 {
			p.ring(0, o.Body.Orbit, world.ToneNone, 60, false)
		}
		p.planet(x, y, math.Max(1, r), o.Tone)
	case world.KindComet:
		if p.f.Kind == world.FrameSystem {
			p.comet(o, x, y)
			break
		}
		if o.Body != nil {
			p.ring(0, o.Body.Orbit, world.ToneNone, 30, false)
		}
		p.disc(x, y, 1, o.Tone, o.Weight)
	case world.KindStation, world.KindShip:
		p.rect(x, y, math.Min(8, math.Max(2, r)), o.Tone, o.Weight)
	case world.KindStar:
		p.star(x, y, o)
	case world.KindColony:
		p.colonyMark(x, y, o)
	default: // systems, colonies, rifts, the gate
		p.disc(x, y, math.Max(1, r), o.Tone, o.Weight)
	}
}

// ring draws a circle of radius ru (frame units) round the frame's centre
// point (the origin); solid says every pixel or every other one.
func (p *pen) ring(_ int, ru float64, t world.Tone, w uint8, solid bool) {
	x0, y0 := p.at(world.Vec{})
	r := squeeze(p.f, ru) * p.ppu
	if r < 3 || x0+r < 0 || y0+r < 0 || x0-r > float64(p.w) || y0-r > float64(p.h) {
		return
	}
	col := p.pal.Mix(t, w)
	if t == world.ToneNone {
		col = p.pal.Mix(world.ToneStar, w)
	}
	n := int(math.Min(2*math.Pi*r*2, 20000))
	for k := 0; k < n; k++ {
		if !solid && k&1 == 1 {
			continue
		}
		a := 2 * math.Pi * float64(k) / float64(n)
		p.px(int(math.Round(x0+r*math.Cos(a))), int(math.Round(y0+r*math.Sin(a))), col)
	}
}
