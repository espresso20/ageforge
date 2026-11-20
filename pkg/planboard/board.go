package planboard

import (
	"image"
	"image/color"
)

// The palette. The panel sets its words on these grounds.
var (
	Wood           = [3]color.RGBA{Hex(0xcfae76), Hex(0xb8985f), Hex(0x8f6f3f)}
	Sheet          = Hex(0x13233f)
	Sheet2         = Hex(0x1b2748)
	Grid           = Hex(0x192b4a)
	Band           = Hex(0x27446f)
	Ink, Dim       = Hex(0xdce6f5), Hex(0x8196b6)
	Brass          = Hex(0xc9a24a)
	Paper          = Hex(0xe9dfc3)
	Card           = Hex(0xb9b08f)
	PInk, PDim     = Hex(0x2a2418), Hex(0x5e5240)
	Ok, Wait       = Hex(0x5fc98a), Hex(0xe8b64a)
	Bad, Late      = Hex(0xd4605a), Hex(0xd08ad8)
	Steel, Grey    = Hex(0x6f9fd8), Hex(0x7d8794)
	Dark           = Hex(0x0e1a30)
	Ruler          = [2]color.RGBA{Hex(0x1d3357), Hex(0x284676)}
	Button, Strike = Hex(0x2b4a7a), Hex(0x4b3f2a)
	HintInk        = Hex(0x4a3a1e)
)

// Board is the static picture of a layout: the wood edge with its two clips,
// the sheet with its ruled border and register marks. It lies under
// everything, so it is cached per layout.
func (l Layout) Board() *image.RGBA {
	c := part(image.Rect(0, 0, l.Cols, l.Rows), l.CW, l.CH)
	w, h := l.Cols*l.CW, l.Rows*l.CH
	r := rng(l.Cols*1000 + l.Rows)
	c.rect(0, 0, w, h, Wood[0])
	for y := 0; y < h; y++ {
		if r.n(3) == 0 {
			continue
		}
		for x := r.n(40) - 20; x < w; {
			n := 10 + r.n(90)
			if r.n(4) > 0 {
				c.hline(x, y, n, Wood[1+r.n(3)/2])
			}
			x += n + 6 + r.n(60)
		}
	}
	c.hline(0, 0, w, Hex(0xe6cc98))
	c.hline(0, h-1, w, Wood[2])
	c.cells(l.Sheet, Sheet)
	c.frame(l.Sheet, Wood[2])
	c.bulldog(w * 46 / 100)
	c.flatClip(w * 566 / 1000)
	pale := Mix(Sheet, Ink, .35)
	x0, y0 := l.Sheet.Min.X*l.CW+4, l.Sheet.Min.Y*l.CH+4
	x1, y1 := l.Sheet.Max.X*l.CW-5, l.Sheet.Max.Y*l.CH-6
	c.hline(x0, y0, x1-x0, pale)
	c.hline(x0, y1, x1-x0, pale)
	c.vline(x0, y0, y1-y0, pale)
	c.vline(x1, y0, y1-y0+1, pale)
	for _, p := range [][2]int{{x0, y0}, {x1, y0}, {x0, y1}, {x1, y1}} {
		c.hline(p[0]-4, p[1], 9, pale)
		c.vline(p[0], p[1]-4, 9, pale)
		c.disc(p[0], p[1], 3, pale)
		c.disc(p[0], p[1], 2, Sheet)
	}
	return c.img
}

// bulldog is a steel bulldog clip with two wire loops, 63 by 45 px.
func (c *canvas) bulldog(x int) {
	st, hi, dk := Hex(0x8a9099), Hex(0xc4cad2), Hex(0x474c55)
	for _, wx := range []int{x + 9, x + 44} {
		c.rect(wx, 0, 2, 18, dk)
		c.rect(wx+8, 0, 2, 18, dk)
		c.rect(wx, 0, 10, 2, dk)
		c.rect(wx+2, 2, 1, 14, hi)
	}
	c.rect(x, 14, 63, 22, st)
	c.rect(x, 14, 63, 2, hi)
	c.rect(x, 34, 63, 2, dk)
	c.rect(x+6, 36, 51, 9, dk)
	c.rect(x+6, 36, 51, 2, st)
	for i := 0; i < 5; i++ {
		c.rect(x+14+i*9, 20, 2, 10, dk)
	}
}

// flatClip is a brass plate clip with a black spring leaf and two rivets.
func (c *canvas) flatClip(x int) {
	br, hi, dk := Hex(0xcaa43c), Hex(0xecd47e), Hex(0x6e5518)
	c.rect(x, 6, 63, 14, br)
	c.rect(x, 6, 63, 2, hi)
	c.rect(x, 18, 63, 2, dk)
	c.rect(x+3, 20, 57, 7, Hex(0x2c2a26))
	c.rect(x+3, 20, 57, 1, Hex(0x5a574e))
	c.rect(x+10, 27, 43, 14, br)
	c.rect(x+10, 27, 43, 2, hi)
	c.rect(x+10, 39, 43, 2, dk)
	c.disc(x+8, 12, 2, dk)
	c.disc(x+55, 12, 2, dk)
	c.rect(x+28, 10, 7, 3, dk)
}
