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
type nativeHost struct {
	*tview.Box
	root *kit.Root
	buf  kit.Buffer
	// at is the host's corner on the screen and seen the frame it was last
	// drawn in, for the mouse (see nativeScreen).
	at   image.Point
	seen uint64
}

// newNativeHost hosts w. A screen that embeds the host calls it last, once
// its own fields are set: the tree is laid out here.
func newNativeHost(w kit.Widget) *nativeHost {
	return &nativeHost{Box: tview.NewBox(), root: kit.NewRoot(w)}
}

// Draw lays the tree out in the host's rectangle, paints it and copies the
// cells to the screen.
func (h *nativeHost) Draw(scr tcell.Screen) {
	x, y, w, ht := h.GetRect()
	if w <= 0 || ht <= 0 {
		return
	}
	h.root.Resize(w, ht)
	h.root.Paint(&h.buf)
	if ns, ok := scr.(*nativeScreen); ok {
		ns.begin(h, image.Rect(x, y, x+w, y+ht))
		defer ns.end()
	}
	for cy := 0; cy < h.buf.H; cy++ {
		for cx := 0; cx < h.buf.W; cx++ {
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
	if h.buf.CursorOn && h.HasFocus() {
		scr.ShowCursor(x+h.buf.Cursor.X, y+h.buf.Cursor.Y)
	}
}

// InputHandler gives the tree the keys tview gives the host.
func (h *nativeHost) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) {
	return h.WrapInputHandler(func(ev *tcell.EventKey, _ func(tview.Primitive)) {
		if k, ok := kitKey(ev); ok {
			h.root.Key(k)
		}
	})
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

// keyCodes are tcell's keys that are not text, as the kit's.
var keyCodes = map[tcell.Key]kit.KeyCode{
	tcell.KeyEnter: kit.KeyEnter, tcell.KeyEsc: kit.KeyEsc, tcell.KeyTab: kit.KeyTab, tcell.KeyBacktab: kit.KeyBacktab,
	tcell.KeyBackspace: kit.KeyBackspace, tcell.KeyBackspace2: kit.KeyBackspace, tcell.KeyDelete: kit.KeyDelete, tcell.KeyInsert: kit.KeyInsert,
	tcell.KeyUp: kit.KeyUp, tcell.KeyDown: kit.KeyDown, tcell.KeyLeft: kit.KeyLeft, tcell.KeyRight: kit.KeyRight,
	tcell.KeyHome: kit.KeyHome, tcell.KeyEnd: kit.KeyEnd, tcell.KeyPgUp: kit.KeyPgUp, tcell.KeyPgDn: kit.KeyPgDn,
	tcell.KeyF1: kit.KeyF1, tcell.KeyF2: kit.KeyF2, tcell.KeyF3: kit.KeyF3, tcell.KeyF4: kit.KeyF4, tcell.KeyF5: kit.KeyF5, tcell.KeyF6: kit.KeyF6,
	tcell.KeyF7: kit.KeyF7, tcell.KeyF8: kit.KeyF8, tcell.KeyF9: kit.KeyF9, tcell.KeyF10: kit.KeyF10, tcell.KeyF11: kit.KeyF11, tcell.KeyF12: kit.KeyF12,
}

// kitKey is a tcell key event as the kit's; ok is false for a key the kit
// has no name for. Ctrl with a letter comes out as the letter with ModCtrl,
// whichever of tcell's two ways it arrived.
func kitKey(ev *tcell.EventKey) (kit.Key, bool) {
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
		return kit.Key{Code: kit.KeyRune, Rune: ev.Rune(), Mod: mod}, true
	}
	if code, ok := keyCodes[k]; ok {
		return kit.Key{Code: code, Mod: mod}, true
	}
	if k >= tcell.KeyCtrlA && k <= tcell.KeyCtrlZ {
		return kit.Key{Code: kit.KeyRune, Rune: 'a' + rune(k-tcell.KeyCtrlA), Mod: mod | kit.ModCtrl}, true
	}
	return kit.Key{}, false
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

// begin marks r as h's cells: what tview draws over them afterward takes
// them back.
func (s *nativeScreen) begin(h *nativeHost, r image.Rectangle) {
	if w, ht := s.Size(); w != s.w || ht != s.h {
		s.w, s.h = w, ht
		s.owner = make([]*nativeHost, w*ht)
	}
	h.at, h.seen, s.cur = r.Min, s.frame, h
	r = r.Intersect(image.Rect(0, 0, s.w, s.h))
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			s.owner[y*s.w+x] = h
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
