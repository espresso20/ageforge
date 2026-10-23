package skyline

import (
	"math"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
)

// fit.go is the far zoom: the whole town fitted to the screen's width,
// whatever its size. The world frame is the wide level's (twice the screen's
// columns, twice its rows, shrunk 2x2 into the screen), but a column of it
// stands for kx world columns, kx being the town's width over the frame's.
// Only the town and its ground are read that way: the sky, the ridge and the
// weather are composed in the frame as they are at any other level, and
// everything on the screen afterwards (markers, bars, the cursor) is placed
// by the world column through wsx.
//
// The town is the skyline of what stands tallest on each run of world
// columns, in the colours of that building's own cells, so a district reads
// by its height and its age's materials, and at night by the windows lit in
// it. The present age's wonder is always drawn, at its own place and
// whatever stands beside it.

// fitOK reports whether the far level is a level of its own for a screen sw
// wide: the town is wider than the wide level shows.
func (v *view) fitOK(sw int) bool { return sw > 0 && v.worldW > 2*sw }

// fitting reports whether the view is at the far level and showing the whole
// town.
func (v *view) fitting() bool { return v.zoom == zFar && v.fitOK(v.sw) }

// camX is the world column of the view's left edge: the whole town is in
// view at the far level.
func (v *view) camX() int {
	if v.fitting() {
		return 0
	}
	return v.cam
}

// frameW is the width of the world frame the last frame was composed in: the
// view's world columns at every level but far, where the town is squeezed
// into twice the screen's.
func (v *view) frameW() int {
	if v.fitting() {
		return 2 * v.sw
	}
	return v.viewW()
}

// fx is the world column of the frame column x (the camera at the town's
// west edge) at the far level: the world columns of a far frame, which the
// cursor maths use.
func (v *view) fx(x int) int {
	if v.fitting() && v.sw > 0 {
		return int(float64(x) * float64(v.worldW) / float64(2*v.sw))
	}
	return x
}

// fx is the frame column of the world column wx (relative to the camera):
// the same column at every level but far.
func (s *scene) fx(wx int) int {
	if !s.fit {
		return wx
	}
	return int(float64(wx) / s.kx)
}

// wsx is the screen column of the world column wx, relative to the camera.
func (s *scene) wsx(wx int) int { return s.sx(s.fx(wx)) }

// fitCol is what stands tallest on the world columns one frame column
// stands for: the lot, the world column to read its sprite at, and the
// height there.
type fitCol struct {
	lot int32 // -1 for none
	at  int32
	h   int32
}

// fitCols reads the town as n columns, cached on the layout.
func (l *layout) fitCols(n int) []fitCol {
	if l.fitN == n && l.fitC != nil {
		return l.fitC
	}
	kx := float64(l.width) / float64(n)
	out := make([]fitCol, n)
	for i := range out {
		a := int(float64(i) * kx)
		b := max(a+1, int(float64(i+1)*kx))
		best := fitCol{lot: -1}
		for w := a; w < b && w < len(l.prof); w++ {
			if h := l.prof[w]; h > int(best.h) && l.tallest[w] >= 0 {
				best = fitCol{lot: l.tallest[w], at: int32(w), h: int32(h)}
			}
		}
		out[i] = best
	}
	l.fitN, l.fitC = n, out
	return out
}

