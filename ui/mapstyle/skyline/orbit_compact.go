package skyline

import (
	"math"
	"strconv"
	"strings"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
)

// orbit_compact.go is the sky scenes squeezed into the mini map (40x15, and
// 65x9 in the old layout): the void and a few stars, the age's signature
// (the planet's arc under the truss in the Space Age; the sun, a nebula and
// the gate in the Interstellar; the starbase and a warp streak in the
// Galactic; the flicker in the Quantum; a small mandala in the
// Transcendent) and the whole panorama's profile on the baseline, with the
// full view's signals riding on it: lit windows, ▼ new, ★ wonders. The rare
// visitor never shows here.

func (v *view) composeSkyCompact(f mapstyle.Frame, W, H int, sc mapmodel.SkyScene) {
	m := f.Model
	lay := v.skyLayoutFor(m, sc, compactGround, true)
	p, mp := v.palettes(m)
	v.fb.reset(W, H)
	S := H - 1
	s := &scene{v: v, m: m, lay: lay, p: p, mp: mp, fb: &v.fb, tier: f.Tier, anim: f.Anim, W: W, H: H,
		S: S, top: 1, groundY: compactBase(S), sel: -1, band: bandOf(m.AgeIdx)}
	if S >= 1 {
		o := newOrb(s, sc)
		o.compactScene()
	}
	bg := chromeBg()
	for x := 0; x < W; x++ {
		s.fb.set(x, 0, ' ', bg, bg, dTop)
	}
	name := strings.ToUpper(strings.TrimSuffix(m.AgeName, " Age"))
	acc := role(theme.RoleAccent, bg)
	x := s.fb.text(0, 0, clip(" "+name+" ", W), theme.Legible(bg, acc, 4.5), acc, dTop)
	parts := []seg{{" " + m.Clock.String(), theme.RoleText}}
	if n := m.Recap.NewCount; v.changes && n > 0 {
		parts = append(parts, seg{" ▼" + strconv.Itoa(n), theme.RolePositive})
	}
	if n := idleProducers(m); n > 0 {
		parts = append(parts, seg{" " + strconv.Itoa(n) + " idle", theme.RoleWarning})
	}
	if m.Harbinger != nil {
		parts = append(parts, seg{" " + string(mapmodel.R(mapmodel.SymHarbinger, f.Tier)), theme.RoleWarning})
	}
	if m.Catastrophe.Pending != "" {
		parts = append(parts, seg{" " + string(mapmodel.R(mapmodel.SymFire, f.Tier)), theme.RoleNegative})
	}
	s.segs(x, 0, bg, parts)
}

// compactBase is the compact view's baseline row for S scene rows: two
// thirds of the way down, leaving room below for the planet or the
// reflection.
func compactBase(S int) int {
	if S <= 2 {
		return max(0, S-1)
	}
	return clampInt(int(float64(S)*0.66+0.5), 1, S-1)
}

func (o *orb) compactScene() {
	S, W, b := o.S, o.W, o.groundY
	void := o.c(mapmodel.InkVoid, iBack, 0)
	for y := 0; y < S; y++ {
		for x := 0; x < W; x++ {
			o.fb.set(x, o.Y(y), ' ', void, void, dSky)
		}
	}
	star, bright := o.c(mapmodel.InkStar, iEmit, 0), o.c(mapmodel.InkStarBright, iEmit, 0)
	density := 0.05
	if o.sc == mapmodel.SkyMandala {
		density = 0.01
	}
	for y := 0; y < b; y++ {
		for x := 0; x < W; x++ {
			if hv := hashf(x, y, 991); hv < density {
				c := theme.Mix(void, star, 0.6)
				if hash(x, y, o.anim/5)%11 == 0 {
					c = bright
				}
				o.fb.fg(x, o.Y(y), '·', c, dStar)
			}
		}
	}
	switch o.sc {
	case mapmodel.SkyOrbit:
		o.compactOrbit()
	case mapmodel.SkyDeep:
		o.compactDeep()
	case mapmodel.SkyGalaxy:
		o.compactGalaxy()
	case mapmodel.SkyQuantum:
		o.compactFringes()
	case mapmodel.SkyMandala:
		o.compactMandala()
	}
	o.compactProfile()
	if o.sc == mapmodel.SkyMandala {
		o.compactReflect()
	}
}

