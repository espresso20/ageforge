package ui

import (
	"image"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/window/kit"
)

// native.go is the bridge between the kit (window/kit) and tview, for as long
// as both are here. A screen moved to the kit still has to sit in tview's
// pages and take tview's focus, so it is hosted: a nativeHost is a tview
// primitive that shows a kit tree, by painting the tree into a buffer of
// cells and copying them to the screen. The window and the terminal both show
// it, which is why a ported screen needs no tview version beside it.
//
// A ported screen embeds *nativeHost (so it is still the primitive the rest
// of the UI knows) and is the kit widget the host shows: Draw becomes Paint,
// InputHandler becomes Key, and Mouse is new. See design-private/port.

// A nativeHost shows a kit tree where tview expects a primitive.
//
// While tview is here it still says who has the keyboard: each primitive's
// own focus flag is the truth, and the tree's focus is read from the flags
// before a key goes in (syncFocus).
type nativeHost struct {
	*tview.Box
	root *kit.Root
	self kit.Widget // the widget the host shows
	buf  kit.Buffer
	// capture sees the host's keys before its widget does (SetInputCapture).
	capture func(*kit.Key) *kit.Key
	// at is the host's corner on the screen and seen the frame it was last
	// drawn in, for the mouse (see nativeScreen).
	at   image.Point
	seen uint64
}

// newNativeHost hosts w. A screen that embeds the host calls it last, once
// its own fields are set: the tree is laid out here.
func newNativeHost(w kit.Widget) *nativeHost {
	h := &nativeHost{Box: tview.NewBox()}
	h.showWidget(w)
	return h
}

// showWidget makes w the widget the host shows. A type that embeds one of the
// drop-in widgets and adds to its Layout, Paint or Key calls it with itself,
// or the host would go on showing the widget inside.
func (h *nativeHost) showWidget(w kit.Widget) {
	h.self, h.root = w, kit.NewRoot(w)
	h.root.Focus(w)
}

// Draw lays the tree out in the host's rectangle, paints it and puts it on
// the screen, back to front: the cells of what the kit draws (all of them,
// or only the rectangles a widget says it covers), and in its place among
// them anything in the tree that still draws itself.
func (h *nativeHost) Draw(scr tcell.Screen) {
	x, y, w, ht := h.GetRect()
	h.root.Resize(max(w, 0), max(ht, 0))
	h.root.Paint(&h.buf) // with no room too: what is inside is still placed
	if w <= 0 || ht <= 0 {
		return
	}
	ns, _ := scr.(*nativeScreen)
	if ns != nil {
		ns.begin(h, image.Pt(x, y))
	}
	h.emit(scr, ns, h.self)
	if ns != nil {
		ns.end()
	}
	// The widget answers for the focus, not the host's own box: a container
	// has it when something inside it does.
	focused := h.Box.HasFocus()
	if f, ok := h.self.(interface{ HasFocus() bool }); ok {
		focused = f.HasFocus()
	}
	if h.buf.CursorOn && focused {
		scr.ShowCursor(x+h.buf.Cursor.X, y+h.buf.Cursor.Y)
	}
}

// emit puts w on the screen. A widget with nothing in it that draws itself
// is its covered cells, copied at once. Otherwise it is its own cells, then
// what is inside it in order, so a primitive that draws itself lies over
// what is under it and under what is over it.
func (h *nativeHost) emit(scr tcell.Screen, ns *nativeScreen, w kit.Widget) {
	if f, ok := w.(*foreign); ok {
		if ns != nil {
			ns.end() // its cells are tview's, not this host's
		}
		f.prim.Draw(scr)
		if ns != nil {
			ns.begin(h, h.at)
		}
		return
	}
	if !hasForeign(w) {
		h.copy(scr, ns, coveredBy(w)...)
		return
	}
	if o, ok := w.(interface{ own() []image.Rectangle }); ok {
		h.copy(scr, ns, o.own()...)
	}
	if p, ok := w.(kit.Parent); ok {
		for _, c := range p.Children() {
			h.emit(scr, ns, c)
		}
	}
}

