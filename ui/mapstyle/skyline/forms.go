package skyline

import "math"

// forms.go is the building grammar: parametric archetypes drawn in blocks
// and line art. A building is (archetype, height, seed): the archetype
// comes from its lineage and period (catalog.go), the height from its period
// and the scene scale, the seed from the lot. There are no per-building
// sprites; three hundred buildings map onto these forms.

type archFn func(r *rnd, h int) *sprite

// ---------------------------------------------------------------- early

func aHut(r *rnd, h int) *sprite {
	w := 7 + 2*r.n(2)
	s := newSprite(w, 5)
	s.cone(0, w, 0, 7)
	s.hline(1, w-2, 4, '█', sWall, sNone)
	s.put(w-2, 4, '█', sWallShade, sNone)
	s.put(w/2, 4, '▄', sGlow, sWallDark)
	s.put(w/2, 3, '▄', sWallDark, s.at(w/2, 3).fg)
	s.clear(0, 4)
	s.clear(w-1, 4)
	if r.n(2) == 0 {
		s.smoke = append(s.smoke, pt{w / 2, 0})
	}
	return s
}

func aTent(r *rnd, h int) *sprite {
	s := newSprite(8, 4)
	s.cone(0, 5, 1, 7)
	s.put(2, 0, '╨', sTrunk, sNone)
	s.put(2, 3, '▄', sWallDark, sRoof)
	s.put(2, 2, '▄', sWallDark, sRoof)
	s.put(6, 2, '╥', sTrunk, sNone)
	s.put(6, 3, '║', sTrunk, sNone)
	s.put(7, 2, '╥', sTrunk, sNone)
	s.put(7, 3, '║', sTrunk, sNone)
	s.flame = append(s.flame, pt{5, 3})
	return s
}

// tree draws a conifer or a broadleaf with its foot on row base.
func tree(s *sprite, x, base, h int, r *rnd) {
	if r.n(3) > 0 {
		top, bot := 2*(base-h)+1, 2*base-1
		cx := float64(x) + 1.5
		for py := top; py <= bot; py++ {
			hw := 0.35 + float64(float64(py-top)*0.34)
			for xx := x; xx < x+3; xx++ {
				if d := float64(xx) + 0.5 - cx; math.Abs(d) <= hw {
					sl := sLeaf
					if d > 0.2 {
						sl = sLeafDark
					}
					s.pset(xx, py, sl)
				}
			}
		}
	} else {
		cy := float64(2*base) - 3
		for py := 2*base - 6; py < 2*base; py++ {
			for xx := x - 1; xx < x+4; xx++ {
				dx := (float64(xx) + 0.5 - (float64(x) + 1.5)) / 2.2
				dy := (float64(py) + 0.5 - cy) / 2.6
				if float64(dx*dx)+float64(dy*dy) <= 1 {
					sl := sLeaf
					if dx > 0.25 || dy > 0.4 {
						sl = sLeafDark
					}
					s.pset(xx, py, sl)
				}
			}
		}
	}
	s.pset(x+1, 2*base, sTrunk)
	s.pset(x+1, 2*base+1, sTrunk)
}

func aWoodCamp(r *rnd, h int) *sprite {
	s := newSprite(9, 5)
	tree(s, 0, 4, 3, r)
	tree(s, 6, 4, 4, r)
	s.text(3, 3, "▄▄▄", sTrunk, sNone)
	s.text(3, 4, "◘◘◘", sTrunk, sWallDark)
	return s
}

func aStones(r *rnd, h int) *sprite {
	s := newSprite(7, 3)
	s.text(0, 0, "▄▀▀▀▀▀▄", sRock, sNone)
	for _, y := range []int{1, 2} {
		s.put(0, y, '█', sRock, sNone)
		s.put(6, y, '█', sRockDark, sNone)
	}
	s.put(3, 2, '▄', sRock, sNone)
	s.put(3, 1, '▲', sGlow, sNone)
	return s
}

func aFirepit(r *rnd, h int) *sprite {
	s := newSprite(9, 3)
	for _, x := range []int{1, 3, 5, 7} {
		s.put(x, 2, '☻', sWallDark, sNone)
	}
	s.put(4, 2, '▄', sGlow, sNone)
	s.put(2, 2, '▄', sTrunk, sNone)
	s.put(6, 2, '▄', sTrunk, sNone)
	s.flame = append(s.flame, pt{4, 1})
	s.smoke = append(s.smoke, pt{4, 0})
	return s
}

func aPile(r *rnd, h int) *sprite {
	s := newSprite(4, 2)
	s.text(1, 0, "▄▄", sRoof, sNone)
	s.text(0, 1, "▐▓▓▌", sRoof, sRoofShade)
	return s
}

func aLonghouse(r *rnd, h int) *sprite {
	w := r.rng(9, 12)
	s := newSprite(w, 4)
	s.hline(2, w-3, 0, '▄', sRoof, sNone)
	s.hline(1, w-2, 1, '█', sRoof, sNone)
	s.put(0, 1, '▄', sRoof, sNone)
	s.put(w-1, 1, '▄', sRoofShade, sNone)
	s.hline(w/2+1, w-2, 1, '█', sRoofShade, sNone)
	s.hline(0, w-1, 2, '▀', sRoofShade, sWall)
	s.hline(0, w/2, 2, '▀', sRoof, sWall)
	s.body(0, 3, w, 1)
	s.put(w/2, 3, '▄', sGlow, sWallDark)
	s.put(2, 3, '║', sWallDark, sWall)
	s.put(w-3, 3, '║', sWallDark, sWallShade)
	s.smoke = append(s.smoke, pt{w/2 - 2, 0})
	return s
}

func aStonesPit(r *rnd, h int) *sprite {
	w := r.rng(6, 9)
	s := newSprite(w, 3)
	for x := 0; x < w; x++ {
		hh := 1 + r.n(2)
		if x == 0 || x == w-1 {
			hh = 1
		}
		sl := sRock
		if x > w/2 {
			sl = sRockDark
		}
		for y := 3 - hh; y < 3; y++ {
			s.put(x, y, '█', sl, sNone)
		}
		if hh == 1 {
			s.put(x, 1, '▄', sRock, sNone)
		}
	}
	s.put(w/2, 1, '╥', sTrunk, sNone)
	s.put(w/2, 0, '╓', sTrunk, sNone)
	return s
}

