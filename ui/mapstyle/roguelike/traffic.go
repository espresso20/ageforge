package roguelike

import (
	"math"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/ui/mapstyle"
)

// traffic.go is the age's traffic, from the roster both styles share
// (mapmodel.MoversAt): carts, trams and cars on the streets, boats on the
// river and open water, trains on the railway, planes, drones and
// satellites overhead. The scene lays the lanes once (the street network,
// the water, the railway, straight lines across the sky); a frame only
// indexes them. Movers draw on streets, rails, water and open sky, never
// over a building or a label, and nothing appears before the age that
// introduces it.

// mover is one moving thing and the lane it runs.
type mover struct {
	k    mapmodel.Mover
	info mapmodel.MoverInfo
	// lanes is the lane it runs, or for a one-way sky mover the few lines
	// it takes in turn, one per crossing.
	lanes [][]mapmodel.Pt
	long  int  // the longest lane
	cars  int  // cells long
	dwell int  // steps at rest at each end, or between one-way runs
	once  bool // one way along the lane, then rest; else there and back
	off   int  // phase, in steps
	n     int  // its number among its kind (picks its inspect line)
}

// period is the mover's cycle in steps (one step is info.Pace frames).
func (mv *mover) period() int {
	if mv.once {
		return mv.long + mv.cars - 1 + mv.dwell
	}
	return 2 * (max(0, len(mv.lanes[0])-mv.cars) + mv.dwell)
}

// cells calls fn for every cell the mover covers at frame f, front first
// (j is 0 at the front), with its direction along the lane (+1 or -1).
// It calls nothing while the mover rests off the map.
func (mv *mover) cells(f int, fn func(p mapmodel.Pt, j, dir int)) {
	per := mv.period()
	if per <= 0 || len(mv.lanes) == 0 {
		return
	}
	t := f/mv.info.Pace + mv.off
	if mv.once {
		lane := mv.lanes[(t/per)%len(mv.lanes)]
		head := t % per
		for j := 0; j < mv.cars; j++ {
			if i := head - j; i >= 0 && i < len(lane) {
				fn(lane[i], j, 1)
			}
		}
		return
	}
	lane, span := mv.lanes[0], max(0, len(mv.lanes[0])-mv.cars)
	t %= per
	s, dir := 0, 1
	switch {
	case t < mv.dwell:
	case t < mv.dwell+span:
		s = t - mv.dwell
	case t < 2*mv.dwell+span:
		s, dir = span, -1
	default:
		s, dir = span-(t-2*mv.dwell-span), -1
	}
	for j := 0; j < mv.cars && j < len(lane); j++ {
		i := s + j // dir < 0: the front is the low end
		if dir > 0 {
			i = s + min(mv.cars, len(lane)) - 1 - j
		}
		fn(lane[i], j, dir)
	}
}

// at reports whether the mover covers p at frame f.
func (mv *mover) at(f int, p mapmodel.Pt) bool {
	hit := false
	mv.cells(f, func(q mapmodel.Pt, _, _ int) { hit = hit || q == p })
	return hit
}

// Traffic limits: at most this many movers of one kind, and street movers in
// all, so the traffic livens the streets without burying the buildings.
const (
	perKindMax   = 3
	streetMax    = 10
	roadsPerCart = 24 // road cells for each street mover
)

// addFn adds mover number i of kind k on its lanes.
type addFn func(k mapmodel.Mover, lanes [][]mapmodel.Pt, i int, once bool, dwell int)

