package roguelike

import (
	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapmodel"
)

// glyph.go resolves one world tile to a glyph, and keeps the legend of what
// the frame actually drew.

type lgID uint8

const (
	lgWater lgID = iota
	lgShallows
	lgForest
	lgHills
	lgMountains
	lgStreet
	lgAvenue
	lgBridge
	lgTrail
	lgPalisade
	lgWall
	lgTower
	lgCentre
	lgWonder
	lgRising
	lgRuin
	lgLegacy
	lgUnder
	lgFresh
	lgFlagged
	lgCiv
	lgWorker
	lgIdle
	lgCaravan
	lgStalled
	lgScout
	lgRaider
	lgWar
	lgHarbinger
	lgHazard
	lgSmoke
	lgRail
	lgGuideway
	// lgMover plus a mapmodel.Mover is that mover's row (the walkers keep
	// lgWorker); lgLineage plus an index in mapmodel.LineageOrder is that
	// lineage's.
	lgMover
	lgLineage = lgMover + lgID(mapmodel.NumMovers)
	numLg     = lgLineage + 18
)

var lgInfo = [lgMover]struct {
	label string
	group uint8
}{
	{"water", 0}, {"shallows", 0}, {"forest", 0}, {"hills", 0}, {"mountains", 0},
	{"street", 2}, {"avenue", 2}, {"bridge or ford", 2}, {"trail to a civ", 2}, {"palisade", 2}, {"city wall", 2},
	{"wall tower", 2}, {"town square", 1}, {"wonder", 1}, {"wonder rising", 1}, {"ruins", 4}, {"legacy (old town)", 1},
	{"understaffed (dim)", 4}, {"new since last visit", 4}, {"short of hands", 4}, {"civ settlement", 3},
	{"worker at work", 5}, {"idle worker", 5}, {"caravan", 5}, {"route disrupted", 5}, {"scouts", 5}, {"raiders", 5},
	{"war on this trail", 3}, {"harbinger", 4}, {"catastrophe", 4}, {"smoke over works", 5},
	{"railway", 2}, {"maglev line", 2},
}

// lgTraffic is the legend group the movers list under.
const lgTraffic = 6

var lgGroups = [7]string{"land", "buildings", "ways and walls", "civs", "state", "life", "traffic"}

// lgLabel is a legend row's label and group ("" for a row with none).
func lgLabel(id lgID) (string, uint8) {
	switch {
	case id < lgMover:
		return lgInfo[id].label, lgInfo[id].group
	case id < lgLineage:
		if k := mapmodel.Mover(id - lgMover); k != mapmodel.MoverWalker {
			return k.Info().Name, lgTraffic
		}
		return "", lgTraffic
	case int(id-lgLineage) < len(mapmodel.LineageOrder):
		return lineageLabel(mapmodel.LineageOrder[id-lgLineage]), 1
	}
	return "", 1
}

type lgEntry struct {
	on bool
	r  rune
	st tcell.Style
}

func (v *view) reg(id lgID, r rune, st tcell.Style) {
	if id < numLg && !v.seen[id].on {
		v.seen[id] = lgEntry{true, r, st}
	}
}

func (v *view) regG(id lgID, g glyph) {
	if id < numLg && !v.seen[id].on {
		v.seen[id] = lgEntry{true, g.r, v.style(g)}
	}
}

func lineageLabel(lin string) string {
	if lin == mapmodel.LinHarbor {
		return "harbor"
	}
	return mapmodel.LineageNames[lin]
}

func lineageIdx(lin string) int {
	for i, l := range mapmodel.LineageOrder {
		if l == lin {
			return i
		}
	}
	return 0
}

var (
	boxLight  = []rune("·│─└││┌├─┘─┴┐┤┬┼") // by mask N1 E2 S4 W8
	boxHeavy  = []rune("·┃━┗┃┃┏┣━┛━┻┓┫┳╋")
	boxDouble = []rune("·║═╚║║╔╠═╝═╩╗╣╦╬")
)

func (v *view) mask(x, y int, like func(kind) bool) int {
	m := 0
	for bit, d := range dirs4 {
		if like(v.sc.at(x+d[0], y+d[1]).k) {
			m |= 1 << bit
		}
	}
	return m
}

