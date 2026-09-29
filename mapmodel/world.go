package mapmodel

import (
	"container/heap"
	"math"
)

// Terrain is a world tile's ground type. The world is pure data derived from
// the run seed: it never depends on game state, so it never reshuffles.
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

// Water reports whether a tile is sea, shallows or river.
func (t Terrain) Water() bool { return t <= TRiver }

// Land reports whether a tile is dry ground.
func (t Terrain) Land() bool { return t >= TBeach }

// World size in tiles. Fixed, so placement never depends on the terminal.
const (
	WorldW = 232
	WorldH = 84
)

// Pt is a tile coordinate.
type Pt struct{ X, Y int }

// World is the seeded land the settlement sits in: terrain, rivers, the
// settlement's centre, its name and one site per civilization slot.
type World struct {
	W, H   int
	Seed   int64
	T      []Terrain
	CX, CY int
	Name   string
	Sites  []Pt // civ sites, indexed like Catalog.Factions
	// Coastal is true when open water lies within a short walk of the centre
	// (the skyline gets a bay, the region map a harbour).
	Coastal bool
}

// In reports whether (x, y) is on the map.
func (w *World) In(x, y int) bool { return x >= 0 && y >= 0 && x < w.W && y < w.H }

// At returns the terrain at (x, y); off the map is deep water.
func (w *World) At(x, y int) Terrain {
	if !w.In(x, y) {
		return TDeep
	}
	return w.T[y*w.W+x]
}

// NewWorld generates the terrain for a seed with nSites civ sites.
func NewWorld(seed int64, nSites int) *World {
	w, h := WorldW, WorldH
	wd := &World{W: w, H: h, Seed: seed, T: make([]Terrain, w*h)}
	elev := make([]float64, w*h)
	cx, cy := float64(w)/2, float64(h)/2
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			nx, ny := float64(x)/26, float64(y)/13
			e := fbm(seed, nx, ny, 5)
			// Continent falloff: the map is an island, so the region view has
			// coasts on every side.
			dx, dy := (float64(x)-cx)/cx, (float64(y)-cy)/cy
			d2 := float64(float64(dx*dx)*0.9) + float64(dy*dy)
			e = float64(e*1.05) - float64(0.34*d2) + 0.13
			elev[y*w+x] = e
			m := fbm(seed+101, float64(nx*1.3), float64(ny*1.3), 4)
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
	wd.placeSites(nSites)
	wd.Name = cityName(seed)
	for dy := -7; dy <= 7 && !wd.Coastal; dy++ {
		for dx := -16; dx <= 16; dx++ {
			if t := wd.At(wd.CX+dx, wd.CY+dy); t == TShallow || t == TDeep {
				wd.Coastal = true
				break
			}
		}
	}
	return wd
}

