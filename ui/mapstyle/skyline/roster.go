package skyline

import (
	"math"
	"strings"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
)

// roster.go is the foreground's own life, from the full view's 40 rows up:
// people as two-cell figures (a head over a body) walking in groups of two
// to five, and the traffic of every age on three lanes.
//
//	road  the street: people, animals, carts, carriages, cars (rows ground
//	      and ground+1, the vehicle's feet on the road)
//	rail  the railway on the verge, from the Industrial Age (rows ground+1
//	      and ground+2); trains are an engine with up to four carriages
//	sky   above the buildings, from the age of flight (the Electric Age);
//	      it runs behind every building row, so a tall tower hides it
//
// One table (rosterEntries) says which age each vehicle belongs to: its
// first and last common age, and the last age it still turns up rarely (a
// horse in the Modern Age). Every drawing is 1 to 4 cells wide and 1 or 2
// tall, in the Unicode tier, with its own rows for the ASCII tier when the
// folded ones would not do, and a second frame for the step or the wheels.
// A train is built from an engine and a carriage drawing, each of that size.
// Below 40 rows the full view keeps its one-cell markers (traffic.go).
//
// The Space Age on, the sky scenes (orbit.go) draw their own traffic and have
// no street; their rows here are the roster the foreground would show.

// the ages, by index (checked against config.AgeOrder by the tests)
const (
	agePrimitive = iota
	ageStone
	ageBronze
	ageIron
	ageClassical
	ageMedieval
	ageRenaissance
	ageColonial
	ageIndustrial
	ageVictorian
	ageElectric
	ageAtomic
	ageModern
	ageInformation
	ageDigital
	ageCyberpunk
	ageFusion
	ageSpace
	ageInterstellar
	ageGalactic
	ageQuantum
	ageTranscendent
	numAges
)

// peopleMinH is the full view's height from which people are two cells
// tall; shorter, they stay the one-cell markers.
const peopleMinH = 40

// bigFigures reports whether a frame with the ground at scene row groundY
// (the full view puts it six rows above the bottom) draws two-cell people
// and the roster's lanes.
func bigFigures(groundY int) bool { return groundY+6 >= peopleMinH }

type laneKind uint8

const (
	laneRoad laneKind = iota
	laneRail
	laneSky
	numLanes
)

// peopleHues are the clothes of the people in the street: the skyline's own
// reds, blues, greens and golds, so a theme re-colours them with the rest.
var peopleHues = [...]theme.SkyHue{theme.SkyCarRed, theme.SkyCarBlue, theme.SkyCargoGreen, theme.SkyTradeGold,
	theme.SkyCargoRed, theme.SkyTram, theme.SkyBalloon, theme.SkyCarWhite}

// rosterMirror is what the roster's drawings flip besides mirrorRunes: the
// legs of a walker and the half discs of a wheel.
var rosterMirror = map[rune]rune{'▙': '▟', '▟': '▙', '▛': '▜', '▜': '▛', '◖': '◗', '◗': '◖'}

// ------------------------------------------------------------- drawings

// rt builds a roster drawing: Unicode rows and their mask, then optional
// second frame, ASCII rows and ASCII second frame through the chained setters.
func rt(rows, mask []string, body, accent theme.SkyHue, sym mapmodel.Sym) *vtemplate {
	t := vt(rows, mask, body, accent, sym)
	return &t
}

func (t *vtemplate) step(alt []string) *vtemplate  { t.alt = alt; return t }
func (t *vtemplate) asc(a, aa []string) *vtemplate { t.ascii, t.asciiAlt = a, aa; return t }
func (t *vtemplate) painted() *vtemplate           { t.paint = true; return t }
func (t *vtemplate) lit() *vtemplate               { t.emit = true; return t }
func (t *vtemplate) smoking() *vtemplate           { t.steam = true; return t }

func rs(r ...string) []string { return r }

