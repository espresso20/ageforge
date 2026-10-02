package roguelike

import (
	"strings"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
)

// cityview.go draws the Earth arc's city (city.go): the age's inks, a cell
// of its land or water, a line over it and the smog, plus its legend rows
// and what the cursor says about it.

// cityPal is the city's inks on one theme and age: every city colour made
// legible on the palette background, and the backgrounds the city lays
// under its cells.
type cityPal struct {
	fg                                                [theme.NumCityHues]tcell.Color
	tower, lit, night, toxic, cooling, concrete, quay tcell.Color
	power, ring                                       tcell.Color
	slick, glint                                      tcell.Color // the toxic canals' slicks, the cooling water's glints
	street, block, lawn, glassBg                      tcell.Color // the age's streets, blocks, lawns and glass, under their glyphs
	smog, glow, halo                                  [4]tcell.Color
	arco                                              [3]tcell.Color
	biome                                             [numBiomes]tcell.Color // the region zoom's satellite colours
}

// city biomes, for the region zoom's satellite mosaic
const (
	bNature uint8 = iota
	bBuilt
	bGreen
	bWaste
	bDark
	numBiomes
)

// cityLook tunes the palette to an Earth-arc age: night falls on the map
// for the megacity (on dark themes), the streets, walls and water take the
// age's colours, the farms stop being green once the green is gone, and
// the city's inks are resolved.
func (p *pal) cityLook(look mapmodel.CityLook) {
	if look.Night && !p.Light {
		p.Bg = theme.Mix(p.Bg, theme.CityColor(theme.CityNight), 0.5)
		for c := mapmodel.Class(0); c < mapmodel.NumClasses; c++ {
			p.Fg[c] = theme.Legible(p.Fg[c], p.Bg, mapmodel.ClassSpecs[c].MinContrast)
		}
		p.WaterBg = theme.Mix(p.Bg, theme.MapHueColor(theme.HueWaterDeep), 0.10)
		p.FreshBg = theme.Mix(p.Bg, theme.Color(theme.RolePositive), 0.25)
		p.FlowBg = theme.Mix(p.Bg, theme.Color(theme.RoleWarning), 0.22)
	}
	set := func(c mapmodel.Class, h theme.CityHue, k float64) {
		p.Fg[c] = theme.Legible(theme.Mix(p.Fg[c], theme.CityColor(h), k), p.Bg, mapmodel.ClassSpecs[c].MinContrast)
	}
	switch look.Key {
	case "modern_age":
		set(mapmodel.CRoad, theme.CitySteel, 0.35)
		set(mapmodel.CWall, theme.CityGlass, 0.35)
	case "information_age":
		set(mapmodel.CRoad, theme.CityServer, 0.3)
		set(mapmodel.CWall, theme.CitySteel, 0.45)
		set(mapmodel.CGround, theme.CitySmog, 0.45)
		set(mapmodel.CHill, theme.CitySmog, 0.35)
	case "digital_age":
		set(mapmodel.CRoad, theme.CityConcrete, 0.5)
		set(mapmodel.CWall, theme.CityConcrete, 0.5)
		set(mapmodel.CGround, theme.CityConcrete, 0.6)
		set(mapmodel.CHill, theme.CityConcreteDark, 0.5)
		set(mapmodel.CWater, theme.CitySmog, 0.3)
	case "cyberpunk_age":
		set(mapmodel.CRoad, theme.CityNeonMagenta, 0.3)
		set(mapmodel.CWall, theme.CityNeonMagenta, 0.65)
		set(mapmodel.CWater, theme.CityToxic, 0.8)
		set(mapmodel.CFlora, theme.CityScreen, 0.85) // vat farms, not fields
		set(mapmodel.CGround, theme.CityToxic, 0.4)
		set(mapmodel.CRock, theme.CityConcreteDark, 0.4)
		set(mapmodel.CHill, theme.CityConcreteDark, 0.5)
	case "fusion_age":
		set(mapmodel.CRoad, theme.CityElectric, 0.35)
		set(mapmodel.CWall, theme.CityElectric, 0.65)
		set(mapmodel.CWater, theme.CityElectric, 0.6)
		set(mapmodel.CFlora, theme.CityReactor, 0.85) // bio reactors, not fields
		set(mapmodel.CGround, theme.CityConduit, 0.3)
		set(mapmodel.CRock, theme.CityConcreteDark, 0.4)
		set(mapmodel.CHill, theme.CityConcreteDark, 0.5)
	}
	p.city = newCityPal(p.Bg, p.Light, look.Key)
}

