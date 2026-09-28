package main

import (
	"math"
	"sort"
)

// The world: a seeded planet surface of continuous functions (height,
// temperature, moisture) plus a few structures computed once on a coarse
// 1-unit grid (rivers, regions, sea names, distance to the coast). World
// units are square; a terminal cell covers s units across and 2s down.

const (
	worldW = 360
	worldH = 220
)

// Pt is a point in world units.
type Pt struct{ X, Y float64 }

func (p Pt) Dist(q Pt) float64 { return math.Hypot(p.X-q.X, p.Y-q.Y) }

type blob struct{ x, y, rx, ry, w float64 }

// Region is a named Voronoi cell of land.
type Region struct {
	ID       int
	Seed     Pt
	Centroid Pt
	Name     string
	Area     int
	MeanH    float64
	Coastal  bool
	Biome    string
	Adj      map[int]bool
}

// SeaName is a labelled body of water.
type SeaName struct {
	Name string
	At   Pt
}

// World is the procedural geography for one seed.
type World struct {
	Seed    int64
	blobs   []blob
	h       []float32 // height on the 1-unit grid
	coastD  []float32 // distance to land, sea cells only (0 on land)
	region  []int16   // region id on the 1-unit grid, -1 at sea
	Regions []*Region
	Rivers  [][]Pt
	Seas    []SeaName
}

func newWorld(seed int64) *World {
	w := &World{Seed: seed}
	j := func(i int, amp float64) float64 { return (rnd01(seed, i, 99) - 0.5) * 2 * amp }
	w.blobs = []blob{
		{150 + j(1, 12), 118 + j(2, 8), 100, 64, 1.0}, // the home continent
		{112 + j(3, 10), 70 + j(4, 8), 55, 32, 0.8},   // its northern shoulder
		{300 + j(5, 10), 82 + j(6, 10), 52, 50, 0.95}, // the far continent
		{268 + j(7, 12), 184 + j(8, 6), 48, 24, 0.85}, // the southern land
		{52 + j(9, 8), 42 + j(10, 6), 30, 18, 0.8},    // northwest isles
		{34 + j(11, 6), 158 + j(12, 10), 20, 30, 0.8}, // western island
	}
	for i := 0; i < 9; i++ { // scattered islands
		x := 20 + rnd01(seed, i, 7)*320
		y := 16 + rnd01(seed, i, 8)*188
		w.blobs = append(w.blobs, blob{x, y, 6 + rnd01(seed, i, 9)*8, 5 + rnd01(seed, i, 10)*6, 0.7})
	}
	w.h = make([]float32, worldW*worldH)
	for y := 0; y < worldH; y++ {
		for x := 0; x < worldW; x++ {
			w.h[y*worldW+x] = float32(w.Height(float64(x), float64(y)))
		}
	}
	w.computeCoastDist()
	w.computeRivers()
	w.computeRegions()
	w.computeSeas()
	return w
}

// Height is the continuous elevation: < 0 is sea, 0..~1 is land.
func (w *World) Height(x, y float64) float64 {
	s := w.Seed
	wx := x + 22*(fbm(s+1, x/64, y/64, 3)-0.5)
	wy := y + 22*(fbm(s+2, x/64, y/64, 3)-0.5)
	c := 0.0
	for _, b := range w.blobs {
		dx, dy := (wx-b.x)/b.rx, (wy-b.y)/b.ry
		c += b.w * math.Exp(-(dx*dx+dy*dy)*1.7)
	}
	c = 1 - math.Exp(-c*1.6) // soft union
	n := fbm(s+3, wx/34, wy/34, 5)
	h := c*1.0 + (n-0.5)*0.8 - 0.5
	// the edges of the world fall away to ocean
	ex := math.Min(x, worldW-x) / 18
	ey := math.Min(y, worldH-y) / 14
	if e := math.Min(ex, ey); e < 1 {
		h -= (1 - e) * 0.6
	}
	if h > 0 {
		r := ridged(s+4, wx/50, wy/50, 4)
		h += r * r * r * 0.9 * clamp(h*3.5, 0, 1)
	}
	return h
}

