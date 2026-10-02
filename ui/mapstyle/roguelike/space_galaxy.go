package roguelike

import (
	"math"

	"github.com/espresso20/ageforge/mapmodel"
)

// space_galaxy.go is the Galactic Age: the starbase hub. The galaxy's arms
// sweep behind it and full-colour nebulae drift across the dark. At the
// centre the starbase: a gold core, a white habitat ring, a gold docking
// ring strung with cyan running lights, and six pylons where the fleet
// docks. The habitat lineages crowd the decks between the rings, trade and
// the embassies the docking ring, engineering the pylons; industry works a
// pulsar of its own in the corner, its taps feeding the base. Every civ met
// is a star system on the rim with a trade lane to the base. Starships come
// and go and jump to warp in a streak of light; the alien saucers, once a
// rare sight, are ordinary traffic now, each species in its own colour.
// Transporter sparkles wink round the base.

// Starbase layout, in plane cells.
const (
	galX, galY         = skyW / 2, 20
	galInRX, galInRY   = 11.0, 4.2 // the habitat ring
	galOutRX, galOutRY = 21.0, 7.5 // the docking ring
	galPulsarX         = 14.0
	galPulsarY         = 34.0
)

// galSystems are the star systems' sites on the rim, by civ.
var galSystems = []mapmodel.Pt{{X: 6, Y: 4}, {X: 26, Y: 3}, {X: 44, Y: 4}, {X: 88, Y: 4}, {X: 108, Y: 3},
	{X: 126, Y: 6}, {X: 124, Y: 17}, {X: 126, Y: 29}, {X: 108, Y: 38}, {X: 84, Y: 39}, {X: 42, Y: 38}}

var (
	galInner = []string{mapmodel.LinHousing, mapmodel.LinFood, mapmodel.LinKnowledge, mapmodel.LinStorage,
		mapmodel.LinCulture, mapmodel.LinFaith}
	galOuter = []string{mapmodel.LinTrade, mapmodel.LinHarbor, mapmodel.LinHacker, mapmodel.LinDiplomacy,
		mapmodel.LinMonument}
)

// layGalaxy lays the Galactic Age's base.
func (b *skyBase) layGalaxy() {
	b.spiral()
	b.nebulae()
	b.starbase()
	b.pulsar()
	b.systems = galSystems
}

// spiral draws the galaxy behind everything: two arms winding out of a
// core hidden behind the starbase, stippled, brighter toward the middle.
func (b *skyBase) spiral() {
	cx, cy := float64(galX)+6, float64(galY)+1
	for y := 0; y < skyH; y++ {
		for x := 0; x < skyW; x++ {
			if !b.free(x, y) {
				continue
			}
			dx, dy := float64(x)-cx, 2*(float64(y)-cy)
			r := math.Sqrt(dx*dx + dy*dy)
			if r < 3 {
				continue
			}
			ph := 2*turns(dx, dy) - 0.85*mapmodel.Log2(r)
			arm := 0.5 + 0.5*mapmodel.Cos(ph)
			fall := 1 - r/150
			v := arm * arm * fall
			h := mapmodel.HashF(b.seed, 571, int64(x), int64(y))
			switch {
			case v > 0.62 && h < 0.35:
				b.put(x, y, skyCell{r: '░', ink: mapmodel.InkHaze, lv: 2, k: skSpiral})
			case v > 0.5 && h < 0.5:
				b.put(x, y, skyCell{r: '·', ink: mapmodel.InkRock, lv: 1, k: skSpiral})
			case v > 0.38 && h < 0.22:
				b.put(x, y, skyCell{r: '·', ink: mapmodel.InkHaze, lv: 2, k: skSpiral})
			}
		}
	}
}

// nebulae tints three soft clouds round the rim, magenta, blue and gold,
// fading into a fainter haze at their edges.
func (b *skyBase) nebulae() {
	clouds := [3]struct {
		cx, cy, rx, ry float64
		ink            mapmodel.SkyInk
	}{
		{17, 13, 15, 5.5, mapmodel.InkCloud}, {118, 31, 15, 6, mapmodel.InkSurface}, {102, 7, 12, 3.5, mapmodel.InkSurface2},
	}
	for i, c := range clouds {
		for y := int(c.cy - c.ry - 2); y <= int(c.cy+c.ry+2); y++ {
			for x := int(c.cx - c.rx - 3); x <= int(c.cx+c.rx+3); x++ {
				if !b.in(x, y) {
					continue
				}
				d := sq((float64(x)-c.cx)/c.rx) + sq((float64(y)-c.cy)/c.ry)
				n := mapmodel.Noise(b.seed+int64(580+i), float64(x)/5, float64(y)/2)
				d += 0.7 * (n - 0.5)
				if d > 1 {
					continue
				}
				cell := b.at(x, y)
				if cell.k != skVoid && cell.k != skStar && cell.k != skSpiral {
					continue
				}
				cell.bg, cell.soft = c.ink, true
				if d > 0.55 {
					cell.bg = mapmodel.InkHaze
				}
				if cell.k == skVoid {
					cell.k = skNebula
				}
				if n > 0.7 && d < 0.5 && cell.r == ' ' {
					cell.r, cell.ink, cell.lv, cell.k = '░', c.ink, 1, skNebula
				}
			}
		}
	}
}

