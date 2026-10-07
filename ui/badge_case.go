package ui

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/theme"
)

// badge_case.go draws the badge case: the account's badges as a grid of
// small badges under family tabs, with the selected one at full size
// beside it.
//
// Everything here is a pure function of the badge list, a size, a view
// (the tab, the selection, how far the grid is scrolled) and a frame
// number. buildCase lays the badges out; renderCase draws the panel into a
// cell grid; the panel (badge_panel.go) keeps the view and paints the
// cells. The badges arrive with the spoiler rules applied (game.BadgeView):
// a withheld badge has no name, no description and no emblem in it, so
// nothing here decides what to hide. What it does decide is not to count
// the withheld ones out loud: while the account has not seen the last age
// they are drawn as one slab a family, never one each.

// The tabs that are not a family.
const (
	caseTabAll  = ""
	caseTabNext = "next"
)

// The case's geometry.
const (
	caseTop    = 3         // the body's first row: under the title bar, the tabs and the count line
	caseFoot   = 2         // the status line and the key bar
	casePitchX = miniW + 1 // a mini and the gap after it
	casePitchY = miniH + 1
	caseMinW   = 30
	caseMinH   = 10
	// caseSideW is the least width that puts the selected badge beside the
	// grid; under it the badge goes below. caseListW is the least width
	// that draws a grid at all; under it the badges are a list of names.
	caseSideW = 80
	caseListW = 60
)

// caseView is what the panel shows of the case.
type caseView struct {
	tab string // caseTabAll, a family's key or caseTabNext
	sel string // the selected item's id
	vy  int    // rows the grid is scrolled by
	// card: the selected badge's card is open over the grid.
	card   bool
	tier   mapmodel.GlyphTier
	prompt bool   // something is typed in the command bar
	note   string // a line for the status row: what a command found
}

func (v caseView) plain() bool { return v.tier == mapmodel.TierASCII }

// caseLayout is where the case's parts sit at a size.
type caseLayout struct {
	w, h int
	// list: the badges are a list of names, not a grid.
	list bool
	cols int // minis a line
	// The grid's box, and the selected badge's: beside the grid (side) or
	// under it.
	gx, gy, gw, gh int
	side           bool
	px, py, pw, ph int
}

func caseLayoutFor(w, h int, tab string) caseLayout {
	l := caseLayout{w: w, h: h, gx: 1, gy: caseTop}
	body := max(h-caseTop-caseFoot, 0)
	l.list = w < caseListW || tab == caseTabNext
	if w >= caseSideW {
		l.side = true
		l.pw = min(46, max(33, w*3/8))
		l.gw = w - 4 - l.pw
		l.gh = body
		l.px, l.py, l.ph = l.gx+l.gw+3, caseTop, body
	} else {
		l.gw = max(w-2, 0)
		l.ph = min(4, max(body-casePitchY, 0))
		l.gh = body - l.ph
		l.px, l.py, l.pw = 1, caseTop+l.gh, l.gw
	}
	l.cols = max((l.gw+1)/casePitchX, 1)
	return l
}

// caseItem is one thing the selection can sit on: a badge, or a family's
// withheld badges as one.
type caseItem struct {
	id     string
	v      game.BadgeView
	group  bool
	family string
	line   int
	col    int
}

type caseLineKind uint8

const (
	lineHeading caseLineKind = iota // a family's name and count
	lineBadges                      // a line of minis
	lineRow                         // one badge, in a list
)

// caseLine is one line of the grid's flow.
type caseLine struct {
	kind  caseLineKind
	y     int // rows from the top of the flow
	title string
	count string
	items []int
	// A ladder's line carries its name and the first rung not earned
	// (nil when every rung is).
	ladder string
	next   *game.BadgeView
}

func (l caseLine) height() int {
	if l.kind == lineBadges {
		return casePitchY
	}
	return 1
}

// need is the rows of the line that must show for it to be drawn: a line
// of minis can lose the gap under it.
func (l caseLine) need() int {
	if l.kind == lineBadges {
		return miniH
	}
	return 1
}

// caseTab is a tab and its family's counts.
type caseTab struct {
	key, title            string
	earned, shown, hidden int
}

// caseModel is the case laid out for one tab at one size.
type caseModel struct {
	account bool
	sum     game.BadgeSummary
	tabs    []caseTab
	items   []caseItem
	lines   []caseLine
	by      map[string]int
	rows    int
	lay     caseLayout
	// rarity is the earned badges by rarity, common first.
	rarity [5]int
	// closest is the badge nearest its count, of those not earned.
	closest *game.BadgeView
}

// caseRarities are the rarities the count line lists, in order.
var caseRarities = [5]string{
	config.BadgeCommon.Name(), config.BadgeUncommon.Name(), config.BadgeRare.Name(),
	config.BadgeEpic.Name(), config.BadgeMythic.Name(),
}

