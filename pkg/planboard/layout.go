// Package planboard draws the Plan panel's pictures: the board, the schedule,
// the job sheet, the gate card, the list of what the plan waits on and the
// tray. It knows nothing of the game: callers hand it plain rows. It paints
// pixels only; the words are the caller's, written as cell text on flat
// plates over these pictures, at the places Layout gives.
package planboard

import "image"

// Layout is where everything lies, in whole cells of a rect cols by rows.
type Layout struct {
	Cols, Rows, CW, CH int
	Compact            bool
	RowH               int             // rows an item takes: 2, or 1 when compact
	Sheet, Inner       image.Rectangle // the paper, and the cells its margin leaves for words
	Hint               int             // the row under the sheet
	ListX, ListW       int
	ListY              int // the first list row: the section header
	SchedEnd           int // the row after the last schedule row
	Lanes, Ruler, Park image.Rectangle
	RulerCells         int
	Right              image.Rectangle // the column of the job sheet and gate card
	Strip              image.Rectangle // waits-on and tray
	Waits, Tray        image.Rectangle
	TabW               int // one tray tab
}

// The smallest rect that gets a board, and the smallest for the full one.
const (
	MinCols, MinRows   = 150, 38
	FullCols, FullRows = 230, 54
)

// NewLayout lays a cols by rows rect out; ok is false below the smallest.
func NewLayout(cols, rows, cw, ch int) (l Layout, ok bool) {
	if cols < MinCols || rows < MinRows {
		return Layout{Cols: cols, Rows: rows}, false
	}
	l = Layout{Cols: cols, Rows: rows, CW: cw, CH: ch, RowH: 2, ListX: 3, ListW: 57, ListY: 6, TabW: 20}
	l.Compact = cols < FullCols || rows < FullRows
	l.Sheet = image.Rect(2, 1, cols-2, rows-1)
	l.Inner = image.Rect(3, 2, cols-3, rows-2)
	l.Hint = rows - 1
	rightW, park, stripH, trayW := 78, 11, 12, 153
	if rows >= 71 {
		stripH = 13
	}
	if l.Compact {
		l.RowH, l.ListW, rightW, park, stripH, trayW, l.TabW = 1, 44, 50, 6, 7, 90, 15
	}
	l.SchedEnd = l.Inner.Max.Y - stripH - 2
	rightX := l.Inner.Max.X - rightW
	laneX := l.ListX + l.ListW
	l.Lanes = image.Rect(laneX, l.ListY, rightX-1, l.SchedEnd)
	l.Ruler = image.Rect(laneX, 4, rightX-1, 6)
	l.Park = image.Rect(rightX-1-park, l.ListY, rightX-1, l.SchedEnd)
	l.RulerCells = l.Park.Min.X - (laneX + 1)
	l.Right = image.Rect(rightX, 4, l.Inner.Max.X, l.SchedEnd)
	l.Strip = image.Rect(3, l.SchedEnd+2, l.Inner.Max.X, l.Inner.Max.Y)
	l.Tray = image.Rect(l.Inner.Max.X-trayW, l.Strip.Min.Y, l.Inner.Max.X, l.Strip.Max.Y)
	l.Waits = image.Rect(3, l.Strip.Min.Y, l.Tray.Min.X-1, l.Strip.Max.Y)
	return l, true
}

// Cards cuts the right column for a job sheet jobH rows tall: the sheet, and
// the gate card under it (below an arch three rows tall). Compact has room
// for one: the gate card takes the whole column while the advance is
// selected (jobH 0).
func (l Layout) Cards(jobH int) (job, gate image.Rectangle) {
	r := l.Right
	if l.Compact {
		job = image.Rect(r.Min.X, r.Min.Y, r.Max.X, r.Max.Y)
		gate = image.Rect(r.Min.X, r.Min.Y+3, r.Max.X, r.Max.Y)
		return job, gate
	}
	job = image.Rect(r.Min.X, r.Min.Y, r.Max.X, min(r.Min.Y+jobH, r.Max.Y))
	gate = image.Rect(r.Min.X, min(job.Max.Y+4, r.Max.Y), r.Max.X, r.Max.Y)
	return job, gate
}

// A Slot is one row of the list: an item, or the header of a section (Item
// -1; Age counts the advances before it).
type Slot struct{ Item, Y, Age int }

// Slots places the items from first on. A header leads the first item shown
// and follows each advance (adv lists their indexes). Items that do not all
// fit are cut at the last whole one, with a row kept for "more below"
// (more is how many are left out).
func (l Layout) Slots(first, n int, adv []int) (slots []Slot, more int) {
	for _, reserve := range []int{0, 1} {
		slots, more = slots[:0], 0
		y, age, shown := l.ListY, 0, 0
		for _, a := range adv {
			if a < first {
				age++
			}
		}
		next := map[int]bool{}
		for _, a := range adv {
			next[a+1] = true
		}
		for i := first; i < n; i++ {
			if i == first || next[i] {
				if y+1+l.RowH > l.SchedEnd-reserve {
					break
				}
				slots = append(slots, Slot{-1, y, age})
				y++
			}
			if y+l.RowH > l.SchedEnd-reserve {
				break
			}
			slots = append(slots, Slot{i, y, age})
			y += l.RowH
			shown++
			for _, a := range adv {
				if a == i {
					age++
				}
			}
		}
		if more = n - first - shown; more == 0 {
			break
		}
	}
	return slots, more
}
