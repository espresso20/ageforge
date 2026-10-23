package skyline

// wonders_early.go: the landmarks of the first ages, drawn rows tall (the
// sprite is exactly that high, its foot on the ground row).

// grove: the Sacred Grove, one great tree whose roots hold a lit shrine, a
// young conifer on one side and a lesser tree on the other, lanterns hung in
// the boughs and a ring of standing stones, two of them carved with runes.
func grove(h int) *sprite {
	w := atLeast(h*2+1, 15) | 1
	s := newSprite(w, h)
	base := h - 1
	cx := w / 2
	canopy := atLeast(h*68/100, 4) // rows of canopy
	rx, ry := float64(w)*0.31, float64(canopy)
	// a crown of leaves: ragged at the edge, darker low and to the right
	crown := func(ccx, ccy, crx, cry float64) {
		ellipse(ccx, ccy, crx, cry, func(x, py int, dx, dy float64) {
			if dx*dx+dy*dy > 0.74 && hash(x, py, 5)%3 == 0 {
				return
			}
			sl := sLeaf
			if hash(x, py, 6)%4 == 0 || (dy > 0 && dx > -0.15 && hash(x, py, 7)%2 == 0) {
				sl = sLeafDark
			}
			s.pset(x, py, sl)
		})
	}
	// the young conifer, left: a stack of widening boughs on a short trunk
	conH := atLeast(h*45/100, 3)
	for i := 0; i < conH-1; i++ {
		half := 1 + i*3/atLeast(conH-1, 1)
		for x := 2 - half; x <= 2+half; x++ {
			s.put(x, base-conH+1+i, '█', []slot{sLeaf, sLeafDark}[(x+i)%2], sNone)
		}
	}
	s.put(2, base-1, '█', sTrunk, sNone)
	s.put(2, base, '█', sTrunk, sNone)
	// the lesser tree, right: a round crown on a thin trunk
	lh := atLeast(h*40/100, 3)
	lx := w - 1 - atLeast(w/10, 3)
	crown(float64(lx)+0.5, float64(2*(base-lh)+lh), float64(lh)*0.8+0.5, float64(lh)+0.5)
	for y := base - lh + lh/2; y <= base; y++ {
		s.put(lx, y, '█', sTrunk, sNone)
	}
	// the great trunk, flared at the roots, and the canopy over it
	tw := 3
	if h < 9 {
		tw = 2
	}
	for y := canopy - 1; y <= base; y++ {
		flare := 0
		if y >= base-1 {
			flare = 1
		}
		for x := cx - tw/2 - flare; x <= cx+tw/2+flare; x++ {
			sl := sTrunk
			if x > cx {
				sl = sWallShade
			}
			s.put(x, y, '█', sl, sNone)
		}
	}
	crown(float64(cx)+0.5, ry, rx, ry)
	// the shrine in the roots: a low arch, lit within
	s.put(cx, base, '▄', sGlow, sWallDark)
	s.put(cx, base-1, '▀', sWallDark, sTrunk)
	// lanterns hanging under the canopy
	for _, d := range []float64{-0.58, 0.55} {
		x := cx + int(d*rx)
		y := min(canopy+canopy*2/5, base-2)
		s.put(x, y, '•', sGlow, sNone)
		s.put(x, y-1, '│', sTrunk, sNone)
	}
	// a ring of standing stones in front, some carved
	for i, dx := range []int{-9, -6, 6, 9} {
		x := cx + dx*w/31
		if x < 5 || x > w-4 {
			continue
		}
		sh := 1 + i%2
		for y := base - sh + 1; y <= base; y++ {
			s.put(x, y, '█', sRock, sNone)
			s.put(x+1, y, '█', sRockDark, sNone)
		}
		if sh == 2 {
			s.put(x, base-1, '▀', sGlow, sRock) // a rune on the stone
		}
	}
	s.text(0, base, "▄▄", sLeafDark, sNone)
	// close: leaf and bark detail
	dark := func(sl slot) slot {
		if sl == sLeaf {
			return sLeafDark
		}
		return sLeaf
	}
	for y := 0; y < base; y++ {
		for x := 0; x < w; x++ {
			if c := s.at(x, y); c.ch == '█' && (c.fg == sLeaf || c.fg == sLeafDark) && hash(x, y, 9)%3 == 0 {
				s.detail(x, y, []rune("▘▗▖▝")[hash(x, y, 4)%4], dark(c.fg))
			}
		}
	}
	for y := canopy; y < base; y++ {
		s.detail(cx-tw/2, y, '▌', sWallShade)
	}
	return s
}

