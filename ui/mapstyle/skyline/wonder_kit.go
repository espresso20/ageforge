package skyline

import "math"

// wonder_kit.go is what the landmarks (wonders.go) are drawn with beyond the
// building grammar: lamp beams that sweep the sky at night, the detail only
// the close zoom shows, a glow round the silhouette, and the drawing
// helpers every landmark shares.

// beamKind is how a lamp throws its light.
type beamKind uint8

const (
	beamFan    beamKind = iota // a lighthouse's: a flat fan that sweeps out to both sides
	beamSky                    // a searchlight: a slanted shaft that swings across the sky
	beamPillar                 // a column of light straight up
	beamDown                   // a lamp at the top of a tall landmark: a fan that sweeps out and down over the roofs
)

// beam is a lamp on a sprite: its cell, how it throws its light and how far
// (in cells).
type beam struct {
	x, y int
	kind beamKind
	n    int
}

// fineCell is one cell of close-zoom detail: it replaces the solid cell
// under it only while the view is at the close zoom, where a quadrant glyph
// is exactly one screen cell and so reads as a window or a carving a whole
// cell would be too coarse for.
type fineCell struct {
	x, y int
	c    scell
}

// carryExtras copies o's lamp beams and close-zoom detail onto s, moved by
// (dx, dy); what falls off the sprite is dropped.
func (s *sprite) carryExtras(o *sprite, dx, dy int) {
	for _, b := range o.beams {
		b.x += dx
		b.y += dy
		if b.y >= 0 && b.y < s.h && b.x >= 0 && b.x < s.w {
			s.beams = append(s.beams, b)
		}
	}
	for _, f := range o.fine {
		f.x += dx
		f.y += dy
		if s.in(f.x, f.y) {
			s.fine = append(s.fine, f)
		}
	}
}

// lamp puts a beam on the sprite.
func (s *sprite) lamp(x, y int, k beamKind, n int) {
	s.beams = append(s.beams, beam{x, y, k, n})
}

// detail marks a cell for the close zoom: the quadrant glyph ch in fg over
// whatever solid face is there.
func (s *sprite) detail(x, y int, ch rune, fg slot) {
	if s.in(x, y) {
		s.fine = append(s.fine, fineCell{x, y, scell{ch, fg, sNone}})
	}
}

// closeUp is the sprite as the close zoom draws it: the same cells with the
// fine detail laid on the solid faces. Cached on the sprite.
func (s *sprite) closeUp() *sprite {
	if len(s.fine) == 0 {
		return s
	}
	if s.close == nil {
		o := newSprite(s.w, s.h)
		copy(o.c, s.c)
		o.smoke, o.beacons, o.blades, o.flame, o.beams = s.smoke, s.beacons, s.blades, s.flame, s.beams
		for _, f := range s.fine {
			base := o.at(f.x, f.y)
			if base.ch != '█' {
				continue
			}
			c := f.c
			if c.bg == sNone {
				c.bg = base.fg
			}
			o.c[f.y*o.w+f.x] = c
		}
		s.close = o
	}
	return s.close
}

// haloCells are the empty cells within one step of the silhouette (the sprite's
// own box grown by one), where a night glow is drawn. Cached.
func (s *sprite) haloCells() []pt {
	if s.halo != nil {
		return s.halo
	}
	out := []pt{}
	for y := -1; y <= s.h; y++ {
		for x := -1; x <= s.w; x++ {
			if s.at(x, y).ch != 0 {
				continue
			}
			for _, d := range [4]pt{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} {
				if s.at(x+d.x, y+d.y).ch != 0 {
					out = append(out, pt{x, y})
					break
				}
			}
		}
	}
	s.halo = out
	return out
}

// ------------------------------------------------------------ drawing

// pfill fills columns x0..x1 over half rows p0..p1 (inclusive) with sl.
func (s *sprite) pfill(x0, x1, p0, p1 int, sl slot) {
	for p := p0; p <= p1; p++ {
		for x := x0; x <= x1; x++ {
			s.pset(x, p, sl)
		}
	}
}

// block fills whole cells x0..x1, y0..y1 with a full block of sl.
func (s *sprite) block(x0, x1, y0, y1 int, sl slot) {
	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			s.put(x, y, '█', sl, sNone)
		}
	}
}

// shaft draws a lit-and-shaded column w wide from row y0 to y1 (a pillar, a
// tower leg): a bright left cell, a shaded right one, the body between.
func (s *sprite) shaft(x, y0, y1, w int, lit, body, shade slot) {
	for y := y0; y <= y1; y++ {
		for i := 0; i < w; i++ {
			sl := body
			switch {
			case i == 0 && w > 1:
				sl = lit
			case i == w-1 && w > 2, i == w-1 && w == 2:
				sl = shade
			}
			s.put(x+i, y, '█', sl, sNone)
		}
	}
}

// ellipse calls f for every cell (x, half row) inside the ellipse centred on
// (cx, cpy) in cells and half rows, radii rx cells and ry half rows; dx and
// dy are the offsets from the centre in units of the radii (-1..1).
func ellipse(cx, cpy, rx, ry float64, f func(x, py int, dx, dy float64)) {
	for py := int(math.Floor(cpy - ry)); py <= int(math.Ceil(cpy+ry)); py++ {
		for x := int(math.Floor(cx - rx)); x <= int(math.Ceil(cx+rx)); x++ {
			dx := (float64(x) + 0.5 - cx) / rx
			dy := (float64(py) + 0.5 - cpy) / ry
			if dx*dx+dy*dy <= 1 {
				f(x, py, dx, dy)
			}
		}
	}
}

// unit is a landmark's size against the size it was drawn for: n rows of
// design at rows tall, never under min.
func unit(rows int, design float64) func(n float64) int {
	k := float64(rows) / design
	return func(n float64) int { return int(math.Floor(n*k + 0.5)) }
}

// atLeast returns v or lo, whichever is larger.
func atLeast(v, lo int) int {
	if v < lo {
		return lo
	}
	return v
}
