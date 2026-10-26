package mapgen

import "github.com/espresso20/ageforge/mapmodel"

// Generator draws maps. It keeps the land and the plan of the last seed, which
// only change with the seed, so drawing a frame is the cheap part. The zero
// value is ready to use; it is not safe for concurrent use.
type Generator struct {
	seed   int64
	cat    *mapmodel.Catalog
	fields *Fields
	ground *Plan
	orbit  *Plan
}

// frame is one picture being drawn, layer by layer, on a grid the size of
// the whole world.
type frame struct {
	m      *mapmodel.Model
	st     *AgeStyle
	plan   *Plan
	fields *Fields
	g      *Grid
	anim   int
	night  int // 0 day .. 256 full night
	placed []placed
	// shown is how many cells of each road are laid; roadAt is the tier+1
	// of each laid road cell of the world.
	shown  []int
	roadAt []int8
	// busy marks cells something is built or laid on; glow lists the cells
	// that light up at night.
	busy []bool
	glow []Pt
	// vents are the chimneys that smoke.
	vents []Pt
	// lights are planet cities that shine whenever it is night.
	lights []Pt
}

// Render draws the model as a w by h grid at animation frame anim. The
// picture is a window onto a world of fixed size, centred on the settlement.
func (gen *Generator) Render(m *mapmodel.Model, w, h, anim int) *Grid {
	f := gen.frame(m, anim)
	f.draw()
	x0, y0 := f.camera(w, h)
	return f.g.window(x0, y0, w, h)
}

// camera is where the window's top-left corner goes: centred on the
// settlement, nudged a little (the nearest nudge that does best) to take in
// the most of what is built: the wonder being built first, then the rest.
func (f *frame) camera(w, h int) (x0, y0 int) {
	c := f.plan.Centre
	best := -1
	nudges := [][2]int{{0, 0}, {-8, 0}, {8, 0}, {0, -4}, {0, 4}, {-8, -4}, {8, -4}, {-8, 4}, {8, 4}, {-16, 0}, {16, 0}, {-16, -4}, {16, -4}, {-16, 4}, {16, 4}, {-24, 0}, {24, 0}}
	if f.plan.Orbit {
		nudges = nudges[:1]
	}
	for _, n := range nudges {
		x, y := clamp(c.X+n[0]-w/2, 0, worldW-w), clamp(c.Y+n[1]-h/2, 0, worldH-h)
		in := 0
		inside := func(r Rect) bool { return r.X >= x && r.X+r.W <= x+w && r.Y >= y && r.Y+r.H <= y+h }
		for _, pl := range f.placed {
			if inside(f.plan.Lots[pl.lot].Rect) {
				in++
			}
		}
		for _, ws := range f.wonders() {
			switch {
			case inside(ws.plot.Rect) && ws.w.Current:
				in += 1000 // the wonder being built comes first
			case inside(ws.plot.Rect):
				in += 4
			}
		}
		if in > best {
			x0, y0, best = x, y, in
		}
	}
	return x0, y0
}

// frame sets up a picture of the model: its style, plan and placements.
func (gen *Generator) frame(m *mapmodel.Model, anim int) *frame {
	gen.sync(m)
	st := styleFor(func(k string) int { return m.Catalog.AgeIdx[k] }, m.AgeIdx)
	f := &frame{m: m, st: st, g: newGrid(worldW, worldH), anim: anim,
		night: int(m.Clock.Night * 256), roadAt: make([]int8, worldW*worldH), busy: make([]bool, worldW*worldH)}
	if st.Scene == Orbit {
		f.plan = gen.orbit
	} else {
		f.plan, f.fields = gen.ground, gen.fields
	}
	f.placed = f.plan.place(m)
	f.layRoads()
	return f
}

// draw paints every layer, bottom to top.
func (f *frame) draw() {
	f.paintGround()
	f.paintTown()
	f.paintDetail()
	f.paintSmoke()
	f.paintAgents()
	f.paintLight()
}

// sync makes the land and plans match the model's seed and catalogue.
func (gen *Generator) sync(m *mapmodel.Model) {
	if gen.fields != nil && gen.seed == m.Seed && gen.cat == m.Catalog {
		return
	}
	gen.seed, gen.cat = m.Seed, m.Catalog
	gen.fields = newFields(m.Seed)
	gen.ground = NewGroundPlan(m.Seed, gen.fields, m.Catalog)
	gen.orbit = NewOrbitPlan(m.Seed, m.Catalog)
}
