package skyline

import (
	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
)

// dress.go is what makes a street of one kind of building read as a street
// of different ones: every silhouette takes one of three variants, chosen
// by a stable hash of its copy's seed, that sets its height, its window
// pattern and what stands on its roof. It also finds a sprite's outline (the
// cells that are its edge) and gives the lit windows their light and flicker.

const numVariants = 3

// variantOf is the variant of the copy with this seed (0, 1 or 2): the same
// for ever, and different for neighbouring copies.
func variantOf(seed uint64) int { return int(hash(int(seed), 41) % numVariants) }

// variantHeight scales a silhouette's height by its variant: as drawn, low,
// tall.
var variantHeight = [numVariants]float64{1, 0.84, 1.16}

// variantRoom is the most rows a variant's roof furniture adds.
var variantRoom = [numVariants]int{0, 2, 3}

func isWall(sl slot) bool { return sl == sWall || sl == sWallLit || sl == sWallShade }

// raise adds n empty rows on top of a sprite, shifting its marks down.
func (s *sprite) raise(n int) *sprite {
	if n <= 0 {
		return s
	}
	o := newSprite(s.w, s.h+n)
	copy(o.c[n*s.w:], s.c)
	sh := func(ps []pt) []pt {
		out := make([]pt, 0, len(ps))
		for _, p := range ps {
			out = append(out, pt{p.x, p.y + n})
		}
		return out
	}
	o.smoke, o.beacons, o.blades, o.flame = sh(s.smoke), sh(s.beacons), sh(s.blades), sh(s.flame)
	o.carryExtras(s, 0, n)
	return o
}

// solidWall reports a cell that is plain facade: a full block of wall.
func (s *sprite) solidWall(x, y int) bool {
	c := s.at(x, y)
	return c.ch == '█' && isWall(c.fg)
}

// dress gives a built sprite its variant: window rows (punched where the
// form drew none, thinned where it drew many), and for the taller and the
// lower variant a chimney or an aerial (a flag, early on) on the roof. Anything that is not a
// building (fields, tents, trees, rock) comes back as it was. maxH is the
// scene's room; the roof furniture is dropped rather than overflow it.
func dress(s *sprite, v int, seed uint64, age, maxH int) *sprite {
	walls, wins := 0, 0
	for _, c := range s.c {
		switch {
		case c.ch == '█' && isWall(c.fg):
			walls++
		case c.fg == sWin || c.bg == sWin:
			wins++
		}
	}
	if walls < 12 || s.w < 5 || s.h < 4 {
		return s
	}
	if wins == 0 {
		pat := [numVariants][4]int{{2, 2, 1, 1}, {3, 2, 2, 1}, {2, 3, 1, 2}}[v]
		dx, dy, ox, oy := pat[0], pat[1], pat[2], pat[3]
		top := s.h
		for x := 0; x < s.w; x++ {
			if t := s.topAt(x); t >= 0 && t < top {
				top = t
			}
		}
		for y := top + oy; y < s.h-1; y += dy {
			for x := ox; x < s.w-1; x += dx {
				// only inside the facade: wall on both sides and above
				if s.solidWall(x, y) && s.solidWall(x-1, y) && s.solidWall(x+1, y) && s.solidWall(x, y-1) {
					c := s.at(x, y)
					s.put(x, y, '■', sWin, c.fg)
				}
			}
		}
	} else if v == 1 {
		for y := 0; y < s.h; y++ { // a sparser pattern: every other window out
			for x := 0; x < s.w; x++ {
				c := s.at(x, y)
				if (c.fg == sWin && c.bg != sNone && (x/2+y)%2 == 1) && isWall(c.bg) {
					s.put(x, y, '█', c.bg, sNone)
				}
			}
		}
	}
	extra := variantRoom[v]
	if extra == 0 || s.h+extra > maxH {
		return s
	}
	// the roof furniture stands on a column away from the edges
	cx := 1 + int(hash(int(seed), 43)%uint64(s.w-2))
	t := s.topAt(cx)
	if t < 0 {
		return s
	}
	if sl := s.at(cx, t).fg; !isWall(sl) && sl != sRoof && sl != sRoofShade && sl != sTrim {
		return s
	}
	if t < extra {
		s = s.raise(extra - t)
		t = extra
	}
	switch v {
	case 1: // a chimney stack
		for k := 1; k <= 2; k++ {
			s.put(cx, t-k, '█', sTrim, sNone)
		}
		s.put(cx, t-2, '▄', sTrim, sNone)
	default: // a flagpole before the electric age, an aerial after
		s.put(cx, t-1, '│', sMetalDark, sNone)
		s.put(cx, t-2, '│', sMetalDark, sNone)
		if age <= 9 {
			s.put(cx, t-3, '▸', sTrim, sNone)
		} else {
			s.put(cx, t-3, '┬', sMetal, sNone)
		}
	}
	return s
}

// outlined reports a slot that takes an outline: the built materials. The
// living, the water, the smoke and every light keep their own edges.
func outlined(sl slot) bool {
	switch sl {
	case sWall, sWallLit, sWallShade, sWallDark, sRoof, sRoofShade, sTrim, sMetal, sMetalDark, sGlass, sGlassHi, sRock, sRockDark:
		return true
	}
	return false
}

