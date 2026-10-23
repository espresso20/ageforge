package ui

import (
	"fmt"
	"strings"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/vibe"
)

// advance_view.go draws the Advance panel: a pure function of the gate
// (game.AdvanceView) and what the player has picked, into a grid the size of
// the panel (the Expeditions panel's cell grid and palette). From the top:
// the march of ages with the calendar year, the Charter (the four pillars,
// the picked one in detail), what the next age brings and what else stands in
// the way, and the line that says whether Enter advances. The layout is worked
// out at 100 columns by 30 rows and shrinks to 80 by 21: the pillars go from
// three rows each to one, and the boxes on the right give way to each other.

// advanceKeys are the panel's keys as the Help panel lists them.
var advanceKeys = []keyHelp{
	{"Up/Down", vibe.E("ui.advance_view.advanceKeys.02", "Pick a pillar to see what it asks for")},
	{"Enter", vibe.E("ui.advance_view.advanceKeys.04", "Advance to the next age, when all four pillars stand (the same as the advance command)")},
	{"Esc", vibe.E("ui.advance_view.advanceKeys.06", "Close the panel")},
}

// advanceUI is what the player has picked and what the panel last said.
type advanceUI struct {
	sel       int    // the picked pillar
	flash     string // a line for the foot, in place of the status
	flashGood bool
}

const (
	advMinW = 60
	advMinH = 14
)

// renderAdvance draws the panel w by h.
func renderAdvance(v game.AdvanceView, u advanceUI, w, h int) *xgrid {
	g := newXGrid(w, h)
	title := vibe.Line("ui.advance_view.renderAdvance.01", "Advance · ") + v.AgeName
	if !v.Final {
		title += vibe.Line("ui.advance_view.renderAdvance.02", " to ") + v.NextName
	}
	g.box(0, 0, w, h, xBorder, title, xTitle, false)
	if w < advMinW || h < advMinH {
		g.put(2, 2, fitText(vibe.Line("ui.advance_view.renderAdvance.03", "The window is too small for the Advance panel. Esc closes it."), w-4), xBad)
		return g
	}
	u.sel = clamp(u.sel, 0, max(0, len(v.Pillars)-1))

	// The line under the title: what the panel is, and how many pillars stand.
	if v.Final {
		g.put(2, 1, vibe.Line("ui.advance_view.renderAdvance.04", "The final age"), xDim)
	} else {
		g.putFit(2, 1, w-22, vibe.Line("ui.advance_view.renderAdvance.05", "What stands between you and the ")+v.NextName, xDim)
		k := xText
		if v.AllStand {
			k = xGood
		}
		g.rput(w-3, 1, vibe.Sprintf("ui.advance_view.renderAdvance.06", "%d of %d stand", v.Stand, len(v.Pillars)), k)
	}

	mh := 6
	if h < 22 {
		mh = 5
	}
	advMarch(g, v, w, mh)

	y0 := 2 + mh
	bh := h - 3 - y0 // the body, down to the row above the status line
	rw := clamp(w*37/100, 30, 50)
	lw := w - 3 - rw
	if v.Final {
		advCharter(g, v, u, 1, y0, w-2, bh)
	} else {
		advCharter(g, v, u, 1, y0, lw, bh)
		advRight(g, v, lw+2, y0, rw, bh)
	}
	advFoot(g, v, u, w, h)
	return g
}

// ===== The march of ages =====

// marchName is an age's name on the track: "Bronze Age" is "Bronze".
func marchName(name string) string { return strings.TrimSuffix(name, " Age") }