func newCityPal(bg tcell.Color, light bool, key string) *cityPal {
	cp := &cityPal{}
	duo := theme.Active().Duotone
	col := func(h theme.CityHue) tcell.Color {
		c := theme.CityColor(h)
		if duo { // one ink on one paper: keep only how light the colour is
			c = theme.Mix(theme.Color(theme.RoleBackground), theme.Color(theme.RoleText),
				float64(0.2+float64(0.8*theme.RelativeLuminance(c))))
		}
		return c
	}
	for h := range cp.fg {
		cp.fg[h] = theme.Legible(col(theme.CityHue(h)), bg, 3)
	}
	k := func(dark, lt float64) float64 {
		if light {
			return lt
		}
		return dark
	}
	cp.tower = theme.Mix(bg, col(theme.CityTower), k(0.8, 0.14))
	cp.lit = theme.Mix(bg, col(theme.CityTowerLit), k(0.5, 0.22))
	cp.night = theme.Mix(bg, col(theme.CityNight), k(0.6, 0.04))
	if key == "fusion_age" { // the powered city: deep blue towers, navy streets
		cp.tower = theme.Mix(bg, col(theme.CityCooling), k(0.62, 0.14))
		cp.lit = theme.Mix(bg, col(theme.CityElectric), k(0.42, 0.2))
		cp.night = theme.Mix(bg, theme.Mix(col(theme.CityNight), col(theme.CityCooling), 0.35), k(0.7, 0.05))
	}
	// each age paves its streets and builds its blocks in its own colour:
	// the Modern Age's asphalt among green lawns, the Information Age's
	// indigo, the Digital Age's pale concrete
	cp.lawn = theme.Mix(bg, col(theme.CityPark), k(0.16, 0.12))
	cp.glassBg = theme.Mix(bg, col(theme.CityGlass), k(0.3, 0.16))
	switch key {
	case "modern_age":
		cp.street = theme.Mix(bg, col(theme.CityAsphalt), k(0.7, 0.18))
		cp.block = cp.glassBg
	case "information_age":
		cp.street = theme.Mix(bg, col(theme.CityServer), k(0.32, 0.14))
		cp.block = theme.Mix(bg, col(theme.CityConcreteDark), k(0.55, 0.16))
	case "digital_age":
		cp.street = theme.Mix(bg, col(theme.CityConcreteDark), k(0.5, 0.16))
		cp.block = theme.Mix(bg, col(theme.CityConcrete), k(0.34, 0.2))
	default:
		cp.street, cp.block = cp.night, cp.tower
	}
	cp.toxic = theme.Mix(bg, col(theme.CityToxicDeep), k(0.85, 0.22))
	cp.cooling = theme.Mix(bg, col(theme.CityCooling), k(0.55, 0.18))
	cp.concrete = theme.Mix(bg, col(theme.CityConcrete), k(0.16, 0.12))
	cp.quay = theme.Mix(bg, col(theme.CityConcreteDark), k(0.5, 0.16))
	cp.power = theme.Mix(bg, col(theme.CityElectric), k(0.14, 0.08))
	cp.ring = theme.Mix(bg, col(theme.CityElectric), k(0.28, 0.14))
	cp.slick = theme.Legible(theme.Mix(cp.toxic, col(theme.CityToxic), 0.6), cp.toxic, 2)
	cp.glint = theme.Legible(theme.Mix(cp.cooling, col(theme.CityElectric), 0.7), cp.cooling, 2)
	for l := 1; l <= 3; l++ {
		cp.smog[l] = theme.Mix(bg, col(theme.CitySmog), float64(k(0.08, 0.07)*float64(l)))
		cp.glow[l] = theme.Mix(bg, col(theme.CityGlow), float64(k(0.09, 0.07)*float64(l)))
		cp.halo[l] = theme.Mix(bg, col(theme.CityElectric), float64(k(0.1, 0.06)*float64(l)))
	}
	cp.arco = [3]tcell.Color{theme.Mix(bg, col(theme.CityTower), k(0.95, 0.18)),
		theme.Mix(bg, col(theme.CityTowerLit), k(0.4, 0.22)), theme.Mix(bg, col(theme.CityTowerLit), k(0.7, 0.32))}
	cp.biome = [numBiomes]tcell.Color{
		bNature: 0,
		bBuilt:  theme.Mix(bg, col(theme.CityConcrete), k(0.26, 0.2)),
		bGreen:  theme.Mix(bg, col(theme.CityPark), k(0.3, 0.22)),
		bWaste:  theme.Mix(bg, col(theme.CityLandfill), k(0.3, 0.22)),
		bDark:   theme.Mix(bg, col(theme.CityServer), k(0.34, 0.22)),
	}
	return cp
}

