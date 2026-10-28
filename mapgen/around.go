package mapgen

import (
	"cmp"
	"slices"

	"github.com/espresso20/ageforge/mapmodel"
)

// The world round the town. What the header and the news report, the map
// shows: each civilization met stands at the edge of the picture in its own
// direction, coloured by how it stands toward you; a trade route is a line of
// dots from the square toward its partner with caravans going back and forth
// on it; scouts and an army walk out along a road and back; the harbinger
// stands by the square, and idle workers loiter round it. In space the civs
// are other stars (on the star map, the systems themselves), the routes are
// lanes and the caravans ships. Like traffic, every figure is where it is by
// a function of its number and the frame alone, and none covers a building.

// How the figures of the world round the town look. A caravan is by cart,
// truck, hovercar, freighter or fleet by the age; a site is a flag on the
// land and a star in space.
var (
	byCart    = Figure{Glyphs: "&", Fg: P(HGold, 255)}
	byTruck   = Figure{Glyphs: "■", Fg: P(HGold, 255)}
	byHover   = Figure{Glyphs: "◆", Fg: P(HGold, 255)}
	byFreight = Figure{Glyphs: "►◄▼▲", Fg: P(HAmber, 255)}
	byFleet   = Figure{Glyphs: "◆", Fg: P(HGold, 255)}
	siteFlag  = Figure{Glyphs: "[▪]"}
	siteStar  = Figure{Glyphs: "✦"}
	scout     = Figure{Glyphs: "@", Fg: P(HLife, 255)}
	raider    = Figure{Glyphs: "»", Fg: P(HRed, 255)}
	idler     = Figure{Glyphs: "☺☻", Fg: P(HIdle, 255)}
)

// omen is the harbinger's figure in an age's colour.
func omen(hue Hue) Figure { return Figure{Glyphs: "Ψ", Fg: P(hue, 255)} }

// relationHue is the colour of a civilization by how it stands toward you.
var relationHue = [...]Hue{
	mapmodel.RelNeutral: HNeutral, mapmodel.RelFriendly: HFriend, mapmodel.RelAllied: HAlly,
	mapmodel.RelRival: HRival, mapmodel.RelEmbargo: HRival, mapmodel.RelWar: HWar,
}

// trailGlyph is the dot a route is drawn with, by scale.
var trailGlyph = [numScales]rune{Ground: '·', City: '·', Orbit: '∙', System: '∙', Stars: '∙'}

// edgeGapX and edgeGapY are how far the edge of the picture lies beyond the
// town, in cells.
const (
	edgeGapX, edgeGapY = 8, 3
	siteGap            = 9 // units of border between two sites
)

// site is a civilization met and where it stands: at the edge of the picture,
// or, on the star map, at a system (node).
type site struct {
	civ  *mapmodel.Faction
	at   Pt
	node int
}

// paintAround draws the world round the town, over the town and its traffic.
func (f *frame) paintAround() {
	f.sites = f.civSites()
	f.paintTrails()
	f.paintSites()
	f.paintParties()
	f.paintSquare()
}

// put draws a glyph at a cell, unless a building stands there, and says what
// it is; it reports whether it drew.
func (f *frame) put(p Pt, r rune, ink Paint, th Thing) bool {
	if !f.g.in(p.X, p.Y) || f.solid(p.X, p.Y) {
		return false
	}
	f.g.draw(p.X, p.Y, r, ink)
	f.g.mark(p.X, p.Y, th)
	return true
}

// open says whether a line of dots may be drawn on a cell: nothing is built
// or laid there, and in space it is empty sky.
func (f *frame) open(p Pt) bool {
	if !f.g.in(p.X, p.Y) || f.busy[at(p.X, p.Y)] || f.solid(p.X, p.Y) {
		return false
	}
	th := f.g.ThingAt(p.X, p.Y)
	return f.st.Scale <= City || th.Kind == KindTerrain && th.Key == "space"
}

// square is the town's heart: the square on the land, the station at the head
// of the elevator in orbit and the solar system.
func (f *frame) square() Pt {
	if f.st.Scale == Orbit || f.st.Scale == System {
		return f.plan.Roads[0].Cells[0]
	}
	return f.plan.Centre
}

// edge is the box the picture of the world reaches to: the town and a gap
// round it, kept inside the world.
func (f *frame) edge() Rect {
	if f.box.W == 0 {
		b := f.bounds()
		x0, y0 := max(b.X-edgeGapX, 1), max(b.Y-edgeGapY, 1)
		x1, y1 := min(b.X+b.W+edgeGapX, worldW-1), min(b.Y+b.H+edgeGapY, worldH-1)
		f.box = Rect{x0, y0, x1 - x0, y1 - y0}
	}
	return f.box
}

