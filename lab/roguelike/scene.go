package main

import (
	"math"
	"sort"

	"github.com/espresso20/ageforge/config"
)

// Kind is what occupies a tile on top of the terrain.
type Kind uint8

const (
	KNone Kind = iota
	KStreet
	KRoad // trail or road outside the settlement
	KBridge
	KWall
	KTower
	KPlaza
	KCentre
	KBuilding
	KWonder
	KRuin
	KSite
)

// Cell is one tile's overlay.
type Cell struct {
	K      Kind
	G      rune
	C      Class
	B      int16 // building index + 1
	F      int16 // faction index + 1
	Fresh  bool  // built since the last check-in
	Anchor bool  // first tile of a building cluster
	Under  bool  // understaffed
	Major  bool  // an avenue: the ring or a road out to a civ
}

// Bld is a building type laid out on the map.
type Bld struct {
	V      BuildingView
	G      rune
	C      Class
	Tiles  []Pt
	Delta  int // instances added since the last check-in
	Wonder bool
}

// tilesFor is how many tiles a building type gets: one per instance while
// counts are small (a primitive camp shows every hut), then sub-linear so
// a 400-house city still fits the screen.
func tilesFor(n int) int {
	if n <= 0 {
		return 0
	}
	if n <= 6 {
		return n
	}
	t := 6 + int(math.Ceil(4*math.Log2(float64(n)/6)))
	if t > 30 {
		t = 30
	}
	return t
}

func tilesForLineage(lin string, n int) int {
	t := tilesFor(n)
	if lin == "food" {
		t = t * 3
		if t > 60 {
			t = 60
		}
	}
	return t
}

// Scene is one state laid out on one world, ready to draw.
type Scene struct {
	W       *World
	P       *Plan
	V       MapView
	D       Dial
	Cells   []Cell
	Blds    []Bld
	Vis     []uint8 // 0 unknown, 1 remembered, 2 in sight
	Trails  [][]Pt  // per faction index; nil if unknown
	Walks   [][]Pt
	Idle    []Pt
	Scout   []Pt
	Raid    []Pt
	WallR   float64
	WallC   Pt // centre of the wall ring
	Hazard  []Pt
	Changes []string
}

func (s *Scene) idx(x, y int) int { return y*s.W.W + x }
func (s *Scene) cell(x, y int) *Cell {
	if !s.W.in(x, y) {
		return &Cell{}
	}
	return &s.Cells[y*s.W.W+x]
}

// epochOfTier finds the era a wonder was raised in, so older wonders keep
// their period look (a ziggurat stays a ziggurat in the neon city). A
// wonder's lineage tier is its age index; k is the fallback.
func epochOfTier(tier, k int) string {
	if order := config.AgeOrder(); tier > 0 && tier < len(order) {
		return config.EpochForAge(order[tier])
	}
	switch {
	case k < 3:
		return "stone_era"
	case k < 6:
		return "iron_era"
	case k < 9:
		return "steel_era"
	case k < 12:
		return "electric_era"
	case k < 15:
		return "digital_era"
	case k < 18:
		return "neon_era"
	}
	return "cosmic_era"
}

func (s *Scene) built(x, y int) bool {
	k := s.cell(x, y).K
	return k == KBuilding || k == KWonder || k == KPlaza || k == KCentre || k == KRuin
}

// streets opens the grid streets that touch something built.
func (s *Scene) streets() {
	w := s.W
	for y := 0; y < w.H; y++ {
		for x := 0; x < w.W; x++ {
			i := y*w.W + x
			if !s.P.Street[i] || s.Cells[i].K != KNone || !w.T[i].land() && w.T[i] != TRiver {
				continue
			}
			if w.T[i] == TRiver && s.D.Road == 0 {
				continue // no bridges before the Iron Age
			}
			n := 0
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					if s.built(x+dx, y+dy) {
						n++
					}
				}
			}
			if n >= 1 {
				if w.T[i] == TRiver {
					s.Cells[i] = Cell{K: KBridge}
				} else {
					s.Cells[i] = Cell{K: KStreet}
				}
			}
		}
	}
	// prune dead-end stubs so the grid does not bristle
	for pass := 0; pass < 2; pass++ {
		for y := 0; y < w.H; y++ {
			for x := 0; x < w.W; x++ {
				c := s.cell(x, y)
				if c.K != KStreet {
					continue
				}
				roads := 0
				for _, d := range [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
					if roadLike(s.cell(x+d[0], y+d[1]).K) {
						roads++
					}
				}
				b := 0
				for _, d := range [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
					if s.built(x+d[0], y+d[1]) {
						b++
					}
				}
				if roads <= 1 && b == 0 {
					*c = Cell{}
				}
			}
		}
	}
}

