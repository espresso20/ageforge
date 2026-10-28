package procmap

import (
	"strconv"
	"strings"

	"github.com/espresso20/ageforge/mapgen"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/ui/mapstyle"
	"github.com/espresso20/ageforge/vibe"
)

// Inspect reports what the cursor is on; ok is false while it is hidden. It
// says only what the game already says: a building's name and count, a
// wonder's name and age, the square, a road, the ground.
func (v *view) Inspect(f mapstyle.Frame) (mapstyle.Inspection, bool) {
	sc := v.scene(f)
	if sc == nil || !v.inspect {
		return mapstyle.Inspection{}, false
	}
	p := v.cur
	if v.overview && v.ov != nil { // the most important thing in the cell
		if x, y, ok := v.ov.CellOf(p); ok {
			p = v.ov.Pick(x, y)
		}
	}
	return describe(f.Model, sc.Grid.ThingAt(p.X, p.Y)), true
}

// describe is what the game says of a thing the engine drew.
func describe(m *mapmodel.Model, th mapgen.Thing) mapstyle.Inspection {
	switch th.Kind {
	case mapgen.KindBuilding:
		return describeBuilding(m, th.Key, th.Lineage, th.Ruin)
	case mapgen.KindWonder:
		return describeWonder(m, th.Key)
	case mapgen.KindAmbient:
		return mapstyle.Inspection{Title: mapmodel.LineageName(mapmodel.LinHousing)}
	case mapgen.KindRoad, mapgen.KindAgent:
		return describeRoad(m, th.Key)
	case mapgen.KindTerrain:
		if t, ok := mapgen.TerrainOf(th.Key); ok {
			return describeTerrain(t)
		}
	}
	return mapstyle.Inspection{Title: m.TownName}
}

// describeBuilding says what the cursor is on when it is on a building type
// (or its ruins): its name, the count, the line, and what to type for it.
func describeBuilding(m *mapmodel.Model, key, lineage string, ruin bool) mapstyle.Inspection {
	b := m.Building(key)
	if b == nil {
		return mapstyle.Inspection{Title: vibe.Line("ui.mapstyle.roguelike.inspect.describeTile.01", "Ruins")}
	}
	command := m.BuildingCommand(b)
	if command == "" {
		command = mapmodel.CmdStatus // a legacy type with nothing to do
	}
	line := lineageLabel(lineage)
	if ruin {
		return mapstyle.Inspection{
			Title:   vibe.Line("ui.mapstyle.roguelike.inspect.describeTile.02", "Ruins of ") + b.Name,
			Command: command,
			Lines:   []string{strconv.Itoa(b.Ruins) + vibe.Line("ui.mapstyle.roguelike.inspect.describeTile.03", " in ruins"), line},
		}
	}
	if b.Legacy {
		line += vibe.Line("ui.mapstyle.roguelike.inspect.describeTile.04", ", legacy")
	}
	in := mapstyle.Inspection{Title: b.Name + " ×" + strconv.Itoa(b.Count), Command: command, Lines: []string{line}}
	add := func(ok bool, text string) {
		if ok {
			in.Lines = append(in.Lines, text)
		}
	}
	staff := strconv.Itoa(b.Workers) + "/" + strconv.Itoa(b.Capacity) + vibe.Line("ui.mapstyle.roguelike.inspect.describeTile.05", " workers")
	if b.Understaffed() {
		staff += vibe.Line("ui.mapstyle.roguelike.inspect.describeTile.06", ", short of hands")
	}
	add(b.Capacity > 0, staff)
	add(b.Rate > 0 && b.RateName != "", "+"+shortRate(b.Rate)+" "+strings.ToLower(b.RateName)+"/s")
	if upgrade := m.Catalog.Defs[b.Upgrade]; upgrade != nil {
		add(true, vibe.Line("ui.mapstyle.roguelike.inspect.describeTile.07", "upgrades to ")+upgrade.Name)
	} else {
		add(b.Legacy, vibe.Line("ui.mapstyle.roguelike.inspect.describeTile.08", "superseded: kept in the old town"))
	}
	add(b.Delta > 0, "+"+strconv.Itoa(b.Delta)+vibe.Line("ui.mapstyle.roguelike.inspect.describeTile.09", " since your last visit"))
	add(b.Delta < 0, strconv.Itoa(-b.Delta)+vibe.Line("ui.mapstyle.roguelike.inspect.describeTile.10", " fewer since your last visit"))
	add(b.Ruins > 0, strconv.Itoa(b.Ruins)+vibe.Line("ui.mapstyle.roguelike.inspect.describeTile.11", " in ruins"))
	return in
}

