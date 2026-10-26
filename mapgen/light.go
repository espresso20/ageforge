package mapgen

// paintLight is the last layer: dusk and night as a pass over the finished
// picture. Colours slide toward the style's night tint, then lit windows and
// street lamps glow on top.
func (f *frame) paintLight() {
	if f.night == 0 {
		return
	}
	n := f.st.Night
	dim := n.Dim * f.night / 256
	for i := range f.g.Cells {
		c := &f.g.Cells[i]
		c.Bg = mix(c.Bg, n.Tint, dim)
		c.Fg = mix(c.Fg, n.Tint, dim*3/4)
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

// glowAt lights a cell: its glyph at full light, its background warmed.
func (f *frame) glowAt(p Pt, light RGB, warm int) {
	c := &f.g.Cells[at(p.X, p.Y)]
	c.Fg = light
	c.Bg = mix(c.Bg, light, warm*f.night/256)
}

// lampRoads sets a lamp on every few laid road cells.
func (f *frame) lampRoads(light RGB) {
	for y := 0; y < worldH; y++ {
		for x := 0; x < worldW; x++ {
			if f.laid(x, y) && (x*5+y*3)%11 == 0 {
				f.glowAt(Pt{x, y}, light, 70)
			}
		}
	}
}
