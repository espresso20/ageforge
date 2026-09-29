package roguelike

import (
	"math"
	"sort"

	"github.com/espresso20/ageforge/mapmodel"
)

// scene.go lays one model out on its world: streets, walls, trails to the
// civs, fog, and the paths the ticking life walks. It is built once per
// model and only read while drawing.

type kind uint8

const (
	kNone kind = iota
	kStreet
	kRoad // trail outside the settlement
	kBridge
	kWall
	kTower
	kPlaza
	kCentre
	kTile   // ref: index into Town.Tiles
	kWonder // ref: index into Town.Wonders
	kSite   // ref: faction index, -1 on the outskirts
)

type cell struct {
	k     kind
	major bool  // an avenue: the ring, or a trail through town
	civ   int16 // faction index + 1 on a trail or a site
	ref   int32
}

// dial is everything an epoch changes: one grammar, a few dials.
type dial struct {
	road   int // 0 trodden ground, 1 light box, 2 heavy box, 3 double box
	wall   int // 0 none, 1 palisade, 2 stone wall, 3 ring boulevard
	night  bool
	smoke  string
	hazard string // pulse sequence of the catastrophe ring
	threat string // what the ring is
}

var dials = [7]dial{
	{road: 0, wall: 1, smoke: "∙°˚", hazard: "*·˙*", threat: "meteor fall"},
	{road: 1, wall: 2, smoke: "∙°˚", hazard: "×x»×", threat: "raiding host"},
	{road: 1, wall: 2, smoke: "°○∙", hazard: "▒░▒░", threat: "choking smog"},
	{road: 2, wall: 3, smoke: "°○∙", hazard: "░▒░·", threat: "fallout"},
	{road: 2, wall: 3, smoke: "∙·", hazard: "#%&0", threat: "glitch storm"},
	{road: 3, wall: 3, night: true, smoke: "·∙", hazard: "☼*☼·", threat: "solar flare"},
	{road: 3, wall: 3, night: true, smoke: "·", hazard: "╳╱╲╳", threat: "reality fracture"},
}

func epochIdx(e int) int { return clamp(e, 0, 6) }

type smokeSrc struct {
	p  mapmodel.Pt
	ph int
}

type scene struct {
	m                         *mapmodel.Model
	w                         *mapmodel.World
	d                         dial
	epoch                     int
	cells                     []cell
	vis                       []uint8 // 0 unknown, 1 remembered, 2 in sight
	visR, wallR               float64
	wallC, harb               mapmodel.Pt
	hasHb                     bool
	trails                    [][]mapmodel.Pt // per faction index; nil when unknown
	walks                     [][]mapmodel.Pt
	idle, scout, raid, hazard []mapmodel.Pt
	smoke                     []smokeSrc
	x0, y0, x1, y1            int // built bounds, for the compact fit
	targets                   []mapmodel.Pt
}

func (s *scene) in(x, y int) bool { return s.w.In(x, y) }

// at is the cell at (x, y); off the map it is a scratch cell.
func (s *scene) at(x, y int) *cell {
	if !s.w.In(x, y) {
		return &cell{}
	}
	return &s.cells[y*s.w.W+x]
}

func (s *scene) seen(p mapmodel.Pt) uint8 {
	if !s.in(p.X, p.Y) {
		return 0
	}
	return s.vis[p.Y*s.w.W+p.X]
}

func rdist(ax, ay, bx, by int) float64 {
	dx, dy := float64(ax-bx)/2, float64(ay-by)
	return math.Sqrt(float64(dx*dx) + float64(dy*dy))
}

func pt(x, y int) mapmodel.Pt { return mapmodel.Pt{X: x, Y: y} }

