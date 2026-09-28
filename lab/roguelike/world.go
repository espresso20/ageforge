package main

import (
	"container/heap"
	"math"
)

// Terrain is a world tile's ground type. The world is pure data derived
// from the run seed: it never depends on game state, so it never reshuffles.
type Terrain uint8

const (
	TDeep Terrain = iota
	TShallow
	TRiver
	TBeach
	TGrass
	TForest
	THills
	TMountain
)

func (t Terrain) water() bool { return t <= TRiver }
func (t Terrain) land() bool  { return t >= TBeach }

// World is the seeded terrain the settlement sits in.
type World struct {
	W, H   int
	Seed   int64
	T      []Terrain
	CX, CY int    // settlement centre
	Name   string // settlement name
	Sites  []Pt   // faction sites, one per civ slot
}

// Pt is a tile coordinate.
type Pt struct{ X, Y int }

func (w *World) in(x, y int) bool { return x >= 0 && y >= 0 && x < w.W && y < w.H }
func (w *World) at(x, y int) Terrain {
	if !w.in(x, y) {
		return TDeep
	}
	return w.T[y*w.W+x]
}

// hash is a small stateless integer hash, the basis of every "random"
// choice the map makes: same inputs, same glyph, forever.
func hash(vals ...int64) uint64 {
	h := uint64(1469598103934665603)
	for _, v := range vals {
		h ^= uint64(v)
		h *= 1099511628211
		h ^= h >> 29
		h *= 0xbf58476d1ce4e5b9
		h ^= h >> 32
	}
	return h
}

func hf(vals ...int64) float64 { return float64(hash(vals...)%1_000_000) / 1_000_000 }

// valueNoise is bilinear value noise on an integer lattice.
func valueNoise(seed int64, x, y float64) float64 {
	x0, y0 := math.Floor(x), math.Floor(y)
	fx, fy := x-x0, y-y0
	sx, sy := fx*fx*(3-2*fx), fy*fy*(3-2*fy)
	ix, iy := int64(x0), int64(y0)
	a := hf(seed, ix, iy)
	b := hf(seed, ix+1, iy)
	c := hf(seed, ix, iy+1)
	d := hf(seed, ix+1, iy+1)
	return (a*(1-sx)+b*sx)*(1-sy) + (c*(1-sx)+d*sx)*sy
}

func fbm(seed int64, x, y float64, oct int) float64 {
	sum, amp, norm := 0.0, 1.0, 0.0
	for i := 0; i < oct; i++ {
		sum += amp * valueNoise(seed+int64(i)*7919, x, y)
		norm += amp
		amp *= 0.5
		x, y = x*2.03, y*2.03
	}
	return sum / norm
}

// rdist is the "round" distance between two tiles: a terminal cell is about
// twice as tall as it is wide, so two columns count as one row. Circles in
// rdist look round on screen.
func rdist(ax, ay, bx, by int) float64 {
	dx := float64(ax-bx) / 2
	dy := float64(ay - by)
	return math.Sqrt(dx*dx + dy*dy)
}

// NewWorld generates the terrain for a seed.
func NewWorld(seed int64, w, h int) *World {
	wd := &World{W: w, H: h, Seed: seed, T: make([]Terrain, w*h)}
	elev := make([]float64, w*h)
	cx, cy := float64(w)/2, float64(h)/2
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			nx, ny := float64(x)/26, float64(y)/13
			e := fbm(seed, nx, ny, 5)
			// continent falloff: the map is an island so the region view
			// has coasts on every side.
			dx, dy := (float64(x)-cx)/cx, (float64(y)-cy)/cy
			d := math.Sqrt(dx*dx*0.9 + dy*dy)
			e = e*1.05 - 0.34*d*d + 0.13
			elev[y*w+x] = e
			m := fbm(seed+101, nx*1.3, ny*1.3, 4)
			var t Terrain
			switch {
			case e < 0.30:
				t = TDeep
			case e < 0.36:
				t = TShallow
			case e < 0.385:
				t = TBeach
			case e > 0.82:
				t = TMountain
			case e > 0.71:
				t = THills
			case m > 0.54:
				t = TForest
			default:
				t = TGrass
			}
			wd.T[y*w+x] = t
		}
	}
	wd.rivers(elev)
	wd.placeCentre()
	wd.placeSites()
	wd.Name = cityName(seed)
	return wd
}

