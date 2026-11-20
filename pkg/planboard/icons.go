package planboard

import "image/color"

// Signs: one per kind, and one per building category for a build.
var kindCol = map[string]color.RGBA{"build": Hex(0x7fb0e8), "upgrade": Hex(0x7fd0a0), "research": Hex(0xc79be8),
	"advance": Hex(0xe8c46a), "expedition": Hex(0x6fd0d0), "deal": Hex(0xe89a6a), "wonder": Hex(0xe8e0f0)}

// kindColour is the sign's colour, from the part of its name before a dash.
func kindColour(sign string) color.RGBA {
	for i := 0; i < len(sign); i++ {
		if sign[i] == '-' {
			sign = sign[:i]
			break
		}
	}
	return kindCol[sign]
}

// sign draws a 21 by 30 px tile; j (0 to 2) varies the small details.
func (c *canvas) sign(kind string, x, y, j int, ground color.RGBA) {
	k := Mix(kindColour(kind), white, float64(j%3)*0.06)
	lt, bg := Mix(k, white, .45), Mix(ground, k, .16)
	c.rect(x, y, 21, 30, bg)
	switch kind {
	case "build":
		c.tri(x+10, y+5, x+2, y+14, x+18, y+14, lt)
		c.rect(x+4, y+14, 13, 11, k)
		c.rect(x+8+j%2, y+18, 4, 7, bg)
		c.rect(x+13, y+16, 2, 2, lt)
	case "build-p":
		c.rect(x+3, y+14, 15, 11, k)
		c.tri(x+3, y+14, x+3, y+8, x+10, y+14, lt)
		c.tri(x+10, y+14, x+10, y+8, x+17, y+14, lt)
		c.rect(x+14, y+4+j%2, 3, 8, lt)
		c.rect(x+6, y+18, 3, 3, bg)
	case "build-s":
		c.rect(x+3, y+9, 15, 16, k)
		c.rect(x+3, y+16, 15, 2, lt)
		c.rect(x+10, y+9, 2, 16, lt)
		c.rect(x+3, y+9, 15, 2, lt)
	case "build-r":
		c.rect(x+6, y+8, 9, 17, k)
		c.rect(x+6, y+5, 3, 3, lt)
		c.rect(x+12, y+5, 3, 3, lt)
		c.rect(x+9, y+12, 3, 6, bg)
		c.rect(x+4, y+25, 13, 2, lt)
	case "build-m":
		c.tri(x+3, y+7, x+17, y+7, x+10, y+25, k)
		c.rect(x+3, y+5, 15, 3, lt)
		c.rect(x+9, y+9, 3, 9, lt)
	case "upgrade":
		c.tri(x+10, y+4, x+2, y+15, x+18, y+15, lt)
		c.rect(x+7, y+15, 7, 10, k)
		c.rect(x+3, y+26, 15, 2, k)
	case "research":
		c.rect(x+7, y+4, 7, 2, lt)
		c.rect(x+8, y+6, 5, 6, k)
		c.tri(x+8, y+11, x+2, y+25, x+18, y+25, k)
		c.rect(x+5+j%2, y+20, 11, 4, lt)
	case "advance":
		c.rect(x+5, y+4, 2, 22, lt)
		c.tri(x+7, y+5, x+7, y+16, x+18, y+10, k)
		c.rect(x+3, y+25, 7, 2, k)
	case "expedition":
		c.tri(x+11, y+4, x+11, y+19, x+18, y+19, k)
		c.tri(x+9, y+8, x+9, y+19, x+3, y+19, lt)
		c.tri(x+2, y+21, x+19, y+21, x+15, y+26, k)
		c.rect(x+10, y+3, 1, 20, lt)
	case "deal":
		c.disc(x+8, y+11, 6, k)
		c.disc(x+8, y+11, 3, bg)
		c.disc(x+13, y+19, 6, lt)
		c.disc(x+13, y+19, 3, bg)
	case "wonder":
		c.tri(x+10, y+4, x+1, y+11, x+19, y+11, lt)
		for _, dx := range []int{3, 9, 15} {
			c.rect(x+dx, y+12, 3, 11, k)
		}
		c.rect(x+1, y+23, 19, 3, lt)
	}
}

// signSmall is the sign at half height, for a one-row list: every second pixel.
func (c *canvas) signSmall(kind string, x, y, j int, ground color.RGBA) {
	t := tile(c.cw, c.ch)
	t.sign(kind, 0, 0, j, ground)
	for yy := 0; yy < 15; yy++ {
		for xx := 0; xx < 11; xx++ {
			c.pt(x+xx, y+yy, t.img.RGBAAt(xx*2, yy*2))
		}
	}
}

// signBig is the sign at twice the size.
func (c *canvas) signBig(kind string, x, y, j int, ground color.RGBA) {
	t := tile(c.cw, c.ch)
	t.sign(kind, 0, 0, j, ground)
	for yy := 0; yy < 30; yy++ {
		for xx := 0; xx < 21; xx++ {
			c.rect(x+xx*2, y+yy*2, 2, 2, t.img.RGBAAt(xx, yy))
		}
	}
}