func newScene(m *mapmodel.Model) *scene {
	w, p := m.Town.World, m.Town.Plan
	if w == nil || p == nil || len(p.Plaza) != w.W*w.H || len(p.Street) != w.W*w.H {
		return nil
	}
	s := &scene{m: m, w: w, epoch: epochIdx(m.Epoch), cells: make([]cell, w.W*w.H)}
	s.d = dials[s.epoch]
	for i, pl := range p.Plaza {
		if pl {
			s.cells[i].k = kPlaza
		}
	}
	*s.at(w.CX, w.CY) = cell{k: kCentre}
	for i, t := range m.Town.Tiles {
		*s.at(t.X, t.Y) = cell{k: kTile, ref: int32(i)}
	}
	for i, wd := range m.Town.Wonders {
		for j := 0; j < mapmodel.WonderW*mapmodel.WonderH; j++ {
			*s.at(wd.At.X+j%mapmodel.WonderW, wd.At.Y+j/mapmodel.WonderW) = cell{k: kWonder, ref: int32(i)}
		}
	}
	s.streets()
	s.walls()
	s.trailsOut()
	s.visibility()
	s.life()
	s.hazards()
	s.harbinger()
	s.bounds()
	return s
}

func (s *scene) built(x, y int) bool {
	k := s.at(x, y).k
	return k == kTile || k == kWonder || k == kPlaza || k == kCentre
}

func roadLike(k kind) bool {
	return k == kStreet || k == kRoad || k == kBridge || k == kPlaza || k == kCentre
}

var dirs4 = [4][2]int{{0, -1}, {1, 0}, {0, 1}, {-1, 0}}

// streets opens the plan's grid streets that touch something built (bridges
// from the Iron era), then prunes dead-end stubs.
func (s *scene) streets() {
	w, p := s.w, s.m.Town.Plan
	for i, t := range w.T {
		x, y := i%w.W, i/w.W
		if !p.Street[i] || s.cells[i].k != kNone || !t.Land() && (t != mapmodel.TRiver || s.d.road == 0) {
			continue
		}
		near := false
		for j := 0; j < 9 && !near; j++ {
			near = s.built(x+j%3-1, y+j/3-1)
		}
		if near {
			s.cells[i].k = kStreet
			if t == mapmodel.TRiver {
				s.cells[i].k = kBridge
			}
		}
	}
	for pass := 0; pass < 2; pass++ {
		for i := range s.cells {
			x, y := i%w.W, i/w.W
			if s.cells[i].k != kStreet {
				continue
			}
			roads, b := 0, 0
			for _, d := range dirs4 {
				if roadLike(s.at(x+d[0], y+d[1]).k) {
					roads++
				}
				if s.built(x+d[0], y+d[1]) {
					b++
				}
			}
			if roads <= 1 && b == 0 {
				s.cells[i] = cell{}
			}
		}
	}
}

func coreLineage(lin string) bool {
	return lin != mapmodel.LinFood && lin != mapmodel.LinHarbor && lin != mapmodel.LinWood && lin != mapmodel.LinMines
}

// walls rings the core: a palisade from the Bronze Age, a stone wall with
// towers, then a ring boulevard where the wall stood. The ring is centred
// on the town square and its radius snaps up a fixed ladder of sizes, so
// it holds still as the town grows and steps outward only when the core
// outgrows it, never creeping by a tile at a time.
func (s *scene) walls() {
	w := s.w
	mode := s.d.wall
	if s.epoch == 0 && s.m.AgeIdx < 2 {
		mode = 0
	}
	cx, cy := w.CX, w.CY
	var ds []float64
	for _, t := range s.m.Town.Tiles {
		if coreLineage(t.Lineage) {
			ds = append(ds, rdist(t.X, t.Y, cx, cy))
		}
	}
	sort.Float64s(ds)
	need := 2.6
	if len(ds) > 0 {
		need = math.Max(need, ds[len(ds)*17/20])
	}
	r := wallStep(need + 1.1)
	s.wallR, s.wallC = r, pt(cx, cy)
	if mode == 0 {
		return
	}
	inside := func(x, y int) bool { return rdist(x, y, cx, cy) < r }
	R := int(r) + 3
	for y := cy - R; y <= cy+R; y++ {
		for x := cx - 2*R; x <= cx+2*R; x++ {
			if !s.in(x, y) || !onRing(x, y, inside) {
				continue
			}
			c, t := s.at(x, y), w.At(x, y)
			switch {
			case c.k == kTile || c.k == kWonder:
			case c.k == kStreet || c.k == kBridge:
				c.major = c.major || mode == 3 // a gate, or the ring itself
			case !t.Land():
				if mode == 3 && t == mapmodel.TRiver {
					*c = cell{k: kBridge, major: true}
				}
			case mode == 3:
				*c = cell{k: kStreet, major: true}
			case mode == 2 && mapmodel.Hash(w.Seed, 70, int64(x), int64(y))%13 == 0:
				*c = cell{k: kTower}
			default:
				*c = cell{k: kWall}
			}
		}
	}
}

