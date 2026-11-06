package mapgen

// The close-up of the space ages: a calm Earth with its Moon, a few stations
// with blinking lights, drawn in quadrant blocks (quad.go). Nothing of the
// town stands here; the town is at the space port, one zoom in. Everything
// about the picture (sizes, kinds, phases, the lights) comes from the seed.
// Any other planet of the system gets the same close-up: its own face and its
// own moons (its spec's count), and no stations.

// station is one kind of station: a disc, a truss with panels, or a cross.
type station struct {
	shape int // 0 disc, 1 truss, 2 cross
	size  int // pixels across half of it
	hull  Paint
	trim  Hue
	orbit int // pixels above the planet's rim
	speed int // frames per step of its orbit
}

var stationKinds = [...]station{
	{0, 3, P(HSteel, 255), HCyan, 7, 7},
	{1, 5, P(HWhite, 255), HGlass, 10, 9},
	{2, 4, P(HMetal, 255), HAmber, 13, 6},
	{1, 4, P(HSteel, 255), HSun, 17, 11},
}

// calm says whether this frame is a planet close-up: in the orbit, the home
// world with no district chosen; in the solar system, any planet but the home
// world.
func (f *frame) calm() bool {
	return f.close && (f.st.Scale == Orbit && f.district == "" || f.st.Scale == System && f.planet > 0)
}

// atPort says whether this frame is the space port: in the orbit, the deepest
// zoom with a building line in focus; in the solar system, the close-up of
// the home world.
func (f *frame) atPort() bool {
	return f.close && (f.st.Scale == Orbit && f.district != "" || f.st.Scale == System && f.planet == 0)
}

// earthFace is the planet's face at (nx, ny), -256 to 256 across: sea, land
// and cloud on an n by n grid of blotches, the cloud drifting by drift.
func earthFace(seed int64, nx, ny, n, drift int) Paint {
	if abs(ny) > 215 {
		return P(HWhite, 245)
	}
	cell := func(salt int64, d int) int {
		return roll(10, seed, 77+salt, int64((nx+256+d)*n/512), int64((ny+256)*n/512))
	}
	switch {
	case cell(0, 0) < 3:
		return P(HHydro, 235)
	case cell(1, drift) == 0:
		return P(HWhite, 225)
	}
	return P(HPanel, 255)
}

// lighting shades a paint for the seed's light: night, dusk or day, by where
// the point (nx, ny) on a body's face lies.
func lighting(seed int64) func(c Paint, nx, ny int) Paint {
	la := roll(turn, seed, 5201) // the light comes from this way
	lx, ly := cos4096(la)*256/1024, 2*sin4096(la)*256/1024
	return func(c Paint, nx, ny int) Paint {
		z := isqrt(max(0, 65536-nx*nx-ny*ny))
		lvl := 2
		switch dot := (nx*lx+ny*ly)/256 + z*90/256; {
		case dot < -30:
			lvl = 0
		case dot < 90:
			lvl = 1
		}
		return Paint{c.Hue, uint8(int(c.Weight) * [3]int{120, 180, 256}[lvl] >> 8)}
	}
}

// planet paints planet i of a system as the thing th and, for a world with
// seas, its air, centred on (sx, sy) in 16ths of a pixel; the shading is
// lighting's.
func (p *Pixels) planet(seed int64, anim int, pl *PlanetSpec, i int, th Thing, sx, sy, R int, shade func(Paint, int, int) Paint) {
	own := p.Owner(th, 5)
	if pl.Ring {
		p.satRing(sx, sy, R, own, -1)
	}
	p.Disc(sx, sy, R, own, func(nx, ny int) Paint {
		return shade(pl.surface(seed, i, nx, ny, 6, anim/12), nx, ny)
	})
	if pl.Ring {
		p.satRing(sx, sy, R, own, 1)
	}
	if pl.World {
		p.Glow(sx, sy, R, R+3*pxUnit, HAir, 150, 3, own)
	}
}

// moon paints a moon, a satellite of a planet, in the table's tone k.
func (p *Pixels) moon(x, y, r, k int, shade func(Paint, int, int) Paint) {
	p.Disc(x, y, r, p.Owner(Thing{Kind: KindTerrain, Key: "satellite"}, 5), func(nx, ny int) Paint {
		return shade(moonTone[k%len(moonTone)], nx, ny)
	})
}