// Temp is 0 (polar) .. 1 (tropical), cooled by altitude.
func (w *World) Temp(x, y, h float64) float64 {
	lat := math.Abs(y/worldH-0.56) * 2 // equator a little south of centre
	t := 1 - lat*1.05 + (fbm(w.Seed+5, x/40, y/40, 3)-0.5)*0.3
	return clamp(t-math.Max(h, 0)*0.55, 0, 1)
}

// Moist is 0 (arid) .. 1 (wet).
func (w *World) Moist(x, y float64) float64 {
	return fbm(w.Seed+6, x/46, y/46, 4)
}

// Biome classifies a land point.
func (w *World) Biome(x, y, h float64) string {
	t := w.Temp(x, y, h)
	m := w.Moist(x, y)
	switch {
	case h > 0.62:
		if t < 0.35 {
			return "peaks"
		}
		return "mountains"
	case t < 0.14:
		return "ice"
	case t < 0.3:
		return "tundra"
	case h > 0.38:
		return "hills"
	case t > 0.62 && m < 0.44:
		return "desert"
	case m > 0.56:
		return "forest"
	case m < 0.4:
		return "steppe"
	}
	return "grassland"
}

func (w *World) gh(x, y int) float64 {
	if x < 0 || y < 0 || x >= worldW || y >= worldH {
		return -1
	}
	return float64(w.h[y*worldW+x])
}

// H samples the grid height bilinearly (cheaper than Height, used for
// terrain shading where exactness does not matter).
func (w *World) H(x, y float64) float64 {
	x0, y0 := math.Floor(x), math.Floor(y)
	fx, fy := x-x0, y-y0
	ix, iy := int(x0), int(y0)
	return lerp(lerp(w.gh(ix, iy), w.gh(ix+1, iy), fx), lerp(w.gh(ix, iy+1), w.gh(ix+1, iy+1), fx), fy)
}

// CoastDist is the distance from (x, y) to the nearest land, 0 on land.
func (w *World) CoastDist(x, y float64) float64 {
	ix, iy := int(x), int(y)
	if ix < 0 || iy < 0 || ix >= worldW || iy >= worldH {
		return 99
	}
	return float64(w.coastD[iy*worldW+ix])
}

// RegionAt is the region id at (x, y), -1 at sea.
func (w *World) RegionAt(x, y float64) int {
	ix, iy := int(x), int(y)
	if ix < 0 || iy < 0 || ix >= worldW || iy >= worldH {
		return -1
	}
	return int(w.region[iy*worldW+ix])
}

func (w *World) computeCoastDist() {
	const inf = 1e9
	d := make([]float64, worldW*worldH)
	for i := range d {
		if w.h[i] > 0 {
			d[i] = 0
		} else {
			d[i] = inf
		}
	}
	// two-pass chamfer distance
	for pass := 0; pass < 2; pass++ {
		for y := 0; y < worldH; y++ {
			for x := 0; x < worldW; x++ {
				i := y*worldW + x
				if pass == 0 {
					if x > 0 {
						d[i] = math.Min(d[i], d[i-1]+1)
					}
					if y > 0 {
						d[i] = math.Min(d[i], d[i-worldW]+1)
						if x > 0 {
							d[i] = math.Min(d[i], d[i-worldW-1]+1.414)
						}
						if x < worldW-1 {
							d[i] = math.Min(d[i], d[i-worldW+1]+1.414)
						}
					}
				}
			}
		}
		for y := worldH - 1; y >= 0; y-- {
			for x := worldW - 1; x >= 0; x-- {
				i := y*worldW + x
				if x < worldW-1 {
					d[i] = math.Min(d[i], d[i+1]+1)
				}
				if y < worldH-1 {
					d[i] = math.Min(d[i], d[i+worldW]+1)
					if x < worldW-1 {
						d[i] = math.Min(d[i], d[i+worldW+1]+1.414)
					}
					if x > 0 {
						d[i] = math.Min(d[i], d[i+worldW-1]+1.414)
					}
				}
			}
		}
	}
	w.coastD = make([]float32, len(d))
	for i, v := range d {
		w.coastD[i] = float32(v)
	}
}

