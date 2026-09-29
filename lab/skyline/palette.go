package main

import (
	"math"

	"github.com/espresso20/ageforge/theme"
)

// palette.go turns a handful of dials (the age family, the time of day, the
// theme, catastrophe pressure) into a small table of colour classes once
// per frame. Cells never do their own colour maths: a cell is (material,
// slot, depth) and looks its colour up here, which is why a theme switch or
// the 16-colour fallback costs nothing per cell.

// Material is one building finish. Buildings pick a family by their own age
// and a variant by hash, so a Georgian terrace and a glass tower standing
// side by side each keep their period.
type Material struct {
	Wall, Roof, Trim, Metal, Glass, GlassHi, Neon1, Neon2, Neon3, Glow RGB
}

// family index by building age index (0..21).
func familyOf(age int) int {
	switch {
	case age <= 1:
		return 0
	case age <= 3:
		return 1
	case age == 4:
		return 2
	case age == 5:
		return 3
	case age <= 7:
		return 4
	case age <= 9:
		return 5
	case age <= 11:
		return 6
	case age <= 14:
		return 7
	case age == 15:
		return 8
	case age <= 17:
		return 9
	default:
		return 10
	}
}

const variants = 3

func m(wall, roof, trim, metal, glass, glassHi, n1, n2, n3, glow uint32) Material {
	return Material{hex(wall), hex(roof), hex(trim), hex(metal), hex(glass), hex(glassHi), hex(n1), hex(n2), hex(n3), hex(glow)}
}

