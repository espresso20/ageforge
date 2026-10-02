package skyline

import (
	"math"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
)

// city.go is the Earth arc's skyline, the Modern Age to the Fusion Age,
// drawn from the shared look (mapmodel.CityLookAt):
//
//   - Modern: street trees and parks fenced along the verge, an elevated
//     freeway with its cars streaming (headlights at night).
//   - Information: half the trees left, dishes on the roofs, gardens on a
//     few, screens flickering in the windows, a thin smog band.
//   - Digital: one walled reserve in the foreground and nothing else
//     green, a thicker smog over the outskirts, the data glow round the
//     data halls, drones.
//   - Cyberpunk: the megacity: the far ridge and the near hills built over
//     with layered towers, holo-ads in the air, three to five elevated
//     rails with sky trains on each, acid rain day and night, drone swarms
//     and hovercars, and no green anywhere.
//   - Fusion: the megacity powered: reactor domes glowing in the near
//     layer and round the reactor lots, launch towers, the space elevator's
//     tether rising off the top of the screen with its climbers, the road
//     carrying plasma, the smog thinned and the rain gone.
//
// Everything is a pure function of the model, the frame and the camera.

// cityHue resolves an Earth-arc city colour in a mode at a haze depth.
func (p *pal) cityHue(h theme.CityHue, mode hueMode, depth int) tcell.Color {
	depth = clampInt(depth, 0, numHaze-1)
	i := (int(h)*int(numModes)+int(mode))*numHaze + depth
	if p.cityCache == nil {
		p.cityCache = make([]tcell.Color, theme.NumCityHues*int(numModes)*numHaze)
	}
	if i < 0 || i >= len(p.cityCache) {
		return p.final(theme.CityColor(h))
	}
	if c := p.cityCache[i]; c != 0 {
		return c
	}
	c := theme.CityColor(h)
	switch mode {
	case mLit:
		c = theme.Mix(theme.Tint(c, p.lightTint), p.horizon, p.haze(depth))
	case mEmit:
		c = theme.Mix(c, p.horizon, float64(p.haze(depth)*0.8))
	}
	c = p.final(c)
	p.cityCache[i] = c
	return c
}

// megacity reports the megacity's frames (the Cyberpunk and Fusion Ages).
func (s *scene) megacity() bool { return s.inCity && s.city.Night }

func (s *scene) fusion() bool { return s.inCity && s.city.Key == "fusion_age" }

func (s *scene) cyberpunk() bool { return s.inCity && s.city.Key == "cyberpunk_age" }

// greenKept reports whether a piece of greenery hashed to h survives the
// age's greenery fade (every piece in the Modern Age, about half in the
// Information Age, none after).
func (s *scene) greenKept(h float64) bool {
	return h < mapmodel.Greenery(s.m.AgeIdx)
}

// ------------------------------------------------------------ the ridge

// megacityRidge builds the far ridge over: a skyline of distant towers in
// blocks of two to four columns, lit at night, the tallest with a beacon.
func (s *scene) megacityRidge() {
	gy := s.groundY
	amp := float64(gy) * 0.46
	tops := make([]float64, s.W)
	hc := theme.CityTower
	if s.fusion() {
		hc = theme.CityCooling
	}
	col := s.p.final(theme.Mix(s.p.cityHue(hc, mLit, 3), s.p.horizon, 0.35))
	win := s.p.cityHue(theme.CityTowerLit, mEmit, 3)
	beacon := s.p.hue(theme.SkyBeacon, mEmit, 3)
	shift := int(float64(s.cam) * 0.15)
	seed := int(s.m.Seed % 1000)
	for x := 0; x < s.W; x++ {
		wx := x + shift
		blk := wx / 3
		if hash(blk, 7, seed)%3 == 0 {
			blk = wx / 2
		}
		n := fbm1(float64(wx)/30, 300+seed, 3)
		h := float64(float64(0.35+float64(0.65*n))*amp) * (0.55 + float64(0.45*hashf(blk, 11, seed)))
		tops[x] = float64(gy-1) - h
	}
	s.paintRidge(tops, col, dRidge)
	for x, t := range tops {
		top := int(t)
		wx := x + shift
		if s.p.night > 0.3 {
			for y := top + 1; y < s.groundY; y += 2 {
				if hash(wx, y, 31)%5 == 0 {
					s.fb.set(x, s.Y(y), '·', win, col, dRidge)
				}
			}
		}
		if top < s.groundY-int(amp*0.8) && hash(wx/2, 33)%4 == 0 && (s.anim/4+wx)%6 < 3 {
			s.fb.fg(x, s.Y(top-1), '•', beacon, dRidge)
		}
	}
	s.farTops = tops
}

