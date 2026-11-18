package cabinet

import (
	"image"
	"math"
)

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
