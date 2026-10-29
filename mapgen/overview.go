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

// Overview squeezes the town of a scene into w by h cells.
func (s *Scene) Overview(w, h int) *Overview {
	b := s.Bounds
	step := max(256, (b.W*256+w-1)/w, (b.H*256+h-1)/h)
	if !s.Fill {
		step = (step + 255) / 256 * 256 // whole cells to a block
	}
	o := &Overview{Grid: newGrid(w, h), Scale: (step + 255) / 256, step: step, picks: make([]Pt, w*h)}
	o.Grid.Wall, o.Grid.Tint = s.Grid.Wall, s.Grid.Tint
	o.Grid.Things = s.Grid.Things
	o.X0 = clamp(b.X+b.W/2-w*step/512, 0, max(0, worldW-w*step/256))
	o.Y0 = clamp(b.Y+b.H/2-h*step/512, 0, max(0, worldH-h*step/256))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			p := s.mostImportant(o.blockX(x), o.blockY(y), o.blockX(x+1), o.blockY(y+1))
			o.picks[y*w+x] = p
			o.Grid.paint(x, y, s.summary(p))
			if s.Grid.in(p.X, p.Y) { // a block past the edge of the world stays blank
				o.Grid.Hit[y*w+x] = s.Grid.Hit[at(p.X, p.Y)]
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
	best, bestSal := Pt{(x0 + x1) / 2, (y0 + y1) / 2}, 0
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			if !s.Grid.in(x, y) {
				continue
			}
			th := s.Grid.Things[s.Grid.Hit[at(x, y)]]
			sal := salience[th.Kind]
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
	return best
}

// summary is the cell drawn for a world cell standing for a block.
func (s *Scene) summary(p Pt) Cell {
	c := s.Grid.At(p.X, p.Y)
	th := s.Grid.ThingAt(p.X, p.Y)
	switch {
	case th.Kind == KindBuilding && !th.Ruin:
		c.R, c.Fg = lineSigns[th.Lineage], P(lineHue(th.Lineage), 255)
	case th.Kind == KindWonder:
		c.R, c.Fg = '★', P(HGold, 255)
	}
	return c
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