// megacityNear is the megacity's near layer: the distant town (midTown)
// carries it, built up with neon (megacityGlow); in the Fusion Age reactor
// domes glow in front of it and launch towers stand ready, and in the
// Cyberpunk Age holo-ads hang in the air.
func (s *scene) megacityNear() {
	if s.fusion() {
		s.megacityFusion(int(float64(s.cam) * 0.4))
	}
	if s.cyberpunk() {
		s.holoAds()
	}
}

// megacityGlow dresses one column of the distant town in the megacity's
// neon: a strip up the edge of every third block, and lit windows in the
// age's colours (magenta and screens; white and electric blue once fusion
// powers the city).
func (s *scene) megacityGlow(x, top, blk int, first bool, col tcell.Color) {
	neons := [3]tcell.Color{s.p.cityHue(theme.CityNeonMagenta, mEmit, 3), s.p.cityHue(theme.CityNeonCyan, mEmit, 3),
		s.p.cityHue(theme.CityNeonYellow, mEmit, 3)}
	if s.fusion() {
		neons = [3]tcell.Color{s.p.cityHue(theme.CityNeonCyan, mEmit, 3), s.p.cityHue(theme.CityPlasma, mEmit, 3),
			s.p.cityHue(theme.CityElectric, mEmit, 3)}
	}
	if first && hash(blk, 23)%3 == 0 {
		nc := neons[hash(blk, 24)%3]
		for y := top; y < s.groundY; y++ {
			s.fb.set(x, s.Y(y), '▌', nc, col, dTown)
		}
		return
	}
	for y := top + 1; y < s.groundY; y += 2 {
		if hash(x+s.cam, y, 26)%5 == 0 {
			s.fb.set(x, s.Y(y), '▪', neons[hash(x+s.cam, y, 27)%3], col, dTown)
		}
	}
}

// megacityFusion raises the Fusion Age's reactor domes and launch towers in
// the near layer, where megacityNear left room.
func (s *scene) megacityFusion(shift int) {
	gy := s.groundY
	dome := s.p.final(theme.Mix(s.p.cityHue(theme.CitySteel, mLit, 2), s.p.horizon, 0.2))
	plasma := s.p.cityHue(theme.CityPlasma, mEmit, 1)
	ring := s.p.cityHue(theme.CityReactor, mEmit, 1)
	steel := s.p.cityHue(theme.CitySteel, mLit, 2)
	beacon := s.p.hue(theme.SkyBeacon, mEmit, 2)
	pulse := 0.5 + float64(0.5*mapmodel.Sin(float64(s.anim)/24))
	for blk := (shift - 10) / 5; blk*5-shift < s.W+10; blk++ {
		if hash(blk, 27)%9 != 0 || hash(blk-1, 27)%9 == 0 {
			continue
		}
		x0 := blk*5 - shift
		if hash(blk, 28)%2 == 0 { // a reactor dome, its plasma ring glowing
			r := 4
			for dx := -r; dx <= r+4; dx++ {
				x := x0 + dx
				u := float64(dx-2) / float64(r+2)
				hgt := int(math.Round(float64(float64(r) * math.Sqrt(math.Max(0, 1-float64(u*u))))))
				for k := 0; k < hgt; k++ {
					s.fb.fill(x, s.Y(gy-1-k), dome, dTown-1)
				}
				if hgt > 1 {
					s.fb.fg(x, s.Y(gy-1-hgt/2), '═', theme.Mix(ring, plasma, pulse), dTown-1)
				}
			}
			for dx := -r - 1; dx <= r+5; dx++ { // the glow round it
				for y := gy - r - 2; y < gy; y++ {
					s.fb.tint(x0+dx, s.Y(y), plasma, float64(0.08+float64(0.1*pulse)), dTown)
				}
			}
			continue
		}
		// a launch tower: a gantry and a rocket on its pad
		h := min(gy-2, int(float64(gy)*0.45))
		for y := gy - h; y < gy; y++ {
			s.fb.fg(x0, s.Y(y), '╫', steel, dTown-1)
		}
		for y := gy - h + 2; y < gy; y++ {
			s.fb.fill(x0+2, s.Y(y), dome, dTown-1)
		}
		s.fb.fg(x0+2, s.Y(gy-h+1), '▲', plasma, dTown-1)
		s.fb.fg(x0+1, s.Y(gy-h+3), '─', steel, dTown-1)
		if (s.anim/4)%3 != 0 {
			s.fb.fg(x0, s.Y(gy-h-1), '•', beacon, dTown-1)
		}
	}
}

