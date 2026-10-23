package roguelike

import (
	"strconv"
	"strings"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/ui/mapstyle"
	"github.com/espresso20/ageforge/vibe"
)

// space_inspect.go says what the sky cursor is on, in the game's own words,
// with the command a player would type for it (from the model's helpers).

// Inspect reports what the cursor is on; ok is false while it is hidden.
func (v *skyView) Inspect(f mapstyle.Frame) (mapstyle.Inspection, bool) {
	s := v.sceneFor(f.Model)
	if s == nil || !v.g.inspect {
		return mapstyle.Inspection{}, false
	}
	v.anim, v.visit, v.clock, v.tier = f.Anim, f.VisitFrame(), f.Clock, f.Tier
	return v.describe(s, v.cur), true
}

// skyVoidNames are what open space is called in each scene.
var skyVoidNames = [mapmodel.NumSkyScenes][2]vibe.Entry{
	{}, {vibe.E("ui.mapstyle.roguelike.space_inspect.skyVoidNames.01", "Open space"), vibe.E("ui.mapstyle.roguelike.space_inspect.skyVoidNames.02", "low orbit, quiet tonight")}, {vibe.E("ui.mapstyle.roguelike.space_inspect.skyVoidNames.03", "Deep space"), vibe.E("ui.mapstyle.roguelike.space_inspect.skyVoidNames.04", "light-years of nothing much")},
	{vibe.E("ui.mapstyle.roguelike.space_inspect.skyVoidNames.05", "Between the stars"), vibe.E("ui.mapstyle.roguelike.space_inspect.skyVoidNames.06", "the lanes run elsewhere")}, {vibe.E("ui.mapstyle.roguelike.space_inspect.skyVoidNames.07", "A maybe"), vibe.E("ui.mapstyle.roguelike.space_inspect.skyVoidNames.08", "nothing is certain here")},
	{vibe.E("ui.mapstyle.roguelike.space_inspect.skyVoidNames.09", "Light"), vibe.E("ui.mapstyle.roguelike.space_inspect.skyVoidNames.10", "everything you built, remembered")},
}

func (v *skyView) describe(s *skyScene, p mapmodel.Pt) mapstyle.Inspection {
	m := s.m
	if in, ok := v.skyVisitorAt(p); ok {
		return in
	}
	if in, ok := v.moverAt(p); ok {
		return in
	}
	if s.sky == mapmodel.SkyOrbit {
		for k := 0; k < orbDebris; k++ {
			if q, ok := v.debrisAt(k); ok && q == p {
				return mapstyle.Inspection{Title: vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.01", "Debris"), Lines: []string{vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.02", "spent stages and old satellites, still circling")},
					Kind: mapstyle.KindMover}
			}
		}
	}
	if s.sky == mapmodel.SkyDeep {
		for k := range deepOrbits {
			if v.orreryPlanet(k) == p {
				if k == 2 {
					return mapstyle.Inspection{Title: vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.03", "The homeworld"), Command: m.SquareCommand(),
						Lines: []string{vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.04", "where it all began"), plural(m.Workers.Pop, vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.05", "person"), vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.06", "people")) + vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.07", " in all")}}
				}
				return mapstyle.Inspection{Title: vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.08", "A planet of the home system"), Lines: []string{vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.09", "it keeps its old orbit")}}
			}
		}
	}
	c := s.at(p.X, p.Y)
	world := ""
	if m.Town.World != nil {
		world = m.Town.World.Name
	}
	switch c.k {
	case skUnit:
		return describeUnit(s, &m.Town.Tiles[c.ref])
	case skMark, skCore:
		if in, ok := v.describeMark(s, c, p); ok {
			return in
		}
	case skWonder:
		if int(c.ref) < len(m.Wonders) {
			return describeSkyWonder(m, &m.Wonders[c.ref])
		}
	case skCiv:
		if f := &m.Factions[c.ref]; f.Discovered {
			return describeCiv(m, f)
		}
	case skLane:
		if f := &m.Factions[c.ref]; f.Discovered {
			return describeLane(m, f)
		}
	case skHub:
		return v.describeHub(s, world)
	case skFrame, skSlot:
		if int(c.ref) < len(s.b.frames) {
			fr := s.b.frames[c.ref]
			in := mapstyle.Inspection{Title: fr.name, Lines: []string{fr.line}}
			if fr.lin != "" {
				if l := m.Lineage(fr.lin); l != nil {
					in.Lines = append(in.Lines, plural(l.Count, vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.10", "building"), vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.11", "buildings"))+vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.12", " of ")+lineageLabel(fr.lin))
				}
			}
			if c.k == skSlot {
				in.Lines = append(in.Lines, vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.13", "room to build"))
			}
			return in
		}
	case skTether:
		return mapstyle.Inspection{Title: vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.14", "Space elevator"), Lines: []string{vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.15", "the tether from ") + world + vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.16", " to the ring"),
			vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.17", "climbers ride it day and night")}}
	case skPlanet, skLimb:
		return mapstyle.Inspection{Title: world + vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.18", ", from orbit"), Command: m.SquareCommand(),
			Lines: []string{plural(m.Workers.Pop, vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.19", "person"), vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.20", "people")) + vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.21", " below"), vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.22", "night side")}}
	case skCity:
		return mapstyle.Inspection{Title: vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.23", "City lights of ") + world, Command: m.SquareCommand(),
			Lines: []string{plural(m.TotalBuildings(), vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.24", "building"), vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.25", "buildings")) + ", " + strconv.Itoa(m.Workers.Idle) + vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.26", " idle")}}
	case skRock:
		return mapstyle.Inspection{Title: vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.27", "Asteroid belt"), Lines: []string{vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.28", "ore for the taking")}}
	case skMoon:
		return mapstyle.Inspection{Title: vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.29", "The moon"), Lines: []string{vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.30", "it used to be out of reach")}}
	case skSun:
		return mapstyle.Inspection{Title: vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.31", "The home sun"), Lines: []string{vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.32", "the orrery of the home system"), world + vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.33", " still turns round it")}}
	case skOrbit:
		return mapstyle.Inspection{Title: vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.34", "The home system"), Lines: []string{vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.35", "a world on every ring")}}
	case skColony:
		return mapstyle.Inspection{Title: vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.36", "Colony world"), Lines: []string{vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.37", "a world of your own, far out")}}
	case skBeacon:
		return mapstyle.Inspection{Title: vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.38", "Colony beacon"), Lines: []string{vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.39", "a world calling home")}}
	case skField:
		return mapstyle.Inspection{Title: vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.40", "Warp gate field"), Lines: []string{vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.41", "open: the stars are a step away")}}
	case skNebula:
		return mapstyle.Inspection{Title: vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.42", "Nebula"), Lines: []string{vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.43", "the first color out here")}}
	case skSpiral:
		return mapstyle.Inspection{Title: vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.44", "The galaxy"), Lines: []string{vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.45", "its arms turn slower than history")}}
	case skCloud:
		return mapstyle.Inspection{Title: vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.46", "Probability cloud"), Lines: []string{vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.47", "where the station might be")}}
	case skEcho:
		if int(c.ref) < len(m.Town.Tiles) {
			tt := &m.Town.Tiles[c.ref]
			name := tt.Key
			if b := m.Building(tt.Key); b != nil {
				name = b.Name
			}
			return mapstyle.Inspection{Title: vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.48", "Echo of the old town"), Lines: []string{name + vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.49", ", as it stood on the ground")}}
		}
	case skRing:
		if k := int(c.ref); k >= 0 && k < len(s.rings) {
			return describeRing(s, k)
		}
	}
	in := mapstyle.Inspection{Title: skyVoidNames[s.sky][0].String(), Lines: []string{skyVoidNames[s.sky][1].String()}}
	if c.k == skStar {
		in.Title = vibe.Line("ui.mapstyle.roguelike.space_inspect.describe.50", "A star")
	}
	return in
}