// ---------------------------------------------------------------- bronze .. medieval

func aMudHouse(r *rnd, h int) *sprite {
	w := r.rng(5, 8)
	hh := max(3, min(h, r.rng(3, 5)))
	s := newSprite(w+3, hh+1)
	s.body(0, 1, w, hh)
	s.hline(0, w-1, 0, '▄', sTrim, sNone)
	s.windows(1, 2, w-2, hh-1, 3, 2, 0, 0, '■')
	s.put(w/2, hh, '▄', sWallDark, sWall)
	s.body(w, 3, 3, hh-2)
	s.hline(w, w+2, 2, '▄', sTrim, sNone)
	if r.n(2) == 0 {
		s.put(1, 0, '╥', sTrunk, sNone)
	}
	return s
}

func aZiggurat(r *rnd, h int) *sprite {
	tiers := max(3, min(5, h/2))
	w := tiers*4 + 1
	s := newSprite(w, tiers*2+2)
	base := s.h - 1
	for t := 0; t < tiers; t++ {
		x0, tw, y := t*2, w-t*4, base-t*2
		s.body(x0, y-1, tw, 2)
		s.hline(x0, x0+tw-1, y-1, '▀', sTrim, sWall)
	}
	cx := w / 2
	for y := base; y > base-tiers*2; y-- {
		s.put(cx, y, '▒', sTrim, sWallShade)
	}
	top := base - tiers*2
	s.text(cx-1, top, "▄█▄", sTrim, sNone)
	s.put(cx, top, '▄', sGlow, sNone)
	s.flame = append(s.flame, pt{cx, top - 1})
	return s
}

func aClassical(r *rnd, h int) *sprite {
	cols := r.rng(3, 6)
	w := cols*2 + 3
	colH := max(2, min(h-3, 3))
	s := newSprite(w, colH+5)
	top := 2
	s.roof(0, w, top, 2.6)
	for y := 0; y <= top; y++ {
		for x := 0; x < w; x++ {
			c := s.at(x, y)
			if c.ch == 0 {
				continue
			}
			fg, bg := sTrim, c.bg
			if bg != sNone {
				bg = sTrim
			}
			if x > w*2/3 {
				fg = sWallShade
			}
			s.put(x, y, c.ch, fg, bg)
		}
	}
	s.hline(0, w-1, top+1, '▀', sTrim, sWallDark)
	for y := top + 2; y < top+2+colH; y++ {
		for x := 0; x < w; x++ {
			if x%2 == 1 {
				s.put(x, y, '█', sTrim, sNone)
			} else {
				s.put(x, y, ' ', sWallDark, sWallDark)
			}
		}
	}
	s.hline(0, w-1, s.h-1, '▀', sTrim, sWallShade)
	s.put(0, s.h-1, '▄', sTrim, sNone)
	s.put(w-1, s.h-1, '▄', sTrim, sNone)
	return s.trimTop()
}

func aVilla(r *rnd, h int) *sprite {
	w := r.rng(6, 9) | 1
	floors := max(1, min(3, (h-2)/2))
	bh := floors * 2
	s := newSprite(w+1, bh+w/2+2)
	base := s.h - 1
	s.body(0, base-bh+1, w, bh)
	s.windows(1, base-bh+1, w-2, bh, 2, 2, 0, 0, '■')
	s.put(w/2, base, '▄', sWallDark, sWall)
	s.roof(0, w, base-bh, 1)
	if r.n(2) == 0 {
		s.put(w-2, base-bh-2, '▐', sWallDark, sNone)
		s.smoke = append(s.smoke, pt{w - 2, base - bh - 3})
	}
	return s.trimTop()
}

func aTimber(r *rnd, h int) *sprite {
	w := r.rng(5, 7)
	floors := max(2, min(4, h/2))
	s := newSprite(w, floors+w)
	base := s.h - 1
	for f := 0; f < floors; f++ {
		y, x0, ww := base-f, 0, w
		if f == 0 {
			x0, ww = 1, w-2 // jettied upper floors overhang the street
		}
		s.body(x0, y, ww, 1)
		for x := x0; x < x0+ww; x++ {
			if (x-x0)%2 == 0 {
				s.put(x, y, '║', sTrim, sWall)
			} else if f > 0 {
				s.put(x, y, '■', sWin, sWall)
			}
		}
	}
	s.put(w/2, base, '▄', sWallDark, sWall)
	s.roof(0, w, base-floors, 0.62)
	return s.trimTop()
}

func aTower(r *rnd, h int) *sprite {
	w := r.rng(3, 5)
	th := max(4, h)
	s := newSprite(w+2, th+2)
	base := s.h - 1
	s.body(1, base-th+2, w, th-1)
	s.crenels(0, w+2, base-th+1, sWall)
	for y := base - th + 3; y < base; y += 2 {
		s.put(1+w/2, y, '▌', sWin, sWall)
	}
	s.put(1+w/2, base, '▄', sWallDark, sWall)
	if r.n(2) == 0 {
		s.put(1+w/2, base-th, '►', sNeon2, sNone)
	}
	return s.trimTop()
}

func aCastle(r *rnd, h int) *sprite {
	th := max(5, h)
	w := 15
	s := newSprite(w, th+1)
	base := s.h - 1
	lt := aTower(newRnd(uint64(r.u())), th)
	s.blit(lt, 0, s.h-lt.h)
	rt := aTower(newRnd(uint64(r.u())), th-1)
	s.blit(rt, w-rt.w, s.h-rt.h)
	wall := th / 2
	s.body(lt.w, base-wall+1, w-lt.w-rt.w, wall)
	s.crenels(lt.w, w-lt.w-rt.w, base-wall, sWall)
	s.put(w/2, base, '█', sWallDark, sNone)
	s.put(w/2, base-1, '▄', sWallDark, sWall)
	return s
}