// holoAds hang the megacity's holo-ads in the air over the near layer: a
// word or a megacorp's initials in neon, flickering, anchored to the world
// so they scroll with the city.
func (s *scene) holoAds() {
	gy := s.groundY
	shift := int(float64(s.cam) * 0.5)
	words := []string{"SOMA", "24H", "KOG", "NVT", "LIVE", "HXD", "SALE", "VKS", "CHIP", "OPEN", "AUGS", "ZNA"}
	neons := [3]tcell.Color{s.p.cityHue(theme.CityNeonMagenta, mEmit, 0), s.p.cityHue(theme.CityNeonCyan, mEmit, 0),
		s.p.cityHue(theme.CityNeonYellow, mEmit, 0)}
	back := s.p.cityHue(theme.CityNight, mFlat, 0)
	for blk := (shift - 60) / 31; blk*31-shift < s.W+20; blk++ {
		if hash(blk, 41)%4 == 0 {
			continue
		}
		w := words[hash(blk, 42)%uint64(len(words))]
		x0 := blk*31 - shift + int(hash(blk, 43)%13)
		y0 := 3 + int(hash(blk, 44)%uint64(max(1, gy/3)))
		nc := neons[hash(blk, 45)%3]
		if hash(blk, s.anim/3, 46)%17 == 0 {
			continue // the ad flickers out
		}
		frame := theme.Mix(back, nc, 0.45)
		for dx := -2; dx <= 2*len(w); dx++ {
			for dy := -1; dy <= 1; dy++ {
				x, y := x0+dx, y0+dy
				switch {
				case dy == 0 && dx >= 0 && dx < 2*len(w) && dx%2 == 0:
					s.fb.set(x, s.Y(y), rune(w[dx/2]), nc, theme.Mix(back, nc, 0.22), dAirLow+1)
				case dy == 0 && dx >= 0 && dx < 2*len(w):
					s.fb.set(x, s.Y(y), ' ', nc, theme.Mix(back, nc, 0.22), dAirLow+1)
				case dy != 0:
					s.fb.fg(x, s.Y(y), [2]rune{'▄', '▀'}[(dy+1)/2], frame, dAirLow+1)
				default:
					s.fb.fg(x, s.Y(y), [2]rune{'▐', '▌'}[min(1, max(0, dx))], frame, dAirLow+1)
				}
			}
		}
	}
}

// ------------------------------------------------------------ the ground

// cityGround plants the Earth arc's foreground green, by the greenery fade:
// street trees in front of the buildings and parks fenced along the verge
// in the Modern Age, about half of them in the Information Age, and in the
// Digital Age only the walled reserve; then the road's own light (the data
// age's glowing marks, the Fusion Age's plasma).
func (s *scene) cityGround() {
	if !s.inCity {
		return
	}
	gy := s.groundY
	leaf := s.p.cityHue(theme.CityPark, mLit, 0)
	trunk := s.p.hue(theme.SkyTrunk, mLit, 0)
	hedge := s.p.cityHue(theme.CityHedge, mLit, 0)
	grass := s.p.cityHue(theme.CityReserve, mLit, 0)
	r0, r1 := s.reserveSpan()
	for x := 0; x < s.W; x++ {
		wx := s.wx(x)
		di := s.m.Skyline.DistrictAt(wx)
		if di >= 0 {
			if d := s.m.Skyline.Districts[di]; d.Bay && wx >= d.BayX && wx < d.BayX+d.BayW {
				continue
			}
		}
		if wx >= r0 && wx < r1 {
			s.reserveColumn(x, wx, r0, r1)
			continue
		}
		if s.city.Key == "digital_age" || s.city.Greenery == 0 {
			continue // the reserve is all the green the Digital Age keeps
		}
		if hash(wx/12, 803)%3 != 2 && s.greenKept(hashf(wx/12, 804)) { // a park along the verge
			ch := '"'
			if wx%3 == 0 {
				ch = '┼'
			}
			fg := leaf
			if ch == '┼' {
				fg = hedge
			}
			s.fb.set(x, s.Y(gy+2), ch, fg, theme.Mix(grass, s.fb.showAt(x, s.Y(gy+2)), 0.35), dGround)
		}
		if hash(wx, 801)%7 == 0 && s.greenKept(hashf(wx, 802)) { // a street tree
			s.fb.fg(x, s.Y(gy-1), '│', trunk, dLamp)
			s.fb.fg(x, s.Y(gy-2), '♣', leaf, dLamp)
		}
	}
	if s.inCity && s.city.Key == "digital_age" || s.fusion() {
		s.roadLight()
	}
}

