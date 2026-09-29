package skyline

import (
	"math"
	"strings"
)

// sprite.go is the cell grammar every building, wonder and vehicle is drawn
// in. A sprite cell names what it is made of (a slot: wall, roof, window,
// neon…), never a colour: the frame palette decides the colour from the
// building's material, the time of day, the haze of its depth row and the
// theme. That is what lets one sprite serve day, night, smog, a light theme
// and a monochrome one.

// slot is a material slot.
type slot uint8

const (
	sNone      slot = iota // transparent: whatever is behind shows
	sWall                  // main facade
	sWallLit               // sun-facing edge
	sWallShade             // side away from the sun
	sWallDark              // recesses, doorways, beams
	sRoof
	sRoofShade
	sTrim // cornices, columns, stone dressing
	sWin  // a window: dark glass by day, lit at night when staffed
	sMetal
	sMetalDark
	sGlass
	sGlassHi
	sNeon1 // emissive signage
	sNeon2
	sNeon3
	sGlow   // emissive warm light (forge mouths, lanterns, reactors)
	sBeacon // aircraft warning light
	sLeaf
	sLeafDark
	sTrunk
	sField1
	sField2
	sRock
	sRockDark
	sFire
	sInk
	sWater
	sSmoke
	numSlots
)

// emissive reports a slot that lights itself.
func (s slot) emissive() bool {
	switch s {
	case sNeon1, sNeon2, sNeon3, sGlow, sBeacon, sFire:
		return true
	}
	return false
}

// scell is one sprite cell; ch == 0 is transparent.
type scell struct {
	ch     rune
	fg, bg slot
}

type pt struct{ x, y int }

// sprite is a drawing in cells, its bottom row on the ground.
type sprite struct {
	w, h    int
	c       []scell
	smoke   []pt // chimney mouths
	beacons []pt // blinking lights
	blades  []pt // windmill hubs
	flame   []pt // torches and braziers
}

func newSprite(w, h int) *sprite {
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	return &sprite{w: w, h: h, c: make([]scell, w*h)}
}

func (s *sprite) in(x, y int) bool { return x >= 0 && y >= 0 && x < s.w && y < s.h }

func (s *sprite) at(x, y int) scell {
	if !s.in(x, y) {
		return scell{}
	}
	return s.c[y*s.w+x]
}

func (s *sprite) put(x, y int, ch rune, fg, bg slot) {
	if s.in(x, y) {
		s.c[y*s.w+x] = scell{ch, fg, bg}
	}
}

func (s *sprite) clear(x, y int) { s.put(x, y, 0, sNone, sNone) }

func (s *sprite) hline(x0, x1, y int, ch rune, fg, bg slot) {
	for x := x0; x <= x1; x++ {
		s.put(x, y, ch, fg, bg)
	}
}

func (s *sprite) vline(x, y0, y1 int, ch rune, fg, bg slot) {
	for y := y0; y <= y1; y++ {
		s.put(x, y, ch, fg, bg)
	}
}

// text writes the non-space runes of str.
func (s *sprite) text(x, y int, str string, fg, bg slot) {
	for _, r := range str {
		if r != ' ' {
			s.put(x, y, r, fg, bg)
		}
		x++
	}
}

// body fills a shaded block: a lit left column, the facade and a shaded
// right strip, which is most of what makes a flat silhouette a volume.
func (s *sprite) body(x, y, w, h int) {
	for yy := y; yy < y+h; yy++ {
		for xx := x; xx < x+w; xx++ {
			sl := sWall
			switch {
			case xx == x && w > 2:
				sl = sWallLit
			case xx >= x+w-1-w/5 && w > 3:
				sl = sWallShade
			}
			s.put(xx, yy, '█', sl, sNone)
		}
	}
}

// windows punches a window grid into solid cells.
func (s *sprite) windows(x, y, w, h, dx, dy, ox, oy int, ch rune) {
	for yy := y + oy; yy < y+h; yy += dy {
		for xx := x + ox; xx < x+w; xx += dx {
			if c := s.at(xx, yy); c.ch == '█' {
				s.put(xx, yy, ch, sWin, c.fg)
			}
		}
	}
}

// blit copies o into s at (x, y), skipping transparent cells.
func (s *sprite) blit(o *sprite, x, y int) {
	for yy := 0; yy < o.h; yy++ {
		for xx := 0; xx < o.w; xx++ {
			if c := o.c[yy*o.w+xx]; c.ch != 0 {
				s.put(x+xx, y+yy, c.ch, c.fg, c.bg)
			}
		}
	}
	mv := func(ps []pt) []pt {
		out := make([]pt, 0, len(ps))
		for _, p := range ps {
			out = append(out, pt{p.x + x, p.y + y})
		}
		return out
	}
	s.smoke = append(s.smoke, mv(o.smoke)...)
	s.beacons = append(s.beacons, mv(o.beacons)...)
	s.blades = append(s.blades, mv(o.blades)...)
	s.flame = append(s.flame, mv(o.flame)...)
}

