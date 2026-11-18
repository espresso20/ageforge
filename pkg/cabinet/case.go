package cabinet

import "image"

// carcass paints the rails, stiles and the dark cut round every front.
func (c *canvas) carcass(l Layout) {
	w := l.Cells.X * l.Cell.X
	top, bot := l.cornice*l.Cell.Y, l.base*l.Cell.Y
	f := l.Fronts[0]
	gx, gy := l.gap.X*l.Cell.X, l.gap.Y*l.Cell.Y
	rail := []float64{0.05, -0.05, 0.08, -0.03, 0.04, -0.06}
	for j := 0; j <= Rows; j++ {
		y := top + j*(f.Dy()+gy)
		r := image.Rect(0, y, w, y+gy)
		c.board(r, 300+j*11, rail[j], false)
		c.edge(r, false)
	}
	stile := []float64{0.02, 0.06, -0.04, 0.05, -0.05, 0.02}
	for j := 0; j <= Cols; j++ {
		x0, x1 := 0, l.Fronts[0].Min.X
		if j > 0 {
			x0 = l.Fronts[j-1].Max.X
			x1 = x0 + gx
		}
		if j == Cols {
			x1 = w
		}
		r := image.Rect(x0, top, x1, bot)
		c.board(r, 400+j*13, stile[j], true)
		c.edge(r, false)
		c.edge(r.Inset(1), false)
	}
	cut := min(max(2, c.sc(4)), min(gx, gy)/3)
	for _, f := range l.Fronts {
		o := f.Inset(-cut)
		c.fill(o, black)
		c.fill(image.Rect(o.Min.X, o.Max.Y-1, o.Max.X, o.Max.Y), walnut[3])
		c.fill(image.Rect(o.Max.X-1, o.Min.Y, o.Max.X, o.Max.Y), walnut[2])
	}
}

// bands fills rows y0..y1 with walnut steps (index -1 is black, 9 a bead).
func (c *canvas) bands(w, y0, y1 int, idx []int) {
	for y := y0; y < y1; y++ {
		k := idx[(y-y0)*len(idx)/(y1-y0)]
		for x := 0; x < w; x++ {
			col := black
			if k == 9 {
				col = walnut[3+(x/max(1, c.sc(3)))%2]
			} else if k >= 0 {
				col = walnut[k]
			}
			c.px(x, y, col)
		}
	}
}

// cornice is a crown moulding with the totals plate on its frieze.
func (c *canvas) cornice(l Layout) {
	w, h := l.Cells.X*l.Cell.X, l.cornice*l.Cell.Y
	c.board(image.Rect(0, 0, w, h), 4, 0.04, false)
	c.bands(w, 0, max(4, c.sc(13)), []int{3, 4, 0, 1, 2, 4, 0})
	c.bands(w, h-max(5, c.sc(21)), h, []int{0, 3, 4, 2, 1, 9, 0, -1})
	c.plate(l.pix(l.Title), false, 9001)
}

// plinth is the step, the skirt with the hint plate, the shaped apron and
// two bracket feet standing on the floor.
func (c *canvas) plinth(l Layout) {
	w, h := l.Cells.X*l.Cell.X, l.Cells.Y*l.Cell.Y
	y0 := l.base * l.Cell.Y
	floor := hx(0x1c1e25)
	c.fill(image.Rect(0, y0, w, h), floor)
	st := max(4, c.sc(8))
	r := image.Rect(0, y0, w, y0+st)
	c.board(r, 51, 0.05, false)
	c.edge(r, false)
	c.fill(image.Rect(0, y0+st, w, y0+st+max(1, c.sc(2))), black)
	sy0 := y0 + st + max(1, c.sc(2))
	sy1 := min(h, y0+2*l.Cell.Y+max(4, c.sc(6)))
	r = image.Rect(0, sy0, w, sy1)
	c.board(r, 52, -0.04, false)
	c.edge(r, false)
	depth := max(2, c.sc(6))
	for x := w * 230 / 1920; x < w*1690/1920; x++ {
		dx := float64(x-w/2) / float64(w*730/1920)
		cut := sy1 - depth + int(dx*dx*float64(depth)+0.5)
		for y := cut; y < min(h, sy1+depth); y++ {
			col := floor
			if y < cut+max(1, c.sc(3)) {
				col = black
			}
			c.px(x, y, col)
		}
		c.px(x, cut-1, walnut[0])
		c.px(x, cut-2, walnut[4])
	}
	top, lit, front, low := hx(0x93694a), hx(0x7a5638), hx(0x58391f), hx(0x3a2616)
	fw, fh := c.sc(190), c.sc(26)
	for k, fx := range []int{c.sc(40), w - c.sc(40) - fw} {
		for y := sy1 - fh/2; y < min(h, sy1+fh/2); y++ {
			for x := fx; x < fx+fw; x++ {
				col := front
				switch {
				case y < sy1-fh*9/26:
					col = top
				case y > sy1+fh*9/26:
					col = low
				case (k == 0 && x-fx < max(2, c.sc(5))) || (k == 1 && fx+fw-1-x < max(2, c.sc(5))):
					col = lit
				}
				c.px(x, y, col)
			}
		}
		c.fill(image.Rect(fx+max(2, c.sc(4)), sy1+fh/2, fx+fw, sy1+fh/2+max(2, c.sc(8))), black)
	}
	c.plate(l.pix(l.Hint), false, 9002)
}