// advMarch draws the march of ages box, mh rows tall: the ages as stops on a
// track, this one and the next lit, the town between them where the gate has
// brought it, and the year.
func advMarch(g *xgrid, v game.AdvanceView, w, mh int) {
	g.box(1, 2, w-2, mh, xBorder, vibe.Line("ui.advance_view.advMarch.01", "The march of ages"), xTitle, false)
	total := len(v.March)
	if total == 0 {
		return
	}
	const x0, lead = 9, 13
	slots := min(total, max(2, (w-x0-10)/lead+1))
	step := lead
	if slots > 1 {
		step = (w - x0 - 10) / (slots - 1)
	}
	before := min(v.Here, (slots-1)/2)
	first := max(0, min(v.Here-before, total-slots))

	// Where the town is between this age and the next, in cells.
	c := 0
	if !v.Final {
		c = clamp(int(v.Share*float64(step)+0.5), 0, step)
	}

	const nameY, trackY = 3, 4
	yearY, sumY := 5, 6
	if mh < 6 {
		yearY, sumY = -1, 5
	}
	for i := 0; i < slots; i++ {
		a := first + i
		x := x0 + i*step
		stop := v.March[a]

		// The track to the right of the stop.
		if i < slots-1 {
			for k := 1; k < step; k++ {
				r, kind := '─', xFaint
				switch {
				case a < v.Here:
					r, kind = '━', xGood
				case a == v.Here && !v.Final && k < c:
					r, kind = '━', xAccent
				}
				g.set(x+k, trackY, r, kind)
			}
			if a == v.Here && !v.Final && c > 0 && c < step {
				g.set(x+c, trackY, '◉', xBright)
			}
		}

		// The stop.
		mark, kind := '○', xFaint
		switch {
		case a < v.Here:
			mark, kind = '●', xGood
		case a == v.Here:
			mark, kind = '●', xAccent
			if c == 0 || v.Final {
				mark, kind = '◉', xBright
			}
		case a == v.Here+1:
			mark, kind = '○', xBright
			if c == step {
				mark = '◉'
			}
		}
		g.set(x, trackY, mark, kind)

		// Its name and its year, centred on it.
		name, nk := marchName(stop.Name), xDim
		switch {
		case stop.Name == "":
			name, nk = "?", xFaint
		case a < v.Here:
			nk = xText
		case a == v.Here:
			nk = xAccent
		case a == v.Here+1:
			nk = xBright
		}
		name = fitText(name, step-1)
		g.put(max(2, x-runeLen(name)/2), nameY, name, nk)
		if yearY > 0 && stop.Year != "" {
			yk := xDim
			if a > v.Here+1 {
				yk = xFaint
			}
			year := fitText(stop.Year, step-1)
			g.put(max(2, x-runeLen(year)/2), yearY, year, yk)
		}
	}
	if first > 0 {
		g.set(2, trackY, '◂', xDim)
	}
	if first+slots < total {
		g.set(w-4, trackY, '▸', xDim)
	}

	// The year, and how far the gate has brought the town.
	var parts []xseg
	if v.Final {
		parts = []xseg{{vibe.Line("ui.advance_view.advMarch.02", "The final age"), xText}, {" · ", xFaint}, {v.YearText, xBright}}
	} else {
		parts = []xseg{{vibe.Line("ui.advance_view.advMarch.03", "Year "), xDim}, {v.YearText, xBright},
			{vibe.Sprintf("ui.advance_view.advMarch.04", " · %d%% of the way to the %s", sharePercent(v.Share, v.Share >= 1), v.NextName), xText}}
	}
	end := g.segsFit(3, sumY, w-8, parts...)
	legend := vibe.Line("ui.advance_view.advMarch.05", "● passed  ◉ now  ○ ahead")
	if start := w - 4 - runeLen(legend); start > end+2 {
		g.put(start, sumY, legend, xFaint)
	}
}

// sharePercent is a share in whole percent, floored, and 100 only when done.
func sharePercent(p float64, done bool) int {
	if done {
		return 100
	}
	return clamp(int(p*100), 0, 99)
}

// ===== The Charter =====

// advCharter draws the Charter box: the four pillars, and the picked one in
// detail below them.
func advCharter(g *xgrid, v game.AdvanceView, u advanceUI, x, y, bw, bh int) {
	g.box(x, y, bw, bh, xBorder, vibe.Line("ui.advance_view.advCharter.01", "The Charter"), xTitle, false)
	if bh < 4 {
		return
	}
	iw := bw - 4
	if v.Final {
		g.putFit(x+2, y+2, iw, vibe.Line("ui.advance_view.advCharter.02", "This is the final age. There is no age beyond it."), xText)
		return
	}
	// Three rows to a pillar (name, bar, a blank) when the box is tall, two
	// when it is not quite, one when it is short; the detail takes the rest.
	pitch := 1
	switch inner := bh - 2; {
	case inner >= 20:
		pitch = 3
	case inner >= 13:
		pitch = 2
	}
	top := y + 1
	for i, p := range v.Pillars {
		advPillar(g, p, i == u.sel, x+2, top+i*pitch, iw, pitch > 1)
	}
	dy := top + len(v.Pillars)*pitch
	if pitch < 3 {
		dy++ // a blank row before the detail
	}
	if dy >= y+bh-1 || u.sel >= len(v.Pillars) {
		return
	}
	advDetail(g, v.Pillars[u.sel], x+2, dy, iw, y+bh-1-dy)
}

