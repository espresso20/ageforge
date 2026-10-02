package roguelike

import (
	"math"

	"github.com/espresso20/ageforge/mapmodel"
)

// space_deep.go is the Interstellar Age: deep space. The home system has
// shrunk to an orrery in the corner, the home sun ringed by its planets and
// a swarm of stellar collectors. Violet and teal nebula wisps, the first
// colour out here, thread the dark. Colony worlds hang far off, one per
// lineage, each growing as its lineage grows and calling home with a
// beacon; a relay chain links them. A fleet of colony arks waits in
// formation, generation ships cross with engine trails longer than towns,
// escort frigates stand by, and at the far side the warp gate's ring rises
// segment by segment from the engineering lineage, finished (its field
// opening) by the Warp Nexus.

// Deep space layout, in plane cells.
const (
	deepSunX, deepSunY = 11, 7
	deepGateX          = 99.0
	deepGateY          = 20.0
	deepGateRX         = 15.0
	deepGateRY         = 7.5
	deepGateCells      = 2 // gate ring cells one engineering unit raises
)

// colonySite is a colony world: its centre, its lineage's face colour and
// the frames for the inspector.
type colonySite struct {
	c     mapmodel.Pt
	bg    mapmodel.SkyInk
	frame int32
}

// deepColonies are the lineages that settle worlds of their own, and the
// sites they take.
var deepColonies = []struct {
	lin string
	at  mapmodel.Pt
	bg  mapmodel.SkyInk
}{
	{mapmodel.LinFood, mapmodel.Pt{X: 42, Y: 6}, mapmodel.InkSurface2},
	{mapmodel.LinKnowledge, mapmodel.Pt{X: 62, Y: 5}, mapmodel.InkSurface},
	{mapmodel.LinMines, mapmodel.Pt{X: 80, Y: 6}, mapmodel.InkRock},
	{mapmodel.LinTrade, mapmodel.Pt{X: 123, Y: 6}, mapmodel.InkSurface},
	{mapmodel.LinMetal, mapmodel.Pt{X: 47, Y: 17}, mapmodel.InkRock},
	{mapmodel.LinWood, mapmodel.Pt{X: 64, Y: 15}, mapmodel.InkSurface2},
	{mapmodel.LinCulture, mapmodel.Pt{X: 124, Y: 21}, mapmodel.InkSurface},
	{mapmodel.LinStorage, mapmodel.Pt{X: 33, Y: 24}, mapmodel.InkRock},
	{mapmodel.LinFaith, mapmodel.Pt{X: 58, Y: 27}, mapmodel.InkSurface},
	{mapmodel.LinHarbor, mapmodel.Pt{X: 76, Y: 34}, mapmodel.InkSurface2},
	{mapmodel.LinDiplomacy, mapmodel.Pt{X: 100, Y: 36}, mapmodel.InkSurface},
	{mapmodel.LinMonument, mapmodel.Pt{X: 123, Y: 35}, mapmodel.InkRock},
}

// layDeep lays the Interstellar Age's base.
func (b *skyBase) layDeep() {
	b.wisps()
	b.orrery()
	b.gateFrame()
	b.colonies()
	b.relayChain()
	b.arks()
	b.escort()
}

// wisps threads the nebula through the dark: long thin filaments along the
// contours of two broad noise fields, violet and teal, only where a third
// field lets them gather (so the dark stays dark), and never in the
// orrery's corner.
func (b *skyBase) wisps() {
	for y := 0; y < skyH; y++ {
		for x := 0; x < skyW; x++ {
			if x < 30 && y < 15 {
				continue
			}
			mask := mapmodel.Noise(b.seed+29, float64(x)/40, float64(y)/14)
			if mask < 0.48 {
				continue
			}
			n1 := fbm2(b.seed+31, float64(x)/30, float64(y)/9)
			n2 := fbm2(b.seed+37, float64(x)/24, float64(y)/7)
			c := b.at(x, y)
			switch {
			case math.Abs(n1-0.52) < 0.014:
				*c = skyCell{r: '░', ink: mapmodel.InkCloud, lv: 2, k: skNebula, fx: fxWisp, ph: uint8(x / 5)}
			case math.Abs(n2-0.47) < 0.012:
				*c = skyCell{r: '░', ink: mapmodel.InkCloud2, lv: 2, k: skNebula, fx: fxWisp, ph: uint8(y)}
			case math.Abs(n1-0.52) < 0.034 && mapmodel.Hash(b.seed, 551, int64(x), int64(y))%4 == 0:
				*c = skyCell{r: '·', ink: mapmodel.InkCloud, lv: 1, k: skNebula}
			case math.Abs(n2-0.47) < 0.03 && mapmodel.Hash(b.seed, 552, int64(x), int64(y))%5 == 0:
				*c = skyCell{r: '·', ink: mapmodel.InkCloud2, lv: 1, k: skNebula}
			}
		}
	}
}