func aCathedral(r *rnd, h int) *sprite {
	th := max(8, h)
	w := 11
	s := newSprite(w, th+1)
	base := s.h - 1
	nave := max(4, th/2)
	s.body(0, base-nave+1, w, nave)
	s.roof(0, w, base-nave, 1)
	tx, tw := 2, 3
	tt := base - th + 3
	s.body(tx, tt, tw, base-tt)
	s.put(tx+1, tt-1, '█', sRoof, sNone)
	s.put(tx, tt-1, '▄', sRoof, sNone)
	s.put(tx+2, tt-1, '▄', sRoofShade, sNone)
	s.put(tx+1, tt-2, '█', sRoof, sNone)
	s.put(tx+1, tt-3, '▲', sRoof, sNone)
	s.put(tx+1, tt-4, '┼', sTrim, sNone)
	for y := tt + 1; y < base-1; y += 2 {
		s.put(tx+1, y, '▌', sWin, sWall)
	}
	s.put(7, base-nave+1, 'o', sGlow, sWall)
	for x := 6; x <= 9; x++ {
		s.put(x, base-1, '▌', sWin, sWall)
	}
	s.put(7, base, '▲', sWallDark, sWall)
	return s.trimTop()
}

func aDomed(r *rnd, h int) *sprite {
	w := r.rng(9, 13) | 1
	bh := max(3, min(h-4, 5))
	s := newSprite(w, bh+8)
	base := s.h - 1
	s.body(0, base-bh+1, w, bh)
	for x := 1; x < w-1; x += 2 {
		s.put(x, base-1, '█', sTrim, sWallDark)
		if bh > 3 {
			s.put(x, base-2, '█', sTrim, sWallDark)
		}
	}
	s.hline(0, w-1, base-bh+1, '▀', sTrim, sWall)
	s.windows(1, base-bh+2, w-2, max(0, bh-4), 2, 1, 0, 0, '■')
	dw, dy := w-4, base-bh
	s.body(2, dy, dw, 1)
	s.windows(2, dy, dw, 1, 2, 1, 1, 0, '▌')
	dr := s.dome(2, dw, dy-1, sMetal)
	s.put(w/2, dy-1-dr, '▲', sTrim, sNone)
	return s.trimTop()
}

func aWindmill(r *rnd, h int) *sprite {
	s := newSprite(7, 7)
	s.text(2, 1, "▄█▄", sRoof, sNone)
	for y := 2; y < 7; y++ {
		s.put(2, y, '█', sWall, sNone)
		s.put(3, y, '█', sWall, sNone)
		s.put(4, y, '█', sWallShade, sNone)
		if y >= 4 {
			s.put(1, y, '▐', sWall, sNone)
			s.put(5, y, '▌', sWallShade, sNone)
		}
	}
	s.put(3, 6, '▄', sWallDark, sWall)
	s.put(3, 4, '■', sWin, sWall)
	s.blades = append(s.blades, pt{3, 2})
	return s
}

func aFields(r *rnd, h int) *sprite {
	w := r.rng(10, 15)
	s := newSprite(w, 4)
	for x := 0; x < w; x++ {
		sl := sField1
		if (x/3)%2 == 1 {
			sl = sField2
		}
		ch := '▓'
		if x%2 == 0 {
			ch = '▒'
		}
		s.put(x, 3, ch, sl, sLeafDark)
	}
	b := r.n(w - 5)
	s.body(b, 2, 4, 1)
	s.put(b+1, 2, '▄', sWallDark, sWall)
	s.hline(b, b+3, 1, '▄', sRoof, sNone)
	s.put(b+2, 1, '▄', sRoofShade, sNone)
	s.put(b+3, 1, '▄', sRoofShade, sNone)
	if r.n(2) == 0 {
		s.put(b+5, 2, '▄', sMetal, sNone)
		s.put(b+5, 1, '▄', sMetal, sNone)
	}
	return s
}

func aSmithy(r *rnd, h int) *sprite {
	w := r.rng(6, 8)
	s := newSprite(w, 6)
	s.body(0, 3, w, 3)
	s.hline(0, w-1, 2, '▄', sRoof, sNone)
	s.hline(w/2, w-1, 2, '▄', sRoofShade, sNone)
	s.put(1, 4, '▄', sGlow, sWallDark)
	s.put(2, 4, '▄', sGlow, sWallDark)
	s.put(1, 5, '█', sWallDark, sNone)
	s.put(2, 5, '█', sWallDark, sNone)
	s.put(w-2, 1, '█', sWallDark, sNone)
	s.put(w-2, 2, '█', sWallDark, sNone)
	s.smoke = append(s.smoke, pt{w - 2, 0})
	s.windows(4, 4, w-5, 1, 2, 1, 0, 0, '■')
	return s
}

func aMarket(r *rnd, h int) *sprite {
	stalls := r.rng(2, 4)
	s := newSprite(stalls*4, 3)
	for i := 0; i < stalls; i++ {
		x := i * 4
		c := []slot{sRoof, sTrim, sNeon1, sNeon2}[r.n(4)]
		s.text(x, 0, "▄▄▄", c, sNone)
		s.text(x, 1, "▀▀▀", c, sNone)
		s.put(x+1, 1, '▀', sTrim, sNone)
		s.put(x, 2, '│', sTrunk, sNone)
		s.put(x+2, 2, '│', sTrunk, sNone)
		s.put(x+1, 2, r.pick("◘▪•o"), sField2, sNone)
	}
	return s
}

func aAqueduct(r *rnd, h int) *sprite {
	w := r.rng(11, 15)
	s := newSprite(w, 5)
	s.hline(0, w-1, 0, '▄', sTrim, sNone)
	for x := 0; x < w; x++ {
		if x%3 == 0 {
			s.vline(x, 2, 4, '█', sTrim, sNone)
		} else {
			s.put(x, 2, '▀', sTrim, sNone)
		}
	}
	s.hline(0, w-1, 1, '▄', sWater, sTrim)
	return s
}

