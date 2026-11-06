package mapgen

// An overview is the whole town squeezed into a small window: each cell
// stands for a square block of the world and shows the most important thing
// in it. Buildings show as the sign of their line, wonders as a star, so a
// town that fills a screen reads at the size of a postage stamp, and the
// cursor can still ask what any cell is.

// Overview is a compressed view of a scene.
type Overview struct {
	Grid *Grid
	// X0, Y0 is the world cell at the top-left of the first block, and Scale
	// the side of a block in world cells, rounded up. The blocks themselves
	// are a fraction of a cell wider than that, in 256ths, so that the town
	// fills the window instead of the nearest whole step below it.
	X0, Y0, Scale int
	step          int  // the side of a block, in 256ths of a world cell
	picks         []Pt // the world cell each cell shows
}

// SystemViewMode selects the framing for the astronomical system map.
type SystemViewMode int

const (
	SystemViewFull   SystemViewMode = 0 // Full 8-planet planetary system (Sol out to Neptune & Pluto)
	SystemViewCosmos SystemViewMode = 1 // Cosmic deep space perspective with outer Kuiper padding
	SystemViewInner  SystemViewMode = 2 // Inner solar system (Sol, Mercury, Venus, Earth, Mars, the edge of the belt)
)

// Overview squeezes the town of a scene into w by h cells.
func (s *Scene) Overview(w, h int) *Overview {
	if s.Scale == System {
		return s.overviewSystem(w, h, SystemViewFull)
	}
	return s.overviewPadded(w, h, 0, 0)
}

// SystemOverview squeezes the solar system in the given view into w by h
// cells; other scales have no system and give the plain overview.
func (s *Scene) SystemOverview(w, h int, mode SystemViewMode) *Overview {
	if s.Scale != System {
		return s.Overview(w, h)
	}
	return s.overviewSystem(w, h, mode)
}

// OverviewCosmos squeezes the scene with additional cosmic padding into w by h cells,
// giving a wide deep space perspective.
func (s *Scene) OverviewCosmos(w, h int) *Overview {
	if s.Scale == System {
		return s.overviewSystem(w, h, SystemViewCosmos)
	}
	return s.overviewPadded(w, h, 100, 50)
}

// OverviewInner focuses on the inner planetary system (Sol through Jupiter)
// with generous orbital spacing, allowing players to see inside the Sol system
// while maintaining an astronomical perspective.
func (s *Scene) OverviewInner(w, h int) *Overview {
	if s.Scale == System {
		return s.overviewSystem(w, h, SystemViewInner)
	}
	return s.overviewPadded(w, h, 0, 0)
}

func (s *Scene) overviewPadded(w, h, padX, padY int) *Overview {
	b := s.Bounds
	bw, bh := b.W+padX, b.H+padY
	gw, gh := s.Grid.W, s.Grid.H
	if s.Grid.Space && b.W >= gw {
		bw += 16
		bh += 8
	}
	step := max(256, (bw*256+w-1)/w, (bh*256+h-1)/h)
	if !s.Fill {
		step = (step + 255) / 256 * 256 // whole cells to a block
	}
	o := &Overview{Grid: newGrid(w, h), Scale: max(1, step/256), step: step, picks: make([]Pt, w*h)}
	o.Grid.Wall, o.Grid.Tint = s.Grid.Wall, s.Grid.Tint
	o.Grid.Space = s.Grid.Space
	o.Grid.Things = s.Grid.Things[:len(s.Grid.Things):len(s.Grid.Things)] // appends must not reach the scene's
	spanX := w * step / 256
	if spanX > gw {
		o.X0 = (gw - spanX) / 2
	} else {
		o.X0 = clamp(b.X+b.W/2-spanX/2, 0, gw-spanX)
	}
	spanY := h * step / 256
	if spanY > gh {
		o.Y0 = (gh - spanY) / 2
	} else {
		o.Y0 = clamp(b.Y+b.H/2-spanY/2, 0, gh-spanY)
	}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			x0, y0 := o.blockX(x), o.blockY(y)
			x1, y1 := o.blockX(x+1), o.blockY(y+1)
			if x1 > 0 && x0 < s.Grid.W && y1 > 0 && y0 < s.Grid.H {
				p := s.mostImportant(x0, y0, x1, y1)
				o.picks[y*w+x] = p
				o.Grid.paint(x, y, s.summary(p))
				o.Grid.Hit[y*w+x] = s.Grid.Hit[p.Y*s.Grid.W+p.X]
			} else {
				bx := x0 + step/512
				by := y0 + step/512
				o.picks[y*w+x] = Pt{bx, by}
				if s.SpaceCell != nil {
					o.Grid.paint(x, y, s.SpaceCell(bx, by))
				} else if s.GroundCell != nil {
					o.Grid.paint(x, y, s.GroundCell(bx, by))
				} else {
					o.Grid.paint(x, y, s.groundEdgeCell(bx, by))
				}
				o.Grid.Hit[y*w+x] = 0
			}
		}
	}
	return o
}

// blockX and blockY are the world column and row where a block begins.
func (o *Overview) blockX(x int) int { return o.X0 + x*o.step/256 }
func (o *Overview) blockY(y int) int { return o.Y0 + y*o.step/256 }

