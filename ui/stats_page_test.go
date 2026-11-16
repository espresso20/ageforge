package ui

import (
	"strings"

	"github.com/espresso20/ageforge/game"
)

// gridLines is a grid's rows as text, trailing blanks trimmed.
func gridLines(g *xgrid) []string {
	lines := make([]string, g.h)
	for y := range lines {
		row := make([]rune, g.w)
		for x := range row {
			row[x] = g.c[y*g.w+x].r
			if row[x] == 0 {
				row[x] = ' '
			}
		}
		lines[y] = strings.TrimRight(string(row), " ")
	}
	return lines
}

// statsRowsText is a table's rows as text, a row to a line.
func statsRowsText(rows []statsRow) string {
	var sb strings.Builder
	for _, r := range rows {
		sb.WriteString(strings.TrimSpace(r.label + " " + r.value))
		sb.WriteByte('\n')
	}
	return sb.String()
}

// statsPageText is the text of a part of the page in cells only, w wide.
func statsPageText(st game.GameState, part, w int) string {
	return strings.Join(gridLines(buildStats(st, part, w, w >= statsWide, false, 0).g), "\n")
}
