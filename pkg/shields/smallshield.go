package shields

import (
	"image"
	"image/color"
	"math"
)

// The small shield is 16 by 18 pixels of art, set in a 21 by 30 picture, and
// is not the large one shrunk: its rim is two pixels, its emblem is drawn on
// its own, and each tier adds one clear mark.

const (
	sx0, sy0 = 2, 6 // where the 16 by 18 art starts in the picture
)

// The shield is 16 wide and 16 high, rows 2 to 17 of the art; the crown and
// the wing ticks are one or two pixels each, so the emblem keeps the field.
var smallWidths = []int{14, 16, 16, 16, 16, 16, 16, 16, 16, 16, 14, 12, 10, 8, 4, 2}

func drawSmall(t Tier, d Domain, _ int) *image.RGBA {
	p := &pals[t]
	img := image.NewRGBA(image.Rect(0, 0, smallW, smallH))
	if t >= Platinum {
		w := newMask(1, 4)
		for y, c := range []uint8{tLight, tMain, tMain, tShade} {
			w.set(0, y, c)
		}
		blit(img, w, sx0-1, sy0+7, p.metalTone, p.outline, false)
		blit(img, w, sx0+16, sy0+7, p.metalTone, p.outline, false)
	}
	in := func(x, y int) bool {
		r := y - 2
		if r < 0 || r >= len(smallWidths) {
			return false
		}
		w := smallWidths[r]
		return x >= 8-w/2 && x < 8+w/2
	}
	levels := []color.RGBA{p.field[0], p.field[1], p.field[1]}
	for y := 2; y < 2+len(smallWidths); y++ {
		for x := 0; x < 16; x++ {
			if !in(x, y) {
				continue
			}
			best, ox, oy := 99.0, 0, 0
			for dy := -3; dy <= 3; dy++ {
				for dx := -3; dx <= 3; dx++ {
					if dd := math.Hypot(float64(dx), float64(dy)); dd < best && !in(x+dx, y+dy) {
						best, ox, oy = dd, dx, dy
					}
				}
			}
			var c color.RGBA
			switch band := int(best) - 1; {
			case band == 0:
				c = p.outline
			case band == 1:
				face := -float64(ox+oy) / (math.Hypot(float64(ox), float64(oy)) * math.Sqrt2)
				switch {
				case face > .3:
					c = p.metal[0]
				case face < -.3:
					c = p.metal[2]
				default:
					c = p.metal[1]
				}
			default:
				c = levels[min(2, (x-1+y-2)/7)]
			}
			img.SetRGBA(sx0+x, sy0+y, c)
		}
	}
	// the emblem, centred in the field
	rows := smallEmblems[d]
	oy := 5 + (8-len(rows))/2
	m := newMask(8, len(rows))
	for y, row := range rows {
		for x := range row {
			m.set(x, y, smallTone(row[x]))
		}
	}
	for y := 0; y < m.h; y++ {
		for x := 0; x < m.w; x++ {
			if tt := m.get(x, y); tt != 0 {
				c := p.emTone(tt)
				if tt == tNoBevl {
					c = p.em[1]
					if m.get(x, y-1) == 0 || m.get(x-1, y) == 0 {
						c = p.em[0]
					}
				}
				put(img, sx0+4+x, sy0+oy+y, c)
			}
		}
	}
	if t >= Silver { // rivets in the upper corners of the field
		put(img, sx0+3, sy0+4, p.metal[0])
		put(img, sx0+12, sy0+4, p.metal[0])
	}
	if t >= Gold { // a three point crown on the rim
		c := newMask(6, 2)
		for y, row := range []string{"h.hh.h", "XXggXX"} {
			for x := range row {
				switch row[x] {
				case 'X':
					c.set(x, y, tMain)
				case 'h':
					c.set(x, y, tLight)
				case 'g':
					c.set(x, y, tGem)
				}
			}
		}
		if t >= Platinum {
			c.set(2, 1, tGemHi)
			c.set(3, 1, tGemHi)
		}
		blit(img, c, sx0+5, sy0, p.metalTone, p.outline, false)
	}
	if t == Legendary {
		// a sunburst: rays over the crown, under the point, one each side
		for _, r := range [][2]int{
			{7, -2}, {7, -3}, {7, -4}, {8, -2}, {8, -3}, {8, -4},
			{3, -1}, {2, -2}, {1, -3}, {0, -4}, {12, -1}, {13, -2}, {14, -3}, {15, -4},
			{7, 19}, {7, 20}, {7, 21}, {8, 19}, {8, 20}, {8, 21},
			{5, 18}, {4, 19}, {3, 20}, {2, 21}, {10, 18}, {11, 19}, {12, 20}, {13, 21},
		} {
			put(img, sx0+r[0], sy0+r[1], p.ray)
		}
		put(img, sx0+3, sy0+5, p.glint)
		put(img, sx0+4, sy0+4, p.glint)
	}
	return img
}