// fitTown draws the town into the far frame.
func (s *scene) fitTown() {
	p := s.p
	lay := s.lay
	dark := math.Max(p.night, p.twilight*0.8)
	mega := s.megacity()
	s.vis = lay.visible(0, lay.width+lay.maxW, s.vis)
	paint := func(x int, lv *lotView, sx int) {
		spr := lv.spr
		y0 := s.groundY - spr.h
		hz := lv.ml.Row
		tint := s.districtTint(lv.ml.Age, hz)
		gold := lv.ml.Wonder && dark > 0.2
		for sy := 0; sy < spr.h; sy++ {
			c := spr.c[sy*spr.w+sx]
			if c.ch == 0 {
				continue
			}
			fg := p.col(lv.fam, lv.vr, c.fg, hz)
			if tint != 0 && !c.fg.emissive() {
				fg = theme.Mix(fg, tint, 0.32)
			}
			if mega && (c.fg.emissive() || c.fg == sGlass || c.fg == sGlassHi) {
				fg = s.neonFix(fg, hz)
			}
			if c.fg == sWin && dark > 0.25 && s.windowLit(lv, sx, sy) {
				lit := p.hue(windowHue(lv.ml.Age, hash(int(lv.ml.Seed), sx, sy)), mEmit, hz)
				fg = theme.Mix(fg, lit, math.Min(1, (dark-0.25)*2))
			}
			if gold { // a landmark glows at night
				fg = theme.Mix(fg, s.wonderGlow(lv, hz), 0.5*clamp01((dark-0.2)*1.6))
			}
			bg := fg
			if c.bg != sNone {
				bg = p.col(lv.fam, lv.vr, c.bg, hz)
			}
			s.fb.set(x, s.Y(y0+sy), '█', fg, bg, lv.depth)
		}
	}
	for x, fc := range lay.fitCols(s.W) {
		if fc.lot < 0 {
			continue
		}
		lv := &lay.lots[fc.lot]
		paint(x, lv, int(fc.at)-lv.x0)
	}
	// this age's wonder, on its own column however low it stands, and the
	// scaffold of one going up
	one := func(lv *lotView) {
		x := s.fx(lv.ml.X - s.cam)
		if x < 0 || x >= s.W {
			return
		}
		sx := lv.spr.w / 2
		paint(x, lv, sx)
		if lv.spr.w >= 3 { // a column either side keeps it from a single needle
			paint(x-1, lv, sx-lv.spr.w/4)
			paint(x+1, lv, sx+lv.spr.w/4)
		}
	}
	for i := range lay.lots {
		if lv := &lay.lots[i]; lv.ml.Wonder {
			one(lv)
		}
	}
	for i := range lay.sites {
		one(&lay.sites[i])
	}
}

// fitLots draws the sky scene's forms into the far frame: the same read of
// the town as fitTown, in the scene's own inks.
func (o *orb) fitLots() {
	lay := o.lay
	o.vis = lay.visible(0, lay.width+lay.maxW, o.vis)
	paint := func(x int, lv *lotView, sx int) {
		spr := lv.spr
		y0 := o.groundY - spr.h
		hz := lv.ml.Row
		warm := o.c(mapmodel.InkLight, iEmit, hz)
		for sy := 0; sy < spr.h; sy++ {
			c := spr.c[sy*spr.w+sx]
			if c.ch == 0 {
				continue
			}
			fg := o.lotCol(lv, c.fg, hz)
			if hz > 0 && !c.fg.emissive() {
				fg = theme.Shade(fg, rowDim[min(hz, 2)])
			}
			if c.fg == sWin && o.windowLit(lv, sx, sy) {
				fg = warm
			}
			bg := fg
			if c.bg != sNone {
				bg = o.lotCol(lv, c.bg, hz)
			}
			o.fb.set(x, o.Y(y0+sy), '█', fg, bg, lv.depth)
		}
	}
	for x, fc := range lay.fitCols(o.W) {
		if fc.lot >= 0 {
			lv := &lay.lots[fc.lot]
			paint(x, lv, int(fc.at)-lv.x0)
		}
	}
	for i := range lay.lots { // this age's wonder, however low it stands
		if lv := &lay.lots[i]; lv.ml.Wonder {
			x := o.fx(lv.ml.X - o.cam)
			if x >= 0 && x < o.W {
				paint(x, lv, lv.spr.w/2)
			}
		}
	}
}

