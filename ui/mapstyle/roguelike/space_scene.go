package roguelike

import (
	"math"
	"sort"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/ui/mapstyle"
)

// space_scene.go lays a sky scene out on a fixed plane, so placement never
// depends on the terminal: skyW x skyH cells, the size of the map area of a
// 160x48 view with its legend. A bigger view shows more of the starfield
// round it, a smaller one scrolls with the cursor.
//
// A scene is built in two layers. The base (skyBase) is everything that
// depends on the seed and the scene only: the backdrop, the structures and
// every lineage's reserved slots, nearest the structure's heart first. The
// scene (skyScene) puts the model in it: each lineage's units take its
// slots in order (unit n always lands on slot n, mapmodel.SkyLineages), and
// the scene adds what changes with play: city lights, civs, wonders, the
// lanes the traffic runs. Every cell is a glyph that means something.

const (
	skyW = 132
	skyH = 42
)

// skyKind is what a cell is.
type skyKind uint8

const (
	skVoid   skyKind = iota
	skStar           // a star
	skWonder         // a wonder's star (ref: index into Model.Wonders)
	skPlanet         // the homeworld's face (Space)
	skLimb           // the atmosphere at the planet's limb
	skCity           // city lights (ref: lit tiles)
	skCiv            // a civ (ref: faction index)
	skTether         // the space elevator
	skFrame          // a structure (ref: index into skyBase.frames)
	skUnit           // a building's unit (ref: index into Town.Tiles)
	skSlot           // an empty berth of a structure (ref: frame)
	skRock           // a rock in the belt
	skMoon           // the moon
	skNebula         // a nebula wisp or cloud
	skOrbit          // an orbit line of the orrery
	skSun            // the home sun
	skColony         // a colony world (ref: frame)
	skBeacon         // a colony beacon (ref: frame)
	skLane           // a trade lane (ref: faction index)
	skSpiral         // the galaxy's arms
	skCloud          // a probability cloud
	skEcho           // a time echo of the old town (ref: index into Town.Tiles)
	skRing           // a mandala ring (ref: era)
	skMark           // a mandala mark (ref: ring<<10 | mark)
	skCore           // the heart of the mandala or the recursion
	skHub            // the heart of a scene: the station hub, the starbase core
	skField          // inside the finished warp gate
	numSkyKinds
)

// skyFx is what a cell does from frame to frame.
type skyFx uint8

const (
	fxNone    skyFx = iota
	fxTwinkle       // stars: now and then a glint
	fxBlink         // beacons and running lights
	fxCloud         // the planet's clouds drift over it
	fxSwirl         // the finished gate's field turns
	fxSpark         // a welding spark at a construction front
	fxFlicker       // superposition: the cell shows r, then alt
	fxFringe        // a probability cloud's fringes shimmer
	fxBreathe       // the mandala breathes
	fxIri           // iridescence runs along a structure
	fxSparkle       // transporter sparkles
	fxWisp          // nebula wisps drift
)

// skyCell is one plane cell.
type skyCell struct {
	r    rune
	sym  mapmodel.Sym // a model symbol, drawn in the frame's tier (overrides r)
	alt  rune         // superposition's other state
	ink  mapmodel.SkyInk
	bg   mapmodel.SkyInk // InkVoid: no background of its own
	lv   uint8           // brightness 0 faint, 1 dim, 2 normal, 3 bright
	fx   skyFx
	k    skyKind
	bold bool
	ph   uint8 // the effect's phase
	soft bool  // the background is a faint tint, not a surface
	ref  int32
}

