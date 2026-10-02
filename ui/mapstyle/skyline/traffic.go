package skyline

import (
	"math"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
)

// traffic.go is the ambient life: what moves through the city, per age, and
// how much of it. Every stream comes from real state: a running trade route
// is a caravan, a ship, a train or a flying freighter; people at work are
// foot traffic (sub-linear); soldiers patrol and a war sends a warband;
// wealth buys private traffic. Everything has a depth: lanes run between
// the building rows (behind the front row, behind the middle row, behind
// every row), so traffic weaves between the towers instead of over them.

// vkind is what drives a vehicle.
type vkind uint8

const (
	vkRoute   vkind = iota // one stream per running trade route
	vkFoot                 // people at work
	vkArmy                 // patrols and warbands
	vkPrivate              // wealth
	vkAmbient              // the age's own life: birds, balloons, aircraft
	vkCity                 // the Earth arc's city: freeway cars, sky trains, swarms (city.go)
)

// vtemplate is a vehicle drawing facing east. mask letters: b body,
// a accent, d ink, k wood, w canvas or sail, g trade gold, r war red,
// m metal, s steam, l warm light, n neon, c cyan light.
type vtemplate struct {
	rows, mask   []string
	body, accent theme.SkyHue
	sym          mapmodel.Sym
}

func vt(rows, mask []string, body, accent theme.SkyHue, sym mapmodel.Sym) vtemplate {
	return vtemplate{rows, mask, body, accent, sym}
}

func one(r, m string) ([]string, []string) { return []string{r}, []string{m} }