// reserveSpan is the Digital Age's walled reserve on the panorama: a strip
// in front of the Digital district, about 15% of the Modern Age's green, or
// nothing in any other age.
func (s *scene) reserveSpan() (int, int) {
	if !s.inCity || s.city.Key != "digital_age" || len(s.m.Skyline.Districts) <= s.city.Age {
		return 0, 0
	}
	d := s.m.Skyline.Districts[s.city.Age]
	w := max(6, d.LandW/8)
	x0 := d.X0 + d.LandW/3 - w/2
	return x0, x0 + w
}

// reserveColumn draws one column of the reserve: its wall at the ends, old
// trees packed inside, grass under them.
func (s *scene) reserveColumn(x, wx, r0, r1 int) {
	gy := s.groundY
	wall := s.p.cityHue(theme.CityReserveWall, mLit, 0)
	leaf := s.p.cityHue(theme.CityReserve, mLit, 0)
	dark := s.p.hue(theme.SkyLeafDark, mLit, 0)
	trunk := s.p.hue(theme.SkyTrunk, mLit, 0)
	if wx == r0 || wx == r1-1 {
		for y := gy - 3; y < gy; y++ {
			s.fb.fill(x, s.Y(y), wall, dLamp)
		}
		s.fb.fg(x, s.Y(gy-4), '▄', wall, dLamp)
		return
	}
	s.fb.fg(x, s.Y(gy-1), '▀', wall, dLamp) // the low wall along the front
	h := 2 + int(hash(wx, 805)%2)
	if wx%2 == 0 {
		s.fb.fg(x, s.Y(gy-1), '│', trunk, dLamp)
	}
	for k := 1; k <= h; k++ {
		c := leaf
		if k == h || hash(wx, k, 806)%3 == 0 {
			c = dark
		}
		s.fb.fill(x, s.Y(gy-1-k), c, dLamp)
	}
	s.fb.fg(x, s.Y(gy-2-h), '▄', leaf, dLamp)
	s.fb.set(x, s.Y(gy+2), '"', leaf, theme.Mix(leaf, s.fb.showAt(x, s.Y(gy+2)), 0.6), dGround)
}

// roadLight lights the road's marks: the Digital Age's data running cyan
// down the street, the Fusion Age's plasma pulsing in toward the city.
func (s *scene) roadLight() {
	gy := s.groundY
	h, step := theme.CityData, 1
	if s.fusion() {
		h, step = theme.CityConduit, -1
	}
	c := s.p.cityHue(h, mEmit, 0)
	hot := s.p.cityHue(theme.CityPlasma, mEmit, 0)
	for x := 0; x < s.W; x++ {
		wx := s.wx(x)
		cl := s.fb.at(x, s.Y(gy+1))
		if cl == nil || cl.d != dGround || cl.ch != '─' && cl.ch != ' ' {
			continue
		}
		fg := c
		if (wx+step*s.anim/2)%11 == 0 {
			fg = hot
		}
		if (wx/3)%2 == 0 {
			s.fb.set(x, s.Y(gy+1), '─', fg, cl.bg, dGround)
		}
	}
}

// ------------------------------------------------------------ the roofs