// monolith: the Great Monolith, a black slab on a stepped plinth, its runes
// lit, a pillar of light over it at night.
func monolith(h int) *sprite {
	w := atLeast(h/2, 7) | 1
	s := newSprite(w, h)
	base := h - 1
	cx := w / 2
	sw := 5
	if h < 12 {
		sw = 3
	}
	top := 3 // the pyramidion
	// the plinth: three steps
	s.block(cx-w/2, cx+w/2, base, base, sRockDark)
	s.block(cx-w/2+1, cx+w/2-1, base-1, base-1, sRock)
	s.hline(cx-sw/2-1, cx+sw/2+1, base-2, '▄', sRock, sNone)
	// the slab
	for y := top; y <= base-2; y++ {
		for x := cx - sw/2; x <= cx+sw/2; x++ {
			sl := sInk
			if x == cx-sw/2 {
				sl = sWallDark
			}
			if x == cx+sw/2 {
				sl = sRockDark
			}
			s.put(x, y, '█', sl, sNone)
		}
	}
	// the cap
	s.hline(cx-sw/2+1, cx+sw/2-1, top-1, '▄', sInk, sNone)
	s.put(cx, top-2, '▄', sWallDark, sNone)
	if top >= 3 {
		s.put(cx, top-3+1, '▲', sGlow, sNone)
	}
	// the runes: a lit groove down the middle, a mark beside it
	for y := top + 1; y <= base-3; y++ {
		switch {
		case y%3 == 1:
			s.put(cx, y, '≡', sGlow, sInk)
		case y%3 == 0 && sw >= 5:
			s.put(cx-1, y, '▪', sNeon2, sInk)
			s.put(cx+1, y, '▪', sNeon2, sInk)
		default:
			s.put(cx, y, '│', sGlow, sInk)
		}
	}
	s.lamp(cx, 0, beamPillar, min(10, h/2))
	// close: the runes carved smaller, down both faces
	for y := top + 1; y <= base-3; y += 2 {
		for _, dx := range []int{-2, 2} {
			if sw >= 5 {
				s.detail(cx+dx, y, []rune("▘▗▝▖")[hash(y, dx)%4], sNeon2)
			}
		}
	}
	return s
}

// henge: Stonehenge, a horseshoe of trilithons, the great one in the middle
// framing a low sun, a lesser pair either side, outliers beyond them and a
// back row of smaller stones showing between.
func henge(h int) *sprite {
	w := atLeast(h*7/4+1, 21) | 1
	s := newSprite(w, h)
	base := h - 1
	cx := w / 2
	k := float64(w) / 39
	at := func(n float64) int { return int(n*k + 0.5) }
	stone := func(x, y0, y1, pw int, back bool) {
		for y := y0; y <= y1; y++ {
			for i := 0; i < pw; i++ {
				sl := sRock
				switch {
				case back:
					sl = sRockDark
				case i == 0 && pw > 1:
					sl = sTrim
				case i == pw-1 && pw > 1:
					sl = sRockDark
				}
				s.put(x+i, y, '█', sl, sNone)
			}
		}
	}
	// the back row, between the great stones
	bh := atLeast(h*48/100, 3)
	for _, dx := range []int{-9, -5, 4, 8} {
		stone(cx+at(float64(dx)), base-bh+1, base, atLeast(at(2.5), 2), true)
	}
	// the low sun in the great gap
	gap := atLeast(at(5.5), 3)
	sr := float64(gap)/2 + 0.2
	ellipse(float64(cx)+0.5, float64(2*base+2)-sr*2, sr, sr, func(x, py int, dx, dy float64) {
		s.pset(x, py, sGlow)
	})
	tri := func(cxp, pw, gp, ph int) { // a trilithon: pillars pw wide, gp apart, ph rows to the lintel
		xl := cxp - gp/2 - pw
		xr := cxp + (gp+1)/2
		stone(xl, base-ph+1, base, pw, false)
		stone(xr, base-ph+1, base, pw, false)
		ly := base - ph
		for x := xl - 1; x <= xr+pw; x++ {
			s.put(x, ly, '█', sRock, sNone)
			s.put(x, ly-1, '▄', sRock, sNone)
		}
		s.put(xl-1, ly, '█', sTrim, sNone)
	}
	big := atLeast(h*76/100, 5)
	mid := atLeast(h*58/100, 4)
	tri(cx, atLeast(at(6.5), 3), gap, big-2)
	tri(cx-at(16), atLeast(at(4), 2), atLeast(at(3), 2), mid-2)
	tri(cx+at(16), atLeast(at(4), 2), atLeast(at(3), 2), mid-2)
	// outliers
	oh := atLeast(h*34/100, 3)
	stone(cx-at(22), base-oh+1, base, atLeast(at(3.5), 2), false)
	stone(cx+at(20), base-oh, base, atLeast(at(3.5), 2), false)
	s.put(cx+at(20), base-oh, '▄', sRock, sNone)
	s.hline(0, w-1, base, '▄', sRockDark, sNone)
	// close: weathering and lichen on the faces
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if c := s.at(x, y); c.ch == '█' && c.fg == sRock && hash(x, y, 2)%5 == 0 {
				s.detail(x, y, []rune("▘▗▖▝")[hash(x, y, 8)%4], sRockDark)
			}
		}
	}
	return s
}