// the vehicle book, by band
var (
	tHunter     = vtpl1("☻╱", "dk", theme.SkyInk, theme.SkyTrunk, mapmodel.SymWorker)
	tPorters    = vtpl1("▪☻▪☻", "gdgd", theme.SkyInk, theme.SkyTradeGold, mapmodel.SymCaravan)
	tWarriors   = vtpl1("☻†☻", "dmd", theme.SkyInk, theme.SkyRock, mapmodel.SymMilitary)
	tWarband    = vtpl1("►☻☻☻", "rddd", theme.SkyInk, theme.SkyWarRed, mapmodel.SymRaider)
	tWalker     = vtpl1("☻", "d", theme.SkyInk, theme.SkyInk, mapmodel.SymWorker)
	tPair       = vtpl1("☻ ☻", "d d", theme.SkyInk, theme.SkyInk, mapmodel.SymWorker)
	tCanoe      = vt([]string{" ☻ ", "▀▀▀"}, []string{" d ", "kkk"}, theme.SkyTrunk, theme.SkyInk, mapmodel.SymShip)
	tOxCart     = vtpl1("▄█▄▐▀▌", "bbbggg", theme.SkyOx, theme.SkyTradeGold, mapmodel.SymCart)
	tDonkey     = vtpl1("▄█▄◘", "bbbk", theme.SkyOx, theme.SkyTrunk, mapmodel.SymCart)
	tChariot    = vtpl1("▄█▀◘", "bbrm", theme.SkyHorse, theme.SkyWarRed, mapmodel.SymMilitary)
	tTrireme    = vt([]string{"  █▌ ", "▀▀▀▀▀"}, []string{"  ww ", "bbbbb"}, theme.SkyHull, theme.SkySail, mapmodel.SymShip)
	tCart       = vtpl1("▄█▄▐▀▌", "bbbwww", theme.SkyHorse, theme.SkyCanvas, mapmodel.SymCart)
	tTradeCart  = vtpl1("▄█▄▐▀▌", "bbbggg", theme.SkyHorse, theme.SkyTradeGold, mapmodel.SymCart)
	tRider      = vt([]string{" ☻ ", "▄█▀"}, []string{" d ", "bbb"}, theme.SkyHorse, theme.SkyInk, mapmodel.SymMilitary)
	tSail       = vt([]string{" ▲▌ ", "▀██▀"}, []string{" ww ", "bbbb"}, theme.SkyHull, theme.SkySail, mapmodel.SymShip)
	tGalleon    = vt([]string{" ▲▲▲ ", "▀████▀"}, []string{" www ", "bbbbbb"}, theme.SkyHull, theme.SkySail, mapmodel.SymShip)
	tCoach      = vtpl1("▄█▐██▌", "bbaaaa", theme.SkyHorse, theme.SkyFunnel, mapmodel.SymCart)
	tTradeCoach = vtpl1("▄█▐██▌", "bbgggg", theme.SkyHorse, theme.SkyTradeGold, mapmodel.SymCart)
	tSteamTrain = vt([]string{"▄  ", "█▄▄▐▀▀▌"}, []string{"s  ", "mmmgggg"}, theme.SkyHullSteam, theme.SkyTradeGold, mapmodel.SymTrain)
	tBarge      = vtpl1("▀███▀", "bbbbb", theme.SkyHullSteam, theme.SkyHullSteam, mapmodel.SymShip)
	tSteamer    = vt([]string{"  █  ", "▀███▀"}, []string{"  a  ", "bbbbb"}, theme.SkyHullSteam, theme.SkyFunnel, mapmodel.SymShip)
	tBalloon    = vt([]string{"▄█▄", " ▀ ", " ▪ "}, []string{"bbb", " b ", " k "}, theme.SkyBalloon, theme.SkyTrunk, mapmodel.SymPlane)
	tZeppelin   = vt([]string{"▄████▄", "▀████▀", "  ▪   "}, []string{"bbbbbb", "bbbbbb", "  d   "}, theme.SkyZeppelin, theme.SkyInk, mapmodel.SymPlane)
	tTram       = vt([]string{" ╱  ", "▐██▌"}, []string{" m  ", "bbbb"}, theme.SkyTram, theme.SkyRail, mapmodel.SymTrain)
	tStreetcar  = vtpl1("▐█▌", "bbb", theme.SkyTram, theme.SkyTram, mapmodel.SymCar)
	tBiplane    = vtpl1("═╬═►", "bbbb", theme.SkyCanvas, theme.SkyCanvas, mapmodel.SymPlane)
	tSoldiers   = vtpl1("☻†☻†", "dmdm", theme.SkyInk, theme.SkyRock, mapmodel.SymMilitary)
	tCar        = vtpl1("▄█▄", "bab", theme.SkyCarRed, theme.SkyAircraft, mapmodel.SymCar)
	tTruck      = vtpl1("▄▄▄█▄", "gggba", theme.SkyCarWhite, theme.SkyAircraft, mapmodel.SymCar)
	tJeep       = vtpl1("▄█▄", "bdb", theme.SkyLeafDark, theme.SkyInk, mapmodel.SymMilitary)
	tJet        = vtpl1("───═►", "sssbb", theme.SkyAircraft, theme.SkyWhite, mapmodel.SymPlane)
	tAirliner   = vt([]string{"  ▄  ", "▀▀█▀►"}, []string{"  b  ", "bbbbb"}, theme.SkyAircraft, theme.SkyAircraft, mapmodel.SymPlane)
	tCargoJet   = vt([]string{"  ▄  ", "▀▀█▀►"}, []string{"  g  ", "bbbbb"}, theme.SkyAircraft, theme.SkyTradeGold, mapmodel.SymPlane)
	tHeli       = vt([]string{"─┼─", "▀█▄"}, []string{"mmm", "bbb"}, theme.SkyHullModern, theme.SkyRail, mapmodel.SymHeli)
	tFreighter  = vt([]string{" ▄▄▄▄ ", "▀████▀"}, []string{" aaaa ", "bbbbbb"}, theme.SkyHullModern, theme.SkyCargoRed, mapmodel.SymShip)
	tDrone      = vtpl1("✕", "c", theme.SkyNeonCyan, theme.SkyNeonCyan, mapmodel.SymDrone)
	tFlyCar     = vtpl1("▄█▄", "bnb", theme.SkyCarBlue, theme.SkyNeonPink, mapmodel.SymCar)
	tFlyCargo   = vtpl1("▄██▄", "gnng", theme.SkyHullModern, theme.SkyNeonCyan, mapmodel.SymCar)
	tMaglev     = vtpl1("▐████►", "mbbbbn", theme.SkyCarWhite, theme.SkyNeonCyan, mapmodel.SymTrain)
	tSkyTrain   = vtpl1("▐▀▀▀▌▐▀▀▀▌▐▀▀▀►", "mnnnmmnnnmmnnnb", theme.SkyHullModern, theme.SkyNeonPink, mapmodel.SymSkyTrain)
	tClimber    = vtpl1("◘", "c", theme.SkyWhite, theme.SkyWhite, mapmodel.SymClimber) // drawn by traffic(), on the tether
	tHoverTank  = vtpl1("▄▀█▄", "rbbb", theme.SkyHullModern, theme.SkyWarRed, mapmodel.SymMilitary)
	tHoverTram  = vtpl1("▐▀▀▀▀▌", "bggggb", theme.SkyCarWhite, theme.SkyNeonCyan, mapmodel.SymTrain)
	tHoverCar   = vtpl1("◄▄►", "cbc", theme.SkyCarWhite, theme.SkyNeonCyan, mapmodel.SymCar)
	tShuttle    = vt([]string{"▲", "█", "▓", "░"}, []string{"b", "b", "l", "s"}, theme.SkyCarWhite, theme.SkyExhaust, mapmodel.SymRocket)
	tStarship   = vt([]string{"  ▄██▄ ", "◄█████►"}, []string{"  bbbb ", "cbbbbbc"}, theme.SkyAircraftLate, theme.SkyNeonCyan, mapmodel.SymRocket)
	tWarpShip   = vtpl1("◄██►", "cbbc", theme.SkyAircraftLate, theme.SkyTradeGold, mapmodel.SymRocket)
	tMote       = vtpl1("·", "c", theme.SkyWarp, theme.SkyWarp, mapmodel.SymWorker)
	tBird       = vtpl1("v v", "d d", theme.SkyBird, theme.SkyBird, mapmodel.SymNone)
	tSatellite  = vtpl1("•", "c", theme.SkyWhite, theme.SkyWhite, mapmodel.SymSatellite)
)