// cityRoofs dresses the visible roofs: the Information and Digital Ages'
// dishes on the taller modern towers, the Information Age's gardens on a
// few roofs, the Digital Age's data glow round its data halls and the
// Fusion Age's glow round its reactors.
func (s *scene) cityRoofs() {
	if !s.inCity {
		return
	}
	key := s.city.Key
	dish := s.p.cityHue(theme.CitySteel, mLit, 0)
	leaf := s.p.cityHue(theme.CityPark, mLit, 0)
	glow := s.p.cityHue(theme.CityGlow, mEmit, 0)
	plasma := s.p.cityHue(theme.CityPlasma, mEmit, 0)
	pulse := 0.5 + float64(0.5*mapmodel.Sin(float64(s.anim)/40))
	for _, li := range s.vis {
		lv := &s.lay.lots[li]
		x0, y0 := lv.x0-s.cam, s.groundY-lv.spr.h
		seed := int(lv.ml.Seed)
		modern := lv.ml.Age >= 12 && !lv.ml.Wonder
		switch {
		case (key == "information_age" || key == "digital_age") && modern && lv.spr.h >= 7 && hash(seed, 811)%3 != 2:
			sx := 1 + int(hash(seed, 812)%uint64(max(1, lv.spr.w-2)))
			if t := lv.spr.topAt(sx); t >= 0 {
				s.fb.fg(x0+sx, s.Y(y0+t-1), '◠', dish, lv.depth)
			}
		}
		if key == "information_age" && modern && hash(seed, 813)%8 == 0 && s.greenKept(hashf(seed, 814)) {
			for sx := 0; sx < lv.spr.w; sx++ { // a garden on the roof
				if t := lv.spr.topAt(sx); t >= 0 && sx%2 == 0 && t == lv.spr.topAt(0) {
					s.fb.fg(x0+sx, s.Y(y0+t-1), '▄', leaf, lv.depth)
				}
			}
		}
		if key == "digital_age" && lv.b != nil && lv.b.Lineage == mapmodel.LinHacker {
			for dx := -2; dx < lv.spr.w+2; dx++ {
				for y := y0 - 1; y < s.groundY; y++ {
					s.fb.tint(x0+dx, s.Y(y), glow, float64(0.1+float64(0.12*pulse)), lv.depth)
				}
			}
		}
		if key == "fusion_age" && lv.b != nil && (lv.b.Key == "fusion_reactor" || lv.b.Key == "fusion_reactor_array") {
			for dx := -3; dx < lv.spr.w+3; dx++ {
				for y := y0 - 2; y < s.groundY; y++ {
					s.fb.tint(x0+dx, s.Y(y), plasma, float64(0.08+float64(0.14*pulse)), lv.depth+1)
				}
			}
		}
	}
}

// districtTint is the finish an Earth-arc district's own buildings wear
// over their period's materials (the Modern, Information and Digital Ages
// share one): the Information Age's towers in indigo data glass, the
// Digital Age's in bare grey concrete. 0 leaves a building as it is.
func (s *scene) districtTint(age, hz int) tcell.Color {
	switch {
	case age == s.cityAge("information_age"):
		return s.p.cityHue(theme.CityServer, mLit, hz)
	case age == s.cityAge("digital_age"):
		return s.p.cityHue(theme.CityConcrete, mLit, hz)
	}
	return 0
}

// cityAge is the index of an Earth-arc age by its key (-1 if unknown).
func (s *scene) cityAge(key string) int {
	if i, ok := s.m.Catalog.AgeIdx[key]; ok {
		return i
	}
	return -1
}

// neonFix keeps the megacity's neon off the greens: a lime or pale green
// sign burns yellow in the Cyberpunk Age and cyan once fusion powers the
// city, so not a cell of the megacity reads as greenery.
func (s *scene) neonFix(c tcell.Color, hz int) tcell.Color {
	r, g, b := c.RGB()
	if r < 0 || g <= r+18 || g <= b+18 {
		return c
	}
	if s.fusion() {
		return s.p.cityHue(theme.CityNeonCyan, mEmit, hz)
	}
	return s.p.cityHue(theme.CityNeonYellow, mEmit, hz)
}