// ---------------------------------------------------------------- industrial .. atomic

func aMineHead(r *rnd, h int) *sprite {
	s := newSprite(9, 7)
	for y := 1; y < 7; y++ {
		s.put(1, y, '╱', sMetal, sNone)
		s.put(3, y, '║', sMetal, sNone)
	}
	s.put(3, 0, 'O', sMetal, sNone)
	s.hline(1, 3, 1, '═', sMetal, sNone)
	s.body(4, 4, 5, 3)
	s.hline(4, 8, 3, '▄', sRoof, sNone)
	s.put(7, 2, '█', sWallDark, sNone)
	s.put(7, 3, '█', sWallDark, sNone)
	s.smoke = append(s.smoke, pt{7, 1})
	s.put(5, 5, '■', sWin, sWall)
	s.put(1, 6, '▄', sRock, sNone)
	s.put(0, 6, '▄', sRockDark, sNone)
	return s
}

func aFactory(r *rnd, h int) *sprite {
	teeth := r.rng(3, 5)
	w := teeth*3 + 1
	bh := max(3, min(h-2, 5))
	stacks := r.rng(1, 3)
	sh := bh + r.rng(3, 6)
	s := newSprite(w+2, sh+2)
	base := s.h - 1
	s.body(0, base-bh+1, w, bh)
	for t := 0; t < teeth; t++ {
		x, y := t*3, base-bh
		s.put(x, y, '█', sRoof, sNone)
		s.put(x, y-1, '▄', sRoof, sNone)
		s.put(x+1, y, '▄', sGlass, sNone)
		s.put(x+2, y, '▄', sRoofShade, sNone)
	}
	s.windows(1, base-bh+2, w-2, bh-2, 2, 2, 0, 0, '■')
	s.put(2, base, '▄', sWallDark, sWall)
	for i := 0; i < stacks; i++ {
		x := w - 2 - i*3
		if x < 1 {
			break
		}
		top := base - sh + i
		s.vline(x, top, base-bh, '█', sWallShade, sNone)
		s.put(x, top+1, '▀', sTrim, sWallShade)
		s.smoke = append(s.smoke, pt{x, top - 1})
	}
	return s.trimTop()
}

func aRowHouse(r *rnd, h int) *sprite {
	units := r.rng(2, 4)
	floors := max(2, min(5, h-1))
	w := units * 3
	s := newSprite(w, floors+3)
	base := s.h - 1
	s.body(0, base-floors+1, w, floors)
	s.hline(0, w-1, base-floors, '▄', sTrim, sNone)
	for u := 0; u < units; u++ {
		x := u * 3
		for f := 0; f < floors; f++ {
			s.put(x+1, base-f, '■', sWin, sWall)
		}
		s.put(x+1, base, '▄', sWallDark, sWall)
		if u%2 == 0 {
			s.put(x, base-floors-1, '▐', sWallShade, sNone)
			s.put(x, base-floors, '█', sWallShade, sNone)
			if r.n(2) == 0 {
				s.smoke = append(s.smoke, pt{x, base - floors - 2})
			}
		}
	}
	return s.trimTop()
}

func aCooling(r *rnd, h int) *sprite {
	th := max(6, min(h, 9))
	s := newSprite(9, th+1)
	base := s.h - 1
	for i := 0; i < th; i++ {
		t := float64(i) / float64(th-1)
		wd := 4 - int(float64(float64(3*t)*float64((1.35-t)*1.9))+0.2)
		if wd < 2 {
			wd = 2
		}
		for dx := -wd; dx <= wd; dx++ {
			sl := sWall
			if dx > wd/2 {
				sl = sWallShade
			} else if dx == -wd {
				sl = sWallLit
			}
			s.put(4+dx, base-i, '█', sl, sNone)
		}
	}
	s.smoke = append(s.smoke, pt{3, base - th}, pt{5, base - th})
	return s.trimTop()
}

func aDerrick(r *rnd, h int) *sprite {
	th := max(5, min(h, 9))
	s := newSprite(7, th+1)
	base := s.h - 1
	for i := 0; i < th; i++ {
		y := base - i
		off := (th - i) * 3 / th
		s.put(3-off, y, '╱', sMetal, sNone)
		s.put(3+off, y, '╲', sMetal, sNone)
		if off > 0 && i%2 == 1 {
			s.hline(3-off+1, 3+off-1, y, '─', sMetalDark, sNone)
		}
	}
	s.put(3, base-th, '▲', sMetal, sNone)
	s.beacons = append(s.beacons, pt{3, base - th - 1})
	return s
}

func aDeco(r *rnd, h int) *sprite {
	w := r.rng(5, 7) | 1
	th := max(8, h)
	s := newSprite(w, th+3)
	base := s.h - 1
	for _, g := range []struct{ inset, top int }{{0, th * 6 / 10}, {1, th * 85 / 100}, {2, th}} {
		ww := w - 2*g.inset
		if ww < 1 {
			continue
		}
		y0 := base - g.top + 1
		s.body(g.inset, y0, ww, base-y0+1)
		s.hline(g.inset, g.inset+ww-1, y0, '▀', sTrim, sWall)
	}
	for x := 1; x < w-1; x += 2 {
		for y := base - th + 2; y < base; y++ {
			if c := s.at(x, y); c.ch == '█' {
				s.put(x, y, '▌', sWin, c.fg)
			}
		}
	}
	cx := w / 2
	s.put(cx, base-th, '█', sTrim, sNone)
	s.put(cx, base-th-1, '▌', sTrim, sNone)
	s.put(cx, base-th-2, '│', sMetal, sNone)
	s.beacons = append(s.beacons, pt{cx, base - th - 2})
	s.put(cx, base, '▄', sWallDark, sWall)
	return s.trimTop()
}

