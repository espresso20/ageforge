package roguelike

import (
	"slices"
	"strconv"
	"strings"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/ui/mapstyle"
	"github.com/espresso20/ageforge/vibe"
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
			in.Lines = append(in.Lines, vibe.Line("ui.mapstyle.roguelike.inspect.Inspect.01", "grid ref ")+pad2((cx-v.g.x)/pl.grid[0])+pad2((cy-v.g.y)/pl.grid[1]))
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

func has(ps []mapmodel.Pt, p mapmodel.Pt) bool { return slices.Contains(ps, p) }

var terrainNames = [8]vibe.Entry{vibe.E("ui.mapstyle.roguelike.inspect.terrainNames.01", "Open water"), vibe.E("ui.mapstyle.roguelike.inspect.terrainNames.02", "Shallows"), vibe.E("ui.mapstyle.roguelike.inspect.terrainNames.03", "River"), vibe.E("ui.mapstyle.roguelike.inspect.terrainNames.04", "Shore"), vibe.E("ui.mapstyle.roguelike.inspect.terrainNames.05", "Grassland"), vibe.E("ui.mapstyle.roguelike.inspect.terrainNames.06", "Forest"), vibe.E("ui.mapstyle.roguelike.inspect.terrainNames.07", "Hills"), vibe.E("ui.mapstyle.roguelike.inspect.terrainNames.08", "Mountains")}
var terrainNotes = [8]vibe.Entry{2: vibe.E("ui.mapstyle.roguelike.inspect.terrainNotes.01", "harbors and mills settle on its banks"), 5: vibe.E("ui.mapstyle.roguelike.inspect.terrainNotes.02", "wood: woodcutters settle at its edge"),
	6: vibe.E("ui.mapstyle.roguelike.inspect.terrainNotes.03", "stone and ore: quarries dig here"), 7: vibe.E("ui.mapstyle.roguelike.inspect.terrainNotes.04", "stone and ore: mines dig here")}

func (v *view) describe(s *scene, p mapmodel.Pt) insp {
	m, w := s.m, s.w
	if !s.in(p.X, p.Y) {
		return insp{Title: vibe.Line("ui.mapstyle.roguelike.inspect.describe.01", "The edge of the known world")}
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
			in.Lines = append(in.Lines, vibe.Line("ui.mapstyle.roguelike.inspect.describe.02", "odds ")+strconv.Itoa(int(float64(h.Probability*100)+0.5))+"%")
		}
		if h.Tier != "" {
			in.Lines = append(in.Lines, vibe.Line("ui.mapstyle.roguelike.inspect.describe.03", "threat ")+h.Tier)
		}
		return in
	case has(s.idle, p):
		return insp{Title: vibe.Line("ui.mapstyle.roguelike.inspect.describe.04", "Idle workers"), Command: m.IdleCommand(),
			Lines: []string{plural(m.Workers.Idle, vibe.Line("ui.mapstyle.roguelike.inspect.describe.05", "person"), vibe.Line("ui.mapstyle.roguelike.inspect.describe.06", "people")) + vibe.Line("ui.mapstyle.roguelike.inspect.describe.07", " with nothing to do")}}
	case has(s.hazard, p):
		return insp{Title: m.Catastrophe.PendingName, Command: mapmodel.CmdCatastrophe,
			Lines: []string{s.d.threat.String() + vibe.Line("ui.mapstyle.roguelike.inspect.describe.08", " pressing on the town")}}
	case vis == 0:
		return insp{Title: vibe.Line("ui.mapstyle.roguelike.inspect.describe.09", "Unexplored"), Command: mapmodel.CmdExpedition, Lines: []string{vibe.Line("ui.mapstyle.roguelike.inspect.describe.10", "send an expedition to lift the fog")}}
	case m.Expeditions.Scout != nil && (c.k == kNone || c.k == kRoad) && has(s.scout, p):
		return insp{Title: vibe.Line("ui.mapstyle.roguelike.inspect.describe.11", "Scouts: ") + m.Expeditions.Scout.Name, Command: mapmodel.CmdExpedition,
			Lines: []string{strconv.Itoa(m.Expeditions.Scout.TicksLeft) + vibe.Line("ui.mapstyle.roguelike.inspect.describe.12", " ticks until they report back")}}
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
		return describeSquare(m, w.Name, s.epoch)
	case kSite:
		if f := &m.Factions[c.civ-1]; f.Discovered {
			return describeCiv(m, f)
		}
	case kRoad, kStreet, kBridge:
		if c.civ > 0 && s.trails[c.civ-1] != nil {
			return describeTrail(m, &m.Factions[c.civ-1])
		}
		return describeRoad(w.Name, c.k, c.major, s.d.wall == 3)
	case kWall, kTower:
		name := map[bool]string{true: vibe.Line("ui.mapstyle.roguelike.inspect.describe.24", "Wall tower"), false: vibe.Line("ui.mapstyle.roguelike.inspect.describe.25", "City wall")}[c.k == kTower]
		if s.d.wall == 1 {
			name = vibe.Line("ui.mapstyle.roguelike.inspect.describe.26", "Palisade")
		}
		return insp{Title: name, Lines: []string{vibe.Line("ui.mapstyle.roguelike.inspect.describe.27", "the town is growing past it")}}
	}
	in := insp{Title: terrainNames[t].String()}
	if n := terrainNotes[t].String(); n != "" {
		in.Lines = append(in.Lines, n)
	}
	if vis == 1 {
		in.Lines = append(in.Lines, vibe.Line("ui.mapstyle.roguelike.inspect.describe.28", "seen once, not in sight"))
	}
	return in
}

