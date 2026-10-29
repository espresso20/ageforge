package mapgen

// The ground of the space scales: a void with a drift of nebula and a
// sprinkle of stars, then the scale's own bodies on it.

// paintSpace draws the void and the bodies of an orbit, a solar system or a
// star map.
func (f *frame) paintSpace() {
	if f.st.Scale == Stars {
		f.paintGalaxy()
		return
	}
	f.paintVoid()
	f.paintStarField()
	f.paintTracks()
	switch f.st.Scale {
	case Orbit:
		f.paintHomePlanet()
	case System:
		f.paintSystem()
	}
}

// paintGalaxy lays down the galaxy of the seed, already painted in the plan.
func (f *frame) paintGalaxy() {
	thing := Thing{Kind: KindTerrain, Key: "space"}
	sky := f.plan.Galaxy.Sky
	if f.close {
		sky = f.plan.Galaxy.Deep
	}
	for i, c := range sky {
		x, y := i%worldW, i/worldW
		if (c.R == '✦' || c.R == '✧') && (f.anim/9+x*7+y*13)%7 == 0 {
			c.R = '+' // the bright stars take turns to twinkle
		}
		if f.close && c.R == '∙' && (f.anim/9+x*7+y*13)%5 == 0 {
			c.R = '·' // and the dim ones of the close sky blink down a step
		}
		f.g.paint(x, y, c)
		f.g.mark(x, y, thing)
	}
}

func (f *frame) paintVoid() {
	sky := f.st.Sky
	thing := Thing{Kind: KindTerrain, Key: "space"}
	for y := 0; y < worldH; y++ {
		for x := 0; x < worldW; x++ {
			bg := sky.Bg
			if haze := max(0, fbm(f.m.Seed+11, x*256/24, y*256/12, 3)-540) * 60 / (noiseMax - 540); haze > 0 {
				bg = P(HNebula, haze)
			}
			f.g.paint(x, y, Cell{R: ' ', Fg: sky.Fg, Bg: bg})
			f.g.mark(x, y, thing)
		}
	}
}

// paintStarField puts one candidate star in each small block of the sky.
func (f *frame) paintStarField() {
	sky := f.st.Sky
	for by := 0; by < worldH/3; by++ {
		for bx := 0; bx < worldW/5; bx++ {
			x := bx*5 + roll(5, f.m.Seed, 12, int64(bx), int64(by))
			y := by*3 + roll(3, f.m.Seed, 13, int64(bx), int64(by))
			if !chance(sky.Dense, f.m.Seed, 14, int64(bx), int64(by)) {
				continue
			}
			h := hash(f.m.Seed, 15, int64(x), int64(y))
			shine := 120 + int(h>>8)%136
			if (f.anim/9+int(h>>20))%7 == 0 {
				shine /= 2 // a slow twinkle
			}
			f.g.draw(x, y, pick(sky.Glyphs, h>>30), sky.Fg.fade(shine))
		}
	}
}

// paintTracks draws every orbit as a faint dotted line, laid or not; the
// lanes that are laid are drawn over it.
func (f *frame) paintTracks() {
	if f.st.Scale != Orbit && f.st.Scale != System {
		return
	}
	for _, r := range f.plan.Roads[2:] {
		for _, c := range r.Cells {
			f.g.draw(c.X, c.Y, '·', P(HOrbit, 110))
		}
	}
}

// paintRelics shows the wonders of the ages before the space ages as bright
// stars where the people can see them: on the home planet, on the orbit of
// the home world, or among the stars near home. A wonder's star.
func (f *frame) paintRelics() {
	if f.st.Scale <= City {
		return
	}
	for _, w := range f.m.Wonders {
		if w.Age >= groundAges && f.st.Scale != Stars || !w.Built {
			continue
		}
		for try := 0; try < 60; try++ {
			p := f.relicSpot(w.Age, try)
			if f.busy[at(p.X, p.Y)] || f.g.ThingAt(p.X, p.Y).Kind == KindWonder || f.st.Scale == Stars && (!f.sky(p) || f.plan.reserved(p)) {
				continue
			}
			f.g.draw(p.X, p.Y, '★', P(HGold, 255))
			f.g.mark(p.X, p.Y, Thing{Kind: KindWonder, Key: w.Key})
			f.targets = append(f.targets, p)
			if !f.close || f.st.Scale != Stars { // the star needs no box of light on plain space
				f.lights = append(f.lights, p)
			}
			break
		}
	}
}

// relicSpot is the try-th place for the star of the wonder of an age.
func (f *frame) relicSpot(age, try int) Pt {
	c, seed := f.plan.Centre, f.m.Seed
	switch f.st.Scale {
	case Orbit:
		return f.orbitRelic(age, try)
	case System:
		home := f.plan.Roads[2+2*homeRing+roll(2, seed, 19, int64(age))].Cells // the orbit of the third world
		return home[roll(len(home), seed, 20, int64(age), int64(try))]
	}
	return Pt{c.X - 12 + roll(25, seed, 21, int64(age), int64(try)), c.Y - 5 + roll(11, seed, 22, int64(age), int64(try))}
}
