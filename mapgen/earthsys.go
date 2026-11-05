package mapgen

// The close-up of the space ages: a calm Earth with its Moon, a few stations
// and a few ships passing, drawn in quadrant blocks (quad.go). Nothing of the
// town stands here; the town is at the space port, one zoom in. Everything
// about the picture (sizes, kinds, phases, the ships) comes from the seed.

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

// calm says whether this frame is the Earth close-up: a close look at the
// orbit or the solar system, with no district chosen.
func (f *frame) calm() bool {
	return f.close && f.district == "" && (f.st.Scale == Orbit || f.st.Scale == System)
}

// atPort says whether this frame is the space port: the deepest zoom of the
// orbit and the solar system, with a building line in focus.
func (f *frame) atPort() bool {
	return f.close && f.district != "" && (f.st.Scale == Orbit || f.st.Scale == System)
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

// earth paints the planet and its air, centred on (sx, sy) in 16ths of a
// pixel; the shading is lighting's.
func (p *Pixels) earth(seed int64, anim, sx, sy, R int, shade func(Paint, int, int) Paint) {
	own := p.Owner(Thing{Kind: KindTerrain, Key: "home"}, 5)
	p.Disc(sx, sy, R, own, func(nx, ny int) Paint {
		return shade(earthFace(seed, nx, ny, 6, anim/12), nx, ny)
	})
	p.Glow(sx, sy, R, R+3*pxUnit, HAir, 150, 3, own)
}

// moon paints the Moon, a satellite of the home world.
func (p *Pixels) moon(x, y, r int, shade func(Paint, int, int) Paint) {
	p.Disc(x, y, r, p.Owner(Thing{Kind: KindTerrain, Key: "satellite"}, 5), func(nx, ny int) Paint {
		return shade(P(HWhite, 235), nx, ny)
	})
}

// paintEarthSystem draws the close-up over the sky already in the frame.
func (f *frame) paintEarthSystem() {
	seed, anim := f.m.Seed, f.anim
	pix := NewPixels(f.g.W, f.g.H)
	c := f.plan.Centre
	sx, sy := (2*c.X+1)*pxUnit, (2*c.Y+1)*pxUnit
	R := (26 + roll(7, seed, 5200)) * pxUnit
	shade := lighting(seed)

	stationOwn := pix.Owner(Thing{Kind: KindTerrain, Key: "space_station"}, 5)

	// where the bodies on orbits are: the angle, and whether behind the planet
	at := func(a, rx, tilt int) (x, y int, behind bool) {
		return sx + rx*cos4096(a)/1024, sy + rx*tilt/100*sin4096(a)/2048, sin4096(a) < 0
	}
	moonR := (5 + roll(4, seed, 5202)) * pxUnit
	moonRx := R + (30+roll(20, seed, 5203))*pxUnit
	mx, my, moonBehind := at(roll(turn, seed, 5204)+anim/8*16, moonRx, 60+roll(40, seed, 5205))
	moon := func() { pix.moon(mx, my, moonR, shade) }
	n := 2 + roll(2, seed, 5206)
	drawStations := func(front bool) {
		for i := 0; i < n; i++ {
			k := stationKinds[(roll(len(stationKinds), seed, 5207)+i)%len(stationKinds)]
			rx := R + (k.orbit+roll(5, seed, 5208, int64(i)))*pxUnit
			x, y, behind := at(roll(turn, seed, 5209, int64(i))+anim/k.speed*16, rx, 40+roll(60, seed, 5210, int64(i)))
			if behind != front {
				pix.station(k, floorDiv(x, pxUnit), floorDiv(y, pxUnit), stationOwn)
			}
		}
	}

	if moonBehind {
		moon()
	}
	drawStations(false)
	pix.earth(seed, anim, sx, sy, R, shade)
	if !moonBehind {
		moon()
	}
	drawStations(true)

	bodies := []shipBody{{sx, sy, R}, {mx, my, moonR}}
	for _, side := range []int{-1, 1} { // the edges of the picture
		bodies = append(bodies, shipBody{sx + side*88*pxUnit, sy + (roll(21, seed, 5211, int64(side))-10)*pxUnit, 1})
	}
	pix.Ships(seed, anim, bodies, nil, nil)

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
