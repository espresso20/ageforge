package roguelike

import (
	"strconv"
	"strings"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/ui/mapstyle"
)

// space_inspect.go says what the sky cursor is on, in the game's own words,
// with the command a player would type for it (from the model's helpers).

// Inspect reports what the cursor is on; ok is false while it is hidden.
func (v *skyView) Inspect(f mapstyle.Frame) (mapstyle.Inspection, bool) {
	s := v.sceneFor(f.Model)
	if s == nil || !v.g.inspect {
		return mapstyle.Inspection{}, false
	}
	v.anim, v.tier = f.Anim, f.Tier
	return v.describe(s, v.cur), true
}

// skyVoidNames are what open space is called in each scene.
var skyVoidNames = [mapmodel.NumSkyScenes][2]string{
	{}, {"Open space", "low orbit, quiet tonight"}, {"Deep space", "light-years of nothing much"},
	{"Between the stars", "the lanes run elsewhere"}, {"A maybe", "nothing is certain here"},
	{"Light", "everything you built, remembered"},
}

func (v *skyView) describe(s *skyScene, p mapmodel.Pt) mapstyle.Inspection {
	m := s.m
	if in, ok := v.skyVisitorAt(p); ok {
		return in
	}
	if in, ok := v.moverAt(p); ok {
		return in
	}
	if s.sky == mapmodel.SkyDeep {
		for k := range deepOrbits {
			if v.orreryPlanet(k) == p {
				if k == 2 {
					return mapstyle.Inspection{Title: "The homeworld", Command: m.SquareCommand(),
						Lines: []string{"where it all began", plural(m.Workers.Pop, "person", "people") + " in all"}}
				}
				return mapstyle.Inspection{Title: "A planet of the home system", Lines: []string{"it keeps its old orbit"}}
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
					in.Lines = append(in.Lines, plural(l.Count, "building", "buildings")+" of "+lineageLabel(fr.lin))
				}
			}
			if c.k == skSlot {
				in.Lines = append(in.Lines, "room to build")
			}
			return in
		}
	case skTether:
		return mapstyle.Inspection{Title: "Space elevator", Lines: []string{"the tether from " + world + " to the ring",
			"climbers ride it day and night"}}
	case skPlanet, skLimb:
		return mapstyle.Inspection{Title: world + ", from orbit", Command: m.SquareCommand(),
			Lines: []string{plural(m.Workers.Pop, "person", "people") + " below", "night side"}}
	case skCity:
		return mapstyle.Inspection{Title: "City lights of " + world, Command: m.SquareCommand(),
			Lines: []string{plural(m.TotalBuildings(), "building", "buildings") + ", " + strconv.Itoa(m.Workers.Idle) + " idle"}}
	case skRock:
		return mapstyle.Inspection{Title: "Asteroid belt", Lines: []string{"ore for the taking"}}
	case skMoon:
		return mapstyle.Inspection{Title: "The moon", Lines: []string{"it used to be out of reach"}}
	case skSun:
		return mapstyle.Inspection{Title: "The home sun", Lines: []string{"the orrery of the home system", world + " still turns round it"}}
	case skOrbit:
		return mapstyle.Inspection{Title: "The home system", Lines: []string{"a world on every ring"}}
	case skColony:
		return mapstyle.Inspection{Title: "Colony world", Lines: []string{"a world of your own, far out"}}
	case skBeacon:
		return mapstyle.Inspection{Title: "Colony beacon", Lines: []string{"a world calling home"}}
	case skField:
		return mapstyle.Inspection{Title: "Warp gate field", Lines: []string{"open: the stars are a step away"}}
	case skNebula:
		return mapstyle.Inspection{Title: "Nebula", Lines: []string{"the first color out here"}}
	case skSpiral:
		return mapstyle.Inspection{Title: "The galaxy", Lines: []string{"its arms turn slower than history"}}
	case skCloud:
		return mapstyle.Inspection{Title: "Probability cloud", Lines: []string{"where the station might be"}}
	case skEcho:
		if int(c.ref) < len(m.Town.Tiles) {
			tt := &m.Town.Tiles[c.ref]
			name := tt.Key
			if b := m.Building(tt.Key); b != nil {
				name = b.Name
			}
			return mapstyle.Inspection{Title: "Echo of the old town", Lines: []string{name + ", as it stood on the ground"}}
		}
	case skRing:
		if e := int(c.ref); e >= 0 && e < len(m.Catalog.EpochName) {
			return mapstyle.Inspection{Title: "Ring of the " + m.Catalog.EpochName[e],
				Lines: []string{plural(len(s.marks[e]), "kind of building", "kinds of building") + " you raised"}}
		}
	}
	in := mapstyle.Inspection{Title: skyVoidNames[s.sky][0], Lines: []string{skyVoidNames[s.sky][1]}}
	if c.k == skStar {
		in.Title = "A star"
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
		return mapstyle.Inspection{Title: "Ruins", Lines: []string{part}}
	}
	cmd := m.BuildingCommand(b)
	if cmd == "" {
		cmd = mapmodel.CmdStatus
	}
	if tt.Ruin {
		return mapstyle.Inspection{Title: "Ruins of " + b.Name, Command: cmd,
			Lines: []string{strconv.Itoa(b.Ruins) + " in ruins", part}}
	}
	if b.Legacy {
		part += ", an older section"
	}
	in := mapstyle.Inspection{Title: b.Name + " ×" + strconv.Itoa(b.Count), Command: cmd, Lines: []string{part}}
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
	}
	add(b.Delta > 0, "+"+strconv.Itoa(b.Delta)+" since your last visit")
	add(b.Delta < 0, strconv.Itoa(-b.Delta)+" fewer since your last visit")
	add(b.Ruins > 0, strconv.Itoa(b.Ruins)+" in ruins")
	return in
}

