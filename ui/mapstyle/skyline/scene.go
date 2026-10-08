package skyline

import (
	"math"
	"sort"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
)

// scene.go composes one frame, back to front: sky, sky objects, clouds, the
// far ridge with the civ towns and the harbinger, the near hills and the
// distant town, the three rows of real buildings, the ground and the bay,
// traffic, smoke, weather, the catastrophe overlay, markers and labels.

type emitter struct {
	x, y     int // screen column, scene row
	strength float64
	heavy    bool
	seed     int
	depth    uint8
}

type label struct {
	x, y int
	text string
	fg   tcell.Color
}

type scene struct {
	v       *view
	m       *mapmodel.Model
	lay     *layout
	p       *pal
	mp      *mapstyle.Palette
	fb      *fb
	tier    mapmodel.GlyphTier
	anim    int
	visit   int // the frame the rare visitor is drawn at (Frame.VisitFrame)
	W, H    int
	S       int // scene rows
	top     int // screen row of scene row 0
	groundY int
	cam     int
	band    int // traffic band of the current age
	gridOut bool
	farTops []float64
	emit    []emitter
	labels  []label
	vis     []int
	sel     int // the lot under the cursor, -1
	ridge   []ridgeItem
	// city is the Earth arc's look for this frame (city.go), inCity
	// whether the age has one.
	city   mapmodel.CityLook
	inCity bool
}

// Y converts a scene row to a frame row.
func (s *scene) Y(y int) int { return y + s.top }

func (s *scene) wx(x int) int { return x + s.cam }

// ------------------------------------------------------------------ sky

var dither = [4]rune{' ', '░', '▒', '▓'}

func (s *scene) sky() {
	gy := s.groundY
	const bands = 7
	for y := 0; y < gy; y++ {
		f := float64(y) / float64(max(1, gy-1)) * bands
		b := int(f)
		t := f - float64(b)
		a := s.p.skyAt(b*(gy-1)/bands, gy)
		c := s.p.skyAt((b+1)*(gy-1)/bands, gy)
		di := 0
		if t >= 0.52 {
			di = min(3, 1+int((t-0.52)/0.48*3))
		}
		for x := 0; x < s.W; x++ {
			s.fb.set(x, s.Y(y), dither[di], c, a, dSky)
		}
	}
}

func (s *scene) celestial() {
	p := s.p
	horizon := float64(s.groundY - 2)
	if p.stars > 0.05 {
		warm, cool := theme.SkyColor(theme.SkyStarWarm), theme.SkyColor(theme.SkyStarCool)
		for y := 0; y < int(horizon*0.8); y++ {
			lim := 0.022 * p.stars * (1 - float64(y)/horizon)
			for x := 0; x < s.W; x++ {
				hv := hashf(x, y, 991)
				if hv > lim {
					continue
				}
				tw := hash(x, y, s.anim/6) % 9
				ch := '·'
				switch {
				case hv < 0.0015:
					ch = '*'
				case hv < 0.004:
					ch = '+'
				case tw == 0:
					ch = '∙'
				}
				c := theme.Mix(warm, cool, hashf(x, y, 5))
				if tw == 1 {
					c = theme.Shade(c, 0.6)
				}
				if cl := s.fb.at(x, s.Y(y)); cl != nil {
					s.fb.fg(x, s.Y(y), ch, p.final(theme.Mix(cl.bg, c, 0.4+0.6*p.stars)), dStar)
				}
			}
		}
	}
	arc := func(u float64) (int, int) {
		return int(float64(s.W) * (0.08 + 0.84*u)), int(horizon - mapmodel.Sin(u/2)*horizon*0.85)
	}
	t := s.m.Clock.TOD
	if u := (t - 0.23) / 0.54; u >= 0 && u <= 1 {
		x, y := arc(u)
		low := 1 - mapmodel.Sin(u/2)
		c := p.final(theme.Mix(theme.SkyColor(theme.SkySun), theme.SkyColor(theme.SkySunLow), low*low))
		if p.gloom { // a pale disc behind the smog
			c = theme.Mix(c, p.skyAt(max(0, y), s.groundY), 0.75)
		}
		for dx := -1; dx <= 1; dx++ {
			s.fb.fg(x+dx, s.Y(y-1), '▄', c, dSkyObj)
			s.fb.fill(x+dx, s.Y(y), c, dSkyObj)
			s.fb.fg(x+dx, s.Y(y+1), '▀', c, dSkyObj)
		}
		s.fb.fg(x-2, s.Y(y), '▐', c, dSkyObj)
		s.fb.fg(x+2, s.Y(y), '▌', c, dSkyObj)
	}
	mt := t + 0.5
	if mt >= 1 {
		mt--
	}
	if u := (mt - 0.23) / 0.54; u >= 0 && u <= 1 {
		if x, y := arc(u); y >= 1 {
			h := theme.SkyMoon
			if p.light {
				h = theme.SkyMoonPale
			}
			c := p.hue(h, mFlat, 0)
			s.fb.fg(x-1, s.Y(y), '▄', c, dSkyObj)
			s.fb.fg(x, s.Y(y), '█', c, dSkyObj)
			s.fb.fg(x+1, s.Y(y), '▄', c, dSkyObj)
			s.fb.fg(x-1, s.Y(y+1), '▀', c, dSkyObj)
			s.fb.fg(x, s.Y(y+1), '█', c, dSkyObj)
			s.fb.fg(x+1, s.Y(y+1), '▀', c, dSkyObj)
		}
	}
}

// skyStructures: the ringed planet and the orbital ring over the late city,
// and the space elevator's tether.
func (s *scene) skyStructures() {
	a := s.m.AgeIdx
	gy := s.groundY
	if a >= 18 {
		cx, cy := s.W*3/4-int(float64(s.cam)*0.02), 3
		pc := s.p.final(theme.Mix(theme.SkyColor(theme.SkyPlanet), s.p.sky[0], 0.3))
		rc := s.p.final(theme.Mix(theme.SkyColor(theme.SkyPlanetRing), s.p.sky[0], 0.3))
		for dy := -1; dy <= 1; dy++ {
			for dx := -3; dx <= 3; dx++ {
				if float64(dx*dx)/4+float64(dy*dy) <= 2.56 {
					s.fb.fill(cx+dx, s.Y(cy+dy), pc, dSkyObj)
				}
			}
		}
		for dx := -2; dx <= 0; dx++ {
			s.fb.fg(cx+dx, s.Y(cy-2), '▄', pc, dSkyObj)
		}
		for dx := -6; dx <= 6; dx++ {
			if dx >= -3 && dx <= 3 {
				s.fb.set(cx+dx, s.Y(cy), '▀', rc, pc, dSkyObj)
			} else {
				s.fb.fg(cx+dx, s.Y(cy), '─', rc, dSkyObj)
			}
		}
	}
	if a >= 17 {
		rc := s.p.final(theme.Mix(theme.SkyColor(theme.SkyOrbitalRing), s.p.sky[1], 0.35))
		lc := s.p.hue(theme.SkyOrbitalLight, mEmit, 3)
		for x := 0; x < s.W; x++ {
			fy, y := s.orbitY(x)
			if y >= gy/2 {
				continue
			}
			ch := '▀'
			if fy-float64(y) >= 0.5 {
				ch = '▄'
			}
			s.fb.fg(x, s.Y(y), ch, rc, dSkyObj)
			if s.p.night > 0.4 && (x+s.anim/4)%9 == 0 {
				s.fb.fg(x, s.Y(y), '•', lc, dSkyObj)
			}
		}
	}
	if x := s.elevatorX(); x >= 0 && x < s.W { // the space elevator, from the Fusion Age
		tc := s.p.col(9, 0, sMetal, 2)
		if s.fusion() { // newly strung: a beam of plasma light off the top of the screen
			tc = s.p.cityHue(theme.CityTether, mEmit, 1)
			glow := s.p.cityHue(theme.CityElectric, mEmit, 1)
			for y := 0; y < gy-1; y++ {
				s.fb.tint(x-1, s.Y(y), glow, 0.18, dRow2+1)
				s.fb.tint(x+1, s.Y(y), glow, 0.18, dRow2+1)
			}
		}
		for y := 0; y < gy-1; y++ {
			s.fb.fg(x, s.Y(y), '│', tc, dRow2)
		}
		s.fb.set(x-1, s.Y(gy-1), '▄', tc, s.fb.showAt(x-1, s.Y(gy-1)), dRow2)
		s.fb.set(x+1, s.Y(gy-1), '▄', tc, s.fb.showAt(x+1, s.Y(gy-1)), dRow2)
		s.fb.fill(x, s.Y(gy-1), tc, dRow2)
	}
}

