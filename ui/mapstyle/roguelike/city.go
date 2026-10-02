package roguelike

import (
	"math"
	"sort"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
)

// city.go is the Earth arc's land, the Modern Age to the Fusion Age: what
// the city makes of every cell of land round the town. The town itself
// still grows from the real buildings; the city is the land between and
// beyond them, read off the seed's world and the age alone, so it is the
// same for the same save and never shuffles as the town grows.
//
//   - Modern: a glass belt round the ring boulevard, suburbs of houses and
//     gardens beyond it, parks fenced into their blocks, highways out of
//     town, and the countryside past the suburbs.
//   - Information: the city covers most of the land: office blocks, server
//     farms, dishes and fiber lines. Half the green is left, as parks and
//     rooftop gardens, and smog gathers at the city's edge.
//   - Digital: the last green is one walled reserve (15% of the Modern
//     green); the rest is concrete, data halls and, past the city, landfill
//     under the smog.
//   - Cyberpunk: the megacity. Every cell is built or derelict: megablocks,
//     megacorp towers with their initials in neon, arcologies, neon signs
//     and holo-ads, vents, scrap and toxic ground; toxic canals for water;
//     three to five sky rails over it all.
//   - Fusion: the megacity powered: reactors in their plasma rings, power
//     conduits, maglev lines, launch towers and the space elevator's
//     tether, with the smog thinned and the water running clean and cold.
//
// The land is laid once per scene (cityLand) and a frame only reads it.

// use is what the city has made of a cell of land.
type use uint8

const (
	uNone        use = iota // the land as it was: terrain draws it
	uStreet                 // a city street between blocks
	uYard                   // a suburban garden
	uHouse                  // a suburban house
	uPark                   // a park, fenced into its block
	uFence                  // a park's fence
	uGlass                  // a glass tower
	uOffice                 // an office block
	uServer                 // a server farm's racks
	uDish                   // a satellite dish
	uGarden                 // a rooftop garden
	uScrub                  // dry ground past the city: the green gone
	uConcrete               // bare concrete
	uLandfill               // landfill
	uDataHall               // a data hall (it glows)
	uReserve                // the reserve's old forest
	uReserveWall            // the reserve's wall
	uAlley                  // a megacity street
	uTrash                  // trash in an alley
	uMega                   // a megablock's roof
	uLit                    // a lit window on it (aux: the ink)
	uNeon                   // a neon sign (aux: the glyph and the ink)
	uAd                     // a holo-ad (aux: the ad and the letter)
	uCorp                   // a megacorp tower's initial (aux: the corp and the letter)
	uArco                   // an arcology (aux: how deep in it)
	uVent                   // a steam vent
	uScrap                  // a scrap heap
	uToxic                  // toxic ground
	uSlag                   // a strip-mined peak
	uQuay                   // a canal's concrete edge
	uYardPower              // a power yard (the Fusion Age's derelict blocks, rebuilt)
	uReactor                // a reactor's core
	uRing                   // its plasma ring (aux: which part)
	uLaunch                 // a launch tower (aux: which part)
	numUses
)

// green reports a use the greenery fade counts.
func (u use) green() bool { return u == uYard || u == uPark || u == uGarden || u == uReserve }

// useFeature is the city feature a use belongs to (its legend row and its
// inspect title), FeatNone for the land as it was.
var useFeature = [numUses]mapmodel.CityFeature{
	uStreet: mapmodel.FeatNone, uYard: mapmodel.FeatSuburbs, uHouse: mapmodel.FeatSuburbs,
	uPark: mapmodel.FeatPark, uFence: mapmodel.FeatPark, uGlass: mapmodel.FeatGlassTower,
	uOffice: mapmodel.FeatOffice, uServer: mapmodel.FeatServerFarm, uDish: mapmodel.FeatDish,
	uGarden: mapmodel.FeatGarden, uScrub: mapmodel.FeatScrub, uConcrete: mapmodel.FeatConcrete,
	uLandfill: mapmodel.FeatLandfill, uDataHall: mapmodel.FeatDataGlow, uReserve: mapmodel.FeatReserve,
	uReserveWall: mapmodel.FeatReserve, uAlley: mapmodel.FeatAlley, uTrash: mapmodel.FeatAlley,
	uMega: mapmodel.FeatMegablock, uLit: mapmodel.FeatMegablock, uNeon: mapmodel.FeatNeonSign,
	uAd: mapmodel.FeatNeonSign, uCorp: mapmodel.FeatCorpTower, uArco: mapmodel.FeatArcology,
	uVent: mapmodel.FeatVent, uScrap: mapmodel.FeatScrap, uToxic: mapmodel.FeatToxic,
	uSlag: mapmodel.FeatMegablock, uQuay: mapmodel.FeatMegablock, uYardPower: mapmodel.FeatConduit,
	uReactor: mapmodel.FeatReactor, uRing: mapmodel.FeatReactor, uLaunch: mapmodel.FeatLaunchTower,
}

