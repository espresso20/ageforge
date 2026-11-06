package mapgen

// The solar system drawn in quadrant blocks (see quad.go) from SolSpec: the
// sun with its glow, the planets as shaded discs, thin orbit lines, the
// speckled belts and, at the cosmos zoom, the comet. Sizes are in 16ths of a pixel, across the
// body, and grow with the zoom: at the cosmos zoom a planet is a speck, at the
// inner zoom Mars shows its cap. The inner zoom frames the sun, the four
// inner planets and the edge of the belt; the cosmos zoom shrinks the system
// to the middle of the window and rings it with the Kuiper belt.

// orreryLook is how big the bodies are at one zoom: the star and the planets
// as a percent of the spec's sizes, the dwarf planet outright.
type orreryLook struct{ sun, planet, dwarf int }

var orreryLooks = [...]orreryLook{
	SystemViewFull:   {100, 100, 18},
	SystemViewCosmos: {73, 73, 16},
	SystemViewInner:  {164, 178, 24},
}

// planet i of the home system, by place; the one the town stands on is 2
const (
	planetEarth = 2
	planetMars  = 3
)

// PlanetThing is the thing the orrery draws planet i of the home system as
// (0 Mercury to 7 Neptune): the home world is "home", the others "planet"
// numbered by Ref.
func PlanetThing(i int) Thing { return SolSpec.ThingOf(i) }

// PlanetOf says which planet of its system a thing is; ok is false for
// anything else.
func PlanetOf(th Thing) (i int, ok bool) {
	switch {
	case th == PlanetThing(planetEarth):
		return planetEarth, true
	case th.Kind == KindTerrain && th.Key == "planet" && th.Ref > 0:
		return th.Ref - 1, true
	}
	return 0, false
}

// orrery is one picture being drawn.
type orrery struct {
	s      *Scene
	pix    *Pixels
	spec   *SystemSpec
	seed   int64
	mode   SystemViewMode
	look   orreryLook
	sun    int   // the star's radius
	rad    []int // the planets' radii
	cx, cy int   // the star's cell
	sx, sy int   // the star's centre, in 16ths of a pixel
}

