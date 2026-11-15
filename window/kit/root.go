package kit

import (
	"image"
	"time"

	"github.com/espresso20/ageforge/theme"
)

// doubleClick is how close two presses on one cell must be to count as a
// double click.
const doubleClick = 400 * time.Millisecond

// A Root is one screen: a stack of layers, the top one taking the input. The
// first layer is the screen itself; a Modal pushed over it is the next.
//
// A key goes down through the top layer to the focused widget, past each
// KeyFilter on the way, then to that widget and outward through the widgets
// around it, then to Tab's walk of the focus, then to Keys. The mouse
// goes to the deepest widget under the pointer in the top layer and outward
// from there. Layers underneath get nothing.
type Root struct {
	// Keys sees a key that nothing in the tree used.
	Keys func(k Key) bool
	// Now is the clock double clicks are timed by; nil means time.Now.
	Now func() time.Time

	layers  []layer
	size    image.Point
	hover   Widget // the widget the pointer is over
	capture Widget // the widget that took a press, until the button comes up
	at      image.Point
	in      bool    // the pointer is on the screen, at at
	shape   Pointer // the shape it was last given
	press   struct {
		at time.Time
		pt image.Point
	}
}

type layer struct {
	w     Widget
	focus Widget
}

// NewRoot is a screen showing base.
func NewRoot(base Widget) *Root {
	r := &Root{}
	r.Push(base)
	return r
}

// Push shows w over everything, and gives it the input. The focus goes to
// the first widget in it that takes one.
func (r *Root) Push(w Widget) {
	r.dropPointer()
	setFocus(r.Focused(), false)
	r.layers = append(r.layers, layer{w: w})
	r.layout()
	r.Focus(first(focusables(w)))
}

// Pop takes the top layer away (never the screen itself); the layer under it
// has the input again, and its focus where it was.
func (r *Root) Pop() {
	if len(r.layers) > 1 {
		r.Remove(r.layers[len(r.layers)-1].w)
	}
}

// Remove takes the layer showing w away, wherever it is in the stack.
func (r *Root) Remove(w Widget) {
	for i := 1; i < len(r.layers); i++ {
		if r.layers[i].w != w {
			continue
		}
		r.dropPointer()
		setFocus(r.Focused(), false)
		r.layers = append(r.layers[:i], r.layers[i+1:]...)
		setFocus(r.Focused(), true)
		return
	}
}

// Top is the layer that has the input.
func (r *Root) Top() Widget { return r.layers[len(r.layers)-1].w }

// Depth is the number of layers: 1 with no modal up.
func (r *Root) Depth() int { return len(r.layers) }

func first(ws []Widget) Widget {
	if len(ws) == 0 {
		return nil
	}
	return ws[0]
}

// focusables are the widgets in w that can hold the keyboard, in tree order.
func focusables(w Widget) []Widget {
	var out []Widget
	walk(w, func(c Widget) {
		if f, ok := c.(Focusable); ok && f.CanFocus() {
			out = append(out, c)
		}
	})
	return out
}

// Focused is the widget holding the keyboard; nil when none does.
func (r *Root) Focused() Widget {
	if len(r.layers) == 0 {
		return nil
	}
	return r.layers[len(r.layers)-1].focus
}

// Focus gives the keyboard to w, which must be in the top layer; nil takes
// it from everything. A Focusable that cannot take it now is refused. A
// widget that is not a Focusable may still be given the keys: it is never
// told, and Tab and the mouse never choose it.
func (r *Root) Focus(w Widget) {
	top := &r.layers[len(r.layers)-1]
	if w != nil {
		if f, ok := w.(Focusable); ok && !f.CanFocus() || pathTo(top.w, w) == nil {
			return
		}
	}
	if top.focus == w {
		return
	}
	setFocus(top.focus, false)
	top.focus = w
	setFocus(w, true)
}

// setFocus tells w the focus has come or gone, if it wants to know.
func setFocus(w Widget, on bool) {
	if f, ok := w.(Focusable); ok {
		f.SetFocus(on)
	}
}

// Resize sets the screen's size in cells.
func (r *Root) Resize(w, h int) { r.size = image.Pt(w, h) }

func (r *Root) layout() {
	box := image.Rectangle{Max: r.size}
	for _, l := range r.layers {
		l.w.Layout(box)
	}
}

// Paint lays the layers out and paints them into b, the screen first. The
// buffer is made the root's size if it is not.
func (r *Root) Paint(b *Buffer) {
	if b.W != r.size.X || b.H != r.size.Y {
		b.Resize(r.size.X, r.size.Y)
	} else {
		b.Reset()
	}
	r.layout()
	p := NewPainter(b)
	for _, l := range r.layers {
		l.w.Paint(p)
	}
	// What is under a still pointer may have changed with this paint.
	r.shape = r.pointer()
	b.Pointer = r.shape
}

// pointer is the shape the mouse pointer takes where it is now. While a
// press is held it keeps the shape it had.
func (r *Root) pointer() Pointer {
	switch {
	case !r.in:
		return PointerArrow
	case r.capture != nil:
		return r.shape
	}
	return pointerAt(pathAt(r.Top(), r.at), r.at)
}

// Key gives the tree a key, and reports whether anything used it.
func (r *Root) Key(k Key) bool {
	top := r.layers[len(r.layers)-1]
	path := pathTo(top.w, top.focus)
	if top.focus == nil || path == nil {
		path = []Widget{top.w} // nothing focused: the layer itself still hears keys
	}
	for _, w := range path {
		if f, ok := w.(KeyFilter); ok {
			var on bool
			if k, on = f.FilterKey(k); !on {
				return true
			}
		}
	}
	for i := len(path) - 1; i >= 0; i-- {
		if h, ok := path[i].(KeyHandler); ok && h.Key(k) {
			return true
		}
	}
	if k.Code == KeyTab || k.Code == KeyBacktab {
		if r.tab(k.Code == KeyBacktab) {
			return true
		}
	}
	return r.Keys != nil && r.Keys(k)
}

