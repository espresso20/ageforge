package mapgen

import (
	"image"
	"image/color"
	"runtime"
	"sync"
	"sync/atomic"
)

// The Galactic void again, for a terminal that can show a picture. It is the
// same layout: void() runs once more with a recorder on the canvas, and what it
// noted (soft ellipses of light, stars) is drawn here at the pixel size of the
// window instead of two pixels a cell.

// wop is one note of the recorder: an ellipse of light in channel k, or a star.
// Positions are isotropic, where a cell is two wide and four tall.
type wop struct {
	cx, cy, rx, ry float64
	phi, k         int
	amp            float64
	star           bool
	p              Paint
}

// VoidImage draws the void between galaxies at w by h cells of cellW by cellH
// pixels: the seeded layout of OverviewVoid, as smooth light on bg, with hue
// saying what each kind of light looks like. It is the same for the same
// arguments; a star does not twinkle.
func (s *Scene) VoidImage(w, h, cellW, cellH int, bg color.RGBA, hue func(Hue) color.RGBA) *image.RGBA {
	var ops []wop
	c := &wcanvas{w: 2 * w, h: 2 * h, rec: &ops}
	c.void(s.Seed, func(uint8, int, int, int) {})
	img := image.NewRGBA(image.Rect(0, 0, w*cellW, h*cellH))
	r := &raster{ops: ops, seed: s.Seed, w: w * cellW, h: h * cellH, sx: float64(cellW) / 2, sy: float64(cellH) / 4, bg: bg}
	for i := range r.hue {
		r.hue[i] = hue(Hue(i))
	}
	for k := 0; k < nch; k++ {
		r.ch[k] = hue(chHue[k])
	}
	const strip = 8
	var next atomic.Int64
	var wg sync.WaitGroup
	for n := runtime.GOMAXPROCS(0); n > 0; n-- {
		wg.Add(1)
		go func() {
			defer wg.Done()
			acc := make([]float32, (nch+1)*r.w*strip)
			rgb := make([]float32, 3*r.w*strip)
			for {
				y0 := int(next.Add(1)-1) * strip
				if y0 >= r.h {
					return
				}
				r.strip(img, y0, min(y0+strip, r.h), acc, rgb)
			}
		}()
	}
	wg.Wait()
	return img
}

type raster struct {
	ops    []wop
	seed   int64
	w, h   int
	sx, sy float64 // pixels per isotropic unit
	bg     color.RGBA
	hue    [NumHues]color.RGBA
	ch     [nch]color.RGBA
}

// strip draws pixel rows y0..y1 into img. acc and rgb are scratch.
func (r *raster) strip(img *image.RGBA, y0, y1 int, acc, rgb []float32) {
	n := r.w * (y1 - y0)
	clear(acc[:(nch+1)*n])
	for _, o := range r.ops {
		if !o.star {
			r.light(o, y0, y1, acc)
		}
	}
	bg := [3]float32{float32(r.bg.R), float32(r.bg.G), float32(r.bg.B)}
	for i := 0; i < n; i++ {
		a := acc[i*(nch+1) : (i+1)*(nch+1)]
		var tot float32
		for k := 0; k < nch; k++ {
			tot += a[k]
		}
		out := bg
		if tot > 0 {
			dust := 1 - min(a[cDust], 235)/256
			w := tot * dust
			f := float32(0) // a soft shoulder, so bright light keeps its shape instead of going flat
			if w > 2 {
				f = 255 * (1 - float32(expNeg(float64(w-2)/170))) / tot
			}
			for k := 0; k < nch; k++ {
				if a[k] == 0 {
					continue
				}
				m := a[k] * f / 255
				c := r.ch[k]
				out[0] += (float32(c.R) - bg[0]) * m
				out[1] += (float32(c.G) - bg[1]) * m
				out[2] += (float32(c.B) - bg[2]) * m
			}
		}
		copy(rgb[3*i:], out[:])
	}
	for i, o := range r.ops {
		if o.star {
			r.dot(i, o, y0, y1, rgb)
		}
	}
	for y := y0; y < y1; y++ {
		for x := 0; x < r.w; x++ {
			j := 3 * ((y-y0)*r.w + x)
			d := noise(x, y)
			img.SetRGBA(x, y, color.RGBA{q8(rgb[j], d), q8(rgb[j+1], noise(y, x)), q8(rgb[j+2], d), 255})
		}
	}
}

