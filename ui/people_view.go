package ui

import (
	"regexp"
	"sort"
	"strings"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/vibe"
)

// people_view.go draws the People panel: a pure function of the game's state
// and what the player has picked, into a grid the size of the panel (the
// Expeditions panel's cell grid and palette). Four tabs: Overview (the whole
// town at a glance), Work (the roster and the work areas), Army (soldiers
// and the buildings that make them) and Scholars (desks and research). The
// layout is worked out at 100 columns by 30 rows and shrinks to 80 by 24:
// boxes lose blank lines before they lose facts.

// peopleTabs are the tab names, in order.
var peopleTabs = []vibe.Entry{vibe.E("ui.people_view.peopleTabs.01", "OVERVIEW"), vibe.E("ui.people_view.peopleTabs.02", "WORK"), vibe.E("ui.people_view.peopleTabs.03", "ARMY"), vibe.E("ui.people_view.peopleTabs.04", "SCHOLARS")}

const (
	tabOverview = iota
	tabWork
	tabArmy
	tabScholars
)

// peopleView is what the player has picked.
type peopleView struct {
	tab int // the tab
	row int // the picked row on the tab (a card, an area, a building)
}

// peopleKeys are the panel's keys as the Help panel lists them.
var peopleKeys = []keyHelp{
	{"Left/Right", vibe.E("ui.people_view.peopleKeys.01", "Switch between Overview, Work, Army and Scholars")},
	{"Up/Down", vibe.E("ui.people_view.peopleKeys.02", "Pick a card, a work area or a building")},
	{"Esc", vibe.E("ui.people_view.peopleKeys.04", "Close the panel")},
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

// renderPeople draws the panel w by h.
func renderPeople(st game.GameState, v peopleView, w, h int) *xgrid {
	g := newXGrid(w, h)
	g.box(0, 0, w, h, xBorder, vibe.Line("ui.people_view.renderPeople.01", "People · ")+st.AgeName, xTitle, false)
	if w < 60 || h < 14 {
		g.put(2, 2, fitText(vibe.Line("ui.people_view.renderPeople.02", "The window is too small for the People panel. Esc closes it."), w-4), xBad)
		return g
	}
	v.tab = clamp(v.tab, 0, len(peopleTabs)-1)
	if n := peopleRows(st, v.tab); n > 0 {
		v.row = clamp(v.row, 0, n-1)
	} else {
		v.row = 0
	}
	peopleHeader(g, st, v)
	c := peopleBody{g: g, st: st, v: v, x: 1, y: 2, w: w - 2, h: h - 4}
	switch v.tab {
	case tabOverview:
		c.overview()
	case tabWork:
		c.work()
	case tabArmy:
		c.army()
	case tabScholars:
		c.scholars()
	}
	return g
}

// peopleHeader is the tab row and the headcount.
func peopleHeader(g *xgrid, st game.GameState, v peopleView) {
	x := 2
	for i, tabName := range peopleTabs {
		name := tabName.String()
		k := xDim
		if i == v.tab {
			k = xSel
		}
		x = g.put(x, 1, " "+name+" ", k) + 1
	}
	ws := st.Workers
	pop := FormatNumber(float64(ws.TotalPop))
	homes := FormatNumber(float64(ws.MaxPop))
	n := runeLen(vibe.Line("ui.people_view.peopleHeader.01", "People ") + pop + vibe.Line("ui.people_view.peopleHeader.02", " of ") + homes + vibe.Line("ui.people_view.peopleHeader.03", " homes"))
	if start := g.w - 3 - n + 1; start > x+1 {
		k := xBright
		if ws.TotalPop > ws.MaxPop {
			k = xBad
		}
		g.segs(start, 1, xseg{vibe.Line("ui.people_view.peopleHeader.04", "People "), xDim}, xseg{pop, k}, xseg{vibe.Line("ui.people_view.peopleHeader.05", " of ") + homes + vibe.Line("ui.people_view.peopleHeader.06", " homes"), xText})
	}
}

// peopleBody is the part of the panel under the tabs: the rectangle the tab
// draws into, and what it draws from.
type peopleBody struct {
	g          *xgrid
	st         game.GameState
	v          peopleView
	x, y, w, h int
}

// footer writes the key bar on the panel's last row.
func (c peopleBody) footer(keys [][2]string, hint string) {
	expKeyBar(c.g, expLayout{w: c.g.w, h: c.g.h, footY: c.g.h - 2}, keys, hint)
}

// stackBar draws the homes as one bar: each kind of person in its colour,
// the empty homes dim. It scales to the larger of the homes and the
// population, so a town with more people than homes still fits.
func stackBar(g *xgrid, x, y, w int, parts []stackPart, homes int) {
	counts := make([]int, len(parts))
	total := 0
	for i, p := range parts {
		counts[i] = p.n
		total += p.n
	}
	if max(homes, total) <= 0 || w <= 0 {
		return
	}
	used := 0
	for i, cells := range stackWidths(counts, homes, w) {
		for j := 0; j < cells; j++ {
			g.set(x+used+j, y, '█', parts[i].kind)
		}
		used += cells
	}
	for i := used; i < w; i++ {
		g.set(x+i, y, '░', xFaint)
	}
}

type stackPart struct {
	n    int
	kind xk
}

// kindCount is a legend entry: the block in the kind's colour, the name, the
// count.
func kindLegend(g *xgrid, x, y int, k peopleKind, housed bool) int {
	mark := "█"
	if !housed {
		mark = "·"
	}
	x = g.put(x, y, mark+" "+k.name+" ", k.kind)
	return g.put(x, y, FormatNumber(float64(k.count)), xText)
}

// A log line is picked for a panel by what it is about (LogEntry.About): by
// its words if it was written in the game's own English, as it always was, and
// by the tags it was given where it was written if it was written under
// VibeMode's script, whose words are its own.

// peopleEvent says whether a log line is about people: someone recruited,
// moved in, starved or fell.
func peopleEvent(e game.LogEntry) bool { return e.About(game.TagPeople) }

// armyEvent says whether a log line is about soldiers.
func armyEvent(e game.LogEntry) bool { return e.About(game.TagArmy) }

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