// paintOrrery draws the system (spec) into o over the star field that is
// already in it, reaching maxRY rows out.
func (s *Scene) paintOrrery(o *Overview, spec *SystemSpec, mode SystemViewMode, cx, cy, maxRY int) {
	var beltRX, beltRY int
	w, h := o.Grid.W, o.Grid.H
	r := &orrery{s: s, pix: NewPixels(w, h), spec: spec, seed: s.Seed, mode: mode, look: orreryLooks[mode], cx: cx, cy: cy}
	// a bigger window gets bigger bodies, a small one smaller
	k := clamp(cy*256/20, 205, 384)
	r.sun = spec.Star.Size * r.look.sun / 100 * k >> 8
	r.look.dwarf = r.look.dwarf * k >> 8
	planets := spec.Planets
	if mode == SystemViewInner { // the sun and the four inner planets
		planets = planets[:planetMars+1]
	}
	for _, p := range planets {
		r.rad = append(r.rad, p.Size*r.look.planet/100*k>>8)
	}
	r.sx, r.sy = (2*cx+1)*pxUnit, (2*cy+1)*pxUnit

	// how far out each orbit is: the inner zoom spreads them to fill the
	// window with the belt at its edge
	const spread = 170
	reach := func(orbit int) (rx, ry int) {
		if mode == SystemViewInner {
			orbit = orbit * spread / 100
		}
		return maxRY * orbit / 100 * 230 / 100, maxRY * orbit / 100
	}
	sunOwn := r.pix.Owner(sunThing, 4)
	orbitOwn := r.pix.Owner(Thing{Kind: KindRoad, Key: "orbit"}, 2)

	// the star's glow first, so everything else lies over it
	r.glow(r.sx, r.sy, r.sun, spec.Star)

	// the orbit lines, the home world's brighter
	rX, rY := make([]int, len(planets)), make([]int, len(planets))
	for i, p := range planets {
		rX[i], rY[i] = reach(p.Orbit)
		rY[i] = max(3+i, rY[i])
		rX[i] = max(rX[i], rY[i]*230/100)
		c := P(HOrbit, 175)
		if p.Home {
			c = P(HOrbit, 255)
		}
		r.pix.Ellipse(r.sx, r.sy, 32*rX[i], 32*rY[i], c, orbitOwn, 0, 0)
	}
	for _, b := range spec.Belts {
		rx, ry := reach(b)
		r.belt(rx, ry, orbitOwn)
		beltRX, beltRY = rx, ry
	}

	// where the planets are: the slower the further out
	x16, y16 := make([]int, len(planets)), make([]int, len(planets))
	cells := make([]Pt, len(planets))
	for i, p := range planets {
		phaseFrac := (int(s.Seed%1000) + i*382) % 1000
		ang := ((phaseFrac*4096)/1000 + (s.Anim/max(10, p.Orbit*p.Orbit/37))*16) % 4096
		x16[i] = r.sx + rX[i]*cos4096(ang)/32
		y16[i] = r.sy + rY[i]*sin4096(ang)/32
		cells[i] = Pt{floorDiv(x16[i], 2*pxUnit), floorDiv(y16[i], 2*pxUnit)}
	}
	r.body(sunOwn)
	for i := range planets {
		r.planet(i, x16[i], y16[i])
	}
	if mode == SystemViewCosmos { // the Kuiper belt, out past the dwarf planet
		r.belt(maxRY*138/100*230/100, maxRY*138/100, orbitOwn)
	}

	if spec.Dwarf && mode != SystemViewInner { // out beyond the rest
		plutoRY := (maxRY * 108) / 100
		plutoA := (1280 + s.Anim/360*16) % 4096
		px := r.sx + plutoRY*230/100*cos4096(plutoA)/32
		py := r.sy + plutoRY*sin4096(plutoA)/32
		plutoOwn := r.pix.Owner(Thing{Kind: KindTerrain, Key: "pluto"}, 5)
		r.pix.Disc(px, py, r.look.dwarf, plutoOwn, func(nx, ny int) Paint { return P(HAmber, 230) })
	}
	if spec.Comet && mode == SystemViewCosmos { // it crosses the cosmic view only
		r.comet()
	}
	// a hull pixel with a blinking light stands by the planets that have one
	stOwn := r.pix.Owner(Thing{Kind: KindTerrain, Key: "space_station"}, 6)
	lamp := 0
	for i, p := range planets {
		if p.Outpost == 1 {
			sx, sy := floorDiv(x16[i]-r.rad[i]-3*pxUnit, pxUnit), floorDiv(y16[i], pxUnit)
			r.pix.Set(sx, sy, P(HSteel, 255), stOwn)
			r.pix.light(r.seed, s.Anim, lamp, sx, sy-1, HRed, stOwn)
			lamp++
		}
	}
	r.pix.Resolve(o.Grid)
	r.pix.Release()

	// Glyph overlays that are not part of the picture: the stations at the
	// planets and the Dyson and mining works. They stand clear of the discs.
	if mode == SystemViewInner {
		sunR := (r.sun + 2*pxUnit - 1) / (2 * pxUnit)
		for _, dx := range []int{-(sunR + 2), sunR + 2} {
			hx, hy := cx+dx, cy
			if o.Grid.in(hx, hy) {
				o.Grid.paint(hx, hy, Cell{R: '⬢', Fg: P(HGold, 255), Bg: o.Grid.At(hx, hy).Bg})
				o.Grid.mark(hx, hy, Thing{Kind: KindBuilding, Key: "dyson_harvester"})
			}
		}
		for _, a := range []int{1024, 3072} {
			mx := cx + (beltRX*cos4096(a))/1024
			my := cy + (beltRY*sin4096(a))/1024
			if len(spec.Belts) > 0 && o.Grid.in(mx, my) {
				o.Grid.paint(mx, my, Cell{R: '⬡', Fg: P(HSteel, 240), Bg: o.Grid.At(mx, my).Bg})
				o.Grid.mark(mx, my, Thing{Kind: KindBuilding, Key: "asteroid_mine"})
			}
		}
	}
	for i, p := range planets {
		across := (r.rad[i] + 2*pxUnit - 1) / (2 * pxUnit) // cells from centre to rim
		if x := cells[i].X + across + 1; p.Outpost == 2 && o.Grid.in(x, cells[i].Y) {
			o.Grid.paint(x, cells[i].Y, Cell{R: '◈', Fg: P(HAmber, 240), Bg: o.Grid.At(x, cells[i].Y).Bg})
			o.Grid.mark(x, cells[i].Y, Thing{Kind: KindBuilding, Key: "space_station"})
		}
	}
}

// glow is the halo of the sun, round its middle (x, y) with radius R.
func (r *orrery) glow(x, y, R int, st StarSpec) {
	r.pix.Glow(x, y, R, R*19/10+2*pxUnit, st.Rim, 150, 3, r.pix.Owner(sunHaloThing, 1))
}

// disc draws star st, as thing own, at (x, y) with radius R: a heart, a body
// and a rim.
func (r *orrery) disc(own uint8, x, y, R int, st StarSpec) {
	r.pix.Disc(x, y, R, own, func(nx, ny int) Paint {
		switch d := isqrt(nx*nx + ny*ny); {
		case d < 100:
			return P(st.Core, 255)
		case d < 205:
			return P(st.Mid, 255)
		}
		return P(st.Rim, 255)
	})
}

