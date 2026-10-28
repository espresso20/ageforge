package ui

import (
	"fmt"
	"strings"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/vibe"
)

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

// ===== Overview =====

func (c peopleBody) overview() {
	g, st := c.g, c.st
	ws := st.Workers
	kinds := peopleKinds(st)
	iw := c.w - 4

	// Your people: homes, the bar over them, the four kinds, food.
	legend := [][]xseg{}
	line := []xseg{}
	used := 0
	for _, k := range kinds {
		text := []xseg{{"█ " + k.name + " ", k.kind}, {num(k.count), xText}}
		n := runeLen(text[0].s) + runeLen(text[1].s)
		if len(line) > 0 && used+3+n > iw {
			legend = append(legend, line)
			line, used = nil, 0
		}
		if len(line) > 0 {
			line = append(line, xseg{"   ", xText})
			used += 3
		}
		line = append(line, text...)
		used += n
	}
	legend = append(legend, line)
	rows := 3 + len(legend) // homes, bar, legend, food
	if c.h < 16 {
		rows--
	}
	ph := rows + 2
	g.box(c.x, c.y, c.w, ph, xBorder, vibe.Line("ui.people_tabs.overview.01", "Your people"), xTitle, false)
	y := c.y + 1
	held := vibe.Line("ui.people_tabs.overview.02", "nothing, growing")
	switch ws.HeldBy {
	case game.HeldHousing:
		held = "housing"
	case game.HeldFood:
		held = "food"
		if t := heldByText(ws); strings.Contains(t, "(") {
			held = vibe.Line("ui.people_tabs.overview.03", "food ") + t[strings.Index(t, "("):]
		}
	}
	g.segsFit(3, y, iw, xseg{vibe.Line("ui.people_tabs.overview.04", "Homes "), xDim}, xseg{num(ws.MaxPop), xText}, xseg{vibe.Line("ui.people_tabs.overview.05", "   empty "), xDim},
		xseg{num(max(0, ws.MaxPop-ws.TotalPop)), xText}, xseg{vibe.Line("ui.people_tabs.overview.06", "   held by "), xDim}, xseg{held, xText})
	y++
	stackBar(g, 3, y, iw, []stackPart{{ws.Working(), xWork}, {ws.Reserve(), xRes}, {ws.Soldiers, xArmy}, {ws.Scholars, xLore}}, ws.MaxPop)
	y++
	for _, l := range legend {
		g.segsFit(3, y, iw, l...)
		y++
	}
	if c.h >= 16 {
		c.foodLine(3, y, iw)
	}

	// One card for each kind.
	cy := c.y + ph
	ch := 5
	if c.h >= 22 {
		ch = 7
	}
	cw := c.w / 4
	for i, k := range kinds {
		x := c.x + i*cw
		w := cw
		if i == 3 {
			w = c.w - 3*cw
		}
		bk, tk := xBorder, xTitle
		if i == c.v.row {
			bk, tk = xAccent, xBright
		}
		g.box(x, cy, w, ch, bk, k.name, tk, false)
		for j, l := range c.card(i, k) {
			if j >= ch-2 {
				break
			}
			kind := xDim
			if j == 0 {
				kind = xBright
			}
			g.putFit(x+2, cy+1+j, w-4, l, kind)
		}
	}

	// Lately: what happened to the people.
	ly := cy + ch
	if lh := c.y + c.h - ly; lh >= 4 {
		g.box(c.x, ly, c.w, lh, xBorder, vibe.Line("ui.people_tabs.overview.07", "Lately"), xTitle, false)
		drawLog(g, 3, ly+1, c.w-4, lh-2, recentLog(st, peopleEvent, lh-2), vibe.Line("ui.people_tabs.overview.08", "Nothing yet. People move in as you build homes."))
	}
	c.footer([][2]string{{"←→", "tabs"}, {"↑↓", "pick"}, {"Esc", "close"}}, "people")
}

