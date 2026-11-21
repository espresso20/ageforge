// Package peopledash paints the pixel parts of the People panel (numerals,
// blocks, the morale dial, trend lines, bars and two status signs) onto an
// image. It knows nothing of the game: callers pass plain numbers.
package peopledash

import (
	"image"
	"image/color"
	"math"
)

// Hex turns 0xRRGGBB into an opaque colour.
func Hex(v uint32) color.RGBA {
	return color.RGBA{R: uint8(v >> 16), G: uint8(v >> 8), B: uint8(v), A: 255}
}

// Palette is the colours of the panel: the ground and tile it sits on, the
// ink, one colour per group and the morale zones.
type Palette struct {
	Ground, Tile, Rule, Ink, Text, Dim       color.RGBA
	Work, Waiting, Soldiers, Scholars, Homes color.RGBA
	Green, Amber, Alarm, Red                 color.RGBA
}

// Dark is the palette on a dark ground.
var Dark = Palette{
	Ground: Hex(0x080b12), Tile: Hex(0x18223a), Rule: Hex(0x2c3a5c), Ink: Hex(0x080b12), Text: Hex(0xf2f0e8), Dim: Hex(0x8c96ae),
	Work: Hex(0xe8b030), Waiting: Hex(0xf0e0b8), Soldiers: Hex(0xc8523e), Scholars: Hex(0x30c0b4), Homes: Hex(0x8fa8d8),
	Green: Hex(0x3fc06c), Amber: Hex(0xf0a020), Alarm: Hex(0xff4636), Red: Hex(0xd23a2c),
}

// Mix blends a toward b by t (0 is a, 1 is b).
func Mix(a, b color.RGBA, t float64) color.RGBA {
	m := func(x, y uint8) uint8 { return uint8(math.Round(float64(x)*(1-t) + float64(y)*t)) }
	return color.RGBA{m(a.R, b.R), m(a.G, b.G), m(a.B, b.B), 255}
}

// Canvas is an image with the drawing calls the panel needs. Every call
// clips to the image.
type Canvas struct {
	Img *image.RGBA
	Pal Palette
}

// New makes a canvas of w by h pixels in the palette p.
func New(w, h int, p Palette) *Canvas {
	return &Canvas{Img: image.NewRGBA(image.Rect(0, 0, w, h)), Pal: p}
}

// Rect fills a rectangle.
func (c *Canvas) Rect(x, y, w, h int, col color.RGBA) {
	r := image.Rect(x, y, x+w, y+h).Intersect(c.Img.Rect)
	for yy := r.Min.Y; yy < r.Max.Y; yy++ {
		for xx := r.Min.X; xx < r.Max.X; xx++ {
			c.Img.SetRGBA(xx, yy, col)
		}
	}
}

// Pt sets one pixel.
func (c *Canvas) Pt(x, y int, col color.RGBA) { c.Rect(x, y, 1, 1, col) }

// Line draws a one pixel line.
func (c *Canvas) Line(x0, y0, x1, y1 int, col color.RGBA) {
	dx, dy := x1-x0, y1-y0
	n := max(abs(dx), abs(dy))
	if n == 0 {
		c.Pt(x0, y0, col)
		return
	}
	for i := 0; i <= n; i++ {
		c.Pt(x0+int(math.Round(float64(dx*i)/float64(n))), y0+int(math.Round(float64(dy*i)/float64(n))), col)
	}
}

// Disc fills a disc.
func (c *Canvas) Disc(cx, cy, r int, col color.RGBA) {
	for y := -r; y <= r; y++ {
		for x := -r; x <= r; x++ {
			if x*x+y*y <= r*r+r/2 {
				c.Pt(cx+x, cy+y, col)
			}
		}
	}
}

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

// BlockPer is how many people or jobs one block stands for.
const BlockPer = 5

// BlockCount is the number of blocks for n people or jobs.
func BlockCount(n int) int {
	if n <= 0 {
		return 0
	}
	return (n + BlockPer - 1) / BlockPer
}

// Blocks lays n blocks of bw by h pixels at the given pitch, the first fill of
// them solid and the rest hollow. It returns the x after the last one.
func (c *Canvas) Blocks(x, y, h, n, fill int, col, ground color.RGBA, bw, pitch int) int {
	lt, dk := Mix(col, Hex(0xffffff), .3), Mix(col, Hex(0), .35)
	for i := 0; i < n; i++ {
		bx := x + i*pitch
		if i < fill {
			c.Rect(bx, y, bw, h, col)
			c.Rect(bx, y, bw, 2, lt)
			c.Rect(bx, y+h-2, bw, 2, dk)
		} else {
			c.Rect(bx, y, bw, h, Mix(col, ground, .55))
			c.Rect(bx+2, y+2, bw-4, h-4, ground)
		}
	}
	return x + n*pitch - (pitch - bw)
}

// Bar is a track of w by h with its first frac part filled.
func (c *Canvas) Bar(x, y, w, h int, frac float64, col, track color.RGBA) {
	c.Rect(x, y, w, h, track)
	c.Rect(x, y, int(math.Round(float64(w)*math.Min(math.Max(frac, 0), 1))), h, col)
}

