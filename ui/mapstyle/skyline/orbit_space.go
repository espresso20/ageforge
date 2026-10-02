package skyline

import (
	"math"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
)

// orbit_space.go is the Space Age: actual space. The ground has dropped
// away: the planet's curved night side fills the bottom of the view, its
// atmosphere a bright line along the limb, its cities a band of amber
// lights under it, densest under the districts with the most built; above
// it the station's truss runs from the oldest district to the frontier
// with every building on it as a module; the space elevator's tether
// climbs from the limb through the truss and off the top; the moon hangs
// upper right with its base lit by the military; an asteroid belt crosses
// the far sky.

// spaceBackdrop draws everything behind the lots.
func (o *orb) spaceBackdrop() {
	o.stars(0, o.S, 0.022, 0.02, 991)
	o.belt()
	o.moon()
	limb := o.limb()
	o.planet(limb)
	o.tether(limb)
	o.truss()
}

// limb is the planet's limb per screen column (limbOf).
func (o *orb) limb() []float64 { return limbOf(o.W, o.S, o.groundY) }

// limbOf is the planet's limb per screen column of a view W wide with S
// scene rows and its baseline at row gy, in half rows from the top of the
// scene: a circle whose top sits three rows under the baseline at the
// middle of the view and falls away to the bottom corners.
func limbOf(W, S, gy int) []float64 {
	top := float64(2 * (gy + 3))
	edge := float64(2 * (S - 1))
	if edge <= top+2 {
		edge = top + 2
	}
	half := float64(W) / 2
	drop := edge - top
	R := (half*half + drop*drop) / (2 * drop)
	cy := top + R
	out := make([]float64, W)
	for x := range out {
		dx := float64(x) + 0.5 - half
		if v := R*R - dx*dx; v > 0 {
			out[x] = cy - math.Sqrt(v)
		} else {
			out[x] = 1e9
		}
	}
	return out
}

