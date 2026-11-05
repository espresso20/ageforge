package mapgen

// The Sol system drawn in quadrant blocks (see quad.go): the sun with its
// glow, the planets as shaded discs, thin orbit lines, the speckled belt and
// the comet. Sizes are in 16ths of a pixel, across the body, and grow with
// the zoom: at the cosmos zoom a planet is a speck, at the inner zoom Jupiter
// shows its bands.

// orreryLook is how big the bodies are at one zoom, in pixels across the
// radius.
type orreryLook struct {
	sun    int
	planet [8]int
	pluto  int
	comet  int
}

var orreryLooks = map[SystemViewMode]orreryLook{
	SystemViewCosmos: {sun: 4 * 16, planet: [8]int{16, 24, 26, 20, 48, 42, 30, 30}, pluto: 16, comet: 20},
	SystemViewFull:   {sun: 88, planet: [8]int{22, 32, 35, 27, 68, 58, 42, 42}, pluto: 18, comet: 24},
	SystemViewInner:  {sun: 9 * 16, planet: [8]int{38, 58, 61, 48, 128, 104, 72, 72}, pluto: 24, comet: 28},
}

// planet owners' keys: every planet but the home world is "planet".
const (
	planetMercury = iota
	planetVenus
	planetEarth
	planetMars
	planetJupiter
	planetSaturn
	planetUranus
	planetNeptune
)

// PlanetThing is the thing the orrery draws planet i as (0 Mercury to 7
// Neptune): the home world is "home", the others "planet" numbered by Ref.
func PlanetThing(i int) Thing {
	if i == planetEarth {
		return Thing{Kind: KindTerrain, Key: "home"}
	}
	return Thing{Kind: KindTerrain, Key: "planet", Ref: i + 1}
}

// PlanetOf says which planet a thing is; ok is false for anything else.
func PlanetOf(th Thing) (i int, ok bool) {
	for i := 0; i < len(moonCount); i++ {
		if PlanetThing(i) == th {
			return i, true
		}
	}
	return 0, false
}

// orrery is one picture being drawn.
type orrery struct {
	s      *Scene
	pix    *Pixels
	mode   SystemViewMode
	look   orreryLook
	cx, cy int // the sun's cell
	sx, sy int // the sun's centre, in 16ths of a pixel
}