// compactSpan maps compact column x onto the panorama.
func (o *orb) compactCols(x int) (int, int) {
	a0, a1 := compactSpan(o.m)
	a := a0 + x*(a1-a0)/max(1, o.W)
	return a, max(a+1, a0+(x+1)*(a1-a0)/max(1, o.W))
}

// compactProfile is the panorama's profile standing on the baseline, in
// eighth blocks, coloured as its tallest lot, with lit windows, ▼ on what
// is new and ★ on the wonders; in the Quantum Age it flickers.
func (o *orb) compactProfile() {
	lay, b := o.lay, o.groundY
	if b <= 0 {
		return
	}
	a0, a1 := compactSpan(o.m)
	maxH := 1
	for x := a0; x < a1 && x < len(lay.prof); x++ {
		maxH = max(maxH, lay.prof[x])
	}
	top := float64(b) * 0.85
	switch o.sc {
	case mapmodel.SkyMandala:
		top = float64(b) * 0.45 // short light, the mandala over it
	case mapmodel.SkyGalaxy:
		top = float64(b) * 0.7 // the starbase stands over it
	}
	wonderAt := make([]bool, o.W) // the columns the recent landmarks stand in
	recent := o.present.Age - 2
	for i := range lay.lots {
		if ml := lay.lots[i].ml; ml.Wonder && a1 > a0 && (ml.Age >= recent || ml.New) {
			if c := (ml.X - a0) * o.W / (a1 - a0); c >= 0 && c < o.W {
				wonderAt[c] = true
			}
		}
	}
	lit := o.c(mapmodel.InkLight, iEmit, 0)
	fresh := o.mp.Fg[mapmodel.CFresh]
	gold := o.c(mapmodel.InkAccent, iEmit, 0)
	if o.sc == mapmodel.SkyOrbit {
		gold = o.c(mapmodel.InkLight, iEmit, 0)
	}
	echo := o.c(mapmodel.InkEcho, iEmit, 0)
	for x := 0; x < o.W; x++ {
		a, bb := o.compactCols(x)
		h, li, isNew := 0, int32(-1), false
		for i := a; i < bb && i < len(lay.prof); i++ {
			if i < 0 {
				continue
			}
			if lay.prof[i] > h {
				h, li = lay.prof[i], lay.tallest[i]
			}
			if t := lay.tallest[i]; t >= 0 && lay.lots[t].ml.New {
				isNew = true
			}
		}
		if h == 0 || li < 0 {
			continue
		}
		lv := &lay.lots[li]
		wall := o.lotCol(lv, sWall, 0)
		if o.sc == mapmodel.SkyDeep { // small worlds: blue, green and rust
			wall = o.lotCol(lv, [3]slot{sWater, sLeaf, sRock}[hash(int(lv.ml.Seed), 3)%3], 0)
		}
		hh := float64(h) / float64(maxH) * top
		if o.sc == mapmodel.SkyQuantum {
			wall = o.iriCol(x/2+o.anim/4, sWall, 0)
			if ((o.anim/8+x/3)%4+4)%4 >= 2 { // the other self
				wall = theme.Mix(echo, wall, 0.35)
				hh *= 0.6 + 0.4*hashf(x, 7)
			}
		}
		full := int(hh)
		for k := 0; k < full; k++ {
			y := b - 1 - k
			o.fb.fill(x, o.Y(y), wall, dRow0)
			if k < full-1 && lv.staff > 0 && hashf(x, y, int(o.m.Seed)) < 0.1+0.6*lv.staff && o.sc != mapmodel.SkyMandala {
				o.fb.set(x, o.Y(y), '·', lit, wall, dRow0)
			}
		}
		if f := hh - float64(full); f > 0.06 {
			o.fb.fg(x, o.Y(b-1-full), eighths[clampInt(int(f*8), 1, 7)], wall, dRow0)
		}
		if ay := b - 2 - full; ay >= 0 {
			switch {
			case o.v.changes && isNew:
				o.fb.fg(x, o.Y(ay), '▼', fresh, dTag)
			case wonderAt[x]:
				o.fb.fg(x, o.Y(ay), mapmodel.R(mapmodel.SymWonder, o.tier), gold, dTag)
			}
		}
	}
	// the baseline itself
	line := o.c(mapmodel.InkFrame, iLit, 0)
	ch, ld := '═', dTruss
	switch o.sc {
	case mapmodel.SkyDeep:
		line, ch, ld = o.c(mapmodel.InkFrameDim, iLit, 0), '·', dFormation
	case mapmodel.SkyGalaxy:
		line, ch, ld = o.c(mapmodel.InkAccent, iLit, 0), '▀', dDockRing
	case mapmodel.SkyQuantum:
		line, ch, ld = echo, '─', dEchoLine
	case mapmodel.SkyMandala:
		line, ch, ld = o.c(mapmodel.InkAccent, iEmit, 1), '─', dMirror
	}
	if b < o.S {
		for x := 0; x < o.W; x++ {
			if o.sc == mapmodel.SkyDeep && x%2 == 1 {
				continue
			}
			o.fb.fg(x, o.Y(b), ch, line, ld)
		}
	}
	if o.m.Catastrophe.Pending != "" && b > 2 {
		crack := o.p.hue(theme.SkyCrack, mEmit, 0)
		x := int(hash(o.anim/6, 3) % uint64(max(1, o.W)))
		for y := 0; y < b-1; y++ {
			if hash(y, x)%2 == 0 {
				x++
			} else {
				x--
			}
			o.fb.fg(x, o.Y(y), '╱', crack, dCrack)
		}
	}
}

