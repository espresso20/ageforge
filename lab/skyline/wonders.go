package main

import "math"

// wonders.go: the 22 landmarks. Each is a short drawing routine on the same
// cell grammar as the buildings (a few are hand-drawn art), sized by the
// scene scale so a small terminal still gets the whole silhouette.

func wonderSprite(key string, scale float64) *Sprite {
	sz := func(n float64) int { return max(3, int(n*scale+0.5)) }
	r := newRnd(len(key), int(key[0]))
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
			s.put(2, y, '█', SInk, SNone)
			s.put(3, y, '█', SInk, SNone)
			s.put(4, y, '█', SRockDark, SNone)
			if y%3 == 1 {
				s.put(3, y, '≡', SGlow, SInk)
			}
		}
		s.text(2, 0, "▄▄▄", SInk, SNone)
		s.text(0, h, "▄▄███▄▄", SRock, SNone)
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
		return arcade(sz(8), 21, 3, true)
	case "parthenon":
		s := aClassical(newRnd(9), 9)
		return widen(s, 2)
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
		return dam(sz(9))
	case "particle_accelerator":
		return accelerator()
	case "space_program":
		s := aLaunch(r, sz(20))
		return s
	case "global_network":
		return mast(sz(24))
	case "world_simulation":
		return monolithCube(sz(18))
	case "neon_citadel":
		return citadel(sz(28))
	case "stellar_cradle":
		return cradle(sz(22))
	case "dyson_scaffold":
		return lattice(sz(24), 11)
	case "warp_nexus":
		return ring(sz(16))
	case "cosmic_beacon":
		s := aSpire(r, sz(26))
		return s
	case "reality_anchor":
		return anchor(sz(20))
	case "singularity_core":
		return singularity(sz(22))
	}
	return aObelisk(r, sz(10))
}

// arcade: tiers of arches (the Colosseum), the top tier ruined on one side.
func arcade(h, w, tiers int, ruined bool) *Sprite {
	th := max(2, h/tiers)
	s := newSprite(w, th*tiers+1)
	base := s.H - 1
	for t := 0; t < tiers; t++ {
		y1 := base - t*th
		y0 := y1 - th + 1
		for y := y0; y <= y1; y++ {
			for x := 0; x < w; x++ {
				slot := SWall
				if x == 0 {
					slot = SWallLit
				} else if x > w*3/4 {
					slot = SWallShade
				}
				switch {
				case y == y0:
					s.put(x, y, '▄', STrim, SNone)
					if t > 0 || y0 == y1 {
						s.put(x, y, '█', STrim, SNone)
					}
				case x%2 == 1:
					s.put(x, y, '▀', slot, SWallDark)
				default:
					s.put(x, y, '█', slot, SNone)
				}
			}
		}
	}
	if ruined {
		top := base - tiers*th + 1
		for x := w / 2; x < w; x++ {
			cut := (x - w/2) * th / max(1, w/2)
			for y := top; y < top+cut && y < s.H; y++ {
				s.put(x, y, 0, SNone, SNone)
			}
		}
	}
	return s.trimTop()
}

// widen stretches a sprite by repeating its middle column n times.
func widen(s *Sprite, n int) *Sprite {
	o := newSprite(s.W+n*2, s.H)
	mid := s.W / 2
	for y := 0; y < s.H; y++ {
		for x := 0; x < o.W; x++ {
			sx := x
			switch {
			case x > mid+n*2:
				sx = x - n*2
			case x > mid:
				sx = mid - (x-mid)%2
			}
			c := s.at(sx, y)
			o.put(x, y, c.Ch, c.Fg, c.Bg)
		}
	}
	return o
}