// cityLine is a line laid over the city: a highway, a fiber line, a sky
// rail or maglev line, a power conduit or the tether. Its points run in
// order, so its traffic can ride it.
type cityLine struct {
	feat  mapmodel.CityFeature
	pts   []mapmodel.Pt
	ink   theme.CityHue
	elev  bool   // elevated: it crosses streets, walls and the railway too
	layer uint8  // which line shows where two cross
	name  string // "the magenta line"
}

// corp is a megacorp with a tower in the megacity.
type corp struct{ name, initials string }

// megacorps are the megacity's corporations (invented; none is a civ).
var megacorps = []corp{
	{"Kessler-Oda Group", "KOG"}, {"Novatek Holdings", "NVT"}, {"Helix Dynamics", "HXD"},
	{"Ombra Industries", "OMB"}, {"Sunrise Heavy Industries", "SHI"}, {"Vektor Systems", "VKS"},
	{"Halcyon Data", "HCD"}, {"Mirai Kogyo", "MKY"}, {"Black Lotus Holdings", "BLH"},
	{"Zenith Arms", "ZNA"}, {"Pacifica Telecom", "PFT"}, {"Talos Biotech", "TLB"},
	{"Yamanaka Heavy", "YMH"}, {"Aurum Logistics", "AUL"}, {"Crimson Spire Media", "CSM"},
	{"Kuroda Neural", "KRN"},
}

// holoAds are the holo-ads' words: four letters at most, a block wide.
var holoAds = []string{"OPEN", "24H", "SALE", "LIVE", "BAR", "CHIP", "AUGS", "SOMA", "HOT", "BUY", "NEW", "VR"}

// neonRunes are the neon signs' glyphs (each folds to ASCII).
var neonRunes = []rune("¥£ƒΣΩ∞≡♦§¤$%&")

// cityLand is the Earth arc's city over one scene.
type cityLand struct {
	look mapmodel.CityLook
	key  string // the look's age key
	use  []use
	aux  []uint8
	ov   []uint8 // the top line on a cell, as its index + 1
	ovm  []uint8 // the box mask of the lines on a cell (N1 E2 S4 W8)
	glow []uint8 // the data glow, 0..3
	smog []uint8 // the smog, 0..3
	mark []bool  // a structure the region zoom shows
	// lines is every line laid over the city, oldest kinds first.
	lines []cityLine
	// vents are the steam vents, reactors the reactor cores, for the
	// ambient effects.
	vents, reactors []mapmodel.Pt
	// tether is the space elevator, from the square to the top of the map.
	tether int // index into lines + 1, 0 for none
	// win is the window the greenery fade is measured over (x0, y0, x1,
	// y1, inclusive): about what the settlement zoom shows round the square;
	// measure marks the cells of it the measure reads: open land the town
	// can never build on (no lot of any quarter, no wonder plot, no street
	// of the plan, not the railway), so a new building never shifts it.
	win     [4]int
	measure []bool
	// lots marks where a building or a landmark can ever stand, from the
	// seed's plan: every quarter lot and wonder plot, the square, and the
	// railway. The city lays its structures and its reserve off them, so
	// the town's growth never meets either.
	lots []bool
	// lanes are the megacity's alley lanes (citytraffic.go).
	lanes     [][]mapmodel.Pt
	lanesDone bool
}

// Greenery is measured over the settlement view round the square.
const (
	greenWinW = 132
	greenWinH = 42
)

func greenWindow(w *mapmodel.World) [4]int {
	x0 := clamp(w.CX-greenWinW/2, 0, max(0, w.W-greenWinW))
	y0 := clamp(w.CY-greenWinH/2, 0, max(0, w.H-greenWinH))
	return [4]int{x0, y0, min(w.W-1, x0+greenWinW-1), min(w.H-1, y0+greenWinH-1)}
}

func (cl *cityLand) inWin(x, y int) bool {
	return x >= cl.win[0] && x <= cl.win[2] && y >= cl.win[1] && y <= cl.win[3]
}

