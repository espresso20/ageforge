// Package shields draws the badge shields in code, as pixel art: crisp one
// pixel outlines, flat fills, and shading by bands and dither. No image files.
//
// A shield is a picture of 70 by 75 pixels (the large one, ten cells by five)
// or 21 by 30 (the small one, three cells by two, 16 by 18 of art in it).
// The window draws a picture over whole cells, so Draw paints one at a cell.
package shields

import (
	"image"
	"image/color"
	"math"
	"sync"

	"github.com/espresso20/ageforge/window/kit"
)

// Tier is how far a badge has been taken.
type Tier int

const (
	Bronze Tier = iota
	Silver
	Gold
	Platinum
	Legendary
	numTiers
)

var tierNames = [numTiers]string{"Bronze", "Silver", "Gold", "Platinum", "Legendary"}

func (t Tier) String() string { return tierNames[t] }

// Size is which shield to draw: the large one, the small one for list rows,
// or the large one at twice the scale for the earn card.
type Size int

const (
	Large Size = iota
	Small
	Huge
)

// The window's cell, in pixels, which a picture is sized to.
const (
	cellW, cellH   = 7, 15
	largeW, largeH = 70, 75
	smallW, smallH = 21, 30
)

// Frames is how many frames a tier's shield has: only the legendary moves.
func Frames(t Tier) int {
	if t == Legendary {
		return 6
	}
	return 1
}

// haloPad is the room a legendary shield's rays take on each side, in pixels.
const haloPadX, haloPadY = 14, 15

// Cells is the cells a shield takes; a legendary large one is wider and taller
// to leave room for its rays.
func Cells(t Tier, s Size) (cols, rows int) {
	if s == Small {
		return smallW / cellW, smallH / cellH
	}
	cols, rows = largeW/cellW, largeH/cellH
	if t == Legendary {
		cols, rows = cols+2*haloPadX/cellW, rows+2*haloPadY/cellH
	}
	if s == Huge {
		cols, rows = cols*2, rows*2
	}
	return
}

// Draw paints the shield with its top left at cell (x, y).
func Draw(p *kit.Painter, x, y int, t Tier, d Domain, frame int, s Size) {
	cols, rows := Cells(t, s)
	p.Picture(image.Rect(x, y, x+cols, y+rows), Image(t, d, frame, s))
}

type key struct {
	t Tier
	d Domain
	f int
	s Size
}

var cache struct {
	sync.Mutex
	m map[key]*image.RGBA
}

// Image is the shield as a picture; the same picture is given every time.
func Image(t Tier, d Domain, frame int, s Size) *image.RGBA {
	frame = ((frame % Frames(t)) + Frames(t)) % Frames(t)
	k := key{t, d, frame, s}
	var big *image.RGBA
	if s == Huge {
		big = Image(t, d, frame, Large)
	}
	cache.Lock()
	defer cache.Unlock()
	if img, ok := cache.m[k]; ok {
		return img
	}
	if cache.m == nil {
		cache.m = map[key]*image.RGBA{}
	}
	var img *image.RGBA
	switch s {
	case Small:
		img = drawSmall(t, d, frame)
	case Huge:
		img = scale2(big)
	default:
		img = drawLarge(t, d, frame)
	}
	cache.m[k] = img
	return img
}

func scale2(a *image.RGBA) *image.RGBA {
	b := a.Bounds()
	out := image.NewRGBA(image.Rect(0, 0, 2*b.Dx(), 2*b.Dy()))
	for y := 0; y < out.Rect.Dy(); y++ {
		for x := 0; x < out.Rect.Dx(); x++ {
			out.SetRGBA(x, y, a.RGBAAt(x/2, y/2))
		}
	}
	return out
}

func put(img *image.RGBA, x, y int, c color.RGBA) {
	if image.Pt(x, y).In(img.Rect) {
		img.SetRGBA(x, y, c)
	}
}

func mix(a, b color.RGBA, t float64) color.RGBA {
	m := func(x, y uint8) uint8 { return uint8(float64(x)*(1-t) + float64(y)*t + .5) }
	return color.RGBA{m(a.R, b.R), m(a.G, b.G), m(a.B, b.B), 255}
}