func roadLike(k Kind) bool {
	return k == KStreet || k == KRoad || k == KBridge || k == KPlaza || k == KCentre
}

// walls rings the core with a palisade, a stone wall, or (once the city
// outgrows its walls) a ring boulevard where the wall used to be.
func (s *Scene) walls() {
	mode := s.D.Wall
	if s.D.Epoch == "stone_era" && s.V.AgeIndex < 2 {
		return // no palisade before the Bronze Age
	}
	w := s.W
	// The wall hugs most of the core; a few late buildings sit outside it,
	// the way towns always spilled past their walls.
	// It is centred on the built core, not the square, so a town that grew
	// lopsided along its river gets a lopsided-looking ring.
	var core []Pt
	sx, sy := float64(w.CX), float64(w.CY)
	for _, b := range s.Blds {
		switch b.V.Lineage {
		case "food", "harbor", "organic_extraction", "geological_extraction":
			continue
		}
		for _, t := range b.Tiles {
			core = append(core, t)
			sx += float64(t.X)
			sy += float64(t.Y)
		}
	}
	cx := int(math.Round(sx / float64(len(core)+1)))
	cy := int(math.Round(sy / float64(len(core)+1)))
	ds := make([]float64, 0, len(core))
	for _, t := range core {
		ds = append(ds, rdist(t.X, t.Y, cx, cy))
	}
	sort.Float64s(ds)
	r := math.Max(2.6, rdist(w.CX, w.CY, cx, cy)+2)
	if len(ds) > 0 {
		r = math.Max(r, ds[len(ds)*17/20])
	}
	r += 1.1
	s.WallR, s.WallC = r, Pt{cx, cy}
	inside := func(x, y int) bool { return rdist(x, y, cx, cy) < r }
	for y := 0; y < w.H; y++ {
		for x := 0; x < w.W; x++ {
			if !onRing(x, y, inside) {
				continue
			}
			i := y*w.W + x
			c := &s.Cells[i]
			t := w.T[i]
			switch {
			case c.K == KBuilding || c.K == KWonder || c.K == KRuin:
				continue
			case c.K == KStreet || c.K == KBridge:
				c.Major = c.Major || mode == 3
				continue // a gate
			case !t.land():
				if mode == 3 && t == TRiver {
					*c = Cell{K: KBridge, Major: true}
				}
				continue
			}
			if mode == 3 {
				*c = Cell{K: KStreet, Major: true}
			} else if mode == 2 && hash(w.Seed, 70, int64(x), int64(y))%13 == 0 {
				*c = Cell{K: KTower}
			} else {
				*c = Cell{K: KWall}
			}
		}
	}
	if mode == 3 {
		s.dropParallel()
	}
}

// dropParallel removes minor streets that run right alongside the ring
// boulevard: the boulevard serves those lots now, and two parallel lines
// one cell apart would autotile into a comb of T-junctions.
func (s *Scene) dropParallel() {
	w := s.W
	isStreet := func(x, y int) bool { k := s.cell(x, y).K; return k == KStreet || k == KBridge }
	major := func(x, y int) bool { return isStreet(x, y) && s.cell(x, y).Major }
	var drop []Pt
	for y := 0; y < w.H; y++ {
		for x := 0; x < w.W; x++ {
			c := s.cell(x, y)
			if c.K != KStreet || c.Major {
				continue
			}
			minor := func(x, y int) bool { return isStreet(x, y) && !s.cell(x, y).Major }
			horiz := minor(x-1, y) || minor(x+1, y)
			vert := minor(x, y-1) || minor(x, y+1)
			if horiz && !vert && (major(x, y-1) && (major(x-1, y-1) || major(x+1, y-1)) ||
				major(x, y+1) && (major(x-1, y+1) || major(x+1, y+1))) {
				drop = append(drop, Pt{x, y})
			}
			if vert && !horiz && (major(x-1, y) && (major(x-1, y-1) || major(x-1, y+1)) ||
				major(x+1, y) && (major(x+1, y-1) || major(x+1, y+1))) {
				drop = append(drop, Pt{x, y})
			}
		}
	}
	for _, p := range drop {
		*s.cell(p.X, p.Y) = Cell{}
	}
}

// onRing reports whether an outside tile is on the one-tile-thin boundary
// of the inside region: it touches the inside edge-on, or it is the corner
// that joins two boundary tiles meeting only diagonally. Thin and
// 4-connected is what box-drawing needs to draw an unbroken wall.
func onRing(x, y int, inside func(x, y int) bool) bool {
	if inside(x, y) {
		return false
	}
	edge := func(x, y int) bool {
		if inside(x, y) {
			return false
		}
		return inside(x+1, y) || inside(x-1, y) || inside(x, y+1) || inside(x, y-1)
	}
	if edge(x, y) {
		return true
	}
	for _, d := range [][2]int{{1, 1}, {1, -1}, {-1, 1}, {-1, -1}} {
		if inside(x+d[0], y+d[1]) && edge(x+d[0], y) && edge(x, y+d[1]) {
			return true
		}
	}
	return false
}

