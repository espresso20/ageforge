package main

import (
	"math"
	"sort"
)

// A genome is the full potential shape of one limb (or of the root
// system): every segment it could ever grow, in the order it grows them. It
// depends only on the run's seed and the lineage key, never on counts, so a
// new building reveals the next segment and never moves one already drawn.
//
// Directions are stored relative to the limb's sector of the crown (-1 is
// the sector's left edge, +1 its right), not as absolute angles. A limb's
// sector widens with its share of the economy, so each lineage owns a wedge
// of the crown in proportion to its size; the limb's shape inside its wedge
// never changes.

type gseg struct {
	Rel    float64 // direction within the sector, -1..1
	Len    float64 // length as a fraction of the bough
	Depth  int
	Parent int     // index into the genome, -1 for the bough
	Bend   float64 // sideways bow of the curve, as a fraction of length
}

type genome struct {
	segs []gseg // in growth order
}

// seg is a laid-out segment in tree units (ground at y=0, up is +y).
type seg struct {
	X0, Y0, X1, Y1 float64
	Depth, Parent  int
	Bend           float64
}

// rng is a small deterministic generator (splitmix64).
type rng struct{ s uint64 }

func (r *rng) next() uint64 {
	r.s += 0x9e3779b97f4a7c15
	z := r.s
	z = (z ^ (z >> 30)) * 0xbf58476d1ce4e5b9
	z = (z ^ (z >> 27)) * 0x94d049bb133111eb
	return z ^ (z >> 31)
}

func (r *rng) f() float64 { return float64(r.next()>>11) / float64(1<<53) }

// between returns a value in [a, b).
func (r *rng) between(a, b float64) float64 { return a + (b-a)*r.f() }

// grow builds a genome maxDepth generations deep. Within a generation the
// growth order is a seeded shuffle, so a limb fills out unevenly, like a
// real one.
func grow(seed uint64, maxDepth int) genome {
	r := &rng{s: seed}
	type node struct {
		rel, len      float64
		depth, parent int
		prio          float64
	}
	var segs []gseg
	gen := []node{{0, 1, 0, -1, 0}}
	for depth := 0; depth <= maxDepth && len(gen) > 0; depth++ {
		sort.SliceStable(gen, func(i, j int) bool { return gen[i].prio < gen[j].prio })
		var next []node
		for _, n := range gen {
			idx := len(segs)
			segs = append(segs, gseg{Rel: n.rel, Len: n.len, Depth: depth, Parent: n.parent, Bend: r.between(-0.2, 0.2)})
			kids := 2
			if r.f() < 0.3 {
				kids = 3
			}
			// Children fan out from the parent's direction, narrower as the
			// limb refines, and stay inside the sector.
			spread := r.between(0.5, 0.95) / (1 + 0.35*float64(depth))
			for k := 0; k < kids; k++ {
				off := spread * (2*float64(k)/float64(kids-1) - 1)
				rel := math.Max(-1, math.Min(1, n.rel+off+r.between(-0.1, 0.1)))
				next = append(next, node{rel, n.len * r.between(0.68, 0.82), depth + 1, idx, r.f() + float64(k)*0.05})
			}
		}
		gen = next
	}
	return genome{segs: segs}
}

// layout places the first n segments of g: the bough leaves (ax, ay), the
// sector is centred on angle centre (radians from vertical, + right) with
// half-width half, the bough is L long, and up tilts the finer segments
// back toward vertical (phototropism). Roots pass centre = pi and up = 0.
func (g genome) layout(n int, ax, ay, centre, half, L, up float64) []seg {
	out := make([]seg, n)
	for i := 0; i < n; i++ {
		s := g.segs[i]
		x0, y0 := ax, ay
		if s.Parent >= 0 {
			x0, y0 = out[s.Parent].X1, out[s.Parent].Y1
		}
		a := centre + s.Rel*half
		a *= 1 - up*float64(s.Depth)/7
		l := L * s.Len
		out[i] = seg{X0: x0, Y0: y0, X1: x0 + math.Sin(a)*l, Y1: y0 + math.Cos(a)*l, Depth: s.Depth, Parent: s.Parent, Bend: s.Bend}
	}
	return out
}

// reveal is how many segments a limb of count buildings shows: about one
// per building early, sub-linear later, so the primitive camp's two
// buildings already make a visible limb and the cosmic empire's hundreds do
// not choke the crown.
func reveal(count, pool int) int {
	if count <= 0 {
		return 0
	}
	n := int(math.Round(1 + 1.6*math.Pow(float64(count), 0.8)))
	return min(n, pool)
}

// tips returns the revealed segments with no revealed child: where the
// leaves grow.
func tips(segs []seg) []int {
	hasKid := make([]bool, len(segs))
	for _, s := range segs {
		if s.Parent >= 0 {
			hasKid[s.Parent] = true
		}
	}
	var out []int
	for i := range segs {
		if !hasKid[i] {
			out = append(out, i)
		}
	}
	return out
}