// familyTitle is a family's heading.
func familyTitle(key string) string {
	for _, f := range badgeFamilyTitles {
		if f.key == key {
			return f.title
		}
	}
	return textfmt.Capitalize(strings.ReplaceAll(key, "_", " "))
}

// caseTabs lists the tabs: All, each family that has a badge to list in
// the families' own order, and Next.
func caseTabs(views []game.BadgeView) []caseTab {
	stat := map[string]*caseTab{}
	var seen []string
	for _, v := range views {
		t := stat[v.Family]
		if t == nil {
			t = &caseTab{key: v.Family, title: familyTitle(v.Family)}
			stat[v.Family] = t
			seen = append(seen, v.Family)
		}
		switch {
		case v.Integrity:
		case v.Earned:
			t.earned++
			t.shown++
		case v.Hidden:
			t.hidden++
		default:
			t.shown++
		}
	}
	tabs := []caseTab{{key: caseTabAll, title: "All"}}
	listed := map[string]bool{}
	add := func(key string) {
		if t := stat[key]; t != nil && !listed[key] {
			listed[key] = true
			tabs = append(tabs, *t)
		}
	}
	for _, f := range badgeFamilyTitles {
		add(f.key)
	}
	for _, key := range seen {
		add(key)
	}
	return append(tabs, caseTab{key: caseTabNext, title: "Next"})
}

// groupID is the id of a family's withheld badges.
func groupID(family string) string { return "hidden." + family }

// buildCase lays the badges out for a tab at a size.
func buildCase(views []game.BadgeView, sum game.BadgeSummary, account bool, tab string, lay caseLayout) *caseModel {
	m := &caseModel{account: account, sum: sum, tabs: caseTabs(views), lay: lay, by: map[string]int{}}
	known := false
	for _, t := range m.tabs {
		known = known || t.key == tab
	}
	if !known {
		tab = caseTabAll
	}
	best := -1.0
	for i := range views {
		v := &views[i]
		if v.Earned && !v.Integrity {
			for r, name := range caseRarities {
				if v.Rarity == name {
					m.rarity[r]++
				}
			}
		}
		if !v.Earned && !v.Hidden && v.Target > 0 {
			if f := v.Progress / v.Target; f > best {
				best, m.closest = f, v
			}
		}
	}

	y := 0
	addLine := func(l caseLine) int {
		l.y = y
		y += l.height()
		m.lines = append(m.lines, l)
		return len(m.lines) - 1
	}
	addItem := func(it caseItem, line int) {
		it.line = line
		it.col = len(m.lines[line].items)
		m.by[it.id] = len(m.items)
		m.lines[line].items = append(m.lines[line].items, len(m.items))
		m.items = append(m.items, it)
	}

	if tab == caseTabNext {
		// The rungs nearest their count, nearest first.
		var near []*game.BadgeView
		for i := range views {
			if v := &views[i]; !v.Earned && !v.Hidden && v.Target > 0 {
				near = append(near, v)
			}
		}
		sort.SliceStable(near, func(a, b int) bool {
			return near[a].Progress/near[a].Target > near[b].Progress/near[b].Target
		})
		for _, v := range near {
			addItem(caseItem{id: v.Key, v: *v, family: v.Family}, addLine(caseLine{kind: lineRow}))
		}
		m.rows = y
		return m
	}

	for _, t := range m.tabs {
		if t.key == caseTabAll || t.key == caseTabNext || (tab != caseTabAll && tab != t.key) {
			continue
		}
		if tab == caseTabAll {
			addLine(caseLine{kind: lineHeading, title: t.title, count: familyCount(t, sum, "")})
		}
		loose := -1 // the line loose badges are filling
		place := func(it caseItem) {
			if lay.list {
				addItem(it, addLine(caseLine{kind: lineRow}))
				return
			}
			if loose < 0 || len(m.lines[loose].items) >= lay.cols {
				loose = addLine(caseLine{kind: lineBadges})
			}
			addItem(it, loose)
		}
		withheld := 0
		for i := 0; i < len(views); i++ {
			v := views[i]
			if v.Family != t.key {
				continue
			}
			if v.Hidden && !v.Secret && !sum.HiddenCounted {
				withheld++
				continue
			}
			if v.Ladder == "" || v.Hidden || lay.list {
				place(caseItem{id: v.Key, v: v, family: t.key})
				continue
			}
			// A ladder: its rungs on a line of their own, in order.
			var rungs []game.BadgeView
			j := i
			for ; j < len(views) && views[j].Family == t.key && views[j].Ladder == v.Ladder && !views[j].Hidden; j++ {
				rungs = append(rungs, views[j])
			}
			i = j - 1
			var next *game.BadgeView
			for k := range rungs {
				if !rungs[k].Earned {
					next = &rungs[k]
					break
				}
			}
			line := -1
			for _, r := range rungs {
				if line < 0 || len(m.lines[line].items) >= lay.cols {
					line = addLine(caseLine{kind: lineBadges, ladder: v.Ladder, next: next})
				}
				addItem(caseItem{id: r.Key, v: r, family: t.key}, line)
			}
			loose = -1
		}
		if withheld > 0 {
			place(caseItem{id: groupID(t.key), group: true, family: t.key, v: game.BadgeView{Family: t.key, Name: game.BadgeHiddenName, Hidden: true}})
		}
	}
	m.rows = y
	return m
}

