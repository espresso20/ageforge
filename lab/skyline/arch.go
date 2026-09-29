package main

import "math"

// arch.go is the building grammar: ~30 parametric archetypes drawn in
// CP437 blocks and line art. A building is (archetype, height, seed): the
// archetype comes from its lineage and period, the height from its period
// and the scene scale, the seed from the save's seed and the lot. There are
// no per-building sprites; 301 buildings map onto this handful of forms.

type rnd struct{ s uint32 }

func newRnd(v ...int) *rnd { return &rnd{hash(v...) | 1} }

func (r *rnd) u() uint32 {
	r.s ^= r.s << 13
	r.s ^= r.s >> 17
	r.s ^= r.s << 5
	return r.s
}
func (r *rnd) n(k int) int {
	if k <= 0 {
		return 0
	}
	return int(r.u() % uint32(k))
}
func (r *rnd) f() float64       { return float64(r.u()%10000) / 10000 }
func (r *rnd) rng(a, b int) int { return a + r.n(b-a+1) }
func (r *rnd) pick(s string) rune {
	rs := []rune(s)
	return rs[r.n(len(rs))]
}

type archFn func(r *rnd, h int) *Sprite

// gable draws a pitched roof over [x, x+w) with its eave on cell row y,
// rasterised in half-cell pixels so the slope is smooth. steep roofs rise
// one column per pixel row (a Gothic pitch), others two. Returns the cell
// rows it used.
func (s *Sprite) gable(x, w, y int, steep bool) int {
	if steep {
		return s.roof(x, w, y, 0.62)
	}
	return s.roof(x, w, y, 1.0)
}

// roof is gable with an explicit slope in columns per half-row.
func (s *Sprite) roof(x, w, y int, slope float64) int {
	cx := float64(x) + float64(w)/2
	half := float64(w) / 2
	rows := int(math.Ceil(half / slope))
	bot := 2*y + 1
	for i := 0; i < rows; i++ {
		py := bot - (rows - 1 - i)
		hw := (float64(i) + 1) * slope
		for xx := x; xx < x+w; xx++ {
			d := math.Abs(float64(xx) + 0.5 - cx)
			if d <= hw+0.01 {
				slot := SRoof
				if float64(xx)+0.5 > cx {
					slot = SRoofShade
				}
				s.pset(xx, py, slot)
			}
		}
	}
	return (rows + 1) / 2
}

// crenels draws battlements on row y across [x, x+w).
func (s *Sprite) crenels(x, w, y int, slot Slot) {
	for i := 0; i < w; i++ {
		if i%2 == 0 {
			s.put(x+i, y, '█', slot, SNone)
		} else {
			s.put(x+i, y, '▄', slot, SNone)
		}
	}
}

func (s *Sprite) dome(x, w, y int, slot Slot) int {
	// y is the cell row the dome sits on top of (its bottom row)
	rx := float64(w) / 2
	ry := rx * 0.95 // in pixels: a dome is about as tall as half its width
	cx := float64(x) + rx
	bot := 2*y + 1
	n := int(math.Ceil(ry))
	for i := 0; i < n; i++ {
		py := bot - i
		yy := (float64(i) + 0.5) / ry
		if yy > 1 {
			continue
		}
		hw := rx * math.Sqrt(1-yy*yy)
		for xx := x; xx < x+w; xx++ {
			d := float64(xx) + 0.5 - cx
			if math.Abs(d) <= hw+0.15 {
				sl := slot
				if d > rx*0.35 {
					sl = SMetalDark
				}
				s.pset(xx, py, sl)
			}
		}
	}
	return (n + 1) / 2
}

// ---------------------------------------------------------------- early

// cone rasterises a conical roof: apex at pixel row top, full width w at
// pixel row bot, lit on the left, shaded on the right.
func (s *Sprite) cone(x, w, top, bot int) {
	cx := float64(x) + float64(w)/2
	n := float64(bot - top)
	for py := top; py <= bot; py++ {
		hw := 0.5 + (float64(w)/2-0.5)*float64(py-top)/n
		for xx := x; xx < x+w; xx++ {
			d := float64(xx) + 0.5 - cx
			if math.Abs(d) <= hw {
				sl := SRoof
				if d > 0.4 {
					sl = SRoofShade
				}
				s.pset(xx, py, sl)
			}
		}
	}
}

func aHut(r *rnd, h int) *Sprite {
	w := 7 + 2*r.n(2)
	s := newSprite(w, 5)
	s.cone(0, w, 0, 7)
	// eaves hang over a low wattle wall with a lit doorway
	s.hline(1, w-2, 4, '█', SWall, SNone)
	s.put(w-2, 4, '█', SWallShade, SNone)
	s.put(w/2, 4, '▄', SGlow, SWallDark)
	s.put(w/2, 3, '▄', SWallDark, s.at(w/2, 3).Fg)
	s.put(0, 4, ' ', SNone, SNone)
	s.put(w-1, 4, ' ', SNone, SNone)
	if r.n(2) == 0 {
		s.Smoke = append(s.Smoke, pt{w / 2, -1 + 1})
	}
	return s
}

func aTent(r *rnd, h int) *Sprite {
	s := newSprite(8, 4)
	s.cone(0, 5, 1, 7)
	s.put(2, 0, '╨', STrunk, SNone) // poles through the smoke hole
	s.put(2, 3, '▄', SWallDark, SRoof)
	s.put(2, 2, '▄', SWallDark, SRoof)
	// a drying rack and the cooking fire
	s.put(6, 2, '╥', STrunk, SNone)
	s.put(6, 3, '║', STrunk, SNone)
	s.put(7, 2, '╥', STrunk, SNone)
	s.put(7, 3, '║', STrunk, SNone)
	s.Flame = append(s.Flame, pt{5, 3})
	return s
}