// copy writes the buffer's cells in rects to the screen.
func (h *nativeHost) copy(scr tcell.Screen, ns *nativeScreen, rects ...image.Rectangle) {
	x, y, _, _ := h.GetRect()
	full := image.Rect(0, 0, h.buf.W, h.buf.H)
	for _, r := range rects {
		r = r.Intersect(full)
		if ns != nil {
			ns.claim(r.Add(image.Pt(x, y)))
		}
		for cy := r.Min.Y; cy < r.Max.Y; cy++ {
			for cx := r.Min.X; cx < r.Max.X; cx++ {
				c := &h.buf.Cells[cy*h.buf.W+cx]
				if c.Cont {
					continue // the right half of a wide glyph: the screen keeps it for the left
				}
				var comb []rune
				if c.Comb != "" {
					comb = []rune(c.Comb)
				}
				scr.SetContent(x+cx, y+cy, c.R, comb, cellStyle(c))
			}
		}
	}
}

// hasForeign reports whether w is, or has in it, a primitive that draws
// itself.
func hasForeign(w kit.Widget) (found bool) {
	walkWidgets(w, func(c kit.Widget) {
		if _, ok := c.(*foreign); ok {
			found = true
		}
	})
	return found
}

// InputHandler gives the tree the keys tview gives the host.
func (h *nativeHost) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) {
	return func(ev *tcell.EventKey, _ func(tview.Primitive)) {
		h.syncFocus()
		h.root.Key(kitKey(ev))
	}
}

// boxFocused reports whether tview has given this very primitive the focus
// (HasFocus also answers for what is inside a container).
func (h *nativeHost) boxFocused() bool { return h.Box.HasFocus() }

// syncFocus points the tree's keys at the primitive in it that tview has
// given the focus, or at the host's own widget when none has it.
func (h *nativeHost) syncFocus() {
	at := h.self
	walkWidgets(h.self, func(w kit.Widget) {
		if f, ok := w.(interface{ boxFocused() bool }); ok && f.boxFocused() {
			at = w
		}
	})
	h.root.Focus(at)
}

// SetInputCapture sets what sees the host's keys before its widget does. It
// returns the key to go on with, or nil to stop it there.
func (h *nativeHost) SetInputCapture(fn func(*kit.Key) *kit.Key) *nativeHost {
	h.capture = fn
	return h
}

// GetInputCapture is the capture as tview hands one out, for the tests that
// ask a screen what it does with a key.
func (h *nativeHost) GetInputCapture() func(*tcell.EventKey) *tcell.EventKey {
	return tcellCapture(h.capture)
}

// FilterKey is the capture, where the kit asks for it: on a key's way down
// to whatever has the focus.
func (h *nativeHost) FilterKey(k kit.Key) (kit.Key, bool) {
	if h.capture == nil {
		return k, true
	}
	out := h.capture(&k)
	if out == nil {
		return k, false
	}
	return *out, true
}

// keysIn is the key handler tview asks of box, for a handler that reads the
// kit's keys: every screen does, and tview's keys stop here.
func keysIn(box *tview.Box, fn func(k *kit.Key)) func(*tcell.EventKey, func(tview.Primitive)) {
	return box.WrapInputHandler(func(ev *tcell.EventKey, _ func(tview.Primitive)) {
		k := kitKey(ev)
		fn(&k)
	})
}

// tcellCapture is a capture that reads the kit's keys, as tview wants one:
// for the boxes tview still owns.
func tcellCapture(fn func(*kit.Key) *kit.Key) func(*tcell.EventKey) *tcell.EventKey {
	if fn == nil {
		return nil
	}
	return func(ev *tcell.EventKey) *tcell.EventKey {
		k := kitKey(ev)
		in := k
		switch out := fn(&k); {
		case out == nil:
			return nil
		case *out == in:
			return ev
		default:
			return tcellKey(*out)
		}
	}
}

// newKey is a key press, for the code that makes one up.
func newKey(code kit.KeyCode, r rune, mod kit.Mod) *kit.Key {
	return &kit.Key{Code: code, Rune: r, Mod: mod}
}

// cellAttrs pairs the kit's attributes with tcell's. Underline is not here:
// tcell keeps it apart from the others.
var cellAttrs = [...]struct {
	kit   kit.Attr
	tcell tcell.AttrMask
}{
	{kit.Bold, tcell.AttrBold}, {kit.Italic, tcell.AttrItalic}, {kit.Dim, tcell.AttrDim},
	{kit.Reverse, tcell.AttrReverse}, {kit.Strike, tcell.AttrStrikeThrough}, {kit.Blink, tcell.AttrBlink},
}