// townLots marks where the town can ever build (cityLand.lots).
func (s *scene) townLots() []bool {
	w, p := s.w, s.m.Town.Plan
	lots := make([]bool, w.W*w.H)
	for i := range lots {
		lots[i] = p.Plaza[i] || s.cells[i].rail
	}
	for _, q := range p.Quarter {
		for _, l := range q {
			lots[l.Y*w.W+l.X] = true
		}
	}
	for _, at := range p.WonderPlots {
		for dy := 0; dy < mapmodel.WonderH; dy++ {
			for dx := 0; dx < mapmodel.WonderW; dx++ {
				if w.In(at.X+dx, at.Y+dy) {
					lots[(at.Y+dy)*w.W+at.X+dx] = true
				}
			}
		}
	}
	return lots
}

// greenMeasure marks the cells the greenery fade is measured over: the
// land in the window, but for the plan's streets and square, its wonder
// plots and the railway. It reads the land and the seed's plan only, not
// what stands on it: a building hides the city's land without changing it.
func (s *scene) greenMeasure(cl *cityLand) []bool {
	w, p := s.w, s.m.Town.Plan
	out := make([]bool, len(cl.use))
	for y := cl.win[1]; y <= cl.win[3]; y++ {
		for x := cl.win[0]; x <= cl.win[2]; x++ {
			i := y*w.W + x
			out[i] = w.T[i].Land() && !p.Plaza[i] && !p.Street[i] && !s.cells[i].rail
		}
	}
	for _, at := range p.WonderPlots {
		for dy := 0; dy < mapmodel.WonderH; dy++ {
			for dx := 0; dx < mapmodel.WonderW; dx++ {
				if w.In(at.X+dx, at.Y+dy) {
					out[(at.Y+dy)*w.W+at.X+dx] = false
				}
			}
		}
	}
	return out
}

// gridCell is where a cell falls on the city's street grid: the town
// plan's staggered grid (blocks of 4x2 between streets), run out over the
// whole world so the city's blocks line up with the town's.
type gridCell struct {
	street bool
	bx, by int    // position in the block, 0..3 and 0..1
	id     uint64 // the block
	col    int
	band   int
}

func (s *scene) gridAt(x, y int) gridCell {
	w := s.w
	ry := y - w.CY + 1
	band := floorDiv(ry, 3)
	warp := int(mapmodel.Hash(w.Seed, 31, int64(band))%3) - 1
	gx := x - w.CX + warp + 2
	g := gridCell{street: mod(ry, 3) == 0 || mod(gx, 5) == 0, bx: mod(gx, 5) - 1, by: mod(ry, 3) - 1,
		col: floorDiv(gx, 5), band: band}
	g.id = mapmodel.Hash(w.Seed, 300, int64(band), int64(g.col))
	return g
}

func mod(a, b int) int { return ((a % b) + b) % b }

// free reports a cell of open land the city may build on: no building, no
// street, wall or trail of the town, and dry.
func (s *scene) free(i int) bool {
	return s.cells[i].k == kNone && s.w.T[i].Land()
}

// land reports a cell of land the city lays a use on: dry, and not one of
// the town's own streets, walls or trails (a building's cell takes a use
// too, hidden under it, so what the city makes of the land never depends
// on what the town has built).
func (s *scene) land(i int) bool {
	switch s.cells[i].k {
	case kNone, kTile, kWonder:
		return s.w.T[i].Land()
	}
	return false
}

// layCity lays the Earth arc's city over the scene. Ages outside the arc
// keep the land as it was.
func (s *scene) layCity() {
	look, ok := mapmodel.CityLookAt(s.m.AgeIdx)
	if !ok {
		return
	}
	s.city = s.cityFor(look)
}

// cityFor lays the city of a look over the scene: the scene's own age's,
// or (for the greenery fade's baseline) another's.
func (s *scene) cityFor(look mapmodel.CityLook) *cityLand {
	n := s.w.W * s.w.H
	cl := &cityLand{look: look, key: look.Key, use: make([]use, n), aux: make([]uint8, n), ov: make([]uint8, n),
		ovm: make([]uint8, n), glow: make([]uint8, n), smog: make([]uint8, n), mark: make([]bool, n),
		win: greenWindow(s.w)}
	cl.lots = s.townLots()
	cl.measure = s.greenMeasure(cl)
	switch look.Key {
	case "modern_age":
		s.cityModern(cl)
	case "information_age":
		s.cityInformation(cl)
	case "digital_age":
		s.cityDigital(cl)
	case "cyberpunk_age":
		s.cityMegacity(cl, false)
	case "fusion_age":
		s.cityMegacity(cl, true)
	}
	s.citySmog(cl)
	return cl
}