func tree(s *Sprite, x, base, h int, r *rnd) {
	if r.n(3) > 0 { // conifer: a half-block triangle on a stub of trunk
		top := 2*(base-h) + 1
		bot := 2*base - 1
		cx := float64(x) + 1.5
		for py := top; py <= bot; py++ {
			hw := 0.35 + float64(py-top)*0.34
			for xx := x; xx < x+3; xx++ {
				d := float64(xx) + 0.5 - cx
				if math.Abs(d) <= hw {
					sl := SLeaf
					if d > 0.2 {
						sl = SLeafDark
					}
					s.pset(xx, py, sl)
				}
			}
		}
		s.pset(x+1, 2*base, STrunk)
		s.pset(x+1, 2*base+1, STrunk)
		return
	}
	// broadleaf: a lumpy ellipse
	cy := float64(2*base) - 3
	for py := 2*base - 6; py < 2*base; py++ {
		for xx := x - 1; xx < x+4; xx++ {
			dx := (float64(xx) + 0.5 - (float64(x) + 1.5)) / 2.2
			dy := (float64(py) + 0.5 - cy) / 2.6
			if dx*dx+dy*dy <= 1 {
				sl := SLeaf
				if dx > 0.25 || dy > 0.4 {
					sl = SLeafDark
				}
				s.pset(xx, py, sl)
			}
		}
	}
	s.pset(x+1, 2*base, STrunk)
	s.pset(x+1, 2*base+1, STrunk)
}

func aGrove(r *rnd, h int) *Sprite {
	w := r.rng(6, 9)
	s := newSprite(w, 5)
	for x := 0; x+2 < w; x += r.rng(2, 3) {
		tree(s, x, 4, r.rng(2, 4), r)
	}
	return s
}

func aWoodCamp(r *rnd, h int) *Sprite {
	s := newSprite(9, 5)
	tree(s, 0, 4, 3, r)
	tree(s, 6, 4, 4, r)
	s.text(3, 3, "▄▄▄", STrunk, SNone)
	s.text(3, 4, "◘◘◘", STrunk, SWallDark)
	return s
}

func aStones(r *rnd, h int) *Sprite {
	s := newSprite(7, 3)
	s.text(0, 0, "▄▀▀▀▀▀▄", SRock, SNone)
	s.put(0, 1, '█', SRock, SNone)
	s.put(6, 1, '█', SRockDark, SNone)
	s.put(0, 2, '█', SRock, SNone)
	s.put(6, 2, '█', SRockDark, SNone)
	s.put(3, 2, '▄', SRock, SNone)
	s.put(3, 1, '▲', SGlow, SNone)
	return s
}

func aFirepit(r *rnd, h int) *Sprite {
	// a story circle: people round a fire, logs as benches
	s := newSprite(9, 3)
	s.put(1, 2, '☻', SWallDark, SNone)
	s.put(3, 2, '☻', SWallDark, SNone)
	s.put(5, 2, '☻', SWallDark, SNone)
	s.put(7, 2, '☻', SWallDark, SNone)
	s.put(4, 2, '▄', SGlow, SNone)
	s.put(2, 2, '▄', STrunk, SNone)
	s.put(6, 2, '▄', STrunk, SNone)
	s.Flame = append(s.Flame, pt{4, 1})
	s.Smoke = append(s.Smoke, pt{4, 0})
	return s
}

func aPile(r *rnd, h int) *Sprite {
	s := newSprite(4, 2)
	s.text(0, 0, " ▄▄ ", SRoof, SNone)
	s.text(0, 1, "▐▓▓▌", SRoof, SRoofShade)
	return s
}

func aLonghouse(r *rnd, h int) *Sprite {
	w := r.rng(9, 12)
	s := newSprite(w, 4)
	s.hline(2, w-3, 0, '▄', SRoof, SNone)
	s.hline(1, w-2, 1, '█', SRoof, SNone)
	s.put(0, 1, '▄', SRoof, SNone)
	s.put(w-1, 1, '▄', SRoofShade, SNone)
	s.hline(w/2+1, w-2, 1, '█', SRoofShade, SNone)
	s.hline(0, w-1, 2, '▀', SRoofShade, SWall)
	s.hline(0, w/2, 2, '▀', SRoof, SWall)
	s.body(0, 3, w, 1)
	s.put(w/2, 3, '▄', SGlow, SWallDark)
	s.put(2, 3, '║', SWallDark, SWall)
	s.put(w-3, 3, '║', SWallDark, SWallShade)
	s.Smoke = append(s.Smoke, pt{w/2 - 2, 0})
	return s
}

// ---------------------------------------------------------------- bronze .. medieval

func aMudHouse(r *rnd, h int) *Sprite {
	w := r.rng(5, 8)
	hh := max(3, min(h, r.rng(3, 5)))
	s := newSprite(w+3, hh+1)
	s.body(0, 1, w, hh)
	s.hline(0, w-1, 0, '▄', STrim, SNone)
	s.windows(1, 2, w-2, hh-1, 3, 2, 0, 0, '■')
	s.put(w/2, hh, '▄', SWallDark, SWall)
	// an annexe
	aw := 3
	s.body(w, 3, aw, hh-2)
	s.hline(w, w+aw-1, 2, '▄', STrim, SNone)
	if r.n(2) == 0 {
		s.put(1, 0, '╥', STrunk, SNone) // a ladder to the roof
	}
	return s
}