// people: a head (h) over a body in clothes (1)
var (
	pMan    = rt(rs("●", "▙"), rs("h", "1"), theme.SkyInk, theme.SkyInk, mapmodel.SymNone).step(rs("●", "▟")).asc(rs("o", "/"), rs("o", "\\"))
	pWoman  = rt(rs("●", "▲"), rs("h", "1"), theme.SkyInk, theme.SkyInk, mapmodel.SymNone).asc(rs("o", "A"), nil)
	pChild  = rt(rs("☻"), rs("1"), theme.SkyInk, theme.SkyInk, mapmodel.SymNone).asc(rs("o"), nil)
	pBundle = rt(rs("●▄", "▙ "), rs("hk", "1 "), theme.SkyInk, theme.SkyTrunk, mapmodel.SymNone).step(rs("●▄", "▟ ")).asc(rs("o_", "/ "), rs("o_", "\\ "))
	pSpear  = rt(rs("●╱", "▙ "), rs("hk", "1 "), theme.SkyInk, theme.SkyTrunk, mapmodel.SymNone).step(rs("●╱", "▟ ")).asc(rs("o/", "/ "), rs("o/", "\\ "))
)

// the road
var (
	tpDog  = rt(rs("▄▟"), rs("bb"), theme.SkyOx, theme.SkyOx, mapmodel.SymNone).step(rs("▄▙")).asc(rs("_>"), rs("_)"))
	tpGoat = rt(rs(" ▖", "▟▙"), rs(" b", "bb"), theme.SkyCanvas, theme.SkyCanvas, mapmodel.SymNone).step(rs(" ▖", "▙▟")).
		asc(rs(" ,", "/\\"), rs(" ,", "\\/"))
	tpOxCart = rt(rs("▄▄ ▖", "◘◘▟█"), rs("gg b", "kkbb"), theme.SkyOx, theme.SkyTradeGold, mapmodel.SymOxCart).step(rs("▄▄ ▖", "◘◘▙█")).
			asc(rs("__ ,", "OO/#"), rs("__ ,", "OO\\#"))
	tpCamel = rt(rs(" ▄ ▗", "▟██▌"), rs(" b b", "bbbb"), theme.SkyField2, theme.SkyField2, mapmodel.SymNone).step(rs(" ▄ ▗", "▙██▐")).
		asc(rs(" n ,", "/##)"), rs(" n ,", "\\##("))
	tpChariot = rt(rs("●▄ ▗", "◘█▟▙"), rs("hr b", "mrbb"), theme.SkyHorse, theme.SkyWarRed, mapmodel.SymMilitary).step(rs("●▄ ▗", "◘█▙▟")).
			asc(rs("o_ ,", "O#/\\"), rs("o_ ,", "O#\\/"))
	tpRider = rt(rs(" ●▗", "▟█▌"), rs(" hb", "bbb"), theme.SkyHorse, theme.SkyInk, mapmodel.SymRider).step(rs(" ●▗", "▙█▐")).
		asc(rs(" o,", "/#)"), rs(" o,", "\\#("))
	tpWagon = rt(rs("▄█▄▗", "◘◘▟▙"), rs("wwwb", "kkbb"), theme.SkyHorse, theme.SkyCanvas, mapmodel.SymWagon).step(rs("▄█▄▗", "◘◘▙▟")).
		asc(rs("_#_,", "OO/\\"), rs("_#_,", "OO\\/"))
	tpCarriage = rt(rs("▟▫▙▗", "◘◘▟▙"), rs("alab", "kkbb"), theme.SkyHorse, theme.SkyFunnel, mapmodel.SymCart).step(rs("▟▫▙▗", "◘◘▙▟")).
			asc(rs("/.\\,", "OO/\\"), rs("/.\\,", "OO\\/"))
	tpTram = rt(rs(" ╱  ", "▐▫▫▌"), rs(" m  ", "baab"), theme.SkyTram, theme.SkyWinWarm, mapmodel.SymTram).
		asc(rs(" /  ", "[..]"), nil)
	tpEarlyCar = rt(rs("▗▄▖", "◘█◘"), rs("bbb", "dbd"), theme.SkyCarRed, theme.SkyAircraft, mapmodel.SymAuto).painted().
			asc(rs(" _ ", "O#O"), nil)
	tpCar = rt(rs(" ▄▄ ", "◘██◘"), rs(" aa ", "dbbd"), theme.SkyCarRed, theme.SkyAircraft, mapmodel.SymAuto).painted().
		asc(rs(" __ ", "O##O"), nil)
	tpBus = rt(rs("█▫▫█", "◘██◘"), rs("baab", "dbbd"), theme.SkyCarYellow, theme.SkyAircraft, mapmodel.SymAuto).
		asc(rs("#..#", "O##O"), nil)
	tpHoverBike = rt(rs(" ●▗", "═▀▀"), rs(" hb", "cbb"), theme.SkyNeonPink, theme.SkyNeonCyan, mapmodel.SymHovercar).lit().
			asc(rs(" o,", "=\"\""), nil)
	tpHoverCar = rt(rs(" ▄▄ ", "◄██►"), rs(" aa ", "cbbc"), theme.SkyCarBlue, theme.SkyNeonCyan, mapmodel.SymHovercar).painted().lit().
			asc(rs(" __ ", "<##>"), nil)
)

