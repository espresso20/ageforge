package ui

import (
	"fmt"
	"image/color"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/rowid"
)

var txtNext = rowid.E("ui.badge_case.caseTabs.02", "Next")

// people_tabs.go is the four tabs of the People panel.

// num prints a whole count, written short when it is long.
func num(n int) string { return textfmt.Number(float64(n)) }

// rightOf is the column a right-aligned text of the box starting at x with
// width w ends at (inside its frame).
func rightOf(x, w int) int { return x + w - 3 }

// window is the first row of a list of n rows, rows to a page, that keeps
// the picked row in view.
func window(n, rows, picked int) int {
	if n <= rows || rows <= 0 {
		return 0
	}
	return clamp(picked-rows/2, 0, n-rows)
}

// pickName writes a list row's name, inverse when picked.
func pickName(g *xgrid, x, y, room int, name string, picked bool) {
	mark, k := "  ", xAccent
	if picked {
		mark, k = "▸ ", xSel
	}
	g.put(x, y, mark, xAccent)
	g.putFit(x+2, y, room-2, name, k)
}

// txtFood is said the same way on several screens, so it has one id.
var txtFood = rowid.E("ui.people_tabs.foodLine.01", "Food ")

// ===== Work =====

// recruitText is what auto-recruit is doing, in plain words.
func recruitText(st game.GameState) string {
	switch game.RecruitStatus(st) {
	case game.RecruitOff:
		return rowid.Line("ui.people_tabs.recruitText.01", "Auto-recruit is off. Turn it on with: workers auto-recruit on")
	case game.RecruitHeld:
		return rowid.Line("ui.people_tabs.recruitText.02", "Auto-recruit is waiting after your last worker command, ") + formatTicks(st.Workers.HoldTicks, st) + rowid.Line("ui.people_tabs.recruitText.03", " more.")
	case game.RecruitHousing:
		return rowid.Line("ui.people_tabs.recruitText.04", "Auto-recruit is held by housing: build more homes for more people.")
	case game.RecruitFood:
		return rowid.Line("ui.people_tabs.recruitText.05", "Auto-recruit is held by food: build or staff food buildings.")
	case game.RecruitWork:
		return rowid.Line("ui.people_tabs.recruitText.08", "Auto-recruit is paused: no work for more people.")
	case game.RecruitStarve:
		return rowid.Line("ui.people_tabs.recruitText.06", "Auto-recruit is paused: food has run out.")
	}
	return rowid.Line("ui.people_tabs.recruitText.07", "Auto-recruit is on: the town is growing toward its homes.")
}

// pane is a tile on the tab's picture with its title in the top row.
func (c peopleBody) pane(x, y, w, h int, title string, col color.RGBA) {
	c.s.pane(x, y, w, h, col)
	if title != "" {
		c.g.put(x+2, y, " "+title+" ", xTitle)
	}
}