func vtpl1(r, m string, body, accent theme.SkyHue, sym mapmodel.Sym) vtemplate {
	rows, mask := one(r, m)
	return vt(rows, mask, body, accent, sym)
}

// moverTags ties the vehicle book to the shared traffic roster
// (mapmodel/movers.go). A tagged vehicle never shows before the age that
// introduces its mover (the no-spoilers rule): a band that straddles that
// age falls back to the band below's vehicle. Untagged drawings (porters,
// warbands, canoes, barges, balloons, zeppelins, biplanes, jeeps,
// helicopters, starships and the like) are the skyline's own.
var moverTags = map[*vtemplate]mapmodel.Mover{
	&tHunter: mapmodel.MoverHunter, &tWalker: mapmodel.MoverWalker, &tPair: mapmodel.MoverWalker,
	&tOxCart: mapmodel.MoverOxCart, &tCart: mapmodel.MoverOxCart, &tTradeCart: mapmodel.MoverOxCart,
	&tRider: mapmodel.MoverRider, &tTrireme: mapmodel.MoverRowboat,
	&tSail: mapmodel.MoverSailShip, &tGalleon: mapmodel.MoverSailShip,
	&tCoach: mapmodel.MoverWagon, &tTradeCoach: mapmodel.MoverWagon,
	&tSteamTrain: mapmodel.MoverSteamTrain, &tTram: mapmodel.MoverTram, &tStreetcar: mapmodel.MoverTram,
	&tSteamer: mapmodel.MoverSteamship, &tCar: mapmodel.MoverEarlyCar, &tTruck: mapmodel.MoverTruck,
	&tJet: mapmodel.MoverPlane, &tAirliner: mapmodel.MoverPlane, &tCargoJet: mapmodel.MoverPlane,
	&tFreighter: mapmodel.MoverBoxShip, &tMaglev: mapmodel.MoverMaglev, &tHoverTram: mapmodel.MoverMaglev,
	&tDrone: mapmodel.MoverDrone, &tFlyCar: mapmodel.MoverHovercar, &tFlyCargo: mapmodel.MoverHovercar,
	&tHoverCar: mapmodel.MoverHovercar, &tHoverTank: mapmodel.MoverHovercar,
	&tShuttle: mapmodel.MoverShuttle, &tSatellite: mapmodel.MoverSatellite,
	&tSkyTrain: mapmodel.MoverSkyTrain, &tHeli: mapmodel.MoverNewsHeli, &tClimber: mapmodel.MoverClimber,
}

// introduced reports whether the age with index age may show vehicle t:
// its mover has been introduced by then, or it has none.
func introduced(t *vtemplate, age int) bool {
	k, ok := moverTags[t]
	return !ok || mapmodel.Introduced(k, age)
}

// fromBook is a band table's vehicle for band: its own or, while that is
// not introduced yet, the nearest band below's.
func fromBook(book *[9]*vtemplate, band, age int) *vtemplate {
	b := band
	for b > 0 && !introduced(book[b], age) {
		b--
	}
	return book[b]
}

// the band tables: soldiers, the ships of the trade routes, the bay's boats
var (
	armyBook = [9]*vtemplate{&tWarriors, &tChariot, &tRider, &tRider, &tSoldiers, &tJeep, &tHoverTank, &tHoverTank, &tHoverTank}
	shipBook = [9]*vtemplate{&tCanoe, &tTrireme, &tSail, &tGalleon, &tSteamer, &tFreighter, &tFreighter, &tFreighter, &tFreighter}
	boatBook = [9]*vtemplate{&tCanoe, &tTrireme, &tSail, &tBarge, &tBarge, &tBarge, &tHoverCar, &tHoverCar, &tHoverCar}
)

// hasViaduct reports whether a frame of the age draws the viaduct: the
// colonial to electric bands, from the railway's coming (the Industrial
// Age, when the roster brings the steam train).
func hasViaduct(age int) bool {
	b := bandOf(age)
	return (b == 3 || b == 4) && mapmodel.Introduced(mapmodel.MoverSteamTrain, age)
}

// hasRails reports whether a frame of the age draws the neon city's maglev
// rails: the digital band, once the maglev has come (the Cyberpunk Age).
func hasRails(age int) bool {
	return bandOf(age) == 6 && mapmodel.Introduced(mapmodel.MoverMaglev, age)
}

