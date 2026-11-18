package cabinet

import (
	"image"
	"image/color"
	"math"
)

func hx(u uint32) color.RGBA { return color.RGBA{uint8(u >> 16), uint8(u >> 8), uint8(u), 255} }

func ramp(u ...uint32) []color.RGBA {
	r := make([]color.RGBA, len(u))
	for i, v := range u {
		r[i] = hx(v)
	}
	return r
}

// The cabinet keeps its own colours under every theme.
var (
	walnut = ramp(0x1a110a, 0x24170e, 0x2f1e13, 0x3a2618, 0x46301f)
	brass  = ramp(0x2a2110, 0x463915, 0x665326, 0x866f36, 0xa38c4a)
	wax    = ramp(0x2a0c0c, 0x4a1414, 0x6c1e1c, 0x8c2c28, 0xa8423a)
	black  = hx(0x040405)
)

var bayer = [4][4]float64{
	{0 / 16.0, 8 / 16.0, 2 / 16.0, 10 / 16.0},
	{12 / 16.0, 4 / 16.0, 14 / 16.0, 6 / 16.0},
	{3 / 16.0, 11 / 16.0, 1 / 16.0, 9 / 16.0},
	{15 / 16.0, 7 / 16.0, 13 / 16.0, 5 / 16.0},
}

func hash(x, y, seed int) float64 {
	n := uint32(x*374761393+y*668265263+seed*1442695041) ^ 0x9e3779b9
	n = (n ^ (n >> 13)) * 1274126177
	n ^= n >> 16
	return float64(n&0xffff) / 65535.0
}

func smooth(a, b, x float64) float64 {
	t := math.Max(0, math.Min(1, (x-a)/(b-a)))
	return t * t * (3 - 2*t)
}

func clamp(x, a, b float64) float64 { return math.Max(a, math.Min(b, x)) }

func valueNoise(x, y float64, seed int) float64 {
	x0, y0 := math.Floor(x), math.Floor(y)
	fx, fy := x-x0, y-y0
	fx, fy = fx*fx*(3-2*fx), fy*fy*(3-2*fy)
	ix, iy := int(x0), int(y0)
	a, b := hash(ix, iy, seed), hash(ix+1, iy, seed)
	c, d := hash(ix, iy+1, seed), hash(ix+1, iy+1, seed)
	return (a*(1-fx)+b*fx)*(1-fy) + (c*(1-fx)+d*fx)*fy
}

// tone picks a ramp step for v in 0..1, stippling only the seams.
func tone(r []color.RGBA, v float64, x, y int) color.RGBA {
	v = clamp(v, 0, 1) * float64(len(r)-1)
	lo := int(v)
	if lo >= len(r)-1 {
		return r[len(r)-1]
	}
	if v-float64(lo) > bayer[y&3][x&3] {
		lo++
	}
	return r[lo]
}

// mul scales a colour's light, keeping its hue.
func mul(c color.RGBA, f float64) color.RGBA {
	m := func(v uint8) uint8 { return uint8(clamp(float64(v)*f+0.5, 0, 255)) }
	return color.RGBA{m(c.R), m(c.G), m(c.B), 255}
}

// canvas is a picture with its origin in panel pixels, so a patch and the
// whole front dither and grain identically. s is pixels per design unit.
type canvas struct {
	img  *image.RGBA
	org  image.Point
	s    float64
	lamp [2]float64
}

func newCanvas(r image.Rectangle, s float64) *canvas {
	return &canvas{img: image.NewRGBA(image.Rect(0, 0, r.Dx(), r.Dy())), org: r.Min, s: s, lamp: [2]float64{260 * s, -80 * s}}
}

// sc converts design units to whole pixels, never below one.
func (c *canvas) sc(v float64) int { return max(1, int(math.Round(v*c.s))) }

func (c *canvas) px(x, y int, col color.RGBA) {
	p := image.Pt(x-c.org.X, y-c.org.Y)
	if p.In(c.img.Rect) {
		c.img.SetRGBA(p.X, p.Y, col)
	}
}

func (c *canvas) get(x, y int) color.RGBA {
	p := image.Pt(x-c.org.X, y-c.org.Y)
	if p.In(c.img.Rect) {
		return c.img.RGBAAt(p.X, p.Y)
	}
	return color.RGBA{}
}

func (c *canvas) bounds() image.Rectangle { return c.img.Rect.Add(c.org) }

func (c *canvas) fill(r image.Rectangle, col color.RGBA) {
	r = r.Intersect(c.bounds())
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			c.px(x, y, col)
		}
	}
}

// fall is 0 near the lamp and 1 far from it.
func (c *canvas) fall(x, y int) float64 {
	return smooth(500*c.s, 2400*c.s, math.Hypot(float64(x)-c.lamp[0], float64(y)-c.lamp[1]))
}

// board paints a walnut board with its own grain (seed), one step of tone
// (shift) and a grain direction.
func (c *canvas) board(r image.Rectangle, seed int, shift float64, vert bool) {
	r = r.Intersect(c.bounds())
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			u, v := float64(x), float64(y)
			if vert {
				u, v = v, u
			}
			n := 0.6*valueNoise(u/46, v/1.7, seed) + 0.25*valueNoise(u/9, v/0.9, seed+9) + 0.15*valueNoise(u/3, v/0.5, seed+4)
			t := clamp((n-0.28)*1.9, 0, 1)
			ring := math.Sin(valueNoise(u/70, v/9, seed+3) * 14)
			t = 0.22 + 0.58*clamp(t*0.8+0.2*(ring*0.5+0.5), 0, 1)
			p := tone(walnut, t, x, y)
			if c.fall(x, y)*0.9 > bayer[y&3][x&3] {
				p = mul(p, 0.7)
			}
			if shift != 0 {
				p = mul(p, 1+shift)
			}
			c.px(x, y, p)
		}
	}
}

// edge draws a one pixel line along each side of r, lit at the top and left
// and dark at the bottom and right (swapped when sunk), each taken from the
// board's own pixel so the grain shows through.
func (c *canvas) edge(r image.Rectangle, sunk bool) {
	lit, dk := 1.6, 0.4
	if sunk {
		lit, dk = dk, lit
	}
	for x := r.Min.X; x < r.Max.X; x++ {
		c.px(x, r.Min.Y, mul(c.get(x, r.Min.Y), lit))
		c.px(x, r.Max.Y-1, mul(c.get(x, r.Max.Y-1), dk))
	}
	for y := r.Min.Y; y < r.Max.Y; y++ {
		c.px(r.Min.X, y, mul(c.get(r.Min.X, y), lit))
		c.px(r.Max.X-1, y, mul(c.get(r.Max.X-1, y), dk))
	}
}