// rivers traces a few rivers from high ground down the steepest slope to
// the sea.
func (wd *World) rivers(elev []float64) {
	type cand struct {
		x, y int
		e    float64
	}
	var srcs []cand
	for i := 0; i < 400 && len(srcs) < 4; i++ {
		x := int(hash(wd.Seed, 77, int64(i)) % uint64(wd.W))
		y := int(hash(wd.Seed, 78, int64(i)) % uint64(wd.H))
		e := elev[y*wd.W+x]
		if e > 0.62 && e < 0.8 {
			ok := true
			for _, s := range srcs {
				if rdist(s.x, s.y, x, y) < 14 {
					ok = false
				}
			}
			if ok {
				srcs = append(srcs, cand{x, y, e})
			}
		}
	}
	for ri, s := range srcs {
		x, y := s.x, s.y
		seen := map[int]bool{}
		for step := 0; step < 400; step++ {
			i := y*wd.W + x
			if wd.T[i].water() && wd.T[i] != TRiver {
				break
			}
			seen[i] = true
			if wd.T[i] != TMountain {
				wd.T[i] = TRiver
			}
			bx, by, be := -1, -1, math.Inf(1)
			for k, d := range [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
				nx, ny := x+d[0], y+d[1]
				if !wd.in(nx, ny) || seen[ny*wd.W+nx] {
					continue
				}
				// horizontal steps are cheap so rivers meander sideways
				// instead of plunging straight down the screen.
				e := elev[ny*wd.W+nx] + 0.012*hf(wd.Seed, int64(ri), int64(step), int64(k))
				if d[1] != 0 {
					e += 0.004
				}
				if e < be {
					bx, by, be = nx, ny, e
				}
			}
			if bx < 0 {
				break
			}
			// dig through small basins so the river reaches the sea
			if be > elev[i] {
				elev[by*wd.W+bx] = elev[i] - 0.0005
			}
			x, y = bx, by
		}
	}
}

// placeCentre picks flat ground near a river, near the middle of the map.
func (wd *World) placeCentre() {
	best, bx, by := math.Inf(1), wd.W/2, wd.H/2
	for y := 6; y < wd.H-6; y++ {
		for x := 12; x < wd.W-12; x++ {
			t := wd.at(x, y)
			if t != TGrass && t != TForest {
				continue
			}
			rd := 99.0
			land := 0
			for dy := -6; dy <= 6; dy++ {
				for dx := -12; dx <= 12; dx++ {
					tt := wd.at(x+dx, y+dy)
					if tt == TRiver {
						if d := rdist(x, y, x+dx, y+dy); d < rd {
							rd = d
						}
					}
					if tt == TGrass || tt == TForest || tt == THills {
						land++
					}
				}
			}
			score := rdist(x, y, wd.W/2, wd.H/2)*0.6 + math.Abs(rd-3)*1.5 - float64(land)*0.08
			if score < best {
				best, bx, by = score, x, y
			}
		}
	}
	wd.CX, wd.CY = bx, by
}

// placeSites spreads eleven civ sites on a ring around the settlement.
func (wd *World) placeSites() {
	n := 11
	for i := 0; i < n; i++ {
		a := 2*math.Pi*float64(i)/float64(n) + hf(wd.Seed, 900, int64(i))*0.4
		r := 0.62 + 0.3*hf(wd.Seed, 901, int64(i))
		rx := math.Min(float64(wd.CX-4), float64(wd.W-wd.CX-5))
		ry := math.Min(float64(wd.CY-3), float64(wd.H-wd.CY-4))
		if rx < 30 {
			rx = float64(wd.W)/2 - 5
		}
		if ry < 12 {
			ry = float64(wd.H)/2 - 4
		}
		tx := wd.CX + int(math.Cos(a)*rx*r)
		ty := wd.CY + int(math.Sin(a)*ry*r)
		// walk to the nearest dry land
		best, bp := math.Inf(1), Pt{tx, ty}
		for dy := -8; dy <= 8; dy++ {
			for dx := -16; dx <= 16; dx++ {
				x, y := tx+dx, ty+dy
				if x < 2 || y < 1 || x >= wd.W-3 || y >= wd.H-2 {
					continue
				}
				t := wd.at(x, y)
				if t == TGrass || t == TForest || t == TBeach || t == THills {
					if d := rdist(x, y, tx, ty); d < best {
						best, bp = d, Pt{x, y}
					}
				}
			}
		}
		wd.Sites = append(wd.Sites, bp)
	}
}

