// Package wirebadge draws badges as open one-pixel wire art: every shape is a
// short list of points joined by straight lines, mirrored where symmetric.
// Tiers and looks are data; nothing is filled, shaded or smoothed.
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

// Look is the colour treatment: the same shapes in every look.
type Look int

const (
	Mono    Look = iota // one tier colour
	TwoTone             // emblem in a lighter tint
	Neon                // tier colour over a dim one-pixel halo
)

// Domain indexes Emblems.
type Domain int

const (
	Food Domain = iota
	Lumber
	Masonry
	Faith
	Trade
	Engineering
	Energy
	Hacker
	Astronaut
	Military
	Metallurgy
	NumDomains
)

var tierColor = [...]color.RGBA{
	{205, 127, 50, 255}, {196, 205, 218, 255}, {255, 205, 64, 255},
	{120, 232, 240, 255}, {255, 96, 210, 255},
}

// shape is a set of polylines as x,y pairs, mirrored across x=0 when m is set.
type shape struct {
	m bool
	p [][]int
}

// shield is the outline's half (x >= 0); the other half is its mirror.
var shield = shape{true, [][]int{{0, 2, 11, 0, 20, 0, 20, 22, 16, 38, 0, 54}}}

var crown = shape{true, [][]int{{0, -10, 4, -5, 8, -9, 8, -3, 0, -3}}}

var wings = shape{true, [][]int{{21, 18, 28, 10, 36, 4, 41, -2}, {21, 24, 30, 18, 38, 13}, {20, 30, 27, 26, 33, 22}}}

// Emblems are drawn about (0,0), within about 10 pixels, one per domain.
var Emblems = [NumDomains]shape{
	Food:        {true, [][]int{{0, -10, 0, 10}, {0, 6, 5, 2}, {0, 1, 5, -3}, {0, -4, 5, -8}}},
	Lumber:      {false, [][]int{{-5, 9, 3, -9}, {3, -9, 9, -5, 8, 1, 2, -1}}},
	Masonry:     {true, [][]int{{8, 9, 8, -1, 5, -6, 0, -8}, {4, 9, 4, 0, 0, -3}, {0, -8, 0, -4}}},
	Faith:       {true, [][]int{{0, -10, 3, -4, 7, 2, 5, 8, 0, 10}, {0, 2, 2, 6, 0, 8}}},
	Trade:       {true, [][]int{{0, -5, 8, -5, 8, 2}, {0, -8, 0, 9}, {0, 9, 4, 9}, {9, 3, 6, 6, 3, 3, 9, 3}}},
	Engineering: {true, [][]int{{0, -8, 3, -8, 8, -3, 8, 3, 3, 8, 0, 8}, {0, -8, 0, -10}, {8, 0, 10, 0}, {0, -3, 3, 0, 0, 3}, {5, -6, 7, -8}, {5, 6, 7, 8}}},
	Energy:      {false, [][]int{{3, -10, -4, 1, 0, 1, -3, 10, 5, -2, 1, -2, 3, -10}}},
	Hacker:      {false, [][]int{{-9, -7, 9, -7, 9, 8, -9, 8, -9, -7}, {-6, -3, -2, 0, -6, 3}, {1, 4, 6, 4}}},
	Astronaut:   {true, [][]int{{0, -10, 3, -5, 3, 6, 0, 6}, {3, 2, 7, 8, 3, 6}, {0, 7, 0, 10}, {0, -2, 0, -2}}},
	Military:    {false, [][]int{{9, -9, -1, 1}, {-4, -2, 2, 4}, {-1, 1, -7, 7}, {-8, 7, -6, 9}}},
	Metallurgy:  {false, [][]int{{-9, -4, 4, -4, 10, -1, 4, 0}, {-9, -4, -6, 0, -3, 0, -3, 5, -6, 9, 6, 9, 3, 5, 3, 0, 4, 0}}},
}

type pen struct {
	set    func(x, y int, c color.RGBA)
	ox, oy int
	s      float64 // scale about the shield's centre (0,28)
	c      color.RGBA
	trace  *[][2]int
}

func (p *pen) px(x, y int) {
	p.set(p.ox+x, p.oy+y, p.c)
	if p.trace != nil {
		*p.trace = append(*p.trace, [2]int{x, y})
	}
}