// planet paints the night side under the limb: the atmosphere's line and
// halo, ocean and continents scrolling at a slow parallax, clouds drifting,
// and the cities' lights.
func (o *orb) planet(limb []float64) {
	S := o.S
	void := o.c(mapmodel.InkVoid, iBack, 0)
	glow := o.c(mapmodel.InkGlow, iEmit, 0)
	haze := o.c(mapmodel.InkCloud2, iEmit, 0)
	ocean := o.c(mapmodel.InkSurface, iBack, 0)
	land := o.c(mapmodel.InkSurface2, iBack, 0)
	cloud := o.c(mapmodel.InkCloud, iBack, 0)
	landHi := o.resolve(theme.Shade(theme.SpaceColor(o.pal[mapmodel.InkSurface2]), 1.45), iBack, 0)
	shelf := o.resolve(theme.Shade(theme.SpaceColor(o.pal[mapmodel.InkSurface]), 1.5), iBack, 0)
	city := o.c(mapmodel.InkLight, iEmit, 0)
	cityHi := theme.Mix(city, o.c(mapmodel.InkStarBright, iEmit, 0), 0.45)
	halo := [3]tcell.Color{theme.Mix(void, glow, 0.08), theme.Mix(void, glow, 0.16), theme.Mix(void, glow, 0.3)}
	seed := int64(o.m.Seed%9973) + 41
	dens := o.cityDensity()
	pu := float64(o.cam) * 0.22 // the surface turns slower than the station passes
	drift := float64(o.anim) * 0.04
	pix := func(x int, py int, lp float64) (tcell.Color, int) {
		d := float64(py) + 0.5 - lp
		switch {
		case d < -3:
			return 0, 0 // space
		case d < 0:
			return halo[clampInt(int(d+3), 0, 2)], 1
		case d < 1.1:
			return glow, 2
		case d < 2.4:
			return haze, 2
		}
		return ocean, 3
	}
	for x := 0; x < o.W; x++ {
		lp := limb[x]
		if lp > float64(2*S) {
			continue
		}
		y0 := max(0, int(lp/2)-2)
		u := float64(x) + pu
		for y := y0; y < S; y++ {
			ct, kt := pix(x, 2*y, lp)
			cb, kb := pix(x, 2*y+1, lp)
			if kt == 0 && kb == 0 {
				continue
			}
			if kt == 0 {
				ct = void
			}
			if kb == 3 && kt == 3 {
				// the surface: continents, clouds, city lights
				dd := float64(2*y) + 1 - lp
				v := math.Sqrt(math.Max(0, dd)) * 2.2 // foreshortened toward the limb
				n := float64(0.65*mapmodel.Noise(seed, u*0.05, v*0.21)) + float64(0.35*mapmodel.Noise(seed+7, u*0.13, v*0.5))
				bg := ocean
				isLand := n > 0.53
				if isLand {
					// plains, forests and highlands
					bg = theme.Mix(land, landHi, clamp01((mapmodel.Noise(seed+11, u*0.3, v*0.8)-0.35)*1.6))
				} else if n > 0.47 {
					bg = theme.Mix(ocean, shelf, (n-0.47)/0.06) // the shallows off a coast
				}
				cn := float64(0.7*mapmodel.Noise(seed+19, (u+drift)*0.09, v*0.3)) + float64(0.3*mapmodel.Noise(seed+23, (u+drift)*0.25, v*0.9))
				ch, fg := ' ', bg
				if cn > 0.6 {
					bg = theme.Mix(bg, cloud, clamp01((cn-0.6)*3.2))
					if cn > 0.69 {
						ch, fg = '░', theme.Mix(cloud, o.c(mapmodel.InkStar, iEmit, 2), 0.25)
					}
				}
				if p := dens[x]; p > 0 && dd < 20 && cn < 0.72 {
					// cities cluster on the coasts and plains, thickest
					// just under the limb where the station looks down
					k := p * (1 - dd/24)
					if !isLand {
						k *= 0.1
					}
					k *= 0.45 + 2.3*math.Max(0, mapmodel.Noise(seed+31, u*0.22, v*0.55)-0.35)
					if hv := hashf(int(u), y, 77); hv < k {
						bg = theme.Mix(bg, city, 0.14) // the glow of a city at night
						ch, fg = '·', city
						if hv < k*0.2 {
							ch, fg = '•', cityHi
						} else if hv < k*0.5 {
							ch = '∙'
						}
					}
				}
				// the night deepens away from the limb
				if dd > 6 {
					bg = theme.Mix(bg, void, math.Min(0.42, (dd-6)/60))
					if ch == ' ' || ch == '░' {
						fg = theme.Mix(fg, void, math.Min(0.42, (dd-6)/60))
					}
				}
				o.fb.set(x, o.Y(y), ch, fg, bg, dPlanet)
				continue
			}
			if kb == 0 {
				cb = void
			}
			if ct == cb {
				o.fb.set(x, o.Y(y), ' ', ct, ct, dPlanet)
			} else {
				o.fb.set(x, o.Y(y), '▀', ct, cb, dPlanet)
			}
		}
	}
}

// cityDensity is how thick the cities' lights are under each screen
// column: in proportion to the lots of the district above it, so more
// buildings mean more lights.
func (o *orb) cityDensity() []float64 {
	out := make([]float64, o.W)
	ds := o.m.Skyline.Districts
	if len(ds) == 0 {
		return out
	}
	per := make([]int, len(ds))
	for i := range o.m.Skyline.Lots {
		l := &o.m.Skyline.Lots[i]
		if l.Age >= 0 && l.Age < len(per) {
			per[l.Age]++
		}
	}
	di := -1
	for x := range out {
		wx := x + o.cam
		if di < 0 || di >= len(ds) || !inDistrict(ds[di], wx) {
			di = o.m.Skyline.DistrictAt(wx)
		}
		if di < 0 {
			continue
		}
		d := ds[di]
		n := 0
		if d.Age >= 0 && d.Age < len(per) {
			n = per[d.Age]
		}
		out[x] = clamp01(0.08 + float64(n)/float64(max(1, d.LandW))*0.9)
	}
	return out
}