// pillarNote is the short line under a pillar's name: how many of its
// requirements stand.
func pillarNote(p game.GatePillar) string {
	if len(p.Rows) == 0 {
		return vibe.Line("ui.advance_view.pillarNote.01", "nothing needed this age")
	}
	met := 0
	for _, r := range p.Rows {
		if r.Met {
			met++
		}
	}
	return vibe.Sprintf("ui.advance_view.pillarNote.02", "%d of %d met", met, len(p.Rows))
}

// pillarStatus is what a pillar's bar ends with: "✓ stands" or its percent.
func pillarStatus(p game.GatePillar) (string, xk) {
	if p.Met {
		return vibe.Line("ui.advance_view.pillarStatus.01", "✓ stands"), xGood
	}
	return fmt.Sprintf("%d%%", sharePercent(p.Progress, false)), xText
}

// advBar draws a bar of w cells filled to share.
func advBar(g *xgrid, x, y, w int, share float64, k xk) {
	fill := clamp(int(share*float64(w)+0.5), 0, w)
	for i := 0; i < w; i++ {
		if i < fill {
			g.set(x+i, y, '█', k)
		} else {
			g.set(x+i, y, '░', xFaint)
		}
	}
}

// advPillar draws one pillar at (x, y), iw cells wide: two rows when roomy
// (name and note, bar and status), one when not.
func advPillar(g *xgrid, p game.GatePillar, picked bool, x, y, iw int, roomy bool) {
	mark, nk := "  ", xAccent
	switch {
	case picked:
		mark, nk = "▸ ", xSel
	case p.Met:
		nk = xGood
	}
	name := strings.ToUpper(p.Name)
	for runeLen(name) < 10 {
		name += " "
	}
	barK := xTitle
	if p.Met {
		barK = xGood
	}
	status, sk := pillarStatus(p)
	note := pillarNote(p)
	if roomy {
		g.put(x, y, mark+name, nk)
		g.putFit(x+14, y, iw-14, note, xDim)
		bw := clamp(iw-2-len([]rune(vibe.Line("ui.advance_view.advPillar.01", "✓ stands")))-1, 10, 48)
		advBar(g, x+2, y+1, bw, p.Progress, barK)
		g.put(x+2+bw+1, y+1, status, sk)
		return
	}
	g.put(x, y, mark+name, nk)
	const nameW = 12
	room := iw - nameW
	bw := clamp((room-len([]rune(vibe.Line("ui.advance_view.advPillar.02", "✓ stands")))-1)/2, 6, 24)
	advBar(g, x+nameW, y, bw, p.Progress, barK)
	sx := g.put(x+nameW+bw+1, y, status, sk)
	if left := x + iw - sx - 1; left >= 8 {
		g.putFit(sx+1, y, left, note, xDim)
	}
}

// advDetail lists a pillar's requirements in rows rows from (x, y): a mark,
// what it is, and where it stands. When they do not all fit, the ones still
// to do come first and the rest are counted.
func advDetail(g *xgrid, p game.GatePillar, x, y, iw, rows int) {
	if rows < 1 {
		return
	}
	g.putFit(x, y, iw, p.Name+vibe.Line("ui.advance_view.advDetail.01", " in detail"), xAccent)
	rows--
	y++
	if len(p.Rows) == 0 {
		if rows > 0 {
			g.put(x, y, vibe.Line("ui.advance_view.advDetail.02", "Nothing needed this age."), xDim)
		}
		return
	}
	show := p.Rows
	more := 0
	if len(show) > rows {
		var open, done []game.GateRow
		for _, r := range p.Rows {
			if r.Met {
				done = append(done, r)
			} else {
				open = append(open, r)
			}
		}
		keep := max(0, rows-1) // the last row says how many are left out
		show = append(open, done...)
		if len(show) > keep {
			more = len(show) - keep
			show = show[:keep]
		}
	}
	for i, r := range show {
		advRow(g, r, x, y+i, iw)
	}
	y += len(show)
	rows -= len(show)
	if more > 0 && rows > 0 {
		g.put(x, y, vibe.Sprintf("ui.advance_view.advDetail.03", "+%d more", more), xDim)
		return
	}
	for _, f := range p.Facts {
		if rows < 1 {
			return
		}
		g.putFit(x, y, iw, f, xDim)
		y++
		rows--
	}
}