// cityLegend says which city features take a legend row (the rest are
// read off their neighbours: an alley is a megablock's street).
var cityLegend = [mapmodel.NumCityFeatures]bool{
	mapmodel.FeatSuburbs: true, mapmodel.FeatPark: true, mapmodel.FeatHighway: true, mapmodel.FeatGlassTower: true,
	mapmodel.FeatServerFarm: true, mapmodel.FeatDish: true, mapmodel.FeatFiber: true, mapmodel.FeatGarden: true,
	mapmodel.FeatOffice: true, mapmodel.FeatSmog: true, mapmodel.FeatConcrete: true, mapmodel.FeatLandfill: true,
	mapmodel.FeatReserve: true, mapmodel.FeatDataGlow: true, mapmodel.FeatMegablock: true, mapmodel.FeatCorpTower: true,
	mapmodel.FeatArcology: true, mapmodel.FeatNeonSign: true,
	mapmodel.FeatCanal: true, mapmodel.FeatSkyRail: true, mapmodel.FeatMaglevLine: true, mapmodel.FeatReactor: true,
	mapmodel.FeatConduit: true, mapmodel.FeatLaunchTower: true, mapmodel.FeatTether: true, mapmodel.FeatCooling: true,
}

// regCity claims a city feature's legend row.
func (v *view) regCity(f mapmodel.CityFeature, g glyph) {
	if f != mapmodel.FeatNone && cityLegend[f] {
		v.regG(lgCity+lgID(f), g)
	}
}

// megacity reports the megacity's looks (the Cyberpunk and Fusion Ages).
func (cl *cityLand) megacity() bool { return cl.look.Night }

func (cl *cityLand) fusion() bool { return cl.key == "fusion_age" }

// cityWater is the megacity's water: toxic canals, then clean cooling
// channels once fusion powers the city.
func (v *view) cityWater(x, y int, t mapmodel.Terrain) glyph {
	cp := v.pal.city
	g := glyph{r: '≈', fg: cp.slick, bg: cp.toxic, sal: [3]int{12, 14, 35}[t]}
	f := mapmodel.FeatCanal
	if v.sc.city.fusion() {
		g.fg, g.bg, f = cp.glint, cp.cooling, mapmodel.FeatCooling
	}
	if t == mapmodel.TShallow {
		g.r = '~'
	}
	v.regCity(f, g)
	// mostly dark, still water; a slick or a glint drifts over it now and then
	h := mapmodel.Hash(v.sc.w.Seed, 12, int64(x), int64(y))
	if (h+uint64(v.anim/6))%[3]uint64{5, 3, 2}[t] != 0 {
		g.r = ' '
	}
	return g
}