// wallRadii is the ladder the wall's radius snaps to (each about a third
// bigger than the last).
var wallRadii = []float64{3.7, 5, 6.5, 8.5, 11, 14, 18, 23, 29, 37, 47, 60}

// wallStep is the smallest ladder radius that holds r.
func wallStep(r float64) float64 {
	for _, x := range wallRadii {
		if r <= x {
			return x
		}
	}
	return r
}

// onRing reports whether an outside tile is on the one-tile-thin,
// 4-connected boundary of the inside region (what box drawing needs to draw
// an unbroken wall).
func onRing(x, y int, inside func(x, y int) bool) bool {
	edge := func(x, y int) bool {
		return !inside(x, y) && (inside(x+1, y) || inside(x-1, y) || inside(x, y+1) || inside(x, y-1))
	}
	if inside(x, y) || edge(x, y) {
		return !inside(x, y)
	}
	for _, d := range [4][2]int{{1, 1}, {1, -1}, {-1, 1}, {-1, -1}} {
		if inside(x+d[0], y+d[1]) && edge(x+d[0], y) && edge(x, y+d[1]) {
			return true
		}
	}
	return false
}

func (s *scene) site(f *mapmodel.Faction) (mapmodel.Pt, bool) {
	if f.Site < 0 || f.Site >= len(s.w.Sites) {
		return mapmodel.Pt{}, false
	}
	p := s.w.Sites[f.Site]
	return p, s.in(p.X, p.Y)
}

// trailsOut marks every civ's site and runs a trail to each one met.
func (s *scene) trailsOut() {
	w, m := s.w, s.m
	s.trails = make([][]mapmodel.Pt, len(m.Factions))
	cost := func(i int) float64 {
		switch s.cells[i].k {
		case kStreet, kRoad, kBridge, kPlaza, kCentre:
			return 0.4
		case kTile, kWonder, kWall, kTower:
			return -1
		}
		return mapmodel.MoveCost(w.T[i])
	}
	for fi := range m.Factions {
		f := &m.Factions[fi]
		site, ok := s.site(f)
		if !ok {
			continue
		}
		*s.at(site.X, site.Y) = cell{k: kSite, civ: int16(fi + 1), ref: int32(fi)}
		for k, d := range [5][2]int{{-2, 0}, {2, 0}, {-1, 1}, {1, -1}, {-3, 1}} {
			x, y := site.X+d[0], site.Y+d[1]
			if int(mapmodel.Hash(w.Seed, 80, int64(fi), int64(k))%5) < f.Strength+1 && w.At(x, y).Land() && s.at(x, y).k == kNone {
				*s.at(x, y) = cell{k: kSite, civ: int16(fi + 1), ref: -1}
			}
		}
	}
	for fi := range m.Factions {
		site, ok := s.site(&m.Factions[fi])
		if !ok || !m.Factions[fi].Discovered {
			continue
		}
		s.trails[fi] = w.Path(pt(w.CX, w.CY), site, cost)
		for _, q := range s.trails[fi] {
			c := s.at(q.X, q.Y)
			switch {
			case c.k == kNone:
				c.k, c.civ = kRoad, int16(fi+1)
				if w.At(q.X, q.Y).Water() {
					c.k = kBridge
				}
			case roadLike(c.k):
				c.major = c.major || c.k == kStreet || c.k == kBridge
				if c.civ == 0 {
					c.civ = int16(fi + 1)
				}
			}
		}
	}
}