// tab moves the focus to the next widget that takes it (the one before, when
// back), round the ends. It reports whether the focus moved.
func (r *Root) tab(back bool) bool {
	fs := focusables(r.Top())
	if len(fs) == 0 {
		return false
	}
	at := -1
	for i, w := range fs {
		if w == r.Focused() {
			at = i
		}
	}
	next := (at + 1) % len(fs)
	if back {
		next = (max(at, 0) + len(fs) - 1) % len(fs)
	}
	if fs[next] == r.Focused() {
		return false
	}
	r.Focus(fs[next])
	return true
}

func (r *Root) now() time.Time {
	if r.Now != nil {
		return r.Now()
	}
	return time.Now()
}

// setHover notes the widget under the pointer, telling the one it left.
func (r *Root) setHover(w Widget, m Mouse) (changed bool) {
	if r.hover == w {
		return false
	}
	if r.hover != nil {
		m.Kind = MouseLeave
		r.hover.(MouseHandler).Mouse(m)
	}
	r.hover = w
	return true
}

// dropPointer forgets the pointer: nothing is hovered or holding a press.
func (r *Root) dropPointer() {
	r.setHover(nil, Mouse{X: -1, Y: -1})
	r.capture = nil
}

// Mouse gives the tree a mouse event, and reports whether the screen may
// look different for it (something used it, the hover moved, or the pointer
// takes another shape). A MouseLeave says the pointer has left the screen.
func (r *Root) Mouse(m Mouse) bool {
	changed := r.mouse(m)
	if r.in = m.Kind != MouseLeave; r.in {
		r.at = m.Pt()
	}
	if s := r.pointer(); s != r.shape {
		r.shape, changed = s, true
	}
	return changed
}

func (r *Root) mouse(m Mouse) bool {
	if m.Kind == MouseLeave {
		// A press that was still held ends here: its widget is told the
		// button came up, away from it.
		held := r.capture
		r.capture = nil
		if held != nil {
			held.(MouseHandler).Mouse(Mouse{Kind: MouseRelease, X: -1, Y: -1})
		}
		return r.setHover(nil, m) || held != nil
	}
	if r.capture != nil && (m.Kind == MouseMove || m.Kind == MouseRelease) {
		c := r.capture
		if m.Kind == MouseRelease {
			r.capture = nil
		}
		c.(MouseHandler).Mouse(m)
		return true
	}
	path := pathAt(r.Top(), m.Pt())
	var over Widget
	for i := len(path) - 1; i >= 0 && over == nil; i-- {
		if _, ok := path[i].(MouseHandler); ok {
			over = path[i]
		}
	}
	changed := r.setHover(over, m)
	if m.Kind == MousePress {
		now := r.now()
		m.Clicks = 1
		if m.Button == ButtonLeft && m.Pt() == r.press.pt && now.Sub(r.press.at) < doubleClick {
			m.Clicks = 2
			now = time.Time{} // a third press starts again
		}
		r.press.at, r.press.pt = now, m.Pt()
		for i := len(path) - 1; i >= 0; i-- {
			if f, ok := path[i].(Focusable); ok && f.CanFocus() {
				r.Focus(path[i])
				changed = true
				break
			}
		}
	}
	for i := len(path) - 1; i >= 0; i-- {
		if h, ok := path[i].(MouseHandler); ok && h.Mouse(m) {
			if m.Kind == MousePress {
				r.capture = path[i]
			}
			return true
		}
	}
	return changed
}

// Modal is a layer over the screen: its child in the middle, the screen
// behind it dimmed, and all the input its own. Push it on the Root to show
// it and Remove it to close it.
type Modal struct {
	Base
	Child Widget
	// W and H are the child's size in cells; zero asks the child, or takes
	// the whole way.
	W, H int
	// Dim is how far the screen behind is mixed toward the ground, 0 to 1;
	// zero means the usual amount and a negative number none.
	Dim float64
	// OnDismiss, when set, is called for Esc that the child did not use and
	// for a press outside the child.
	OnDismiss func()
}

// modalDim is how far a modal dims the screen behind it when it does not say.
const modalDim = 0.45

func (d *Modal) Children() []Widget { return []Widget{d.Child} }

func (d *Modal) Layout(r image.Rectangle) {
	d.Place(r)
	d.Child.Layout(centered(r, d.Child, d.W, d.H))
}

func (d *Modal) Paint(p *Painter) {
	dim := d.Dim
	if dim == 0 {
		dim = modalDim
	}
	p.Shade(d.rect, Role(theme.RoleBackground), dim)
	d.Child.Paint(p)
}

// Key takes every key the child left, but Tab, which still walks the
// focus inside it: a modal's keys go no further.
func (d *Modal) Key(k Key) bool {
	if k.Code == KeyEsc && d.OnDismiss != nil {
		d.OnDismiss()
	}
	return k.Code != KeyTab && k.Code != KeyBacktab
}

// Mouse takes the mouse wherever the child is not.
func (d *Modal) Mouse(m Mouse) bool {
	if m.Kind == MousePress && d.OnDismiss != nil && !m.Pt().In(d.Child.Bounds()) {
		d.OnDismiss()
	}
	return true
}