// hiddenText is how many badges are withheld, in words: a number only
// once the account may know it. "" for none.
func hiddenText(n int, sum game.BadgeSummary) string {
	switch {
	case n <= 0:
		return ""
	case sum.HiddenCounted:
		return wholeNumber(n) + " hidden"
	}
	return game.BadgeHiddenName + " hidden"
}

// familyCount is a family's count: "2 of 3", then its withheld badges.
// earned is the word after the count, if any ("2 of 3 earned").
func familyCount(t caseTab, sum game.BadgeSummary, earned string) string {
	s := fmt.Sprintf("%d of %d%s", t.earned, t.shown, earned)
	if hidden := hiddenText(t.hidden, sum); hidden != "" {
		s += " · " + hidden
	}
	return s
}

// home is the item the case opens on: the first of the tab.
func (m *caseModel) home() string {
	if len(m.items) == 0 {
		return ""
	}
	return m.items[0].id
}

func (m *caseModel) selected(v caseView) *caseItem {
	if i, ok := m.by[v.sel]; ok {
		return &m.items[i]
	}
	return nil
}

// follow scrolls the grid so the selection's line shows, with its
// family's heading over it when that fits.
func (m *caseModel) follow(v *caseView) {
	gh := m.lay.gh
	if it := m.selected(*v); it != nil && gh > 0 {
		ln := m.lines[it.line]
		top := ln.y
		if it.line > 0 && m.lines[it.line-1].kind == lineHeading && ln.need()+1 <= gh {
			top--
		}
		if top < v.vy {
			v.vy = top
		}
		if end := ln.y + ln.need(); end > v.vy+gh {
			v.vy = end - gh
		}
	}
	v.vy = max(min(v.vy, m.rows-gh), 0)
	// Start the view on a line, not part way down one: a line that does
	// not show whole is not drawn, and the rows it leaves belong at the
	// foot of the grid, not the head.
	for _, ln := range m.lines {
		if ln.y < v.vy {
			continue
		}
		if it := m.selected(*v); it == nil || ln.y <= m.lines[it.line].y {
			v.vy = ln.y
		}
		break
	}
}

// visible reports whether a line is drawn at a scroll: whole, or not at
// all. A badge is never cut.
func (m *caseModel) visible(l caseLine, vy int) bool {
	return l.y >= vy && l.y+l.need() <= vy+m.lay.gh
}

// shows is visible for line i, and for a heading asks that the line under
// it shows too: a family's name is not left alone at the foot.
func (m *caseModel) shows(i, vy int) bool {
	if !m.visible(m.lines[i], vy) {
		return false
	}
	if m.lines[i].kind == lineHeading && i+1 < len(m.lines) {
		return m.visible(m.lines[i+1], vy)
	}
	return true
}

// ----- text -----

// seg is a run of text in one style, or in an ink.
type seg struct {
	text string
	st   tStyle
	ink  ink
	bold bool
}

func (g *tGrid) segs(x, y, room int, line []seg) int {
	for _, s := range line {
		for _, r := range s.text {
			if room <= 0 {
				return x
			}
			fl := uint8(0)
			if s.bold {
				fl = cfBold
			}
			if s.ink != inkNone {
				g.paint(x, y, r, s.ink, s.st, fl)
			} else {
				g.put(x, y, r, s.st, -1)
				g.flag(x, y, fl)
			}
			x++
			room--
		}
	}
	return x
}

func segsLen(line []seg) int {
	n := 0
	for _, s := range line {
		n += runeLen(s.text)
	}
	return n
}

// wrapCells wraps s to lines of at most w cells, on spaces. A word longer
// than a line is cut.
func wrapCells(s string, w int) []string {
	if w < 1 {
		return nil
	}
	var lines []string
	line := ""
	for _, word := range strings.Fields(s) {
		for runeLen(word) > w {
			if line != "" {
				lines = append(lines, line)
				line = ""
			}
			rs := []rune(word)
			lines = append(lines, string(rs[:w]))
			word = string(rs[w:])
		}
		switch {
		case line == "":
			line = word
		case runeLen(line)+1+runeLen(word) <= w:
			line += " " + word
		default:
			lines = append(lines, line)
			line = word
		}
	}
	if line != "" {
		lines = append(lines, line)
	}
	return lines
}

