package mapgen

import (
	"cmp"
	"slices"

	"github.com/espresso20/ageforge/mapmodel"
)

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
	// mapmodel.LineageOrder; Order lists each line's lots, nearest its
	// heart first. Hearts are where each line's district is rooted.
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
	plotW, plotH = 11, 6
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
	b.plan.assign(seed, cat)
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

var (
	east  = [2]int{1, 0}
	west  = [2]int{-1, 0}
	south = [2]int{0, 1}
	north = [2]int{0, -1}
)

// layArms puts down the plaza and the four trunk arms that leave it.
func (b *planner) layArms() {
	p := b.plan
	p.Roads = append(p.Roads, Road{Cells: []Pt{p.Centre}, Parent: -1})
	b.claimRoad(p.Centre, 0)
	for i, arm := range [][2]int{east, west, south, north} {
		reach := 56 + roll(30, b.seed, 20, int64(i))
		if arm[0] == 0 {
			reach = 30 + roll(14, b.seed, 20, int64(i))
		}
		b.arms[i] = b.grow(0, 0, arm, reach, 0)
	}
}

// layStreets grows streets off the trunk arms and lanes off the streets.
func (b *planner) layStreets() {
	for _, arm := range b.arms {
		if arm >= 0 {
			b.branch(arm, 1)
		}
	}
	streets := len(b.plan.Roads)
	for road := 1; road < streets; road++ {
		if b.plan.Roads[road].Tier == 1 {
			b.branch(road, 2)
		}
	}
}

func (b *planner) claimRoad(c Pt, road int) {
	b.claim[at(c.X, c.Y)], b.roadOf[at(c.X, c.Y)] = roadCell, road
}

// branch spawns the children of a road: streets off a trunk, lanes off a
// street, on both sides at irregular gaps. A street across a trunk runs
// long; a lane stops where it meets the next street.
func (b *planner) branch(road, tier int) {
	r := b.plan.Roads[road]
	dir := b.heading(r)
	left, right := [2]int{dir[1], -dir[0]}, [2]int{-dir[1], dir[0]}
	gap, span := 13, 3
	if tier == 2 {
		gap, span = 9, 3
	}
	for pos := 4; pos < len(r.Cells); pos += gap + roll(span, b.seed, 44, int64(road), int64(pos)) {
		for s, side := range [][2]int{left, right} {
			if !chance(92, b.seed, 45, int64(road), int64(pos), int64(s)) {
				continue
			}
			length := 24 + roll(24, b.seed, 46, int64(road), int64(pos), int64(s))
			if side[0] != 0 {
				length *= 2 // a street running along the way the trunks run
			}
			b.grow(road, pos, side, length, tier)
		}
	}
}

// heading is the way a road runs, from its first cell to its last.
func (b *planner) heading(r Road) [2]int {
	first, last := r.Cells[0], r.Cells[len(r.Cells)-1]
	switch {
	case abs(last.X-first.X) >= abs(last.Y-first.Y) && last.X != first.X:
		return [2]int{cmp.Compare(last.X, first.X), 0}
	case last.Y != first.Y:
		return [2]int{0, cmp.Compare(last.Y, first.Y)}
	}
	return east
}

// grow lays one road out of cell `from` of road `parent` in direction dir and
// returns its index (-1 if it could not take a single cell). It goes
// straight, steps sideways round a snag, bridges a river, and stops at water,
// steep ground or another road.
func (b *planner) grow(parent, from int, dir [2]int, length, tier int) int {
	p := b.plan
	index := len(p.Roads)
	cur := p.Roads[parent].Cells[from]
	var cells []Pt
	for len(cells) < length {
		next := Pt{cur.X + dir[0], cur.Y + dir[1]}
		if inWorld(next.X, next.Y) && b.claim[at(next.X, next.Y)] == roadCell {
			break // it has met another road
		}
		if !b.canExtend(next, dir, index, parent) {
			next = b.nudge(cur, dir, index, parent)
			if next == cur {
				break
			}
		}
		if b.f.Biome[at(next.X, next.Y)] == River {
			p.Bridge[next] = true
		}
		b.claimRoad(next, index)
		cells = append(cells, next)
		cur = next
	}
	if len(cells) == 0 {
		return -1
	}
	p.Roads = append(p.Roads, Road{Cells: cells, Parent: parent, From: from, Tier: tier})
	return index
}