// vehicle is one placed vehicle in frame terms.
type vehicle struct {
	kind  vkind
	t     *vtemplate
	x, y  int // screen column of its left edge, scene row of its bottom row
	depth uint8
	west  bool // facing west (mirrored)
	seed  int
	body  theme.SkyHue // overrides the template body (car colours)
}

// cells calls fn for every drawn cell of the vehicle, in frame terms.
func (v *vehicle) cells(fn func(x, y int, ch rune, mk byte)) {
	n := len(v.t.rows)
	for r, row := range v.t.rows {
		rs, ms := []rune(row), []rune(v.t.mask[r])
		for i := range rs {
			j := i
			ch := rs[i]
			if v.west {
				j = len(rs) - 1 - i
				if m, ok := mirrorRunes[ch]; ok {
					ch = m
				}
			}
			if ch == ' ' {
				continue
			}
			mk := byte('b')
			if i < len(ms) {
				mk = byte(ms[i])
			}
			fn(v.x+j, v.y-(n-1-r), ch, mk)
		}
	}
}

// width is the vehicle's widest row.
func (v *vehicle) width() int {
	w := 0
	for _, r := range v.t.rows {
		w = max(w, len([]rune(r)))
	}
	return w
}

// bandOf is the traffic band of an age: 0 primitive/stone, 1 bronze to
// classical, 2 medieval/renaissance, 3 colonial/industrial,
// 4 victorian/electric, 5 atomic to information, 6 digital/cyberpunk,
// 7 fusion/space, 8 interstellar to transcendent.
func bandOf(age int) int {
	switch {
	case age <= 1:
		return 0
	case age <= 4:
		return 1
	case age <= 6:
		return 2
	case age <= 8:
		return 3
	case age <= 10:
		return 4
	case age <= 13:
		return 5
	case age <= 15:
		return 6
	case age <= 17:
		return 7
	}
	return 8
}

// runningModes lists the modes of the first n running (not disrupted)
// routes, in model order.
func runningModes(m *mapmodel.Model, n int) []mapmodel.RouteMode {
	var out []mapmodel.RouteMode
	for _, r := range m.Routes {
		if !r.Disrupted && len(out) < n {
			out = append(out, r.Mode)
		}
	}
	return out
}

// counts is how much of each kind moves, from real state.
type counts struct{ route, foot, army, war, private, ambient int }

func trafficCounts(m *mapmodel.Model) counts {
	a := m.Activity
	c := counts{route: min(a.Routes, 8), war: min(a.Wars, 4)}
	if a.Staffed > 0 {
		c.foot = min(12, int(mapmodel.Log2(1+float64(a.Staffed))*1.1))
	}
	if a.Soldiers > 0 {
		c.army = min(4, (a.Soldiers+39)/40)
	}
	c.private = int(a.Wealth*8 + 0.5)
	c.ambient = 2 + int(a.Traffic*4+0.5)
	return c
}

// lane is where a stream runs: a scene row and a depth.
type lane struct {
	y int
	d uint8
}

// lanes is one frame's picking context: the lanes its streams run on, the
// works it draws (viaduct, maglev rails) and the age whose roster it shows.
type lanes struct {
	age, groundY       int
	street, back, high lane
	sky                [3]lane
	day                bool // birds fly
	viaduct, rails     bool
}

// walk picks stream i's vehicle in band: pick gives a band's own choice
// (nil: that band has nothing for the stream, which is skipped), and
// while the choice is a mover the age has not introduced yet, the band
// below's is taken instead.
func (l *lanes) walk(band, i int, pick func(l *lanes, b, i int) (*vtemplate, lane)) (*vtemplate, lane) {
	for b := band; b >= 0; b-- {
		t, ln := pick(l, b, i)
		if t == nil || introduced(t, l.age) {
			return t, ln
		}
	}
	return nil, lane{}
}

// landRoute is band b's vehicle for land trade route i. The steam train
// takes the viaduct and the maglev its rails only once the frame draws
// them; before that the band's road traffic carries the goods.
func (l *lanes) landRoute(b, i int) (*vtemplate, lane) {
	switch b {
	case 0:
		return &tPorters, l.street
	case 1:
		return &tOxCart, l.street
	case 2:
		return &tTradeCart, l.street
	case 3:
		if i%3 == 2 && l.viaduct {
			return &tSteamTrain, lane{viaductY(l.groundY) - 1, dLane0}
		}
		return &tTradeCoach, [2]lane{l.street, l.back}[i%2]
	case 4:
		if i%2 == 0 && l.viaduct {
			return &tSteamTrain, lane{viaductY(l.groundY) - 1, dLane0}
		}
		return &tTram, l.street
	case 5:
		return &tTruck, l.street
	case 6:
		if !l.rails {
			return &tTruck, l.street
		}
		ry := railYs(l.groundY)
		r := ry[i%len(ry)]
		return &tMaglev, lane{r.y - 1, r.d - 1}
	}
	return &tHoverTram, lane{l.groundY - 4, dLane0}
}

