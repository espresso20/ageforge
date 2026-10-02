package skyline

import (
	"math"
	"sort"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
)

// orbit_galaxy.go is the Galactic Age: sci-fi at its most confident. Rich
// gold, cyan and crimson on a deep blue-black; a spiral galaxy hangs in
// the far sky over full-colour nebulae; the civs met are star systems
// joined by gold trade lanes; the baseline is the starbase's docking ring,
// gold-trimmed white with cyan running lights; and over the present stands
// the starbase hub, a spindle through a great disc with docking pylons
// arching out, the Cosmic Beacon blazing on its crown once it is built.
// Starships cruise and jump to warp; alien saucers are ordinary traffic;
// transporter sparkles come and go by the decks.

// galaxyBackdrop draws everything behind the lots.
func (o *orb) galaxyBackdrop() {
	o.galNebulae()
	o.spiral()
	o.stars(0, o.S, 0.026, 0.02, 733)
	o.hub()
	o.dockRing()
}

const galNebW = 480

// galNebulae are soft patches of magenta, blue and gold, far off.
func (o *orb) galNebulae() {
	seed := int64(o.m.Seed%6007) + 11
	t := o.v.texture("gneb", seed, galNebW, o.S, func(x, y int) (uint8, uint8) {
		fx, fy := float64(x), float64(y)
		n := float64(0.7*pnoise(seed, fx/40, fy/9, galNebW/40)) + float64(0.3*pnoise(seed+3, fx/16, fy/4, galNebW/16))
		v := smoothstep(0.5, 0.82, n)
		c := pnoise(seed+5, fx/80, fy/20, galNebW/80)
		return uint8(v * 255), uint8(c * 255)
	})
	void := o.c(mapmodel.InkVoid, iBack, 0)
	cols := [3]tcell.Color{o.c(mapmodel.InkCloud, iBack, 0), o.c(mapmodel.InkSurface, iBack, 0), o.c(mapmodel.InkSurface2, iBack, 0)}
	if o.light {
		cols = [3]tcell.Color{o.c(mapmodel.InkCloud, iEmit, 3), o.c(mapmodel.InkSurface, iEmit, 3), o.c(mapmodel.InkSurface2, iEmit, 3)}
	}
	off := int(float64(o.cam) * 0.06)
	for y := 0; y < o.S; y++ {
		for x := 0; x < o.W; x++ {
			a, b := texAt(t, x+off, y)
			if a < 12 {
				continue
			}
			v := float64(a) / 255
			col := cols[min(2, int(b)*3/256)]
			bg := theme.Mix(void, col, float64(v*0.5))
			if v > 0.7 && hash(x+off, y, 3)%3 == 0 {
				o.fb.set(x, o.Y(y), '░', theme.Mix(void, col, 0.85), bg, dGalNebula)
			} else {
				o.fb.set(x, o.Y(y), ' ', bg, bg, dGalNebula)
			}
		}
	}
}

// spiralGeom is the galaxy's centre and size: screen anchored, upper left
// of the middle.
func (o *orb) spiralGeom() (float64, float64, float64, float64) {
	gy := o.groundY
	a := math.Min(float64(o.W)*0.24, float64(gy)*1.5)
	b := a * 0.27
	return float64(o.W) * 0.28, float64(gy) * 0.3, a, b
}

