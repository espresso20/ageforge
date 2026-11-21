package ui

import (
	"image"
	"strings"

	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/window/kit"
)

// A nativeText is a text view. It answers the calls the UI makes of one, and
// is drawn by the kit: a
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
	drawFunc   func(x, y, w, h int)
	// wheel, when set, takes a turn of the wheel the scroll view had no use
	// for: a box that pages its own text has nothing for the view to move.
	wheel func(rows int) bool
	// click, when set, is what a click on the box does when nothing in
	// its text took the click.
	click func()
	// rest, when set, is told whether the pointer is on the box, so a box a
	// click acts on can show the line it runs.
	rest func(on bool)
	// over, when set, paints on the text once the box has: pictures that go
	// over the cells. It is given the rectangle the text shows in and the
	// text's first row in view.
	over func(p *kit.Painter, text image.Rectangle, top int)
	// pick and hand, when set, are the buttons that over painted: pick takes
	// a mouse event first, hand says where the pointer is a hand.
	pick func(m kit.Mouse) bool
	hand func(at image.Point) kit.Pointer
}

// newNativeText is a text view: scrollable, wrapping at
// the edge, tags not read until SetDynamicColors says so.
func newNativeText() *nativeText {
	t := &nativeText{body: &kit.Text{Wrap: kit.WrapChar}, wrap: true, scrollable: true}
	t.view = kit.NewScroll(t.body)
	t.view.Focus, t.view.NoBar = true, true
	t.box = &kit.Panel{Child: t.view, Ground: kit.Role(theme.RoleBackground), TitleAlign: kit.AlignCenter}
	t.nativeHost = newNativeHost()
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
		t.drawFunc(x, y, w, h)
	}
	t.box.Border = focusBorder(t.border, t.HasFocus())
	t.box.Layout(r)
}

// focusBorder is the border of a box that has one: a line, doubled while the
// box has the focus.
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
// scrolling keys.
func (t *nativeText) SetScrollable(on bool) *nativeText {
	t.scrollable, t.view.Follow = on, !on
	return t
}

// SetMaxLines keeps the text to its last n lines (0: all of it). The older
// lines go after a draw, so text that was just set is shown
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

func (t *nativeText) SetTextColor(c kit.Color) *nativeText {
	t.body.Style.Fg = c
	return t
}

func (t *nativeText) SetBackgroundColor(c kit.Color) *nativeText {
	t.box.Ground = c
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

func (t *nativeText) SetBorderColor(c kit.Color) *nativeText {
	t.box.BorderStyle.Fg = c
	return t
}

func (t *nativeText) SetTitleColor(c kit.Color) *nativeText {
	t.box.TitleStyle.Fg = c
	return t
}

// SetDrawFunc calls fn with the box's rectangle each time it is laid out, for
// a view that wants to know its size.
func (t *nativeText) SetDrawFunc(fn func(x, y, w, h int)) *nativeText {
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

// scrollSpan is where the view stands in its text: the first row in view, how
// many rows show and how many the text has when it wraps to the view's width.
func (t *nativeText) scrollSpan() (top, view, total int) {
	r := t.view.Bounds()
	return t.view.Top(), r.Dy(), t.body.Size(image.Pt(r.Dx(), 1<<20)).Y
}

// Paint draws the box.
func (t *nativeText) Paint(p *kit.Painter) {
	t.box.Paint(p)
	if t.over != nil {
		t.over(p, t.view.Bounds(), t.view.Top())
	}
	if t.maxLines > 0 {
		if ls := strings.Split(t.raw, "\n"); len(ls) > t.maxLines {
			t.raw = strings.Join(ls[len(ls)-t.maxLines:], "\n")
			t.show()
		}
	}
}

// Mouse gives the wheel to the box's own rule and a click to its own
// action, where it has them and nothing in its text took them.
func (t *nativeText) Mouse(m kit.Mouse) bool {
	if t.rest != nil && (m.Kind == kit.MouseMove || m.Kind == kit.MouseLeave) {
		t.rest(m.Kind == kit.MouseMove)
	}
	if t.pick != nil && t.pick(m) {
		return true
	}
	switch {
	case m.Kind == kit.MouseWheel:
		return t.wheel != nil && t.wheel(m.Wheel)
	case m.Kind == kit.MousePress && m.Button == kit.ButtonLeft && t.click != nil:
		t.click()
		return true
	}
	return false
}

// PointerAt: the hand over a box a click acts on.
func (t *nativeText) PointerAt(at image.Point) kit.Pointer {
	if t.hand != nil && t.hand(at) == kit.PointerHand {
		return kit.PointerHand
	}
	return handIf(t.click != nil)
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
