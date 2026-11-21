package ui

import (
	"image"
	"regexp"
	"sort"
	"strings"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/rowid"
)

// people_view.go draws the People panel: a pure function of the game's state
// and what the player has picked, into a grid the size of the panel (the
// Expeditions panel's cell grid and palette). Four tabs: Overview (the whole
// town at a glance), Work (the roster and the work areas), Army (soldiers
// and the buildings that make them) and Scholars (desks and research). The
// layout is worked out at 100 columns by 30 rows and shrinks to 80 by 24:
// boxes lose blank lines before they lose facts.

// peopleTabs are the tab names, in order.
var peopleTabs = []rowid.Entry{rowid.E("ui.people_view.peopleTabs.01", "OVERVIEW"), rowid.E("ui.people_view.peopleTabs.02", "WORK"), rowid.E("ui.people_view.peopleTabs.03", "ARMY"), rowid.E("ui.people_view.peopleTabs.04", "SCHOLARS")}

const (
	tabOverview = iota
	tabWork
	tabArmy
	tabScholars
)

// What a click on the panel can pick.
const (
	hitRow = iota // a card, a work area or a building
	hitSub        // a building of the picked work area
)

// peopleView is what the player has picked.
type peopleView struct {
	tab  int  // the tab
	row  int  // the picked row on the tab (a card, an area, a building)
	sub  int  // the picked building of the picked work area
	shut bool // the picked work area's detail is closed
}

// peopleKeys are the panel's keys as the Help panel lists them.
var peopleKeys = []keyHelp{
	{"Left/Right", rowid.E("ui.people_view.peopleKeys.01", "Switch between Overview, Work, Army and Scholars")},
	{"Up/Down", rowid.E("ui.people_view.peopleKeys.02", "Pick a card, a work area or a building")},
	{"A", rowid.E("ui.people_view.peopleKeys.05", "Work: put every area on auto")},
	{"B", rowid.E("ui.people_view.peopleKeys.06", "Work: open the buildings")},
	{"G", rowid.E("ui.people_view.peopleKeys.07", "Work: turn automatic growth on or off")},
	{"T", rowid.E("ui.people_view.peopleKeys.08", "Pick research")},
	{"Enter", rowid.E("ui.people_view.peopleKeys.09", "Work: open or close the picked area")},
	{"Esc", rowid.E("ui.people_view.peopleKeys.04", "Close the panel")},
}

// peopleRows is how many rows the tab has to pick from.
func peopleRows(st game.GameState, tab int) int {
	switch tab {
	case tabOverview:
		return 4
	case tabWork:
		return len(game.ShareRows(st))
	case tabArmy:
		return len(militaryBuildings(st))
	case tabScholars:
		return len(knowledgeBuildings(st))
	}
	return 0
}

// peopleBuilding is a built building as a list row.
type peopleBuilding struct {
	key string
	game.BuildingState
}

// builtWhere lists the built buildings that keep, by name.
func builtWhere(st game.GameState, keep func(game.BuildingState) bool) []peopleBuilding {
	var out []peopleBuilding
	for k, bs := range st.Buildings {
		if bs.Count > 0 && keep(bs) {
			out = append(out, peopleBuilding{k, bs})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Name != out[j].Name {
			return out[i].Name < out[j].Name
		}
		return out[i].key < out[j].key
	})
	return out
}

// militaryBuildings are the built buildings of the military category.
func militaryBuildings(st game.GameState) []peopleBuilding {
	return builtWhere(st, func(b game.BuildingState) bool { return b.Category == "military" })
}

// knowledgeBuildings are the built buildings that come with scholars.
func knowledgeBuildings(st game.GameState) []peopleBuilding {
	return builtWhere(st, func(b game.BuildingState) bool { return b.Scholars > 0 })
}