// canExtend says whether a road heading dir may take cell c: it must be
// passable, a river must have dry ground behind it, and no other road may run
// beside it.
func (b *planner) canExtend(c Pt, dir [2]int, road, parent int) bool {
	if !b.passable(c.X, c.Y) || b.claim[at(c.X, c.Y)] == roadCell {
		return false
	}
	if b.f.Biome[at(c.X, c.Y)] == River && !b.buildable(c.X+dir[0], c.Y+dir[1]) {
		return false
	}
	for _, side := range [][2]int{{dir[1], -dir[0]}, {-dir[1], dir[0]}} {
		s := Pt{c.X + side[0], c.Y + side[1]}
		if inWorld(s.X, s.Y) && b.claim[at(s.X, s.Y)] == roadCell && b.roadOf[at(s.X, s.Y)] != road && b.roadOf[at(s.X, s.Y)] != parent {
			return false
		}
	}
	return true
}

// nudge finds a cell one step to the side from which the road can carry on;
// it returns cur when there is none.
func (b *planner) nudge(cur Pt, dir [2]int, road, parent int) Pt {
	sides := [][2]int{{dir[1], -dir[0]}, {-dir[1], dir[0]}}
	if roll(2, b.seed, 47, int64(cur.X), int64(cur.Y)) == 1 {
		sides[0], sides[1] = sides[1], sides[0]
	}
	for _, side := range sides {
		c := Pt{cur.X + side[0], cur.Y + side[1]}
		ahead := Pt{c.X + dir[0], c.Y + dir[1]}
		if b.canExtend(c, dir, road, parent) && b.canExtend(ahead, dir, road, parent) {
			return c
		}
	}
	return cur
}

// reservePlots keeps ground for the wonders along the trunk arms, one plot
// per age the ground can show, the nearest plots for the earliest ages. Plot
// n sits on arm n%4, on the first side and place from its turn that fits.
func (b *planner) reservePlots() {
	p := b.plan
	for n := 0; n < groundAges; n++ {
		p.Plots = append(p.Plots, b.findPlot(b.arms[n%4], n/4, n))
	}
}

// findPlot looks along a trunk arm for room for the plot that is the
// step-th on that arm; none (a zero Lot) if the arm has no room left.
func (b *planner) findPlot(road, step, n int) Lot {
	if road < 0 {
		return Lot{}
	}
	r := b.plan.Roads[road]
	horizontal := b.heading(r)[1] == 0
	span := plotH + 2
	if horizontal {
		span = plotW + 2
	}
	for i := span/2 + 3 + step*span; i < len(r.Cells); i++ {
		for s := 0; s < 2; s++ {
			side := (n/4 + n%4 + s) % 2
			lot := Lot{Rect: plotBeside(r.Cells[i], horizontal, side), Door: r.Cells[i], Road: road, At: i}
			if b.free(lot.Rect) {
				b.take(lot.Rect)
				return lot
			}
		}
	}
	return Lot{}
}

// plotBeside is the wonder plot on one side of road cell c.
func plotBeside(c Pt, horizontal bool, side int) Rect {
	switch {
	case horizontal && side == 0:
		return Rect{c.X - plotW/2, c.Y - plotH, plotW, plotH}
	case horizontal:
		return Rect{c.X - plotW/2, c.Y + 1, plotW, plotH}
	case side == 0:
		return Rect{c.X - plotW, c.Y - plotH/2, plotW, plotH}
	}
	return Rect{c.X + 1, c.Y - plotH/2, plotW, plotH}
}

// free says whether every cell of r is dry, buildable and unclaimed.
func (b *planner) free(r Rect) bool {
	ok := true
	r.cells(func(x, y int) {
		if !b.buildable(x, y) || b.claim[at(x, y)] != freeCell {
			ok = false
		}
	})
	return ok
}

func (b *planner) take(r Rect) {
	r.cells(func(x, y int) { b.claim[at(x, y)] = lotCell })
}

// cutLots lines every road with parcels on both sides, packing each as close
// as the last one allows.
func (b *planner) cutLots() {
	p := b.plan
	for ri, r := range p.Roads {
		if len(r.Cells) < 2 {
			continue
		}
		horizontal := b.heading(r)[1] == 0
		step, offset := lotH, lotH/2
		if horizontal {
			step, offset = lotW, lotW/2
		}
		next := [2]int{offset, offset}
		for i, c := range r.Cells {
			if p.Bridge[c] {
				continue
			}
			for side := 0; side < 2; side++ {
				if i < next[side] {
					continue
				}
				lot := lotBeside(c, horizontal, side)
				if !b.free(lot) {
					next[side] = i + 1
					continue
				}
				b.take(lot)
				p.Lots = append(p.Lots, Lot{Rect: lot, Door: c, Road: ri, At: i})
				next[side] = i + step
			}
		}
	}
}