// tierInk is the ink a tier's name is written in.
func tierInk(t config.BadgeTier) ink {
	_, base, _ := tierInks(t)
	return base
}

// barSegs is a progress bar of w cells: the part done in an ink, the rest
// dim.
func barSegs(progress, target float64, w int, k ink) []seg {
	if w < 1 {
		return nil
	}
	done := 0
	if target > 0 {
		done = int(math.Floor(float64(w) * progress / target))
	}
	done = max(min(done, w), 0)
	if done == w && progress < target {
		done = w - 1 // full only when it is
	}
	return []seg{{text: strings.Repeat("█", done), st: tsHi, ink: k}, {text: strings.Repeat("░", w-done), st: tsDim}}
}

func countText(progress, target float64) string {
	return FormatNumber(progress) + " / " + FormatNumber(target)
}

// metaLine is a badge's tier, rarity and points on one line.
func metaLine(it *caseItem) []seg {
	v := it.v
	dot := seg{text: " · ", st: tsDim}
	switch {
	case it.group:
		return []seg{{text: "Hidden badges", st: tsDim}}
	case v.Hidden && v.Secret:
		return []seg{{text: "Secret", st: tsDim}}
	case v.Hidden:
		return []seg{{text: "Hidden", st: tsDim}}
	case v.Integrity:
		return []seg{{text: "Integrity", st: tsLabel}, dot, {text: "no points", st: tsDim}}
	}
	out := []seg{{text: textfmt.Capitalize(v.Tier), st: tsText, ink: tierInk(v.Level), bold: true}}
	if v.Rarity != "" {
		out = append(out, dot, seg{text: textfmt.Capitalize(v.Rarity), st: tsLabel})
	}
	return append(out, dot, seg{text: wholeNumber(v.Points), st: tsHi}, seg{text: " pts", st: tsDim})
}

// descText is what a badge says of itself, and what stands in for a
// withheld one.
func descText(it *caseItem) string {
	v := it.v
	switch {
	case it.group:
		return "More badges of this family show as you play on."
	case v.Hidden && v.Desc != "":
		return v.Desc
	case v.Hidden:
		return "Not revealed yet."
	}
	return v.Desc
}

// statusLines are the lines under a badge's description: when it was
// earned or how far it is, where it sits on its ladder and what it gives.
// bar is the width of a progress bar, 0 for none.
func statusLines(it *caseItem, room, bar int) [][]seg {
	v := it.v
	var out [][]seg
	switch {
	case it.group, v.Hidden:
	case v.Earned:
		line := []seg{{text: "Earned", st: tsGood}}
		if !v.At.IsZero() {
			line = append(line, seg{text: " " + v.At.Format("2 Jan 2006"), st: tsText})
		}
		if v.Run != "" {
			line = append(line, seg{text: " · run ", st: tsDim}, seg{text: v.Run, st: tsLabel})
		}
		out = append(out, line)
		if v.Crossed {
			for _, l := range wrapCells("Earned in a modified game: no points.", room) {
				out = append(out, []seg{{text: l, st: tsBad}})
			}
		}
	case v.Target > 0:
		line := barSegs(v.Progress, v.Target, bar, tierInk(v.Level))
		if bar > 0 {
			line = append(line, seg{text: " ", st: tsText})
		}
		out = append(out, append(line, seg{text: countText(v.Progress, v.Target), st: tsHi}))
	default:
		out = append(out, []seg{{text: "Not earned yet.", st: tsDim}})
	}
	if v.Ladder != "" && v.Rungs > 0 && !v.Hidden {
		out = append(out, []seg{{text: fmt.Sprintf("Rung %d of %d · %s", v.Rung, v.Rungs, v.Ladder), st: tsDim}})
	}
	if !v.Hidden {
		if v.RewardTheme != "" {
			name := v.RewardTheme
			if t, ok := theme.ByKey(v.RewardTheme); ok {
				name = t.Name
			}
			out = append(out, []seg{{text: "Gives the ", st: tsDim}, {text: name, st: tsHi}, {text: " theme.", st: tsDim}})
		}
		if v.RewardTitle != "" {
			out = append(out, []seg{{text: "Gives the title ", st: tsDim}, {text: v.RewardTitle, st: tsHi}, {text: ".", st: tsDim}})
		}
	}
	return out
}