// colosseum: three tiers of arches and an attic, the far end fallen: the top
// tier gone from the middle on, the middle tier ragged beyond that, rubble
// at the foot. The arches are lit from inside at night.
func colosseum(h int) *sprite {
	w := atLeast(h*5/2, 22)
	w = w/3*3 + 2
	s := newSprite(w, h)
	base := h - 1
	attic := atLeast(h/6, 2)
	th := atLeast((h-attic-1-atLeast(h/6, 2))/3, 2)
	type tier struct{ y0, y1 int }
	var tiers [3]tier
	for t := 0; t < 3; t++ {
		y1 := base - 1 - t*th
		tiers[t] = tier{y1 - th + 1, y1}
	}
	// where the wall has fallen: columns from fall on lose their upper tiers
	fall := w * 62 / 100
	standing := func(x, y int) bool {
		switch {
		case x < fall:
			return true
		case x < fall+w/9:
			return y > tiers[2].y0-1 // the top tier gone, the attic with it
		case x < fall+w*2/9:
			return y >= tiers[1].y0+th/2
		case x < fall+w*3/9:
			return y > tiers[0].y0+th/3
		}
		return y > tiers[0].y1-1
	}
	wallAt := func(x int) slot {
		switch {
		case x <= 1:
			return sWallLit
		case x > w*3/4:
			return sWallShade
		}
		return sWall
	}
	// the stylobate
	for x := 0; x < w; x++ {
		s.put(x, base, '▄', sTrim, sNone)
	}
	for t := 0; t < 3; t++ {
		for y := tiers[t].y0; y <= tiers[t].y1; y++ {
			for x := 0; x < w; x++ {
				if !standing(x, y) {
					continue
				}
				bay := x % 3
				switch {
				case y == tiers[t].y0:
					s.put(x, y, '█', sTrim, sNone) // the cornice
				case bay == 0:
					s.put(x, y, '█', wallAt(x), sNone) // the pier
				case y == tiers[t].y0+1:
					s.put(x, y, '▄', sWallDark, wallAt(x)) // the crown of the arch
				case y == tiers[t].y1:
					s.put(x, y, '▄', sWin, sWallDark) // lit within
				default:
					s.put(x, y, '█', sWallDark, sNone)
				}
			}
		}
	}
	// the attic
	for y := tiers[2].y0 - attic; y < tiers[2].y0; y++ {
		for x := 0; x < w; x++ {
			if !standing(x, y) || x >= fall+w/9 {
				continue
			}
			sl := wallAt(x)
			if y == tiers[2].y0-attic {
				s.put(x, y, '█', sTrim, sNone)
				continue
			}
			if x%3 == 1 && y == tiers[2].y0-1 {
				s.put(x, y, '■', sWin, sl)
				continue
			}
			s.put(x, y, '█', sl, sNone)
		}
	}
	// the velarium's masts along the rim, a few rows of them over the standing part
	mh := atLeast(h/6, 2)
	mastTop := tiers[2].y0 - attic - mh
	for x := 2; x < fall+w/9-2; x += 3 {
		for y := max(0, mastTop); y < tiers[2].y0-attic; y++ {
			s.put(x, y, '│', sTrunk, sNone)
		}
	}
	// the fallen top edge, ragged
	for x := fall + w/9 - 1; x < fall+w/9+3 && x < w; x++ {
		if hash(x, 1, 6)%2 == 0 {
			s.put(x, tiers[2].y0, '▀', sTrim, sNone)
		}
	}
	// rubble at the foot
	for x := fall + 1; x < w; x++ {
		if hash(x, 2, 6)%3 != 0 {
			s.put(x, base-1, '▄', sRock, sNone)
		}
	}
	// close: carved frieze on each cornice, small windows in the attic
	for t := 0; t < 3; t++ {
		for x := 1; x < w-1; x += 2 {
			if standing(x, tiers[t].y0) {
				s.detail(x, tiers[t].y0, '▖', sWallDark)
			}
		}
	}
	for x := 1; x < fall+w/9-1; x += 3 {
		s.detail(x, tiers[2].y0-1, '▘', sWin)
	}
	return s
}