// traffic plans the age's movers on the scene's lanes.
func (s *scene) traffic() {
	m, w := s.m, s.w
	busy := m.Activity.Traffic
	add := func(k mapmodel.Mover, lanes [][]mapmodel.Pt, i int, once bool, dwell int) {
		if len(lanes) == 0 {
			return
		}
		info := k.Info()
		mv := mover{k: k, info: info, lanes: lanes, cars: info.Cars, dwell: dwell, once: once, n: i}
		for _, l := range lanes {
			mv.long = max(mv.long, len(l))
		}
		if !once && len(lanes[0]) <= mv.cars {
			return
		}
		if p := mv.period(); p > 0 {
			mv.off = int(mapmodel.Hash(w.Seed, 130, int64(k), int64(i)) % uint64(p))
		}
		s.movers = append(s.movers, mv)
	}
	var street, water []mapmodel.Mover
	for _, k := range mapmodel.MoversAt(m.AgeIdx) {
		switch k.Info().Way {
		case mapmodel.WayStreet:
			street = append(street, k)
		case mapmodel.WayWater:
			water = append(water, k)
		case mapmodel.WayFoot:
			if k == mapmodel.MoverHunter {
				for i, l := range s.hunts(min(3, 1+m.Workers.Pop/40)) {
					add(k, [][]mapmodel.Pt{l}, i, false, 6)
				}
			}
		case mapmodel.WayRail:
			// One train to a line (it is a single track), two cars longer in
			// a busy town; it waits a while at each end of the line.
			if len(s.rail) > 8 {
				add(k, [][]mapmodel.Pt{s.rail}, 0, false, 30)
				if busy > 0.6 {
					s.movers[len(s.movers)-1].cars += 2
				}
			}
		case mapmodel.WaySky:
			s.skyMovers(k, busy, add)
		}
	}
	s.streetMovers(street, min(perKindMax, 1+int(float64(busy*2.5))), add)
	s.waterMovers(water, busy, add)
}

// streetMovers shares the street network among the age's street movers,
// the age's newest first, a few of each and not too many in all.
func (s *scene) streetMovers(kinds []mapmodel.Mover, per int, add addFn) {
	if len(kinds) == 0 {
		return
	}
	lanes := s.streetLanes()
	if len(lanes) == 0 {
		return
	}
	budget := min(streetMax, max(1, s.roads/roadsPerCart))
	count := make([]int, len(kinds))
	for placed, round := 0, 0; placed < budget && round < per; round++ {
		for i := len(kinds) - 1; i >= 0 && placed < budget; i-- {
			count[i]++
			placed++
		}
	}
	lane := 0
	for i, k := range kinds {
		for n := 0; n < count[i]; n++ {
			add(k, [][]mapmodel.Pt{lanes[lane%len(lanes)]}, n, false, 2)
			lane++
		}
	}
}

// waterMovers puts the age's boats on the water: rowboats on the river (or
// the shallows), ships on open water (or, failing that, the river; a
// container ship needs open water).
func (s *scene) waterMovers(kinds []mapmodel.Mover, busy float64, add addFn) {
	if len(kinds) == 0 {
		return
	}
	river := s.waterLanes(func(t mapmodel.Terrain) bool { return t == mapmodel.TRiver }, 3, 8, 140)
	sea := s.waterLanes(func(t mapmodel.Terrain) bool { return t == mapmodel.TDeep || t == mapmodel.TShallow }, 3, 14, 141)
	for _, k := range kinds {
		lanes, crowded := sea, false
		switch {
		case k == mapmodel.MoverRowboat && len(river) > 0:
			lanes = river
		case len(sea) == 0 && k != mapmodel.MoverBoxShip:
			lanes, crowded = river, true // ships up the river: one of each
		}
		n := 1
		if busy > 0.6 && len(lanes) > 1 && !crowded {
			n = 2
		}
		for i := 0; i < n && i < len(lanes); i++ {
			add(k, [][]mapmodel.Pt{lanes[i]}, i, false, 8)
		}
	}
}