// windowLight is a lit window's light in an Earth-arc age: warm in the
// Modern Age, cold screens in the Information Age, the data glow in the
// Digital Age, white plasma once fusion powers the city (the megacity's
// neon windows stay as the Cyberpunk Age's sprites draw them).
func (s *scene) windowLight(lv *lotView, sx, sy int, warm tcell.Color, hz int) tcell.Color {
	k := hash(int(lv.ml.Seed), sx, sy, 816)
	switch s.city.Key {
	case "information_age":
		if k%2 == 0 {
			return s.p.cityHue(theme.CityScreenCool, mEmit, hz)
		}
	case "digital_age":
		if k%5 < 2 {
			return s.p.cityHue(theme.CityGlow, mEmit, hz)
		}
		if k%5 == 2 {
			return s.p.cityHue(theme.CityScreenCool, mEmit, hz)
		}
	case "fusion_age":
		return s.p.cityHue([3]theme.CityHue{theme.CityPlasma, theme.CityReactor, theme.CityElectric}[k%3], mEmit, hz)
	}
	return warm
}

// leafFade is a sprite's leaf, field or crop cell after the greenery fade:
// green while it is kept, else dry and dead (and in the megacity, the vats'
// synthetic amber).
func (s *scene) leafFade(lv *lotView, sx, sy int, fg tcell.Color, hz int) tcell.Color {
	if s.city.Key != "digital_age" && s.greenKept(hashf(int(lv.ml.Seed), sx, sy, 815)) {
		return fg // (the Digital Age keeps no green but its reserve)
	}
	h := theme.CityLandfill
	if s.megacity() {
		h = theme.CityScreen
	}
	return s.p.cityHue(h, mLit, hz)
}

// ------------------------------------------------------------ the air

// citySmog lays the age's smog band along the horizon: a haze over
// everything behind the building rows, thicker and higher as the smog
// grows (the Information Age's thin band to the megacity's brown-violet
// murk), thinning once fusion cleans the air.
func (s *scene) citySmog() {
	if !s.inCity || s.city.Smog <= 0 {
		return
	}
	lvl := s.city.Smog
	gy := s.groundY
	band := int(float64(gy) * float64(0.2+float64(0.35*lvl)))
	c := s.p.cityHue(theme.CitySmog, mLit, 2)
	if s.cyberpunk() {
		c = theme.Mix(c, s.p.cityHue(theme.CityNeonMagenta, mEmit, 2), 0.25)
	}
	for y := gy - band; y < gy; y++ {
		if y < 0 {
			continue
		}
		u := float64(y-(gy-band)) / float64(max(1, band))
		for x := 0; x < s.W; x++ {
			n := 0.75 + float64(0.25*noise1(float64(s.wx(x))/14+float64(s.anim)/400, y))
			s.fb.tint(x, s.Y(y), c, float64(float64(lvl*0.8*u)*n), dLane2+1)
		}
	}
	if s.megacity() {
		return
	}
	// the data age's haze gathers at the edges of the view, over all but
	// the front row: thin in the Information Age, a murk in the Digital Age
	edge := int(float64(s.W) * float64(0.12+float64(0.3*lvl)))
	for x := 0; x < edge; x++ {
		u := 1 - float64(x)/float64(edge)
		a := float64(lvl * 0.75 * float64(u*u))
		for y := 0; y < gy; y++ {
			s.fb.tint(x, s.Y(y), c, a, dLane0)
			s.fb.tint(s.W-1-x, s.Y(y), c, a, dLane0)
		}
	}
}

// cityRain is the megacity's acid rain, day and night (heavier by night).
func (s *scene) cityRain() {
	drops := s.W * s.S / 40
	if s.p.night > 0.5 {
		drops = s.W * s.S / 24
	}
	acid := theme.Mix(s.p.cityHue(theme.CityAcid, mEmit, 0), s.p.hue(theme.SkyWhite, mEmit, 0), 0.25)
	cols := [3]tcell.Color{acid, theme.Mix(acid, s.p.cityHue(theme.CityNeonMagenta, mEmit, 0), 0.35),
		theme.Mix(acid, s.p.cityHue(theme.CityNeonCyan, mEmit, 0), 0.35)} // rain catching the neon
	for i := 0; i < drops; i++ {
		y := int((hash(i, 1)%uint64(s.S+10) + uint64(s.anim*2)) % uint64(s.S+2))
		x := int(hash(i, 2)%uint64(s.W+s.S)) - y/2
		if y < s.S && x >= 0 && x < s.W {
			s.fb.fg(x, s.Y(y), '╱', cols[i%3], dWeather)
		}
	}
}

