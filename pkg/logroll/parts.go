package logroll

import (
	"image"
	"image/color"
)

// Card is one stamp of the rack, w by h pixels. A tall box (60 px or more) is
// the whole slot: a label card with a worn corner, a tack, worn edges and a
// smudge, none alike (the seed is the office and the layout), with the
// stamp's face above it. A short box is the face alone. inkLevel (0 to 1) is
// how much ink the face holds, wax puts the wax dot of recent trouble on it.
func Card(office int, on bool, inkLevel float64, wax bool, seed uint64, w, h int) *image.RGBA {
	r := rng(seedOf(seed, uint64(office), 14))
	c := newCanvas(w, h, Plate)
	paper, ink, block := hx(0xdccb9f), Inks[office%Offices], hx(0x7a5a30)
	if !on {
		paper, ink, block = hx(0xa89878), hx(0x8d7f62), hx(0x948565)
	}
	fw, fh := min(35, w), min(45, h)
	fx, fy := (w-fw)/2, 0
	if h >= 60 {
		fy = 10
		c.rect(2, 2, w-4, h-4, paper)
		for k := 0; k < 4; k++ { // worn edge: a few missing pixels along each side
			c.rect(2+r.n(w-4), 2, 1+r.n(4), 1, Plate)
			c.rect(2+r.n(w-4), h-3, 1+r.n(4), 1, Plate)
			c.rect(2, 2+r.n(h-4), 1, 1+r.n(4), Plate)
			c.rect(w-3, 2+r.n(h-4), 1, 1+r.n(4), Plate)
		}
		cx, cy, cs := []int{2, w - 2}[r.n(2)], []int{2, h - 2}[r.n(2)], 4+r.n(7) // cut corner
		for i := 0; i < cs; i++ {
			c.rect(min(cx, cx+(cs-i)*(1-2*(cx/(w-2)))), cy+i*(1-2*(cy/(h-2))), cs-i, 1, Plate)
		}
		c.disc(10+r.n(w-20), 6+r.n(3), 2, hx(0x3a2a14)) // tack
		sx, sy := w-30+r.n(14), h-14+r.n(6)
		for k := 0; k < 3+r.n(3); k++ { // smudge
			c.hline(sx+r.n(5), sy+k, 3+r.n(7), mix(paper, hx(0x2a2418), .14+r.f()*.1))
		}
	}
	c.rect(fx, fy, fw, fh, block)
	c.rect(fx, fy, fw, 2, mix(block, Chalk, .25))
	fc := mix(block, ink, .4+.6*inkLevel)
	if on {
		fc = mix(block, hx(0xf6eed2), .35+.65*inkLevel)
	}
	face(c, office, fx+3, fy+3, fw-6, fh-6, fc, &r, 1, 0)
	if !on {
		for y := fy; y < fy+fh; y += 6 {
			c.hline(fx, y, fw, mix(block, Plate, .4))
		}
	}
	if wax {
		c.disc(fx+fw-4, fy+5, 4, hx(0xb21e1e))
		c.disc(fx+fw-5, fy+4, 1, hx(0xe06a5a))
	}
	return c.RGBA
}

// Lever is the Detail lever of layout l in pixels, at stop 0 to 2. Full is a
// slot with a knob beside three plates; compact is three blocks on a row.
func Lever(l Layout, cw, ch, stop int) *image.RGBA {
	q := l.Lever
	c := newCanvas(q.W*cw, q.H*ch, Plate)
	pale, dark := hx(0xe6d3a0), hx(0x35240f)
	for i, s := range l.Stops {
		x, y, w, h := (s.X-q.X)*cw, (s.Y-q.Y)*ch, s.W*cw, s.H*ch
		col := dark
		if i == stop {
			col = pale
		}
		if l.Mode == Full {
			c.rect(x, y+2, w, h-4, col)
			c.hline(3*cw+3, y+h/2, x-3*cw-3, hx(0x2a1c0c))
		} else {
			c.rect(x+1, y+1, w-2, h-2, col)
		}
	}
	if l.Mode == Full {
		a, z := (l.Stops[0].Y-q.Y)*ch+l.Stops[0].H*ch/2, (l.Stops[2].Y-q.Y)*ch+l.Stops[2].H*ch/2
		c.rect(cw+2, a, 4, z-a, hx(0x2a1c0c))
		ky := (l.Stops[stop].Y-q.Y)*ch + l.Stops[stop].H*ch/2
		c.disc(cw+4, ky, 8, hx(0x9b1f1f))
		c.disc(cw+2, ky-2, 2, hx(0xd45a4a))
	}
	return c.RGBA
}

// MarkKind says what a tick on the rail stands for.
type MarkKind int

const (
	Headline MarkKind = iota
	Trouble
	Seam
)

// Mark is one tick of the rail at Pos (0 the oldest line kept, 1 the newest).
type Mark struct {
	Pos  float64
	Kind MarkKind
}