// halves decodes a cell into its top and bottom half-cell slots.
func (c scell) halves() (slot, slot) {
	switch c.ch {
	case '█':
		return c.fg, c.fg
	case '▀':
		return c.fg, c.bg
	case '▄':
		return c.bg, c.fg
	}
	return sNone, sNone
}

// pset sets one half-cell pixel (py counts half rows from the top), keeping
// the other half: slopes, domes and trees get smooth edges on the grid.
func (s *sprite) pset(x, py int, sl slot) {
	y := py / 2
	if py < 0 || !s.in(x, y) {
		return
	}
	c := s.at(x, y)
	t, b := c.halves()
	if c.ch != 0 && c.ch != '█' && c.ch != '▀' && c.ch != '▄' {
		t, b = c.bg, c.bg
	}
	if py%2 == 0 {
		t = sl
	} else {
		b = sl
	}
	switch {
	case t == b && t != sNone:
		s.put(x, y, '█', t, sNone)
	case b == sNone:
		s.put(x, y, '▀', t, sNone)
	case t == sNone:
		s.put(x, y, '▄', b, sNone)
	default:
		s.put(x, y, '▀', t, b)
	}
}

// roof rasterises a pitched roof over [x, x+w) with its eave on row y, at
// slope columns per half row. It returns the cell rows it used.
func (s *sprite) roof(x, w, y int, slope float64) int {
	cx := float64(x) + float64(w)/2
	rows := int(math.Ceil(float64(w) / 2 / slope))
	bot := 2*y + 1
	for i := 0; i < rows; i++ {
		py := bot - (rows - 1 - i)
		hw := float64(float64(i+1) * slope)
		for xx := x; xx < x+w; xx++ {
			if math.Abs(float64(xx)+0.5-cx) <= hw+0.01 {
				sl := sRoof
				if float64(xx)+0.5 > cx {
					sl = sRoofShade
				}
				s.pset(xx, py, sl)
			}
		}
	}
	return (rows + 1) / 2
}

// dome sits a half-block dome on row y (its bottom row).
func (s *sprite) dome(x, w, y int, sl slot) int {
	rx := float64(w) / 2
	ry := float64(rx * 0.95)
	cx := float64(x) + rx
	bot := 2*y + 1
	n := int(math.Ceil(ry))
	for i := 0; i < n; i++ {
		yy := (float64(i) + 0.5) / ry
		if yy > 1 {
			continue
		}
		hw := float64(rx * math.Sqrt(1-float64(yy*yy)))
		for xx := x; xx < x+w; xx++ {
			d := float64(xx) + 0.5 - cx
			if math.Abs(d) <= hw+0.15 {
				c := sl
				if d > float64(rx*0.35) {
					c = sMetalDark
				}
				s.pset(xx, bot-i, c)
			}
		}
	}
	return (n + 1) / 2
}

// cone rasterises a conical roof from apex pixel row top to full width at
// pixel row bot.
func (s *sprite) cone(x, w, top, bot int) {
	cx := float64(x) + float64(w)/2
	n := float64(bot - top)
	for py := top; py <= bot; py++ {
		hw := 0.5 + float64((float64(w)/2-0.5)*float64(py-top))/n
		for xx := x; xx < x+w; xx++ {
			d := float64(xx) + 0.5 - cx
			if math.Abs(d) <= hw {
				sl := sRoof
				if d > 0.4 {
					sl = sRoofShade
				}
				s.pset(xx, py, sl)
			}
		}
	}
}

// crenels draws battlements on row y.
func (s *sprite) crenels(x, w, y int, sl slot) {
	for i := 0; i < w; i++ {
		ch := '▄'
		if i%2 == 0 {
			ch = '█'
		}
		s.put(x+i, y, ch, sl, sNone)
	}
}

// maskSlots maps the letters of hand-drawn art masks to slots.
var maskSlots = map[rune]slot{
	'w': sWall, 'W': sWallLit, 's': sWallShade, 'd': sWallDark,
	'r': sRoof, 'R': sRoofShade, 't': sTrim, 'l': sWin,
	'm': sMetal, 'M': sMetalDark, 'g': sGlass, 'G': sGlassHi,
	'n': sNeon1, 'N': sNeon2, 'v': sNeon3, 'o': sGlow, 'b': sBeacon,
	'f': sLeaf, 'F': sLeafDark, 'k': sTrunk, 'c': sField1, 'C': sField2,
	'x': sRock, 'X': sRockDark, 'i': sFire, 'K': sInk, 'a': sWater, 'S': sSmoke,
}

