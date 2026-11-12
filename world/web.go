package world

import (
	"math/bits"
	"slices"

	"github.com/espresso20/ageforge/mapmodel"
)

// web.go: what the Transcendent Age opens. Past the Local Group the sky is the
// cosmic web: clusters of galaxies strung on filaments. The web is the
// player's own network: every filament is tied to one building line, and the
// filament glows as the line is built up. The shape is the save's seed alone,
// so the same seed always weaves the same web; the glow is the only thing the
// game state touches.

// NumLines is how many building lines a palette colours: one per entry of
// mapmodel.LineageOrder (a test holds it there).
const NumLines = 17

// LineIndex is a building line's place in mapmodel.LineageOrder, or -1.
func LineIndex(line string) int { return slices.Index(mapmodel.LineageOrder, line) }

const (
	webNodes  = 30     // clusters, the Local Group among them
	webHalfX  = 230    // Mly: nodes fall in a box this wide and half as tall
	webGap    = 36     // Mly: no two clusters closer
	webNear   = 3      // each cluster reaches out to this many nearest neighbours
	webLGSize = 6      // Mly: the Local Group node's radius, which the frame below is drawn to fill
	webStride = 700000 // room the clusters' made-up name ids have, clear of the stars' and galaxies'
)

// StrandUnbuilt is the weight of the faint thread a line nobody built draws as.
const StrandUnbuilt = 40

// Strength is how strongly a building line glows, 0 for a line with nothing
// built, else 70..255. It is one monotone formula: seventy as soon as there is
// a building, plus 3/4 of sixteen steps per doubling of the count (a hundred
// buildings land near 150, a thousand near 190), plus ten for each tier of the
// best building. A line never glows less for having more, or higher tier.
func Strength(count, tier int) int {
	if count <= 0 {
		return 0
	}
	l := bits.Len(uint(count))
	lg16 := 16*(l-1) + 16*(count-1<<(l-1))>>(l-1) // log2(count) in sixteenths, straight between powers of two
	return min(255, 70+lg16*3/4+10*clamp(tier, 0, 3))
}

// webFrame is the cosmic web with lg, the Local Group, hung on the node it
// stands for.
func webFrame(m *mapmodel.Model, seed int64, lg *Frame) *Frame {
	f := &Frame{Kind: FrameWeb, Key: "cosmic_web", Unit: UnitMly, Radius: webHalfX + 20, Band: Band{Min: 20, Max: 560}, Schem: true,
		Children: []*Frame{lg}}
	lg.Anchor = "local_group"
	type pt struct{ x, y int }
	pts := []pt{{9, -7}} // the Local Group, near the middle
	for i := 0; len(pts) < webNodes && i < 800; i++ {
		p := pt{Roll(2*webHalfX+1, seed, 7700, int64(i)) - webHalfX, Roll(webHalfX+1, seed, 7701, int64(i)) - webHalfX/2}
		free := true
		for _, q := range pts {
			free = free && Sq(p.x-q.x)+Sq(p.y-q.y) >= Sq(webGap)
		}
		if free {
			pts = append(pts, p)
		}
	}
	// the filaments: each cluster to its nearest few, no pair twice
	type edge struct{ a, b int }
	var edges []edge
	for a := range pts {
		near := make([]int, 0, len(pts)-1)
		for b := range pts {
			if b != a {
				near = append(near, b)
			}
		}
		slices.SortFunc(near, func(i, j int) int {
			return (Sq(pts[a].x-pts[i].x) + Sq(pts[a].y-pts[i].y)) - (Sq(pts[a].x-pts[j].x) + Sq(pts[a].y-pts[j].y))
		})
		for _, b := range near[:webNear] {
			e := edge{min(a, b), max(a, b)}
			if !slices.Contains(edges, e) {
				edges = append(edges, e)
			}
		}
	}
	// each filament carries one line: a seeded shuffle of the filaments deals
	// the lines out in turn, so every line has one, and the rest fall by chance
	deal := make([]int, len(edges))
	for i := range deal {
		deal[i] = i
	}
	slices.SortFunc(deal, func(a, b int) int {
		return Roll(1<<30, seed, 7702, int64(a)) - Roll(1<<30, seed, 7702, int64(b))
	})
	lineOf := make([]int, len(edges))
	order := slices.Clone(mapmodel.LineageOrder)
	slices.SortFunc(order, func(a, b string) int {
		return Roll(1<<30, seed, 7703, mapmodel.HashStr(a)) - Roll(1<<30, seed, 7703, mapmodel.HashStr(b))
	})
	n := len(order)
	for j, e := range deal {
		if j < n {
			lineOf[e] = j
		} else {
			lineOf[e] = Roll(n, seed, 7704, int64(e))
		}
	}
	have := map[string]*mapmodel.Lineage{}
	for _, l := range m.Lineages {
		have[l.Key] = l
	}
	for i, e := range edges {
		a, b := pts[e.a], pts[e.b]
		line := order[lineOf[i]]
		count, tier := 0, 0
		if l := have[line]; l != nil && l.Count > 0 {
			count, tier = l.Count, tierOf(m, l)
		}
		w, tone := uint8(StrandUnbuilt), ToneStar
		if s := Strength(count, tier); s > 0 {
			w, tone = uint8(s), ToneGold
		}
		// the filament bends a little, a quadratic curve a to b through c
		dx, dy, bend := b.x-a.x, b.y-a.y, Roll(51, seed, 7705, int64(i))-25
		c := Vec{float64(a.x+b.x)/2 - float64(dy*bend)/100, float64(a.y+b.y)/2 + float64(dx*bend)/100}
		A, B := Vec{float64(a.x), float64(a.y)}, Vec{float64(b.x), float64(b.y)}
		f.Objects = append(f.Objects, Object{Kind: KindStrand, Key: "strand_" + itoa(i), Label: line, Inspect: line, Line: line, Count: count, Tier: tier,
			Pos: Vec{(A.X + 2*c.X + B.X) / 4, (A.Y + 2*c.Y + B.Y) / 4}, Path: []Vec{A, c, B}, Tone: tone, Weight: w, Pick: true, Gated: "lineage"})
	}
	for i, p := range pts { // the clusters, over the filaments
		o := Object{Kind: KindCluster, Key: "web_" + itoa(i), Pos: Vec{float64(p.x), float64(p.y)}, Size: float64(3 + Roll(4, seed, 7706, int64(i))),
			Tone: ToneStar, Weight: 230, Pick: true}
		if i == 0 {
			o.Key, o.Size, o.Tone, o.Label, o.Inspect = "local_group", webLGSize, ToneGold, "local_group", "local_group"
		} else {
			k := StarKey(seed, webStride+i)
			o.Label, o.Inspect, o.Doubt = k, k, uint8(2+Roll(2, seed, 7707, int64(i)))
		}
		o.Seed = vseed(seed, o.Key, 0)
		f.Objects = append(f.Objects, o)
	}
	return f
}
