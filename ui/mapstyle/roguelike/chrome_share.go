package roguelike

import (
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/ui/mapstyle"
	"github.com/espresso20/ageforge/vibe"
)

// chrome_share.go lends the map's words to a picture another renderer draws:
// the header, the news line, the inspector and status line, the flows panel,
// and what the cursor says of a building, a wonder, the square, a road or the
// ground. Every string stays where it is; a second renderer reads them
// through here, so the game has one set of map words.

// Chrome is the frame round a map picture.
type Chrome struct{ v *view }

// NewChrome makes a chrome with its options at their defaults.
func NewChrome() *Chrome { return &Chrome{v: newView()} }

// Begin readies the chrome for a frame and returns the palette it draws in,
// or nil when there is no model.
func (c *Chrome) Begin(f mapstyle.Frame) *mapstyle.Palette {
	if f.Model == nil {
		return nil
	}
	v := c.v
	v.anim, v.visit, v.clock, v.tier = f.Anim, f.VisitFrame(), f.Clock, f.Tier
	v.sc = &scene{m: f.Model, w: f.Model.Town.World, epoch: f.Model.Epoch}
	v.palette(f.Model.Epoch, -1)
	return v.pal.Palette
}

// NoMap says there is no map yet, on a canvas, for a frame with no model.
func (c *Chrome) NoMap(cv *mapstyle.Canvas) {
	c.v.pal = &pal{Palette: mapstyle.NewPalette(0)}
	c.v.noMap(cv)
}

// SetOption sets a shared option (the flows overlay, the changes news).
func (c *Chrome) SetOption(o mapstyle.Option, on bool) { c.v.SetOption(o, on) }

// Flows reports whether the flows overlay is on.
func (c *Chrome) Flows() bool { return c.v.flows }

// Header draws the header row with the zoom's name.
func (c *Chrome) Header(cv *mapstyle.Canvas, w int, zoom string) { c.v.headerWith(cv, w, zoom) }

// Bottom draws the news, the inspector's two lines and the status line.
func (c *Chrome) Bottom(cv *mapstyle.Canvas, w, h int, in mapstyle.Inspection, ok bool, zoom string) {
	c.v.bottomWith(cv, w, h, in, ok, zoom)
}

// Compact draws the first and last lines of the compact view.
func (c *Chrome) Compact(cv *mapstyle.Canvas, r mapstyle.Rect) { c.v.compactBars(cv, r) }

// Sidebar draws the rule and the flows panel (when on) of a sidebar w wide,
// and returns the first row below it, where a legend may go.
func (c *Chrome) Sidebar(cv *mapstyle.Canvas, x, y, w, h int) int {
	return c.v.flowsPanel(cv, x, y, w, h)
}

// Box draws the flows panel in a frame over the picture, for a screen with no
// room for a sidebar.
func (c *Chrome) Box(cv *mapstyle.Canvas, x, y, w, maxH int) { c.v.box(cv, x, y, w, maxH) }

// LegendTitle is the legend's heading.
func LegendTitle() string { return vibe.Line("ui.mapstyle.roguelike.chrome.drawLegend.01", "LEGEND") }

// SideWidth is the width of the sidebar.
const SideWidth = sideW

// DescribeBuilding says what the cursor is on when it is on a building type
// (or its ruins): its name, the count, the line, and what to type for it.
func DescribeBuilding(m *mapmodel.Model, key, lineage string, ruin bool) mapstyle.Inspection {
	return (&view{}).describeTile(&scene{m: m}, &mapmodel.TownTile{Key: key, Lineage: lineage, Ruin: ruin})
}

// DescribeWonder says what the cursor is on when it is on a wonder.
func DescribeWonder(m *mapmodel.Model, key string) mapstyle.Inspection {
	return describeWonder(m, &mapmodel.TownWonder{Key: key})
}

// DescribeSquare says what the cursor is on when it is on the town square.
func DescribeSquare(m *mapmodel.Model) mapstyle.Inspection {
	return describeSquare(m, m.Town.World.Name, m.Epoch)
}

// RoadKind is a sort of road for DescribeRoad.
type RoadKind uint8

const (
	RoadStreet RoadKind = iota
	RoadAvenue
	RoadBridge
)

// DescribeRoad says what the cursor is on when it is on a road.
func DescribeRoad(m *mapmodel.Model, k RoadKind) mapstyle.Inspection {
	kinds := [...]kind{RoadStreet: kStreet, RoadAvenue: kStreet, RoadBridge: kBridge}
	return describeRoad(m.Town.World.Name, kinds[k], k == RoadAvenue, false)
}

// DescribeTerrain says what the cursor is on when it is on open ground of a
// terrain.
func DescribeTerrain(t mapmodel.Terrain) mapstyle.Inspection {
	in := mapstyle.Inspection{Title: terrainNames[t].String()}
	if n := terrainNotes[t].String(); n != "" {
		in.Lines = append(in.Lines, n)
	}
	return in
}

// ZoomName is the name of a zoom: 0 the region, 1 the settlement, 2 the
// district.
func ZoomName(zoom int) string { return zoomNames[zoom].String() }