// compactOrbit: the planet's arc under the truss, its cities lit, the moon
// and a shuttle climbing.
func (o *orb) compactOrbit() {
	S, W, b := o.S, o.W, o.groundY
	if b+1 >= S {
		return
	}
	glow := o.c(mapmodel.InkGlow, iEmit, 0)
	ocean := o.c(mapmodel.InkSurface, iBack, 0)
	land := o.c(mapmodel.InkSurface2, iBack, 0)
	city := o.c(mapmodel.InkLight, iEmit, 0)
	void := o.c(mapmodel.InkVoid, iBack, 0)
	limb := limbOf(W, S, max(0, b-2))
	for x := 0; x < W; x++ {
		lp := limb[x]
		for y := 0; y < S; y++ {
			t, bt := float64(2*y)+0.5-lp, float64(2*y+1)+0.5-lp
			if bt < 0 {
				continue
			}
			surf := ocean
			if mapmodel.Noise(int64(o.m.Seed%97)+5, float64(x)/7, float64(y)/2.5) > 0.55 {
				surf = land
			}
			switch {
			case t < 0:
				o.fb.set(x, o.Y(y), '▄', glow, void, dPlanet)
			case t < 1.2:
				o.fb.set(x, o.Y(y), '▀', glow, surf, dPlanet)
			default:
				ch := ' '
				if t < 9 && hash(x, y, 31)%4 == 0 {
					ch = '·'
				}
				o.fb.set(x, o.Y(y), ch, city, surf, dPlanet)
			}
		}
	}
	moon := o.c(mapmodel.InkMoon, iLit, 0)
	if mx, my := W*5/6, max(0, b/4); my < b {
		o.fb.fg(mx, o.Y(my), '●', moon, dMoon)
	}
	if b > 3 {
		x := int(hash(o.anim/40, 3) % uint64(max(1, W)))
		y := b + 1 - (o.anim%40)*(b+1)/40
		if y >= 0 && y < S {
			o.fb.fg(x, o.Y(y), mapmodel.R(mapmodel.SymShuttle, o.tier), o.c(mapmodel.InkFrame, iLit, 0), dLane2)
		}
	}
}