func (w *World) computeRivers() {
	type cand struct {
		x, y int
		h    float64
	}
	var cands []cand
	for i := 0; i < 900; i++ {
		x := int(rnd01(w.Seed, i, 31) * worldW)
		y := int(rnd01(w.Seed, i, 32) * worldH)
		h := w.gh(x, y)
		if h > 0.3 && h < 0.9 {
			cands = append(cands, cand{x, y, h})
		}
	}
	sort.Slice(cands, func(i, j int) bool { return cands[i].h > cands[j].h })
	onRiver := make(map[int]bool)
	var springs []cand
	for _, c := range cands {
		ok := true
		for _, s := range springs {
			if math.Hypot(float64(c.x-s.x), float64(c.y-s.y)) < 20 {
				ok = false
				break
			}
		}
		if ok {
			springs = append(springs, c)
		}
		if len(springs) >= 34 {
			break
		}
	}
	for _, s := range springs {
		x, y := s.x, s.y
		seen := map[int]bool{}
		path := []Pt{{float64(x), float64(y)}}
		joined := false
		for step := 0; step < 500; step++ {
			seen[y*worldW+x] = true
			bx, by, bh := -1, -1, math.Inf(1)
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					if dx == 0 && dy == 0 {
						continue
					}
					nx, ny := x+dx, y+dy
					if seen[ny*worldW+nx] {
						continue
					}
					h := w.gh(nx, ny)
					if dx != 0 && dy != 0 {
						h += 0.002
					}
					if h < bh {
						bx, by, bh = nx, ny, h
					}
				}
			}
			if bx < 0 {
				break
			}
			x, y = bx, by
			path = append(path, Pt{float64(x), float64(y)})
			if bh < 0 {
				joined = true
				break
			}
			if onRiver[y*worldW+x] {
				joined = true
				break
			}
		}
		if !joined || len(path) < 14 {
			continue
		}
		for _, p := range path {
			onRiver[int(p.Y)*worldW+int(p.X)] = true
		}
		// thin the polyline: every other point, keep the ends
		var thin []Pt
		for i, p := range path {
			if i%2 == 0 || i == len(path)-1 {
				thin = append(thin, Pt{p.X + 0.5, p.Y + 0.5})
			}
		}
		w.Rivers = append(w.Rivers, thin)
	}
}

func (w *World) computeRegions() {
	var seeds []Pt
	for i := 0; i < 4000 && len(seeds) < 70; i++ {
		p := Pt{8 + rnd01(w.Seed, i, 41)*(worldW-16), 8 + rnd01(w.Seed, i, 42)*(worldH-16)}
		if w.gh(int(p.X), int(p.Y)) <= 0.02 {
			continue
		}
		ok := true
		for _, s := range seeds {
			if s.Dist(p) < 25 {
				ok = false
				break
			}
		}
		if ok {
			seeds = append(seeds, p)
		}
	}
	w.region = make([]int16, worldW*worldH)
	for y := 0; y < worldH; y++ {
		for x := 0; x < worldW; x++ {
			i := y*worldW + x
			if w.h[i] <= 0 {
				w.region[i] = -1
				continue
			}
			fx, fy := float64(x), float64(y)
			// warped distance: organic, wandering borders
			wx := fx + 9*(fbm(w.Seed+11, fx/18, fy/18, 3)-0.5)*2
			wy := fy + 9*(fbm(w.Seed+12, fx/18, fy/18, 3)-0.5)*2
			best, bd := 0, math.Inf(1)
			for k, s := range seeds {
				d := (s.X-wx)*(s.X-wx) + (s.Y-wy)*(s.Y-wy)
				if d < bd {
					best, bd = k, d
				}
			}
			w.region[i] = int16(best)
		}
	}
	names := newNamer(w.Seed)
	for k, s := range seeds {
		w.Regions = append(w.Regions, &Region{ID: k, Seed: s, Name: names.region(), Adj: map[int]bool{}})
	}
	sx := make([]float64, len(seeds))
	sy := make([]float64, len(seeds))
	sh := make([]float64, len(seeds))
	for y := 0; y < worldH; y++ {
		for x := 0; x < worldW; x++ {
			r := int(w.region[y*worldW+x])
			if r < 0 {
				continue
			}
			reg := w.Regions[r]
			reg.Area++
			sx[r] += float64(x)
			sy[r] += float64(y)
			sh[r] += float64(w.h[y*worldW+x])
			for _, d := range [][2]int{{1, 0}, {0, 1}, {-1, 0}, {0, -1}} {
				nx, ny := x+d[0], y+d[1]
				if nx < 0 || ny < 0 || nx >= worldW || ny >= worldH {
					continue
				}
				o := int(w.region[ny*worldW+nx])
				if o < 0 {
					reg.Coastal = true
				} else if o != r {
					reg.Adj[o] = true
				}
			}
		}
	}
	for r, reg := range w.Regions {
		if reg.Area == 0 {
			continue
		}
		reg.Centroid = Pt{sx[r] / float64(reg.Area), sy[r] / float64(reg.Area)}
		reg.MeanH = sh[r] / float64(reg.Area)
		// a centroid can fall in the sea for a crescent-shaped region: use the seed
		if w.RegionAt(reg.Centroid.X, reg.Centroid.Y) != r {
			reg.Centroid = reg.Seed
		}
		reg.Biome = w.Biome(reg.Centroid.X, reg.Centroid.Y, w.H(reg.Centroid.X, reg.Centroid.Y))
	}
}

