package main

import "strings"

// Slot is a material slot: a sprite cell names what it is made of (wall,
// roof, window, neon...) and the renderer decides the colour from the
// building's material, the time of day, haze depth and the theme. That is
// what lets one sprite serve day, night, smog, a light theme and a
// monochrome theme without redrawing it.
type Slot uint8

const (
	SNone      Slot = iota // transparent: whatever is behind shows
	SWall                  // main facade
	SWallLit               // sun-facing edge
	SWallShade             // side away from the sun
	SWallDark              // recesses, doorways, beams
	SRoof
	SRoofShade
	STrim // cornices, columns, stone dressing
	SWin  // a window: dark glass by day, maybe lit at night
	SMetal
	SMetalDark
	SGlass
	SGlassHi
	SNeon1 // emissive signage (always lit, brighter at night)
	SNeon2
	SNeon3
	SGlow   // emissive warm (forge mouths, lanterns, reactors)
	SBeacon // aircraft warning light (blinks)
	SLeaf
	SLeafDark
	STrunk
	SField1 // crops
	SField2
	SRock
	SRockDark
	SFire
	SInk // near-black outline
	SWater
	SSmoke
	numSlots
)

// SCell is one sprite cell. Ch == 0 is fully transparent.
type SCell struct {
	Ch     rune
	Fg, Bg Slot
}

type pt struct{ X, Y int }

// Sprite is a building (or prop) drawn in cells, bottom row on the ground.
type Sprite struct {
	W, H    int
	C       []SCell
	Smoke   []pt // chimney mouths: smoke rises from here
	Beacons []pt // blinking lights
	Blades  []pt // windmill hubs (animated sails)
	Flame   []pt // torches / braziers (animated)
}

func newSprite(w, h int) *Sprite {
	return &Sprite{W: w, H: h, C: make([]SCell, w*h)}
}

func (s *Sprite) in(x, y int) bool { return x >= 0 && y >= 0 && x < s.W && y < s.H }

func (s *Sprite) at(x, y int) SCell {
	if !s.in(x, y) {
		return SCell{}
	}
	return s.C[y*s.W+x]
}

func (s *Sprite) put(x, y int, ch rune, fg, bg Slot) {
	if s.in(x, y) {
		s.C[y*s.W+x] = SCell{ch, fg, bg}
	}
}

// over draws ch in fg but keeps the cell's existing background (or makes
// the background the existing solid fill, for half blocks over a body).
func (s *Sprite) over(x, y int, ch rune, fg Slot) {
	if !s.in(x, y) {
		return
	}
	c := s.C[y*s.W+x]
	bg := c.Bg
	if c.Ch == '█' {
		bg = c.Fg
	}
	s.C[y*s.W+x] = SCell{ch, fg, bg}
}

func (s *Sprite) rect(x, y, w, h int, ch rune, fg, bg Slot) {
	for yy := y; yy < y+h; yy++ {
		for xx := x; xx < x+w; xx++ {
			s.put(xx, yy, ch, fg, bg)
		}
	}
}

func (s *Sprite) hline(x0, x1, y int, ch rune, fg, bg Slot) {
	for x := x0; x <= x1; x++ {
		s.put(x, y, ch, fg, bg)
	}
}

func (s *Sprite) vline(x, y0, y1 int, ch rune, fg, bg Slot) {
	for y := y0; y <= y1; y++ {
		s.put(x, y, ch, fg, bg)
	}
}

// text writes runes left to right in fg over bg.
func (s *Sprite) text(x, y int, str string, fg, bg Slot) {
	for _, r := range str {
		if r != ' ' {
			s.put(x, y, r, fg, bg)
		}
		x++
	}
}

// body fills a shaded block: a lit left column, the facade, and a shaded
// right strip, which is most of what makes a flat silhouette read as a
// volume.
func (s *Sprite) body(x, y, w, h int) {
	for yy := y; yy < y+h; yy++ {
		for xx := x; xx < x+w; xx++ {
			slot := SWall
			switch {
			case xx == x && w > 2:
				slot = SWallLit
			case xx >= x+w-1-w/5 && w > 3:
				slot = SWallShade
			}
			s.put(xx, yy, '█', slot, SNone)
		}
	}
}

// windows punches a window grid into a body.
func (s *Sprite) windows(x, y, w, h, dx, dy, ox, oy int, ch rune) {
	for yy := y + oy; yy < y+h; yy += dy {
		for xx := x + ox; xx < x+w; xx += dx {
			c := s.at(xx, yy)
			if c.Ch == '█' {
				s.put(xx, yy, ch, SWin, c.Fg)
			}
		}
	}
}

// blit copies o into s at (x, y), skipping transparent cells.
func (s *Sprite) blit(o *Sprite, x, y int) {
	for yy := 0; yy < o.H; yy++ {
		for xx := 0; xx < o.W; xx++ {
			c := o.C[yy*o.W+xx]
			if c.Ch != 0 {
				s.put(x+xx, y+yy, c.Ch, c.Fg, c.Bg)
			}
		}
	}
	for _, p := range o.Smoke {
		s.Smoke = append(s.Smoke, pt{p.X + x, p.Y + y})
	}
	for _, p := range o.Beacons {
		s.Beacons = append(s.Beacons, pt{p.X + x, p.Y + y})
	}
	for _, p := range o.Blades {
		s.Blades = append(s.Blades, pt{p.X + x, p.Y + y})
	}
	for _, p := range o.Flame {
		s.Flame = append(s.Flame, pt{p.X + x, p.Y + y})
	}
}