// airRoute is band b's aircraft for air trade route i; nil before band 5,
// so an air route walked down that far goes overland.
func (l *lanes) airRoute(b, i int) (*vtemplate, lane) {
	switch {
	case b >= 8:
		return &tWarpShip, l.sky[i%3]
	case b >= 6:
		return &tFlyCargo, l.sky[i%3]
	case b == 5:
		return &tCargoJet, l.high
	}
	return nil, lane{}
}

// privateCar is band b's private vehicle number i.
func (l *lanes) privateCar(b, i int) (*vtemplate, lane) {
	ln := [2]lane{l.street, l.back}[i%2]
	switch b {
	case 0:
		if i > 1 {
			return nil, ln
		}
		return &tWalker, ln
	case 1:
		if i > 2 {
			return nil, ln
		}
		return &tDonkey, ln
	case 2:
		if i > 3 {
			return nil, ln
		}
		return &tCart, ln
	case 3:
		return &tCoach, ln
	case 4:
		return &tStreetcar, ln
	case 5:
		return &tCar, ln
	case 6:
		return &tFlyCar, l.sky[i%3]
	}
	return &tHoverCar, l.sky[i%3]
}

// ambient is band b's own life number i: birds, balloons, aircraft.
func (l *lanes) ambient(b, i int) (*vtemplate, lane) {
	seed := 900 + i
	ln := l.high
	ln.y = max(1, l.groundY/6+int(hash(seed, 37)%uint64(max(1, l.groundY/4))))
	switch b {
	case 0, 1, 2:
		if !l.day || i > 2 {
			return nil, ln
		}
		return &tBird, ln
	case 3:
		if i > 1 {
			return nil, ln
		}
		return &tBalloon, ln
	case 4:
		if i > 2 {
			return nil, ln
		}
		return [2]*vtemplate{&tZeppelin, &tBiplane}[i%2], ln
	case 5:
		// jets, airliners and helicopters: the whole mix waits for the
		// plane, and until it comes the sky keeps band 4's airships
		if !mapmodel.Introduced(mapmodel.MoverPlane, l.age) {
			return l.ambient(4, i)
		}
		t := [3]*vtemplate{&tJet, &tAirliner, &tHeli}[i%3]
		if t == &tHeli && !introduced(&tHeli, l.age) {
			t = &tJet // the news helicopter waits for the Information Age
		}
		if t == &tHeli {
			ln = l.sky[1]
		}
		return t, ln
	case 6:
		ln = l.sky[i%3]
		ln.y -= int(hash(seed, 41) % 3)
		if i%3 == 2 && mapmodel.MoverNewsHeli.Info().In(l.age) {
			return &tHeli, l.sky[1] // the Digital Age's sky still has its news helicopters
		}
		return &tDrone, ln
	case 7:
		if i > 2 {
			return nil, ln
		}
		return &tHoverCar, l.sky[i%3]
	}
	if i > 1 {
		return nil, ln
	}
	return &tStarship, lane{max(1, l.groundY/8+i*3), dAir}
}