// txtFood is said the same way on several screens, so it has one id.
var txtFood = vibe.E("ui.people_tabs.foodLine.01", "Food ")

// foodLine writes what everyone eats, what food nets and how many it feeds.
func (c peopleBody) foodLine(x, y, w int) {
	ws := c.st.Workers
	rs, ok := c.st.Resources["food"]
	if !ok {
		return
	}
	per := 0.0
	if ws.TotalPop > 0 {
		per = ws.FoodDrain / float64(ws.TotalPop)
	}
	net := xGood
	if rs.Rate < 0 {
		net = xBad
	}
	parts := []xseg{{txtFood.String(), xDim}, {textfmt.Rate(-ws.FoodDrain), xText}, {vibe.Line("ui.people_tabs.foodLine.02", " for everyone"), xDim}, {vibe.Line("ui.people_tabs.foodLine.03", "   net "), xDim}, {textfmt.Rate(rs.Rate), net}}
	// Morale, which the old Workers panel showed, before the figure that goes
	// first on a narrow panel.
	mk := xText
	if band := computeMoraleBand(c.st.Morale, c.st.MoraleMultiplier); band.Bonus {
		mk = xGood
	} else if band.Penalty {
		mk = xBad
	}
	parts = append(parts, xseg{vibe.Line("ui.people_tabs.foodLine.04", "   morale "), xDim}, xseg{fmt.Sprintf("%.0f%%", c.st.Morale*100), mk})
	if grown := game.FoodBeforeMastery(rs.Rate, rs.Breakdown); per > 0 && grown > 0 && rs.Rate >= 0 {
		parts = append(parts, xseg{vibe.Line("ui.people_tabs.foodLine.05", "   feeds up to "), xDim}, xseg{num(int(grown / per)), xText})
	}
	c.g.segsFit(x, y, w, parts...)
}

// card is a kind's card: its count, then facts about it.
func (c peopleBody) card(i int, k peopleKind) []string {
	st := c.st
	switch i {
	case 0:
		slots, filled, areas := 0, 0, 0
		for _, r := range game.ShareRows(st) {
			slots += r.Slots
			filled += r.Workers
			areas++
		}
		mode := "roster auto"
		if len(st.Workers.Shares) > 0 {
			mode = vibe.Line("ui.people_tabs.card.01", "roster set")
		}
		return []string{num(k.count) + vibe.Line("ui.people_tabs.card.02", " at work"), vibe.Sprintf("ui.people_tabs.card.03", "slots %s/%s", num(filled), num(slots)), textfmt.Count(areas, vibe.Line("ui.people_tabs.card.04", "work area"), vibe.Line("ui.people_tabs.card.05", "work areas")), mode}
	case 1:
		return []string{num(k.count) + vibe.Line("ui.people_tabs.card.06", " waiting"), vibe.Line("ui.people_tabs.card.07", "fill jobs first"), vibe.Line("ui.people_tabs.card.08", "grow with homes"), vibe.Line("ui.people_tabs.card.09", "see Work")}
	case 2:
		out := vibe.Line("ui.people_tabs.card.10", "none out")
		if k.out > 0 {
			out = num(k.out) + txtOut.String()
		}
		return []string{num(k.count-k.out) + vibe.Line("ui.people_tabs.card.12", " at their posts"), out, vibe.Line("ui.people_tabs.card.13", "defense ") + num(int(st.Military.DefenseRating)), vibe.Line("ui.people_tabs.card.14", "see Army")}
	}
	s := st.Scholars
	return []string{num(k.count) + vibe.Line("ui.people_tabs.card.15", " at desks"), num(s.Researching) + vibe.Line("ui.people_tabs.card.16", " researching"), num(s.Teaching) + vibe.Line("ui.people_tabs.card.17", " teaching"), vibe.Line("ui.people_tabs.card.18", "come with buildings")}
}

// ===== Work =====