func aPylon(r *rnd, h int) *sprite {
	th := max(5, min(h, 8))
	s := newSprite(5, th+1)
	base := s.h - 1
	for i := 0; i < th; i++ {
		y := base - i
		if i < th/2 {
			s.put(1, y, '╱', sMetal, sNone)
			s.put(3, y, '╲', sMetal, sNone)
			s.put(2, y, '╳', sMetalDark, sNone)
		} else {
			s.put(2, y, '║', sMetal, sNone)
		}
	}
	s.text(0, base-th+2, "══╬══", sMetal, sNone)
	return s
}

// ---------------------------------------------------------------- modern .. neon

func aGlass(r *rnd, h int) *sprite {
	w := r.rng(4, 8)
	th := max(8, h)
	s := newSprite(w, th+4)
	base := s.h - 1
	top := base - th + 1
	for y := top; y <= base; y++ {
		for x := 0; x < w; x++ {
			sl := sGlass
			switch {
			case x == 0:
				sl = sGlassHi
			case x == w-1 || (w > 5 && x == w-2):
				sl = sWallShade
			}
			s.put(x, y, '▀', sWin, sl)
		}
	}
	if w >= 6 {
		s.vline(w/2, top, base, '█', sWall, sNone)
	}
	s.hline(0, w-1, base, '█', sWall, sNone)
	s.put(w/2, base, '▄', sGlassHi, sWallDark)
	switch r.n(4) {
	case 0:
		s.vline(w/2, top-2, top-1, '│', sMetal, sNone)
		s.beacons = append(s.beacons, pt{w / 2, top - 3})
	case 1:
		for i := 0; i < w-1; i++ {
			s.put(i, top-1-i/2, '▄', sGlassHi, sNone)
			for y := top - i/2; y < top; y++ {
				s.put(i, y, '█', sGlass, sNone)
			}
		}
	case 2:
		s.hline(1, w-2, top-1, '▄', sWall, sNone)
		s.beacons = append(s.beacons, pt{1, top - 2})
	default:
		s.hline(0, w-1, top-1, '▄', sWall, sNone)
	}
	return s.trimTop()
}

func aDish(r *rnd, h int) *sprite {
	s := newSprite(7, 6)
	s.text(0, 0, "▄▀▀▀▄", sMetal, sNone)
	s.text(0, 1, "▀▄▄▄▀", sMetal, sNone)
	s.put(2, 1, '▄', sMetalDark, sNone)
	s.vline(2, 2, 3, '█', sMetalDark, sNone)
	s.body(0, 4, 6, 2)
	s.windows(1, 4, 4, 2, 2, 1, 0, 0, '■')
	s.beacons = append(s.beacons, pt{4, 0})
	return s
}

func aDataCenter(r *rnd, h int) *sprite {
	w := r.rng(10, 14)
	s := newSprite(w, 5)
	s.body(0, 2, w, 3)
	for x := 1; x < w-1; x++ {
		if x%2 == 1 {
			s.put(x, 3, '·', sNeon2, sWallDark)
			s.put(x, 2, '·', sNeon3, sWallDark)
		} else {
			s.put(x, 3, '█', sWallDark, sNone)
			s.put(x, 2, '█', sWallDark, sNone)
		}
	}
	for x := 1; x < w-1; x += 3 {
		s.text(x, 1, "▄▄", sMetal, sNone)
	}
	s.beacons = append(s.beacons, pt{w - 2, 0})
	return s
}

const neonGlyphs = "¥£ƒΣΩ∞≡♦§¤"

func aCyber(r *rnd, h int) *sprite {
	w := r.rng(5, 9)
	th := max(12, h)
	s := newSprite(w+2, th+4)
	base := s.h - 1
	top := base - th + 1
	for y := top; y <= base; y++ {
		for x := 1; x <= w; x++ {
			sl := sWall
			if x == w {
				sl = sWallShade
			}
			s.put(x, y, '█', sl, sNone)
		}
		if (y-top)%2 == 1 {
			for x := 2; x < w; x++ {
				if r.n(3) != 0 {
					s.put(x, y, '▄', sWin, sWall)
				}
			}
		}
	}
	n := []slot{sNeon1, sNeon2, sNeon3}[r.n(3)]
	s.vline(1, top, base, '▌', n, sWall)
	sy := top + 2 + r.n(max(1, th/3))
	for i := 0; i < r.rng(3, 5); i++ {
		s.put(w+1, sy+i, r.pick(neonGlyphs), sNeon1, sInk)
	}
	s.put(0, top+th/2, '▐', sNeon2, sNone)
	if r.n(2) == 0 {
		s.hline(1, w, top-1, '▄', sNeon2, sNone)
		s.hline(2, w-1, top-2, '▄', sMetalDark, sNone)
	} else {
		s.vline(w/2+1, top-2, top-1, '│', sMetal, sNone)
		s.beacons = append(s.beacons, pt{w/2 + 1, top - 3})
	}
	return s.trimTop()
}

func aVertFarm(r *rnd, h int) *sprite {
	w := r.rng(5, 7)
	th := max(8, h)
	s := newSprite(w, th+1)
	base := s.h - 1
	for y := base - th + 1; y <= base; y++ {
		for x := 0; x < w; x++ {
			sl := sGlass
			if x == w-1 {
				sl = sWallShade
			}
			s.put(x, y, '█', sl, sNone)
		}
		if (base-y)%2 == 0 {
			s.hline(0, w-2, y, '▄', sLeaf, sGlass)
		}
	}
	s.hline(0, w-1, base-th, '▄', sMetal, sNone)
	return s
}

func aBunker(r *rnd, h int) *sprite {
	w := r.rng(8, 11)
	s := newSprite(w, 6)
	s.body(0, 3, w, 3)
	s.hline(1, w-1, 2, '▄', sWall, sNone)
	for x := 1; x < w-1; x += 3 {
		s.put(x, 4, '▬', sWallDark, sWall)
	}
	s.text(w-4, 0, "▀▄▄▀", sMetal, sNone)
	s.put(w-3, 1, '▐', sMetalDark, sNone)
	s.put(w-3, 2, '█', sMetalDark, sNone)
	s.put(1, 1, '►', sBeacon, sNone)
	s.put(1, 2, '│', sMetal, sNone)
	return s
}

