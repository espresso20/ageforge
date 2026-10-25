package skyline

import (
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
	"github.com/espresso20/ageforge/vibe"
)

// compact.go is the glanceable form for a sidebar (40x15). The whole city is
// squeezed into a sparkline skyline: eighth-block tops give each column
// eight steps of height, the facade wears the tallest building's material,
// and the full view's signals ride on it: lit windows (staffed), smoke
// (producing), ▼ (new), wonders, the harbinger on the horizon, fire in a
// catastrophe and a few vehicles on the road, under a one-line header.

const compactGround = 42 // the layout the compact view measures heights in

func (v *view) DrawCompact(scr tcell.Screen, r mapstyle.Rect, f mapstyle.Frame) {
	if r.W <= 0 || r.H <= 0 {
		return
	}
	if f.Model == nil {
		blank(scr, r)
		return
	}
	v.composeCompact(f, r.W, r.H)
	v.fb.blit(scr, r, f.Tier)
}

func (v *view) composeCompact(f mapstyle.Frame, W, H int) {
	if sky := f.Model.Sky(); sky != mapmodel.SkyGround {
		v.composeSkyCompact(f, W, H, sky) // the sky arc: orbit_compact.go
		return
	}
	m := f.Model
	lay := v.clay
	if lay == nil || lay.m != m {
		keep := v.lay
		v.lay = nil
		lay = v.layoutFor(m, compactGround)
		v.clay, v.lay = lay, keep
	}
	p, mp := v.palettes(m)
	v.fb.reset(W, H)
	s := &scene{v: v, m: m, lay: lay, p: p, mp: mp, fb: &v.fb, tier: f.Tier, anim: f.Anim, visit: f.VisitFrame(), W: W, H: H,
		S: H - 1, top: 1, groundY: H - 2, sel: -1, band: bandOf(m.AgeIdx)}
	s.city, s.inCity = mapmodel.CityLookAt(m.AgeIdx)
	bg := chromeBg()
	if H >= 2 {
		s.compactScene()
	}
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
		parts = append(parts, seg{" " + strconv.Itoa(n) + vibe.Line("ui.mapstyle.skyline.compact.composeCompact.01", " idle"), theme.RoleWarning})
	}
	if m.Harbinger != nil {
		parts = append(parts, seg{" " + string(mapmodel.R(mapmodel.SymHarbinger, f.Tier)), theme.RoleWarning})
	}
	if m.Catastrophe.Pending != "" {
		parts = append(parts, seg{" " + string(mapmodel.R(mapmodel.SymFire, f.Tier)), theme.RoleNegative})
	}
	s.segs(x, 0, bg, parts)
}

// compactSpan is the built stretch of the panorama the compact view shows.
func compactSpan(m *mapmodel.Model) (int, int) {
	if len(m.Skyline.Districts) == 0 {
		return 0, max(1, m.Skyline.Width)
	}
	a := max(0, m.Skyline.Districts[0].X0-2)
	return a, max(a+1, m.Skyline.FrontierX+8)
}