// starbase draws the hub: the core, the habitat ring, the docking ring with
// its running lights, the spokes between them and the six pylons; and it
// reserves the decks between the rings for the habitat lineages, the
// berths outside the docking ring for trade and the embassies, the pylons
// for engineering and the berths at their tips for the fleet.
func (b *skyBase) starbase() {
	cx, cy := float64(galX), float64(galY)
	used := map[mapmodel.Pt]bool{}
	ringOf := func(rx, ry float64, ink mapmodel.SkyInk, fr int32, lights bool) map[mapmodel.Pt]bool {
		ps, rs := smoothRing(cx, cy, rx, ry)
		on := map[mapmodel.Pt]bool{}
		for i, p := range ps {
			c := skyCell{r: rs[i], ink: ink, lv: 2, k: skFrame, ref: fr, bold: true}
			if lights && i%7 == 3 {
				c.r, c.ink, c.fx, c.ph, c.lv = '∙', mapmodel.InkAccent2, fxBlink, uint8(i), 3
			}
			b.put(p.X, p.Y, c)
			on[p], used[p] = true, true
		}
		return on
	}
	core := b.frame("Starbase core", "where the base's heart beats", slStarbase, "")
	hab := b.frame("Habitat ring", "decks of homes, farms and halls", slStarbase, "")
	dock := b.frame("Docking ring", "the base's busy edge", slStarbase, "")
	pyl := b.frame("Docking pylon", "ships tie up at its tip", slPylon, mapmodel.LinEngineer)
	ringOf(4, 1.8, mapmodel.InkAccent, core, false)
	inner := ringOf(galInRX, galInRY, mapmodel.InkFrame, hab, false)
	outer := ringOf(galOutRX, galOutRY, mapmodel.InkAccent, dock, true)
	b.hubAt(galX, galY)
	used[pt(galX, galY)] = true
	// the core's glow, and three bridges, core to docking ring, faintly
	for _, p := range discCells(cx, cy, 3.2, 1.2) {
		if c := b.at(p.X, p.Y); c.k != skHub && c.k != skFrame {
			b.put(p.X, p.Y, skyCell{r: ' ', bg: mapmodel.InkAccent, soft: true, k: skFrame, ref: core})
			used[p] = true
		}
	}
	for k := 0; k < 3; k++ {
		t := float64(k)/3 + 1.0/6
		end := pt(int(math.Round(cx+2*galOutRY*mapmodel.Sin(t)*galOutRX/(2*galOutRY))), int(math.Round(cy-galOutRY*mapmodel.Cos(t))))
		for i, p := range lineCells(pt(galX, galY), end)[1:] {
			if c := b.at(p.X, p.Y); c.k != skFrame && c.k != skHub && i%2 == 0 {
				b.put(p.X, p.Y, skyCell{r: '·', ink: mapmodel.InkFrameDim, lv: 2, k: skFrame, ref: hab})
				used[p] = true
			}
		}
	}
	// six pylons, up, down and the four diagonals, each a clean line out
	// from the docking ring: engineering's cells along them, the fleet's
	// berths round the tips
	var pylon, fleet []mapmodel.Pt
	for _, d := range [6][2]int{{0, -1}, {1, -1}, {1, 1}, {0, 1}, {-1, 1}, {-1, -1}} {
		p := pt(galX, galY)
		for inEllipse(p.X, p.Y, cx, cy, galOutRX, galOutRY) || outer[p] {
			p = pt(p.X+d[0]*(1+abs(d[1])*abs(d[0])), p.Y+d[1])
		}
		r := '│'
		if d[0]*d[1] < 0 {
			r = '╱'
		} else if d[0] != 0 {
			r = '╲'
		}
		var cells []mapmodel.Pt
		for i := 0; i < 6; i++ {
			cells = append(cells, p)
			p = pt(p.X+d[0], p.Y+d[1])
		}
		for i, q := range cells {
			b.put(q.X, q.Y, skyCell{r: r, ink: mapmodel.InkFrameDim, lv: 2, k: skFrame, ref: pyl, bold: true})
			used[q] = true
			if i%2 == 1 {
				pylon = append(pylon, q)
			}
		}
		tip := cells[len(cells)-1]
		b.put(tip.X, tip.Y, skyCell{r: '◆', ink: mapmodel.InkAccent, lv: 3, k: skFrame, ref: pyl, bold: true})
		for _, o := range [6][2]int{{-2, 0}, {2, 0}, {-1, -1}, {1, -1}, {-1, 1}, {1, 1}} {
			fleet = append(fleet, pt(tip.X+o[0], tip.Y+o[1]))
		}
	}
	// the deck floor between the rings
	for y := galY - 9; y <= galY+9; y++ {
		for x := galX - int(galOutRX) - 1; x <= galX+int(galOutRX)+1; x++ {
			if inEllipse(x, y, cx, cy, galOutRX, galOutRY) && !inEllipse(x, y, cx, cy, galInRX+0.5, galInRY+0.5) &&
				!used[pt(x, y)] && b.free(x, y) {
				b.put(x, y, skyCell{r: ' ', bg: mapmodel.InkFrameDim, soft: true, k: skFrame, ref: hab})
			}
		}
	}
	// decks between the rings, two layers out from the habitat ring
	var decks [2][]mapmodel.Pt
	in := map[mapmodel.Pt]bool{}
	for p := range inner {
		in[p] = true
	}
	for k := range decks {
		for _, p := range halo(in, cx, cy, galInRX, galInRY) {
			in[p] = true
			if !used[p] && inEllipse(p.X, p.Y, cx, cy, galOutRX-0.6, galOutRY-0.6) {
				decks[k] = append(decks[k], p)
				used[p] = true
			}
		}
	}
	for i, lin := range galInner {
		for _, d := range decks {
			b.addBand(lin, centreOut(sector(rotate(d, len(d)-len(d)/12), i, len(galInner))))
		}
	}
	// berths outside the docking ring, three layers
	out := map[mapmodel.Pt]bool{}
	for p := range outer {
		out[p] = true
	}
	for k := 0; k < 3; k++ {
		var layer []mapmodel.Pt
		for _, p := range halo(out, cx, cy, galOutRX, galOutRY) {
			out[p] = true
			if !used[p] && b.free(p.X, p.Y) {
				layer = append(layer, p)
			}
		}
		for i, lin := range galOuter {
			b.addBand(lin, centreOut(sector(rotate(layer, len(layer)-len(layer)/10), i, len(galOuter))))
		}
	}
	b.addBand(mapmodel.LinEngineer, pylon)
	var keep []mapmodel.Pt
	for _, p := range fleet {
		if b.free(p.X, p.Y) && !used[p] {
			keep = append(keep, p)
			used[p] = true
		}
	}
	b.addBand(mapmodel.LinMilitary, keep)
	// transporter sparkles wink in the void round the base
	for i := 0; i < 70; i++ {
		h := mapmodel.Hash(b.seed, 575, int64(i))
		p := pt(galX-30+int(h%61), galY-11+int(h>>8%23))
		if c := b.at(p.X, p.Y); (c.k == skVoid || c.k == skNebula) && c.r == ' ' {
			c.fx, c.ph, c.ink = fxSparkle, uint8(h>>16), mapmodel.InkMoon
		}
	}
}

