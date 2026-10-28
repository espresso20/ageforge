package mapgen

import (
	"cmp"
	"slices"

	"github.com/espresso20/ageforge/mapmodel"
)

// Scene is a frame drawn over the whole world, with what a view needs to
// choose its own window: where the square is, how far the town reaches and
// the places a cursor can visit.
type Scene struct {
	Grid    *Grid
	Centre  Pt
	Bounds  Rect // the town: its parcels, plots and laid roads
	Targets []Pt // buildings and wonders, nearest the square first
}

// Draw draws the whole world at animation frame anim.
func (gen *Generator) Draw(m *mapmodel.Model, anim int) *Scene {
	return gen.DrawAt(m, anim, anim)
}

// DrawAt draws the whole world at animation frame anim, with the rare visitor
// at frame visit of the world's clock (negative: none is out), which runs on
// when the picture is held still.
func (gen *Generator) DrawAt(m *mapmodel.Model, anim, visit int) *Scene {
	f := gen.frame(m, anim)
	f.visit = visit
	f.draw()
	c := f.plan.Centre
	slices.SortStableFunc(f.targets, func(a, b Pt) int {
		return cmp.Compare(Rect{a.X, a.Y, 1, 1}.dist(c), Rect{b.X, b.Y, 1, 1}.dist(c))
	})
	return &Scene{Grid: f.g, Centre: c, Bounds: f.bounds(), Targets: f.targets}
}

// bounds is the box round everything the town has laid or built, with a
// margin.
func (f *frame) bounds() Rect {
	x0, y0, x1, y1 := f.plan.Centre.X, f.plan.Centre.Y, f.plan.Centre.X, f.plan.Centre.Y
	grow := func(r Rect) {
		x0, y0, x1, y1 = min(x0, r.X), min(y0, r.Y), max(x1, r.X+r.W-1), max(y1, r.Y+r.H-1)
	}
	for i, r := range f.plan.Roads {
		for _, c := range r.Cells[:f.shown[i]] {
			grow(Rect{c.X, c.Y, 1, 1})
		}
	}
	for _, pl := range f.placed {
		grow(f.plan.Lots[pl.lot].Rect)
	}
	for _, lot := range f.filler {
		grow(f.plan.Lots[lot].Rect)
	}
	for _, ws := range f.wonders() {
		grow(ws.plot.Rect)
	}
	if f.st.Pads > 0 && f.plan.Port.Rect.W > 0 {
		grow(f.plan.Port.Rect)
	}
	const margin = 3
	return Rect{x0 - margin, y0 - margin, x1 - x0 + 1 + 2*margin, y1 - y0 + 1 + 2*margin}
}
