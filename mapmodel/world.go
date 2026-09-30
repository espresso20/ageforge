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
	mouths := wd.springs(elev)
	wd.shores(elev)
	wd.rivers(elev, mouths)
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

// river is one river's plan: its spring and where it meets still water.
type river struct{ src, mouth Pt }

// springs picks a few springs on high ground and follows the water from
// each down the steepest slope, the way it would run, to see where it ends
// up: the sea, a lake, or a basin with no way out, which fills and becomes
// a lake. It returns each river's spring and mouth.
func (wd *World) springs(elev []float64) []river {
	var srcs []Pt
	for i := 0; i < 400 && len(srcs) < 4; i++ {
		x := int(Hash(wd.Seed, 77, int64(i)) % uint64(wd.W))
		y := int(Hash(wd.Seed, 78, int64(i)) % uint64(wd.H))
		e := elev[y*wd.W+x]
		if e > 0.62 && e < 0.8 {
			ok := true
			for _, s := range srcs {
				if rdist(s.X, s.Y, x, y) < 14 {
					ok = false
				}
			}
			if ok {
				srcs = append(srcs, Pt{x, y})
			}
		}
	}
	var out []river
	for ri, s := range srcs {
		walk := wd.course(elev, s.X, s.Y, ri)
		if len(walk) < 6 {
			continue
		}
		end := walk[len(walk)-1]
		if !wd.At(end.X, end.Y).Water() {
			// No way out: the water pools at the basin's floor.
			floor := walk[0]
			for _, p := range walk {
				if elev[p.Y*wd.W+p.X] < elev[floor.Y*wd.W+floor.X] {
					floor = p
				}
			}
			if rdist(floor.X, floor.Y, s.X, s.Y) < 8 {
				continue // too near its spring to make a river
			}
			wd.lake(floor, ri)
			end = floor
		}
		out = append(out, river{src: s, mouth: end})
	}
	return out
}

// course walks downhill from (x, y) until it reaches water or has nowhere
// lower to go: the steepest step each time, sideways steps a little
// favoured, digging through small basins on the way.
func (wd *World) course(elev []float64, x, y, ri int) []Pt {
	var out []Pt
	seen := map[int]bool{}
	for step := 0; step < 400; step++ {
		i := y*wd.W + x
		out = append(out, Pt{x, y})
		if wd.T[i].Water() {
			break // into the water it runs
		}
		seen[i] = true
		bx, by, be := -1, -1, math.Inf(1)
		for k, d := range [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
			nx, ny := x+d[0], y+d[1]
			if !wd.In(nx, ny) || seen[ny*wd.W+nx] {
				continue
			}
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
	return out
}

// lake pools water round a basin floor: roughly an ellipse twice as wide as
// it is tall (on screen, round), its shore swung in and out by slow noise.
// shores then gives it a clean edge and a deep middle.
func (wd *World) lake(c Pt, ri int) {
	r := 3.2 + float64(2.6*HashF(wd.Seed, 81, int64(ri))) // in rows
	R := int(r) + 2
	for y := c.Y - R; y <= c.Y+R; y++ {
		for x := c.X - 2*R; x <= c.X+2*R; x++ {
			if !wd.In(x, y) {
				continue
			}
			wob := float64(1.6*Noise(wd.Seed+83+int64(ri), float64(x)/6, float64(y)/3)) - 0.8
			if i := y*wd.W + x; rdist(x, y, c.X, c.Y) <= r+wob && !wd.T[i].Water() {
				wd.T[i] = TShallow
			}
		}
	}
}

// shores rounds off the sea and the lakes. The raw noise leaves ragged
// water: one-tile notches, spurs and specks, and shallows and deep water
// interleaved where the seabed hovers at the threshold. A few majority
// passes over an ellipse twice as wide as it is tall (a cell is about twice
// as tall as it is wide) smooth the water's edge. Depth then follows the
// distance to the shore: an even ring of shallows along every shore, deep
// water beyond (depth by the raw seabed speckled the two together). A beach
// runs along the water wherever the land is low.
func (wd *World) shores(elev []float64) {
	w, h := wd.W, wd.H
	water := make([]bool, w*h)
	for i, t := range wd.T {
		water[i] = t.Water()
	}
	next := make([]bool, w*h)
	for pass := 0; pass < 3; pass++ {
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				n, tot := 0, 0
				for dy := -1; dy <= 1; dy++ {
					for dx := -2; dx <= 2; dx++ {
						if dx*dx+4*dy*dy > 5 { // the ellipse: no far corners
							continue
						}
						tot++
						if nx, ny := x+dx, y+dy; !wd.In(nx, ny) || water[ny*w+nx] {
							n++ // off the map is open sea
						}
					}
				}
				i := y*w + x
				switch {
				case 2*n > tot+2:
					next[i] = true
				case 2*n < tot-2:
					next[i] = false
				default:
					next[i] = water[i] // a near tie keeps its side
				}
			}
		}
		water, next = next, water
	}
	// The majority passes leave the odd one-tile notch (water with land on
	// three sides) or spur (land with water on three sides): nip them off.
	for changed, pass := true, 0; changed && pass < 6; pass++ {
		changed = false
		for y := 1; y < h-1; y++ {
			for x := 1; x < w-1; x++ {
				i, n := y*w+x, 0
				for _, j := range [4]int{i - 1, i + 1, i - w, i + w} {
					if water[j] != water[i] {
						n++
					}
				}
				if n >= 3 {
					water[i], changed = !water[i], true
				}
			}
		}
	}
	dist := shoreDistance(water, w, h)
	for i := range wd.T {
		switch {
		case !water[i]:
			if wd.T[i].Water() {
				wd.T[i] = TBeach // a notch filled in
			}
		case dist[i] < 2.2:
			wd.T[i] = TShallow
		default:
			wd.T[i] = TDeep
		}
	}
	// A beach where low land meets the water; inland, old beach turns to
	// grass.
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			i := y*w + x
			if t := wd.T[i]; t.Water() || t == THills || t == TMountain {
				continue
			}
			wet := false
			for dy := -1; dy <= 1 && !wet; dy++ {
				for dx := -2; dx <= 2; dx++ {
					if nx, ny := x+dx, y+dy; dx*dx+4*dy*dy <= 5 && wd.In(nx, ny) && water[ny*w+nx] {
						wet = true
						break
					}
				}
			}
			switch {
			case wet && elev[i] < 0.45:
				wd.T[i] = TBeach
			case !wet && wd.T[i] == TBeach:
				wd.T[i] = TGrass
			}
		}
	}
}