// describeSkyWonder is a wonder's star.
func describeSkyWonder(m *mapmodel.Model, w *mapmodel.Wonder) mapstyle.Inspection {
	in := mapstyle.Inspection{Title: w.Name, Command: m.WonderCommand(w)}
	if w.Age >= 0 && w.Age < len(m.Catalog.AgeNames) {
		in.Lines = append(in.Lines, "wonder of the "+m.Catalog.AgeNames[w.Age])
	}
	if w.Built {
		in.Lines = append(in.Lines, "it shines among the stars")
	} else {
		in.Title += " (rising)"
		in.Lines = append(in.Lines, strconv.Itoa(int(w.Progress*100))+"% of its cost banked")
	}
	if w.Delta {
		in.Lines = append(in.Lines, "completed since your last visit")
	}
	return in
}

// describeLane is a trade lane to a civ's star system.
func describeLane(m *mapmodel.Model, f *mapmodel.Faction) mapstyle.Inspection {
	in := mapstyle.Inspection{Title: "Lane to the " + f.Name, Command: m.FactionCommand(f)}
	var rt *mapmodel.Route
	for i := range m.Routes {
		if m.Routes[i].Civ == f.Key && rt == nil {
			rt = &m.Routes[i]
			in.Command = m.RouteCommand(rt)
		}
	}
	switch {
	case f.Relation == mapmodel.RelWar:
		in.Lines = []string{"at war: the lane is closed"}
	case rt != nil && rt.Disrupted:
		in.Lines = []string{rt.Name + " is disrupted"}
	case rt != nil && len(rt.Imports) > 0:
		in.Lines = []string{rt.Name + " brings " + strings.Join(rt.Imports, ", ")}
	case rt != nil:
		in.Lines = []string{rt.Name}
	default:
		in.Lines = []string{"no freight on it", f.Relation.String()}
	}
	return in
}

// describeHub is the scene's heart: what the town square is on the ground.
func (v *skyView) describeHub(s *skyScene, world string) mapstyle.Inspection {
	m := s.m
	titles := [mapmodel.NumSkyScenes]string{"", "Ring station of " + world, "Home system of " + world,
		"Starbase " + world, "Anchor of " + world, "Core of " + world}
	in := mapstyle.Inspection{Title: titles[s.sky], Command: m.SquareCommand(),
		Lines: []string{plural(m.Workers.Pop, "person", "people") + ", " + strconv.Itoa(m.Workers.Idle) + " idle"}}
	if m.AgeReady && m.NextAgeName != "" {
		in.Lines = append(in.Lines, m.NextAgeName+" is within reach")
	}
	if m.Flows.Worst != "" {
		in.Lines = append(in.Lines, m.Flows.Worst)
	}
	return in
}