// the rails: an engine and a carriage each
var (
	trSteamEngine = rt(rs("▄▄ ▗", "█◘◘█"), rs("bb d", "bmmb"), theme.SkyHullSteam, theme.SkyFunnel, mapmodel.SymTrain).smoking().step(rs("▄▄ ▗", "█◘█◘")).
			asc(rs("__ ,", "#OO#"), rs("__ ,", "#O#O"))
	trSteamCar = rt(rs("▐▫▫▌", "◘▀▀◘"), rs("baab", "dbbd"), theme.SkyHull, theme.SkyWinWarm, mapmodel.SymTrain).
			asc(rs("[..]", "O\"\"O"), nil)
	trDieselEngine = rt(rs("▐██◣", "◘██◘"), rs("bbbb", "dbbd"), theme.SkyCargoRed, theme.SkyAircraft, mapmodel.SymTrain).
			asc(rs("[##/", "O##O"), nil)
	trDieselCar = rt(rs("▐▫▫▌", "◘██◘"), rs("baab", "dbbd"), theme.SkyCarWhite, theme.SkyWinWarm, mapmodel.SymTrain).
			asc(rs("[..]", "O##O"), nil)
	trBulletHead = rt(rs("▄▄▄◣", "████"), rs("bbbb", "bbbb"), theme.SkyCarWhite, theme.SkyNeonCyan, mapmodel.SymTrain).
			asc(rs("___/", "####"), nil)
	trBulletCar = rt(rs("▄▄▄▄", "████"), rs("aaaa", "bbbb"), theme.SkyCarWhite, theme.SkyNeonCyan, mapmodel.SymTrain).
			asc(rs("____", "####"), nil)
	trMaglevHead = rt(rs("▄▄▄◣", "▀███"), rs("bbbb", "cbbb"), theme.SkyCarWhite, theme.SkyNeonCyan, mapmodel.SymMaglev).lit().
			asc(rs("___/", "\"###"), nil)
	trMaglevCar = rt(rs("▄▄▄▄", "▀▀▀▀"), rs("bbbb", "cccc"), theme.SkyCarWhite, theme.SkyNeonCyan, mapmodel.SymMaglev).lit().
			asc(rs("____", "\"\"\"\""), nil)
	trPortalHead = rt(rs("▐◈◈▌", "◄▀▀►"), rs("bnnb", "cbbc"), theme.SkyAircraftLate, theme.SkyNeonCyan, mapmodel.SymMaglev).lit().
			asc(rs("[**]", "<\"\">"), nil)
	trPortalCar = rt(rs("▐◈◈▌", "▀▀▀▀"), rs("bnnb", "cccc"), theme.SkyAircraftLate, theme.SkyNeonCyan, mapmodel.SymMaglev).lit().
			asc(rs("[**]", "\"\"\"\""), nil)
)