// trails runs a road from the settlement to every civ the player has met.
// The roads are the ones the scouts found; caravans use them.
func (s *Scene) trails() {
	w := s.W
	s.Trails = make([][]Pt, len(s.V.Factions))
	cost := func(i int) float64 {
		c := s.Cells[i]
		switch c.K {
		case KStreet, KRoad, KBridge, KPlaza, KCentre:
			return 0.4
		case KBuilding, KWonder, KWall, KTower, KRuin:
			return -1
		}
		return w.moveCost(w.T[i])
	}
	for fi, f := range s.V.Factions {
		if fi >= len(w.Sites) {
			break
		}
		site := w.Sites[fi]
		sc := &s.Cells[s.idx(site.X, site.Y)]
		letter := 'C'
		for _, r := range f.Name {
			letter = r
			break
		}
		*sc = Cell{K: KSite, G: letter, C: factionClass(f), F: int16(fi + 1)}
		for _, d := range [][2]int{{-2, 0}, {2, 0}, {-1, 1}, {1, -1}, {-3, 1}} {
			if int(hash(w.Seed, 80, int64(fi), int64(d[0]))%5) < f.Strength+1 {
				x, y := site.X+d[0], site.Y+d[1]
				if w.in(x, y) && w.at(x, y).land() && s.cell(x, y).K == KNone {
					s.Cells[s.idx(x, y)] = Cell{K: KSite, G: '▪', C: factionClass(f), F: int16(fi + 1)}
				}
			}
		}
		if !f.Discovered {
			continue
		}
		path := w.path(Pt{w.CX, w.CY}, site, cost)
		s.Trails[fi] = path
		for _, pt := range path {
			c := &s.Cells[s.idx(pt.X, pt.Y)]
			if c.K == KStreet || c.K == KBridge {
				c.Major = true // the road out of town is an avenue inside it
			}
			if c.K != KNone {
				continue
			}
			if w.at(pt.X, pt.Y).water() {
				*c = Cell{K: KBridge}
			} else {
				*c = Cell{K: KRoad}
			}
		}
	}
}

func factionClass(f FactionView) Class {
	switch {
	case f.AtWar || f.Status == "embargo":
		return CDanger
	case f.Status == "allied" || f.Status == "friendly":
		return CFlora
	case f.Status == "rival":
		return CIdle
	}
	return CLabel
}

// visibility lifts the fog: a radius that grows with the age and with
// every expedition, plus the lands of civs you have met and the roads
// to them.
func (s *Scene) visibility() {
	w := s.W
	s.Vis = make([]uint8, w.W*w.H)
	if s.D.Space {
		for i := range s.Vis {
			s.Vis[i] = 2
		}
		return
	}
	vr := 12 + 2.2*float64(s.V.AgeIndex) + 3*math.Log2(1+float64(s.V.Expeditions))
	mark := func(cx, cy int, r float64) {
		for y := cy - int(r) - 6; y <= cy+int(r)+6; y++ {
			for x := cx - 2*int(r) - 12; x <= cx+2*int(r)+12; x++ {
				if !w.in(x, y) {
					continue
				}
				d := rdist(x, y, cx, cy)
				// ragged edge, like torchlight
				d += 1.6 * valueNoise(w.Seed+3, float64(x)/3, float64(y)/1.5)
				i := y*w.W + x
				switch {
				case d <= r && s.Vis[i] < 2:
					s.Vis[i] = 2
				case d <= r+5 && s.Vis[i] < 1:
					s.Vis[i] = 1
				}
			}
		}
	}
	mark(w.CX, w.CY, vr)
	for fi, f := range s.V.Factions {
		if f.Discovered && fi < len(w.Sites) {
			mark(w.Sites[fi].X, w.Sites[fi].Y, 5)
			for k, pt := range s.Trails[fi] {
				if k%3 == 0 {
					mark(pt.X, pt.Y, 2)
				}
			}
		}
	}
}

