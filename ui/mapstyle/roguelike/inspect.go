package roguelike

import (
	"strconv"
	"strings"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/ui/mapstyle"
)

// inspect.go says what the cursor is on, in the game's own words, and which
// command a player would type for it (from the model's helpers only).

type insp = mapstyle.Inspection

// Inspect reports what the cursor is on; ok is false while it is hidden.
func (v *view) Inspect(f mapstyle.Frame) (mapstyle.Inspection, bool) {
	s := v.sceneFor(f.Model)
	if s == nil || !v.inspect {
		return insp{}, false
	}
	v.palette(s.epoch, s.m.AgeIdx)
	v.tier, v.anim, v.visit, v.clock = f.Tier, f.Anim, f.VisitFrame(), f.Clock
	p, region := v.cur, v.zoom == zRegion && v.g.zoom == zRegion
	if region && v.g.scale > 1 { // the most salient tile in the cell
		g, best := &v.g, -3
		x0, y0 := g.vx+floorDiv(p.X-g.vx, g.scale)*g.scale, g.vy+floorDiv(p.Y-g.vy, g.scale)*g.scale
		for j := 0; j < g.scale*g.scale; j++ {
			if q := pt(x0+j%g.scale, y0+j/g.scale); v.tile(q.X, q.Y).sal > best {
				best, p = v.tile(q.X, q.Y).sal, q
			}
		}
	}
	in := v.describe(s, p)
	if pl := plates[s.epoch]; region && pl.refs {
		if cx, cy, ok := v.g.cellOf(v.cur); ok {
			in.Lines = append(in.Lines, "grid ref "+pad2((cx-v.g.x)/pl.grid[0])+pad2((cy-v.g.y)/pl.grid[1]))
		}
	}
	return in, true
}

func floorDiv(a, b int) int {
	if a < 0 {
		return -((-a + b - 1) / b)
	}
	return a / b
}

func has(ps []mapmodel.Pt, p mapmodel.Pt) bool {
	for _, q := range ps {
		if q == p {
			return true
		}
	}
	return false
}

var terrainNames = [8]string{"Open water", "Shallows", "River", "Shore", "Grassland", "Forest", "Hills", "Mountains"}
var terrainNotes = [8]string{2: "harbors and mills settle on its banks", 5: "wood: woodcutters settle at its edge",
	6: "stone and ore: quarries dig here", 7: "stone and ore: mines dig here"}

