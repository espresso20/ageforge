package kit

import "image"

// Scroll is a window onto a child taller than the room there is. The wheel,
// the arrows, the page keys, Home and End move it, and so does dragging the
// bar on its right edge. The child is laid out at its full height and
// painted through the window, so any widget can go inside.
type Scroll struct {
	Base
	Child Widget
	// Focus lets it hold the keyboard, so its keys work without a parent
	// passing them on.
	Focus bool
	// Follow keeps the end in view while the child grows, as long as the
	// player has not scrolled away from it: a log.
	Follow bool
	// NoBar leaves the scroll bar out, so the child keeps the whole width.
	NoBar bool

	top   int  // the child's first row in view
	total int  // the child's height
	stuck bool // the end is in view and the player has not left it (see Follow)
	bar   bool // the child was too tall last time, so a column went to the bar
	held  bool // the thumb is being dragged
	grab  int  // and where in it the pointer took hold, in eighths of a cell
}

// NewScroll is a scroll view of child.
func NewScroll(child Widget) *Scroll { return &Scroll{Child: child} }

func (s *Scroll) Children() []Widget { return []Widget{s.Child} }

func (s *Scroll) CanFocus() bool { return s.Focus }

// Top is the child's first row in view.
func (s *Scroll) Top() int { return s.top }

// AtEnd reports whether the child's last row is in view.
func (s *Scroll) AtEnd() bool { return s.top >= s.total-s.rect.Dy() }

// ScrollTo asks for the child's row n at the top. The next layout holds it to
// what there is to scroll, so it may be called before the child has a size.
func (s *Scroll) ScrollTo(n int) { s.top, s.stuck = max(n, 0), false }

// ScrollBy asks for the view d rows down (up when negative).
func (s *Scroll) ScrollBy(d int) { s.ScrollTo(s.top + d) }

// ToEnd asks for the child's last rows.
func (s *Scroll) ToEnd() { s.top, s.stuck = 1<<30, true }

// move scrolls now, within what the last layout found there is: for the
// keys and the mouse, which come after one.
func (s *Scroll) move(n int) {
	s.top = max(0, min(n, s.total-s.rect.Dy()))
	s.stuck = s.AtEnd()
}

func (s *Scroll) Layout(r image.Rectangle) {
	atEnd := s.Follow && (s.total == 0 || s.stuck)
	s.Place(r)
	// The child's height decides whether there is a bar, and the bar's
	// column changes the child's width. Try the width that held last time
	// first, so a settled view measures its child once.
	w := r.Dx()
	if s.NoBar {
		s.total, s.bar = s.height(w), false
	} else if s.bar && w > 1 {
		if s.total = s.height(w - 1); s.total > r.Dy() {
			w--
		} else {
			s.total, s.bar = s.height(w), false
		}
	} else if s.total = s.height(w); s.total > r.Dy() && w > 1 {
		w--
		s.total, s.bar = s.height(w), true
	}
	if atEnd {
		s.top = s.total
	}
	s.move(s.top)
	y := r.Min.Y - s.top
	s.Child.Layout(image.Rect(r.Min.X, y, r.Min.X+w, y+max(s.total, r.Dy())))
}

func (s *Scroll) height(w int) int {
	if z, ok := s.Child.(Sizer); ok {
		return z.Size(image.Pt(w, 1<<20)).Y
	}
	return s.rect.Dy()
}

func (s *Scroll) Paint(p *Painter) {
	r := s.rect
	p = p.Clip(r)
	s.Child.Paint(p)
	if s.bar {
		p.Bar(r.Max.X-1, r.Min.Y, r.Dy(), s.top, r.Dy(), s.total)
	}
}

func (s *Scroll) Key(k Key) bool {
	page := max(s.rect.Dy()-1, 1)
	was := s.top
	switch k.Code {
	case KeyUp:
		s.move(s.top - 1)
	case KeyDown:
		s.move(s.top + 1)
	case KeyPgUp:
		s.move(s.top - page)
	case KeyPgDn:
		s.move(s.top + page)
	case KeyHome:
		s.move(0)
	case KeyEnd:
		s.move(s.total)
	default:
		return false
	}
	// A key that could not move anything is not used: at the top, Up may
	// mean something to whatever is outside.
	return s.top != was || s.Focus
}

func (s *Scroll) Mouse(m Mouse) bool {
	r := s.rect
	onBar := s.bar && m.X == r.Max.X-1
	switch m.Kind {
	case MouseWheel:
		was := s.top
		s.move(s.top + m.Wheel*wheelRows)
		return s.top != was
	case MousePress:
		if m.Button != ButtonLeft || !onBar {
			return false
		}
		// On the thumb: drag from where it was taken. Off it: the thumb's
		// middle jumps to the pointer, and the drag goes on from there.
		from, to := thumbSpan(r.Dy(), s.top, r.Dy(), s.total)
		row := (m.Y - r.Min.Y) * 8
		s.held, s.grab = true, (to-from)/2
		if row+8 > from && row < to {
			s.grab = row + 4 - from
		}
		s.drag(m.Y)
		return true
	case MouseMove:
		if !s.held {
			return false
		}
		s.drag(m.Y)
		return true
	case MouseRelease:
		held := s.held
		s.held = false
		return held
	}
	return false
}

// drag moves the view so the thumb's grabbed point is at pointer row y.
func (s *Scroll) drag(y int) {
	h := s.rect.Dy()
	from, to := thumbSpan(h, s.top, h, s.total)
	free := h*8 - (to - from)
	if free <= 0 {
		return
	}
	want := (y-s.rect.Min.Y)*8 + 4 - s.grab // where the thumb's top should be, in eighths
	s.move((want*(s.total-h) + free/2) / free)
}