func aZiggurat(r *rnd, h int) *Sprite {
	tiers := max(3, min(5, h/2))
	w := tiers*4 + 1
	s := newSprite(w, tiers*2+2)
	base := s.H - 1
	for t := 0; t < tiers; t++ {
		x0 := t * 2
		tw := w - t*4
		y := base - t*2
		s.body(x0, y-1, tw, 2)
		s.hline(x0, x0+tw-1, y-1, '▀', STrim, SWall)
	}
	cx := w / 2
	for y := base; y > base-tiers*2; y-- {
		s.put(cx, y, '▒', STrim, SWallShade)
	}
	top := base - tiers*2
	s.text(cx-1, top, "▄█▄", STrim, SNone)
	s.put(cx, top, '▄', SGlow, SNone)
	s.Flame = append(s.Flame, pt{cx, top - 1})
	return s
}

func aClassical(r *rnd, h int) *Sprite {
	cols := r.rng(3, 6)
	w := cols*2 + 3
	colH := max(2, min(h-3, 3))
	s := newSprite(w, colH+5)
	// a shallow pediment in dressed stone
	top := 2
	s.roof(0, w, top, 2.6)
	for y := 0; y <= top; y++ {
		for x := 0; x < w; x++ {
			c := s.at(x, y)
			if c.Ch == 0 {
				continue
			}
			fg, bg := STrim, c.Bg
			if bg != SNone {
				bg = STrim
			}
			if x > w*2/3 {
				fg = SWallShade
			}
			s.put(x, y, c.Ch, fg, bg)
		}
	}
	s.hline(0, w-1, top+1, '▀', STrim, SWallDark)
	for y := top + 2; y < top+2+colH; y++ {
		for x := 0; x < w; x++ {
			if x%2 == 1 {
				s.put(x, y, '█', STrim, SNone)
			} else {
				s.put(x, y, ' ', SWallDark, SWallDark)
			}
		}
	}
	s.hline(0, w-1, s.H-1, '▀', STrim, SWallShade)
	s.put(0, s.H-1, '▄', STrim, SNone)
	s.put(w-1, s.H-1, '▄', STrim, SNone)
	return s.trimTop()
}

func aVilla(r *rnd, h int) *Sprite {
	w := r.rng(6, 9) | 1
	floors := max(1, min(3, (h-2)/2))
	bh := floors * 2
	s := newSprite(w+1, bh+w/2+2)
	base := s.H - 1
	s.body(0, base-bh+1, w, bh)
	s.windows(1, base-bh+1, w-2, bh, 2, 2, 0, 0, '■')
	s.put(w/2, base, '▄', SWallDark, SWall)
	s.gable(0, w, base-bh, false)
	if r.n(2) == 0 {
		s.put(w-2, base-bh-2, '▐', SWallDark, SNone)
		s.Smoke = append(s.Smoke, pt{w - 2, base - bh - 3})
	}
	return s.trimTop()
}

func aTimber(r *rnd, h int) *Sprite {
	w := r.rng(5, 7)
	floors := max(2, min(4, h/2))
	s := newSprite(w, floors+w)
	base := s.H - 1
	for f := 0; f < floors; f++ {
		y := base - f
		ww := w
		x0 := 0
		if f == 0 {
			x0, ww = 1, w-2 // jettied upper floors overhang the street
		}
		s.body(x0, y, ww, 1)
		for x := x0; x < x0+ww; x++ {
			if (x-x0)%2 == 0 {
				s.put(x, y, '║', STrim, SWall)
			} else if f > 0 {
				s.put(x, y, '■', SWin, SWall)
			}
		}
	}
	s.put(w/2, base, '▄', SWallDark, SWall)
	s.gable(0, w, base-floors, true)
	return s.trimTop()
}

func aTower(r *rnd, h int) *Sprite {
	w := r.rng(3, 5)
	th := max(4, h)
	s := newSprite(w+2, th+2)
	base := s.H - 1
	s.body(1, base-th+2, w, th-1)
	s.crenels(0, w+2, base-th+1, SWall)
	for y := base - th + 3; y < base; y += 2 {
		s.put(1+w/2, y, '▌', SWin, SWall)
	}
	s.put(1+w/2, base, '▄', SWallDark, SWall)
	if r.n(2) == 0 {
		s.put(1+w/2, base-th, '►', SNeon2, SNone)
		s.put(1+w/2, base-th+1-1, '►', SNeon2, SNone)
	}
	return s.trimTop()
}

func aCastle(r *rnd, h int) *Sprite {
	th := max(5, h)
	w := 15
	s := newSprite(w, th+1)
	base := s.H - 1
	lt := aTower(newRnd(int(r.u())), th)
	s.blit(lt, 0, s.H-lt.H)
	rt := aTower(newRnd(int(r.u())), th-1)
	s.blit(rt, w-rt.W, s.H-rt.H)
	wall := th / 2
	s.body(lt.W, base-wall+1, w-lt.W-rt.W, wall)
	s.crenels(lt.W, w-lt.W-rt.W, base-wall, SWall)
	gx := w / 2
	s.put(gx, base, '█', SWallDark, SNone)
	s.put(gx, base-1, '▄', SWallDark, SWall)
	return s
}

