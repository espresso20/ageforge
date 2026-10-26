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
	"ui.mapstyle.roguelike.chrome.bottom.01":       {"map flows"},
	"ui.mapstyle.roguelike.space_chrome.bottom.01": {"map flows"},
}

// mapWidth is a limit on the rows of the map sheet whose ids start with prefix
// (and, where when is set, that the test holds for): the width the map gives the
// label. The map clips what it draws to the cells it has, so a longer line is cut
// off; a label with a fixed width also pushes the text next to it out of its
// column.
type mapWidth struct {
	prefix string
	max    int
	when   func(*row) bool
}

func idEndsWith(suffix string) func(*row) bool {
	return func(r *row) bool { return strings.HasSuffix(r.id, suffix) }
}

// shortMark is a mark of a few capitals (a corporation's initials), not a name.
func shortMark(r *row) bool { return len([]rune(r.current)) <= 3 }

// mapWidths are the widths the map's code gives labels. The legend's labels
// are cut at 20 cells. The sidebar is 26 cells wide (sideW) and the legend 23:
// it starts 2 cells in and stops 1 before the right edge. A label sits 1 cell
// in and follows a glyph and a space, which leaves 20. The longest labels the
// legend was written with, "new since last visit" and "echo of the old town",
// are exactly 20. A group heading has the full 23. The flows panel keeps its
// tags (DRAIN, IDLE, SHORT, OK) in a column 5 cells wide. The megacity's
// holo-ads and towers draw their letters one to a cell: a block carries four
// letters of an ad and three of a corporation's initials.
var mapWidths = []mapWidth{
	{prefix: "ui.mapstyle.roguelike.glyph.lgInfo.", max: 20},
	{prefix: "ui.mapstyle.roguelike.space_legend.skyLgInfo.", max: 20},
	{prefix: "ui.mapstyle.roguelike.glyph.lgGroups.", max: 23},
	{prefix: "map.city.", max: 20, when: idEndsWith(".name")},  // the legend label of a part of the city
	{prefix: "map.mover.", max: 20, when: idEndsWith(".name")}, // the legend label of a mover
	{prefix: "map.sky.", max: 20},                              // the legend label of a part of a station
	{prefix: "map.lineage.", max: 20},                          // the legend label of a kind of building
	{prefix: "ui.mapstyle.roguelike.chrome.flowLines.01", max: 5}, {prefix: "ui.mapstyle.roguelike.chrome.flowLines.02", max: 5},
	{prefix: "ui.mapstyle.roguelike.chrome.flowLines.05", max: 5}, {prefix: "ui.mapstyle.roguelike.chrome.flowLines.08", max: 5},
	{prefix: "ui.mapstyle.roguelike.space_chrome.flowLines.01", max: 5}, {prefix: "ui.mapstyle.roguelike.space_chrome.flowLines.02", max: 5},
	{prefix: "ui.mapstyle.roguelike.space_chrome.flowLines.05", max: 5}, {prefix: "ui.mapstyle.roguelike.space_chrome.flowLines.08", max: 5},
	{prefix: "ui.mapstyle.roguelike.city.holoAds.", max: 4},
	{prefix: "ui.mapstyle.roguelike.city.megacorps.", max: 3, when: shortMark},
}

// mapWidthOf is the width the map gives a row's text, or 0 when it gives none.
func mapWidthOf(r *row) int {
	for _, w := range mapWidths {
		if strings.HasPrefix(r.id, w.prefix) && (w.when == nil || w.when(r)) {
			return w.max
		}
	}
	return 0
}