// ---------------------------------------------------------------- modern

// modern radii, measured from the square in rows (rdist): the glass belt
// and the suburbs beyond it.
func (s *scene) modernRadii() (glass, suburbs float64) {
	return s.wallR + 5, s.wallR + 28
}

// edgeNoise roughs a radius up, so the city's edge is ragged, not a circle.
func (s *scene) edgeNoise(x, y int, salt int64, amp float64) float64 {
	return float64(amp * (mapmodel.Noise(s.w.Seed+salt, float64(x)/9, float64(y)/4.5) - 0.5) * 2)
}

// ringDist is a cell's distance from the square, roughed up: what every age
// reads to tell the town inside its ring from the city round it.
func (s *scene) ringDist(x, y int) float64 {
	return rdist(x, y, s.w.CX, s.w.CY) + s.edgeNoise(x, y, 401, 2.5)
}

// modernUse is what the Modern Age makes of a free cell: the greenery
// fade's baseline for every later age.
func (s *scene) modernUse(x, y int) use {
	w := s.w
	i := y*w.W + x
	t := w.T[i]
	d := s.ringDist(x, y)
	glass, suburbs := s.modernRadii()
	if d < s.wallR || t == mapmodel.THills || t == mapmodel.TMountain || t == mapmodel.TBeach || d >= suburbs {
		return uNone
	}
	g := s.gridAt(x, y)
	switch {
	case g.street:
		return uStreet
	case t == mapmodel.TForest:
		return uPark
	case d < glass:
		return uGlass
	case (g.bx+g.by)%2 == 0:
		return uHouse
	}
	return uYard
}

// modernGreen reports a cell the Modern Age draws green: a park, a garden,
// or forest and grass the city has not reached.
func (s *scene) modernGreen(i int, u use) bool {
	if u.green() {
		return true
	}
	t := s.w.T[i]
	return u == uNone && (t == mapmodel.TGrass || t == mapmodel.TForest)
}

func (s *scene) cityModern(cl *cityLand) {
	w := s.w
	for i := range cl.use {
		if s.land(i) {
			cl.use[i] = s.modernUse(i%w.W, i/w.W)
		}
	}
	s.fences(cl)
	s.highways(cl)
}

// fences rails the parks in: a street cell beside a park is its fence.
func (s *scene) fences(cl *cityLand) {
	w := s.w
	for i, u := range cl.use {
		if u != uStreet || !s.free(i) {
			continue
		}
		x, y := i%w.W, i/w.W
		for _, d := range dirs4 {
			if nx, ny := x+d[0], y+d[1]; w.In(nx, ny) && cl.use[ny*w.W+nx] == uPark {
				cl.use[i] = uFence
				break
			}
		}
	}
}

// ----------------------------------------------------------- information

// cityRadius is how far the Information and Digital cities reach.
func (s *scene) cityRadius() float64 { return s.wallR + 36 }

// greenKeep returns, for each cell the Modern Age draws green, whether a
// later age keeps it: share of them inside the greenery window, by the
// quantile of a noise that clusters the kept cells into parks and speckles
// them into gardens. It reads the land, not what stands on it, so a new
// building never shifts what the city keeps.
func (s *scene) greenKeep(cl *cityLand, share float64) []bool {
	w := s.w
	keep := make([]bool, len(cl.use))
	score := make([]float64, len(cl.use))
	var inWin []float64
	for i := range cl.use {
		if !w.T[i].Land() {
			continue
		}
		x, y := i%w.W, i/w.W
		if !s.modernGreen(i, s.modernUse(x, y)) {
			continue
		}
		v := float64(0.6*mapmodel.Noise(w.Seed+511, float64(x)/7, float64(y)/3.5)) + float64(0.4*mapmodel.HashF(w.Seed, 512, int64(i)))
		score[i] = v + 1 // 0 marks a cell that was never green
		if cl.measure[i] {
			inWin = append(inWin, v)
		}
	}
	if len(inWin) == 0 {
		return keep
	}
	sort.Float64s(inWin)
	k := int(math.Round(float64(share * float64(len(inWin)))))
	for i, sc := range score {
		if sc == 0 || k == 0 {
			continue
		}
		v := sc - 1
		if k >= len(inWin) || v < inWin[k] {
			keep[i] = true
		}
	}
	return keep
}

