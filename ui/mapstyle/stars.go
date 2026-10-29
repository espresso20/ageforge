package mapstyle

import (
	"cmp"
	"slices"
	"strconv"

	"github.com/espresso20/ageforge/mapgen"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/vibe"
)

// stars.go says what the cursor says of the territory on the star map: a
// sector of a building line, the border round the lot and a lane between hubs.
// A civilization met is described as ever (describeCiv).

// topBuildings is how many of a sector's buildings the line names.
const topBuildings = 3

// describeSector is a sector: its line, how many of its systems shine and
// the buildings it holds the most of.
func describeSector(m *mapmodel.Model, th mapgen.Thing) Inspection {
	if th.Key == "border" {
		return Inspection{Title: vibe.Line("ui.mapstyle.stars.border.01", "Your territory"),
			Lines: []string{vibe.Line("ui.mapstyle.stars.border.02", "the edge of what you hold")}}
	}
	in := Inspection{Title: vibe.Line("ui.mapstyle.stars.sector.01", "Sector: ") + placeName(m, th.Key) + " · " +
		plural(th.Count, vibe.Line("ui.mapstyle.stars.sector.02", "system"), vibe.Line("ui.mapstyle.stars.sector.03", "systems"))}
	for _, l := range m.Lineages {
		if l.Key != th.Key {
			continue
		}
		bs := slices.Clone(l.Buildings)
		slices.SortStableFunc(bs, func(a, b *mapmodel.Building) int { return cmp.Compare(b.Count, a.Count) })
		for _, b := range bs[:min(len(bs), topBuildings)] {
			if b.Count > 0 {
				in.Lines = append(in.Lines, b.Name+" ×"+strconv.Itoa(b.Count))
			}
		}
	}
	return in
}

// describeStarLane is a lane between two hubs of the star map, or from home to
// a civilization, and how busy it is.
func describeStarLane(m *mapmodel.Model, th mapgen.Thing) Inspection {
	end := func(key string) string {
		if key == "home" {
			return m.TownName
		}
		if civ := m.Faction(key); civ != nil {
			return civ.Name
		}
		return placeName(m, key)
	}
	if th.Ref == 1 { // home to a civilization: the key is the civ
		th.Key, th.Lineage = th.Lineage, th.Key
	}
	return Inspection{Title: vibe.Line("ui.mapstyle.stars.lane.01", "Lane: ") + end(th.Key) + vibe.Line("ui.mapstyle.stars.lane.02", " to ") + end(th.Lineage),
		Lines: []string{vibe.Line("ui.mapstyle.stars.lane.03", "traffic ") + strconv.Itoa(th.Count) + vibe.Line("ui.mapstyle.stars.lane.04", " of 3")}}
}