// renderPeople draws the panel w by h. The Overview comes back with its
// picture; the other tabs are cells alone.
func renderPeople(st game.GameState, v peopleView, buy buyChoice, hist *game.RunHistory, w, h int) (*xgrid, *pScene) {
	g := newXGrid(w, h)
	g.shell(rowid.Line("ui.people_view.renderPeople.01", "People · ")+st.AgeName, 0)
	if w < peopleMinW || h < peopleMinH {
		msg := fitText(rowid.Line("ui.people_view.renderPeople.02", "The window is too small for the People panel. Esc closes it."), w-4)
		g.put((w-runeLen(msg))/2, h/2, msg, xBad)
		return g, nil
	}
	v.tab = clamp(v.tab, 0, len(peopleTabs)-1)
	v.row = clamp(v.row, 0, peopleRows(st, v.tab)-1)
	peopleHeader(g, st, v)
	if v.tab == tabOverview {
		g.foot([][2]string{{"↑↓", "pick"}, {"Esc", "close"}}, "people")
		return g, newOverview(g, st, hist, v.row, 3).s
	}
	c := peopleBody{g: g, st: st, v: v, buy: buy, hist: hist, s: newPScene(w, 2, h-2), x: 1, y: 2, w: w - 2, h: h - 4}
	switch v.tab {
	case tabWork:
		c.work()
	case tabArmy:
		c.army()
	case tabScholars:
		c.scholars()
	}
	return g, c.s
}

// peopleHeader is the tab row and the headcount.
func peopleHeader(g *xgrid, st game.GameState, v peopleView) {
	x := 2
	squares := [...]xk{xGold, xWork, xArmy, xLore}
	for i, tabName := range peopleTabs {
		name := tabName.String()
		k := xDim
		if i == v.tab {
			k = xSel
		}
		end := g.put(g.put(x, 1, "■", squares[i]), 1, " "+name+" ", k)
		g.tab(image.Rect(x, 1, end, 2), name, i, i == v.tab)
		x = end + 1
	}
	// The key that steps through the tabs, once, at the end of the row.
	g.keyButton(x, 1, "←→", "tabs")
	if !festivalUnlocked(st) {
		return
	}
	word, f := txtFestivalHold.String(), st.Festival
	off := !f.Ready || f.Culture < f.Cost
	if note := festivalChip(st); note != "" {
		word, off = note, true
	}
	g.chip(g.w-3-runeLen(word)-2, 1, chipSpec{kind: keyAccent, label: word, cmd: festivalCmd, ask: festivalAsk(st), yes: txtFestivalHold.String(), off: off})
}

// peopleBody is the part of the panel under the tabs: the rectangle the tab
// draws into, and what it draws from.
type peopleBody struct {
	g          *xgrid
	st         game.GameState
	v          peopleView
	buy        buyChoice // the shared amount
	hist       *game.RunHistory
	s          *pScene // the tab's picture
	x, y, w, h int
}

// footer writes the key bar on the panel's last row.
func (c peopleBody) footer(keys [][2]string, hint string) {
	c.g.foot(keys, hint)
}

// A log line is picked for a panel by what it is about (LogEntry.About): by
// its words if it was written in the game's own English, as it always was, and
// by the tags it was given where it was written otherwise.

// peopleEvent says whether a log line is about people: someone recruited,
// moved in, starved or fell.
func peopleEvent(e game.LogEntry) bool { return e.Office == game.OfficePeople }

// armyEvent says whether a log line is about soldiers.
func armyEvent(e game.LogEntry) bool { return e.Office == game.OfficeArmy }

// recentLog is the newest n log lines (newest first) that keep accepts.
func recentLog(st game.GameState, keep func(game.LogEntry) bool, n int) []game.LogEntry {
	var out []game.LogEntry
	for i := len(st.Log) - 1; i >= 0 && len(out) < n; i-- {
		e := st.Log[i]
		if e.Type == "debug" || !keep(e) {
			continue
		}
		out = append(out, e)
	}
	return out
}

// logKind is the colour a log line takes: bad news red, a gain green.
func logKind(e game.LogEntry) xk {
	switch {
	case e.About(game.TagBad):
		return xBad
	case e.About(game.TagGood):
		return xGood
	}
	return xText
}

// drawLog writes lines into a box of w cells at (x, y): the log lines that
// fit, or a quiet line when there are none.
func drawLog(g *xgrid, x, y, w, rows int, lines []game.LogEntry, none string) {
	if len(lines) == 0 {
		g.putFit(x, y, w, none, xFaint)
		return
	}
	for i, e := range lines {
		if i >= rows {
			break
		}
		g.putFit(x, y+i, w, strings.TrimSpace(plainLogText(e.Message)), logKind(e))
	}
}

// tagPattern is a colour tag in a log line, such as a flavor tag or "[-]".
var tagPattern = regexp.MustCompile(`\[[a-zA-Z#0-9:_\-]*\]`)

// plainLogText strips the colour tags a log line may carry.
func plainLogText(s string) string { return tagPattern.ReplaceAllString(s, "") }
