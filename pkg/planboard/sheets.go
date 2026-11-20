package planboard

import (
	"image"
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
	c.signBig(sign, (r.Min.X+2)*cw, (r.Min.Y+1)*ch+4, rn.n(3), Paper)
	for i, g := range gauges {
		y := (gy+i*step)*ch + ch/2 - 4
		c.fill(gx*cw, y, (r.Max.X-1-gx)*cw, 8, g)
	}
	return c.img
}

// Gate is the gate card (its picture also covers the three rows above r): an arch over the card and one pillar bar per
// requirement, gw cells wide from column gx, step rows apart from row gy.
func (l Layout) Gate(r image.Rectangle, gx, gw, gy, step int, pillars []Bar) *image.RGBA {
	c := part(image.Rect(r.Min.X, r.Min.Y-3, r.Max.X, r.Max.Y), l.CW, l.CH)
	cw, ch := l.CW, l.CH
	c.cells(r, Card)
	aw := min(r.Dx(), 26) * cw
	ax := r.Min.X*cw + (r.Dx()*cw-aw)/2
	c.arch(ax, (r.Min.Y-3)*ch, aw, 3*ch, Brass)
	c.arch(ax+8, (r.Min.Y-3)*ch+8, aw-16, 3*ch-8, Sheet)
	c.rect(r.Min.X*cw, r.Min.Y*ch, r.Dx()*cw, 3, Brass)
	for i, g := range pillars {
		c.fill(gx*cw, (gy+i*step)*ch+ch/2-4, gw*cw, 8, g)
	}
	return c.img
}

// A WaitRow is one resource the plan waits on, as a gauge row.
type WaitRow struct{ Have, Need float64 }

// WaitsPic draws the gauges of the "waits on" box: gw cells wide from column
// gx, one per row from row gy, step rows apart, coloured by how much of the
// need is covered.
func (l Layout) WaitsPic(r image.Rectangle, gx, gw, gy, step int, w []WaitRow) *image.RGBA {
	c := part(r, l.CW, l.CH)
	cw, ch := l.CW, l.CH
	for i, v := range w {
		c.fill(gx*cw, (gy+i*step)*ch+ch/2-4, gw*cw, 8, Bar{Have: v.Have, Need: v.Need, Short: v.Have < v.Need})
	}
	return c.img
}

// An Entry is one line of the tray's list: its sign and whether it can be
// added.
type Entry struct {
	Sign string
	Dim  bool
}

// TabRect is tray tab i, a cell tall.
func (l Layout) TabRect(i int) image.Rectangle {
	x := l.Tray.Min.X + 1 + i*l.TabW
	return image.Rect(x, l.Tray.Min.Y, x+l.TabW-1, l.Tray.Min.Y+1)
}

// EntryRect is the cells of tray entry i: two columns, filled down the left
// one first, each row RowH tall under the search row.
func (l Layout) EntryRect(i int) (image.Rectangle, bool) {
	top := l.Tray.Min.Y + 2
	per := (l.Tray.Max.Y - top) / l.RowH
	if per <= 0 || i < 0 || i >= 2*per {
		return image.Rectangle{}, false
	}
	half := l.Tray.Dx() / 2
	x := l.Tray.Min.X + i/per*half
	y := top + i%per*l.RowH
	return image.Rect(x, y, x+half-1+i/per, y+l.RowH), true
}

// EntryCount is how many entries the tray's two columns hold.
func (l Layout) EntryCount() int {
	_, ok := l.EntryRect(0)
	if !ok {
		return 0
	}
	return 2 * ((l.Tray.Max.Y - l.Tray.Min.Y - 2) / l.RowH)
}

// TrayPic draws the tab strip, the search row and the entries. The selected
// tab is underlined in brass; each entry has a brass plus button at the
// right of its column.
func (l Layout) TrayPic(r image.Rectangle, signs []string, sel int, es []Entry) *image.RGBA {
	c := part(r, l.CW, l.CH)
	cw, ch := l.CW, l.CH
	c.cells(r, Card)
	for i, sg := range signs {
		t := l.TabRect(i)
		bg := Mix(Card, Wood[2], .25)
		if i == sel {
			bg = Paper
		}
		c.cells(t, bg)
		c.signSmall(sg, t.Min.X*cw, t.Min.Y*ch, 0, bg)
		if i == sel {
			c.rect(t.Min.X*cw, t.Max.Y*ch-3, t.Dx()*cw, 3, Brass)
		}
	}
	for i, e := range es {
		er, ok := l.EntryRect(i)
		if !ok {
			break
		}
		bg := Card
		if i%2 == 0 {
			bg = Mix(Card, Paper, .35)
		}
		c.cells(er, bg)
		rn := rng(seed(e.Sign))
		if l.RowH == 2 {
			c.sign(e.Sign, er.Min.X*cw+2, er.Min.Y*ch, rn.n(3), bg)
		} else {
			c.signSmall(e.Sign, er.Min.X*cw+2, er.Min.Y*ch, rn.n(3), bg)
		}
		col := Brass
		if e.Dim {
			col = Mix(Card, PDim, .5)
		}
		bx, by := (er.Max.X-2)*cw+rn.jit(), er.Min.Y*ch+l.RowH*ch/2-5
		c.rect(bx, by, 14, 10, col)
		c.rect(bx+6, by+2, 2, 6, PInk)
		c.rect(bx+3, by+4, 8, 2, PInk)
	}
	return c.img
}