// cellStyle is a kit cell's colors and attributes as a tcell style. A cell
// with no ground gets theme.Transparent, which the window shows through.
func cellStyle(c *kit.Cell) tcell.Style {
	bg := theme.Transparent
	if !c.Clear() {
		bg = theme.FromRGBA(c.Bg)
	}
	st := tcell.StyleDefault.Foreground(theme.FromRGBA(c.Fg)).Background(bg)
	if c.Attr == 0 {
		return st
	}
	var mask tcell.AttrMask
	for _, a := range cellAttrs {
		if c.Attr&a.kit != 0 {
			mask |= a.tcell
		}
	}
	st = st.Attributes(mask)
	if c.Attr&kit.Underline != 0 {
		st = st.Underline(true)
	}
	return st
}

// kitStyle is a tcell style as the kit's: the reverse of cellStyle.
func kitStyle(st tcell.Style) kit.Style {
	fg, bg, mask := st.Decompose()
	out := kit.Style{Fg: kitColor(fg), Bg: kitColor(bg)}
	for _, a := range cellAttrs {
		if mask&a.tcell != 0 {
			out.Attr |= a.kit
		}
	}
	if st.GetUnderlineStyle() != tcell.UnderlineStyleNone {
		out.Attr |= kit.Underline
	}
	return out
}

// kitColor is a tcell color as the kit's: a theme sentinel is its role (so it
// still follows the theme), the default is nothing, and the rest are literal.
func kitColor(c tcell.Color) kit.Color {
	if role, ok := theme.RefRole(c); ok {
		return kit.Role(role)
	}
	if c == theme.Transparent {
		return kit.Clear
	}
	r, g, b := c.RGB()
	if r < 0 {
		return kit.Color{}
	}
	return kit.RGB(uint8(r), uint8(g), uint8(b))
}

// keyCodes are tcell's keys that are not text, as the kit's.
var keyCodes = map[tcell.Key]kit.KeyCode{
	tcell.KeyEnter: kit.KeyEnter, tcell.KeyEsc: kit.KeyEsc, tcell.KeyTab: kit.KeyTab, tcell.KeyBacktab: kit.KeyBacktab,
	tcell.KeyBackspace: kit.KeyBackspace, tcell.KeyBackspace2: kit.KeyBackspace, tcell.KeyDelete: kit.KeyDelete, tcell.KeyInsert: kit.KeyInsert,
	tcell.KeyUp: kit.KeyUp, tcell.KeyDown: kit.KeyDown, tcell.KeyLeft: kit.KeyLeft, tcell.KeyRight: kit.KeyRight,
	tcell.KeyHome: kit.KeyHome, tcell.KeyEnd: kit.KeyEnd, tcell.KeyPgUp: kit.KeyPgUp, tcell.KeyPgDn: kit.KeyPgDn,
	tcell.KeyF1: kit.KeyF1, tcell.KeyF2: kit.KeyF2, tcell.KeyF3: kit.KeyF3, tcell.KeyF4: kit.KeyF4, tcell.KeyF5: kit.KeyF5, tcell.KeyF6: kit.KeyF6,
	tcell.KeyF7: kit.KeyF7, tcell.KeyF8: kit.KeyF8, tcell.KeyF9: kit.KeyF9, tcell.KeyF10: kit.KeyF10, tcell.KeyF11: kit.KeyF11, tcell.KeyF12: kit.KeyF12,
}

// kitKey is a tcell key event as the kit's. A character with Ctrl or Alt held
// comes out as a chord, whichever of tcell's two ways it arrived (the terminal
// and the window differ), and a key the kit has no name for as kit.KeyOther.
func kitKey(ev *tcell.EventKey) kit.Key {
	var mod kit.Mod
	if m := ev.Modifiers(); m != 0 {
		if m&tcell.ModShift != 0 {
			mod |= kit.ModShift
		}
		if m&tcell.ModCtrl != 0 {
			mod |= kit.ModCtrl
		}
		if m&tcell.ModAlt != 0 {
			mod |= kit.ModAlt
		}
	}
	k := ev.Key()
	if k == tcell.KeyRune {
		if mod&(kit.ModCtrl|kit.ModAlt) != 0 {
			return kit.Key{Code: kit.KeyChord, Rune: ev.Rune(), Mod: mod}
		}
		return kit.Key{Code: kit.KeyRune, Rune: ev.Rune(), Mod: mod}
	}
	if code, ok := keyCodes[k]; ok {
		return kit.Key{Code: code, Mod: mod}
	}
	if k >= tcell.KeyCtrlA && k <= tcell.KeyCtrlZ {
		return kit.Key{Code: kit.KeyChord, Rune: 'a' + rune(k-tcell.KeyCtrlA), Mod: mod | kit.ModCtrl}
	}
	return kit.Key{Code: kit.KeyOther, Mod: mod}
}

