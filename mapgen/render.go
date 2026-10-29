package mapgen

import "github.com/espresso20/ageforge/mapmodel"

// Generator draws maps. It keeps the land and the plan of the last seed, which
// only change with the seed, so drawing a frame is the cheap part. The zero
// value is ready to use; it is not safe for concurrent use.
type Generator struct {
	seed   int64
	cat    *mapmodel.Catalog
	fields *Fields
	plans  [numScales]*Plan // by scale; the ground and the city share one
}

// frame is one picture being drawn, layer by layer, on a grid the size of
// the whole world.
type frame struct {
	m        *mapmodel.Model
	st       *AgeStyle
	plan     *Plan
	fields   *Fields
	g        *Grid
	anim     int
	visit    int // the frame the rare visitor is drawn at; negative: none
	sites    []site
	fresh    bool   // light what is new since the last visit
	district string // the building line in focus, if any
	close    bool   // a close look: the star map leaves its arms off the sky
	box      Rect   // the edge of the picture, once worked out
	night    int    // 0 day .. 256 full night
	placed   []placed
	filler   []int // lots given to ordinary houses
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
	// lights are planet cities that shine whenever it is night; targets are
	// the places the cursor visits.
	lights  []Pt
	targets []Pt
}

// Render draws the model as a w by h grid at animation frame anim. The
// picture is a window onto a world of fixed size, centred on the settlement.
func (gen *Generator) Render(m *mapmodel.Model, w, h, anim int) *Grid {
	f := gen.frame(m, anim)
	f.draw()
	x0, y0 := f.camera(w, h)
	return f.g.Window(x0, y0, w, h)
}

// camera is where the window's top-left corner goes: centred on the
// settlement, nudged a little (the nearest nudge that does best) to take in
// the most of what is built: the wonder being built first, then the rest.
func (f *frame) camera(w, h int) (x0, y0 int) {
	nudges := [][2]int{{0, 0}, {-8, 0}, {8, 0}, {0, -4}, {0, 4}, {-8, -4}, {8, -4}, {-8, 4}, {8, 4}, {-16, 0}, {16, 0}, {-16, -4}, {16, -4}, {-16, 4}, {16, 4}, {-24, 0}, {24, 0}}
	if f.st.Scale > City {
		nudges = nudges[:1]
	}
	best := -1
	for _, n := range nudges {
		x := clamp(f.plan.Centre.X+n[0]-w/2, 0, worldW-w)
		y := clamp(f.plan.Centre.Y+n[1]-h/2, 0, worldH-h)
		if score := f.inView(Rect{x, y, w, h}); score > best {
			x0, y0, best = x, y, score
		}
	}
	return x0, y0
}

// inView scores a window by what it takes in whole: one for each building,
// four for a wonder, a thousand for the wonder being built.
func (f *frame) inView(window Rect) int {
	inside := func(r Rect) bool {
		return r.X >= window.X && r.X+r.W <= window.X+window.W && r.Y >= window.Y && r.Y+r.H <= window.Y+window.H
	}
	score := 0
	for _, pl := range f.placed {
		if inside(f.plan.Lots[pl.lot].Rect) {
			score++
		}
	}
	for _, ws := range f.wonders() {
		switch {
		case inside(ws.plot.Rect) && ws.w.Current:
			score += 1000
		case inside(ws.plot.Rect):
			score += 4
		}
	}
	return score
}

// frame sets up a picture of the model: its style, plan and placements.
func (gen *Generator) frame(m *mapmodel.Model, anim int) *frame {
	gen.sync(m)
	st := styleFor(func(k string) int { return m.Catalog.AgeIdx[k] }, m.AgeIdx)
	f := &frame{m: m, st: st, g: newGrid(worldW, worldH), anim: anim, visit: anim,
		night: int(m.Clock.Night * 256), roadAt: make([]int8, worldW*worldH), busy: make([]bool, worldW*worldH)}
	f.plan, f.fields = gen.plans[st.Scale], gen.fields
	f.g.Wall, f.g.Tint = st.Wall, st.Night.Tint
	f.placed = f.plan.place(m)
	f.filler = f.fillerLots()
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
	f.paintBodies()
	f.paintAround()
	f.paintFocus()
	f.paintLight()
	f.paintVisitor()
}

// sync makes the land and plans match the model's seed and catalogue.
func (gen *Generator) sync(m *mapmodel.Model) {
	if gen.fields != nil && gen.seed == m.Seed && gen.cat == m.Catalog {
		return
	}
	gen.seed, gen.cat = m.Seed, m.Catalog
	gen.fields = newFields(m.Seed)
	ground := NewGroundPlan(m.Seed, gen.fields, m.Catalog)
	gen.plans = [numScales]*Plan{Ground: ground, City: ground,
		Orbit: NewOrbitPlan(m.Seed, m.Catalog), System: NewSystemPlan(m.Seed, m.Catalog), Stars: NewStarPlan(m.Seed, m.Catalog)}
}