// recruitText is what auto-recruit is doing, in plain words.
func recruitText(st game.GameState) string {
	switch game.RecruitStatus(st) {
	case game.RecruitOff:
		return vibe.Line("ui.people_tabs.recruitText.01", "Auto-recruit is off. Turn it on with: workers auto-recruit on")
	case game.RecruitHeld:
		return vibe.Line("ui.people_tabs.recruitText.02", "Auto-recruit is waiting after your last worker command, ") + formatTicks(st.Workers.HoldTicks, st) + vibe.Line("ui.people_tabs.recruitText.03", " more.")
	case game.RecruitHousing:
		return vibe.Line("ui.people_tabs.recruitText.04", "Auto-recruit is held by housing: build more homes for more people.")
	case game.RecruitFood:
		return vibe.Line("ui.people_tabs.recruitText.05", "Auto-recruit is held by food: build or staff food buildings.")
	case game.RecruitStarve:
		return vibe.Line("ui.people_tabs.recruitText.06", "Auto-recruit is paused: food has run out.")
	}
	return vibe.Line("ui.people_tabs.recruitText.07", "Auto-recruit is on: the town is growing toward its homes.")
}

func (c peopleBody) work() {
	g, st := c.g, c.st
	rows := game.ShareRows(st)
	lw := c.w * 11 / 20
	rw := c.w - lw
	g.box(c.x, c.y, lw, c.h, xBorder, vibe.Line("ui.people_tabs.work.01", "Work areas"), xTitle, false)
	g.box(c.x+lw, c.y, rw, c.h, xBorder, "", xTitle, false)
	if len(rows) == 0 {
		g.putFit(c.x+2, c.y+1, lw-4, vibe.Line("ui.people_tabs.work.02", "No worker buildings yet: every area is on auto."), xFaint)
		c.footer([][2]string{{"←→", "tabs"}, {"Esc", "close"}}, "roster <area> <%|auto>")
		return
	}
	inner := lw - 4
	g.segsFit(c.x+2, c.y+1, inner, xseg{vibe.Line("ui.people_tabs.work.06", "  Area"), xDim})
	g.put(c.x+2+2+11+1, c.y+1, vibe.Line("ui.people_tabs.work.07", "Share"), xDim)
	g.rput(rightOf(c.x, lw), c.y+1, vibe.Line("ui.people_tabs.work.08", "Filled"), xDim)
	per := (c.h - 3) / 2
	first := window(len(rows), per, c.v.row)
	for n := 0; n < per && first+n < len(rows); n++ {
		i := first + n
		r := rows[i]
		y := c.y + 2 + n*2
		pickName(g, c.x+2, y, 13, r.Name, i == c.v.row)
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
		g.putFit(c.x+4, y+1, inner-2, c.areaBuildings(r.Domain), xDim)
	}
	// The picked area.
	r := rows[clamp(c.v.row, 0, len(rows)-1)]
	g.put(c.x+lw+2, c.y, " "+fitText(r.Name, rw-6)+" ", xTitle)
	rx, ri := c.x+lw+2, rw-4
	y := c.y + 1
	bs := c.areaList(r.Domain)
	for i, b := range bs {
		if y >= c.y+c.h-8 {
			g.putFit(rx, y, ri, vibe.Sprintf("ui.people_tabs.work.09", "and %d more", len(bs)-i), xFaint)
			y++
			break
		}
		count := fmt.Sprintf("%s/%s", num(b.WorkersAssigned), num(b.WorkerCapacity*b.Count))
		g.putFit(rx, y, ri-runeLen(count)-1, fmt.Sprintf("%s x%d", b.Name, b.Count), xText)
		g.rput(rx+ri-1, y, count, xGood)
		y++
	}
	if len(bs) == 0 {
		g.putFit(rx, y, ri, vibe.Line("ui.people_tabs.work.10", "No buildings here yet."), xFaint)
		y++
	}
	y++
	share := vibe.Line("ui.people_tabs.work.11", "auto, by its buildings' slots")
	if r.Set {
		share = vibe.Line("ui.people_tabs.work.12", "set to ") + game.SharePercent(float64(int(r.Percent+0.5)))
	}
	g.segsFit(rx, y, ri, xseg{vibe.Line("ui.people_tabs.work.13", "Share  "), xDim}, xseg{share, xText})
	y++
	g.segsFit(rx, y, ri, xseg{vibe.Line("ui.people_tabs.work.14", "Set it  "), xDim}, xseg{"roster " + strings.ToLower(r.Name) + " <%|auto>", xText})
	y += 2
	lines := wrapText(recruitText(st), ri)
	openAt := len(lines) // the first line of the note after it is gold
	lines = append(lines, wrapText(vibe.Line("ui.people_tabs.work.17", "Open slots fill from the reserve."), ri)...)
	for i, l := range lines {
		if y >= c.y+c.h-1 {
			break
		}
		k := xDim
		if i == openAt {
			k = xGold
		}
		g.put(rx, y, l, k)
		y++
	}
	c.footer([][2]string{{"←→", "tabs"}, {"↑↓", vibe.Line("ui.people_tabs.work.19", "pick an area")}}, "roster <area> <%|auto>")
}