// spiral is a spiral-arm galaxy seen at a slant, precomputed per size: arms
// in the haze ink round a bright core.
func (o *orb) spiral() {
	cx, cy, a, b := o.spiralGeom()
	w, h := int(2*a)+3, int(2*b)+3
	t := o.v.texture("spiral", int64(o.m.Seed%13), w, h, func(x, y int) (uint8, uint8) {
		dx := (float64(x) + 0.5 - float64(w)/2) / a
		dy := (float64(y) + 0.5 - float64(h)/2) / b
		r := math.Sqrt(float64(dx*dx) + float64(dy*dy))
		if r >= 1 {
			return 0, 0
		}
		th := turns(dx, dy)
		arm := mapmodel.Cos(float64(2*th) - float64(0.62*mapmodel.Log2(r+0.03)))
		armV := math.Max(0, arm)
		armV = float64(float64(armV*armV)*armV) * (1 - r) * math.Sqrt(1-r) * 1.25
		core := 0.0
		if r < 0.3 {
			core = (1 - r/0.3) * (1 - r/0.3)
		}
		return uint8(math.Min(1, armV*1.15+0.1*(1-r)) * 255), uint8(core * 255)
	})
	void := o.c(mapmodel.InkVoid, iBack, 0)
	arm := theme.Mix(o.c(mapmodel.InkHaze, iEmit, 0), o.c(mapmodel.InkRock, iEmit, 1), 0.35)
	core := o.c(mapmodel.InkRock, iEmit, 0)
	hot := o.c(mapmodel.InkStarBright, iEmit, 1)
	x0, y0 := int(cx)-w/2, int(cy)-h/2
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			v, c := texAt(t, x, y)
			if v < 10 && c == 0 {
				continue
			}
			X, Y := x0+x, y0+y
			if X < 0 || X >= o.W || Y < 0 || Y >= o.groundY {
				continue
			}
			k := float64(v) / 255
			col := theme.Mix(void, arm, math.Min(1, k*1.3))
			if c > 0 {
				col = theme.Mix(col, core, float64(c)/255)
			}
			ch, fg := ' ', col
			switch {
			case c > 200:
				ch, fg = '█', theme.Mix(core, hot, 0.5)
			case c > 120:
				ch, fg = '▓', theme.Mix(core, hot, 0.3)
			case k > 0.55:
				ch, fg = '░', theme.Mix(col, hot, 0.25)
			case k > 0.2 && hash(X, Y, 9)%5 == 0:
				ch, fg = '·', theme.Mix(col, hot, 0.4)
			}
			o.fb.set(X, o.Y(Y), ch, fg, col, dSpiral)
		}
	}
}

// starSystems are the civs met as star systems in the far layer, each a
// bright star with its name and a pennant in your standing's colour,
// joined west to east by dotted gold trade lanes (bright where you trade).
func (o *orb) starSystems() {
	type sys struct {
		x, y int
		f    *mapmodel.Faction
		i    int
	}
	var ss []sys
	for _, it := range o.ridge {
		if it.fac != nil {
			ss = append(ss, sys{it.x, o.farRow(it), it.fac, it.i})
		}
	}
	sort.Slice(ss, func(i, j int) bool { return ss[i].x < ss[j].x })
	lane := o.c(mapmodel.InkAccent, iEmit, 3)
	busy := o.c(mapmodel.InkAccent, iEmit, 1)
	for i := 1; i < len(ss); i++ {
		a, b := ss[i-1], ss[i]
		c := lane
		if a.f.TradeCount > 0 && b.f.TradeCount > 0 {
			c = busy
		}
		n := max(abs(b.x-a.x), 1)
		for k := 2; k < n-1; k += 2 {
			x := a.x + k*(b.x-a.x)/n
			y := a.y + k*(b.y-a.y)/n
			o.fb.fg(x, o.Y(y), '·', c, dSystem)
		}
	}
	tints := [3]mapmodel.SkyInk{mapmodel.InkStarBright, mapmodel.InkAccent, mapmodel.InkAccent2}
	for _, s := range ss {
		c := o.c(tints[s.i%3], iEmit, 0)
		glow := theme.Mix(o.c(mapmodel.InkVoid, iBack, 0), c, 0.35)
		o.fb.fg(s.x-1, o.Y(s.y), '─', glow, dSystem)
		o.fb.fg(s.x+1, o.Y(s.y), '─', glow, dSystem)
		o.fb.fg(s.x, o.Y(s.y-1), '│', glow, dSystem)
		o.fb.fg(s.x, o.Y(s.y+1), '│', glow, dSystem)
		o.fb.fg(s.x, o.Y(s.y), mapmodel.R(mapmodel.SymStarBig, o.tier), c, dSystem)
		rc := o.mp.Fg[mapmodel.RelationClass(s.f.Relation)]
		flag := '►'
		if s.f.Relation == mapmodel.RelWar {
			flag = mapmodel.R(mapmodel.SymWar, o.tier)
		} else if o.tier == mapmodel.TierNerd {
			flag = mapmodel.R(mapmodel.SymCiv, o.tier)
		}
		o.fb.fg(s.x+2, o.Y(s.y-1), flag, rc, dSystem)
		o.labels = append(o.labels, label{s.x - textLen(s.f.Name)/2, s.y - 2, s.f.Name, rc})
	}
}

