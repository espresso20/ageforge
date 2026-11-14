package ui

import (
	"image"
	"strings"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/window/kit"
)

// A nativeText stands where a tview.TextView stood. It answers the calls the
// UI makes of a text view, with the same names, and is drawn by the kit: a
// kit.Text in a kit.Scroll in a kit.Panel, on a host. So a text view moves
// to the kit by changing its constructor, and gets the mouse wheel with it.
//
// It goes when the screens around it are kit trees: its Panel then sits in
// their layout directly.
type nativeText struct {
	*nativeHost
	kit.Base
	body *kit.Text
	view *kit.Scroll
	box  *kit.Panel

	raw        string // the text as it was set
	dynamic    bool   // tags are read; otherwise the text prints as written
	wrap, word bool
	border     bool
	scrollable bool
	maxLines   int
	changed    func()
	done       func(kit.KeyCode)
	drawFunc   func(scr tcell.Screen, x, y, w, h int) (int, int, int, int)
}

// newNativeText is a text view as tview makes one: scrollable, wrapping at
// the edge, tags not read until SetDynamicColors says so.
func newNativeText() *nativeText {
	t := &nativeText{body: &kit.Text{Wrap: kit.WrapChar}, wrap: true, scrollable: true}
	t.view = kit.NewScroll(t.body)
	t.view.Focus, t.view.NoBar = true, true
	t.box = &kit.Panel{Child: t.view, Ground: kit.Role(theme.RoleBackground), TitleAlign: kit.AlignCenter}
	t.nativeHost = newNativeHost(t)
	return t
}

func (t *nativeText) Children() []kit.Widget { return []kit.Widget{t.box} }

// Layout places the box, after telling the draw func (if any) the size.
func (t *nativeText) Layout(r image.Rectangle) {
	t.Place(r)
	if t.nativeHost == nil {
		return // newNativeHost lays the tree out before the host is set
	}
	if t.drawFunc != nil {
		x, y, w, h := t.GetRect()
		t.drawFunc(nil, x, y, w, h)
	}
	t.box.Border = focusBorder(t.border, t.HasFocus())
	t.box.Layout(r)
}

// focusBorder is the border of a box that has one: a line, doubled while the
// box has the focus, as tview draws it.
func focusBorder(on, focus bool) kit.Border {
	switch {
	case !on:
		return kit.BorderNone
	case focus:
		return kit.BorderDouble
	}
	return kit.BorderLine
}

func (t *nativeText) SetDynamicColors(on bool) *nativeText {
	t.dynamic = on
	return t.show()
}

// SetRegions is accepted and does nothing: the kit reads region tags always.
func (t *nativeText) SetRegions(bool) *nativeText { return t }

func (t *nativeText) SetWrap(on bool) *nativeText {
	t.wrap = on
	return t.rewrap()
}

func (t *nativeText) SetWordWrap(on bool) *nativeText {
	t.word = on
	return t.rewrap()
}

func (t *nativeText) rewrap() *nativeText {
	switch {
	case !t.wrap:
		t.body.Wrap = kit.WrapNone
	case t.word:
		t.body.Wrap = kit.WrapWord
	default:
		t.body.Wrap = kit.WrapChar
	}
	return t
}

func (t *nativeText) SetTextAlign(align kit.Align) *nativeText {
	t.body.Align = align
	return t
}

// SetScrollable(false) makes it a view of the text's end that takes no
// scrolling keys, as tview does.
func (t *nativeText) SetScrollable(on bool) *nativeText {
	t.scrollable, t.view.Follow = on, !on
	return t
}

// SetMaxLines keeps the text to its last n lines (0: all of it). As in
// tview the older lines go after a draw, so text that was just set is shown
// whole once.
func (t *nativeText) SetMaxLines(n int) *nativeText {
	t.maxLines = n
	return t
}

func (t *nativeText) SetText(s string) *nativeText {
	t.raw = s
	t.show()
	if t.changed != nil {
		t.changed()
	}
	return t
}

// show gives the kit the text as it should print.
func (t *nativeText) show() *nativeText {
	s := t.raw
	if !t.dynamic {
		s = kit.Escape(s)
	}
	t.body.SetText(s)
	return t
}

// GetText is the text as it was set; with strip, without its tags.
func (t *nativeText) GetText(strip bool) string {
	if strip && t.dynamic {
		return kit.Strip(t.raw)
	}
	return t.raw
}