func (s *scene) cityInformation(cl *cityLand) {
	w := s.w
	keep := s.greenKeep(cl, mapmodel.Greenery(cl.look.Age))
	glass, _ := s.modernRadii()
	reach := s.cityRadius()
	for i := range cl.use {
		if !s.land(i) {
			continue
		}
		x, y := i%w.W, i/w.W
		t := w.T[i]
		mu := s.modernUse(x, y)
		d := s.ringDist(x, y)
		inCity := d >= s.wallR && d+s.edgeNoise(x, y, 402, 3) < reach && t != mapmodel.TMountain
		grassy := t == mapmodel.TGrass || t == mapmodel.TForest
		switch {
		case keep[i] && (mu == uPark || t == mapmodel.TForest) && inCity:
			cl.use[i] = uPark
		case keep[i] && inCity:
			cl.use[i] = uGarden
		case keep[i]:
			cl.use[i] = uNone // still green: the countryside, or a lawn inside the ring
		case d < s.wallR && grassy:
			cl.use[i] = uStreet // paved inside the ring
		case !inCity && grassy:
			cl.use[i] = uScrub
		case !inCity:
			cl.use[i] = uNone
		default:
			cl.use[i] = s.officeBlock(x, y, d < glass, 16, 22, false)
		}
	}
	s.fences(cl)
	s.highways(cl)
	s.fiber(cl)
}

// officeBlock is a city block's cell in the Information and Digital
// cities: glass by the ring, else an office, a server farm (servers% of
// blocks) or a dish block (up to dishes%); paved says the Digital Age's
// concrete and data halls stand in for the dishes.
func (s *scene) officeBlock(x, y int, glass bool, servers, dishes uint64, paved bool) use {
	g := s.gridAt(x, y)
	if g.street {
		return uStreet
	}
	if glass {
		return uGlass
	}
	h := mapmodel.Hash(int64(g.id), 7) % 100
	switch {
	case h < servers:
		return uServer
	case h < dishes && paved:
		return uDataHall
	case h < dishes && (g.bx+3*g.by)%3 == 1:
		return uDish
	}
	return uOffice
}

// --------------------------------------------------------------- digital

func (s *scene) cityDigital(cl *cityLand) {
	w := s.w
	s.highways(cl) // first, so the reserve grows round them
	reserve := s.reserve(cl, mapmodel.Greenery(cl.look.Age))
	glass, _ := s.modernRadii()
	reach := s.cityRadius()
	for i := range cl.use {
		if !s.land(i) {
			continue
		}
		x, y := i%w.W, i/w.W
		t := w.T[i]
		d := s.ringDist(x, y)
		switch {
		case reserve[i] == 1:
			cl.use[i] = uReserve
		case reserve[i] == 2:
			cl.use[i] = uReserveWall
		case t == mapmodel.TMountain:
			cl.use[i] = uNone
		case d+s.edgeNoise(x, y, 403, 3) < reach:
			cl.use[i] = s.officeBlock(x, y, d < glass && d >= s.wallR, 14, 26, true)
		case mapmodel.Noise(w.Seed+404, float64(x)/6, float64(y)/3) > 0.52:
			cl.use[i] = uLandfill
		default:
			cl.use[i] = uConcrete
		}
	}
	s.dataGlow(cl)
}

