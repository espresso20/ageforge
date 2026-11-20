package logroll

import (
	"image"
	"image/color"
	"math"
)

// Office numbers, as the game numbers them (1 to 10), less one.
const Offices = 10

// in reports whether the pictogram of office o (0 to 9) covers the point
// (x, y) of a 16 by 16 unit square.
func in(o int, x, y float64) bool {
	rc := func(a, b, c, d float64) bool { return x >= a && x <= c && y >= b && y <= d }
	ci := func(cx, cy, r float64) bool { return math.Hypot(x-cx, y-cy) <= r }
	ln := func(ax, ay, bx, by, w float64) bool { return seg(x, y, ax, ay, bx, by) <= w }
	switch o {
	case 0: // hammer
		return rc(2, 2, 12, 6.5) || rc(12, 3.2, 15.2, 5.4) || rc(6.2, 6.5, 8.8, 15)
	case 1: // lamp
		return ci(8, 10.2, 3.8) || rc(4.5, 13.5, 11.5, 15) || rc(6.8, 4.8, 9.2, 7) || (math.Abs(x-8)+0.9*math.Abs(y-3) <= 2.4 && y < 5)
	case 2: // dividers
		return ci(8, 2.6, 1.9) || ln(8, 3, 3.6, 15, .95) || ln(8, 3, 12.4, 15, .95) || rc(5.8, 8.6, 10.2, 9.6)
	case 3: // two figures
		return ci(4.8, 4.2, 2.3) || ci(11.2, 4.2, 2.3) || rc(2.4, 7.6, 7.2, 15) || rc(8.8, 7.6, 13.6, 15)
	case 4: // barrel
		w := 3.8 + 2*(1-math.Pow((y-8.5)/6.5, 2))
		return y >= 2 && y <= 15 && math.Abs(x-8) <= w && math.Abs(y-5.3) > .55 && math.Abs(y-11.7) > .55
	case 5: // scales
		return rc(7.4, 2, 8.6, 14) || rc(4, 13.5, 12, 15) || rc(2, 4, 14, 5.2) || (ci(4.5, 8, 3) && y > 8) || (ci(11.5, 8, 3) && y > 8)
	case 6: // banner
		return rc(3, 1, 4.2, 15) || (rc(4.2, 2, 14, 10) && !(x > 11 && math.Abs(y-6) < (x-11)*1.3))
	case 7: // shield
		top := rc(3, 2, 13, 8) || (y > 8 && math.Abs(x-8) <= 5*(15-y)/7+.5)
		return top && !(rc(5, 6, 11, 7.4) || rc(7.2, 4, 8.8, 12))
	case 8: // comet
		return ci(11, 11, 3.2) || ln(9, 9, 2, 3, .5) || ln(9, 10, 1, 8, .5) || ln(10, 9, 6, 1.5, .5)
	}
	t := math.Max(0, math.Min(1, ((x-13)*-8+(y-2)*9)/145)) // quill
	return seg(x, y, 13, 2, 5, 11) <= 2.6*math.Sin(math.Pi*t) || ln(14, 1, 2, 15, .4)
}

// face paints the pictogram of office o into the box x, y, w, h. The seed
// moves it by up to shift pixels and takes away about wear in 150 pixels.
func face(c canvas, o, x, y, w, h int, ink color.RGBA, r *rng, shift, wear int) {
	s := float64(min(w, h)) / 16
	n := int(16 * s)
	ox := x + (w-n)/2 + r.n(2*shift+1) - shift
	oy := y + (h-n)/2 + r.n(2*shift+1) - shift
	for py := 0; py < n; py++ {
		for px := 0; px < n; px++ {
			if in(o, (float64(px)+.5)/s, (float64(py)+.5)/s) && (wear == 0 || r.n(150) >= wear) {
				c.pt(ox+px, oy+py, ink)
			}
		}
	}
}

// Impression is the mark of office o on the roll (a w by h box on bg). Its
// shift, pressure and wear come from the seed, which the panel takes from the
// entry and not from the row it sits on.
func Impression(office int, seed uint64, w, h int, ink, bg color.RGBA) *image.RGBA {
	r := rng(seedOf(seed, uint64(office), 7))
	c := newCanvas(w, h, bg)
	p := mix(bg, ink, .72+.28*r.f()) // pressure
	face(c, office, 0, 0, w, h, p, &r, 1, r.n(14))
	return c.RGBA
}

// Tally draws n strokes (15 at most) in fives, each 2 px wide and 11 px tall,
// the fifth crossing the four before it. Spacing and lean come from the seed.
func Tally(n int, seed uint64, w, h int, ink, bg color.RGBA) *image.RGBA {
	r := rng(seedOf(seed, 12))
	c := newCanvas(w, h, bg)
	x := 1
	for n = min(n, 15); n > 0; n -= 5 {
		sh, y, lean := 11-r.n(2), (h-11)/2+r.n(2)-1, r.n(3)-1
		for i := 0; i < min(n, 4); i++ {
			c.rect(x+i*4, y, 2, sh, ink)
		}
		for i := 0; n >= 5 && i < 18; i++ {
			c.rect(x-2+i, y+sh-2-i*(sh-3+lean)/18, 1, 2, ink)
		}
		x += 20 + r.n(2)
	}
	return c.RGBA
}