// belt is the asteroid belt across the far sky, clumped and drifting.
func (o *orb) belt() {
	gy := o.groundY
	b0 := max(1, gy/6)
	b1 := b0 + max(2, gy/7)
	rock := o.c(mapmodel.InkRock, iLit, 2)
	dark := o.c(mapmodel.InkRockDark, iLit, 2)
	shift := float64(o.cam)*0.05 + float64(o.anim)*0.02
	small := mapmodel.R(mapmodel.SymSkyAsteroid, o.tier)
	for y := b0; y <= b1 && y < o.S; y++ {
		mid := 1 - math.Abs(float64(2*y-b0-b1))/float64(max(1, b1-b0+1))
		for x := 0; x < o.W; x++ {
			bx := int(float64(x) + shift)
			p := float64(0.05+0.22*mapmodel.Noise(77, float64(bx)/23, 0.5)) * (0.4 + 0.6*mid)
			hv := hashf(bx, y, 313)
			if hv >= p {
				continue
			}
			ch, c := '·', dark
			switch k := hash(bx, y, 5) % 10; {
			case k == 0:
				ch, c = '▪', rock
			case k < 3:
				ch, c = small, rock
			case k < 5:
				ch = '∙'
			case k < 6:
				ch, c = '▖', rock
			}
			o.fb.fg(x, o.Y(y), ch, c, dBelt)
		}
	}
}

// moon is the moon, upper right with a little parallax: a cratered disc,
// lit from the left, its base's lights growing with the military.
func (o *orb) moon() {
	gy := o.groundY
	rr := clampInt(gy/7, 2, 5)
	cx := float64(o.W)*0.84 - float64(o.cam)*0.01
	cpy := float64(2*max(rr+1, gy/4)) + 1
	R := float64(2 * rr)
	moonC := theme.SpaceColor(o.pal[mapmodel.InkMoon])
	lit := o.resolve(moonC, iLit, 1)
	shade := o.resolve(theme.Shade(moonC, 0.62), iLit, 1)
	night := o.resolve(theme.Shade(moonC, 0.34), iLit, 1)
	crater := o.resolve(theme.Shade(moonC, 0.78), iLit, 1)
	px := func(x, py int) (tcell.Color, bool) {
		dx := (float64(x) + 0.5 - cx) / R
		dy := (float64(py) + 0.5 - cpy) / R
		if float64(dx*dx)+float64(dy*dy) > 1 {
			return 0, false
		}
		c := lit
		switch {
		case dx > 0.45:
			c = night
		case dx > 0.1:
			c = shade
		}
		if mapmodel.Noise(int64(o.m.Seed%101)+3, (dx+1)*3.1, (dy+1)*3.1) > 0.66 && dx < 0.45 {
			c = crater
		}
		return c, true
	}
	x0, x1 := int(cx-R)-1, int(cx+R)+1
	y0, y1 := int((cpy-R)/2)-1, int((cpy+R)/2)+1
	for y := max(0, y0); y <= y1 && y < o.groundY; y++ {
		for x := max(0, x0); x <= x1 && x < o.W; x++ {
			ct, ot := px(x, 2*y)
			cb, ob := px(x, 2*y+1)
			switch {
			case ot && ob:
				if ct == cb {
					o.fb.fill(x, o.Y(y), ct, dMoon)
				} else {
					o.fb.set(x, o.Y(y), '▀', ct, cb, dMoon)
				}
			case ot:
				o.fb.fg(x, o.Y(y), '▀', ct, dMoon)
			case ob:
				o.fb.fg(x, o.Y(y), '▄', cb, dMoon)
			}
		}
	}
	// the moon base: its lights grow with the military lineage
	n := 0
	for _, l := range o.m.SkyLineages() {
		if l.Key == mapmodel.LinMilitary {
			n = len(l.Units)
		}
	}
	if n == 0 {
		return
	}
	light := o.c(mapmodel.InkLight, iEmit, 1)
	by := int(cpy+R*0.45) / 2
	for i := 0; i < min(n, 10); i++ {
		x := int(cx-R*0.55) + i%5
		y := by - i/5
		if c := o.fb.at(x, o.Y(y)); c != nil && c.d == dMoon {
			o.fb.fg(x, o.Y(y), '·', light, dMoon)
		}
	}
	if (o.anim/4)%3 != 0 {
		o.fb.fg(int(cx-R*0.55)+2, o.Y(by-2), '•', o.c(mapmodel.InkAccent3, iEmit, 1), dMoon)
	}
}