func aCrane(r *rnd, h int) *sprite {
	s := newSprite(11, 8)
	s.vline(2, 1, 7, '╫', sNeon2, sNone)
	s.hline(0, 10, 1, '═', sNeon2, sNone)
	s.put(8, 2, '│', sMetalDark, sNone)
	s.put(8, 3, '▄', sMetal, sNone)
	cols := []slot{sRoof, sTrim, sNeon1, sMetal, sGlass}
	for i := 0; i < 6; i++ {
		x, y := 4+(i%3)*2, 7-i/3
		s.put(x, y, '█', cols[r.n(len(cols))], sNone)
		s.put(x+1, y, '▌', sWallDark, sNone)
	}
	s.beacons = append(s.beacons, pt{2, 0})
	return s
}

func aSilo(r *rnd, h int) *sprite {
	n := r.rng(2, 3)
	th := max(3, min(h, 6))
	s := newSprite(n*3, th+1)
	base := s.h - 1
	for i := 0; i < n; i++ {
		x := i * 3
		s.vline(x, base-th+1, base, '█', sWallLit, sNone)
		s.vline(x+1, base-th+1, base, '█', sWallShade, sNone)
		s.put(x, base-th, '▄', sRoof, sNone)
		s.put(x+1, base-th, '▄', sRoofShade, sNone)
		s.put(x, base-th/2, '▀', sTrim, sWallLit)
		s.put(x+1, base-th/2, '▀', sTrim, sWallShade)
	}
	return s
}

func aWarehouse(r *rnd, h int) *sprite {
	w := r.rng(7, 10)
	s := newSprite(w, 4)
	s.body(0, 1, w, 3)
	s.hline(0, w-1, 0, '▄', sRoof, sNone)
	s.hline(w/2, w-1, 0, '▄', sRoofShade, sNone)
	s.put(2, 3, '█', sWallDark, sNone)
	s.put(3, 3, '█', sWallDark, sNone)
	s.put(2, 2, '▄', sWallDark, sWall)
	s.put(3, 2, '▄', sWallDark, sWall)
	s.windows(5, 2, w-6, 1, 2, 1, 0, 0, '■')
	return s
}

func aVault(r *rnd, h int) *sprite {
	w := r.rng(6, 8)
	s := newSprite(w, 4)
	s.body(0, 1, w, 3)
	s.hline(0, w-1, 0, '▄', sTrim, sNone)
	s.hline(0, w-1, 1, '▀', sTrim, sWall)
	s.put(w/2, 2, 'o', sMetal, sWallDark)
	s.put(w/2, 3, '▀', sWallDark, sWall)
	return s
}

func aSolar(r *rnd, h int) *sprite {
	w := r.rng(8, 12)
	s := newSprite(w, 2)
	for x := 0; x < w; x++ {
		if x%4 != 3 {
			s.put(x, 0, '▄', sGlass, sNone)
			s.put(x, 1, '▀', sGlassHi, sNone)
		} else {
			s.put(x, 1, '│', sMetal, sNone)
		}
	}
	return s
}

func aObelisk(r *rnd, h int) *sprite {
	th := max(5, min(h, 9))
	s := newSprite(3, th+1)
	base := s.h - 1
	s.put(1, base-th+1, '▲', sTrim, sNone)
	s.vline(1, base-th+2, base-1, '█', sTrim, sNone)
	s.text(0, base, "▄█▄", sTrim, sNone)
	return s
}

func aHabDome(r *rnd, h int) *sprite {
	w := r.rng(7, 11) | 1
	s := newSprite(w, 5)
	s.body(0, 4, w, 1)
	s.dome(0, w, 3, sGlass)
	for x := 2; x < w-2; x += 2 {
		s.put(x, 3, '▪', sNeon3, sGlass)
	}
	s.hline(0, w-1, 4, '▀', sTrim, sWall)
	return s.trimTop()
}

func aFusion(r *rnd, h int) *sprite {
	s := newSprite(11, 7)
	s.body(0, 4, 11, 3)
	s.dome(1, 9, 3, sMetal)
	s.hline(2, 8, 2, '═', sGlow, sMetal)
	s.windows(1, 5, 9, 2, 2, 1, 0, 0, '▀')
	s.beacons = append(s.beacons, pt{10, 1})
	return s.trimTop()
}

func aLaunch(r *rnd, h int) *sprite {
	th := max(10, min(h, 16))
	s := newSprite(9, th+2)
	base := s.h - 1
	s.vline(1, base-th+1, base, '╫', sMetal, sNone)
	for y := base - th + 3; y <= base; y += 3 {
		s.put(2, y, '─', sMetal, sNone)
	}
	rt := base - th + 2
	s.put(4, rt, '▲', sTrim, sNone)
	for y := rt + 1; y < base; y++ {
		s.put(4, y, '█', sTrim, sNone)
		s.put(3, y, '▐', sWallShade, sNone)
		s.put(5, y, '▌', sWallShade, sNone)
	}
	s.put(4, rt+3, '▀', sNeon2, sTrim)
	s.put(3, base-1, '▄', sWallShade, sNone)
	s.put(5, base-1, '▄', sWallShade, sNone)
	s.hline(0, 8, base, '▀', sMetalDark, sNone)
	s.smoke = append(s.smoke, pt{6, base - 1})
	return s
}

// ---------------------------------------------------------------- fusion .. transcendent