func (w *World) computeSeas() {
	type c struct {
		p Pt
		d float64
	}
	var cs []c
	for y := 6; y < worldH-6; y += 3 {
		for x := 6; x < worldW-6; x += 3 {
			d := w.CoastDist(float64(x), float64(y))
			if d > 9 {
				cs = append(cs, c{Pt{float64(x), float64(y)}, d})
			}
		}
	}
	sort.Slice(cs, func(i, j int) bool { return cs[i].d > cs[j].d })
	names := newNamer(w.Seed + 77)
	for _, cand := range cs {
		ok := true
		for _, s := range w.Seas {
			if s.At.Dist(cand.p) < 70 {
				ok = false
				break
			}
		}
		if ok {
			w.Seas = append(w.Seas, SeaName{names.sea(), cand.p})
		}
		if len(w.Seas) >= 6 {
			break
		}
	}
}

// ---------------------------------------------------------------------------
// names

type namer struct {
	seed int64
	n    int
	used map[string]bool
}

func newNamer(seed int64) *namer { return &namer{seed: seed, used: map[string]bool{}} }

func (n *namer) pick(list []string) string {
	n.n++
	return list[int(rnd01(n.seed, n.n, 5)*float64(len(list)))]
}

var (
	nmHead = []string{"Ald", "Bry", "Cael", "Dun", "Eld", "Fen", "Gal", "Har", "Ith", "Kor", "Lin", "Mor", "Nor", "Os", "Pel", "Quen", "Ral", "Sar", "Tal", "Ul", "Var", "Wen", "Yr", "Zan", "Ash", "Bel", "Cor", "Dra", "Thal", "Ves", "Mar", "Sol"}
	nmMid  = []string{"a", "e", "i", "o", "an", "or", "el", "ir", "", "", "u", "ae"}
	nmTail = []string{"mere", "vale", "wold", "moor", "ria", "dor", "heim", "land", "mark", "fell", "reach", "shire", "dale", "garde", "holm", "wick", "stan", "thal", "ia", "os", "en", "ath"}
	nmTown = []string{"ford", "bury", "ton", "stead", "wick", "port", "haven", "cross", "gate", "by", "minster", "field"}
	nmSeaA = []string{"Amber", "Grey", "Whispering", "Sunken", "Pale", "Iron", "Western", "Southern", "Morning", "Glass", "Salt", "Silent", "Endless", "Cold"}
	nmSeaB = []string{"Sea", "Sea", "Ocean", "Gulf", "Deep", "Waters"}
)

func (n *namer) word() string {
	for tries := 0; ; tries++ {
		s := n.pick(nmHead) + n.pick(nmMid) + n.pick(nmTail)
		if len(s) <= 11 && (!n.used[s] || tries > 20) {
			n.used[s] = true
			return s
		}
	}
}

func (n *namer) region() string { return n.word() }

func (n *namer) town() string {
	for tries := 0; ; tries++ {
		s := n.pick(nmHead) + n.pick(nmMid) + n.pick(nmTown)
		if len(s) <= 12 && (!n.used[s] || tries > 20) {
			n.used[s] = true
			return s
		}
	}
}

func (n *namer) sea() string {
	for {
		s := "The " + n.pick(nmSeaA) + " " + n.pick(nmSeaB)
		if !n.used[s] {
			n.used[s] = true
			return s
		}
	}
}