// body is the system's own star.
func (r *orrery) body(own uint8) { r.disc(own, r.sx, r.sy, r.sun, r.spec.Star) }

// belt scatters the asteroids: a ring of specks that drift round slowly.
func (r *orrery) belt(rx, ry int, own uint8) {
	n := max(80, rx*3)
	drift := r.s.Anim / 48
	for i := 0; i < n; i++ {
		a := i*turn/n + drift
		j := roll(7, r.seed, int64(i)) - 3
		x := r.sx + (32*rx+j*24)*cos4096(a)/1024
		y := r.sy + (32*ry+j*12)*sin4096(a)/1024
		c := P(HRock, 120+roll(136, r.seed, int64(i), 1))
		if roll(9, r.seed, int64(i), 2) == 0 {
			c = P(HSand, 210)
		}
		r.pix.Set(floorDiv(x, pxUnit), floorDiv(y, pxUnit), c, own)
	}
}

// planet draws planet i (and what travels with it) centred at (x, y).
func (r *orrery) planet(i, x, y int) {
	p := &r.spec.Planets[i]
	own := r.pix.Owner(r.spec.ThingOf(i), 5)
	R := r.rad[i]
	// the light comes from the star
	lx, ly := r.sx-x, 2*(r.sy-y)
	l := max(1, isqrt(lx*lx+ly*ly))
	lx, ly = lx*256/l, ly*256/l

	if p.Ring { // the back of the ring, behind the planet
		r.pix.satRing(x, y, R, own, -1)
	}
	r.pix.Disc(x, y, R, own, func(nx, ny int) Paint {
		c := p.surface(r.seed, i, nx, ny, 4, 0)
		z := isqrt(max(0, 65536-nx*nx-ny*ny))
		dot := (nx*lx+ny*ly)/256 + z*90/256
		lvl := 256
		switch {
		case dot < -30:
			lvl = 120
		case dot < 90:
			lvl = 180
		}
		return Paint{c.Hue, uint8(int(c.Weight) * lvl >> 8)}
	})
	if p.Ring { // and the front of it, across the planet
		r.pix.satRing(x, y, R, own, 1)
	}
	if r.mode == SystemViewCosmos {
		return
	}
	// the moons: a dot each on a slow orbit, wherever the disc is big enough,
	// the first eight of them
	moon := r.pix.Owner(Thing{Kind: KindTerrain, Key: "satellite"}, 6)
	for k := 0; k < min(p.Moons, 8) && R >= 3*pxUnit/2; k++ {
		a := (r.s.Anim/(3+2*k)*16 + roll(turn, r.seed, 92, int64(i), int64(k))) % 4096
		rm := R + (3+2*k)*pxUnit
		r.pix.Set(floorDiv(x+rm*cos4096(a)/1024, pxUnit), floorDiv(y+rm*sin4096(a)/2048, pxUnit), moonTone[k%len(moonTone)], moon)
	}
}

var moonTone = [...]Paint{P(HWhite, 235), P(HSun, 255), P(HRock, 255), P(HSand, 240), P(HCool, 240)}

// satRing draws a planet's ring: two thin lines round it, the half behind it
// (side -1) or the half in front (side 1).
func (p *Pixels) satRing(x, y, R int, own uint8, side int) {
	p.Ellipse(x, y, R*19/10, R*7/10, P(HSand, 215), own, 0, side)
	p.Ellipse(x, y, R*24/10, R*9/10, P(HGold, 170), own, 0, side)
}

// surface is what the planet (the ith of its system) is made of at (nx, ny)
// on its face, before the light falls on it; fine is how many blotches the
// seas and the flecks have across, and drift moves the clouds.
func (p *PlanetSpec) surface(seed int64, i, nx, ny, fine, drift int) Paint {
	switch {
	case p.World:
		return earthFace(seed, nx, ny, fine, drift)
	case p.Cap && ny < -205:
		return P(HWhite, 245)
	case p.Spot && nx > 20 && nx < 150 && ny > 50 && ny < 120:
		return P(HRed, 255) // the great storm
	case len(p.Bands) > 0:
		return p.Bands[clamp((ny+256)*len(p.Bands)/512, 0, len(p.Bands)-1)]
	case p.Blotch.Weight > 0 && roll(5, seed, 91, int64(i), int64((nx+256)*(fine-1)/512), int64((ny+256)*(fine-1)/512)) == 0:
		return p.Blotch
	}
	return p.Base
}