// orbitY is the orbital ring's row at screen column x.
func (s *scene) orbitY(x int) (float64, int) {
	cx := float64(s.W)/2 - float64(s.cam)*0.03
	d := (float64(x) - cx) / float64(s.W)
	fy := 1.5 + float64(d*d)*float64(s.groundY)*0.9
	return fy, int(fy)
}

// elevatorX is the screen column of the space elevator, or -1.
func (s *scene) elevatorX() int {
	if x := tetherX(s.m); x >= 0 {
		return x - s.cam
	}
	return -1
}

// tetherX is the world column of the space elevator, or -1: it rises from
// the Fusion district, east of its middle, in the Fusion Age and stays
// there through the Space Age, where the sky arc (orbit_space.go) carries
// it from the planet up to the station. From the Interstellar Age the map
// is out in deep space, with no planet to anchor it.
func tetherX(m *mapmodel.Model) int {
	fusion, ok := m.Catalog.AgeIdx["fusion_age"]
	if !ok || m.AgeIdx < fusion || len(m.Skyline.Districts) <= fusion || m.Sky() > mapmodel.SkyOrbit {
		return -1
	}
	d := m.Skyline.Districts[fusion]
	return d.X0 + d.LandW*3/4
}

func (s *scene) clouds() {
	n := [...]int{3, 6, 9, 11, 8, 7}[s.m.Weather.Kind]
	if s.m.AgeIdx >= 18 {
		n /= 2
	}
	span := s.W + 60
	top := max(1, s.groundY/2)
	base := theme.Tint(theme.SkyColor(theme.SkyCloud), s.p.lightTint)
	if s.p.twilight > 0.3 {
		base = theme.Mix(base, theme.SkyColor(theme.SkyCloudDusk), float64(s.p.twilight*0.6))
	}
	if k := s.m.Weather.Kind; k == mapmodel.Rain || k == mapmodel.Storm {
		base = theme.Shade(base, 0.7)
	}
	if s.p.gloom { // the megacity's clouds are smog
		base = theme.Mix(base, theme.CityColor(theme.CitySmog), 0.6)
		base = theme.Shade(base, 0.55)
	} else if s.inCity && s.city.Smog > 0 { // and the data age's are browning
		base = theme.Mix(base, theme.CityColor(theme.CitySmog), float64(0.8*s.city.Smog))
	}
	c, under := s.p.final(base), s.p.final(theme.Shade(base, 0.78))
	for i := 0; i < n; i++ {
		cw := 12 + int(hash(i, 41)%18)
		ch := 1 + int(hash(i, 43)%2)
		speed := 0.03 + hashf(i, 47)*0.05
		x0 := int(float64(hash(i, 49)%uint64(span))+float64(s.anim)*speed-float64(s.cam)*0.08) % span
		if x0 < 0 {
			x0 += span
		}
		x0 -= 30
		y0 := 1 + int(hash(i, 53)%uint64(top))
		for dx := 0; dx < cw; dx++ {
			u := float64(dx) / float64(cw-1)
			bump := mapmodel.Sin(u/2) * (0.9 + 0.8*noise1(float64(dx)*0.5, i))
			hgt := bump * (1.1 + float64(ch)*0.55)
			x := x0 + dx
			for k := 0; k < int(math.Ceil(hgt)); k++ {
				y := y0 + ch - k
				if y < 0 || y >= s.groundY {
					continue
				}
				col := c
				if k == 0 {
					col = under
				}
				if hgt-float64(k) < 0.5 {
					s.fb.fg(x, s.Y(y), '▄', col, dCloud)
				} else {
					s.fb.fill(x, s.Y(y), col, dCloud)
				}
			}
			if y := y0 + ch + 1; bump > 0.2 && y < s.groundY && hash(x, i, 3)%2 == 0 {
				s.fb.fg(x, s.Y(y), '▀', under, dCloud)
			}
		}
	}
}

// ------------------------------------------------------------ landscape

func noise1(x float64, seed int) float64 { return mapmodel.Noise(int64(seed), x, 0.5) }

func fbm1(x float64, seed, oct int) float64 {
	v, amp, tot := 0.0, 1.0, 0.0
	for o := 0; o < oct; o++ {
		v += noise1(x, seed+o*101) * amp
		tot += amp
		x *= 2.03
		amp *= 0.5
	}
	return v / tot
}

func (s *scene) ridgeTops(layer int, par, amp, freq float64) []float64 {
	h := make([]float64, s.W)
	seed := 300 + layer*77 + int(s.m.Seed%1000)
	for x := range h {
		n := fbm1((float64(x)+float64(s.cam)*par)*freq, seed, 4)
		h[x] = float64(s.groundY-1) - n*math.Sqrt(n)*amp
	}
	return h
}

func (s *scene) paintRidge(tops []float64, col tcell.Color, d uint8) {
	for x, t := range tops {
		ti := int(math.Floor(t))
		for y := max(0, ti); y < s.groundY; y++ {
			s.fb.fill(x, s.Y(y), col, d)
		}
		if t-float64(ti) < 0.5 && ti-1 >= 0 {
			s.fb.fg(x, s.Y(ti-1), '▄', col, d)
		}
	}
}

func (s *scene) farRidge() {
	if s.megacity() { // the megacity has built the ridge over
		s.megacityRidge()
		s.ridgeTowns()
		s.harbinger()
		return
	}
	amp := float64(s.groundY) * 0.42
	tops := s.ridgeTops(0, 0.15, amp, 1.0/30)
	h := theme.SkyRidge
	if s.m.AgeIdx >= 18 {
		h = theme.SkyRidgeCosmic
	}
	col := s.p.hill(h, 0.42)
	if s.inCity && s.city.Smog > 0 { // the outskirts under the smog
		col = theme.Mix(col, s.p.cityHue(theme.CitySmog, mLit, 3), float64(0.7*s.city.Smog))
	}
	s.paintRidge(tops, col, dRidge)
	snow := s.p.hill(theme.SkySnowCap, 0.3)
	for x, t := range tops {
		if t < float64(s.groundY)-amp*0.72 {
			s.fb.set(x, s.Y(int(t)), '▀', snow, col, dRidge)
		}
	}
	s.farTops = tops
	s.ridgeTowns()
	s.harbinger()
}