// reserve finds the Digital Age's one walled park: a round stretch of the
// old green near the town, grown until it holds share of the Modern Age's
// green inside the greenery window, and the wall round it. It keeps off the
// town's quarter lots (the town can never build inside it), its plan, the
// railway and the highways, and it reads the land and the seed only, so the
// town's growth never moves it. It returns 1 for a cell inside, 2 for the
// wall.
func (s *scene) reserve(cl *cityLand, share float64) []uint8 {
	w := s.w
	out := make([]uint8, len(cl.use))
	green := make([]bool, len(cl.use))
	total := 0
	for i := range cl.use {
		if !w.T[i].Land() { // the seed's land: a trail or street the town opens never moves the target
			continue
		}
		x, y := i%w.W, i/w.W
		if s.modernGreen(i, s.modernUse(x, y)) {
			green[i] = true
			if cl.measure[i] {
				total++
			}
		}
	}
	target := int(math.Round(float64(share * float64(total))))
	if target == 0 {
		return out
	}
	open := func(i int) bool { // a cell the reserve may take: the seed's land, never a lot the town may build on
		return w.T[i].Land() && !cl.lots[i] && cl.ov[i] == 0
	}
	// the centre: the densest open green within reach of the town, far
	// enough inside the window for a round park of the target's size
	best, bc := -1, mapmodel.Pt{}
	lo, hi := s.wallR+5, s.wallR+30
	rr := int(math.Ceil(float64(1.5 * math.Sqrt(float64(target)/(2*math.Pi)))))
	for y := cl.win[1] + rr + 1; y <= cl.win[3]-rr-1; y++ {
		for x := cl.win[0] + 2*rr + 2; x <= cl.win[2]-2*rr-2; x++ {
			d := rdist(x, y, w.CX, w.CY)
			if i := y*w.W + x; !green[i] || !open(i) || d < lo || d > hi {
				continue
			}
			n := 0
			for dy := -rr; dy <= rr; dy++ {
				for dx := -2 * rr; dx <= 2*rr; dx++ {
					if j := (y+dy)*w.W + x + dx; w.In(x+dx, y+dy) && open(j) && cl.measure[j] {
						n++
					}
				}
			}
			n = 4*n + int(mapmodel.Hash(w.Seed, 405, int64(x), int64(y))%4)
			if n > best {
				best, bc = n, pt(x, y)
			}
		}
	}
	if best < 0 {
		return out
	}
	// grow a round park until it holds the target in the window
	type cand struct {
		i int
		d float64
	}
	var cs []cand
	R := 3*rr + 6
	for y := bc.Y - R; y <= bc.Y+R; y++ {
		for x := bc.X - 2*R; x <= bc.X+2*R; x++ {
			if i := y*w.W + x; w.In(x, y) && open(i) {
				cs = append(cs, cand{i, rdist(x, y, bc.X, bc.Y) +
					float64(0.8*mapmodel.Noise(w.Seed+406, float64(x)/4, float64(y)/2))})
			}
		}
	}
	sort.Slice(cs, func(a, b int) bool {
		if cs[a].d != cs[b].d {
			return cs[a].d < cs[b].d
		}
		return cs[a].i < cs[b].i
	})
	got, r := 0, 0.0
	for _, c := range cs {
		if got >= target {
			break
		}
		out[c.i], r = 1, c.d
		if cl.measure[c.i] {
			got++
		}
	}
	// the wall: open cells just outside, one thin unbroken ring
	inside := func(x, y int) bool { return w.In(x, y) && out[y*w.W+x] == 1 }
	R2 := int(r) + 3
	for y := bc.Y - R2; y <= bc.Y+R2; y++ {
		for x := bc.X - 2*R2; x <= bc.X+2*R2; x++ {
			if i := y*w.W + x; w.In(x, y) && out[i] == 0 && open(i) && onRing(x, y, inside) {
				out[i] = 2
			}
		}
	}
	return out
}

// dataGlow marks the Digital Age's data glow round its data halls and the
// town's hacker dens.
func (s *scene) dataGlow(cl *cityLand) {
	w := s.w
	var src []mapmodel.Pt
	for i, u := range cl.use {
		if u == uDataHall {
			src = append(src, pt(i%w.W, i/w.W))
		}
	}
	for _, t := range s.m.Town.Tiles {
		if t.Lineage == mapmodel.LinHacker && !t.Ruin {
			src = append(src, t.Pt)
		}
	}
	for _, p := range src {
		for dy := -1; dy <= 1; dy++ {
			for dx := -2; dx <= 2; dx++ {
				if x, y := p.X+dx, p.Y+dy; w.In(x, y) {
					i := y*w.W + x
					lv := uint8(3 - max(abs(dx)/2+abs(dy), abs(dx+dy)/2))
					if dx == 0 && dy == 0 {
						lv = 3
					}
					cl.glow[i] = max(cl.glow[i], min(lv, 3))
				}
			}
		}
	}
}

// -------------------------------------------------------------- megacity

func (s *scene) cityMegacity(cl *cityLand, fusion bool) {
	w := s.w
	for i := range cl.use {
		if !s.land(i) {
			continue
		}
		x, y := i%w.W, i/w.W
		switch w.T[i] {
		case mapmodel.TMountain:
			cl.use[i] = uSlag
			continue
		case mapmodel.TBeach:
			cl.use[i] = uQuay
			continue
		}
		s.megaBlock(cl, x, y, i, fusion)
	}
	s.arcologies(cl)
	if fusion {
		s.reactorsAndTowers(cl)
	}
	s.skyRails(cl, fusion)
	if fusion {
		s.conduits(cl)
		s.tetherLine(cl)
	}
	for i, u := range cl.use {
		if u == uVent {
			cl.vents = append(cl.vents, pt(i%w.W, i/w.W))
		}
	}
}