// halves decodes a cell into its top and bottom half-pixel slots.
func (c SCell) halves() (Slot, Slot) {
	switch c.Ch {
	case '█':
		return c.Fg, c.Fg
	case '▀':
		return c.Fg, c.Bg
	case '▄':
		return c.Bg, c.Fg
	}
	return SNone, SNone
}

// pset sets one half-cell pixel (py counts half rows from the top), keeping
// the other half. This is how slopes, domes and trees get smooth edges
// while staying on the character grid.
func (s *Sprite) pset(x, py int, slot Slot) {
	y := py / 2
	if !s.in(x, y) || py < 0 {
		return
	}
	c := s.at(x, y)
	t, b := c.halves()
	if c.Ch != 0 && c.Ch != '█' && c.Ch != '▀' && c.Ch != '▄' {
		t, b = c.Bg, c.Bg
	}
	if py%2 == 0 {
		t = slot
	} else {
		b = slot
	}
	switch {
	case t == b && t != SNone:
		s.put(x, y, '█', t, SNone)
	case b == SNone:
		s.put(x, y, '▀', t, SNone)
	case t == SNone:
		s.put(x, y, '▄', b, SNone)
	default:
		s.put(x, y, '▀', t, b)
	}
}

// maskSlots maps the letters used in hand-drawn art masks to slots.
var maskSlots = map[rune]Slot{
	'w': SWall, 'W': SWallLit, 's': SWallShade, 'd': SWallDark,
	'r': SRoof, 'R': SRoofShade, 't': STrim, 'l': SWin,
	'm': SMetal, 'M': SMetalDark, 'g': SGlass, 'G': SGlassHi,
	'n': SNeon1, 'N': SNeon2, 'v': SNeon3, 'o': SGlow, 'b': SBeacon,
	'f': SLeaf, 'F': SLeafDark, 'k': STrunk, 'c': SField1, 'C': SField2,
	'x': SRock, 'X': SRockDark, 'i': SFire, 'K': SInk, 'a': SWater, 'S': SSmoke,
}

// art builds a sprite from hand-drawn lines plus a foreground mask and an
// optional background mask (same shape). A space in the art is
// transparent; '.' in the bg mask (or no bg mask) leaves the background
// transparent. Markers in the art: '*' becomes a smoke mouth, '+' a beacon,
// '@' a windmill hub, '^' a flame (each drawn as nothing until animated).
func art(lines, fgMask, bgMask string) *Sprite {
	al := splitArt(lines)
	fl := splitArt(fgMask)
	bl := splitArt(bgMask)
	h := len(al)
	w := 0
	for _, l := range al {
		if n := len([]rune(l)); n > w {
			w = n
		}
	}
	s := newSprite(w, h)
	for y, l := range al {
		fr := runesAt(fl, y)
		br := runesAt(bl, y)
		for x, ch := range []rune(l) {
			switch ch {
			case ' ':
				continue
			case '*':
				s.Smoke = append(s.Smoke, pt{x, y})
				continue
			case '+':
				s.Beacons = append(s.Beacons, pt{x, y})
				continue
			case '@':
				s.Blades = append(s.Blades, pt{x, y})
				ch = '▪'
			case '^':
				s.Flame = append(s.Flame, pt{x, y})
				continue
			}
			fg, bg := SWall, SNone
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
	s = strings.TrimPrefix(s, "\n")
	s = strings.TrimSuffix(s, "\n")
	return strings.Split(s, "\n")
}

func runesAt(ls []string, y int) []rune {
	if y < len(ls) {
		return []rune(ls[y])
	}
	return nil
}

// mirror flips a sprite left to right (glyphs with a handedness swap).
func (s *Sprite) mirror() *Sprite {
	o := newSprite(s.W, s.H)
	swap := map[rune]rune{'▌': '▐', '▐': '▌', '/': '\\', '\\': '/', '╱': '╲', '╲': '╱', '(': ')', ')': '(',
		'◄': '►', '►': '◄', '┌': '┐', '┐': '┌', '└': '┘', '┘': '└', '╔': '╗', '╗': '╔', '╚': '╝', '╝': '╚'}
	for y := 0; y < s.H; y++ {
		for x := 0; x < s.W; x++ {
			c := s.C[y*s.W+x]
			if r, ok := swap[c.Ch]; ok {
				c.Ch = r
			}
			// lit and shaded faces swap sides with the mirror
			if c.Fg == SWallLit {
				c.Fg = SWallShade
			} else if c.Fg == SWallShade {
				c.Fg = SWallLit
			}
			o.C[y*s.W+(s.W-1-x)] = c
		}
	}
	for _, p := range s.Smoke {
		o.Smoke = append(o.Smoke, pt{s.W - 1 - p.X, p.Y})
	}
	for _, p := range s.Beacons {
		o.Beacons = append(o.Beacons, pt{s.W - 1 - p.X, p.Y})
	}
	for _, p := range s.Blades {
		o.Blades = append(o.Blades, pt{s.W - 1 - p.X, p.Y})
	}
	for _, p := range s.Flame {
		o.Flame = append(o.Flame, pt{s.W - 1 - p.X, p.Y})
	}
	return o
}