// cables strings the Information Age's cable and fiber lines across the
// panorama: spans sagging between masts, a pulse of data running down them
// now and then.
func (s *scene) cables() {
	if !s.inCity || s.city.Key != "information_age" {
		return
	}
	gy := s.groundY
	line := s.p.cityHue(theme.CityConcreteDark, mLit, 1)
	mast := s.p.cityHue(theme.CitySteel, mLit, 1)
	data := s.p.cityHue(theme.CityData, mEmit, 1)
	for k, base := range []int{gy * 52 / 100, gy * 66 / 100} {
		span := 26 + 6*k
		for x := 0; x < s.W; x++ {
			wx := s.wx(x) + k*11
			u := float64(((wx%span)+span)%span) / float64(span) // 0..1 along the span
			sag := float64(3 * (1 - float64((2*u-1)*(2*u-1))))
			y := base + int(math.Round(sag))
			ch := '─'
			if u < 0.03 || u > 0.97 {
				for yy := y; yy < gy; yy++ { // a mast
					s.fb.fg(x, s.Y(yy), '│', mast, dLane1+1)
				}
				s.fb.fg(x, s.Y(y-1), '┬', mast, dLane1+1)
				continue
			}
			c := line
			if (wx+s.anim/2)%37 == 0 {
				c, ch = data, '•'
			}
			s.fb.fg(x, s.Y(y), ch, c, dLane1+1)
		}
	}
}

// ------------------------------------------------------------ the works

// freewayY is the Modern freeway's deck row.
func freewayY(groundY int) int { return viaductY(groundY) - 1 }

// hasFreeway reports whether a frame of the age draws the elevated freeway:
// the Modern to the Digital Age, while the highways stand.
func hasFreeway(age int) bool { return mapmodel.FeatHighway.Info().In(age) }

// freeway draws the elevated freeway: a concrete deck on pillars, its lane
// paint dashed along the top.
func (s *scene) freeway() {
	gy := s.groundY
	y := freewayY(gy)
	deck := s.p.cityHue(theme.CityConcrete, mLit, 1)
	under := s.p.cityHue(theme.CityConcreteDark, mLit, 1)
	paint := s.p.cityHue(theme.CityLaneMark, mLit, 1)
	for x := 0; x < s.W; x++ {
		wx := s.wx(x)
		s.fb.set(x, s.Y(y), '▀', deck, under, dLane0+1)
		if (wx/2)%3 == 0 {
			s.fb.set(x, s.Y(y), '▀', theme.Mix(deck, paint, 0.6), under, dLane0+1)
		}
		if wx%14 == 0 {
			for yy := y + 1; yy < gy; yy++ {
				s.fb.fill(x, s.Y(yy), under, dLane0+1)
			}
		}
	}
}

// ------------------------------------------------------------ traffic

// cityRailYs are the megacity's elevated rails, three to five by the
// frame's height, spread between the rows at their depths.
func cityRailYs(groundY int) []lane {
	n := 3
	if groundY >= 40 {
		n = 4
	}
	if groundY >= 52 {
		n = 5
	}
	out := make([]lane, 0, n)
	for i := 0; i < n; i++ {
		f := 0.24 + float64(0.5*float64(i)/float64(max(1, n-1)))
		y := groundY - max(4+3*i, int(float64(float64(groundY)*f)))
		out = append(out, lane{y, [3]uint8{dLane0, dLane1, dLane2}[i%3] + 2})
	}
	return out
}

// cityAdd places one of the city's vehicles: seed fixes where it starts,
// dx shifts it along (a swarm's formation).
type cityAdd func(t *vtemplate, seed int, ln lane, speed float64, west bool, dx int)

