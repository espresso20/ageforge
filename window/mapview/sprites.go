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

// galaxy is a filled ellipse of its size, squashed by its tilt, dithered
// thin toward the rim. The home galaxy's disc reads its shape for the same.
func (p *pen) galaxy(o *world.Object, x, y, r float64) {
	if r < 1.5 {
		p.disc(x, y, 1, o.Tone, o.Weight)
		return
	}
	q, ang := 1.0, 0.0
	switch {
	case o.Galaxy != nil:
		q, ang = float64(max(o.Galaxy.Q, 16))/256, float64(o.Galaxy.Phi)*2*math.Pi/world.Turn
	case o.Shape != nil:
		q, ang = float64(max(o.Shape.Flat, 8))/100, float64(o.Shape.Tilt)*2*math.Pi/world.Turn
	}
	if x+r < 0 || y+r < 0 || x-r > float64(p.w) || y-r > float64(p.h) {
		return
	}
	sin, cos := math.Sincos(ang)
	for yy := max(0, int(y-r)); yy <= min(p.h-1, int(y+r)); yy++ {
		for xx := max(0, int(x-r)); xx <= min(p.w-1, int(x+r)); xx++ {
			dx, dy := float64(xx)-x, float64(yy)-y
			u, v := (dx*cos+dy*sin)/r, (-dx*sin+dy*cos)/(r*q)
			d := u*u + v*v
			if d > 1 {
				continue
			}
			k := (1 - d) * float64(o.Weight) / 255
			if k > Bayer(xx, yy) {
				p.px(xx, yy, p.pal.Mix(o.Tone, uint8(120+135*(1-d))))
			}
		}
	}
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