func library(h int) *Sprite {
	w := 21
	s := newSprite(w, h+2)
	base := s.H - 1
	bh := max(4, h/2)
	s.body(0, base-bh+1, w, bh)
	s.hline(0, w-1, base-bh+1, '▀', STrim, SWall)
	for x := 1; x < w-1; x += 2 {
		for y := base - bh + 2; y < base; y++ {
			s.put(x, y, '█', STrim, SWallDark)
		}
	}
	// towers at both ends
	for _, tx := range []int{0, w - 3} {
		s.body(tx, base-bh-2, 3, bh+2)
		s.crenels(tx, 3, base-bh-3, SWall)
	}
	s.body(5, base-bh-1, 11, 2)
	d := s.dome(5, 11, base-bh-2, SMetal)
	s.put(10, base-bh-2-d, '▲', STrim, SNone)
	s.put(10, base, '▄', SGlow, SWallDark)
	return s.trimTop()
}

func basilica(h int) *Sprite {
	w := 23
	s := newSprite(w, h+3)
	base := s.H - 1
	bh := max(4, h*4/10)
	s.body(0, base-bh+1, w, bh)
	s.hline(0, w-1, base-bh+1, '▀', STrim, SWall)
	for x := 1; x < w-1; x += 2 {
		s.put(x, base-1, '█', STrim, SWallDark)
		s.put(x, base-2, '█', STrim, SWallDark)
	}
	s.windows(1, base-bh+2, w-2, max(0, bh-4), 2, 1, 0, 0, '■')
	s.gable(4, 15, base-bh, false)
	// drum, dome, lantern
	dy := base - bh - 3
	s.body(6, dy-1, 11, 3)
	s.windows(6, dy-1, 11, 2, 2, 1, 1, 0, '▌')
	dh := s.dome(5, 13, dy-2, SMetal)
	for y := dy - 2 - dh; y > dy-2-dh-max(1, h/8); y-- {
		s.put(11, y, '█', STrim, SNone)
	}
	s.put(11, dy-3-dh-max(1, h/8)+1, '▲', STrim, SNone)
	s.put(11, dy-4-dh-max(1, h/8)+1, '┼', STrim, SNone)
	return s.trimTop()
}

func lighthouse(h int) *Sprite {
	w := 9
	s := newSprite(w, h+2)
	base := s.H - 1
	for i := 0; i < h-3; i++ {
		y := base - i
		wd := 3 - i*2/max(1, h-3)
		for dx := -wd; dx <= wd; dx++ {
			slot := SWall
			if (i/2)%2 == 1 {
				slot = SRoof
			}
			if dx > 0 {
				if slot == SWall {
					slot = SWallShade
				} else {
					slot = SRoofShade
				}
			}
			s.put(4+dx, y, '█', slot, SNone)
		}
		if i%4 == 2 {
			s.put(4, y, '■', SWin, s.at(4, y).Fg)
		}
	}
	top := base - (h - 3)
	s.text(2, top, "▀▀▀▀▀", SMetalDark, SNone)
	s.put(3, top-1, '▐', SMetal, SNone)
	s.put(4, top-1, '█', SGlow, SNone)
	s.put(5, top-1, '▌', SMetal, SNone)
	s.put(4, top-2, '▲', SMetalDark, SNone)
	s.Beacons = append(s.Beacons, pt{4, top - 3})
	return s.trimTop()
}

func glassHall(h int) *Sprite {
	w := 29
	s := newSprite(w, h+3)
	base := s.H - 1
	bh := max(3, h-3)
	for y := base - bh + 1; y <= base; y++ {
		for x := 0; x < w; x++ {
			if x%2 == 0 {
				s.put(x, y, '▐', SMetal, SGlass)
			} else {
				s.put(x, y, '▀', SGlassHi, SGlass)
			}
		}
	}
	// barrel vault over the nave, a taller one over the transept
	for x := 0; x < w; x++ {
		s.put(x, base-bh, '▄', SGlassHi, SNone)
	}
	cx := w / 2
	for dx := -5; dx <= 5; dx++ {
		hh := int(3.2 * math.Sqrt(1-float64(dx*dx)/36))
		for k := 0; k <= hh; k++ {
			y := base - bh - k
			ch := '█'
			if k == hh {
				ch = '▄'
			}
			slot := SGlass
			if (dx+5)%3 == 0 {
				slot = SMetal
			}
			s.put(cx+dx, y, ch, slot, SNone)
		}
	}
	s.put(cx, base-bh-4, '►', SNeon1, SNone)
	return s.trimTop()
}