// Trend draws samples as a stepped line over a faint fill, with the last
// sample marked. Fewer than two samples leave the area bare. Guides are values
// drawn as dotted lines.
func (c *Canvas) Trend(x, y, w, h int, v []float64, lo, hi float64, col, ground color.RGBA, guides ...float64) {
	if len(v) < 2 || hi <= lo {
		return
	}
	py := func(f float64) int {
		f = math.Min(math.Max(f, lo), hi)
		return y + h - 3 - int(math.Round((f-lo)/(hi-lo)*float64(h-6)))
	}
	px := func(i int) int { return x + 2 + i*(w-6)/(len(v)-1) }
	c.Rect(x, y+h-1, w, 1, c.Pal.Rule)
	c.Rect(x, y, 1, h, c.Pal.Rule)
	for _, g := range guides {
		for xx := x + 1; xx < x+w; xx += 5 {
			c.Rect(xx, py(g), 2, 1, c.Pal.Rule)
		}
	}
	fill := Mix(ground, col, .2)
	for i := 1; i < len(v); i++ {
		yy := py(v[i-1])
		c.Rect(px(i-1), yy, px(i)-px(i-1)+1, y+h-1-yy, fill)
	}
	for i := 1; i < len(v); i++ {
		for d := 0; d < 2; d++ {
			c.Line(px(i-1), py(v[i-1])+d, px(i), py(v[i-1])+d, col)
			c.Line(px(i), py(v[i-1])+d, px(i), py(v[i])+d, col)
		}
	}
	c.Rect(px(len(v)-1)-3, py(v[len(v)-1])-3, 7, 7, col)
}

// MoraleZone is the colour of a morale level: red under 40 percent, amber to
// 50, green above.
func (p Palette) MoraleZone(v float64) color.RGBA {
	switch {
	case v < .40:
		return p.Red
	case v < .50:
		return p.Amber
	}
	return p.Green
}

// DialPoint is the pixel at morale level v on a radius r from the hub, for a
// dial running from floor to cap.
func DialPoint(cx, cy int, v, floor, cap, r float64) (int, int) {
	f := (v - floor) / (cap - floor)
	return cx + int(math.Round(r*math.Cos(math.Pi*(1-f)))), cy - int(math.Round(r*math.Sin(math.Pi*(1-f))))
}

// Dial paints the morale half ring of outer radius r around the hub cx,cy:
// the zones, ticks at the floor, 40, 50 and the cap, and the needle at level.
// Levels above level fade into the tile colour.
func (c *Canvas) Dial(cx, cy, r int, level, floor, cap float64) {
	t := r / 5
	for y := -r; y <= 0; y++ {
		for x := -r; x <= r; x++ {
			d := math.Hypot(float64(x), float64(y))
			if d > float64(r) || d < float64(r-t) {
				continue
			}
			v := floor + (1-math.Atan2(float64(-y), float64(x))/math.Pi)*(cap-floor)
			if math.Mod(v, .025) < .003 {
				continue
			}
			col := c.Pal.MoraleZone(v)
			if v > level {
				col = Mix(col, c.Pal.Tile, .74)
			}
			c.Pt(cx+x, cy+y, col)
		}
	}
	for _, v := range []float64{floor, .40, .50, cap} {
		x0, y0 := DialPoint(cx, cy, v, floor, cap, float64(r+3))
		x1, y1 := DialPoint(cx, cy, v, floor, cap, float64(r+11))
		c.Line(x0, y0, x1, y1, c.Pal.Text)
	}
	f := (level - floor) / (cap - floor)
	for d := 0.0; d <= float64(r-t-6); d++ {
		c.Disc(cx+int(math.Round(d*math.Cos(math.Pi*(1-f)))), cy-int(math.Round(d*math.Sin(math.Pi*(1-f)))), 2, c.Pal.Text)
	}
	c.Disc(cx, cy, 10, c.Pal.Text)
	c.Disc(cx, cy, 4, c.Pal.Tile)
}

// Warn paints a warning triangle 17 pixels wide with its top at y.
func (c *Canvas) Warn(x, y int, col color.RGBA) {
	for i := 0; i < 15; i++ {
		w := 1 + i*17/15
		c.Rect(x-w/2, y+i, w, 1, col)
	}
	c.Rect(x-1, y+5, 3, 5, c.Pal.Ink)
	c.Rect(x-1, y+11, 3, 2, c.Pal.Ink)
}

// Arrow paints a falling (down) or rising arrow 17 pixels wide and 19 tall.
func (c *Canvas) Arrow(x, y int, down bool, col color.RGBA) {
	for i := 0; i < 9; i++ {
		if down {
			c.Rect(x-(17-2*i)/2, y+10+i, 17-2*i, 1, col)
		} else {
			c.Rect(x-(1+2*i)/2, y+i, 1+2*i, 1, col)
		}
	}
	if down {
		c.Rect(x-2, y, 5, 10, col)
	} else {
		c.Rect(x-2, y+9, 5, 10, col)
	}
}
