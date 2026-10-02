package skyline

import (
	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
)

// palette.go turns a few dials (epoch, time of day, weather, catastrophe
// pressure, the theme) into the frame's colour table. Cells never do their
// own colour maths: a sprite cell is (material, slot, depth) and looks its
// colour up here, cached, so a theme switch or a new hour costs nothing
// per cell.

type hueMode uint8

const (
	mLit  hueMode = iota // lit by the ambient light and hazed
	mEmit                // self-lit, hazed a little
	mFlat                // as it is (then themed)
	numModes
)

// depth indexes for haze: 0 front row … 4 far hills.
const numHaze = 5

var hazeK = [numHaze]float64{0, 0.10, 0.22, 0.52, 0.72}

type pal struct {
	day, twilight, night, stars float64
	light, mono                 bool
	gloom                       bool // the megacity's smog-dark day: neon never washes out
	duoLo, duoHi                tcell.Color
	sky                         [4]tcell.Color
	lightTint, horizon          tcell.Color
	hazeMul                     float64
	mats                        [theme.SkyFamilies][theme.SkyVariants]theme.SkyMaterial
	matCache                    []tcell.Color
	hueCache                    []tcell.Color
	cityCache                   []tcell.Color // the Earth arc's city colours (city.go)
}

// newPal builds the frame palette for a model on the active theme.
func newPal(m *mapmodel.Model) *pal {
	c := m.Clock
	p := &pal{day: c.Daylight, twilight: c.Twilight, night: c.Night, hazeMul: 1,
		matCache: make([]tcell.Color, theme.SkyFamilies*theme.SkyVariants*int(numSlots)*numHaze),
		hueCache: make([]tcell.Color, int(numHues)*int(numModes)*numHaze)}
	for f := 0; f < theme.SkyFamilies; f++ {
		for v := 0; v < theme.SkyVariants; v++ {
			p.mats[f][v] = theme.SkyMaterialFor(f, v)
		}
	}
	keys := theme.SkyKeysFor(m.Epoch)
	tw := keys.Dusk
	if c.TOD < 0.5 {
		tw = keys.Dawn
	}
	for i := range p.sky {
		v := theme.Mix(keys.Night[i], keys.Day[i], p.day)
		p.sky[i] = theme.Mix(v, tw[i], float64(p.twilight*0.85))
	}
	switch m.Weather.Kind {
	case mapmodel.Rain, mapmodel.Storm, mapmodel.Snow:
		g := theme.Shade(gray(p.sky[2]), 0.85)
		for i := range p.sky {
			p.sky[i] = theme.Mix(p.sky[i], g, 0.55)
		}
	case mapmodel.Cloudy:
		g := gray(p.sky[2])
		for i := range p.sky {
			p.sky[i] = theme.Mix(p.sky[i], g, 0.3)
		}
	case mapmodel.Smog:
		sm := theme.SkyColor(theme.SkySmokeHeavy)
		for i := range p.sky {
			p.sky[i] = theme.Mix(p.sky[i], sm, float64(0.25+float64(0.1*float64(i))))
		}
		p.hazeMul = 1.4
	}
	if pr := m.Catastrophe.Pressure; pr > 0 {
		p.sky[3] = theme.Mix(p.sky[3], theme.SkyColor(theme.SkyBruise), float64(0.45*pr))
		p.sky[2] = theme.Mix(p.sky[2], theme.SkyColor(theme.SkyBruiseHigh), float64(0.30*pr))
	}
	p.citySky(m)
	p.lightTint = theme.Mix(theme.SkyColor(theme.SkyLightNight), theme.SkyColor(theme.SkyLightDay), p.day)
	p.lightTint = theme.Mix(p.lightTint, theme.SkyColor(theme.SkyLightDusk), float64(p.twilight*0.6))
	p.stars = smoothstep(0.35, 0.9, p.night)
	if m.AgeIdx >= 18 && p.stars < 0.35 {
		p.stars = 0.35 // thin air: the late sky shows stars by day
	}
	bg := theme.Color(theme.RoleBackground)
	p.light = theme.IsLight()
	if p.light {
		// A light page wants a daylight sky: lift everything toward the page
		// and turn night into a blue hour, so silhouettes read as ink on
		// paper instead of a black hole in a white UI.
		for i := range p.sky {
			floor := theme.Mix(theme.SkyColor(theme.SkyBlueHour), bg, float64(i)/4)
			v := p.sky[i]
			if luma(v) < luma(floor) {
				v = theme.Mix(v, floor, 0.75)
			}
			p.sky[i] = theme.Mix(v, bg, 0.28)
		}
		p.stars *= 0.3
	}
	if theme.Active().Duotone {
		p.mono = true
		p.duoLo, p.duoHi = bg, theme.Color(theme.RoleText)
		if luma(p.duoLo) > luma(p.duoHi) {
			p.duoLo, p.duoHi = p.duoHi, p.duoLo
		}
		for i := range p.sky {
			p.sky[i] = p.final(p.sky[i])
		}
	}
	p.horizon = theme.Mix(p.sky[3], p.sky[2], 0.35)
	return p
}