// fbm2 is two octaves of the model's value noise.
func fbm2(seed int64, x, y float64) float64 {
	return (2*mapmodel.Noise(seed, x, y) + mapmodel.Noise(seed+1, x*2.1, y*2.1)) / 3
}

// orrery draws the home system in the corner: the sun, its corona and the
// orbits its planets ride (the planets move: deepOverlay), and reserves the
// stellar collectors' swarm between the inner orbits.
func (b *skyBase) orrery() {
	sx, sy := float64(deepSunX), float64(deepSunY)
	for _, o := range deepOrbits {
		ps, _ := smoothRing(sx, sy, o[0], o[1])
		for i, p := range ps {
			if i%2 == 0 && b.free(p.X, p.Y) {
				b.put(p.X, p.Y, skyCell{r: '·', ink: mapmodel.InkFrameDim, lv: 1, k: skOrbit})
			}
		}
	}
	for _, d := range [8][2]int{{-2, 0}, {2, 0}, {-1, -1}, {1, -1}, {-1, 1}, {1, 1}, {0, -1}, {0, 1}} {
		b.put(deepSunX+d[0], deepSunY+d[1], skyCell{r: '·', ink: mapmodel.InkAccent3, lv: 2, k: skSun})
	}
	b.hub = pt(deepSunX, deepSunY)
	b.put(deepSunX, deepSunY, skyCell{sym: mapmodel.SymSun, ink: mapmodel.InkAccent, lv: 3, k: skHub, bold: true})
	var swarm []mapmodel.Pt
	for _, r := range [][2]float64{{6.5, 2.6}, {9.6, 3.8}} {
		ps, _ := smoothRing(sx, sy, r[0], r[1])
		swarm = append(swarm, ps...)
	}
	var keep []mapmodel.Pt
	for _, p := range swarm {
		if b.free(p.X, p.Y) || b.at(p.X, p.Y).k == skOrbit {
			keep = append(keep, p)
		}
	}
	b.addBand(mapmodel.LinEnergy, keep)
}

// deepOrbits are the orrery's orbits: radii in columns and rows.
var deepOrbits = [4][2]float64{{4, 1.7}, {8, 3.2}, {11.5, 4.5}, {15, 6}}

// gateFrame lays the warp gate's ring out in build order (from its foot up
// both sides at once) and draws it all as scaffold; the engineering units
// raise it a few cells each.
func (b *skyBase) gateFrame() {
	ps, rs := smoothRing(deepGateX, deepGateY, deepGateRX, deepGateRY)
	foot := 0
	for i, p := range ps {
		if p.Y > ps[foot].Y || p.Y == ps[foot].Y && abs(p.X-int(deepGateX)) < abs(ps[foot].X-int(deepGateX)) {
			foot = i
		}
	}
	scaffold := b.frame("Gate scaffold", "the warp gate's ring, still to raise", slScaffold, mapmodel.LinEngineer)
	b.frame("Warp gate", "its ring rises segment by segment", slGate, mapmodel.LinEngineer)
	n := len(ps)
	order := make([]int, 0, n)
	for k := 0; len(order) < n; k++ {
		if k == 0 {
			order = append(order, foot)
			continue
		}
		order = append(order, (foot+k)%n)
		if len(order) < n {
			order = append(order, (foot-k+n)%n)
		}
	}
	b.gate = make([]mapmodel.Pt, n)
	b.gateStroke = make([]rune, n)
	for i, j := range order {
		b.gate[i], b.gateStroke[i] = ps[j], rs[j]
		b.put(ps[j].X, ps[j].Y, skyCell{r: '·', ink: mapmodel.InkFrameDim, lv: 1, k: skSlot, ref: scaffold})
	}
	var slots []mapmodel.Pt
	for i := 0; i < gateEngCells(n); i += deepGateCells {
		slots = append(slots, b.gate[i])
	}
	b.addBand(mapmodel.LinEngineer, slots)
}

// gateEngCells is how much of a gate ring of n cells engineering can raise:
// three quarters of it. The keystone arc at the top is the Warp Nexus's.
func gateEngCells(n int) int { return n * 3 / 4 }

// gateRune is the stroke the gate's ring takes where engineering unit n
// stands (one unit raises deepGateCells cells).
func gateRune(b *skyBase, n int) rune {
	if i := n * deepGateCells; i < len(b.gateStroke) {
		return b.gateStroke[i]
	}
	return '·'
}

