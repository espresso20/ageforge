package mapgen

import "github.com/espresso20/ageforge/mapmodel"

// Pt is a cell of the world.
type Pt struct{ X, Y int }

// Rect is a block of cells.
type Rect struct{ X, Y, W, H int }

func (r Rect) cells(f func(x, y int)) {
	for y := r.Y; y < r.Y+r.H; y++ {
		for x := r.X; x < r.X+r.W; x++ {
			f(x, y)
		}
	}
}

// Road is a line of road cells growing out of another road.
type Road struct {
	Cells  []Pt
	Parent int // the road it leaves, -1 for the first
	From   int // the parent cell it leaves from
	Tier   int // 0 trunk, 1 street, 2 lane
}

// Lot is a parcel: a rectangle that opens on a road cell.
type Lot struct {
	Rect Rect
	Door Pt  // the road cell it opens on
	Road int // that road
	At   int // the door's place in the road's cells
}

// Plan is the settlement's skeleton: its roads, its parcels, the plots kept
// for wonders and which building line owns which parcel. It is a pure
// function of the seed, so it never changes while a game goes on; what is
// built only decides how much of it shows.
type Plan struct {
	Orbit  bool
	Centre Pt
	Roads  []Road
	Lots   []Lot
	// Plots are the wonder plots, indexed by the wonder's age.
	Plots []Lot
	// Owner is the building line of each lot, an index into
	// mapmodel.LineageOrder; Order lists each line's lots, nearest the
	// middle first. Hearts are where each line's district is rooted.
	Owner  []int
	Order  [][]int
	Hearts []Pt
	// Bridge marks road cells that cross a river.
	Bridge map[Pt]bool
}

// Parcel sizes. A cell is about twice as tall as wide, so these read as
// roughly square on screen.
const (
	lotW, lotH   = 6, 4
	plotW, plotH = 10, 5
)

// groundAges is how many ages the ground scenes cover; later ages are in
// orbit and keep their wonders there.
const groundAges = 17

const (
	freeCell = iota
	roadCell
	lotCell
)

type planner struct {
	seed   int64
	f      *Fields
	plan   *Plan
	claim  []uint8
	roadOf []int
	arms   [4]int // road index of the east, west, south and north trunk arms, -1 if none
}

func newPlanner(seed int64) *planner {
	return &planner{seed: seed, plan: &Plan{Bridge: map[Pt]bool{}},
		claim: make([]uint8, worldW*worldH), roadOf: make([]int, worldW*worldH)}
}

// NewGroundPlan lays out a settlement on the land of a seed.
func NewGroundPlan(seed int64, f *Fields, cat *mapmodel.Catalog) *Plan {
	b := newPlanner(seed)
	b.f = f
	b.plan.Centre = b.findCentre()
	b.layArms()
	b.reservePlots()
	b.layStreets()
	b.cutLots()
	b.plan.assign(seed, cat, 12, 5)
	return b.plan
}

// findCentre picks the flattest dry spot near water that is near the middle
// of the world: the settlement's square.
func (b *planner) findCentre() Pt {
	best, bestScore := Pt{worldW / 2, worldH / 2}, 1<<62
	for y := worldH/2 - 12; y <= worldH/2+12; y++ {
		for x := worldW/2 - 36; x <= worldW/2+36; x++ {
			if !b.buildable(x, y) {
				continue
			}
			score := b.surroundings(x, y) + abs(x-worldW/2)*3 + abs(y-worldH/2)*6 + max(0, int(b.f.WaterDist[at(x, y)])-14)*30
			if score < bestScore {
				best, bestScore = Pt{x, y}, score
			}
		}
	}
	return best
}

// surroundings scores the ground a settlement would cover around (x, y):
// steep and wet cells cost.
func (b *planner) surroundings(x, y int) int {
	cost := 0
	for yy := y - 8; yy <= y+8; yy += 2 {
		for xx := x - 24; xx <= x+24; xx += 3 {
			switch {
			case !inWorld(xx, yy) || b.f.Biome[at(xx, yy)].Water():
				cost += 90
			case !b.buildable(xx, yy):
				cost += 40
			default:
				cost += b.f.Slope(xx, yy) / 4
			}
		}
	}
	return cost
}

func (b *planner) buildable(x, y int) bool {
	if !inWorld(x, y) {
		return false
	}
	switch b.f.Biome[at(x, y)] {
	case Sand, Scrub, Grass, Forest, Hills:
		return b.f.Slope(x, y) <= 45
	}
	return false
}

// passable is where a road can go: buildable ground, or a river it bridges.
func (b *planner) passable(x, y int) bool {
	return inWorld(x, y) && b.claim[at(x, y)] != lotCell && (b.buildable(x, y) || b.f.Biome[at(x, y)] == River)
}
