package skyline

import "math"

// wonders.go: the landmarks. Each is a short drawing on the same cell
// grammar as the buildings, sized by the scene scale so a small terminal
// still gets the whole silhouette.

// wonderFamily lets a landmark wear a later period's finish.
var wonderFamily = map[string]int{"crystal_palace": 7, "great_monolith": 8, "hoover_dam": 7}

func wonderSprite(key string, scale float64) *sprite {
	sz := func(n float64) int { return max(3, int(float64(n*scale)+0.5)) }
	r := newRnd(hash(len(key), int(key[0])))
	switch key {
	case "sacred_grove":
		return art(`
    ▄▄███▄▄
  ▄█████████▄
 ████▓████▓███
 ▀███████▓███▀
  ▀▀██▀██▀██▀
    ▀▄▐█▌▄▀
  ·   ▐█▌   ·
 ▄▄▄▄▄███▄▄▄▄▄`, `
    fffFfff
  ffffffFffFF
 fffffffffFFFF
 FffffffFfFFFF
  FFffFkfFFFF
    kkkkkkk
  o   kkk   o
 xxxxxkkkXXXXX`, `
    .......
  ...........
 ....f....F...
 .............
  ...........
    ...k.k.
  .....k.....
 .............`)
	case "great_monolith":
		h := sz(11)
		s := newSprite(7, h+1)
		for y := 1; y < h; y++ {
			s.put(2, y, '█', sInk, sNone)
			s.put(3, y, '█', sInk, sNone)
			s.put(4, y, '█', sRockDark, sNone)
			if y%3 == 1 {
				s.put(3, y, '≡', sGlow, sInk)
			}
		}
		s.text(2, 0, "▄▄▄", sInk, sNone)
		s.text(0, h, "▄▄███▄▄", sRock, sNone)
		return s
	case "stonehenge":
		return art(`
▄▄▄▄▄   ▄▄▄▄▄   ▄▄▄▄▄
██ ██ ▄ ██ ██ ▄ ██ ██
██ ██ █ ██ ██ █ ██ ██
██ ██ █ ██ ██ █ ██ ██`, `
xxxxX   xxxxX   xxxxX
xX xX x xX xX X xX xX
xX xX x xX xX X xX xX
xX xX x xX xX X xX xX`, "")
	case "colosseum":
		return arcade(sz(8), 21, 3)
	case "parthenon":
		return widen(aClassical(newRnd(9), 9), 2)
	case "great_library":
		return library(sz(12))
	case "sistine_chapel":
		return basilica(sz(15))
	case "grand_lighthouse":
		return lighthouse(sz(17))
	case "crystal_palace":
		return glassHall(sz(8))
	case "eiffel_tower":
		return lattice(sz(22), 15)
	case "hoover_dam":
		return dam(sz(10))
	case "particle_accelerator":
		return accelerator(sz(9))
	case "space_program":
		return aLaunch(r, sz(20))
	case "global_network":
		return mast(sz(24))
	case "world_simulation":
		return dataCube(sz(18))
	case "neon_citadel":
		return citadel(sz(28))
	case "stellar_cradle":
		return cradle(sz(22))
	case "dyson_scaffold":
		return dyson(sz(20))
	case "warp_nexus":
		return ringGate(sz(16))
	case "cosmic_beacon":
		return beacon(sz(28))
	case "reality_anchor":
		return anchor(sz(20))
	case "singularity_core":
		return singularity(sz(20))
	}
	return aObelisk(r, sz(10))
}

// arcade: tiers of arches (the Colosseum), the top tier ruined on one side.
func arcade(h, w, tiers int) *sprite {
	th := max(2, h/tiers)
	s := newSprite(w, th*tiers+1)
	base := s.h - 1
	for t := 0; t < tiers; t++ {
		y1 := base - t*th
		y0 := y1 - th + 1
		for y := y0; y <= y1; y++ {
			for x := 0; x < w; x++ {
				sl := sWall
				if x == 0 {
					sl = sWallLit
				} else if x > w*3/4 {
					sl = sWallShade
				}
				switch {
				case y == y0 && (t > 0 || y0 == y1):
					s.put(x, y, '█', sTrim, sNone)
				case y == y0:
					s.put(x, y, '▄', sTrim, sNone)
				case x%2 == 1:
					s.put(x, y, '▀', sl, sWallDark)
				default:
					s.put(x, y, '█', sl, sNone)
				}
			}
		}
	}
	top := base - tiers*th + 1
	for x := w / 2; x < w; x++ {
		cut := (x - w/2) * th / max(1, w/2)
		for y := top; y < top+cut && y < s.h; y++ {
			s.clear(x, y)
		}
	}
	return s.trimTop()
}