// advRow draws one requirement: ✓ or ✗, its label, and its count or note at
// the right edge. Green when met.
func advRow(g *xgrid, r game.GateRow, x, y, iw int) {
	mark, mk, tk := "✗", xBad, xText
	if r.Met {
		mark, mk, tk = "✓", xGood, xGood
	}
	var val string
	switch r.Kind {
	case game.GateRowCount:
		val = vibe.Sprintf("ui.advance_view.advRow.01", "%s of %s", num(r.Have), num(r.Need))
	case game.GateRowShare:
		val = fmt.Sprintf("%d%%", r.Have)
	default:
		val = r.Note
	}
	val = fitText(val, max(0, iw-4-6))
	g.put(x, y, mark, mk)
	g.putFit(x+2, y, iw-2-runeLen(val)-1, r.Label, tk)
	g.rput(x+iw-1, y, val, tk)
}

// ===== The right column =====

// advRight draws the boxes on the right: what the next age brings, and
// under it what else stands in the way when there is room.
func advRight(g *xgrid, v game.AdvanceView, x, y, bw, bh int) {
	vw := bw - 4 - 11
	type brow struct {
		label string
		lines []string
	}
	b := v.Brings
	buildings := vibe.Line("ui.advance_view.advRight.01", "none new")
	if b.Buildings > 0 {
		buildings = vibe.Sprintf("ui.advance_view.advRight.02", "%d new", b.Buildings)
	}
	resources := vibe.Line("ui.advance_view.advRight.03", "none new")
	if len(b.Resources) > 0 {
		resources = strings.Join(b.Resources, ", ")
	}
	resLines := []string{resources}
	if len(b.Resources) > 0 {
		resLines = wrapItems(b.Resources, vw)
	}
	rows := []brow{
		{vibe.Line("ui.advance_view.advRight.04", "Buildings"), []string{buildings}},
		{vibe.Line("ui.advance_view.advRight.05", "Resources"), resLines},
		{vibe.Line("ui.advance_view.advRight.06", "People"), wrapItems(b.People, vw)},
	}
	if b.Wonder != "" {
		rows = append(rows, brow{vibe.Line("ui.advance_view.advRight.07", "Wonder"), wrapText(b.Wonder, vw)})
	}
	need := 2
	for _, r := range rows {
		need += len(r.lines)
	}
	// What stands in the way gets the room under it, as much as it says and
	// no more, when there are four rows to spare for it.
	wayH := 0
	if spare := bh - need; spare >= 4 {
		wayH = min(spare, 2+max(1, wayLines(v, bw-4)))
	}
	boxH := need
	if wayH == 0 {
		boxH = bh
	}

	title := vibe.Line("ui.advance_view.advRight.08", "The ") + v.NextName + vibe.Line("ui.advance_view.advRight.09", " brings")
	if runeLen(title) > bw-6 {
		title = v.NextName + vibe.Line("ui.advance_view.advRight.10", " brings")
	}
	if runeLen(title) > bw-6 {
		title = vibe.Line("ui.advance_view.advRight.11", "Brings")
	}
	g.box(x, y, bw, boxH, xBorder, title, xTitle, false)
	ly := y + 1
	for _, r := range rows {
		for i, l := range r.lines {
			if ly >= y+boxH-1 {
				break
			}
			if i == 0 {
				g.put(x+2, ly, r.label, xDim)
			}
			g.putFit(x+2+11, ly, vw, l, xText)
			ly++
		}
	}
	if wayH > 0 {
		advWay(g, v, x, y+boxH, bw, wayH)
	}
}