// shoreDistance is each water tile's distance to the nearest land, in rows
// (a column counts half a row), by a two-pass chamfer; land is 0.
func shoreDistance(water []bool, w, h int) []float64 {
	const side, down, diag = 0.5, 1.0, 1.118
	d := make([]float64, w*h)
	for i, wet := range water {
		if wet {
			d[i] = math.Inf(1)
		}
	}
	at := func(x, y int) float64 {
		if x < 0 || y < 0 || x >= w || y >= h {
			return math.Inf(1) // off the map is open sea, far from land
		}
		return d[y*w+x]
	}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			i := y*w + x
			d[i] = math.Min(d[i], math.Min(math.Min(at(x-1, y)+side, at(x, y-1)+down),
				math.Min(at(x-1, y-1)+diag, at(x+1, y-1)+diag)))
		}
	}
	for y := h - 1; y >= 0; y-- {
		for x := w - 1; x >= 0; x-- {
			i := y*w + x
			d[i] = math.Min(d[i], math.Min(math.Min(at(x+1, y)+side, at(x, y+1)+down),
				math.Min(at(x+1, y+1)+diag, at(x-1, y+1)+diag)))
		}
	}
	return d
}

// rivers draws each river from its spring to its mouth. The river takes the
// cheapest way through low ground (so it follows the valleys without the
// downhill walk's dithering across flat land), is smoothed, given a gentle
// meander, and drawn as a band that widens downstream: about two cells
// across where it runs north-south and one row where it runs east-west,
// which read the same on screen.
func (wd *World) rivers(elev []float64, rs []river) {
	valley := func(i int) float64 {
		if t := wd.T[i]; t.Water() && t != TRiver {
			return 60 // across still water only to reach the mouth
		}
		return 1 + float64(40*math.Max(0, elev[i]-0.3))
	}
	before := append([]Terrain(nil), wd.T...)
	for ri, r := range rs {
		mouth, ok := wd.nearestWater(r.mouth, 6)
		if !ok {
			continue
		}
		if bed := wd.Path(r.src, mouth, valley); len(bed) >= 6 {
			wd.paintRiver(meander(wd.Seed, ri, smoothCourse(bed)))
		}
	}
	wd.stitchRivers()
	wd.pruneRivers(before)
}

// stitchRivers joins river tiles that touch only at a corner: on screen a
// band stepping down a row reads as broken there, so the lower row's run
// takes one more tile, under the upper one.
func (wd *World) stitchRivers() {
	for y := 0; y+1 < wd.H; y++ {
		for x := 0; x < wd.W; x++ {
			if wd.T[y*wd.W+x] != TRiver {
				continue
			}
			for _, dx := range [2]int{-1, 1} {
				nx := x + dx
				if !wd.In(nx, y+1) || wd.T[(y+1)*wd.W+nx] != TRiver {
					continue
				}
				side, below := wd.T[y*wd.W+nx], wd.T[(y+1)*wd.W+x]
				if !side.Water() && !below.Water() {
					wd.T[(y+1)*wd.W+x] = TRiver
				}
			}
		}
	}
}

