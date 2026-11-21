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

var smallWidths = []int{10, 12, 14, 14, 14, 14, 14, 14, 13, 12, 10, 8, 6, 4, 2}

func drawSmall(t Tier, d Domain, _ int) *image.RGBA {
	p := &pals[t]
	img := image.NewRGBA(image.Rect(0, 0, smallW, smallH))
	if t >= Platinum {
		w := newMask(3, 6)
		for y, row := range []string{"h..", "XX.", "XXX", "XXX", ".XX", "..X"} {
			for x := range row {
				if row[x] != '.' {
					w.set(x, y, tMain)
				}
			}
		}
		w.bevel()
		blit(img, w, sx0-2, sy0+4, p.metalTone, p.outline, false)
		blit(img, w, sx0+16-1, sy0+4, p.metalTone, p.outline, true)
	}
	in := func(x, y int) bool {
		r := y - 3
		if r < 0 || r >= len(smallWidths) {
			return false
		}
		w := smallWidths[r]
		return x >= 8-w/2 && x < 8+w/2
	}
	levels := []color.RGBA{p.field[0], p.field[1], p.field[1]}
	for y := 3; y < 3+len(smallWidths); y++ {
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
				c = levels[min(2, (x-1+y-3)/7)]
				if false {
					c = levels[min(2, (x-1+y-3)/7)-1]
				}
			}
			img.SetRGBA(sx0+x, sy0+y, c)
		}
	}
	// the emblem, centred in the field
	rows := smallEmblems[d]
	m := newMask(8, len(rows))
	for y, row := range rows {
		for x := range row {
			m.set(x+(8-len(row))/2, y, smallTone(row[x]))
		}
	}
	oy := 5 + (7-len(rows))/2
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
	if t >= Silver {
		r := newMask(12, 2)
		r.rect(0, 0, 12, 1, tRibHi)
		r.rect(0, 1, 12, 1, tRibLo)
		blit(img, r, sx0+2, sy0+14, p.metalTone, p.outline, false)
	}
	if t >= Gold {
		c := newMask(8, 4)
		for y, row := range []string{"h..hh..h", "XX.XX.XX", "XXXXXXXX", "XgXggXgX"} {
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
			c.set(3, 3, tGemHi)
			c.set(4, 3, tGemHi)
		}
		blit(img, c, sx0+4, sy0-1, p.metalTone, p.outline, false)
	}
	if t == Legendary {
		for _, q := range [][3]int{{-1, 2, 1}, {16, 2, -1}, {-1, 17, 1}, {16, 17, -1}} {
			put(img, sx0+q[0], sy0+q[1], p.ray)
			put(img, sx0+q[0]+q[2], sy0+q[1]-1, p.ray)
		}
		put(img, sx0+3, sy0+4, p.glint)
		put(img, sx0+4, sy0+3, p.glint)
	}
	return img
}