func (v *view) describeTile(s *scene, tt *mapmodel.TownTile) insp {
	m := s.m
	b := m.Building(tt.Key)
	if b == nil {
		return insp{Title: vibe.Line("ui.mapstyle.roguelike.inspect.describeTile.01", "Ruins")}
	}
	cmd := m.BuildingCommand(b)
	if cmd == "" {
		cmd = mapmodel.CmdStatus // a legacy type with nothing to do
	}
	lin := lineageLabel(tt.Lineage)
	if tt.Ruin {
		return insp{Title: vibe.Line("ui.mapstyle.roguelike.inspect.describeTile.02", "Ruins of ") + b.Name, Command: cmd, Lines: []string{strconv.Itoa(b.Ruins) + vibe.Line("ui.mapstyle.roguelike.inspect.describeTile.03", " in ruins"), lin}}
	}
	if b.Legacy {
		lin += vibe.Line("ui.mapstyle.roguelike.inspect.describeTile.04", ", legacy")
	}
	in := insp{Title: b.Name + " ×" + strconv.Itoa(b.Count), Command: cmd, Lines: []string{lin}}
	add := func(ok bool, s string) {
		if ok {
			in.Lines = append(in.Lines, s)
		}
	}
	staff := strconv.Itoa(b.Workers) + "/" + strconv.Itoa(b.Capacity) + vibe.Line("ui.mapstyle.roguelike.inspect.describeTile.05", " workers")
	if b.Understaffed() {
		staff += vibe.Line("ui.mapstyle.roguelike.inspect.describeTile.06", ", short of hands")
	}
	add(b.Capacity > 0, staff)
	add(b.Rate > 0 && b.RateName != "", "+"+short(b.Rate)+" "+strings.ToLower(b.RateName)+"/s")
	if d := m.Catalog.Defs[b.Upgrade]; d != nil {
		add(true, vibe.Line("ui.mapstyle.roguelike.inspect.describeTile.07", "upgrades to ")+d.Name)
	} else {
		add(b.Legacy, vibe.Line("ui.mapstyle.roguelike.inspect.describeTile.08", "superseded: kept in the old town"))
	}
	add(b.Delta > 0, "+"+strconv.Itoa(b.Delta)+vibe.Line("ui.mapstyle.roguelike.inspect.describeTile.09", " since your last visit"))
	add(b.Delta < 0, strconv.Itoa(-b.Delta)+vibe.Line("ui.mapstyle.roguelike.inspect.describeTile.10", " fewer since your last visit"))
	add(b.Ruins > 0, strconv.Itoa(b.Ruins)+vibe.Line("ui.mapstyle.roguelike.inspect.describeTile.11", " in ruins"))
	return in
}

// describeRoad is a street, an avenue, a ring boulevard or a bridge.
func describeRoad(town string, k kind, major, ring bool) insp {
	name := vibe.Line("ui.mapstyle.roguelike.inspect.describe.19", "Bridge")
	switch {
	case k == kStreet && major && ring:
		name = vibe.Line("ui.mapstyle.roguelike.inspect.describe.20", "Ring boulevard")
	case k == kStreet && major:
		name = vibe.Line("ui.mapstyle.roguelike.inspect.describe.21", "Avenue")
	case k == kStreet:
		name = vibe.Line("ui.mapstyle.roguelike.inspect.describe.22", "Street")
	}
	return insp{Title: name, Lines: []string{vibe.Line("ui.mapstyle.roguelike.inspect.describe.23", "in ") + town}}
}