// blit sets the mask's tones as colours at (ox, oy) and gives the shape a one
// pixel outline all round.
func blit(img *image.RGBA, m *mask, ox, oy int, tone func(uint8) color.RGBA, outline color.RGBA, flip bool) {
	at := func(x, y int) uint8 {
		if flip {
			x = m.w - 1 - x
		}
		return m.get(x, y)
	}
	for y := -1; y <= m.h; y++ {
		for x := -1; x <= m.w; x++ {
			switch t := at(x, y); {
			case t != 0:
				put(img, ox+x, oy+y, tone(t))
			case at(x-1, y) != 0 || at(x+1, y) != 0 || at(x, y-1) != 0 || at(x, y+1) != 0:
				put(img, ox+x, oy+y, outline)
			}
		}
	}
}

// ---- the large shield ----

const (
	bx0, by0 = 13, 14 // the shield's top left
	bw, bh   = 44, 57
	midX     = 35 // the pixel edge the shield is mirrored about
)

// bodyWidth is the width of row r of the shield: straight to the shoulders
// and then a curve that closes to a point.
func bodyWidth(r int) int {
	switch {
	case r == 0:
		return 38
	case r == 1:
		return 42
	case r < 27:
		return bw
	}
	t := (float64(r-27) + .5) / float64(bh-27)
	w := int(math.Round(float64(bw)*math.Pow(1-math.Pow(t, 1.75), .62)/2)) * 2
	return max(w, 2)
}

// rim lists, from the edge in, what each pixel of the rim is: o outline,
// b bevel (light, mid or shade by the way the edge faces), g groove,
// i inner line, l light line.
var rimSpec = [numTiers]string{"obbgi", "obbgbbi", "obbgbbil", "oblbgbbi", "obbgbbil"}

type body struct {
	band [largeH][largeW]int // pixels in from the edge, -1 outside the shield
}

