package mapview

import (
	"image/color"
	"math"

	"github.com/espresso20/ageforge/world"
)

func (p *pen) px(x, y int, c color.RGBA) {
	if x < 0 || y < 0 || x >= p.w || y >= p.h {
		return
	}
	i := p.img.PixOffset(x, y)
	p.img.Pix[i], p.img.Pix[i+1], p.img.Pix[i+2], p.img.Pix[i+3] = c.R, c.G, c.B, 255
}

// disc is a filled circle; a sprite under a pixel is a single dot.
func (p *pen) disc(x, y, r float64, t world.Tone, w uint8) {
	col := p.pal.Mix(t, w)
	if r > 6000 || x+r < 0 || y+r < 0 || x-r > float64(p.w) || y-r > float64(p.h) {
		return
	}
	if r <= 1 {
		p.px(int(math.Round(x)), int(math.Round(y)), col)
		return
	}
	for yy := max(0, int(y-r)); yy <= min(p.h-1, int(y+r)); yy++ {
		for xx := max(0, int(x-r)); xx <= min(p.w-1, int(x+r)); xx++ {
			if dx, dy := float64(xx)-x, float64(yy)-y; dx*dx+dy*dy <= r*r {
				p.px(xx, yy, col)
			}
		}
	}
}

// blob is a soft circle: dithered thinner toward its edge.
func (p *pen) blob(x, y, r float64, t world.Tone, w uint8) {
	if r > 6000 || x+r < 0 || y+r < 0 || x-r > float64(p.w) || y-r > float64(p.h) {
		return
	}
	col := p.pal.Mix(t, w)
	for yy := max(0, int(y-r)); yy <= min(p.h-1, int(y+r)); yy++ {
		for xx := max(0, int(x-r)); xx <= min(p.w-1, int(x+r)); xx++ {
			dx, dy := float64(xx)-x, float64(yy)-y
			if d := (dx*dx + dy*dy) / (r * r); d <= 1 && 1-d > Bayer(xx, yy) {
				p.px(xx, yy, col)
			}
		}
	}
}

func (p *pen) rect(x, y, half float64, t world.Tone, w uint8) {
	col := p.pal.Mix(t, w)
	h := int(math.Min(half, 50))
	for dy := -h + 1; dy < h; dy++ {
		for dx := -h + 1; dx < h; dx++ {
			p.px(int(math.Round(x))+dx, int(math.Round(y))+dy, col)
		}
	}
	p.px(int(math.Round(x)), int(math.Round(y)), col)
}

// line clips to the canvas and draws a one-pixel line, dim by weight.
func (p *pen) line(x0, y0, x1, y1 float64, t world.Tone, w uint8) {
	t0, t1 := 0.0, 1.0
	dx, dy := x1-x0, y1-y0
	for _, e := range [4][2]float64{{-dx, x0}, {dx, float64(p.w-1) - x0}, {-dy, y0}, {dy, float64(p.h-1) - y0}} {
		pp, qq := e[0], e[1]
		if pp == 0 {
			if qq < 0 {
				return
			}
			continue
		}
		r := qq / pp
		if pp < 0 {
			t0 = math.Max(t0, r)
		} else {
			t1 = math.Min(t1, r)
		}
	}
	if t0 >= t1 {
		return
	}
	x0, y0, x1, y1 = x0+t0*dx, y0+t0*dy, x0+t1*dx, y0+t1*dy
	col := p.pal.Mix(t, w)
	n := int(math.Max(math.Abs(x1-x0), math.Abs(y1-y0))) + 1
	for i := 0; i <= n; i++ {
		f := float64(i) / float64(n)
		p.px(int(math.Round(x0+(x1-x0)*f)), int(math.Round(y0+(y1-y0)*f)), col)
	}
}

// ring2 is a one pixel circle of radius r about x, y in colour col: the
// selection mark.
func (p *pen) ring2(x, y, r float64, col color.RGBA) {
	n := int(math.Min(2*math.Pi*r*2, 4000))
	for k := 0; k < n; k++ {
		a := 2 * math.Pi * float64(k) / float64(n)
		p.px(int(math.Round(x+r*math.Cos(a))), int(math.Round(y+r*math.Sin(a))), col)
	}
}

// star is a lit disc with a dithered glow ring round it, three art pixels
// across at least; size is the star class (1 a dwarf, 3 the sun's size).
func (p *pen) star(x, y float64, o *world.Object) {
	r := starRadius(o)
	p.blob(x, y, r+2.5, o.Tone, o.Weight/2)
	p.disc(x, y, r, o.Tone, o.Weight)
	if o.Star != nil && o.Star.Extra != world.ToneNone && r >= 2 {
		p.disc(x, y, r*.45, o.Star.Extra, 255) // the hot core
	}
}

func starRadius(o *world.Object) float64 {
	if o.Star == nil {
		return 1.5
	}
	return 1.5 + float64(o.Star.Size-1)*.75
}

// planetRamp is four steps from the night side to the lit side. The lit side
// is lifted toward white until it is as bright as the palette's mid tone, so
// a dark colour (Neptune's blue) still shows against the void.
func planetRamp(pal *world.Palette, t world.Tone) []color.RGBA {
	lit, mid := pal.Mix(t, 255), pal.Mix(world.ToneWhite, 128)
	lum := func(c color.RGBA) int { return 3*int(c.R) + 6*int(c.G) + int(c.B) }
	for k := 0; lum(lit) < lum(mid) && k < 255; k += 15 {
		w := uint8(min(255, k))
		white := pal.Tones[world.ToneWhite]
		mix := func(a, b uint8) uint8 { return uint8((int(a)*(255-int(w)) + int(b)*int(w)) / 255) }
		lit = color.RGBA{R: mix(lit.R, white.R), G: mix(lit.G, white.G), B: mix(lit.B, white.B), A: 255}
	}
	out := make([]color.RGBA, 4)
	for i := range out {
		f := []int{70, 130, 190, 255}[i]
		mix := func(a, b uint8) uint8 { return uint8((int(a)*(255-f) + int(b)*f + 127) / 255) }
		out[i] = color.RGBA{R: mix(pal.Bg.R, lit.R), G: mix(pal.Bg.G, lit.G), B: mix(pal.Bg.B, lit.B), A: 255}
	}
	return out
}

// planet is a sphere lit from the frame's centre (the star), shaded in four
// steps with an ordered dither at the terminator.
func (p *pen) planet(x, y, r float64, t world.Tone) {
	ramp := planetRamp(p.pal, t)
	if r < 2 {
		p.px(int(math.Round(x)), int(math.Round(y)), ramp[3])
		return
	}
	cx, cy := p.at(world.Vec{})
	lx, ly := cx-x, cy-y
	if d := math.Hypot(lx, ly); d > 0 {
		lx, ly = lx/d, ly/d
	} else {
		lx, ly = -.6, -.6
	}
	xi, yi := int(math.Round(x)), int(math.Round(y))
	ri := int(math.Ceil(r))
	for dy := -ri; dy <= ri; dy++ {
		for dx := -ri; dx <= ri; dx++ {
			nx, ny := float64(dx)/r, float64(dy)/r
			if nx*nx+ny*ny > 1 {
				continue
			}
			nz := math.Sqrt(1 - nx*nx - ny*ny)
			lam := nx*lx*.8 + ny*ly*.8 + nz*.6
			i := int(math.Max(0, lam)*3.2 + Bayer(xi+dx, yi+dy) - .2)
			p.px(xi+dx, yi+dy, ramp[min(3, max(0, i))])
		}
	}
}
