package ui

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/peopledash"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/rowid"
)

// people_army.go is the Army and Scholars tabs: a band, a list of buildings
// with blocks, a pane of figures, and a pane along the foot.

// figure paints a bold numeral at a cell with its label under it.
func (c peopleBody) figure(x, y int, value, label string, col color.RGBA) int {
	nx := c.s.num(x, y, value, col, true, 3)
	c.s.tx(x, y+c.s.numH(true, 3), label, c.s.pal.Dim, false)
	return max(nx, x+runeLen(label)) + 2
}

// panes cuts the tab's body into a list on the left, figures on the right and
// a pane along the foot; it returns the heights and the left width.
func (c peopleBody) panes(col color.RGBA, left, right, foot string) (topH, lw, rw int) {
	topH = max(8, c.h*3/5)
	lw = c.w * 11 / 20
	rw = c.w - lw
	c.pane(c.x, c.y, lw, topH, left, col)
	c.pane(c.x+lw, c.y, rw, topH, right, col)
	if c.h-topH >= 4 {
		c.pane(c.x, c.y+topH, c.w, c.h-topH, foot, col)
	}
	return
}

// listRow draws a building's picked name and its figure at the right, and
// makes the two rows a click target.
func (c peopleBody) listRow(i, y, lw int, b peopleBuilding, right string, rk xk) {
	g := c.g
	pickName(g, c.x+2, y, lw-6-runeLen(right), fmt.Sprintf("%s x%d", b.Name, b.Count), i == c.v.row)
	g.hit(hitRow, i, c.x+1, y, lw-2, 2)
	g.rput(rightOf(c.x, lw), y, right, rk)
}

func (c peopleBody) army() {
	g, st := c.g, c.st
	mil, soldiers := st.Military, peopleKinds(st)[2]
	blds := militaryBuildings(st)
	posts := map[string]game.SoldierPost{}
	parts := make([]bandPart, len(blds))
	for _, p := range mil.Posts {
		posts[p.Key] = p
	}
	for i, b := range blds {
		parts[i] = bandPart{posts[b.key].Posts, posts[b.key].Posted}
	}
	chip := ""
	if soldiers.out > 0 {
		chip = rowid.Sprintf("ui.people_army.words.01", "%s away", num(soldiers.out))
	}
	c.band(num(soldiers.count), rowid.Sprintf("ui.people_army.band.01", "posted of %s", num(mil.SoldierCap)), chip,
		rankName(st, mil.Rank)+" · "+rankProgress(st)+rowid.Line("ui.people_tabs.army.10", " experience"), c.s.pal.Soldiers, parts, "", "")
	c.y, c.h = c.y+workBandH(g.h)+1, c.h-workBandH(g.h)-1
	topH, lw, rw := c.panes(c.s.pal.Soldiers, rowid.Line("ui.people_tabs.army.01", "Military buildings"), rowid.Line("ui.people_army.title.01", "Rank and defense"), rowid.Line("ui.people_tabs.army.14", "Fallen and filled"))

	if len(blds) == 0 {
		g.putFit(c.x+2, c.y+1, lw-4, rowid.Line("ui.people_tabs.army.03", "No military buildings yet."), xFaint)
	}
	per := max(0, (topH-2)/2)
	first := window(len(blds), per, c.v.row)
	for n := 0; n < per && first+n < len(blds); n++ {
		i, y := first+n, c.y+1+n*2
		b, p := blds[first+n], posts[blds[first+n].key]
		sk := xGood
		if p.Posted < p.Posts {
			sk = xBad
		}
		c.listRow(i, y, lw, b, rowid.Sprintf("ui.people_tabs.army.04", "%s/%s posted", num(p.Posted), num(p.Posts)), sk)
		c.s.blocks(c.x+4, y+1, 1, peopledash.BlockCount(p.Posts), peopledash.BlockCount(p.Posted), c.s.pal.Soldiers, c.s.pal.Tile, 12)
		g.putFit(c.x+18, y+1, lw-22, b.Description, xDim)
	}

	x, ri := c.x+lw+2, rw-4
	fx := c.figure(x, c.y+1, num(int(mil.DefenseRating)), rowid.Line("ui.people_army.words.02", "Defense"), c.s.pal.Text)
	fx = c.figure(fx, c.y+1, num(int(mil.Threat)), rowid.Line("ui.people_army.words.03", "Threat"), c.s.pal.Alarm)
	c.figure(fx, c.y+1, textfmt.Percent(mil.Mitigation), rowid.Line("ui.people_army.words.04", "Blunted"), c.s.pal.Green)
	g.segsFit(x, c.y+5, ri, xseg{rowid.Line("ui.people_tabs.army.11", "Gives "), xDim}, xseg{rankGives(mil.Rank), xText})
	g.putFit(x, c.y+6, ri, rankNext(st), xDim)
	y := c.y + 8
	g.put(x, y, rowid.Sprintf("ui.people_tabs.army.02", "Away · %s", num(soldiers.out)), xDim)
	y++
	for _, m := range mil.Active {
		if m == nil || y+1 >= c.y+topH-1 {
			continue
		}
		g.putFit(x, y, ri, fmt.Sprintf("%s · %s", m.Name, rowid.Sprintf("ui.people_tabs.army.12", "%s · back in %s", textfmt.Count(m.Soldiers, soldiers.one, soldiers.name), formatTicks(m.TicksLeft, st))), xText)
		y++
	}
	if y == c.y+9 {
		g.putFit(x, y, ri, rowid.Line("ui.people_tabs.army.13", "Nobody is out."), xFaint)
	}
	if bh := c.h - topH; bh >= 4 {
		drawLog(g, 3, c.y+topH+1, c.w-4, bh-2, recentLog(st, armyEvent, bh-2), rowid.Line("ui.people_tabs.army.15", "No news from the army."))
	}
	c.footer([][2]string{{"↑↓", rowid.Line("ui.people_tabs.army.17", "pick a building")}, escClose()}, "army")
}

