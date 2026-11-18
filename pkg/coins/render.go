package coins

import (
	"image"
	"image/color"
	"math"
)

// ground is the height of an untouched field.
const ground = 2.0

// ramp is five tones, dark to light.
type ramp [5]color.RGBA

func newRamp(a, b, c, d, e uint32) ramp {
	f := func(h uint32) color.RGBA { return color.RGBA{uint8(h >> 16), uint8(h >> 8), uint8(h), 255} }
	return ramp{f(a), f(b), f(c), f(d), f(e)}
}

func mixColor(a, b color.RGBA, t float64) color.RGBA {
	m := func(x, y uint8) uint8 { return uint8(float64(x)*(1-t) + float64(y)*t + 0.5) }
	return color.RGBA{m(a.R, b.R), m(a.G, b.G), m(a.B, b.B), 255}
}

// mixRamp pulls ramp a part-way to ramp b, step for step.
func mixRamp(a, b ramp, t float64) ramp {
	var r ramp
	for i := range r {
		r[i] = mixColor(a[i], b[i], t)
	}
	return r
}

var bayer = [4][4]float64{
	{0 / 16.0, 8 / 16.0, 2 / 16.0, 10 / 16.0},
	{12 / 16.0, 4 / 16.0, 14 / 16.0, 6 / 16.0},
	{3 / 16.0, 11 / 16.0, 1 / 16.0, 9 / 16.0},
	{15 / 16.0, 7 / 16.0, 13 / 16.0, 5 / 16.0},
}

// pick quantises tone s (0..1) to a ramp step; only the seams between steps
// are stippled.
func (r ramp) pick(s float64, x, y int) color.RGBA {
	if s <= 0 {
		return r[0]
	}
	if s >= 1 {
		return r[4]
	}
	t := s * 4
	i := int(t)
	f := t - float64(i)
	const lo, hi = 0.3, 0.7
	if f < lo {
		return r[i]
	}
	if f > hi {
		return r[i+1]
	}
	if (f-lo)/(hi-lo) > bayer[y&3][x&3] {
		return r[i+1]
	}
	return r[i]
}

// sample is what a face says about one point: its height and which of the
// four ramps colours it. A tone above 0 is an inlay lit at exactly that tone;
// seal draws the point flat, with no shading from its neighbours.
type sample struct {
	h    float64
	z    int
	tone float64
	seal bool
}

// spec is one piece: an outline, a face, and how it is shaded.
type spec struct {
	sd        func(x, y float64) float64 // outline, negative inside
	face      func(x, y float64) sample
	ramp      [4]ramp
	edgeR     *ramp
	edgeT     float64
	edge      float64
	reed      int // >0: reeding round the edge, this many sectors
	edgeSeg   int // >0: reeding on every other one of this many edge segments
	rimW      float64
	rimH      float64
	field     int // 0 matte, 1 satin, 2 mirror, 3 matte with no stipple (one pure tone)
	fieldTone float64
	gain      float64
	seed      int
	bright    float64
	band      [2]float64                            // mirror band, in 128 space
	tonF      func(x, y, d float64, px, py int) int // whole ramp steps of toning at a pixel (0 none), from tonR
	tonR      ramp
}

// lineWidth keeps a hairline at least as wide as a pixel of the 48 px version.
func lineWidth(lod int) func(float64) float64 {
	return func(w float64) float64 {
		if lod == 1 {
			return math.Max(w, 2.7)
		}
		return w
	}
}

// level picks the value for level l (1-based) from a list.
func level(l int, a ...float64) float64 { return a[min(l, len(a))-1] }

// wear is how worn a hand-struck piece of level l is.
func wear(l int) float64 { return level(l, 0.95, 0.75, 0.5, 0.3, 0.12, 0) }

// ---- small maths ----

func smooth(a, b, x float64) float64 {
	t := min(max((x-a)/(b-a), 0), 1)
	return t * t * (3 - 2*t)
}

func clamp(x, a, b float64) float64 { return math.Max(a, math.Min(b, x)) }

func hash(x, y, seed int) float64 {
	n := uint32(x*374761393+y*668265263+seed*1442695041) ^ 0x9e3779b9
	n = (n ^ (n >> 13)) * 1274126177
	n ^= n >> 16
	return float64(n&0xffff) / 65535.0
}

func valueNoise(x, y, cell float64, seed int) float64 {
	gx, gy := math.Floor(x/cell), math.Floor(y/cell)
	fx, fy := x/cell-gx, y/cell-gy
	fx, fy = fx*fx*(3-2*fx), fy*fy*(3-2*fy)
	a, b := hash(int(gx), int(gy), seed), hash(int(gx)+1, int(gy), seed)
	c, d := hash(int(gx), int(gy)+1, seed), hash(int(gx)+1, int(gy)+1, seed)
	return (a*(1-fx)+b*fx)*(1-fy) + (c*(1-fx)+d*fx)*fy
}

