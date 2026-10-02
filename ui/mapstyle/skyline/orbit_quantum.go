package skyline

import (
	"math"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
)

// orbit_quantum.go is the Quantum Age: reality bends. Over a near-black
// void everything shifts through an iridescent ramp, cell by cell and frame
// by frame. Every lot is in superposition, flickering between its sky form
// and the building it was on the ground, in a wave that sweeps along the
// panorama; probability clouds shimmer round them in interference fringes;
// behind it all the old town stands as a faint time echo on its ground
// line; and at the present a starbase-like landmark repeats smaller and
// smaller inside itself. Phase ships tunnel through in bursts.

// iriSlot is how each slot runs through the iridescent ramp: its offset
// along the ramp, its shade and how it is lit.
var iriSlot = [numSlots]struct {
	off  int
	k    float64
	mode int
}{
	sWall: {0, 0.42, iLit}, sWallLit: {0, 0.8, iLit}, sWallShade: {0, 0.27, iLit}, sWallDark: {0, 0.17, iLit},
	sRoof: {0, 0.52, iLit}, sRoofShade: {0, 0.3, iLit}, sTrim: {1, 0.66, iLit}, sWin: {0, 0.14, iLit},
	sMetal: {1, 0.46, iLit}, sMetalDark: {1, 0.27, iLit}, sGlass: {2, 0.56, iLit}, sGlassHi: {2, 1, iEmit},
	sNeon1: {3, 1, iEmit}, sNeon2: {4, 1, iEmit}, sNeon3: {5, 1, iEmit}, sGlow: {2, 1, iEmit},
	sBeacon: {5, 1, iEmit}, sLeaf: {2, 0.5, iLit}, sLeafDark: {2, 0.3, iLit}, sTrunk: {1, 0.25, iLit},
	sField1: {3, 0.45, iLit}, sField2: {3, 0.3, iLit}, sRock: {1, 0.38, iLit}, sRockDark: {1, 0.24, iLit},
	sFire: {5, 1, iEmit}, sInk: {0, 0.1, iLit}, sWater: {2, 0.4, iLit}, sSmoke: {1, 0.2, iLit},
}

// iri is step i of the iridescent ramp, lit in mode at a depth row.
func (o *orb) iri(i, mode, depth int) tcell.Color {
	if o.iriTab == nil {
		o.iriTab = make([]tcell.Color, 8*numInkModes*numHaze)
		for h := 0; h < 8; h++ {
			for md := 0; md < numInkModes; md++ {
				for d := 0; d < numHaze; d++ {
					o.iriTab[(h*numInkModes+md)*numHaze+d] = o.raw(mapmodel.SkyIridescent[h], md, d)
				}
			}
		}
	}
	return o.iriTab[((i&7)*numInkModes+clampInt(mode, 0, numInkModes-1))*numHaze+clampInt(depth, 0, numHaze-1)]
}

// iriCol is a sprite slot's iridescent colour at ramp step i.
func (o *orb) iriCol(i int, sl slot, depth int) tcell.Color {
	if o.iq == nil {
		o.iq = make([]tcell.Color, 8*int(numSlots)*numHaze)
	}
	sp := iriSlot[sl]
	if sp.k == 0 {
		sp = iriSlot[sWall]
	}
	h := (i + sp.off) & 7
	depth = clampInt(depth, 0, numHaze-1)
	j := (h*int(numSlots)+int(sl))*numHaze + depth
	if c := o.iq[j]; c != 0 {
		return c
	}
	c := o.resolve(theme.Shade(theme.SpaceColor(mapmodel.SkyIridescent[h]), sp.k), sp.mode, depth)
	o.iq[j] = c
	return c
}

// quantumBackdrop draws everything behind the lots.
func (o *orb) quantumBackdrop() {
	o.stars(0, o.groundY, 0.009, 0.01, 271)
	o.fringes()
	o.echo()
	o.fractal()
	o.echoLine()
}