// edges returns the sprite's outline mask, one flag per cell, cached: a cell
// is an edge when it is built and has no neighbour to its left, to its right
// or above it. The foot is not an edge; the ground is the sprite's own line.
func (s *sprite) edges() []bool {
	if s.edge != nil {
		return s.edge
	}
	e := make([]bool, s.w*s.h)
	for y := 0; y < s.h; y++ {
		for x := 0; x < s.w; x++ {
			c := s.c[y*s.w+x]
			if c.ch == 0 || !outlined(c.fg) {
				continue
			}
			e[y*s.w+x] = s.at(x-1, y).ch == 0 || s.at(x+1, y).ch == 0 || s.at(x, y-1).ch == 0
		}
	}
	s.edge = e
	return e
}

// outlineK is how much darker a silhouette's edge is than its face.
const outlineK = 0.6

// rowTone is how much a back row is dimmed and cooled so the three rows read
// as depth: the front row as it is, the middle a step back, the back row two.
var (
	rowDim  = [numHaze]float64{1, 0.88, 0.76, 1, 1}
	rowCool = [numHaze]float64{0, 0.08, 0.16, 0, 0}
)

// tone leans a built colour cooler for its row: in a dark theme by a blue filter,
// in a light one toward the horizon, which is how paper reads as far.
func (p *pal) tone(c tcell.Color, depth int) tcell.Color {
	if depth <= 0 || depth > 2 {
		return c
	}
	if p.light {
		return theme.Mix(c, p.horizon, rowCool[depth]*1.4)
	}
	// a cool filter: white leaning toward the water's blue, multiplied in
	k := rowCool[depth]
	return theme.Tint(c, theme.Mix(theme.SkyColor(theme.SkyWhite), theme.SkyColor(theme.SkyWater), k*1.6))
}

// rowCap is how bright, at most, a row is against the row in front of it.
var rowCap = [numHaze]float64{1, 0.9, 0.78, 1, 1}

// recede dims a hazed colour to its row's share of the front row's light: a
// dark theme's back rows are dimmer than the front, by day too. A light theme
// keeps the haze alone (fading toward the paper is its dimming), and a
// self-lit slot keeps its own light.
func (p *pal) recede(c, front tcell.Color, sl slot, depth int) tcell.Color {
	if p.light || sl.emissive() {
		return c
	}
	if lf, lc := luma(front)*rowCap[depth], luma(c); lc > lf && lc > 0 {
		return theme.Shade(c, lf/lc)
	}
	return c
}

// windowStyle names how an age lights its windows.
type windowStyle uint8

const (
	wTorch    windowStyle = iota // stone to iron: torches and fires
	wLamp                        // classical to the Victorian: oil and gas lamps
	wElectric                    // electric and atomic: warm bulbs
	wLED                         // modern to digital: cool white light and screens
	wNeon                        // cyberpunk on: neon and plasma
)

func styleOfAge(age int) windowStyle {
	switch {
	case age <= 3:
		return wTorch
	case age <= 9:
		return wLamp
	case age <= 11:
		return wElectric
	case age <= 14:
		return wLED
	}
	return wNeon
}

// windowHue is the light of one window of an age: torches burn gold and
// orange, lamps are warm, bulbs warm, modern light cool and white, neon cold.
func windowHue(age int, k uint64) theme.SkyHue {
	switch styleOfAge(age) {
	case wTorch:
		return [3]theme.SkyHue{theme.SkyFireGold, theme.SkyFire, theme.SkyFireGold}[k%3]
	case wLamp:
		if k%5 == 0 {
			return theme.SkyWinPale
		}
		return theme.SkyWinWarm
	case wElectric:
		if k%4 == 0 {
			return theme.SkyWinPale
		}
		return theme.SkyWinWarm
	case wLED:
		if k%5 == 0 {
			return theme.SkyWinWarm
		}
		if k%2 == 0 {
			return theme.SkyWinNeon
		}
		return theme.SkyWinPale
	}
	if k%2 == 0 {
		return theme.SkyWinNeon
	}
	return theme.SkyWinPale
}

// windowGlow is how bright one lit window burns (0.3 to 1) at animation
// frame anim: a torch wavers, a lamp breathes, a bulb holds and now and then
// browns out, a sign stutters. Slow throughout, and still when the animation
// is (the map's motion setting holds anim at one frame).
func windowGlow(age int, seed, sx, sy, anim int) float64 {
	ph := hashf(seed, sx, sy, 61) // (a phase in turns: mapmodel.Sin takes turns)
	switch styleOfAge(age) {
	case wTorch:
		return 0.72 + 0.28*float64(mapmodel.Sin(float64(anim)/32+ph))
	case wLamp:
		return 0.86 + 0.14*float64(mapmodel.Sin(float64(anim)/72+ph))
	case wElectric, wLED:
		if hash(seed, sx, sy, anim/40)%53 == 0 {
			return 0.5
		}
		return 1
	}
	if hash(seed, sx, sy, anim/6)%31 == 0 {
		return 0.35
	}
	return 0.82 + 0.18*float64(mapmodel.Sin(float64(anim)/56+ph))
}