// salience ranks a cell for down-sampling (the mini map) and labels.
func (c *skyCell) salience() int {
	switch c.k {
	case skHub, skCore:
		return 95
	case skCiv:
		return 90
	case skUnit, skMark:
		return 60
	case skWonder:
		return 58
	case skTether, skFrame, skRing:
		return 50
	case skCity, skBeacon, skColony:
		return 45
	case skSlot, skSun, skMoon:
		return 40
	case skLimb:
		return 35
	case skPlanet, skRock, skLane, skField:
		return 20
	case skNebula, skSpiral, skCloud, skEcho, skOrbit:
		return 8
	case skStar:
		return 5
	}
	return 0
}

// skyFrame names a structure for the inspector and the legend.
type skyFrame struct {
	name, line string
	lg         skyLgID
	lin        string // the lineage whose units it holds, "" for none
}

// skyBase is a scene's seed-only layer.
type skyBase struct {
	seed  int64
	sky   mapmodel.SkyScene
	cells []skyCell
	hub   mapmodel.Pt
	// slots are each lineage's reserved cells, the heart of its structure
	// first. bands cut a lineage's slots into the groups a structure grows
	// by (an outer ring's sector, a slip): band j is slots[bands[j-1]:
	// bands[j]], and the scene draws a band's frame once a unit stands in
	// it.
	slots  map[string][]mapmodel.Pt
	bands  map[string][]int
	frames []skyFrame
	// the scene's own data
	planet     planetGeom    // Space: the homeworld below
	gate       []mapmodel.Pt // Interstellar: the warp gate's ring, in build order
	gateStroke []rune        // and its strokes
	colony     map[string]colonySite
	systems    []mapmodel.Pt // Galactic: star system sites, by faction slot
}

// skyScene is one model laid out on its base.
type skyScene struct {
	m       *mapmodel.Model
	b       *skyBase
	sky     mapmodel.SkyScene
	cells   []skyCell
	hub     mapmodel.Pt
	targets []mapmodel.Pt
	labels  []skyLabel
	movers  []skyMover
	units   int // units placed
	// unitAt maps a building key to its first unit's cell (inspect, labels)
	anchor map[string]mapmodel.Pt
	// the mandala (Transcendent): the eras' rings, the crown, the layout
	rings []mapmodel.MandalaRing
	crown []mapmodel.MandalaMark
	md    mapstyle.MandalaGeom
}

// skyLabel is a name the close zoom may print.
type skyLabel struct {
	p    mapmodel.Pt
	text string
	ink  mapmodel.SkyInk
	rel  int // a civ's relation class + 1, 0 for none
	far  bool
}

func (b *skyBase) in(x, y int) bool { return x >= 0 && y >= 0 && x < skyW && y < skyH }

func (b *skyBase) at(x, y int) *skyCell {
	if !b.in(x, y) {
		return &skyCell{}
	}
	return &b.cells[y*skyW+x]
}

func (s *skyScene) in(x, y int) bool { return x >= 0 && y >= 0 && x < skyW && y < skyH }

// at is the cell at (x, y); off the plane it is a star field of its own.
func (s *skyScene) at(x, y int) skyCell {
	if !s.in(x, y) {
		return offPlane(s.b.seed, s.sky, x, y)
	}
	return s.cells[y*skyW+x]
}

func (s *skyScene) cell(x, y int) *skyCell {
	if !s.in(x, y) {
		return &skyCell{}
	}
	return &s.cells[y*skyW+x]
}

// frame registers a structure and returns its index.
func (b *skyBase) frame(name, line string, lg skyLgID, lin string) int32 {
	b.frames = append(b.frames, skyFrame{name: name, line: line, lg: lg, lin: lin})
	return int32(len(b.frames) - 1)
}

// put sets a cell of the base.
func (b *skyBase) put(x, y int, c skyCell) {
	if b.in(x, y) {
		b.cells[y*skyW+x] = c
	}
}

// free reports whether a base cell holds only backdrop.
func (b *skyBase) free(x, y int) bool {
	if !b.in(x, y) {
		return false
	}
	switch b.cells[y*skyW+x].k {
	case skVoid, skStar, skNebula, skSpiral, skCloud, skEcho:
		return true
	}
	return false
}