// visibility lifts the fog: a radius that grows with the age and every
// expedition, plus the lands of civs met and the trails to them. From the
// Cosmic era everything is known.
func (s *scene) visibility() {
	w, m := s.w, s.m
	s.vis = make([]uint8, w.W*w.H)
	if s.epoch >= 6 {
		for i := range s.vis {
			s.vis[i] = 2
		}
		return
	}
	s.visR = 12 + float64(2.2*float64(m.AgeIdx)) + float64(3*mapmodel.Log2(1+float64(m.Expeditions.Completed)))
	mark := func(cx, cy int, r float64) {
		R := int(r) + 6
		for y := cy - R; y <= cy+R; y++ {
			for x := cx - 2*R; x <= cx+2*R; x++ {
				if !s.in(x, y) {
					continue
				}
				d := rdist(x, y, cx, cy) + float64(1.6*mapmodel.Noise(w.Seed+3, float64(x)/3, float64(y)/1.5))
				switch i := y*w.W + x; {
				case d <= r:
					s.vis[i] = 2
				case d <= r+5 && s.vis[i] < 1:
					s.vis[i] = 1
				}
			}
		}
	}
	mark(w.CX, w.CY, s.visR)
	for fi := range m.Factions {
		if site, ok := s.site(&m.Factions[fi]); ok && m.Factions[fi].Discovered {
			mark(site.X, site.Y, 5)
			for k, q := range s.trails[fi] {
				if k%3 == 0 {
					mark(q.X, q.Y, 2)
				}
			}
		}
	}
}

func smoky(lin string) bool {
	return lin == mapmodel.LinMetal || lin == mapmodel.LinEnergy || lin == mapmodel.LinEngineer || lin == mapmodel.LinMines
}

// life plans the moving things: workers between homes and staffed works,
// idle hands in the square, smoke over busy works, scouts and raiders.
func (s *scene) life() {
	w, m := s.w, s.m
	var homes, works []mapmodel.Pt
	run, last := 0, ""
	for _, t := range m.Town.Tiles {
		if t.Key != last {
			run, last = 0, t.Key
		}
		run++
		b := m.Building(t.Key)
		switch {
		case t.Ruin || b == nil:
		case t.Lineage == mapmodel.LinHousing:
			homes = append(homes, t.Pt)
		case b.Workers > 0:
			works = append(works, t.Pt)
			if smoky(t.Lineage) && run <= 3 && len(s.smoke) < 48 {
				s.smoke = append(s.smoke, smokeSrc{t.Pt, int(mapmodel.Hash(int64(t.X), int64(t.Y)) % 7)})
			}
		}
	}
	if len(homes) == 0 {
		homes = []mapmodel.Pt{pt(w.CX, w.CY)}
	}
	if m.AgeIdx < 3 { // early on, people walk out to the woods to gather
		for y := w.CY - 10; y <= w.CY+10; y++ {
			for x := w.CX - 20; x <= w.CX+20; x++ {
				if w.At(x, y) == mapmodel.TForest && mapmodel.Hash(w.Seed, 90, int64(x), int64(y))%4 == 0 {
					works = append(works, pt(x, y))
				}
			}
		}
	}
	busy := m.Workers.Staffed
	n := min(18, int(float64(math.Sqrt(float64(busy))*0.9)))
	if busy > 0 && n < 2 {
		n = 2
	}
	cost := func(i int) float64 {
		switch s.cells[i].k {
		case kStreet, kRoad, kBridge, kPlaza, kCentre:
			return 0.5
		case kTile, kWonder, kWall, kTower, kSite:
			return 30
		}
		if c := mapmodel.MoveCost(w.T[i]); c < 0 {
			return c
		}
		return mapmodel.MoveCost(w.T[i]) + 1
	}
	for i := 0; i < n && len(works) > 0; i++ {
		h := homes[mapmodel.Hash(w.Seed, 91, int64(i))%uint64(len(homes))]
		t := works[mapmodel.Hash(w.Seed, 92, int64(i))%uint64(len(works))]
		if p := w.Path(h, t, cost); len(p) > 2 {
			s.walks = append(s.walks, p)
		}
	}
	var ring []mapmodel.Pt // idle hands stand about the square
	for y := w.CY - 2; y <= w.CY+2; y++ {
		for x := w.CX - 4; x <= w.CX+4; x++ {
			if s.at(x, y).k == kPlaza {
				ring = append(ring, pt(x, y))
			}
		}
	}
	h := func(p mapmodel.Pt) uint64 { return mapmodel.Hash(w.Seed, 93, int64(p.X), int64(p.Y)) }
	sort.Slice(ring, func(a, b int) bool { return h(ring[a]) < h(ring[b]) })
	if m.Workers.Idle > 0 {
		s.idle = ring[:min(len(ring), 9, 1+m.Workers.Idle/6)]
	}
	if sc := m.Expeditions.Scout; sc != nil { // scouts walk out toward the fog
		a := mapmodel.HashF(w.Seed, 94, mapmodel.HashStr(sc.Name))
		r := math.Max(s.visR, 14)
		tx := clamp(w.CX+int(float64(mapmodel.Cos(a)*r)*1.9), 1, w.W-2)
		ty := clamp(w.CY+int(float64(mapmodel.Sin(a)*r)*0.95), 1, w.H-2)
		s.scout = w.Path(pt(w.CX, w.CY), pt(tx, ty), cost)
	}
	if m.Expeditions.Military != nil { // raiders march on the least liked civ
		best := -1
		for fi, f := range m.Factions {
			if s.trails[fi] != nil && (best < 0 || f.Opinion < m.Factions[best].Opinion) {
				best = fi
			}
		}
		if best >= 0 {
			s.raid = s.trails[best][:len(s.trails[best])*2/3]
		}
	}
}

