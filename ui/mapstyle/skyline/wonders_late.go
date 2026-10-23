package skyline

import "math"

// wonders_late.go: the landmarks of the modern and the future ages.

// rocket: the Space Program, a staged rocket on its pad beside the service
// gantry with its swing arms out to the hull, the lightning mast at the other
// side; the engines glow and the pad steams at night.
func rocket(h int) *sprite {
	w := atLeast(h*7/10, 17) | 1
	s := newSprite(w, h)
	base := h - 1
	cx := w / 2
	u := unit(h, 34)
	pad := 2
	s1 := atLeast(u(14), 4)
	s2 := atLeast(u(9), 3)
	cap := atLeast(u(4), 2)
	tip := atLeast(u(4), 2)
	rw := 3 // half-width of the first stage
	// the pad
	s.block(cx-7, cx+7, base, base, sWall)
	s.block(cx-5, cx+5, base-1, base-1, sMetalDark)
	// the first stage, banded; fins at its foot
	y0 := base - pad
	for i := 0; i < s1; i++ {
		y := y0 - i
		for dx := -rw; dx <= rw; dx++ {
			sl := sWallLit
			switch {
			case dx < -rw+1:
				sl = sWallLit
			case dx > rw-2:
				sl = sWallShade
			default:
				sl = sWall
			}
			if (i/4)%2 == 1 {
				sl = sInk
				if dx == -rw {
					sl = sWallDark
				}
			}
			s.put(cx+dx, y, '█', sl, sNone)
		}
		if i%4 == 1 && i > 0 && (i/4)%2 == 0 {
			s.put(cx, y, '▪', sInk, sWall)
		}
	}
	for i := 0; i < atLeast(u(4), 2); i++ {
		s.put(cx-rw-1-(atLeast(u(4), 2)-1-i)/2, y0-i, '◢', sMetal, sNone)
		s.put(cx+rw+1+(atLeast(u(4), 2)-1-i)/2, y0-i, '◣', sMetalDark, sNone)
	}
	// the interstage ring, the second stage, the capsule and its escape tower
	y := y0 - s1
	s.hline(cx-rw, cx+rw, y, '▄', sMetalDark, sNone)
	for i := 0; i < s2; i++ {
		for dx := -2; dx <= 2; dx++ {
			sl := sWall
			switch {
			case dx == -2:
				sl = sWallLit
			case dx == 2:
				sl = sWallShade
			}
			if (i/3)%2 == 0 && i > 0 {
				sl = sInk
			}
			s.put(cx+dx, y-1-i, '█', sl, sNone)
		}
	}
	y -= s2 + 1
	for i := 0; i < cap; i++ {
		half := 2 - i*2/atLeast(cap, 1)
		for dx := -half; dx <= half; dx++ {
			sl := sWall
			if dx > 0 {
				sl = sWallShade
			}
			s.put(cx+dx, y-i, '█', sl, sNone)
		}
	}
	s.put(cx, y-cap, '▲', sWallLit, sNone)
	for i := 1; i <= tip; i++ {
		s.put(cx, y-cap-i, '│', sMetal, sNone)
	}
	s.beacons = append(s.beacons, pt{cx, max(0, y-cap-tip-1)})
	// the engines: a glow under the stage, steam round the pad
	s.put(cx-1, y0+1, '▀', sGlow, sNone)
	s.put(cx, y0+1, '▀', sGlow, sNone)
	s.put(cx+1, y0+1, '▀', sGlow, sNone)
	for _, dx := range []int{-6, -4, 4, 6} {
		s.put(cx+dx, base-1, '▒', sSmoke, sNone)
	}
	// the service gantry, left: a lattice with its arms out to the rocket
	gx := cx - rw - 5
	gh := atLeast(u(27), 6)
	for i := 0; i < gh; i++ {
		yy := base - 1 - i
		s.put(gx, yy, '█', sMetal, sNone)
		s.put(gx+3, yy, '█', sMetalDark, sNone)
		if i%2 == 0 {
			s.put(gx+1, yy, '╲', sMetalDark, sNone)
			s.put(gx+2, yy, '╱', sMetalDark, sNone)
		}
	}
	for _, ai := range []int{u(10), u(17), u(23)} {
		if ai < gh {
			ay := base - 1 - ai
			for x := gx + 4; x < cx-rw; x++ {
				s.put(x, ay, '═', sMetal, sNone)
			}
		}
	}
	s.put(gx+1, base-1-gh, '┬', sMetal, sNone)
	s.beacons = append(s.beacons, pt{gx + 1, max(0, base-2-gh)})
	// the lightning mast, right
	mx := cx + rw + 5
	for i := 0; i < atLeast(u(24), 5); i++ {
		s.put(mx, base-1-i, '║', sMetal, sNone)
	}
	s.beacons = append(s.beacons, pt{mx, max(0, base-1-atLeast(u(24), 5))})
	// close: the roll pattern and rivets on the first stage, the gantry's rungs
	for i := 0; i < s1; i++ {
		if (i/4)%2 == 0 {
			s.detail(cx-rw+1, y0-i, '▘', sWallDark)
			s.detail(cx+rw-1, y0-i, '▗', sWallDark)
		}
	}
	for i := 0; i < gh; i++ {
		s.detail(gx+1, base-1-i, '▌', sMetalDark)
	}
	return s
}