// newSkyBase lays the seed-only layer of a scene.
func newSkyBase(m *mapmodel.Model, sky mapmodel.SkyScene) *skyBase {
	b := &skyBase{seed: m.Seed, sky: sky, cells: make([]skyCell, skyW*skyH), slots: map[string][]mapmodel.Pt{},
		bands: map[string][]int{}, colony: map[string]colonySite{}, hub: pt(skyW/2, skyH/2)}
	b.starfield()
	switch sky {
	case mapmodel.SkyOrbit:
		b.layOrbit(m)
	case mapmodel.SkyDeep:
		b.layDeep()
	case mapmodel.SkyGalaxy:
		b.layGalaxy()
	case mapmodel.SkyQuantum:
		b.layQuantum()
	case mapmodel.SkyMandala:
		b.layMandala()
	}
	return b
}

// starfield scatters the stars every scene starts from (the mandala keeps
// only a few faint ones).
func (b *skyBase) starfield() {
	for y := 0; y < skyH; y++ {
		for x := 0; x < skyW; x++ {
			b.cells[y*skyW+x] = starAt(b.seed, b.sky, x, y)
		}
	}
}

// starAt is the backdrop cell at (x, y): a star or the void. It is a pure
// function of the seed, so the field runs on past the plane's edge.
func starAt(seed int64, sky mapmodel.SkyScene, x, y int) skyCell {
	h := mapmodel.Hash(seed, 501, int64(x), int64(y))
	per := uint64(34)
	if sky == mapmodel.SkyMandala {
		per = 140
	}
	c := skyCell{r: ' ', k: skVoid}
	switch v := h % per; {
	case v == 0:
		c = skyCell{r: '·', ink: mapmodel.InkStar, lv: 2, k: skStar, fx: fxTwinkle, ph: uint8(h >> 20)}
	case v == 1:
		c = skyCell{r: '.', ink: mapmodel.InkStar, lv: 1, k: skStar}
	case v == 2 && h>>12%3 == 0:
		c = skyCell{r: '∙', ink: mapmodel.InkStarBright, lv: 2, k: skStar, fx: fxTwinkle, ph: uint8(h >> 24)}
	case v == 3 && h>>12%4 == 0:
		c = skyCell{r: '·', ink: mapmodel.InkStar, lv: 0, k: skStar}
	}
	return c
}

// offPlane is a cell beyond the plane: stars only.
func offPlane(seed int64, sky mapmodel.SkyScene, x, y int) skyCell { return starAt(seed, sky, x, y) }

// newSkyScene puts a model on its base.
func newSkyScene(m *mapmodel.Model, b *skyBase) *skyScene {
	s := &skyScene{m: m, b: b, sky: b.sky, cells: append([]skyCell(nil), b.cells...), hub: b.hub,
		anchor: map[string]mapmodel.Pt{}}
	if b.sky == mapmodel.SkyMandala {
		s.mandala()
	} else {
		s.placeUnits()
	}
	switch b.sky {
	case mapmodel.SkyOrbit:
		s.orbitLife()
	case mapmodel.SkyDeep:
		s.deepLife()
	case mapmodel.SkyGalaxy:
		s.galaxyLife()
	case mapmodel.SkyQuantum:
		s.quantumLife()
	case mapmodel.SkyMandala:
		s.mandalaLife()
	}
	s.wonderStars()
	s.findTargets()
	return s
}