// final applies the theme's duotone. Every scene colour leaves through it.
func (p *pal) final(c tcell.Color) tcell.Color {
	if p.mono {
		return theme.Mix(p.duoLo, p.duoHi, luma(c)/255)
	}
	return c
}

func (p *pal) haze(d int) float64 {
	if d < 0 {
		d = 0
	}
	if d >= numHaze {
		d = numHaze - 1
	}
	k := float64(hazeK[d] * p.hazeMul)
	if k > 0.9 {
		k = 0.9
	}
	return k
}

func (p *pal) raw(m theme.SkyMaterial, s slot) tcell.Color {
	switch s {
	case sWall:
		return m.Wall
	case sWallLit:
		return theme.Mix(theme.Shade(m.Wall, 1.16), theme.SkyColor(theme.SkyWhite), 0.05)
	case sWallShade:
		return theme.Shade(m.Wall, 0.70)
	case sWallDark:
		return theme.Shade(m.Wall, 0.42)
	case sRoof:
		return m.Roof
	case sRoofShade:
		return theme.Shade(m.Roof, 0.68)
	case sTrim:
		return m.Trim
	case sWin:
		return theme.Mix(theme.Shade(m.Wall, 0.30), m.Glass, 0.5)
	case sMetal:
		return m.Metal
	case sMetalDark:
		return theme.Shade(m.Metal, 0.6)
	case sGlass:
		return m.Glass
	case sGlassHi:
		return m.GlassHi
	case sNeon1:
		return m.Neon1
	case sNeon2:
		return m.Neon2
	case sNeon3:
		return m.Neon3
	case sGlow:
		return m.Glow
	}
	return theme.SkyColor(slotHue[s])
}

// slotHue gives the nature slots their fixed colours.
var slotHue = [numSlots]theme.SkyHue{
	sBeacon: theme.SkyBeacon, sLeaf: theme.SkyLeaf, sLeafDark: theme.SkyLeafDark, sTrunk: theme.SkyTrunk,
	sField1: theme.SkyField1, sField2: theme.SkyField2, sRock: theme.SkyRock, sRockDark: theme.SkyRockDark,
	sFire: theme.SkyFire, sInk: theme.SkyInk, sWater: theme.SkyWater, sSmoke: theme.SkySmoke,
}

// col resolves (material, slot, depth) for this frame.
func (p *pal) col(fam, variant int, s slot, depth int) tcell.Color {
	fam, variant = clampInt(fam, 0, theme.SkyFamilies-1), clampInt(variant, 0, theme.SkyVariants-1)
	depth = clampInt(depth, 0, numHaze-1)
	i := ((fam*theme.SkyVariants+variant)*int(numSlots)+int(s))*numHaze + depth
	if c := p.matCache[i]; c != 0 {
		return c
	}
	m := p.mats[fam][variant]
	c := p.raw(m, s)
	if !s.emissive() {
		c = theme.Tint(c, p.lightTint)
		if p.light {
			c = theme.Mix(c, theme.Shade(c, 0.8), p.night) // ink on paper
		}
	} else if p.day > 0.5 && !p.gloom {
		c = theme.Mix(c, theme.Tint(m.Wall, p.lightTint), float64(0.35*p.day)) // neon washes out by day
	}
	c = p.final(theme.Mix(c, p.horizon, p.haze(depth)))
	p.matCache[i] = c
	return c
}

// hue resolves a fixed scene colour in a mode at a haze depth.
func (p *pal) hue(h theme.SkyHue, mode hueMode, depth int) tcell.Color {
	depth = clampInt(depth, 0, numHaze-1)
	i := (int(h)*int(numModes)+int(mode))*numHaze + depth
	if i < 0 || i >= len(p.hueCache) {
		return p.final(theme.SkyColor(h))
	}
	if c := p.hueCache[i]; c != 0 {
		return c
	}
	c := theme.SkyColor(h)
	switch mode {
	case mLit:
		c = theme.Mix(theme.Tint(c, p.lightTint), p.horizon, p.haze(depth))
	case mEmit:
		c = theme.Mix(c, p.horizon, float64(p.haze(depth)*0.8))
	}
	c = p.final(c)
	p.hueCache[i] = c
	return c
}