// materials[family][variant]
var materials = [][variants]Material{
	// 0 primitive/stone: hide, wood, thatch
	{m(0x7a5236, 0xa8843f, 0xd8c8a0, 0x6b6258, 0x3a2a20, 0x5a4535, 0xffb347, 0xff7b2e, 0xffe08a, 0xffa53a),
		m(0x6a4630, 0xc19a4b, 0xcbb892, 0x6b6258, 0x3a2a20, 0x5a4535, 0xffb347, 0xff7b2e, 0xffe08a, 0xffa53a),
		m(0x8a6242, 0x9a7a3a, 0xe0d2ae, 0x6b6258, 0x3a2a20, 0x5a4535, 0xffb347, 0xff7b2e, 0xffe08a, 0xffa53a)},
	// 1 bronze/iron: mudbrick, flat roofs
	{m(0xb38456, 0x8a6a45, 0xe0cfa4, 0xa0703a, 0x3b2b20, 0x5a4535, 0xffb347, 0xd9822b, 0xffe08a, 0xffa53a),
		m(0xc49a6c, 0x7d5e3e, 0xeadcb8, 0xa0703a, 0x3b2b20, 0x5a4535, 0xffb347, 0xd9822b, 0xffe08a, 0xffa53a),
		m(0x9d9384, 0x7a6a55, 0xd8ccb0, 0xa0703a, 0x3b2b20, 0x5a4535, 0xffb347, 0xd9822b, 0xffe08a, 0xffa53a)},
	// 2 classical: marble, terracotta
	{m(0xe2dccb, 0xb4553a, 0xfff6e2, 0xb08d57, 0x4a3b30, 0x6a5a4a, 0xffc766, 0xd9822b, 0xffe08a, 0xffb050),
		m(0xd6cdb8, 0xa04a34, 0xf4ecd8, 0xb08d57, 0x4a3b30, 0x6a5a4a, 0xffc766, 0xd9822b, 0xffe08a, 0xffb050),
		m(0xcfc3a8, 0xc2653f, 0xfaf0dc, 0xb08d57, 0x4a3b30, 0x6a5a4a, 0xffc766, 0xd9822b, 0xffe08a, 0xffb050)},
	// 3 medieval: plaster and beams, rubble stone, slate
	{m(0xd3c6a6, 0x6f3a2a, 0x4a3423, 0x6d6a66, 0x2e2a30, 0x4a4650, 0xffc766, 0xff8c3a, 0xffe08a, 0xffb050),
		m(0x8f8a80, 0x4d5260, 0x5d5850, 0x6d6a66, 0x2e2a30, 0x4a4650, 0xffc766, 0xff8c3a, 0xffe08a, 0xffb050),
		m(0x9c958a, 0x5a3a30, 0x6a6258, 0x6d6a66, 0x2e2a30, 0x4a4650, 0xffc766, 0xff8c3a, 0xffe08a, 0xffb050)},
	// 4 renaissance/colonial: ochre stucco, copper domes, whitewash
	{m(0xd6a861, 0x9b4a2f, 0xefe3c4, 0x5f9c88, 0x2e3440, 0x5a6878, 0xffcf70, 0xff8c3a, 0xffe08a, 0xffb050),
		m(0xe6dfd0, 0x4a5a6a, 0xffffff, 0x5f9c88, 0x2e3440, 0x5a6878, 0xffcf70, 0xff8c3a, 0xffe08a, 0xffb050),
		m(0xc98f5f, 0x7a3a28, 0xf0e2c0, 0x5f9c88, 0x2e3440, 0x5a6878, 0xffcf70, 0xff8c3a, 0xffe08a, 0xffb050)},
	// 5 industrial/victorian: brick, soot, iron
	{m(0x8f3c2a, 0x3c3838, 0xb8a48a, 0x55575c, 0x2a2c30, 0x50555c, 0xffd27a, 0xff6a2a, 0xffe08a, 0xff8c3a),
		m(0x6f3226, 0x2e2c2c, 0xa89880, 0x4a4c50, 0x2a2c30, 0x50555c, 0xffd27a, 0xff6a2a, 0xffe08a, 0xff8c3a),
		m(0x7c6a5a, 0x3a3a3e, 0xc0b098, 0x55575c, 0x2a2c30, 0x50555c, 0xffd27a, 0xff6a2a, 0xffe08a, 0xff8c3a)},
	// 6 electric/atomic: limestone deco, concrete, chrome
	{m(0xc9bfa6, 0x5e5e62, 0xc7a24a, 0x7a7f86, 0x2c3440, 0x6a8098, 0xffe28a, 0x5ae0ff, 0xff5a5a, 0xfff0b0),
		m(0x9a9a95, 0x4e4e52, 0xd0d0c8, 0x7a7f86, 0x2c3440, 0x6a8098, 0xffe28a, 0x5ae0ff, 0xff5a5a, 0xfff0b0),
		m(0xb0a48c, 0x5a5048, 0xe0c870, 0x7a7f86, 0x2c3440, 0x6a8098, 0xffe28a, 0x5ae0ff, 0xff5a5a, 0xfff0b0)},
	// 7 modern/information/digital: concrete and glass
	{m(0x8e9499, 0x5a6066, 0xc8ced4, 0xb0b7bd, 0x3f6f96, 0x8fc0e0, 0xfff0b0, 0x46d0ff, 0xff4a6a, 0xe8f6ff),
		m(0x4a6a86, 0x3a4a5a, 0xa8c8e0, 0xb0b7bd, 0x2f5f86, 0x9fd0f0, 0xfff0b0, 0x46d0ff, 0xff4a6a, 0xe8f6ff),
		m(0x6e767e, 0x4a5056, 0xd8dde2, 0xb0b7bd, 0x35607a, 0x7fb0d0, 0xfff0b0, 0x46d0ff, 0xff4a6a, 0xe8f6ff)},
	// 8 cyberpunk: dark towers, neon
	{m(0x2c2a45, 0x1d1b2e, 0x4a4570, 0x5a5878, 0x3a2f63, 0x6a5aa3, 0xff3ea5, 0x29f0ff, 0xb4ff39, 0xff9ad5),
		m(0x23213a, 0x16142a, 0x3d3a60, 0x5a5878, 0x2a3f63, 0x5a8ab3, 0x29f0ff, 0xff3ea5, 0xffe23a, 0x9af5ff),
		m(0x34304a, 0x201c30, 0x524a78, 0x5a5878, 0x402a58, 0x8a5aa3, 0xb4ff39, 0xff3ea5, 0x29f0ff, 0xe0ffa0)},
	// 9 fusion/space: white composites, glass, cyan glow
	{m(0xd9dee6, 0xaab4c2, 0xf4f8ff, 0x8a96a8, 0x4fa3c9, 0xaee8f8, 0x7ff0ff, 0xffb35a, 0xa0ffa0, 0x9ff6ff),
		m(0xc2cad6, 0x8894a6, 0xe8eef8, 0x8a96a8, 0x3f8fb9, 0x9edcf0, 0x7ff0ff, 0xffb35a, 0xa0ffa0, 0x9ff6ff),
		m(0xe6e2d8, 0xb0a894, 0xffffff, 0x8a96a8, 0x5fb3a9, 0xbef0e8, 0x7ff0ff, 0xffb35a, 0xa0ffa0, 0x9ff6ff)},
	// 10 cosmic: iridescent alloys, violet and gold light
	{m(0xe4e0f6, 0xb9b2e0, 0xfff4d0, 0x9a92c8, 0x5a4ab0, 0xc8b8ff, 0xb28cff, 0xffd479, 0x7ff0ff, 0xe8d8ff),
		m(0xc9c2ec, 0x9a90d0, 0xffe8a8, 0x9a92c8, 0x4a3aa0, 0xb8a8ff, 0xffd479, 0xb28cff, 0x7ff0ff, 0xfff0c0),
		m(0xf0ecff, 0xd0c8f0, 0xd8f8ff, 0x9a92c8, 0x3a6ab0, 0xa8d8ff, 0x7ff0ff, 0xb28cff, 0xffd479, 0xd8f8ff)},
}