// wonder prefabs, [epoch][unicode, ascii]: each wonder keeps its era's look.
var prefabs = [7][2]string{
	{" ▲ ▲█▲", " A A#A"}, {"╓╥╖╨╨╨", ".n.|||"}, {"┌╬┐▐█▌", "+#+(#)"}, {" ▲ ▐▓▌", " ^ (#)"},
	{" ║ ▐█▌", " | (#)"}, {" ◊ ▀█▀", " I \"#\""}, {"(◙) ║ ", "(O) | "},
}
var scaffold = [2]string{"┌┬┐╫╫╫", "/-\\|#|"}

func runeAt(s string, i int) rune {
	for k, r := range []rune(s) {
		if k == i {
			return r
		}
	}
	return ' '
}

func superscript(n int) rune {
	if n < 10 {
		return [10]rune{'⁰', '¹', '²', '³', '⁴', '⁵', '⁶', '⁷', '⁸', '⁹'}[max(n, 0)]
	}
	return '⁺'
}

// flagged reports a building the flows overlay marks: short of hands.
func flagged(b *mapmodel.Building) bool {
	return b != nil && b.Count > 0 && (b.Understaffed() || b.Staffing == 0 && b.Capacity > 0)
}

// tile resolves one world tile at settlement scale.
func (v *view) tile(x, y int) glyph {
	s := v.sc
	if !s.in(x, y) || s.vis[y*s.w.W+x] == 0 {
		return glyph{r: ' ', sal: -1}
	}
	i := y*s.w.W + x
	t, c := s.w.T[i], s.cells[i]
	if s.vis[i] == 1 { // remembered, not in sight
		g := v.terrain(x, y, t, false)
		return glyph{r: g.r, c: mapmodel.CMemory, sal: min(g.sal, 1)}
	}
	if c.rail && len(s.rail) > 0 && c.k != kTile && c.k != kWonder && c.k != kSite && c.k != kCentre {
		return v.railGlyph(t)
	}
	var g glyph
	switch c.k {
	case kRoad:
		g = glyph{r: '·', c: mapmodel.CRoad, sal: 40}
		v.regG(lgTrail, g)
	case kStreet:
		if s.d.road == 0 { // no streets yet: trampled ground
			g = v.terrain(x, y, t, true)
			g.c, g.sal = mapmodel.CGround, 2
			break
		}
		mk := v.mask(x, y, roadLike)
		g = glyph{r: boxLight[mk], c: mapmodel.CRoad, sal: 40}
		switch {
		case !c.major:
			v.regG(lgStreet, glyph{r: '┼', c: g.c})
		case s.d.road == 1:
			g.c = mapmodel.CWall
		case s.d.road == 2:
			g.r, g.c = boxHeavy[mk], mapmodel.CWall
		default:
			g.r, g.c = boxDouble[mk], mapmodel.CWall
		}
		if c.major {
			v.regG(lgAvenue, glyph{r: [4]rune{'┼', '┼', '╋', '╬'}[s.d.road], c: g.c})
		}
	case kBridge:
		g = glyph{r: '═', c: mapmodel.CWall, bg: v.pal.WaterBg, sal: 41}
		if mk := v.mask(x, y, roadLike); mk&5 != 0 && mk&10 == 0 {
			g.r = '║'
		}
		if s.d.road == 0 {
			g.r = '='
		}
		v.regG(lgBridge, g)
	case kWall:
		if s.d.wall == 1 {
			g = glyph{r: '#', c: mapmodel.CRock, sal: 50}
			v.regG(lgPalisade, g)
			break
		}
		mk := v.mask(x, y, func(k kind) bool { return k == kWall || k == kTower })
		g = glyph{r: boxDouble[mk], c: mapmodel.CWall, sal: 50}
		if mk == 0 {
			g.r = '■'
		}
		v.regG(lgWall, glyph{r: '═', c: g.c})
	case kTower:
		g = glyph{r: '■', c: mapmodel.CWall, attr: tcell.AttrBold, sal: 51}
		v.regG(lgTower, g)
	case kPlaza:
		g = glyph{r: '·', c: mapmodel.CRoad, sal: 30}
		if s.epoch == 0 {
			g = v.terrain(x, y, t, true)
			g.c, g.sal = mapmodel.CGround, 30
		}
	case kCentre:
		g = glyph{r: mapmodel.R(mapmodel.CentreSym(s.epoch), v.tier), c: mapmodel.CWealth, attr: tcell.AttrBold, sal: 95}
		if s.epoch == 0 { // the hearth flickers
			g.c = [3]mapmodel.Class{mapmodel.CIdle, mapmodel.CIdle, mapmodel.CDanger}[(v.anim/2+1)%3]
		}
		v.regG(lgCentre, g)
	case kTile:
		g = v.building(x, y, &s.m.Town.Tiles[c.ref])
	case kWonder:
		g = v.wonder(x, y, &s.m.Town.Wonders[c.ref])
	case kSite:
		if f := &s.m.Factions[c.civ-1]; f.Discovered {
			g = glyph{r: '▪', c: mapmodel.RelationClass(f.Relation), sal: 100}
			v.regG(lgCiv, g)
			if c.ref >= 0 { // the seat: its initial, or the nerd icon
				g.r, g.attr = []rune(f.Name + "C")[0], tcell.AttrBold|tcell.AttrReverse
				if v.tier == mapmodel.TierNerd {
					g.r, g.attr = mapmodel.R(mapmodel.SymCiv, v.tier), tcell.AttrBold
				}
			}
			break
		}
		fallthrough
	default:
		g = v.terrain(x, y, t, false)
	}
	return g
}

