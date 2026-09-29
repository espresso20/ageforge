package skyline

import (
	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
)

// fb.go is the frame buffer the scene composes into, with a depth per cell.
// Everything that draws names its depth (smaller is nearer) and a write only
// lands on a cell whose current owner is no nearer. The scene draws back to
// front anyway; the depth test is what lets traffic, smoke and fire drawn
// late still pass behind a nearer tower and in front of a farther one.

// Depths, nearest first.
const (
	dTop       uint8 = 1  // cursor, markers, chrome
	dWeather   uint8 = 4  // rain, snow, glitch bands
	dTag       uint8 = 6  // flows tags and new markers
	dShip      uint8 = 12 // ships on the bay
	dStreet    uint8 = 13 // the road lane
	dGround    uint8 = 15 // ground, road, verge, water
	dLamp      uint8 = 17 // street furniture
	dWonder    uint8 = 19
	dRow0      uint8 = 20
	dLane0     uint8 = 25 // behind the front row
	dRow1      uint8 = 30
	dLane1     uint8 = 35 // behind the middle row
	dRow2      uint8 = 40
	dLane2     uint8 = 45 // behind every row
	dAirLow    uint8 = 48 // aircraft over the city
	dLabel     uint8 = 55
	dTown      uint8 = 60 // the distant town
	dHills     uint8 = 62
	dHarbinger uint8 = 67
	dRidgeTown uint8 = 68
	dRidge     uint8 = 70
	dAir       uint8 = 78 // high aircraft
	dCloud     uint8 = 85
	dCrack     uint8 = 88
	dSkyObj    uint8 = 92
	dStar      uint8 = 95
	dSky       uint8 = 100
	dEmpty     uint8 = 255
)

// rowDepth is the depth of a lot row (0 front … 2 back).
func rowDepth(row int) uint8 {
	switch row {
	case 0:
		return dRow0
	case 1:
		return dRow1
	}
	return dRow2
}

type cell struct {
	ch     rune
	fg, bg tcell.Color
	d      uint8
}

type fb struct {
	w, h int
	c    []cell
}

func (f *fb) reset(w, h int) {
	f.w, f.h = w, h
	if cap(f.c) < w*h {
		f.c = make([]cell, w*h)
	}
	f.c = f.c[:w*h]
	for i := range f.c {
		f.c[i] = cell{ch: ' ', d: dEmpty}
	}
}

func (f *fb) at(x, y int) *cell {
	if x < 0 || y < 0 || x >= f.w || y >= f.h {
		return nil
	}
	return &f.c[y*f.w+x]
}

// set writes a whole cell if depth d is no farther than the cell's owner.
func (f *fb) set(x, y int, ch rune, fg, bg tcell.Color, d uint8) {
	if p := f.at(x, y); p != nil && d <= p.d {
		*p = cell{ch, fg, bg, d}
	}
}

// fill paints a cell solid.
func (f *fb) fill(x, y int, c tcell.Color, d uint8) { f.set(x, y, '█', c, c, d) }

// fg draws a glyph over the cell's background (partial glyphs let the scene
// behind show through).
func (f *fb) fg(x, y int, ch rune, fg tcell.Color, d uint8) {
	p := f.at(x, y)
	if p == nil || d > p.d {
		return
	}
	p.bg = p.show()
	p.ch, p.fg, p.d = ch, fg, d
}

// tint leans a cell toward c by a (smoke, haze, lamplight).
func (f *fb) tint(x, y int, c tcell.Color, a float64, d uint8) {
	if p := f.at(x, y); p != nil && d <= p.d {
		p.fg = theme.Mix(p.fg, c, a)
		p.bg = theme.Mix(p.bg, c, a)
	}
}

// text writes s at depth d and returns the column after it.
func (f *fb) text(x, y int, s string, fg, bg tcell.Color, d uint8) int {
	for _, r := range s {
		f.set(x, y, r, fg, bg, d)
		x++
	}
	return x
}

// show is the colour a cell shows as a background (a full block is its
// foreground colour).
func (c *cell) show() tcell.Color {
	if c.ch == '█' {
		return c.fg
	}
	return c.bg
}

// showAt is the background colour a cell shows (0 off the buffer).
func (f *fb) showAt(x, y int) tcell.Color {
	if p := f.at(x, y); p != nil {
		return p.show()
	}
	return 0
}

// blit writes the buffer into r through a canvas, every cell exactly once.
// A full block becomes a space on its colour (no font seams), the ASCII tier
// turns shades into spaces on a blended colour, and a glyph whose ink would
// vanish into its background is made legible or dropped.
func (f *fb) blit(scr tcell.Screen, r mapstyle.Rect, tier mapmodel.GlyphTier) {
	cv := &mapstyle.Canvas{Scr: scr, R: r, Tier: tier}
	for y := 0; y < r.H && y < f.h; y++ {
		for x := 0; x < r.W && x < f.w; x++ {
			c := f.c[y*f.w+x]
			ch, fg, bg := c.ch, c.fg, c.bg
			switch {
			case ch == 0:
				ch = ' '
			case ch == '█':
				ch, bg = ' ', fg
			case tier == mapmodel.TierASCII && ch == '░':
				ch, bg = ' ', theme.Mix(bg, fg, 0.25)
			case tier == mapmodel.TierASCII && ch == '▒':
				ch, bg = ' ', theme.Mix(bg, fg, 0.5)
			case tier == mapmodel.TierASCII && ch == '▓':
				ch, bg = ' ', theme.Mix(bg, fg, 0.75)
			}
			if ch != ' ' && fg == bg {
				if ch >= 0x2580 && ch <= 0x259F {
					ch = ' '
				} else {
					fg = theme.Legible(fg, bg, 2)
				}
			}
			if !fg.Valid() {
				fg = bg
			}
			cv.Put(x, y, ch, tcell.StyleDefault.Foreground(fg).Background(bg))
		}
	}
}