// Nature and fixed colours, lit like everything else.
var (
	cLeaf      = hex(0x3f7a3a)
	cLeafDark  = hex(0x2a5530)
	cTrunk     = hex(0x5a3d26)
	cField1    = hex(0x9fae4a)
	cField2    = hex(0xc9a84a)
	cRock      = hex(0x8a8580)
	cRockDark  = hex(0x5e5a58)
	cInk       = hex(0x141218)
	cWater     = hex(0x2d5f8a)
	cSmoke     = hex(0x8e8a88)
	cFire      = hex(0xff7a1a)
	cBeacon    = hex(0xff3030)
	cWinLitA   = hex(0xffd27a)
	cWinLitB   = hex(0xfff0c0)
	cWinLitNeo = hex(0x9af5ff)
)

// skyKey is a four-stop vertical gradient, top to horizon.
type skyKey [4]RGB

func sk(a, b, c, d uint32) skyKey { return skyKey{hex(a), hex(b), hex(c), hex(d)} }

func (a skyKey) lerp(b skyKey, t float64) skyKey {
	var o skyKey
	for i := range a {
		o[i] = a[i].Lerp(b[i], t)
	}
	return o
}

// skies per epoch: day, dusk, night.
type epochSky struct{ day, dusk, dawn, night skyKey }

var skies = map[string]epochSky{
	"stone_era": {sk(0x2f6fd1, 0x5a9be6, 0x9fcdf2, 0xdcefff), sk(0x1c1b4a, 0x5b3072, 0xd0585a, 0xffb257),
		sk(0x243a78, 0x6a5a9a, 0xe0889a, 0xffc890), sk(0x03060f, 0x0a1330, 0x16224a, 0x27305c)},
	"iron_era": {sk(0x2a68c8, 0x5898e0, 0xa0cdf0, 0xe8f0f0), sk(0x1c1b4a, 0x5b3072, 0xd8604a, 0xffba60),
		sk(0x243a78, 0x6a5a9a, 0xe0889a, 0xffc890), sk(0x03060f, 0x0a1330, 0x16224a, 0x2a3058)},
	"steel_era": {sk(0x5a7896, 0x7f93a6, 0xb2b0a4, 0xd8c8a8), sk(0x2a2240, 0x6a3a5a, 0xc0604a, 0xe89a58),
		sk(0x3a4060, 0x7a6a80, 0xc08a80, 0xe0b080), sk(0x06070c, 0x121624, 0x262430, 0x4a3428)},
	"electric_era": {sk(0x3f6fa8, 0x6a92bc, 0xa8bccc, 0xdcd6c4), sk(0x201c46, 0x60306a, 0xd05a50, 0xffa860),
		sk(0x2a3a70, 0x6a5a90, 0xd08890, 0xffc088), sk(0x04060e, 0x0e1428, 0x1e2440, 0x42384a)},
	"digital_era": {sk(0x2a6ad8, 0x4f94ec, 0x98caf6, 0xe0f2ff), sk(0x1a1a50, 0x582f78, 0xe0585e, 0xffb060),
		sk(0x243a80, 0x6a5aa0, 0xe888a0, 0xffc890), sk(0x02040c, 0x0a1230, 0x16204a, 0x2a3a66)},
	"neon_era": {sk(0x1f5d7a, 0x3f8a98, 0x88b8b0, 0xd7e6c8), sk(0x1a0c3a, 0x5a1a6a, 0xd0306a, 0xff8a50),
		sk(0x1a1a50, 0x4a2a78, 0xb04a8a, 0xff9a80), sk(0x07031a, 0x1c0838, 0x40104f, 0x8a1a60)},
	"cosmic_era": {sk(0x14103a, 0x3a3a9a, 0x8a8fe0, 0xe0d8ff), sk(0x0e0826, 0x3a1a60, 0xa04a90, 0xffa080),
		sk(0x0e0e30, 0x3a3a80, 0x9a70c0, 0xffc0a0), sk(0x020108, 0x0a0620, 0x1a0e3a, 0x3a1e5a)},
}

