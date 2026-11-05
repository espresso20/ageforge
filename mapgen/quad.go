package mapgen

import "sync"

// Quadrant blocks: a picture drawn at twice the grid's resolution each way.
// Every cell holds a 2 by 2 block of pixels, and a quadrant glyph (▘ ▝ ▀ ▖
// ▌ ▞ ▛ ▗ ▚ ▐ ▜ ▄ ▙ ▟ █) shows any pattern of them in two colours, a
// foreground and a background. Space draws its bodies this way: a planet is
// a small shaded disc, an orbit a thin line, not a letter.
//
// The canvas lives in mapgen, not beside the screen code, because the engine
// is what knows the picture (and what a pixel of it is, for the inspect line)
// and the screen layer only has to turn a cell's paints into colours. A
// resolved canvas is an ordinary Grid, so every theme, tier and test that
// handles a Grid handles this too.
//
// Geometry is whole numbers, like the noise: positions and sizes are in 16ths
// of a pixel. A pixel is half a cell wide and half a cell tall, so it is as
// tall as it is wide only on a cell that is twice as tall as it is wide;
// shapes below are drawn for that, and a disc is round on screen.

// pxUnit is one pixel in the 16ths the drawing calls use.
const pxUnit = 16

// quadGlyph is the glyph for a mask of the pixels that show the foreground:
// bit 0 the top left, 1 the top right, 2 the bottom left, 3 the bottom right.
var quadGlyph = [16]rune{
	' ', '▘', '▝', '▀',
	'▖', '▌', '▞', '▛',
	'▗', '▚', '▐', '▜',
	'▄', '▙', '▟', '█',
}

// QuadGlyph is the glyph that draws the pixels of mask in the foreground.
func QuadGlyph(mask int) rune { return quadGlyph[mask&15] }

type pixOwner struct {
	thing Thing
	prio  uint8
}

// Pixels is a canvas of 2w by 2h pixels over a w by h grid. A pixel with no
// hue is empty and shows whatever the cell under it has for a background.
type Pixels struct {
	W, H   int // in pixels
	px     []Paint
	own    []uint8
	owners []pixOwner // owners[0] is nobody
}

// NewPixels makes an empty canvas for a grid of w by h cells.
func NewPixels(w, h int) *Pixels {
	n := 4 * w * h
	p, _ := pixPool.Get().(*Pixels)
	if p == nil || cap(p.px) < n {
		p = &Pixels{px: make([]Paint, n), own: make([]uint8, n), owners: make([]pixOwner, 1, 16)}
	}
	p.W, p.H = 2*w, 2*h
	p.px, p.own, p.owners = p.px[:n], p.own[:n], p.owners[:1]
	clear(p.px)
	clear(p.own)
	return p
}

var pixPool sync.Pool

// Release hands the canvas back for the next one to reuse; it is not to be
// used afterwards.
func (p *Pixels) Release() { pixPool.Put(p) }

// Owner registers a thing the inspect line names, and returns the id to set
// pixels with. Where several things share a cell, the one with the highest
// prio is the one the cell reports.
func (p *Pixels) Owner(t Thing, prio int) uint8 {
	for i, o := range p.owners {
		if i > 0 && o.thing == t && o.prio == uint8(prio) {
			return uint8(i)
		}
	}
	if len(p.owners) >= 255 {
		return 0
	}
	p.owners = append(p.owners, pixOwner{t, uint8(clamp(prio, 0, 255))})
	return uint8(len(p.owners) - 1)
}

// Set paints a pixel; a paint with no hue or no weight leaves it as it was.
func (p *Pixels) Set(x, y int, c Paint, own uint8) {
	if x < 0 || y < 0 || x >= p.W || y >= p.H || c.Hue == HNone || c.Weight == 0 {
		return
	}
	i := y*p.W + x
	p.px[i], p.own[i] = c, own
}

// At is the paint of a pixel; off the canvas it is empty.
func (p *Pixels) At(x, y int) Paint {
	if x < 0 || y < 0 || x >= p.W || y >= p.H {
		return Paint{}
	}
	return p.px[y*p.W+x]
}