func (v *view) describe(s *scene, p mapmodel.Pt) insp {
	m, w := s.m, s.w
	if !s.in(p.X, p.Y) {
		return insp{Title: "The edge of the known world"}
	}
	vis, c := s.seen(p), s.at(p.X, p.Y)
	if in, ok := v.visitorAt(p); ok {
		return in
	}
	switch {
	case s.hasHb && p == s.harb && vis == 2:
		h := m.Harbinger
		in := insp{Title: h.Name, Command: mapmodel.CmdHarbinger, Lines: []string{h.WarningLine()}}
		if h.Numeric {
			in.Lines = append(in.Lines, "odds "+strconv.Itoa(int(float64(h.Probability*100)+0.5))+"%")
		}
		if h.Tier != "" {
			in.Lines = append(in.Lines, "threat "+h.Tier)
		}
		return in
	case has(s.idle, p):
		return insp{Title: "Idle workers", Command: m.IdleCommand(),
			Lines: []string{plural(m.Workers.Idle, "person", "people") + " with nothing to do"}}
	case has(s.hazard, p):
		return insp{Title: m.Catastrophe.PendingName, Command: mapmodel.CmdCatastrophe,
			Lines: []string{s.d.threat + " pressing on the town"}}
	case vis == 0:
		return insp{Title: "Unexplored", Command: mapmodel.CmdExpedition, Lines: []string{"send an expedition to lift the fog"}}
	case m.Expeditions.Scout != nil && (c.k == kNone || c.k == kRoad) && has(s.scout, p):
		return insp{Title: "Scouts: " + m.Expeditions.Scout.Name, Command: mapmodel.CmdExpedition,
			Lines: []string{strconv.Itoa(m.Expeditions.Scout.TicksLeft) + " ticks until they report back"}}
	}
	if vis == 2 && v.g.zoom != zRegion {
		if in, ok := v.moverAt(p); ok {
			return in
		}
	}
	t := w.At(p.X, p.Y)
	if s.city != nil && c.k != kTile && c.k != kWonder && c.k != kSite && c.k != kCentre &&
		(s.city.ov[p.Y*w.W+p.X] != 0 || !c.rail) { // a line over the railway, or no railway
		if in, ok := v.describeCity(s, p, c.k); ok {
			return in
		}
	}
	if c.rail && len(s.rail) > 0 && c.k != kTile && c.k != kWonder && c.k != kSite && c.k != kCentre {
		return insp{Title: v.railName(), Lines: []string{v.railLine()}}
	}
	switch c.k {
	case kTile:
		return v.describeTile(s, &m.Town.Tiles[c.ref])
	case kWonder:
		return describeWonder(m, &m.Town.Wonders[c.ref])
	case kCentre, kPlaza:
		in := insp{Title: "Town square of " + w.Name, Command: m.SquareCommand(),
			Lines: []string{plural(m.Workers.Pop, "person", "people") + ", " + strconv.Itoa(m.Workers.Idle) + " idle"}}
		if s.epoch == 0 {
			in.Title = "The hearth of " + w.Name
		}
		if m.AgeReady && m.NextAgeName != "" {
			in.Lines = append(in.Lines, m.NextAgeName+" is within reach")
		}
		if m.Flows.Worst != "" {
			in.Lines = append(in.Lines, m.Flows.Worst)
		}
		return in
	case kSite:
		if f := &m.Factions[c.civ-1]; f.Discovered {
			return describeCiv(m, f)
		}
	case kRoad, kStreet, kBridge:
		if c.civ > 0 && s.trails[c.civ-1] != nil {
			return describeTrail(m, &m.Factions[c.civ-1])
		}
		name := "Bridge"
		switch {
		case c.k == kStreet && c.major && s.d.wall == 3:
			name = "Ring boulevard"
		case c.k == kStreet && c.major:
			name = "Avenue"
		case c.k == kStreet:
			name = "Street"
		}
		return insp{Title: name, Lines: []string{"in " + w.Name}}
	case kWall, kTower:
		name := map[bool]string{true: "Wall tower", false: "City wall"}[c.k == kTower]
		if s.d.wall == 1 {
			name = "Palisade"
		}
		return insp{Title: name, Lines: []string{"the town is growing past it"}}
	}
	in := insp{Title: terrainNames[t]}
	if n := terrainNotes[t]; n != "" {
		in.Lines = append(in.Lines, n)
	}
	if vis == 1 {
		in.Lines = append(in.Lines, "seen once, not in sight")
	}
	return in
}

func (v *view) describeTile(s *scene, tt *mapmodel.TownTile) insp {
	m := s.m
	b := m.Building(tt.Key)
	if b == nil {
		return insp{Title: "Ruins"}
	}
	cmd := m.BuildingCommand(b)
	if cmd == "" {
		cmd = mapmodel.CmdStatus // a legacy type with nothing to do
	}
	lin := lineageLabel(tt.Lineage)
	if tt.Ruin {
		return insp{Title: "Ruins of " + b.Name, Command: cmd, Lines: []string{strconv.Itoa(b.Ruins) + " in ruins", lin}}
	}
	if b.Legacy {
		lin += ", legacy"
	}
	in := insp{Title: b.Name + " ×" + strconv.Itoa(b.Count), Command: cmd, Lines: []string{lin}}
	add := func(ok bool, s string) {
		if ok {
			in.Lines = append(in.Lines, s)
		}
	}
	staff := strconv.Itoa(b.Workers) + "/" + strconv.Itoa(b.Capacity) + " workers"
	if b.Understaffed() {
		staff += ", short of hands"
	}
	add(b.Capacity > 0, staff)
	add(b.Rate > 0 && b.RateName != "", "+"+short(b.Rate)+" "+strings.ToLower(b.RateName)+"/s")
	if d := m.Catalog.Defs[b.Upgrade]; d != nil {
		add(true, "upgrades to "+d.Name)
	} else {
		add(b.Legacy, "superseded: kept in the old town")
	}
	add(b.Delta > 0, "+"+strconv.Itoa(b.Delta)+" since your last visit")
	add(b.Delta < 0, strconv.Itoa(-b.Delta)+" fewer since your last visit")
	add(b.Ruins > 0, strconv.Itoa(b.Ruins)+" in ruins")
	return in
}