// describeUnit is a building's unit: the type and its count, what it is up
// here, its workers and output, and the command for it.
func describeUnit(s *skyScene, tt *mapmodel.TownTile) mapstyle.Inspection {
	m := s.m
	b := m.Building(tt.Key)
	part := mapmodel.SkyPartOf(s.sky, tt.Lineage).Name
	if b == nil {
		return mapstyle.Inspection{Title: vibe.Line("ui.mapstyle.roguelike.space_inspect.describeUnit.01", "Ruins"), Lines: []string{part}}
	}
	cmd := m.BuildingCommand(b)
	if cmd == "" {
		cmd = mapmodel.CmdStatus
	}
	if tt.Ruin {
		return mapstyle.Inspection{Title: vibe.Line("ui.mapstyle.roguelike.space_inspect.describeUnit.02", "Ruins of ") + b.Name, Command: cmd,
			Lines: []string{strconv.Itoa(b.Ruins) + vibe.Line("ui.mapstyle.roguelike.space_inspect.describeUnit.03", " in ruins"), part}}
	}
	if b.Legacy {
		part += vibe.Line("ui.mapstyle.roguelike.space_inspect.describeUnit.04", ", an older section")
	}
	in := mapstyle.Inspection{Title: b.Name + " ×" + strconv.Itoa(b.Count), Command: cmd, Lines: []string{part}}
	add := func(ok bool, s string) {
		if ok {
			in.Lines = append(in.Lines, s)
		}
	}
	staff := strconv.Itoa(b.Workers) + "/" + strconv.Itoa(b.Capacity) + vibe.Line("ui.mapstyle.roguelike.space_inspect.describeUnit.05", " workers")
	if b.Understaffed() {
		staff += vibe.Line("ui.mapstyle.roguelike.space_inspect.describeUnit.06", ", short of hands")
	}
	add(b.Capacity > 0, staff)
	add(b.Rate > 0 && b.RateName != "", "+"+short(b.Rate)+" "+strings.ToLower(b.RateName)+"/s")
	if d := m.Catalog.Defs[b.Upgrade]; d != nil {
		add(true, vibe.Line("ui.mapstyle.roguelike.space_inspect.describeUnit.07", "upgrades to ")+d.Name)
	}
	add(b.Delta > 0, "+"+strconv.Itoa(b.Delta)+vibe.Line("ui.mapstyle.roguelike.space_inspect.describeUnit.08", " since your last visit"))
	add(b.Delta < 0, strconv.Itoa(-b.Delta)+vibe.Line("ui.mapstyle.roguelike.space_inspect.describeUnit.09", " fewer since your last visit"))
	add(b.Ruins > 0, strconv.Itoa(b.Ruins)+vibe.Line("ui.mapstyle.roguelike.space_inspect.describeUnit.10", " in ruins"))
	return in
}

