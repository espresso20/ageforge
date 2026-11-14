package kit

import "image"

// A Widget is one thing on the screen. Layout gives it its rectangle (and it
// places its children); Paint draws it there.
type Widget interface {
	Layout(r image.Rectangle)
	Paint(p *Painter)
	Bounds() image.Rectangle
}

// The rest is optional: a widget is what it implements.

// A Sizer knows the size it would like, given the most it can have.
type Sizer interface {
	Size(most image.Point) image.Point
}

// A Parent has widgets inside it, listed back to front.
type Parent interface {
	Children() []Widget
}

// A KeyHandler takes keys while it, or something inside it, has the focus.
// It returns true when it used the key.
type KeyHandler interface {
	Key(k Key) bool
}

// A MouseHandler takes the mouse while the pointer is over it. It returns
// true when it used the event.
type MouseHandler interface {
	Mouse(m Mouse) bool
}

// A Focusable can hold the keyboard.
type Focusable interface {
	CanFocus() bool
	SetFocus(on bool)
}

// KeyCode is which key; KeyRune is a character, in Key.Rune.
type KeyCode uint8

const (
	KeyRune KeyCode = iota
	// KeyChord is a character key pressed with Ctrl or Alt held: the
	// character is in Key.Rune and the modifiers in Key.Mod. It is never
	// text, so a widget that reads KeyRune need not look at the modifiers.
	KeyChord
	KeyEnter
	KeyEsc
	KeyTab
	KeyBacktab
	KeyBackspace
	KeyDelete
	KeyInsert
	KeyUp
	KeyDown
	KeyLeft
	KeyRight
	KeyHome
	KeyEnd
	KeyPgUp
	KeyPgDn
	KeyF1
	KeyF2
	KeyF3
	KeyF4
	KeyF5
	KeyF6
	KeyF7
	KeyF8
	KeyF9
	KeyF10
	KeyF11
	KeyF12
	// KeyOther is a key the kit has no name for. A screen that moves on at
	// any key still hears it.
	KeyOther
)

// Mod is the modifier keys held.
type Mod uint8

const (
	ModShift Mod = 1 << iota
	ModCtrl
	ModAlt
)

// A Key is one key press.
type Key struct {
	Code KeyCode
	Rune rune
	Mod  Mod
}

// Is reports whether k is the plain character r.
func (k Key) Is(r rune) bool { return k.Code == KeyRune && k.Rune == r && k.Mod&^ModShift == 0 }

// Ctrl reports whether k is Ctrl with the letter r, given in lower case.
func (k Key) Ctrl(r rune) bool {
	return k.Code == KeyChord && k.Mod&ModCtrl != 0 && (k.Rune == r || k.Rune == r-'a'+'A')
}

// MouseKind is what the mouse did.
type MouseKind uint8

const (
	MouseMove    MouseKind = iota // the pointer is over a new cell
	MousePress                    // a button went down
	MouseRelease                  // and came up
	MouseWheel                    // the wheel turned: Wheel rows, down positive
	MouseLeave                    // the pointer is no longer over the widget
)

// MouseButton is which button.
type MouseButton uint8

const (
	ButtonLeft MouseButton = iota
	ButtonRight
	ButtonMiddle
)

// A Mouse is one mouse event, at a cell.
type Mouse struct {
	Kind   MouseKind
	X, Y   int
	Button MouseButton
	Wheel  int
	Mod    Mod
	// Clicks is 2 on the second press of a double click, else 1. The root
	// sets it.
	Clicks int
}

// Pt is where the event is.
func (m Mouse) Pt() image.Point { return image.Pt(m.X, m.Y) }

// Base is the part every widget has: its rectangle, whether it is hidden,
// and whether it has the focus. Widgets embed it.
type Base struct {
	rect    image.Rectangle
	focused bool
	// Hide takes the widget off the screen: it is not painted, takes no
	// room in a Row or Col, and gets no input.
	Hide bool
}

// Bounds is where the widget was last placed.
func (b *Base) Bounds() image.Rectangle { return b.rect }

// Place sets the rectangle: a widget's Layout calls it first.
func (b *Base) Place(r image.Rectangle) { b.rect = r }

// Hidden reports whether the widget is off the screen.
func (b *Base) Hidden() bool { return b.Hide }

// SetFocus is called by the root when the focus comes or goes.
func (b *Base) SetFocus(on bool) { b.focused = on }

// Focused reports whether the widget holds the keyboard.
func (b *Base) Focused() bool { return b.focused }

func hidden(w Widget) bool {
	h, ok := w.(interface{ Hidden() bool })
	return w == nil || (ok && h.Hidden())
}

func children(w Widget) []Widget {
	if p, ok := w.(Parent); ok {
		return p.Children()
	}
	return nil
}

// walk visits w and everything shown inside it, parents first.
func walk(w Widget, fn func(Widget)) {
	if hidden(w) {
		return
	}
	fn(w)
	for _, c := range children(w) {
		walk(c, fn)
	}
}

// pathTo is the chain of widgets from root down to target; nil when target
// is not in the tree.
func pathTo(root, target Widget) []Widget {
	if hidden(root) {
		return nil
	}
	if root == target {
		return []Widget{root}
	}
	for _, c := range children(root) {
		if p := pathTo(c, target); p != nil {
			return append([]Widget{root}, p...)
		}
	}
	return nil
}

// pathAt is the chain of widgets under a point, outermost first. Later
// children are in front, so they are asked first.
func pathAt(root Widget, pt image.Point) []Widget {
	if hidden(root) || !pt.In(root.Bounds()) {
		return nil
	}
	cs := children(root)
	for i := len(cs) - 1; i >= 0; i-- {
		if p := pathAt(cs[i], pt); p != nil {
			return append([]Widget{root}, p...)
		}
	}
	return []Widget{root}
}

// Custom is a widget made of functions, for a screen that composes its own
// cells or needs one odd part. Any function may be nil.
type Custom struct {
	Base
	Draw    func(p *Painter, r image.Rectangle)
	OnKey   func(k Key) bool
	OnMouse func(m Mouse) bool
	Natural func(most image.Point) image.Point
	// Focus lets the widget hold the keyboard.
	Focus bool
}

func (c *Custom) Layout(r image.Rectangle) { c.Place(r) }

func (c *Custom) Paint(p *Painter) {
	if c.Draw != nil {
		c.Draw(p.Clip(c.rect), c.rect)
	}
}

func (c *Custom) Key(k Key) bool { return c.OnKey != nil && c.OnKey(k) }

func (c *Custom) Mouse(m Mouse) bool { return c.OnMouse != nil && c.OnMouse(m) }

func (c *Custom) CanFocus() bool { return c.Focus }

func (c *Custom) Size(most image.Point) image.Point {
	if c.Natural != nil {
		return c.Natural(most)
	}
	return most
}

// Hole is a window onto what is drawn under the cells: it clears its
// rectangle and tells OnPlace where that is, so the map (or any picture the
// window draws itself) can go there.
type Hole struct {
	Base
	OnPlace func(r image.Rectangle)
}

func (h *Hole) Layout(r image.Rectangle) { h.Place(r) }

func (h *Hole) Paint(p *Painter) {
	p.Fill(h.rect, Clear)
	if h.OnPlace != nil {
		h.OnPlace(h.rect)
	}
}