// widen stretches a sprite by repeating its middle columns.
func widen(s *sprite, n int) *sprite {
	o := newSprite(s.w+n*2, s.h)
	mid := s.w / 2
	for y := 0; y < s.h; y++ {
		for x := 0; x < o.w; x++ {
			sx := x
			switch {
			case x > mid+n*2:
				sx = x - n*2
			case x > mid:
				sx = mid - (x-mid)%2
			}
			c := s.at(sx, y)
			o.put(x, y, c.ch, c.fg, c.bg)
		}
	}
	return o
}

func library(h int) *sprite {
	w := 21
	s := newSprite(w, h+2)
	base := s.h - 1
	bh := max(4, h/2)
	s.body(0, base-bh+1, w, bh)
	s.hline(0, w-1, base-bh+1, '▀', sTrim, sWall)
	for x := 1; x < w-1; x += 2 {
		s.vline(x, base-bh+2, base-1, '█', sTrim, sWallDark)
	}
	for _, tx := range []int{0, w - 3} {
		s.body(tx, base-bh-2, 3, bh+2)
		s.crenels(tx, 3, base-bh-3, sWall)
	}
	s.body(5, base-bh-1, 11, 2)
	d := s.dome(5, 11, base-bh-2, sMetal)
	s.put(10, base-bh-2-d, '▲', sTrim, sNone)
	s.put(10, base, '▄', sGlow, sWallDark)
	return s.trimTop()
}

func basilica(h int) *sprite {
	w := 23
	s := newSprite(w, h+3)
	base := s.h - 1
	bh := max(4, h*4/10)
	s.body(0, base-bh+1, w, bh)
	s.hline(0, w-1, base-bh+1, '▀', sTrim, sWall)
	for x := 1; x < w-1; x += 2 {
		s.put(x, base-1, '█', sTrim, sWallDark)
		s.put(x, base-2, '█', sTrim, sWallDark)
	}
	s.windows(1, base-bh+2, w-2, max(0, bh-4), 2, 1, 0, 0, '■')
	s.roof(4, 15, base-bh, 1)
	dy := base - bh - 3
	s.body(6, dy-1, 11, 3)
	s.windows(6, dy-1, 11, 2, 2, 1, 1, 0, '▌')
	dh := s.dome(5, 13, dy-2, sMetal)
	lan := max(1, h/8)
	for y := dy - 2 - dh; y > dy-2-dh-lan; y-- {
		s.put(11, y, '█', sTrim, sNone)
	}
	s.put(11, dy-2-dh-lan, '▲', sTrim, sNone)
	s.put(11, dy-3-dh-lan, '┼', sTrim, sNone)
	return s.trimTop()
}

func lighthouse(h int) *sprite {
	s := newSprite(9, h+2)
	base := s.h - 1
	for i := 0; i < h-3; i++ {
		y := base - i
		wd := 3 - i*2/max(1, h-3)
		for dx := -wd; dx <= wd; dx++ {
			striped := (i/2)%2 == 1
			sl := sWall
			switch {
			case striped && dx > 0:
				sl = sRoofShade
			case striped:
				sl = sRoof
			case dx > 0:
				sl = sWallShade
			}
			s.put(4+dx, y, '█', sl, sNone)
		}
		if i%4 == 2 {
			s.put(4, y, '■', sWin, s.at(4, y).fg)
		}
	}
	top := base - (h - 3)
	s.text(2, top, "▀▀▀▀▀", sMetalDark, sNone)
	s.put(3, top-1, '▐', sMetal, sNone)
	s.put(4, top-1, '█', sGlow, sNone)
	s.put(5, top-1, '▌', sMetal, sNone)
	s.put(4, top-2, '▲', sMetalDark, sNone)
	s.beacons = append(s.beacons, pt{4, top - 3})
	return s.trimTop()
}