func (s *scene) ridgeTop(x int) int {
	if len(s.farTops) == 0 {
		return s.groundY - 2
	}
	return int(s.farTops[clampInt(x, 0, len(s.farTops)-1)])
}

// ridgeTowns: every civ you have met stands on the far ridge as a small
// town, sized by its strength, its pennant in the colour of your relations.
func (s *scene) ridgeTowns() {
	col := s.p.hill(theme.SkyRidgeTown, 0.3)
	win := s.p.hue(theme.SkyWinWarm, mEmit, 4)
	for _, it := range s.ridge {
		if it.fac == nil {
			continue
		}
		f := it.fac
		w := townWidth(f)
		x0 := it.x - w/2
		for dx := 0; dx < w; dx++ {
			x := x0 + dx
			if x < 0 || x >= s.W {
				continue
			}
			top := s.ridgeTop(x)
			hgt := 1 + int(hash(it.i, dx, 3)%uint64(1+clampInt(f.Strength, 1, 5)))
			for y := top - hgt; y < top; y++ {
				if y >= 0 {
					s.fb.fill(x, s.Y(y), col, dRidgeTown)
				}
			}
			if s.p.night > 0.4 && hash(it.i, dx)%2 == 0 && top-1 >= 0 {
				s.fb.set(x, s.Y(top-1), '·', win, col, dRidgeTown)
			}
		}
		py := s.ridgeTop(it.x) - 5
		rc := s.mp.Fg[mapmodel.RelationClass(f.Relation)]
		s.fb.fg(it.x, s.Y(py+1), '│', col, dRidgeTown)
		s.fb.fg(it.x, s.Y(py+2), '│', col, dRidgeTown)
		s.fb.fg(it.x, s.Y(py+3), '│', col, dRidgeTown)
		flag := '►'
		if f.Relation == mapmodel.RelWar {
			flag = mapmodel.R(mapmodel.SymWar, s.tier)
		} else if s.tier == mapmodel.TierNerd {
			flag = mapmodel.R(mapmodel.SymCiv, s.tier)
		}
		s.fb.fg(it.x+1, s.Y(py+1), flag, rc, dRidgeTown)
		s.labels = append(s.labels, label{it.x - textLen(f.Name)/2, py, f.Name, rc})
	}
}

func townWidth(f *mapmodel.Faction) int { return clampInt(4+f.Strength*2, 5, 14) }

// harbinger: a lone figure on the far ridge, facing the city. From the
// Digital Era it is a flickering hologram.
func (s *scene) harbinger() {
	h := s.m.Harbinger
	if h == nil {
		return
	}
	var x int
	for _, it := range s.ridge {
		if it.fac == nil {
			x = it.x
		}
	}
	y := s.ridgeTop(x) - 1
	ink := s.p.final(theme.SkyColor(theme.SkyHarbinger))
	if s.p.light {
		ink = s.p.final(theme.SkyColor(theme.SkyHarbingerPale))
	}
	if s.m.AgeIdx >= 13 {
		if hash(s.anim/2, 3)%5 == 0 {
			ink = theme.Mix(ink, s.fb.showAt(clampInt(x, 0, s.W-1), s.Y(max(0, y))), 0.6)
		} else {
			ink = s.p.hue(theme.SkyHarbingerHolo, mEmit, 1)
		}
	}
	s.fb.fg(x, s.Y(y-2), '▄', ink, dHarbinger)
	s.fb.fg(x-1, s.Y(y-1), '▐', ink, dHarbinger)
	s.fb.fill(x, s.Y(y-1), ink, dHarbinger)
	s.fb.fg(x+1, s.Y(y-1), '▌', ink, dHarbinger)
	s.fb.fg(x-1, s.Y(y), '▐', ink, dHarbinger)
	s.fb.fill(x, s.Y(y), ink, dHarbinger)
	s.fb.fg(x+1, s.Y(y), '▌', ink, dHarbinger)
	for dy := 0; dy < 3; dy++ {
		s.fb.fg(x+2, s.Y(y-dy), '│', ink, dHarbinger)
	}
	if s.p.night > 0.4 {
		s.fb.fg(x, s.Y(y-2), '▄', s.p.hue(theme.SkyHarbingerEyes, mEmit, 2), dHarbinger)
	}
	if s.tier == mapmodel.TierNerd {
		s.fb.fg(x, s.Y(y-3), mapmodel.R(mapmodel.SymHarbinger, s.tier), s.mp.Fg[mapmodel.CDanger], dHarbinger)
	}
	s.labels = append(s.labels, label{x - textLen(h.Name)/2, y - 5, h.Name, s.mp.Fg[mapmodel.CIdle]})
}

func (s *scene) nearLayer() {
	if s.megacity() {
		s.megacityNear()
		s.midTown()
		return
	}
	a := s.m.AgeIdx
	tops := s.ridgeTops(1, 0.4, float64(s.groundY)*0.16, 1.0/18)
	h := theme.SkyHillYoung
	switch {
	case a > 11:
		h = theme.SkyHillModern
	case a > 7:
		h = theme.SkyHillIndustrial
	}
	hc := s.p.hill(h, 0.22)
	if s.inCity && s.city.Smog > 0 {
		hc = theme.Mix(hc, s.p.cityHue(theme.CitySmog, mLit, 2), float64(0.5*s.city.Smog))
	}
	s.paintRidge(tops, hc, dHills)
	if a <= 9 {
		tc := s.p.hill(theme.SkyForest, 0.15)
		for x, t := range tops {
			if hash(x+int(float64(s.cam)*0.4), 5)%3 == 0 && int(t)-1 > 0 {
				s.fb.fg(x, s.Y(int(t)-1), '▲', tc, dHills)
			}
		}
	}
	s.midTown()
}

