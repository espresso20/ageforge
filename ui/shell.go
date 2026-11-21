package ui

import (
	"image"
	"strings"

	"github.com/espresso20/ageforge/rowid"
)

// shell.go is the frame every full panel is drawn in: a box with the panel's
// name in its top edge and the close button at its right, the panel's tabs on
// the row under it, and the keys along the foot. A panel draws its own
// content in the cells the shell hands back and nothing of the frame itself,
// so what a panel's frame looks like is decided here and nowhere else.

// shell draws the frame on the whole grid: the box, title in its top edge,
// the close button, and the tabs (none for a panel without) with the one
// numbered open showing. It returns the panel's own cells: inside the edges
// with a column of air each side, under the tabs and above the key row.
func (g *xgrid) shell(title string, open int, tabs ...string) image.Rectangle {
	g.box(0, 0, g.w, g.h, xBorder, title, xTitle, false)
	g.closeButton(g.w)
	top, x := 1, panelPadX
	for i, name := range tabs {
		k := xDim
		if i == open {
			k = xSel
		}
		end := g.put(x, 1, " "+name+" ", k)
		g.tab(image.Rect(x, 1, end, 2), name, i, i == open)
		x, top = end+1, 2
	}
	return image.Rect(panelPadX, top, max(panelPadX+1, g.w-panelPadX), max(top+1, g.h-2))
}

// foot writes the panel's keys on the shell's key row, on a grid tall enough
// to have one.
func (g *xgrid) foot(keys [][2]string, hint string) {
	if g.h >= panelMinRows {
		g.keyRow(g.h-2, keys, hint)
	}
}

// keyRow writes keys along row y as buttons, each a key and what it does. The
// last one (how to close the panel) is there whenever it fits; the others are
// those that fit before it, in order. hint is a command the row says may be
// typed, dropped first: a click on it writes it in the command bar, up to the
// first part the player has to fill in.
func (g *xgrid) keyRow(y int, keys [][2]string, hint string) {
	x := 2
	room := g.w - 2
	last, closeW := len(keys)-1, 0
	if last >= 0 {
		closeW = keyButtonW(keys[last][0], keys[last][1])
	}
	for i, kv := range keys {
		at := x
		if x > 2 {
			at++
		}
		end := at + keyButtonW(kv[0], kv[1])
		if i < last {
			end += 1 + closeW
		}
		if end <= room {
			x = g.keyButton(at, y, kv[0], kv[1])
		}
	}
	if hint == "" {
		return
	}
	lead := rowid.Line("ui.expedition_view.expKeyBar.01", " · or type ")
	if x+runeLen(lead)+chipW(hint, "") <= room {
		line, _, _ := strings.Cut(hint, "<")
		g.chip(g.put(x, y, lead, xFaint), y, chipSpec{kind: keySteel, label: hint, cmd: line, fill: true})
	}
}