func (s *scene) compactScene() {
	m, p, lay := s.m, s.p, s.lay
	W := s.W
	S := s.groundY // sky and city rows: scene rows 0..S-1; the road is row S
	for y := 0; y < S; y++ {
		c := p.skyAt(y, S)
		for x := 0; x < W; x++ {
			s.fb.fill(x, s.Y(y), c, dSky)
		}
	}
	if p.stars > 0.2 {
		star := p.hue(theme.SkyStarWarm, mFlat, 0)
		for y := 0; y < S/2; y++ {
			for x := 0; x < W; x++ {
				if hashf(x, y, 991) < 0.03*p.stars {
					s.fb.fg(x, s.Y(y), '·', star, dStar)
				}
			}
		}
	}
	if S >= 3 {
		sx := int(float64(W) * (0.1 + 0.8*math.Mod(m.Clock.TOD+0.27, 0.5)/0.5))
		if p.day > 0.3 {
			s.fb.fg(sx, s.Y(1), '●', p.hue(theme.SkySun, mFlat, 0), dSkyObj)
		} else {
			s.fb.fg(sx, s.Y(1), mapmodel.R(mapmodel.SymMoon, s.tier), p.hue(theme.SkyMoon, mFlat, 0), dSkyObj)
		}
	}
	ridge := p.hill(theme.SkyRidge, 0.45)
	ridgeTop := func(x int) float64 {
		n := fbm1(float64(x)/9, 300+int(m.Seed%1000), 3)
		return float64(S) - 1 - n*float64(S)*0.35
	}
	for x := 0; x < W; x++ {
		top := ridgeTop(x)
		for y := int(math.Ceil(top)); y < S; y++ {
			s.fb.fill(x, s.Y(y), ridge, dRidge)
		}
		if fl := math.Floor(top); top-fl > 0.01 && fl >= 0 {
			s.fb.fg(x, s.Y(int(fl)), eighths[clampInt(int((math.Ceil(top)-top)*8), 1, 7)], ridge, dRidge)
		}
	}
	if h := m.Harbinger; h != nil && S >= 2 {
		// a lone figure standing on the ridge, not hanging in the sky
		hx := W/6 + int(hash(int(mapmodel.HashStr(h.Key)))%uint64(max(1, W/2)))
		hy := clampInt(int(math.Ceil(ridgeTop(hx)))-1, 0, S-1)
		s.fb.fg(hx, s.Y(hy), mapmodel.R(mapmodel.SymHarbinger, s.tier), s.mp.Fg[mapmodel.CIdle], dHarbinger)
	}
	a0, a1 := compactSpan(m)
	// Each strip column is the tallest building in the stretch of town it
	// stands for. The tallest of all would set the scale for the rest (one
	// tower or wonder flattens a town into a band), so the scale is held to
	// the tall end of the ordinary columns: those fill most of the strip and a
	// landmark reaches its top.
	type colInfo struct {
		h     int
		lot   int32
		isNew bool
	}
	cols := make([]colInfo, W)
	var hs []int
	for x := 0; x < W; x++ {
		a := a0 + x*(a1-a0)/W
		b := max(a+1, a0+(x+1)*(a1-a0)/W)
		ci := colInfo{lot: -1}
		for i := a; i < b && i < len(lay.prof); i++ {
			if lay.prof[i] > ci.h {
				ci.h, ci.lot = lay.prof[i], lay.tallest[i]
			}
			if t := lay.tallest[i]; t >= 0 && lay.lots[t].ml.New {
				ci.isNew = true
			}
		}
		cols[x] = ci
		if ci.h > 0 && ci.lot >= 0 {
			hs = append(hs, ci.h)
		}
	}
	ref := 1
	if len(hs) > 0 {
		sort.Ints(hs)
		// the tallest sets the scale, but no more than half again the tall
		// end of the ordinary columns: a landmark reaches the top and the
		// rest still fill the strip
		ref = max(1, min(hs[len(hs)-1], hs[len(hs)*85/100]*3/2))
	}
	top := float64(max(1, S-1))
	fire := m.Catastrophe.Pending != ""
	win := p.hue(theme.SkyWinWarm, mEmit, 0)
	if s.inCity { // the age's own light in the windows (city.go)
		win = s.compactWindow()
	}
	fresh := s.mp.Fg[mapmodel.CFresh]
	gold := p.hue(theme.SkyWonderGold, mEmit, 0)
	for x := 0; x < W; x++ {
		ci := cols[x]
		if ci.h == 0 || ci.lot < 0 {
			continue
		}
		a := a0 + x*(a1-a0)/W
		b := max(a+1, a0+(x+1)*(a1-a0)/W)
		lv := &lay.lots[ci.lot]
		wall := p.col(lv.fam, lv.vr, sWall, 0)
		hh := math.Min(float64(ci.h)/float64(ref)*top, float64(max(1, S-1)))
		full := int(hh)
		for k := 0; k < full; k++ {
			y := S - 1 - k
			s.fb.fill(x, s.Y(y), wall, dRow0)
			if p.night > 0.35 && k < full-1 && lv.staff > 0 && hashf(x, y, int(m.Seed)) < 0.1+0.6*lv.staff {
				s.fb.set(x, s.Y(y), '·', win, wall, dRow0)
			}
		}
		if f := hh - float64(full); f > 0.06 || full == 0 {
			s.fb.fg(x, s.Y(S-1-full), eighths[clampInt(int(f*8), 1, 7)], wall, dRow0)
		}
		// What the building says about itself is drawn in its own top cell,
		// never above it: nothing floats over the skyline.
		ty := S - 1 - full
		if f := hh - float64(full); f <= 0.06 && full > 0 {
			ty = S - full
		}
		switch {
		case s.v.changes && ci.isNew:
			s.stripCap(x, ty, '▼', fresh)
		case lv.ml.Wonder && lv.ml.X >= a && lv.ml.X < b: // one star, on the wonder's own column
			s.stripCap(x, ty, mapmodel.R(mapmodel.SymWonder, s.tier), gold)
		case fire && hash(x, 5)%5 == 0:
			k := int(hash(x, s.anim/2) % 4)
			s.stripCap(x, ty, flameGlyphs[k], p.hue(flameHues[k], mEmit, 0))
		}
	}
	s.compactCity(a0, a1)
	// the road, with the traffic on it
	rh, mh := roadHues(m.AgeIdx)
	road := p.hue(rh, mLit, 0)
	for x := 0; x < W; x++ {
		s.fb.set(x, s.Y(S), ' ', p.hue(mh, mLit, 0), road, dGround)
	}
	c := trafficCounts(m)
	for i := 0; i < c.route+min(3, c.foot/3)+c.war; i++ {
		x := (int(hash(i, 23)%uint64(W)) + s.anim/(3+i%3)) % W
		sym, h := mapmodel.SymCaravan, theme.SkyTradeGold
		switch {
		case i >= c.route+min(3, c.foot/3):
			sym, h = mapmodel.SymRaider, theme.SkyWarRed
		case i >= c.route:
			sym, h = mapmodel.SymWorker, theme.SkyInk
		}
		ch := mapmodel.R(sym, s.tier)
		switch {
		case s.tier != mapmodel.TierNerd && sym == mapmodel.SymCaravan:
			ch = '▪'
		case sym == mapmodel.SymWorker: // a person is a speck at this scale, not a white face as big as a house
			ch = '·'
		}
		s.fb.fg(x, s.Y(S), ch, theme.Legible(p.hue(h, mEmit, 0), road, 2), dStreet)
	}
	if sym := compactSkySym(s.band, m.AgeIdx); sym != mapmodel.SymNone && S >= 4 {
		for i := 0; i < 1+c.private/4; i++ {
			x := (int(hash(i, 29)%uint64(W)) + s.anim/(2+i%2)) % W
			y := 1 + int(hash(i, 31)%uint64(max(1, S/2)))
			s.fb.fg(x, s.Y(y), mapmodel.R(sym, s.tier), p.hue(theme.SkyAircraft, mEmit, 0), dAirLow)
		}
	}
}

