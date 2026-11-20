// Package logroll paints the Logs panel's pictures: the desk, the roll's
// small marks, the rack, the lever, the rail and the slate. It is pure: it
// knows cells and pixels, never words and never the game.
package logroll

import (
	"image"
	"image/color"
	"math"
)

// Palette. Panels write their text in these inks over the pictures.
var (
	DeskBG  = hx(0x2e2216)
	DeskInk = hx(0xe6d8b8)
	Paper   = hx(0xefe6c8)
	Band    = hx(0xcfc29c)
	Head    = hx(0xb9ab82)
	Light   = hx(0xf4eed9)
	Ink     = hx(0x2a2418)
	Dim     = hx(0x74664c)
	Rust    = hx(0xb8641a)
	Red     = hx(0x8c1414)
	Navy    = hx(0x2d3a5e)
	NavyInk = hx(0xf0e2b0)
	Hit     = hx(0xf2d24c)
	Slate   = hx(0x232e2c)
	Chalk   = hx(0xe8eadf)
	ChalkY  = hx(0xe6cf6a)
	ChalkD  = hx(0xa3ad9f)
	Refuse  = hx(0x6b5f55)
	Brass   = hx(0xa87a22)
	Plate   = hx(0x4a3520)
	Strip   = hx(0x3d2f1d)

	// Inks holds one ink per office, in office order.
	Inks = [10]color.RGBA{hx(0x7a3b1d), hx(0x4b3a7a), hx(0x24507a), hx(0x1f6a5a), hx(0x6a5a1c),
		hx(0x2f6a2f), hx(0x7a2f5a), hx(0x8a2424), hx(0x3a4a6a), hx(0x555555)}
	wood = [4]color.RGBA{hx(0x5a4326), hx(0x634a2a), hx(0x553e22), hx(0x5f462a)}
)

func hx(v uint32) color.RGBA { return color.RGBA{uint8(v >> 16), uint8(v >> 8), uint8(v), 255} }

func mix(a, b color.RGBA, t float64) color.RGBA {
	f := func(x, y uint8) uint8 { return uint8(float64(x)*(1-t) + float64(y)*t + .5) }
	return color.RGBA{f(a.R, b.R), f(a.G, b.G), f(a.B, b.B), 255}
}

// rng is a small seeded generator: the same seed paints the same picture.
type rng uint64

func (r *rng) next() uint64 {
	*r += 0x9e3779b97f4a7c15
	z := uint64(*r)
	z = (z ^ z>>30) * 0xbf58476d1ce4e5b9
	z = (z ^ z>>27) * 0x94d049bb133111eb
	return z ^ z>>31
}
func (r *rng) n(k int) int { return int(r.next() % uint64(max(k, 1))) }
func (r *rng) f() float64  { return float64(r.next()>>11) / (1 << 53) }

// seedOf mixes seed parts into one number.
func seedOf(parts ...uint64) uint64 {
	r := rng(0x1234abcd)
	for _, p := range parts {
		r += rng(p)
		r.next()
	}
	return r.next()
}

type canvas struct{ *image.RGBA }

func newCanvas(w, h int, bg color.RGBA) canvas {
	c := canvas{image.NewRGBA(image.Rect(0, 0, max(w, 1), max(h, 1)))}
	c.rect(0, 0, w, h, bg)
	return c
}

func (c canvas) rect(x, y, w, h int, col color.RGBA) {
	r := image.Rect(x, y, x+w, y+h).Intersect(c.Rect)
	for yy := r.Min.Y; yy < r.Max.Y; yy++ {
		o := c.PixOffset(r.Min.X, yy)
		for xx := r.Min.X; xx < r.Max.X; xx++ {
			c.Pix[o], c.Pix[o+1], c.Pix[o+2], c.Pix[o+3] = col.R, col.G, col.B, 255
			o += 4
		}
	}
}
func (c canvas) pt(x, y int, col color.RGBA)       { c.rect(x, y, 1, 1, col) }
func (c canvas) hline(x, y, w int, col color.RGBA) { c.rect(x, y, w, 1, col) }
func (c canvas) vline(x, y, h int, col color.RGBA) { c.rect(x, y, 1, h, col) }

func (c canvas) disc(cx, cy, r int, col color.RGBA) {
	for y := -r; y <= r; y++ {
		for x := -r; x <= r; x++ {
			if x*x+y*y <= r*r {
				c.pt(cx+x, cy+y, col)
			}
		}
	}
}

// tri fills the triangle with the given corners.
func (c canvas) tri(x0, y0, x1, y1, x2, y2 int, col color.RGBA) {
	ymin, ymax := min(y0, y1, y2), max(y0, y1, y2)
	xmin, xmax := min(x0, x1, x2), max(x0, x1, x2)
	e := func(ax, ay, bx, by, px, py int) int { return (bx-ax)*(py-ay) - (by-ay)*(px-ax) }
	for y := ymin; y <= ymax; y++ {
		for x := xmin; x <= xmax; x++ {
			a, b, d := e(x0, y0, x1, y1, x, y), e(x1, y1, x2, y2, x, y), e(x2, y2, x0, y0, x, y)
			if (a >= 0 && b >= 0 && d >= 0) || (a <= 0 && b <= 0 && d <= 0) {
				c.pt(x, y, col)
			}
		}
	}
}

// seg is the distance from point p to the segment a-b.
func seg(px, py, ax, ay, bx, by float64) float64 {
	dx, dy := bx-ax, by-ay
	t := math.Max(0, math.Min(1, ((px-ax)*dx+(py-ay)*dy)/(dx*dx+dy*dy)))
	return math.Hypot(px-ax-t*dx, py-ay-t*dy)
}
