package mapgen

import "github.com/espresso20/ageforge/mapmodel"

// Orbit swaps the ground for space. The planet is a big ellipse low in the
// world, its limb curving across the bottom of the picture. Orbits are
// concentric elliptical rings round it, drawn as ring roads, joined by one
// space elevator that rises from the planet to a hub. Stations stand on the
// rings the way buildings stand on streets, so the same placement, roads,
// joins, traffic and light serve both scenes.
const (
	planetX, planetY   = worldW / 2, 78
	planetRX, planetRY = 58, 24
	ringCount          = 7
	stationW, stationH = 6, 3
)

func ringRX(ring int) int { return planetRX + 8*ring }
func ringRY(ring int) int { return planetRY + 2 + 3*ring }

// NewOrbitPlan lays out the rings, the elevator, the wonder plots and the
// station parcels, and hands the parcels to the building lines.
func NewOrbitPlan(seed int64, cat *mapmodel.Catalog) *Plan {
	b := newPlanner(seed)
	p := b.plan
	p.Orbit = true
	hub := Pt{planetX, planetY - ringRY(ringCount) - 1}
	p.Centre = Pt{planetX, planetY - 26}
	p.Roads = []Road{{Cells: []Pt{hub}, Parent: -1}}
	elevator := Road{Parent: 0, Tier: 1}
	for y := hub.Y + 1; y < planetY-planetRY; y++ {
		elevator.Cells = append(elevator.Cells, Pt{planetX, y})
	}
	p.Roads = append(p.Roads, elevator)
	for _, c := range append([]Pt{hub}, elevator.Cells...) {
		b.claimRoad(c, 1)
	}
	for ring := 1; ring <= ringCount; ring++ {
		crossing := planetY - ringRY(ring) - hub.Y - 1
		for _, side := range []int{1, -1} {
			road := Road{Cells: arc(ring, side), Parent: 1, From: crossing}
			for _, c := range road.Cells {
				b.claimRoad(c, len(p.Roads))
			}
			p.Roads = append(p.Roads, road)
		}
	}
	b.reserveStations()
	b.cutStations()
	p.assign(seed, cat, 48, 16)
	return p
}

// ringOf is which ring a road index belongs to (roads 2 and 3 are ring 1).
func ringOf(road int) int { return (road-2)/2 + 1 }

// arc is the upper half of a ring from the elevator out to one side, one
// step at a time so every cell touches the one before: along a row, then up
// or down to the next height.
func arc(ring, side int) []Pt {
	rx, ry := ringRX(ring), ringRY(ring)
	var cells []Pt
	height := ry
	for dx := 1; dx <= rx; dx++ {
		x := planetX + side*dx
		if x < 0 || x >= worldW {
			break
		}
		next := isqrt(ry * ry * (rx*rx - dx*dx) / (rx * rx))
		for h := height; h >= next; h-- {
			cells = append(cells, Pt{x, planetY - h})
		}
		height = next
	}
	return cells
}

// stationRoom says whether a station parcel fits on ring n: free of other
// parcels and of every road but its own ring.
func (b *planner) stationRoom(r Rect, ring int) bool {
	ok := true
	r.cells(func(x, y int) {
		switch {
		case !inWorld(x, y) || b.claim[at(x, y)] == lotCell:
			ok = false
		case b.claim[at(x, y)] == roadCell && (b.roadOf[at(x, y)] < 2 || ringOf(b.roadOf[at(x, y)]) != ring):
			ok = false
		}
	})
	return ok
}

// reserveStations keeps a plot on a ring for each wonder of the orbit ages: a
// room on the ring, the ring running through it, wide and tall enough for a
// template.
func (b *planner) reserveStations() {
	p := b.plan
	p.Plots = make([]Lot, 22)
	for age := groundAges; age < len(p.Plots); age++ {
		k := age - groundAges
		road := 2 + 2*(1+k%3) + k%2
		for i := 8 + k/3*14; i < len(p.Roads[road].Cells); i++ {
			c := p.Roads[road].Cells[i]
			rect := Rect{c.X - plotW/2, c.Y - 2, plotW, 5}
			if b.stationRoom(rect, ringOf(road)) {
				b.take(rect)
				p.Plots[age] = Lot{Rect: rect, Door: c, Road: road, At: i}
				break
			}
		}
	}
}

// cutStations sets parcels along every ring, packing each in as close as
// the last allows, clear of the elevator.
func (b *planner) cutStations() {
	p := b.plan
	for road := 2; road < len(p.Roads); road++ {
		next := 3
		for i, c := range p.Roads[road].Cells {
			if i < next {
				continue
			}
			rect := Rect{c.X - stationW/2, c.Y - 1, stationW, stationH}
			if abs(c.X-planetX) < 5 || !b.stationRoom(rect, ringOf(road)) {
				next = i + 1
				continue
			}
			b.take(rect)
			p.Lots = append(p.Lots, Lot{Rect: rect, Door: c, Road: road, At: i})
			next = i + stationW + 1
		}
	}
}

