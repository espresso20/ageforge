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
//   - close composes a view half as wide and shows a window half as tall,
//     every cell drawn as 2x2 (a half block becomes two solid cells, so a
//     roofline is chunky and exact; a shape glyph such as a head is redrawn
//     at twice the size, and any other single glyph is drawn once, never
//     four times: see zoomCell). The window starts at the foot of the
//     scene and the up and down arrows raise it over a tall tower (pan).
//   - wide composes a view twice as wide and twice as tall (the buildings
//     keep the size the screen's own height gives them) and shrinks it 2x2
//     into one cell, quadrant by quadrant, the nearest thing in each
//     quadrant winning, so thin towers and rooflines survive as a skyline.
//   - far fits the whole town to the screen's width (fit.go): the frame is
//     the wide one's size, but each of its columns stands for as many world
//     columns as it takes, and the nearest-wins shrink reads it out.
//
// Text (labels, tags, the cursor's name) and the bars are never scaled:
// they are drawn on the screen afterwards (scene.present).

type zoom uint8

const (
	zMid   zoom = iota // today's scale: the zero value
	zFar               // the whole town fits the screen's width
	zClose             // double scale
	zWide              // half scale
)

// zoomSteps is the levels from the widest view to the closest.
var zoomSteps = [...]zoom{zFar, zWide, zMid, zClose}

func (z zoom) String() string {
	switch z {
	case zFar:
		return "far"
	case zWide:
		return "wide"
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
	case "wide":
		return zWide, true
	case "close":
		return zClose, true
	}
	return zMid, false
}

// levels is the zoom levels a view steps through, widest first. A town
// that is no wider than two screens has no far level of its own: the wide
// one already shows all of it.
func levels(fit bool) []zoom {
	if fit {
		return zoomSteps[:]
	}
	return zoomSteps[1:]
}

// in is the next level closer, out the next one wider (both stop at the
// ends); fit says whether the far level is a level of its own.
func (z zoom) in(fit bool) zoom  { return z.step(1, fit) }
func (z zoom) out(fit bool) zoom { return z.step(-1, fit) }

func (z zoom) step(d int, fit bool) zoom {
	ls := levels(fit)
	if z == zFar && !fit {
		z = zWide
	}
	for i, s := range ls {
		if s == z {
			return ls[clampInt(i+d, 0, len(ls)-1)]
		}
	}
	return z
}

// cols is how many columns the world frame is for a screen sw wide: the
// world columns the view shows, but at far, where it is the frame's
// columns (the town's are more: fit.go).
func (z zoom) cols(sw int) int {
	switch z {
	case zFar, zWide:
		return 2 * sw
	case zClose:
		return (sw + 1) / 2
	}
	return sw
}

// frameRows is the height of the world's own frame for a screen sh high:
// the same at mid and close, twice the scene at wide and far (the bars stay
// three rows, unused).
func (z zoom) frameRows(sh int) int {
	if z == zFar || z == zWide {
		return 2*(sh-3) + 3
	}
	return sh
}

// scrollStep is the world columns an arrow key moves: four screen columns.
func (z zoom) scrollStep() int {
	switch z {
	case zFar, zWide:
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
		return 2 * (x - s.offX)
	case zFar, zWide:
		return floorDiv2(x)
	}
	return x
}

// sy is the screen scene row (before the header row) of the world frame's
// scene row y (its upper cell, at close, where the window is raised by pan).
func (s *scene) sy(y int) int {
	switch s.zm {
	case zClose:
		return 2*(y+s.pan) - s.ls
	case zFar, zWide:
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
	if s.sS > 0 {
		sc.S = s.sS
	}
	sc.fb.reset(sc.W, sc.H)
	if s.zm == zClose {
		magnify(sc.fb, &s.v.lfb, sc.S, s.pan, s.offX, s.tier)
	} else {
		shrink(sc.fb, &s.v.lfb, s.tier == mapmodel.TierASCII)
	}
	return &sc
}

// panBar shows where the close level's window stands over a scene bigger
// than it: a thin line down the right edge with a thick stretch where the
// window is, and for the mandala, which is wider than the window too, the
// same along the bottom. It is drawn on the screen, after the scene is
// scaled.
func (s *scene) panBar() {
	if s.zm != zClose || s.W < 2 {
		return
	}
	acc, dim := theme.Color(theme.RoleAccent), theme.Color(theme.RoleDim)
	if S := s.ls; s.v.panHi > s.v.panLo {
		hi := S - 1 - s.pan // the window's lowest world row
		lo := hi - (S-1)/2  // and its highest
		x := s.W - 1
		for r := 0; r < S; r++ {
			c := s.fb.at(x, s.Y(r))
			if c == nil {
				continue
			}
			bg := c.show()
			if r >= lo && r <= hi {
				s.fb.set(x, s.Y(r), '▐', theme.Legible(acc, bg, 3), bg, dBar)
			} else {
				s.fb.set(x, s.Y(r), '▕', theme.Legible(dim, bg, 2), bg, dBar)
			}
		}
	}
	if s.v.panXHi > s.v.panXLo && s.ls > 0 {
		lo, hi := s.offX, s.offX+(s.W+1)/2-1
		y := s.Y(s.ls - 1)
		for x := 0; x < s.W-1; x++ { // (the corner is the other bar's)
			c := s.fb.at(x, y)
			if c == nil {
				continue
			}
			bg := c.show()
			if x >= lo && x <= hi {
				s.fb.set(x, y, '▄', theme.Legible(acc, bg, 3), bg, dBar)
			} else {
				s.fb.set(x, y, '▁', theme.Legible(dim, bg, 2), bg, dBar)
			}
		}
	}
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

// magnify draws a window of src (the world frame) at 2x into dst: every
// cell is a 2x2 block. The window is the foot of the frame, raised by pan
// world rows and starting offX columns in; S is the scene rows of dst. A
// block glyph is exact (each of its four quadrants becomes one solid cell);
// the other glyphs are scaled by what they are (zoomCell).
func magnify(dst, src *fb, S, pan, offX int, tier mapmodel.GlyphTier) {
	srows := src.h - (dst.h - S) // the frame's scene rows: it has the same bars
	for r := 0; r < S; r++ {
		k := S - 1 - r // rows above the bottom one
		ly := srows - 1 - k/2 - pan
		if ly < 0 {
			continue
		}
		low := k%2 == 0
		for x := 0; x < dst.w; x++ {
			c := src.at(x/2+offX, ly+1)
			if c == nil {
				continue
			}
			q := 0
			if low {
				q = 2
			}
			q += x & 1
			dst.c[(r+1)*dst.w+x] = zoomCell(*c, q, tier)
		}
	}
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
	var out tcell.Color
	n := 0
	for _, c := range cs {
		if !c.Valid() {
			continue
		}
		n++
		if n == 1 {
			out = c
		} else {
			out = theme.Mix(out, c, 1/float64(n))
		}
	}
	if n == 0 && len(cs) > 0 {
		return cs[0]
	}
	return out
}
