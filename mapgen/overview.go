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
	// the side of a block in world cells.
	X0, Y0, Scale int
	picks         []Pt // the world cell each cell shows
}

// Overview squeezes the town of a scene into w by h cells.
func (s *Scene) Overview(w, h int) *Overview {
	b := s.Bounds
	scale := max(1, (b.W+w-1)/w, (b.H+h-1)/h)
	o := &Overview{Grid: newGrid(w, h), Scale: scale, picks: make([]Pt, w*h)}
	o.Grid.Wall, o.Grid.Tint = s.Grid.Wall, s.Grid.Tint
	o.Grid.Things = s.Grid.Things
	o.X0 = clamp(b.X+b.W/2-w*scale/2, 0, max(0, worldW-w*scale))
	o.Y0 = clamp(b.Y+b.H/2-h*scale/2, 0, max(0, worldH-h*scale))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			p := s.mostImportant(o.X0+x*scale, o.Y0+y*scale, scale)
			o.picks[y*w+x] = p
			o.Grid.paint(x, y, s.summary(p))
			if s.Grid.in(p.X, p.Y) { // a block past the edge of the world stays blank
				o.Grid.Hit[y*w+x] = s.Grid.Hit[at(p.X, p.Y)]
			}
		}
	}
	return o
}

// salience says how much a kind of thing deserves to stand for a block.
var salience = [...]int{KindNone: 0, KindTerrain: 0, KindRoad: 2, KindAgent: 1, KindAmbient: 3, KindBuilding: 5, KindWonder: 6}

// mostImportant is the cell of a block that stands for it: the most
// important thing in it, or the cell in its middle when there is nothing.
func (s *Scene) mostImportant(x0, y0, scale int) Pt {
	best, bestSal := Pt{x0 + scale/2, y0 + scale/2}, 0
	for y := y0; y < y0+scale; y++ {
		for x := x0; x < x0+scale; x++ {
			if !s.Grid.in(x, y) {
				continue
			}
			th := s.Grid.Things[s.Grid.Hit[at(x, y)]]
			sal := salience[th.Kind]
			if th.Kind == KindBuilding && th.Ruin {
				sal = 2
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
	x, y = floorDiv(p.X-o.X0, o.Scale), floorDiv(p.Y-o.Y0, o.Scale)
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
