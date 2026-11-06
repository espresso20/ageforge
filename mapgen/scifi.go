package mapgen

// What makes the Interstellar close-up read as an age of starflight: city
// lights on Earth's night side, space elevators to the ring, big stations
// standing round the world, a starship docked at the ring and a few dim
// satellites. Everything is still or very slow; the lights are the slow
// beacons of lights.go. All of it is drawn in pixels, from the seed.

// line paints a straight run of pixels from (x0, y0) to (x1, y1).
func (p *Pixels) line(x0, y0, x1, y1 int, c Paint, own uint8) {
	n := max(abs(x1-x0), abs(y1-y0))
	for i := 0; i <= n; i++ {
		if n == 0 {
			p.Set(x0, y0, c, own)
			return
		}
		p.Set(x0+(x1-x0)*i/n, y0+(y1-y0)*i/n, c, own)
	}
}

// setFree paints a pixel only where nothing is yet, so a dim extra never hides
// or disowns a module.
func (p *Pixels) setFree(x, y int, c Paint, own uint8) {
	if x >= 0 && y >= 0 && x < p.W && y < p.H && p.px[y*p.W+x].Hue == HNone {
		p.Set(x, y, c, own)
	}
}

// cityLights scatters warm pixels over the land on Earth's dark side.
func (f *frame) cityLights(pix *Pixels, pp *port, shade func(Paint, int, int) Paint) {
	seed := f.m.Seed
	own := pix.Owner(SolSpec.ThingOf(planetEarth), 5)
	pl := &SolSpec.Planets[planetEarth]
	for y := -pp.R; y <= pp.R; y++ {
		for x := -pp.R; x <= pp.R; x++ {
			nx, ny := x*256/pp.R, y*256/pp.R
			if nx*nx+ny*ny > 240*240 || shade(P(HPanel, 255), nx, ny).Weight > 130 || pl.surface(seed, planetEarth, nx, ny, 6, 0).Hue != HPanel {
				continue
			}
			px, py := pp.cx+x, pp.cy+y
			if roll(3, seed, 5420, int64(px/3), int64(py/3)) != 0 || roll(2, seed, 5421, int64(px), int64(py)) != 0 {
				continue
			}
			c := P(HAmber, 255)
			if roll(3, seed, 5422, int64(px), int64(py)) == 0 {
				c = P(HSun, 255)
			}
			pix.Set(px, py, c, own)
		}
	}
}

// shield is a faint shimmer on a circle just outside the air: a few dim
// pixels that slowly brighten and fade.
func (f *frame) shield(pix *Pixels, pp *port) {
	seed, own := f.m.Seed, pix.Owner(SolSpec.ThingOf(planetEarth), 5)
	for i := 0; i < 28; i++ {
		a := roll(turn, seed, 5450, int64(i))
		per := 48 + roll(32, seed, 5451, int64(i))
		ph := (f.anim + roll(per, seed, 5452, int64(i))) % per
		w := 10 + 70*min(ph, per-ph)*2/per
		r := pp.R + 5
		pix.setFree(pp.cx+r*cos4096(a)/1024, pp.cy+r*sin4096(a)/1024, P(HAir, w), own)
	}
}

// tether is one space elevator: its foot on Earth, its head on the ring.
type tether struct{ x0, y0, x1, y1 int }

// tethers lays two or three elevators from the planet's surface to the near
// side of the ring.
func (pp *port) tethers(seed int64) []tether {
	n, a0 := 2+roll(2, seed, 5430), roll(turn, seed, 5431)
	var out []tether
	last := -9
	for i := 0; i < 24 && len(out) < n; i++ {
		x, y, front := pp.at(a0+i*turn/24, 100)
		dx, dy := x-pp.cx, y-pp.cy
		d := isqrt(dx*dx + dy*dy)
		if !front || d < pp.R+8 || i < last+3 {
			continue
		}
		last = i
		out = append(out, tether{pp.cx + dx*pp.R*9/10/d, pp.cy + dy*pp.R*9/10/d, x, y})
	}
	return out
}