// the sky
var (
	skBiplane = rt(rs(" ═╬═", "─▀█►"), rs(" bmb", "bbba"), theme.SkyCanvas, theme.SkyCargoRed, mapmodel.SymPlane).
			asc(rs(" =+=", "-\"#>"), nil)
	skProp = rt(rs("▄ ┼ ", "▀██►"), rs("b m ", "bbba"), theme.SkyAircraft, theme.SkyAircraft, mapmodel.SymPlane).
		asc(rs("_ + ", "\"##>"), nil)
	skJet = rt(rs("◢██ ", "▀██►"), rs("bbb ", "bbba"), theme.SkyAircraft, theme.SkyWhite, mapmodel.SymPlane).
		asc(rs("/## ", "\"##>"), nil)
	skAirliner = rt(rs("▗▟█▖", "▀██►"), rs("bbbb", "bbba"), theme.SkyAircraft, theme.SkyAircraft, mapmodel.SymPlane).
			asc(rs("_/#_", "\"##>"), nil)
	skHeli = rt(rs(" ─┼─", "─▄█▌"), rs(" mmm", "bbbb"), theme.SkyHullModern, theme.SkyRail, mapmodel.SymHeli).step(rs(" ═┼═", "─▄█▌")).
		asc(rs(" -+-", "-_#]"), rs(" =+=", "-_#]"))
	skDrone = rt(rs("▂✕▂", " ▾ "), rs("ccc", " c "), theme.SkyNeonCyan, theme.SkyNeonCyan, mapmodel.SymDrone).lit().step(rs("▂╳▂", " ▾ ")).
		asc(rs("_x_", " v "), rs("_X_", " v "))
	skShuttle = rt(rs("░◢█▄", "▒▀██"), rs("sbbb", "lbbb"), theme.SkyCarWhite, theme.SkyExhaust, mapmodel.SymShuttle).
			asc(rs("./#_", ":\"##"), nil)
	skStarship = rt(rs("═▄█▄", "═▀██"), rs("cbbb", "cbbb"), theme.SkyAircraftLate, theme.SkyNeonCyan, mapmodel.SymStarship).lit().
			asc(rs("=_#_", "=\"##"), nil)
)

// ---------------------------------------------------------- the table

// rentry is one kind of traffic: the drawings it uses (a train has one per
// length), the lane it runs in, the ages it belongs to and its speed in
// columns a frame. The colour role is the drawing's body and accent hue.
type rentry struct {
	name   string
	lane   laneKind
	art    []*vtemplate // a variant is picked per vehicle; trains: 2, 3 and 4 carriages
	parts  []*vtemplate // the units the art is built from, each at most 4 x 2
	from   int          // the first age it is common in
	until  int          // the last age it is common in
	rare   int          // the last age it still turns up rarely (>= until)
	lo, hi float64
}

// train builds the drawings of a train: its carriages, then the engine at
// the front (the east end), with 2, 3 and 4 carriages.
func train(engine, car *vtemplate) []*vtemplate {
	var out []*vtemplate
	for n := 2; n <= 4; n++ {
		out = append(out, join(append(carriages(car, n), engine)))
	}
	return out
}

func carriages(t *vtemplate, n int) []*vtemplate {
	out := make([]*vtemplate, n)
	for i := range out {
		out[i] = t
	}
	return out
}

// join sets drawings side by side, bottoms aligned: a train, a group.
// gap cells of space go between them.
func join(ps []*vtemplate) *vtemplate { return joinGap(ps, 0, nil) }

