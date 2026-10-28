package mapgen

import "github.com/espresso20/ageforge/mapmodel"

// The rare visitor. The model's schedule (mapmodel.Model.SightingAt) says
// when one is about; this layer says where. Most visits are a visitor: on the
// land a traveller who comes in along a trunk road from its far end and stops
// near the square, in space a ship that flies in from the edge of the map to
// a lane. One visit in alienOneIn is an alien's craft instead: a saucer that
// crosses the town, or drops in over it, hangs there and shoots off. Where a
// visitor is a function of the visit and the frame alone, so any frame draws
// without the ones before it. It is the last layer: it shines through the
// night and never covers a building.

// alienOneIn is how rare an alien is among the visits.
const alienOneIn = 3

// arrivePct is how much of a visit the visitor spends arriving, in percent;
// it stays where it stopped for the rest and is then gone.
const arrivePct = 60

// stopAt is how far from the square, in cells of the road, a traveller stops.
const stopAt = 3

// saucerReach is how far either side of the square a saucer crosses, and
// hoverRise how far above its spot it drops in from.
const (
	saucerReach = 60
	hoverRise   = 12
)

// How visits look: the traveller on foot, the ship in space and the alien's
// saucer, a dome with a rim to either side.
var (
	onFoot = Visitor{Glyphs: "☿", Fg: P(HFresh, 255)}
	byShip = Visitor{Glyphs: "►◄▼▲", Fg: P(HFresh, 255)}
	saucer = Visitor{Glyphs: "◄◉►", Fg: P(HFresh, 255)}
)

// paintVisitor draws the visit that is on at the frame's visit frame, if any.
func (f *frame) paintVisitor() {
	sg, ok := f.m.SightingAt(f.visit)
	if !ok {
		return
	}
	thing := Thing{Kind: KindVisitor, Key: "visitor", Ref: int(sg.Roll >> 24 % 1024)}
	look, cells, runes := f.st.Visitor, []Pt(nil), []rune(nil)
	if sg.Roll>>16%alienOneIn == 0 {
		thing.Key, look = "alien", saucer
		cells, runes = f.saucerCells(sg), []rune(saucer.Glyphs)
	} else {
		at, dx, dy, ok := f.arrival(sg)
		if !ok {
			return
		}
		cells, runes = []Pt{at}, []rune{heading(look.Glyphs, dx, dy)}
	}
	for i, p := range cells {
		if !f.g.in(p.X, p.Y) || f.solid(p.X, p.Y) {
			continue
		}
		f.g.draw(p.X, p.Y, runes[i], look.Fg)
		f.g.mark(p.X, p.Y, thing)
		f.g.Cells[at(p.X, p.Y)].Dim = 0
	}
}

// solid says whether a building stands at a cell: a visitor never covers one.
func (f *frame) solid(x, y int) bool {
	switch f.g.ThingAt(x, y).Kind {
	case KindBuilding, KindWonder, KindAmbient:
		return true
	}
	return false
}

// arrival is where the visitor who arrives is at the visit frame, and the way
// it is heading. It comes along its route in the first arrivePct of the visit
// and stays at the end of it; ok is false when it has no way in.
func (f *frame) arrival(sg mapmodel.Sighting) (at Pt, dx, dy int, ok bool) {
	route := f.footRoute(sg)
	if f.st.Scale > City {
		route = f.shipRoute(sg)
	}
	if len(route) < 2 {
		return Pt{}, 0, 0, false
	}
	last := len(route) - 1
	now := clamp((f.visit-sg.Start)*last/max(1, sg.Frames*arrivePct/100), 0, last)
	from := max(now-1, 0)
	return route[now], route[from+1].X - route[from].X, route[from+1].Y - route[from].Y, true
}

// footRoute is how a traveller walks in: along one of the trunk roads, from
// its far end to stopAt cells from the square.
func (f *frame) footRoute(sg mapmodel.Sighting) []Pt {
	var trunks [][]Pt
	for i, road := range f.plan.Roads {
		if i > 0 && road.Parent == 0 && road.Tier == 0 && len(road.Cells) > stopAt+4 {
			trunks = append(trunks, road.Cells)
		}
	}
	if len(trunks) == 0 {
		return nil
	}
	cells := trunks[roll(len(trunks), int64(sg.Roll), 1)]
	var route []Pt
	for i := len(cells) - 1; i >= stopAt; i-- {
		route = append(route, cells[i])
	}
	return route
}

// shipRoute is how a ship flies in: a straight stair from a spot on the left
// or the right edge of the map to a cell of one of the lanes that are laid.
func (f *frame) shipRoute(sg mapmodel.Sighting) []Pt {
	var lanes []int
	for i, n := range f.shown {
		if n >= 4 {
			lanes = append(lanes, i)
		}
	}
	target := f.plan.Centre
	if len(lanes) > 0 {
		lane := lanes[roll(len(lanes), int64(sg.Roll), 2)]
		target = f.plan.Roads[lane].Cells[roll(f.shown[lane], int64(sg.Roll), 3)]
	}
	edge := Pt{0, roll(worldH, int64(sg.Roll), 4)}
	if sg.Roll>>8&1 == 1 {
		edge.X = worldW - 1
	}
	return append([]Pt{edge}, staircase(edge, target)...)
}

// saucerCells are the three cells of the alien's saucer at the visit frame.
// It crosses the town above the square, or, on a hover, drops in over open
// ground near it, wobbles there and shoots off upward.
func (f *frame) saucerCells(sg mapmodel.Sighting) []Pt {
	t, c := f.visit-sg.Start, f.plan.Centre
	spot, found := Pt{}, false
	if sg.Kind != mapmodel.SightFlyby {
		want := Pt{c.X - 18 + int(sg.Roll>>4%37), c.Y - 7 + int(sg.Roll>>12%15)}
		spot, found = f.openSpot(want)
	}
	var dome Pt
	switch pct := t * 100 / sg.Frames; {
	case !found:
		x := c.X - saucerReach + t*2*saucerReach/sg.Frames
		if sg.Roll&1 == 1 {
			x = c.X + saucerReach - t*2*saucerReach/sg.Frames
		}
		drift := (int(sg.Roll>>8%7) - 3) * t / sg.Frames
		dome = Pt{x, clamp(c.Y-10+int(sg.Roll>>4%5)+drift, 1, worldH-2)}
	case pct < 15:
		dome = Pt{spot.X, spot.Y - hoverRise + hoverRise*pct/15}
	case pct < 85:
		dome = Pt{spot.X + [4]int{0, 1, 0, -1}[t/10%4], spot.Y}
	default:
		dome = Pt{spot.X, spot.Y - hoverRise*(pct-85)/15}
	}
	return []Pt{{dome.X - 1, dome.Y}, dome, {dome.X + 1, dome.Y}}
}

// openSpot is the cell nearest want with open ground to either side as well,
// searched outward in rings (a ring is twice as wide as tall), if the town
// leaves one.
func (f *frame) openSpot(want Pt) (Pt, bool) {
	open := func(x, y int) bool {
		for dx := -1; dx <= 1; dx++ {
			if !inWorld(x+dx, y) || f.busy[at(x+dx, y)] || f.solid(x+dx, y) {
				return false
			}
		}
		return true
	}
	for ring := 0; ring <= 6; ring++ {
		for dy := -ring; dy <= ring; dy++ {
			for dx := -2 * ring; dx <= 2*ring; dx++ {
				if open(want.X+dx, want.Y+dy) {
					return Pt{want.X + dx, want.Y + dy}, true
				}
			}
		}
	}
	return Pt{}, false
}