// network: the Global Network, a lattice mast carrying dishes, topped with a
// wire-frame globe whose nodes burn at night.
func network(h int) *sprite {
	w := atLeast(h*5/8, 21) | 1
	s := newSprite(w, h)
	base := h - 1
	cx := w / 2
	u := unit(h, 34)
	hub := 2
	gr := atLeast(u(9), 4) // the globe's height in rows (and its radius in columns)
	mast := h - hub - gr - 2
	if mast < 4 {
		mast = 4
	}
	// the hub at the foot
	s.block(cx-5, cx+5, base-hub+1, base, sWall)
	s.hline(cx-6, cx+6, base-hub, '▄', sTrim, sNone)
	for x := cx - 4; x <= cx+4; x += 2 {
		s.put(x, base-1, '▀', sWin, sWall)
	}
	s.put(cx, base, '▄', sGlow, sWallDark)
	// the mast: a narrowing lattice
	mb := base - hub
	for i := 0; i < mast; i++ {
		y := mb - i
		t := float64(i) / float64(mast)
		half := 3 - int(2.4*t)
		if half > 0 {
			s.put(cx-half, y, '█', sMetalDark, sNone)
			s.put(cx+half, y, '█', sWallDark, sNone)
			for x := cx - half + 1; x < cx+half; x++ {
				s.put(x, y, '▒', sWallDark, sNone)
			}
		} else {
			s.put(cx, y, '█', sMetalDark, sNone)
		}
	}
	// the cross-arms with dishes
	for k, ai := range []int{u(9), u(15)} {
		if ai >= mast {
			continue
		}
		y := mb - ai
		half := 4 + k
		s.hline(cx-half, cx+half, y, '═', sMetal, sNone)
		for _, dx := range []int{-half, half - 1} {
			s.put(cx+dx, y-1, '◣', sMetalDark, sNone)
			s.put(cx+dx+1, y-1, '◢', sMetal, sNone)
		}
		s.put(cx-half, y-2, '┬', sMetal, sNone)
		s.put(cx+half, y-2, '┬', sMetal, sNone)
	}
	// the globe: a glass sphere in a frame of meridians, nodes lit
	gcy := float64(2*(mb-mast)) - float64(gr)
	rg := float64(gr)
	ellipse(float64(cx)+0.5, gcy, rg, rg, func(x, py int, dx, dy float64) {
		r2 := dx*dx + dy*dy
		switch {
		case r2 > 0.80:
			s.pset(x, py, sMetal)
		case math.Abs(dx) < 0.08 || math.Abs(dy) < 0.10:
			s.pset(x, py, sNeon1) // the equator and the prime meridian
		case math.Abs(math.Abs(dx)-0.52) < 0.07 || math.Abs(math.Abs(dy)-0.55) < 0.09:
			s.pset(x, py, sNeon1)
		case hash(x, py, 3)%19 == 0:
			s.pset(x, py, sGlow)
		default:
			s.pset(x, py, sGlass)
		}
	})
	gtop := mb - mast - gr
	for _, p := range []pt{{cx - gr + 1, gtop + gr/2}, {cx + gr - 1, gtop + gr/2}, {cx, gtop}} {
		s.put(p.x, p.y, '•', sGlow, sNone)
	}
	s.put(cx, gtop-1, '│', sMetal, sNone)
	s.beacons = append(s.beacons, pt{cx, max(0, gtop-2)})
	// close: the globe's continents, rivets on the mast
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if c := s.at(x, y); c.ch == '█' && c.fg == sGlass && hash(x, y, 8)%3 == 0 {
				s.detail(x, y, '▘', sNeon1)
			}
		}
	}
	return s
}

