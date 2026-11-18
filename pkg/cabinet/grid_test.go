package cabinet

import (
	"image"
	"testing"
)

func TestPlanGridShapes(t *testing.T) {
	for _, s := range traySizes[:2] {
		g, ok := PlanGrid(s[0], s[1], s[2], s[3])
		if !ok || g.Per < 8 || len(g.Spots) != g.Per {
			t.Fatalf("PlanGrid%v: ok=%v per=%d", s, ok, g.Per)
		}
		for i := range g.Spots {
			w, p := g.well(i), g.pix(g.Plates[i])
			if !w.In(g.lining) || !p.In(g.lining) || w.Overlaps(p) {
				t.Fatalf("%v well %d: %v plate %v lining %v", s, i, w, p, g.lining)
			}
			for j := i + 1; j < g.Per; j++ {
				if w.Overlaps(g.well(j)) || p.Overlaps(g.pix(g.Plates[j])) {
					t.Fatalf("%v wells %d and %d overlap", s, i, j)
				}
			}
			c := g.Spots[i]
			if got := g.At(c.X/g.Cell.X, c.Y/g.Cell.Y); got != i {
				t.Fatalf("%v At centre of %d = %d", s, i, got)
			}
		}
	}
}

func TestGridPaints(t *testing.T) {
	s := traySizes[1]
	g, _ := PlanGrid(s[0], s[1], s[2], s[3])
	gr := Grid{Slot: 23, Held: []bool{true}}
	if im := GridOpen(g, gr, g.Per-1); im.Bounds().Size() != image.Pt(s[0]*s[2], s[1]*s[3]) {
		t.Fatalf("open size %v", im.Bounds())
	}
	cells, im := GridLift(g, gr, g.Per, 1)
	if im.Bounds().Size() != image.Pt(cells.Dx()*g.Cell.X, cells.Dy()*g.Cell.Y) {
		t.Fatalf("lift size %v for %v", im.Bounds(), cells)
	}
}