// Disc paints a round body centred at (cx, cy) with a radius of r, all in
// 16ths of a pixel and measured across (a disc r wide is r/2 tall in pixels).
// shade says what each pixel is made of from its place on the face: nx and ny
// run from -256 to 256 across it, y up the screen being negative; it returns
// an empty paint to leave the pixel alone. The pixel the centre falls in is
// always painted, so the smallest body is one pixel.
func (p *Pixels) Disc(cx, cy, r int, own uint8, shade func(nx, ny int) Paint) {
	r = max(r, 1)
	x0, x1 := (cx-r)/pxUnit-1, (cx+r)/pxUnit+1
	y0, y1 := (cy-r/2)/pxUnit-1, (cy+r/2)/pxUnit+1
	ccx, ccy := floorDiv(cx, pxUnit), floorDiv(cy, pxUnit)
	for y := max(y0, 0); y <= min(y1, p.H-1); y++ {
		dy := y*pxUnit + pxUnit/2 - cy
		for x := max(x0, 0); x <= min(x1, p.W-1); x++ {
			dx := x*pxUnit + pxUnit/2 - cx
			if dx*dx+4*dy*dy > r*r && !(x == ccx && y == ccy) {
				continue
			}
			p.Set(x, y, shade(clamp(dx*256/r, -256, 256), clamp(2*dy*256/r, -256, 256)), own)
		}
	}
}

// Glow lays soft rings of c round (cx, cy), from radius r0 out to r1 (16ths
// of a pixel, across): steps bands that fade from weight w0 toward nothing.
// It only paints pixels that are still empty, so draw it first.
func (p *Pixels) Glow(cx, cy, r0, r1 int, c Hue, w0, steps int, own uint8) {
	if steps < 1 || r1 <= r0 {
		return
	}
	x0, x1 := (cx-r1)/pxUnit-1, (cx+r1)/pxUnit+1
	y0, y1 := (cy-r1/2)/pxUnit-1, (cy+r1/2)/pxUnit+1
	for y := max(y0, 0); y <= min(y1, p.H-1); y++ {
		dy := y*pxUnit + pxUnit/2 - cy
		for x := max(x0, 0); x <= min(x1, p.W-1); x++ {
			dx := x*pxUnit + pxUnit/2 - cx
			d2 := dx*dx + 4*dy*dy
			if d2 > r1*r1 || p.px[y*p.W+x].Hue != HNone {
				continue
			}
			d := isqrt(d2)
			if d < r0 {
				d = r0
			}
			band := (d - r0) * steps / (r1 - r0)
			p.Set(x, y, P(c, w0*(steps-band)/(steps+1)), own)
		}
	}
}

// Ellipse paints the line of an ellipse of radii rx and ry (16ths of a
// pixel) round (cx, cy), one pixel thick, dashed where gap is more than zero
// (that many pixels off for every one on). half limits it to the lower half
// (1) or the upper half (-1) of the ellipse, 0 for all of it.
func (p *Pixels) Ellipse(cx, cy, rx, ry int, c Paint, own uint8, gap, half int) {
	n := max(24, (rx+ry)*2/pxUnit*3/2)
	n = min(n, 4096)
	lastX, lastY := -1<<30, -1<<30
	on := 0
	for i := 0; i < n; i++ {
		a := i * turn / n
		s := sin4096(a)
		if half > 0 && s < 0 || half < 0 && s > 0 {
			continue
		}
		x := floorDiv(cx+rx*cos4096(a)/1024, pxUnit)
		y := floorDiv(cy+ry*s/1024, pxUnit)
		if x == lastX && y == lastY {
			continue
		}
		lastX, lastY = x, y
		on++
		if gap > 0 && on%(gap+1) != 1 {
			continue
		}
		p.Set(x, y, c, own)
	}
}

// Resolve turns the canvas into cells of g: every cell with a pixel in it
// becomes a quadrant glyph in the two colours that stand for its four pixels,
// laid over the cell's own background wherever a pixel is empty. Cells with
// no pixel are left as they are, so what was drawn under the canvas (a star
// field) shows round the bodies. A cell also reports the thing of its most
// important pixel to the inspect line.
func (p *Pixels) Resolve(g *Grid) {
	var hit [256]uint16 // owner -> index in g.Things, 0 until first used
	for cy := 0; cy < p.H/2 && cy < g.H; cy++ {
		for cx := 0; cx < p.W/2 && cx < g.W; cx++ {
			i := 2*cy*p.W + 2*cx
			px := [4]Paint{p.px[i], p.px[i+1], p.px[i+p.W], p.px[i+p.W+1]}
			if px[0].Hue == HNone && px[1].Hue == HNone && px[2].Hue == HNone && px[3].Hue == HNone {
				continue
			}
			ci := cy*g.W + cx
			base := g.Cells[ci].Bg
			for k := range px {
				if px[k].Hue == HNone {
					px[k] = base
				}
			}
			r, fg, bg := reduceQuad(px, base)
			g.Cells[ci] = Cell{R: r, Fg: fg, Bg: bg}

			best, prio := uint8(0), -1
			for _, o := range [4]uint8{p.own[i], p.own[i+1], p.own[i+p.W], p.own[i+p.W+1]} {
				if o != 0 && int(p.owners[o].prio) > prio {
					best, prio = o, int(p.owners[o].prio)
				}
			}
			if best != 0 {
				if hit[best] == 0 {
					hit[best] = g.thingIndex(p.owners[best].thing)
				}
				g.Hit[ci] = hit[best]
			}
		}
	}
}