// lotBeside is the parcel on one side of road cell c.
func lotBeside(c Pt, horizontal bool, side int) Rect {
	if horizontal {
		if side == 0 {
			return Rect{c.X - lotW/2, c.Y - lotH, lotW, lotH}
		}
		return Rect{c.X - lotW/2, c.Y + 1, lotW, lotH}
	}
	if side == 0 {
		return Rect{c.X - lotW, c.Y - lotH/2, lotW, lotH}
	}
	return Rect{c.X + 1, c.Y - lotH/2, lotW, lotH}
}

// assign gives every lot to a building line. Each line has a heart, kept
// near the middle for the earliest lines and further out for the later ones,
// and a share of the lots in proportion to how many building types it has.
// Lots go to the nearest heart that still has room, so districts grow round
// their hearts, and a line's lots are listed nearest its heart first.
func (p *Plan) assign(seed int64, cat *mapmodel.Catalog) {
	lines := mapmodel.LineageOrder
	p.Hearts = p.placeHearts(seed, len(lines))
	quota := make([]int, len(lines))
	types := 0
	for _, l := range lines {
		types += len(cat.ByLineage[l])
	}
	for i, l := range lines {
		quota[i] = (len(p.Lots)*len(cat.ByLineage[l]) + types - 1) / types
	}
	p.Owner = make([]int, len(p.Lots))
	p.Order = make([][]int, len(lines))
	byNearest := make([]int, len(p.Lots))
	for i := range byNearest {
		byNearest[i] = i
	}
	nearest := func(lot int) int {
		best := 1 << 62
		for _, h := range p.Hearts {
			best = min(best, p.Lots[lot].Rect.dist(h))
		}
		return best
	}
	slices.SortStableFunc(byNearest, func(a, b int) int { return cmp.Compare(nearest(a), nearest(b)) })
	for _, lot := range byNearest {
		heartsByDistance := make([]int, len(lines))
		for i := range heartsByDistance {
			heartsByDistance[i] = i
		}
		slices.SortStableFunc(heartsByDistance, func(a, b int) int {
			return cmp.Compare(p.Lots[lot].Rect.dist(p.Hearts[a]), p.Lots[lot].Rect.dist(p.Hearts[b]))
		})
		for _, l := range heartsByDistance {
			if len(p.Order[l]) < quota[l] {
				p.Owner[lot] = l
				p.Order[l] = append(p.Order[l], lot)
				break
			}
		}
	}
	for l := range p.Order {
		slices.SortStableFunc(p.Order[l], func(a, b int) int {
			return cmp.Compare(p.Lots[a].Rect.dist(p.Hearts[l]), p.Lots[b].Rect.dist(p.Hearts[l]))
		})
	}
}

// dist is the squared distance from the middle of r to c, with rows counting
// double because cells are tall.
func (r Rect) dist(c Pt) int {
	dx, dy := r.X+r.W/2-c.X, r.Y+r.H/2-c.Y
	return dx*dx + 4*dy*dy
}

// placeHearts spreads one heart per line round the centre by best candidate:
// of a dozen hashed spots, take the one furthest from the hearts already set.
func (p *Plan) placeHearts(seed int64, n int) []Pt {
	hearts := make([]Pt, 0, n)
	for i := 0; i < n; i++ {
		rx, ry := 14+i*3, 5+i/2
		best, bestGap := p.Centre, -1
		for try := 0; try < 12; try++ {
			c := Pt{p.Centre.X + roll(2*rx+1, seed, 50, int64(i), int64(try)) - rx,
				p.Centre.Y + roll(2*ry+1, seed, 51, int64(i), int64(try)) - ry}
			gap := 1 << 40
			for _, h := range hearts {
				gap = min(gap, Rect{c.X, c.Y, 1, 1}.dist(h))
			}
			if gap > bestGap {
				best, bestGap = c, gap
			}
		}
		hearts = append(hearts, best)
	}
	return hearts
}
