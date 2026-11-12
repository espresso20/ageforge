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
	deep   deepLayer        // the dark ground of the star map's close views, kept between frames
	wide   wideLayer        // the star map's wide views, kept between frames
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
	clock    int // the world's clock when the picture is held still, else 0
	sites    []site
	fresh    bool   // light what is new since the last visit
	district string // the building line in focus, if any
	close    bool   // a close look: the star map leaves its arms off the sky
	planet   int    // the planet the close look is on (1 Mercury .. 8 Neptune), 0 the home world
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
	pp      *port // the space port, once laid out
}

// frame sets up a picture of the model: its style, plan and placements.
func (gen *Generator) frame(m *mapmodel.Model, anim int) *frame {
	gen.sync(m)
	st := styleFor(func(k string) int { return m.Catalog.AgeIdx[k] }, m.AgeIdx)
	gw, gh := worldW, worldH
	if st.Scale == System {
		gw, gh = SystemW, SystemH
	}
	f := &frame{m: m, st: st, g: newGrid(gw, gh), anim: anim, visit: anim,
		night: int(m.Clock.Night * 256), roadAt: make([]int8, gw*gh), busy: make([]bool, gw*gh)}
	f.plan, f.fields = gen.plans[st.Scale], gen.fields
	f.g.Wall, f.g.Tint = st.Wall, st.Night.Tint
	f.g.Space = st.Scale >= Orbit
	f.placed = f.plan.place(m)
	f.filler = f.fillerLots()
	f.layRoads()
	return f
}

func (f *frame) at(x, y int) int { return y*f.g.W + x }

// draw paints every layer, bottom to top.
func (f *frame) draw() {
	if f.calm() || f.atPort() {
		f.paintVoid()
		f.paintStarField()
		if f.atPort() {
			f.paintSpacePort()
		} else {
			f.paintEarthSystem()
		}
		f.paintAround()
		f.paintFocus()
		f.paintVisitor()
		return
	}
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