// cityGlyph draws a cell of the city's land or the megacity's water; ok is
// false where the land is as it was.
func (v *view) cityGlyph(x, y int, t mapmodel.Terrain) (glyph, bool) {
	s := v.sc
	cl, cp := s.city, v.pal.city
	i := y*s.w.W + x
	if t.Water() {
		if cl.megacity() {
			return v.cityWater(x, y, t), true
		}
		return glyph{}, false
	}
	u := cl.use[i]
	h := mapmodel.Hash(s.w.Seed, 11, int64(x), int64(y))
	f := v.anim
	var g glyph
	switch u {
	case uNone:
		return glyph{}, false
	case uStreet:
		g = glyph{r: ' ', bg: cp.street, sal: 2}
	case uHouse: // slate roofs, and a red one here and there, on their lawns
		g = glyph{r: '⌂', fg: cp.fg[theme.CitySteel], bg: cp.lawn, sal: 14}
		if h%3 == 0 {
			g.fg = cp.fg[theme.CityRoof]
		}
	case uYard:
		g = glyph{r: ' ', fg: cp.fg[theme.CityHedge], bg: cp.lawn, sal: 3}
		if h%3 == 0 {
			g.r = [2]rune{',', '\''}[h>>4%2]
		}
	case uPark:
		g = glyph{r: '♣', fg: cp.fg[theme.CityPark], sal: 18}
		if h%5 == 0 {
			g.r = 'τ'
		}
	case uFence:
		g = glyph{r: '┊', fg: cp.fg[theme.CityHedge], sal: 10}
		if cl.parkAt(s, x, y-1) || cl.parkAt(s, x, y+1) {
			g.r = '┈'
		}
	case uGlass:
		g = glyph{r: '▦', fg: cp.fg[theme.CityGlass], bg: cp.glassBg, attr: tcell.AttrBold, sal: 22}
	case uOffice:
		g = glyph{r: '▪', fg: cp.fg[theme.CitySteel], bg: cp.block, sal: 14}
		if cl.key == "digital_age" { // blank concrete slabs, a window here and there
			g.r, g.fg = ' ', cp.fg[theme.CityConcreteDark]
			if h%4 == 0 {
				g.r = '▫'
			}
		} else if mapmodel.Hash(int64(i), int64(f/3))%37 == 0 { // a screen flickers in a window
			g.fg, g.attr = cp.fg[[2]theme.CityHue{theme.CityScreen, theme.CityScreenCool}[h%2]], tcell.AttrBold
		}
	case uServer:
		g = glyph{r: '▤', fg: cp.fg[theme.CityServer], bg: cp.block, sal: 16}
		if mapmodel.Hash(int64(i), int64(f/2))%9 == 0 {
			g.fg = cp.fg[theme.CityData]
		}
	case uDish:
		g = glyph{r: mapmodel.R(mapmodel.SymDish, v.tier), fg: cp.fg[theme.CitySteel], bg: cp.block, attr: tcell.AttrBold, sal: 22}
	case uGarden: // a garden on a roof: a speck of green among the vents
		g = glyph{r: ' ', fg: cp.fg[theme.CityPark], bg: cp.block, sal: 10}
		if h%4 == 0 {
			g.r = '♧'
		} else if h%4 == 1 {
			g.r = '·'
		}
	case uScrub:
		g = glyph{r: ' ', fg: cp.fg[theme.CitySmog], sal: 2}
		if h%4 == 0 {
			g.r = [2]rune{'\'', '.'}[h>>4%2]
		}
	case uConcrete:
		g = glyph{r: ' ', fg: cp.fg[theme.CityConcreteDark], bg: cp.concrete, sal: 4}
		if h%11 == 0 {
			g.r = '·'
		}
	case uLandfill:
		g = glyph{r: []rune("▒▒%·░")[h%5], fg: cp.fg[theme.CityLandfill], sal: 6}
	case uDataHall:
		g = glyph{r: '▧', fg: cp.fg[theme.CityGlow], bg: cp.block, attr: tcell.AttrBold, sal: 18}
	case uReserve:
		g = glyph{r: '♣', fg: cp.fg[theme.CityReserve], attr: tcell.AttrBold, sal: 24}
		if h%5 == 0 {
			g.r = 'τ'
		}
	case uReserveWall:
		mk := 0
		for bit, d := range dirs4 {
			if s.w.In(x+d[0], y+d[1]) && cl.use[(y+d[1])*s.w.W+x+d[0]] == uReserveWall {
				mk |= 1 << bit
			}
		}
		g = glyph{r: boxDouble[mk], fg: cp.fg[theme.CityReserveWall], sal: 30}
		if mk == 0 {
			g.r = '■'
		}
	case uAlley:
		g = glyph{r: ' ', bg: cp.night, sal: 2}
	case uTrash:
		g = glyph{r: []rune("·,·%")[h%4], fg: cp.fg[theme.CityRust], bg: cp.night, sal: 3}
	case uMega:
		g = glyph{r: ' ', bg: cp.tower, sal: 8}
	case uLit:
		g = glyph{r: '·', fg: v.litInk(cl, cl.aux[i]), bg: cp.tower, sal: 9}
		if h%3 == 0 {
			g.r = '▪'
		}
	case uVent:
		g = glyph{r: '○', fg: cp.fg[theme.CitySteam], bg: cp.tower, sal: 9}
	case uNeon:
		g = glyph{r: neonRunes[int(cl.aux[i]>>2)%len(neonRunes)], fg: v.neonInk(cl, int(cl.aux[i]&3)), bg: cp.tower,
			attr: tcell.AttrBold, sal: 12}
		if mapmodel.Hash(int64(i), int64(f/2))%23 == 0 { // a flicker
			g.fg, g.attr = cp.lit, 0
		}
	case uAd:
		word := holoAds[(int(cl.aux[i]>>3)+f/40)%len(holoAds)]
		g = glyph{r: ' ', fg: v.neonInk(cl, int(cl.aux[i]>>3)), bg: cp.night, attr: tcell.AttrBold, sal: 12}
		if p := int(cl.aux[i] & 7); p < len(word) {
			g.r = rune(word[p])
		}
	case uCorp:
		c := megacorps[int(cl.aux[i]>>2)%len(megacorps)]
		g = glyph{r: rune(c.initials[int(cl.aux[i]&3)%len(c.initials)]), fg: v.neonInk(cl, int(cl.aux[i]>>2)),
			bg: cp.lit, attr: tcell.AttrBold, sal: 33}
		if mapmodel.Hash(int64(cl.aux[i]>>2), int64(f/3))%41 == 0 {
			g.attr = 0
		}
	case uArco:
		switch cl.aux[i] {
		case 0:
			g = glyph{r: '▫', fg: cp.fg[theme.CityTowerLit], bg: cp.arco[0], sal: 26}
		case 1:
			g = glyph{r: '▒', fg: v.neonInk(cl, 1), bg: cp.arco[1], sal: 28}
		default:
			g = glyph{r: mapmodel.R(mapmodel.SymArcology, v.tier), fg: v.neonInk(cl, 0), bg: cp.arco[2],
				attr: tcell.AttrBold, sal: 34}
		}
	case uScrap:
		g = glyph{r: '‰', fg: cp.fg[theme.CityRust], bg: cp.night, sal: 6}
	case uToxic:
		g = glyph{r: '░', fg: cp.fg[theme.CityToxic], bg: cp.toxic, sal: 5}
	case uSlag:
		g = glyph{r: '▲', fg: cp.fg[theme.CityConcreteDark], sal: 20}
	case uQuay:
		g = glyph{r: ' ', bg: cp.quay, sal: 4}
	case uYardPower:
		g = glyph{r: '≡', fg: cp.fg[theme.CityConduit], bg: cp.power, sal: 6}
	case uReactor:
		g = glyph{r: mapmodel.R(mapmodel.SymReactor, v.tier), fg: cp.fg[theme.CityPlasma], bg: cp.ring,
			attr: tcell.AttrBold, sal: 40}
		if (f/3+x)%4 == 0 {
			g.fg = cp.fg[theme.CityReactor]
		}
	case uRing:
		a := int(cl.aux[i])
		r := runeAt(reactorRing[min(a/9, 2)], a%9)
		g = glyph{r: r, fg: cp.fg[theme.CityReactor], bg: cp.ring, sal: 36}
		if r == '═' {
			g.fg, g.attr = cp.fg[theme.CityPlasma], tcell.AttrBold
		}
		if (a%9+f/2)%8 == 0 { // the plasma runs round the ring
			g.fg, g.attr = cp.fg[theme.CityPlasma], tcell.AttrBold
		}
	case uLaunch:
		a := int(cl.aux[i])
		r := runeAt(launchPad[min(a/3, 2)], a%3)
		g = glyph{r: r, fg: cp.fg[theme.CitySteel], sal: 34}
		switch r {
		case '▲':
			g.r, g.fg, g.attr = mapmodel.R(mapmodel.SymLaunchTower, v.tier), cp.fg[theme.CityPlasma], tcell.AttrBold
			if v.tier != mapmodel.TierNerd {
				g.r = '▲'
			}
		case '·':
			g.fg = cp.fg[theme.CityTaillight]
			if (f/4)%3 == 0 {
				g.r = ' '
			}
		case '▀':
			g.fg = cp.fg[theme.CityConcreteDark]
		}
	default:
		return glyph{}, false
	}
	v.regCity(useFeature[u], g)
	if lvl := cl.glow[i]; lvl > 0 && g.bg == 0 || lvl > 0 && cl.fusion() && u != uRing && u != uReactor {
		ph := float64(mapmodel.Sin(float64(f)/48 + float64(h%100)/100)) // the glow breathes
		if l := int(lvl) + int(ph+0.5); l > 0 {
			if cl.fusion() {
				g.bg = cp.halo[min(l, 3)]
			} else if g.bg == 0 {
				g.bg = cp.glow[min(l, 3)]
			}
		}
	}
	return g, true
}