// line is Bresenham between two points.
func (p *pen) line(x0, y0, x1, y1 int) {
	dx, dy := abs(x1-x0), -abs(y1-y0)
	sx, sy := sign(x1-x0), sign(y1-y0)
	for e := dx + dy; ; {
		p.px(x0, y0)
		if x0 == x1 && y0 == y1 {
			return
		}
		e2 := 2 * e
		if e2 >= dy {
			e, x0 = e+dy, x0+sx
		}
		if e2 <= dx {
			e, y0 = e+dx, y0+sy
		}
	}
}

func (p *pen) draw(sh shape, around int) {
	for _, pl := range sh.p {
		for side := 0; side < 1+b2i(sh.m); side++ {
			var qx, qy int
			for i := 0; i+1 < len(pl); i += 2 {
				x, y := pl[i], pl[i+1]
				if side == 1 {
					x = -x
				}
				x, y = int(math.Round(float64(x)*p.s)), around+int(math.Round(float64(y-around)*p.s))
				if i > 0 {
					p.line(qx, qy, x, y)
				} else if len(pl) == 2 {
					p.px(x, y)
				}
				qx, qy = x, y
			}
		}
	}
}

func (p *pen) spark(x, y int) {
	p.line(x-2, y, x+2, y)
	p.line(x, y-2, x, y+2)
}

func mix(a, b color.RGBA, t float64) color.RGBA {
	m := func(x, y uint8) uint8 { return uint8(float64(x)*(1-t) + float64(y)*t + .5) }
	return color.RGBA{m(a.R, b.R), m(a.G, b.G), m(a.B, b.B), 255}
}

// Draw paints a badge through set with the shield's top centre at (cx,cy)+(0,0);
// art spans about x -50..50 and y -10..54 times scale (sparks and rays stay
// full size, so scale below 1 is for bronze and silver). frame animates the legendary tier.
func Draw(set func(x, y int, c color.RGBA), cx, cy int, t Tier, d Domain, look Look, frame int, scale float64) {
	col, white := tierColor[t], color.RGBA{255, 255, 255, 255}
	if t == Legendary {
		col = mix(col, white, .15+.3*(1+math.Sin(float64(frame)*.6))/2)
	}
	if look == Neon {
		for _, o := range [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
			render(&pen{set: set, ox: cx + o[0], oy: cy + o[1], s: scale, c: mix(col, color.RGBA{12, 14, 20, 255}, .72)}, t, d, col, false, frame)
		}
	}
	emblem := col
	if look == TwoTone {
		emblem = mix(col, white, .55)
	}
	render(&pen{set: set, ox: cx, oy: cy, s: scale, c: col}, t, d, emblem, true, frame)
}

func render(p *pen, t Tier, d Domain, emblem color.RGBA, main bool, frame int) {
	col := p.c
	var tr [][2]int
	p.trace = &tr
	p.draw(shield, 28)
	p.trace = nil
	if t >= Silver {
		s := p.s
		p.s *= .8
		p.draw(shield, 28)
		p.s = s
	}
	if t >= Gold {
		p.draw(crown, 28)
		for _, s := range [][2]int{{27, 38}, {-27, 38}, {21, 51}, {-21, 51}} {
			p.spark(s[0], s[1])
		}
	}
	if t >= Platinum {
		p.draw(wings, 28)
	}
	if t == Legendary {
		for _, deg := range []float64{22.5, 67.5, 112.5, 157.5, 247.5, 292.5} {
			a := deg * math.Pi / 180
			c, s := math.Cos(a), math.Sin(a)
			p.line(int(36*c), 28+int(36*s), int(52*c), 28+int(52*s))
		}
		if main {
			p.c = color.RGBA{255, 255, 255, 255}
			for k := 0; k < 10; k++ {
				q := tr[(frame*9+k)%len(tr)]
				p.px(q[0], q[1])
			}
		}
	}
	if main {
		oy := p.oy
		sc := p.s
		p.s = math.Min(1, sc*1.1)
		p.c, p.oy = emblem, oy+28
		p.draw(Emblems[d], 0)
		p.s = sc
		p.c, p.oy = col, oy
	}
}

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

func sign(a int) int {
	if a < 0 {
		return -1
	}
	if a > 0 {
		return 1
	}
	return 0
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}
