// Package wirebadge draws badges as open one-pixel wire art: every shape is a
// short list of points joined by straight lines, mirrored where symmetric.
// Each tier adds strokes to the one below; nothing is filled or shaded.
package wirebadge

import (
	"image/color"
	"math"
)

// Tier is a badge rank, each adding strokes to the one below.
type Tier int

const (
	Bronze Tier = iota
	Silver
	Gold
	Platinum
	Legendary
)

var tierColor = [...]color.RGBA{
	{205, 127, 50, 255}, {196, 205, 218, 255}, {255, 205, 64, 255},
	{120, 232, 240, 255}, {255, 96, 210, 255},
}

// shape is a set of polylines as x,y pairs, mirrored across x=0 when m is set.
// Coordinates are screen pixels at scale 1.
type shape struct {
	m bool
	p [][]int
}

// poly is a closed polygon of n points about (0,0), radius r(i), as a polyline.
func poly(n int, r func(i int) float64) []int {
	var o []int
	for i := 0; i <= n; i++ {
		a, k := 2*math.Pi*float64(i%n)/float64(n)-math.Pi/2, r(i%n)
		o = append(o, int(math.Round(k*math.Cos(a))), int(math.Round(k*math.Sin(a))))
	}
	return o
}

// kernels are the grains of an ear, one pair of leaves for each row y.
func kernels(ys ...int) (o [][]int) {
	for _, y := range ys {
		o = append(o, []int{0, y + 6, 11, y - 1, 13, y - 12, 4, y - 7, 0, y + 6})
	}
	return o
}

// arc lists n+1 points of an ellipse about (cx,cy) from angle a to b degrees (0 is +x, 90 is down).
func arc(cx, cy, rx, ry, a, b float64, n int) (o []int) {
	for i := 0; i <= n; i++ {
		t := (a + (b-a)*float64(i)/float64(n)) * math.Pi / 180
		o = append(o, int(math.Round(cx+rx*math.Cos(t))), int(math.Round(cy+ry*math.Sin(t))))
	}
	return o
}

// sil is a domain's outer outline; top is where its crown sits, em the emblem's scale.
type sil struct {
	shape
	top int
	em  float64
}

// silhouettes give each domain its own shape: heater, tower, gothic arch, kite,
// roundel, octagon, Spanish shield, screen, lozenge, swallowtail, house.
var silhouettes = map[string]sil{
	"food":        {shape{true, [][]int{append([]int{0, 0, 60, 0}, arc(-57, 60, 117, 117, 0, 61, 12)...)}}, 0, 1},
	"masonry":     {shape{true, [][]int{{0, 0, 13, 0, 13, 14, 26, 14, 26, 0, 39, 0, 39, 14, 52, 14, 52, 162, 0, 162}}}, 0, 1},
	"faith":       {shape{true, [][]int{append(arc(-47, 96, 107, 107, -64, 0, 12), arc(0, 120, 60, 42, 0, 90, 10)...)}}, 0, 1},
	"military":    {shape{true, [][]int{{0, 0, 62, 56, 0, 162}}}, 0, .85},
	"trade":       {shape{true, [][]int{arc(0, 84, 62, 62, -90, 90, 24)}}, 22, 1},
	"engineering": {shape{true, [][]int{{0, 22, 26, 22, 62, 58, 62, 110, 26, 146, 0, 146}}}, 22, 1},
	"energy":      {shape{true, [][]int{append([]int{0, 0, 60, 0}, arc(0, 80, 60, 82, 0, 90, 16)...)}}, 0, 1},
	"hacker":      {shape{true, [][]int{append(append([]int{0, 16, 46, 16}, arc(46, 30, 14, 14, -90, 0, 5)...), append(arc(46, 138, 14, 14, 0, 90, 5), 0, 152)...)}}, 16, .95},
	"astronaut":   {shape{true, [][]int{{0, 0, 46, 81, 0, 162}}}, 0, .8},
	"lumber":      {shape{true, [][]int{{0, 0, 60, 0, 60, 162, 0, 120}}}, 0, 1},
	"metallurgy":  {shape{true, [][]int{{0, 0, 60, 56, 60, 162, 0, 162}}}, 0, 1},
}

// leaf is a pointed leaf from b along unit d, length l, half width hw: an
// open outline, the base left to the stem it hangs from.
func leaf(bx, by, dx, dy, l, hw float64) (o []int) {
	u, w := []float64{0, .25, .6, 1, .6, .25, 0}, []float64{0, .8, .75, 0, -.75, -.8, 0}
	for i := range u {
		o = append(o, int(math.Round(bx+dx*u[i]*l+dy*w[i]*hw)), int(math.Round(by+dy*u[i]*l-dx*w[i]*hw)))
	}
	return o
}

