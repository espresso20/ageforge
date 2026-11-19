package cabinet

import "image"

// grid.go is the plain tray of a drawer with no legendary: wells in rows on
// the lining, a plate under each, and pages when they do not all fit.

// gridLabelW is the cells under a grid well: wide enough for a tally name.
const gridLabelW = 30

// Grid is what an open ageless drawer shows: Slot picks the lining and the
// frame, Held says which of its wells have a coin.
type Grid struct {
	Slot int
	Held []bool
}

// GridLayout is where a grid's wells sit. Rail is the open tray's rail.
type GridLayout struct {
	TrayLayout
	Per, Cols int
	Spots     []image.Point // centre of each well on the page, in pixels
	Plates    []image.Rectangle
	reach     int
}

// PlanGrid places up to n wells per page in a panel of cols by rows cells; it
// reports false when the panel is too small to hold one.
func PlanGrid(cols, rows, cw, ch int) (GridLayout, bool) {
	t, ok := PlanTray(cols, rows, cw, ch)
	if !ok {
		return GridLayout{}, false
	}
	reach := t.piece/2 + t.grow + 4*t.step + 1
	my := max(2, int(t.s*3))
	colW, rowH := max((gridLabelW+2)*cw, 2*reach)+2*my, 2*reach+3*my+ch
	g := GridLayout{TrayLayout: t, Cols: t.lining.Dx() / colW, reach: reach}
	rowsN := t.lining.Dy() / rowH
	if g.Cols < 1 || rowsN < 1 {
		return GridLayout{}, false
	}
	g.Per = g.Cols * rowsN
	x0 := t.lining.Min.X + (t.lining.Dx()-g.Cols*colW)/2
	y0 := t.lining.Min.Y + (t.lining.Dy()-rowsN*rowH)/2
	for i := 0; i < g.Per; i++ {
		c := image.Pt(x0+i%g.Cols*colW+colW/2, y0+i/g.Cols*rowH+reach+my)
		row := (c.Y + reach + my + ch - 1) / ch
		px := c.X/cw - gridLabelW/2
		g.Spots = append(g.Spots, c)
		g.Plates = append(g.Plates, image.Rect(px, row, px+gridLabelW, row+1))
	}
	return g, true
}

// At gives the well under a cell, or -1.
func (g GridLayout) At(cellX, cellY int) int {
	p := image.Pt(cellX*g.Cell.X+g.Cell.X/2, cellY*g.Cell.Y+g.Cell.Y/2)
	for i, c := range g.Spots {
		if p.In(image.Rect(c.X-g.reach, c.Y-g.reach, c.X+g.reach, c.Y+g.reach)) {
			return i
		}
	}
	return -1
}

func (g GridLayout) well(i int) image.Rectangle {
	c := g.Spots[i]
	return image.Rect(c.X-g.reach, c.Y-g.reach, c.X+g.reach, c.Y+g.reach)
}

// grid paints the frame, the lining, the first n wells and their plates.
func (c *canvas) grid(l GridLayout, gr Grid, n, sel int) {
	whole := c.img.Rect.Add(c.org)
	rp := c.frame(l.TrayLayout, gr.Slot)
	for i := 0; i < n && i < len(l.Spots); i++ {
		if l.well(i).Overlaps(whole) {
			p, seed := l.piece, gr.Slot*64+i+1
			m := shape(nil, p, []image.Point{{}}).grow(l.grow, seed)
			c.cut(m, l.Spots[i].X-p/2-l.grow, l.Spots[i].Y-p/2-l.grow, rp, sel == i, true, seed)
		}
	}
	c.rail(l.TrayLayout, gr.Slot)
	for i := 0; i < n && i < len(l.Plates); i++ {
		c.plate(l.pix(l.Plates[i]), i >= len(gr.Held) || !gr.Held[i], gr.Slot*20+5+i)
	}
}

// GridOpen paints the whole panel for an open grid with n wells, nothing selected.
func GridOpen(l GridLayout, g Grid, n int) *image.RGBA {
	c := newCanvas(image.Rect(0, 0, l.Cells.X*l.Cell.X, l.Cells.Y*l.Cell.Y), l.s)
	c.grid(l, g, n, -1)
	return c.img
}

// GridLift paints well i selected, over the cells it covers and one more each side.
func GridLift(l GridLayout, g Grid, n, i int) (cells image.Rectangle, img *image.RGBA) {
	r := l.well(i)
	cells = image.Rect(r.Min.X/l.Cell.X-1, r.Min.Y/l.Cell.Y-1, (r.Max.X+l.Cell.X-1)/l.Cell.X+1, (r.Max.Y+l.Cell.Y-1)/l.Cell.Y+1)
	c := newCanvas(l.pix(cells), l.s)
	c.grid(l, g, n, i)
	return cells, c.img
}
