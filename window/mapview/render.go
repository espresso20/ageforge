package mapview

import (
	"image"
	"image/color"
	"math"

	"github.com/espresso20/ageforge/world"
)

// regionPos is the display position of a point in a frame. The Orion arm's
// features are at true distance, which would put every near one on top of
// Sol, so the region squeezes distance with a square root like the terminal
// map did; this is a drawing choice and the world keeps the true numbers.
func regionPos(f *world.Frame, p world.Vec) world.Vec {
	if f.Kind != world.FrameRegion {
		return p
	}
	r := math.Hypot(p.X, p.Y)
	if r == 0 {
		return p
	}
	s := math.Sqrt(r/regionSqueeze) * regionSqueeze / r
	return world.Vec{X: p.X * s, Y: p.Y * s}
}

const regionSqueeze = 8500 // ly: where the squeeze leaves distance alone

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
		ppu: float64(b.Dx()) * v.ch[i].mx / v.width(), w: b.Dx(), h: b.Dy(), tex: &v.tex}
	p.cx, p.cy = float64(p.w)/2, float64(p.h)/2
	return p
}

// radius is how far out an object draws, in art pixels. Size is a true radius
// in the galaxy frames and the close-up, but an art size in the system (a
// planet's 68 is not 68 AU): there a body is a dot whose radius grows with it.
func (p *pen) radius(o *world.Object) float64 {
	if o.Kind == world.KindStar {
		return starRadius(o)
	}
	if p.f.Kind == world.FrameSystem && o.Kind != world.KindBelt {
		return 1 + o.Size/24
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
	for i := range p.f.Objects {
		p.object(&p.f.Objects[i])
	}
}

func (p *pen) object(o *world.Object) {
	x, y := p.at(o.Pos)
	r := p.radius(o)
	switch o.Kind {
	case world.KindGalaxy:
		p.galaxy(o, x, y, r)
	case world.KindFeature:
		switch {
		case len(o.Path) > 1: // the home arm: the galaxy's own light draws it
		case o.To != (world.Vec{}) || o.Key == "arm_band":
			tx, ty := p.at(o.To)
			p.line(x, y, tx, ty, o.Tone, o.Weight)
		default:
			p.blob(x, y, math.Max(r, 1), o.Tone, o.Weight)
		}
	case world.KindLane:
		tx, ty := p.at(o.To)
		p.line(x, y, tx, ty, o.Tone, o.Weight)
	case world.KindBelt:
		p.ring(0, o.Size, o.Tone, o.Weight, true)
	case world.KindPlanet, world.KindMoon:
		if o.Body != nil && o.Body.Orbit > 0 {
			p.ring(0, o.Body.Orbit, world.ToneNone, 60, false)
		}
		p.planet(x, y, math.Max(1, r), o.Tone)
	case world.KindComet:
		if o.Body != nil {
			p.ring(0, o.Body.Orbit, world.ToneNone, 30, false)
		}
		p.disc(x, y, 1, o.Tone, o.Weight)
	case world.KindStation, world.KindShip:
		p.rect(x, y, math.Min(8, math.Max(2, r)), o.Tone, o.Weight)
	case world.KindStar:
		p.star(x, y, o)
	default: // systems, colonies, rifts, the gate
		p.disc(x, y, math.Max(1, r), o.Tone, o.Weight)
	}
}

// ring draws a circle of radius ru (frame units) round the frame's centre
// point (the origin); solid says every pixel or every other one.
func (p *pen) ring(_ int, ru float64, t world.Tone, w uint8, solid bool) {
	x0, y0 := p.at(world.Vec{})
	r := ru * p.ppu
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