func (t *nativeText) Clear() *nativeText { return t.SetText("") }

func (t *nativeText) ScrollToBeginning() *nativeText {
	t.view.ScrollTo(0)
	return t
}

func (t *nativeText) ScrollToEnd() *nativeText {
	t.view.ToEnd()
	return t
}

// ScrollTo puts row at the top. The column is not kept: the kit's text does
// not scroll sideways.
func (t *nativeText) ScrollTo(row, _ int) *nativeText {
	t.view.ScrollTo(row)
	return t
}

func (t *nativeText) GetScrollOffset() (row, column int) { return t.view.Top(), 0 }

func (t *nativeText) SetChangedFunc(fn func()) *nativeText {
	t.changed = fn
	return t
}

// SetDoneFunc sets what Esc, Enter, Tab and Shift+Tab call.
func (t *nativeText) SetDoneFunc(fn func(key kit.KeyCode)) *nativeText {
	t.done = fn
	return t
}

func (t *nativeText) SetTextColor(c tcell.Color) *nativeText {
	t.body.Style.Fg = kitColor(c)
	return t
}

func (t *nativeText) SetBackgroundColor(c tcell.Color) *nativeText {
	t.box.Ground = kitColor(c)
	return t
}

func (t *nativeText) SetBorder(on bool) *nativeText {
	t.border = on
	return t
}

func (t *nativeText) SetTitle(s string) *nativeText {
	t.box.Title = s
	return t
}

func (t *nativeText) GetTitle() string { return t.box.Title }

func (t *nativeText) SetTitleAlign(align kit.Align) *nativeText {
	t.box.TitleAlign = align
	return t
}

func (t *nativeText) SetBorderColor(c tcell.Color) *nativeText {
	t.box.BorderStyle.Fg = kitColor(c)
	return t
}

func (t *nativeText) SetTitleColor(c tcell.Color) *nativeText {
	t.box.TitleStyle.Fg = kitColor(c)
	return t
}

// SetDrawFunc calls fn with the box's rectangle each time it is laid out, the
// screen nil. Unlike tview's it cannot move the text: the rectangle it
// returns is ignored, so it serves a view that only wants to know its size.
func (t *nativeText) SetDrawFunc(fn func(scr tcell.Screen, x, y, w, h int) (int, int, int, int)) *nativeText {
	t.drawFunc = fn
	return t
}

// Write adds p to the text, so a text view can be a Fprint target.
func (t *nativeText) Write(p []byte) (int, error) {
	t.SetText(t.raw + string(p))
	return len(p), nil
}

// GetInnerRect is the rectangle the text has: the box, less its border.
func (t *nativeText) GetInnerRect() (x, y, w, h int) {
	x, y, w, h = t.GetRect()
	if t.border {
		x, y, w, h = x+1, y+1, w-2, h-2
	}
	return x, y, max(w, 0), max(h, 0)
}

// Paint draws the box.
func (t *nativeText) Paint(p *kit.Painter) {
	t.box.Paint(p)
	if t.maxLines > 0 {
		if ls := strings.Split(t.raw, "\n"); len(ls) > t.maxLines {
			t.raw = strings.Join(ls[len(ls)-t.maxLines:], "\n")
			t.show()
		}
	}
}

// viKeys are the letters a text view scrolls by, as the keys they stand for.
var viKeys = map[rune]kit.KeyCode{'j': kit.KeyDown, 'k': kit.KeyUp, 'g': kit.KeyHome, 'G': kit.KeyEnd}

// Key takes a text view's keys: Esc, Enter and Tab are done, and the arrows,
// the page keys, Home, End and j, k, g, G scroll.
func (t *nativeText) Key(k kit.Key) bool {
	switch k.Code {
	case kit.KeyEsc, kit.KeyEnter, kit.KeyTab, kit.KeyBacktab:
		if t.done != nil {
			t.done(k.Code)
		}
		return true
	}
	if !t.scrollable {
		return false
	}
	switch code, vi := viKeys[k.Rune]; {
	case k.Ctrl('f'):
		k = kit.Key{Code: kit.KeyPgDn}
	case k.Ctrl('b'):
		k = kit.Key{Code: kit.KeyPgUp}
	case vi && k.Code == kit.KeyRune && k.Mod == 0:
		k = kit.Key{Code: code}
	}
	return t.view.Key(k)
}