// trafficFor returns the vehicles for a model at an animation frame, for a
// view of width w scrolled to cam with the ground at scene row groundY.
// Streams wrap per screen but are anchored to the world, so scrolling moves
// them with the buildings. It is a pure function of its arguments.
func trafficFor(m *mapmodel.Model, anim, cam, w, groundY int) []vehicle {
	if m == nil || w <= 0 || groundY < 6 {
		return nil
	}
	band := bandOf(m.AgeIdx)
	c := trafficCounts(m)
	gy := float64(groundY)
	ls := lanes{
		age:     m.AgeIdx,
		groundY: groundY,
		street:  lane{groundY + 1, dStreet},
		back:    lane{groundY - 1, dLane0},
		high:    lane{max(1, groundY/6), dAirLow},
		sky:     [3]lane{{int(gy * 0.62), dLane0}, {int(gy * 0.48), dLane1}, {int(gy * 0.34), dLane2}},
		day:     m.Clock.Daylight > 0.5 && m.Weather.Kind != mapmodel.Rain && m.Weather.Kind != mapmodel.Storm,
		viaduct: hasViaduct(m.AgeIdx),
		rails:   hasRails(m.AgeIdx),
	}
	var out []vehicle
	add := func(kind vkind, t *vtemplate, seed int, ln lane, speed float64) {
		v := vehicle{kind: kind, t: t, seed: seed, depth: ln.d, y: ln.y}
		v.west = hash(seed, 17)%2 == 0 && kind != vkArmy
		if kind == vkArmy && t == &tWarband {
			v.west = false // warbands march in from the west
		}
		if v.west {
			speed = -speed
		}
		pad := 12
		span := w + 2*pad
		pos := float64(hash(seed, 23)%uint64(span)) + float64(anim)*speed - float64(cam)
		x := int(math.Floor(pos)) % span
		if x < 0 {
			x += span
		}
		v.x = x - pad
		if t == &tCar || t == &tFlyCar {
			v.body = [4]theme.SkyHue{theme.SkyCarRed, theme.SkyCarBlue, theme.SkyCarWhite, theme.SkyCarYellow}[hash(seed, 3)%4]
		}
		out = append(out, v)
	}
	sp := func(seed int, lo, hi float64) float64 { return lo + (hi-lo)*hashf(seed, 13) }

	// trade routes: one stream per running route, drawn by how its goods
	// travel (mapmodel.Route.Mode). Sea routes sail the bays (up to three
	// ships a bay); with no bay they come overland. Air routes fly once the
	// age has aircraft (the plane, from the Modern Age), and walk or roll
	// before that.
	modes := runningModes(m, c.route)
	bays := false
	for _, d := range m.Skyline.Districts {
		bays = bays || d.Bay
	}
	sea := 0
	for i, mode := range modes {
		if mode == mapmodel.ModeSea && bays {
			sea++
			continue
		}
		seed := 100 + i
		var t *vtemplate
		var ln lane
		if mode == mapmodel.ModeAir {
			t, ln = ls.walk(band, i, (*lanes).airRoute)
		}
		if t == nil {
			t, ln = ls.walk(band, i, (*lanes).landRoute)
		}
		if t != nil {
			add(vkRoute, t, seed, ln, sp(seed, 0.25, 0.45))
		}
	}
	// people at work
	for i := 0; i < c.foot; i++ {
		seed := 300 + i
		t := &tWalker
		switch {
		case band == 0:
			t = &tHunter
		case i%3 == 0:
			t = &tPair
		case band == 8 && i%2 == 0:
			t = &tMote
		}
		ln := [2]lane{ls.street, ls.back}[i%2]
		add(vkFoot, t, seed, ln, sp(seed, 0.08, 0.18))
	}
	// soldiers and wars
	for i := 0; i < c.army; i++ {
		seed := 500 + i
		add(vkArmy, fromBook(&armyBook, band, m.AgeIdx), seed, ls.street, sp(seed, 0.12, 0.25))
	}
	for i := 0; i < c.war; i++ {
		add(vkArmy, &tWarband, 600+i, ls.street, 0.22)
	}
	// private traffic
	for i := 0; i < c.private; i++ {
		seed := 700 + i
		if t, ln := ls.walk(band, i, (*lanes).privateCar); t != nil {
			add(vkPrivate, t, seed, ln, sp(seed, 0.3, 0.6))
		}
	}
	// the age's own life
	for i := 0; i < c.ambient; i++ {
		seed := 900 + i
		if t, ln := ls.walk(band, i, (*lanes).ambient); t != nil {
			add(vkAmbient, t, seed, ln, sp(seed, 0.2, 0.7))
		}
	}
	// the Earth arc's city: freeway cars, sky trains, swarms and hovercars
	cityTraffic(m, groundY, func(t *vtemplate, seed int, ln lane, speed float64, west bool, dx int) {
		v := vehicle{kind: vkCity, t: t, seed: seed, depth: ln.d, y: ln.y, west: west}
		if west {
			speed = -speed
		}
		pad := 12
		span := w + 2*pad
		pos := float64(hash(seed, 23)%uint64(span)) + float64(float64(anim)*speed) - float64(cam)
		x := int(math.Floor(pos)) % span
		if x < 0 {
			x += span
		}
		v.x = x - pad + dx
		if t == &tCar || t == &tFlyCar {
			v.body = [4]theme.SkyHue{theme.SkyCarRed, theme.SkyCarBlue, theme.SkyCarWhite, theme.SkyCarYellow}[hash(seed, 3)%4]
		}
		out = append(out, v)
	})
	out = append(out, bayTraffic(m, anim, cam, w, groundY, band, sea)...)
	out = append(out, launches(m, anim, cam, w, groundY, band)...)
	return out
}

// bayTraffic: route ships (one per running sea route, up to three per
// bay) and one boat of the age on every bay in view.
func bayTraffic(m *mapmodel.Model, anim, cam, w, groundY, band, routes int) []vehicle {
	var out []vehicle
	ship, boat := fromBook(&shipBook, band, m.AgeIdx), fromBook(&boatBook, band, m.AgeIdx)
	for di, d := range m.Skyline.Districts {
		if !d.Bay || d.BayX+d.BayW < cam || d.BayX > cam+w {
			continue
		}
		for i := 0; i <= min(routes, 3); i++ {
			kind, t := vkRoute, ship
			if i == routes || i == 3 {
				kind, t = vkAmbient, boat
			}
			seed := di*10 + i
			v := vehicle{kind: kind, t: t, seed: seed, depth: dShip, y: groundY}
			spn := float64(max(1, d.BayW-8))
			ph := math.Mod(float64(hash(seed, 5)%100)+float64(anim)*(0.08+hashf(seed, 3)*0.1), spn*2)
			off := ph
			if ph > spn {
				off, v.west = spn*2-ph, true
			}
			v.x = d.BayX + 2 + int(off) - cam
			out = append(out, v)
			if kind == vkAmbient {
				break
			}
		}
	}
	return out
}