// describeSkyWonder is a wonder's star.
func describeSkyWonder(m *mapmodel.Model, w *mapmodel.Wonder) mapstyle.Inspection {
	in := mapstyle.Inspection{Title: w.Name, Command: m.WonderCommand(w)}
	if w.Age >= 0 && w.Age < len(m.Catalog.AgeNames) {
		in.Lines = append(in.Lines, vibe.Line("ui.mapstyle.roguelike.space_inspect.describeSkyWonder.01", "wonder of the ")+m.Catalog.AgeNames[w.Age])
	}
	if w.Built {
		in.Lines = append(in.Lines, vibe.Line("ui.mapstyle.roguelike.space_inspect.describeSkyWonder.02", "it shines among the stars"))
	} else {
		in.Title += vibe.Line("ui.mapstyle.roguelike.space_inspect.describeSkyWonder.03", " (rising)")
		in.Lines = append(in.Lines, strconv.Itoa(int(w.Progress*100))+vibe.Line("ui.mapstyle.roguelike.space_inspect.describeSkyWonder.04", "% of its cost banked"))
	}
	if w.Delta {
		in.Lines = append(in.Lines, vibe.Line("ui.mapstyle.roguelike.space_inspect.describeSkyWonder.05", "completed since your last visit"))
	}
	return in
}

// describeLane is a trade lane to a civ's star system.
func describeLane(m *mapmodel.Model, f *mapmodel.Faction) mapstyle.Inspection {
	in := mapstyle.Inspection{Title: vibe.Line("ui.mapstyle.roguelike.space_inspect.describeLane.01", "Lane to the ") + f.Name, Command: m.FactionCommand(f)}
	var rt *mapmodel.Route
	for i := range m.Routes {
		if m.Routes[i].Civ == f.Key && rt == nil {
			rt = &m.Routes[i]
			in.Command = m.RouteCommand(rt)
		}
	}
	switch {
	case f.Relation == mapmodel.RelWar:
		in.Lines = []string{vibe.Line("ui.mapstyle.roguelike.space_inspect.describeLane.02", "at war: the lane is closed")}
	case rt != nil && rt.Disrupted:
		in.Lines = []string{rt.Name + vibe.Line("ui.mapstyle.roguelike.space_inspect.describeLane.03", " is disrupted")}
	case rt != nil && len(rt.Imports) > 0:
		in.Lines = []string{rt.Name + vibe.Line("ui.mapstyle.roguelike.space_inspect.describeLane.04", " brings ") + strings.Join(rt.Imports, ", ")}
	case rt != nil:
		in.Lines = []string{rt.Name}
	default:
		in.Lines = []string{vibe.Line("ui.mapstyle.roguelike.space_inspect.describeLane.05", "no freight on it"), f.Relation.String()}
	}
	return in
}

// describeHub is the scene's heart: what the town square is on the ground.
func (v *skyView) describeHub(s *skyScene, world string) mapstyle.Inspection {
	m := s.m
	titles := [mapmodel.NumSkyScenes]string{"", vibe.Line("ui.mapstyle.roguelike.space_inspect.describeHub.01", "Ring station of ") + world, vibe.Line("ui.mapstyle.roguelike.space_inspect.describeHub.02", "Home system of ") + world,
		vibe.Line("ui.mapstyle.roguelike.space_inspect.describeHub.03", "Starbase ") + world, vibe.Line("ui.mapstyle.roguelike.space_inspect.describeHub.04", "Anchor of ") + world, vibe.Line("ui.mapstyle.roguelike.space_inspect.describeHub.05", "Core of ") + world}
	in := mapstyle.Inspection{Title: titles[s.sky], Command: m.SquareCommand(),
		Lines: []string{plural(m.Workers.Pop, vibe.Line("ui.mapstyle.roguelike.space_inspect.describeHub.06", "person"), vibe.Line("ui.mapstyle.roguelike.space_inspect.describeHub.07", "people")) + ", " + strconv.Itoa(m.Workers.Idle) + vibe.Line("ui.mapstyle.roguelike.space_inspect.describeHub.08", " idle")}}
	if m.AgeReady && m.NextAgeName != "" {
		in.Lines = append(in.Lines, m.NextAgeName+vibe.Line("ui.mapstyle.roguelike.space_inspect.describeHub.09", " is within reach"))
	}
	if m.Flows.Worst != "" {
		in.Lines = append(in.Lines, m.Flows.Worst)
	}
	return in
}
