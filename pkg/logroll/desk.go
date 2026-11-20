package logroll

import (
	"image"
	"image/color"
)

var BrassD = hx(0x7d5a12)

// Desk paints everything that does not change while the log does: the wood,
// the roll's paper between two rollers with its column rules, the beds of the
// rail, the rack and the lever, and the slate with its frame, cord and ledge.
// It is one picture per layout; cache it by Layout.Key.
func Desk(l Layout, cw, ch int, seed uint64) *image.RGBA {
	r := rng(seedOf(seed, uint64(l.Cols), uint64(l.Rows), 1))
	c := newCanvas(l.Cols*cw, l.Rows*ch, wood[0])
	cells := func(q Rect, col color.RGBA) { c.rect(q.X*cw, q.Y*ch, q.W*cw, q.H*ch, col) }
	if l.Mode == Placeholder {
		return c.RGBA
	}
	for y := 0; y < c.Rect.Dy(); y++ {
		c.hline(0, y, c.Rect.Dx(), wood[y/(2*ch)%4])
	}
	for i := 0; i < l.Cols*l.Rows/3; i++ { // runs of darker grain, none on a period
		c.hline(r.n(c.Rect.Dx()), r.n(c.Rect.Dy()), 3+r.n(18), mix(wood[r.n(4)], hx(0x120c05), .3+r.f()*.25))
	}
	cells(l.Title, Strip)
	cells(l.Head, Strip)
	cells(l.Rail, Strip)
	cells(l.Rack, Plate)
	cells(l.Lever, Plate)

	// the roll: paper with a deckle edge and a roller at each end
	x0, x1 := l.Roll.X*cw, l.Roll.Right()*cw
	top, bot := (l.Roll.Y+1)*ch, (l.Roll.Bottom()-1)*ch
	c.rect(x0, top, x1-x0, bot-top, Paper)
	for y := top; y < bot; y++ {
		c.hline(x0-1-r.n(3), y, 1, hx(0x241a0e))
		c.hline(x1+r.n(3), y, 1, hx(0x241a0e))
		c.hline(x0, y, r.n(2), mix(Paper, Band, .7))
		c.hline(x1-1-r.n(2), y, 1+r.n(2), mix(Paper, Band, .7))
	}
	for _, y := range []int{l.Roll.Y * ch, (l.Roll.Bottom() - 1) * ch} {
		for i, col := range [4]color.RGBA{hx(0xe0c27f), hx(0xb98f4e), hx(0x8f6a36), hx(0x5e441f)} {
			c.rect(x0, y+i*ch/4, x1-x0, (i+1)*ch/4-i*ch/4, col)
		}
		for _, cx := range []int{x0 - 6, x1 - 6} {
			c.rect(cx, y, 12, ch, hx(0x4a3318))
			k := r.n(2)
			c.rect(cx+5-k, y+3, 2+k, ch-6, hx(0x6e5420))
		}
	}
	for _, q := range []Rect{l.Stamp, l.Text, l.Paid, l.Gained, l.Times} {
		if q.W > 0 {
			c.vline((q.X-1)*cw+cw/2, top, bot-top, Band)
		}
	}

	// the slate: wood frame, slate face, ledge, cord
	s := l.Slate
	cells(s, hx(0x6a4d28))
	c.rect(s.X*cw, s.Y*ch, s.W*cw, 2, hx(0x8a6a3c))
	c.rect(s.X*cw, s.Y*ch, 2, s.H*ch, hx(0x8a6a3c))
	cells(l.SlateBody, Slate)
	cells(l.Ledge, hx(0x7a5a30))
	c.hline(l.Ledge.X*cw, l.Ledge.Y*ch, l.Ledge.W*cw, hx(0x2c1f0e))
	ledge(c, l.Ledge, cw, ch, &r)
	nx, ny := (s.X+s.W/2)*cw, ch/2
	for _, ax := range []int{s.X*cw + s.W*cw/5, s.Right()*cw - s.W*cw/5} {
		cord(c, ax, s.Y*ch+3, nx, ny, hx(0xb89c5e), hx(0x6e5a2e))
	}
	c.disc(nx, ny, 4, hx(0x2c2c2c))
	c.disc(nx-1, ny-1, 1, hx(0x8a8a8a))
	return c.RGBA
}

// cord is a twisted rope 2 px wide from (x0, y0) to (x1, y1).
func cord(c canvas, x0, y0, x1, y1 int, a, b color.RGBA) {
	n := max(abs(x1-x0), abs(y1-y0))
	for i := 0; i <= n; i++ {
		x, y := x0+(x1-x0)*i/max(n, 1), y0+(y1-y0)*i/max(n, 1)
		col := a
		if i/3%2 == 1 {
			col = b
		}
		c.rect(x, y, 2, 2, col)
	}
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// ledge puts a chalk stick (56 by 10 px), a folded rag (90 by 28 px) and a
// little chalk dust on the slate's ledge.
func ledge(c canvas, q Rect, cw, ch int, r *rng) {
	x, y := q.X*cw, q.Y*ch
	w, h := q.W*cw, q.H*ch
	sx, sy := x+16+r.n(24), y+h/2-5+r.n(3)
	c.rect(sx, sy, 56, 10, Chalk)
	c.rect(sx, sy+8, 56, 2, ChalkD)
	rx, ry := x+w-130-r.n(40), y+h/2-14+r.n(3)
	c.rect(rx, ry, 90, 28, hx(0xbfb49a))
	for i := 1; i <= 3; i++ { // three folds
		fy := ry + i*7
		for xx := 0; xx < 90; xx++ {
			c.pt(rx+xx, fy+(xx*(i+1)/23)%2, hx(0x8e836a))
		}
	}
	for k := 0; k < 14; k++ { // ragged edge
		c.rect(rx+r.n(90), ry+r.n(2)*27, 1+r.n(3), 1, hx(0x7a5a30))
	}
	c.rect(rx+10, ry+4, 38, 3, mix(hx(0xbfb49a), Chalk, .55)) // chalk on the cloth
	for k := 0; k < 3+r.n(3); k++ {
		c.rect(x+60+r.n(w-140), y+3+r.n(h-6), 2+r.n(4), 1+r.n(2), mix(hx(0x7a5a30), Chalk, .5))
	}
}
