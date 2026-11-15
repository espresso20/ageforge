package ui

// dash_layout.go is the dashboard's body: the economy tab on the left and
// the Panels column on the right, laid out for the terminal they are drawn
// on. A big terminal gets the rich layout; a small one keeps the classic
// one, which fits 80x24 and is the layout the dashboard always had.
//
//	classic (under richMinCols by richMinRows)    rich
//	Resources | Buildings | Panels                 Resources | Buildings | Panels
//	Construction (whole row)  | People             Construction | People  |  (the whole
//	Log                       |                    Log                    |  column)
//
// The rich layout widens the Panels column so each panel can say what it is
// for, gives that whole column to the Panels list, and moves the People box
// into the right half of the Under construction row, where it has the rows
// for the bar, the posts and what holds the population back.

const (
	// richMinCols and richMinRows are the smallest terminal the rich layout
	// is drawn on. Below it the Panels column is sidebarW wide and the People
	// box sits under the list, as it always did.
	richMinCols = 120
	richMinRows = 36
	// richSidebarW is the Panels column's width in the rich layout: room for
	// a panel's name and a short line on what it is for (panelsRichWidth).
	richSidebarW = 33
	// richPeopleRows is the People box's height, border included, in the rich
	// layout: the headcount, the bar, the four kinds and one line on what
	// holds the population back. richPeopleDetailRows adds the second line of
	// that (the food detail), on a terminal richDetailMinRows tall or more.
	richPeopleRows       = 9
	richPeopleDetailRows = 10
	richDetailMinRows    = 42
)

// miniMapMinCols and miniMapMinRows are the smallest terminal the mini map
// shows on in the classic layout (under richMinCols columns; from 100
// columns it takes one more row, 37): the Panels list in its shortest form,
// a box of miniShortestInnerH picture rows under it, and the People box
// under that. The rich layout shows the mini map at every size it is drawn
// on. The wiki and the minimap command quote them (TestMiniMapAppearsWhereTheDocsSay).
const (
	miniMapMinCols = 80
	miniMapMinRows = 36
)

// dashRich says whether a w by h terminal gets the rich layout.
func dashRich(w, h int) bool { return w >= richMinCols && h >= richMinRows }

// richPeopleHeight is the People box's height, border included, in the rich
// layout on a terminal h rows tall.
func richPeopleHeight(termH int) int {
	if termH >= richDetailMinRows {
		return richPeopleDetailRows
	}
	return richPeopleRows
}

// dashBody is the dashboard's content area: the economy tab and, beside it,
// the Panels column. It chooses the layout on every draw from the size of the
// terminal (screen), so a resize shows the right one at once.
type dashBody struct {
	*nativeGroup
	econ, panels, people Primitive
	// mini is the mini map, a square box under the Panels list; miniOff is
	// the minimap setting turned off. miniShown reports whether the last
	// layout gave it room (miniFits: the size has room, whatever the model
	// and the setting say), miniDrawn that there was a layout at all (before
	// it, there is no real size).
	mini                           Primitive
	miniOff                        bool
	miniShown, miniFits, miniDrawn bool
	// screen is the terminal's size.
	screen func() (w, h int)
	// classicPeople is the People box's height, border included, in the
	// classic layout (it grows by the lines that say what holds the
	// population back).
	classicPeople func() int
	// switched is told when the layout changes, before the boxes are drawn
	// in it: the People box is written differently in each, and may come out
	// the same size.
	switched func()
	rich     bool
	// modelReady reports whether the mini map has a model to draw.
	modelReady func() bool
}

func newDashBody(econ, panels, people, mini Primitive, screen func() (w, h int), classicPeople func() int) *dashBody {
	b := &dashBody{econ: econ, panels: panels, people: people, mini: mini, screen: screen, classicPeople: classicPeople}
	b.nativeGroup = newNativeGroup(b.arrange)
	return b
}

func (b *dashBody) arrange(x, y, w, h int) {
	if w <= 0 || h <= 0 {
		return
	}
	tw, th := b.screen()
	rich := dashRich(tw, th)
	if rich != b.rich {
		b.rich = rich
		if b.switched != nil {
			b.switched()
		}
	}
	side := sidebarW
	if rich {
		side = richSidebarW
	}
	side = min(side, w)
	b.put(b.econ, x, y, w-side, h)
	// The right column: the Panels list, then the mini map, then (in the
	// classic layout) the People box. In the rich layout the People box is
	// in the economy tab's Under construction row.
	ph := 0
	if !rich {
		ph = min(b.classicPeople(), h)
	}
	region := h - ph
	mh := miniBoxHeight(side, region)
	b.miniFits = mh > 0
	if b.miniOff || b.mini == nil || !b.miniReady() {
		mh = 0
	}
	b.miniShown, b.miniDrawn = mh > 0, true
	b.put(b.panels, x+w-side, y, side, region-mh)
	if mh > 0 {
		b.put(b.mini, x+w-side, y+region-mh, side, mh)
	}
	if ph > 0 {
		b.put(b.people, x+w-side, y+h-ph, side, ph)
	}
}

// miniReady is set by the dashboard: false until a refresh has built the
// mini map's model.
func (b *dashBody) miniReady() bool { return b.modelReady != nil && b.modelReady() }

// wantsModel reports whether the mini map needs a model: it is on and
// room, or nothing has been laid out yet.
func (b *dashBody) wantsModel() bool { return !b.miniOff && (!b.miniDrawn || b.miniFits) }

// miniShortestInnerH is the fewest picture rows a mini map shorter than
// square is shown with.
const miniShortestInnerH = 8

// miniBoxHeight is the mini map's height, border included, in a right
// column w wide with h rows for the Panels list and the mini map together,
// or 0 (hidden). The Panels list always lists every panel. Each form of the
// list is tried in turn, from the form it has by itself (the rows
// sidebarText picks a form from) to the shortest: with a box about square
// on screen (a cell is about 2.1 times taller than wide), else with a
// shorter box of at least miniShortestInnerH picture rows. The first that
// fits wins, and when none does there is no mini map.
func miniBoxHeight(w, h int) int {
	square := int(float64(w)/2.1 + 0.5)
	if square < miniMinInnerH {
		return 0
	}
	forms := panelsFormRows(w - 2)
	own := 0
	for own < len(forms)-1 && h-2 < forms[own] {
		own++
	}
	for f := own; f < len(forms); f++ {
		picture := h - (forms[f] + 2) - 2
		if picture >= square {
			return square + 2
		}
		if picture >= miniShortestInnerH {
			return picture + 2
		}
	}
	return 0
}

// panelsFormRows is the fewest rows inside the box each form of the Panels
// list needs, in w cells, longest form first: the grouped list (when w has
// room for it), a name a row, and two columns.
func panelsFormRows(w int) []int {
	n := len(sidebarPanels)
	forms := []int{n, (n + 1) / 2}
	if w >= panelsRichWidth() {
		forms = append([]int{panelsRichRows()}, forms...)
	}
	return forms
}