// fringes fill the dark under the baseline with an interference pattern,
// as if the whole panorama had come through two slits: the cosine of the
// difference of the distances to two points on the baseline, with a slow
// phase, fading with depth, in the probability inks and a breath of the
// ramp.
func (o *orb) fringes() {
	gy := o.groundY
	ax := float64(o.W)*0.38 - float64(o.cam)*0.02
	bx := float64(o.W)*0.62 - float64(o.cam)*0.02
	c1, c2 := o.c(mapmodel.InkCloud, iEmit, 1), o.c(mapmodel.InkCloud2, iEmit, 1)
	ph := float64(o.anim) / 36
	for y := gy + 2; y < o.S; y++ {
		fy := float64(2 * (y - gy))
		fall := clamp01(1.15 - float64(y-gy)/float64(max(1, o.S-gy)))
		for x := 0; x < o.W; x++ {
			da := math.Sqrt(float64((float64(x)-ax)*(float64(x)-ax)) + float64(fy*fy))
			db := math.Sqrt(float64((float64(x)-bx)*(float64(x)-bx)) + float64(fy*fy))
			v := mapmodel.Cos((da-db)/5.5-ph) * fall
			switch {
			case v > 0.8:
				o.fb.fg(x, o.Y(y), '▒', theme.Mix(c2, o.iri(x/9+o.anim/10, iEmit, 2), 0.3), dFringe)
			case v > 0.45:
				o.fb.fg(x, o.Y(y), '░', c1, dFringe)
			case v > 0.3 && (x+y)%2 == 0:
				o.fb.fg(x, o.Y(y), '·', c1, dFringe)
			}
		}
	}
}

// echo is the time echo: the town on the ground, as it was, a faint ghost
// behind everything (the ground layout at the same columns), a slow scan
// of memory passing over it.
func (o *orb) echo() {
	g := o.v.glay
	if g == nil || g.m != o.m {
		return
	}
	faint := theme.Mix(o.c(mapmodel.InkEcho, iBack, 0), o.c(mapmodel.InkHaze, iEmit, 0), 0.35)
	if o.light {
		faint = o.c(mapmodel.InkEcho, iLit, 2)
	}
	bright := theme.Mix(faint, o.c(mapmodel.InkHaze, iEmit, 0), 0.6)
	scan := (o.anim/2)%(o.W+60) - 30
	vis := g.visible(o.cam-16, o.cam+o.W+16, nil)
	for _, li := range vis {
		lv := &g.lots[li]
		spr := lv.spr
		x0 := lv.ml.X - spr.w/2 - o.cam
		y0 := o.groundY - spr.h
		for sy := 0; sy < spr.h; sy++ {
			for sx := 0; sx < spr.w; sx++ {
				c := spr.c[sy*spr.w+sx]
				X := x0 + sx
				if c.ch == 0 || X < 0 || X >= o.W {
					continue
				}
				col := faint
				if d := X - scan; d > -6 && d < 6 {
					col = theme.Mix(faint, bright, 1-math.Abs(float64(d))/6)
				}
				ch := c.ch
				if ch == '█' {
					o.fb.set(X, o.Y(y0+sy), ' ', col, col, dEcho)
					continue
				}
				o.fb.fg(X, o.Y(y0+sy), ch, col, dEcho)
			}
		}
	}
}

// echoLine is the echo's faint ground line, which is the baseline the
// lots stand on.
func (o *orb) echoLine() {
	gy := o.groundY
	c := o.c(mapmodel.InkEcho, iEmit, 0)
	for x := 0; x < o.W; x++ {
		wx := x + o.cam
		if wx < o.x0 || wx > o.x1 {
			continue
		}
		ch := '─'
		if wx%9 == 0 {
			ch = '┴'
		}
		o.fb.fg(x, o.Y(gy), ch, theme.Mix(c, o.iri(wx/3+o.anim/4, iEmit, 2), 0.25), dEchoLine)
	}
}