func (v *view) building(x, y int, tt *mapmodel.TownTile) glyph {
	p := v.pal
	b := v.sc.m.Building(tt.Key)
	if tt.Ruin || b == nil {
		g := glyph{r: mapmodel.R(mapmodel.SymRuin, v.tier), c: mapmodel.CRock, sal: 45}
		v.regG(lgRuin, g)
		return g
	}
	g := glyph{r: mapmodel.R(mapmodel.LineageSym(tt.Lineage, v.sc.epoch), v.tier), c: mapmodel.LineageClass(tt.Lineage), sal: 60}
	if tt.Lineage == mapmodel.LinHousing || tt.Lineage == mapmodel.LinFood {
		g.sal = 55
	}
	if g.r == '"' && mapmodel.Hash(int64(x), int64(y), int64(v.anim/2))%9 == 0 {
		g.r = '\'' // wind in the crops
	}
	switch {
	case tt.Legacy:
		g.c = mapmodel.CLegacy
		v.regG(lgLegacy, g)
	case b.Understaffed():
		v.regG(lgLineage+lgID(lineageIdx(tt.Lineage)), g)
		g.fg = p.dim
		v.regG(lgUnder, g)
	default:
		v.regG(lgLineage+lgID(lineageIdx(tt.Lineage)), g)
	}
	switch {
	case v.flows && flagged(b):
		g.bg = p.FlowBg
		warn := glyph{r: mapmodel.R(mapmodel.SymWarning, v.tier), c: mapmodel.CIdle, bg: p.FlowBg, attr: tcell.AttrBold, sal: g.sal}
		v.regG(lgFlagged, warn)
		if tt.Anchor {
			return warn
		}
	case v.changes && tt.Fresh:
		g.bg, g.sal = p.FreshBg, g.sal+5
		v.regG(lgFresh, glyph{r: ' ', bg: p.FreshBg})
	}
	return g
}

func (v *view) wonder(x, y int, wd *mapmodel.TownWonder) glyph {
	i, asc := (y-wd.At.Y)*mapmodel.WonderW+x-wd.At.X, 0
	if v.tier == mapmodel.TierASCII {
		asc = 1
	}
	g := glyph{r: runeAt(scaffold[asc], i), c: mapmodel.CWork, sal: 90}
	if wd.Built {
		ep := v.sc.epoch
		if d := v.sc.m.Catalog.Defs[wd.Key]; d != nil {
			ep = epochIdx(d.Epoch)
		}
		g.r, g.c, g.attr = runeAt(prefabs[ep][asc], i), mapmodel.CWealth, tcell.AttrBold
		if v.tier == mapmodel.TierNerd && i == 1 {
			g.r = mapmodel.R(mapmodel.SymWonder, v.tier)
		}
		v.regG(lgWonder, glyph{r: mapmodel.R(mapmodel.SymWonder, v.tier), c: g.c, attr: g.attr})
	} else {
		v.regG(lgRising, glyph{r: runeAt(scaffold[asc], 4), c: g.c})
	}
	if g.r == ' ' {
		g.sal = 25
	}
	if v.changes && wd.Fresh {
		g.bg = v.pal.FreshBg
		v.regG(lgFresh, glyph{r: ' ', bg: g.bg})
	}
	return g
}