// joinGap is join with gap blank cells between the drawings; digits, if not
// nil, gives each drawing's clothes digit for the mask letter 1.
func joinGap(ps []*vtemplate, gap int, digits []byte) *vtemplate {
	h := 1
	for _, p := range ps {
		h = max(h, len(p.rows))
	}
	var rows, mask, alt, asc, ascAlt []string
	for r := 0; r < h; r++ {
		var u, m, a, c, ca []rune
		for i, p := range ps {
			if i > 0 {
				for k := 0; k < gap; k++ {
					u, m, a, c, ca = append(u, ' '), append(m, ' '), append(a, ' '), append(c, ' '), append(ca, ' ')
				}
			}
			w := p.width()
			pick := func(rows []string, fallback []string) []rune {
				if rows == nil {
					rows = fallback
				}
				off := h - len(rows) // bottoms aligned: short drawings start lower
				if r < off {
					return []rune(strings.Repeat(" ", w))
				}
				return []rune(rows[r-off])
			}
			u = append(u, pick(p.rows, nil)...)
			mm := pick(p.mask, nil)
			if digits != nil {
				for j := range mm {
					if mm[j] == '1' {
						mm[j] = rune(digits[i])
					}
				}
			}
			m = append(m, mm...)
			a = append(a, pick(p.alt, p.rows)...)
			c = append(c, pick(p.ascii, p.rows)...)
			ca = append(ca, pick(firstRows(p.asciiAlt, p.ascii, p.alt, p.rows), nil)...)
		}
		rows, mask = append(rows, string(u)), append(mask, string(m))
		alt, asc, ascAlt = append(alt, string(a)), append(asc, string(c)), append(ascAlt, string(ca))
	}
	t := vt(rows, mask, ps[0].body, ps[0].accent, ps[len(ps)-1].sym)
	last := ps[len(ps)-1]
	t.alt, t.ascii, t.asciiAlt = alt, asc, ascAlt
	t.paint, t.emit, t.steam = last.paint, last.emit, last.steam
	t.people = digits != nil
	return &t
}

func firstRows(l ...[]string) []string {
	for _, r := range l {
		if r != nil {
			return r
		}
	}
	return nil
}

// width is the drawing's widest row.
func (t *vtemplate) width() int {
	w := 0
	for _, r := range t.rows {
		w = max(w, len([]rune(r)))
	}
	return w
}

func one1(t *vtemplate) []*vtemplate { return []*vtemplate{t} }

// rosterEntries is the roster: what each age's three lanes carry. A lane
// with no entry for an age stays empty.
var rosterEntries = []rentry{
	// the road
	{"bundle walker", laneRoad, one1(pBundle), one1(pBundle), agePrimitive, ageStone, ageIron, 0.05, 0.10},
	{"spear walker", laneRoad, one1(pSpear), one1(pSpear), agePrimitive, ageStone, ageBronze, 0.05, 0.10},
	{"pack dog", laneRoad, one1(tpDog), one1(tpDog), agePrimitive, ageBronze, ageMedieval, 0.10, 0.20},
	{"goat", laneRoad, one1(tpGoat), one1(tpGoat), agePrimitive, ageIron, ageMedieval, 0.08, 0.16},
	{"ox cart", laneRoad, one1(tpOxCart), one1(tpOxCart), ageBronze, ageClassical, ageIndustrial, 0.12, 0.22},
	{"camel", laneRoad, one1(tpCamel), one1(tpCamel), ageBronze, ageIron, ageMedieval, 0.10, 0.18},
	{"chariot", laneRoad, one1(tpChariot), one1(tpChariot), ageBronze, ageIron, ageClassical, 0.25, 0.40},
	{"horse and rider", laneRoad, one1(tpRider), one1(tpRider), ageClassical, ageRenaissance, ageModern, 0.20, 0.35},
	{"wagon", laneRoad, one1(tpWagon), one1(tpWagon), ageClassical, ageRenaissance, ageIndustrial, 0.14, 0.24},
	{"carriage", laneRoad, one1(tpCarriage), one1(tpCarriage), ageRenaissance, ageIndustrial, ageElectric, 0.20, 0.32},
	{"tram", laneRoad, one1(tpTram), one1(tpTram), ageVictorian, ageElectric, ageAtomic, 0.18, 0.28},
	{"motor car", laneRoad, one1(tpEarlyCar), one1(tpEarlyCar), ageElectric, ageElectric, ageAtomic, 0.22, 0.34},
	{"car", laneRoad, one1(tpCar), one1(tpCar), ageAtomic, ageDigital, ageCyberpunk, 0.35, 0.70},
	{"bus", laneRoad, one1(tpBus), one1(tpBus), ageAtomic, ageModern, ageDigital, 0.25, 0.40},
	{"hover bike", laneRoad, one1(tpHoverBike), one1(tpHoverBike), ageCyberpunk, ageCyberpunk, ageFusion, 0.50, 0.90},
	{"hover car", laneRoad, one1(tpHoverCar), one1(tpHoverCar), ageFusion, ageTranscendent, ageTranscendent, 0.40, 0.80},
	// the rails
	{"steam train", laneRail, train(trSteamEngine, trSteamCar), []*vtemplate{trSteamEngine, trSteamCar}, ageIndustrial, ageElectric, ageModern, 0.16, 0.28},
	{"train", laneRail, train(trDieselEngine, trDieselCar), []*vtemplate{trDieselEngine, trDieselCar}, ageAtomic, ageModern, ageModern, 0.25, 0.40},
	{"bullet train", laneRail, train(trBulletHead, trBulletCar), []*vtemplate{trBulletHead, trBulletCar}, ageInformation, ageDigital, ageDigital, 0.60, 0.90},
	{"maglev", laneRail, train(trMaglevHead, trMaglevCar), []*vtemplate{trMaglevHead, trMaglevCar}, ageCyberpunk, ageCyberpunk, ageFusion, 0.70, 1.10},
	{"portal tram", laneRail, train(trPortalHead, trPortalCar), []*vtemplate{trPortalHead, trPortalCar}, ageInterstellar, ageTranscendent, ageTranscendent, 0.60, 1.00},
	// the sky
	{"biplane", laneSky, one1(skBiplane), one1(skBiplane), ageElectric, ageElectric, ageAtomic, 0.25, 0.45},
	{"propeller plane", laneSky, one1(skProp), one1(skProp), ageAtomic, ageModern, ageDigital, 0.30, 0.55},
	{"jet", laneSky, one1(skJet), one1(skJet), ageAtomic, ageDigital, ageDigital, 0.50, 0.90},
	{"airliner", laneSky, one1(skAirliner), one1(skAirliner), ageModern, ageDigital, ageDigital, 0.40, 0.70},
	{"helicopter", laneSky, one1(skHeli), one1(skHeli), ageInformation, ageDigital, ageDigital, 0.20, 0.40},
	{"drone", laneSky, one1(skDrone), one1(skDrone), ageCyberpunk, ageCyberpunk, ageFusion, 0.30, 0.60},
	{"shuttle", laneSky, one1(skShuttle), one1(skShuttle), ageFusion, ageSpace, ageInterstellar, 0.40, 0.70},
	{"starship", laneSky, one1(skStarship), one1(skStarship), ageInterstellar, ageTranscendent, ageTranscendent, 0.50, 0.90},
}

