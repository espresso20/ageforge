package skyline

import "math"

// wonders_mid.go: the landmarks from the Renaissance to the Atomic Age.

// ribbedDome sits a dome of sMetal w wide on row y (its bottom row) with
// gilded ribs every third column, and returns its height in cells.
func (s *sprite) ribbedDome(x, w, y int) int {
	dh := s.dome(x, w, y, sMetal)
	for yy := y - dh; yy <= y; yy++ {
		for xx := x; xx < x+w; xx++ {
			if (xx-x)%3 == 1 {
				if c := s.at(xx, yy); c.ch != 0 && (c.fg == sMetal || c.fg == sMetalDark) {
					s.put(xx, yy, c.ch, sTrim, c.bg)
				}
			}
		}
	}
	return dh
}

// basilica: the Sistine Chapel, a long nave under a great ribbed dome on its
// drum with a lantern and a cross, a bell tower with a spire at each end,
// the rose window and the glazing lit at night.
func basilica(h int) *sprite {
	w := atLeast(h, 27) | 1
	s := newSprite(w, h)
	base := h - 1
	cx := w / 2
	u := unit(h, 30)
	nav := atLeast(u(10), 4)
	drum := atLeast(u(7), 2)
	lan := atLeast(u(4), 2)
	ty := base - nav + 1 // the nave's top row
	// the nave: pilasters, two tiers of arched windows
	for y := ty; y <= base; y++ {
		for x := 3; x < w-3; x++ {
			sl := sWall
			switch {
			case x == 3:
				sl = sWallLit
			case x == w-4:
				sl = sWallShade
			}
			s.put(x, y, '█', sl, sNone)
		}
	}
	s.hline(2, w-3, ty, '█', sTrim, sNone)
	s.hline(3, w-4, base, '▄', sTrim, sNone)
	for x := 5; x < w-5; x += 3 {
		if x > cx-3 && x < cx+3 {
			continue
		}
		for y := ty + 2; y < base-1; y++ {
			if (y-ty)%4 < 3 {
				s.put(x, y, '▌', sWin, sWall)
				s.put(x+1, y, '▐', sWin, sWall)
			}
		}
	}
	// the great door and the rose window over it
	s.block(cx-1, cx+1, base-1, base, sWallDark)
	s.put(cx, base-1, '▄', sGlow, sWallDark)
	if nav >= 6 {
		ellipse(float64(cx)+0.5, float64(2*ty+5), 2.4, 2.4, func(x, py int, dx, dy float64) {
			if dx*dx+dy*dy > 0.35 || hash(x, py, 2)%5 != 0 {
				s.pset(x, py, sGlow)
			}
		})
	}
	// the drum with its pilasters and windows
	dx0, dx1 := cx-u(6), cx+u(6)
	dy := ty - drum
	s.block(dx0, dx1, dy, ty-1, sWall)
	for x := dx0 + 1; x < dx1; x += 2 {
		for y := dy + 1; y < ty-1; y++ {
			s.put(x, y, '▌', sWin, sWall)
		}
	}
	s.hline(dx0-1, dx1+1, dy, '▄', sTrim, sNone)
	s.hline(dx0-1, dx1+1, ty-1, '▄', sTrim, sNone)
	// the dome, the lantern, the cross
	dh := s.ribbedDome(dx0, dx1-dx0+1, dy-1)
	ly := dy - 1 - dh
	for y := ly; y > ly-lan; y-- {
		s.put(cx, y, '█', sGlow, sNone)
		s.put(cx-1, y, '▐', sTrim, sNone)
		s.put(cx+1, y, '▌', sTrim, sNone)
	}
	s.put(cx, ly-lan, '▲', sTrim, sNone)
	s.put(cx, ly-lan-1, '┼', sTrim, sNone)
	// a bell tower with a spire at each end
	th := atLeast(u(20), 6)
	for _, tx := range []int{0, w - 5} {
		for y := base - th + 1; y <= base; y++ {
			for x := tx; x < tx+5; x++ {
				sl := sWall
				switch {
				case x == tx:
					sl = sWallLit
				case x == tx+4:
					sl = sWallShade
				}
				s.put(x, y, '█', sl, sNone)
			}
			if (base-y)%4 == 2 && y > base-th+4 {
				s.put(tx+2, y, '▌', sWin, sWall)
				s.put(tx+3, y, '▐', sWin, sWall)
			}
		}
		s.hline(tx-0, tx+4, base-th, '▀', sTrim, sNone)
		s.cone(tx, 5, 2*(base-th)-2*atLeast(u(7), 3), 2*(base-th)-1)
		s.put(tx+2, base-th-atLeast(u(7), 3)-1, '┼', sTrim, sNone)
	}
	// close: carved pilasters and the window bars
	for x := 5; x < w-5; x += 3 {
		for y := ty + 1; y < base-1; y += 2 {
			if !(x > cx-3 && x < cx+3) {
				s.detail(x-1, y, '▌', sWallShade)
			}
		}
	}
	for x := dx0 + 1; x < dx1; x += 2 {
		for y := dy + 1; y < ty-1; y++ {
			s.detail(x, y, '▖', sTrim)
		}
	}
	return s
}