// laurel is a stem along the shield with five pairs of leaves pointing up.
func laurel() (o [][]int) {
	st := [][2]float64{{83, 48}, {82, 72}, {76, 98}, {63, 124}, {44, 146}, {22, 162}, {6, 167}}
	var f []int
	for _, c := range st {
		f = append(f, int(c[0]), int(c[1]))
	}
	o = [][]int{f}
	for i := 0; i < 6; i++ {
		tx, ty := st[i][0]-st[i+1][0], st[i][1]-st[i+1][1]
		h := math.Hypot(tx, ty)
		tx, ty = tx/h, ty/h
		mx, my := (st[i][0]+st[i+1][0])/2, (st[i][1]+st[i+1][1])/2
		for _, a := range []float64{.52, -.52} {
			o = append(o, leaf(mx, my, tx*math.Cos(a)-ty*math.Sin(a), tx*math.Sin(a)+ty*math.Cos(a), 20, 5))
		}
	}
	return o
}

var (
	crown  = shape{true, [][]int{{0, -46, 9, -31, 18, -42, 25, -28, 27, -20, 0, -20}, {26, -26, 0, -26}, {0, -52, 3, -49, 0, -46}}}
	wreath = shape{true, laurel()}
	// sparks are small stars at the shield's flanks.
	sparks = shape{true, [][]int{{66, 114, 84, 114}, {75, 105, 75, 123}, {70, 109, 80, 119}, {80, 109, 70, 119}, {44, 148, 56, 148}, {50, 142, 50, 154}}}
	// rays taper to a point: a kite each, at three angles.
	rays = shape{true, [][]int{{0, -102, 4, -62, 0, -62}, {75, -82, 54, -49, 48, -53, 75, -82}, {105, -53, 68, -27, 64, -33, 105, -53}}}
)

// emblems are drawn about (0,0) within about 30 pixels, by worker domain key.
var emblems = map[string]shape{
	"food":        {true, append([][]int{{0, -32, 5, -24, 0, -16}, {0, -30, 0, 32}}, kernels(-8, 6, 20)...)},
	"lumber":      {false, [][]int{{-15, 27, 9, -27}, {-12, 29, 12, -25}, {-18, 27, -13, 31}, {9, -27, 20, -30, 30, -18, 27, 2, 6, -3}, {17, -22, 25, -8}}},
	"masonry":     {true, [][]int{{30, 27, 0, 27}, {24, 27, 24, -3, 15, -18, 0, -24, 0, -9}, {12, 27, 12, 0, 0, -9}, {24, 8, 12, 8}, {24, 18, 12, 18}, {21, -9, 9, -5}}},
	"faith":       {true, [][]int{{0, -30, 6, -20, 15, -8, 21, 6, 18, 18, 9, 27, 0, 30}, {0, -6, 7, 6, 7, 16, 0, 22}, {0, 6, 3, 14, 0, 18}}},
	"trade":       {true, [][]int{{0, -30, 0, 27}, {0, 27, 12, 27}, {27, -18, 0, -18}, {27, -18, 15, 3}, {27, -18, 39, 3}, {13, 3, 18, 10, 36, 10, 41, 3, 13, 3}, {0, -30, 3, -27, 0, -24}}},
	"engineering": {false, [][]int{poly(32, func(i int) float64 { return []float64{26, 26, 19, 19}[i%4] }), poly(12, func(int) float64 { return 8 }), poly(6, func(int) float64 { return 13 })}},
	"energy":      {false, [][]int{{9, -30, -12, 3, 0, 3, -9, 30, 15, -6, 3, -6, 9, -30}, {6, -20, -4, -1}, {2, 3, -3, 14}, {8, 2, 4, 2}}},
	"hacker":      {false, [][]int{{-27, -24, 27, -24, 27, 24, -27, 24, -27, -24}, {-27, -15, 27, -15}, {-22, -20, -20, -20}, {-16, -20, -14, -20}, {-18, -6, -8, 2, -18, 10}, {-3, 10, 10, 10}, {2, -6, 20, -6}, {2, 0, 14, 0}}},
	"astronaut":   {true, [][]int{{0, -30, 9, -15, 9, 18, 0, 18}, {9, 6, 21, 24, 9, 18}, {4, 18, 5, 24, 0, 24}, {0, 26, 0, 33}, {3, 26, 0, 31}, {0, -12, 4, -8, 4, -4, 0, 0}, {-9, 9, 9, 9}}},
	"military":    {true, [][]int{{0, -32, 4, -26, 4, 6}, {0, -24, 0, 2}, {16, 9, 16, 13, 0, 13}, {16, 9, 0, 9}, {3, 13, 3, 24}, {0, 32, 5, 28, 0, 24}}},
	"metallurgy":  {false, [][]int{{-27, -12, 12, -12, 30, -3, 12, 0}, {-27, -12, -18, 0, -9, 0, -9, 15, -18, 27, 18, 27, 9, 15, 9, 0, 12, 0}, {-21, -8, 10, -8}, {-12, 22, 12, 22}, {-3, -6, 3, -6}}},
}