// dockRing is the starbase's docking ring along the baseline: a strong
// white band trimmed in gold, its cyan running lights chasing along it.
func (o *orb) dockRing() {
	gy := o.groundY
	gold := o.c(mapmodel.InkAccent, iLit, 0)
	white := o.c(mapmodel.InkFrame, iLit, 0)
	dim := o.c(mapmodel.InkFrameDim, iLit, 0)
	cyan := o.c(mapmodel.InkAccent2, iEmit, 0)
	void := o.c(mapmodel.InkVoid, iBack, 0)
	for x := 0; x < o.W; x++ {
		wx := x + o.cam
		if wx < o.x0 || wx > o.x1 {
			continue
		}
		o.fb.set(x, o.Y(gy), '▀', gold, white, dDockRing)
		ch, fg := '▀', dim
		if (wx-o.anim/2)%7 == 0 {
			ch, fg = '▪', cyan
		} else if wx%7 == 3 {
			ch, fg = '▀', theme.Mix(dim, cyan, 0.3)
		}
		o.fb.set(x, o.Y(gy+1), ch, fg, theme.Mix(void, dim, 0.25), dDockRing)
	}
}

// hub draws the starbase hub over the present district: a spindle through
// a great disc and its habitat ring, docking pylons arching up and down,
// gold trim and cyan lights, and atop it the Cosmic Beacon once it is
// built.
func (o *orb) hub() {
	gy := o.groundY
	H := clampInt(gy-1, 8, 40)
	W := clampInt(int(float64(H)*2.9), 20, 110)
	k := spriteKey{"\x00hub", 0, W<<8 | H}
	if o.v.sprites == nil {
		o.v.sprites = map[spriteKey]*sprite{}
	}
	spr := o.v.sprites[k]
	if spr == nil {
		spr = hubSprite(W, H)
		o.v.sprites[k] = spr
	}
	x0 := o.present.Centre - spr.w/2 - o.cam
	if x0+spr.w < 0 || x0 > o.W {
		return
	}
	y0 := gy - spr.h
	white := o.c(mapmodel.InkStarBright, iEmit, 0)
	for sy := 0; sy < spr.h; sy++ {
		for sx := 0; sx < spr.w; sx++ {
			c := spr.c[sy*spr.w+sx]
			if c.ch == 0 {
				continue
			}
			X, Y := x0+sx, o.Y(y0+sy)
			fg := o.hubCol(c.fg)
			if c.fg == sWin && hash(sx, sy, o.anim/40)%4 != 0 {
				fg = theme.Mix(o.c(mapmodel.InkLight, iEmit, 1), white, 0.4)
			}
			if c.fg == sNeon1 && (o.anim/4+sx)%6 == 0 {
				fg = white
			}
			if c.bg == sNone {
				if c.ch == '█' {
					o.fb.fill(X, Y, fg, dHub)
				} else {
					o.fb.fg(X, Y, c.ch, fg, dHub)
				}
				continue
			}
			o.fb.set(X, Y, c.ch, fg, o.hubCol(c.bg), dHub)
		}
	}
	// the Cosmic Beacon, lit atop the spindle once it is built
	for _, w := range o.m.Wonders {
		if w.Key != "cosmic_beacon" || !w.Built {
			continue
		}
		cx := x0 + spr.w/2
		beam := o.c(mapmodel.InkGlow, iEmit, 0)
		top := y0 + spr.topAt(spr.w/2)
		pulse := 0.5 + 0.5*mapmodel.Sin(float64(o.anim)/32)
		for y := 0; y < top-1; y++ {
			ch := '│'
			if (y+o.anim/2)%4 == 0 {
				ch = '║'
			}
			o.fb.fg(cx, o.Y(y), ch, theme.Mix(o.c(mapmodel.InkVoid, iBack, 0), beam, 0.45+0.55*pulse), dHub)
		}
		o.fb.fg(cx-1, o.Y(top-1), '◢', o.c(mapmodel.InkAccent, iEmit, 0), dHub)
		o.fb.fg(cx, o.Y(top-1), '◆', white, dHub)
		o.fb.fg(cx+1, o.Y(top-1), '◣', o.c(mapmodel.InkAccent, iEmit, 0), dHub)
	}
}

