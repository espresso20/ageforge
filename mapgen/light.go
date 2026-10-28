package mapgen

// paintLight is the last layer: dusk and night as a pass over the finished
// picture. Every cell is pulled toward the dark by a dim that follows the
// clock; windows, lamps and the cities of a planet are exempt and glow.
func (f *frame) paintLight() {
	f.g.Wall, f.g.Tint = f.st.Wall, f.st.Night.Tint
	if f.night == 0 {
		return
	}
	n := f.st.Night
	dusk := clamp(n.Dim*f.night/256, 0, 255)
	for i := range f.g.Cells {
		f.g.Cells[i].Dim = dimmer(f.g.Cells[i].Dim, uint8(dusk))
	}
	if f.night < 128 {
		return
	}
	for _, p := range f.lights {
		f.glowAt(p, n.Light, 15)
	}
	lit := 35 + min(55, f.m.Workers.Pop/60)
	for _, p := range f.glow {
		if chance(lit, f.m.Seed, 80, int64(p.X), int64(p.Y)) {
			f.glowAt(p, n.Light, 40)
		}
	}
	if n.Lamps {
		f.lampRoads(n.Light)
	}
}

// dimmer is two dims together: each takes its share of what the other left.
func dimmer(a, b uint8) uint8 { return uint8(255 - (255-int(a))*(255-int(b))/255) }

// glowAt lights a cell: its glyph in the light, its ground warmed by it, and
// no dark over it.
func (f *frame) glowAt(p Pt, light Hue, warm int) {
	c := &f.g.Cells[at(p.X, p.Y)]
	c.Fg, c.Bg, c.Dim = P(light, 255), P(light, warm*f.night/256), 0
}

// lampRoads sets a lamp on every few laid road cells.
func (f *frame) lampRoads(light Hue) {
	for y := 0; y < worldH; y++ {
		for x := 0; x < worldW; x++ {
			if f.laid(x, y) && (x*5+y*3)%11 == 0 {
				f.glowAt(Pt{x, y}, light, 70)
			}
		}
	}
}