// skyMovers puts the age's aircraft and spacecraft overhead: planes on
// straight lines across the map, drones over the streets, satellites high
// and slow, an orbital habitat drifting by, shuttles climbing from the pad.
func (s *scene) skyMovers(k mapmodel.Mover, busy float64, add addFn) {
	w := s.w
	line := func(salt int64, slope bool) []mapmodel.Pt {
		h := mapmodel.Hash(w.Seed, salt)
		y0 := w.CY - 16 + int(h%33)
		dy := 0
		if slope {
			dy = int(h>>8%23) - 11
		}
		x0, x1 := max(0, w.CX-110), min(w.W-1, w.CX+110)
		out := make([]mapmodel.Pt, 0, x1-x0+1)
		for x := x0; x <= x1; x++ {
			y := y0 + dy*(x-x0)/max(1, x1-x0)
			out = append(out, pt(x, clamp(y, 0, w.H-1)))
		}
		if h>>16%2 == 0 {
			return reversed(out)
		}
		return out
	}
	switch k {
	case mapmodel.MoverPlane:
		for i := 0; i < 1+int(float64(busy*1.5)); i++ {
			var ls [][]mapmodel.Pt
			for v := 0; v < 4; v++ {
				ls = append(ls, line(int64(200+i*8+v), true))
			}
			add(k, ls, i, true, 300+int(mapmodel.Hash(w.Seed, 210, int64(i))%300))
		}
	case mapmodel.MoverSatellite:
		for i := 0; i < 2; i++ {
			var ls [][]mapmodel.Pt
			for v := 0; v < 3; v++ {
				ls = append(ls, line(int64(220+i*8+v), true))
			}
			add(k, ls, i, true, 60+int(mapmodel.Hash(w.Seed, 230, int64(i))%120))
		}
	case mapmodel.MoverHabitat:
		add(k, [][]mapmodel.Pt{line(240, false)}, 0, true, 40)
	case mapmodel.MoverShuttle:
		if pad, ok := s.launchPad(); ok {
			var l []mapmodel.Pt
			for y := pad.Y - 1; y >= max(0, pad.Y-34); y-- {
				l = append(l, pt(pad.X, y))
			}
			add(k, [][]mapmodel.Pt{l}, 0, true, 400+int(mapmodel.Hash(w.Seed, 250)%300))
		}
	case mapmodel.MoverDrone:
		if lanes := s.streetLanes(); len(lanes) > 0 {
			for i := 0; i < 1+int(float64(busy*2)); i++ {
				add(k, [][]mapmodel.Pt{lanes[(i*5+3)%len(lanes)]}, i, false, 3)
			}
		}
	}
}

// launchPad is where shuttles lift off: the launch complex, else a space
// force base or the space program, else open ground outside the town.
func (s *scene) launchPad() (mapmodel.Pt, bool) {
	for _, key := range []string{"launch_complex", "space_force_base"} {
		for _, t := range s.m.Town.Tiles {
			if t.Key == key && t.Anchor && !t.Ruin {
				return t.Pt, true
			}
		}
	}
	for _, wd := range s.m.Town.Wonders {
		if wd.Key == "space_program" && wd.Built {
			return pt(wd.At.X+1, wd.At.Y), true
		}
	}
	w := s.w
	a, r := mapmodel.HashF(w.Seed, 251), math.Max(s.wallR, 4)+5
	c := pt(w.CX+int(float64(mapmodel.Cos(a)*r)*2), w.CY+int(float64(mapmodel.Sin(a)*r)))
	for rad := 0; rad <= 5; rad++ {
		for dy := -rad; dy <= rad; dy++ {
			for dx := -2 * rad; dx <= 2*rad; dx++ {
				if q := pt(c.X+dx, c.Y+dy); s.in(q.X, q.Y) && s.at(q.X, q.Y).k == kNone && s.w.At(q.X, q.Y).Land() {
					return q, true
				}
			}
		}
	}
	return mapmodel.Pt{}, false
}

