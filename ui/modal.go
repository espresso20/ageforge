package ui

import (
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/theme"
)

// modal.go is what the windows that float over a page have in common: the
// opaque rows they are built from, the frame, the centring, the keys that
// close them, the ring of buttons the keyboard moves around, and the plain
// button dialog. It holds no player text; every string comes from the caller.

// solidBox is an opaque filler in a role's colour. tview only paints cells a
// primitive draws, so every row and column inside a window must belong to a
// primitive that clears it, or the page beneath shows through.
func solidBox(role theme.Role) *tview.Box {
	return tview.NewBox().SetBackgroundColor(theme.Color(role))
}

// surfaceBox is solidBox in the modal surface colour.
func surfaceBox() *tview.Box { return solidBox(theme.RoleSurface) }

// surfaceText is a line (or lines) of text on the modal surface.
func surfaceText(text string, align int) *tview.TextView {
	tv := tview.NewTextView().SetDynamicColors(true).SetTextAlign(align).SetWrap(false).SetText(text)
	tv.SetBackgroundColor(theme.Color(theme.RoleSurface))
	return tv
}

// framed gives a window's rows a border and a title, both in the role's colour.
func framed(inner *tview.Flex, title string, role theme.Role) *tview.Flex {
	inner.SetBorder(true).
		SetTitle(title).
		SetTitleColor(theme.Color(role)).
		SetBorderColor(theme.Color(role))
	return inner
}

// centerModal centers inner at width by height cells. A margin that is nil
// draws nothing, so the page beneath stays visible around the window; a Box
// margin clears it.
func centerModal(inner tview.Primitive, width, height int, clearMargins bool) *tview.Flex {
	margin := func() tview.Primitive {
		if clearMargins {
			return tview.NewBox()
		}
		return nil
	}
	return tview.NewFlex().
		AddItem(margin(), 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(margin(), 0, 1, false).
			AddItem(inner, height, 0, true).
			AddItem(margin(), 0, 1, false),
			width, 0, true).
		AddItem(margin(), 0, 1, false)
}

// centeredModal is a window with the page beneath cleared around it.
func centeredModal(inner tview.Primitive, width, height int) *tview.Flex {
	return centerModal(inner, width, height, true)
}

// floatingModal is a window with the page beneath still showing around it.
func floatingModal(inner tview.Primitive, width, height int) *tview.Flex {
	return centerModal(inner, width, height, false)
}

// keyCapturer is a primitive whose keys can be taken before it sees them.
type keyCapturer interface {
	SetInputCapture(capture func(event *tcell.EventKey) *tcell.EventKey) *tview.Box
}

// takeKeys makes each key in keys run its function on the window and be
// taken; any other key goes on to whatever has the focus.
func takeKeys(window keyCapturer, keys map[tcell.Key]func()) {
	window.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		run, ok := keys[ev.Key()]
		if !ok {
			return ev
		}
		run()
		return nil
	})
}

// showButtonModal shows text with buttons over the page. A button, or Esc
// (button -1), removes the window and then calls done with the button's index
// and label.
func showButtonModal(pages *tview.Pages, page, text string, buttons []string, done func(button int, label string)) *tview.Modal {
	modal := tview.NewModal().
		SetText(text).
		AddButtons(buttons).
		SetDoneFunc(func(button int, label string) {
			pages.RemovePage(page)
			done(button, label)
		})
	pages.AddPage(page, modal, true, true)
	return modal
}

// buttonRing is the buttons of a window and the one the keyboard is on.
// Tab and the arrows move it; a button's letter presses that button.
type buttonRing struct {
	app     *tview.Application
	buttons []*tview.Button
	letters []rune // lower case, one to a button; 0 for none
	at      int
	// moved is told which button has the keyboard, so the window can hand it
	// back there after another window closes.
	moved func(tview.Primitive)
}

func newButtonRing(app *tview.Application, moved func(tview.Primitive)) *buttonRing {
	return &buttonRing{app: app, moved: moved}
}

// add puts a button on the ring; letter presses it (0: no letter).
func (r *buttonRing) add(b *tview.Button, letter rune) {
	r.buttons = append(r.buttons, b)
	r.letters = append(r.letters, unicode.ToLower(letter))
}

// clear empties the ring, for a window whose buttons change.
func (r *buttonRing) clear() {
	r.buttons, r.letters = r.buttons[:0], r.letters[:0]
}

// focus puts the keyboard on button i, counting round the ring.
func (r *buttonRing) focus(i int) {
	n := len(r.buttons)
	if n == 0 {
		return
	}
	r.at = (i%n + n) % n
	b := r.buttons[r.at]
	if r.moved != nil {
		r.moved(b)
	}
	r.app.SetFocus(b)
}

// press does what the button with that letter does.
func (r *buttonRing) press(letter rune) bool {
	for i, l := range r.letters {
		if l != 0 && l == unicode.ToLower(letter) {
			r.buttons[i].InputHandler()(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), nil)
			return true
		}
	}
	return false
}

// key takes Tab and Shift-Tab, Right and Left, and the buttons' letters, and
// reports whether the key was one of them.
func (r *buttonRing) key(ev *tcell.EventKey) bool {
	switch ev.Key() {
	case tcell.KeyTab, tcell.KeyRight:
		r.focus(r.at + 1)
	case tcell.KeyBacktab, tcell.KeyLeft:
		r.focus(r.at - 1)
	case tcell.KeyRune:
		return r.press(ev.Rune())
	default:
		return false
	}
	return true
}