// areaList is the built buildings of a work area that take workers.
func (c peopleBody) areaList(domain string) []peopleBuilding {
	return builtWhere(c.st, func(b game.BuildingState) bool { return b.WorkerDomain == domain && b.WorkerCapacity > 0 })
}

// areaBuildings is a work area's buildings on one line: "Farm 10/10 · Smithy
// 64/64".
func (c peopleBody) areaBuildings(domain string) string {
	var parts []string
	for _, b := range c.areaList(domain) {
		parts = append(parts, fmt.Sprintf("%s %s/%s", b.Name, num(b.WorkersAssigned), num(b.WorkerCapacity*b.Count)))
	}
	return strings.Join(parts, " · ")
}

// ===== Army =====

func (c peopleBody) army() {
	g, st := c.g, c.st
	mil := st.Military
	kinds := peopleKinds(st)
	soldiers := kinds[2]
	topH := max(8, c.h*3/5)
	lw := c.w * 11 / 20
	rw := c.w - lw
	g.box(c.x, c.y, lw, topH, xBorder, vibe.Line("ui.people_tabs.army.01", "Military buildings"), xTitle, false)
	g.box(c.x+lw, c.y, rw, topH, xBorder, vibe.Sprintf("ui.people_tabs.army.02", "Away · %s", num(soldiers.out)), xTitle, false)

	// The buildings with their posts, then the army and the defense under them.
	blds := militaryBuildings(st)
	per := max(0, (topH-2-5)/2)
	if len(blds) == 0 {
		g.putFit(c.x+2, c.y+1, lw-4, vibe.Line("ui.people_tabs.army.03", "No military buildings yet."), xFaint)
	}
	posts := map[string]game.SoldierPost{}
	for _, p := range mil.Posts {
		posts[p.Key] = p
	}
	first := window(len(blds), per, c.v.row)
	for n := 0; n < per && first+n < len(blds); n++ {
		i := first + n
		b := blds[i]
		y := c.y + 1 + n*2
		p := posts[b.key]
		staff := vibe.Sprintf("ui.people_tabs.army.04", "%s/%s posted", num(p.Posted), num(p.Posts))
		pickName(g, c.x+2, y, lw-6-runeLen(staff), fmt.Sprintf("%s x%d", b.Name, b.Count), i == c.v.row)
		sk := xGood
		if p.Posted < p.Posts {
			sk = xBad
		}
		g.rput(rightOf(c.x, lw), y, staff, sk)
		g.putFit(c.x+4, y+1, lw-8, b.Description, xDim)
	}
	sy := c.y + topH - 6
	g.segsFit(c.x+2, sy, lw-4, xseg{soldiers.name + " ", xDim}, xseg{num(soldiers.count) + vibe.Line("ui.people_tabs.army.05", " posted of ") + num(mil.SoldierCap), xText})
	g.segsFit(c.x+2, sy+1, lw-4, xseg{vibe.Line("ui.people_tabs.army.06", "Defense "), xDim}, xseg{num(int(mil.DefenseRating)), xText},
		xseg{vibe.Line("ui.people_tabs.army.07", "  threat ") + num(int(mil.Threat)) + vibe.Line("ui.people_tabs.army.08", ", blunts ") + textfmt.Percent(mil.Mitigation), xDim})
	g.segsFit(c.x+2, sy+2, lw-4, xseg{vibe.Line("ui.people_tabs.army.09", "Rank "), xDim}, xseg{rankName(st, mil.Rank), xBright}, xseg{" · " + rankProgress(st) + vibe.Line("ui.people_tabs.army.10", " experience"), xText})
	g.segsFit(c.x+2, sy+3, lw-4, xseg{vibe.Line("ui.people_tabs.army.11", "Gives "), xDim}, xseg{rankGives(mil.Rank), xText})
	g.segsFit(c.x+2, sy+4, lw-4, xseg{rankNext(st), xDim})

	// Away: each party out.
	y := c.y + 1
	away := 0
	for _, m := range mil.Active {
		if m == nil {
			continue
		}
		away++
		if y+1 >= c.y+topH-1 {
			break
		}
		g.putFit(c.x+lw+2, y, rw-4, m.Name, xBright)
		g.putFit(c.x+lw+4, y+1, rw-6, vibe.Sprintf("ui.people_tabs.army.12", "%s · back in %s", textfmt.Count(m.Soldiers, soldiers.one, soldiers.name), formatTicks(m.TicksLeft, st)), xDim)
		y += 2
	}
	if away == 0 {
		g.putFit(c.x+lw+2, y, rw-4, vibe.Line("ui.people_tabs.army.13", "Nobody is out."), xFaint)
	}

	// The soldiers' news.
	by := c.y + topH
	if bh := c.h - topH; bh >= 4 {
		g.box(c.x, by, c.w, bh, xBorder, vibe.Line("ui.people_tabs.army.14", "Fallen and filled"), xTitle, false)
		drawLog(g, 3, by+1, c.w-4, bh-2, recentLog(st, armyEvent, bh-2), vibe.Line("ui.people_tabs.army.15", "No news from the army."))
	}
	c.footer([][2]string{{"←→", "tabs"}, {"↑↓", vibe.Line("ui.people_tabs.army.17", "pick a building")}}, "army")
}