// placeUnits puts every lineage's units on its slots, unit n on slot n.
// Fields are two tiles a copy on the ground; up here a bay is one unit, so
// food keeps every other tile. Units past a lineage's last slot wait, out
// of sight, for room the scene never needs in practice.
func (s *skyScene) placeUnits() {
	m := s.m
	for _, l := range m.SkyLineages() {
		slots := s.b.slots[l.Key]
		if len(slots) == 0 {
			continue
		}
		for _, u := range l.Units {
			n := u.N
			if l.Key == mapmodel.LinFood {
				if n%2 == 1 {
					continue
				}
				n /= 2
			}
			if n >= len(slots) {
				continue
			}
			p := slots[n]
			c := s.cell(p.X, p.Y)
			tt := &m.Town.Tiles[u.Tile]
			bg, soft := c.bg, c.soft // a unit stands on whatever floor it lands on
			*c = skyCell{sym: l.Part.Sym, ink: unitInk(s.sky, l.Key), lv: 2, k: skUnit, ref: int32(u.Tile), bg: bg, soft: soft}
			if r := s.unitRune(l.Key, n); r != 0 {
				c.sym, c.r = mapmodel.SymNone, r
			}
			if tt.Anchor {
				if _, ok := s.anchor[tt.Key]; !ok {
					s.anchor[tt.Key] = p
				}
			}
			s.units++
		}
	}
}

// unitRune lets a scene draw a lineage's units in its own art (0: the
// part's symbol).
func (s *skyScene) unitRune(lin string, n int) rune {
	switch {
	case s.sky == mapmodel.SkyOrbit && lin == mapmodel.LinEngineer:
		return 0
	case s.sky == mapmodel.SkyDeep && lin == mapmodel.LinEngineer:
		return gateRune(s.b, n)
	}
	return 0
}

// unitInk is the ink a lineage's units are drawn in.
func unitInk(sky mapmodel.SkyScene, lin string) mapmodel.SkyInk {
	switch sky {
	case mapmodel.SkyOrbit:
		switch lin {
		case mapmodel.LinFood, mapmodel.LinWood:
			return mapmodel.InkAccent2
		case mapmodel.LinEnergy:
			return mapmodel.InkAccent
		case mapmodel.LinMines, mapmodel.LinMetal:
			return mapmodel.InkLight
		case mapmodel.LinMilitary:
			return mapmodel.InkAccent3
		case mapmodel.LinEngineer:
			return mapmodel.InkFrameDim
		}
		return mapmodel.InkFrame
	case mapmodel.SkyDeep:
		switch lin {
		case mapmodel.LinEngineer:
			return mapmodel.InkFrame
		case mapmodel.LinEnergy:
			return mapmodel.InkAccent
		case mapmodel.LinHousing:
			return mapmodel.InkMoon
		case mapmodel.LinHacker:
			return mapmodel.InkAccent2
		case mapmodel.LinMilitary:
			return mapmodel.InkGlow
		}
		return mapmodel.InkLight
	case mapmodel.SkyGalaxy:
		switch lin {
		case mapmodel.LinMilitary, mapmodel.LinTrade, mapmodel.LinHarbor:
			return mapmodel.InkAccent
		case mapmodel.LinKnowledge, mapmodel.LinHacker, mapmodel.LinCulture:
			return mapmodel.InkAccent2
		case mapmodel.LinMines, mapmodel.LinWood, mapmodel.LinEnergy, mapmodel.LinMetal:
			return mapmodel.InkAccent3
		}
		return mapmodel.InkFrame
	case mapmodel.SkyQuantum:
		return mapmodel.InkFrame
	}
	return mapmodel.InkLight
}

// wonderStars hangs every wonder in the sky as a bright star, at a seeded
// spot on open space: the constellation of what the player has raised.
func (s *skyScene) wonderStars() {
	if s.sky == mapmodel.SkyMandala {
		return // the mandala carries them on its rings
	}
	for i, w := range s.m.Wonders {
		h := mapmodel.Hash(s.b.seed, 520, mapmodel.HashStr(w.Key))
		for try := 0; try < 40; try++ {
			x := int((h >> 3) % skyW)
			y := int((h >> 17) % (skyH * 3 / 5))
			if s.openAt(x, y) && s.openAt(x-1, y) && s.openAt(x+1, y) {
				c := skyCell{sym: mapmodel.SymWonder, ink: mapmodel.InkStarBright, lv: 3, k: skWonder, ref: int32(i), bold: true}
				if !w.Built {
					c.lv, c.fx = 1, fxBlink
				}
				*s.cell(x, y) = c
				s.labels = append(s.labels, skyLabel{p: pt(x, y), text: w.Name, ink: mapmodel.InkStar, far: true})
				break
			}
			h = mapmodel.Hash(int64(h), int64(try))
		}
	}
}