// Scene dials for one frame.
type Dials struct {
	AgeIdx   int
	Epoch    string
	TOD      float64 // time of day 0..1, 0 = midnight
	Theme    theme.Theme
	Colors   int     // 0/256 = truecolour or 256, 16 = ANSI
	Pressure float64 // catastrophe pressure 0..1 (reddens the horizon)
	Weather  int     // 0 clear, 1 cloudy, 2 rain, 3 storm
}

// Pal is the frame's resolved colour table.
type Pal struct {
	D        Dials
	Sky      skyKey
	Day      float64 // 1 at noon, 0 at night
	Twilight float64
	Night    float64
	Light    RGB // ambient light tint
	Stars    float64
	light    bool // light theme
	mono     bool
	duoLo    RGB
	duoHi    RGB
	cache    map[uint32]RGB
}

func twilightAmt(t float64) float64 {
	g := func(c float64) float64 { d := (t - c) / 0.045; return math.Exp(-d * d) }
	return math.Max(g(0.25), g(0.75))
}

func newPal(d Dials) *Pal {
	p := &Pal{D: d, cache: map[uint32]RGB{}}
	es, ok := skies[d.Epoch]
	if !ok {
		es = skies["stone_era"]
	}
	t := d.TOD
	p.Day = smoothstep(0.21, 0.31, t) - smoothstep(0.69, 0.79, t)
	p.Twilight = twilightAmt(t)
	p.Night = 1 - math.Max(p.Day, p.Twilight*0.85)
	sky := es.night.lerp(es.day, p.Day)
	tw := es.dusk
	if t < 0.5 {
		tw = es.dawn
	}
	p.Sky = sky.lerp(tw, p.Twilight*0.85)
	if d.Weather >= 2 { // overcast: flatten the sky toward grey
		g := p.Sky[2].Gray().Mul(0.85)
		for i := range p.Sky {
			p.Sky[i] = p.Sky[i].Lerp(g, 0.55)
		}
	}
	if d.Pressure > 0 { // the harbinger's weather: a bruised horizon
		p.Sky[3] = p.Sky[3].Lerp(hex(0xa0302a), 0.45*d.Pressure)
		p.Sky[2] = p.Sky[2].Lerp(hex(0x6a2030), 0.30*d.Pressure)
	}
	p.Light = hex(0x5a6690).Lerp(hex(0xfffaf0), p.Day)
	p.Light = p.Light.Lerp(hex(0xe8a080), p.Twilight*0.6)
	p.Stars = smoothstep(0.35, 0.9, p.Night)
	if d.AgeIdx >= 18 { // thin air: stars show in the upper sky by day too
		p.Stars = math.Max(p.Stars, 0.35)
	}

	p.light = d.Theme.IsLight()
	if p.light {
		// A light page wants a daylight sky: lift everything toward the
		// canvas, and turn night into a blue hour so silhouettes read as
		// ink on paper instead of a black hole in a white UI.
		bg := fromTC(d.Theme.Color(theme.RoleBackground))
		for i := range p.Sky {
			floor := hex(0x8a90bc).Lerp(bg, float64(i)/4)
			v := p.Sky[i]
			if v.Luma() < floor.Luma() {
				v = v.Lerp(floor, 0.75)
			}
			p.Sky[i] = v.Lerp(bg, 0.28)
		}
		p.Stars *= 0.3
	}
	switch d.Theme.Key {
	case "monochrome", "parchment":
		p.mono = true
		p.duoLo = fromTC(d.Theme.Color(theme.RoleBackground))
		p.duoHi = fromTC(d.Theme.Color(theme.RoleText))
		if p.duoLo.Luma() > p.duoHi.Luma() {
			p.duoLo, p.duoHi = p.duoHi, p.duoLo
		}
	}
	return p
}

// final applies the theme's duotone and the colour-depth fallback. Every
// colour leaves the palette through here.
func (p *Pal) final(c RGB) RGB {
	if p.mono {
		c = p.duoLo.Lerp(p.duoHi, c.Luma()/255)
	}
	if p.D.Colors == 16 {
		c = nearestANSI(c.Sat(1.9).clamp())
	}
	return c
}

// haze by depth: 0 front ... 4 far hills.
var hazeK = [...]float64{0, 0.10, 0.22, 0.52, 0.72}

// SkyAt is the sky colour at scene row y of h.
func (p *Pal) SkyAt(y, h int) RGB {
	f := float64(y) / float64(max(1, h-1)) * 3
	i := int(f)
	if i >= 3 {
		return p.Sky[3]
	}
	return p.Sky[i].Lerp(p.Sky[i+1], f-float64(i))
}