// lighthouse: the Grand Lighthouse on its rock, a striped tapering tower
// with a gallery, a glazed lantern and a roof, the keeper's cottage at the
// foot; at night the lamp throws a beam across the bay.
func lighthouse(h int) *sprite {
	w := atLeast(h*5/8, 15) | 1
	s := newSprite(w, h)
	base := h - 1
	cx := w/2 - 1
	u := unit(h, 32)
	rock := atLeast(u(5), 2)
	lan := atLeast(u(4), 2)
	roof := 2
	gal := 1
	tower := h - rock - lan - roof - gal
	if tower < 3 {
		tower = 3
	}
	// the rock, with foam at its foot
	for y := base - rock + 1; y <= base; y++ {
		d := y - (base - rock + 1) // 0 at the top
		half := w/2 - 1 - (rock-1-d)*(w/2-3)/atLeast(rock, 1)
		half = clampInt(half, 3, w/2)
		for x := cx - half; x <= cx+half+1; x++ {
			sl := sRock
			if x > cx || hash(x, y, 4)%3 == 0 {
				sl = sRockDark
			}
			s.put(x, y, '█', sl, sNone)
		}
	}
	for x := 0; x < w; x++ {
		if hash(x, 3, 8)%3 == 0 {
			s.put(x, base, '▄', sWater, sNone)
		}
	}
	// the tower: stripes, a door, a few windows
	ty := base - rock
	for i := 0; i < tower; i++ {
		y := ty - i
		half := 3 - i*2/atLeast(tower, 1)
		for dx := -half; dx <= half+1; dx++ {
			striped := (i/3)%2 == 1
			sl := sWall
			switch {
			case striped && dx > 0:
				sl = sRoofShade
			case striped:
				sl = sRoof
			case dx > 0:
				sl = sWallShade
			}
			s.put(cx+dx, y, '█', sl, sNone)
		}
		if i%5 == 3 && i > 1 {
			s.put(cx, y, '▌', sWin, s.at(cx, y).fg)
		}
	}
	s.put(cx, ty, '▀', sWallDark, sWall) // the door
	s.put(cx+1, ty, '▀', sWallDark, sWall)
	// the gallery, the lantern room and its roof
	gy := ty - tower
	s.hline(cx-4, cx+5, gy, '▀', sMetalDark, sNone)
	s.put(cx-4, gy-1, '┬', sMetal, sNone)
	s.put(cx+5, gy-1, '┬', sMetal, sNone)
	for y := gy - lan; y < gy; y++ {
		s.put(cx-2, y, '▐', sMetal, sNone)
		s.put(cx+3, y, '▌', sMetal, sNone)
		for x := cx - 1; x <= cx+2; x++ {
			s.put(x, y, '█', sGlass, sNone)
		}
		s.put(cx, y, '█', sGlow, sNone)
		s.put(cx+1, y, '█', sGlow, sNone)
	}
	ry := gy - lan
	s.hline(cx-2, cx+3, ry-1, '▄', sMetalDark, sNone)
	s.put(cx, ry-2, '▄', sMetalDark, sNone)
	s.put(cx+1, ry-2, '▄', sMetalDark, sNone)
	s.put(cx, ry-3, '│', sMetal, sNone)
	s.beacons = append(s.beacons, pt{cx, max(0, ry-3)})
	s.lamp(cx, gy-lan/2-1, beamFan, atLeast(u(18), 8))
	// the keeper's cottage
	if w >= 19 {
		cxx := cx + w/2 - 3
		s.block(cxx, cxx+3, ty-1, ty, sWall)
		s.put(cxx+1, ty-1, '■', sWin, sWall)
		s.put(cxx, ty-2, '▄', sRoof, sNone)
		s.put(cxx+1, ty-2, '█', sRoof, sNone)
		s.put(cxx+2, ty-2, '█', sRoofShade, sNone)
		s.put(cxx+3, ty-2, '▄', sRoofShade, sNone)
	}
	// close: brickwork courses on the tower, panes in the lantern
	for i := 0; i < tower; i++ {
		y := ty - i
		for dx := -1; dx <= 1; dx++ {
			if (i+dx)%2 == 0 {
				s.detail(cx+dx, y, '▖', sWallDark)
			}
		}
	}
	for y := gy - lan; y < gy; y++ {
		s.detail(cx-1, y, '▌', sMetal)
		s.detail(cx+2, y, '▐', sMetal)
	}
	return s
}