// salience says how much a kind of thing deserves to stand for a block.
var salience = [...]int{KindNone: 0, KindTerrain: 0, KindRoad: 2, KindAgent: 1, KindAmbient: 3, KindBuilding: 5, KindWonder: 6, KindVisitor: 7, KindCiv: 6, KindRoute: 2, KindMission: 4, KindHarbinger: 6, KindIdle: 4, KindSector: 5, KindLane: 2}

// mostImportant is the cell of a block that stands for it: the most
// important thing in it, or the cell in its middle when there is nothing.
func (s *Scene) mostImportant(x0, y0, x1, y1 int) Pt {
	x1, y1 = max(x1, x0+1), max(y1, y0+1)
	cx, cy := floorDiv(x0+x1, 2), floorDiv(y0+y1, 2)
	best, bestSal := Pt{cx, cy}, 0
	foundInGrid := false
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			if !s.Grid.in(x, y) {
				continue
			}
			foundInGrid = true
			th := s.Grid.Things[s.Grid.Hit[y*s.Grid.W+x]]
			sal := salience[th.Kind]
			if th.Kind == KindTerrain && (th.Key == "planet" || th.Key == "home" || th.Key == "sun" || th.Key == "star" || th.Key == "space_station") {
				sal = 4
			}
			if th.Kind == KindBuilding && th.Ruin {
				sal = 2
			}
			if th.Kind == KindSector && th.Key == "border" {
				sal = 1
			}
			if th.Kind == KindSector && th.Ref == 1 {
				sal++ // a hub stands for its block before the systems round it
			}
			if sal > bestSal {
				best, bestSal = Pt{x, y}, sal
			}
		}
	}
	if !s.Grid.in(best.X, best.Y) && foundInGrid {
		best = Pt{clamp(cx, 0, s.Grid.W-1), clamp(cy, 0, s.Grid.H-1)}
	}
	return best
}

// summary is the cell drawn for a world cell standing for a block.
func (s *Scene) summary(p Pt) Cell {
	if !s.Grid.in(p.X, p.Y) {
		if s.SpaceCell != nil {
			return s.SpaceCell(p.X, p.Y)
		}
		if s.GroundCell != nil {
			return s.GroundCell(p.X, p.Y)
		}
		return s.groundEdgeCell(p.X, p.Y)
	}
	c := s.Grid.At(p.X, p.Y)
	if s.SpaceCell != nil {
		return c
	}
	th := s.Grid.ThingAt(p.X, p.Y)
	switch {
	case th.Kind == KindBuilding && !th.Ruin:
		c.R, c.Fg = lineSigns[th.Lineage], P(lineHue(th.Lineage), 255)
	case th.Kind == KindWonder:
		c.R, c.Fg = '★', P(HGold, 255)
	}
	return c
}

// groundEdgeCell samples mirrored terrain or clamped edge ground for coordinates outside the world.
func (s *Scene) groundEdgeCell(wx, wy int) Cell {
	if s.Grid.W == 0 || s.Grid.H == 0 {
		return Cell{}
	}
	mx := mirrorCoord(wx, s.Grid.W)
	my := mirrorCoord(wy, s.Grid.H)
	i := my*s.Grid.W + mx
	terrain := func(idx int) bool {
		th := s.Grid.Things[s.Grid.Hit[idx]]
		if th.Key == "home" {
			return false
		}
		return th.Kind == KindNone || th.Kind == KindTerrain
	}
	if terrain(i) {
		return s.Grid.Cells[i]
	}
	dx := 1
	if mx >= s.Grid.W/2 {
		dx = -1
	}
	sx := mx
	for n := 0; n < 64 && !terrain(i) && sx+dx >= 0 && sx+dx < s.Grid.W; n++ {
		sx += dx
		i = my*s.Grid.W + sx
	}
	if !terrain(i) {
		dx = -dx
		sx = mx
		i = my*s.Grid.W + sx
		for n := 0; n < 64 && !terrain(i) && sx+dx >= 0 && sx+dx < s.Grid.W; n++ {
			sx += dx
			i = my*s.Grid.W + sx
		}
	}
	if terrain(i) {
		return s.Grid.Cells[i]
	}
	cx := clamp(wx, 0, s.Grid.W-1)
	cy := clamp(wy, 0, s.Grid.H-1)
	c := s.Grid.Cells[cy*s.Grid.W+cx]
	if !terrain(cy*s.Grid.W + cx) {
		c.R = ' '
	}
	return c
}

func mirrorCoord(v, n int) int {
	if n <= 0 {
		return 0
	}
	v = ((v % (2 * n)) + 2*n) % (2 * n)
	if v >= n {
		v = 2*n - 1 - v
	}
	return v
}

