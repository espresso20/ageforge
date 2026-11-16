package kit

import (
	"image"

	"github.com/espresso20/ageforge/theme"
)

// wheelRows is the rows one notch of the wheel moves a list or a scroll view.
const wheelRows = 3

// List is rows of tagged text, one of them selected. The arrows, Home, End
// and the page keys move the selection; Enter chooses it. The mouse shows
// the row it is over, a click selects it and a double click chooses it.
type List struct {
	Base
	// OnChoose is called with the row the player chose; OnMove whenever the
	// selection lands on another row.
	OnChoose func(i int)
	OnMove   func(i int)
	// Loop makes moving past either end come round to the other.
	Loop bool
	// ClickChooses makes one click choose, as a menu does.
	ClickChooses bool
	// Bar shows a scroll bar down the right edge when rows are out of view.
	Bar bool
	// Style is how a row is painted; SelectedStyle replaces the theme's
	// selection colors when its Fg is set.
	Style, SelectedStyle Style
	// TextOnly colors the selected row's text, not the whole row.
	TextOnly bool
	// AfterRow, when set, is called after each row is painted with the
	// cells of the row, its number, whether it is the selected one and
	// whether the pointer is over it: how a screen lights a row its own way.
	AfterRow func(p *Painter, row image.Rectangle, i int, selected, hover bool)

	items []string
	sel   int
	top   int
	hover int // one more than the row under the pointer; 0 for none
	// lost is set when a Select could not scroll (no height yet): the next
	// Layout brings the selection into view.
	lost bool
}

// NewList is a list of rows.
func NewList(items ...string) *List { return &List{items: items, Bar: true} }

// SetItems changes the rows. The selection stays on its row number, or the
// last row when there are fewer.
func (l *List) SetItems(items []string) {
	l.items = items
	l.sel = max(0, min(l.sel, len(items)-1))
	l.hover = 0
}

// Len is the number of rows.
func (l *List) Len() int { return len(l.items) }

// Selected is the selected row; -1 when the list is empty.
func (l *List) Selected() int {
	if len(l.items) == 0 {
		return -1
	}
	return l.sel
}

// Select moves the selection to row i (held inside the list) and brings it
// into view.
func (l *List) Select(i int) {
	if len(l.items) == 0 {
		return
	}
	i = max(0, min(i, len(l.items)-1))
	moved := i != l.sel
	l.sel = i
	l.show()
	if moved && l.OnMove != nil {
		l.OnMove(i)
	}
}

// show scrolls just far enough to have the selection in view.
func (l *List) show() {
	h := l.rect.Dy()
	if h <= 0 {
		l.lost = true
		return
	}
	l.lost = false
	if l.sel < l.top {
		l.top = l.sel
	}
	if l.sel >= l.top+h {
		l.top = l.sel - h + 1
	}
	l.clamp()
}

func (l *List) clamp() {
	l.top = max(0, min(l.top, len(l.items)-l.rect.Dy()))
}

func (l *List) CanFocus() bool { return len(l.items) > 0 }

func (l *List) Layout(r image.Rectangle) {
	l.Place(r)
	if l.lost {
		l.show()
	}
	l.clamp()
}

func (l *List) Size(most image.Point) image.Point {
	w := 0
	for _, s := range l.items {
		w = max(w, Width(s))
	}
	return image.Pt(min(w, most.X), min(len(l.items), most.Y))
}

// textWidth is the cells a row's text has: the box, less the bar's column.
func (l *List) textWidth() int {
	if l.Bar && len(l.items) > l.rect.Dy() {
		return l.rect.Dx() - 1
	}
	return l.rect.Dx()
}

func (l *List) Paint(p *Painter) {
	r := l.rect
	p = p.Clip(r)
	w := l.textWidth()
	for y := 0; y < r.Dy() && l.top+y < len(l.items); y++ {
		i := l.top + y
		st := l.Style
		switch {
		case i == l.sel && l.SelectedStyle.Fg.IsSet():
			st = l.SelectedStyle
		case i == l.sel:
			st = Style{Fg: Role(theme.RoleSelectionText), Bg: Role(theme.RoleSelection)}
		case i == l.hover-1:
			st.Bg = Role(theme.RoleChip)
		}
		row := image.Rect(r.Min.X, r.Min.Y+y, r.Min.X+w, r.Min.Y+y+1)
		if !l.TextOnly || i != l.sel {
			p.Fill(row, st.Bg)
		}
		p.Print(row.Min.X, row.Min.Y, w, l.items[i], st)
		if l.AfterRow != nil {
			l.AfterRow(p, row, i, i == l.sel, i == l.hover-1)
		}
	}
	if w < r.Dx() {
		p.Bar(r.Max.X-1, r.Min.Y, r.Dy(), l.top, r.Dy(), len(l.items))
	}
}

func (l *List) Key(k Key) bool {
	n := len(l.items)
	if n == 0 {
		return false
	}
	page := max(l.rect.Dy()-1, 1)
	to := l.sel
	switch k.Code {
	case KeyUp:
		to--
	case KeyDown:
		to++
	case KeyPgUp:
		to = max(to-page, 0)
	case KeyPgDn:
		to = min(to+page, n-1)
	case KeyHome:
		to = 0
	case KeyEnd:
		to = n - 1
	case KeyEnter:
		if l.OnChoose != nil {
			l.OnChoose(l.sel)
		}
		return true
	default:
		return false
	}
	if l.Loop {
		to = (to + n) % n
	}
	l.Select(to)
	return true
}

// rowAt is the row whose text is at pt, or -1.
func (l *List) rowAt(pt image.Point) int {
	row := l.top + pt.Y - l.rect.Min.Y
	if row < 0 || row >= len(l.items) || pt.X >= l.rect.Min.X+l.textWidth() {
		return -1
	}
	return row
}

func (l *List) Mouse(m Mouse) bool {
	row := l.rowAt(m.Pt())
	switch m.Kind {
	case MouseLeave:
		l.hover = 0
	case MouseMove:
		l.hover = row + 1
	case MouseWheel:
		l.top += m.Wheel * wheelRows
		l.clamp()
		l.hover = 0
	case MousePress:
		if m.Button != ButtonLeft || row < 0 {
			return false
		}
		l.Select(row)
		if (l.ClickChooses || m.Clicks >= 2) && l.OnChoose != nil {
			l.OnChoose(row)
		}
	default:
		return false
	}
	return true
}