// fractalSprite is the landmark at one scale: a ring with its spindle and
// disc, hollow at the heart where the next smaller copy sits.
func fractalSprite(R float64) *sprite {
	w := int(R*3.2) + 3
	h := int(R*1.3) + 2
	s := newSprite(w, h)
	cx, cpy := float64(w)/2, float64(h)
	thick := math.Max(1, R*0.13)
	hollow := R * 0.5
	for py := 0; py < 2*h; py++ {
		for x := 0; x < w; x++ {
			dx := float64(x) + 0.5 - cx
			dy := float64(py) + 0.5 - cpy
			d := math.Sqrt(float64(dx*dx) + float64(dy*dy))
			switch {
			case d <= R && d >= R-thick:
				sl := sWall
				if dx < -R*0.5 {
					sl = sWallLit
				} else if dx > R*0.4 {
					sl = sWallShade
				}
				s.pset(x, py, sl)
			case math.Abs(dy) <= math.Max(0.6, R*0.07) && math.Abs(dx) <= R*1.55 && d > R:
				s.pset(x, py, sMetal) // the disc, out past the ring
			case math.Abs(dx) <= math.Max(0.5, R*0.05) && math.Abs(dy) <= R*1.25 && d > hollow && d < R-thick:
				s.pset(x, py, sTrim) // the spindle, stopping at the hollow
			case d > R && d < R+thick*1.5 && math.Abs(math.Abs(dx)-math.Abs(dy)) < thick*0.8:
				s.pset(x, py, sNeon1) // pylons at the diagonals
			}
		}
	}
	return s
}

// fractal draws the landmark at the present three times, each copy inside
// the last, its colours flowing inward as if the view were falling into it.
func (o *orb) fractal() {
	gy := o.groundY
	R := float64(clampInt(gy*62/100, 5, 24))
	cx := o.present.Centre - o.cam
	cy := float64(gy)*0.4 + 0.5
	if cx+int(R*2) < 0 || cx-int(R*2) > o.W {
		return
	}
	if o.v.sprites == nil {
		o.v.sprites = map[spriteKey]*sprite{}
	}
	for lvl := 0; lvl < 3; lvl++ {
		r := R * [3]float64{1, 0.46, 0.21}[lvl]
		k := spriteKey{"\x00fractal", uint64(lvl), int(r * 10)}
		spr := o.v.sprites[k]
		if spr == nil {
			spr = fractalSprite(r)
			o.v.sprites[k] = spr
		}
		x0 := cx - spr.w/2
		y0 := int(cy) - spr.h/2
		step := lvl*3 - o.anim/6
		for sy := 0; sy < spr.h; sy++ {
			for sx := 0; sx < spr.w; sx++ {
				c := spr.c[sy*spr.w+sx]
				if c.ch == 0 {
					continue
				}
				X, Y := x0+sx, y0+sy
				if Y < 0 || Y >= gy {
					continue
				}
				fg := o.iriCol(step+sx/4, c.fg, 1+lvl/2)
				d := dFractal // each copy over the last
				if c.bg == sNone {
					if c.ch == '█' {
						o.fb.fill(X, o.Y(Y), fg, d)
					} else {
						o.fb.fg(X, o.Y(Y), c.ch, fg, d)
					}
					continue
				}
				o.fb.set(X, o.Y(Y), c.ch, fg, o.iriCol(step+sx/4, c.bg, 1+lvl/2), d)
			}
		}
	}
}

