package roguelike

import (
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
)

// citylines.go lays the lines over the Earth arc's city: the highways out
// of the Modern town, the Information Age's fiber, the megacity's sky
// rails (maglev lines once fusion powers the city), the Fusion Age's power
// conduits and the space elevator's tether. A line keeps its points in
// order, so its traffic can ride it; it draws only where it may (a ground
// line on open land, an elevated one over streets and the railway too, and
// none over a building, a wonder, a civ or the square).

// addLine lays a line: each cell it may draw on takes its mask (the line's
// neighbours along it), and the line shows there unless a higher one does.
func (s *scene) addLine(cl *cityLand, ln cityLine) {
	if len(ln.pts) < 2 || len(cl.lines) >= 250 {
		return
	}
	w := s.w
	cl.lines = append(cl.lines, ln)
	id := uint8(len(cl.lines))
	may := func(p mapmodel.Pt) bool {
		if !w.In(p.X, p.Y) {
			return false
		}
		i := p.Y*w.W + p.X
		switch k := s.cells[i].k; {
		case k == kTile || k == kWonder || k == kSite || k == kCentre:
			return false
		case ln.elev:
			return true
		case k != kNone || s.cells[i].rail:
			return false
		case ln.feat == mapmodel.FeatFiber:
			return cl.use[i] == uStreet
		}
		return w.T[i].Land() || ln.feat == mapmodel.FeatHighway // highways bridge the water
	}
	bit := func(a, b mapmodel.Pt) uint8 {
		switch {
		case b.Y < a.Y:
			return 1
		case b.X > a.X:
			return 2
		case b.Y > a.Y:
			return 4
		}
		return 8
	}
	for j, p := range ln.pts {
		if !may(p) {
			continue
		}
		i := p.Y*w.W + p.X
		var m uint8
		if j > 0 {
			m |= bit(p, ln.pts[j-1])
		}
		if j+1 < len(ln.pts) {
			m |= bit(p, ln.pts[j+1])
		}
		cl.ovm[i] |= m
		if top := cl.ov[i]; top == 0 || cl.lines[top-1].layer <= ln.layer {
			cl.ov[i] = id
		}
		if ln.feat != mapmodel.FeatFiber {
			cl.mark[i] = true
		}
	}
}

// run walks from (x, y) one cell at a time in direction (dx, dy) to the
// map's edge or open sea (more than six cells of deep water), and returns
// the cells it crossed.
func (s *scene) run(x, y, dx, dy int) []mapmodel.Pt {
	w := s.w
	var out []mapmodel.Pt
	deep := 0
	for w.In(x, y) {
		if w.At(x, y) == mapmodel.TDeep {
			if deep++; deep > 6 {
				break
			}
		} else {
			deep = 0
		}
		out = append(out, pt(x, y))
		x, y = x+dx, y+dy
	}
	// leave no deep water dangling at the open end
	for len(out) > 0 && w.At(out[len(out)-1].X, out[len(out)-1].Y) == mapmodel.TDeep {
		out = out[:len(out)-1]
	}
	return out
}

// railRowY is the railway's row, or -1 before the railway is laid.
func (s *scene) railRowY() int {
	if len(s.rail) > 0 {
		return s.rail[0].Y
	}
	return -1
}

// highways runs the Modern Age's highways out of town: east and west along
// a street row by the square, north and south down the clearest column,
// each from the ring boulevard to the edge of the map.
func (s *scene) highways(cl *cityLand) {
	w := s.w
	rail := s.railRowY()
	yh := -1
	for _, k := range []int{0, 1, -1, 2, -2, 3} {
		if y := w.CY - 1 + 3*k; y != rail && y > 1 && y < w.H-2 {
			yh = y
			break
		}
	}
	out := s.wallR + 0.5
	start := func(dx, dy, x, y int) (int, int) {
		for w.In(x, y) && rdist(x, y, w.CX, w.CY) < out {
			x, y = x+dx, y+dy
		}
		return x, y
	}
	add := func(pts []mapmodel.Pt) {
		if len(pts) >= 8 {
			s.addLine(cl, cityLine{feat: mapmodel.FeatHighway, pts: pts, ink: theme.CityAsphalt, layer: 1})
		}
	}
	if yh >= 0 {
		x, y := start(1, 0, w.CX, yh)
		add(s.run(x, y, 1, 0))
		x, y = start(-1, 0, w.CX, yh)
		add(s.run(x, y, -1, 0))
	}
	// north and south: the column near the square with the fewest
	// buildings in its way
	best, bx := 1<<30, w.CX
	for k := 0; k <= 12; k++ {
		dx := (k + 1) / 2
		if k%2 == 1 {
			dx = -dx
		}
		x := w.CX + dx
		cost := 0
		for y := 0; y < w.H; y++ {
			if rdist(x, y, w.CX, w.CY) < out {
				continue
			}
			if c := s.at(x, y).k; c == kTile || c == kWonder || c == kSite {
				cost++
			}
		}
		if cost*4+abs(dx) < best {
			best, bx = cost*4+abs(dx), x
		}
	}
	x, y := start(0, -1, bx, w.CY)
	add(s.run(x, y, 0, -1))
	x, y = start(0, 1, bx, w.CY)
	add(s.run(x, y, 0, 1))
}

