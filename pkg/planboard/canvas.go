package planboard

import (
	"image"
	"image/color"
	"math"
)

// canvas is a pixel picture drawn with hard edges only.
type canvas struct {
	img    *image.RGBA
	cw, ch int
	org    image.Point // the pixel of the panel the picture's corner lies on
}

// part is a canvas for the cells r of the panel; drawing uses panel pixels.
func part(r image.Rectangle, cw, ch int) *canvas {
	return &canvas{image.NewRGBA(image.Rect(0, 0, r.Dx()*cw, r.Dy()*ch)), cw, ch, image.Pt(r.Min.X*cw, r.Min.Y*ch)}
}

// tile is a small canvas for one sign.
func tile(cw, ch int) *canvas {
	return &canvas{image.NewRGBA(image.Rect(0, 0, 21, 30)), cw, ch, image.Point{}}
}

// Hex is the colour 0xRRGGBB.
func Hex(v uint32) color.RGBA { return color.RGBA{uint8(v >> 16), uint8(v >> 8), uint8(v), 255} }

// Mix is the colour t of the way from a to b.
func Mix(a, b color.RGBA, t float64) color.RGBA {
	m := func(x, y uint8) uint8 { return uint8(float64(x)*(1-t) + float64(y)*t + 0.5) }
	return color.RGBA{m(a.R, b.R), m(a.G, b.G), m(a.B, b.B), 255}
}

var (
	white = Hex(0xffffff)
	black = Hex(0x000000)
)

func (c *canvas) rect(x, y, w, h int, col color.RGBA) {
	x, y = x-c.org.X, y-c.org.Y
	r := image.Rect(x, y, x+w, y+h).Intersect(c.img.Rect)
	for yy := r.Min.Y; yy < r.Max.Y; yy++ {
		for xx := r.Min.X; xx < r.Max.X; xx++ {
			c.img.SetRGBA(xx, yy, col)
		}
	}
}

// cells fills whole cells.
func (c *canvas) cells(r image.Rectangle, col color.RGBA) {
	c.rect(r.Min.X*c.cw, r.Min.Y*c.ch, r.Dx()*c.cw, r.Dy()*c.ch, col)
}

func (c *canvas) pt(x, y int, col color.RGBA)       { c.rect(x, y, 1, 1, col) }
func (c *canvas) hline(x, y, w int, col color.RGBA) { c.rect(x, y, w, 1, col) }
func (c *canvas) vline(x, y, h int, col color.RGBA) { c.rect(x, y, 1, h, col) }

// dotted runs n pixels along x (horizontal) or y: on pixels, then off.
func (c *canvas) dotted(x, y, n int, horiz bool, col color.RGBA, on, off int) {
	for i := 0; i < n; i++ {
		if i%(on+off) < on {
			if horiz {
				c.pt(x+i, y, col)
			} else {
				c.pt(x, y+i, col)
			}
		}
	}
}

func (c *canvas) disc(cx, cy, r int, col color.RGBA) {
	for y := -r; y <= r; y++ {
		for x := -r; x <= r; x++ {
			if x*x+y*y <= r*r+r/2 {
				c.pt(cx+x, cy+y, col)
			}
		}
	}
}

// tri fills a triangle by testing each pixel centre.
func (c *canvas) tri(x0, y0, x1, y1, x2, y2 int, col color.RGBA) {
	e := func(ax, ay, bx, by, px, py int) int { return (bx-ax)*(py-ay) - (by-ay)*(px-ax) }
	for y := min(y0, y1, y2); y <= max(y0, y1, y2); y++ {
		for x := min(x0, x1, x2); x <= max(x0, x1, x2); x++ {
			a, b, d := e(x0, y0, x1, y1, x, y), e(x1, y1, x2, y2, x, y), e(x2, y2, x0, y0, x, y)
			if (a >= 0 && b >= 0 && d >= 0) || (a <= 0 && b <= 0 && d <= 0) {
				c.pt(x, y, col)
			}
		}
	}
}

// hatch lays diagonal stripes of col over r, 2 px wide every period px.
func (c *canvas) hatch(r image.Rectangle, col color.RGBA, period, phase int) {
	r = r.Sub(c.org).Intersect(c.img.Rect)
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			if ((x-y+phase)%period+period)%period < 2 {
				c.img.SetRGBA(x, y, col)
			}
		}
	}
}

// frame is a one pixel outline just outside the cells r (in the ring of
// cells round them, which must carry no text).
func (c *canvas) frame(r image.Rectangle, col color.RGBA) {
	x0, y0, x1, y1 := r.Min.X*c.cw-1, r.Min.Y*c.ch-1, r.Max.X*c.cw, r.Max.Y*c.ch
	c.hline(x0, y0, x1-x0+1, col)
	c.hline(x0, y1, x1-x0+1, col)
	c.vline(x0, y0, y1-y0+1, col)
	c.vline(x1, y0, y1-y0+1, col)
}

// outline is a one pixel box inside the pixel rectangle r.
func (c *canvas) outline(r image.Rectangle, col color.RGBA) {
	c.hline(r.Min.X, r.Min.Y, r.Dx(), col)
	c.hline(r.Min.X, r.Max.Y-1, r.Dx(), col)
	c.vline(r.Min.X, r.Min.Y, r.Dy(), col)
	c.vline(r.Max.X-1, r.Min.Y, r.Dy(), col)
}

// arch fills a doorway: a round top of radius w/2 over straight sides.
func (c *canvas) arch(x, y, w, h int, col color.RGBA) {
	for dx := 0; dx < w; dx++ {
		u := (float64(2*dx) - float64(w-1)) / float64(w)
		ey := int(float64(w) / 2 * (1 - math.Sqrt(1-u*u)))
		c.rect(x+dx, y+ey, 1, h-ey, col)
	}
}

// block is a bar of height 12 with a light top edge and a dark underside.
func (c *canvas) block(x, yc, w int, col color.RGBA, hollow bool) {
	if hollow {
		c.outline(image.Rect(x, yc-6, x+w, yc+6), col)
		return
	}
	c.rect(x, yc-6, w, 12, col)
	c.hline(x, yc-6, w, Mix(col, white, .35))
	c.hline(x, yc+5, w, Mix(col, black, .45))
}

// seed is a stable number for a string: decoration varies with it and never
// with a row number or a frame.
func seed(s string) uint32 {
	h := uint32(2166136261)
	for i := 0; i < len(s); i++ {
		h = (h ^ uint32(s[i])) * 16777619
	}
	return h
}

// rng is a small deterministic generator.
type rng uint32

func (r *rng) next() uint32 {
	*r = *r*1664525 + 1013904223
	return uint32(*r >> 8)
}

// n is a number in [0, k).
func (r *rng) n(k int) int { return int(r.next() % uint32(max(k, 1))) }

// jit is -1, 0 or 1.
func (r *rng) jit() int { return r.n(3) - 1 }