// compactSky is what flies over the compact view from band 5 on, with the
// mover each stands for: planes, drones, hovercars, then starships.
var compactSky = [4]struct {
	sym   mapmodel.Sym
	mover mapmodel.Mover
}{
	{mapmodel.SymPlane, mapmodel.MoverPlane}, {mapmodel.SymDrone, mapmodel.MoverDrone},
	{mapmodel.SymCar, mapmodel.MoverHovercar}, {mapmodel.SymRocket, mapmodel.MoverNone},
}

// compactSkySym is the compact view's sky marker for a band and age:
// band's own or, while its mover is not introduced yet, the band below's
// (SymNone before band 5 and before the first plane).
func compactSkySym(band, age int) mapmodel.Sym {
	for i := min(3, band-5); i >= 0; i-- {
		if c := compactSky[i]; c.mover == mapmodel.MoverNone || mapmodel.Introduced(c.mover, age) {
			return c.sym
		}
	}
	return mapmodel.SymNone
}

// stripCap draws a mark in the top cell of a strip column: the glyph in its
// colour over what that cell shows, so it is the top of the skyline and not
// something hung above it.
func (s *scene) stripCap(x, y int, ch rune, fg tcell.Color) {
	if y < 0 {
		return
	}
	bg := s.fb.showAt(x, s.Y(y))
	s.fb.set(x, s.Y(y), ch, theme.Legible(fg, bg, 3), bg, dTag)
}
