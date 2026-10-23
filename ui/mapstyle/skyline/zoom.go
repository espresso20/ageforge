package skyline

import (
	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
)

// zoom.go is the skyline's zoom. The world is always composed at today's
// scale (every layer, the traffic, the weather, the cursor maths) into a
// frame of its own, and the zoom is how that frame reaches the screen:
//
//   - mid is the frame itself.
//   - close composes a view half as wide and shows its foot, every cell
//     drawn as 2x2 (a half block becomes two solid cells, so a roofline is
//     chunky and exact), cropped at the top: the ground stays at the bottom
//     and a tall tower runs off the top of the screen.
//   - far composes a view twice as wide and twice as tall (the buildings
//     keep the size the screen's own height gives them) and shrinks it 2x2
//     into one cell, quadrant by quadrant, the nearest thing in each
//     quadrant winning, so thin towers and rooflines survive as a skyline.
//
// Text (labels, tags, the cursor's name) and the bars are never scaled:
// they are drawn on the screen afterwards (scene.present).

type zoom uint8

const (
	zMid   zoom = iota // today's scale: the zero value
	zFar               // half scale
	zClose             // double scale
)

// zoomSteps is the levels from the widest view to the closest.
var zoomSteps = [...]zoom{zFar, zMid, zClose}

func (z zoom) String() string {
	switch z {
	case zFar:
		return "far"
	case zClose:
		return "close"
	}
	return "mid"
}

// parseZoom reads a level's name; the empty name is the default, mid.
func parseZoom(s string) (zoom, bool) {
	switch s {
	case "", "mid":
		return zMid, true
	case "far":
		return zFar, true
	case "close":
		return zClose, true
	}
	return zMid, false
}

// in is the next level closer, out the next one wider (both stop at the ends).
func (z zoom) in() zoom {
	for i, s := range zoomSteps {
		if s == z {
			return zoomSteps[min(i+1, len(zoomSteps)-1)]
		}
	}
	return z
}

func (z zoom) out() zoom {
	for i, s := range zoomSteps {
		if s == z {
			return zoomSteps[max(i-1, 0)]
		}
	}
	return z
}

// cols is how many world columns a screen sw wide shows.
func (z zoom) cols(sw int) int {
	switch z {
	case zFar:
		return 2 * sw
	case zClose:
		return (sw + 1) / 2
	}
	return sw
}

// frameRows is the height of the world's own frame for a screen sh high:
// the same at mid and close, twice the scene at far (the bars stay three
// rows, unused).
func (z zoom) frameRows(sh int) int {
	if z == zFar {
		return 2*(sh-3) + 3
	}
	return sh
}

// scrollStep is the world columns an arrow key moves: four screen columns.
func (z zoom) scrollStep() int {
	switch z {
	case zFar:
		return 8
	case zClose:
		return 2
	}
	return 4
}

func floorDiv2(x int) int {
	if x < 0 {
		return -((-x + 1) / 2)
	}
	return x / 2
}

// sx is the screen column of the world view's column x (its left cell, at close).
func (s *scene) sx(x int) int {
	switch s.zm {
	case zClose:
		return 2 * x
	case zFar:
		return floorDiv2(x)
	}
	return x
}

// sy is the screen scene row (before the header row) of the world frame's
// scene row y (its upper cell, at close).
func (s *scene) sy(y int) int {
	switch s.zm {
	case zClose:
		return 2*y - s.ls
	case zFar:
		return floorDiv2(y)
	}
	return y
}

// present puts the world frame on the screen. At mid it is the screen; at
// the other levels it returns the scene's screen-sized twin, whose frame is
// the view's own, scaled, for the markers and the bars to draw on.
func (s *scene) present() *scene {
	if s.zm == zMid {
		return s
	}
	sc := *s
	sc.fb = &s.v.fb
	sc.W, sc.H, sc.S = s.sw, s.sh, s.sh-3
	sc.fb.reset(sc.W, sc.H)
	if s.zm == zClose {
		magnify(sc.fb, &s.v.lfb)
	} else {
		shrink(sc.fb, &s.v.lfb, s.tier == mapmodel.TierASCII)
	}
	return &sc
}

// ------------------------------------------------------------ quadrants

// quadRunes is the block glyph for each set of ink quadrants: bit 0 the
// upper left, 1 the upper right, 2 the lower left, 3 the lower right.
var quadRunes = [16]rune{' ', '▘', '▝', '▀', '▖', '▌', '▞', '▛', '▗', '▚', '▐', '▜', '▄', '▙', '▟', '█'}

var quadMask = func() map[rune]uint8 {
	m := make(map[rune]uint8, 17)
	for i, r := range quadRunes {
		m[r] = uint8(i)
	}
	m[0] = 0
	return m
}()