func (c peopleBody) work() {
	g, st := c.g, c.st
	rows := game.ShareRows(st)
	c.workBand(rows)
	c.y, c.h = c.y+workBandH(c.g.h)+1, c.h-workBandH(c.g.h)-1
	lw := c.w * 11 / 20
	if c.v.shut {
		lw = c.w
	}
	rw := c.w - lw
	if rw > 0 {
		c.pane(c.x+lw, c.y, rw, c.h, "", c.s.pal.Work)
	}
	title := rowid.Line("ui.people_tabs.work.01", "Work areas")
	if len(rows) == 0 {
		c.pane(c.x, c.y, lw, c.h, title, c.s.pal.Work)
		g.putFit(c.x+2, c.y+1, lw-4, rowid.Line("ui.people_tabs.work.02", "No worker buildings yet: every area is on auto."), xFaint)
		c.footer([][2]string{escClose()}, rowid.Line("ui.people_tabs.work.03", "roster <area> <%|auto>"))
		return
	}
	inner := lw - 4
	g.segsFit(c.x+2, c.y+1, inner, xseg{rowid.Line("ui.people_tabs.work.06", "  Area"), xDim})
	g.put(c.x+2+2+11+1, c.y+1, rowid.Line("ui.people_tabs.work.07", "Share"), xDim)
	g.rput(rightOf(c.x, lw), c.y+1, rowid.Line("ui.people_tabs.work.08", "Filled"), xDim)
	// An area takes a row and the rows its buildings need; the window
	// starts where it would with two rows each, and moves on until the
	// picked area is whole.
	notes := make([][]string, len(rows))
	for i, r := range rows {
		notes[i] = c.areaBuildings(r.Domain, inner-2)
	}
	avail := c.h - 3
	first := window(len(rows), avail/2, c.v.row)
	for used := 0; first < c.v.row; first++ {
		used = 0
		for i := first; i <= c.v.row; i++ {
			used += 1 + len(notes[i])
		}
		if used <= avail {
			break
		}
	}
	// The table's tile ends with its last row.
	shown := 0
	for i := first; i < len(rows) && shown+1+len(notes[i]) <= avail; i++ {
		shown += 1 + len(notes[i])
	}
	c.pane(c.x, c.y, lw, shown+3, title, c.s.pal.Work)
	y := c.y + 2
	for i := first; i < len(rows) && y+1+len(notes[i]) <= c.y+2+avail; i++ {
		r := rows[i]
		pickName(g, c.x+2, y, 13, r.Name, i == c.v.row)
		g.hit(hitRow, i, c.x+1, y, lw-2, 1+len(notes[i]))
		share := "auto"
		sk := xDim
		if r.Set {
			share, sk = game.SharePercent(float64(int(r.Percent+0.5))), xGold
		}
		sx := c.x + 2 + 2 + 11 + 1
		g.put(sx, y, share, sk)
		count := fmt.Sprintf("%s/%s", num(r.Workers), num(r.Slots))
		ck := xText
		if r.Slots > 0 && r.Workers >= r.Slots {
			ck = xGood
		}
		cx := rightOf(c.x, lw) - runeLen(count) + 1
		g.put(cx, y, count, ck)
		if bw := cx - (sx + 5) - 1; bw >= 3 {
			fill := 0
			if r.Slots > 0 {
				fill = bw * r.Workers / r.Slots
			}
			for b := 0; b < bw; b++ {
				if b < fill {
					g.set(sx+5+b, y, '█', xWork)
				} else {
					g.set(sx+5+b, y, '░', xFaint)
				}
			}
		}
		for j, l := range notes[i] {
			g.put(c.x+4, y+1+j, l, xDim)
		}
		y += 1 + len(notes[i])
	}
	keys := [][2]string{{"↑↓", rowid.Line("ui.people_tabs.work.19", "pick an area")}, {"Enter", rowid.Line("ui.people_tabs.work.20", "open or close it")}, escClose()}
	if c.v.shut {
		c.footer(keys, "roster <area> <%|auto>")
		return
	}
	// The picked area.
	r := rows[clamp(c.v.row, 0, len(rows)-1)]
	g.put(c.x+lw+2, c.y, " "+fitText(r.Name, rw-6)+" ", xTitle)
	rx, ri := c.x+lw+2, rw-4
	y = c.y + 1
	bs := c.areaList(r.Domain)
	recruitY := c.y + c.h - 1 - c.recruitRows()
	y = c.workList(rx, y, ri, recruitY-6, bs)
	if len(bs) == 0 {
		g.putFit(rx, y, ri, rowid.Line("ui.people_tabs.work.10", "No buildings here yet."), xFaint)
		y++
	}
	y++
	c.shareRow(rx, y, ri, r)
	y += 2
	lines := wrapText(recruitText(st), ri)
	openAt := len(lines) // the first line of the note after it is gold
	lines = append(lines, wrapText(rowid.Line("ui.people_tabs.work.17", "Open slots fill from the reserve."), ri)...)
	for i, l := range lines {
		if y >= recruitY-1 {
			break
		}
		k := xDim
		if i == openAt {
			k = xGold
		}
		g.put(rx, y, l, k)
		y++
	}
	c.recruitBlock(rx, recruitY, ri)
	c.footer(keys, "roster <area> <%|auto>")
}

// areaList is the built buildings of a work area that take workers.
func (c peopleBody) areaList(domain string) []peopleBuilding {
	return builtWhere(c.st, func(b game.BuildingState) bool { return b.WorkerDomain == domain && b.WorkerCapacity > 0 })
}

// areaBuildings is a work area's buildings in rows at most w cells wide,
// broken between buildings: "Farm 10/10 · Smithy 64/64".
func (c peopleBody) areaBuildings(domain string, w int) []string {
	var parts []string
	for _, b := range c.areaList(domain) {
		parts = append(parts, fmt.Sprintf("%s %s/%s", b.Name, num(b.WorkersAssigned), num(b.WorkerCapacity*b.Count)))
	}
	for i := 0; i < len(parts)-1; i++ {
		parts[i] += " ·"
	}
	return wrapLines(parts, w, runeLen)
}