// hill is scenery at an explicit haze k.
func (p *pal) hill(h theme.SkyHue, k float64) tcell.Color {
	return p.final(theme.Mix(theme.Tint(theme.SkyColor(h), p.lightTint), p.horizon, k))
}

// skyAt is the sky colour at scene row y of rows.
func (p *pal) skyAt(y, rows int) tcell.Color {
	f := float64(y) / float64(max(1, rows-1)) * 3
	i := int(f)
	if i >= 3 {
		return p.sky[3]
	}
	if i < 0 {
		return p.sky[0]
	}
	return theme.Mix(p.sky[i], p.sky[i+1], f-float64(i))
}

// numHues is the size of the fixed-hue table.
const numHues = theme.SkyExhaust + 1

func luma(c tcell.Color) float64 {
	r, g, b := c.RGB()
	if r < 0 {
		return 0
	}
	return float64(0.2126*float64(r)) + float64(0.7152*float64(g)) + float64(0.0722*float64(b))
}

func gray(c tcell.Color) tcell.Color {
	l := luma(c) / 255
	return theme.Mix(theme.SkyColor(theme.SkyInk), theme.SkyColor(theme.SkyWhite), l)
}

func smoothstep(a, b, x float64) float64 {
	t := (x - a) / (b - a)
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	return float64(t*t) * (3 - float64(2*t))
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// hash is a small stateless integer hash, the source of every "random"
// choice in the scene, so a state always draws the same city.
func hash(v ...int) uint64 {
	h := uint64(0x9e3779b97f4a7c15)
	for _, x := range v {
		h ^= uint64(int64(x)) + 0x9e3779b97f4a7c15 + (h << 6) + (h >> 2)
		h *= 0xbf58476d1ce4e5b9
		h ^= h >> 31
	}
	return h ^ h>>32
}

func hashf(v ...int) float64 { return float64(hash(v...)%100000) / 100000 }

// citySky tints the Earth arc's skies: smog browning the horizon from the
// Information Age, the megacity's day a smog-dark dusk the sun never gets
// through, the powered city's horizon glowing electric blue.
func (p *pal) citySky(m *mapmodel.Model) {
	look, ok := mapmodel.CityLookAt(m.AgeIdx)
	if !ok {
		return
	}
	c := theme.CityColor
	switch look.Key {
	case "cyberpunk_age":
		murk := [4]tcell.Color{c(theme.CityNight), c(theme.CityTower), theme.Mix(c(theme.CityTower), c(theme.CitySmog), 0.5),
			theme.Mix(c(theme.CitySmog), c(theme.CityNeonMagenta), 0.3)}
		for i := range p.sky {
			p.sky[i] = theme.Mix(p.sky[i], murk[i], float64(0.7*p.day))
		}
		p.gloom = true
	case "fusion_age": // the plasma's glow on the horizon, electric blue
		p.sky[3] = theme.Mix(p.sky[3], c(theme.CityElectric), float64(0.15+float64(0.4*p.night)))
		p.sky[2] = theme.Mix(p.sky[2], c(theme.CityElectric), float64(0.22*p.night))
		p.sky[1] = theme.Mix(p.sky[1], c(theme.CityCooling), float64(0.3*p.night))
	case "modern_age": // a crisp blue day; by night the glass city lights its own horizon
		glow := theme.Mix(c(theme.CityHeadlight), c(theme.CityScreen), 0.5)
		p.sky[3] = theme.Mix(p.sky[3], glow, float64(0.45*p.night))
		p.sky[2] = theme.Mix(p.sky[2], glow, float64(0.2*p.night))
	default: // the smog: a haze paling the sky and browning the horizon, a grey-brown sky in the Digital Age
		k := [4]float64{0.26, 0.36, 0.5, 0.62}
		light := c(theme.CityScreen) // the Information Age's sodium light in the smog by night
		if look.Key == "digital_age" {
			k = [4]float64{0.35, 0.5, 0.62, 0.72}
			light = c(theme.CityGlow) // the Digital Age's data glow
		}
		smog := theme.Mix(c(theme.CitySmog), light, float64(0.45*p.night))
		for i := range p.sky {
			p.sky[i] = theme.Mix(p.sky[i], smog, k[i])
		}
	}
}
