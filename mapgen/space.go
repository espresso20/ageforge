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
	if f.st.Scale == Orbit {
		f.paintHomePlanet()
		f.paintOrbitShips()
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
		if !f.close {
			c = f.breatheSky(x, y, c)
		}
		f.g.paint(x, y, c)
		f.g.mark(x, y, thing)
	}
}

func (f *frame) paintVoid() {
	sky := f.st.Sky
	thing := Thing{Kind: KindTerrain, Key: "space"}
	for y := 0; y < f.g.H; y++ {
		for x := 0; x < f.g.W; x++ {
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
	return Pt{c.X - 12 + roll(25, seed, 21, int64(age), int64(try)), c.Y - 5 + roll(11, seed, 22, int64(age), int64(try))}
}

// breatheSky is the overview's slow life: the core swelling and fading and the
// glow of the dust shimmering, each place at its own phase.
func (f *frame) breatheSky(x, y int, c Cell) Cell {
	switch w := int(c.Bg.Weight); {
	case c.Bg.Hue == HCore:
		c.Bg = P(HCore, w-w*tri(f.anim/4, 12)/60) // up to a fifth down, over about twelve seconds
	case c.Bg.Hue == HNebula && w > 18:
		ph := int(hash(f.m.Seed, 64, int64(floorDiv(x, 3)), int64(floorDiv(y, 2))) % 24)
		c.Bg = P(HNebula, w-w*tri(f.anim/5+ph, 12)/48)
	}
	return c
}

// spaceCellFunc returns a function that evaluates procedural space cells at any coordinate (wx, wy).
func (f *frame) spaceCellFunc() func(wx, wy int) Cell {
	seed := f.m.Seed
	sky := f.st.Sky
	scale := f.st.Scale
	anim := f.anim
	close := f.close
	if scale == Stars {
		return func(wx, wy int) Cell {
			return starsSpaceCell(seed, anim, close, wx, wy)
		}
	}
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

// starsSpaceCell generates procedural deep space outside the galaxy for the Stars scale.
func starsSpaceCell(seed int64, anim int, close bool, wx, wy int) Cell {
	h8 := func(salt int64, x, y int) int { return int(hash(seed, salt, int64(x), int64(y)) % 256) }
	fx := floorDiv(wx*256, 80)
	fy := floorDiv(wy*256, 40)
	tone := 6 + clamp(wx, 0, worldW-1)*7/worldW + clamp(wy, 0, worldH-1)*5/worldH + fbm(seed+31, fx, fy, 2)*3/noiseMax + h8(41, wx, wy)/140
	bg := P(HNebula, tone)
	c := Cell{R: ' ', Fg: P(HStar, 0), Bg: bg}
	if close {
		dh := int(hash(seed, 63, int64(wx), int64(wy)) % 1000)
		switch {
		case dh < 14:
			c.R, c.Fg = '∙', P(HStar, 60+h8(61, wx, wy)*40/256)
		case dh < 40:
			c.R, c.Fg = '·', P(HStar, 50+h8(62, wx, wy)*40/256)
		}
		if c.R == '∙' && (anim/9+wx*7+wy*13)%5 == 0 {
			c.R = '·'
		}
	} else {
		h := int(hash(seed, 44, int64(wx), int64(wy)) % 1000)
		switch {
		case h < 3:
			c.R, c.Fg = pick("✦✧", uint64(h8(46, wx, wy))), P(HStar, 255)
		case h < 17:
			c.R, c.Fg = pick("∙∙+", uint64(h8(47, wx, wy))), P(HStar, 150+h8(48, wx, wy)*80/256)
		case h < 72:
			c.R, c.Fg = '·', P(HStar, 70+h8(49, wx, wy)*70/256)
		}
		if (c.R == '✦' || c.R == '✧') && (anim/9+wx*7+wy*13)%7 == 0 {
			c.R = '+'
		}
		if w := int(c.Bg.Weight); c.Bg.Hue == HNebula && w > 18 {
			ph := int(hash(seed, 64, int64(floorDiv(wx, 3)), int64(floorDiv(wy, 2))) % 24)
			c.Bg = P(HNebula, w-w*tri(anim/5+ph, 12)/48)
		}
	}
	return c
}

// paintOrbitShips flies the pixel ships over the home planet's disc: between
// the planet and the two edges of the picture, behind the planet where they
// pass it.
func (f *frame) paintOrbitShips() {
	c, seed := f.plan.Centre, f.m.Seed
	cx, cy := (2*c.X+1)*pxUnit, (2*c.Y+1)*pxUnit
	bodies := []shipBody{{cx, cy, 2 * discRX * pxUnit}}
	for _, side := range []int{-1, 1} {
		bodies = append(bodies, shipBody{cx + side*(2*discRX+60)*pxUnit, cy + (roll(21, seed, 5212, int64(side))-10)*pxUnit, 1})
	}
	pix := NewPixels(f.g.W, f.g.H)
	pix.Ships(seed, f.anim, bodies, nil, nil)
	pix.Resolve(f.g)
	pix.Release()
}