// megaBlock is a megacity cell: an alley, or part of a block (a megablock,
// a megacorp tower, a derelict block, a vent block, a block of neon).
func (s *scene) megaBlock(cl *cityLand, x, y, i int, fusion bool) {
	w := s.w
	g := s.gridAt(x, y)
	hc := mapmodel.Hash(w.Seed, 310, int64(i))
	if g.street {
		cl.use[i] = uAlley
		if hc%13 == 0 && !fusion || hc%29 == 0 {
			cl.use[i] = uTrash
		}
		return
	}
	hb := mapmodel.Hash(int64(g.id), 11) % 100
	derelict, corpN := uint64(12), uint64(18)
	if fusion {
		derelict = 4
	}
	edge := s.streetSide(x, y)
	switch {
	case hb < derelict:
		switch {
		case fusion:
			cl.use[i] = uYardPower
		case hc%3 == 0:
			cl.use[i] = uScrap
		case hc%3 == 1:
			cl.use[i] = uToxic
		default:
			cl.use[i] = uTrash
		}
	case hb < corpN && g.by == 0 && g.bx < 3:
		ci := int(mapmodel.Hash(int64(g.id), 12) % uint64(len(megacorps)))
		cl.use[i], cl.aux[i] = uCorp, uint8(ci<<2|g.bx)
		cl.mark[i] = true
	case hb < corpN:
		cl.use[i], cl.aux[i] = uLit, uint8(hc%4)
	case hb < corpN+4 && !fusion && (g.bx+g.by)%2 == 0:
		cl.use[i] = uVent
	case hb < corpN+12 && edge:
		cl.use[i], cl.aux[i] = uNeon, uint8(int(hc%uint64(len(neonRunes)))<<2|int(hc>>8%3))
	case hb >= 96 && g.by == 0:
		ad := int(mapmodel.Hash(int64(g.id), 13) % uint64(len(holoAds)))
		cl.use[i], cl.aux[i] = uAd, uint8(ad<<3|g.bx)
	case edge && hc%9 == 0:
		cl.use[i], cl.aux[i] = uNeon, uint8(int(hc%uint64(len(neonRunes)))<<2|int(hc>>8%3))
	case hc%4 == 0:
		cl.use[i], cl.aux[i] = uLit, uint8(hc>>4%4)
	default:
		cl.use[i] = uMega
	}
}

// streetSide reports a block cell with a street beside it (where the neon
// goes).
func (s *scene) streetSide(x, y int) bool {
	for _, d := range dirs4 {
		if s.gridAt(x+d[0], y+d[1]).street {
			return true
		}
	}
	return false
}

// arcologies raises the megacity's arcologies: a few super-blocks (two by
// two blocks and the streets between them, 9x5) round the town, terraced
// toward a lit crown.
func (s *scene) arcologies(cl *cityLand) {
	w := s.w
	n := 4 + int(mapmodel.Hash(w.Seed, 320)%3)
	for k := 0; k < n*16 && n > 0; k++ {
		a := mapmodel.HashF(w.Seed, 321, int64(k))
		r := 12 + float64(24*mapmodel.HashF(w.Seed, 322, int64(k)))
		cx := w.CX + int(float64(float64(mapmodel.Cos(a)*r)*2))
		cy := w.CY + int(float64(mapmodel.Sin(a)*r))
		// snap to a block's top-left corner, then take 9x5 from it
		g := s.gridAt(cx, cy)
		x0, y0 := cx-g.bx, cy-g.by
		if g.street {
			continue
		}
		if !s.placeable(cl, x0, y0, 9, 5) {
			continue
		}
		for dy := 0; dy < 5; dy++ {
			for dx := 0; dx < 9; dx++ {
				i := (y0+dy)*w.W + x0 + dx
				depth := min(min(dx, 8-dx)/2, min(dy, 4-dy))
				cl.use[i], cl.aux[i] = uArco, uint8(min(depth, 2))
				cl.mark[i] = depth == 2
			}
		}
		n--
	}
}