// detailLines is everything the case says of a badge, wrapped to room: its
// name, its tier line, its description and its status. most is the most
// rows there are; when the whole does not fit, the description is cut
// short with an ellipsis, so the name and the status always show.
func detailLines(it *caseItem, room, most, bar int) [][]seg {
	var name [][]seg
	for _, l := range wrapCells(it.v.Name, room) {
		name = append(name, []seg{{text: l, st: tsBright, bold: true}})
	}
	if len(name) > 2 {
		name = name[:2]
	}
	head := append(name, metaLine(it))
	desc := wrapCells(descText(it), room)
	tail := statusLines(it, room, bar)
	if free := most - len(head) - len(tail); len(desc) > free {
		if free < 1 {
			desc = nil
		} else {
			desc = desc[:free]
			rs := []rune(desc[free-1])
			if len(rs) >= room {
				rs = rs[:max(room-1, 0)]
			}
			desc[free-1] = strings.TrimRight(string(rs), " ,.") + "…"
		}
	}
	out := head
	for _, l := range desc {
		out = append(out, []seg{{text: l, st: tsText}})
	}
	out = append(out, tail...)
	if len(out) > most {
		out = out[:max(most, 0)]
	}
	return out
}

// ----- drawing -----

// caseFrame is the frame number a badge of the case is drawn at: n for
// one that moves, and atRest for the rest and whenever n is atRest.
func caseFrame(n int, moves bool) int {
	if !moves {
		return atRest
	}
	return n
}

// drawItemMini draws an item's mini at (x, y).
func drawItemMini(g *tGrid, x, y int, it *caseItem, v caseView, n int) {
	md := medalOf(it.v, v.tier)
	drawMini(g, x, y, md, caseFrame(n, md.state == medalEarned && md.tier == config.BadgeLegendary))
	if it.group {
		// More than one, and how many is not said.
		g.put(x+miniW/2, y+miniH/2, '…', tsDim, -1)
	}
}

// renderCase draws the panel, w by h. n is the frame number, or atRest
// when nothing moves.
func renderCase(m *caseModel, v caseView, n int) *tGrid {
	w, h := m.lay.w, m.lay.h
	out := newTGrid(w, h)
	if w < caseMinW || h < caseMinH {
		out.text(0, 0, clipRunes("Badges: make the window larger.", w), tsText, -1)
		return out
	}
	for x := 0; x < w; x++ {
		out.put(x, 0, ' ', tsChip, -1)
		out.put(x, h-1, ' ', tsChip, -1)
	}
	m.drawTitle(out)
	m.drawKeys(out, v)
	if !m.account {
		out.text(1, caseTop, clipRunes("No account is loaded, so there are no badges to show.", w-2), tsDim, -1)
		return out
	}
	m.drawTabs(out, v)
	m.drawCounts(out, v)
	lay := m.lay
	sel := m.selected(v)

	// The grid.
	for i, ln := range m.lines {
		if !m.shows(i, v.vy) {
			continue
		}
		y := lay.gy + ln.y - v.vy
		switch ln.kind {
		case lineHeading:
			x := out.text(lay.gx, y, clipRunes(ln.title, lay.gw), tsBright, -1)
			x = out.text(x+2, y, clipRunes(ln.count, max(lay.gx+lay.gw-x-2, 0)), tsDim, -1)
			for x++; x < lay.gx+lay.gw; x++ {
				out.put(x, y, '─', tsDim, -1)
			}
		case lineBadges:
			for _, i := range ln.items {
				it := &m.items[i]
				x := lay.gx + it.col*casePitchX
				drawItemMini(out, x, y, it, v, n)
				if sel != nil && it.id == sel.id {
					for dy := 0; dy < miniH; dy++ {
						for dx := -1; dx <= miniW; dx++ {
							out.flag(x+dx, y+dy, cfSel)
						}
					}
					out.put(x-1, y+miniH/2, '▸', tsLane, -1)
					out.put(x+miniW, y+miniH/2, '◂', tsLane, -1)
					out.flag(x-1, y+miniH/2, cfSel|cfBold)
					out.flag(x+miniW, y+miniH/2, cfSel|cfBold)
				}
			}
			m.drawLadderLabel(out, ln, lay.gx+len(ln.items)*casePitchX+1, y)
		case lineRow:
			if len(ln.items) > 0 {
				it := &m.items[ln.items[0]]
				m.drawRow(out, it, y, sel != nil && it.id == sel.id, v)
			}
		}
	}
	if len(m.items) == 0 {
		text := "No badges to show here yet."
		if v.tab == caseTabNext {
			text = "No counted badge is in sight yet."
		}
		out.text(lay.gx, lay.gy, clipRunes(text, lay.gw), tsDim, -1)
	}
	// More above and below.
	edge := lay.gx + lay.gw
	if v.vy > 0 {
		out.put(edge, lay.gy, '▲', tsHi, -1)
	}
	if v.vy+lay.gh < m.rows {
		out.put(edge, lay.gy+lay.gh-1, '▼', tsHi, -1)
	}

	// The selected badge, beside the grid or under it.
	if lay.side {
		for y := 0; y < lay.gh; y++ {
			if c := out.at(edge+1, lay.gy+y); c.r == ' ' {
				out.put(edge+1, lay.gy+y, '│', tsBorder, -1)
			}
		}
	} else if lay.ph > 0 {
		for x := lay.gx; x < lay.gx+lay.gw; x++ {
			out.put(x, lay.py, '─', tsBorder, -1)
		}
	}
	if sel != nil {
		m.drawPane(out, sel, v, n)
	}

	// The status line.
	out.segs(1, h-caseFoot, w-2, m.statusLine(sel, v, w-2))

	if v.card && sel != nil {
		// The detail takes the body: one badge, and nothing behind it.
		for y := caseTop; y < h-caseFoot; y++ {
			for x := 0; x < w; x++ {
				out.put(x, y, ' ', tsText, -1)
			}
		}
		m.drawCard(out, sel, v, n)
	}
	return out
}

