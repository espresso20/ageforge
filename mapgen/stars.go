package mapgen

import (
	"cmp"
	"slices"

	"github.com/espresso20/ageforge/mapmodel"
)

// The star map is the last scale: a seeded spiral galaxy (galaxy.go), the
// systems the people have along its home arm joined by routes, and the region
// their buildings reach. Parcels are marks round the
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
	p.Galaxy = NewGalaxy(seed)
	p.Nodes = append([]Pt{p.Centre}, p.Galaxy.homeNodes(seed, starNodes)...)
	route := b.routes()
	p.NodeRoad = route
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

// chain is the cells of a road from the home system out: its parents' cells up
// to where it leaves them, then its own.
func (p *Plan) chain(road int) []Pt {
	r := p.Roads[road]
	if r.Parent < 0 {
		return r.Cells
	}
	return append(append([]Pt(nil), p.chain(r.Parent)[:r.From+1]...), r.Cells...)
}