func describeWonder(m *mapmodel.Model, tw *mapmodel.TownWonder) insp {
	in := insp{Title: tw.Key, Command: m.WonderCommand(nil)}
	for i := range m.Wonders {
		wd := &m.Wonders[i]
		if wd.Key != tw.Key {
			continue
		}
		in.Title, in.Command = wd.Name, m.WonderCommand(wd)
		if wd.Age >= 0 && wd.Age < len(m.Catalog.AgeNames) {
			in.Lines = append(in.Lines, "wonder of the "+m.Catalog.AgeNames[wd.Age])
		}
		if wd.Built {
			in.Lines = append(in.Lines, "standing")
		} else {
			in.Title += " (rising)"
			in.Lines = append(in.Lines, strconv.Itoa(int(float64(wd.Progress*100)))+"% of its cost banked")
		}
		if wd.Delta {
			in.Lines = append(in.Lines, "completed since your last visit")
		}
	}
	return in
}

func describeCiv(m *mapmodel.Model, f *mapmodel.Faction) insp {
	in := insp{Title: f.Name, Command: m.FactionCommand(f),
		Lines: []string{f.Relation.String() + ", opinion " + strconv.Itoa(f.Opinion), "strength " + strconv.Itoa(f.Strength)}}
	if f.Specialty != "" {
		in.Lines[1] += ", trades in " + f.Specialty
	}
	if f.TradeCount > 0 {
		in.Lines = append(in.Lines, plural(f.TradeCount, "trade", "trades")+" so far")
	}
	if f.LentWorkers > 0 {
		in.Lines = append(in.Lines, plural(f.LentWorkers, "worker", "workers")+" lent to you")
	}
	return in
}

func describeTrail(m *mapmodel.Model, f *mapmodel.Faction) insp {
	in := insp{Title: "Trail to the " + f.Name, Command: m.FactionCommand(f)}
	var rt *mapmodel.Route
	for i := range m.Routes {
		if m.Routes[i].Civ == f.Key && rt == nil {
			rt = &m.Routes[i]
			in.Command = m.RouteCommand(rt)
		}
	}
	switch {
	case f.Relation == mapmodel.RelWar:
		in.Lines = []string{"at war: the trail is closed"}
	case rt != nil && rt.Disrupted:
		in.Lines = []string{rt.Name + " is disrupted"}
	case rt != nil && len(rt.Imports) > 0:
		in.Lines = []string{rt.Name + " caravans bring " + strings.Join(rt.Imports, ", ")}
	case rt != nil:
		in.Lines = []string{rt.Name + " caravans"}
	default:
		in.Lines = []string{"no caravans on it", f.Relation.String()}
	}
	return in
}

func plural(n int, one, many string) string {
	if n == 1 {
		return "1 " + one
	}
	return strconv.Itoa(n) + " " + many
}

// short renders a rate: 3.2, 748, 12.4k, 3.1M.
func short(f float64) string {
	if f >= 999.5e12 {
		return textfmt.Number(f) // past the T of this table: the game's own units
	}
	unit := ""
	for i, u := range [4]float64{1e12, 1e9, 1e6, 1e3} {
		if f >= u {
			f, unit = f/u, [4]string{"T", "B", "M", "k"}[i]
			break
		}
	}
	if f < 10 || unit != "" && f < 100 {
		return strconv.FormatFloat(f, 'f', 1, 64) + unit
	}
	return strconv.Itoa(int(f+0.5)) + unit
}
