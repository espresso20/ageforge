package cabinet

import (
	"image"
	"math"
)

// stud is a round brass head with one hard highlight; a missing one is a dark
// hole. Each head takes its own variation from the seed.
func (c *canvas) stud(x, y int, held bool, seed int) {
	d := max(4, c.sc(8))
	r := float64(d) / 2
	h := hash(x, y, seed)
	hl := d*3/8 + int(h*2) // highlight reach
	body := brass[3-int(h+0.2)]
	lip := float64(max(1, c.sc(1.5))) + 0.5
	for yy := 0; yy <= d; yy++ {
		for xx := 0; xx <= d; xx++ {
			dist := math.Hypot(float64(xx)-r+0.5, float64(yy)-r+0.5)
			in := dist <= r+0.5 && xx < d && yy < d
			switch {
			case in && held:
				col := body
				if xx+yy > d {
					col = brass[1]
				}
				if xx <= hl && yy <= hl && xx+yy <= hl+1 && xx+yy >= 2-int(h*1.5) {
					col = hx(0xd9c27a)
				}
				c.px(x+xx, y+yy, col)
			case in:
				c.px(x+xx, y+yy, black)
			case held && dist <= r+lip && xx+yy > d:
				c.px(x+xx, y+yy, black)
			case !held && dist <= r+lip && xx+yy > d:
				c.px(x+xx, y+yy, walnut[4])
			}
		}
	}
}

// studs sets the drawer's count of pieces as a row centred on cx, each stud
// a pixel off the line by its own seed.
func (c *canvas) studs(cx, cy, total, held, maxW, seed int) {
	d := max(4, c.sc(8))
	pitch := min(c.sc(13), (maxW-d)/max(total, 1))
	x0 := cx - ((total-1)*pitch+d)/2
	for i := 0; i < total; i++ {
		jy := int(hash(i, seed, 5) * 3)
		c.stud(x0+i*pitch, cy-d/2+jy-1, i < held, seed+i*31)
	}
}

// ring is a plain brass ring pull on a small mount.
func (c *canvas) ring(cx, cy, dia, seed int) {
	r := float64(dia) * 0.27
	t := math.Max(1.5, r*0.14*(0.85+0.4*hash(seed, 2, 2)))
	mw := max(3, int(r*0.42)) + int(hash(seed, 1, 2)*3)
	top, mh := cy-int(r*1.62), int(r*0.8)
	cut, shape := max(1, mw/2), int(hash(seed, 6, 2)*3)
	for y := top; y < top+mh; y++ {
		for x := cx - mw; x < cx+mw; x++ {
			dx, dy := x-cx, y-top-mh/2
			corner := (dx < cut-mw || dx >= mw-cut) && (y < top+cut || y >= top+mh-cut)
			if (shape == 1 && corner) || (shape == 2 && math.Hypot(float64(dx)+.5, float64(dy)+.5) > float64(mw)) {
				continue
			}
			col := brass[2]
			if dx+dy < -mw/2 {
				col = brass[4]
			} else if dx+dy > mw/2 {
				col = brass[1]
			}
			c.px(x, y, col)
		}
	}
	u := c.sc(1)
	cx, cy = cx+(int(hash(seed, 4, 2)*3)-1)*u, cy+int(hash(seed, 3, 2)*3)*u
	th := math.Pi/4 + (hash(seed, 5, 2)-0.5)*0.87
	cs, sn := math.Sqrt2*math.Cos(th), math.Sqrt2*math.Sin(th)
	sh := max(1, c.sc(4))
	for pass := 0; pass < 2; pass++ {
		for y := -int(r + t + 1); y <= int(r+t+1); y++ {
			for x := -int(r + t + 1); x <= int(r+t+1); x++ {
				d := math.Hypot(float64(x), float64(y))
				if math.Abs(d-r) > t {
					continue
				}
				if pass == 0 {
					c.px(cx+x+sh, cy+y+sh, black)
					continue
				}
				q := float64(x)*cs + float64(y)*sn
				col := tone(brass, 0.55-0.5*q/(2*r), cx+x, cy+y)
				if (d > r+t*0.45 && q < 0) || (d < r-t*0.45 && q > 0) {
					col = brass[4]
				}
				c.px(cx+x, cy+y, col)
			}
		}
	}
}