// describeSquare is the town square: its people, the next age if it is in
// reach, and what is stuck.
func describeSquare(m *mapmodel.Model, town string, epoch int) insp {
	in := insp{Title: vibe.Line("ui.mapstyle.roguelike.inspect.describe.13", "Town square of ") + town, Command: m.SquareCommand(),
		Lines: []string{plural(m.Workers.Pop, vibe.Line("ui.mapstyle.roguelike.inspect.describe.14", "person"), vibe.Line("ui.mapstyle.roguelike.inspect.describe.15", "people")) + ", " + strconv.Itoa(m.Workers.Idle) + vibe.Line("ui.mapstyle.roguelike.inspect.describe.16", " idle")}}
	if epoch == 0 {
		in.Title = vibe.Line("ui.mapstyle.roguelike.inspect.describe.17", "The hearth of ") + town
	}
	if m.AgeReady && m.NextAgeName != "" {
		in.Lines = append(in.Lines, m.NextAgeName+vibe.Line("ui.mapstyle.roguelike.inspect.describe.18", " is within reach"))
	}
	if m.Flows.Worst != "" {
		in.Lines = append(in.Lines, m.Flows.Worst)
	}
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
			in.Lines = append(in.Lines, vibe.Line("ui.mapstyle.roguelike.inspect.describeWonder.01", "wonder of the ")+m.Catalog.AgeNames[wd.Age])
		}
		if wd.Built {
			in.Lines = append(in.Lines, "standing")
		} else {
			in.Title += vibe.Line("ui.mapstyle.roguelike.inspect.describeWonder.03", " (rising)")
			in.Lines = append(in.Lines, strconv.Itoa(int(float64(wd.Progress*100)))+vibe.Line("ui.mapstyle.roguelike.inspect.describeWonder.04", "% of its cost banked"))
		}
		if wd.Delta {
			in.Lines = append(in.Lines, vibe.Line("ui.mapstyle.roguelike.inspect.describeWonder.05", "completed since your last visit"))
		}
	}
	return in
}

func describeCiv(m *mapmodel.Model, f *mapmodel.Faction) insp {
	in := insp{Title: f.Name, Command: m.FactionCommand(f),
		Lines: []string{f.Relation.String() + vibe.Line("ui.mapstyle.roguelike.inspect.describeCiv.01", ", opinion ") + strconv.Itoa(f.Opinion), vibe.Line("ui.mapstyle.roguelike.inspect.describeCiv.02", "strength ") + strconv.Itoa(f.Strength)}}
	if f.Specialty != "" {
		in.Lines[1] += vibe.Line("ui.mapstyle.roguelike.inspect.describeCiv.03", ", trades in ") + f.Specialty
	}
	if f.TradeCount > 0 {
		in.Lines = append(in.Lines, plural(f.TradeCount, vibe.Line("ui.mapstyle.roguelike.inspect.describeCiv.04", "trade"), vibe.Line("ui.mapstyle.roguelike.inspect.describeCiv.05", "trades"))+vibe.Line("ui.mapstyle.roguelike.inspect.describeCiv.06", " so far"))
	}
	if f.LentWorkers > 0 {
		in.Lines = append(in.Lines, plural(f.LentWorkers, vibe.Line("ui.mapstyle.roguelike.inspect.describeCiv.07", "worker"), vibe.Line("ui.mapstyle.roguelike.inspect.describeCiv.08", "workers"))+vibe.Line("ui.mapstyle.roguelike.inspect.describeCiv.09", " lent to you"))
	}
	return in
}

func describeTrail(m *mapmodel.Model, f *mapmodel.Faction) insp {
	in := insp{Title: vibe.Line("ui.mapstyle.roguelike.inspect.describeTrail.01", "Trail to the ") + f.Name, Command: m.FactionCommand(f)}
	var rt *mapmodel.Route
	for i := range m.Routes {
		if m.Routes[i].Civ == f.Key && rt == nil {
			rt = &m.Routes[i]
			in.Command = m.RouteCommand(rt)
		}
	}
	switch {
	case f.Relation == mapmodel.RelWar:
		in.Lines = []string{vibe.Line("ui.mapstyle.roguelike.inspect.describeTrail.02", "at war: the trail is closed")}
	case rt != nil && rt.Disrupted:
		in.Lines = []string{rt.Name + vibe.Line("ui.mapstyle.roguelike.inspect.describeTrail.03", " is disrupted")}
	case rt != nil && len(rt.Imports) > 0:
		in.Lines = []string{rt.Name + vibe.Line("ui.mapstyle.roguelike.inspect.describeTrail.04", " caravans bring ") + strings.Join(rt.Imports, ", ")}
	case rt != nil:
		in.Lines = []string{rt.Name + vibe.Line("ui.mapstyle.roguelike.inspect.describeTrail.05", " caravans")}
	default:
		in.Lines = []string{vibe.Line("ui.mapstyle.roguelike.inspect.describeTrail.06", "no caravans on it"), f.Relation.String()}
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
	if f >= 999.5*trillion {
		return textfmt.Number(f) // past the T of this table: the game's own units
	}
	unit := ""
	for i, u := range [4]float64{trillion, billion, million, 1000} {
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