// openAt reports a scene cell with only the backdrop on it.
func (s *skyScene) openAt(x, y int) bool {
	if !s.in(x, y) {
		return false
	}
	switch s.cells[y*skyW+x].k {
	case skVoid, skStar, skNebula, skSpiral:
		return true
	}
	return false
}

// findTargets lists what Tab steps through: the hub, every building type's
// first unit, the civs and the wonders, nearest the hub first.
func (s *skyScene) findTargets() {
	seen := map[mapmodel.Pt]bool{}
	add := func(p mapmodel.Pt) {
		if !seen[p] && s.in(p.X, p.Y) {
			seen[p] = true
			s.targets = append(s.targets, p)
		}
	}
	add(s.hub)
	keys := make([]string, 0, len(s.anchor))
	for k := range s.anchor {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var rest []mapmodel.Pt
	for _, k := range keys {
		rest = append(rest, s.anchor[k])
	}
	for i := range s.cells {
		if k := s.cells[i].k; k == skCiv && s.cells[i].bold || k == skWonder {
			rest = append(rest, pt(i%skyW, i/skyW))
		}
	}
	h := s.hub
	sort.SliceStable(rest, func(i, j int) bool {
		a, b := rest[i], rest[j]
		if da, db := rdist(a.X, a.Y, h.X, h.Y), rdist(b.X, b.Y, h.X, h.Y); da != db {
			return da < db
		}
		return a.Y < b.Y || a.Y == b.Y && a.X < b.X
	})
	for _, p := range rest {
		add(p)
	}
}

// ------------------------------------------------------------- geometry

// sq is x squared.
func sq(x float64) float64 { return x * x }

// inEllipse reports whether a cell centre lies inside the ellipse.
func inEllipse(x, y int, cx, cy, rx, ry float64) bool {
	return sq((float64(x)-cx)/rx)+sq((float64(y)-cy)/ry) < 1
}

// ringCells is an ellipse's outline as the one-cell, 4-connected boundary
// of its inside (what box drawing needs to draw it unbroken), clockwise
// from the top.
func ringCells(cx, cy, rx, ry float64) []mapmodel.Pt {
	inside := func(x, y int) bool { return inEllipse(x, y, cx, cy, rx, ry) }
	var out []mapmodel.Pt
	for y := int(cy-ry) - 2; y <= int(cy+ry)+2; y++ {
		for x := int(cx-rx) - 2; x <= int(cx+rx)+2; x++ {
			if onRing(x, y, inside) {
				out = append(out, pt(x, y))
			}
		}
	}
	sortClockwise(out, cx, cy)
	return out
}

// sortClockwise orders points by their angle round (cx, cy), clockwise from
// the top, nearest first on ties.
func sortClockwise(ps []mapmodel.Pt, cx, cy float64) {
	key := func(p mapmodel.Pt) float64 { return turns(float64(p.X)-cx, 2*(float64(p.Y)-cy)) }
	sort.SliceStable(ps, func(i, j int) bool {
		a, b := key(ps[i]), key(ps[j])
		if a != b {
			return a < b
		}
		if da, db := rdist(ps[i].X, ps[i].Y, int(cx), int(cy)), rdist(ps[j].X, ps[j].Y, int(cx), int(cy)); da != db {
			return da < db
		}
		return ps[i].Y < ps[j].Y || ps[i].Y == ps[j].Y && ps[i].X < ps[j].X
	})
}

// turns is the angle of (dx, dy) in turns [0, 1), clockwise from up (dy
// grows downward, as rows do). A polynomial arctangent, accurate to about
// 1e-5 turns, keeps package math's transcendentals out of the maps.
func turns(dx, dy float64) float64 {
	if dx == 0 && dy == 0 {
		return 0
	}
	ax, ay := math.Abs(dx), math.Abs(dy)
	lo, hi := math.Min(ax, ay), math.Max(ax, ay)
	t := lo / hi
	t2 := t * t
	a := t * (0.9998660 + t2*(-0.3302995+t2*(0.1801410+t2*(-0.0851330+t2*0.0208351)))) // radians, t in [0, 1]
	if ay > ax {
		a = math.Pi/2 - a
	}
	// a is now the angle from the x axis in the first quadrant; turn it into
	// a bearing clockwise from up
	var b float64
	switch {
	case dx >= 0 && dy < 0: // upper right
		b = math.Pi/2 - a
	case dx >= 0: // lower right
		b = math.Pi/2 + a
	case dy >= 0: // lower left
		b = 3*math.Pi/2 - a
	default: // upper left
		b = 3*math.Pi/2 + a
	}
	tu := b / (2 * math.Pi)
	if tu >= 1 {
		tu--
	}
	if tu < 0 {
		tu++
	}
	return tu
}

// discCells is every cell of a filled ellipse, from the centre outward
// (then clockwise): a growing thing that fills it in this order never moves
// what it has already filled.
func discCells(cx, cy, rx, ry float64) []mapmodel.Pt {
	var out []mapmodel.Pt
	for y := int(cy-ry) - 1; y <= int(cy+ry)+1; y++ {
		for x := int(cx-rx) - 1; x <= int(cx+rx)+1; x++ {
			if inEllipse(x, y, cx, cy, rx, ry) {
				out = append(out, pt(x, y))
			}
		}
	}
	d := func(p mapmodel.Pt) float64 { return sq((float64(p.X)-cx)/rx) + sq((float64(p.Y)-cy)/ry) }
	sort.SliceStable(out, func(i, j int) bool {
		a, b := d(out[i]), d(out[j])
		if a != b {
			return a < b
		}
		return turns(float64(out[i].X)-cx, 2*(float64(out[i].Y)-cy)) < turns(float64(out[j].X)-cx, 2*(float64(out[j].Y)-cy))
	})
	return out
}

// lineCells is the straight run of cells from a to b (Bresenham).
func lineCells(a, b mapmodel.Pt) []mapmodel.Pt {
	dx, dy := abs(b.X-a.X), -abs(b.Y-a.Y)
	sx, sy := 1, 1
	if a.X > b.X {
		sx = -1
	}
	if a.Y > b.Y {
		sy = -1
	}
	err := dx + dy
	out := []mapmodel.Pt{a}
	for p := a; p != b; {
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			p.X += sx
		}
		if e2 <= dx {
			err += dx
			p.Y += sy
		}
		out = append(out, p)
	}
	return out
}

