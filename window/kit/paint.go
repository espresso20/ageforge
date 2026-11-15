package kit

import (
	"image"
	"image/color"
	"math"

	"github.com/espresso20/ageforge/theme"
)

// A Painter is what a widget paints with: the buffer, and the rectangle it
// may touch. Coordinates are the buffer's cells.
type Painter struct {
	buf  *Buffer
	clip image.Rectangle
}

// NewPainter paints anywhere in b.
func NewPainter(b *Buffer) *Painter {
	return &Painter{buf: b, clip: image.Rect(0, 0, b.W, b.H)}
}

// Clip is a painter that may touch only r (and only where p may).
func (p *Painter) Clip(r image.Rectangle) *Painter {
	return &Painter{buf: p.buf, clip: p.clip.Intersect(r)}
}

// Bounds is the rectangle the painter may touch.
func (p *Painter) Bounds() image.Rectangle { return p.clip }

// At is the cell at x, y as it has been painted so far; a blank outside the
// buffer. It reads past the clip: a widget may look at what is beside it.
func (p *Painter) At(x, y int) Cell { return p.buf.At(x, y) }

func (p *Painter) cell(x, y int) *Cell {
	if !image.Pt(x, y).In(p.clip) {
		return nil
	}
	return &p.buf.Cells[y*p.buf.W+x]
}

// unpair blanks the other half of a wide glyph that a write to x, y cuts in
// two.
func (p *Painter) unpair(x, y int, c *Cell) {
	if c.Cont && x > 0 {
		l := &p.buf.Cells[y*p.buf.W+x-1]
		l.R, l.Comb, l.Wide = ' ', "", false
	}
	if c.Wide && x+1 < p.buf.W {
		r := &p.buf.Cells[y*p.buf.W+x+1]
		r.R, r.Comb, r.Cont = ' ', "", false
	}
	c.Wide, c.Cont = false, false
}

// ink sets a cell's colors and attributes from a style: an unset foreground
// is the theme's text, an unset ground stays.
func ink(c *Cell, st Style) {
	if fg, ok := st.Fg.RGBA(); ok {
		c.Fg = fg
	} else {
		c.Fg = theme.RGBA(theme.RoleText)
	}
	if bg, ok := st.Bg.RGBA(); ok {
		c.Bg = bg
	}
	c.Attr = st.Attr
}

// Fill blanks r and gives it the ground bg (Clear makes a hole).
func (p *Painter) Fill(r image.Rectangle, bg Color) {
	r = r.Intersect(p.clip)
	c, ok := bg.RGBA()
	if !ok || r.Empty() {
		return
	}
	for y := r.Min.Y; y < r.Max.Y; y++ {
		row := p.buf.Cells[y*p.buf.W+r.Min.X : y*p.buf.W+r.Max.X]
		for x := range row {
			p.unpair(r.Min.X+x, y, &row[x])
			row[x].R, row[x].Comb, row[x].Bg, row[x].Attr = ' ', "", c, 0
		}
	}
}

// Put writes one glyph.
func (p *Painter) Put(x, y int, r rune, st Style) {
	if c := p.cell(x, y); c != nil {
		p.unpair(x, y, c)
		c.R, c.Comb = r, ""
		ink(c, st)
	}
}

// put writes one cluster and reports the cells it took. A wide cluster with
// no room for both halves leaves a blank.
func (p *Painter) put(x, y int, g *glyph, st Style) {
	c := p.cell(x, y)
	if c == nil {
		return
	}
	p.unpair(x, y, c)
	ink(c, st)
	if g.s == "\t" || g.w == 0 {
		c.R, c.Comb = ' ', ""
		return
	}
	rs := []rune(g.s)
	if g.w < 2 {
		c.R, c.Comb = rs[0], string(rs[1:])
		return
	}
	r := p.cell(x+1, y)
	if r == nil {
		c.R, c.Comb = ' ', ""
		return
	}
	p.unpair(x+1, y, r)
	c.R, c.Comb, c.Wide = rs[0], string(rs[1:]), true
	*r = Cell{R: ' ', Fg: c.Fg, Bg: c.Bg, Attr: c.Attr, Cont: true}
}

// line paints a laid-out line from x, leaving out the clusters in its first
// skip cells and stopping after maxW cells (below 0: no limit). It returns
// the cells painted.
func (p *Painter) line(x, y int, l *Line, skip, maxW int) int {
	n := 0
	for i := range l.g {
		g := &l.g[i]
		if !printable(g) && g.s != "\t" {
			continue
		}
		if skip > 0 {
			skip -= g.w
			continue
		}
		if maxW >= 0 && n+g.w > maxW {
			break
		}
		p.put(x+n, y, g, g.st)
		for k := 1; k < g.w && g.s == "\t"; k++ { // a tab is that many blanks
			p.put(x+n+k, y, g, g.st)
		}
		n += g.w
	}
	return n
}

// Print paints tagged text on one line from x, y, cut after maxW cells
// (below 0: no limit), and returns the cells it took. Line breaks in s are
// not followed.
func (p *Painter) Print(x, y, maxW int, s string, st Style) int {
	if y < p.clip.Min.Y || y >= p.clip.Max.Y || s == "" {
		return 0
	}
	ps := parse(s, st)
	l := Line{g: glyphs(&ps, st)}
	return p.line(x, y, &l, 0, maxW)
}