func glassHall(h int) *sprite {
	w := 29
	s := newSprite(w, h+3)
	base := s.h - 1
	bh := max(3, h-3)
	for y := base - bh + 1; y <= base; y++ {
		for x := 0; x < w; x++ {
			if x%2 == 0 {
				s.put(x, y, '▐', sMetal, sGlass)
			} else {
				s.put(x, y, '▀', sGlassHi, sGlass)
			}
		}
	}
	s.hline(0, w-1, base-bh, '▄', sGlassHi, sNone)
	cx := w / 2
	for dx := -5; dx <= 5; dx++ {
		hh := int(float64(3.2 * math.Sqrt(1-float64(dx*dx)/36)))
		for k := 0; k <= hh; k++ {
			ch := '█'
			if k == hh {
				ch = '▄'
			}
			sl := sGlass
			if (dx+5)%3 == 0 {
				sl = sMetal
			}
			s.put(cx+dx, base-bh-k, ch, sl, sNone)
		}
	}
	s.put(cx, base-bh-4, '►', sNeon1, sNone)
	return s.trimTop()
}

// lattice: the iron tower on four splayed legs.
func lattice(h, w int) *sprite {
	s := newSprite(w, h+2)
	base := s.h - 1
	cx := w / 2
	for i := 0; i < h; i++ {
		y := base - i
		u := 1 - float64(i)/float64(h)
		half := int(float64(float64(w/2)*float64(u*u)) * (0.85 + float64(0.15*u)))
		if half < 1 && i < h-3 {
			half = 1
		}
		if half == 0 {
			s.put(cx, y, '║', sMetal, sNone)
			continue
		}
		l, rr := cx-half, cx+half
		s.put(l, y, '╱', sMetal, sNone)
		s.put(rr, y, '╲', sMetalDark, sNone)
		for x := l + 1; x < rr; x++ {
			if i < h/6 && x > cx-half/2 && x < cx+half/2 {
				continue // the great arch between the legs
			}
			if (x+i)%2 == 0 {
				s.put(x, y, '╳', sMetalDark, sNone)
			}
		}
		if i == h/6 || i == h*2/5 || i == h*7/10 {
			s.hline(l, rr, y, '═', sMetal, sNone)
		}
	}
	s.put(cx, base-h, '│', sMetal, sNone)
	s.beacons = append(s.beacons, pt{cx, base - h - 1})
	return s.trimTop()
}

// dam: seen from downstream, a banded concrete face filling a V canyon,
// the lake glinting over the crest, towers on the crest, white water down
// the spillways and the powerhouse at the foot.
func dam(h int) *sprite {
	w := 31
	h = max(6, h)
	s := newSprite(w, h+2)
	base := s.h - 1
	top := 2
	cx := w / 2
	gapTop, gapBot := 12, 4
	rows := base - top
	for y := top; y <= base; y++ {
		d := y - top
		half := gapTop - (gapTop-gapBot)*d/max(1, rows)
		for x := 0; x < w; x++ {
			dx := x - cx
			switch {
			case dx < -half || dx > half:
				sl := sRock
				if dx > 0 {
					sl = sRockDark
				}
				if hash(x, y, 3)%5 == 0 {
					s.put(x, y, '▓', sl, sRockDark)
				} else {
					s.put(x, y, '█', sl, sNone)
				}
			case y == top:
				s.put(x, y, '▄', sTrim, sWater)
			default:
				sl := sWall
				switch {
				case dx < -half/2:
					sl = sWallLit
				case dx > half/2:
					sl = sWallShade
				}
				if d%3 == 0 {
					s.put(x, y, '▀', sTrim, sl)
				} else {
					s.put(x, y, '█', sl, sNone)
				}
			}
		}
		if y > top && y < base-1 {
			s.put(cx-half+1, y, '║', sWater, s.at(cx-half+1, y).fg)
			s.put(cx+half-1, y, '║', sWater, s.at(cx+half-1, y).fg)
		}
	}
	for x := cx - gapTop - 2; x <= cx+gapTop+2; x++ {
		s.put(x, top-1, '▄', sWater, sNone)
	}
	for _, tx := range []int{cx - 6, cx - 3, cx + 3, cx + 6} {
		s.put(tx, top-1, '█', sTrim, sNone)
		s.put(tx, top-2, '▄', sTrim, sNone)
	}
	s.body(cx-4, base-1, 9, 2)
	s.windows(cx-3, base-1, 7, 1, 2, 1, 0, 0, '▀')
	s.put(cx, base, '▄', sGlow, sWallDark)
	s.beacons = append(s.beacons, pt{cx - 6, top - 3}, pt{cx + 6, top - 3})
	return s
}