// noise is a repeatable value in [-0.5, 0.5) for a pixel, to break up banding
// before the colours are rounded.
func noise(x, y int) float32 {
	h := uint32(x)*0x9E3779B1 ^ uint32(y)*0x85EBCA6B
	h ^= h >> 15
	h *= 0x2C1B3C6D
	h ^= h >> 12
	return float32(h&0xFFFF)/65536 - 0.5
}

func q8(v, d float32) uint8 {
	return uint8(min(255, max(0, v+d+0.5)))
}

// light adds the ellipse o to the channel sums of rows y0..y1: amp times
// (1-rho^2)^2, rho the distance from its middle in its own radii.
func (r *raster) light(o wop, y0, y1 int, acc []float32) {
	ext := max(o.rx, o.ry)
	xa, xb := floorF((o.cx-ext)*r.sx), floorF((o.cx+ext)*r.sx)+1
	ya, yb := max(y0, floorF((o.cy-ext)*r.sy)), min(y1-1, floorF((o.cy+ext)*r.sy)+1)
	xa, xb = max(xa, 0), min(xb, r.w-1)
	cs, sn := float64(cos4096(o.phi))/1024, float64(sin4096(o.phi))/1024
	for y := ya; y <= yb; y++ {
		dy := (float64(y)+0.5)/r.sy - o.cy
		row := (y - y0) * r.w
		for x := xa; x <= xb; x++ {
			dx := (float64(x)+0.5)/r.sx - o.cx
			u, v := (dx*cs+dy*sn)/o.rx, (dy*cs-dx*sn)/o.ry
			d := u*u + v*v
			if d >= 1 {
				continue
			}
			t := 1 - d
			acc[(row+x)*(nch+1)+o.k] += float32(o.amp * t * t)
		}
	}
}

// dot lays star o (the i-th note) over rgb: a soft point a pixel or two wide, a
// little wider the brighter it is, with a faint glow round the brightest.
func (r *raster) dot(i int, o wop, y0, y1 int, rgb []float32) {
	wt := float64(o.p.Weight) / 255
	// the star sits a pixel or so off its lattice point, so it is not on a grid
	jx := (float64(hash(r.seed, 7090, int64(i))%1000)/1000 - 0.5) * 2
	jy := (float64(hash(r.seed, 7091, int64(i))%1000)/1000 - 0.5) * 2
	px, py := o.cx*r.sx+jx, o.cy*r.sy+jy
	sigma := 0.55 + 0.7*wt
	glow := o.p.Weight >= 200
	reach := 3 * sigma
	if glow {
		reach = 7
	}
	col := r.hue[o.p.Hue]
	cr, cg, cb := float32(col.R), float32(col.G), float32(col.B)
	for y := max(y0, floorF(py-reach)); y <= min(y1-1, floorF(py+reach)); y++ {
		for x := max(0, floorF(px-reach)); x <= min(r.w-1, floorF(px+reach)); x++ {
			dx, dy := float64(x)+0.5-px, float64(y)+0.5-py
			d2 := dx*dx + dy*dy
			a := expNeg(d2 / (2 * sigma * sigma))
			if glow {
				a += 0.22 * expNeg(d2/(2*3.2*3.2))
			}
			a = min(1, a*min(1, wt*1.5+0.1))
			if a < 0.01 {
				continue
			}
			j := 3 * ((y-y0)*r.w + x)
			f := float32(a)
			rgb[j] += (cr - rgb[j]) * f
			rgb[j+1] += (cg - rgb[j+1]) * f
			rgb[j+2] += (cb - rgb[j+2]) * f
		}
	}
}

// The engine keeps to integers and its own tables, so the picture is the same
// everywhere; these are the few real-number helpers it needs.

func floorF(x float64) int {
	i := int(x)
	if float64(i) > x {
		i--
	}
	return i
}

// expNeg is e^-x for x >= 0, close enough for a falloff and cut off at 0 past 6.
func expNeg(x float64) float64 {
	if x >= 6 {
		return 0
	}
	return 1 / (1 + x + x*x/2 + x*x*x/6 + x*x*x*x/24 + x*x*x*x*x/120)
}

// sinF and cosF are the sine tables read between their steps; a is in turns of 4096.
func sinF(a float64) float64 {
	i := floorF(a)
	f := a - float64(i)
	return (float64(sin4096(i))*(1-f) + float64(sin4096(i+1))*f) / 1024
}

func cosF(a float64) float64 { return sinF(a + turn/4) }
