package mapgen

import (
	"cmp"
	"slices"

	"github.com/espresso20/ageforge/mapmodel"
)

// The star map is the last scale: every star of the sky from the hash, the
// systems the people have among them joined by routes, and the region their
// buildings reach, a glow that grows with them. Parcels are marks round the
// systems; a wonder is a landmark beside one.

// starNodes is how many systems the map holds besides home.
const starNodes = 27

// starSlots are the places round a system, nearest first, where its marks go.
var starSlots = [][2]int{
	{2, 0}, {-2, 0}, {4, 0}, {-4, 0}, {0, -1}, {0, 1}, {2, -1}, {-2, -1}, {2, 1}, {-2, 1},
	{4, -1}, {-4, -1}, {4, 1}, {-4, 1}, {6, 0}, {-6, 0},
}

// NewStarPlan lays out the systems, the routes that join them and the marks
// round each, and the plots of the wonders at the three nearest systems.
func NewStarPlan(seed int64, cat *mapmodel.Catalog) *Plan {
	b := newPlanner(seed)
	p := b.plan
	p.Scale, p.Centre = Stars, Pt{worldW / 2, worldH / 2}
	p.Roads = []Road{{Cells: []Pt{p.Centre}, Parent: -1}}
	b.claimRoad(p.Centre, 0)
	p.Nodes = append([]Pt{p.Centre}, bestCandidates(seed, p.Centre, starNodes)...)
	route := b.routes()
	b.plotsBeside(route)
	b.cutStarLots(route)
	p.assign(seed, cat, 14, 5)
	p.Plots = padPlots(p.Plots)
	return p
}

func padPlots(plots []Lot) []Lot {
	for len(plots) < 22 {
		plots = append(plots, Lot{})
	}
	return plots
}

// bestCandidates picks n points in an ellipse round c, each of a dozen
// hashed spots the one furthest from the points already picked.
func bestCandidates(seed int64, c Pt, n int) []Pt {
	const rx, ry = 62, 22
	var out []Pt
	for i := 0; i < n; i++ {
		best, bestGap := c, -1
		for try := 0; try < 12; try++ {
			q := Pt{c.X + roll(2*rx+1, seed, 90, int64(i), int64(try)) - rx, c.Y + roll(2*ry+1, seed, 91, int64(i), int64(try)) - ry}
			if rim(q.X, q.Y, c.X, c.Y, rx, ry) > 1000 {
				continue
			}
			gap := Rect{q.X, q.Y, 1, 1}.dist(c)
			for _, o := range out {
				gap = min(gap, Rect{q.X, q.Y, 1, 1}.dist(o))
			}
			if gap > bestGap {
				best, bestGap = q, gap
			}
		}
		out = append(out, best)
	}
	return out
}

// routes joins every system to the nearest one already joined, hub first, and
// returns each system's road: the one that ends at it.
func (b *planner) routes() []int {
	p := b.plan
	route := make([]int, len(p.Nodes))
	joined := []int{0}
	left := map[int]bool{}
	for n := 1; n < len(p.Nodes); n++ {
		left[n] = true
	}
	for len(joined) < len(p.Nodes) {
		from, to, gap := 0, 0, 1<<60
		for n := 1; n < len(p.Nodes); n++ {
			if !left[n] {
				continue
			}
			for _, j := range joined {
				if d := (Rect{p.Nodes[n].X, p.Nodes[n].Y, 1, 1}).dist(p.Nodes[j]); d < gap {
					from, to, gap = j, n, d
				}
			}
		}
		parent := route[from]
		road := Road{Cells: staircase(p.Nodes[from], p.Nodes[to]), Parent: parent, From: len(p.Roads[parent].Cells) - 1}
		route[to] = len(p.Roads)
		for _, c := range road.Cells {
			b.claimRoad(c, len(p.Roads))
		}
		p.Roads = append(p.Roads, road)
		joined = append(joined, to)
		delete(left, to)
	}
	return route
}

// staircase is the path from a to b, not counting a, one step at a time along
// whichever axis has further to go (a row counts double, being the taller).
func staircase(a, b Pt) []Pt {
	var cells []Pt
	for a != b {
		if abs(b.X-a.X) >= 2*abs(b.Y-a.Y) {
			a.X += cmp.Compare(b.X, a.X)
		} else {
			a.Y += cmp.Compare(b.Y, a.Y)
		}
		cells = append(cells, a)
	}
	return cells
}

// plotsBeside keeps ground for the wonders of the last three ages beside the
// systems nearest home, the first system with room for each.
func (b *planner) plotsBeside(route []int) {
	p := b.plan
	near := make([]int, 0, len(p.Nodes)-1)
	for n := 1; n < len(p.Nodes); n++ {
		near = append(near, n)
	}
	slices.SortStableFunc(near, func(x, y int) int {
		return cmp.Compare(Rect{p.Nodes[x].X, p.Nodes[x].Y, 1, 1}.dist(p.Centre), Rect{p.Nodes[y].X, p.Nodes[y].Y, 1, 1}.dist(p.Centre))
	})
	p.Plots = make([]Lot, 22)
	for _, age := range []int{19, 20, 21} {
		for len(near) > 0 && p.Plots[age].Rect.W == 0 {
			n := near[0]
			near = near[1:]
			node := p.Nodes[n]
			for _, off := range [][2]int{{1, -2}, {-plotW, -2}, {1, 0}, {-plotW, 0}, {1, -4}, {-plotW, -4}} {
				rect := Rect{node.X + off[0], node.Y + off[1], plotW, plotH}
				if b.free2(rect) {
					b.take(rect)
					p.Plots[age] = Lot{Rect: rect, Door: node, Road: route[n], At: len(p.Roads[route[n]].Cells) - 1}
					break
				}
			}
		}
	}
}

// cutStarLots sets a mark at every free slot round every system.
func (b *planner) cutStarLots(route []int) {
	p := b.plan
	for n, node := range p.Nodes {
		road := route[n]
		for _, s := range starSlots {
			rect := Rect{node.X + s[0], node.Y + s[1], 1, 1}
			if b.free2(rect) {
				b.take(rect)
				p.Lots = append(p.Lots, Lot{Rect: rect, Door: node, Road: road, At: len(p.Roads[road].Cells) - 1})
			}
		}
	}
}

// reach is the radius, in columns, of the region the people's buildings
// reach: out to the furthest parcel in use, and a little more.
func (f *frame) reach() int {
	c, r := f.plan.Centre, 14
	use := func(rect Rect) {
		r = max(r, isqrt(rect.dist(c))+8)
	}
	for _, pl := range f.placed {
		use(f.plan.Lots[pl.lot].Rect)
	}
	for _, lot := range f.filler {
		use(f.plan.Lots[lot].Rect)
	}
	return r
}

// paintStarMap draws the glow of the people's reach; the systems themselves
// are drawn over the routes (paintBodies).
func (f *frame) paintStarMap() {
	c, r := f.plan.Centre, f.reach()
	rx, ry := r, max(1, r/2)
	for y := c.Y - ry; y <= c.Y+ry; y++ {
		for x := c.X - rx; x <= c.X+rx; x++ {
			if r := rim(x, y, c.X, c.Y, rx, ry); inWorld(x, y) && r <= 1000 {
				f.g.Cells[at(x, y)].Bg = f.st.Reach.fade(64 + (1000-r)*192/1000) // brightest at home
			}
		}
	}
}
