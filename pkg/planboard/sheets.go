package planboard

import (
	"image"
	"image/color"
)

// A Bar is one gauge: a label row's bar, full when Have reaches Need. Held
// marks the part already spent on the item; Short colours a shortfall.
type Bar struct {
	Have, Need float64
	Held       float64
	Short      bool
}

func (b Bar) frac() float64 {
	if b.Need <= 0 {
		return 1
	}
	return min(b.Have/b.Need, 1)
}

// fill draws a gauge of w pixels in the colour for its ratio.
func (c *canvas) fill(x, y, w, h int, b Bar) {
	col := Ok
	switch f := b.frac(); {
	case b.Short && f < .34:
		col = Bad
	case b.Short:
		col = Wait
	}
	c.rect(x, y, w, h, Mix(Paper, PInk, .18))
	n := int(float64(w) * b.frac())
	c.rect(x, y, n, h, col)
	if b.Held > 0 {
		hw := min(int(float64(w)*min(b.Held/max(b.Need, 1e-9), 1)), n)
		c.hatch(image.Rect(x+n-hw, y, x+n, y+h), Mix(col, PInk, .5), 4, 1)
	}
	c.outline(image.Rect(x, y, x+w, y+h), PDim)
}

// Slip is the job sheet: a paper slip with a perforated top, a shadow, the big
// sign, and one gauge per entry. Gauges lie in the cells from gx to the right
// edge, one per row starting at row gy, spaced by step rows.
func (l Layout) Slip(r image.Rectangle, sign string, gx, gy, step int, gauges []Bar) *image.RGBA {
	c := part(r, l.CW, l.CH)
	cw, ch := l.CW, l.CH
	c.rect(r.Min.X*cw+4, r.Min.Y*ch+4, r.Dx()*cw, r.Dy()*ch, Mix(Sheet, black, .45))
	c.cells(r, Paper)
	rn := rng(seed(sign))
	for x := r.Min.X*cw + 6; x < r.Max.X*cw-6; x += 9 {
		c.disc(x+rn.jit(), r.Min.Y*ch+4+rn.jit(), 2, Sheet)
	}
	c.sign(sign, (r.Min.X+2)*cw, (r.Min.Y+2)*ch, rn.n(3), Paper)
	for i, g := range gauges {
		y := (gy+i*step)*ch + ch/2 - 4
		c.fill(gx*cw, y, (r.Max.X-1-gx)*cw, 8, g)
	}
	return c.img
}

// Gate is the gate card: an arch over the card and one pillar bar per
// requirement, each a row apart from row gy.
func (l Layout) Gate(r image.Rectangle, gx, gy, step int, pillars []Bar) *image.RGBA {
	c := part(r, l.CW, l.CH)
	cw, ch := l.CW, l.CH
	c.cells(r, Card)
	c.arch(r.Min.X*cw, (r.Min.Y-3)*ch, r.Dx()*cw, 3*ch, Brass)
	c.arch(r.Min.X*cw+8, (r.Min.Y-3)*ch+8, r.Dx()*cw-16, 3*ch-8, Sheet)
	c.rect(r.Min.X*cw, r.Min.Y*ch, r.Dx()*cw, 3, Brass)
	for i, g := range pillars {
		c.fill(gx*cw, (gy+i*step)*ch+ch/2-4, (r.Max.X-1-gx)*cw, 8, g)
	}
	return c.img
}

// A WaitRow is one resource the plan waits on, as a gauge row.
type WaitRow struct{ Have, Need float64 }

// WaitsPic draws the gauges of the "waits on" box, one per row from row gy, with
// the colour chosen by how much of the need is covered.
func (l Layout) WaitsPic(r image.Rectangle, gx, gy int, w []WaitRow) *image.RGBA {
	c := part(r, l.CW, l.CH)
	cw, ch := l.CW, l.CH
	c.cells(r, Card)
	c.frame(r, PDim)
	for i, v := range w {
		c.fill(gx*cw, (gy+i)*ch+ch/2-4, (r.Max.X-1-gx)*cw, 8, Bar{Have: v.Have, Need: v.Need, Short: v.Have < v.Need})
	}
	return c.img
}

// An Entry is one line of the tray's list: its sign and whether it can be
// added.
type Entry struct {
	Sign string
	Dim  bool
}

// TrayPic draws the tab strip and the entry list. Tabs are tabW cells wide on the
// strip's first row; the selected tab is underlined in brass. Entries sit one
// per row from row ey (two rows each when the layout's rows are tall), with a
// brass plus button in the last column.
func (l Layout) TrayPic(r image.Rectangle, tabs, sel, ey int, es []Entry) *image.RGBA {
	c := part(r, l.CW, l.CH)
	cw, ch := l.CW, l.CH
	c.cells(r, Card)
	for i := 0; i < tabs; i++ {
		x := r.Min.X + 1 + i*l.TabW
		c.cells(image.Rect(x, r.Min.Y, x+l.TabW-1, r.Min.Y+1), Mix(Card, Wood[2], .25))
		if i == sel {
			c.cells(image.Rect(x, r.Min.Y, x+l.TabW-1, r.Min.Y+1), Paper)
			c.rect(x*cw, (r.Min.Y+1)*ch-3, (l.TabW-1)*cw, 3, Brass)
		}
	}
	for i, e := range es {
		y := ey + i*l.RowH
		if y+l.RowH > r.Max.Y {
			break
		}
		if i%2 == 0 {
			c.cells(image.Rect(r.Min.X+1, y, r.Max.X-1, y+l.RowH), Mix(Card, Paper, .35))
		}
		rn := rng(seed(e.Sign))
		bg := color.RGBA(Mix(Card, Paper, .35))
		if i%2 == 1 {
			bg = Card
		}
		if l.RowH == 2 {
			c.sign(e.Sign, (r.Min.X+2)*cw, y*ch, rn.n(3), bg)
		} else {
			c.signSmall(e.Sign, (r.Min.X+2)*cw, y*ch, rn.n(3), bg)
		}
		bx := (r.Max.X-3)*cw + rn.jit()
		col := Brass
		if e.Dim {
			col = Mix(Card, PDim, .5)
		}
		c.rect(bx, y*ch+l.RowH*ch/2-5, 14, 10, col)
		c.rect(bx+6, y*ch+l.RowH*ch/2-3, 2, 6, PInk)
		c.rect(bx+3, y*ch+l.RowH*ch/2-1, 8, 2, PInk)
	}
	return c.img
}