var syllA = []string{"Ash", "Bram", "Cor", "Dun", "Eld", "Fen", "Gal", "Hol", "Ir", "Kest", "Lun", "Mor", "Nor", "Ost", "Rav", "Sel", "Thorn", "Umb", "Vel", "Wyn"}
var syllB = []string{"vale", "ford", "hold", "mere", "wick", "stead", "haven", "gate", "moor", "crest", "fall", "reach"}

func cityName(seed int64) string {
	return syllA[hash(seed, 5)%uint64(len(syllA))] + syllB[hash(seed, 6)%uint64(len(syllB))]
}

// moveCost is the A* cost of stepping onto a tile; <0 is impassable.
func (wd *World) moveCost(t Terrain) float64 {
	switch t {
	case TDeep:
		return -1
	case TShallow:
		return 14
	case TRiver:
		return 5
	case TMountain:
		return 10
	case THills:
		return 3
	case TForest:
		return 2
	}
	return 1
}

type pqItem struct {
	i int
	f float64
}
type pq []pqItem

func (p pq) Len() int            { return len(p) }
func (p pq) Less(a, b int) bool  { return p[a].f < p[b].f }
func (p pq) Swap(a, b int)       { p[a], p[b] = p[b], p[a] }
func (p *pq) Push(x interface{}) { *p = append(*p, x.(pqItem)) }
func (p *pq) Pop() interface{} {
	o := *p
	it := o[len(o)-1]
	*p = o[:len(o)-1]
	return it
}

// path finds a cheap 4-connected path; cost(i) < 0 blocks a tile.
func (wd *World) path(from, to Pt, cost func(i int) float64) []Pt {
	n := wd.W * wd.H
	g := make([]float64, n)
	prev := make([]int32, n)
	for i := range g {
		g[i] = math.Inf(1)
		prev[i] = -1
	}
	s, t := from.Y*wd.W+from.X, to.Y*wd.W+to.X
	g[s] = 0
	h := func(i int) float64 {
		x, y := i%wd.W, i/wd.W
		return (math.Abs(float64(x-to.X)) + math.Abs(float64(y-to.Y))) * 0.5
	}
	open := &pq{{s, h(s)}}
	for open.Len() > 0 {
		it := heap.Pop(open).(pqItem)
		if it.i == t {
			break
		}
		if it.f-h(it.i) > g[it.i]+1e-9 {
			continue
		}
		x, y := it.i%wd.W, it.i/wd.W
		for _, d := range [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
			nx, ny := x+d[0], y+d[1]
			if !wd.in(nx, ny) {
				continue
			}
			j := ny*wd.W + nx
			c := cost(j)
			if j == t && c < 0 {
				c = 1
			}
			if c < 0 {
				continue
			}
			if ng := g[it.i] + c; ng < g[j] {
				g[j] = ng
				prev[j] = int32(it.i)
				heap.Push(open, pqItem{j, ng + h(j)})
			}
		}
	}
	if prev[t] < 0 && s != t {
		return nil
	}
	var out []Pt
	for i := t; i != -1; i = int(prev[i]) {
		out = append(out, Pt{i % wd.W, i / wd.W})
		if i == s {
			break
		}
	}
	for a, b := 0, len(out)-1; a < b; a, b = a+1, b-1 {
		out[a], out[b] = out[b], out[a]
	}
	return out
}