// perimeter is the length of the border of r in units: a row counts for two,
// being twice as tall as a column is wide.
func (r Rect) perimeter() int { return 2 * (max(1, r.W-1) + 2*max(1, r.H-1)) }

// borderAt is the cell t units round the border of r: along the top, down the
// right side, back along the bottom and up the left.
func (r Rect) borderAt(t int) Pt {
	w, h := max(1, r.W-1), 2*max(1, r.H-1)
	switch t %= 2 * (w + h); {
	case t < w:
		return Pt{r.X + t, r.Y}
	case t < w+h:
		return Pt{r.X + w, r.Y + (t-w)/2}
	case t < 2*w+h:
		return Pt{r.X + w - (t - w - h), r.Y + r.H - 1}
	}
	return Pt{r.X, r.Y + r.H - 1 - (t-2*w-h)/2}
}

// civSites places every civilization met, in the order of the model's list.
func (f *frame) civSites() []site {
	var out []site
	for i := range f.m.Factions {
		if f.m.Factions[i].Discovered {
			out = append(out, f.siteOf(&f.m.Factions[i], out))
		}
	}
	return out
}

// siteOf is where a civilization stands, clear of the sites already placed and
// of the town: at
// a system of its own on the star map, else at a spot of the border of the
// picture that its key picks, in its direction from the town.
func (f *frame) siteOf(civ *mapmodel.Faction, placed []site) site {
	key, seed := mapmodel.HashStr(civ.Key), f.m.Seed
	if f.st.Scale == Stars {
		nodes := len(f.plan.Nodes) - 1
		n := 1 + roll(nodes, seed, 100, key)
		for taken := true; taken; {
			taken = slices.ContainsFunc(placed, func(s site) bool { return s.node == n })
			if taken {
				n = n%nodes + 1
			}
		}
		return site{civ, f.plan.Nodes[n], n}
	}
	edge := f.edge()
	t := roll(edge.perimeter(), seed, 101, key)
	var at Pt
	for try := 0; try < 24; try++ {
		at = edge.borderAt(t + try*siteGap)
		at.X = clamp(at.X, edge.X+1, edge.X+edge.W-2)
		crowded := slices.ContainsFunc(placed, func(s site) bool { return abs(s.at.X-at.X) < 5 && abs(s.at.Y-at.Y) < 2 })
		if !crowded && !f.solid(at.X-1, at.Y) && !f.solid(at.X, at.Y) && !f.solid(at.X+1, at.Y) {
			break
		}
	}
	return site{civ, at, 0}
}

// siteOfKey is the site of the civ with key k.
func (f *frame) siteOfKey(k string) (site, bool) {
	i := slices.IndexFunc(f.sites, func(s site) bool { return s.civ.Key == k })
	if i < 0 {
		return site{}, false
	}
	return f.sites[i], true
}

// journey is the way from the square to a place, not counting the square: the
// cells of the road toward it and the open ground beyond, a straight stair
// across the sky, or, on the star map, the routes from the home system to the
// node.
func (f *frame) journey(to Pt, node int) []Pt {
	home := f.square()
	switch {
	case f.st.Scale == Stars:
		return f.plan.chain(f.plan.NodeRoad[node])[1:]
	case f.st.Scale > City:
		return staircase(home, to)
	}
	arm, gap := []Pt(nil), 1<<60
	for i, road := range f.plan.Roads {
		end := road.Cells[len(road.Cells)-1]
		if d := (Rect{end.X, end.Y, 1, 1}).dist(to); i > 0 && road.Parent == 0 && road.Tier == 0 && d < gap {
			arm, gap = road.Cells, d
		}
	}
	if len(arm) == 0 {
		return staircase(home, to)
	}
	return append(slices.Clone(arm), staircase(arm[len(arm)-1], to)...)
}

// frontier is a place at the edge of the picture a named expedition makes for
// (a system, on the star map).
func (f *frame) frontier(name string) (Pt, int) {
	key := mapmodel.HashStr(name)
	if f.st.Scale == Stars {
		n := 1 + roll(len(f.plan.Nodes)-1, f.m.Seed, 102, key)
		return f.plan.Nodes[n], n
	}
	edge := f.edge()
	return edge.borderAt(roll(edge.perimeter(), f.m.Seed, 102, key)), 0
}

// paintTrails draws each trade route's line once, to its partner, and the
// caravans that go back and forth on every route.
func (f *frame) paintTrails() {
	drawn := map[string]bool{}
	lane := 0
	if f.st.Scale > City {
		lane = 1
	}
	for i, route := range f.m.Routes {
		s, ok := f.siteOfKey(route.Civ)
		if !ok {
			continue
		}
		cells, thing := f.journey(s.at, s.node), Thing{Kind: KindRoute, Key: route.Civ, Ref: lane}
		if len(cells) < 3 {
			continue
		}
		if !drawn[route.Civ] {
			drawn[route.Civ] = true
			for _, p := range cells[:len(cells)-1] {
				if f.open(p) {
					f.g.draw(p.X, p.Y, trailGlyph[f.st.Scale], P(relationHue[s.civ.Relation], 200))
					f.g.mark(p.X, p.Y, thing)
				}
			}
		}
		f.paintCaravans(cells, i, route.Disrupted, thing)
	}
}