func aCathedral(r *rnd, h int) *Sprite {
	th := max(8, h)
	w := 11
	s := newSprite(w, th+1)
	base := s.H - 1
	nave := max(4, th/2)
	s.body(0, base-nave+1, w, nave)
	s.gable(0, w, base-nave, false)
	// the spire tower, left of centre
	tx := 2
	tw := 3
	towerTop := base - th + 3
	s.body(tx, towerTop, tw, base-towerTop)
	s.put(tx+1, towerTop-1, '█', SRoof, SNone)
	s.put(tx, towerTop-1, '▄', SRoof, SNone)
	s.put(tx+2, towerTop-1, '▄', SRoofShade, SNone)
	s.put(tx+1, towerTop-2, '█', SRoof, SNone)
	s.put(tx+1, towerTop-3, '▲', SRoof, SNone)
	s.put(tx+1, towerTop-4, '┼', STrim, SNone)
	for y := towerTop + 1; y < base-1; y += 2 {
		s.put(tx+1, y, '▌', SWin, SWall)
	}
	// rose window and lancets
	s.put(7, base-nave+1, 'o', SGlow, SWall)
	for x := 6; x <= 9; x++ {
		if x != 7 || nave > 4 {
			s.put(x, base-1, '▌', SWin, SWall)
		}
	}
	s.put(7, base, '▲', SWallDark, SWall)
	return s.trimTop()
}

func aDomed(r *rnd, h int) *Sprite {
	w := r.rng(9, 13) | 1
	bh := max(3, min(h-4, 5))
	s := newSprite(w, bh+8)
	base := s.H - 1
	s.body(0, base-bh+1, w, bh)
	for x := 1; x < w-1; x += 2 {
		s.put(x, base-1, '█', STrim, SWallDark)
		if bh > 3 {
			s.put(x, base-2, '█', STrim, SWallDark)
		}
	}
	s.hline(0, w-1, base-bh+1, '▀', STrim, SWall)
	s.windows(1, base-bh+2, w-2, max(0, bh-4), 2, 1, 0, 0, '■')
	// drum and dome
	dw := w - 4
	dy := base - bh
	s.body(2, dy, dw, 1)
	s.windows(2, dy, dw, 1, 2, 1, 1, 0, '▌')
	dr := s.dome(2, dw, dy-1, SMetal)
	s.put(w/2, dy-1-dr, '▲', STrim, SNone)
	return s.trimTop()
}

func aWindmill(r *rnd, h int) *Sprite {
	s := newSprite(7, 7)
	s.text(2, 1, "▄█▄", SRoof, SNone)
	s.put(3, 1, '█', SRoof, SNone)
	for y := 2; y < 7; y++ {
		s.put(2, y, '█', SWall, SNone)
		s.put(3, y, '█', SWall, SNone)
		s.put(4, y, '█', SWallShade, SNone)
		if y >= 4 {
			s.put(1, y, '▐', SWall, SNone)
			s.put(5, y, '▌', SWallShade, SNone)
		}
	}
	s.put(3, 6, '▄', SWallDark, SWall)
	s.put(3, 4, '■', SWin, SWall)
	s.Blades = append(s.Blades, pt{3, 2})
	return s
}

func aFields(r *rnd, h int) *Sprite {
	w := r.rng(10, 15)
	s := newSprite(w, 4)
	for x := 0; x < w; x++ {
		ch := '▄'
		slot := SField1
		if (x/3)%2 == 1 {
			slot = SField2
		}
		s.put(x, 3, '▓', slot, SLeafDark)
		if x%2 == 0 {
			s.put(x, 3, '▒', slot, SLeafDark)
		}
		_ = ch
	}
	// a barn
	b := r.n(w - 5)
	s.body(b, 2, 4, 1)
	s.put(b+1, 2, '▄', SWallDark, SWall)
	s.hline(b, b+3, 1, '▄', SRoof, SNone)
	s.put(b+2, 1, '▄', SRoofShade, SNone)
	s.put(b+3, 1, '▄', SRoofShade, SNone)
	if r.n(2) == 0 {
		s.put(b+5, 2, '▄', SMetal, SNone) // a silo
		s.put(b+5, 1, '▄', SMetal, SNone)
	}
	return s
}

func aSmithy(r *rnd, h int) *Sprite {
	w := r.rng(6, 8)
	s := newSprite(w, 6)
	base := 5
	s.body(0, 3, w, 3)
	s.hline(0, w-1, 2, '▄', SRoof, SNone)
	s.hline(w/2, w-1, 2, '▄', SRoofShade, SNone)
	s.put(1, 4, '▄', SGlow, SWallDark)
	s.put(2, 4, '▄', SGlow, SWallDark)
	s.put(1, base, '█', SWallDark, SNone)
	s.put(2, base, '█', SWallDark, SNone)
	s.put(w-2, 1, '█', SWallDark, SNone)
	s.put(w-2, 2, '█', SWallDark, SNone)
	s.Smoke = append(s.Smoke, pt{w - 2, 0})
	s.windows(4, 4, w-5, 1, 2, 1, 0, 0, '■')
	return s
}

func aMarket(r *rnd, h int) *Sprite {
	stalls := r.rng(2, 4)
	w := stalls * 4
	s := newSprite(w, 3)
	for i := 0; i < stalls; i++ {
		x := i * 4
		c := []Slot{SRoof, STrim, SNeon1, SNeon2}[r.n(4)]
		s.text(x, 0, "▄▄▄", c, SNone)
		s.text(x, 1, "▀▀▀", c, SNone)
		s.put(x+1, 1, '▀', STrim, SNone)
		s.put(x, 2, '│', STrunk, SNone)
		s.put(x+2, 2, '│', STrunk, SNone)
		s.put(x+1, 2, r.pick("◘▪•o"), SField2, SNone)
	}
	return s
}

// ---------------------------------------------------------------- industrial .. atomic

