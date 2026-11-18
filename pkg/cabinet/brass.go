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
	body := brass[3]
	if h > 0.8 {
		body = brass[2]
	}
	lip := max(1, c.sc(1.5))
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
				col := black
				if xx+yy < 3 {
					col = hx(0)
				}
				c.px(x+xx, y+yy, col)
			case held && dist <= r+0.5+float64(lip) && xx+yy > d:
				c.px(x+xx, y+yy, black)
			case !held && dist <= r+0.5+float64(lip) && xx+yy > d:
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
	t := math.Max(1.5, r*0.14)
	mw := max(2, int(r*0.35)) + int(hash(seed, 1, 2)*2)
	top := cy - int(r) - int(r*0.38)
	c.fill(image.Rect(cx-mw, top, cx+mw, top+int(r*0.7)), brass[2])
	c.fill(image.Rect(cx-mw, top, cx+mw, top+1), brass[4])
	c.fill(image.Rect(cx+mw-1, top, cx+mw, top+int(r*0.7)), brass[1])
	reach := int(r + t + 1)
	for pass := 0; pass < 2; pass++ {
		for y := -reach; y <= reach; y++ {
			for x := -reach; x <= reach; x++ {
				d := math.Hypot(float64(x), float64(y))
				if math.Abs(d-r) > t {
					continue
				}
				if pass == 0 {
					sh := max(1, c.sc(4))
					c.px(cx+x+sh, cy+y+sh, black)
					continue
				}
				col := tone(brass, 0.55-0.5*float64(x+y)/(2*r), cx+x, cy+y)
				if (d > r+t*0.45 && x+y < 0) || (d < r-t*0.45 && x+y > 0) {
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
				c.px(cx+x, cy+y, brass[4]) // the gold line, all the way round
			case d <= r+float64(cut)*0.8 && x+y > 0:
				c.px(cx+x, cy+y, walnut[4])
			}
		}
	}
}

// outline lays the shape of a piece in a recess: its edge dim, its main inner
// contour dimmer and its body one step above the floor.
func (c *canvas) outline(cx, cy int, im *image.RGBA) {
	n := im.Rect.Dx()
	solid := func(x, y int) bool { return image.Pt(x, y).In(im.Rect) && im.RGBAAt(x, y).A > 0 }
	lum := func(x, y int) float64 {
		p := im.RGBAAt(x, y)
		return (float64(p.R)*0.3 + float64(p.G)*0.59 + float64(p.B)*0.11) / 255
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
			case solid(x-2, y) && solid(x+2, y) && solid(x, y-2) && solid(x, y+2) && x+2 < n && y+2 < n:
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

// blit lays the opaque pixels of im with its centre at cx, cy.
func (c *canvas) blit(im *image.RGBA, cx, cy int) {
	b := im.Rect
	for y := 0; y < b.Dy(); y++ {
		for x := 0; x < b.Dx(); x++ {
			if p := im.RGBAAt(b.Min.X+x, b.Min.Y+y); p.A > 0 {
				c.px(cx-b.Dx()/2+x, cy-b.Dy()/2+y, p)
			}
		}
	}
}

// cast draws the hard shadow of im, offset down and right.
func (c *canvas) cast(im *image.RGBA, cx, cy, d int) {
	b := im.Rect
	for y := 0; y < b.Dy(); y++ {
		for x := 0; x < b.Dx(); x++ {
			if im.RGBAAt(b.Min.X+x, b.Min.Y+y).A > 0 {
				c.px(cx-b.Dx()/2+x+d, cy-b.Dy()/2+y+d, black)
			}
		}
	}
}

// seal is a wax seal whose outline, lobes and rim vary with the seed.
func (c *canvas) seal(cx, cy, seed int) {
	r := max(5, c.sc(13))
	lobes := float64(6 + seed%3)
	amp := (1.2 + 0.8*hash(seed, 3, 7)) * c.s
	phase := hash(seed, 4, 7) * math.Pi
	ring := 4 + hash(seed, 5, 7)*2
	sh := max(1, c.sc(3))
	for pass := 0; pass < 2; pass++ {
		for y := -r - 3; y <= r+3; y++ {
			for x := -r - 3; x <= r+3; x++ {
				d := math.Hypot(float64(x), float64(y))
				rr := float64(r) + amp*math.Sin(lobes*math.Atan2(float64(y), float64(x))+phase)
				if d > rr {
					continue
				}
				if pass == 0 {
					c.px(cx+x+sh, cy+y+sh, black)
					continue
				}
				col := tone(wax, 0.55-0.45*float64(x+y)/float64(2*r), cx+x, cy+y)
				if d > rr-math.Max(1, 2.2*c.s) {
					col = wax[1]
					if x+y < 0 {
						col = wax[4]
					}
				}
				if d < ring*float64(r)/13+1 && d > ring*float64(r)/13-1.5 {
					col = wax[1]
				}
				c.px(cx+x, cy+y, col)
			}
		}
	}
}

// plate paints the brass round the cell box face: a hard shadow, the frame,
// four pins by the seed, and the flat face the caller writes words on.
func (c *canvas) plate(face image.Rectangle, dim bool, seed int) {
	mx, my := max(4, c.sc(8)), max(2, c.sc(3))
	o := image.Rect(face.Min.X-mx, face.Min.Y-my, face.Max.X+mx, face.Max.Y+my)
	lit, dk, pin := brass[4], brass[0], brass[3]
	if dim {
		lit, pin = brass[3], brass[2]
	}
	sh := max(1, c.sc(3))
	c.fill(o.Add(image.Pt(sh, sh)), black)
	c.fill(o, PlateFace(dim))
	for x := o.Min.X; x < o.Max.X; x++ {
		c.px(x, o.Min.Y, lit)
		c.px(x, o.Max.Y-1, dk)
	}
	for y := o.Min.Y; y < o.Max.Y; y++ {
		c.px(o.Min.X, y, lit)
		c.px(o.Max.X-1, y, dk)
	}
	ps := max(1, c.sc(2))
	jit := func(k int, span int) int { return int(hash(seed, k, 11) * float64(span+1)) }
	for k := 0; k < 4; k++ {
		x := o.Min.X + 1 + jit(k, mx-ps-3)
		if k&1 == 1 {
			x = o.Max.X - 1 - ps - jit(k, mx-ps-3)
		}
		y := o.Min.Y + 1 + jit(k+4, 1)
		if k&2 == 2 {
			y = o.Max.Y - 1 - ps - jit(k+4, 1)
		}
		col := pin
		if hash(seed, k, 13) > 0.7 {
			col = brass[4]
		}
		c.fill(image.Rect(x, y, x+ps, y+ps), col)
	}
	c.fill(face, PlateFace(dim))
}
