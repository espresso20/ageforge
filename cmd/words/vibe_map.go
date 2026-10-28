package main

import "strings"

// vibe_map.go: what vibe-wire decides about the map sheet that the syntax
// alone does not say. The rest of the sheet is decided like the screens: a text
// in a function is wired as a vibe call, and a text in a table made once becomes
// an entry that carries its id (vibe.E), or vibe.Same where the table holds a
// string that has no row (see vibe/vibe.go).

// typedIn are the commands the player types that a sentence of the map names
// as they are typed (the hint line that ends in "map flows"), by row id. The
// keep column cannot tell one from the words around it, so the sheet lists them
// with the keys and the verbs, and vibe-lint holds a line to them.
var typedIn = map[string][]string{
	"ui.mapstyle.roguelike.chrome.bottom.01": {"map flows"},
}

// mapWidth is a limit on the rows of the map sheet whose ids start with prefix:
// the width the map gives the label. The map clips what it draws to the cells it
// has, so a longer line is cut off; a label with a fixed width also pushes the
// text next to it out of its column.
type mapWidth struct {
	prefix string
	max    int
}

// mapWidths are the widths the map's code gives labels. The sidebar is 26
// cells wide (sideWidth) and the legend 23: it starts 2 cells in and stops 1
// before the right edge. A label sits 1 cell in and follows a glyph and a
// space, which leaves 20. The flows panel keeps its tags (DRAIN, IDLE, SHORT,
// OK) in a column 5 cells wide.
var mapWidths = []mapWidth{
	{prefix: "map.lineage.", max: 20}, // the legend label of a kind of building
	{prefix: "ui.mapstyle.roguelike.chrome.flowLines.01", max: 5}, {prefix: "ui.mapstyle.roguelike.chrome.flowLines.02", max: 5},
	{prefix: "ui.mapstyle.roguelike.chrome.flowLines.05", max: 5}, {prefix: "ui.mapstyle.roguelike.chrome.flowLines.08", max: 5},
}

// mapWidthOf is the width the map gives a row's text, or 0 when it gives none.
func mapWidthOf(r *row) int {
	for _, w := range mapWidths {
		if strings.HasPrefix(r.id, w.prefix) {
			return w.max
		}
	}
	return 0
}