// glassHall: the Crystal Palace, a great barrel-vaulted transept of glass in
// iron ribs between two lower naves of glazing, a brick water tower at each
// end; the panes glow at night.
func glassHall(h int) *sprite {
	w := atLeast(h*3/2, 31) | 1
	s := newSprite(w, h)
	base := h - 1
	cx := w / 2
	u := unit(h, 28)
	nave := atLeast(u(6), 3)
	clere := atLeast(u(3), 1)
	tw := atLeast(u(20), 11) | 1 // the transept's width
	vault := atLeast(u(15), 5)   // its height over the naves' eaves
	tower := atLeast(u(25), 8)
	// the naves, two tiers of glazing with an iron rib every third column
	for x := 5; x < w-5; x++ {
		for y := base - nave + 1; y <= base; y++ {
			switch {
			case (x-5)%3 == 0:
				s.put(x, y, '█', sMetal, sNone)
			case hash(x, y, 2)%6 == 0:
				s.put(x, y, '█', sWin, sNone)
			default:
				s.put(x, y, '█', sGlass, sNone)
			}
		}
		if x > 6 && x < w-7 {
			for y := base - nave - clere + 1; y <= base-nave; y++ {
				if (x-5)%3 == 0 {
					s.put(x, y, '█', sMetal, sNone)
				} else {
					s.put(x, y, '▀', sGlassHi, sGlass)
				}
			}
		}
	}
	s.hline(5, w-6, base-nave-clere, '▄', sMetalDark, sNone)
	// the transept's vault: a half-ellipse of glazing in ribs
	eaves := base - nave - clere
	ellipse(float64(cx)+0.5, float64(2*eaves+1), float64(tw)/2, float64(vault*2), func(x, py int, dx, dy float64) {
		if py > 2*eaves+1 {
			return
		}
		switch {
		case (x-cx)%3 == 0 || dx*dx+dy*dy > 0.86:
			s.pset(x, py, sMetal)
		case hash(x, py, 5)%7 == 0:
			s.pset(x, py, sWin)
		default:
			s.pset(x, py, sGlass)
		}
	})
	// the facade glazing under the vault
	for x := cx - tw/2 + 1; x <= cx+tw/2-1; x++ {
		for y := eaves + 1; y <= base; y++ {
			if y > base-nave {
				continue
			}
		}
	}
	// a pennant on the crown
	s.put(cx, 0, '▸', sNeon1, sNone)
	s.put(cx, 1, '│', sMetal, sNone)
	// the brick water towers at each end
	for _, tx := range []int{0, w - 5} {
		for y := base - tower + 1; y <= base; y++ {
			for x := tx + 1; x < tx+4; x++ {
				sl := sWall
				switch {
				case x == tx+1:
					sl = sWallLit
				case x == tx+3:
					sl = sWallShade
				}
				s.put(x, y, '█', sl, sNone)
			}
			if (base-y)%5 == 3 {
				s.put(tx+2, y, '▌', sWin, sWall)
			}
		}
		ty := base - tower
		s.hline(tx, tx+4, ty, '█', sMetalDark, sNone) // the tank
		s.hline(tx, tx+4, ty-1, '▀', sMetal, sMetalDark)
		s.put(tx+2, ty-2, '▲', sMetalDark, sNone)
	}
	// close: panes in the glazing, brick courses on the towers
	for x := 5; x < w-5; x++ {
		for y := base - nave + 1; y <= base; y++ {
			if c := s.at(x, y); c.ch == '█' && c.fg == sGlass && (x+y)%2 == 0 {
				s.detail(x, y, '▘', sGlassHi)
			}
		}
	}
	for _, tx := range []int{0, w - 5} {
		for y := base - tower + 1; y <= base; y++ {
			s.detail(tx+1+(y%2), y, '▖', sWallDark)
		}
	}
	return s
}

