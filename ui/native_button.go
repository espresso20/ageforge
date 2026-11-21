package ui

import (
	"image"

	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/window/kit"
)

// A nativeButton is a button drawn by the kit, and a click presses it. The
// keyboard presses it with Enter only.
type nativeButton struct {
	*nativeHost
	kit.Base
	btn *kit.Button
	// act, when set, draws the button as a one-row button of that kind
	// (chip.go); plates are the pictures it has drawn.
	act    keyKind
	acting bool
	// key is the key the button names on its cap ("" for none).
	key string
	// off draws the button unavailable.
	off    bool
	ground kit.Color
	plates keyPresser
}

// newNativeButton is a button: the label on the contrast
// ground, inverted while it has the focus.
func newNativeButton(label string) *nativeButton {
	b := &nativeButton{btn: kit.NewButton(label, nil)}
	b.btn.Style = kit.Style{Fg: kit.Role(theme.RoleText), Bg: kit.Role(theme.RoleSelection)}
	b.btn.FocusStyle = kit.Style{Fg: kit.Role(theme.RoleBackground), Bg: kit.Role(theme.RoleText)}
	b.nativeHost = newNativeHost()
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

// SetBackgroundColor is accepted and does nothing: its
// draw overwrites the box ground with the style's, so the fill styleFilledButton
// asks for has never shown at rest. The screens look as they did.
func (b *nativeButton) SetBackgroundColor(kit.Color) *nativeButton { return b }

func (b *nativeButton) Children() []kit.Widget { return []kit.Widget{b.btn} }

func (b *nativeButton) Layout(r image.Rectangle) {
	b.Place(r)
	b.btn.Layout(r)
}

// asAction makes the button a plate of kind, with the key it names.
func (b *nativeButton) asAction(kind keyKind, ground kit.Color, key string) *nativeButton {
	b.act, b.acting, b.ground, b.key = kind, true, ground, key
	return b
}

// actionWidth is the cells the button takes as a plate.
func (b *nativeButton) actionWidth() int { return chipW(b.btn.Label, b.key) }

// Paint shows the button as it is now: with the focus the primitive holds.
func (b *nativeButton) Paint(p *kit.Painter) {
	b.btn.SetFocus(b.HasFocus())
	if b.acting {
		b.paintAction(p)
		return
	}
	b.btn.Paint(p)
}

// paintAction draws the button as a plate: lit while the pointer
// or the keyboard is on it, sunk while it is held.
func (b *nativeButton) paintAction(p *kit.Painter) {
	r := b.Bounds()
	hover, down := b.btn.State()
	st := keyRest
	switch {
	case b.off:
		st = keyOff
	case down:
		st = keyPressed
	case hover || b.HasFocus():
		st = keyHover
	}
	g := newXGrid(r.Dx(), r.Dy())
	g.chip(0, 0, chipSpec{kind: b.act, label: b.btn.Label, key: b.key, off: b.off})
	pal := newExpPalette()
	pal.styles[xText] = pal.styles[xText].Background(b.ground)
	b.plates.paintKey(p, g.keyset(), r.Min, pal, g.keys[0], st, pictureCell != nil)
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