// aRingSpire: a tapering shaft carrying stacked rings, each lit by a band
// of light underneath: the late city's signature tower.
func aRingSpire(r *rnd, h int) *sprite {
	th := max(12, h)
	w := 13
	s := newSprite(w, th+3)
	base := s.h - 1
	top := base - th + 1
	cx := w / 2
	gap := max(3, th/4)
	for y := top; y <= base; y++ {
		t := float64(y-top) / float64(th)
		half := 1 + int(float64(t*1.6)+0.3)
		if t < 0.12 {
			half = 0 // only the very tip is a needle
		}
		for dx := -half; dx <= half; dx++ {
			sl := sWall
			switch {
			case dx < 0 && dx == -half:
				sl = sWallLit
			case dx > 0:
				sl = sWallShade
			}
			s.put(cx+dx, y, '█', sl, sNone)
		}
		if half >= 1 && (y-top)%2 == 1 {
			s.put(cx, y, '▪', sWin, sWall)
		}
		if k := y - top - 2; k > 0 && k%gap == 0 && y < base-1 {
			rw := half + 3 + r.n(2)
			s.hline(cx-rw, cx+rw, y, '▄', sMetal, sNone)
			for dx := -rw; dx <= rw; dx++ {
				if c := s.at(cx+dx, y+1); c.ch == 0 {
					s.put(cx+dx, y+1, '▀', sNeon1, sNone)
				} else {
					s.put(cx+dx, y+1, '▀', sNeon1, c.fg)
				}
			}
			s.put(cx+rw, y, '▄', sMetalDark, sNone)
		}
	}
	s.text(cx-1, top-1, "▄█▄", sGlow, sNone)
	s.put(cx, top-2, '│', sMetal, sNone)
	s.beacons = append(s.beacons, pt{cx, top - 3})
	s.put(cx, base, '▄', sGlow, sWallDark)
	return s.trimTop()
}

// aSaucer: habitat discs held up on a stem; tall ones carry a second disc.
func aSaucer(r *rnd, h int) *sprite {
	th := max(6, h)
	w := r.rng(9, 13) | 1
	s := newSprite(w, th+4)
	base := s.h - 1
	cx := w / 2
	disc := func(y, dw int) {
		x0 := cx - dw/2
		s.hline(x0, x0+dw-1, y, '█', sWall, sNone)
		s.put(x0, y, '▄', sWallLit, sNone)
		s.put(x0+dw-1, y, '▄', sWallShade, sNone)
		for x := x0 + 1; x < x0+dw-1; x += 2 {
			s.put(x, y, '▪', sWin, sWall)
		}
		s.hline(x0+1, x0+dw-2, y+1, '▀', sNeon1, sNone)
		s.dome(x0+2, dw-4, y-1, sGlass)
	}
	stem := base - th*45/100 + 1
	s.vline(cx, stem, base, '█', sWallShade, sNone)
	s.put(cx-1, base, '▄', sWallShade, sNone)
	s.put(cx+1, base, '▄', sWallShade, sNone)
	disc(stem-1, w)
	if th >= 12 {
		up := stem - 1 - th/3
		s.vline(cx, up+1, stem-2, '▐', sWallShade, sNone)
		disc(up, w-4)
	}
	if r.n(2) == 0 {
		s.beacons = append(s.beacons, pt{cx, 0})
	}
	return s.trimTop()
}

// aArcology: a terraced ziggurat of towers with hanging gardens and a
// band of light at every terrace.
func aArcology(r *rnd, h int) *sprite {
	th := max(10, h)
	w := th + 5
	if w%2 == 0 {
		w++
	}
	s := newSprite(w, th+3)
	base := s.h - 1
	tiers := 4
	for t := 0; t < tiers; t++ {
		inset := t * (w / 9)
		y0 := base - (t+1)*th/tiers + 1
		y1 := base - t*th/tiers
		s.body(inset, y0, w-2*inset, y1-y0+1)
		s.hline(inset, w-1-inset, y0, '▄', sLeaf, sNone)
		for x := inset; x < w-inset; x += 3 {
			s.put(x, y0+1, '▀', sLeafDark, s.at(x, y0+1).fg)
		}
		for y := y0 + 2; y <= y1; y++ {
			for x := inset + 1; x < w-1-inset; x += 2 {
				s.put(x, y, '▀', sWin, s.at(x, y).fg)
			}
		}
		if y1 < base {
			s.hline(inset, w-1-inset, y1, '▀', sNeon3, sWall)
		}
	}
	cx := w / 2
	s.put(cx, base-th, '▲', sGlassHi, sNone)
	s.put(cx, base-th-1, '│', sMetal, sNone)
	s.beacons = append(s.beacons, pt{cx, base - th - 2})
	return s.trimTop()
}

// aCrystal: a cluster of glass prisms with pointed tops, facet lines and a
// lit core.
func aCrystal(r *rnd, h int) *sprite {
	th := max(8, h)
	n := r.rng(2, 3)
	s := newSprite(n*4+1, th+3)
	base := s.h - 1
	for i := 0; i < n; i++ {
		x := i * 4
		ph := th - r.n(max(1, th/2))
		if i == n/2 {
			ph = th
		}
		top := base - ph + 1
		for y := top + 1; y <= base; y++ {
			s.put(x, y, '█', sGlassHi, sNone)
			s.put(x+1, y, '█', sGlass, sNone)
			s.put(x+2, y, '█', sWallShade, sNone)
			if (y+i)%4 == 0 {
				s.put(x+1, y, '╱', sGlassHi, sGlass)
			}
		}
		s.put(x, top, '▄', sGlassHi, sNone)
		s.put(x+1, top, '█', sGlass, sNone)
		s.put(x+2, top, '▄', sWallShade, sNone)
		s.put(x+1, top-1, '▲', sGlassHi, sNone)
		s.vline(x+1, top+2, base-1, '│', sNeon3, sGlass)
	}
	s.hline(0, s.w-1, base, '▀', sMetalDark, sGlass)
	return s.trimTop()
}

// aHabRing: a habitat torus, seen edge on, held up on a pylon.
func aHabRing(r *rnd, h int) *sprite {
	th := max(7, h)
	w := r.rng(11, 15) | 1
	s := newSprite(w, th+2)
	base := s.h - 1
	cx := w / 2
	ry := base - th + 1
	s.text(1, ry, "▄"+string(repeat('▀', w-4))+"▄", sMetal, sNone)
	s.put(0, ry+1, '█', sWallLit, sNone)
	s.put(w-1, ry+1, '█', sWallShade, sNone)
	for x := 1; x < w-1; x++ {
		s.put(x, ry+1, '▪', sWin, sWall)
		if x%2 == 0 {
			s.put(x, ry+1, '█', sWall, sNone)
		}
	}
	s.text(1, ry+2, "▀"+string(repeat('▄', w-4))+"▀", sMetalDark, sNone)
	s.hline(2, w-3, ry+2, '▀', sNeon1, sNone)
	s.vline(cx, ry+3, base, '█', sWallShade, sNone)
	s.put(cx-1, ry+3, '╲', sMetal, sNone)
	s.put(cx+1, ry+3, '╱', sMetal, sNone)
	s.text(cx-1, base, "▄█▄", sWallShade, sNone)
	s.beacons = append(s.beacons, pt{cx, ry - 1})
	return s.trimTop()
}