// simulation: the World Simulation, a black tower of servers with data
// streaming down it, a hovering crystal above it, held on a beam of light.
func simulation(h int) *sprite {
	w := atLeast(h/2, 13) | 1
	s := newSprite(w, h)
	base := h - 1
	cx := w / 2
	u := unit(h, 32)
	cr := atLeast(u(8), 4) // the crystal's rows
	tower := h - cr - atLeast(u(2), 1) - 1
	tw := atLeast(w*5/9, 7) | 1
	hx := tw / 2
	// the plinth of server racks
	s.block(cx-hx-2, cx+hx+2, base-1, base, sWallDark)
	for x := cx - hx - 2; x <= cx+hx+2; x += 2 {
		s.put(x, base-1, '▀', sNeon1, sWallDark)
		s.put(x+1, base, '▄', sNeon2, sWallDark)
	}
	// the tower
	glyphs := []rune("01ƒΣ≡∞¥")
	for y := base - tower - 1; y <= base-2; y++ {
		for x := cx - hx; x <= cx+hx; x++ {
			sl := sInk
			switch {
			case x == cx-hx:
				sl = sWallShade
			case x == cx+hx:
				sl = sWallDark
			}
			s.put(x, y, '█', sl, sNone)
		}
		for x := cx - hx + 1; x < cx+hx; x += 2 {
			if hash(x, y, 1)%3 == 0 {
				s.put(x, y, glyphs[hash(x, y, 2)%uint64(len(glyphs))], sNeon2, sInk)
			}
		}
	}
	s.hline(cx-hx, cx+hx, base-tower-1, '▀', sNeon1, sInk)
	// the beam up to the crystal
	for y := cr + 1; y < base-tower-1; y++ {
		s.put(cx, y, '│', sGlow, sNone)
	}
	// the crystal: a rhombus of triangles round a core
	half := atLeast(cr/2+1, 3)
	for i := 0; i < cr; i++ {
		d := i
		if i >= cr/2 {
			d = cr - 1 - i
		}
		hw := 1 + d*(half-1)/atLeast(cr/2, 1)
		for dx := -hw; dx <= hw; dx++ {
			ch := '█'
			switch {
			case dx == -hw && i < cr/2:
				ch = '◢'
			case dx == hw && i < cr/2:
				ch = '◣'
			case dx == -hw:
				ch = '◥'
			case dx == hw:
				ch = '◤'
			}
			sl := sNeon2
			if dx > 0 {
				sl = sNeon1
			}
			if abs(dx) <= hw/3 && abs(i-cr/2) <= 1 {
				sl = sGlow
			}
			s.put(cx+dx, i, ch, sl, sNone)
		}
	}
	s.lamp(cx, 0, beamPillar, atLeast(u(6), 3))
	// close: more data, finer
	for y := base - tower - 1; y <= base-2; y++ {
		for x := cx - hx + 1; x < cx+hx; x++ {
			if c := s.at(x, y); c.ch == '█' && c.fg == sInk && hash(x, y, 5)%4 == 0 {
				s.detail(x, y, []rune("▘▗▖▝")[hash(x, y, 6)%4], sNeon2)
			}
		}
	}
	return s
}

// citadel: the Neon Citadel, a stepped megatower edged in neon, its sign
// down the face, skybridges to two needles, a crown of aerials; at night a
// pillar of light stands over it and searchlights swing across the roofs.
func citadel(h int) *sprite {
	w := atLeast(h*7/10, 25) | 1
	s := newSprite(w, h)
	base := h - 1
	cx := w / 2
	u := unit(h, 34)
	// the flanking needles and the bridges to them
	for _, nx := range []int{1, w - 4} {
		nh := atLeast(u(20), 5)
		for y := base - nh + 1; y <= base; y++ {
			s.put(nx, y, '█', sWallLit, sNone)
			s.put(nx+1, y, '█', sWall, sNone)
			s.put(nx+2, y, '█', sWallShade, sNone)
			if (base-y)%3 == 1 {
				s.put(nx+1, y, '■', sWin, sWall)
			}
		}
		s.put(nx+1, base-nh, '▲', sNeon2, sNone)
		s.put(nx+1, base-nh-1, '│', sMetal, sNone)
	}
	// the tiers, each set back from the one under it
	type tier struct{ in, top int }
	tiers := []tier{{3, u(12)}, {5, u(21)}, {7, u(28)}, {9, u(32)}}
	for ti, t := range tiers {
		y0 := base - t.top + 1
		for y := y0; y <= base; y++ {
			for x := t.in; x < w-t.in; x++ {
				sl := sWall
				switch {
				case x == t.in:
					sl = sWallLit
				case x > cx+1:
					sl = sWallShade
				}
				s.put(x, y, '█', sl, sNone)
				if (y-y0)%2 == 1 && x > t.in && x < w-t.in-1 && hash(x, y, ti)%4 != 0 {
					s.put(x, y, '▄', sWin, sl)
				}
			}
		}
		s.hline(t.in, w-t.in-1, y0, '▀', sNeon1, sWall)
		s.put(t.in, y0, '▐', sNeon2, sNone)
		s.put(w-t.in-1, y0, '▌', sNeon2, sNone)
		for y := y0 + 1; y <= base; y++ {
			s.put(t.in, y, '▐', sNeon1, sWallLit) // the lit edge
		}
	}
	// the sign, down the face
	top := tiers[len(tiers)-1].top
	for i, c := range "CITADEL" {
		if y := base - top + 3 + i; y < base-1 {
			s.put(cx, y, c, sNeon2, sInk)
		}
	}
	// the bridges
	by := base - atLeast(u(8), 2)
	for x := 4; x < tiers[0].in; x++ {
		s.put(x, by, '═', sMetal, sNone)
	}
	for x := w - tiers[0].in; x < w-4; x++ {
		s.put(x, by, '═', sMetal, sNone)
	}
	// the crown: aerials and the beacon
	ct := base - top
	s.put(cx, ct, '▲', sNeon1, sNone)
	for y := ct - 1; y >= max(0, ct-atLeast(u(3), 2)); y-- {
		s.put(cx, y, '│', sMetal, sNone)
	}
	s.put(cx-2, ct, '┬', sMetal, sNone)
	s.put(cx+2, ct, '┬', sMetal, sNone)
	s.beacons = append(s.beacons, pt{cx, max(0, ct-atLeast(u(3), 2)-1)}, pt{cx - 2, max(0, ct-1)}, pt{cx + 2, max(0, ct-1)})
	s.lamp(cx, max(0, ct-atLeast(u(3), 2)-1), beamPillar, atLeast(u(10), 4))
	s.lamp(cx-2, max(0, ct-1), beamSky, atLeast(u(14), 6))
	s.lamp(cx+2, max(0, ct-1), beamSky, atLeast(u(14), 6))
	// close: the windows in the glass, panelling on the edges
	for y := base - top + 1; y <= base; y++ {
		for x := 4; x < w-4; x++ {
			if c := s.at(x, y); c.ch == '█' && isWall(c.fg) && hash(x, y, 9)%3 == 0 {
				s.detail(x, y, []rune("▘▗▖▝")[hash(x, y, 6)%4], sWin)
			}
		}
	}
	return s
}