// lattice: the Eiffel-style iron tower (also the Dyson scaffold's anchor).
func lattice(h, w int) *Sprite {
	s := newSprite(w, h+2)
	base := s.H - 1
	cx := w / 2
	for i := 0; i < h; i++ {
		y := base - i
		t := float64(i) / float64(h)
		half := int(float64(w/2) * math.Pow(1-t, 2.2))
		if half < 1 && i < h-3 {
			half = 1
		}
		l, r := cx-half, cx+half
		if half == 0 {
			s.put(cx, y, '║', SMetal, SNone)
			continue
		}
		s.put(l, y, '╱', SMetal, SNone)
		s.put(r, y, '╲', SMetalDark, SNone)
		for x := l + 1; x < r; x++ {
			if i < h/6 && x > cx-half/2 && x < cx+half/2 {
				continue // the great arch between the legs
			}
			if (x+i)%2 == 0 {
				s.put(x, y, '╳', SMetalDark, SNone)
			}
		}
		if i == h/6 || i == h*2/5 || i == h*7/10 {
			s.hline(l, r, y, '═', SMetal, SNone)
		}
	}
	s.put(cx, base-h, '│', SMetal, SNone)
	s.Beacons = append(s.Beacons, pt{cx, base - h - 1})
	return s.trimTop()
}

// dam: side-on, a gravity wedge between canyon walls, the reservoir brim
// full on the upstream side, the powerhouse and tailrace below.
func dam(h int) *Sprite {
	w := 29
	s := newSprite(w, h+2)
	base := s.H - 1
	top := base - h + 1
	up := 11 // the vertical upstream face
	for y := top; y <= base; y++ {
		d := y - top
		// canyon walls, ragged
		lw := 2 + d*3/h + int(hash(y, 1)%2)
		for x := 0; x < lw; x++ {
			s.put(x, y, '█', SRock, SNone)
		}
		rw := 2 + d*3/h + int(hash(y, 2)%2)
		for x := w - rw; x < w; x++ {
			s.put(x, y, '█', SRockDark, SNone)
		}
		// the reservoir behind the dam
		if d >= 1 {
			for x := lw; x < up; x++ {
				ch := ' '
				if d == 1 {
					ch = '▀'
				}
				s.put(x, y, ch, STrim, SWater)
			}
		}
		// the wedge
		dw := 2 + d*9/max(1, h)
		for x := up; x < up+dw && x < w-rw; x++ {
			slot := SWall
			if x == up {
				slot = SWallLit
			} else if x >= up+dw-2 {
				slot = SWallShade
			}
			s.put(x, y, '█', slot, SNone)
		}
		if up+dw < w-rw && d == 0 {
			s.put(up+dw, y, '▄', SWallShade, SNone)
		}
	}
	s.hline(up-1, up+2, top-1, '▄', STrim, SNone) // the roadway on the crest
	// intake towers standing in the lake
	for _, tx := range []int{5, 8} {
		for y := top - 1; y < top+h/2; y++ {
			s.put(tx, y, '█', STrim, SNone)
		}
		s.put(tx, top-2, '▄', STrim, SNone)
		s.put(tx, top, '■', SWin, STrim)
	}
	// powerhouse and tailrace at the foot
	ph := up + 10
	s.body(ph, base-1, 5, 2)
	s.windows(ph, base-1, 5, 1, 2, 1, 1, 0, '▀')
	for x := ph - 2; x < ph; x++ {
		s.put(x, base, '▀', STrim, SWater)
	}
	s.Beacons = append(s.Beacons, pt{up + 1, top - 2})
	return s.trimTop()
}