// rivers traces a few rivers from high ground down the steepest slope.
func (wd *World) rivers(elev []float64) {
	type cand struct{ x, y int }
	var srcs []cand
	for i := 0; i < 400 && len(srcs) < 4; i++ {
		x := int(Hash(wd.Seed, 77, int64(i)) % uint64(wd.W))
		y := int(Hash(wd.Seed, 78, int64(i)) % uint64(wd.H))
		e := elev[y*wd.W+x]
		if e > 0.62 && e < 0.8 {
			ok := true
			for _, s := range srcs {
				if rdist(s.x, s.y, x, y) < 14 {
					ok = false
				}
			}
			if ok {
				srcs = append(srcs, cand{x, y})
			}
		}
	}
	for ri, s := range srcs {
		x, y := s.x, s.y
		seen := map[int]bool{}
		for step := 0; step < 400; step++ {
			i := y*wd.W + x
			if wd.T[i].Water() && wd.T[i] != TRiver {
				break
			}
			seen[i] = true
			if wd.T[i] != TMountain {
				wd.T[i] = TRiver
			}
			bx, by, be := -1, -1, math.Inf(1)
			for k, d := range [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
				nx, ny := x+d[0], y+d[1]
				if !wd.In(nx, ny) || seen[ny*wd.W+nx] {
					continue
				}
				// Sideways steps are cheap so rivers meander instead of
				// plunging straight down the screen.
				e := elev[ny*wd.W+nx] + float64(0.012*HashF(wd.Seed, int64(ri), int64(step), int64(k)))
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
			if be > elev[i] {
				elev[by*wd.W+bx] = elev[i] - 0.0005 // dig through small basins
			}
			x, y = bx, by
		}
	}
}

// placeCentre picks flat ground near a river, near the middle of the map.
func (wd *World) placeCentre() {
	best, bx, by := math.Inf(1), wd.W/2, wd.H/2
	for y := 8; y < wd.H-8; y++ {
		for x := 30; x < wd.W-30; x++ {
			t := wd.At(x, y)
			if t != TGrass && t != TForest {
				continue
			}
			rd := 99.0
			land := 0
			for dy := -6; dy <= 6; dy++ {
				for dx := -12; dx <= 12; dx++ {
					tt := wd.At(x+dx, y+dy)
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
			score := float64(rdist(x, y, wd.W/2, wd.H/2)*0.6) + float64(math.Abs(rd-3)*1.5) - float64(float64(land)*0.08)
			if score < best {
				best, bx, by = score, x, y
			}
		}
	}
	wd.CX, wd.CY = bx, by
}

// placeSites spreads the civ sites on a ring around the settlement.
func (wd *World) placeSites(n int) {
	for i := 0; i < n; i++ {
		a := float64(i)/float64(n) + float64(HashF(wd.Seed, 900, int64(i))*0.064)
		r := 0.62 + float64(0.3*HashF(wd.Seed, 901, int64(i)))
		rx := math.Min(float64(wd.CX-4), float64(wd.W-wd.CX-5))
		ry := math.Min(float64(wd.CY-3), float64(wd.H-wd.CY-4))
		if rx < 30 {
			rx = float64(float64(wd.W)/2) - 5
		}
		if ry < 12 {
			ry = float64(float64(wd.H)/2) - 4
		}
		tx := wd.CX + int(float64(float64(Cos(a)*rx)*r))
		ty := wd.CY + int(float64(float64(Sin(a)*ry)*r))
		best, bp := math.Inf(1), Pt{tx, ty}
		for dy := -8; dy <= 8; dy++ {
			for dx := -16; dx <= 16; dx++ {
				x, y := tx+dx, ty+dy
				if x < 2 || y < 1 || x >= wd.W-3 || y >= wd.H-2 {
					continue
				}
				t := wd.At(x, y)
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
	return syllA[Hash(seed, 5)%uint64(len(syllA))] + syllB[Hash(seed, 6)%uint64(len(syllB))]
}

// MoveCost is the path cost of stepping onto terrain; <0 is impassable.
func MoveCost(t Terrain) float64 {
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

func (p pq) Len() int { return len(p) }
func (p pq) Less(a, b int) bool {
	if p[a].f != p[b].f {
		return p[a].f < p[b].f
	}
	return p[a].i < p[b].i // a total order keeps the path identical everywhere
}
func (p pq) Swap(a, b int)       { p[a], p[b] = p[b], p[a] }
func (p *pq) Push(x interface{}) { *p = append(*p, x.(pqItem)) }
func (p *pq) Pop() interface{} {
	o := *p
	it := o[len(o)-1]
	*p = o[:len(o)-1]
	return it
}

// Path finds a cheap 4-connected path; cost(i) < 0 blocks a tile (the
// destination is always enterable). nil when there is none.
func (w *World) Path(from, to Pt, cost func(i int) float64) []Pt {
	n := w.W * w.H
	g := make([]float64, n)
	prev := make([]int32, n)
	for i := range g {
		g[i] = math.Inf(1)
		prev[i] = -1
	}
	s, t := from.Y*w.W+from.X, to.Y*w.W+to.X
	g[s] = 0
	h := func(i int) float64 {
		x, y := i%w.W, i/w.W
		return float64(float64(math.Abs(float64(x-to.X))+math.Abs(float64(y-to.Y))) * 0.5)
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
		x, y := it.i%w.W, it.i/w.W
		for _, d := range [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
			nx, ny := x+d[0], y+d[1]
			if !w.In(nx, ny) {
				continue
			}
			j := ny*w.W + nx
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
		out = append(out, Pt{i % w.W, i / w.W})
		if i == s {
			break
		}
	}
	for a, b := 0, len(out)-1; a < b; a, b = a+1, b-1 {
		out[a], out[b] = out[b], out[a]
	}
	return out
}