// compactDeep: the far sun, a breath of nebula and the warp gate standing
// over the present.
func (o *orb) compactDeep() {
	W, b := o.W, o.groundY
	void := o.c(mapmodel.InkVoid, iBack, 0)
	vio, teal := o.c(mapmodel.InkCloud, iBack, 0), o.c(mapmodel.InkCloud2, iBack, 0)
	for y := 0; y < b; y++ {
		for x := 0; x < W; x++ {
			n := mapmodel.Noise(int64(o.m.Seed%89)+9, (float64(x)+float64(o.anim)*0.02)/9, float64(y)/2.4)
			if n > 0.62 {
				c := vio
				if (x/9)%2 == 1 {
					c = teal
				}
				bg := theme.Mix(void, c, (n-0.62)*1.1)
				if cl := o.fb.at(x, o.Y(y)); cl != nil {
					cl.bg = bg
					if cl.ch == ' ' {
						cl.fg = bg
					}
				}
			}
		}
	}
	if b >= 3 {
		sx, sy := max(1, W/8), max(0, b/4)
		o.fb.fg(sx, o.Y(sy), '●', o.c(mapmodel.InkAccent, iEmit, 0), dOrrSun)
		o.fb.fg(sx-1, o.Y(sy), '·', o.c(mapmodel.InkAccent3, iEmit, 1), dOrrery)
		o.fb.fg(sx+1, o.Y(sy), '·', o.c(mapmodel.InkAccent3, iEmit, 1), dOrrery)
	}
	// the gate, at the present's end of the panorama
	frac, done := gateFrac(o.m)
	gx := W - max(4, W/7)
	R := math.Max(1.5, math.Min(float64(b)*0.42, 4))
	cy := float64(b) - R - 0.5
	plate := o.c(mapmodel.InkFrame, iLit, 0)
	scaf := o.c(mapmodel.InkFrameDim, iLit, 0)
	field := o.c(mapmodel.InkAccent2, iEmit, 1)
	const dg = dTag + 2 // the gate stands in front of the profile
	steps := int(R * 16)
	for i := 0; i < steps; i++ {
		t := float64(i) / float64(steps)
		from := math.Abs(t - 0.75)
		if from > 0.5 {
			from = 1 - from
		}
		x := int(math.Round(float64(gx) + 2*R*mapmodel.Cos(t)))
		y := int(math.Round(cy - R*mapmodel.Sin(t)))
		if y < 0 || y >= b {
			continue
		}
		if from <= frac/2 {
			o.fb.fg(x, o.Y(y), '●', plate, dg)
		} else if i%2 == 0 {
			o.fb.fg(x, o.Y(y), '·', scaf, dg)
		}
	}
	if done {
		for y := int(cy - R + 1); y <= int(cy+R-1); y++ {
			for x := gx - int(2*R) + 2; x <= gx+int(2*R)-2; x++ {
				dx, dy := float64(x-gx)/(2*R), (float64(y)-cy)/R
				if float64(dx*dx)+float64(dy*dy) < 0.55 && y >= 0 && y < b && (x+y+o.anim/4)%3 == 0 {
					o.fb.fg(x, o.Y(y), '░', field, dg)
				}
			}
		}
	}
}

// compactGalaxy: the starbase standing over the present, a warp streak now
// and then, and the saucers that are traffic now.
func (o *orb) compactGalaxy() {
	W, b := o.W, o.groundY
	white := o.c(mapmodel.InkFrame, iLit, 0)
	gold := o.c(mapmodel.InkAccent, iLit, 0)
	cyan := o.c(mapmodel.InkAccent2, iEmit, 0)
	hx := W - max(5, W/6)
	const dh = dTag + 2 // the starbase stands in front of the profile
	if b >= 4 {
		top := 0
		for y := top; y < b; y++ {
			o.fb.fg(hx, o.Y(y), '│', white, dh)
		}
		dy := max(1, b*2/5)
		for x := hx - 5; x <= hx+5; x++ {
			o.fb.fg(x, o.Y(dy), '▬', gold, dh)
		}
		o.fb.fg(hx-3, o.Y(dy-1), '╱', white, dh)
		o.fb.fg(hx+3, o.Y(dy-1), '╲', white, dh)
		o.fb.fg(hx-1, o.Y(dy-1), '▄', white, dh)
		o.fb.fg(hx+1, o.Y(dy-1), '▄', white, dh)
		o.fb.fg(hx, o.Y(dy), '◘', cyan, dh)
	}
	if (o.anim/4)%10 < 3 && b > 2 {
		y := 1 + int(hash(o.anim/40, 7)%uint64(max(1, b-2)))
		x0 := int(hash(o.anim/40, 9) % uint64(max(1, W/2)))
		streak := o.c(mapmodel.InkGlow, iEmit, 0)
		for k := 0; k < W/3; k++ {
			o.fb.fg(x0+k, o.Y(y), '━', theme.Mix(streak, o.c(mapmodel.InkVoid, iBack, 0), float64(k)/float64(W/3+1)), dTrail)
		}
	}
	if b > 3 {
		for i, a := range mapmodel.Aliens[:min(2, len(mapmodel.Aliens))] {
			x := (int(hash(i, 51)%uint64(max(1, W))) + o.anim/(4+i)) % max(1, W)
			y := 1 + int(hash(i, 53)%uint64(max(1, b-3)))
			c := o.raw(a.Hue, iEmit, 0)
			o.fb.fg(x, o.Y(y), mapmodel.R(mapmodel.SymUFO, o.tier), c, dLane)
		}
	}
}

