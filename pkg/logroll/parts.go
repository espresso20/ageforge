package logroll

import (
	"image"
	"image/color"
)

// Tile is one office of the rack, w by h pixels: the office's face on a block
// at the left (35 px wide) and a paper tab at the right for its name and line
// count, a worn corner and a smudge, none alike (the seed is the office).
// inkLevel (0 to 1) is how much ink the face holds; dot puts the red dot of
// something new in the tab's corner. An office that is off is muted grey.
func Tile(office int, on bool, inkLevel float64, dot bool, seed uint64, w, h int) *image.RGBA {
	r := rng(seedOf(seed, uint64(office), 14))
	c := newCanvas(w, h, Plate)
	paper, ink, block := hx(0xdccb9f), Inks[office%Offices], hx(0x7a5a30)
	if !on {
		paper, ink, block = hx(0xa89878), hx(0x8d7f62), hx(0x948565)
	}
	fw := min(35, w)
	if w > fw {
		c.rect(fw, 1, w-fw-1, h-2, paper)
		cx := w - 2 - r.n(3)
		for i := 0; i < 3+r.n(2); i++ { // worn corner
			c.rect(cx-i, h-2-(3-i), 1+i, 1, Plate)
		}
		sx := fw + 8 + r.n(max(1, w-fw-30))
		for k := 0; k < 2+r.n(2); k++ { // smudge
			c.hline(sx+r.n(4), h-6+k, 3+r.n(6), mix(paper, hx(0x2a2418), .12+r.f()*.08))
		}
	}
	c.rect(0, 0, fw, h, block)
	c.rect(0, 0, fw, 2, mix(block, Chalk, .25))
	fc := mix(block, ink, .4+.6*inkLevel)
	if on {
		fc = mix(block, hx(0xf6eed2), .35+.65*inkLevel)
	}
	face(c, office, 3, 3, fw-6, h-6, fc, &r, 1, 0)
	if dot && w > fw {
		c.disc(w-8, 7, 4, hx(0xb21e1e))
		c.disc(w-9, 6, 1, hx(0xe06a5a))
	}
	return c.RGBA
}

// Lever is the Detail lever of layout l in pixels, at stop 0 to 2. Full is a
// brass line down the left with a notch to each stop's number box, a red knob
// at the chosen one and a pale plate behind it; compact is three small blocks
// on a row.
func Lever(l Layout, cw, ch, stop int) *image.RGBA {
	q := l.Lever
	c := newCanvas(q.W*cw, q.H*ch, Plate)
	pale, dark := hx(0xe6d3a0), hx(0x35240f)
	if l.Mode != Full {
		for i, s := range l.Stops {
			col := dark
			if i == stop {
				col = pale
			}
			c.rect((s.X-q.X)*cw+1, (s.Y-q.Y)*ch+1, s.W*cw-2, s.H*ch-2, col)
		}
		return c.RGBA
	}
	lx := 2*cw + cw/2
	a, z := (l.Stops[0].Y-q.Y)*ch+ch/2, (l.Stops[2].Y-q.Y)*ch+ch/2
	c.rect(lx-1, a, 3, z-a, BrassD)
	for i, s := range l.Stops {
		x, y := (s.X-q.X)*cw, (s.Y-q.Y)*ch
		if i == stop {
			c.rect(x, y, s.W*cw, s.H*ch, pale)
		}
		c.hline(lx, y+ch/2, x+cw-lx, BrassD)
		c.rect(x+cw/2, y+2, 3*cw, ch-4, Brass)
	}
	ky := (l.Stops[stop].Y-q.Y)*ch + ch/2
	c.disc(lx, ky, 6, hx(0x9b1f1f))
	c.disc(lx-2, ky-2, 2, hx(0xd45a4a))
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

// Warn is a chalk warning triangle with a bar and a dot, w by h pixels.
func Warn(w, h int, col color.RGBA) *image.RGBA {
	c := newCanvas(w, h, Slate)
	x, y := (w-12)/2, (h-11)/2
	c.tri(x+6, y, x, y+10, x+12, y+10, col)
	c.rect(x+5, y+4, 2, 3, Slate)
	c.rect(x+5, y+8, 2, 1, Slate)
	return c.RGBA
}

// SlateBare is the slate's unused part, w by h pixels: faint small smudges of
// chalk dust and no text.
func SlateBare(w, h int, seed uint64) *image.RGBA {
	r := rng(seedOf(seed, 19))
	c := newCanvas(w, h, Slate)
	for i, n := 0, 6+r.n(5); i < n; i++ {
		sw, sh := 10+r.n(24), 3+r.n(5)
		x0, y0 := r.n(max(w-sw, 1)), r.n(max(h-sh, 1))
		col := mix(Slate, Chalk, .04+.04*r.f())
		c.rect(x0, y0, sw, sh, col)
		c.rect(x0+3, y0+sh, sw-6, 1, col)
	}
	return c.RGBA
}