// CellOf is the cell that shows a world cell, and whether any does.
func (o *Overview) CellOf(p Pt) (x, y int, ok bool) {
	x, y = floorDiv((p.X-o.X0)*256, o.step), floorDiv((p.Y-o.Y0)*256, o.step)
	for x > 0 && o.blockX(x) > p.X {
		x--
	}
	for o.blockX(x+1) <= p.X && x < o.Grid.W {
		x++
	}
	for y > 0 && o.blockY(y) > p.Y {
		y--
	}
	for o.blockY(y+1) <= p.Y && y < o.Grid.H {
		y++
	}
	return x, y, x >= 0 && y >= 0 && x < o.Grid.W && y < o.Grid.H
}

// Pick is the world cell a cell shows.
func (o *Overview) Pick(x, y int) Pt {
	if x < 0 || y < 0 || x >= o.Grid.W || y >= o.Grid.H {
		return Pt{-1, -1}
	}
	return o.picks[y*o.Grid.W+x]
}

func floorDiv(a, b int) int {
	if a < 0 {
		return -((-a + b - 1) / b)
	}
	return a / b
}

// overviewSystem draws the orrery of the solar system at far zoom, avoiding
// point-sampling artifacts on detailed 3D spheres.
func (s *Scene) overviewSystem(w, h int, mode SystemViewMode) *Overview {
	gw, gh := s.Grid.W, s.Grid.H
	var bw, bh int
	switch mode {
	case SystemViewCosmos:
		bw, bh = gw+80, gh+40
	case SystemViewInner:
		bw, bh = (gw*58)/100, (gh*58)/100
	default:
		bw, bh = gw, gh
	}
	step := max(256, (bw*256+w-1)/w, (bh*256+h-1)/h)
	o := &Overview{Grid: newGrid(w, h), Scale: max(1, step/256), step: step, picks: make([]Pt, w*h)}
	o.Grid.Wall, o.Grid.Tint = s.Grid.Wall, s.Grid.Tint
	o.Grid.Space = true
	o.Grid.Things = s.Grid.Things[:len(s.Grid.Things):len(s.Grid.Things)] // appends must not reach the scene's

	spanX := w * step / 256
	o.X0 = (gw - spanX) / 2
	spanY := h * step / 256
	o.Y0 = (gh - spanY) / 2

	// 1. Fill base space background with stars & nebula
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			wx := o.blockX(x)
			wy := o.blockY(y)
			o.picks[y*w+x] = Pt{wx, wy}
			if s.SpaceCell != nil {
				c := s.SpaceCell(wx, wy)
				if mode == SystemViewCosmos && c.Bg.Hue == HNebula { // the Milky Way is the glow here
					c.Bg = P(HVoid, 255)
				}
				o.Grid.paint(x, y, c)
			} else {
				o.Grid.paint(x, y, Cell{Bg: P(HVoid, 255)})
			}
		}
	}

	cx := w / 2
	cy := h / 2

	// 2. How far out the system reaches; at the cosmos zoom it leaves room
	// for the Kuiper belt round it.
	maxRY := cy * 88 / 100
	switch mode {
	case SystemViewCosmos:
		maxRY = min(cy*40/100, cx*13/100)
	case SystemViewInner:
		maxRY = cy * 92 / 100
	}
	maxRY = max(8, maxRY)

	// 3. The system itself, drawn in quadrant blocks (orrery.go): glow, orbit
	// lines, belts, star, planets with their moons, the dwarf planet, the comet.
	s.paintOrrery(o, &SolSpec, mode, cx, cy, maxRY)

	// 7. Outer Megastructures: Warp Gate
	warpRY := (maxRY * 115) / 100
	warpY := cy + warpRY
	if s.Gate && o.Grid.in(cx, warpY) {
		if o.Grid.in(cx-1, warpY) {
			o.Grid.paint(cx-1, warpY, Cell{R: '⫷', Fg: P(HCyan, 255), Bg: o.Grid.At(cx-1, warpY).Bg})
			o.Grid.mark(cx-1, warpY, Thing{Kind: KindBuilding, Key: "warp_gate"})
		}
		o.Grid.paint(cx, warpY, Cell{R: '◎', Fg: P(HNeon, 255), Bg: P(HCyan, 220)})
		o.Grid.mark(cx, warpY, Thing{Kind: KindBuilding, Key: "warp_gate"})
		if o.Grid.in(cx+1, warpY) {
			o.Grid.paint(cx+1, warpY, Cell{R: '⫸', Fg: P(HCyan, 255), Bg: o.Grid.At(cx+1, warpY).Bg})
			o.Grid.mark(cx+1, warpY, Thing{Kind: KindBuilding, Key: "warp_gate"})
		}
	}

	// 10. Rare Visitors & Alien Craft (Sync with Scene Grid)
	if s.Grid != nil {
		for i := range s.Grid.Cells {
			th := s.Grid.Things[s.Grid.Hit[i]]
			if th.Kind == KindVisitor {
				wx, wy := i%s.Grid.W, i/s.Grid.W
				ox, oy, ok := o.CellOf(Pt{wx, wy})
				if ok && o.Grid.in(ox, oy) {
					o.Grid.paint(ox, oy, s.Grid.At(wx, wy))
					o.Grid.mark(ox, oy, th)
					o.picks[oy*w+ox] = Pt{wx, wy}
					o.Grid.Hit[oy*w+ox] = s.Grid.Hit[i]
				}
			}
		}
	}

	return o
}