// accelerator: the collider ring laid out on the ground in perspective,
// the lab hall at its heart and the detector bulging at the front, with
// beam lights running round it.
func accelerator(h int) *sprite {
	w := 33
	ringH := max(3, h/2)
	s := newSprite(w, ringH+8)
	base := s.h - 1
	rx := float64(w)/2 - 0.5
	ry := float64(ringH)
	cy := float64(2*base) - ry
	cx := float64(w) / 2
	for x := 0; x < w; x++ {
		dx := (float64(x) + 0.5 - cx) / rx
		if dx < -1 || dx > 1 {
			continue
		}
		dy := float64(ry * math.Sqrt(1-float64(dx*dx)))
		back, front := int(cy-dy), int(cy+dy)
		s.pset(x, back, sMetalDark)
		s.pset(x, front, sMetal)
		s.pset(x, front-1, sMetal)
	}
	hx := w/2 - 4
	hb := base - ringH/2
	s.body(hx, hb-3, 9, 4)
	s.windows(hx+1, hb-2, 7, 2, 2, 1, 0, 0, '▀')
	s.dome(hx+1, 7, hb-4, sMetal)
	s.put(w/2, base, '◘', sGlow, sMetalDark)
	s.put(w/2-1, base, '▐', sMetal, sNone)
	s.put(w/2+1, base, '▌', sMetal, sNone)
	s.beacons = append(s.beacons, pt{2, base - ringH/2}, pt{w - 3, base - ringH/2}, pt{w/2 - 8, base})
	return s.trimTop()
}

func mast(h int) *sprite {
	w := 11
	s := newSprite(w, h+2)
	base := s.h - 1
	cx := w / 2
	for i := 0; i < h; i++ {
		y := base - i
		half := int(float64(3 * (1 - float64(i)/float64(h))))
		s.put(cx-half, y, '╱', sMetal, sNone)
		s.put(cx+half, y, '╲', sMetalDark, sNone)
		if half > 0 && i%2 == 0 {
			s.put(cx, y, '╳', sMetal, sNone)
		}
		if half == 0 {
			s.put(cx, y, '║', sMetal, sNone)
		}
		if i == h/3 || i == h*2/3 {
			s.text(cx-4, y, "▀▄▄▀", sMetal, sNone)
			s.text(cx+1, y, "▀▄▄▀", sMetal, sNone)
		}
	}
	s.beacons = append(s.beacons, pt{cx, base - h}, pt{cx, base - h/2})
	s.body(1, base-1, w-2, 2)
	s.windows(1, base-1, w-2, 2, 2, 1, 0, 0, '▀')
	return s.trimTop()
}

// dataCube: the World Simulation, a black cube with data running down it.
func dataCube(h int) *sprite {
	w := 15
	s := newSprite(w, h+1)
	base := s.h - 1
	for y := base - h + 1; y <= base; y++ {
		for x := 0; x < w; x++ {
			sl := sInk
			if x == 0 {
				sl = sWallShade
			}
			s.put(x, y, '█', sl, sNone)
		}
	}
	glyphs := []rune("01ƒΣ≡∞¥")
	for x := 2; x < w-1; x += 2 {
		for y := base - h + 2; y < base; y++ {
			if hash(x, y)%3 == 0 {
				s.put(x, y, glyphs[hash(x, y, 1)%uint64(len(glyphs))], sNeon2, sInk)
			}
		}
	}
	s.hline(0, w-1, base-h+1, '▀', sNeon1, sInk)
	return s
}