func fbm(x, y, cell float64, seed int) float64 {
	return 0.6*valueNoise(x, y, cell, seed) + 0.4*valueNoise(x, y, cell/2.1, seed+7)
}

func angle(x, y float64) float64 { return math.Atan2(y, x) }

func wrap(a float64) float64 {
	for a > math.Pi {
		a -= 2 * math.Pi
	}
	for a < -math.Pi {
		a += 2 * math.Pi
	}
	return a
}

func rotate(x, y, a float64) (float64, float64) {
	c, s := math.Cos(a), math.Sin(a)
	return x*c + y*s, -x*s + y*c
}

// around folds the plane into one of n sectors, so a shape drawn once repeats
// n times round the centre.
func around(x, y float64, n int, rot float64) (float64, float64) {
	a := angle(x, y) - rot
	sec := 2 * math.Pi / float64(n)
	a = math.Mod(a+sec/2+10*math.Pi*2, sec) - sec/2
	r := math.Hypot(x, y)
	return r * math.Cos(a), r * math.Sin(a)
}

// sector is the index of the nearest of n items round the centre, rot as in around.
func sector(x, y float64, n int, rot float64) int {
	a := angle(x, y) - rot
	sec := 2 * math.Pi / float64(n)
	return int(math.Floor((a+sec/2+10*math.Pi*2)/sec)) % n
}

// ---- signed distances, negative inside ----

type vec struct{ x, y float64 }

func sdCircle(x, y, cx, cy, r float64) float64 { return math.Hypot(x-cx, y-cy) - r }

func sdBox(x, y, cx, cy, hx, hy float64) float64 {
	dx, dy := math.Abs(x-cx)-hx, math.Abs(y-cy)-hy
	return math.Hypot(math.Max(dx, 0), math.Max(dy, 0)) + math.Min(math.Max(dx, dy), 0)
}

func sdRoundBox(x, y, cx, cy, hx, hy, r float64) float64 {
	return sdBox(x, y, cx, cy, hx-r, hy-r) - r
}

func sdSeg(x, y, ax, ay, bx, by, r float64) float64 {
	px, py := x-ax, y-ay
	dx, dy := bx-ax, by-ay
	t := clamp((px*dx+py*dy)/(dx*dx+dy*dy+1e-9), 0, 1)
	return math.Hypot(px-dx*t, py-dy*t) - r
}

func sdEllipse(x, y, cx, cy, a, b, rot float64) float64 {
	lx, ly := rotate(x-cx, y-cy, rot)
	k := math.Hypot(lx/a, ly/b)
	return (k - 1) * math.Min(a, b)
}

func sdPoly(x, y float64, p []vec) float64 {
	n := len(p)
	d := (x-p[0].x)*(x-p[0].x) + (y-p[0].y)*(y-p[0].y)
	s := 1.0
	for i, j := 0, n-1; i < n; j, i = i, i+1 {
		ex, ey := p[j].x-p[i].x, p[j].y-p[i].y
		wx, wy := x-p[i].x, y-p[i].y
		t := clamp((wx*ex+wy*ey)/(ex*ex+ey*ey), 0, 1)
		bx, by := wx-ex*t, wy-ey*t
		d = math.Min(d, bx*bx+by*by)
		c1, c2, c3 := y >= p[i].y, y < p[j].y, ex*wy > ey*wx
		if (c1 && c2 && c3) || (!c1 && !c2 && !c3) {
			s = -s
		}
	}
	return s * math.Sqrt(d)
}

// sdTrueEllipse is an ellipse boundary with a distance corrected for stretch.
func sdTrueEllipse(x, y, a, b, rot float64) float64 {
	lx, ly := rotate(x, y, rot)
	k := math.Hypot(lx/a, ly/b)
	g := math.Hypot(lx/(a*a), ly/(b*b)) / math.Max(k, 1e-6)
	return (k - 1) / math.Max(g, 1e-6)
}

// polyline is the distance to a chain of segments, w wide.
func polyline(x, y float64, pts []vec, w float64) float64 {
	m := 1e9
	for i := 0; i+1 < len(pts); i++ {
		m = math.Min(m, sdSeg(x, y, pts[i].x, pts[i].y, pts[i+1].x, pts[i+1].y, w))
	}
	return m
}

func sdSub(a, b float64) float64 { return math.Max(a, -b) }

// bevel is a height rising from the edge of a shape (sd < 0) over w units.
func bevel(sd, w float64) float64 { return smooth(0, 1, clamp(-sd/w, 0, 1)) }

