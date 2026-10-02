package roguelike

import (
	"math"

	"github.com/espresso20/ageforge/mapmodel"
)

// space_orbit.go is the Space Age: actual space. The camera has left the
// ground. The homeworld's curved limb fills the bottom of the map, its
// night side showing the player's own world (land and sea from the seed's
// terrain), cloud swirls drifting over it and amber city lights where the
// town stands, with the civs met glowing in their colours. The space
// elevator's tether rises from the town to the ring station, whose hub and
// spokes hold a ring of modules, one sector per lineage, growing outward
// ring by ring. Solar wings run out along the truss, hulls rise on the
// slips, foundry tanks glow, relay satellites ride a low orbit, the
// asteroid belt crosses the sky with mining outposts on its rocks, and the
// moon carries the base.

// Orbit layout, in plane cells.
const (
	orbHubX, orbHubY = skyW / 2, 14
	orbRX, orbRY     = 22.0, 5.0 // the main ring
	orbRYc           = 5         // its top and bottom rows are this far from the hub
	orbLimb          = 31.0      // the limb's top row, under the hub
	orbPlanetR       = 50.0      // the planet's radius in rows
	orbWingIn        = 27        // the solar wings start this far out
	orbMoonX         = 116.0
	orbMoonY         = 6.5
)

// orbRingLins are the lineages whose units live on the ring, clockwise from
// the top.
var orbRingLins = []string{mapmodel.LinHousing, mapmodel.LinFood, mapmodel.LinWood, mapmodel.LinStorage,
	mapmodel.LinKnowledge, mapmodel.LinFaith, mapmodel.LinCulture, mapmodel.LinMonument, mapmodel.LinTrade,
	mapmodel.LinHarbor, mapmodel.LinDiplomacy}

// planetGeom is the homeworld below: a circle of radius r rows centred
// below the plane (a column is half a row), with the world projected on its
// night side.
type planetGeom struct {
	cx, cy, r float64
	wcx, wcy  int // the world's centre
	ww, wh    int // and size
	foot      mapmodel.Pt
}

// limb is the planet's top edge at column x, in rows.
func (p *planetGeom) limb(x int) float64 {
	dx := (float64(x) - p.cx) / 2
	d := p.r*p.r - dx*dx
	if d <= 0 {
		return 1e9
	}
	return p.cy - math.Sqrt(d)
}

// toPlane is where world tile (wx, wy) shows on the night side: the world
// spread across the planet's face, its north at the limb.
func (p *planetGeom) toPlane(wx, wy int) mapmodel.Pt {
	u := float64(wx-p.wcx) / float64(max(1, p.ww/2))
	px := int(math.Round(p.cx + u*float64(skyW/2)))
	yl := math.Ceil(p.limb(px))
	v := 0.5 + float64(wy-p.wcy)/float64(max(1, p.wh))
	return pt(px, int(yl+v*(float64(skyH)-yl)))
}

// toWorld is the world tile under plane cell (px, py) on the planet.
func (p *planetGeom) toWorld(px, py int) (int, int) {
	u := (float64(px) - p.cx) / float64(skyW/2)
	yl := math.Ceil(p.limb(px))
	v := (float64(py) - yl) / math.Max(1, float64(skyH)-yl)
	return p.wcx + int(u*float64(p.ww/2)), p.wcy + int((v-0.5)*float64(p.wh))
}