// paintTethers draws the elevators, each with a climber that creeps up it a
// pixel every few frames.
func (f *frame) paintTethers(pix *Pixels, pp *port) {
	own := pix.Owner(Thing{Kind: KindRoad, Key: "tether"}, 5)
	for i, t := range pp.tethers(f.m.Seed) {
		pix.line(t.x0, t.y0, t.x1, t.y1, P(HSteel, 200), own)
		n := max(abs(t.x1-t.x0), abs(t.y1-t.y0))
		k := (f.anim/6 + roll(n+1, f.m.Seed, 5432, int64(i))) % (n + 1)
		pix.Set(t.x0+(t.x1-t.x0)*k/n, t.y0+(t.y1-t.y0)*k/n, P(HWhite, 255), own)
	}
}

// satellites the swarm round the ring has.
const swarmSize = 10

// sat is satellite i of the swarm at frame anim: its pixel and how bright it
// is (0 while dark). It drifts along the ring a pixel every few dozen frames
// and is lit for half of a 5 to 8 second cycle, fading in and out.
func (pp *port) sat(seed int64, anim, i int) (x, y, w int) {
	x, y, _ = pp.at(roll(turn, seed, 5310, int64(i))+anim/6, 112+roll(45, seed, 5311, int64(i)))
	per := 40 + roll(25, seed, 5312, int64(i))
	ph, half := (anim+roll(per, seed, 5313, int64(i)))%per, per/2
	if ph >= half {
		return x, y, 0
	}
	return x, y, 6 + 110*sin4096(ph*turn/2/half)/1024
}

func (f *frame) swarm(pix *Pixels, pp *port) {
	for i := 0; i < swarmSize; i++ {
		if x, y, w := pp.sat(f.m.Seed, f.anim, i); w > 0 {
			pix.setFree(x, y, P(HWhite, w), 0)
		}
	}
}

// starship is the one big interstellar ship, docked at the ring and never
// moving: a blocky hull nose-in, a wider drive section and a glow behind it
// that pulses gently.
func (f *frame) starship(pix *Pixels, pp *port) {
	seed := f.m.Seed
	own := pix.Owner(Thing{Kind: KindTerrain, Key: "space_station"}, 6)
	// the berth: the dock farthest to the side with room for the hull beside it
	// that covers no other module; else a spot at the ring's edge
	bx, by, hw, found := 0, 0, 3, false
	free := func(x, y, d int) bool {
		for _, o := range pp.mods {
			if (o.x-x)*d > 0 && (o.x-x)*d < 28 && abs(o.y-y) < 8 {
				return false
			}
		}
		return true
	}
	for _, m := range pp.mods {
		d := 1
		if m.x < pp.cx {
			d = -1
		}
		if m.m.shape == shDock && free(m.x, m.y, d) && (!found || abs(m.x-pp.cx) > abs(bx-pp.cx)) {
			bx, by, hw, found = m.x, m.y, (m.m.w+(m.tier+1)/2)/2+1, true
		}
	}
	if !found {
		bx, by, _ = pp.at(roll(2, seed, 5460)*turn/2+roll(200, seed, 5461)-100, 100)
	}
	dir := 1
	if bx < pp.cx {
		dir = -1
	}
	bx += dir * hw
	rows := [...]int{1, 2, 3, 3, 4, 4, 4, 4, 4, 4, 4, 4, 6, 6, 6, 6}
	pix.line(bx, by, bx+2*dir, by, P(HSteel, 255), own)
	for i, h := range rows {
		x := bx + (3+i)*dir
		for r := 0; r < h; r++ {
			c := P(HWhite, 235)
			switch {
			case i == 0:
				c = P(HCyan, 255)
			case i >= 12:
				c = P(HSteel, 255)
				if r == 0 {
					c = P(HGold, 255)
				}
			case i == 4 || i == 8:
				c = P(HGold, 255)
			case r == h-1 && h > 2:
				c = P(HSteel, 255)
			case r == h/2 && i%4 != 0 && i > 4:
				c = P(HCyan, 255)
			}
			pix.Set(x, by-h/2+r, c, own)
		}
	}
	// the drive glow: the same pixels all the time, only the brightness breathes
	pulse := 190 + 60*sin4096(f.anim*48)/1024
	for i := 0; i < 5; i++ {
		for r := 0; r < 4-i/2; r++ {
			pix.Set(bx+(3+len(rows)+i)*dir, by-(4-i/2)/2+r, P(HPlasma, pulse*(5-i)/5), own)
		}
	}
}