// strokeOf is the line stroke for a direction t turns clockwise from up.
func strokeOf(t float64) rune {
	t -= math.Floor(t)
	switch k := int(math.Round(t * 8)); k % 4 {
	case 0:
		return '│'
	case 1:
		return '╱'
	case 2:
		return '─'
	}
	return '╲'
}

// pulsar draws industry's own star in the corner, a pulsar with its beams,
// and reserves its works on three orbits (mines, metal, organic) and the
// energy taps along a conduit to the base.
func (b *skyBase) pulsar() {
	px, py := galPulsarX, galPulsarY
	fr := b.frame("The pulsar", "industry works its light and its matter", slSun, "")
	b.put(int(px), int(py), skyCell{r: '✦', ink: mapmodel.InkStarBright, lv: 3, k: skSun, fx: fxBlink, bold: true, ref: fr})
	for k := 1; k <= 3; k++ { // the beams, flashing
		for _, d := range [2][2]int{{1, -1}, {-1, 1}} {
			x, y := int(px)+d[0]*2*k, int(py)+d[1]*k
			b.put(x, y, skyCell{r: '·', ink: mapmodel.InkAccent2, lv: 2, k: skSun, fx: fxBlink, ph: uint8(k), ref: fr})
		}
	}
	lins := []string{mapmodel.LinMines, mapmodel.LinMetal, mapmodel.LinWood}
	for i, o := range [3][2]float64{{5.5, 2.2}, {8.5, 3.4}, {11.5, 4.6}} {
		ps, _ := smoothRing(px, py, o[0], o[1])
		var keep []mapmodel.Pt
		for j, p := range ps {
			if !b.free(p.X, p.Y) {
				continue
			}
			if j%2 == 0 {
				b.put(p.X, p.Y, skyCell{r: '·', ink: mapmodel.InkFrameDim, lv: 0, k: skOrbit})
			}
			keep = append(keep, p)
		}
		b.addBand(lins[i], keep)
	}
	ps, _ := smoothRing(px, py, 14.5, 5.8)
	var taps []mapmodel.Pt
	for _, p := range ps {
		if b.free(p.X, p.Y) {
			taps = append(taps, p)
		}
	}
	b.addBand(mapmodel.LinEnergy, taps)
}