func accelerator() *Sprite {
	s := art(`
    ▄▄▄▀▀▀▀▀▀▀▀▀▀▀▀▄▄▄
 ▄▀▀  ▄▄▄▄▄▄▄▄▄▄▄▄   ▀▀▄
▐▌  ▄█▀▀▀▀▀▀▀▀▀▀▀▀█▄   ▐▌
 ▀▄▄█  ▒▓█o█▓▒     █▄▄▀
    █▀▀▀▀▀▀▀▀▀▀▀▀▀▀█
    █ ■ ■ ■ ▄ ■ ■ ■█`, `
    mmmmmmmmmmmmmmmmmm
 mmm  wwwwwwwwwwww   mmm
mm  wwwwwwwwwwwwwwww   MM
 mmmw  nnnononn     wMMM
    wtttttttttttttts
    wwlwlwlwdwlwlwls`, `
    ..................
 ...  ............   ...
..  .................   ..
 ....dddddddddddddddd....
    ...............d
    w.w.w.w.w.w.w.w.`)
	return s
}

func mast(h int) *Sprite {
	w := 11
	s := newSprite(w, h+2)
	base := s.H - 1
	cx := w / 2
	for i := 0; i < h; i++ {
		y := base - i
		half := int(3 * (1 - float64(i)/float64(h)))
		s.put(cx-half, y, '╱', SMetal, SNone)
		s.put(cx+half, y, '╲', SMetalDark, SNone)
		if half > 0 && i%2 == 0 {
			s.put(cx, y, '╳', SMetal, SNone)
		}
		if half == 0 {
			s.put(cx, y, '║', SMetal, SNone)
		}
		if i == h/3 || i == h*2/3 {
			s.text(cx-3, y, "▀▄▄▀", SMetal, SNone)
			s.text(cx+1, y, "▀▄▄▀", SMetal, SNone)
		}
	}
	s.Beacons = append(s.Beacons, pt{cx, base - h}, pt{cx, base - h/2})
	s.body(1, base-1, w-2, 2)
	s.windows(1, base-1, w-2, 2, 2, 1, 0, 0, '▀')
	return s.trimTop()
}

func monolithCube(h int) *Sprite {
	w := 15
	s := newSprite(w, h+1)
	base := s.H - 1
	for y := base - h + 1; y <= base; y++ {
		for x := 0; x < w; x++ {
			slot := SInk
			if x == 0 {
				slot = SWallShade
			}
			s.put(x, y, '█', slot, SNone)
		}
	}
	// data rain on the face: neon glyph columns
	for x := 2; x < w-1; x += 2 {
		for y := base - h + 2; y < base; y++ {
			if hash(x, y)%3 == 0 {
				s.put(x, y, []rune("01ƒΣ≡∞¥")[hash(x, y, 1)%7], SNeon2, SInk)
			}
		}
	}
	s.hline(0, w-1, base-h+1, '▀', SNeon1, SInk)
	return s
}

func citadel(h int) *Sprite {
	w := 25
	s := newSprite(w, h+3)
	base := s.H - 1
	steps := []struct{ in, top int }{{0, h / 3}, {3, h * 2 / 3}, {6, h * 9 / 10}, {9, h}}
	for _, st := range steps {
		y0 := base - st.top + 1
		for y := y0; y <= base; y++ {
			for x := st.in; x < w-st.in; x++ {
				slot := SWall
				if x > w/2+2 {
					slot = SWallShade
				}
				s.put(x, y, '█', slot, SNone)
				if (y-y0)%2 == 1 && x > st.in && x < w-st.in-1 && hash(x, y)%4 != 0 {
					s.put(x, y, '▄', SWin, slot)
				}
			}
		}
		s.hline(st.in, w-st.in-1, y0, '▀', SNeon1, SWall)
		s.put(st.in, y0, '▐', SNeon2, SNone)
		s.put(w-st.in-1, y0, '▌', SNeon2, SNone)
	}
	cx := w / 2
	for i, r := range "CITADEL" {
		if base-h+3+i < base {
			s.put(cx, base-h+3+i, r, SNeon2, SInk)
		}
	}
	s.put(cx, base-h, '▲', SNeon1, SNone)
	s.put(cx, base-h-1, '│', SMetal, SNone)
	s.Beacons = append(s.Beacons, pt{cx, base - h - 2})
	return s.trimTop()
}