// parthenon: a temple on its rock: the hill, steps, a front of eight fluted
// columns with the dark cella behind and its lit door, a frieze of triglyphs
// and a pediment, braziers on the steps.
func parthenon(h int) *sprite {
	w := atLeast(h*2, 25) | 1
	s := newSprite(w, h)
	base := h - 1
	cx := w / 2
	hill := atLeast(h*30/100, 2)
	steps := 2
	ped := atLeast(h*15/100, 2)
	ent := 2
	colH := h - hill - steps - ped - ent
	if colH < 3 {
		colH = 3
		steps = 1
		ent = 1
	}
	// the hill: a rough slope up to a flat top
	topY := base - hill + 1
	for x := 0; x < w; x++ {
		d := cx - x
		if d < 0 {
			d = -d
		}
		ht := hill
		if edge := w/2 - d; edge < hill*2 {
			ht = atLeast(edge/2, 1)
		}
		ht = atLeast(ht+int(hash(x, 4, 7)%2), 1)
		if ht > hill {
			ht = hill
		}
		for y := base - ht + 1; y <= base; y++ {
			sl := sRock
			if x > cx || (y > base-ht+1 && hash(x, y, 8)%3 == 0) {
				sl = sRockDark
			}
			s.put(x, y, '█', sl, sNone)
		}
	}
	for x := 1; x < w-1; x++ {
		if hash(x, 1, 9)%7 == 0 {
			s.put(x, topY-1, '▄', sLeafDark, sNone)
		}
	}
	// the steps
	for i := 0; i < steps; i++ {
		y := topY - 1 - i
		inset := steps - 1 - i
		for x := 2 + inset; x < w-2-inset; x++ {
			sl := sTrim
			if i > 0 {
				sl = sWall
			}
			s.put(x, y, '█', sl, sNone)
		}
	}
	stylo := topY - steps
	// eight columns three wide with one between
	cw, step, n := 3, 4, 8
	if w < 33 {
		cw, step = 2, 3
	}
	span := (n-1)*step + cw
	x0 := cx - span/2
	x1 := x0 + span - 1
	colTop := stylo - colH + 1
	for y := colTop; y <= stylo; y++ {
		for x := x0; x <= x1; x++ {
			s.put(x, y, '█', sWallDark, sNone) // the cella wall behind
		}
	}
	dh := atLeast(colH*2/3, 2)
	for y := stylo - dh + 1; y <= stylo; y++ {
		s.put(cx, y, '█', sGlow, sNone)
		s.put(cx+1, y, '█', sWin, sNone)
	}
	for i := 0; i < n; i++ {
		s.shaft(x0+i*step, colTop, stylo, cw, sTrim, sWall, sWallShade)
	}
	// the entablature: architrave and a frieze of triglyphs
	ay := colTop - 1
	for x := x0 - 1; x <= x1+1; x++ {
		s.put(x, ay, '█', sTrim, sNone)
	}
	if ent == 2 {
		for x := x0 - 1; x <= x1+1; x++ {
			sl := sWall
			if x%3 == 0 {
				sl = sWallShade
			}
			s.put(x, ay-1, '█', sl, sNone)
		}
		ay--
	}
	// the pediment
	pw := x1 - x0 + 3
	for r := 0; r < ped*2; r++ {
		py := 2*ay - 1 - r
		half := (pw/2 - 1) - r*(pw/2-1)/(ped*2)
		for x := cx - half; x <= cx+half; x++ {
			sl := sWallShade
			if x <= cx-half+1 || x >= cx+half-1 || r == 0 {
				sl = sTrim
			}
			s.pset(x, py, sl)
		}
	}
	// braziers on the steps
	s.flame = append(s.flame, pt{x0 - 2, topY - 2}, pt{x1 + 2, topY - 2})
	// close: fluting on every column, carved metopes
	for i := 0; i < n; i++ {
		for y := colTop; y <= stylo; y++ {
			s.detail(x0+i*step, y, '▌', sWallShade)
			if cw == 3 {
				s.detail(x0+i*step+1, y, '▐', sWallShade)
			}
		}
	}
	if ent == 2 {
		for x := x0 + 1; x <= x1; x += 3 {
			s.detail(x, ay, '▝', sWallDark)
		}
	}
	return s
}