// hazards marks where a pending catastrophe presses on the town.
func (s *scene) hazards() {
	if s.m.Catastrophe.Pending == "" {
		return
	}
	r, c := s.wallR, s.wallC
	if r < 5 {
		r = 7
	}
	R := int(r) + 4
	for y := c.Y - R; y <= c.Y+R; y++ {
		for x := c.X - 2*R; x <= c.X+2*R; x++ {
			if d := rdist(x, y, c.X, c.Y); s.in(x, y) && d > r+0.5 && d < r+3.5 && mapmodel.Hash(s.w.Seed, 95, int64(x), int64(y))%5 == 0 {
				s.hazard = append(s.hazard, pt(x, y))
			}
		}
	}
}

// harbinger stands the omen just outside the walls, on open ground.
func (s *scene) harbinger() {
	if s.m.Harbinger == nil {
		return
	}
	a, r := mapmodel.HashF(s.w.Seed, 97), math.Max(s.wallR, 4)+2.5
	p := pt(s.wallC.X+int(float64(mapmodel.Cos(a)*r)*2), s.wallC.Y+int(float64(mapmodel.Sin(a)*r)))
	for rad := 0; rad <= 4; rad++ {
		for dy := -rad; dy <= rad; dy++ {
			for dx := -2 * rad; dx <= 2*rad; dx++ {
				if q := pt(p.X+dx, p.Y+dy); s.in(q.X, q.Y) && s.at(q.X, q.Y).k == kNone && s.w.At(q.X, q.Y).Land() {
					s.harb, s.hasHb = q, true
					return
				}
			}
		}
	}
}

// bounds finds the built area and the inspect targets (building anchors and
// wonders, nearest the square first).
func (s *scene) bounds() {
	w := s.w
	s.x0, s.y0, s.x1, s.y1 = w.CX-8, w.CY-3, w.CX+8, w.CY+3
	grow := func(x, y int) {
		s.x0, s.x1, s.y0, s.y1 = min(s.x0, x), max(s.x1, x), min(s.y0, y), max(s.y1, y)
	}
	for _, t := range s.m.Town.Tiles {
		grow(t.X, t.Y)
		if t.Anchor {
			s.targets = append(s.targets, t.Pt)
		}
	}
	for _, wd := range s.m.Town.Wonders {
		grow(wd.At.X, wd.At.Y)
		grow(wd.At.X+mapmodel.WonderW-1, wd.At.Y+mapmodel.WonderH-1)
		s.targets = append(s.targets, pt(wd.At.X+1, wd.At.Y))
	}
	sort.SliceStable(s.targets, func(i, j int) bool {
		a, b := s.targets[i], s.targets[j]
		if da, db := rdist(a.X, a.Y, w.CX, w.CY), rdist(b.X, b.Y, w.CX, w.CY); da != db {
			return da < db
		}
		return a.Y < b.Y || a.Y == b.Y && a.X < b.X
	})
}

// siteTargets lists the met civs' sites (region zoom tabs through them).
func (s *scene) siteTargets() []mapmodel.Pt {
	var out []mapmodel.Pt
	for i := range s.m.Factions {
		if p, ok := s.site(&s.m.Factions[i]); ok && s.m.Factions[i].Discovered {
			out = append(out, p)
		}
	}
	return out
}

func clamp(a, lo, hi int) int {
	if a < lo {
		return lo
	}
	if a > hi {
		return hi
	}
	return a
}