// hubCol colours the hub's slots.
func (o *orb) hubCol(sl slot) tcell.Color {
	switch sl {
	case sWall:
		return o.c(mapmodel.InkFrame, iLit, 1)
	case sWallLit:
		return o.c(mapmodel.InkFrame, iLit, 0)
	case sWallShade:
		return o.c(mapmodel.InkFrameDim, iLit, 1)
	case sTrim:
		return o.c(mapmodel.InkAccent, iLit, 1)
	case sNeon1:
		return o.c(mapmodel.InkAccent2, iEmit, 1)
	case sNeon3:
		return o.c(mapmodel.InkAccent3, iEmit, 1)
	case sGlow:
		return o.c(mapmodel.InkGlow, iEmit, 1)
	case sWin:
		return o.c(mapmodel.InkFrameDim, iLit, 2)
	case sMetal:
		return o.c(mapmodel.InkFrameDim, iLit, 1)
	}
	return o.c(mapmodel.InkFrameDim, iLit, 2)
}

// hubSprite draws the starbase hub W wide and H tall in half-block
// pixels: the rings sit two fifths of the way down, so the upper pylons and
// the crown stand clear of the decks docked in front.
func hubSprite(W, H int) *sprite {
	s := newSprite(W, H)
	cx := float64(W) / 2
	P := float64(2 * H) // pixel rows
	mid := P * 0.4
	set := func(x int, py int, sl slot) {
		if x >= 0 && x < W && py >= 0 && py < 2*H {
			s.pset(x, py, sl)
		}
	}
	ellipse := func(ex, ey, rx, ry float64, fill func(dx, dy float64) slot) {
		for py := int(ey - ry - 1); py <= int(ey+ry+1); py++ {
			for x := int(ex - rx - 1); x <= int(ex+rx+1); x++ {
				dx := (float64(x) + 0.5 - ex) / rx
				dy := (float64(py) + 0.5 - ey) / ry
				if float64(dx*dx)+float64(dy*dy) <= 1 {
					if sl := fill(dx, dy); sl != sNone {
						set(x, py, sl)
					}
				}
			}
		}
	}
	curve := func(x0, y0, x1, y1, x2, y2 float64, sl slot, thick int) {
		for i := 0; i <= 80; i++ {
			t := float64(i) / 80
			u := 1 - t
			x := float64(u*u)*x0 + float64(2*u*t)*x1 + float64(t*t)*x2
			y := float64(u*u)*y0 + float64(2*u*t)*y1 + float64(t*t)*y2
			for k := 0; k < thick; k++ {
				set(int(x)+k, int(y), sl)
			}
		}
	}
	// the spindle, crown to keel
	for py := 1; py < int(P)-1; py++ {
		w := 1
		if py > int(mid-P*0.16) && py < int(mid+P*0.2) {
			w = 2
		}
		for k := -w; k <= w; k++ {
			sl := sWall
			switch {
			case k == -w:
				sl = sWallLit
			case k == w:
				sl = sWallShade
			}
			set(int(cx)+k, py, sl)
		}
	}
	// the docking pylons arching up from the outer ring, and down
	for _, side := range []float64{-1, 1} {
		ox := cx + side*float64(W)*0.4
		curve(ox, mid-1, ox-side*float64(W)*0.02, mid-P*0.36, cx+side*float64(W)*0.15, mid-P*0.37, sWallShade, 2)
		curve(ox, mid+2, ox-side*float64(W)*0.04, mid+P*0.4, cx+side*float64(W)*0.13, mid+P*0.42, sWallShade, 1)
		tx := int(cx + side*float64(W)*0.15)
		set(tx, int(mid-P*0.37)-1, sNeon1)
		set(tx+1, int(mid-P*0.37)-1, sNeon1)
	}
	// the docking ring, a great flat ellipse with a gold rim
	ellipse(cx, mid, float64(W)*0.47, P*0.075, func(dx, dy float64) slot {
		switch {
		case dy < -0.25:
			return sTrim
		case math.Abs(dx) < 0.45 && dy < 0.4:
			return sWallShade
		}
		return sWall
	})
	// the habitat ring above it and the core's bulge
	ellipse(cx, mid-P*0.09, float64(W)*0.22, P*0.06, lit(sWallLit, sWall, sWallShade))
	ellipse(cx, mid+P*0.05, float64(W)*0.06, P*0.12, lit(sWallLit, sWall, sWallShade))
	// windows on the habitat ring, lights on the rim, the reactor's glow
	wy := int(mid-P*0.09) / 2
	for x := int(cx - float64(W)*0.2); x < int(cx+float64(W)*0.2); x += 2 {
		if c := s.at(x, wy); c.ch != 0 {
			s.put(x, wy, '▪', sWin, sWall)
		}
	}
	ry0 := int(mid+1) / 2
	for x := int(cx - float64(W)*0.46); x < int(cx+float64(W)*0.46); x += 5 {
		if c := s.at(x, ry0); c.ch != 0 {
			s.put(x, ry0, '•', sNeon1, c.fg)
		}
	}
	s.put(int(cx), int(mid+P*0.07)/2, '◘', sGlow, sWall)
	if top := s.topAt(int(cx)); top > 0 {
		s.put(int(cx), top-1, '•', sNeon3, sNone)
	}
	return s.trimTop()
}