var eighthsUp = [9]rune{' ', '▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

// layOrbit lays the Space Age's base.
func (b *skyBase) layOrbit(m *mapmodel.Model) {
	b.belt()
	b.moon()
	b.layPlanet(m)
	b.station()
	b.wings()
	b.slips()
	b.foundry()
	b.relays()
}

// layPlanet draws the homeworld's face and the atmosphere at its limb.
func (b *skyBase) layPlanet(m *mapmodel.Model) {
	w := m.Town.World
	p := planetGeom{cx: float64(skyW / 2), cy: orbLimb + orbPlanetR, r: orbPlanetR, ww: mapmodel.WorldW, wh: mapmodel.WorldH,
		wcx: mapmodel.WorldW / 2, wcy: mapmodel.WorldH / 2}
	if w != nil {
		p.wcx, p.wcy, p.ww, p.wh = w.CX, w.CY, w.W, w.H
	}
	p.foot = p.toPlane(p.wcx, p.wcy)
	b.planet = p
	for x := 0; x < skyW; x++ {
		yl := p.limb(x)
		top := int(math.Ceil(yl))
		if fl := int(math.Floor(yl)); yl-float64(fl) > 0.01 && fl >= 0 && fl < skyH {
			if k := int(math.Round((1 - (yl - float64(fl))) * 8)); k > 0 {
				b.put(x, fl, skyCell{r: eighthsUp[min(8, k)], ink: mapmodel.InkGlow, lv: 2, k: skLimb})
			}
		}
		if y := int(math.Floor(yl)) - 1; y >= 0 && y < skyH && b.free(x, y) {
			b.put(x, y, skyCell{r: '▁', ink: mapmodel.InkCloud2, lv: 1, k: skLimb})
		}
		for y := max(0, top); y < skyH; y++ {
			c := skyCell{r: ' ', bg: mapmodel.InkSurface, k: skPlanet, fx: fxCloud}
			if w != nil {
				if wx, wy := p.toWorld(x, y); w.At(wx, wy).Land() {
					c.bg = mapmodel.InkSurface2
				}
			}
			if y == top {
				c.r, c.ink, c.lv = '▀', mapmodel.InkGlow, 1 // the glow runs into the air's first row
			}
			b.put(x, y, c)
		}
	}
}

// belt scatters the asteroid belt across the top of the sky; mining
// outposts take rocks on its middle line, nearest the hub first.
func (b *skyBase) belt() {
	ph := mapmodel.HashF(b.seed, 540)
	yb := func(x int) float64 { return 2.6 + 1.3*mapmodel.Sin(float64(x)/52+ph) }
	for x := 0; x < skyW; x++ {
		c := yb(x)
		for y := int(c) - 1; y <= int(c)+1; y++ {
			h := mapmodel.Hash(b.seed, 541, int64(x), int64(y))
			core := float64(y) <= c+0.5 && float64(y) >= c-0.5
			if (core && h%100 < 52 || h%100 < 22) && b.free(x, y) {
				ink := mapmodel.InkRock
				if h>>16%3 == 0 {
					ink = mapmodel.InkRockDark
				}
				b.put(x, y, skyCell{r: []rune("o•◦▪o∙")[h>>8%6], ink: ink, lv: 2, k: skRock})
			}
		}
	}
	var slots []mapmodel.Pt
	for k := 0; ; k++ {
		done := 0
		for _, x := range []int{orbHubX - 2 - 4*k, orbHubX + 2 + 4*k} {
			if x < 2 || x > 104 {
				done++
				continue
			}
			slots = append(slots, pt(x, int(math.Round(yb(x)))))
		}
		if done == 2 {
			break
		}
	}
	b.addBand(mapmodel.LinMines, slots)
}

// moon draws the moon, lit from the left, and reserves its base's cells,
// spreading out from the lit side.
func (b *skyBase) moon() {
	cx, cy, rx, ry := orbMoonX, orbMoonY, 8.0, 3.6
	for _, p := range discCells(cx, cy, rx, ry) {
		c := skyCell{r: ' ', bg: mapmodel.InkMoon, k: skMoon}
		if float64(p.X) > cx+rx*0.42 {
			c.bg = mapmodel.InkRockDark // the night side
		}
		if mapmodel.Hash(b.seed, 545, int64(p.X), int64(p.Y))%9 == 0 {
			c.r, c.ink, c.lv = '∘', mapmodel.InkRockDark, 1 // a crater
		}
		b.put(p.X, p.Y, c)
	}
	var base []mapmodel.Pt
	for _, p := range discCells(cx-3, cy+0.5, rx-1, ry-0.6) {
		if inEllipse(p.X, p.Y, cx, cy, rx-0.8, ry-0.4) && (p.X+p.Y)%2 == 0 {
			base = append(base, p)
		}
	}
	b.addBand(mapmodel.LinMilitary, base[:min(len(base), 24)])
}

// station draws the hub, its spokes and the main ring, and reserves each
// ring lineage's sector in the three layers of berths round the ring (a
// band per layer): modules dock outside the ring like beads on a string,
// and a lineage that outgrows its berths takes the layer beyond.
func (b *skyBase) station() {
	hx, hy := orbHubX, orbHubY
	fx, fy := float64(hx), float64(hy)
	ring, strokes := smoothRing(fx, fy, orbRX, orbRY)
	fr := b.frame("Station ring", "the main ring of the station", slRing, "")
	spoke := b.frame("Spoke", "from the hub to the ring", 0, "")
	on := map[mapmodel.Pt]bool{}
	for i, p := range ring {
		on[p] = true
		b.put(p.X, p.Y, skyCell{r: strokes[i], ink: mapmodel.InkFrame, lv: 2, k: skFrame, ref: fr})
	}
	b.hubAt(hx, hy)
	for x := hx - 1; !on[pt(x, hy)] && x > 0; x-- {
		b.put(x, hy, skyCell{r: '─', ink: mapmodel.InkFrameDim, lv: 1, k: skFrame, ref: spoke})
	}
	for x := hx + 1; !on[pt(x, hy)] && x < skyW; x++ {
		b.put(x, hy, skyCell{r: '─', ink: mapmodel.InkFrameDim, lv: 1, k: skFrame, ref: spoke})
	}
	for y := hy - 1; !on[pt(hx, y)] && y > 0; y-- {
		b.put(hx, y, skyCell{r: '│', ink: mapmodel.InkFrameDim, lv: 1, k: skFrame, ref: spoke})
	}
	for y := hy + 1; !on[pt(hx, y)] && y < skyH; y++ {
		b.put(hx, y, skyCell{r: '│', ink: mapmodel.InkFrameDim, lv: 1, k: skFrame, ref: spoke})
	}
	// the berths: three layers outside the ring, minus the tether's column
	// below it and the truss's row beside it
	used := map[mapmodel.Pt]bool{}
	for p := range on {
		used[p] = true
	}
	var layers [3][]mapmodel.Pt
	for k := range layers {
		var keep []mapmodel.Pt
		for _, p := range halo(used, fx, fy, orbRX, orbRY) {
			used[p] = true
			if p.X == hx && p.Y > hy || p.Y == hy {
				continue
			}
			keep = append(keep, p)
		}
		layers[k] = rotate(keep, len(keep)-len(keep)/(2*len(orbRingLins)))
	}
	for i, lin := range orbRingLins {
		for _, l := range layers {
			b.addBand(lin, centreOut(sector(l, i, len(orbRingLins))))
		}
	}
	b.frame("Outer ring", "the station grows outward", slArc, "")
}

// rotate starts a clockwise ring n cells later (so no sector straddles the
// top).
func rotate(ps []mapmodel.Pt, n int) []mapmodel.Pt {
	if len(ps) == 0 {
		return ps
	}
	n %= len(ps)
	return append(append([]mapmodel.Pt{}, ps[n:]...), ps[:n]...)
}

// hubAt sets the scene's heart.
func (b *skyBase) hubAt(x, y int) {
	b.hub = pt(x, y)
	b.put(x, y, skyCell{sym: mapmodel.SymSkyHub, ink: mapmodel.InkFrame, lv: 3, k: skHub, bold: true})
}

// frameNamed is the index of the frame with legend row lg (the first), or
// -1.
func (b *skyBase) frameNamed(lg skyLgID) int32 {
	for i, f := range b.frames {
		if f.lg == lg {
			return int32(i)
		}
	}
	return -1
}

// wings lays the truss from the ring to the solar wings and reserves the
// wings' panels, both sides alike, inner rows first, a band per column pair
// and a strut every sixth column.
func (b *skyBase) wings() {
	hy := orbHubY
	truss := b.frame("Truss", "it carries the solar wings", slTruss, mapmodel.LinEnergy)
	for x := orbHubX - orbWingIn; x < skyW && b.at(x, hy).k != skFrame; x++ {
		b.put(x, hy, skyCell{r: '═', ink: mapmodel.InkFrameDim, lv: 1, k: skFrame, ref: truss})
	}
	for x := orbHubX + orbWingIn; x >= 0 && b.at(x, hy).k != skFrame; x-- {
		b.put(x, hy, skyCell{r: '═', ink: mapmodel.InkFrameDim, lv: 1, k: skFrame, ref: truss})
	}
	for col := 0; ; col++ {
		xl, xr := orbHubX-orbWingIn-1-col, orbHubX+orbWingIn+1+col
		if xl < 1 && xr > skyW-2 {
			break
		}
		if col%6 == 5 {
			continue // a strut
		}
		var band []mapmodel.Pt
		for _, x := range []int{xl, xr} {
			if x < 1 || x > skyW-2 {
				continue
			}
			for _, y := range []int{hy - 1, hy + 1, hy - 2, hy + 2} {
				band = append(band, pt(x, y))
			}
		}
		b.addBand(mapmodel.LinEnergy, band)
	}
}

// orbSlips are the shipyard's slips: each one's top-left corner (five hull
// cells under a gantry), nearest the station first.
var orbSlips = [8]mapmodel.Pt{{X: 33, Y: 23}, {X: 24, Y: 23}, {X: 15, Y: 23}, {X: 6, Y: 23},
	{X: 33, Y: 27}, {X: 24, Y: 27}, {X: 15, Y: 27}, {X: 6, Y: 27}}

// slips reserves the shipyard's hull cells, a band per slip.
func (b *skyBase) slips() {
	for _, o := range orbSlips {
		var band []mapmodel.Pt
		for i := 1; i <= 5; i++ {
			band = append(band, pt(o.X+i, o.Y+1))
		}
		b.addBand(mapmodel.LinEngineer, band)
	}
	b.frame("Slip", "a hull is rising here", slSlip, mapmodel.LinEngineer)
}

// foundry reserves the foundry's tanks, three rows deep, nearest the
// station first, a band per row.
func (b *skyBase) foundry() {
	for _, y := range []int{24, 26, 28} {
		var band []mapmodel.Pt
		for x := 95; x <= 125; x += 2 {
			band = append(band, pt(x, y))
		}
		b.addBand(mapmodel.LinMetal, band)
	}
	b.frame("Foundry pipes", "ore in, metal out", slTruss, mapmodel.LinMetal)
}

// relays dots the relay satellites' low orbit over the limb and reserves
// its stations, nearest the tether first.
func (b *skyBase) relays() {
	p := &b.planet
	orbit := b.frame("Relay orbit", "the relay satellites' track", slOrbitLine, mapmodel.LinHacker)
	at := func(x int) mapmodel.Pt { return pt(x, int(math.Round(p.limb(x)-3))) }
	for x := 1; x < skyW-1; x += 2 {
		if q := at(x); x != orbHubX && b.free(q.X, q.Y) {
			b.put(q.X, q.Y, skyCell{r: '·', ink: mapmodel.InkFrameDim, lv: 0, k: skFrame, ref: orbit})
		}
	}
	var slots []mapmodel.Pt
	for k := 0; ; k++ {
		done := 0
		for _, x := range []int{orbHubX - 4 - 4*k, orbHubX + 4 + 4*k} {
			if x < 2 || x > skyW-3 {
				done++
				continue
			}
			slots = append(slots, at(x))
		}
		if done == 2 {
			break
		}
	}
	b.addBand(mapmodel.LinHacker, slots)
}

// orbitLife puts what changes with play on the Space Age: the frames of
// the outer rings, slips, wings and foundry that hold units, the city
// lights and the civs on the night side, the tether, and the traffic.
func (s *skyScene) orbitLife() {
	s.orbitSlips()
	s.orbitWings()
	s.orbitFoundry()
	s.orbitPlanet()
	s.orbitTraffic()
}

// orbitSlips draws each slip with a hull on it: the gantry, and the hull
// cells still to come as scaffold.
func (s *skyScene) orbitSlips() {
	slip := s.b.frameNamed(slSlip)
	for j, o := range orbSlips {
		if !s.started(mapmodel.LinEngineer, j) {
			continue
		}
		frame := func(x, y int, r rune) {
			*s.cell(x, y) = skyCell{r: r, ink: mapmodel.InkFrameDim, lv: 2, k: skFrame, ref: slip}
		}
		frame(o.X, o.Y, '╒')
		frame(o.X+6, o.Y, '╕')
		frame(o.X, o.Y+1, '│')
		frame(o.X+6, o.Y+1, '│')
		frame(o.X, o.Y+2, '╘')
		frame(o.X+6, o.Y+2, '╛')
		for i := 1; i <= 5; i++ {
			frame(o.X+i, o.Y, '═')
			frame(o.X+i, o.Y+2, '═')
			if c := s.cell(o.X+i, o.Y+1); c.k != skUnit {
				*c = skyCell{r: '░', ink: mapmodel.InkFrameDim, lv: 1, k: skSlot, ref: slip}
			}
		}
	}
}

// orbitWings runs the truss out to the last column of panels on each side,
// with a strut where the panels pause.
func (s *skyScene) orbitWings() {
	truss := s.b.frameNamed(slTruss)
	hy := orbHubY
	far := [2]int{orbHubX - orbWingIn, orbHubX + orbWingIn}
	for _, p := range s.b.slots[mapmodel.LinEnergy] {
		if s.at(p.X, p.Y).k != skUnit {
			continue
		}
		if p.X < orbHubX {
			far[0] = min(far[0], p.X-1)
		} else {
			far[1] = max(far[1], p.X+1)
		}
	}
	for x := max(0, far[0]); x < orbHubX-orbWingIn; x++ {
		r := '═'
		if (orbHubX-orbWingIn-1-x)%6 == 5 {
			r = '╪'
		}
		*s.cell(x, hy) = skyCell{r: r, ink: mapmodel.InkFrameDim, lv: 1, k: skFrame, ref: truss}
	}
	for x := orbHubX + orbWingIn + 1; x <= min(skyW-1, far[1]); x++ {
		r := '═'
		if (x-orbHubX-orbWingIn-1)%6 == 5 {
			r = '╪'
		}
		*s.cell(x, hy) = skyCell{r: r, ink: mapmodel.InkFrameDim, lv: 1, k: skFrame, ref: truss}
	}
}

// orbitFoundry links each row of tanks with a pipe, and the rows with a
// riser.
func (s *skyScene) orbitFoundry() {
	pipe := s.b.frameNamed(slTruss) // the foundry's pipes share the truss's row
	for i := range s.b.frames {
		if s.b.frames[i].lin == mapmodel.LinMetal {
			pipe = int32(i)
		}
	}
	rows := []int{}
	for j, y := range []int{24, 26, 28} {
		band := s.b.band(mapmodel.LinMetal, j)
		last := -1
		for _, p := range band {
			if s.at(p.X, p.Y).k == skUnit {
				last = p.X
			}
		}
		if last < 0 {
			continue
		}
		rows = append(rows, y)
		for x := 93; x <= last; x++ {
			if c := s.cell(x, y); c.k != skUnit {
				*c = skyCell{r: '─', ink: mapmodel.InkFrameDim, lv: 1, k: skFrame, ref: pipe}
			}
		}
	}
	if len(rows) > 0 {
		for y := orbHubY + 1; y <= rows[len(rows)-1]; y++ {
			c := s.cell(93, y)
			switch {
			case y == orbHubY+1 || c.k == skVoid || c.k == skStar:
				*c = skyCell{r: '│', ink: mapmodel.InkFrameDim, lv: 1, k: skFrame, ref: pipe}
			case c.k == skFrame && c.r == '─':
				c.r = '├'
			}
		}
	}
}

// orbitPlanet lights the night side: the town where it stands, the civs
// met in their colours, and the tether's foot in the heart of the town.
func (s *skyScene) orbitPlanet() {
	p := &s.b.planet
	m := s.m
	lit := map[mapmodel.Pt]int{}
	for _, t := range m.Town.Tiles {
		if !t.Ruin {
			lit[p.toPlane(t.X, t.Y)]++
		}
	}
	for q, n := range lit {
		c := s.cell(q.X, q.Y)
		if c.k != skPlanet {
			continue
		}
		c.k, c.ink, c.lv, c.ref, c.fx = skCity, mapmodel.InkLight, 2, int32(n), fxNone
		switch {
		case n >= 6:
			c.r, c.lv, c.bold = '•', 3, true
		case n >= 3:
			c.r = '•'
		default:
			c.r = '∙'
		}
	}
	for fi := range m.Factions {
		f := &m.Factions[fi]
		w := m.Town.World
		if !f.Discovered || w == nil || f.Site < 0 || f.Site >= len(w.Sites) {
			continue
		}
		q := p.toPlane(w.Sites[f.Site].X, w.Sites[f.Site].Y)
		ink := inkRel + mapmodel.SkyInk(f.Relation)
		for dx := -1; dx <= 1; dx++ {
			if c := s.cell(q.X+dx, q.Y); c.k == skPlanet || c.k == skCity {
				*c = skyCell{r: '∙', ink: ink, lv: 2, k: skCiv, ref: int32(fi), bg: c.bg}
			}
		}
		if c := s.cell(q.X, q.Y); c.k == skCiv {
			c.r, c.bold = '•', true
			if f.Relation == mapmodel.RelWar {
				c.sym = mapmodel.SymWar
			}
		}
		s.labels = append(s.labels, skyLabel{p: q, text: f.Name, rel: int(f.Relation) + 1})
	}
	// the tether, from the heart of the town to the ring
	foot := p.foot
	for y := orbHubY + orbRYc + 1; y < foot.Y; y++ {
		*s.cell(foot.X, y) = skyCell{r: '║', ink: mapmodel.InkFrame, lv: 2, k: skTether}
	}
	*s.cell(foot.X, foot.Y) = skyCell{r: '╨', ink: mapmodel.InkLight, lv: 3, k: skTether, bold: true}
}

// orbitTraffic lays the Space Age's lanes: climbers on the tether, shuttles
// up from the town, satellites in a low orbit and mining drones off the
// belt, as many as the town is busy.
func (s *skyScene) orbitTraffic() {
	p := &s.b.planet
	busy := s.m.Activity.Traffic
	var tether []mapmodel.Pt
	for y := p.foot.Y - 1; y > orbHubY+orbRYc; y-- {
		tether = append(tether, pt(p.foot.X, y))
	}
	for i := 0; i < 2; i++ {
		s.addMover(smClimber, tether, i, false, 6)
	}
	target := pt(orbHubX-6, orbHubY+orbRYc+1)
	if s.started(mapmodel.LinEngineer, 0) {
		target = pt(orbSlips[0].X+3, orbSlips[0].Y-1)
	}
	for i := 0; i < 1+int(busy*1.5); i++ {
		x := p.foot.X - 5 - 4*i
		from := pt(x, int(math.Ceil(p.limb(x))))
		s.addMover(smShuttle, lineCells(from, target), i, true, 160+int(mapmodel.Hash(s.b.seed, 620, int64(i))%240))
	}
	for i := 0; i < 2; i++ {
		var arc []mapmodel.Pt
		for x := 1; x < skyW-1; x++ {
			arc = append(arc, pt(x, int(math.Round(p.limb(x)-5-float64(i)))))
		}
		if mapmodel.Hash(s.b.seed, 621, int64(i))%2 == 0 {
			arc = reversed(arc)
		}
		s.addMover(smSatellite, arc, i, true, 40+int(mapmodel.Hash(s.b.seed, 622, int64(i))%160))
	}
	mines := s.b.slots[mapmodel.LinMines]
	drop := pt(orbHubX+int(orbRX)+8, orbHubY-3)
	if s.started(mapmodel.LinMetal, 0) {
		drop = pt(95, 23)
	}
	for i := 0; i < min(3, 1+len(mines)/10); i++ {
		if i >= len(mines) {
			break
		}
		o := mines[(i*3)%len(mines)]
		s.addMover(smMiningDrone, lineCells(pt(o.X, o.Y+1), drop), i, false, 8)
	}
}