// reduceQuad picks the glyph and the two colours for four pixels, in the
// order top left, top right, bottom left, bottom right. base is the colour of
// the empty ground, which the glyph's background keeps when it can. The two
// colours that fill most of the block win; a pixel of any other colour joins
// whichever of the two it is nearer to, and a group of one hue shows as the
// average of its weights.
func reduceQuad(px [4]Paint, base Paint) (r rune, fg, bg Paint) {
	var cols [4]Paint
	var cnt [4]int
	n := 0
	for _, c := range px {
		k := 0
		for ; k < n && cols[k] != c; k++ {
		}
		if k == n {
			cols[n] = c
			n++
		}
		cnt[k]++
	}
	if n == 1 {
		return ' ', cols[0], cols[0]
	}
	// the two that win: most pixels, then not the ground, then the brighter
	better := func(a, b int) bool {
		if cnt[a] != cnt[b] {
			return cnt[a] > cnt[b]
		}
		if (cols[a] == base) != (cols[b] == base) {
			return cols[b] == base
		}
		return paintLuma(cols[a]) > paintLuma(cols[b])
	}
	a, b := 0, 1
	if better(b, a) {
		a, b = b, a
	}
	for k := 2; k < n; k++ {
		switch {
		case better(k, a):
			a, b = k, a
		case better(k, b):
			b = k
		}
	}
	// group every pixel with the nearer of the two
	var inA [4]bool
	sumA, nA, sumB, nB := 0, 0, 0, 0
	allA, allB := true, true
	for k, c := range px {
		toA := c == cols[a] || c != cols[b] && paintDist(c, cols[a]) <= paintDist(c, cols[b])
		inA[k] = toA
		if toA {
			sumA += int(c.Weight)
			nA++
			allA = allA && c.Hue == cols[a].Hue
		} else {
			sumB += int(c.Weight)
			nB++
			allB = allB && c.Hue == cols[b].Hue
		}
	}
	ca, cb := cols[a], cols[b]
	if allA && nA > 1 {
		ca = Paint{ca.Hue, uint8(sumA / nA)}
	}
	if allB && nB > 1 {
		cb = Paint{cb.Hue, uint8(sumB / nB)}
	}
	if nB == 0 {
		return ' ', ca, ca
	}
	// the foreground is the group that is not the ground, else the brighter
	aFront := paintLuma(ca) >= paintLuma(cb)
	if aGround, bGround := cols[a] == base, cols[b] == base; aGround != bGround {
		aFront = bGround
	}
	mask := 0
	for k := range px {
		if inA[k] == aFront {
			mask |= 1 << k
		}
	}
	if aFront {
		return quadGlyph[mask], ca, cb
	}
	return quadGlyph[mask], cb, ca
}

// spaceRGB is a rough colour for each hue that space draws, enough to say
// which of two paints is nearer to a third. The screen layer has the real
// colours, which depend on the theme; the order of them hardly does.
var spaceRGB = [NumHues][3]int{
	HSun: {255, 210, 122}, HCore: {255, 148, 64}, HWhite: {255, 255, 255},
	HRock: {143, 129, 114}, HRed: {200, 70, 60}, HAmber: {255, 179, 71},
	HAir: {79, 176, 255}, HSea: {11, 38, 72}, HLand: {27, 50, 38},
	HCloud: {70, 98, 138}, HPanel: {63, 120, 216}, HHydro: {127, 216, 142},
	HCool: {120, 200, 230}, HCyan: {120, 240, 255}, HOrbit: {76, 70, 118},
	HGold: {230, 190, 80}, HSand: {220, 190, 130}, HStar: {184, 198, 230},
	HNebula: {138, 70, 208}, HEngine: {120, 240, 255}, HSteel: {150, 160, 175},
	HVoid: {4, 6, 12}, HLife: {120, 220, 140}, HWater: {40, 90, 160},
}

func paintRGB(p Paint) (r, g, b int) {
	c := spaceRGB[p.Hue]
	if c == [3]int{} {
		c = [3]int{128, 128, 128}
	}
	w := int(p.Weight)
	return c[0] * w / 255, c[1] * w / 255, c[2] * w / 255
}

func paintLuma(p Paint) int {
	r, g, b := paintRGB(p)
	return 3*r + 6*g + b
}

func paintDist(a, b Paint) int {
	ar, ag, ab := paintRGB(a)
	br, bg, bb := paintRGB(b)
	return (ar-br)*(ar-br) + (ag-bg)*(ag-bg) + (ab-bb)*(ab-bb)
}