// paintEarthSystem draws the close-up over the sky already in the frame: the
// home world with its stations, or the planet the player zoomed on.
func (f *frame) paintEarthSystem() {
	anim := f.anim
	sp, seed := &SolSpec, f.m.Seed
	pl := planetEarth
	if f.planet > 0 {
		pl = f.planet - 1
	}
	ps := &sp.Planets[pl]
	pix := NewPixels(f.g.W, f.g.H)
	c := f.plan.Centre
	sx, sy := (2*c.X+1)*pxUnit, (2*c.Y+1)*pxUnit
	R := (26 + roll(7, seed, 5200)) * pxUnit * ps.Close / 100
	shade := lighting(seed)

	stationOwn := pix.Owner(Thing{Kind: KindTerrain, Key: "space_station"}, 5)

	// where the bodies on orbits are: the angle, and whether behind the planet
	at := func(a, rx, tilt int) (x, y int, behind bool) {
		return sx + rx*cos4096(a)/1024, sy + rx*tilt/100*sin4096(a)/2048, sin4096(a) < 0
	}
	// the planet's moons, from its spec; moon 0 of the home world
	// is the Moon. Moon k keeps its first roll when k is 0.
	rk := func(n int, salt int64, k int) int {
		if k == 0 {
			return roll(n, seed, salt)
		}
		return roll(n, seed, salt, int64(k))
	}
	moons := func(front bool) {
		for k := 0; k < min(ps.Moons, 14); k++ {
			r := max(pxUnit, (5+rk(4, 5202, k)-k)*pxUnit)
			rx := R + (30+rk(20, 5203, k)+10*min(k, 5)+3*max(0, k-5))*pxUnit
			x, y, behind := at(rk(turn, 5204, k)+anim/(8+2*k)*16, rx, 60+rk(40, 5205, k))
			if behind != front {
				pix.moon(x, y, r, k, shade)
			}
		}
	}
	n := ps.Stations
	if ps.Home {
		n += roll(2, seed, 5206)
	}
	drawStations := func(front bool) {
		for i := 0; i < n; i++ {
			k := stationKinds[(roll(len(stationKinds), seed, 5207)+i)%len(stationKinds)]
			rx := R + (k.orbit+roll(5, seed, 5208, int64(i)))*pxUnit
			x, y, behind := at(roll(turn, seed, 5209, int64(i))+anim/k.speed*16, rx, 40+roll(60, seed, 5210, int64(i)))
			if behind != front {
				x, y = floorDiv(x, pxUnit), floorDiv(y, pxUnit)
				pix.station(k, x, y, stationOwn)
				// nav lights at the tips, and a white strobe on the largest
				tip := [3]int{k.size - 1, k.size, k.size / 2 * 2}[k.shape]
				pix.light(seed, anim, 4*i, x-tip, y, HRed, stationOwn)
				pix.light(seed, anim, 4*i+1, x+tip, y, HLife, stationOwn)
				if k.size == 5 {
					pix.light(seed, anim, 4*i+2, x, y-2, HWhite, stationOwn)
				}
			}
		}
	}

	moons(false)
	drawStations(false)
	pix.planet(seed, anim, ps, pl, sp.ThingOf(pl), sx, sy, R, shade)
	moons(true)
	drawStations(true)

	pix.Resolve(f.g)
	pix.Release()
}

// station draws a station centred on pixel (x, y).
func (p *Pixels) station(k station, x, y int, own uint8) {
	switch k.shape {
	case 0:
		p.Disc(x*pxUnit+pxUnit/2, y*pxUnit+pxUnit/2, k.size*pxUnit, own, func(nx, ny int) Paint {
			if abs(nx) < 90 && abs(ny) < 90 {
				return P(k.trim, 255)
			}
			return k.hull
		})
	case 1:
		for dx := -k.size; dx <= k.size; dx++ {
			p.Set(x+dx, y, k.hull, own)
		}
		for _, e := range [2]int{-k.size, k.size} {
			p.Set(x+e, y-1, P(k.trim, 255), own)
			p.Set(x+e, y+1, P(k.trim, 255), own)
		}
		p.Set(x, y, P(HWhite, 255), own)
	default:
		for d := -k.size / 2; d <= k.size/2; d++ {
			p.Set(x+2*d, y, k.hull, own)
			p.Set(x, y+d, k.hull, own)
		}
		p.Set(x, y, P(k.trim, 255), own)
	}
}