// tetherX is the space elevator's screen column: at the Space Age
// district's east end, where the Fusion Age's tether went up. -1 before
// the Space Age.
func (o *orb) tetherX() int {
	ds := o.m.Skyline.Districts
	sa := o.m.Catalog.SpaceAge()
	if sa < 0 || sa >= len(ds) {
		return -1
	}
	d := ds[sa]
	return d.X0 + d.LandW - 4 - o.cam
}

// tether is the space elevator's cable, from its anchor on the limb up
// through the truss and on off the top of the view.
func (o *orb) tether(limb []float64) {
	x := o.tetherX()
	if x < 0 || x >= o.W {
		return
	}
	tc := o.c(mapmodel.InkFrameDim, iLit, 1)
	hi := o.c(mapmodel.InkFrame, iLit, 0)
	foot := min(o.S-1, int(limb[x]/2))
	for y := 0; y < foot; y++ {
		o.fb.fg(x, o.Y(y), '│', tc, dTether)
	}
	o.fb.fg(x, o.Y(foot), '▲', hi, dTether)
	o.fb.fg(x-1, o.Y(foot), '▗', tc, dTether)
	o.fb.fg(x+1, o.Y(foot), '▖', tc, dTether)
}

// truss is the station's spine along the baseline from the oldest district
// to the frontier: a white beam with a node every six columns, a steel
// lattice under it, nav lights, and an unfinished end past the frontier.
func (o *orb) truss() {
	gy := o.groundY
	beam := o.c(mapmodel.InkFrame, iLit, 0)
	steel := o.c(mapmodel.InkFrameDim, iLit, 0)
	dark := o.c(mapmodel.InkEcho, iLit, 0)
	green := o.c(mapmodel.InkAccent2, iEmit, 0)
	red := o.c(mapmodel.InkAccent3, iEmit, 0)
	end := o.x1 - 8
	for x := 0; x < o.W; x++ {
		wx := x + o.cam
		if wx < o.x0 || wx > o.x1 {
			continue
		}
		bg := o.fb.showAt(x, o.Y(gy))
		switch {
		case wx == o.x0:
			o.fb.set(x, o.Y(gy), '╞', beam, bg, dTruss)
			o.fb.fg(x, o.Y(gy+1), '▀', steel, dTruss)
			continue
		case wx > end:
			if wx%2 == 0 {
				o.fb.fg(x, o.Y(gy), '┄', steel, dTruss)
			}
			if wx%4 == 1 {
				o.fb.fg(x, o.Y(gy+1), '╱', dark, dTruss)
			}
			continue
		}
		ch := '═'
		if wx%6 == 0 {
			ch = '╪'
		}
		o.fb.set(x, o.Y(gy), ch, beam, steel, dTruss)
		lch := '╲'
		if wx%2 == 1 {
			lch = '╱'
		}
		if wx%6 == 0 {
			lch = '┴'
		}
		o.fb.fg(x, o.Y(gy+1), lch, steel, dTruss)
		if wx%12 == 3 && (o.anim/3+wx/12)%6 < 2 {
			c := green
			if (wx/12)%2 == 1 {
				c = red
			}
			o.fb.fg(x, o.Y(gy+1), '•', c, dTruss)
		}
	}
	if x := o.tetherX(); x >= 0 && x < o.W { // where the tether meets the truss
		hub := mapmodel.R(mapmodel.SymSkyHub, o.tier)
		o.fb.set(x, o.Y(gy), hub, beam, steel, dTruss)
	}
}