// cradle: the Stellar Cradle, a starship standing in its launch arch between
// two gantries, the engines lit, running lights along the arch.
func cradle(h int) *sprite {
	w := atLeast(h*7/10, 21) | 1
	s := newSprite(w, h)
	base := h - 1
	cx := w / 2
	u := unit(h, 34)
	// the two gantries and the arch between them
	gh := h - atLeast(u(3), 2)
	for _, side := range []int{-1, 1} {
		for i := 0; i < gh; i++ {
			y := base - i
			t := float64(i) / float64(gh)
			// each gantry leans in toward the top, as one half of an arch
			x := cx + side*(w/2-1-int(float64(w/2-3)*math.Pow(t, 2.2)))
			s.put(x, y, '█', sMetal, sNone)
			s.put(x-side, y, '▐', sMetalDark, sNone)
			if side > 0 {
				s.put(x, y, '█', sMetalDark, sNone)
				s.put(x-side, y, '▌', sMetal, sNone)
			}
			if i%4 == 2 {
				s.put(x-2*side, y, '═', sNeon1, sNone)
			}
		}
	}
	// the ship: a slender hull with a swept pair of fins and a lit nose
	sh := atLeast(u(28), 7)
	for i := 0; i < sh; i++ {
		y := base - 2 - i
		t := float64(i) / float64(sh)
		half := 1
		if t < 0.7 {
			half = 2
		}
		if t > 0.88 {
			half = 0
		}
		for dx := -half; dx <= half; dx++ {
			sl := sWall
			switch {
			case dx == -half && half > 0:
				sl = sWallLit
			case dx == half && half > 0:
				sl = sWallShade
			}
			if i%7 == 3 {
				sl = sNeon1
			}
			s.put(cx+dx, y, '█', sl, sNone)
		}
	}
	s.put(cx, base-2-sh, '▲', sWallLit, sNone)
	for i := 0; i < atLeast(sh/4, 2); i++ {
		s.put(cx-3-(atLeast(sh/4, 2)-1-i)/2, base-2-i, '◢', sMetal, sNone)
		s.put(cx+3+(atLeast(sh/4, 2)-1-i)/2, base-2-i, '◣', sMetalDark, sNone)
	}
	s.put(cx, base-1, '▀', sGlow, sNone)
	s.put(cx-1, base-1, '▀', sGlow, sNone)
	s.put(cx+1, base-1, '▀', sGlow, sNone)
	s.hline(0, w-1, base, '▀', sMetalDark, sNone)
	s.beacons = append(s.beacons, pt{cx, max(0, base-3-sh)})
	s.lamp(cx, max(0, base-3-sh), beamPillar, atLeast(u(8), 3))
	// close: the hull's panelling, rungs on the gantries
	for i := 0; i < sh; i++ {
		s.detail(cx, base-2-i, '▌', sWallDark)
	}
	return s
}