// horizon is where haze fades to.
func (p *Pal) horizon() RGB { return p.Sky[3].Lerp(p.Sky[2], 0.35) }

func (p *Pal) raw(mat Material, s Slot) (RGB, bool) {
	switch s {
	case SWall:
		return mat.Wall, false
	case SWallLit:
		return mat.Wall.Mul(1.16).Lerp(hex(0xffffff), 0.05), false
	case SWallShade:
		return mat.Wall.Mul(0.70), false
	case SWallDark:
		return mat.Wall.Mul(0.42), false
	case SRoof:
		return mat.Roof, false
	case SRoofShade:
		return mat.Roof.Mul(0.68), false
	case STrim:
		return mat.Trim, false
	case SWin:
		return mat.Wall.Mul(0.30).Lerp(mat.Glass, 0.5), false
	case SMetal:
		return mat.Metal, false
	case SMetalDark:
		return mat.Metal.Mul(0.6), false
	case SGlass:
		return mat.Glass, false
	case SGlassHi:
		return mat.GlassHi, false
	case SNeon1:
		return mat.Neon1, true
	case SNeon2:
		return mat.Neon2, true
	case SNeon3:
		return mat.Neon3, true
	case SGlow:
		return mat.Glow, true
	case SBeacon:
		return cBeacon, true
	case SLeaf:
		return cLeaf, false
	case SLeafDark:
		return cLeafDark, false
	case STrunk:
		return cTrunk, false
	case SField1:
		return cField1, false
	case SField2:
		return cField2, false
	case SRock:
		return cRock, false
	case SRockDark:
		return cRockDark, false
	case SFire:
		return cFire, true
	case SInk:
		return cInk, false
	case SWater:
		return cWater, false
	case SSmoke:
		return cSmoke, false
	}
	return mat.Wall, false
}

// Col resolves (material, slot, depth) for this frame, cached.
func (p *Pal) Col(fam, variant int, s Slot, depth int) RGB {
	key := uint32(fam)<<24 | uint32(variant)<<16 | uint32(s)<<8 | uint32(depth)
	if c, ok := p.cache[key]; ok {
		return c
	}
	mat := materials[fam][variant%variants]
	c, emissive := p.raw(mat, s)
	if !emissive {
		c = c.Tint(p.Light)
		if p.light {
			// ink on paper: keep facades from going muddy on a pale sky
			c = c.Lerp(c.Mul(0.8), p.Night)
		}
	} else if p.Day > 0.5 {
		c = c.Lerp(mat.Wall.Tint(p.Light), 0.35*p.Day) // neon washes out in daylight
	}
	c = c.Lerp(p.horizon(), hazeK[min(depth, len(hazeK)-1)])
	c = p.final(c)
	p.cache[key] = c
	return c
}

// Emit resolves an emissive colour (lit windows, fire, stars), unaffected
// by ambient light but still hazed and themed.
func (p *Pal) Emit(c RGB, depth int) RGB {
	c = c.Lerp(p.horizon(), hazeK[min(depth, len(hazeK)-1)]*0.8)
	return p.final(c)
}

// Lit returns an ambient-lit, hazed colour for scenery (hills, ground).
func (p *Pal) Lit(c RGB, depth int) RGB {
	c = c.Tint(p.Light)
	c = c.Lerp(p.horizon(), hazeK[min(depth, len(hazeK)-1)])
	return p.final(c)
}

// Hill is scenery at an explicit haze k (0 crisp .. 1 lost in the sky).
func (p *Pal) Hill(c RGB, k float64) RGB {
	return p.final(c.Tint(p.Light).Lerp(p.horizon(), k))
}

// ansi16 is the classic CGA/VGA text palette BBS art was drawn in.
var ansi16 = []RGB{
	hex(0x000000), hex(0xaa0000), hex(0x00aa00), hex(0xaa5500), hex(0x0000aa), hex(0xaa00aa), hex(0x00aaaa), hex(0xaaaaaa),
	hex(0x555555), hex(0xff5555), hex(0x55ff55), hex(0xffff55), hex(0x5555ff), hex(0xff55ff), hex(0x55ffff), hex(0xffffff),
}

func nearestANSI(c RGB) RGB {
	best, bd := ansi16[0], math.MaxFloat64
	for _, a := range ansi16 {
		dr, dg, db := (c.R-a.R)*0.30, (c.G-a.G)*0.59, (c.B-a.B)*0.11
		if d := dr*dr + dg*dg + db*db; d < bd {
			best, bd = a, d
		}
	}
	return best
}