func aFactory(r *rnd, h int) *Sprite {
	teeth := r.rng(3, 5)
	w := teeth*3 + 1
	bh := max(3, min(h-2, 5))
	stacks := r.rng(1, 3)
	sh := bh + r.rng(3, 6)
	s := newSprite(w+2, sh+2)
	base := s.H - 1
	s.body(0, base-bh+1, w, bh)
	for t := 0; t < teeth; t++ {
		x := t * 3
		y := base - bh
		s.put(x, y, '█', SRoof, SNone)
		s.put(x+1, y, '▄', SRoofShade, SNone)
		s.put(x, y-1, '▄', SRoof, SNone)
		s.put(x+1, y, '▄', SGlass, SNone)
		s.put(x+2, y, '▄', SRoofShade, SNone)
	}
	s.windows(1, base-bh+2, w-2, bh-2, 2, 2, 0, 0, '■')
	s.put(2, base, '▄', SWallDark, SWall)
	for i := 0; i < stacks; i++ {
		x := w - 2 - i*3
		if x < 1 {
			break
		}
		top := base - sh + i
		s.vline(x, top, base-bh, '█', SWallShade, SNone)
		s.put(x, top+1, '▀', STrim, SWallShade)
		s.Smoke = append(s.Smoke, pt{x, top - 1})
	}
	return s.trimTop()
}

func aRowHouse(r *rnd, h int) *Sprite {
	units := r.rng(2, 4)
	uw := 3
	floors := max(2, min(5, h-1))
	w := units * uw
	s := newSprite(w, floors+3)
	base := s.H - 1
	s.body(0, base-floors+1, w, floors)
	s.hline(0, w-1, base-floors, '▄', STrim, SNone)
	for u := 0; u < units; u++ {
		x := u * uw
		for f := 0; f < floors; f++ {
			s.put(x+1, base-f, '■', SWin, SWall)
		}
		s.put(x+1, base, '▄', SWallDark, SWall)
		if u%2 == 0 {
			s.put(x, base-floors-1, '▐', SWallShade, SNone)
			s.put(x, base-floors, '█', SWallShade, SNone)
			if r.n(2) == 0 {
				s.Smoke = append(s.Smoke, pt{x, base - floors - 2})
			}
		}
	}
	return s.trimTop()
}

func aCooling(r *rnd, h int) *Sprite {
	th := max(6, min(h, 9))
	s := newSprite(9, th+1)
	base := s.H - 1
	for i := 0; i < th; i++ {
		y := base - i
		t := float64(i) / float64(th-1)
		// hyperboloid: wide foot, waist at 2/3, flared lip
		wd := 4 - int(3*t*(1.35-t)*1.9+0.2)
		if wd < 2 {
			wd = 2
		}
		for dx := -wd; dx <= wd; dx++ {
			slot := SWall
			if dx > wd/2 {
				slot = SWallShade
			} else if dx == -wd {
				slot = SWallLit
			}
			s.put(4+dx, y, '█', slot, SNone)
		}
	}
	s.Smoke = append(s.Smoke, pt{3, base - th}, pt{5, base - th})
	return s.trimTop()
}

func aDerrick(r *rnd, h int) *Sprite {
	th := max(5, min(h, 9))
	s := newSprite(7, th+1)
	base := s.H - 1
	for i := 0; i < th; i++ {
		y := base - i
		off := (th - i) * 3 / th
		s.put(3-off, y, '╱', SMetal, SNone)
		s.put(3+off, y, '╲', SMetal, SNone)
		if off > 0 && i%2 == 1 {
			for x := 3 - off + 1; x < 3+off; x++ {
				s.put(x, y, '─', SMetalDark, SNone)
			}
		}
	}
	s.put(3, base-th, '▲', SMetal, SNone)
	s.put(3, base-th+1, '+', SBeacon, SNone)
	s.Beacons = append(s.Beacons, pt{3, base - th})
	return s
}

func aDeco(r *rnd, h int) *Sprite {
	w := r.rng(5, 7) | 1
	th := max(8, h)
	s := newSprite(w, th+3)
	base := s.H - 1
	// three setbacks
	segs := []struct{ inset, top int }{{0, th * 6 / 10}, {1, th * 85 / 100}, {2, th}}
	for _, g := range segs {
		ww := w - 2*g.inset
		if ww < 1 {
			continue
		}
		y0 := base - g.top + 1
		s.body(g.inset, y0, ww, base-y0+1)
		s.hline(g.inset, g.inset+ww-1, y0, '▀', STrim, SWall)
	}
	for x := 1; x < w-1; x += 2 {
		for y := base - th + 2; y < base; y++ {
			if s.at(x, y).Ch == '█' {
				s.put(x, y, '▌', SWin, s.at(x, y).Fg)
			}
		}
	}
	cx := w / 2
	s.put(cx, base-th, '█', STrim, SNone)
	s.put(cx, base-th-1, '▌', STrim, SNone)
	s.put(cx, base-th-2, '│', SMetal, SNone)
	s.Beacons = append(s.Beacons, pt{cx, base - th - 2})
	s.put(cx, base, '▄', SWallDark, SWall)
	return s.trimTop()
}

func aPylon(r *rnd, h int) *Sprite {
	th := max(5, min(h, 8))
	s := newSprite(5, th+1)
	base := s.H - 1
	for i := 0; i < th; i++ {
		y := base - i
		if i < th/2 {
			s.put(1, y, '╱', SMetal, SNone)
			s.put(3, y, '╲', SMetal, SNone)
			s.put(2, y, '╳', SMetalDark, SNone)
		} else {
			s.put(2, y, '║', SMetal, SNone)
		}
	}
	s.text(0, base-th+2, "═╬═", SMetal, SNone)
	s.text(0, base-th+2, "═", SMetal, SNone)
	s.put(4, base-th+2, '═', SMetal, SNone)
	s.put(3, base-th+2, '═', SMetal, SNone)
	// a low turbine hall beside it
	return s
}

// ---------------------------------------------------------------- modern .. cosmic