// art builds a sprite from drawn lines, a foreground mask and an optional
// background mask. A space is transparent; '*' marks a smoke mouth, '+' a
// beacon, '^' a flame.
func art(lines, fgMask, bgMask string) *sprite {
	al, fl, bl := splitArt(lines), splitArt(fgMask), splitArt(bgMask)
	w := 0
	for _, l := range al {
		if n := len([]rune(l)); n > w {
			w = n
		}
	}
	s := newSprite(w, len(al))
	for y, l := range al {
		fr, br := runesAt(fl, y), runesAt(bl, y)
		for x, ch := range []rune(l) {
			switch ch {
			case ' ':
				continue
			case '*':
				s.smoke = append(s.smoke, pt{x, y})
				continue
			case '+':
				s.beacons = append(s.beacons, pt{x, y})
				continue
			case '^':
				s.flame = append(s.flame, pt{x, y})
				continue
			}
			fg, bg := sWall, sNone
			if x < len(fr) {
				if v, ok := maskSlots[fr[x]]; ok {
					fg = v
				}
			}
			if x < len(br) {
				if v, ok := maskSlots[br[x]]; ok {
					bg = v
				}
			}
			s.put(x, y, ch, fg, bg)
		}
	}
	return s
}

func splitArt(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(strings.Trim(s, "\n"), "\n")
}

func runesAt(ls []string, y int) []rune {
	if y < len(ls) {
		return []rune(ls[y])
	}
	return nil
}

var mirrorRunes = map[rune]rune{'▌': '▐', '▐': '▌', '/': '\\', '\\': '/', '╱': '╲', '╲': '╱', '(': ')', ')': '(',
	'◄': '►', '►': '◄', '┌': '┐', '┐': '┌', '└': '┘', '┘': '└', '╔': '╗', '╗': '╔', '╚': '╝', '╝': '╚',
	'◢': '◣', '◣': '◢', '◤': '◥', '◥': '◤', '▖': '▗', '▗': '▖', '▘': '▝', '▝': '▘'}

// mirror flips a sprite left to right.
func (s *sprite) mirror() *sprite {
	o := newSprite(s.w, s.h)
	for y := 0; y < s.h; y++ {
		for x := 0; x < s.w; x++ {
			c := s.c[y*s.w+x]
			if r, ok := mirrorRunes[c.ch]; ok {
				c.ch = r
			}
			switch c.fg { // lit and shaded faces swap sides
			case sWallLit:
				c.fg = sWallShade
			case sWallShade:
				c.fg = sWallLit
			}
			o.c[y*s.w+(s.w-1-x)] = c
		}
	}
	fl := func(ps []pt) []pt {
		out := make([]pt, 0, len(ps))
		for _, p := range ps {
			out = append(out, pt{s.w - 1 - p.x, p.y})
		}
		return out
	}
	o.smoke, o.beacons, o.blades, o.flame = fl(s.smoke), fl(s.beacons), fl(s.blades), fl(s.flame)
	return o
}

// trimTop drops empty rows above the art so heights are honest.
func (s *sprite) trimTop() *sprite {
	marked := func(y int) bool {
		for _, ps := range [][]pt{s.smoke, s.beacons, s.flame, s.blades} {
			for _, p := range ps {
				if p.y == y {
					return true
				}
			}
		}
		return false
	}
	top := 0
	for top < s.h-1 && !marked(top) {
		empty := true
		for x := 0; x < s.w; x++ {
			if s.c[top*s.w+x].ch != 0 {
				empty = false
				break
			}
		}
		if !empty {
			break
		}
		top++
	}
	if top == 0 {
		return s
	}
	o := newSprite(s.w, s.h-top)
	copy(o.c, s.c[top*s.w:])
	sh := func(ps []pt) []pt {
		out := make([]pt, 0, len(ps))
		for _, p := range ps {
			out = append(out, pt{p.x, p.y - top})
		}
		return out
	}
	o.smoke, o.beacons, o.blades, o.flame = sh(s.smoke), sh(s.beacons), sh(s.blades), sh(s.flame)
	return o
}

// topAt is the first solid row of column x, or -1.
func (s *sprite) topAt(x int) int {
	for y := 0; y < s.h; y++ {
		if s.at(x, y).ch != 0 {
			return y
		}
	}
	return -1
}

// rnd is a small xorshift stream seeded from a hash: a sprite's jitter.
type rnd struct{ s uint32 }

func newRnd(seed uint64) *rnd { return &rnd{uint32(seed^seed>>32) | 1} }

func (r *rnd) u() uint32 {
	r.s ^= r.s << 13
	r.s ^= r.s >> 17
	r.s ^= r.s << 5
	return r.s
}

func (r *rnd) n(k int) int {
	if k <= 0 {
		return 0
	}
	return int(r.u() % uint32(k))
}

func (r *rnd) rng(a, b int) int { return a + r.n(b-a+1) }

func (r *rnd) pick(s string) rune {
	rs := []rune(s)
	return rs[r.n(len(rs))]
}