// aObservatory: a slit dome with the telescope showing, on a drum.
func aObservatory(r *rnd, h int) *sprite {
	w := r.rng(9, 11) | 1
	bh := max(3, min(h-4, 6))
	s := newSprite(w, bh+7)
	base := s.h - 1
	s.body(0, base-bh+1, w, bh)
	s.windows(1, base-bh+2, w-2, bh-1, 2, 2, 0, 0, '■')
	d := s.dome(1, w-2, base-bh, sMetal)
	cx := w / 2
	for y := base - bh - d + 1; y <= base-bh; y++ {
		s.put(cx, y, '█', sWallDark, sNone)
	}
	s.put(cx+1, base-bh-d, '╱', sGlassHi, sNone)
	s.put(cx+2, base-bh-d-1, '╱', sGlassHi, sNone)
	return s.trimTop()
}

// aTesseract: nested frames that seem to hang in the air over a plinth.
func aTesseract(r *rnd, h int) *sprite {
	sz := max(5, min(h-2, 11))
	w := sz*2 - 1
	s := newSprite(w, sz+3)
	base := s.h - 1
	top := 0
	s.hline(0, w-1, top, '▄', sMetal, sNone)
	s.hline(0, w-1, top+sz-1, '▀', sMetal, sNone)
	s.vline(0, top+1, top+sz-2, '█', sWallLit, sNone)
	s.vline(w-1, top+1, top+sz-2, '█', sWallShade, sNone)
	in := sz / 2
	if in >= 2 {
		s.hline(in, w-1-in, top+in/2+1, '▄', sNeon1, sNone)
		s.hline(in, w-1-in, top+sz-in/2-2, '▀', sNeon1, sNone)
		s.vline(in, top+in/2+2, top+sz-in/2-3, '▐', sNeon1, sNone)
		s.vline(w-1-in, top+in/2+2, top+sz-in/2-3, '▌', sNeon1, sNone)
		s.put(1, top+1, '╲', sNeon2, sNone)
		s.put(w-2, top+1, '╱', sNeon2, sNone)
		s.put(1, top+sz-2, '╱', sNeon2, sNone)
		s.put(w-2, top+sz-2, '╲', sNeon2, sNone)
	}
	s.put(w/2, top+sz/2, '◊', sGlow, sNone)
	s.put(w/2, base-1, '▐', sMetalDark, sNone)
	s.text(w/2-1, base, "▄█▄", sMetalDark, sNone)
	return s
}

// aPrism: a glass pyramid sending a beam of light straight up.
func aPrism(r *rnd, h int) *sprite {
	pw := r.rng(9, 13) | 1
	beam := max(2, h-pw/2-1)
	s := newSprite(pw, pw/2+beam+2)
	base := s.h - 1
	cx := pw / 2
	for i := 0; i <= cx; i++ {
		y := base - cx + i
		for x := cx - i; x <= cx+i; x++ {
			sl := sGlass
			switch {
			case x == cx-i:
				sl = sGlassHi
			case x > cx:
				sl = sWallShade
			}
			s.put(x, y, '█', sl, sNone)
		}
		if i > 0 {
			s.put(cx-i, y, '◢', sGlassHi, sNone)
			s.put(cx+i, y, '◣', sWallShade, sNone)
		}
	}
	s.vline(cx, base-cx+1, base, '│', sGlow, sGlass)
	s.vline(cx, base-cx-beam, base-cx-1, '│', sGlow, sNone)
	return s.trimTop()
}

// aPillar: a column of light held by rings, the transcendent city.
func aPillar(r *rnd, h int) *sprite {
	th := max(10, h)
	s := newSprite(7, th+1)
	base := s.h - 1
	s.vline(3, base-th+1, base, '║', sGlow, sNone)
	for y := base - 2; y > base-th+1; y -= max(2, th/5) {
		s.text(1, y, "═◊═", sMetal, sNone)
		s.put(0, y, '▐', sNeon1, sNone)
		s.put(6, y, '▌', sNeon1, sNone)
		s.put(3, y, '◊', sNeon2, sNone)
	}
	s.text(1, base, "▄▀▀▀▄", sMetalDark, sNone)
	s.put(3, base-th, '·', sGlow, sNone)
	return s
}

// aGenShip: a generation ship on its building cradle.
func aGenShip(r *rnd, h int) *sprite {
	w := r.rng(15, 19)
	s := newSprite(w, 7)
	s.text(1, 1, "▄"+string(repeat('▄', w-5))+"▄", sWall, sNone)
	s.hline(0, w-2, 2, '█', sWall, sNone)
	s.put(w-1, 2, '►', sWallShade, sNone)
	for x := 2; x < w-3; x += 2 {
		s.put(x, 2, '▪', sWin, sWall)
	}
	s.hline(1, w-3, 3, '▀', sWallShade, sNone)
	s.put(0, 1, '▄', sGlow, sNone)
	s.put(0, 3, '▀', sGlow, sNone)
	s.put(w/2, 0, '▲', sMetal, sNone)
	for x := 3; x < w-3; x += 4 {
		s.vline(x, 4, 6, '║', sMetalDark, sNone)
	}
	s.hline(0, w-1, 6, '▀', sMetalDark, sNone)
	return s
}

func repeat(r rune, n int) []rune {
	if n < 0 {
		n = 0
	}
	out := make([]rune, n)
	for i := range out {
		out[i] = r
	}
	return out
}