// tcellKey is a kit key as tcell's: the reverse of kitKey, for the few
// places that still hand a key to tview.
func tcellKey(k kit.Key) *tcell.EventKey {
	var mod tcell.ModMask
	if k.Mod&kit.ModShift != 0 {
		mod |= tcell.ModShift
	}
	if k.Mod&kit.ModCtrl != 0 {
		mod |= tcell.ModCtrl
	}
	if k.Mod&kit.ModAlt != 0 {
		mod |= tcell.ModAlt
	}
	switch {
	case k.Code == kit.KeyChord && mod == tcell.ModCtrl && k.Rune >= 'a' && k.Rune <= 'z':
		return tcell.NewEventKey(tcell.KeyCtrlA+tcell.Key(k.Rune-'a'), 0, mod)
	case k.Code == kit.KeyRune, k.Code == kit.KeyChord:
		return tcell.NewEventKey(tcell.KeyRune, k.Rune, mod)
	case k.Code == kit.KeyBackspace:
		return tcell.NewEventKey(tcell.KeyBackspace2, 0, mod)
	}
	for tk, code := range keyCodes {
		if code == k.Code {
			return tcell.NewEventKey(tk, 0, mod)
		}
	}
	return tcell.NewEventKey(tcell.KeyF64, 0, mod)
}

// A nativeScreen is the app's screen in a window. It notes which host wrote
// each cell last, so the mouse goes to the host that is really under the
// pointer and not to one a tview page has covered.
type nativeScreen struct {
	tcell.Screen
	w, h  int
	owner []*nativeHost // who wrote each cell last; nil for tview's own widgets
	cur   *nativeHost   // the host being drawn
	frame uint64        // frames shown so far
	over  *nativeHost   // the host the pointer is over
	held  *nativeHost   // the host that took a press, until the button comes up
}

func newNativeScreen(s tcell.Screen) *nativeScreen { return &nativeScreen{Screen: s} }

// begin says h is being drawn, its corner at pt. What tview draws afterward
// over the cells h claims takes them back.
func (s *nativeScreen) begin(h *nativeHost, pt image.Point) {
	if w, ht := s.Size(); w != s.w || ht != s.h {
		s.w, s.h = w, ht
		s.owner = make([]*nativeHost, w*ht)
	}
	h.at, h.seen, s.cur = pt, s.frame, h
}

// claim marks r as the cells of the host being drawn.
func (s *nativeScreen) claim(r image.Rectangle) {
	r = r.Intersect(image.Rect(0, 0, s.w, s.h))
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			s.owner[y*s.w+x] = s.cur
		}
	}
}

func (s *nativeScreen) end() { s.cur = nil }

func (s *nativeScreen) SetContent(x, y int, r rune, comb []rune, st tcell.Style) {
	if x >= 0 && y >= 0 && x < s.w && y < s.h {
		s.owner[y*s.w+x] = s.cur
	}
	s.Screen.SetContent(x, y, r, comb, st)
}

func (s *nativeScreen) Show() {
	s.frame++
	s.Screen.Show()
}

// hostAt is the host whose cell is under x, y in the frame on show.
func (s *nativeScreen) hostAt(x, y int) *nativeHost {
	if x < 0 || y < 0 || x >= s.w || y >= s.h {
		return nil
	}
	if h := s.owner[y*s.w+x]; h != nil && h.seen+1 == s.frame {
		return h
	}
	return nil
}

// mouse gives a mouse event (in screen cells) to the host under the pointer,
// or to the one holding a press, and reports whether the screen may look
// different for it.
func (s *nativeScreen) mouse(m kit.Mouse) (changed bool) {
	if m.Kind == kit.MouseLeave {
		s.held = nil // the pointer left the window: the press is over
	}
	to := s.held
	if to == nil && m.Kind != kit.MouseLeave {
		to = s.hostAt(m.X, m.Y)
	}
	if s.over != nil && s.over != to {
		changed = s.over.root.Mouse(kit.Mouse{Kind: kit.MouseLeave})
	}
	s.over = to
	if to == nil {
		return changed
	}
	switch m.Kind {
	case kit.MousePress:
		s.held = to
	case kit.MouseRelease:
		s.held = nil
	}
	m.X, m.Y = m.X-to.at.X, m.Y-to.at.Y
	return to.root.Mouse(m) || changed
}