// lattice: the Eiffel Tower, four curved legs round a great arch, three
// platforms, the lattice closing to a spire; a searchlight at the top
// sweeps the sky at night and the legs are strung with lights.
func lattice(h int) *sprite {
	w := atLeast(h*3/4, 17) | 1
	s := newSprite(w, h)
	base := h - 1
	cx := w / 2
	u := unit(h, 34)
	spire := atLeast(u(5), 2)
	body := h - spire
	archH := atLeast(u(6), 2)
	plats := [3]int{u(9), u(18), u(26)}
	half := func(i int) int {
		t := float64(i) / float64(body)
		return 1 + int(float64(w/2-1)*math.Pow(1-t, 2.0)+0.5)
	}
	for i := 0; i < body; i++ {
		y := base - i
		hw := half(i)
		lw := 1
		switch {
		case i < body*28/100:
			lw = 3
		case i < body*58/100:
			lw = 2
		}
		if hw <= 1 {
			s.put(cx, y, '█', sMetal, sNone)
			continue
		}
		lw = min(lw, hw)
		for k := 0; k < lw; k++ {
			sl, sr := sMetal, sMetalDark
			if k == lw-1 && lw > 1 {
				sl, sr = sMetalDark, sMetal
			}
			s.put(cx-hw+k, y, '█', sl, sNone)
			s.put(cx+hw-k, y, '█', sr, sNone)
		}
		// the lattice between the legs, above the arch
		for x := cx - hw + lw; x <= cx+hw-lw; x++ {
			if i < archH {
				a := float64(hw-lw) * math.Sqrt(math.Max(0, 1-float64(i*i)/float64(archH*archH)))
				if math.Abs(float64(x-cx)) < a {
					continue
				}
			}
			if (x+i)%3 != 0 {
				s.put(x, y, '▒', sMetalDark, sNone)
			}
		}
	}
	// the platforms
	for k, pi := range plats {
		if pi >= body {
			continue
		}
		y := base - pi
		hw := half(pi) + 2
		if k == 2 {
			hw = half(pi) + 1
		}
		s.hline(cx-hw, cx+hw, y, '█', sMetalDark, sNone)
		s.hline(cx-hw, cx+hw, y-1, '▀', sMetal, sNone)
		if k < 2 {
			for dx := -hw; dx <= hw; dx += 3 {
				s.put(cx+dx, y-1, '•', sGlow, sNone)
			}
		}
	}
	// the spire and its searchlight
	for y := base - body; y > base-body-spire; y-- {
		s.put(cx, y, '█', sMetal, sNone)
	}
	s.put(cx, 0, '┼', sMetal, sNone)
	s.beacons = append(s.beacons, pt{cx, 0})
	s.lamp(cx, 1, beamDown, atLeast(u(16), 8))
	// strings of lights up the legs
	for i := 2; i < body; i += 3 {
		hw := half(i)
		if hw > 1 {
			s.put(cx-hw, base-i, '•', sGlow, sNone)
			s.put(cx+hw, base-i, '•', sGlow, sNone)
		}
	}
	// close: rivets down the legs
	for i := 1; i < body*60/100; i++ {
		hw := half(i)
		if hw > 2 {
			s.detail(cx-hw+1, base-i, '▘', sMetal)
			s.detail(cx+hw-1, base-i, '▗', sMetal)
		}
	}
	return s
}