// colonies reserves each colony world's cells, from its heart outward; the
// world is drawn as large as its lineage needs (deepLife).
func (b *skyBase) colonies() {
	for _, c := range deepColonies {
		fr := b.frame(mapmodel.SkyPartOf(mapmodel.SkyDeep, c.lin).Name, "a colony world of your own", slColony, c.lin)
		b.colony[c.lin] = colonySite{c: c.at, bg: c.bg, frame: fr}
		var slots []mapmodel.Pt
		for _, p := range discCells(float64(c.at.X), float64(c.at.Y), 8, 3.6) {
			if b.in(p.X, p.Y) {
				slots = append(slots, p)
			}
		}
		b.addBand(c.lin, slots[:min(len(slots), 40)])
	}
}

// colonyRadius is how large a colony world with n units is drawn: big
// enough to hold them all.
func colonyRadius(n int) float64 {
	switch {
	case n <= 2:
		return 1
	case n <= 9:
		return 1.6
	case n <= 20:
		return 2.4
	}
	return 3.2
}

// relayChain runs the relay chain from home through every colony site to
// the gate, faintly, and reserves its relay beacons along it.
func (b *skyBase) relayChain() {
	fr := b.frame("Relay chain", "it carries word from world to world", slRelay, mapmodel.LinHacker)
	stops := []mapmodel.Pt{pt(deepSunX+4, deepSunY+1)}
	for _, c := range deepColonies {
		stops = append(stops, c.at)
	}
	stops = append(stops, pt(int(deepGateX-deepGateRX)-1, int(deepGateY)))
	// visit the sites nearest-first from home, so the chain wanders outward
	route := []mapmodel.Pt{stops[0]}
	left := append([]mapmodel.Pt{}, stops[1:len(stops)-1]...)
	for len(left) > 0 {
		cur, bi := route[len(route)-1], 0
		for i, p := range left {
			if rdist(p.X, p.Y, cur.X, cur.Y) < rdist(left[bi].X, left[bi].Y, cur.X, cur.Y) {
				bi = i
			}
		}
		route = append(route, left[bi])
		left = append(left[:bi], left[bi+1:]...)
	}
	route = append(route, stops[len(stops)-1])
	var slots []mapmodel.Pt
	step := 0
	for i := 1; i < len(route); i++ {
		for _, p := range lineCells(route[i-1], route[i])[1:] {
			if !b.free(p.X, p.Y) || inEllipse(p.X, p.Y, float64(route[i].X), float64(route[i].Y), 9, 4.2) {
				continue
			}
			if step%2 == 0 {
				b.put(p.X, p.Y, skyCell{r: '·', ink: mapmodel.InkFrameDim, lv: 0, k: skFrame, ref: fr})
			}
			if step%5 == 2 {
				slots = append(slots, p)
			}
			step++
		}
	}
	b.addBand(mapmodel.LinHacker, slots)
}

// arks reserves the colony arks' formation in the lower left, the front
// rank first.
func (b *skyBase) arks() {
	var slots []mapmodel.Pt
	for _, y := range []int{32, 34, 36, 38} {
		for x := 6; x <= 40; x += 2 {
			if b.free(x+(y/2)%2, y) {
				slots = append(slots, pt(x+(y/2)%2, y))
			}
		}
	}
	b.addBand(mapmodel.LinHousing, slots)
}

// escort reserves the escort frigates' chevrons before the gate.
func (b *skyBase) escort() {
	var slots []mapmodel.Pt
	for rank := 0; rank < 3; rank++ {
		tip := pt(int(deepGateX-deepGateRX)-4-rank*7, int(deepGateY))
		slots = append(slots, tip)
		for k := 1; k <= 5; k++ {
			slots = append(slots, pt(tip.X-2*k, tip.Y-k), pt(tip.X-2*k, tip.Y+k))
		}
	}
	b.addBand(mapmodel.LinMilitary, slots)
}