// fitGround draws the ground under the far frame: grass, cobble or slab by
// the age of the district a column stands in, the road, and the bay.
func (s *scene) fitGround() {
	a := s.m.AgeIdx
	p := s.p
	gy := s.groundY
	rh, mh := roadHues(a)
	rc := p.hue(rh, mLit, 0)
	mc := p.hue(mh, mLit, 0)
	if a > 14 {
		mc = p.hue(mh, mEmit, 0)
	}
	paved := s.inCity && mapmodel.Greenery(a) <= 0.15
	di := -1
	for x := 0; x < s.W; x++ {
		wx := int((float64(x) + 0.5) * s.kx)
		if di < 0 || di >= len(s.m.Skyline.Districts) || !inDistrict(s.m.Skyline.Districts[di], wx) {
			di = s.m.Skyline.DistrictAt(wx)
		}
		dAge := -1
		if di >= 0 {
			d := s.m.Skyline.Districts[di]
			dAge = d.Age
			if d.Bay && wx >= d.BayX && wx < d.BayX+d.BayW {
				w := p.final(theme.Tint(theme.SkyColor(theme.SkyWater), p.lightTint))
				for y := gy; y < s.S; y++ {
					s.fb.fill(x, s.Y(y), w, dGround)
				}
				continue
			}
		}
		top, under := theme.SkyGroundGrass, theme.SkyGroundSoil
		switch {
		case dAge > 11 || paved:
			top, under = theme.SkyGroundSlab, theme.SkyGroundSlabDark
		case dAge > 7:
			top, under = theme.SkyGroundBrick, theme.SkyGroundBrickDark
		case dAge > 3:
			top, under = theme.SkyGroundCobble, theme.SkyGroundCobbleDark
		}
		s.fb.set(x, s.Y(gy), '▀', p.hue(top, mLit, 0), p.hue(under, mLit, 0), dGround)
		s.fb.set(x, s.Y(gy+1), ' ', mc, rc, dGround)
		vg, vb := theme.SkyVerge, theme.SkyVergeDark
		if a >= 12 && dAge >= 0 {
			vg, vb = theme.SkyVergeUrban, theme.SkyVergeUrbanDark
		}
		s.fb.set(x, s.Y(gy+2), ' ', p.hue(vg, mLit, 0), p.hue(vb, mLit, 0), dGround)
		for y := gy + 3; y < s.S; y++ {
			s.fb.fill(x, s.Y(y), p.hue(vb, mLit, 0), dGround)
		}
	}
}

// districtNames writes each district's name under the far view, off rows
// below the ground or the baseline, where it fits between its neighbours.
func (s *scene) districtNames(off int) {
	y := s.Y(s.sy(s.groundY + off))
	cbg := chromeBg()
	for _, d := range s.m.Skyline.Districts {
		a, b := s.wsx(d.X0), s.wsx(d.X0+d.W)
		name := s.m.Catalog.AgeNames[d.Age]
		for _, t := range []string{name, ageWord(name)} {
			n := textLen(t)
			if b-a < n+2 {
				continue
			}
			x := a + (b-a-n)/2
			for i, r := range []rune(t) {
				s.fb.set(x+i, y, r, theme.Legible(theme.Color(theme.RoleText), cbg, 4.5), cbg, dTag)
			}
			break
		}
	}
}

// farFlags marks, on the far view, the present age's wonder with a gold
// star above its column (the lit rim is too fine to find at this scale).
func (s *scene) farFlags() {
	gold := s.p.hue(theme.SkyWonderGold, mEmit, 0)
	sym := mapmodel.R(mapmodel.SymWonder, s.tier)
	mark := func(lv *lotView) {
		x := s.wsx(lv.ml.X - s.cam)
		y := max(0, s.sy(s.groundY-lv.spr.h-2))
		s.fb.fg(x, s.Y(y), sym, gold, dTag)
	}
	for i := range s.lay.lots {
		if lv := &s.lay.lots[i]; lv.ml.Wonder {
			mark(lv)
		}
	}
	for i := range s.lay.sites {
		mark(&s.lay.sites[i])
	}
}

// ridgeItemsFar places the civs met and the harbinger across a far frame
// w wide, evenly: the ridge is not scrolled there.
func ridgeItemsFar(m *mapmodel.Model, w int) []ridgeItem {
	return ridgeItemsSpan(m, w, 0, w)
}