// sparkles are transporter sparkles: a few cells by the decks twinkling
// through their shapes for a moment, now here, now there.
func (o *orb) sparkles() {
	if len(o.vis) == 0 {
		return
	}
	seq := [4]rune{'·', '✧', '˙', '·'}
	cyan, gold := o.c(mapmodel.InkAccent2, iEmit, 0), o.c(mapmodel.InkAccent, iEmit, 0)
	for k := 0; k < 3; k++ {
		slot := o.anim/16 + k*7
		if hash(slot, k, 3)%3 == 0 {
			continue
		}
		lv := &o.lay.lots[o.vis[int(hash(slot, k)%uint64(len(o.vis)))]]
		ph := (o.anim + 5*k) % 16
		if ph < 0 || ph >= 8 {
			continue
		}
		c := cyan
		if hash(slot, 9)%2 == 0 {
			c = gold
		}
		x0, y0 := lv.x0-o.cam, o.groundY-lv.spr.h
		for i := 0; i < 4; i++ {
			x := x0 + int(hash(slot, i, 1)%uint64(max(1, lv.spr.w)))
			y := y0 + int(hash(slot, i, 2)%uint64(max(1, lv.spr.h)))
			o.fb.fg(x, o.Y(y), seq[(ph/2+i)%4], c, lv.depth)
		}
	}
}
