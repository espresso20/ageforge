package mapgen

import (
	"cmp"
	"slices"

	"github.com/espresso20/ageforge/mapmodel"
)

// assign gives every lot to a building line. Each line has a heart, kept
// near the middle for the earliest lines and further out for the later ones
// (spreadX and spreadY cells further for the last), and a share of the lots in
// proportion to how many building types it has. Lots are dealt from the
// middle outward, each to the nearest heart whose line is not already ahead
// of its fair turn, so every line gets its share of the near lots and its
// district grows round its heart. A line's lots are listed nearest the
// middle first.
func (p *Plan) assign(seed int64, cat *mapmodel.Catalog, spreadX, spreadY int) {
	lines := mapmodel.LineageOrder
	p.Hearts = p.placeHearts(seed, len(lines), spreadX, spreadY)
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
	fromMiddle := make([]int, len(p.Lots))
	for i := range fromMiddle {
		fromMiddle[i] = i
	}
	slices.SortStableFunc(fromMiddle, func(a, b int) int {
		return cmp.Compare(p.Lots[a].Rect.dist(p.Centre), p.Lots[b].Rect.dist(p.Centre))
	})
	p.Near = fromMiddle
	for turn, lot := range fromMiddle {
		heartsByDistance := make([]int, len(lines))
		for i := range heartsByDistance {
			heartsByDistance[i] = i
		}
		slices.SortStableFunc(heartsByDistance, func(a, b int) int {
			return cmp.Compare(p.Lots[lot].Rect.dist(p.Hearts[a]), p.Lots[lot].Rect.dist(p.Hearts[b]))
		})
		deal := func(room func(l int) bool) bool {
			for _, l := range heartsByDistance {
				if room(l) {
					p.Owner[lot] = l
					p.Order[l] = append(p.Order[l], lot)
					return true
				}
			}
			return false
		}
		turnShare := turn/len(lines) + 2
		if !deal(func(l int) bool { return len(p.Order[l]) < min(quota[l], turnShare) }) {
			deal(func(l int) bool { return len(p.Order[l]) < quota[l] })
		}
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
func (p *Plan) placeHearts(seed int64, n, spreadX, spreadY int) []Pt {
	hearts := make([]Pt, 0, n)
	for i := 0; i < n; i++ {
		rx, ry := 8+i*spreadX/n, 3+i*spreadY/n
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