// drawTitle writes the title bar: the counts, the score and the title it
// holds.
func (m *caseModel) drawTitle(out *tGrid) {
	w := m.lay.w
	x := out.text(0, 0, " BADGES", tsChipKey, -1)
	if !m.account {
		return
	}
	sum := m.sum
	parts := []string{fmt.Sprintf("%d of %d", sum.Earned, sum.Shown), hiddenText(sum.Hidden, sum), wholeNumber(sum.Points) + " pts"}
	left := ""
	for _, p := range parts {
		if p != "" {
			left += " · " + p
		}
	}
	left = "  " + strings.TrimPrefix(left, " · ")
	right := sum.Title + " "
	if sum.NextTitle != "" && w >= 100 {
		right = fmt.Sprintf("%s · %s at %s ", sum.Title, sum.NextTitle, wholeNumber(sum.NextTitleAt))
	}
	if x+runeLen(left)+2+runeLen(right) > w {
		left = fmt.Sprintf("  %d/%d · %s pts", sum.Earned, sum.Shown, wholeNumber(sum.Points))
	}
	x = out.text(x, 0, clipRunes(left, max(w-x, 0)), tsChip, -1)
	if rx := w - runeLen(right); rx > x+1 {
		rx = out.text(rx, 0, sum.Title, tsChipKey, -1)
		out.text(rx, 0, strings.TrimPrefix(right, sum.Title), tsChipDim, -1)
	}
}

// drawTabs writes the tab row: the open tab in brackets, the row scrolled
// so it shows, and an arrow at an end that has more.
func (m *caseModel) drawTabs(out *tGrid, v caseView) {
	w := m.lay.w
	type span struct{ a, z int }
	spans := make([]span, len(m.tabs))
	x, open := 0, 0
	for i, t := range m.tabs {
		spans[i] = span{x, x + runeLen(t.title) + 2}
		x = spans[i].z + 1
		if t.key == v.tab {
			open = i
		}
	}
	room := w - 4
	off := 0
	if x-1 > room {
		off = min(max(spans[open].z-room, 0), spans[open].a)
	}
	for i, t := range m.tabs {
		a := 2 + spans[i].a - off
		if a < 2 || a+runeLen(t.title)+2 > w-2 {
			continue
		}
		if i == open {
			out.text(a, 1, "["+t.title+"]", tsLane, -1)
			for dx := 0; dx < runeLen(t.title)+2; dx++ {
				out.flag(a+dx, 1, cfBold)
			}
		} else {
			out.text(a+1, 1, t.title, tsDim, -1)
		}
	}
	if off > 0 {
		out.put(0, 1, '◂', tsHi, -1)
	}
	if x-1-off > room {
		out.put(w-1, 1, '▸', tsHi, -1)
	}
}

// drawCounts writes the line under the tabs: what the account holds by
// rarity on All, a family's count on its tab.
func (m *caseModel) drawCounts(out *tGrid, v caseView) {
	w := m.lay.w
	var line []seg
	switch v.tab {
	case caseTabAll:
		for i, name := range caseRarities {
			if i > 0 {
				line = append(line, seg{text: " · ", st: tsDim})
			}
			line = append(line, seg{text: name + " ", st: tsDim}, seg{text: wholeNumber(m.rarity[i]), st: tsHi})
		}
		if segsLen(line) > w-2 {
			line = line[:0]
			for i, name := range caseRarities {
				if i > 0 {
					line = append(line, seg{text: " ", st: tsDim})
				}
				line = append(line, seg{text: name[:1] + " ", st: tsDim}, seg{text: wholeNumber(m.rarity[i]), st: tsHi})
			}
		}
	case caseTabNext:
		line = []seg{{text: "The badges you are closest to, nearest first.", st: tsDim}}
	default:
		for _, t := range m.tabs {
			if t.key == v.tab {
				line = []seg{{text: familyCount(t, m.sum, " earned"), st: tsDim}}
			}
		}
	}
	out.segs(1, 2, w-2, line)
}

