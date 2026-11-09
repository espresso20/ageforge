package mapgen

import "github.com/espresso20/ageforge/mapmodel"

// The ring scales (orbit and the solar system) see the world from straight
// above: a body in the middle, rings round it, and parcels strung along the
// rings the way buildings stand along streets. A cell is twice as tall as it
// is wide, so a circle is twice as wide as tall. One spoke rises from the
// body through every ring; it is the root of the road tree, and the two
// halves of each ring grow out of it.

// ringSpec describes a ring scale.
type ringSpec struct {
	cx, cy     int
	hub        int   // rows from the centre to the edge of the body in the middle
	rings      []int // each ring's radius in rows, nearest first
	lotW, lotH int   // a parcel on a ring
	step       int   // cells of ring from one parcel to the next
	spread     [2]int
}

// halfRing is one side of a ring of radius ry rows, from the top (not
// included) round to the bottom (not included), one step at a time so every
// cell touches the one before: along a row, then up or down to the next.
func halfRing(cx, cy, ry, side int) []Pt {
	rx := (ry * 237) / 100
	if rx > cx-4 {
		rx = cx - 4
	}
	var upper []Pt
	height := ry
	for dx := 1; dx <= rx; dx++ {
		next := isqrt(ry * ry * (rx*rx - dx*dx) / (rx * rx))
		for h := height; h >= next; h-- {
			upper = append(upper, Pt{X: cx + side*dx, Y: cy - h})
		}
		height = next
	}
	cells := upper
	for i := len(upper) - 2; i >= 0; i-- { // the lower half mirrors the upper
		cells = append(cells, Pt{X: upper[i].X, Y: 2*cy - upper[i].Y})
	}
	return cells
}

// ringOf is which ring a road index belongs to: roads 2 and 3 are ring 0.
func ringOf(road int) int { return (road - 2) / 2 }

// newRingPlan lays out the hub, the spoke and the rings, and no parcels yet.
func newRingPlan(seed int64, sp ringSpec, scale Scale) *planner {
	gw, gh := worldW, worldH
	if scale == System {
		gw, gh = SystemW, SystemH
	}
	b := newPlannerSized(seed, gw, gh)
	p := b.plan
	p.Scale, p.Centre = scale, Pt{X: sp.cx, Y: sp.cy}
	hub := Pt{X: sp.cx, Y: sp.cy - sp.hub - 1}
	p.Roads = []Road{{Cells: []Pt{hub}, Parent: -1}}
	spoke := Road{Parent: 0, Tier: 1}
	for y := hub.Y - 1; y >= max(0, sp.cy-sp.rings[len(sp.rings)-1]); y-- {
		spoke.Cells = append(spoke.Cells, Pt{X: sp.cx, Y: y})
	}
	p.Roads = append(p.Roads, spoke)
	b.claimRoad(hub, 0)
	for _, c := range spoke.Cells {
		b.claimRoad(c, 1)
	}
	for _, ry := range sp.rings {
		crossing := hub.Y - 1 - (sp.cy - ry) // the spoke cell on the ring's top
		for _, side := range []int{1, -1} {
			cells := halfRing(sp.cx, sp.cy, ry, side)
			if side == 1 {
				cells = append(cells, Pt{X: sp.cx, Y: sp.cy + ry})
			}
			road := Road{Cells: cells, Parent: 1, From: crossing}
			for _, c := range road.Cells {
				b.claimRoad(c, len(p.Roads))
			}
			p.Roads = append(p.Roads, road)
		}
	}
	p.Plots = make([]Lot, 22)
	return b
}

// room says whether a parcel fits on ring n: free of other parcels and of
// every road but its own ring.
func (b *planner) room(r Rect, ring int) bool {
	ok := true
	r.cells(func(x, y int) {
		switch {
		case !b.in(x, y) || b.claim[b.at(x, y)] == lotCell:
			ok = false
		case b.claim[b.at(x, y)] == roadCell && (b.roadOf[b.at(x, y)] < 2 || ringOf(b.roadOf[b.at(x, y)]) != ring):
			ok = false
		}
	})
	return ok
}

// cutRingLots sets parcels along every ring, packing each in as close as the
// last allows, and then hands them to the building lines.
func (b *planner) cutRingLots(seed int64, cat *mapmodel.Catalog, sp ringSpec) {
	p := b.plan
	for road := 2; road < len(p.Roads); road++ {
		next := 3
		for i, c := range p.Roads[road].Cells {
			if i < next {
				continue
			}
			rect := Rect{c.X - sp.lotW/2, c.Y - sp.lotH/2, sp.lotW, sp.lotH}
			if !b.room(rect, ringOf(road)) {
				next = i + 1
				continue
			}
			b.take(rect)
			p.Lots = append(p.Lots, Lot{Rect: rect, Door: c, Road: road, At: i})
			next = i + sp.step
		}
	}
	p.assign(seed, cat, sp.spread[0], sp.spread[1])
}

// plotBesideRing keeps a wonder's ground just outside a ring, at its east
// end (or west end if east is off-world), for the wonder of an age.
func (b *planner) plotBesideRing(age, ring int) {
	p := b.plan
	for r := ring; r >= 0; r-- {
		for _, road := range []int{2 + 2*r, 2 + 2*r + 1} {
			cells := p.Roads[road].Cells
			for i, c := range cells {
				for _, rect := range []Rect{
					{c.X + 1, c.Y - plotH/2, plotW, plotH},
					{c.X - plotW, c.Y - plotH/2, plotW, plotH},
					{c.X - plotW/2, c.Y + 1, plotW, plotH},
					{c.X - plotW/2, c.Y - plotH, plotW, plotH},
				} {
					if b.free2(rect) {
						b.take(rect)
						p.Plots[age] = Lot{Rect: rect, Door: c, Road: road, At: i}
						return
					}
				}
			}
		}
	}
}

// free2 says whether every cell of r is in the world and unclaimed.
func (b *planner) free2(r Rect) bool {
	ok := true
	r.cells(func(x, y int) {
		if !b.in(x, y) || b.claim[b.at(x, y)] != freeCell {
			ok = false
		}
	})
	return ok
}
