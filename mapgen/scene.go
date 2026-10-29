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
	Fill    bool // the overview may scale in fractions, to fill the window (the star map)
	Targets []Pt // buildings and wonders, nearest the square first
	// Districts are the building lines that have something standing, in the
	// order of the legend.
	Districts []District
}

// Moment is when and how a picture is drawn: the animation frame, the frame
// of the world's clock the rare visitor is at (negative: none is out, and the
// clock runs on when the picture is held still), whether what is new since the
// last visit is lit, and the building line whose district is in focus ("": the
// whole town).
type Moment struct {
	Anim, Visit int
	Fresh       bool
	District    string
	Close       bool // finer than the overview: the star map shows its territory on plain dark space
}

// Draw draws the whole world at animation frame anim.
func (gen *Generator) Draw(m *mapmodel.Model, anim int) *Scene {
	return gen.DrawAt(m, Moment{Anim: anim, Visit: anim})
}

// DrawAt draws the whole world at a moment.
func (gen *Generator) DrawAt(m *mapmodel.Model, at Moment) *Scene {
	f := gen.frame(m, at.Anim)
	f.visit, f.fresh, f.district, f.close = at.Visit, at.Fresh, at.District, at.Close
	f.draw()
	c := f.plan.Centre
	slices.SortStableFunc(f.targets, func(a, b Pt) int {
		return cmp.Compare(Rect{a.X, a.Y, 1, 1}.dist(c), Rect{b.X, b.Y, 1, 1}.dist(c))
	})
	bounds := f.bounds()
	if len(f.sites) > 0 && f.st.Scale != Stars {
		bounds = f.edge() // the civs stand at the edge of the picture
	}
	if f.st.Scale == Stars {
		c = f.plan.territoryCentre() // the camera starts on the town, not the core
	}
	return &Scene{Grid: f.g, Centre: c, Bounds: bounds, Fill: f.st.Scale == Stars, Targets: f.targets, Districts: f.districts()}
}

// starFrame is how much of the galaxy's reach the overview frames, in percent:
// the faint rim is dark, and the territory is held in whatever it covers.
const starFrame = 84

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
	if f.st.Scale == Stars { // the body of the galaxy is the map, and all the territory
		r := f.plan.Galaxy.reach()
		cx, cy := worldW/2, worldH/2
		for _, p := range r {
			grow(Rect{cx + (p.X-cx)*starFrame/100, cy + (p.Y-cy)*starFrame/100, 1, 1})
		}
		for _, list := range [][]Sector{f.plan.Sectors, f.plan.Rivals} {
			for _, s := range list {
				for _, c := range s.Cells {
					grow(Rect{c.X, c.Y, 1, 1})
				}
			}
		}
		if f.m.Harbinger != nil {
			for _, c := range f.plan.Rift {
				grow(Rect{c.X, c.Y, 1, 1})
			}
		}
	}
	const margin = 3
	return Rect{x0 - margin, y0 - margin, x1 - x0 + 1 + 2*margin, y1 - y0 + 1 + 2*margin}
}
