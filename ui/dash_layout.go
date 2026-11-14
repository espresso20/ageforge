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
}

func newDashBody(econ, panels, people Primitive, screen func() (w, h int), classicPeople func() int) *dashBody {
	b := &dashBody{econ: econ, panels: panels, people: people, screen: screen, classicPeople: classicPeople}
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
	if rich {
		// The People box is in the economy tab's Under construction row.
		b.put(b.panels, x+w-side, y, side, h)
		return
	}
	ph := min(b.classicPeople(), h)
	b.put(b.panels, x+w-side, y, side, h-ph)
	b.put(b.people, x+w-side, y+h-ph, side, ph)
}