// drawLadderLabel writes a ladder's name and how far its next rung is, to
// the right of its rungs, when the line has the room.
func (m *caseModel) drawLadderLabel(out *tGrid, ln caseLine, x, y int) {
	room := m.lay.gx + m.lay.gw - x
	if ln.ladder == "" || room < 9 {
		return
	}
	out.text(x, y, clipRunes(ln.ladder, room), tsText, -1)
	if ln.next == nil {
		out.text(x, y+1, clipRunes("complete", room), tsGood, -1)
		return
	}
	out.text(x, y+1, clipRunes(countText(ln.next.Progress, ln.next.Target), room), tsHi, -1)
	out.segs(x, y+2, room, barSegs(ln.next.Progress, ln.next.Target, min(room, 12), tierInk(ln.next.Level)))
}

// rowMark is a badge's mark in a list: earned, in sight or withheld.
func rowMark(v game.BadgeView) seg {
	switch {
	case v.Hidden:
		return seg{text: "?", st: tsDim}
	case !v.Earned:
		return seg{text: "☆", st: tsDim}
	case v.Crossed:
		return seg{text: "╱", st: tsBad}
	case v.Integrity:
		return seg{text: "★", st: tsText, ink: inkCyan}
	}
	return seg{text: "★", st: tsText, ink: tierInk(v.Level), bold: true}
}

// drawRow writes one badge of a list: its mark, its name, its tier and,
// with the room, how far it is.
func (m *caseModel) drawRow(out *tGrid, it *caseItem, y int, selected bool, v caseView) {
	lay := m.lay
	bv := it.v
	tier := bv.Tier
	if bv.Integrity {
		tier = "integrity"
	}
	if bv.Hidden {
		tier = ""
	}
	var right []seg
	if tier != "" {
		right = append(right, seg{text: fmt.Sprintf("%-9s", tier), st: tsDim})
	}
	if bv.Target > 0 && !bv.Earned && !bv.Hidden {
		if lay.gw >= 54 {
			right = append(right, seg{text: " ", st: tsDim})
			right = append(right, barSegs(bv.Progress, bv.Target, 10, tierInk(bv.Level))...)
		}
		if lay.gw >= 40 {
			right = append(right, seg{text: fmt.Sprintf(" %13s", countText(bv.Progress, bv.Target)), st: tsHi})
		}
	}
	name := bv.Name
	if it.group {
		name = "Hidden badges"
	}
	nameRoom := lay.gw - 2 - segsLen(right) - 1
	if nameRoom < 8 {
		right, nameRoom = nil, lay.gw-2
	}
	out.segs(lay.gx, y, 1, []seg{rowMark(bv)})
	st := tsText
	if !bv.Earned {
		st = tsDim
	}
	out.text(lay.gx+2, y, clipRunes(name, nameRoom), st, -1)
	out.segs(lay.gx+lay.gw-segsLen(right), y, segsLen(right), right)
	if selected {
		out.put(lay.gx-1, y, '▸', tsLane, -1)
		for x := lay.gx - 1; x < lay.gx+lay.gw; x++ {
			out.flag(x, y, cfSel)
		}
	}
}

// drawPane draws the selected badge beside the grid: its full art, then
// what the case says of it. Under the grid (a narrow window) there is no
// room for the art: the lines alone.
func (m *caseModel) drawPane(out *tGrid, it *caseItem, v caseView, n int) {
	lay := m.lay
	if !lay.side {
		if lay.ph < 2 {
			return
		}
		for i, l := range detailLines(it, lay.pw, lay.ph-1, min(12, lay.pw/3)) {
			out.segs(lay.px, lay.py+1+i, lay.pw, l)
		}
		return
	}
	md := medalOf(it.v, v.tier)
	aw, ah := medalSize(md)
	y := lay.py
	bar := min(14, lay.pw/3)
	if ah+3 <= lay.ph && aw <= lay.pw {
		drawMedal(out, lay.px+(lay.pw-aw)/2, y, md, caseFrame(n, md.animated()), v.plain())
		y += ah
		if len(detailLines(it, lay.pw, 1<<20, bar)) < lay.ph-ah {
			y++ // a row of air under the art when there is one to spare
		}
	}
	for i, l := range detailLines(it, lay.pw, lay.py+lay.ph-y, bar) {
		out.segs(lay.px, y+i, lay.pw, l)
	}
}