func citadel(h int) *sprite {
	w := 25
	s := newSprite(w, h+3)
	base := s.h - 1
	for _, st := range []struct{ in, top int }{{0, h / 3}, {3, h * 2 / 3}, {6, h * 9 / 10}, {9, h}} {
		y0 := base - st.top + 1
		for y := y0; y <= base; y++ {
			for x := st.in; x < w-st.in; x++ {
				sl := sWall
				if x > w/2+2 {
					sl = sWallShade
				}
				s.put(x, y, '█', sl, sNone)
				if (y-y0)%2 == 1 && x > st.in && x < w-st.in-1 && hash(x, y)%4 != 0 {
					s.put(x, y, '▄', sWin, sl)
				}
			}
		}
		s.hline(st.in, w-st.in-1, y0, '▀', sNeon1, sWall)
		s.put(st.in, y0, '▐', sNeon2, sNone)
		s.put(w-st.in-1, y0, '▌', sNeon2, sNone)
	}
	cx := w / 2
	for i, c := range "CITADEL" {
		if y := base - h + 3 + i; y < base {
			s.put(cx, y, c, sNeon2, sInk)
		}
	}
	s.put(cx, base-h, '▲', sNeon1, sNone)
	s.put(cx, base-h-1, '│', sMetal, sNone)
	s.beacons = append(s.beacons, pt{cx, base - h - 2})
	return s.trimTop()
}

// cradle: a ship standing in its launch arch, gantries either side.
func cradle(h int) *sprite {
	w := 21
	s := newSprite(w, h+2)
	base := s.h - 1
	cx := w / 2
	for i := 0; i < h; i++ {
		y := base - i
		t := float64(i) / float64(h)
		half := int(float64(float64(w/2-1) * math.Sqrt(math.Max(0, 1-float64(t*t)))))
		s.put(cx-half, y, '█', sTrim, sNone)
		s.put(cx+half, y, '█', sWallShade, sNone)
		if half > 1 {
			s.put(cx-half+1, y, '▌', sMetal, sNone)
			s.put(cx+half-1, y, '▐', sMetalDark, sNone)
		}
		if i%4 == 2 && half > 3 {
			s.put(cx-half+2, y, '═', sNeon1, sNone)
			s.put(cx+half-2, y, '═', sNeon1, sNone)
		}
	}
	for y := base - h*2/3; y < base-1; y++ {
		s.put(cx, y, '█', sWall, sNone)
		if y > base-h/2 {
			s.put(cx-1, y, '▐', sWallShade, sNone)
			s.put(cx+1, y, '▌', sWallShade, sNone)
		}
	}
	s.put(cx, base-h*2/3-1, '▲', sWall, sNone)
	s.put(cx, base-h/3, '■', sNeon1, sWall)
	s.put(cx, base-1, '▀', sGlow, sNone)
	s.hline(0, w-1, base, '▀', sMetalDark, sNone)
	s.beacons = append(s.beacons, pt{cx, base - h})
	return s.trimTop()
}

// dyson: a captive star half wrapped in lattice bands, held up by a mast.
func dyson(h int) *sprite {
	w := h*2 + 3
	s := newSprite(w, h+2)
	base := s.h - 1
	cx := float64(w) / 2
	R := float64(h) * 0.42
	cyp := float64(2*(base-2)) - float64(2*R)
	for py := 0; py < 2*base; py++ {
		for x := 0; x < w; x++ {
			dx := (float64(x) + 0.5 - cx) / 2
			dy := (float64(py) + 0.5 - cyp) / 2
			d := math.Sqrt(float64(dx*dx) + float64(dy*dy))
			switch {
			case d <= R*0.45:
				s.pset(x, py, sGlow)
			case d <= R*0.6:
				s.pset(x, py, sNeon2)
			case d > R*0.86 && d <= R && (dx < R*0.3 || py%4 < 2):
				sl := sMetal
				if dx > 0 {
					sl = sMetalDark
				}
				s.pset(x, py, sl)
			}
		}
	}
	for y := int(cyp/2 + R/2 + 1); y < base; y++ {
		s.put(w/2, y, '╫', sMetal, sNone)
	}
	s.text(w/2-3, base, "▄▄███▄▄", sMetalDark, sNone)
	return s.trimTop()
}

