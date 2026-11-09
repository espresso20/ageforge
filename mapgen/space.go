package mapgen

// The ground of the space scales: a void with a drift of nebula and a
// sprinkle of stars, then the scale's own bodies on it.

// paintSpace draws the void and the bodies of an orbit, a solar system or a
// star map.
func (f *frame) paintSpace() {
	if f.st.Scale == Stars {
		f.paintVoid()
		return
	}
	f.paintVoid()
	f.paintStarField()
	f.paintTracks()
	if f.st.Scale == Orbit {
		f.paintHomePlanet()
	}
}

func (f *frame) paintVoid() {
	sky := f.st.Sky
	thing := Thing{Kind: KindTerrain, Key: "space"}
	for y := 0; y < f.g.H; y++ {
		for x := 0; x < f.g.W; x++ {
			bg := sky.Bg
			if f.st.Scale == Stars {
				// the star map's looks are the wide views; this is only its flat ground
			} else if haze := max(0, fbm(f.m.Seed+11, x*256/24, y*256/12, 3)-540) * 60 / (noiseMax - 540); haze > 0 {
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
	for by := 0; by < f.g.H/3; by++ {
		for bx := 0; bx < f.g.W/5; bx++ {
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
	if f.st.Scale != Orbit {
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
			if f.busy[f.at(p.X, p.Y)] || f.g.ThingAt(p.X, p.Y).Kind == KindWonder || f.st.Scale == Stars && (!f.sky(p) || f.plan.reserved(p)) {
				continue
			}
			f.g.draw(p.X, p.Y, '★', P(HGold, 255))
			f.g.mark(p.X, p.Y, Thing{Kind: KindWonder, Key: w.Key})
			f.targets = append(f.targets, p)
			if f.st.Scale != Stars { // a star on the star map needs no box of light
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
	return Pt{X: c.X - 12 + roll(25, seed, 21, int64(age), int64(try)), Y: c.Y - 5 + roll(11, seed, 22, int64(age), int64(try))}
}

// spaceCellFunc returns a function that evaluates procedural space cells at any coordinate (wx, wy).
func (f *frame) spaceCellFunc() func(wx, wy int) Cell {
	seed := f.m.Seed
	sky := f.st.Sky
	anim := f.anim
	return func(wx, wy int) Cell {
		return voidSpaceCell(seed, sky, anim, wx, wy)
	}
}

// voidSpaceCell generates the procedural nebula background and stars for Orbit and System scales.
func voidSpaceCell(seed int64, sky Look, anim int, wx, wy int) Cell {
	fx := floorDiv(wx*256, 24)
	fy := floorDiv(wy*256, 12)
	bg := sky.Bg
	if haze := max(0, fbm(seed+11, fx, fy, 3)-540) * 60 / (noiseMax - 540); haze > 0 {
		bg = P(HNebula, haze)
	}
	cell := Cell{R: ' ', Fg: sky.Fg, Bg: bg}

	bx := floorDiv(wx, 5)
	by := floorDiv(wy, 3)
	starX := bx*5 + roll(5, seed, 12, int64(bx), int64(by))
	starY := by*3 + roll(3, seed, 13, int64(bx), int64(by))
	if starX == wx && starY == wy && chance(sky.Dense, seed, 14, int64(bx), int64(by)) {
		h := hash(seed, 15, int64(wx), int64(wy))
		shine := 120 + int(h>>8)%136
		if (anim/9+int(h>>20))%7 == 0 {
			shine /= 2 // a slow twinkle
		}
		if len(sky.Glyphs) > 0 {
			cell.R = pick(sky.Glyphs, h>>30)
		}
		cell.Fg = sky.Fg.fade(shine)
	}
	return cell
}

// paintOrbitStation sets one station beside the home planet: a fixed hull
// pixel with a blinking light above it.
func (f *frame) paintOrbitStation() {
	c, seed := f.plan.Centre, f.m.Seed
	x, y := (2*c.X+1)-2*discRX-10, (2*c.Y+1)-8
	pix := NewPixels(f.g.W, f.g.H)
	own := pix.Owner(Thing{Kind: KindTerrain, Key: "space_station"}, 6)
	pix.Set(x, y, P(HSteel, 255), own)
	pix.light(seed, f.anim, 0, x, y-1, HRed, own)
	pix.Resolve(f.g)
	pix.Release()
}