// ageLane is what one lane carries in one age: the common entries and the
// rare older ones.
type ageLane struct{ common, rare []*rentry }

// ageRoster is rosterEntries by age and lane.
var ageRoster = func() (out [numAges][numLanes]ageLane) {
	for i := range rosterEntries {
		e := &rosterEntries[i]
		for a := e.from; a <= e.rare && a < numAges; a++ {
			l := &out[a][e.lane]
			if a <= e.until {
				l.common = append(l.common, e)
			} else {
				l.rare = append(l.rare, e)
			}
		}
	}
	return out
}()

// groupArt is the drawings of a group of people: by era (0 the first ages
// with their bundles and spears, 1 the rest), size (2 to 5) and a few
// mixes of men, women and children.
var groupArt = func() (out [2][4][6]*vtemplate) {
	pools := [2][]*vtemplate{{pBundle, pSpear, pMan, pWoman, pChild, pBundle}, {pMan, pWoman, pChild, pMan, pWoman}}
	for e := range out {
		for n := range out[e] {
			for v := range out[e][n] {
				ps := make([]*vtemplate, n+2)
				digits := make([]byte, n+2)
				for j := range ps {
					ps[j] = pools[e][hash(e, n, v, j)%uint64(len(pools[e]))]
					digits[j] = byte('1' + (v+j)%5)
				}
				if e == 1 && ps[0] == pChild { // a child walks beside, never ahead of, the grown-ups
					ps[0], ps[len(ps)-1] = ps[len(ps)-1], ps[0]
				}
				out[e][n][v] = joinGap(ps, 1, digits)
			}
		}
	}
	return out
}()