// compactMandala: a small mandala breathing over the baseline.
func (o *orb) compactMandala() {
	W, b := o.W, o.groundY
	if b < 3 {
		return
	}
	cx, cy := float64(W)/2, float64(b-1)*0.5
	R := math.Max(1, math.Min(float64(b-1)*0.48, float64(W)/5))
	gold := o.c(mapmodel.InkAccent, iEmit, 0)
	dim := o.c(mapmodel.InkFrameDim, iEmit, 0)
	void := o.c(mapmodel.InkVoid, iBack, 0)
	for k := 0; k < 3; k++ {
		r := R * float64(k+1) / 3
		br := o.breath(k * 2)
		n := max(6, int(r*6))
		for i := 0; i < n; i++ {
			t := float64(i)/float64(n) + float64(o.anim)/float64(2400+600*k)
			x := int(math.Round(cx + 2*r*mapmodel.Cos(t)))
			y := int(math.Round(cy - r*mapmodel.Sin(t)))
			c := theme.Mix(theme.Mix(void, dim, 0.7), gold, 0.15+0.35*br)
			ch := '·'
			if i%max(1, n/6) == 0 {
				ch, c = '◆', theme.Mix(theme.Mix(void, gold, 0.7), gold, br)
			}
			if y >= 0 && y < b {
				o.fb.fg(x, o.Y(y), ch, c, dMandala)
			}
		}
	}
	o.fb.fg(int(cx), o.Y(int(math.Round(cy))), '✦', o.c(mapmodel.InkGlow, iEmit, 0), dMandala)
}

// compactFringes is the Quantum's interference under the baseline,
// squeezed.
func (o *orb) compactFringes() {
	b := o.groundY
	ax, bx := float64(o.W)*0.35, float64(o.W)*0.65
	c1, c2 := o.c(mapmodel.InkCloud, iEmit, 1), o.c(mapmodel.InkCloud2, iEmit, 0)
	for y := b + 1; y < o.S; y++ {
		fy := float64(2 * (y - b))
		for x := 0; x < o.W; x++ {
			da := math.Sqrt(float64((float64(x)-ax)*(float64(x)-ax)) + float64(fy*fy))
			db := math.Sqrt(float64((float64(x)-bx)*(float64(x)-bx)) + float64(fy*fy))
			switch v := mapmodel.Cos((da-db)/3.5 - float64(o.anim)/36); {
			case v > 0.75:
				o.fb.fg(x, o.Y(y), '▒', c2, dFringe)
			case v > 0.4:
				o.fb.fg(x, o.Y(y), '░', c1, dFringe)
			}
		}
	}
}

// compactReflect mirrors the profile under the mirror line.
func (o *orb) compactReflect() {
	b := o.groundY
	void := o.c(mapmodel.InkVoid, iBack, 0)
	for k := 1; b+k < o.S && b-k >= 0; k++ {
		for x := 0; x < o.W; x++ {
			src := o.fb.at(x, o.Y(b-k))
			if src == nil || src.d != dRow0 {
				continue
			}
			c := theme.Mix(src.show(), void, 0.6+0.1*float64(k))
			o.fb.set(x, o.Y(b+k), ' ', c, c, dReflect)
		}
	}
}