// quantumLots draws the lots in superposition: each shows its sky form or
// the building it was on the ground (the ground layout's sprite), by a
// wave sweeping along the panorama, both at once for a moment as the wave
// passes; probability clouds shimmer round them.
func (o *orb) quantumLots() {
	g := o.v.glay
	if g != nil && g.m != o.m {
		g = nil
	}
	white := o.c(mapmodel.InkStarBright, iEmit, 0)
	void := o.c(mapmodel.InkVoid, iBack, 0)
	wave := func(lv *lotView) int { return ((o.anim/8+lv.ml.X/12)%4 + 4) % 4 } // four steps, sweeping east
	for _, li := range o.vis {
		lv := &o.lay.lots[li]
		if q := wave(lv); q == 1 || q == 3 {
			o.halo(lv) // a cloud of probability round what is about to change
		}
	}
	for _, li := range o.vis {
		lv := &o.lay.lots[li]
		var other *sprite
		if g != nil {
			if gi := g.find(lv.ml.Key, lv.ml.Copy); gi >= 0 {
				other = g.lots[gi].spr
			}
		}
		if other == nil {
			other = lv.spr.mirror()
		}
		q := wave(lv)
		show, ghost := lv.spr, other
		pale := 0.0
		if q >= 2 {
			show, ghost, pale = other, lv.spr, 0.42 // the building it was: a paler possibility
		}
		if (q == 1 || q == 3) && o.anim%8 >= 5 {
			back := *lv
			back.depth = lv.depth + 1
			o.drawSprite(&back, ghost, o.wireCol(ghost, lv.ml.X, 0.6, void), white, false)
		}
		o.drawSprite(lv, show, o.wireCol(show, lv.ml.X, pale, void), white, li == o.sel)
		o.lotLights(lv, show)
	}
}

// wireCol colours a sprite as the Quantum draws it: dark fills, edges lit
// in the iridescent ramp (which runs along the panorama and with time),
// self-lit slots at full strength; pale fades it toward the void.
func (o *orb) wireCol(spr *sprite, wx0 int, pale float64, void tcell.Color) func(sx, sy int, sl slot, hz int) tcell.Color {
	return func(sx, sy int, sl slot, hz int) tcell.Color {
		i := (wx0+sx)/16 + sy/7 + o.anim/8
		var c tcell.Color
		edge := spr.at(sx-1, sy).ch == 0 || spr.at(sx+1, sy).ch == 0 || spr.at(sx, sy-1).ch == 0 || spr.at(sx, sy+1).ch == 0
		switch {
		case !edge:
			c = o.iriCol(i, sWallDark, hz) // the dark inside
		case iriSlot[sl].mode == iEmit:
			c = o.iriCol(i, sl, hz)
		default:
			c = theme.Mix(o.iriCol(i, sWallLit, hz), o.pearl, 0.3) // an edge, lit like mother of pearl
		}
		if pale > 0 {
			c = theme.Mix(c, void, pale)
		}
		return c
	}
}

// halo is a lot's probability cloud: interference fringes from two points
// either side of it (the cosine of the difference of the distances, with a
// slow phase), fading with distance, in the cloud inks.
func (o *orb) halo(lv *lotView) {
	spr := lv.spr
	x0 := lv.x0 - o.cam
	top := o.groundY - spr.h
	ay := float64(top) + float64(spr.h)/2
	ax, bx := float64(x0)-1.5, float64(x0+spr.w)+0.5
	c1, c2 := o.c(mapmodel.InkCloud, iEmit, lv.ml.Row), o.c(mapmodel.InkCloud2, iEmit, lv.ml.Row)
	ph := float64(o.anim)/48 + hashf(int(lv.ml.Seed))
	for y := max(0, top-2); y < o.groundY; y++ {
		for x := max(0, x0-3); x < min(o.W, x0+spr.w+3); x++ {
			fy := float64(2 * (float64(y) - ay))
			da := math.Sqrt(float64((float64(x)-ax)*(float64(x)-ax)) + float64(fy*fy))
			db := math.Sqrt(float64((float64(x)-bx)*(float64(x)-bx)) + float64(fy*fy))
			v := mapmodel.Cos((da-db)/3.2 + ph)
			edge := math.Min(float64(x-(x0-3)), float64(x0+spr.w+2-x))
			fall := clamp01(edge/3) * clamp01(float64(y-(top-3))/3)
			v *= fall
			switch {
			case v > 0.72:
				o.fb.fg(x, o.Y(y), '▒', c2, dHaloQ)
			case v > 0.38:
				o.fb.fg(x, o.Y(y), '░', c1, dHaloQ)
			}
		}
	}
}