// fiber lays the Information Age's fiber lines: four dashed lines along
// street rows, from the ring boulevard out to the city's edge.
func (s *scene) fiber(cl *cityLand) {
	w := s.w
	rail := s.railRowY()
	reach := s.cityRadius()
	ks := []int{2, -2, 4, -4, 6, -6, 3, -3}
	h := mapmodel.Hash(w.Seed, 340)
	n := 0
	for j, k := range ks {
		if n == 4 {
			break
		}
		y := w.CY - 1 + 3*k
		if y == rail || y < 1 || y >= w.H-1 {
			continue
		}
		dx := 1
		if (h>>uint(j))%2 == 0 {
			dx = -1
		}
		x := w.CX
		for w.In(x, y) && rdist(x, y, w.CX, w.CY) < s.wallR+1 {
			x += dx
		}
		var pts []mapmodel.Pt
		for w.In(x, y) && rdist(x, y, w.CX, w.CY) < reach {
			pts = append(pts, pt(x, y))
			x += dx
		}
		if len(pts) >= 8 {
			s.addLine(cl, cityLine{feat: mapmodel.FeatFiber, pts: pts, ink: theme.CityData})
			n++
		}
	}
}

// railInks are the sky rails' colours and names; the maglev lines' once
// fusion powers the city.
var (
	skyRailInks = []struct {
		ink  theme.CityHue
		name string
	}{{theme.CityNeonMagenta, "magenta"}, {theme.CityNeonCyan, "cyan"}, {theme.CityNeonYellow, "yellow"},
		{theme.CityTowerLit, "violet"}, {theme.CityPlasma, "white"}}
	maglevInks = []struct {
		ink  theme.CityHue
		name string
	}{{theme.CityPlasma, "white"}, {theme.CityElectric, "blue"}, {theme.CityConduit, "azure"},
		{theme.CityNeonCyan, "cyan"}, {theme.CityReactor, "silver"}}
)

// skyLineCount is how many elevated lines the megacity runs: three, up to
// five as the traffic grows.
func skyLineCount(m *mapmodel.Model) int {
	return min(5, 3+int(float64(m.Activity.Traffic*2.2)))
}

// skyRails lays the megacity's elevated lines, three to five crossing in
// layers: rows along the street grid, west to east, and straight columns
// north to south, each a colour and a name ("the magenta line"). Once
// fusion powers the city they are maglev lines.
func (s *scene) skyRails(cl *cityLand, fusion bool) {
	w := s.w
	n := skyLineCount(s.m)
	rail := s.railRowY()
	// candidate rows, two blocks out and more, in a seeded order
	var rows []int
	for k := 2; k <= 7; k++ {
		for _, sg := range [2]int{1, -1} {
			if y := w.CY - 1 + 3*k*sg; y > 1 && y < w.H-2 && abs(y-rail) > 2 {
				rows = append(rows, y)
			}
		}
	}
	h := mapmodel.Hash(w.Seed, 350)
	for i := len(rows) - 1; i > 0; i-- {
		j := int(mapmodel.Hash(int64(h), int64(i)) % uint64(i+1))
		rows[i], rows[j] = rows[j], rows[i]
	}
	inks, feat := skyRailInks, mapmodel.FeatSkyRail
	if fusion {
		inks, feat = maglevInks, mapmodel.FeatMaglevLine
	}
	used := []int{}
	far := func(y int) bool {
		for _, u := range used {
			if abs(u-y) < 6 {
				return false
			}
		}
		return true
	}
	ri, side := 0, int(h>>20%2)*2-1
	for k := 0; k < n; k++ {
		ln := cityLine{feat: feat, ink: inks[k%len(inks)].ink, elev: true, layer: uint8(2 + k),
			name: "the " + inks[k%len(inks)].name + " line"}
		if k%2 == 0 { // a row, west to east through the town's latitude
			for ri < len(rows) && !far(rows[ri]) {
				ri++
			}
			if ri >= len(rows) {
				continue
			}
			y := rows[ri]
			used = append(used, y)
			west := s.run(w.CX, y, -1, 0)
			east := s.run(w.CX+1, y, 1, 0)
			ln.pts = append(reversed(west), east...)
		} else { // a column, north to south, east and west of the town by turns
			off := int(2*s.wallR) + 6 + 10*(k/2)
			x := clamp(w.CX+side*off, 2, w.W-3)
			side = -side
			north := s.run(x, w.CY, 0, -1)
			south := s.run(x, w.CY+1, 0, 1)
			ln.pts = append(reversed(north), south...)
		}
		s.addLine(cl, ln)
	}
}

// conduits run the Fusion Age's power conduits from each reactor in to the
// ring boulevard: along the reactor's row toward the town, then up or down
// its column to the ring.
func (s *scene) conduits(cl *cityLand) {
	w := s.w
	ring := s.wallR + 0.5
	for _, r := range cl.reactors {
		dx := 1
		if r.X > w.CX {
			dx = -1
		}
		x, y := r.X+dx*5, r.Y
		var pts []mapmodel.Pt
		for w.In(x, y) && x != w.CX && rdist(x, y, w.CX, w.CY) > ring {
			pts = append(pts, pt(x, y))
			x += dx
		}
		dy := 1
		if y > w.CY {
			dy = -1
		}
		for w.In(x, y) && y != w.CY && rdist(x, y, w.CX, w.CY) > ring {
			pts = append(pts, pt(x, y))
			y += dy
		}
		if len(pts) >= 3 {
			s.addLine(cl, cityLine{feat: mapmodel.FeatConduit, pts: pts, ink: theme.CityConduit, layer: 1})
		}
	}
}

// tetherLine raises the space elevator: a tether from the square straight
// up to the top of the map, over everything but the buildings.
func (s *scene) tetherLine(cl *cityLand) {
	w := s.w
	var pts []mapmodel.Pt
	for y := w.CY - 1; y >= 0; y-- {
		pts = append(pts, pt(w.CX, y))
	}
	if len(pts) < 2 {
		return
	}
	s.addLine(cl, cityLine{feat: mapmodel.FeatTether, pts: pts, ink: theme.CityTether, elev: true, layer: 9,
		name: "the tether"})
}