// ringGate: the warp nexus, a ring on two pylons with light inside it.
func ringGate(h int) *sprite {
	w := h*2 + 1
	s := newSprite(w, h+3)
	cx := w / 2
	cy := float64(h) / 2
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dx := float64(x-cx) / float64(w/2)
			dy := (float64(y) - cy) / cy
			d := math.Sqrt(float64(dx*dx) + float64(dy*dy))
			switch {
			case d > 0.8 && d <= 1.0:
				sl := sMetal
				if dx > 0.2 {
					sl = sMetalDark
				}
				s.put(x, y, '█', sl, sNone)
			case d > 0.7:
				s.put(x, y, '▓', sNeon1, sNone)
			case d > 0.5:
				s.put(x, y, '░', sNeon1, sNone)
			case d <= 0.5 && hash(x, y)%3 == 0:
				s.put(x, y, []rune("·∙°")[hash(x, y, 2)%3], sNeon2, sNone)
			}
		}
	}
	for _, px := range []int{cx - h/2, cx + h/2} {
		s.vline(px, h-1, h+1, '█', sMetalDark, sNone)
		s.put(px-1, h+2, '▄', sMetalDark, sNone)
		s.put(px, h+2, '█', sMetalDark, sNone)
		s.put(px+1, h+2, '▄', sMetalDark, sNone)
	}
	s.beacons = append(s.beacons, pt{cx, 0})
	return s.trimTop()
}

// beacon: a ring spire whose crown throws a beam into the sky.
func beacon(h int) *sprite {
	sp := aRingSpire(newRnd(77), h*2/3)
	s := newSprite(sp.w, sp.h+h/3+1)
	s.blit(sp, 0, s.h-sp.h)
	cx := sp.w / 2
	for y := 0; y < s.h-sp.h; y++ {
		ch := '│'
		if y%3 == 0 {
			ch = '║'
		}
		s.put(cx, y, ch, sGlow, sNone)
	}
	s.put(cx-1, s.h-sp.h-1, '◢', sNeon2, sNone)
	s.put(cx+1, s.h-sp.h-1, '◣', sNeon2, sNone)
	return s
}

// anchor: an inverted pyramid pinned to the ground by pillars, shards of
// light orbiting it.
func anchor(h int) *sprite {
	w := 19
	s := newSprite(w, h+1)
	base := s.h - 1
	top := base - h + 1
	ph := h / 2
	for i := 0; i < ph; i++ {
		y := top + i
		half := w/2 - 1 - i*(w/2-1)/max(1, ph)
		for x := w/2 - half; x <= w/2+half; x++ {
			sl := sWall
			if x > w/2 {
				sl = sWallShade
			}
			s.put(x, y, '█', sl, sNone)
		}
		if i%2 == 1 {
			s.put(w/2, y, '◊', sNeon1, sWall)
		}
	}
	for y := top + ph; y < base; y++ {
		s.put(3, y, '│', sMetalDark, sNone)
		s.put(w-4, y, '│', sMetalDark, sNone)
		if y%2 == 0 {
			s.put(w/2, y, '·', sNeon2, sNone)
		}
	}
	for i, p := range []pt{{0, top + 1}, {w - 1, top + 2}, {1, top + ph}, {w - 2, top + ph + 1}} {
		s.put(p.x, p.y, '◆', []slot{sNeon1, sNeon2}[i%2], sNone)
	}
	s.text(2, base, "▄█▄", sMetal, sNone)
	s.text(w-5, base, "▄█▄", sMetal, sNone)
	return s
}

func singularity(h int) *sprite {
	w := h*2 + 5
	s := newSprite(w, h+1)
	cx := w / 2
	cy := float64(h) * 0.45
	R := float64(h) * 0.28
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dx := float64(x-cx) / 2
			dy := float64(y) - cy
			d := math.Sqrt(float64(dx*dx) + float64(dy*dy))
			switch {
			case d <= R:
				s.put(x, y, '█', sInk, sNone)
			case d <= R+1.2:
				s.put(x, y, '▓', sGlow, sInk)
			}
			ex, ey := dx/float64(R*2.6), float64(math.Abs(dy)*4.5)/float64(R*2.6)
			if float64(ex*ex)+float64(ey*ey) <= 1 && d > R {
				ch := '▀'
				if dy > 0 {
					ch = '▄'
				}
				s.put(x, y, ch, sNeon2, sNone)
			}
		}
	}
	s.text(cx-2, h, "▄███▄", sMetalDark, sNone)
	return s.trimTop()
}