// dome is a rounded bump of radius r and height h centred on cx, cy.
func dome(x, y, cx, cy, r, h float64) float64 {
	d := math.Hypot(x-cx, y-cy) - r
	if d >= 0 {
		return 0
	}
	return h * bevel(d, r)
}

// render shades a spec into a size x size picture: a height field lit from the
// upper left, quantised to the ramps with an ordered dither, the rim and edge
// drawn as the piece seen again, shifted down and right.
func render(sp *spec, n int) *image.RGBA {
	s := float64(n) / 128
	img := image.NewRGBA(image.Rect(0, 0, n, n))
	e := sp.edge
	c0 := float64(n-1)/2 - e*s*0.4
	W := n + 2
	H := make([]float64, W*W)
	Z := make([]int, W*W)
	T := make([]float64, W*W)
	Sl := make([]bool, W*W)
	D := make([]float64, W*W)
	for py := -1; py <= n; py++ {
		for px := -1; px <= n; px++ {
			x, y := (float64(px)-c0)/s, (float64(py)-c0)/s
			d := -sp.sd(x, y)
			i := (py+1)*W + px + 1
			D[i] = d
			st := sp.face(x, y)
			h := st.h
			if sp.rimW > 0 && d >= 0 && d < sp.rimW*1.15 {
				p := smooth(0, sp.rimW*0.45, d) * (1 - smooth(sp.rimW*0.8, sp.rimW*1.15, d))
				h = math.Max(h, ground+sp.rimH*p)
			}
			H[i], Z[i], T[i], Sl[i] = h, st.z, st.tone, st.seal
		}
	}
	lx, ly, lz := -0.5, -0.6, 0.62
	ll := math.Sqrt(lx*lx + ly*ly + lz*lz)
	lx, ly, lz = lx/ll, ly/ll, lz/ll
	gain := 1.6
	if sp.gain > 0 {
		gain *= sp.gain
	}
	switch sp.field {
	case 0:
		gain *= 0.62
	case 1:
		gain *= 0.85
	}
	for py := 0; py < n; py++ {
		for px := 0; px < n; px++ {
			i := (py+1)*W + px + 1
			x, y := (float64(px)-c0)/s, (float64(py)-c0)/s
			if D[i] < 0 {
				ex, ey := x-e*0.8, y-e*0.8
				if sp.sd(ex, ey) < 0 {
					r := sp.ramp[0]
					if sp.edgeR != nil {
						r = *sp.edgeR
					}
					t := 0.16
					if sp.edgeT > 0 {
						t = sp.edgeT
					}
					if sp.reed > 0 && (sp.edgeSeg == 0 || int(math.Floor((angle(x, y)+math.Pi)*float64(sp.edgeSeg)/(2*math.Pi)))&1 == 0) &&
						int(math.Floor((angle(x, y)+math.Pi)*float64(sp.reed)/(2*math.Pi)))&1 == 0 {
						t += 0.16
					}
					img.SetRGBA(px, py, r.pick(t+0.03*float64((px+py)&1), px, py))
				}
				continue
			}
			h := H[i]
			gx := (H[i+1] - H[i-1]) * s / 2
			gy := (H[i+W] - H[i-W]) * s / 2
			if Sl[i] {
				gx, gy = 0, 0
			}
			nx, ny := -gx, -gy
			nl := math.Sqrt(nx*nx + ny*ny + 1)
			dot := (nx*lx + ny*ly + lz) / nl
			sv := 0.5 + (dot-lz)*gain + (h-ground)*0.1
			z := Z[i]
			dd := D[i]
			rel := math.Abs(h-ground) > 0.4
			flat := z == 0 && !rel && !Sl[i] && T[i] == 0 && dd > sp.rimW*1.1
			if flat {
				switch sp.field {
				case 0:
					sv += (hash(px, py, sp.seed+2) - 0.5) * 0.07
				case 1:
					sv += (hash(0, py, sp.seed+4) - 0.5) * 0.16
				case 2:
					sv = sp.fieldTone
					w := (x - y) / 1.414
					if (w > sp.band[0] && w < sp.band[1]) || (w > sp.band[1]+7 && w < sp.band[1]+9.5) {
						sv = 0.84
					}
				}
			}
			if T[i] > 0 {
				sv = T[i]
			}
			sv += sp.bright
			if flat && sp.field == 3 { // one pure tone: snapped to a ramp step, no dither
				sv = math.Floor(sv*4+0.5) / 4
			}
			rp := sp.ramp[z]
			if sp.tonF != nil {
				if m := sp.tonF(x, y, dd, px, py); m > 0 {
					rp = sp.tonR
					sv -= 0.25 * float64(m-1)
				}
			}
			img.SetRGBA(px, py, rp.pick(sv, px, py))
		}
	}
	return img
}
