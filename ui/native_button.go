package ui

import (
	"image"

	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/window/kit"
)

// A nativeButton stands where a tview.Button stood: the same calls, drawn by
// the kit, and a click presses it. Like tview's, it is pressed by Enter only.
type nativeButton struct {
	*nativeHost
	kit.Base
	btn *kit.Button
}

// newNativeButton is a button as tview makes one: the label on the contrast
// ground, inverted while it has the focus.
func newNativeButton(label string) *nativeButton {
	b := &nativeButton{btn: kit.NewButton(label, nil)}
	b.btn.Style = kit.Style{Fg: kit.Role(theme.RoleText), Bg: kit.Role(theme.RoleSelection)}
	b.btn.FocusStyle = kit.Style{Fg: kit.Role(theme.RoleBackground), Bg: kit.Role(theme.RoleText)}
	b.nativeHost = newNativeHost(b)
	b.SetRect(0, 0, kit.Width(label)+4, 1)
	return b
}

func styleOf(fg, bg kit.Color) kit.Style { return kit.Style{Fg: fg, Bg: bg} }

func (b *nativeButton) SetLabel(s string) *nativeButton {
	b.btn.Label = s
	return b
}

func (b *nativeButton) GetLabel() string { return b.btn.Label }

func (b *nativeButton) SetSelectedFunc(fn func()) *nativeButton {
	b.btn.OnPress = fn
	return b
}

func (b *nativeButton) SetLabelColor(c kit.Color) *nativeButton {
	b.btn.Style.Fg = c
	return b
}

// SetBackgroundColor is accepted and does nothing, as on a tview.Button: its
// draw overwrites the box ground with the style's, so the fill styleFilledButton
// asks for has never shown at rest. The screens look as they did.
func (b *nativeButton) SetBackgroundColor(kit.Color) *nativeButton { return b }

func (b *nativeButton) Children() []kit.Widget { return []kit.Widget{b.btn} }

func (b *nativeButton) Layout(r image.Rectangle) {
	b.Place(r)
	b.btn.Layout(r)
}

// Paint shows the button as it is now: with the focus tview's primitive holds.
func (b *nativeButton) Paint(p *kit.Painter) {
	b.btn.SetFocus(b.HasFocus())
	b.btn.Paint(p)
}

// Key: Enter presses; every other key is the caller's.
func (b *nativeButton) Key(k kit.Key) bool {
	if k.Code != kit.KeyEnter {
		return false
	}
	if b.btn.OnPress != nil {
		b.btn.OnPress()
	}
	return true
}