// groupCount is how many groups of people are out: more as the town works.
func groupCount(m *mapmodel.Model) int {
	st := m.Activity.Staffed
	if st <= 0 {
		return 0
	}
	return clampInt(1+int(mapmodel.Log2(1+float64(st))*0.5), 1, 7)
}

// rosterCounts is how many vehicles run in each lane.
func rosterCounts(m *mapmodel.Model) (road, rail, sky int) {
	a := m.Activity
	return 1 + int(a.Wealth*5), 1 + int(a.Traffic*2), 1 + int(a.Traffic*2.5)
}

// hasRailLane reports whether the age's frame lays its railway: the age's
// roster has trains.
func hasRailLane(age int) bool {
	if age < 0 || age >= numAges {
		return false
	}
	l := ageRoster[age][laneRail]
	return len(l.common)+len(l.rare) > 0
}

// pick chooses stream i's entry in the lane for the age: usually a common
// one, one stream in eight a rare older one.
func (l *ageLane) pick(seed int) *rentry {
	pool := l.common
	if len(l.rare) > 0 && (len(pool) == 0 || hash(seed, 71)%8 == 0) {
		pool = l.rare
	}
	if len(pool) == 0 {
		return nil
	}
	return pool[hash(seed, 73)%uint64(len(pool))]
}

// rosterTraffic is the roster's vehicles for a frame, in frame terms:
// groups of people and the age's road, rail and sky traffic.
func rosterTraffic(m *mapmodel.Model, anim, cam, w, groundY int) []vehicle {
	age := clampInt(m.AgeIdx, 0, numAges-1)
	var out []vehicle
	add := func(t *vtemplate, seed, y int, d uint8, lo, hi float64) {
		v := vehicle{kind: vkLife, t: t, seed: seed, depth: d, y: y}
		v.west = hash(seed, 17)%2 == 0
		speed := lo + (hi-lo)*hashf(seed, 13)
		if v.west {
			speed = -speed
		}
		pad := max(12, v.width()+2)
		span := w + 2*pad
		pos := float64(hash(seed, 23)%uint64(span)) + float64(anim)*speed - float64(cam)
		x := int(math.Floor(pos)) % span
		if x < 0 {
			x += span
		}
		v.x = x - pad
		if t.paint {
			v.body = [4]theme.SkyHue{theme.SkyCarRed, theme.SkyCarBlue, theme.SkyCarWhite, theme.SkyCarYellow}[hash(seed, 3)%4]
		}
		out = append(out, v)
	}
	road := groundY + 1
	era := 1
	if age <= ageStone {
		era = 0
	}
	// people, in groups of two to five at their own pace
	for i := 0; i < groupCount(m); i++ {
		seed := 1300 + i
		n := int(hash(seed, 61) % 4)
		add(groupArt[era][n][hash(seed, 62)%6], seed, road, dStreet, 0.05, 0.14)
	}
	nroad, nrail, nsky := rosterCounts(m)
	for i := 0; i < nroad; i++ {
		seed := 1400 + i
		if e := ageRoster[age][laneRoad].pick(seed); e != nil {
			add(e.art[hash(seed, 79)%uint64(len(e.art))], seed, road, dStreet, e.lo, e.hi)
		}
	}
	for i := 0; i < nrail; i++ {
		seed := 1500 + i
		if e := ageRoster[age][laneRail].pick(seed); e != nil {
			add(e.art[hash(seed, 79)%uint64(len(e.art))], seed, groundY+2, dStreet-1, e.lo, e.hi)
		}
	}
	for i := 0; i < nsky; i++ {
		seed := 1600 + i
		if e := ageRoster[age][laneSky].pick(seed); e != nil {
			add(e.art[hash(seed, 79)%uint64(len(e.art))], seed, skyLaneY(groundY, i), dAirLow, e.lo, e.hi)
		}
	}
	return out
}

// skyLaneY is the bottom row of sky vehicle i: three sub-lanes above the
// roofs, well over the street.
func skyLaneY(groundY, i int) int {
	return max(3, int(float64(groundY)*[3]float64{0.52, 0.38, 0.24}[i%3]))
}
