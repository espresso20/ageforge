package logroll

// Mode is how much of the panel fits.
type Mode int

const (
	Placeholder Mode = iota // too small: one centred line
	Compact                 // down to MinCompactCols x MinCompactRows
	Full                    // MinFullCols x MinFullRows and up
)

const (
	MinFullCols, MinFullRows       = 230, 54
	MinCompactCols, MinCompactRows = 150, 38
)

// Rect is a block of whole cells.
type Rect struct{ X, Y, W, H int }

func (r Rect) Right() int  { return r.X + r.W }
func (r Rect) Bottom() int { return r.Y + r.H }
func (r Rect) In(x, y int) bool {
	return x >= r.X && x < r.Right() && y >= r.Y && y < r.Bottom()
}
func (r Rect) Overlaps(o Rect) bool {
	return r.W > 0 && r.H > 0 && o.W > 0 && o.H > 0 &&
		r.X < o.Right() && o.X < r.Right() && r.Y < o.Bottom() && o.Y < r.Bottom()
}
func (r Rect) Holds(o Rect) bool {
	return o.X >= r.X && o.Y >= r.Y && o.Right() <= r.Right() && o.Bottom() <= r.Bottom()
}

// Layout is every block of the panel in cells, relative to the panel's own
// top left corner. Nothing in it comes from a fixed row count: it is planned
// from the size the panel is handed.
type Layout struct {
	Mode       Mode
	Cols, Rows int

	Title, Find, Head, Hint Rect // strip, find slip (inside Title), column heads, hint line
	Rack, Lever             Rect
	Roll, Body, Rail        Rect // Roll holds its two rollers; Body is the lines
	Slate, SlateBody, Ledge Rect // SlateBody and Ledge sit inside Slate

	Stamps [10]Rect // one per office, inside Rack
	Stops  [3]Rect  // one per level, inside Lever

	// Roll columns: X and width in cells. Gained is empty in Compact (its
	// amounts share Paid) and Strokes tells whether Times has room for them.
	Clock, Stamp, Text, Paid, Gained, Times Rect
	Strokes                                 bool
}

// Key tells two layouts apart for picture caches.
func (l Layout) Key() [2]int { return [2]int{l.Cols, l.Rows} }

// Plan lays the panel out in a rect of cols by rows cells.
func Plan(cols, rows int) Layout {
	l := Layout{Cols: cols, Rows: rows}
	switch {
	case cols >= MinFullCols && rows >= MinFullRows:
		l.Mode = Full
	case cols >= MinCompactCols && rows >= MinCompactRows:
		l.Mode = Compact
	default:
		return l
	}
	full := l.Mode == Full
	rackW, rackX, slateW, margin := 26, 2, 79, 2
	if !full {
		rackW, rackX, slateW, margin = 8, 1, 50, 1
	}
	l.Slate = Rect{cols - margin - slateW, 3, slateW, rows - 5}
	l.Rail = Rect{l.Slate.X - 5, 4, 3, rows - 8}
	rollX := rackX + rackW + 2
	l.Roll = Rect{rollX, 3, l.Rail.X - rollX, rows - 5}
	l.Body = Rect{rollX, 4, l.Roll.W, rows - 7}
	l.Head = Rect{rollX, 2, l.Roll.W, 1}
	l.Title = Rect{margin, 0, l.Slate.X - margin - 2, 2}
	l.Find = Rect{rollX, 0, min(40, l.Title.Right()-rollX), 2}
	l.Hint = Rect{margin, rows - 1, cols - 2*margin, 1}
	l.SlateBody = Rect{l.Slate.X + 1, l.Slate.Y + 1, slateW - 2, l.Slate.H - 5}
	l.Ledge = Rect{l.Slate.X + 1, l.Slate.Bottom() - 3, slateW - 2, 3}

	if full {
		l.Rack = Rect{rackX, 4, rackW, 42}
		l.Lever = Rect{rackX, 48, rackW, rows - 51}
		for i := range l.Stamps {
			l.Stamps[i] = Rect{rackX + i%2*13, l.Rack.Y + 2 + i/2*8, 13, 8}
		}
		sh := (l.Lever.H - 2) / 3
		for i := range l.Stops {
			l.Stops[i] = Rect{l.Lever.X + 5, l.Lever.Y + 2 + i*sh, l.Lever.W - 7, sh}
		}
	} else {
		l.Rack = Rect{rackX, 4, rackW, 30}
		l.Lever = Rect{rackX, 34, rackW, 2}
		for i := range l.Stamps {
			l.Stamps[i] = Rect{rackX, l.Rack.Y + 3*i, rackW, 3}
		}
		for i := range l.Stops {
			l.Stops[i] = Rect{rackX + 1 + 2*i, l.Lever.Y, 2, 2}
		}
	}

	// Roll columns, from the right: Times, Gained, Paid; the rest is text.
	paidW, gainedW, timesW := 35, 20, 21
	switch {
	case !full:
		paidW, gainedW, timesW = 22, 0, 8
	case l.Roll.W < 150:
		paidW, gainedW, timesW = 26, 16, 18
	}
	l.Strokes = full
	x := l.Roll.Right() - 1
	l.Times = Rect{x - timesW, 0, timesW, 1}
	x = l.Times.X - 1
	if gainedW > 0 {
		l.Gained = Rect{x - gainedW, 0, gainedW, 1}
		x = l.Gained.X - 1
	}
	l.Paid = Rect{x - paidW, 0, paidW, 1}
	l.Clock = Rect{rollX + 1, 0, 6, 1}
	l.Stamp = Rect{l.Clock.Right() + 1, 0, 3, 1}
	l.Text = Rect{l.Stamp.Right() + 1, 0, l.Paid.X - l.Stamp.Right() - 2, 1}
	return l
}