// megaStations stands two or three huge stations round the world, each a wheel,
// a spindle with docking arms or a cluster of habitat spheres, bigger than the
// Moon, with slow beacons on their tips.
func (f *frame) megaStations(pix *Pixels, pp *port, shade func(Paint, int, int) Paint) {
	seed, anim := f.m.Seed, f.anim
	own := pix.Owner(Thing{Kind: KindTerrain, Key: "space_station"}, 6)
	n, kind, a0 := 2+roll(2, seed, 5440), roll(3, seed, 5441), roll(turn, seed, 5442)
	for i := 0; i < n; i++ {
		x, y, _ := pp.at(a0+i*turn/n+roll(turn/8, seed, 5443, int64(i)), 130+roll(25, seed, 5444, int64(i)))
		x, y = clamp(x, 18, max(18, pix.W-19)), clamp(y, 14, max(14, pix.H-15))
		lamp := func(j, dx, dy int, h Hue) { pix.light(seed, anim, 300+4*i+j, x+dx, y+dy, h, own) }
		switch (kind + i) % 3 {
		case 0: // a wheel
			rx := 11 + roll(3, seed, 5445, int64(i))
			ry := rx * (55 + roll(15, seed, 5446, int64(i))) / 100
			pix.Ellipse(x*pxUnit, y*pxUnit, rx*pxUnit, ry*pxUnit, P(HWhite, 240), own, 0, 0)
			pix.Ellipse(x*pxUnit, y*pxUnit, (rx-2)*pxUnit, (ry-2)*pxUnit, P(HSteel, 255), own, 0, 0)
			pix.line(x-rx+2, y, x+rx-2, y, P(HSteel, 255), own)
			pix.line(x, y-ry+2, x, y+ry-2, P(HSteel, 255), own)
			for dy := -1; dy <= 1; dy++ {
				pix.line(x-1, y+dy, x+1, y+dy, P(HGold, 255), own)
			}
			lamp(0, -rx, 0, HRed)
			lamp(1, rx, 0, HLife)
		case 1: // a spindle with docking arms
			d := roll(9, seed, 5447, int64(i)) - 4
			for o := -1; o <= 1; o++ {
				pix.line(x-14, y-d+o, x+14, y+d+o, P(HSteel, 255), own)
			}
			pix.line(x-14, y-d, x+14, y+d, P(HWhite, 230), own)
			for _, e := range [2]int{-1, 1} {
				for dy := -1; dy <= 1; dy++ {
					pix.line(x+e*14-1, y+e*d+dy, x+e*14+1, y+e*d+dy, P(HGold, 255), own)
				}
			}
			for _, a := range [3]int{-7, 0, 7} {
				pix.line(x+a, y+a*d/14-5, x+a, y+a*d/14+5, P(HGlass, 255), own)
			}
			lamp(0, -7, -d/2-5, HRed)
			lamp(1, 7, d/2+5, HLife)
		default: // habitat spheres on struts
			spheres := [...][3]int{{-9, 2, 5}, {0, -3, 6}, {9, 3, 5}, {2, 9, 3}}
			tone := [...]Hue{HWhite, HCool, HSteel, HWhite}
			for _, j := range [3]int{0, 2, 3} {
				pix.line(x+spheres[1][0], y+spheres[1][1], x+spheres[j][0], y+spheres[j][1], P(HSteel, 255), own)
			}
			for j, s := range spheres {
				h := tone[j]
				pix.Disc((x+s[0])*pxUnit, (y+s[1])*pxUnit, s[2]*pxUnit, own, func(nx, ny int) Paint {
					if abs(ny) < 40 {
						return shade(P(HGlass, 255), nx, ny)
					}
					return shade(P(h, 255), nx, ny)
				})
			}
			lamp(0, 0, -10, HRed)
			lamp(1, 14, 3, HLife)
		}
	}
}