// galaxyLife puts what changes with play on the Galactic Age: the star
// systems of the civs met and their lanes to the base, the Cosmic Beacon
// lit on the core, and the traffic.
func (s *skyScene) galaxyLife() {
	m, b := s.m, s.b
	for _, w := range m.Wonders {
		if w.Key == "cosmic_beacon" && w.Built {
			*s.cell(galX, galY) = skyCell{r: '✦', ink: mapmodel.InkAccent, lv: 3, k: skHub, bold: true, fx: fxBlink}
		}
	}
	route := map[string]bool{}
	for _, r := range m.Routes {
		if !r.Disrupted && r.Civ != "" {
			route[r.Civ] = true
		}
	}
	for fi := range m.Factions {
		f := &m.Factions[fi]
		if !f.Discovered || fi >= len(b.systems) {
			continue
		}
		site := b.systems[fi]
		ink := inkRel + mapmodel.SkyInk(f.Relation)
		// the lane first, so the system draws over its end
		edge := pt(galX+int(float64(site.X-galX)*0.42), galY+int(float64(site.Y-galY)*0.42))
		for i, p := range lineCells(edge, site) {
			if i%2 == 1 || !s.openAt(p.X, p.Y) {
				continue
			}
			c := skyCell{r: '·', ink: mapmodel.InkAccent, lv: 1, k: skLane, ref: int32(fi)}
			switch {
			case f.Relation == mapmodel.RelWar:
				c.ink = mapmodel.InkAccent3
			case route[f.Key]:
				c.r, c.lv = '∙', 2
			}
			*s.cell(p.X, p.Y) = c
		}
		*s.cell(site.X, site.Y) = skyCell{sym: mapmodel.SymStarBig, ink: ink, lv: 3, k: skCiv, ref: int32(fi), bold: true}
		for k, d := range [3][2]int{{-2, 0}, {2, 1}, {1, -1}} {
			if k < 1+f.Strength/2 {
				if c := s.cell(site.X+d[0], site.Y+d[1]); s.openAt(site.X+d[0], site.Y+d[1]) {
					*c = skyCell{r: '·', ink: ink, lv: 2, k: skCiv, ref: int32(fi)}
				}
			}
		}
		s.labels = append(s.labels, skyLabel{p: site, text: f.Name, rel: int(f.Relation) + 1})
	}
	s.galaxyTraffic(route)
}

// galaxyTraffic: starships out from the pylons to the rim, jumping to warp
// at the end; freighters on the lanes with running trade; and the aliens,
// one saucer of each species crossing, more in a busy age.
func (s *skyScene) galaxyTraffic(route map[string]bool) {
	b, m := s.b, s.m
	busy := m.Activity.Traffic
	n := 0
	for k := 0; k < 2+int(busy*3); k++ {
		t := float64(k)/5 + 0.1 + mapmodel.HashF(b.seed, 640, int64(k))*0.15
		from := pt(galX+int(2*(galOutRY+8)*mapmodel.Sin(t)), galY-int((galOutRY+8)*mapmodel.Cos(t)))
		to := pt(galX+int(2*40*mapmodel.Sin(t)), galY-int(40*mapmodel.Cos(t)))
		if mv := s.addMover(smStarship, lineCells(from, to), n, true, 30+int(mapmodel.Hash(b.seed, 641, int64(k))%90)); mv != nil {
			mv.warp = 14
			n++
		}
	}
	for fi := range m.Factions {
		f := &m.Factions[fi]
		if !f.Discovered || !route[f.Key] || fi >= len(b.systems) {
			continue
		}
		site := b.systems[fi]
		edge := pt(galX+int(float64(site.X-galX)*0.42), galY+int(float64(site.Y-galY)*0.42))
		s.addMover(smStarship, lineCells(edge, site), n, false, 12)
		n++
	}
	for k := 0; k < 3+int(busy*2); k++ {
		y := 2 + int(mapmodel.Hash(b.seed, 645, int64(k))%uint64(skyH-4))
		lane := lineCells(pt(-3, y), pt(skyW+2, y+int(mapmodel.Hash(b.seed, 646, int64(k))%7)-3))
		if k%2 == 1 {
			lane = reversed(lane)
		}
		if mv := s.addMover(smAlien, lane, k, true, 80+int(mapmodel.Hash(b.seed, 647, int64(k))%200)); mv != nil {
			mv.alien = k % len(mapmodel.Aliens)
		}
	}
}