// cityTraffic is the Earth arc's own traffic, a stream of its own (vkCity):
// the freeway's cars, both ways, a sky train or two on every megacity rail,
// drone swarms and hovercars over the megacity.
func cityTraffic(m *mapmodel.Model, groundY int, add cityAdd) {
	look, ok := mapmodel.CityLookAt(m.AgeIdx)
	if !ok {
		return
	}
	busy := m.Activity.Traffic
	sp := func(seed int, lo, hi float64) float64 { return lo + float64((hi-lo)*hashf(seed, 13)) }
	if hasFreeway(m.AgeIdx) {
		ln := lane{freewayY(groundY) - 1, dLane0}
		for i := 0; i < 5+int(float64(busy*5)); i++ {
			seed := 1000 + i
			t := &tCar
			if i%4 == 3 && introduced(&tTruck, m.AgeIdx) {
				t = &tTruck
			}
			add(t, seed, ln, sp(seed, 0.35, 0.7), i%2 == 1, 0)
		}
	}
	if look.Key == "cyberpunk_age" && introduced(&tSkyTrain, m.AgeIdx) {
		for ri, r := range cityRailYs(groundY) {
			for j := 0; j < 2; j++ {
				seed := 1100 + ri*4 + j
				add(&tSkyTrain, seed, lane{r.y - 1, r.d - 1}, sp(seed, 0.5, 0.9), (ri+j)%2 == 0, 0)
			}
		}
	}
	if !look.Night {
		return
	}
	swarms, cars := 4, 6+int(float64(busy*4))
	if look.Key == "fusion_age" {
		swarms, cars = 1, 3
	}
	gy := float64(groundY)
	sky := [3]lane{{int(gy * 0.62), dLane0}, {int(gy * 0.48), dLane1}, {int(gy * 0.34), dLane2}}
	for k := 0; k < swarms && introduced(&tDrone, m.AgeIdx); k++ { // a swarm flies in formation
		seed := 1200 + k*8
		ln := sky[k%3]
		ln.y -= int(hash(seed, 41) % 4)
		speed, west := sp(seed, 0.4, 0.8), hash(seed, 17)%2 == 0
		for d := 0; d < 3; d++ {
			add(&tDrone, seed, lane{ln.y - d%2, ln.d}, speed, west, 2*d)
		}
	}
	for i := 0; i < cars && introduced(&tFlyCar, m.AgeIdx); i++ {
		seed := 1300 + i
		add(&tFlyCar, seed, sky[i%3], sp(seed, 0.3, 0.6), hash(seed, 17)%2 == 0, 0)
	}
}

// compactWindow is the mini map's window light for an Earth-arc age.
func (s *scene) compactWindow() tcell.Color {
	switch s.city.Key {
	case "information_age":
		return s.p.cityHue(theme.CityScreenCool, mEmit, 0)
	case "digital_age":
		return s.p.cityHue(theme.CityGlow, mEmit, 0)
	case "cyberpunk_age":
		return s.p.cityHue(theme.CityNeonMagenta, mEmit, 0)
	case "fusion_age":
		return s.p.cityHue(theme.CityPlasma, mEmit, 0)
	}
	return s.p.hue(theme.SkyWinWarm, mEmit, 0)
}

// compactCity is the Earth arc in the mini map: the megacity's acid rain
// and the space elevator's tether, rising from the Fusion district from the
// Fusion Age on (where the full view's elevatorX stands it).
func (s *scene) compactCity(a0, a1 int) {
	S := s.groundY
	if s.cyberpunk() {
		col := s.p.cityHue(theme.CityAcid, mEmit, 0)
		for i := 0; i < s.W*S/30; i++ {
			y := int((hash(i, 1)%uint64(S+4) + uint64(s.anim)) % uint64(S+1))
			x := int(hash(i, 2)%uint64(s.W+S)) - y/2
			if y < S && x >= 0 && x < s.W {
				s.fb.fg(x, s.Y(y), '╱', col, dWeather)
			}
		}
	}
	fusion := s.cityAge("fusion_age")
	if fusion < 0 || s.m.AgeIdx < fusion || len(s.m.Skyline.Districts) <= fusion || a1 <= a0 {
		return
	}
	d := s.m.Skyline.Districts[fusion]
	x := (d.X0 + d.LandW*3/4 - a0) * s.W / (a1 - a0)
	if x < 0 || x >= s.W {
		return
	}
	tc := s.p.cityHue(theme.CityTether, mEmit, 0)
	for y := 0; y < S; y++ {
		s.fb.fg(x, s.Y(y), '│', tc, dRow0)
	}
	s.fb.fg(x, s.Y(S-1-(s.anim/3)%max(1, S-1)), mapmodel.R(mapmodel.SymClimber, s.tier), tc, dRow0)
}