// smoothRing is an ellipse's outline as a thin, 8-connected line, clockwise
// from the top, each cell with the stroke its slope calls for (─ │ ╱ ╲):
// how a hand would draw an ellipse in a terminal.
func smoothRing(cx, cy, rx, ry float64) ([]mapmodel.Pt, []rune) {
	n := int(8*(rx+ry)) + 16
	var ps []mapmodel.Pt
	for i := 0; i <= n; i++ {
		t := float64(i) / float64(n)
		p := pt(int(math.Round(cx+rx*mapmodel.Sin(t))), int(math.Round(cy-ry*mapmodel.Cos(t))))
		if len(ps) > 0 {
			q := ps[len(ps)-1]
			if q == p {
				continue
			}
			if abs(p.X-q.X) > 1 || abs(p.Y-q.Y) > 1 {
				for _, f := range lineCells(q, p)[1:] {
					ps = append(ps, f)
				}
				continue
			}
		}
		ps = append(ps, p)
	}
	if len(ps) > 1 && ps[len(ps)-1] == ps[0] {
		ps = ps[:len(ps)-1]
	}
	// thin it: drop a cell whose two neighbours already touch
	for i := 0; i < len(ps) && len(ps) > 8; {
		a, c := ps[(i+len(ps)-1)%len(ps)], ps[(i+1)%len(ps)]
		if abs(a.X-c.X) <= 1 && abs(a.Y-c.Y) <= 1 {
			ps = append(ps[:i], ps[i+1:]...)
			continue
		}
		i++
	}
	rs := make([]rune, len(ps))
	for i := range ps {
		a, c := ps[(i+len(ps)-1)%len(ps)], ps[(i+1)%len(ps)]
		dx, dy := c.X-a.X, c.Y-a.Y
		switch {
		case dy == 0:
			rs[i] = '─'
		case dx == 0:
			rs[i] = '│'
		case abs(dx) >= 2*abs(dy):
			rs[i] = '─'
		case abs(dy) >= 2*abs(dx):
			rs[i] = '│'
		case dx*dy > 0:
			rs[i] = '╲'
		default:
			rs[i] = '╱'
		}
	}
	return ps, rs
}