func cradle(h int) *Sprite {
	w := 19
	s := newSprite(w, h+2)
	base := s.H - 1
	cx := w / 2
	// the arch
	for i := 0; i < h; i++ {
		y := base - i
		t := float64(i) / float64(h)
		half := int(float64(w/2) * math.Sqrt(math.Max(0, 1-t*t)))
		s.put(cx-half, y, '█', STrim, SNone)
		s.put(cx+half, y, '█', SWallShade, SNone)
		if half > 0 {
			s.put(cx-half+1, y, '▌', SMetal, SNone)
		}
	}
	// the ship in its cradle
	for y := base - h*2/3; y < base-1; y++ {
		s.put(cx, y, '█', SWall, SNone)
		if y > base-h/2 {
			s.put(cx-1, y, '▐', SWallShade, SNone)
			s.put(cx+1, y, '▌', SWallShade, SNone)
		}
	}
	s.put(cx, base-h*2/3-1, '▲', SWall, SNone)
	s.put(cx, base-h/3, '■', SNeon1, SWall)
	s.hline(0, w-1, base, '▀', SMetalDark, SNone)
	s.Beacons = append(s.Beacons, pt{cx, base - h})
	return s.trimTop()
}

func ring(h int) *Sprite {
	w := h*2 + 1
	s := newSprite(w, h+2)
	cx := w / 2
	cy := float64(h) / 2
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dx := float64(x-cx) / float64(w/2)
			dy := (float64(y) - cy) / cy
			d := math.Hypot(dx, dy)
			switch {
			case d > 0.8 && d <= 1.0:
				slot := SMetal
				if dx > 0.2 {
					slot = SMetalDark
				}
				s.put(x, y, '█', slot, SNone)
			case d <= 0.8 && d > 0.74:
				s.put(x, y, '░', SNeon1, SNone)
			case d <= 0.5 && hash(x, y)%3 == 0:
				s.put(x, y, []rune("·∙°")[hash(x, y, 2)%3], SNeon2, SNone)
			}
		}
	}
	s.text(cx-3, h, "▄█▀ ▀█▄", SMetalDark, SNone)
	s.text(cx-4, h+1, "▀▀▀▀▀▀▀▀▀", SMetalDark, SNone)
	return s.trimTop()
}

func anchor(h int) *Sprite {
	w := 17
	s := newSprite(w, h+1)
	base := s.H - 1
	top := base - h + 1
	ph := h / 2
	for i := 0; i < ph; i++ {
		y := top + i
		half := w/2 - i*(w/2)/ph
		for x := w/2 - half; x <= w/2+half; x++ {
			slot := SWall
			if x > w/2 {
				slot = SWallShade
			}
			s.put(x, y, '█', slot, SNone)
		}
		if i%2 == 1 {
			s.put(w/2, y, '◊', SNeon1, SWall)
		}
	}
	for y := top + ph; y < base; y++ {
		s.put(2, y, '│', SMetalDark, SNone)
		s.put(w-3, y, '│', SMetalDark, SNone)
		if (y+1)%2 == 0 {
			s.put(w/2, y, '·', SNeon2, SNone)
		}
	}
	s.text(1, base, "▄█▄", SMetal, SNone)
	s.text(w-4, base, "▄█▄", SMetal, SNone)
	return s
}

func singularity(h int) *Sprite {
	w := h*2 + 5
	s := newSprite(w, h+1)
	cx := w / 2
	cy := float64(h) * 0.45
	R := float64(h) * 0.28
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dx := float64(x-cx) / 2
			dy := float64(y) - cy
			d := math.Hypot(dx, dy)
			switch {
			case d <= R:
				s.put(x, y, '█', SInk, SNone)
			case d <= R+1.2:
				s.put(x, y, '▓', SGlow, SInk)
			}
			// the accretion disk, seen nearly edge-on
			ey := math.Abs(dy) * 4.5
			if math.Hypot(dx/(R*2.6), ey/(R*2.6)) <= 1 && math.Hypot(dx, dy) > R {
				ch := '▀'
				if dy > 0 {
					ch = '▄'
				}
				s.put(x, y, ch, SNeon2, SNone)
			}
		}
	}
	s.text(cx-2, h, "▄███▄", SMetalDark, SNone)
	return s.trimTop()
}