func (c peopleBody) scholars() {
	g, st := c.g, c.st
	sch, s := peopleKinds(st)[3], st.Scholars
	lower := strings.ToLower(sch.name)
	blds := knowledgeBuildings(st)
	parts := make([]bandPart, len(blds))
	for i, b := range blds {
		full := b.Count
		if b.FullRateCopies > 0 {
			full = min(b.Count, b.FullRateCopies)
		}
		parts[i] = bandPart{b.Count * b.Scholars, full * b.Scholars}
	}
	r := st.Research
	note := rowid.Line("ui.people_scholars.words.01", "Nothing is being researched.")
	if r.CurrentTech != "" {
		note = fmt.Sprintf("%s · %d%%, %s", r.CurrentTechName, researchPercent(st), researchLeft(st))
	}
	c.band(num(s.Total), lower, "", note, c.s.pal.Scholars, parts, rowid.Line("ui.people_scholars.words.02", "Pick research"), "research")
	c.y, c.h = c.y+workBandH(g.h)+1, c.h-workBandH(g.h)-1
	topH, lw, rw := c.panes(c.s.pal.Scholars, rowid.Line("ui.people_tabs.scholars.01", "Desks"), rowid.Line("ui.people_tabs.scholars.02", "Research"), rowid.Line("ui.people_tabs.scholars.16", "How to speed research"))

	if len(blds) == 0 {
		g.putFit(c.x+2, c.y+1, lw-4, rowid.Line("ui.people_tabs.scholars.03", "No knowledge buildings yet: each one comes with its ")+lower+".", xFaint)
	}
	per := max(0, (topH-2)/2)
	first := window(len(blds), per, c.v.row)
	for n := 0; n < per && first+n < len(blds); n++ {
		i, y := first+n, c.y+1+n*2
		b := blds[i]
		c.listRow(i, y, lw, b, rowid.Sprintf("ui.people_tabs.scholars.04", "%s each", num(b.Scholars)), xDim)
		c.s.blocks(c.x+4, y+1, 1, peopledash.BlockCount(parts[i].slots), peopledash.BlockCount(parts[i].filled), c.s.pal.Scholars, c.s.pal.Tile, 12)
		total := fmt.Sprintf("%s %s", num(b.Count*b.Scholars), lower)
		if b.FullRateCopies > 0 {
			if b.Count > b.FullRateCopies {
				total += rowid.Sprintf("ui.people_tabs.scholars.23", ", %d copies, %d at half rate", b.Count, b.Count-b.FullRateCopies)
			} else {
				total += rowid.Sprintf("ui.people_tabs.scholars.05", ", %d of %d copies", b.Count, b.FullRateCopies)
			}
		}
		g.putFit(c.x+18, y+1, lw-22, total, xText)
	}

	x, ri := c.x+lw+2, rw-4
	fx := c.figure(x, c.y+1, num(s.Researching), rowid.Line("ui.people_scholars.words.03", "Researching"), c.s.pal.Text)
	fx = c.figure(fx, c.y+1, num(s.Teaching), rowid.Line("ui.people_scholars.words.04", "Teaching"), c.s.pal.Text)
	c.figure(fx, c.y+1, textfmt.SignedPercent(s.TeachingBonus), rowid.Line("ui.people_scholars.words.05", "to workers"), c.s.pal.Green)
	if r.CurrentTech != "" {
		c.s.bar(x, c.y+5, ri, float64(researchPercent(st))/100, c.s.pal.Scholars)
	}
	next := rowid.Line("ui.people_tabs.scholars.07", "nothing planned")
	if q := queueLines(st); len(q) > 0 {
		next = strings.TrimPrefix(q[0], "1. ")
	}
	for i, row := range [][2]string{
		{rowid.Line("ui.people_tabs.scholars.10", "Research"), rateText(s.PerTick, st.TickIntervalMs)},
		{txtNext.String(), next},
	} {
		g.put(x, c.y+7+i, row[0], xDim)
		g.putFit(x+10, c.y+7+i, ri-10, row[1], xText)
	}
	if y := c.y + 10; y < c.y+topH-1 {
		c.splitRow(x, y, ri)
	}
	if bh := c.h - topH; bh >= 4 {
		for i, t := range []string{
			rowid.Line("ui.people_tabs.scholars.17", "Build knowledge buildings: each comes with its ") + lower + ".",
			rowid.Line("ui.people_tabs.scholars.18", "Research techs that make your ") + lower + rowid.Line("ui.people_tabs.scholars.19", " faster."),
			rowid.Line("ui.people_tabs.scholars.20", "Expeditions and events can bring back finds that add to the research in progress."),
		} {
			if i < bh-2 {
				g.putFit(3, c.y+topH+1+i, c.w-4, "· "+t, xText)
			}
		}
	}
	c.footer([][2]string{{"↑↓", rowid.Line("ui.people_tabs.scholars.22", "pick a building")}, escClose()}, "scholars")
}