// parkAt reports a park cell at (x, y).
func (cl *cityLand) parkAt(s *scene, x, y int) bool {
	return s.w.In(x, y) && cl.use[y*s.w.W+x] == uPark
}

// litInk is a lit window's colour: warm screens and cold ones, magenta in
// the megacity; white and electric blue once fusion powers it.
func (v *view) litInk(cl *cityLand, a uint8) tcell.Color {
	cp := v.pal.city
	if cl.fusion() {
		return cp.fg[[4]theme.CityHue{theme.CityPlasma, theme.CityElectric, theme.CityReactor, theme.CityScreenCool}[a%4]]
	}
	return cp.fg[[4]theme.CityHue{theme.CityScreen, theme.CityNeonCyan, theme.CityNeonMagenta, theme.CityScreenCool}[a%4]]
}

// neonInk is a neon colour: magenta, cyan and yellow in the megacity;
// cyan, white and electric blue once fusion powers it.
func (v *view) neonInk(cl *cityLand, k int) tcell.Color {
	cp := v.pal.city
	if cl.fusion() {
		return cp.fg[[3]theme.CityHue{theme.CityNeonCyan, theme.CityPlasma, theme.CityElectric}[abs(k)%3]]
	}
	return cp.fg[[3]theme.CityHue{theme.CityNeonMagenta, theme.CityNeonCyan, theme.CityNeonYellow}[abs(k)%3]]
}