// launches: once the roster brings the shuttle (the Space Age), rockets
// climb from the launch complexes on screen, one at a time.
func launches(m *mapmodel.Model, anim, cam, w, groundY, band int) []vehicle {
	if band != 7 || !introduced(&tShuttle, m.AgeIdx) ||
		(m.Building("launch_complex") == nil && m.Building("space_program") == nil) {
		return nil
	}
	const period = 160
	ph := anim % period
	if ph > groundY*2 {
		return nil
	}
	x := -1
	if len(m.Skyline.Districts) > 17 {
		d := m.Skyline.Districts[17]
		x = d.X0 + d.LandW/3 - cam
	} else if n := len(m.Skyline.Districts); n > 0 {
		d := m.Skyline.Districts[n-1]
		x = d.X0 + d.LandW - 6 - cam
	}
	if x < 0 || x >= w {
		return nil
	}
	return []vehicle{{kind: vkAmbient, t: &tShuttle, x: x, y: groundY - 2 - ph/2, depth: dLane2, seed: anim / period}}
}

// viaductY is the row of the viaduct deck (colonial to electric ages).
func viaductY(groundY int) int { return groundY - max(4, groundY/6) }

// railYs are the maglev rails of the neon cities: rows and depths, the
// megacity's three to five (city.go).
func railYs(groundY int) []lane { return cityRailYs(groundY) }

// structures draws the static traffic works: viaducts, maglev rails and
// the hover-tram guideway, at their depths (so nearer towers hide them).
// The viaduct waits for the railway and the rails for the maglev.
func (s *scene) structures() {
	gy := s.groundY
	metal := s.p.col(5, 0, sMetal, 1)
	if hasFreeway(s.m.AgeIdx) {
		s.freeway() // the Modern to the Digital Age's elevated freeway (city.go)
	}
	s.cables()
	switch s.band {
	case 3, 4:
		if !hasViaduct(s.m.AgeIdx) {
			return
		}
		y := viaductY(gy)
		stone := s.p.col(4, 2, sTrim, 1)
		for x := 0; x < s.W; x++ {
			wx := s.wx(x)
			s.fb.fg(x, s.Y(y), '▀', stone, dLane0+1)
			if wx%9 == 0 {
				for yy := y + 1; yy < gy; yy++ {
					s.fb.fill(x, s.Y(yy), stone, dLane0+1)
				}
			} else if wx%9 == 1 || wx%9 == 8 {
				s.fb.fg(x, s.Y(y+1), '▀', stone, dLane0+1)
			}
		}
	case 6:
		if !hasRails(s.m.AgeIdx) {
			return
		}
		neon := s.p.hue(theme.SkyNeonCyan, mEmit, 1)
		for i, r := range railYs(gy) {
			hz := i + 1
			rail := s.p.col(8, i%3, sMetal, hz)
			for x := 0; x < s.W; x++ {
				wx := s.wx(x)
				s.fb.set(x, s.Y(r.y), '▀', rail, s.fb.showAt(x, s.Y(r.y)), r.d)
				if (wx+i*7)%5 == 0 {
					s.fb.fg(x, s.Y(r.y), '▀', neon, r.d)
				}
				if (wx+i*11)%29 == 0 {
					for yy := r.y + 1; yy < gy; yy++ {
						s.fb.fg(x, s.Y(yy), '│', rail, r.d)
					}
				}
			}
		}
	case 7:
		y := gy - 3
		for x := 0; x < s.W; x++ {
			s.fb.fg(x, s.Y(y), '▀', metal, dLane0+1)
			if s.wx(x)%13 == 0 {
				s.fb.fg(x, s.Y(y+1), '│', metal, dLane0+1)
				s.fb.fg(x, s.Y(y+2), '│', metal, dLane0+1)
			}
		}
	}
}