// dam: the Hoover Dam seen from downstream: a wide banded concrete face
// curving between tall rock cliffs, the lake glinting over the crest, four
// intake towers and the powerhouse at the foot, the spillways white; a
// pylon on each cliff with a lamp lit at night and a wire across the gap.
func dam(h int) *sprite {
	w := atLeast(h*13/10, 31) | 1
	s := newSprite(w, h)
	base := h - 1
	cx := w / 2
	u := unit(h, 32)
	crest := base - atLeast(u(16), 4)
	cliff := atLeast(w/9, 4) // the cliffs' thickness at the top
	crestHalf := cx - cliff
	botHalf := atLeast(crestHalf*38/100, 3)
	rows := base - crest
	faceHalf := func(y int) int {
		d := y - crest
		// the face narrows toward the foot, curving in
		t := float64(d) / float64(atLeast(rows, 1))
		return crestHalf - int(float64(crestHalf-botHalf)*(1-math.Pow(1-t, 1.7))+0.5)
	}
	// the cliffs: everything outside the gap
	for y := 0; y <= base; y++ {
		half := crestHalf
		if y > crest {
			half = faceHalf(y)
		}
		for x := 0; x < w; x++ {
			dx := x - cx
			if dx >= -half && dx <= half {
				continue
			}
			if y < int(hash(x, 3, 4)%2)+cliffTop(u, x, cx, w) {
				continue
			}
			sl := sRock
			if dx > 0 {
				sl = sRockDark
			}
			if (y+x/3)%4 == 0 {
				sl = sRockDark // strata
			}
			if hash(x, y, 3)%5 == 0 {
				s.put(x, y, '▓', sl, sRockDark)
			} else {
				s.put(x, y, '█', sl, sNone)
			}
		}
	}
	// the dam face: concrete in bands, lit on the left, shaded on the right
	for y := crest; y <= base; y++ {
		d := y - crest
		half := faceHalf(y)
		for x := cx - half; x <= cx+half; x++ {
			dx := x - cx
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
		if d > 1 && d < rows-1 { // the white water of the spillways
			s.put(cx-half+1, y, '║', sWater, sWallLit)
			s.put(cx+half-1, y, '║', sWater, sWallShade)
		}
	}
	for x := cx - crestHalf; x <= cx+crestHalf; x++ {
		s.put(x, crest-1, '▄', sWater, sNone) // the lake over the crest
	}
	// the four intake towers
	for _, tx := range []int{cx - crestHalf*3/5, cx - crestHalf/5, cx + crestHalf/5, cx + crestHalf*3/5} {
		th := atLeast(u(6), 3)
		for y := crest - th; y < crest-1; y++ {
			s.put(tx, y, '█', sTrim, sNone)
			s.put(tx+1, y, '█', sWallShade, sNone)
			if (crest-y)%2 == 0 {
				s.put(tx, y, '▌', sWin, sTrim)
			}
		}
		s.put(tx, crest-th-1, '▄', sTrim, sNone)
		s.put(tx+1, crest-th-1, '▄', sWallShade, sNone)
	}
	// the powerhouse at the foot
	s.block(cx-botHalf-2, cx+botHalf+2, base-1, base, sWall)
	s.hline(cx-botHalf-3, cx+botHalf+3, base-2, '▄', sRoof, sNone)
	for x := cx - botHalf - 1; x < cx+botHalf+2; x += 2 {
		s.put(x, base-1, '▀', sWin, sWall)
	}
	s.put(cx, base, '▄', sGlow, sWallDark)
	// a pylon on each cliff, a wire between them
	px0, px1 := cliff/2, w-1-cliff/2
	top := atLeast(u(4), 2)
	for _, px := range []int{px0, px1} {
		for y := top + 1; y < top+atLeast(u(7), 3); y++ {
			s.put(px-1, y, '╱', sMetal, sNone)
			s.put(px+1, y, '╲', sMetalDark, sNone)
			if y%2 == 0 {
				s.put(px, y, '╳', sMetalDark, sNone)
			}
		}
		s.hline(px-2, px+2, top+1, '─', sMetal, sNone)
		s.put(px, top, '┼', sMetal, sNone)
		s.beacons = append(s.beacons, pt{px, max(0, top-1)})
	}
	for x := px0 + 1; x < px1; x++ {
		sag := int(1.5 * math.Sin(math.Pi*float64(x-px0)/float64(px1-px0)))
		s.put(x, top+1+sag, '─', sMetalDark, sNone)
	}
	// close: the form lines in the concrete, spray at the foot
	for y := crest + 1; y <= base-2; y++ {
		d := y - crest
		half := faceHalf(y)
		for x := cx - half + 2; x < cx+half-1; x += 3 {
			if d%3 != 0 {
				s.detail(x, y, '▌', sWallDark)
			}
		}
	}
	return s
}

// cliffTop is the row a cliff column starts at: the cliffs are tallest at the
// edges and a little lower near the gap.
func cliffTop(u func(float64) int, x, cx, w int) int {
	d := float64(abs(x-cx)) / float64(atLeast(w/2, 1))
	return int((1 - d) * float64(u(6)))
}

// accelerator: the collider, a segmented ring standing on its legs over the
// lab hall with the beam pipe running through its middle out both sides, the
// detector at its heart; the beam circles it in lights.
func accelerator(h int) *sprite {
	w := atLeast(h*11/10, 27) | 1
	s := newSprite(w, h)
	base := h - 1
	cx := w / 2
	u := unit(h, 30)
	lab := atLeast(u(5), 3)
	legs := atLeast(u(3), 1)
	ringH := h - lab - legs
	rx := float64(w)/2 - 0.5
	ry := float64(ringH) // half rows: the ring is ringH rows tall
	cpy := ry
	// the beam pipe through the ring's middle
	py := int(cpy)
	pipeY := py / 2
	for x := 0; x < w; x++ {
		sl := sMetal
		if x%5 < 2 {
			sl = sMetalDark
		}
		s.put(x, pipeY, '▄', sl, sNone)
	}
	// the ring: a rim of segments (every fourth one a bright magnet), a thinner
	// ring inside it carrying the beam
	ellipse(float64(w)/2, cpy, rx, ry, func(x, ppy int, dx, dy float64) {
		r2 := dx*dx + dy*dy
		ang := math.Atan2(dy*ry, dx*rx)
		seg := int((ang + math.Pi) / (2 * math.Pi) * 28)
		switch {
		case r2 > 0.80:
			sl := sMetal
			switch {
			case seg%4 == 0:
				sl = sNeon2
			case dx > 0.2:
				sl = sMetalDark
			}
			s.pset(x, ppy, sl)
		case r2 > 0.66 && r2 <= 0.72:
			if seg%2 == 0 {
				s.pset(x, ppy, sGlow)
			}
		case r2 <= 0.50:
			// the detector: concentric layers
			switch {
			case r2 <= 0.07:
				s.pset(x, ppy, sGlow)
			case r2 <= 0.16:
				s.pset(x, ppy, sNeon2)
			case r2 <= 0.30:
				s.pset(x, ppy, sGlass)
			case r2 <= 0.38:
				s.pset(x, ppy, sMetalDark)
			}
		}
	})
	// the legs and the lab hall
	for _, lx := range []int{cx - w/4, cx + w/4} {
		s.shaft(lx-1, ringH, base-lab, 2, sMetal, sMetalDark, sMetalDark)
	}
	s.block(cx-w/4-2, cx+w/4+2, base-lab+1, base, sWall)
	s.hline(cx-w/4-3, cx+w/4+3, base-lab, '▄', sMetalDark, sNone)
	for x := cx - w/4; x < cx+w/4; x += 2 {
		for y := base - lab + 2; y < base; y++ {
			s.put(x, y, '▌', sWin, sWall)
		}
	}
	s.block(cx-1, cx+1, base-1, base, sWallDark)
	s.put(cx, base, '▄', sGlow, sWallDark)
	s.beacons = append(s.beacons, pt{2, max(0, pipeY-1)}, pt{w - 3, max(0, pipeY-1)})
	// close: bolts on the magnets, the hall's windows split
	for x := 1; x < w-1; x++ {
		for y := 0; y < ringH; y++ {
			if c := s.at(x, y); c.ch == '█' && (c.fg == sMetal || c.fg == sMetalDark) && hash(x, y, 6)%4 == 0 {
				s.detail(x, y, '▘', sMetalDark)
			}
		}
	}
	return s
}