// cityLineAt is the line drawn on cell i of kind c, if one is.
func (cl *cityLand) lineAt(s *scene, i int) (*cityLine, bool) {
	top := cl.ov[i]
	if top == 0 {
		return nil, false
	}
	return &cl.lines[top-1], true
}

// cityLineGlyph draws the top line on a cell.
func (v *view) cityLineGlyph(x, y, i int, t mapmodel.Terrain) (glyph, bool) {
	s := v.sc
	cl, cp := s.city, v.pal.city
	ln, ok := cl.lineAt(s, i)
	if !ok {
		return glyph{}, false
	}
	mk := int(cl.ovm[i])
	g := glyph{fg: cp.fg[ln.ink], attr: tcell.AttrBold, sal: 44}
	switch ln.feat {
	case mapmodel.FeatHighway:
		g.r, g.fg, g.sal = boxDouble[mk], cp.fg[theme.CityConcrete], 38
		if mk == 0 {
			g.r = '═'
		}
	case mapmodel.FeatFiber:
		g.r, g.attr, g.sal = '┄', 0, 6
		if mk&10 == 0 {
			g.r = '┆'
		}
		if (x+y+v.anim/2)%11 == 0 { // a pulse of data
			g.attr = tcell.AttrBold
		}
	case mapmodel.FeatSkyRail:
		g.r = boxDouble[mk]
	case mapmodel.FeatMaglevLine:
		g.r, g.bg = boxDouble[mk], cp.power
	case mapmodel.FeatConduit:
		g.r, g.attr, g.sal = boxLight[mk], 0, 36
		if (x+y-v.anim/2)%7 == 0 { // plasma pulses in toward the core
			g.fg, g.attr = cp.fg[theme.CityPlasma], tcell.AttrBold
		}
	case mapmodel.FeatTether: // a beam of light from the square to the sky
		g.r, g.fg, g.bg, g.sal = '║', cp.fg[theme.CityPlasma], cp.ring, 46
		if mk&5 == 0 {
			g.r = '│'
		}
	default:
		return glyph{}, false
	}
	if g.r == boxDouble[0] || g.r == boxLight[0] { // a lone cell: a short run
		g.r = '═'
		if ln.feat == mapmodel.FeatConduit {
			g.r = '─'
		}
	}
	if t.Water() {
		g.bg = v.pal.WaterBg
		if cl.megacity() {
			g.bg = cp.toxic
			if cl.fusion() {
				g.bg = cp.cooling
			}
		}
	}
	v.regCity(ln.feat, glyph{r: g.r, fg: g.fg, attr: g.attr})
	return g, true
}