// midTown is the distant town: flat silhouettes whose height follows the
// real skyline nearby, so a big economy casts a big shadow.
func (s *scene) midTown() {
	a := s.m.AgeIdx
	if a < 3 {
		return
	}
	prof := s.lay.prof
	const par = 0.6
	m := s.p.mats[familyOf(a)][0].Wall
	base := theme.Mix(m, theme.SkyColor(theme.SkyTownFar), 0.55)
	winc := s.p.hue(theme.SkyWinWarm, mEmit, 3)
	if a >= 15 {
		winc = s.p.hue(theme.SkyNeonMagenta, mEmit, 3)
	}
	neon := s.p.hue(theme.SkyNeonCyan, mEmit, 3)
	cols := [3]tcell.Color{}
	for i := range cols {
		cols[i] = s.p.final(theme.Mix(theme.Tint(theme.Shade(base, 0.9+0.1*float64(i)), s.p.lightTint), s.p.horizon, 0.45))
	}
	for x := 0; x < s.W; x++ {
		mx := float64(x) + float64(s.cam)*par
		cellN := int(math.Floor(mx / 8))
		split := 3 + int(hash(cellN, 7)%3)
		blk := cellN * 2
		off := int(mx) - cellN*8
		if off >= split {
			blk++
		}
		first, last := off == 0 || off == split, off == split-1 || off == 7
		wx := int(mx / par)
		dh := 0
		for k := -12; k <= 12; k++ {
			if i := wx + k*3; i >= 0 && i < len(prof) && prof[i] > dh {
				dh = prof[i]
			}
		}
		if dh == 0 {
			continue
		}
		col := cols[hash(blk, 5)%3]
		hgt := int(float64(dh) * (0.4 + 0.5*hashf(blk, 61)))
		if a <= 5 {
			hgt = int(float64(s.groundY)*0.12) + 2*b2i(blk%5 == 0)
		}
		top := s.groundY - hgt
		if last && hgt > 2 {
			col = theme.Shade(col, 0.85)
		}
		for y := top; y < s.groundY; y++ {
			s.fb.fill(x, s.Y(y), col, dTown)
		}
		switch {
		case a <= 5:
			if int(mx)%2 == 0 {
				s.fb.fg(x, s.Y(top-1), '▄', col, dTown)
			}
		case a <= 7:
			if !first && !last {
				s.fb.fg(x, s.Y(top-1), '▄', col, dTown)
			}
		case a <= 11:
			if first && hash(blk, 9)%2 == 0 {
				s.fb.fill(x, s.Y(top-1), col, dTown)
				s.fb.fg(x, s.Y(top-2), '▄', col, dTown)
			}
		default:
			if off == 1 && hash(blk, 9)%3 == 0 {
				s.fb.fg(x, s.Y(top-1), '│', col, dTown)
				s.fb.fg(x, s.Y(top-2), '│', col, dTown)
			}
			if a >= 15 && !first && !last && hash(blk, 11)%3 == 0 {
				s.fb.set(x, s.Y(top), '▀', neon, col, dTown)
			}
		}
		if s.megacity() { // the megacity never sleeps: neon and windows, day and night
			s.megacityGlow(x, top, blk, first, col)
			continue
		}
		if s.p.night > 0.4 && a > 5 && !last {
			for y := top + 1; y < s.groundY; y += 2 {
				if hash(int(mx), y, 71)%4 == 0 {
					s.fb.set(x, s.Y(y), '▪', winc, col, dTown)
				}
			}
		}
	}
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

// ------------------------------------------------------------------ city

// flagged reports a producer the flows overlay marks: idle or short of hands.
func flagged(lv *lotView) bool {
	return lv.b != nil && lv.b.Count > 0 && lv.b.Staffing >= 0 && lv.b.Staffing < 0.5
}

func (s *scene) lots() {
	p := s.p
	night := p.night
	s.vis = s.lay.visible(s.cam-12, s.cam+s.W+12, s.vis)
	white := theme.SkyColor(theme.SkyWhite)
	fade := s.inCity && mapmodel.Greenery(s.m.AgeIdx) < 1 // the greenery fade (city.go)
	screens := s.inCity && s.city.Key == "information_age"
	mega := s.megacity()
	cool := s.p.cityHue(theme.CityScreenCool, mEmit, 0)
	for _, li := range s.vis {
		lv := &s.lay.lots[li]
		spr := lv.spr
		x0 := lv.x0 - s.cam
		y0 := s.groundY - spr.h
		hz := lv.ml.Row
		sel := li == s.sel
		dim := s.v.flows && flagged(lv)
		flick := lv.fam == 8 && hash(int(lv.ml.Seed), s.anim/3)%9 == 0
		tint := s.districtTint(lv.ml.Age, hz)
		phase := 0.0
		if lv.fam == 10 {
			phase = 0.5 + 0.5*mapmodel.Sin(float64(s.anim)/48+hashf(int(lv.ml.Seed)))
		}
		for sy := 0; sy < spr.h; sy++ {
			Y := s.Y(y0 + sy)
			for sx := 0; sx < spr.w; sx++ {
				c := spr.c[sy*spr.w+sx]
				X := x0 + sx
				if c.ch == 0 || X < 0 || X >= s.W {
					continue
				}
				fg := p.col(lv.fam, lv.vr, c.fg, hz)
				if tint != 0 && !c.fg.emissive() {
					fg = theme.Mix(fg, tint, 0.32)
				}
				if mega && (c.fg.emissive() || c.fg == sGlass || c.fg == sGlassHi) {
					fg = s.neonFix(fg, hz)
				}
				ch := c.ch
				switch {
				case c.fg == sWin && night > 0.25 && s.windowLit(lv, sx, sy):
					h := theme.SkyWinWarm
					if k := hash(int(lv.ml.Seed), sx, sy); k%4 == 0 {
						h = theme.SkyWinPale
					} else if lv.ml.Age >= 15 && k%3 == 0 {
						h = theme.SkyWinNeon
					}
					lit := p.hue(h, mEmit, hz)
					if s.inCity { // the age's own light in the windows (city.go)
						lit = s.windowLight(lv, sx, sy, lit, hz)
					}
					fg = theme.Mix(fg, lit, math.Min(1, (night-0.25)*2))
				case screens && c.fg == sWin && lv.ml.Age >= 10 && hash(int(lv.ml.Seed), sx, sy, s.anim/3)%23 == 0:
					fg = cool // a screen flickers in a window
				case flick && (c.fg == sNeon1 || c.fg == sNeon2 || c.fg == sNeon3):
					fg = p.col(lv.fam, lv.vr, sWallDark, hz)
				case phase > 0 && c.fg == sGlow:
					fg = theme.Mix(fg, p.col(lv.fam, lv.vr, sWall, hz), phase*0.6)
				case c.fg == sWater:
					if (sx+s.anim/3)%4 == 0 {
						ch = '▀'
					}
				case fade && (c.fg == sLeaf || c.fg == sLeafDark || c.fg == sField1 || c.fg == sField2):
					fg = s.leafFade(lv, sx, sy, fg, hz)
				}
				if dim {
					fg = theme.Shade(fg, 0.55)
				}
				if sel {
					fg = theme.Mix(fg, white, 0.28)
				}
				if c.bg == sNone {
					if ch == '█' && textured(lv, c.fg, sx, sy) {
						s.fb.set(X, Y, '▓', fg, theme.Shade(fg, 0.82), lv.depth)
					} else if ch == '█' {
						s.fb.fill(X, Y, fg, lv.depth)
					} else {
						s.fb.fg(X, Y, ch, fg, lv.depth)
					}
					continue
				}
				bg := p.col(lv.fam, lv.vr, c.bg, hz)
				if tint != 0 && !c.bg.emissive() {
					bg = theme.Mix(bg, tint, 0.32)
				}
				if mega && (c.bg.emissive() || c.bg == sGlass || c.bg == sGlassHi) {
					bg = s.neonFix(bg, hz)
				}
				if dim {
					bg = theme.Shade(bg, 0.55)
				}
				if sel {
					bg = theme.Mix(bg, white, 0.28)
				}
				s.fb.set(X, Y, ch, fg, bg, lv.depth)
			}
		}
		for _, b := range spr.beacons {
			on := (s.anim/3+int(hash(int(lv.ml.Seed), b.x)%7))%8 < 3
			if on || night < 0.3 {
				c := p.hue(theme.SkyBeacon, mEmit, hz)
				if !on {
					c = theme.Shade(c, 0.55)
				}
				s.fb.fg(x0+b.x, s.Y(y0+b.y), '•', c, lv.depth)
			}
		}
		for _, b := range spr.blades {
			s.blades(x0+b.x, y0+b.y, lv)
		}
		for i, fl := range spr.flame {
			s.flame(x0+fl.x, y0+fl.y, int(lv.ml.Seed)+i, hz, lv.depth)
		}
		for i, e := range spr.smoke {
			if i > 0 && hash(int(lv.ml.Seed), i)%3 != 0 {
				continue // most buildings send up one plume, not one per flue
			}
			busy := lv.staff
			if !lv.producer {
				if hash(int(lv.ml.Seed), 9)%2 == 0 {
					continue // not every hearth is lit
				}
				busy = 0.35
			} else if busy < 0.05 {
				continue // an idle mill does not smoke: that is the point
			}
			s.emit = append(s.emit, emitter{x0 + e.x, y0 + e.y, busy, lv.producer && lv.ml.Age >= 8,
				int(lv.ml.Seed) + i*31, lv.depth - 1})
		}
	}
}

func textured(lv *lotView, sl slot, x, y int) bool {
	switch sl {
	case sRoof, sRoofShade, sRock, sRockDark, sLeaf, sLeafDark:
		return hash(int(lv.ml.Seed), x, y)%4 == 0
	case sWall, sWallShade:
		return lv.fam <= 5 && lv.fam != 2 && hash(int(lv.ml.Seed), x, y)%7 == 0
	}
	return false
}

// windowLit: light means labour. Windows light in proportion to staffing;
// an idle building stays dark.
func (s *scene) windowLit(lv *lotView, x, y int) bool {
	seed := int(lv.ml.Seed)
	if s.gridOut {
		return hashf(seed, x, y, 5) < 0.08
	}
	if lv.staff <= 0 {
		return false
	}
	on := hashf(seed, x, y) < 0.05+0.9*lv.staff
	if hash(seed, x, y, s.anim/48)%29 == 0 {
		on = !on
	}
	return on
}

func (s *scene) blades(x, y int, lv *lotView) {
	c := s.p.col(lv.fam, lv.vr, sTrim, lv.ml.Row)
	d := lv.depth
	put := func(dx, dy int, ch rune) { s.fb.fg(x+dx, s.Y(y+dy), ch, c, d) }
	if lv.staff > 0.05 && (s.anim/3)%2 == 1 {
		put(0, -1, '│')
		put(0, -2, '│')
		put(0, 1, '│')
		for _, dx := range []int{-2, -1, 1, 2} {
			put(dx, 0, '─')
		}
	} else {
		put(-1, -1, '╲')
		put(1, -1, '╱')
		put(-1, 1, '╱')
		put(1, 1, '╲')
		put(-2, -2, '╲')
		put(2, -2, '╱')
	}
	put(0, 0, '•')
}

var flameGlyphs = [4]rune{'▲', '^', '♠', '▲'}
var flameHues = [4]theme.SkyHue{theme.SkyFireHot, theme.SkyFire, theme.SkyFireDeep, theme.SkyFireGold}

func (s *scene) flame(x, y, seed, hz int, d uint8) {
	k := int(hash(seed, s.anim/2) % 4)
	s.fb.fg(x, s.Y(y), flameGlyphs[k], s.p.hue(flameHues[k], mEmit, hz), d)
}

// smoke rises from busy chimneys, leaning downwind and thinning out.
func (s *scene) smoke() {
	wind := 0.7 + 0.3*mapmodel.Sin(float64(s.m.Tick)/2500)
	if s.m.Weather.Kind == mapmodel.Storm {
		wind = 1.1
	}
	light, heavy := s.p.hue(theme.SkySmoke, mLit, 0), s.p.hue(theme.SkySmokeHeavy, mLit, 0)
	furnace := s.p.hue(theme.SkyFurnace, mEmit, 0)
	for _, e := range s.emit {
		n := 2 + int(e.strength*2)
		life := 11
		col := light
		if e.heavy {
			n++
			life = 14
			col = heavy
		}
		for k := 0; k < n; k++ {
			age := (s.anim + k*life/n + int(hash(e.seed)%uint64(life))) % life
			u := float64(age) / float64(life)
			if u > 0.82 {
				continue
			}
			fy := float64(e.y) - float64(age)*0.5
			fx := float64(e.x) + float64(age)*wind*1.3 + mapmodel.Sin(float64(age)*0.14+float64(e.seed%7)/7)*0.35
			x, y := int(math.Round(fx)), int(math.Round(fy))
			if y < 0 || y >= s.groundY {
				continue
			}
			c := col
			if e.heavy && s.p.night > 0.4 && u < 0.3 {
				c = theme.Mix(c, furnace, 0.35)
			}
			spread := int(u * 3)
			alpha := (1 - u) * 0.75
			for dx := 0; dx <= spread; dx++ {
				xx := x + dx - spread/2
				a := alpha
				if dx != spread/2 {
					a *= 0.6
				}
				s.fb.tint(xx, s.Y(y), c, a, e.depth)
				if dx == spread/2 {
					if u < 0.2 {
						s.fb.fg(xx, s.Y(y), '▓', c, e.depth)
					} else if cl := s.fb.at(xx, s.Y(y)); u < 0.45 && cl != nil && cl.ch == ' ' {
						s.fb.fg(xx, s.Y(y), '▒', c, e.depth)
					}
				}
			}
		}
	}
}

// ------------------------------------------------------------------ ground

func roadHues(a int) (theme.SkyHue, theme.SkyHue) {
	switch {
	case a <= 1:
		return theme.SkyRoadDirt, theme.SkyRoadDirtMark
	case a <= 7:
		return theme.SkyRoadCobble, theme.SkyRoadCobbleMark
	case a <= 11:
		return theme.SkyRoadRail, theme.SkyRoadRailMark
	case a <= 14:
		return theme.SkyRoadAsphalt, theme.SkyRoadAsphaltMark
	case a <= 16:
		return theme.SkyRoadNeon, theme.SkyRoadNeonMark
	}
	return theme.SkyRoadLight, theme.SkyRoadLightMark
}

func (s *scene) ground() {
	a := s.m.AgeIdx
	p := s.p
	gy := s.groundY
	rh, mh := roadHues(a)
	rc := p.hue(rh, mLit, 0)
	mode := mLit
	if a > 14 {
		mode = mEmit
	}
	mc := p.hue(mh, mode, 0)
	di := -1
	paved := s.inCity && mapmodel.Greenery(a) <= 0.15 // the city has paved the old grass
	for x := 0; x < s.W; x++ {
		wx := s.wx(x)
		if di < 0 || di >= len(s.m.Skyline.Districts) || !inDistrict(s.m.Skyline.Districts[di], wx) {
			di = s.m.Skyline.DistrictAt(wx)
		}
		dAge := -1
		if di >= 0 {
			d := s.m.Skyline.Districts[di]
			dAge = d.Age
			if d.Bay && wx >= d.BayX && wx < d.BayX+d.BayW {
				s.water(x, wx, d)
				continue
			}
		}
		top, under := theme.SkyGroundGrass, theme.SkyGroundSoil
		switch {
		case dAge > 11 || paved || dAge <= 3 && s.inCity && !s.greenKept(hashf(wx/6, 817)):
			top, under = theme.SkyGroundSlab, theme.SkyGroundSlabDark
		case dAge > 7:
			top, under = theme.SkyGroundBrick, theme.SkyGroundBrickDark
		case dAge > 3:
			top, under = theme.SkyGroundCobble, theme.SkyGroundCobbleDark
		}
		s.fb.set(x, s.Y(gy), '▀', p.hue(top, mLit, 0), p.hue(under, mLit, 0), dGround)
		// the road: re-skinned to the current age through the whole city, so
		// history stands on today's street
		ch := ' '
		switch {
		case a <= 1:
			if hash(wx, 3)%3 == 0 {
				ch = '░'
			}
		case a <= 7:
			if hash(wx, 4)%4 == 0 {
				ch = '·'
			}
		case a <= 11:
			ch = '═'
		case a <= 16:
			if (wx/3)%2 == 0 {
				ch = '─'
			}
		default:
			ch = '═'
		}
		s.fb.set(x, s.Y(gy+1), ch, mc, rc, dGround)
		vg, vb := theme.SkyVerge, theme.SkyVergeDark
		if a >= 12 && (dAge >= 0 || s.inCity && s.city.Greenery < 1) {
			vg, vb = theme.SkyVergeUrban, theme.SkyVergeUrbanDark
		}
		vch := ' '
		if h := hash(wx, 8) % 7; h < 2 && vg == theme.SkyVerge {
			vch = [2]rune{'"', ','}[h]
		}
		s.fb.set(x, s.Y(gy+2), vch, p.hue(vg, mLit, 0), p.hue(vb, mLit, 0), dGround)
		for y := gy + 3; y < s.S; y++ {
			s.fb.fill(x, s.Y(y), p.hue(vb, mLit, 0), dGround)
		}
	}
	s.streetFurniture()
	s.frontier()
	s.cityGround()
}

func inDistrict(d mapmodel.District, x int) bool { return x >= d.X0 && x < d.X0+d.W }

func (s *scene) water(x, wx int, d mapmodel.District) {
	p := s.p
	white := theme.SkyColor(theme.SkyWhite)
	for y := s.groundY; y < s.S; y++ {
		k := float64(y-s.groundY) / 3
		c := theme.Mix(theme.SkyColor(theme.SkyWater), theme.SkyColor(theme.SkyWaterDeep), math.Min(1, k))
		c = theme.Mix(c, p.skyAt(max(0, s.groundY-1-(y-s.groundY)*3), s.groundY), 0.35)
		col := p.final(theme.Tint(c, p.lightTint))
		ch, fg := ' ', theme.Mix(col, white, 0.25)
		switch hash(wx, y, s.anim/4) % 9 {
		case 0:
			ch = '~'
		case 1:
			ch, fg = '▀', theme.Shade(col, 1.15)
		}
		if above := s.fb.at(x, s.Y(s.groundY-1-(y-s.groundY))); above != nil && p.night > 0.4 &&
			luma(above.fg) > luma(col)+60 && hash(wx, y, s.anim/3)%3 != 0 {
			ch, fg = '▒', theme.Mix(above.fg, col, 0.45)
		}
		s.fb.set(x, s.Y(y), ch, fg, col, dGround)
	}
	if wx == d.BayX || wx == d.BayX+d.BayW-1 {
		s.fb.fill(x, s.Y(s.groundY), p.hue(theme.SkyQuay, mLit, 0), dGround)
	}
}

// streetFurniture: lamps from the Victorian age, market pennants in the
// medieval and renaissance towns.
func (s *scene) streetFurniture() {
	a := s.m.AgeIdx
	if a < 5 || (a > 6 && a < 9) {
		return
	}
	post := s.p.hue(theme.SkyLampPost, mLit, 0)
	lamp := s.p.hue(theme.SkyLamp, mEmit, 0)
	flags := [3]tcell.Color{s.p.col(3, 0, sRoof, 0), s.p.col(4, 0, sNeon1, 0), s.p.col(4, 1, sMetal, 0)}
	for x := 0; x < s.W; x++ {
		wx := s.wx(x)
		if wx%16 != 0 || s.m.Skyline.DistrictAt(wx) < 0 {
			continue
		}
		s.fb.fg(x, s.Y(s.groundY-1), '│', post, dLamp)
		if a <= 6 {
			s.fb.fg(x, s.Y(s.groundY-2), '│', post, dLamp)
			s.fb.fg(x+1, s.Y(s.groundY-2), '►', flags[(wx/16)%3], dLamp)
			continue
		}
		head := post
		if s.p.night > 0.3 {
			head = lamp
			s.fb.tint(x, s.Y(s.groundY+1), lamp, 0.25, dGround)
		}
		s.fb.fg(x, s.Y(s.groundY-2), '▀', head, dLamp)
	}
}

// frontier is where the city is still being built: one construction per
// queued build (four at most), in the way its era built. The first ages
// raise poles and stick frames; from the Iron Age timber scaffolding and
// wooden jib cranes go up; from the Industrial Age, steel tower cranes.
func (s *scene) frontier() {
	fx := s.m.Skyline.FrontierX - s.cam + 4
	iron, industrial := s.m.Catalog.AgeIdx["iron_age"], s.m.Catalog.AgeIdx["industrial_age"]
	for i := range s.m.Queue {
		if i >= 4 {
			break
		}
		x := fx + i*7
		if x < -8 || x > s.W+2 {
			continue
		}
		p := s.m.Queue[i].Progress
		switch {
		case s.m.AgeIdx < iron:
			s.stickFrame(x, i, p)
		case s.m.AgeIdx < industrial:
			s.timberWork(x, i, p)
		default:
			s.towerCrane(x, i, p)
		}
	}
}

// stickFrame is early building: a tripod of poles lashed at the top with
// hides going on, or two posts and a lintel with the wall rising between.
func (s *scene) stickFrame(x, i int, p float64) {
	wood := s.p.hue(theme.SkyScaffold, mLit, 0)
	fill := s.p.hue(theme.SkyGroundSoil, mLit, 0)
	g := s.groundY
	if g < 4 {
		return
	}
	put := func(dx, y int, r rune, c tcell.Color) { s.fb.fg(x+dx, s.Y(y), r, c, dRow0) }
	if i%2 == 0 {
		put(2, g-3, '┼', wood)
		put(1, g-2, '╱', wood)
		put(2, g-2, '│', wood)
		put(3, g-2, '╲', wood)
		put(0, g-1, '╱', wood)
		put(2, g-1, '│', wood)
		put(4, g-1, '╲', wood)
		if p >= 0.4 {
			put(1, g-1, '▒', fill)
			put(3, g-1, '▒', fill)
		}
		return
	}
	put(0, g-3, '┬', wood)
	put(1, g-3, '─', wood)
	put(2, g-3, '─', wood)
	put(3, g-3, '─', wood)
	put(4, g-3, '┬', wood)
	for y := g - 2; y < g; y++ {
		put(0, y, '│', wood)
		put(4, y, '│', wood)
	}
	for dx := 1; dx <= 1+int(p*3) && dx <= 3; dx++ {
		put(dx, g-1, '▄', fill)
	}
}

// timberWork is building from the Iron Age to the Industrial: a timber
// scaffold round a rising wall, or a wooden jib crane with its load on a
// rope.
func (s *scene) timberWork(x, i int, p float64) {
	wood := s.p.hue(theme.SkyScaffold, mLit, 0)
	wall := s.p.hue(theme.SkyGroundCobble, mLit, 0)
	g := s.groundY
	if g < 6 {
		return
	}
	put := func(dx, y int, r rune, c tcell.Color) { s.fb.fg(x+dx, s.Y(y), r, c, dRow0) }
	if i%2 == 0 {
		h := 3 + int(p*2) // the scaffold climbs with the wall
		for y := g - h; y < g; y++ {
			row := "├┼┼┤"
			if y == g-h {
				row = "┌┬┬┐"
			}
			for dx, r := range []rune(row) {
				put(dx, y, r, wood)
			}
		}
		for y := g - max(1, int(p*float64(h))); y < g; y++ {
			put(1, y, '▓', wall)
			put(2, y, '▓', wall)
		}
		return
	}
	put(0, g-4, '┌', wood)
	put(1, g-4, '─', wood)
	put(2, g-4, '─', wood)
	put(3, g-4, '┐', wood)
	for y := g - 3; y < g; y++ {
		put(0, y, '│', wood)
	}
	load := g - 3 + (s.anim/6+i)%3 // the load goes up and down
	for y := g - 3; y < load; y++ {
		put(3, y, '┊', wood)
	}
	put(3, load, '▪', wall)
	if p >= 0.5 {
		put(1, g-1, '▓', wall)
		put(2, g-1, '▓', wall)
	}
}

// towerCrane is industrial building: a steel tower crane, its hook
// swinging, over the scaffolded frame of what it builds.
func (s *scene) towerCrane(x, i int, p float64) {
	c := s.p.hue(theme.SkyCrane, mLit, 0)
	sc := s.p.hue(theme.SkyScaffold, mLit, 0)
	h := min(6+int(hash(i, 5)%4), s.groundY-1)
	for y := s.groundY - h; y < s.groundY; y++ {
		s.fb.fg(x, s.Y(y), '╫', c, dRow0)
	}
	for dx := -2; dx <= 4; dx++ {
		s.fb.fg(x+dx, s.Y(s.groundY-h), '═', c, dRow0)
	}
	s.fb.fg(x+3, s.Y(s.groundY-h+1+(s.anim/6+i)%3), '┴', c, dRow0)
	built := int(p*3) + 1
	for y := s.groundY - min(built, 3); y < s.groundY; y++ {
		for dx := 1; dx <= 4; dx++ {
			s.fb.fg(x+dx, s.Y(y), '┼', sc, dRow0)
		}
	}
}

// ------------------------------------------------------------------ weather

func (s *scene) weather() {
	s.citySmog()
	k := s.m.Weather.Kind
	drops := 0
	col := theme.Mix(s.p.hue(theme.SkyRain, mLit, 0), s.p.sky[2], 0.3)
	switch {
	case k == mapmodel.Rain:
		drops = s.W * s.S / 22
	case k == mapmodel.Storm:
		drops = s.W * s.S / 10
	case s.m.Epoch == 5 && s.p.night > 0.5 && k != mapmodel.Snow && !s.inCity:
		drops = s.W * s.S / 70 // the neon era's night drizzle
		col = theme.Mix(col, s.p.hue(theme.SkyNeonMagenta, mEmit, 0), 0.35)
	case s.cyberpunk() && k != mapmodel.Snow:
		s.cityRain() // the megacity's acid rain, day and night
	}
	for i := 0; i < drops; i++ {
		y := int((hash(i, 1)%uint64(s.S+10) + uint64(s.anim*2)) % uint64(s.S+2))
		x := int(hash(i, 2)%uint64(s.W+s.S)) - y/2
		if y < s.S && x >= 0 && x < s.W {
			s.fb.fg(x, s.Y(y), '╱', col, dWeather)
		}
	}
	if k == mapmodel.Snow {
		sc := s.p.hue(theme.SkyWhite, mLit, 0)
		for i := 0; i < s.W*s.S/30; i++ {
			y := int((hash(i, 11)%uint64(s.S+10) + uint64(s.anim/2)) % uint64(s.S+2))
			x := int(hash(i, 12)%uint64(s.W)) + int(mapmodel.Sin(float64(y)/9+hashf(i))*1.5)
			if y < s.S {
				s.fb.fg(x, s.Y(y), [2]rune{'·', '*'}[hash(i)%2], sc, dWeather)
			}
		}
	}
	if k == mapmodel.Smog {
		sm := s.p.hue(theme.SkySmokeHeavy, mLit, 1)
		for y := s.groundY / 2; y < s.groundY; y++ {
			for x := 0; x < s.W; x++ {
				if hash(x/3, y, s.anim/12)%5 == 0 {
					s.fb.tint(x, s.Y(y), sm, 0.18, dLane2)
				}
			}
		}
	}
	if k == mapmodel.Storm && (s.anim/4)%23 == 0 {
		x := int(hash(s.anim/92, 7) % uint64(max(1, s.W)))
		bolt := s.p.hue(theme.SkyWhite, mEmit, 0)
		for y := 0; y < s.groundY-4; y++ {
			ch := '╲'
			if hash(y, x)%2 == 0 {
				ch = '╱'
				x++
			} else {
				x--
			}
			s.fb.fg(x, s.Y(y), ch, bolt, dCrack)
		}
	}
}

// ------------------------------------------------------------ catastrophe

// catastrophe draws the pending catastrophe's overlay for its epoch: roof
// fires, cracks in the sky, glitch bands with the grid down.
func (s *scene) catastrophe() {
	if s.m.Catastrophe.Pending == "" {
		return
	}
	switch s.m.Catalog.EpochIdx[s.m.Catastrophe.Pending] {
	case 0: // meteor impact
		s.meteor()
		s.fires(4)
	case 3: // meltdown
		s.skyCracks()
		s.fires(5)
	case 4, 5: // digital collapse, solar event
		s.glitch()
		s.fires(8)
	case 6: // reality fracture
		s.skyCracks()
		s.glitch()
	default: // barbarians, industrial collapse
		s.fires(3)
	}
}

func (s *scene) fires(every int) {
	black := s.p.hue(theme.SkySmokeBlack, mFlat, 0)
	for _, li := range s.vis {
		lv := &s.lay.lots[li]
		if lv.ml.Wonder || int(hash(int(lv.ml.Seed), 77)%uint64(every)) != 0 {
			continue
		}
		x0, y0 := lv.x0-s.cam, s.groundY-lv.spr.h
		for sx := 0; sx < lv.spr.w; sx++ {
			if t := lv.spr.topAt(sx); t >= 0 && hash(int(lv.ml.Seed), sx)%2 == 0 {
				s.flame(x0+sx, y0+t-1, int(lv.ml.Seed)+sx*7, 0, lv.depth-1)
			}
		}
		for k := 0; k < 7; k++ {
			age := (s.anim + k*3) % 20
			x, y := x0+lv.spr.w/2+age/3, y0-2-age/2
			if y >= 0 {
				s.fb.fg(x, s.Y(y), [4]rune{'█', '▓', '▒', '░'}[min(3, age/5)], black, lv.depth-1)
			}
		}
	}
}

func (s *scene) glitch() {
	cols := [3]tcell.Color{s.p.hue(theme.SkyNeonPink, mEmit, 0), s.p.hue(theme.SkyNeonCyan, mEmit, 0), s.p.hue(theme.SkyWhite, mEmit, 0)}
	for i := 0; i < s.S/2; i++ {
		y := int(hash(i, s.anim/2) % uint64(max(1, s.groundY)))
		x := int(hash(i, 5, s.anim/2) % uint64(max(1, s.W)))
		for k := 0; k < 4+int(hash(i, 9)%16); k++ {
			s.fb.fg(x+k, s.Y(y), []rune("▓▒░▀▄")[hash(k, i)%5], cols[i%3], dWeather)
		}
	}
}

func (s *scene) skyCracks() {
	c := s.p.hue(theme.SkyCrack, mEmit, 0)
	for i := 0; i < 3; i++ {
		x := s.W/4 + i*s.W/4
		for y := 0; y < s.groundY*2/3; y++ {
			if hash(i, y)%2 == 0 {
				x++
				s.fb.fg(x, s.Y(y), '╲', c, dCrack)
			} else {
				x--
				s.fb.fg(x, s.Y(y), '╱', c, dCrack)
			}
		}
	}
}

func (s *scene) meteor() {
	c, tail := s.p.hue(theme.SkyFireHot, mEmit, 0), s.p.hue(theme.SkyFire, mEmit, 0)
	x0, y0 := s.W*2/3, 2
	for i := 0; i < s.groundY/3; i++ {
		ch, col := '░', tail
		if i == s.groundY/3-1 {
			ch, col = '●', c
		} else if i > s.groundY/4 {
			ch = '▒'
		}
		s.fb.fg(x0-i*2, s.Y(y0+i), ch, col, dCrack)
	}
}

// ------------------------------------------------------------ markers

// markers: the ▼ over what is new since the last visit, the flows tags,
// the labels on the grid and the inspect cursor.
func (s *scene) markers() {
	pos := s.mp.Fg[mapmodel.CFresh]
	warn := s.mp.Fg[mapmodel.CIdle]
	full := len(s.m.Flows.Full) > 0
	for _, li := range s.vis {
		lv := &s.lay.lots[li]
		x := lv.x0 - s.cam + lv.spr.w/2
		y := max(0, s.groundY-lv.spr.h-1)
		if s.v.changes && lv.ml.New {
			s.fb.fg(x, s.Y(y), mapmodel.R(mapmodel.SymFresh, s.tier), pos, dTag)
			if s.tier != mapmodel.TierNerd {
				s.fb.fg(x, s.Y(y), '▼', pos, dTag)
			}
			y--
		}
		if s.v.flows && lv.ml.Copy == 0 {
			tag := ""
			switch {
			case flagged(lv) && lv.b.Workers == 0:
				tag = "IDLE"
			case flagged(lv):
				tag = "LOW"
			case full && lv.ml.Lineage == mapmodel.LinStorage:
				tag = "FULL"
			}
			if tag != "" && y >= 0 {
				bg := s.fb.showAt(clampInt(x, 0, s.W-1), s.Y(y))
				s.fb.text(x-len(tag)/2, s.Y(y), tag, theme.Legible(warn, bg, 3), bg, dTag)
			}
		}
	}
	for _, lb := range s.labels {
		for dy := 0; dy < 3; dy++ {
			if s.labelFree(lb.x, lb.y-dy, textLen(lb.text)) {
				for i, r := range []rune(lb.text) {
					if c := s.fb.at(lb.x+i, s.Y(lb.y-dy)); c != nil {
						s.fb.fg(lb.x+i, s.Y(lb.y-dy), r, theme.Legible(lb.fg, c.show(), 3), dLabel)
					}
				}
				break
			}
		}
	}
	s.cursor()
}

func (s *scene) labelFree(x, y, n int) bool {
	if y < 0 || y >= s.groundY {
		return false
	}
	for i := 0; i < n; i++ {
		if c := s.fb.at(x+i, s.Y(y)); c != nil && c.d <= dLabel {
			return false
		}
	}
	return true
}

func (s *scene) cursor() {
	if !s.v.inspect {
		return
	}
	acc := s.mp.Fg[mapmodel.CAccent]
	bgc := s.mp.Bg
	name := ""
	var x0, x1, top int
	switch t := s.v.cur; {
	case s.sel >= 0:
		lv := &s.lay.lots[s.sel]
		x0, x1 = lv.x0-s.cam, lv.x0-s.cam+lv.spr.w-1
		top = s.groundY - lv.spr.h
		name = lv.b.Name
		if lv.ml.Wonder {
			name = s.m.Catalog.Defs[lv.ml.Key].Name
		}
	case t.kind == tTether:
		if x := s.elevatorX(); x >= 0 && x < s.W {
			s.fb.fg(x-1, s.Y(s.groundY/2), '►', acc, dTop)
			lbl := " " + mapmodel.FeatTether.Info().Title + " "
			s.fb.text(clampInt(x-textLen(lbl)/2, 0, max(0, s.W-textLen(lbl))), s.Y(s.groundY/2-2), lbl, bgc, acc, dTop)
		}
		return
	case t.kind == tUFO:
		if x, y, ok := saucerAt(s.m, s.visit, s.W); ok {
			mark, my := '▼', y-1
			if my < 0 {
				mark, my = '▲', y+1
			}
			s.fb.fg(x+1, s.Y(my), mark, acc, dTop)
		}
		return
	case t.kind != tLot:
		for _, it := range s.ridge {
			if (t.kind == tHarbinger && it.fac == nil) || (it.fac != nil && it.fac.Key == t.key) {
				s.fb.fg(it.x, s.Y(s.ridgeTop(it.x)-6), '▼', acc, dTop)
			}
		}
		return
	default:
		return
	}
	for x := x0 + 1; x < x1; x++ {
		s.fb.fg(x, s.Y(s.groundY+2), '─', acc, dTop)
	}
	s.fb.fg(x0, s.Y(s.groundY+2), '└', acc, dTop)
	s.fb.fg(x1, s.Y(s.groundY+2), '┘', acc, dTop)
	s.fb.fg((x0+x1)/2, s.Y(s.groundY+1), '▲', acc, dTop)
	lbl := " " + name + " "
	n := textLen(lbl)
	s.fb.text(clampInt((x0+x1)/2-n/2, 0, max(0, s.W-n)), s.Y(max(0, top-2)), lbl, bgc, acc, dTop)
}

// ------------------------------------------------------------ ridge items

// ridgeItem is a civ town (fac set) or the harbinger on the far ridge.
type ridgeItem struct {
	fac *mapmodel.Faction
	x   int // screen column
	i   int
}

// ridgeItems places the discovered civs west to east by their site, then
// the harbinger, on the far ridge (which scrolls at 0.15 parallax).
func ridgeItems(m *mapmodel.Model, W, cam int) []ridgeItem {
	var civs []*mapmodel.Faction
	for i := range m.Factions {
		if m.Factions[i].Discovered {
			civs = append(civs, &m.Factions[i])
		}
	}
	w := m.Town.World
	siteX := func(f *mapmodel.Faction) int {
		if w == nil || f.Site < 0 || f.Site >= len(w.Sites) {
			return f.Site
		}
		return w.Sites[f.Site].X - w.CX
	}
	sort.SliceStable(civs, func(i, j int) bool { return siteX(civs[i]) < siteX(civs[j]) })
	span := W + int(float64(max(0, m.Skyline.Width-W))*0.15)
	shift := int(float64(cam) * 0.15)
	var out []ridgeItem
	for i, f := range civs {
		x := (i+1)*span/(len(civs)+1) + int(hash(i, 17)%5) - 2
		out = append(out, ridgeItem{fac: f, x: x - shift, i: i})
	}
	if h := m.Harbinger; h != nil {
		x := span/6 + int(hash(int(mapmodel.HashStr(h.Key)))%uint64(max(1, span*2/3)))
		for _, it := range out { // keep off the towns
			if d := x - it.x - shift; d > -8 && d < 8 {
				x += 16
			}
		}
		out = append(out, ridgeItem{x: x - shift, i: len(out)})
	}
	return out
}

func textLen(s string) int { return mapstyle.TextLen(s) }
