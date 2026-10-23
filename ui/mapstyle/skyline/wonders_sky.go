package skyline

import "math"

// wonders_sky.go: the landmarks of the cosmic ages. They also stand in the
// sky arc's scenes (orbit_forms.go), so each is drawn on its own, with no
// ground under it but its foot.

// dyson: the Dyson Scaffold, a captive star half wrapped in lattice bands,
// held up by a mast on a plinth.
func dyson(h int) *sprite {
	w := atLeast(h*2+3, 17)
	s := newSprite(w, h)
	base := h - 1
	cx := float64(w) / 2
	R := float64(h) * 0.40
	cyp := float64(2*(base-2)) - float64(2*R)
	for py := 0; py < 2*base; py++ {
		for x := 0; x < w; x++ {
			dx := (float64(x) + 0.5 - cx) / 2
			dy := (float64(py) + 0.5 - cyp) / 2
			d := math.Sqrt(dx*dx + dy*dy)
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
	// close: rivets and braces on the bands
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if c := s.at(x, y); c.fg == sMetal && c.ch != 0 && hash(x, y, 4)%3 == 0 {
				s.detail(x, y, '▘', sMetalDark)
			}
		}
	}
	return s
}

// ringGate: the Warp Nexus, a ring on two pylons with light inside it.
func ringGate(h int) *sprite {
	w := atLeast(h*2-5, 15) | 1
	s := newSprite(w, h)
	ringH := h - 3
	cx := w / 2
	cy := float64(ringH) / 2
	for y := 0; y < ringH; y++ {
		for x := 0; x < w; x++ {
			dx := float64(x-cx) / float64(w/2)
			dy := (float64(y) - cy) / cy
			d := math.Sqrt(dx*dx + dy*dy)
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
	for _, px := range []int{cx - ringH/2, cx + ringH/2} {
		s.vline(px, ringH-1, h-2, '█', sMetalDark, sNone)
		s.put(px-1, h-1, '▄', sMetalDark, sNone)
		s.put(px, h-1, '█', sMetalDark, sNone)
		s.put(px+1, h-1, '▄', sMetalDark, sNone)
	}
	s.beacons = append(s.beacons, pt{cx, 0})
	for y := 0; y < ringH; y++ {
		for x := 0; x < w; x++ {
			if c := s.at(x, y); c.ch == '█' && (c.fg == sMetal || c.fg == sMetalDark) && hash(x, y, 7)%4 == 0 {
				s.detail(x, y, '▘', sNeon1)
			}
		}
	}
	return s
}

// beacon: the Cosmic Beacon, a ring spire whose crown throws a beam into the sky.
func beacon(h int) *sprite {
	want := h - h/3 - 1
	sp := aRingSpire(newRnd(77), want)
	for want > 4 && sp.h > h-3 {
		want--
		sp = aRingSpire(newRnd(77), want)
	}
	s := newSprite(sp.w, h)
	s.blit(sp, 0, s.h-sp.h)
	cx := sp.w / 2
	for y := 0; y < s.h-sp.h; y++ {
		ch := '│'
		if y%3 == 0 {
			ch = '║'
		}
		s.put(cx, y, ch, sGlow, sNone)
	}
	if s.h > sp.h {
		s.put(cx-1, s.h-sp.h-1, '◢', sNeon2, sNone)
		s.put(cx+1, s.h-sp.h-1, '◣', sNeon2, sNone)
	}
	return s
}

// anchor: the Reality Anchor, an inverted pyramid pinned to the ground by
// pillars, shards of light orbiting it.
func anchor(h int) *sprite {
	w := atLeast(h*4/5, 15) | 1
	s := newSprite(w, h)
	base := h - 1
	top := 0
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
		s.put(w/4, y, '│', sMetalDark, sNone)
		s.put(w-1-w/4, y, '│', sMetalDark, sNone)
		if y%2 == 0 {
			s.put(w/2, y, '·', sNeon2, sNone)
		}
	}
	for i, p := range []pt{{0, top + 1}, {w - 1, top + 2}, {1, top + ph}, {w - 2, top + ph + 1}} {
		s.put(p.x, p.y, '◆', []slot{sNeon1, sNeon2}[i%2], sNone)
	}
	s.text(w/4-1, base, "▄█▄", sMetal, sNone)
	s.text(w-1-w/4-1, base, "▄█▄", sMetal, sNone)
	return s
}

// singularity: the Singularity Core, a black sphere ringed in light with its
// accretion disc.
func singularity(h int) *sprite {
	w := atLeast(h*2+5, 17)
	s := newSprite(w, h)
	cx := w / 2
	cy := float64(h) * 0.45
	R := float64(h) * 0.28
	for y := 0; y < h-1; y++ {
		for x := 0; x < w; x++ {
			dx := float64(x-cx) / 2
			dy := float64(y) - cy
			d := math.Sqrt(dx*dx + dy*dy)
			switch {
			case d <= R:
				s.put(x, y, '█', sInk, sNone)
			case d <= R+1.2:
				s.put(x, y, '▓', sGlow, sInk)
			}
			ex, ey := dx/(R*2.6), math.Abs(dy)*4.5/(R*2.6)
			if ex*ex+ey*ey <= 1 && d > R {
				ch := '▀'
				if dy > 0 {
					ch = '▄'
				}
				s.put(x, y, ch, sNeon2, sNone)
			}
		}
	}
	s.text(cx-2, h-1, "▄███▄", sMetalDark, sNone)
	return s
}