type pen struct {
	set    func(x, y int, c color.RGBA)
	ox, oy int
	s      float64 // scale about the shield's centre (0,84)
	c      color.RGBA
}

// line steps one pixel at a time between two points.
func (p *pen) line(x0, y0, x1, y1 int) {
	n := max(x1-x0, x0-x1, y1-y0, y0-y1)
	for i := 0; i <= n; i++ {
		t := float64(i) / float64(max(n, 1))
		p.set(p.ox+x0+int(math.Round(t*float64(x1-x0))), p.oy+y0+int(math.Round(t*float64(y1-y0))), p.c)
	}
}

// at is the scaled place of a point, scaled by k more about the row around.
func (p *pen) at(x, y, around int, k float64) (int, int) {
	return int(math.Round(float64(x) * p.s * k)), around + int(math.Round(float64(y-around)*p.s*k))
}

// each calls f for the scaled end of every segment of sh, then with only the
// start of its first.
func (p *pen) each(sh shape, around int, k float64, f func(x0, y0, x1, y1 int)) {
	for _, pl := range sh.p {
		for side := 0; side <= b2i(sh.m); side++ {
			sg := 1 - 2*side
			for i := 2; i < len(pl); i += 2 {
				x0, y0 := p.at(sg*pl[i-2], pl[i-1], around, k)
				x1, y1 := p.at(sg*pl[i], pl[i+1], around, k)
				f(x0, y0, x1, y1)
			}
		}
	}
}

// draw joins each polyline's points, scaled about the row around.
func (p *pen) draw(sh shape, around int, k float64) {
	p.each(sh, around, k, p.line)
}

// Mix blends a toward b.
func Mix(a, b color.RGBA, t float64) color.RGBA {
	m := func(x, y uint8) uint8 { return uint8(float64(x)*(1-t) + float64(y)*t + .5) }
	return color.RGBA{m(a.R, b.R), m(a.G, b.G), m(a.B, b.B), 255}
}

// Color is a tier's wire colour; legendary shimmers with frame.
func Color(t Tier, frame int) color.RGBA {
	if t == Legendary {
		return Mix(tierColor[t], color.RGBA{255, 255, 255, 255}, .15+.3*(1+math.Sin(float64(frame)*.6))/2)
	}
	return tierColor[t]
}

// Draw paints a badge through set with the shield's centre at (cx,cy).
// The art spans about x -120..120 and y -102..162 about (0,84) at scale 1, in
// screen pixels; it is wire in col over a halo two pixels out (none when halo
// is transparent).
func Draw(set func(x, y int, c color.RGBA), cx, cy int, t Tier, domain string, col, halo color.RGBA, scale float64) {
	if halo.A != 0 {
		for _, o := range [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}, {1, 1}, {-1, 1}, {1, -1}, {-1, -1}, {2, 0}, {-2, 0}, {0, 2}, {0, -2}} {
			render(&pen{set: set, ox: cx + o[0], oy: cy - 84 + o[1], s: scale, c: halo}, t, domain, false)
		}
	}
	render(&pen{set: set, ox: cx, oy: cy - 84, s: scale, c: col}, t, domain, true)
}

func render(p *pen, t Tier, domain string, emblem bool) {
	sl, ok := silhouettes[domain]
	if !ok {
		sl = silhouettes["food"]
	}
	p.draw(sl.shape, 84, 1)
	if t >= Silver {
		p.draw(sl.shape, 84, .9)
	}
	if t >= Gold {
		// a crown above, sparks at the flanks
		oy := p.oy
		p.oy += int(float64(sl.top) * p.s)
		p.draw(crown, 84, 1)
		if t == Legendary {
			p.draw(rays, 84, 1)
		}
		p.oy = oy
		if t < Platinum {
			p.draw(sparks, 84, 1) // the wreath takes this place
		}
	}
	if t >= Platinum {
		p.draw(sl.shape, 84, .8)
		p.draw(wreath, 84, 1)
	}
	if emblem {
		oy, k := p.oy, p.s
		p.oy, p.s = oy+84, k*1.4*sl.em
		p.draw(emblems[domain], 0, 1)
		p.oy, p.s = oy, k
	}
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}