// paintOrrery draws the system into o over the star field that is already in
// it.
func (s *Scene) paintOrrery(o *Overview, mode SystemViewMode, cx, cy, maxRY int, rX, rY [8]int, beltRX, beltRY int) {
	var planets [8]Pt
	w, h := o.Grid.W, o.Grid.H
	r := &orrery{s: s, pix: NewPixels(w, h), mode: mode, look: orreryLooks[mode], cx: cx, cy: cy}
	// a bigger window gets bigger bodies, a small one smaller
	k := clamp(cy*256/20, 205, 384)
	for i := range r.look.planet {
		r.look.planet[i] = r.look.planet[i] * k >> 8
	}
	r.look.sun, r.look.pluto, r.look.comet = r.look.sun*k>>8, r.look.pluto*k>>8, r.look.comet*k>>8
	r.sx, r.sy = (2*cx+1)*pxUnit, (2*cy+1)*pxUnit

	sunOwn := r.pix.Owner(Thing{Kind: KindTerrain, Key: "sun"}, 4)
	glowOwn := r.pix.Owner(Thing{Kind: KindTerrain, Key: "sun", Ref: 1}, 1) // Ref 1: the halo, not the sun itself
	orbitOwn := r.pix.Owner(Thing{Kind: KindRoad, Key: "orbit"}, 2)

	// the sun's glow first, so everything else lies over it
	r.pix.Glow(r.sx, r.sy, r.look.sun, r.look.sun*19/10+2*pxUnit, HCore, 150, 3, glowOwn)

	// the orbit lines, the home world's brighter
	for i := 0; i < 8; i++ {
		c := P(HOrbit, 175)
		if i == planetEarth {
			c = P(HOrbit, 255)
		}
		r.pix.Ellipse(r.sx, r.sy, 32*rX[i], 32*rY[i], c, orbitOwn, 0, 0)
	}
	r.belt(beltRX, beltRY, orbitOwn)

	// where the planets are
	var x16, y16 [8]int
	divs := [8]int{12, 20, 32, 50, 80, 120, 180, 260}
	for i := 0; i < 8; i++ {
		phaseFrac := (int(s.Seed%1000) + i*382) % 1000
		ang := ((phaseFrac*4096)/1000 + (s.Anim/divs[i])*16) % 4096
		x16[i] = r.sx + rX[i]*cos4096(ang)/32
		y16[i] = r.sy + rY[i]*sin4096(ang)/32
		planets[i] = Pt{floorDiv(x16[i], 2*pxUnit), floorDiv(y16[i], 2*pxUnit)}
	}
	r.sun(sunOwn)
	for i := 0; i < 8; i++ {
		r.planet(i, x16[i], y16[i])
	}

	// Pluto, out beyond the rest
	plutoRY := (maxRY * 108) / 100
	plutoA := (1280 + s.Anim/360*16) % 4096
	px := r.sx + plutoRY*230/100*cos4096(plutoA)/32
	py := r.sy + plutoRY*sin4096(plutoA)/32
	plutoOwn := r.pix.Owner(Thing{Kind: KindTerrain, Key: "pluto"}, 5)
	r.pix.Disc(px, py, r.look.pluto, plutoOwn, func(nx, ny int) Paint { return P(HAmber, 230) })

	r.comet(maxRY)
	// the station by the home world: a fixed hull pixel with a blinking light
	stOwn := r.pix.Owner(Thing{Kind: KindTerrain, Key: "space_station"}, 6)
	sx, sy := floorDiv(x16[planetEarth]-r.look.planet[planetEarth]-3*pxUnit, pxUnit), floorDiv(y16[planetEarth], pxUnit)
	r.pix.Set(sx, sy, P(HSteel, 255), stOwn)
	r.pix.light(s.Seed, s.Anim, 0, sx, sy-1, HRed, stOwn)
	r.pix.Resolve(o.Grid)
	r.pix.Release()

	// Glyph overlays that are not part of the picture: the stations at the
	// planets and the Dyson and mining works. They stand clear of the discs.
	sunR := (r.look.sun + 2*pxUnit - 1) / (2 * pxUnit)
	if mode == SystemViewInner {
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
			if o.Grid.in(mx, my) {
				o.Grid.paint(mx, my, Cell{R: '⬡', Fg: P(HSteel, 240), Bg: o.Grid.At(mx, my).Bg})
				o.Grid.mark(mx, my, Thing{Kind: KindBuilding, Key: "asteroid_mine"})
			}
		}
	}
	across := func(i int) int { return (r.look.planet[i] + 2*pxUnit - 1) / (2 * pxUnit) } // cells from centre to rim
	if p := planets[planetMars]; o.Grid.in(p.X+across(planetMars)+1, p.Y) {
		x := p.X + across(planetMars) + 1
		o.Grid.paint(x, p.Y, Cell{R: '◈', Fg: P(HAmber, 240), Bg: o.Grid.At(x, p.Y).Bg})
		o.Grid.mark(x, p.Y, Thing{Kind: KindBuilding, Key: "space_station"})
	}
}

// sun is the star: a white heart, a yellow body and an orange rim.
func (r *orrery) sun(own uint8) {
	r.pix.Disc(r.sx, r.sy, r.look.sun, own, func(nx, ny int) Paint {
		switch d := isqrt(nx*nx + ny*ny); {
		case d < 100:
			return P(HWhite, 255)
		case d < 205:
			return P(HSun, 255)
		}
		return P(HCore, 255)
	})
}

// belt scatters the asteroids: a ring of specks that drift round slowly.
func (r *orrery) belt(rx, ry int, own uint8) {
	n := max(80, rx*3)
	drift := r.s.Anim / 48
	for i := 0; i < n; i++ {
		a := i*turn/n + drift
		j := roll(7, r.s.Seed, int64(i)) - 3
		x := r.sx + (32*rx+j*24)*cos4096(a)/1024
		y := r.sy + (32*ry+j*12)*sin4096(a)/1024
		c := P(HRock, 120+roll(136, r.s.Seed, int64(i), 1))
		if roll(9, r.s.Seed, int64(i), 2) == 0 {
			c = P(HSand, 210)
		}
		r.pix.Set(floorDiv(x, pxUnit), floorDiv(y, pxUnit), c, own)
	}
}

// comet sweeps round outside the planets, its tail pointing away from the sun.
func (r *orrery) comet(maxRY int) {
	a := (3400 + r.s.Anim/200*16) % 4096
	ry := (maxRY * 118) / 100
	hx := r.sx + ry*230/100*cos4096(a)/32
	hy := r.sy + ry*sin4096(a)/32
	own := r.pix.Owner(Thing{Kind: KindTerrain, Key: "comet"}, 5)
	dx, dy := (hx-r.sx)/pxUnit, (hy-r.sy)/pxUnit
	m := max(1, abs(dx), abs(dy))
	tail := r.look.comet * 6 / pxUnit
	for t := tail; t >= 1; t-- {
		x := floorDiv(hx+dx*t*pxUnit/m, pxUnit)
		y := floorDiv(hy+dy*t*pxUnit/m, pxUnit)
		r.pix.Set(x, y, P(HCyan, 215-t*200/(tail+1)), own)
	}
	r.pix.Disc(hx, hy, r.look.comet, own, func(nx, ny int) Paint { return P(HWhite, 255) })
}

