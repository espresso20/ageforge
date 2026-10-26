package procmap

import (
	"github.com/espresso20/ageforge/mapgen"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/ui/mapstyle"
	"github.com/espresso20/ageforge/ui/mapstyle/roguelike"
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
		return roguelike.DescribeBuilding(m, th.Key, th.Lineage, th.Ruin)
	case mapgen.KindWonder:
		return roguelike.DescribeWonder(m, th.Key)
	case mapgen.KindAmbient:
		return mapstyle.Inspection{Title: mapmodel.LineageName(mapmodel.LinHousing)}
	case mapgen.KindRoad, mapgen.KindAgent:
		return describeRoad(m, th.Key)
	case mapgen.KindTerrain:
		if t, ok := mapgen.TerrainOf(th.Key); ok {
			return roguelike.DescribeTerrain(t)
		}
	}
	return mapstyle.Inspection{Title: m.Town.World.Name}
}

func describeRoad(m *mapmodel.Model, key string) mapstyle.Inspection {
	switch key {
	case "square":
		return roguelike.DescribeSquare(m)
	case "bridge":
		return roguelike.DescribeRoad(m, roguelike.RoadBridge)
	case "main road":
		return roguelike.DescribeRoad(m, roguelike.RoadAvenue)
	}
	return roguelike.DescribeRoad(m, roguelike.RoadStreet)
}