func aGlass(r *rnd, h int) *Sprite {
	w := r.rng(4, 8)
	th := max(8, h)
	s := newSprite(w, th+4)
	base := s.H - 1
	top := base - th + 1
	for y := top; y <= base; y++ {
		for x := 0; x < w; x++ {
			slot := SGlass
			if x == w-1 || (w > 5 && x == w-2) {
				slot = SWallShade
			}
			if x == 0 {
				slot = SGlassHi
			}
			s.put(x, y, '▀', SWin, slot)
		}
	}
	// a mullion stripe
	if w >= 6 {
		for y := top; y <= base; y++ {
			s.put(w/2, y, '█', SWall, SNone)
		}
	}
	s.hline(0, w-1, base, '█', SWall, SNone)
	s.put(w/2, base, '▄', SGlassHi, SWallDark)
	switch r.n(4) {
	case 0: // antenna
		s.put(w/2, top-1, '│', SMetal, SNone)
		s.put(w/2, top-2, '│', SMetal, SNone)
		s.Beacons = append(s.Beacons, pt{w / 2, top - 3})
	case 1: // slanted crown
		for i := 0; i < w-1; i++ {
			s.put(i, top-1-i/2, '▄', SGlassHi, SNone)
			for y := top - i/2; y < top; y++ {
				s.put(i, y, '█', SGlass, SNone)
			}
		}
	case 2: // stepped cap
		s.hline(1, w-2, top-1, '▄', SWall, SNone)
		s.Beacons = append(s.Beacons, pt{1, top - 2})
	default:
		s.hline(0, w-1, top-1, '▄', SWall, SNone)
	}
	return s.trimTop()
}

func aDish(r *rnd, h int) *Sprite {
	s := newSprite(7, 6)
	s.text(0, 0, "▄▀▀▀▄", SMetal, SNone)
	s.text(0, 1, "▀▄▄▄▀", SMetal, SNone)
	s.put(2, 1, '▄', SMetalDark, SNone)
	s.put(2, 2, '█', SMetalDark, SNone)
	s.put(2, 3, '█', SMetalDark, SNone)
	s.body(0, 4, 6, 2)
	s.windows(1, 4, 4, 2, 2, 1, 0, 0, '■')
	s.put(4, 0, '+', SBeacon, SNone)
	return s
}

func aDataCenter(r *rnd, h int) *Sprite {
	w := r.rng(10, 14)
	s := newSprite(w, 5)
	s.body(0, 2, w, 3)
	for x := 1; x < w-1; x++ {
		if x%2 == 1 {
			s.put(x, 3, '·', SNeon2, SWallDark)
			s.put(x, 2, '·', SNeon3, SWallDark)
		} else {
			s.put(x, 3, '█', SWallDark, SNone)
			s.put(x, 2, '█', SWallDark, SNone)
		}
	}
	for x := 1; x < w-1; x += 3 {
		s.text(x, 1, "▄▄", SMetal, SNone)
	}
	s.put(w-2, 0, '+', SBeacon, SNone)
	return s
}

var neonGlyphs = "¥£ƒΣΩ∞≡♦§¤"

func aCyber(r *rnd, h int) *Sprite {
	w := r.rng(5, 9)
	th := max(12, h)
	s := newSprite(w+2, th+4)
	base := s.H - 1
	top := base - th + 1
	for y := top; y <= base; y++ {
		for x := 1; x <= w; x++ {
			slot := SWall
			if x == w {
				slot = SWallShade
			}
			s.put(x, y, '█', slot, SNone)
		}
		if (y-top)%2 == 1 {
			for x := 2; x < w; x += 1 {
				if hash(x, y, int(r.s))%3 != 0 {
					s.put(x, y, '▄', SWin, SWall)
				}
			}
		}
	}
	// neon edge and a vertical sign
	n := []Slot{SNeon1, SNeon2, SNeon3}[r.n(3)]
	for y := top; y <= base; y++ {
		s.put(1, y, '▌', n, SWall)
	}
	sx := w + 1
	sy := top + 2 + r.n(max(1, th/3))
	sl := r.rng(3, 5)
	for i := 0; i < sl; i++ {
		s.put(sx, sy+i, r.pick(neonGlyphs), SNeon1, SInk)
	}
	s.put(0, top+th/2, '▐', SNeon2, SNone)
	// rooftop billboard or antenna
	if r.n(2) == 0 {
		s.hline(1, w, top-1, '▄', SNeon2, SNone)
		s.hline(2, w-1, top-2, '▄', SMetalDark, SNone)
	} else {
		s.put(w/2+1, top-1, '│', SMetal, SNone)
		s.put(w/2+1, top-2, '│', SMetal, SNone)
		s.Beacons = append(s.Beacons, pt{w/2 + 1, top - 3})
	}
	return s.trimTop()
}

func aArcology(r *rnd, h int) *Sprite {
	th := max(12, h)
	w := th + 5
	if w%2 == 0 {
		w++
	}
	s := newSprite(w, th+3)
	base := s.H - 1
	tiers := 4
	for t := 0; t < tiers; t++ {
		inset := t * (w / 9)
		y0 := base - (t+1)*th/tiers + 1
		y1 := base - t*th/tiers
		s.body(inset, y0, w-2*inset, y1-y0+1)
		s.hline(inset, w-1-inset, y0, '▄', SLeaf, SNone)
		for y := y0 + 1; y <= y1; y++ {
			for x := inset + 1; x < w-1-inset; x += 2 {
				s.put(x, y, '▀', SWin, s.at(x, y).Fg)
			}
		}
	}
	cx := w / 2
	topY := base - th
	s.put(cx, topY, '▲', SGlassHi, SNone)
	s.put(cx, topY-1, '│', SMetal, SNone)
	s.Beacons = append(s.Beacons, pt{cx, topY - 2})
	return s.trimTop()
}

