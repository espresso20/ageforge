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
	t := math.Max(1.5, r*0.14)
	mw := max(2, int(r*0.35)) + int(hash(seed, 1, 2)*2)
	top := cy - int(r) - int(r*0.38)
	mh := int(r * 0.7)
	c.fill(image.Rect(cx-mw, top, cx+mw, top+mh), brass[2])
	c.fill(image.Rect(cx-mw, top, cx+mw, top+1), brass[4])
	c.fill(image.Rect(cx+mw-1, top, cx+mw, top+mh), brass[1])
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

// outline lays the shape of a piece in a recess: its edge dim and its body
// one step above the floor.
func (c *canvas) outline(cx, cy int, im *image.RGBA) {
	n := im.Rect.Dx()
	solid := func(x, y int) bool { return image.Pt(x, y).In(im.Rect) && im.RGBAAt(x, y).A > 0 }
	for y := 0; y < n; y++ {
		for x := 0; x < n; x++ {
			if !solid(x, y) {
				continue
			}
			col := hx(0x3a281a)
			if !solid(x-1, y) || !solid(x+1, y) || !solid(x, y-1) || !solid(x, y+1) {
				col = hx(0x84704a)
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
	c.fill(image.Rect(o.Min.X, o.Min.Y, o.Max.X, o.Min.Y+1), lit)
	c.fill(image.Rect(o.Min.X, o.Max.Y-1, o.Max.X, o.Max.Y), dk)
	c.fill(image.Rect(o.Min.X, o.Min.Y, o.Min.X+1, o.Max.Y), lit)
	c.fill(image.Rect(o.Max.X-1, o.Min.Y, o.Max.X, o.Max.Y), dk)
	ps := max(1, c.sc(2))
	jit := func(k, span int) int { return int(hash(seed, k, 11) * float64(span+1)) }
	for k := 0; k < 4; k++ {
		x, y := o.Min.X+1+jit(k, mx-ps-3), o.Min.Y+1+jit(k+4, 1)
		if k&1 == 1 {
			x = o.Max.X - 1 - ps - jit(k, mx-ps-3)
		}
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