// deepLife puts what changes with play on the Interstellar Age: the gate's
// raised segments (all of them once the Warp Nexus stands, its field open),
// the colony worlds at the size their lineages need, with beacons, and the
// traffic.
func (s *skyScene) deepLife() {
	b := s.b
	m := s.m
	built, progress := false, 0.0
	for _, w := range m.Wonders {
		if w.Key == "warp_nexus" {
			built, progress = w.Built, w.Progress
		}
	}
	gate := b.frameNamed(slGate)
	n := len(b.gate)
	eng, key := 0, 0 // cells engineering has raised; keystone cells the wonder has
	for i, p := range b.slots[mapmodel.LinEngineer] {
		if s.at(p.X, p.Y).k == skUnit {
			eng = max(eng, (i+1)*deepGateCells)
		}
	}
	eng = min(eng, gateEngCells(n))
	switch {
	case built:
		eng, key = gateEngCells(n), n-gateEngCells(n)
	default:
		key = int(progress * float64(n-gateEngCells(n)))
	}
	raise := func(i int) {
		p := b.gate[i]
		if c := s.cell(p.X, p.Y); c.k != skUnit {
			*c = skyCell{r: b.gateStroke[i], ink: mapmodel.InkFrame, lv: 2, k: skFrame, ref: gate}
		}
	}
	for i := 0; i < eng; i++ {
		raise(i)
	}
	for i := gateEngCells(n); i < gateEngCells(n)+key && i < n; i++ {
		raise(i)
	}
	raised := eng
	if built {
		for _, p := range discCells(deepGateX, deepGateY, deepGateRX-1.5, deepGateRY-1) {
			if c := s.cell(p.X, p.Y); c.k == skVoid || c.k == skStar || c.k == skNebula {
				dx, dy := float64(p.X)-deepGateX, 2*(float64(p.Y)-deepGateY)
				d := math.Sqrt(dx*dx+dy*dy) / deepGateRX
				ph := uint8(int(turns(dx, dy)*24+d*10) % 24)
				*c = skyCell{r: '░', ink: mapmodel.InkAccent2, lv: 2, k: skField, fx: fxSwirl, ph: ph}
			}
		}
		for _, p := range b.gate {
			if c := s.cell(p.X, p.Y); c.k == skFrame || c.k == skUnit {
				c.ink = mapmodel.InkAccent2
			}
		}
	} else {
		fronts := []int{raised, raised + 1} // welding at both fronts
		if key > 0 {
			fronts = append(fronts, gateEngCells(n)+key, gateEngCells(n)+key+1)
		}
		for _, i := range fronts {
			if i < len(b.gate) {
				p := b.gate[i]
				if c := s.cell(p.X, p.Y); c.k == skSlot {
					c.fx, c.ph = fxSpark, uint8(i*7)
				}
			}
		}
	}
	for _, c := range deepColonies {
		site := b.colony[c.lin]
		n := 0
		for _, p := range b.slots[c.lin] {
			if s.at(p.X, p.Y).k == skUnit {
				n++
			}
		}
		if n == 0 {
			continue
		}
		r := colonyRadius(n)
		top := site.c.Y
		for _, p := range discCells(float64(site.c.X), float64(site.c.Y), 2*r, r) {
			cell := s.cell(p.X, p.Y)
			if cell.k == skUnit {
				cell.bg = site.bg
				if tt := &m.Town.Tiles[cell.ref]; s.anchor[tt.Key] != p {
					cell.sym, cell.r, cell.lv = mapmodel.SymNone, '∙', 2 // a settlement's lights
					if (p.X+p.Y)%3 == 0 {
						cell.r = '•'
					}
				} else {
					cell.bold = true
				}
			} else {
				*cell = skyCell{r: ' ', bg: site.bg, k: skColony, ref: site.frame}
			}
			top = min(top, p.Y)
		}
		if c := s.cell(site.c.X, top-1); c.k != skUnit && c.k != skHub {
			*c = skyCell{r: '•', ink: mapmodel.InkLight, lv: 3, k: skBeacon, fx: fxBlink, ph: uint8(site.c.X), bold: true,
				ref: site.frame}
		}
		s.labels = append(s.labels, skyLabel{p: pt(site.c.X, top-1), text: mapmodel.SkyPartOf(s.sky, c.lin).Name,
			ink: mapmodel.InkLight, far: true})
	}
	s.deepTraffic()
}

// deepTraffic: generation ships from home out to the colonies and the
// gate, as many as the fleet of arks warrants, and mining drones between
// the mining world and home.
func (s *skyScene) deepTraffic() {
	b := s.b
	home := pt(deepSunX+12, deepSunY+3)
	arks := 0
	for _, p := range b.slots[mapmodel.LinHousing] {
		if s.at(p.X, p.Y).k == skUnit {
			arks++
		}
	}
	dests := []mapmodel.Pt{pt(int(deepGateX-deepGateRX)+2, int(deepGateY)), deepColonies[3].at, deepColonies[6].at,
		deepColonies[10].at, deepColonies[11].at}
	for i := 0; i < min(4, 1+arks/10); i++ {
		d := dests[int(mapmodel.Hash(b.seed, 630, int64(i))%uint64(len(dests)))]
		from := pt(home.X, home.Y+i*2)
		if mv := s.addMover(smGenShip, lineCells(from, d), i, true, 60+int(mapmodel.Hash(b.seed, 631, int64(i))%160)); mv != nil {
			mv.trail = 26
		}
	}
	mine := b.colony[mapmodel.LinMines].c
	for i := 0; i < 2; i++ {
		s.addMover(smMiningDrone, lineCells(pt(mine.X-3, mine.Y+2+i), pt(home.X-2, home.Y-1+i)), i, false, 10)
	}
}