// halo is the layer of cells just outside a set of cells (8-adjacent to
// one, in none, and outside the ellipse they ring), clockwise from the top.
func halo(in map[mapmodel.Pt]bool, cx, cy, rx, ry float64) []mapmodel.Pt {
	seen := map[mapmodel.Pt]bool{}
	var out []mapmodel.Pt
	for p := range in {
		for dy := -1; dy <= 1; dy++ {
			for dx := -1; dx <= 1; dx++ {
				q := pt(p.X+dx, p.Y+dy)
				if in[q] || seen[q] || inEllipse(q.X, q.Y, cx, cy, rx, ry) {
					continue
				}
				seen[q] = true
				out = append(out, q)
			}
		}
	}
	sortClockwise(out, cx, cy)
	return out
}

// band is lineage lin's slots in band j (nil past the last band).
func (b *skyBase) band(lin string, j int) []mapmodel.Pt {
	bs := b.bands[lin]
	if j < 0 || j >= len(bs) {
		return nil
	}
	lo := 0
	if j > 0 {
		lo = bs[j-1]
	}
	return b.slots[lin][lo:bs[j]]
}

// addBand appends a band of slots to lineage lin.
func (b *skyBase) addBand(lin string, ps []mapmodel.Pt) {
	b.slots[lin] = append(b.slots[lin], ps...)
	b.bands[lin] = append(b.bands[lin], len(b.slots[lin]))
}

// started reports whether any unit stands in band j of lineage lin.
func (s *skyScene) started(lin string, j int) bool {
	for _, p := range s.b.band(lin, j) {
		if s.at(p.X, p.Y).k == skUnit {
			return true
		}
	}
	return false
}

// sector is slice i of n equal parts of ps.
func sector(ps []mapmodel.Pt, i, n int) []mapmodel.Pt {
	if n <= 0 {
		return nil
	}
	return ps[len(ps)*i/n : len(ps)*(i+1)/n]
}

// centreOut reorders a run of cells from its middle outward, alternating
// sides, so a lineage grows from the heart of its sector.
func centreOut(ps []mapmodel.Pt) []mapmodel.Pt {
	out := make([]mapmodel.Pt, 0, len(ps))
	mid := len(ps) / 2
	for k := 0; len(out) < len(ps); k++ {
		if i := mid + k; i < len(ps) && k >= 0 {
			out = append(out, ps[i])
		}
		if i := mid - k - 1; i >= 0 && len(out) < len(ps) {
			out = append(out, ps[i])
		}
	}
	return out
}