func aHabDome(r *rnd, h int) *Sprite {
	w := r.rng(7, 11) | 1
	s := newSprite(w, 5)
	s.body(0, 4, w, 1)
	s.dome(0, w, 3, SGlass)
	for x := 2; x < w-2; x += 2 {
		s.put(x, 3, '▪', SNeon3, SGlass)
	}
	s.hline(0, w-1, 4, '▀', STrim, SWall)
	return s.trimTop()
}

func aFusion(r *rnd, h int) *Sprite {
	s := newSprite(11, 7)
	s.body(0, 4, 11, 3)
	s.dome(1, 9, 3, SMetal)
	s.hline(2, 8, 2, '═', SGlow, SMetal)
	s.windows(1, 5, 9, 2, 2, 1, 0, 0, '▀')
	s.put(10, 1, '+', SBeacon, SNone)
	return s.trimTop()
}

func aLaunch(r *rnd, h int) *Sprite {
	th := max(10, min(h, 16))
	s := newSprite(9, th+2)
	base := s.H - 1
	// gantry
	for y := base - th + 1; y <= base; y++ {
		s.put(1, y, '╫', SMetal, SNone)
	}
	for y := base - th + 3; y <= base; y += 3 {
		s.put(2, y, '─', SMetal, SNone)
	}
	// the rocket
	rt := base - th + 2
	s.put(4, rt, '▲', STrim, SNone)
	for y := rt + 1; y < base; y++ {
		s.put(4, y, '█', STrim, SNone)
		s.put(3, y, '▐', SWallShade, SNone)
		s.put(5, y, '▌', SWallShade, SNone)
	}
	s.put(4, rt+3, '▀', SNeon2, STrim)
	s.put(3, base-1, '▟', SWallShade, SNone)
	s.put(3, base-1, '▄', SWallShade, SNone)
	s.put(5, base-1, '▄', SWallShade, SNone)
	s.text(0, base, "▀▀▀▀▀▀▀▀▀", SMetalDark, SNone)
	s.Smoke = append(s.Smoke, pt{6, base - 1})
	return s
}

func aSpire(r *rnd, h int) *Sprite {
	th := max(14, h)
	s := newSprite(7, th+3)
	base := s.H - 1
	top := base - th + 1
	for y := top; y <= base; y++ {
		t := float64(y-top) / float64(th)
		wd := int(t*t*3 + 0.3)
		for dx := -wd; dx <= wd; dx++ {
			slot := SWall
			if dx > 0 {
				slot = SWallShade
			}
			if dx < 0 && dx == -wd {
				slot = SWallLit
			}
			s.put(3+dx, y, '█', slot, SNone)
		}
		if wd == 0 {
			s.put(3, y, '▐', SWall, SNone)
		}
	}
	for y := top + 3; y < base-2; y += max(3, th/4) {
		s.put(2, y, '═', SNeon1, SNone)
		s.put(3, y, '═', SNeon1, SNone)
		s.put(4, y, '═', SNeon1, SNone)
	}
	switch r.n(3) {
	case 0: // a glowing orb carried on the needle
		oy := top + th/5
		s.text(2, oy, "▄█▄", SGlow, SNone)
		s.text(2, oy+1, "▀█▀", SGlow, SNone)
		s.put(3, oy+1, '█', SWall, SNone)
	case 1: // a twin needle, shorter, on the shaded side
		for y := top + th/3; y < base-2; y++ {
			if s.at(5, y).Ch == 0 {
				s.put(5, y, '▐', SWallShade, SNone)
			}
		}
		s.put(5, top+th/3-1, '•', SNeon3, SNone)
	}
	s.put(3, top-1, '•', SNeon2, SNone)
	s.put(3, top-2, '·', SNeon2, SNone)
	s.put(3, base, '▄', SGlow, SWallDark)
	return s.trimTop()
}

// aSaucer: a habitat disc held up on a stem, the late city's houses.
func aSaucer(r *rnd, h int) *Sprite {
	th := max(6, h)
	w := r.rng(7, 11) | 1
	s := newSprite(w, th+3)
	base := s.H - 1
	cx := w / 2
	stem := base - th*45/100 + 1
	for y := stem; y <= base; y++ {
		s.put(cx, y, '█', SWallShade, SNone)
		if y == base {
			s.put(cx-1, y, '▄', SWallShade, SNone)
			s.put(cx+1, y, '▄', SWallShade, SNone)
		}
	}
	// the disc: a lit rim, a band of windows, a glass dome on top
	dy := stem - 1
	for x := 0; x < w; x++ {
		s.put(x, dy, '█', SWall, SNone)
		s.put(x, dy+1, '▀', SWallShade, SNone)
	}
	s.put(0, dy, '▄', SWall, SNone)
	s.put(w-1, dy, '▄', SWall, SNone)
	for x := 1; x < w-1; x += 2 {
		s.put(x, dy, '▪', SWin, SWall)
	}
	s.dome(2, w-4, dy-1, SGlass)
	s.hline(1, w-2, dy+1, '▀', SNeon1, SNone)
	if r.n(2) == 0 {
		s.Beacons = append(s.Beacons, pt{cx, dy - 3})
	}
	return s.trimTop()
}