// disc paints a round cut of diameter dia: floor, a dark cut edge upper left
// and a light lip lower right. A gold cut is darker and ringed in brass.
func (c *canvas) disc(cx, cy, dia int, gold bool) {
	r := float64(dia) / 2
	cut := max(2, c.sc(3))
	fl := walnut[0]
	if gold {
		fl = hx(0x050405)
	}
	for y := -dia/2 - cut; y <= dia/2+cut; y++ {
		for x := -dia/2 - cut; x <= dia/2+cut; x++ {
			d := math.Hypot(float64(x), float64(y))
			shade := float64(-x-y) / math.Sqrt2
			switch {
			case d <= r-2 || (!gold && d <= r):
				col := fl
				if shade > 0 && r-d < float64(cut) {
					col = hx(0x0b0704)
					if gold {
						col = black
					}
				}
				c.px(cx+x, cy+y, col)
			case gold && d <= r:
				c.px(cx+x, cy+y, brass[4])
			case d <= r+float64(cut)*0.8 && x+y > 0:
				c.px(cx+x, cy+y, walnut[4])
			}
		}
	}
}

// outline lays the shape of a piece in a recess: its edge dim, its inner
// contours dimmer and its body one step above the floor.
func (c *canvas) outline(cx, cy int, im *image.RGBA) {
	n := im.Rect.Dx()
	solid := func(x, y int) bool { return image.Pt(x, y).In(im.Rect) && im.RGBAAt(x, y).A > 0 }
	lum := func(x, y int) float64 {
		p := im.RGBAAt(x, y)
		return (float64(p.R)*0.3 + float64(p.G)*0.59 + float64(p.B)*0.11) / 255
	}
	near := func(x, y int) bool { // within two pixels of the edge of the shape
		for dy := -2; dy <= 2; dy++ {
			for dx := -2; dx <= 2; dx++ {
				if !solid(x+dx, y+dy) {
					return true
				}
			}
		}
		return false
	}
	for y := 0; y < n; y++ {
		for x := 0; x < n; x++ {
			if !solid(x, y) {
				continue
			}
			col := hx(0x3a281a)
			switch {
			case !solid(x-1, y) || !solid(x+1, y) || !solid(x, y-1) || !solid(x, y+1):
				col = hx(0x84704a)
			case !near(x, y):
				l := lum(x, y)
				if math.Abs(l-lum(x+1, y)) > 0.12 || math.Abs(l-lum(x, y+1)) > 0.12 {
					col = hx(0x5e4a32)
				}
			}
			c.px(cx-n/2+x, cy-n/2+y, col)
		}
	}
}

// covered is the share of a piece's square that it paints.
func covered(im *image.RGBA) float64 {
	n := 0
	for i := 3; i < len(im.Pix); i += 4 {
		if im.Pix[i] > 0 {
			n++
		}
	}
	return float64(n) / float64(len(im.Pix)/4)
}

// lay draws the opaque pixels of im centred on cx, cy and moved by d; a
// shadow draws them all black, the hard cast shadow of the piece.
func (c *canvas) lay(im *image.RGBA, cx, cy, d int, shadow bool) {
	b := im.Rect
	for y := 0; y < b.Dy(); y++ {
		for x := 0; x < b.Dx(); x++ {
			p := im.RGBAAt(b.Min.X+x, b.Min.Y+y)
			if shadow && p.A > 0 {
				p = black
			}
			if p.A > 0 {
				c.px(cx-b.Dx()/2+x+d, cy-b.Dy()/2+y+d, p)
			}
		}
	}
}