// statusLine is the line over the key bar: how far the next rung of the
// selected badge's ladder is, or the badge nearest its count.
func (m *caseModel) statusLine(sel *caseItem, v caseView, room int) []seg {
	if v.note != "" {
		return []seg{{text: clipRunes(v.note, room), st: tsHi}}
	}
	label, target := "", (*game.BadgeView)(nil)
	if sel != nil && !sel.group && !sel.v.Hidden {
		switch {
		case sel.v.Ladder != "":
			label = "Next rung: "
			for i := range m.items {
				if o := &m.items[i]; o.v.Ladder == sel.v.Ladder && o.family == sel.family && !o.v.Earned && !o.v.Hidden {
					target = &o.v
					break
				}
			}
			if target == nil {
				return []seg{{text: clipRunes("Every rung of "+sel.v.Ladder+" is earned.", room), st: tsGood}}
			}
		case !sel.v.Earned && sel.v.Target > 0:
			label, target = "Progress: ", &sel.v
		}
	}
	if target == nil && m.closest != nil {
		label, target = "Closest: ", m.closest
	}
	if target == nil {
		return nil
	}
	count := countText(target.Progress, target.Target)
	bar := min(18, room-runeLen(label)-runeLen(target.Name)-runeLen(count)-4)
	line := []seg{{text: label, st: tsDim}, {text: target.Name, st: tsText}}
	if bar >= 6 {
		line = append(line, seg{text: "  ", st: tsText})
		line = append(line, barSegs(target.Progress, target.Target, bar, tierInk(target.Level))...)
	}
	return append(line, seg{text: "  " + count, st: tsHi})
}

// drawKeys writes the key bar: as many of the keys as the bar holds, and
// always how to close.
func (m *caseModel) drawKeys(out *tGrid, v caseView) {
	w, h := m.lay.w, m.lay.h
	move := [2]string{"←↑↓→", "move"}
	card := [2]string{"←→", "the badge before, the next"}
	if v.plain() {
		move[0], card[0] = "Arrows", "Left Right"
	}
	tab, enter, page, ends, esc := [2]string{"Tab", "family"}, [2]string{"Enter", "detail"}, [2]string{"PgUp PgDn", "page"}, [2]string{"Home End", "first, last"}, [2]string{"Esc", "close"}
	// Each set gives up one more key; the first that fits is drawn.
	sets := [][][2]string{
		{move, tab, enter, page, ends, esc},
		{move, tab, enter, page, esc},
		{move, tab, enter, esc},
		{tab, enter, esc},
		{enter, esc},
		{esc},
	}
	switch {
	case v.prompt:
		sets = [][][2]string{{{"Enter", "runs the command below"}, esc}, {esc}}
	case v.card:
		closes := [2]string{"Esc", "closes the detail"}
		sets = [][][2]string{{{"Enter", "closes the detail"}, card, closes}, {card, closes}, {closes}, {esc}}
	}
	width := func(keys [][2]string) int {
		n := 1
		for _, k := range keys {
			n += runeLen(k[0]) + 1 + runeLen(k[1]) + 2
		}
		return n - 2
	}
	keys := sets[len(sets)-1]
	for _, set := range sets {
		if width(set) <= w {
			keys = set
			break
		}
	}
	x := 1
	for _, k := range keys {
		x = out.text(x, h-1, k[0], tsChipKey, -1)
		x = out.text(x+1, h-1, k[1], tsChipDim, -1) + 2
	}
}

// drawCard draws the selected badge's detail in place of the grid: the
// badge at full size and everything the case says of it, nothing cut.
func (m *caseModel) drawCard(out *tGrid, it *caseItem, v caseView, n int) {
	lay := m.lay
	bw, bh := lay.w-2, lay.h-caseTop-caseFoot
	cw := min(bw-2, 76)
	if cw < 28 || bh < 6 {
		return
	}
	md := medalOf(it.v, v.tier)
	aw, ah := medalSize(md)
	art := cw >= aw+30 && bh >= ah+4
	tx, tw := 3, cw-6
	if art {
		tx, tw = aw+6, cw-aw-9
	}
	lines := detailLines(it, tw, bh-3, min(16, tw/2))
	ch := len(lines) + 4
	if art {
		ch = max(ch, ah+4)
	}
	ch = min(ch, bh)
	x0, y0 := 1+(bw-cw)/2, caseTop+(bh-ch)/2
	for y := 0; y < ch; y++ {
		for x := 0; x < cw; x++ {
			r := ' '
			switch {
			case y == 0 && x == 0:
				r = '╭'
			case y == 0 && x == cw-1:
				r = '╮'
			case y == ch-1 && x == 0:
				r = '╰'
			case y == ch-1 && x == cw-1:
				r = '╯'
			case y == 0 || y == ch-1:
				r = '─'
			case x == 0 || x == cw-1:
				r = '│'
			}
			out.put(x0+x, y0+y, r, tsBorder, -1)
		}
	}
	out.text(x0+2, y0, clipRunes(" "+it.v.Name+" ", cw-4), tsBright, -1)
	if esc := " Esc closes "; cw > runeLen(esc)+runeLen(it.v.Name)+8 {
		out.text(x0+cw-2-runeLen(esc), y0, esc, tsDim, -1)
	}
	if art {
		drawMedal(out, x0+3, y0+2+(ch-4-ah)/2, md, caseFrame(n, md.animated()), v.plain())
	}
	for i, l := range lines {
		if 2+i >= ch-1 {
			break
		}
		out.segs(x0+tx, y0+2+i, tw, l)
	}
}