// describeWonder says what the cursor is on when it is on a wonder.
func describeWonder(m *mapmodel.Model, key string) mapstyle.Inspection {
	in := mapstyle.Inspection{Title: key, Command: m.WonderCommand(nil)}
	for i := range m.Wonders {
		wonder := &m.Wonders[i]
		if wonder.Key != key {
			continue
		}
		in.Title, in.Command = wonder.Name, m.WonderCommand(wonder)
		if wonder.Age >= 0 && wonder.Age < len(m.Catalog.AgeNames) {
			in.Lines = append(in.Lines, vibe.Line("ui.mapstyle.roguelike.inspect.describeWonder.01", "wonder of the ")+m.Catalog.AgeNames[wonder.Age])
		}
		if wonder.Built {
			in.Lines = append(in.Lines, "standing")
		} else {
			in.Title += vibe.Line("ui.mapstyle.roguelike.inspect.describeWonder.03", " (rising)")
			in.Lines = append(in.Lines, strconv.Itoa(int(float64(wonder.Progress*100)))+vibe.Line("ui.mapstyle.roguelike.inspect.describeWonder.04", "% of its cost banked"))
		}
		if wonder.Delta {
			in.Lines = append(in.Lines, vibe.Line("ui.mapstyle.roguelike.inspect.describeWonder.05", "completed since your last visit"))
		}
	}
	return in
}

// describeRoad is what the cursor says on a road, a bridge or the square.
func describeRoad(m *mapmodel.Model, key string) mapstyle.Inspection {
	town := m.TownName
	switch key {
	case "square":
		return describeSquare(m, town)
	case "bridge":
		return roadIn(vibe.Line("ui.mapstyle.roguelike.inspect.describe.19", "Bridge"), town)
	case "main road":
		return roadIn(vibe.Line("ui.mapstyle.roguelike.inspect.describe.21", "Avenue"), town)
	}
	return roadIn(vibe.Line("ui.mapstyle.roguelike.inspect.describe.22", "Street"), town)
}

func roadIn(name, town string) mapstyle.Inspection {
	return mapstyle.Inspection{Title: name, Lines: []string{vibe.Line("ui.mapstyle.roguelike.inspect.describe.23", "in ") + town}}
}

// describeSquare is the town square: its people, the next age if it is in
// reach, and what is stuck.
func describeSquare(m *mapmodel.Model, town string) mapstyle.Inspection {
	people := plural(m.Workers.Pop, vibe.Line("ui.mapstyle.roguelike.inspect.describe.14", "person"), vibe.Line("ui.mapstyle.roguelike.inspect.describe.15", "people"))
	in := mapstyle.Inspection{
		Title:   vibe.Line("ui.mapstyle.roguelike.inspect.describe.13", "Town square of ") + town,
		Command: m.SquareCommand(),
		Lines:   []string{people + ", " + strconv.Itoa(m.Workers.Idle) + vibe.Line("ui.mapstyle.roguelike.inspect.describe.16", " idle")},
	}
	if m.Epoch == 0 {
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

var terrainNames = [8]vibe.Entry{
	vibe.E("ui.mapstyle.roguelike.inspect.terrainNames.01", "Open water"),
	vibe.E("ui.mapstyle.roguelike.inspect.terrainNames.02", "Shallows"),
	vibe.E("ui.mapstyle.roguelike.inspect.terrainNames.03", "River"),
	vibe.E("ui.mapstyle.roguelike.inspect.terrainNames.04", "Shore"),
	vibe.E("ui.mapstyle.roguelike.inspect.terrainNames.05", "Grassland"),
	vibe.E("ui.mapstyle.roguelike.inspect.terrainNames.06", "Forest"),
	vibe.E("ui.mapstyle.roguelike.inspect.terrainNames.07", "Hills"),
	vibe.E("ui.mapstyle.roguelike.inspect.terrainNames.08", "Mountains"),
}

var terrainNotes = [8]vibe.Entry{
	2: vibe.E("ui.mapstyle.roguelike.inspect.terrainNotes.01", "harbors and mills settle on its banks"),
	5: vibe.E("ui.mapstyle.roguelike.inspect.terrainNotes.02", "wood: woodcutters settle at its edge"),
	6: vibe.E("ui.mapstyle.roguelike.inspect.terrainNotes.03", "stone and ore: quarries dig here"),
	7: vibe.E("ui.mapstyle.roguelike.inspect.terrainNotes.04", "stone and ore: mines dig here"),
}

// describeTerrain says what the cursor is on when it is on open ground of a
// terrain.
func describeTerrain(t mapmodel.Terrain) mapstyle.Inspection {
	in := mapstyle.Inspection{Title: terrainNames[t].String()}
	if note := terrainNotes[t].String(); note != "" {
		in.Lines = append(in.Lines, note)
	}
	return in
}

// rateUnits are the suffixes of a rate, largest first.
var rateUnits = []struct {
	size   float64
	suffix string
}{{1e12, "T"}, {1e9, "B"}, {1e6, "M"}, {1e3, "k"}}

// shortRate renders a rate: 3.2, 748, 12.4k, 3.1M.
func shortRate(f float64) string {
	if f >= 999.5e12 {
		return textfmt.Number(f) // past the T of this table: the game's own units
	}
	suffix := ""
	for _, unit := range rateUnits {
		if f >= unit.size {
			f, suffix = f/unit.size, unit.suffix
			break
		}
	}
	if f < 10 || suffix != "" && f < 100 {
		return strconv.FormatFloat(f, 'f', 1, 64) + suffix
	}
	return strconv.Itoa(int(f+0.5)) + suffix
}