// terrain is the ground's glyph; water shimmers with the frame.
func (v *view) terrain(x, y int, t mapmodel.Terrain, trampled bool) glyph {
	h := mapmodel.Hash(v.sc.w.Seed, 11, int64(x), int64(y))
	g := glyph{r: ' ', c: mapmodel.CGround}
	switch t {
	case mapmodel.TDeep, mapmodel.TShallow, mapmodel.TRiver:
		g = glyph{r: '≈', c: mapmodel.CWater, bg: v.pal.WaterBg, sal: [3]int{12, 14, 35}[t]}
		id := lgWater
		if t == mapmodel.TShallow {
			g.r, id = '~', lgShallows
		}
		v.regG(id, g)
		if mapmodel.Hash(int64(x), int64(y), int64(v.anim/2))%29 == 0 {
			g.r = '≈' + '~' - g.r // a glint: swap the two, now and then
		}
	case mapmodel.TBeach:
		g.r, g.sal = '·', 8
	case mapmodel.TForest:
		g = glyph{r: '♣', c: mapmodel.CFlora, sal: 20}
		v.regG(lgForest, g)
		if h%5 == 0 {
			g.r = 'τ'
		}
	case mapmodel.THills:
		g = glyph{r: '^', c: mapmodel.CHill, sal: 25}
		v.regG(lgHills, g)
	case mapmodel.TMountain:
		g = glyph{r: '▲', c: mapmodel.CRock, sal: 30}
		v.regG(lgMountains, g)
	default:
		switch {
		case trampled && h%3 == 0:
			g.r = '.'
		case !trampled && h%19 < 4:
			g.r, g.sal = [4]rune{'.', '.', ',', '\''}[h%19], 1
		}
	}
	return g
}

// rightHalf fills a district tile's second cell: ways and walls continue,
// water and fields repeat, a building's anchor shows its count.
func (v *view) rightHalf(x, y int, g glyph) glyph {
	s := v.sc
	if !s.in(x, y) || s.vis[y*s.w.W+x] == 0 {
		return glyph{r: ' '}
	}
	blank, c, east := glyph{r: ' ', bg: g.bg}, s.at(x, y), s.at(x+1, y).k
	if s.vis[y*s.w.W+x] == 1 {
		return blank
	}
	if c.rail && len(s.rail) > 0 && s.at(x+1, y).rail {
		if g.r == '╪' {
			g.r = '═' // sleepers on every other cell
		}
		return g
	}
	switch c.k {
	case kStreet:
		if s.d.road > 0 && roadLike(east) {
			g.r = '─'
			if c.major {
				g.r = [4]rune{'─', '─', '━', '═'}[s.d.road]
			}
			return g
		}
	case kWall:
		if east == kWall || east == kTower {
			g.r = [4]rune{'═', '#', '═', '═'}[s.d.wall]
			return g
		}
	case kBridge:
		if roadLike(east) {
			return g
		}
	case kTile:
		tt := &s.m.Town.Tiles[c.ref]
		if b := s.m.Building(tt.Key); tt.Anchor && b != nil {
			g.r, g.attr = superscript(b.Count), 0
			if g.c == mapmodel.CIdle {
				g.c = mapmodel.LineageClass(tt.Lineage)
			}
			return g
		}
		if tt.Lineage == mapmodel.LinFood && !tt.Ruin {
			return g
		}
	case kWonder:
		if east == kWonder {
			switch g.r {
			case '█', '▓', '▀', '#', '|':
				return g
			case '╓', '╥', '┌', '╬', '╨', '┬', '╫':
				g.r = '─'
				return g
			}
		}
	case kNone, kPlaza:
		if t := s.w.At(x, y); t.Water() || t == mapmodel.TForest && mapmodel.Hash(int64(x), int64(y))%2 == 0 {
			return g
		}
	}
	return blank
}