// smogged lays the age's smog over a cell: a haze tint on its background.
func (v *view) smogged(i int, g glyph) glyph {
	cl := v.sc.city
	lvl := cl.smog[i]
	if lvl == 0 {
		return g
	}
	cp := v.pal.city
	if g.bg == 0 {
		g.bg = cp.smog[lvl]
	} else {
		g.bg = theme.Mix(g.bg, theme.CityColor(theme.CitySmog), float64(0.06*float64(lvl)))
	}
	if !cl.megacity() { // the megacity's smog is just its air
		v.regCity(mapmodel.FeatSmog, glyph{r: ' ', bg: cp.smog[2]})
	}
	return g
}

// describeLine says what a line is.
func describeLine(ln *cityLine, n int) insp {
	info := ln.feat.Info()
	in := insp{Title: info.Title, Lines: []string{info.Line(n)}}
	switch ln.feat {
	case mapmodel.FeatSkyRail:
		in.Lines = []string{ln.name + ", high over the street"}
	case mapmodel.FeatMaglevLine:
		in.Lines = []string{ln.name + ", humming with power"}
	}
	return in
}

// describeCity says what a cell of the city is: its line, or its land.
func (v *view) describeCity(s *scene, p mapmodel.Pt, k kind) (insp, bool) {
	cl := s.city
	i := p.Y*s.w.W + p.X
	n := int(mapmodel.Hash(int64(p.X), int64(p.Y)) % 97)
	if ln, ok := cl.lineAt(s, i); ok && (k == kNone || ln.elev) {
		return describeLine(ln, n), true
	}
	if k != kNone {
		return insp{}, false
	}
	t := s.w.At(p.X, p.Y)
	if t.Water() {
		if !cl.megacity() {
			return insp{}, false
		}
		f := mapmodel.FeatCanal
		if cl.fusion() {
			f = mapmodel.FeatCooling
		}
		return insp{Title: f.Info().Title, Lines: []string{f.Info().Line(n)}}, true
	}
	u := cl.use[i]
	switch u {
	case uNone:
		return insp{}, false
	case uStreet:
		return insp{Title: "Street", Lines: []string{"between the blocks of " + s.w.Name}}, true
	case uCorp:
		c := megacorps[int(cl.aux[i]>>2)%len(megacorps)]
		return insp{Title: c.name + " tower", Lines: []string{c.initials + " in neon, a block high",
			mapmodel.FeatCorpTower.Info().Line(n)}}, true
	case uAd:
		word := holoAds[(int(cl.aux[i]>>3)+v.anim/40)%len(holoAds)]
		return insp{Title: "Holo-ad", Lines: []string{"flashing " + strings.TrimSpace(word) + " over the street"}}, true
	case uArco:
		return insp{Title: "Arcology", Lines: []string{mapmodel.FeatArcology.Info().Line(n), "home to a hundred thousand"}}, true
	}
	f := useFeature[u]
	if f == mapmodel.FeatNone {
		return insp{}, false
	}
	return insp{Title: f.Info().Title, Lines: []string{f.Info().Line(n)}}, true
}

// biomeAt is what the region zoom's satellite mosaic sees on cell k: the
// city's concrete, its green, its waste or its server halls, or the land
// as it was.
func (cl *cityLand) biomeAt(s *scene, k int) uint8 {
	switch u := cl.use[k]; {
	case u == uNone:
		if cl.megacity() && s.w.T[k].Land() && s.cells[k].k != kNone {
			return bBuilt // the town, inside the megacity
		}
		return bNature
	case u.green():
		return bGreen
	case u == uLandfill || u == uScrub || u == uScrap || u == uToxic:
		return bWaste
	case u == uServer || u == uDataHall || u == uMega || u == uLit || u == uAlley:
		return bDark
	}
	return bBuilt
}

// cityLights is the megacity seen from orbit at night: a carpet of light,
// neon and window glow, twinkling.
func (v *view) cityLights(cx, cy int, h float64) glyph {
	cp := v.pal.city
	out := glyph{r: ' '}
	switch {
	case h < 0.12:
		out.r, out.fg, out.attr = '•', v.neonInk(v.sc.city, int(h*1000)), tcell.AttrBold
	case h < 0.42:
		out.r, out.fg = '·', v.litInk(v.sc.city, uint8(h*100))
	case h < 0.5:
		out.r, out.fg = '∙', cp.fg[theme.CityTowerLit]
	}
	if out.r != ' ' && mapmodel.Hash(int64(cx), int64(cy), int64(v.anim/4))%13 == 0 {
		out.r = '∙' // a twinkle
	}
	return out
}