// placeable reports a w x h stretch from (x0, y0) all of free land the
// city has laid as plain megacity (nothing special on it yet).
func (s *scene) placeable(cl *cityLand, x0, y0, wd, ht int) bool {
	w := s.w
	for dy := 0; dy < ht; dy++ {
		for dx := 0; dx < wd; dx++ {
			x, y := x0+dx, y0+dy
			if !w.In(x, y) {
				return false
			}
			i := y*w.W + x
			if !w.T[i].Land() || cl.lots[i] || s.cells[i].k == kSite {
				return false // where a building may ever stand, or a civ's
			}
			switch cl.use[i] {
			case uArco, uReactor, uRing, uLaunch, uCorp, uSlag, uQuay:
				return false
			}
		}
	}
	return true
}

// reactorRing is a reactor's 9x3 footprint: its plasma ring round the core.
var reactorRing = [3]string{"╭───────╮", "│ ═◉◉◉═ │", "╰───────╯"}

// launchPad is a launch tower's 3x3 footprint: the gantry, the rocket on
// its pad and the beacon.
var launchPad = [3]string{"╫▲ ", "╫█·", "▀▀▀"}

// reactorsAndTowers sets the Fusion Age's reactors on a ring round the
// town, and two launch towers out at the city's edge.
func (s *scene) reactorsAndTowers(cl *cityLand) {
	w := s.w
	want := 4
	for k := 0; k < 120 && len(cl.reactors) < want; k++ {
		a := float64(len(cl.reactors))/float64(want) + float64(0.22*mapmodel.HashF(w.Seed, 331, int64(k)))
		r := 12 + float64(16*mapmodel.HashF(w.Seed, 332, int64(k)))
		x0 := w.CX + int(float64(float64(mapmodel.Cos(a)*r)*2)) - 4
		y0 := w.CY + int(float64(mapmodel.Sin(a)*r)) - 1
		if !s.placeable(cl, x0, y0, 9, 3) {
			continue
		}
		for dy, row := range reactorRing {
			for dx, ch := range []rune(row) {
				i := (y0+dy)*w.W + x0 + dx
				cl.use[i], cl.aux[i], cl.mark[i] = uRing, uint8(dy*9+dx), true
				if ch == '◉' {
					cl.use[i] = uReactor
				}
			}
		}
		cl.reactors = append(cl.reactors, pt(x0+4, y0+1))
		for dy := -2; dy <= 4; dy++ { // the reactor's glow on the blocks round it
			for dx := -3; dx <= 11; dx++ {
				if x, y := x0+dx, y0+dy; w.In(x, y) {
					d := max(max(-dx, dx-8)/2, max(-dy, dy-2))
					cl.glow[y*w.W+x] = max(cl.glow[y*w.W+x], uint8(clamp(3-d, 0, 3)))
				}
			}
		}
	}
	towers := 0
	for k := 0; k < 120 && towers < 2; k++ {
		a := mapmodel.HashF(w.Seed, 333, int64(k))
		r := 14 + float64(22*mapmodel.HashF(w.Seed, 334, int64(k)))
		x0 := w.CX + int(float64(float64(mapmodel.Cos(a)*r)*2)) - 1
		y0 := w.CY + int(float64(mapmodel.Sin(a)*r)) - 1
		if !s.placeable(cl, x0, y0, 3, 3) {
			continue
		}
		for dy, row := range launchPad {
			for dx := range []rune(row) {
				i := (y0+dy)*w.W + x0 + dx
				cl.use[i], cl.aux[i], cl.mark[i] = uLaunch, uint8(dy*3+dx), true
			}
		}
		towers++
	}
}

// ----------------------------------------------------------------- smog

// citySmog lays the age's smog: a haze gathering at the city's edge in the
// Information Age, over the Digital Age's outskirts, heavy over the
// megacity's far blocks, thin once fusion cleans the air.
func (s *scene) citySmog(cl *cityLand) {
	lvl := cl.look.Smog
	if lvl <= 0 {
		return
	}
	w := s.w
	from := s.cityRadius() - 8 // where it begins, from the square
	if cl.key == "cyberpunk_age" || cl.key == "fusion_age" {
		from = s.wallR + 24
	}
	for i := range cl.smog {
		x, y := i%w.W, i/w.W
		d := rdist(x, y, w.CX, w.CY) + float64(6*mapmodel.Noise(w.Seed+407, float64(x)/10, float64(y)/5))
		v := float64(float64(d-from)/12) * lvl
		if cl.key == "information_age" && d > from+20 {
			v = float64(float64(from+32-d)/12) * lvl // a band at the edge, thinning beyond
		}
		cl.smog[i] = uint8(clamp(int(math.Round(float64(v*3))), 0, 3))
	}
}