// pruneRivers gives back to the land any stretch of river that is not part
// of a whole river: a scrap of a few tiles, or a band that never reaches
// the sea or a lake (a brush stroke cut off by the water it crossed).
func (wd *World) pruneRivers(before []Terrain) {
	seen := make([]bool, len(wd.T))
	for i, t := range wd.T {
		if t != TRiver || seen[i] {
			continue
		}
		comp, mouth := []int{i}, false
		seen[i] = true
		for k := 0; k < len(comp); k++ {
			x, y := comp[k]%wd.W, comp[k]/wd.W
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					nx, ny := x+dx, y+dy
					if !wd.In(nx, ny) {
						continue
					}
					j := ny*wd.W + nx
					switch t := wd.T[j]; {
					case t == TRiver && !seen[j]:
						seen[j] = true
						comp = append(comp, j)
					case t == TDeep || t == TShallow:
						mouth = true
					}
				}
			}
		}
		if !mouth || len(comp) < 8 {
			for _, j := range comp {
				wd.T[j] = before[j]
			}
		}
	}
}

// nearestWater is the still water nearest p within r rows (p itself when it
// is water): shores may have filled in the tile a river was headed for.
func (wd *World) nearestWater(p Pt, r int) (Pt, bool) {
	best, bd := p, math.Inf(1)
	for y := p.Y - r; y <= p.Y+r; y++ {
		for x := p.X - 2*r; x <= p.X+2*r; x++ {
			if t := wd.At(x, y); wd.In(x, y) && t.Water() && t != TRiver {
				if d := rdist(x, y, p.X, p.Y); d < bd {
					best, bd = Pt{x, y}, d
				}
			}
		}
	}
	return best, bd <= float64(r)
}

// fpt is a point between tiles.
type fpt struct{ X, Y float64 }

// smoothCourse turns a staircase of tiles into a curve: each point is the
// average of its neighbours along the course (fewer near the ends, which
// stay put, so the river still starts at its spring and meets the water).
func smoothCourse(c []Pt) []fpt {
	const win = 5
	out := make([]fpt, len(c))
	for i := range c {
		r := min(win, i, len(c)-1-i)
		var sx, sy float64
		for j := i - r; j <= i+r; j++ {
			sx += float64(c[j].X)
			sy += float64(c[j].Y)
		}
		n := float64(2*r + 1)
		out[i] = fpt{sx / n, sy / n}
	}
	return out
}

// meander swings a smoothed course from side to side: a long, slow sine
// along its length, about two and a half columns each way on screen (so a
// river running east-west swings a row and a bit), fading in and out at
// the ends so the spring and the mouth stay put.
func meander(seed int64, ri int, c []fpt) []fpt {
	if len(c) < 3 {
		return c
	}
	out := make([]fpt, len(c))
	phase := HashF(seed, 79, int64(ri))
	period := 30 + float64(16*HashF(seed, 80, int64(ri))) // in course steps
	for i := range c {
		a, b := c[max(0, i-3)], c[min(len(c)-1, i+3)]
		dx, dy := b.X-a.X, float64(b.Y-a.Y)*2 // on screen, a row is two columns
		l := math.Sqrt(float64(dx*dx) + float64(dy*dy))
		if l == 0 {
			out[i] = c[i]
			continue
		}
		t := float64(i) / float64(len(c)-1)
		fade := math.Min(1, float64(math.Min(t, 1-t)*5))
		amp := float64(float64(2.6*fade) * Sin(float64(i)/period+phase)) // in columns
		// the normal to the course on screen, back in tile units
		nx, ny := -dy/l, float64(dx/l)/2
		out[i] = fpt{float64(c[i].X) + float64(nx*amp), float64(c[i].Y) + float64(ny*amp)}
	}
	return out
}

// paintRiver draws a course as a band of river tiles: a round brush swept
// along it in screen terms (a row counts as two columns), a little over two
// columns across at the spring and about three at the mouth, so it runs
// unbroken whatever its angle. It cuts through anything on land (a gorge
// through the hills) and stops at the sea and the lakes, which it runs
// into.
func (wd *World) paintRiver(c []fpt) {
	for i := 0; i+1 < len(c); i++ {
		a, b := c[i], c[i+1]
		seg := math.Max(math.Abs(b.X-a.X), float64(math.Abs(b.Y-a.Y)*2))
		steps := max(1, int(math.Ceil(float64(seg*4))))
		for k := 0; k < steps; k++ {
			f := float64(k) / float64(steps)
			x, y := a.X+float64((b.X-a.X)*f), a.Y+float64((b.Y-a.Y)*f)
			t := (float64(i) + f) / float64(len(c)-1)
			hw := 1.05 + float64(0.5*t) // half the width, in columns
			hw2, half := float64(hw*hw), float64(hw*0.5)
			for ty := int(math.Floor(y - half)); ty <= int(math.Ceil(y+half)); ty++ {
				vy := float64((float64(ty) - y) * 2)
				for tx := int(math.Floor(x - hw)); tx <= int(math.Ceil(x+hw)); tx++ {
					vx := float64(tx) - x
					if !wd.In(tx, ty) || float64(vx*vx)+float64(vy*vy) > hw2 {
						continue
					}
					if j := ty*wd.W + tx; !wd.T[j].Water() {
						wd.T[j] = TRiver
					}
				}
			}
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
