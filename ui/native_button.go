package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/window/kit"
)

// A nativeButton stands where a tview.Button stood: the same calls, drawn by
// the kit, and a click presses it. Like tview's, it is pressed by Enter only.
type nativeButton struct {
	*nativeHost
	btn *kit.Button
}

// newNativeButton is a button as tview makes one: the label on the contrast
// ground, inverted while it has the focus.
func newNativeButton(label string) *nativeButton {
	b := &nativeButton{btn: kit.NewButton(label, nil)}
	b.btn.Style = styleOf(tview.Styles.PrimaryTextColor, tview.Styles.ContrastBackgroundColor)
	b.btn.FocusStyle = styleOf(tview.Styles.InverseTextColor, tview.Styles.PrimaryTextColor)
	b.nativeHost = newNativeHost(b.btn)
	b.SetRect(0, 0, kit.Width(label)+4, 1)
	return b
}

func styleOf(fg, bg tcell.Color) kit.Style { return kit.Style{Fg: kitColor(fg), Bg: kitColor(bg)} }

func (b *nativeButton) SetLabel(s string) *nativeButton {
	b.btn.Label = s
	return b
}

func (b *nativeButton) GetLabel() string { return b.btn.Label }

func (b *nativeButton) SetSelectedFunc(fn func()) *nativeButton {
	b.btn.OnPress = fn
	return b
}

func (b *nativeButton) SetLabelColor(c tcell.Color) *nativeButton {
	b.btn.Style.Fg = kitColor(c)
	return b
}

// SetBackgroundColor is accepted and does nothing, as on a tview.Button: its
// draw overwrites the box ground with the style's, so the fill styleFilledButton
// asks for has never shown at rest. The screens look as they did.
func (b *nativeButton) SetBackgroundColor(tcell.Color) *nativeButton { return b }

// Draw shows the button as it is now: with the focus tview's primitive holds.
func (b *nativeButton) Draw(scr tcell.Screen) {
	b.btn.SetFocus(b.HasFocus())
	b.nativeHost.Draw(scr)
}

// InputHandler: Enter presses; every other key is the caller's.
func (b *nativeButton) InputHandler() func(*tcell.EventKey, func(tview.Primitive)) {
	return keysIn(b.Box, func(ev *kit.Key) {
		if ev.Code == kit.KeyEnter && b.btn.OnPress != nil {
			b.btn.OnPress()
		}
	})
}