// Rail paints the rail, w by h pixels: the whole kept log top to bottom, a
// brass bracket (2 px lines) around the part on screen (from to to, 0 to 1),
// a tick per headline and per trouble line and a stitched mark per seam.
func Rail(w, h int, marks []Mark, from, to float64, seed uint64) *image.RGBA {
	r := rng(seedOf(seed, 15))
	c := newCanvas(w, h, Strip)
	c.rect(2, 3, w-4, h-6, Band)
	pos := func(f float64) int { return 4 + int(f*float64(h-8)) }
	for _, m := range marks {
		y := pos(m.Pos)
		switch m.Kind {
		case Seam:
			for x := 3; x < w-3; x += 4 + r.n(2) {
				c.rect(x, y-1, 2+r.n(2), 1, Dim)
			}
		case Trouble:
			c.rect(w/2, y, w/2-3, 3, Rust)
		default:
			c.rect(w/2, y, w/2-3, 3, Navy)
		}
	}
	a, z := pos(from), max(pos(to), pos(from)+8)
	c.rect(2, a, 2, z-a, BrassD)
	c.rect(2, a, 8, 2, BrassD)
	c.rect(2, z-2, 8, 2, BrassD)
	return c.RGBA
}

// RowKind picks a full-width strip of the roll.
type RowKind int

const (
	RowBand   RowKind = iota // a band header
	RowBanner                // an age began
	RowOpen                  // an opened line: brass edge 3 px wide, left and right
)

// RowBack is the picture behind a band header, a banner or an opened line,
// w by h pixels. Edges are cut a little differently by the seed.
func RowBack(kind RowKind, w, h int, seed uint64) *image.RGBA {
	r := rng(seedOf(seed, uint64(kind), 16))
	switch kind {
	case RowBanner:
		c := newCanvas(w, h, Navy)
		c.hline(0, 2, w, Brass)
		c.hline(0, h-3, w, Brass)
		for k := 0; k < 6; k++ {
			c.rect(r.n(w), r.n(h), 1+r.n(3), 1, mix(Navy, NavyInk, .12))
		}
		return c.RGBA
	case RowOpen:
		c := newCanvas(w, h, Light)
		for _, x := range []int{0, w - 3} {
			c.rect(x, 0, 3, h, Brass)
		}
		for _, y := range []int{0, h - 2} { // corner ticks
			c.rect(0, y, 9, 2, Brass)
			c.rect(w-9, y, 9, 2, Brass)
		}
		return c.RGBA
	}
	c := newCanvas(w, h, Head)
	for _, x := range []int{0, w - 1} {
		c.rect(x, r.n(3), 1, h/2, Band)
	}
	c.hline(0, h-1, w, mix(Head, Dim, .35))
	return c.RGBA
}

// SeamRow is the stitched line between two sittings, w by h pixels on paper,
// with a gap of gapW pixels at gapX left for the label.
func SeamRow(w, h, gapX, gapW int, seed uint64) *image.RGBA {
	r := rng(seedOf(seed, 17))
	c := newCanvas(w, h, Paper)
	for x := 2; x < w-6; x += 5 + r.n(3) {
		if x+4 < gapX || x > gapX+gapW {
			c.rect(x, h/2-1+r.n(2), 3+r.n(3), 2, Dim)
		}
	}
	return c.RGBA
}

// Bar is a chalk bar on the slate: a 1 px outline and a solid fill up to frac
// (0 to 1) with a ragged end. h is the box's height, the bar is 11 px tall.
func Bar(w, h int, frac float64, seed uint64) *image.RGBA {
	r := rng(seedOf(seed, uint64(frac*1000), 18))
	c := newCanvas(w, h, Slate)
	y := (h - 11) / 2
	c.hline(0, y, w, ChalkD)
	c.hline(0, y+10, w, ChalkD)
	c.vline(0, y, 11, ChalkD)
	c.vline(w-1, y, 11, ChalkD)
	n := int(float64(w-4) * min(max(frac, 0), 1))
	c.rect(2, y+2, n, 7, Chalk)
	for i := 0; i < 1+r.n(3) && n > 0; i++ {
		c.rect(2+n+i, y+2+r.n(3), 1, 4+r.n(3), Chalk)
	}
	return c.RGBA
}

// Arrow is a chalk arrow, 12 by h pixels: dir 1 up, -1 down, 0 level.
func Arrow(dir, h int, col color.RGBA) *image.RGBA {
	c := newCanvas(12, h, Slate)
	y := (h - 11) / 2
	switch dir {
	case 1:
		c.tri(6, y, 2, y+5, 10, y+5, col)
		c.rect(5, y+5, 3, 6, col)
	case -1:
		c.tri(6, y+11, 2, y+6, 10, y+6, col)
		c.rect(5, y, 3, 6, col)
	default:
		c.rect(0, y+4, 8, 3, col)
		c.tri(8, y+1, 8, y+9, 12, y+5, col)
	}
	return c.RGBA
}

// SlateBare is the slate's unused part, w by h pixels: four to six broad rag
// swipes of different size and angle, plainly visible, and no text.
func SlateBare(w, h int, seed uint64) *image.RGBA {
	r := rng(seedOf(seed, 19))
	c := newCanvas(w, h, Slate)
	for i, n := 0, 4+r.n(3); i < n; i++ {
		th, ln := 14+r.n(24), w/3+r.n(2*w/3)
		x0, y0 := r.n(max(w-ln, 1)), r.n(max(h, 1))
		slope := float64(r.n(9)-4) / 10
		col := mix(Slate, Chalk, .12+.1*r.f())
		for x := 0; x < ln; x++ {
			c.rect(x0+x, y0+int(slope*float64(x)), 1, th, col)
		}
	}
	return c.RGBA
}