// planet draws one planet (and what travels with it) centred at (x, y).
func (r *orrery) planet(i, x, y int) {
	own := r.pix.Owner(PlanetThing(i), 5)
	R := r.look.planet[i]
	// the light comes from the sun
	lx, ly := r.sx-x, 2*(r.sy-y)
	l := max(1, isqrt(lx*lx+ly*ly))
	lx, ly = lx*256/l, ly*256/l

	if i == planetSaturn { // the back of the ring, behind the planet
		r.pix.satRing(x, y, R, own, -1)
	}
	r.pix.Disc(x, y, R, own, func(nx, ny int) Paint {
		c := surface(r.s.Seed, i, nx, ny, 4, 0)
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
	if i == planetSaturn { // and the front of it, across the planet
		r.pix.satRing(x, y, R, own, 1)
	}
	if r.mode == SystemViewCosmos {
		return
	}
	// the moons: a dot each on a slow orbit, wherever the disc is big enough
	moon := r.pix.Owner(Thing{Kind: KindTerrain, Key: "satellite"}, 6)
	for k := 0; k < moonCount[i] && R >= 3*pxUnit/2; k++ {
		a := (r.s.Anim/(3+2*k)*16 + roll(turn, r.s.Seed, 92, int64(i), int64(k))) % 4096
		rm := R + (3+2*k)*pxUnit
		r.pix.Set(floorDiv(x+rm*cos4096(a)/1024, pxUnit), floorDiv(y+rm*sin4096(a)/2048, pxUnit), moonTone[k%len(moonTone)], moon)
	}
}

// moonCount is how many moons each planet shows, as the sky has them.
var moonCount = [8]int{0, 0, 1, 2, 4, 5, 4, 2}

var moonTone = [...]Paint{P(HWhite, 235), P(HSun, 255), P(HRock, 255), P(HSand, 240), P(HCool, 240)}

// satRing draws Saturn's ring: two thin lines round the planet, the half
// behind it (side -1) or the half in front (side 1).
func (p *Pixels) satRing(x, y, R int, own uint8, side int) {
	p.Ellipse(x, y, R*19/10, R*7/10, P(HSand, 215), own, 0, side)
	p.Ellipse(x, y, R*24/10, R*9/10, P(HGold, 170), own, 0, side)
}

// surface is what planet i is made of at (nx, ny) on its face, before the
// light falls on it; fine is how many blotches the home world and Mars have
// across, and drift moves the home world's clouds.
func surface(seed int64, i, nx, ny, fine, drift int) Paint {
	band := (ny + 256) * 9 / 512 // nine bands, pole to pole
	switch i {
	case planetMercury:
		return P(HRock, 255)
	case planetVenus:
		if (ny+256)/90%2 == 1 {
			return P(HSun, 205)
		}
		return P(HSun, 250)
	case planetEarth:
		return earthFace(seed, nx, ny, fine, drift)
	case planetMars:
		if ny < -205 {
			return P(HWhite, 245)
		}
		if roll(5, seed, 91, int64((nx+256)*(fine-1)/512), int64((ny+256)*(fine-1)/512)) == 0 {
			return P(HRed, 200)
		}
		return P(HRed, 255)
	case planetJupiter:
		if nx > 20 && nx < 150 && ny > 50 && ny < 120 {
			return P(HRed, 255) // the great red spot
		}
		return [9]Paint{P(HRock, 235), P(HAmber, 255), P(HSand, 255), P(HRock, 215), P(HAmber, 245),
			P(HSand, 240), P(HRock, 225), P(HAmber, 255), P(HRock, 235)}[clamp(band, 0, 8)]
	case planetSaturn:
		return [9]Paint{P(HSand, 235), P(HGold, 250), P(HSand, 255), P(HGold, 235), P(HSand, 250),
			P(HGold, 255), P(HSand, 240), P(HGold, 240), P(HSand, 230)}[clamp(band, 0, 8)]
	case planetUranus:
		return P(HCool, 255)
	}
	if band%4 == 1 {
		return P(HCyan, 215)
	}
	return P(HPanel, 255)
}