// traffic draws the vehicles, depth tested cell by cell, plus the space
// elevator's climber, orbital traffic and warp flashes.
func (s *scene) traffic() {
	if s.v.noTraffic {
		return
	}
	s.structures()
	vs := trafficFor(s.m, s.anim, s.cam, s.W, s.groundY)
	night := s.p.night > 0.4
	for i := range vs {
		v := &vs[i]
		mode := mLit
		body := v.t.body
		if v.body != 0 {
			body = v.body
		}
		markX, markY := v.x+v.width()/2, v.y
		v.cells(func(x, y int, ch rune, mk byte) {
			h, md := body, mode
			switch mk {
			case 'a':
				h = v.t.accent
			case 'd':
				h = theme.SkyInk
			case 'k':
				h = theme.SkyTrunk
			case 'w':
				h = theme.SkySail
			case 'g':
				h = theme.SkyTradeGold
			case 'r':
				h, md = theme.SkyWarRed, mEmit
			case 'm':
				h = theme.SkyRail
			case 's':
				h = theme.SkySteam
			case 'l':
				h, md = theme.SkyExhaust, mEmit
			case 'n':
				h, md = v.t.accent, mEmit
			case 'c':
				h, md = theme.SkyNeonCyan, mEmit
			}
			if v.t == &tStarship || v.t == &tSatellite || v.t == &tMote {
				md = mEmit
			}
			if s.tier == mapmodel.TierNerd && x == markX && y == markY && v.t.sym != mapmodel.SymNone {
				ch = mapmodel.R(v.t.sym, s.tier)
			}
			s.fb.fg(x, s.Y(y), ch, s.p.hue(h, md, 0), v.depth)
		})
		if night && v.kind == vkCity && (v.t == &tCar || v.t == &tTruck) { // headlights down the freeway
			hx, tx := v.x+v.width(), v.x-1
			if v.west {
				hx, tx = v.x-1, v.x+v.width()
			}
			s.fb.fg(hx, s.Y(v.y), '▪', s.p.cityHue(theme.CityHeadlight, mEmit, 0), v.depth)
			s.fb.fg(tx, s.Y(v.y), '·', s.p.cityHue(theme.CityTaillight, mEmit, 0), v.depth)
		}
		if night && s.band >= 5 && v.depth == dStreet && v.kind != vkFoot {
			hx := v.x + v.width()
			if v.west {
				hx = v.x - 1
			}
			s.fb.tint(hx, s.Y(v.y), s.p.hue(theme.SkyHeadlight, mEmit, 0), 0.3, v.depth)
		}
		if v.t == &tSteamTrain || v.t == &tSteamer {
			sx, dir := v.x, -1
			if v.west {
				sx, dir = v.x+v.width()-1, 1
			}
			steam := s.p.hue(theme.SkySteam, mLit, 0)
			for k := 1; k <= 3; k++ {
				s.fb.fg(sx+dir*k, s.Y(v.y-len(v.t.rows)-k/2), [3]rune{'▓', '▒', '░'}[k-1], steam, v.depth)
			}
		}
	}
	if x := s.elevatorX(); x >= 0 && x < s.W {
		if s.fusion() && introduced(&tClimber, s.m.AgeIdx) { // climbers riding the tether up and away
			span := max(1, s.groundY-1)
			for k := 0; k < 3; k++ {
				y := s.groundY - 2 - (s.anim/3+k*span/3)%span
				s.fb.fg(x, s.Y(y), mapmodel.R(mapmodel.SymClimber, s.tier), s.p.cityHue(theme.CityPlasma, mEmit, 0), dRow2-1)
			}
		} else {
			cy := (s.anim / 2) % (s.groundY * 2)
			if cy >= s.groundY {
				cy = s.groundY*2 - cy
			}
			s.fb.fg(x, s.Y(min(cy, s.groundY-2)), '◘', s.p.hue(theme.SkyHarbingerHolo, mEmit, 0), dRow2-1)
		}
	}
	sats := introduced(&tSatellite, s.m.AgeIdx)
	if sats && s.m.AgeIdx >= 17 {
		for i := 0; i < 3; i++ {
			x := (int(hash(i, 91)%uint64(s.W)) + s.anim/(3+i)) % s.W
			_, y := s.orbitY(x)
			s.fb.fg(x, s.Y(y+1), '·', s.p.hue(theme.SkyWhite, mEmit, 2), dAir)
		}
	} else if sats && s.p.night > 0.5 {
		for i := 0; i < 2; i++ {
			x := (int(hash(i, 91)%uint64(s.W)) + s.anim/5) % s.W
			s.fb.fg(x, s.Y(1+i*2), mapmodel.R(mapmodel.SymSatellite, s.tier), s.p.hue(theme.SkyWhite, mEmit, 2), dAir)
		}
	}
	if s.band == 8 && (s.anim/3)%17 < 2 {
		x := int(hash(s.anim/51, 3) % uint64(s.W))
		y := 2 + int(hash(s.anim/51, 4)%uint64(max(1, s.groundY/3)))
		c := s.p.hue(theme.SkyWarp, mEmit, 1)
		s.fb.fg(x, s.Y(y), '✦', c, dAir)
		for k := 1; k < 6; k++ {
			s.fb.fg(x-k, s.Y(y), '─', theme.Mix(c, s.fb.showAt(clampInt(x-k, 0, s.W-1), s.Y(y)), float64(k)/6), dAir)
		}
	}
}
