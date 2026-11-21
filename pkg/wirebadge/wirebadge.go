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
type shape struct {
	m bool
	p [][]int
}

var (
	shield = shape{true, [][]int{{0, 2, 11, 0, 20, 0, 20, 22, 16, 38, 0, 54}}}
	crown  = shape{true, [][]int{{0, -14, 4, -8, 8, -13, 8, -7, 0, -7}}}
	wings  = shape{true, [][]int{{21, 18, 29, 9, 38, 3}, {21, 28, 31, 21, 38, 15}}}
	// sparks are small crosses at the shield's flanks.
	sparks = shape{true, [][]int{{25, 38, 29, 38}, {27, 36, 27, 40}, {19, 51, 23, 51}, {21, 49, 21, 53}}}
	rays   = shape{true, [][]int{{0, -22, 0, -34}, {17, -17, 24, -26}}}
)

// emblems are drawn about (0,0) within about 10 pixels, by worker domain key.
var emblems = map[string]shape{
	"food":        {true, [][]int{{0, -10, 0, 10}, {0, 6, 5, 2}, {0, 1, 5, -3}, {0, -4, 5, -8}}},
	"lumber":      {false, [][]int{{-5, 9, 3, -9}, {3, -9, 9, -5, 8, 1, 2, -1}}},
	"masonry":     {true, [][]int{{8, 9, 8, -1, 5, -6, 0, -8}, {4, 9, 4, 0, 0, -3}}},
	"faith":       {true, [][]int{{0, -10, 3, -4, 7, 2, 5, 8, 0, 10}, {0, 2, 2, 6, 0, 8}}},
	"trade":       {true, [][]int{{0, -9, 0, 9}, {0, 9, 5, 9}, {0, -6, 9, -6, 9, 1}, {9, 1, 5, 4, 13, 4, 9, 1}}},
	"engineering": {true, [][]int{{0, -7, 5, -5, 7, 0, 5, 5, 0, 7}, {0, -7, 0, -10}, {0, 7, 0, 10}, {7, 0, 10, 0}, {5, -5, 7, -7}, {5, 5, 7, 7}, {0, -3, 3, 0, 0, 3}}},
	"energy":      {false, [][]int{{3, -10, -4, 1, 0, 1, -3, 10, 5, -2, 1, -2, 3, -10}}},
	"hacker":      {false, [][]int{{-9, -8, 9, -8, 9, 8, -9, 8, -9, -8}, {-5, -4, 0, 0, -5, 4}, {-4, -3, 1, 0, -4, 3}, {3, 4, 7, 4}, {-9, -5, 9, -5}}},
	"astronaut":   {true, [][]int{{0, -10, 3, -5, 3, 6, 0, 6}, {3, 2, 7, 8, 3, 6}, {0, 7, 0, 10}}},
	"military":    {true, [][]int{{0, -10, 3, -6, 3, 4}, {0, -6, 0, 3}, {7, 5, 0, 5}, {0, 5, 0, 10}}},
	"metallurgy":  {false, [][]int{{-9, -4, 4, -4, 10, -1, 4, 0}, {-9, -4, -6, 0, -3, 0, -3, 5, -6, 9, 6, 9, 3, 5, 3, 0, 4, 0}}},
}

type pen struct {
	set    func(x, y int, c color.RGBA)
	ox, oy int
	s      float64 // scale about the shield's centre (0,28)
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

// draw joins each polyline's points, scaled about the row around; thick
// strokes are doubled a pixel to the right.
func (p *pen) draw(sh shape, around int, thick bool) {
	for _, pl := range sh.p {
		for side := 0; side <= b2i(sh.m); side++ {
			for w := 0; w <= b2i(thick); w++ {
				for i := 2; i < len(pl); i += 2 {
					var q [4]int
					for j := range q {
						v := pl[i-2+j]
						if j%2 == 0 {
							v *= 1 - 2*side
							q[j] = int(math.Round(float64(v)*p.s)) + w
						} else {
							q[j] = around + int(math.Round(float64(v-around)*p.s))
						}
					}
					p.line(q[0], q[1], q[2], q[3])
				}
			}
		}
	}
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

// Draw paints a badge through set with the shield's top centre at (cx,cy).
// The art spans about x -50..50 and y -34..54 at scale 1; it is wire in col over
// a one-pixel halo (none when halo is transparent). A scale under .6 leaves out
// the sparks and draws the emblem bolder.
func Draw(set func(x, y int, c color.RGBA), cx, cy int, t Tier, domain string, col, halo color.RGBA, scale float64) {
	if halo.A != 0 {
		for _, o := range [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
			render(&pen{set, cx + o[0], cy + o[1], scale, halo}, t, domain, false)
		}
	}
	render(&pen{set, cx, cy, scale, col}, t, domain, true)
}

func render(p *pen, t Tier, domain string, emblem bool) {
	small := p.s < .6
	p.draw(shield, 28, false)
	if t >= Silver {
		s := p.s
		p.s *= .8
		p.draw(shield, 28, false)
		p.s = s
	}
	if t >= Gold {
		p.draw(crown, 28, false)
		if !small {
			p.draw(sparks, 28, false)
		}
	}
	if t >= Platinum {
		p.draw(wings, 28, false)
	}
	if t == Legendary {
		p.draw(rays, 28, false)
	}
	if emblem {
		s, oy := p.s, p.oy
		p.s, p.oy = math.Min(1, math.Max(s*1.1, .7)), oy+28
		p.draw(emblems[domain], 0, small)
		p.s, p.oy = s, oy
	}
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}