// library: the Great Library, a domed reading hall between two wings
// papered with scroll niches, a beacon tower at each end, the portico's
// lamps and the dome's lantern lit at night.
func library(h int) *sprite {
	w := atLeast(h*5/4, 25) | 1
	s := newSprite(w, h)
	base := h - 1
	cx := w / 2
	u := unit(h, 27)
	hall := atLeast(u(10), 4) // the hall's walls
	drum := atLeast(u(5), 2)
	lan := atLeast(u(3), 2)
	wing := atLeast(u(7), 3)
	tower := atLeast(u(15), 6)
	hx0, hx1 := cx-u(8), cx+u(8)
	// the wings: scroll niches
	for _, side := range []int{-1, 1} {
		xa, xb := cx+side*u(5), cx+side*(w/2-1)
		if xa > xb {
			xa, xb = xb, xa
		}
		for y := base - wing; y <= base; y++ {
			for x := max(0, xa); x <= min(w-1, xb); x++ {
				sl := sWall
				if x == xa && side < 0 || x == xb && side > 0 {
					sl = sWallLit
				}
				s.put(x, y, '█', sl, sNone)
				if y > base-wing && y < base && (x+y)%2 == 1 && x > xa && x < xb {
					s.put(x, y, '▪', sWallDark, sWall)
				}
			}
		}
		s.hline(max(0, xa), min(w-1, xb), base-wing, '▀', sTrim, sWall)
	}
	// the beacon towers at the ends
	for _, side := range []int{-1, 1} {
		tx := 0
		if side > 0 {
			tx = w - 5
		}
		for y := base - tower; y <= base; y++ {
			for x := tx; x < tx+5; x++ {
				sl := sWall
				if x == tx {
					sl = sWallLit
				}
				if x == tx+4 {
					sl = sWallShade
				}
				s.put(x, y, '█', sl, sNone)
			}
			if (base-y)%3 == 1 && y > base-tower+2 {
				s.put(tx+2, y, '▌', sWin, sWall)
			}
		}
		s.crenels(tx, 5, base-tower-1, sTrim)
		s.put(tx+2, base-tower-2, '▲', sGlow, sNone)
		s.beacons = append(s.beacons, pt{tx + 2, base - tower - 3})
	}
	// the hall: portico of columns under a cornice
	hy := base - hall + 1
	s.block(hx0, hx1, hy, base, sWall)
	for x := hx0 + 1; x < hx1; x += 3 {
		for y := hy + 1; y < base; y++ {
			s.put(x+1, y, '█', sWallDark, sNone) // the shadow between columns
		}
	}
	for x := hx0; x <= hx1-1; x += 3 {
		s.shaft(x, hy+1, base-1, 2, sTrim, sWall, sWallShade)
	}
	s.hline(hx0-1, hx1+1, hy, '█', sTrim, sNone)
	s.hline(hx0-1, hx1+1, base, '▄', sTrim, sNone)
	for x := hx0 + 2; x < hx1; x += 3 {
		s.put(x, base-1, '▄', sGlow, sWallDark) // lamps between the columns
	}
	// the drum with arched windows, the dome and its lantern
	dx0, dx1 := cx-u(6), cx+u(6)
	dy := hy - drum
	s.block(dx0, dx1, dy, hy-1, sWall)
	for x := dx0 + 1; x < dx1; x += 2 {
		for y := dy + 1; y < hy-1; y++ {
			s.put(x, y, '▌', sWin, sWall)
		}
	}
	s.hline(dx0-1, dx1+1, dy, '▄', sTrim, sNone)
	dh := s.dome(dx0, dx1-dx0+1, dy-1, sMetal)
	ly := dy - 1 - dh
	for y := ly; y > ly-lan; y-- {
		s.put(cx, y, '█', sGlow, sNone)
		s.put(cx-1, y, '▐', sTrim, sNone)
		s.put(cx+1, y, '▌', sTrim, sNone)
	}
	s.put(cx, ly-lan, '▲', sTrim, sNone)
	// close: the scrolls in the niches, the columns' flutes, the drum's glazing bars
	for _, side := range []int{-1, 1} {
		xa, xb := cx+side*u(5), cx+side*(w/2-1)
		if xa > xb {
			xa, xb = xb, xa
		}
		for y := base - wing + 1; y < base; y++ {
			for x := max(0, xa) + 1; x < min(w-1, xb); x++ {
				if (x+y)%2 == 1 {
					s.detail(x, y, '▘', sTrim)
				}
			}
		}
	}
	for x := hx0; x <= hx1; x += 3 {
		for y := hy + 1; y < base; y++ {
			s.detail(x, y, '▌', sWallShade)
		}
	}
	return s
}