// PrintAligned paints tagged text on one line inside the w cells from x.
func (p *Painter) PrintAligned(x, y, w int, s string, st Style, a Align) {
	if n := Width(s); a != AlignLeft && n < w {
		if a == AlignCenter {
			x += w/2 - n/2 // each half rounded down, as titles always were
		} else {
			x += w - n
		}
		w = n
	}
	p.Print(x, y, w, s, st)
}

// Border is the line a box is drawn with.
type Border uint8

const (
	BorderNone Border = iota
	BorderLine
	BorderRound
	BorderDouble
	BorderHeavy
	BorderASCII // for the plain glyph set
)

// borderRunes are each border's runes: across, down, then the corners from
// the top left, clockwise.
var borderRunes = [...][6]rune{
	BorderLine:   {'─', '│', '┌', '┐', '┘', '└'},
	BorderRound:  {'─', '│', '╭', '╮', '╯', '╰'},
	BorderDouble: {'═', '║', '╔', '╗', '╝', '╚'},
	BorderHeavy:  {'━', '┃', '┏', '┓', '┛', '┗'},
	BorderASCII:  {'-', '|', '+', '+', '+', '+'},
}

// Box draws a border just inside r.
func (p *Painter) Box(r image.Rectangle, b Border, st Style) {
	if b == BorderNone || r.Dx() < 2 || r.Dy() < 2 {
		return
	}
	rs := borderRunes[b]
	x0, y0, x1, y1 := r.Min.X, r.Min.Y, r.Max.X-1, r.Max.Y-1
	for x := x0 + 1; x < x1; x++ {
		p.Put(x, y0, rs[0], st)
		p.Put(x, y1, rs[0], st)
	}
	for y := y0 + 1; y < y1; y++ {
		p.Put(x0, y, rs[1], st)
		p.Put(x1, y, rs[1], st)
	}
	p.Put(x0, y0, rs[2], st)
	p.Put(x1, y0, rs[3], st)
	p.Put(x1, y1, rs[4], st)
	p.Put(x0, y1, rs[5], st)
}

func mixRGBA(a, b color.RGBA, t float64) color.RGBA {
	m := func(x, y uint8) uint8 { return uint8(math.Round(float64(x) + (float64(y)-float64(x))*t)) }
	return color.RGBA{m(a.R, b.R), m(a.G, b.G), m(a.B, b.B), a.A}
}

// Shade mixes everything already painted in r toward a color by t (0 leaves
// it, 1 is the color): how a modal dims the screen behind it. Holes stay
// holes.
func (p *Painter) Shade(r image.Rectangle, toward Color, t float64) {
	to, ok := toward.RGBA()
	r = r.Intersect(p.clip)
	if !ok || r.Empty() || t <= 0 {
		return
	}
	for y := r.Min.Y; y < r.Max.Y; y++ {
		row := p.buf.Cells[y*p.buf.W+r.Min.X : y*p.buf.W+r.Max.X]
		for x := range row {
			row[x].Fg = mixRGBA(row[x].Fg, to, t)
			if !row[x].Clear() {
				row[x].Bg = mixRGBA(row[x].Bg, to, t)
			}
		}
	}
}

// Cursor shows the text cursor at x, y, when the painter may touch it.
func (p *Painter) Cursor(x, y int) {
	if image.Pt(x, y).In(p.clip) {
		p.buf.Cursor, p.buf.CursorOn = image.Pt(x, y), true
	}
}

// Bar paints a scroll bar down the h cells from x, y: a view of size rows
// starting at top, over total rows. The thumb's ends land on eighths of a
// cell. Nothing is painted when everything fits.
func (p *Painter) Bar(x, y, h, top, size, total int) {
	if total <= size || h <= 0 {
		return
	}
	track, thumb := Role(theme.RoleChip), Role(theme.RoleDim)
	// the thumb, in eighths of a cell from the bar's top
	from, to := thumbSpan(h, top, size, total)
	for i := 0; i < h; i++ {
		a, b := max(from-i*8, 0), min(to-i*8, 8) // the thumb's part of this cell
		switch {
		case b <= 0 || a >= 8:
			p.Put(x, y+i, ' ', Style{Bg: track})
		case a == 0 && b == 8:
			p.Put(x, y+i, '█', Style{Fg: thumb, Bg: track})
		case a > 0: // starts part way down: a lower block of the thumb
			p.Put(x, y+i, rune(0x2580+8-a), Style{Fg: thumb, Bg: track})
		default: // ends part way down: a lower block of the track under it
			p.Put(x, y+i, rune(0x2580+8-b), Style{Fg: track, Bg: thumb})
		}
	}
}

// thumbSpan is where a scroll bar's thumb starts and ends, in eighths of a
// cell, never shorter than one cell.
func thumbSpan(h, top, size, total int) (from, to int) {
	full := h * 8
	length := max(full*size/total, 8)
	from = 0
	if total > size {
		from = (full - length) * top / (total - size)
	}
	from = max(0, min(from, full-length))
	return from, from + length
}
