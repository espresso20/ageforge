package mapgen

// The life of the star map's sky: a comet now and then, a nova, the dust
// shimmering, the core and the hubs breathing, the ghosts of the quantum age
// and the halo of the transcendent one. Every bit of it is a pure function of
// the animation frame (about 8 a second), so a held frame is a whole picture.

const (
	cometSlot = 72 // frames between comet slots; with the hashed start, 6 to 12 seconds
	cometLife = 18
	novaSlot  = 96 // 8 to 15 seconds between novas
	novaLife  = 14
)

// tri is a triangle wave of n, 0 up to amp and back, over a period of 2*amp.
func tri(n, amp int) int {
	n = ((n % (2 * amp)) + 2*amp) % (2 * amp)
	if n > amp {
		n = 2*amp - n
	}
	return n
}

// breathe is a slow dip of a weight: w down by up to depth over about six
// seconds, each place at its own phase.
func (f *frame) breathe(w, depth, phase int) int {
	return w - depth*tri(f.anim/3+phase, 8)/8
}

// paintSkyLife is the moving sky, over the territory and under the light.
func (f *frame) paintSkyLife() {
	f.paintComet()
	f.paintNova()
	f.paintGhosts()
	f.paintHalo()
}

// skyCell is a cell of open sky.
func (f *frame) skyCell(x, y int) bool { return f.g.in(x, y) && f.sky(Pt{x, y}) }

// paintComet flies one comet across the sky once in a while: a bright head
// and a short tail of fading dots.
func (f *frame) paintComet() {
	if f.anim <= 0 {
		return
	}
	seed := f.m.Seed
	slot := f.anim / cometSlot
	t := f.anim%cometSlot - roll(cometSlot-cometLife, seed, 380, int64(slot))
	if t < 0 || t >= cometLife {
		return
	}
	x0, y0 := 20+roll(worldW-40, seed, 381, int64(slot)), 8+roll(worldH-16, seed, 382, int64(slot))
	vx := []int{-3, -2, 2, 3}[roll(4, seed, 383, int64(slot))]
	vy := roll(3, seed, 384, int64(slot)) - 1
	for k := 4; k >= 0; k-- {
		x, y := x0+vx*(t-k)/2, y0+vy*(t-k)/2
		if !f.skyCell(x, y) {
			continue
		}
		switch k {
		case 0:
			f.g.draw(x, y, '✦', P(HWhite, 255))
		case 1:
			f.g.draw(x, y, '∙', P(HWhite, 200))
		case 2, 3:
			f.g.draw(x, y, '·', P(HStar, 150))
		default:
			f.g.draw(x, y, '·', P(HStar, 90))
		}
	}
}

// paintNova blows a star up for a moment: a flare with four arms that grows
// and then fades back to a dot.
func (f *frame) paintNova() {
	if f.anim <= 0 {
		return
	}
	seed := f.m.Seed
	slot := f.anim / novaSlot
	t := f.anim%novaSlot - roll(novaSlot-novaLife, seed, 385, int64(slot))
	if t < 0 || t >= novaLife {
		return
	}
	var x, y int
	for try := int64(0); try < 12; try++ {
		x, y = 14+roll(worldW-28, seed, 386, int64(slot), try), 6+roll(worldH-12, seed, 387, int64(slot), try)
		if f.skyCell(x, y) && f.skyCell(x-2, y) && f.skyCell(x+2, y) {
			break
		}
	}
	if !f.skyCell(x, y) {
		return
	}
	arm := func(dx, dy int, r rune, w int) {
		if f.skyCell(x+dx, y+dy) {
			f.g.draw(x+dx, y+dy, r, P(HWhite, w))
		}
	}
	switch {
	case t < 3:
		f.g.draw(x, y, '•', P(HWhite, 230))
	case t < 9:
		f.g.draw(x, y, '✦', P(HWhite, 255))
		arm(-2, 0, '+', 220)
		arm(2, 0, '+', 220)
		arm(0, -1, '+', 200)
		arm(0, 1, '+', 200)
		if t > 5 {
			arm(-4, 0, '·', 150)
			arm(4, 0, '·', 150)
		}
	default:
		f.g.draw(x, y, '*', P(HWhite, 200-(t-9)*25))
		arm(-2, 0, '·', 120-(t-9)*20)
		arm(2, 0, '·', 120-(t-9)*20)
	}
}

// homeRing is the outline of the home sectors grown by reach columns (and
// half as many rows): the cells of open sky along it, in scan order.
func (f *frame) homeRing(reach int) []Pt {
	homes := f.homeSectors()
	if len(homes) == 0 {
		return nil
	}
	in := make([]bool, worldW*worldH)
	for _, h := range homes {
		for _, c := range h.cells {
			for dy := -reach / 2; dy <= reach/2; dy++ {
				for dx := -reach; dx <= reach; dx++ {
					if x, y := c.X+dx, c.Y+dy; dx*dx+4*dy*dy <= reach*reach && inWorld(x, y) {
						in[at(x, y)] = true
					}
				}
			}
		}
	}
	var ring []Pt
	for _, p := range edgeOf(in) {
		if f.skyCell(p.X, p.Y) && !f.plan.reserved(p) {
			ring = append(ring, p)
		}
	}
	return ring
}

// paintGhosts is the quantum age's touch: a few systems at the edge of the
// home territory that are there, and then are not.
func (f *frame) paintGhosts() {
	if f.m.Catalog.Ages[f.m.AgeIdx] != "quantum_age" {
		return
	}
	ring := f.homeRing(6)
	if len(ring) == 0 {
		return
	}
	seed := f.m.Seed
	for i := int64(0); i < 10; i++ {
		p := ring[roll(len(ring), seed, 390, i)]
		phase := roll(7, seed, 393, i)
		step := (f.anim/3 + phase) % 3
		if step == 0 {
			continue // gone for a moment
		}
		glyph := '∙'
		if step == 1 {
			glyph = '✧'
		}
		f.g.draw(p.X, p.Y, glyph, P(HNeon, f.breathe(235, 80, int(i))))
	}
}

// paintHalo is the transcendent age's touch: a ring of small circles round the
// home sectors, a few cells outside the border, brighter where a slow swell
// passes.
func (f *frame) paintHalo() {
	if f.m.Catalog.Ages[f.m.AgeIdx] != "transcendent_age" {
		return
	}
	for _, p := range f.homeRing(8) {
		w := 125 + 90*tri(f.anim/2+(p.X+2*p.Y)/3, 8)/8
		f.g.draw(p.X, p.Y, '◦', P(HPlasma, w))
	}
}