// wrapItems breaks a list of items into lines of at most w cells, comma
// separated, never inside an item.
func wrapItems(items []string, w int) []string {
	var out []string
	line := ""
	for i, it := range items {
		if i < len(items)-1 {
			it += ","
		}
		switch {
		case line == "":
			line = it
		case runeLen(line)+1+runeLen(it) > w:
			out = append(out, line)
			line = it
		default:
			line += " " + it
		}
	}
	if line != "" {
		out = append(out, line)
	}
	return out
}

// wayLines is how many rows what stands in the way takes in w cells.
func wayLines(v game.AdvanceView, w int) int {
	n := 0
	for _, t := range append(append([]string(nil), v.Blockers...), v.Notices...) {
		n += len(wrapText(t, w))
	}
	return n
}

// advWay draws what else stands in the way of the advance.
func advWay(g *xgrid, v game.AdvanceView, x, y, bw, bh int) {
	g.box(x, y, bw, bh, xBorder, vibe.Line("ui.advance_view.advWay.01", "In the way"), xTitle, false)
	iw := bw - 4
	ly := y + 1
	put := func(lines []string, k xk) {
		for _, text := range lines {
			for _, l := range wrapText(text, iw) {
				if ly >= y+bh-1 {
					return
				}
				g.put(x+2, ly, l, k)
				ly++
			}
		}
	}
	put(v.Blockers, xBad)
	put(v.Notices, xAccent)
	if len(v.Blockers)+len(v.Notices) == 0 {
		g.putFit(x+2, ly, iw, vibe.Line("ui.advance_view.advWay.02", "Nothing else stands in the way."), xDim)
	}
}

// ===== The foot =====

// advFoot writes the status line and the key bar: the Enter prompt when the
// gate is met, what is still to do when it is not, or the last thing the
// panel was asked to do.
func advFoot(g *xgrid, v game.AdvanceView, u advanceUI, w, h int) {
	y := h - 3
	room := w - 4
	switch {
	case u.flash != "":
		k := xBad
		if u.flashGood {
			k = xGood
		}
		g.putFit(2, y, room, u.flash, k)
	case v.Final:
		if len(v.Blockers) > 0 {
			g.putFit(2, y, room, v.Blockers[0], xDim)
		}
	case v.Ready:
		x := g.put(2, y, " Enter ", xSel)
		x = g.put(x+1, y, vibe.Line("ui.advance_view.advFoot.02", "Advance to the ")+v.NextName, xBright)
		if len(v.Notices) > 0 && x+3 < w-2 {
			g.putFit(x+2, y, w-3-x-2, v.Notices[0], xAccent)
		}
	case v.AllStand && len(v.Blockers) > 0:
		g.putFit(2, y, room, v.Blockers[0], xBad)
	default:
		var open []string
		for _, p := range v.Pillars {
			if !p.Met {
				open = append(open, p.Name)
			}
		}
		x := g.put(2, y, vibe.Line("ui.advance_view.advFoot.03", "Not yet: "), xBad)
		g.putFit(x, y, w-3-x, textfmt.List(open)+vibe.Line("ui.advance_view.advFoot.04", " still to stand."), xText)
	}
	short := make([][2]string, 0, len(advanceKeys))
	for _, k := range advanceKeys {
		if v.Final && k.key != "Esc" {
			continue // nothing to pick and nothing to advance to
		}
		short = append(short, [2]string{shortKeyName(k.key), shortKeyUse(k.key)})
	}
	hint := "advance"
	if v.Final {
		hint = ""
	}
	expKeyBar(g, expLayout{w: w, h: h, footY: h - 2}, short, hint)
}

// shortKeyName and shortKeyUse are the key bar's wording for the panel's
// keys: the Help panel's rows are longer.
func shortKeyName(key string) string {
	if key == "Up/Down" {
		return "↑↓"
	}
	return key
}

func shortKeyUse(key string) string {
	switch key {
	case "Up/Down":
		return vibe.Line("ui.advance_view.shortKeyUse.01", "pick a pillar")
	case "Enter":
		return "advance"
	}
	return "close"
}
