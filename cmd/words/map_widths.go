package main

import (
	"go/ast"
	"go/token"
	"strconv"
	"strings"
)

// map_widths.go: the widths the map gives its labels, read by the export as the
// max column of the map sheet, and the string-literal reader the width check
// uses.

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

// stringOf is the value of a single string literal. Unlike stringLit, it does
// not follow + or parentheses.
func stringOf(e ast.Expr) (string, bool) {
	lit, ok := e.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return "", false
	}
	s, err := strconv.Unquote(lit.Value)
	if err != nil {
		return "", false
	}
	return s, true
}