func aCrane(r *rnd, h int) *Sprite {
	s := newSprite(11, 8)
	for y := 1; y < 8; y++ {
		s.put(2, y, '╫', SNeon2, SNone)
	}
	s.hline(0, 10, 1, '═', SNeon2, SNone)
	s.put(8, 2, '│', SMetalDark, SNone)
	s.put(8, 3, '▄', SMetal, SNone)
	// containers
	cols := []Slot{SRoof, STrim, SNeon1, SMetal, SGlass}
	for i := 0; i < 6; i++ {
		x := 4 + (i%3)*2
		y := 7 - i/3
		s.put(x, y, '█', cols[r.n(len(cols))], SNone)
		s.put(x+1, y, '▌', SWallDark, SNone)
	}
	s.put(2, 0, '+', SBeacon, SNone)
	return s
}

func aSilo(r *rnd, h int) *Sprite {
	n := r.rng(2, 3)
	th := max(3, min(h, 6))
	s := newSprite(n*3, th+1)
	base := s.H - 1
	for i := 0; i < n; i++ {
		x := i * 3
		s.vline(x, base-th+1, base, '█', SWallLit, SNone)
		s.vline(x+1, base-th+1, base, '█', SWallShade, SNone)
		s.put(x, base-th, '▄', SRoof, SNone)
		s.put(x+1, base-th, '▄', SRoofShade, SNone)
		s.put(x, base-th/2, '▀', STrim, SWallLit)
		s.put(x+1, base-th/2, '▀', STrim, SWallShade)
	}
	return s
}

func aWarehouse(r *rnd, h int) *Sprite {
	w := r.rng(7, 10)
	s := newSprite(w, 4)
	s.body(0, 1, w, 3)
	s.hline(0, w-1, 0, '▄', SRoof, SNone)
	s.hline(w/2, w-1, 0, '▄', SRoofShade, SNone)
	s.put(2, 3, '█', SWallDark, SNone)
	s.put(3, 3, '█', SWallDark, SNone)
	s.put(2, 2, '▄', SWallDark, SWall)
	s.put(3, 2, '▄', SWallDark, SWall)
	s.windows(5, 2, w-6, 1, 2, 1, 0, 0, '■')
	return s
}

func aVault(r *rnd, h int) *Sprite {
	w := r.rng(6, 8)
	s := newSprite(w, 4)
	s.body(0, 1, w, 3)
	s.hline(0, w-1, 0, '▄', STrim, SNone)
	s.hline(0, w-1, 1, '▀', STrim, SWall)
	s.put(w/2, 2, 'o', SMetal, SWallDark)
	s.put(w/2, 3, '▀', SWallDark, SWall)
	return s
}

func aBunker(r *rnd, h int) *Sprite {
	w := r.rng(8, 11)
	s := newSprite(w, 6)
	s.body(0, 3, w, 3)
	s.hline(0, w-1, 2, '▄', SWall, SNone)
	s.put(0, 2, ' ', SNone, SNone)
	for x := 1; x < w-1; x += 3 {
		s.put(x, 4, '▬', SWallDark, SWall)
	}
	d := aDishSmall()
	s.blit(d, w-4, 0)
	s.put(1, 1, '►', SBeacon, SNone)
	s.put(1, 2, '│', SMetal, SNone)
	return s
}

func aDishSmall() *Sprite {
	s := newSprite(4, 3)
	s.text(0, 0, "▀▄▄▀", SMetal, SNone)
	s.put(1, 1, '▐', SMetalDark, SNone)
	s.put(1, 2, '█', SMetalDark, SNone)
	return s
}

func aSolar(r *rnd, h int) *Sprite {
	w := r.rng(8, 12)
	s := newSprite(w, 2)
	for x := 0; x < w; x++ {
		if x%4 != 3 {
			s.put(x, 0, '▄', SGlass, SNone)
			s.put(x, 1, '▀', SGlassHi, SNone)
		} else {
			s.put(x, 1, '│', SMetal, SNone)
		}
	}
	return s
}

func aObelisk(r *rnd, h int) *Sprite {
	th := max(5, min(h, 9))
	s := newSprite(3, th+1)
	base := s.H - 1
	s.put(1, base-th+1, '▲', STrim, SNone)
	for y := base - th + 2; y < base; y++ {
		s.put(1, y, '█', STrim, SNone)
	}
	s.text(0, base, "▄█▄", STrim, SNone)
	return s
}

func aVertFarm(r *rnd, h int) *Sprite {
	w := r.rng(5, 7)
	th := max(8, h)
	s := newSprite(w, th+1)
	base := s.H - 1
	for y := base - th + 1; y <= base; y++ {
		for x := 0; x < w; x++ {
			slot := SGlass
			if x == w-1 {
				slot = SWallShade
			}
			s.put(x, y, '█', slot, SNone)
		}
		if (base-y)%2 == 0 {
			for x := 0; x < w-1; x++ {
				s.put(x, y, '▄', SLeaf, SGlass)
			}
		}
	}
	s.hline(0, w-1, base-th, '▄', SMetal, SNone)
	return s
}

// trimTop drops empty rows above the art, so heights are honest.
func (s *Sprite) trimTop() *Sprite {
	top := 0
	for top < s.H-1 {
		empty := true
		for x := 0; x < s.W; x++ {
			if s.C[top*s.W+x].Ch != 0 {
				empty = false
				break
			}
		}
		for _, p := range append(append(append([]pt{}, s.Smoke...), s.Beacons...), s.Flame...) {
			if p.Y == top {
				empty = false
			}
		}
		if !empty {
			break
		}
		top++
	}
	if top == 0 {
		return s
	}
	o := newSprite(s.W, s.H-top)
	copy(o.C, s.C[top*s.W:])
	shift := func(ps []pt) []pt {
		var out []pt
		for _, p := range ps {
			out = append(out, pt{p.X, p.Y - top})
		}
		return out
	}
	o.Smoke, o.Beacons, o.Blades, o.Flame = shift(s.Smoke), shift(s.Beacons), shift(s.Blades), shift(s.Flame)
	return o
}