// hunts are the hunters' trails: from a home out to the woods and hills
// beyond the fields, and back.
func (s *scene) hunts(n int) [][]mapmodel.Pt {
	w, m := s.w, s.m
	var homes, wilds []mapmodel.Pt
	for _, t := range m.Town.Tiles {
		if t.Lineage == mapmodel.LinHousing && !t.Ruin {
			homes = append(homes, t.Pt)
		}
	}
	if len(homes) == 0 {
		homes = []mapmodel.Pt{pt(w.CX, w.CY)}
	}
	for y := w.CY - 14; y <= w.CY+14; y++ {
		for x := w.CX - 30; x <= w.CX+30; x++ {
			d := rdist(x, y, w.CX, w.CY)
			if t := w.At(x, y); (t == mapmodel.TForest || t == mapmodel.THills) && d >= 8 && d <= 15 && s.at(x, y).k == kNone {
				wilds = append(wilds, pt(x, y))
			}
		}
	}
	if len(wilds) == 0 {
		return nil
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
	var out [][]mapmodel.Pt
	for i := 0; i < n; i++ {
		h := homes[mapmodel.Hash(w.Seed, 160, int64(i))%uint64(len(homes))]
		t := wilds[mapmodel.Hash(w.Seed, 161, int64(i))%uint64(len(wilds))]
		if p := w.Path(h, t, cost); len(p) > 4 {
			out = append(out, p)
		}
	}
	return out
}

func reversed(ps []mapmodel.Pt) []mapmodel.Pt {
	out := make([]mapmodel.Pt, len(ps))
	for i, p := range ps {
		out[len(ps)-1-i] = p
	}
	return out
}

// streetLanes are the lanes on the street network near the town (found once
// per scene): the first runs end to end of the network, the rest from a road
// cell picked by hash to the farthest one it reaches.
func (s *scene) streetLanes() [][]mapmodel.Pt {
	if s.lanesDone {
		return s.lanes
	}
	s.lanesDone = true
	w := s.w
	R := clamp(int(s.wallR)+12, 14, 36)
	pass := func(x, y int) bool {
		k := s.at(x, y).k
		return (k == kStreet || k == kRoad || k == kBridge || k == kPlaza) && rdist(x, y, w.CX, w.CY) <= float64(R)
	}
	n := 0
	for y := w.CY - R; y <= w.CY+R; y++ {
		for x := w.CX - 2*R; x <= w.CX+2*R; x++ {
			if s.in(x, y) && pass(x, y) {
				n++
			}
		}
	}
	s.roads = n
	if n < 8 {
		return nil
	}
	s.lanes = s.lanesOn(w.CX-2*R, w.CY-R, w.CX+2*R, w.CY+R, pass, min(12, max(2, n/20)), 6, 170)
	return s.lanes
}

// waterLanes are up to n lanes on the biggest body of water of the kinds ok
// near the town, each at least least cells long.
func (s *scene) waterLanes(ok func(mapmodel.Terrain) bool, n, least int, salt int64) [][]mapmodel.Pt {
	w := s.w
	pass := func(x, y int) bool {
		k := s.at(x, y).k
		return ok(w.At(x, y)) && (k == kNone || k == kBridge)
	}
	return s.lanesOn(w.CX-80, w.CY-26, w.CX+80, w.CY+26, pass, n, least, salt)
}

// lanesOn finds up to n lanes through the cells pass allows inside the box
// (x0, y0)-(x1, y1), all on the biggest connected stretch: the first from
// end to end of it (from its first cell to the farthest, then from there to
// the farthest again), the rest each from a cell picked by hash to the
// farthest it reaches. Lanes shorter than least are dropped. Breadth-first
// and deterministic: the same scene always finds the same lanes.
func (s *scene) lanesOn(x0, y0, x1, y1 int, pass func(x, y int) bool, n, least int, salt int64) [][]mapmodel.Pt {
	w := s.w
	x0, y0, x1, y1 = max(0, x0), max(0, y0), min(w.W-1, x1), min(w.H-1, y1)
	bw, bh := x1-x0+1, y1-y0+1
	if bw <= 0 || bh <= 0 {
		return nil
	}
	ok := make([]bool, bw*bh)
	for i := range ok {
		ok[i] = pass(x0+i%bw, y0+i/bw)
	}
	comp := make([]int32, bw*bh)
	prev := make([]int32, bw*bh)
	queue := make([]int32, 0, 256)
	// bfs walks from start over ok cells, stamping comp with tag, and
	// returns the last cell it reached (the farthest) with prev filled in.
	bfs := func(start int, tag int32) int {
		queue = append(queue[:0], int32(start))
		comp[start], prev[start] = tag, -1
		far := start
		for q := 0; q < len(queue); q++ {
			i := int(queue[q])
			far = i
			x, y := i%bw, i/bw
			for _, d := range dirs4 {
				nx, ny := x+d[0], y+d[1]
				if nx < 0 || ny < 0 || nx >= bw || ny >= bh {
					continue
				}
				if j := ny*bw + nx; ok[j] && comp[j] != tag {
					comp[j], prev[j] = tag, int32(i)
					queue = append(queue, int32(j))
				}
			}
		}
		return far
	}
	// the biggest stretch
	best, bestN, tag := -1, 0, int32(0)
	for i := range ok {
		if ok[i] && comp[i] == 0 {
			tag++
			bfs(i, tag)
			if len(queue) > bestN {
				best, bestN = i, len(queue)
			}
		}
	}
	if best < 0 || bestN < least {
		return nil
	}
	tag++
	bfs(best, tag)
	members := append([]int32(nil), queue...)
	path := func(start int) []mapmodel.Pt {
		tag++
		far := bfs(start, tag)
		var out []mapmodel.Pt
		for i := far; i >= 0; i = int(prev[i]) {
			out = append(out, pt(x0+i%bw, y0+i/bw))
		}
		return out
	}
	var out [][]mapmodel.Pt
	a := path(best)
	if l := path(pt2i(a[0], x0, y0, bw)); len(l) >= least {
		out = append(out, l)
	}
	for k := 1; k < n; k++ {
		st := int(members[mapmodel.Hash(w.Seed, salt, int64(k))%uint64(len(members))])
		if l := path(st); len(l) >= least {
			out = append(out, l)
		}
	}
	return out
}

func pt2i(p mapmodel.Pt, x0, y0, bw int) int { return (p.Y-y0)*bw + p.X - x0 }

// railway lays the line the trains run on, from the Industrial Age: a
// straight track along one of the town plan's street rows, the first in a
// seeded order (two to four blocks out from the square, then farther, then
// nearer) that crosses no wonder plot, the square or a civ's site. Street
// rows never take buildings, and the choice reads only the seed's world and
// plan, so the line never moves as the town grows. It runs out either way to
// the edge of the map or to open sea, and marks its cells with the rail flag
// without changing what they are (a street it crosses stays a street).
func (s *scene) railway() {
	rails := false
	for _, k := range mapmodel.MoversAt(s.m.AgeIdx) {
		rails = rails || k.Info().Way == mapmodel.WayRail
	}
	if !rails {
		return
	}
	w, p := s.w, s.m.Town.Plan
	var near, mid, far []int
	for j := -6; j <= 7; j++ {
		y := w.CY - 1 + 3*j
		switch d := abs(y - w.CY); {
		case d < 4:
		case d <= 11:
			mid = append(mid, y)
		case d <= 17:
			far = append(far, y)
		default:
			near = append(near, y) // out of most screens: last resort
		}
	}
	h := mapmodel.Hash(w.Seed, 120)
	for i := len(mid) - 1; i > 0; i-- { // a seeded shuffle of the preferred rows
		j := int(mapmodel.Hash(int64(h), int64(i)) % uint64(i+1))
		mid[i], mid[j] = mid[j], mid[i]
	}
	for _, y := range append(append(mid, far...), near...) {
		if line, ok := s.railRow(p, y); ok {
			s.rail = line
			for _, q := range line {
				s.at(q.X, q.Y).rail = true
			}
			return
		}
	}
}

// railRow is the track along row y, or false if the row crosses something
// it must not.
func (s *scene) railRow(p *mapmodel.Plan, y int) ([]mapmodel.Pt, bool) {
	w := s.w
	if y < 2 || y >= w.H-2 {
		return nil, false
	}
	plot := func(x int) bool {
		for _, at := range p.WonderPlots {
			if x >= at.X && x < at.X+mapmodel.WonderW && y >= at.Y && y < at.Y+mapmodel.WonderH {
				return true
			}
		}
		for _, st := range w.Sites {
			if abs(st.X-x) <= 3 && abs(st.Y-y) <= 1 {
				return true
			}
		}
		k := s.at(x, y).k
		return p.Plaza[y*w.W+x] || k == kTile || k == kWonder || k == kSite || k == kCentre
	}
	// out from the middle each way, to the map's edge or open sea
	end := func(dir int) (int, bool) {
		last, deep := w.CX, 0
		for x := w.CX; x >= 1 && x < w.W-1; x += dir {
			if plot(x) {
				return 0, false
			}
			if w.At(x, y) == mapmodel.TDeep {
				if deep++; deep > 6 {
					break
				}
				continue
			}
			deep, last = 0, x
		}
		return last, true
	}
	xa, ok1 := end(-1)
	xb, ok2 := end(1)
	if !ok1 || !ok2 || xb-xa < 40 {
		return nil, false
	}
	line := make([]mapmodel.Pt, 0, xb-xa+1)
	for x := xa; x <= xb; x++ {
		line = append(line, pt(x, y))
	}
	return line, true
}

// railGlyph is the track: sleepers on a line, a sleek guideway once the
// maglev runs on it. Labels keep off it (salience 46), so the line reads
// unbroken at every zoom.
func (v *view) railGlyph(t mapmodel.Terrain) glyph {
	g := glyph{r: '╪', c: mapmodel.CWall, sal: 46}
	id := lgRail
	if mapmodel.MoverMaglev.Info().In(v.sc.m.AgeIdx) {
		g.r, g.c, id = '━', mapmodel.CCivic, lgGuideway
	}
	if t.Water() {
		g.bg = v.pal.WaterBg
	}
	v.regG(id, glyph{r: g.r, c: g.c})
	return g
}

// railName is what the railway is called in this age.
func (v *view) railName() string {
	if mapmodel.MoverMaglev.Info().In(v.sc.m.AgeIdx) {
		return "Maglev line"
	}
	return "Railway"
}

// railLine says what runs on the railway now.
func (v *view) railLine() string {
	for _, k := range mapmodel.MoversAt(v.sc.m.AgeIdx) {
		if i := k.Info(); i.Way == mapmodel.WayRail {
			return i.Name + "s run through " + v.sc.w.Name
		}
	}
	return "in " + v.sc.w.Name
}

// trafficPresent works out which movers' lanes cross the view (cached per
// camera), so their legend rows hold steady while they come and go.
func (v *view) trafficPresent() {
	s, g := v.sc, v.g
	key := [6]int{g.vx, g.vy, g.w, g.h, g.cellW, len(s.movers)}
	if v.present.sc == s && v.present.key == key {
		return
	}
	v.present.sc, v.present.key = s, key
	v.present.on = [mapmodel.NumMovers]bool{}
	for i := range s.movers {
		mv := &s.movers[i]
		if v.present.on[mv.k] {
			continue
		}
	lanes:
		for _, l := range mv.lanes {
			for _, p := range l {
				if _, _, ok := g.cellOf(p); ok && s.seen(p) == 2 {
					v.present.on[mv.k] = true
					break lanes
				}
			}
		}
	}
}

// shown reports whether a mover is out at all this frame (satellites keep
// to the night before the Cosmic Era).
func (v *view) shown(mv *mover) bool {
	return !mv.info.Night || v.sc.epoch >= 6 || v.sc.m.Clock.Night > 0.5
}

// drawTraffic draws the movers of one layer: the ground (streets, rails,
// water, paths) or the sky.
func (v *view) drawTraffic(cv *mapstyle.Canvas, sky bool) {
	s, f := v.sc, v.anim
	v.trafficPresent()
	for i := range s.movers {
		mv := &s.movers[i]
		if (mv.info.Way == mapmodel.WaySky) != sky || !v.shown(mv) {
			continue
		}
		id := lgMover + lgID(mv.k)
		r := mapmodel.R(mv.info.Sym, v.tier)
		if v.present.on[mv.k] {
			v.reg(id, r, v.style(glyph{r: r, c: mv.info.Class, attr: tcell.AttrBold}))
		}
		mv.cells(f, func(p mapmodel.Pt, j, dir int) {
			c := s.at(p.X, p.Y)
			if mv.info.Way == mapmodel.WayWater && (c.k != kNone || c.rail) {
				return // under a bridge
			}
			ch := r
			switch {
			case mv.cars > 1 && j == 0 && v.tier != mapmodel.TierNerd: // the engine leads
				ch = '►'
				if mv.east(f, dir) < 0 {
					ch = '◄'
				}
			case mv.cars > 1 && j > 0 && v.tier == mapmodel.TierNerd:
				ch = mapmodel.R(mv.info.Sym, mapmodel.TierUnicode)
			case mv.k == mapmodel.MoverPlane && v.tier == mapmodel.TierASCII && mv.east(f, dir) < 0:
				ch = '<'
			}
			v.onQuiet(cv, p, ch, mv.info.Class)
			if j == 0 {
				v.exhaust(cv, mv, p, mv.east(f, dir))
			}
		})
	}
}

// front is the mover's front cell at frame f, if it is out.
func (mv *mover) front(f int) (mapmodel.Pt, bool) {
	var p mapmodel.Pt
	ok := false
	mv.cells(f, func(q mapmodel.Pt, j, _ int) {
		if j == 0 {
			p, ok = q, true
		}
	})
	return p, ok
}

// east is 1 while the mover heads east and -1 while it heads west, from
// where its front was a step ago. A train reads it off its line (which runs
// west to east), so one waiting at the end faces the way it will leave.
func (mv *mover) east(f, dir int) int {
	if mv.info.Way == mapmodel.WayRail {
		return dir
	}
	now, ok := mv.front(f)
	was, ok2 := mv.front(f - mv.info.Pace)
	if ok && ok2 && now.X < was.X {
		return -1
	}
	return 1
}

// exhaust trails smoke behind a steam engine (heading east or west) and
// fire under a shuttle.
func (v *view) exhaust(cv *mapstyle.Canvas, mv *mover, p mapmodel.Pt, east int) {
	s := v.sc
	switch {
	case mv.info.Smoke:
		puffs := []rune(s.d.smoke)
		if len(puffs) == 0 {
			return
		}
		for k := 1; k <= 3; k++ {
			q := pt(p.X-east*(k+1), p.Y-1-(k-1)/2)
			if (v.anim/3+k)%4 == 0 {
				continue // the plume breaks up as it drifts
			}
			v.onQuiet(cv, q, puffs[(k-1)%len(puffs)], mapmodel.CMemory)
		}
	case mv.k == mapmodel.MoverShuttle:
		for k, r := range []rune{'░', '·'} {
			v.onQuiet(cv, pt(p.X, p.Y+1+k), r, mapmodel.CIdle)
		}
	}
}

// onQuiet draws one mover cell like on, without claiming a legend row
// (rows are claimed per kind, so they hold steady).
func (v *view) onQuiet(cv *mapstyle.Canvas, p mapmodel.Pt, r rune, c mapmodel.Class) {
	s := v.sc
	k := s.at(p.X, p.Y).k
	cx, cy, ok := v.g.cellOf(p)
	if !ok || s.seen(p) < 2 || k == kTile || k == kWonder || k == kSite || k == kCentre {
		return
	}
	_, cur := cv.Get(cx, cy)
	_, bg, _ := cur.Decompose()
	cv.Put(cx, cy, r, v.style(glyph{r: r, c: c, bg: bg, attr: tcell.AttrBold}))
}

// moverAt is the mover (or worker) on p at this frame, topmost first. Only
// what is drawn counts: nothing on a building, no boat under a bridge.
func (v *view) moverAt(p mapmodel.Pt) (insp, bool) {
	s, f := v.sc, v.anim
	if k := s.at(p.X, p.Y).k; k == kTile || k == kWonder || k == kSite || k == kCentre {
		return insp{}, false
	}
	for pass := 0; pass < 2; pass++ {
		for i := len(s.movers) - 1; i >= 0; i-- {
			mv := &s.movers[i]
			if (mv.info.Way == mapmodel.WaySky) == (pass == 0) && v.shown(mv) && mv.at(f, p) {
				if mv.info.Way == mapmodel.WayWater && (s.at(p.X, p.Y).k != kNone || s.at(p.X, p.Y).rail) {
					continue
				}
				return insp{Title: mv.info.Title, Lines: []string{mv.info.Line(mv.n)}, Kind: mapstyle.KindMover}, true
			}
		}
	}
	for i, path := range s.walks {
		if path[walkAt(i, len(path), f)] == p {
			w := mapmodel.MoverWalker.Info()
			return insp{Title: w.Title, Lines: []string{w.Line(i)}, Kind: mapstyle.KindMover}, true
		}
	}
	return insp{}, false
}

// walkAt is the cell index of worker i on a path of n cells at frame f:
// there and back again, a cell every walkFrames frames.
func walkAt(i, n, f int) int {
	if n < 2 {
		return 0
	}
	t := (f/walkFrames + int(mapmodel.Hash(int64(i), 3)%uint64(2*n-2))) % (2*n - 2)
	return min(t, 2*n-2-t)
}
