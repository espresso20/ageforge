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