// rim is 1000 on the planet's edge, less inside it and more outside, on the
// ellipse the planet is.
func rim(x, y int) int {
	dx, dy := x-planetX, y-planetY
	return (dx*dx*planetRY*planetRY + dy*dy*planetRX*planetRX) * 1000 / (planetRX * planetRX * planetRY * planetRY)
}

// paintSpace draws the void, a drift of nebula, the stars and the planet
// with its air.
func (f *frame) paintSpace() {
	for y := 0; y < worldH; y++ {
		for x := 0; x < worldW; x++ {
			r := rim(x, y)
			switch {
			case r < 1000:
				f.paintPlanet(x, y, r)
			case r < 1220:
				f.paintAir(x, y, r)
			default:
				f.paintVoid(x, y)
			}
		}
	}
	f.paintStars()
}

func (f *frame) paintVoid(x, y int) {
	sky := f.st.Sky
	haze := max(0, fbm(f.m.Seed+11, x*256/24, y*256/12, 3)-540) * 90 / (noiseMax - 540)
	bg := sky.Bg
	if haze > 0 {
		bg = P(HNebula, haze)
	}
	f.g.paint(x, y, Cell{R: ' ', Fg: sky.Fg, Bg: bg})
	f.g.mark(x, y, Thing{Kind: KindTerrain, Key: "space"})
}

// paintStars puts one candidate star in each small block of the sky.
func (f *frame) paintStars() {
	sky := f.st.Sky
	for by := 0; by < worldH/3; by++ {
		for bx := 0; bx < worldW/5; bx++ {
			x := bx*5 + roll(5, f.m.Seed, 12, int64(bx), int64(by))
			y := by*3 + roll(3, f.m.Seed, 13, int64(bx), int64(by))
			if rim(x, y) < 1220 || !chance(sky.Dense, f.m.Seed, 14, int64(bx), int64(by)) {
				continue
			}
			h := hash(f.m.Seed, 15, int64(x), int64(y))
			shine := 120 + int(h>>8)%136
			if (f.anim/9+int(h>>20))%7 == 0 {
				shine /= 2 // a slow twinkle
			}
			f.g.draw(x, y, pick(sky.Glyphs, h>>30), sky.Fg.fade(shine))
		}
	}
}

// paintPlanet draws the surface: sea, land and drifting cloud, darkening
// toward the limb; at night the land shows the lights of its cities.
func (f *frame) paintPlanet(x, y, r int) {
	seed := f.m.Seed
	look := f.st.Land[Deep]
	land := false
	if fbm(seed+7, x*256/10, y*256/5, 3) > 540 {
		look, land = f.st.Land[Grass], true
	}
	if fbm(seed+9, x*256/14+f.anim*6, y*256/6, 3) > 610 {
		look, land = f.st.Land[Hills], false
	}
	c := look.at(x, y, 0)
	city := land && r < 900 && chance(6, seed, 16, int64(x), int64(y))
	if city && f.night >= 128 {
		c.R = '∙'
	}
	edge := clamp((r-760)*256/240, 0, 230)
	c.Dim = uint8(edge)
	f.g.paint(x, y, c)
	f.g.mark(x, y, Thing{Kind: KindTerrain, Key: "planet"})
	if city {
		f.lights = append(f.lights, Pt{x, y})
	}
}

// paintAir draws the thin glow of atmosphere round the planet.
func (f *frame) paintAir(x, y, r int) {
	air := f.st.Land[Sand]
	thin := (r - 1000) * 256 / 220
	c := Cell{R: ' ', Fg: air.Fg.fade(256 - thin), Bg: air.Bg.fade(256 - thin)}
	if r < 1060 {
		c.R = pick(air.Glyphs, hash(int64(x), int64(y)))
	}
	f.g.paint(x, y, c)
	f.g.mark(x, y, Thing{Kind: KindTerrain, Key: "atmosphere"})
}

// paintRelics shows the wonders of the ground ages, which have no room on
// the rings, as bright points on the planet: a wonder's star.
func (f *frame) paintRelics() {
	for _, w := range f.m.Wonders {
		if !f.plan.Orbit || w.Age >= groundAges || !w.Built {
			continue
		}
		for try := 0; try < 40; try++ {
			x := planetX - 36 + roll(73, f.m.Seed, 17, int64(w.Age), int64(try))
			y := planetY - planetRY + 3 + roll(9, f.m.Seed, 18, int64(w.Age), int64(try))
			if rim(x, y) < 880 {
				f.g.draw(x, y, '★', f.st.Tones['a'])
				f.g.mark(x, y, Thing{Kind: KindWonder, Key: w.Key})
				f.lights = append(f.lights, Pt{x, y})
				break
			}
		}
	}
}
