package mapgen

// neighbourMask marks which of the four neighbours N, E, S, W of (x, y) satisfy
// test: bits 1, 2, 4, 8, the index of a join glyph set.
func neighbourMask(x, y int, test func(x, y int) bool) int {
	mask := 0
	for i, d := range neighbours {
		if test(x+d[0], y+d[1]) {
			mask |= 1 << i
		}
	}
	return mask
}

// paintGround lays the ground of the scene: land by biome with rivers
// joined up, or space with its planet.
func (f *frame) paintGround() {
	if f.st.Scale > City {
		f.paintSpace()
		return
	}
	wet := func(x, y int) bool { return !inWorld(x, y) || f.fields.Biome[at(x, y)].Water() }
	for y := 0; y < worldH; y++ {
		for x := 0; x < worldW; x++ {
			b := f.groundAt(x, y)
			look := f.st.Land[b]
			var c Cell
			switch b {
			case River:
				c = Cell{R: joinRune(joinsRounded, neighbourMask(x, y, wet), 0), Fg: look.Fg, Bg: look.Bg}
			case Deep, Shallows:
				// each cell changes glyph on its own beat, so the water shimmers
				c = look.at(x, y, int64((f.anim+roll(12, int64(x), int64(y)))/12))
			default:
				c = look.at(x, y, 0)
			}
			f.g.paint(x, y, c)
			f.g.mark(x, y, Thing{Kind: KindTerrain, Key: biomeNames[b]})
		}
	}
}

// groundAt is the biome a cell is drawn as. In a city the people have cleared
// the forest and scrub round the middle into lawns and parks, as far out as
// their numbers reach.
func (f *frame) groundAt(x, y int) Biome {
	b := f.fields.Biome[at(x, y)]
	if f.st.Scale != City || (b != Forest && b != Scrub) {
		return b
	}
	pop := f.m.Workers.Pop
	rx, ry := 24+pop/60, 8+pop/180
	dx, dy := x-f.plan.Centre.X, y-f.plan.Centre.Y
	if dx*dx*ry*ry+dy*dy*rx*rx < rx*rx*ry*ry {
		return Grass
	}
	return b
}

// groundCellFunc evaluates seamless procedural wilderness and ocean outside the
// world grid so wide viewports never mirror continents, rivers, or city yards.
func (f *frame) groundCellFunc() func(wx, wy int) Cell {
	seed := f.m.Seed
	st := *f.st
	anim := f.anim
	dusk := clamp(st.Night.Dim*f.night/256, 0, 255)
	coast := coasts[roll(len(coasts), seed, 1)]
	return func(wx, wy int) Cell {
		noise := fbm(seed, floorDiv(wx*256, 30), floorDiv(wy*256, 15), 4)
		reach := (wx-worldW/2)*coast[0] + 2*(wy-worldH/2)*coast[1]
		h := clamp(noise*2/5+235-max(0, reach-46)*8, 0, noiseMax)
		moist := fbm(seed+101, floorDiv(wx*256, 40), floorDiv(wy*256, 20), 4)
		var b Biome
		switch {
		case h < deepBelow:
			b = Deep
		case h < seaLevel:
			b = Shallows
		case h < beachTop:
			b = Sand
		default:
			wet := clamp(moist*3/noiseMax, 0, 2)
			band := 0
			if h >= hillsAt {
				band = 1
			}
			if h >= rangesAt {
				band = 2
			}
			b = biomeBands[band][wet]
		}
		look := st.Land[b]
		var c Cell
		switch b {
		case Deep, Shallows:
			c = look.at(wx, wy, int64((anim+roll(12, int64(wx), int64(wy)))/12))
		default:
			c = look.at(wx, wy, 0)
		}
		bx, by := floorDiv(wx, spotW), floorDiv(wy, spotH)
		sx := bx*spotW + roll(spotW, seed, 60, int64(bx), int64(by))
		sy := by*spotH + roll(spotH, seed, 61, int64(bx), int64(by))
		if sx == wx && sy == wy {
			for _, s := range st.Scatter {
				if s.On == b && chance(s.Chance, seed, 62, int64(wx), int64(wy)) {
					c.R = pick(s.Glyphs, hash(seed, 63, int64(wx), int64(wy)))
					c.Fg = s.Fg
					break
				}
			}
		}
		if dusk > 0 {
			c.Dim = dimmer(c.Dim, uint8(dusk))
		}
		return c
	}
}