// paintCaravans puts two caravans on a route, half a journey apart; a route
// that is disrupted has one stalled a third of the way.
func (f *frame) paintCaravans(cells []Pt, route int, disrupted bool, th Thing) {
	car := f.st.Caravan
	if disrupted {
		f.put(cells[len(cells)/3], []rune(car.Glyphs)[0], P(HWar, 255), th)
		return
	}
	for c := 0; c < 2; c++ {
		p, dx, dy := pace(cells, c*(len(cells)-1)+route*7, f.anim, 2)
		f.put(p, heading(car.Glyphs, dx, dy), car.Fg, th)
	}
}

// paintSites marks each civilization met, and lets the cursor visit them.
func (f *frame) paintSites() {
	runes := []rune(f.st.Site.Glyphs)
	for _, s := range f.sites {
		thing := Thing{Kind: KindCiv, Key: s.civ.Key}
		for i, r := range runes {
			f.put(Pt{s.at.X - len(runes)/2 + i, s.at.Y}, r, P(relationHue[s.civ.Relation], 255), thing)
		}
		f.targets = append(f.targets, s.at)
	}
}

// paintParties walks the scouts, and the army toward the civilization that
// likes you least, out along their way and back.
func (f *frame) paintParties() {
	e := f.m.Expeditions
	if e.Scout != nil {
		to, node := f.frontier(e.Scout.Name)
		f.paintParty(f.journey(to, node), scout, 3, "scout")
	}
	if e.Military == nil {
		return
	}
	to, node := f.frontier(e.Military.Name)
	if foe, ok := f.leastLiked(); ok {
		to, node = foe.at, foe.node
	}
	cells := f.journey(to, node)
	f.paintParty(cells[:max(2, len(cells)*2/3)], raider, 2, "military")
}

// leastLiked is the civilization met that likes you least, if any.
func (f *frame) leastLiked() (site, bool) {
	if len(f.sites) == 0 {
		return site{}, false
	}
	return slices.MinFunc(f.sites, func(a, b site) int { return cmp.Compare(a.civ.Opinion, b.civ.Opinion) }), true
}

// paintParty puts one figure on a journey, going out and back.
func (f *frame) paintParty(cells []Pt, who Figure, slow int, key string) {
	if len(cells) < 2 {
		return
	}
	p, dx, dy := pace(cells, 0, f.anim, slow)
	f.put(p, heading(who.Glyphs, dx, dy), who.Fg, Thing{Kind: KindMission, Key: key})
}

// paintSquare stands the harbinger by the square, in the age's style, and the
// idle workers about it, one for six idle and nine at most.
func (f *frame) paintSquare() {
	spots := f.squareSpots()
	if f.m.Harbinger != nil && len(spots) > 0 {
		f.put(spots[0], []rune(f.st.Omen.Glyphs)[0], f.st.Omen.Fg, Thing{Kind: KindHarbinger, Key: f.m.Harbinger.Key})
		spots = spots[1:]
	}
	if f.m.Workers.Idle == 0 {
		return
	}
	runes := []rune(idler.Glyphs)
	for j, p := range spots[:min(len(spots), 9, 1+f.m.Workers.Idle/6)] {
		r := runes[0]
		if (f.anim/6+j)%4 == 0 { // now and then one sits down
			r = runes[1]
		}
		f.put(p, r, idler.Fg, Thing{Kind: KindIdle, Key: "idle"})
	}
}

// squareSpots are the cells round the square where nobody has built, nearest
// first.
func (f *frame) squareSpots() []Pt {
	sq := f.square()
	var spots []Pt
	for dy := -2; dy <= 2; dy++ {
		for dx := -4; dx <= 4; dx++ {
			if p := (Pt{sq.X + dx, sq.Y + dy}); (dx != 0 || dy != 0) && f.g.in(p.X, p.Y) && !f.solid(p.X, p.Y) {
				spots = append(spots, p)
			}
		}
	}
	slices.SortStableFunc(spots, func(a, b Pt) int {
		return cmp.Or(cmp.Compare(Rect{a.X, a.Y, 1, 1}.dist(sq), Rect{b.X, b.Y, 1, 1}.dist(sq)),
			cmp.Compare(hash(f.m.Seed, 103, int64(a.X), int64(a.Y)), hash(f.m.Seed, 103, int64(b.X), int64(b.Y))))
	})
	return spots
}