func paintBody(img *image.RGBA, t Tier, p *pal) *body {
	var b body
	in := func(x, y int) bool {
		r := y - by0
		if r < 0 || r >= bh {
			return false
		}
		w := bodyWidth(r)
		return x >= midX-w/2 && x < midX+w/2
	}
	levels, thr := []color.RGBA{p.field[1], p.field[2], p.field[3]}, []int{20, 68}
	if t >= Gold {
		levels, thr = p.field[:], []int{12, 32, 60, 82}
	}
	spec := rimSpec[t]
	for y := 0; y < largeH; y++ {
		for x := 0; x < largeW; x++ {
			b.band[y][x] = -1
			if !in(x, y) {
				continue
			}
			// the nearest pixel outside the shield
			best, ox, oy := 99.0, 0, 0
			for dy := -9; dy <= 9; dy++ {
				for dx := -9; dx <= 9; dx++ {
					if d := math.Hypot(float64(dx), float64(dy)); d < best && !in(x+dx, y+dy) {
						best, ox, oy = d, dx, dy
					}
				}
			}
			band := int(best) - 1
			b.band[y][x] = band
			if band >= len(spec) {
				s := (x - bx0) + (y-by0)*9/10
				i := 0
				for k, th := range thr {
					if s >= th {
						i = k + 1
					} else if s >= th-2 && (x+y)&1 == 0 {
						i = k + 1
					}
					if s >= th && s < th+2 && (x+y)&1 == 1 {
						i = k
					}
				}
				c := levels[i]
				if band == len(spec) { // the field's first pixel sits in the rim's shadow
					c = levels[len(levels)-2+boolInt(t < Gold)]
				}
				img.SetRGBA(x, y, c)
				continue
			}
			var c color.RGBA
			switch spec[band] {
			case 'o':
				c = p.outline
			case 'i':
				c = p.deep
			case 'g':
				c = p.metal[2]
			case 'l':
				c = p.metal[0]
			default:
				face := -float64(ox+oy) / (math.Hypot(float64(ox), float64(oy)) * math.Sqrt2)
				switch {
				case face > .35:
					c = p.metal[0]
				case face < -.35:
					c = p.metal[2]
				default:
					c = p.metal[1]
				}
			}
			img.SetRGBA(x, y, c)
		}
	}
	return &b
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// snap is the pixel nearest (x, y) that lies in rim band n.
func (b *body) snap(x, y, n int) (int, int) {
	bx, by, bd := x, y, 1<<30
	for j := 0; j < largeH; j++ {
		for i := 0; i < largeW; i++ {
			if d := (i-x)*(i-x) + (j-y)*(j-y); b.band[j][i] == n && d < bd {
				bx, by, bd = i, j, d
			}
		}
	}
	return bx, by
}

func crown(t Tier) *mask {
	m := newMask(24, 13)
	spikes := [][2]float64{{3.5, 3}, {12, .5}, {20.5, 3}}
	halfW := 3.5
	if t == Legendary {
		spikes, halfW = [][2]float64{{2, 4}, {7, 2}, {12, 0}, {17, 2}, {22, 4}}, 3
	}
	for _, s := range spikes {
		m.poly(tMain, s[0]-halfW, 9, s[0], s[1], s[0]+halfW, 9)
	}
	m.rect(0, 9, 24, 4, tMain)
	for _, s := range spikes {
		m.disc(s[0], s[1]+.6, 1.5, tLight)
	}
	m.bevel()
	m.rect(0, 9, 24, 1, tLight)
	m.rect(0, 12, 24, 1, tShade)
	gems := []float64{4, 12, 20}
	if t == Legendary {
		gems = []float64{3, 7.5, 12, 16.5, 21}
	}
	for _, g := range gems {
		m.disc(g, 10.9, 1, tGem)
	}
	if t >= Platinum {
		r := 2.4
		if t == Legendary {
			r = 2.9
		}
		m.disc(12, 5.4, r, tGem)
		m.disc(11.2, 4.6, 1, tGemHi)
	}
	return m
}

func wing() *mask {
	m := newMask(14, 31)
	tips := [][2]int{{1, 1}, {0, 8}, {1, 15}, {4, 21}, {8, 27}}
	for i := len(tips) - 1; i >= 0; i-- {
		by := 10 + 4*i
		tp := tips[i]
		m.poly(tMain, 14, float64(by)-2.5, 14, float64(by)+2.5, float64(tp[0])+1, float64(tp[1])+1.5, float64(tp[0]), float64(tp[1]))
		m.line(14, by+2, tp[0]+1, tp[1]+1, tShade)
	}
	m.bevel()
	return m
}

func banner() *mask {
	m := newMask(62, 13)
	m.poly(tRibLo, 0, 3, 12, 3, 12, 12, 0, 12, 4, 7.5)
	m.poly(tRibLo, 62, 3, 50, 3, 50, 12, 62, 12, 58, 7.5)
	m.rect(8, 1, 46, 11, tRib)
	m.rect(8, 2, 46, 1, tRibHi)
	m.rect(8, 1, 46, 1, tLight)
	m.rect(8, 11, 46, 1, tShade)
	m.rect(8, 10, 46, 1, tRibLo)
	for x := 14; x < 25; x += 2 {
		m.set(x, 6, tRibHi)
		m.set(61-x, 6, tRibHi)
	}
	m.poly(tLight, 31, 3.5, 34, 6.5, 31, 9.5, 28, 6.5)
	m.poly(tRib, 31, 5, 32.5, 6.5, 31, 8, 29.5, 6.5)
	return m
}

func sparkle(img *image.RGBA, x, y, n int, c color.RGBA) {
	put(img, x, y, c)
	for i := 1; i <= n; i++ {
		put(img, x-i, y, c)
		put(img, x+i, y, c)
		put(img, x, y-i, c)
		put(img, x, y+i, c)
	}
}

// halo is the rays behind a legendary shield: eight long wedges, each with a
// one pixel outline and a bright core, a pair of them lit at a time.
func halo(img *image.RGBA, p *pal, frame int) {
	const cx, cy, from, to = 35.0 + haloPadX, 38.0 + haloPadY, 22.0, 52.0
	hot := mix(p.ray, color.RGBA{255, 255, 255, 255}, .7)
	for i := 0; i < 8; i++ {
		a := (22.5 + 45*float64(i)) * math.Pi / 180
		ca, sa := math.Cos(a), math.Sin(a)
		lit := (i+frame)%4 == 0
		for y := 0; y < largeH+2*haloPadY; y++ {
			for x := 0; x < largeW+2*haloPadX; x++ {
				dx, dy := float64(x)+.5-cx, float64(y)+.5-cy
				r, d := dx*ca+dy*sa, math.Abs(-dx*sa+dy*ca)
				w := 5.2 - 4.2*(r-from)/(to-from)
				if r < from || r > to || d > w {
					continue
				}
				c := mix(p.ray, p.outline, .1)
				switch {
				case d > w-1:
					c = p.outline
				case d < w-2.4 && lit:
					c = hot
				case d < w-2.4:
					c = mix(p.ray, hot, .45)
				}
				img.SetRGBA(x, y, c)
			}
		}
	}
}

func drawLarge(t Tier, d Domain, frame int) *image.RGBA {
	p := &pals[t]
	shield := image.NewRGBA(image.Rect(0, 0, largeW, largeH))
	metal := p.metalTone
	if t >= Platinum {
		w := wing()
		blit(shield, w, 0, 17, metal, p.outline, false)
		blit(shield, w, largeW-14, 17, metal, p.outline, true)
	}
	b := paintBody(shield, t, p)
	// the emblem, in the field
	em := newMask(emW, emH)
	bigEmblems[d](em)
	em.bevel()
	blit(shield, em, midX-emW/2, 28, p.emTone, p.outline, false)
	if t >= Silver {
		for _, r := range [][2]int{{19, 21}, {51, 21}, {14, 37}, {56, 37}, {21, 52}, {49, 52}} {
			x, y := b.snap(r[0], r[1], 4)
			if t == Silver || t == Gold {
				x, y = b.snap(r[0], r[1], 5-boolInt(t == Silver))
			}
			if t == Legendary {
				g := rimGems[(x+y)%len(rimGems)]
				put(shield, x, y, g[1])
				put(shield, x+1, y, g[0])
				put(shield, x, y+1, g[0])
				put(shield, x+1, y+1, mix(g[0], p.outline, .5))
				continue
			}
			put(shield, x, y, p.metal[0])
			put(shield, x+1, y, p.metal[1])
			put(shield, x, y+1, p.metal[1])
			put(shield, x+1, y+1, p.deep)
		}
		blit(shield, banner(), 4, 58, metal, p.outline, false)
	}
	if t >= Gold {
		blit(shield, crown(t), midX-12, 1, metal, p.outline, false)
	}
	if t == Gold || t == Platinum {
		for _, s := range [][2]int{{8, 11}, {63, 47}, {60, 8}} {
			if t == Gold && s[0] == 60 {
				continue
			}
			sparkle(shield, s[0], s[1], 2, p.twinkle)
			put(shield, s[0], s[1], color.RGBA{255, 255, 255, 255})
		}
	}
	if t == Legendary {
		pos := [][2]int{{6, 12}, {64, 40}, {58, 6}, {8, 56}, {66, 64}, {3, 36}}
		sparkle(shield, pos[frame][0], pos[frame][1], 2, p.twinkle)
		o := pos[(frame+3)%6]
		sparkle(shield, o[0], o[1], 1, p.twinkle)
		// the glint sweeps from the upper left to the lower right
		if frame < 5 {
			c := -4 + frame*(largeW+largeH*7/10)/4
			for y := 0; y < largeH; y++ {
				for x := 0; x < largeW; x++ {
					px := shield.RGBAAt(x, y)
					if px.A == 0 || px == p.outline {
						continue
					}
					switch dist := math.Abs(float64(x) + .7*float64(y) - float64(c)); {
					case dist < 1.6:
						shield.SetRGBA(x, y, mix(px, p.glint, .85))
					case dist < 3.6:
						shield.SetRGBA(x, y, mix(px, p.glint, .35))
					}
				}
			}
		}
		img := image.NewRGBA(image.Rect(0, 0, largeW+2*haloPadX, largeH+2*haloPadY))
		halo(img, p, frame)
		for y := 0; y < largeH; y++ {
			for x := 0; x < largeW; x++ {
				if c := shield.RGBAAt(x, y); c.A != 0 {
					img.SetRGBA(x+haloPadX, y+haloPadY, c)
				}
			}
		}
		return img
	}
	return shield
}