// life plans the moving things: workers walking between homes and work,
// idle hands loitering in the square, scouts and raiders heading out.
func (s *Scene) life() {
	w := s.W
	var homes, works []Pt
	for _, b := range s.Blds {
		if len(b.Tiles) == 0 || b.Wonder {
			continue
		}
		if b.V.Lineage == "housing" {
			homes = append(homes, b.Tiles...)
		} else if b.V.Workers > 0 {
			works = append(works, b.Tiles...)
		}
	}
	if len(homes) == 0 {
		homes = []Pt{{w.CX, w.CY}}
	}
	// In the first ages people also walk out to the forest to gather.
	if s.V.AgeIndex < 3 {
		for y := w.CY - 10; y <= w.CY+10; y++ {
			for x := w.CX - 20; x <= w.CX+20; x++ {
				if w.at(x, y) == TForest && hash(w.Seed, 90, int64(x), int64(y))%4 == 0 {
					works = append(works, Pt{x, y})
				}
			}
		}
	}
	busy := s.V.Pop - s.V.Idle
	n := int(math.Sqrt(float64(busy)) * 0.9)
	if n > 18 {
		n = 18
	}
	if busy > 0 && n < 2 {
		n = 2
	}
	cost := func(i int) float64 {
		c := s.Cells[i]
		switch c.K {
		case KStreet, KRoad, KBridge, KPlaza, KCentre:
			return 0.5
		case KBuilding, KWonder, KWall, KTower, KRuin, KSite:
			return 30
		}
		return w.moveCost(w.T[i]) + 1
	}
	for i := 0; i < n && len(works) > 0; i++ {
		h := homes[hash(w.Seed, 91, int64(i))%uint64(len(homes))]
		t := works[hash(w.Seed, 92, int64(i))%uint64(len(works))]
		if p := w.path(h, t, cost); len(p) > 2 {
			s.Walks = append(s.Walks, p)
		}
	}
	// idle workers stand around the square
	idle := 0
	if s.V.Idle > 0 {
		idle = 1 + s.V.Idle/6
		if idle > 9 {
			idle = 9
		}
	}
	var ring []Pt
	for y := w.CY - 2; y <= w.CY+2; y++ {
		for x := w.CX - 4; x <= w.CX+4; x++ {
			if s.cell(x, y).K == KPlaza {
				ring = append(ring, Pt{x, y})
			}
		}
	}
	sort.Slice(ring, func(a, b int) bool {
		return hash(w.Seed, 93, int64(ring[a].X), int64(ring[a].Y)) < hash(w.Seed, 93, int64(ring[b].X), int64(ring[b].Y))
	})
	if idle > len(ring) {
		idle = len(ring)
	}
	s.Idle = ring[:idle]
	// scouts walk toward the fog, raiders toward the nearest rival
	if s.V.ScoutActive != "" {
		a := hf(w.Seed, 94, int64(len(s.V.ScoutActive))) * 2 * math.Pi
		vr := 12 + 2.2*float64(s.V.AgeIndex) + 3*math.Log2(1+float64(s.V.Expeditions))
		tx := w.CX + int(math.Cos(a)*vr*2*0.95)
		ty := w.CY + int(math.Sin(a)*vr*0.95)
		if tx < 1 {
			tx = 1
		}
		if ty < 1 {
			ty = 1
		}
		if tx > w.W-2 {
			tx = w.W - 2
		}
		if ty > w.H-2 {
			ty = w.H - 2
		}
		s.Scout = w.path(Pt{w.CX, w.CY}, Pt{tx, ty}, cost)
	}
	if s.V.MilitaryActive != "" {
		best := -1
		for fi, f := range s.V.Factions {
			if s.Trails[fi] != nil && (best < 0 || f.Opinion < s.V.Factions[best].Opinion) {
				best = fi
			}
		}
		if best >= 0 {
			tr := s.Trails[best]
			s.Raid = tr[:len(tr)*2/3]
		}
	}
}

// hazards marks where a pending catastrophe is pressing on the city.
func (s *Scene) hazards() {
	if s.V.PendingCatastrophe == "" {
		return
	}
	w := s.W
	r := s.WallR
	if r < 5 {
		r = 7
	}
	c := s.WallC
	for y := c.Y - int(r) - 4; y <= c.Y+int(r)+4; y++ {
		for x := c.X - 2*int(r) - 8; x <= c.X+2*int(r)+8; x++ {
			if !w.in(x, y) {
				continue
			}
			d := rdist(x, y, c.X, c.Y)
			if d > r+0.5 && d < r+3.5 && hash(w.Seed, 95, int64(x), int64(y))%5 == 0 {
				s.Hazard = append(s.Hazard, Pt{x, y})
			}
		}
	}
}

// changes summarises what was built since the last check-in.
func (s *Scene) changes() {
	type ch struct {
		name string
		d    int
	}
	var cs []ch
	for _, b := range s.Blds {
		if b.Delta > 0 {
			cs = append(cs, ch{b.V.Name, b.Delta})
		}
	}
	sort.Slice(cs, func(a, b int) bool {
		if cs[a].d != cs[b].d {
			return cs[a].d > cs[b].d
		}
		return cs[a].name < cs[b].name
	})
	for _, c := range cs {
		s.Changes = append(s.Changes, "+"+itoa(c.d)+" "+c.name)
	}
}