// bgDepth is the depth of what a cell shows through its background: a
// little behind the cell's own glyph.
func bgDepth(d uint8) uint8 {
	if d > 235 {
		return 255
	}
	return d + 20
}

// magnify draws the foot of src (the world frame) at 2x into dst: every
// cell is a 2x2 block. A block glyph is exact (each of its four quadrants
// becomes one solid cell); any other glyph is repeated.
func magnify(dst, src *fb) {
	S := dst.h - 3 // screen scene rows; the world frame has as many
	for r := 0; r < S; r++ {
		k := S - 1 - r // rows above the bottom one
		ly := (src.h - 3) - 1 - k/2
		low := k%2 == 0
		for x := 0; x < dst.w; x++ {
			c := src.at(x/2, ly+1)
			if c == nil {
				continue
			}
			q := 0
			if low {
				q = 2
			}
			q += x & 1
			dst.c[(r+1)*dst.w+x] = zoomCell(*c, q)
		}
	}
}

// zoomCell is the cell of the 2x2 block that quadrant q (0 upper left … 3
// lower right) of c becomes.
func zoomCell(c cell, q int) cell {
	m, ok := quadMask[c.ch]
	if !ok {
		return c
	}
	if m == 0 {
		return cell{ch: ' ', fg: c.fg, bg: c.bg, d: c.d}
	}
	if m == 15 {
		return c
	}
	col := c.bg
	if m&(1<<q) != 0 {
		col = c.fg
	}
	return cell{ch: '█', fg: col, bg: col, d: c.d}
}

// shrink draws src (a world frame twice dst's size) into dst: each 2x2
// block of cells is one cell whose four quadrants are the blocks' cells. A
// quadrant takes the colour of the nearest thing in its cell, and the cell
// is drawn in the quadrant glyph that keeps the nearest things apart from
// the rest (a solid fill in the ASCII tier).
func shrink(dst, src *fb, ascii bool) {
	S := dst.h - 3
	var qc [4]tcell.Color
	var qd [4]uint8
	for y := 0; y < S; y++ {
		for x := 0; x < dst.w; x++ {
			near := uint8(255)
			for k := 0; k < 4; k++ {
				c := src.at(2*x+(k&1), 2*y+(k>>1)+1)
				if c == nil {
					qc[k], qd[k] = 0, 255
				} else {
					qc[k], qd[k] = c.pixel()
				}
				near = min(near, qd[k])
			}
			var mask uint8
			var inkN, bgN int
			var ink, back [4]tcell.Color
			for k := 0; k < 4; k++ {
				if qd[k] == near {
					mask |= 1 << k
					ink[inkN] = qc[k]
					inkN++
				} else {
					back[bgN] = qc[k]
					bgN++
				}
			}
			fg := meanColor(ink[:inkN])
			cl := cell{ch: quadRunes[mask], fg: fg, bg: fg, d: near}
			if mask != 15 {
				cl.bg = meanColor(back[:bgN])
				if ascii { // no quadrant glyphs: the nearest things only
					cl.ch, cl.bg = '█', fg
				}
			} else {
				cl.ch = '█'
			}
			dst.c[(y+1)*dst.w+x] = cl
		}
	}
}

// pixel is the one colour a cell is worth when it shrinks to a quadrant,
// and its depth: ink wins over what shows through (a rooftop cell reads as
// the roof, not the sky behind it), a shade or a thin glyph is a blend.
func (c *cell) pixel() (tcell.Color, uint8) {
	m, ok := quadMask[c.ch]
	switch {
	case ok && m == 0:
		return c.bg, bgDepth(c.d)
	case ok:
		return c.fg, c.d
	}
	a := 0.5
	switch c.ch {
	case '░':
		a = 0.25
	case '▒':
		a = 0.5
	case '▓':
		a = 0.75
	}
	return mixValid(c.bg, c.fg, a), c.d
}

func mixValid(a, b tcell.Color, t float64) tcell.Color {
	if !a.Valid() {
		return b
	}
	if !b.Valid() {
		return a
	}
	return theme.Mix(a, b, t)
}

// meanColor is the average colour of cs (the first when only one).
func meanColor(cs []tcell.Color) tcell.Color {
	var r, g, b, n int32
	for _, c := range cs {
		if !c.Valid() {
			continue
		}
		cr, cg, cb := c.RGB()
		r, g, b, n = r+cr, g+cg, b+cb, n+1
	}
	if n == 0 {
		if len(cs) > 0 {
			return cs[0]
		}
		return 0
	}
	return tcell.NewRGBColor((r+n/2)/n, (g+n/2)/n, (b+n/2)/n)
}