// ===== Scholars =====

func (c peopleBody) scholars() {
	g, st := c.g, c.st
	kinds := peopleKinds(st)
	sch := kinds[3]
	topH := max(8, c.h*3/5)
	lw := c.w * 11 / 20
	rw := c.w - lw
	g.box(c.x, c.y, lw, topH, xBorder, vibe.Line("ui.people_tabs.scholars.01", "Desks"), xTitle, false)
	g.box(c.x+lw, c.y, rw, topH, xBorder, vibe.Line("ui.people_tabs.scholars.02", "Research"), xTitle, false)

	blds := knowledgeBuildings(st)
	per := max(0, (topH-2-1)/2)
	if len(blds) == 0 {
		g.putFit(c.x+2, c.y+1, lw-4, vibe.Line("ui.people_tabs.scholars.03", "No knowledge buildings yet: each one comes with its ")+strings.ToLower(sch.name)+".", xFaint)
	}
	first := window(len(blds), per, c.v.row)
	for n := 0; n < per && first+n < len(blds); n++ {
		i := first + n
		b := blds[i]
		y := c.y + 1 + n*2
		each := vibe.Sprintf("ui.people_tabs.scholars.04", "%s each", num(b.Scholars))
		pickName(g, c.x+2, y, lw-6-runeLen(each), fmt.Sprintf("%s x%d", b.Name, b.Count), i == c.v.row)
		g.rput(rightOf(c.x, lw), y, each, xDim)
		const bw = 10
		fill := bw
		if b.FullRateCopies > 0 {
			fill = bw * min(b.Count, b.FullRateCopies) / b.FullRateCopies
		}
		for k := 0; k < bw; k++ {
			if k < fill {
				g.set(c.x+4+k, y+1, '█', xLore)
			} else {
				g.set(c.x+4+k, y+1, '░', xFaint)
			}
		}
		total := fmt.Sprintf("%s %s", num(b.Count*b.Scholars), strings.ToLower(sch.name))
		if b.FullRateCopies > 0 {
			if b.Count > b.FullRateCopies {
				total += vibe.Sprintf("ui.people_tabs.scholars.23", ", %d copies, %d at half rate", b.Count, b.Count-b.FullRateCopies)
			} else {
				total += vibe.Sprintf("ui.people_tabs.scholars.05", ", %d of %d copies", b.Count, b.FullRateCopies)
			}
		}
		g.putFit(c.x+4+bw+1, y+1, lw-4-bw-4, total, xText)
	}

	// Research.
	r := st.Research
	rx, ri := c.x+lw+2, rw-4
	studying, progress := "nothing", "-"
	if r.CurrentTech != "" {
		studying = r.CurrentTechName
		progress = fmt.Sprintf("%d%%, %s", researchPercent(st), researchLeft(st))
	}
	next := vibe.Line("ui.people_tabs.scholars.07", "nothing planned")
	if q := queueLines(st); len(q) > 0 {
		next = strings.TrimPrefix(q[0], "1. ")
	}
	s := st.Scholars
	for i, row := range [][2]string{
		{vibe.Line("ui.people_tabs.scholars.08", "Studying"), studying},
		{vibe.Line("ui.people_tabs.scholars.09", "Progress"), progress},
		{vibe.Line("ui.people_tabs.scholars.10", "Research"), textfmt.RateValue(s.PerTick) + vibe.Line("ui.people_tabs.scholars.11", " a tick")},
		{sch.name, vibe.Sprintf("ui.people_tabs.scholars.12", "%s: %s research, %s teach", num(s.Total), num(s.Researching), num(s.Teaching))},
		{vibe.Line("ui.people_tabs.scholars.13", "Teaching"), textfmt.SignedPercent(s.TeachingBonus) + vibe.Line("ui.people_tabs.scholars.14", " to what workers make")},
		{txtNext.String(), next},
	} {
		y := c.y + 1 + i
		if y >= c.y+topH-1 {
			break
		}
		lab := fitText(row[0], 9)
		g.put(rx, y, lab, xDim)
		g.putFit(rx+10, y, ri-10, row[1], xText)
	}

	by := c.y + topH
	if bh := c.h - topH; bh >= 4 {
		g.box(c.x, by, c.w, bh, xBorder, vibe.Line("ui.people_tabs.scholars.16", "How to speed research"), xTitle, false)
		tips := []string{
			vibe.Line("ui.people_tabs.scholars.17", "Build knowledge buildings: each comes with its ") + strings.ToLower(sch.name) + ".",
			vibe.Line("ui.people_tabs.scholars.18", "Research techs that make your ") + strings.ToLower(sch.name) + vibe.Line("ui.people_tabs.scholars.19", " faster."),
			vibe.Line("ui.people_tabs.scholars.20", "Expeditions and events can bring back finds that add to the research in progress."),
		}
		for i, t := range tips {
			if i >= bh-2 {
				break
			}
			g.putFit(3, by+1+i, c.w-4, "· "+t, xText)
		}
	}
	c.footer([][2]string{{"←→", "tabs"}, {"↑↓", vibe.Line("ui.people_tabs.scholars.22", "pick a building")}}, "scholars")
}
