package kit

import (
	"image"

	"github.com/espresso20/ageforge/theme"
)

// Button is a label to press: Enter or Space while it has the focus, or a
// click.
type Button struct {
	Base
	// Label is tagged text.
	Label   string
	OnPress func()
	// Disabled greys it out; it takes no focus and no press.
	Disabled bool

	hover, down bool
}

// NewButton is a button that calls fn.
func NewButton(label string, fn func()) *Button { return &Button{Label: label, OnPress: fn} }

func (b *Button) CanFocus() bool { return !b.Disabled }

func (b *Button) Layout(r image.Rectangle) { b.Place(r) }

// Size is the label with a blank cell each side, one row high.
func (b *Button) Size(most image.Point) image.Point {
	return image.Pt(min(Width(b.Label)+2, most.X), min(1, most.Y))
}

func (b *Button) style() Style {
	switch {
	case b.Disabled:
		return Style{Fg: Role(theme.RoleDim), Bg: Role(theme.RoleChip)}
	case b.focused || b.down:
		return Style{Fg: Role(theme.RoleOnAccent), Bg: Role(theme.RoleAccent), Attr: Bold}
	case b.hover:
		return Style{Fg: Role(theme.RoleSelectionText), Bg: Role(theme.RoleSelection)}
	}
	return Style{Fg: Role(theme.RoleText), Bg: Role(theme.RoleChip)}
}

func (b *Button) Paint(p *Painter) {
	r := b.rect
	p = p.Clip(r)
	st := b.style()
	p.Fill(r, st.Bg)
	p.PrintAligned(r.Min.X, r.Min.Y+(r.Dy()-1)/2, r.Dx(), b.Label, st, AlignCenter)
}

func (b *Button) press() {
	if !b.Disabled && b.OnPress != nil {
		b.OnPress()
	}
}

func (b *Button) Key(k Key) bool {
	if k.Code == KeyEnter || k.Is(' ') {
		b.press()
		return true
	}
	return false
}

func (b *Button) Mouse(m Mouse) bool {
	switch m.Kind {
	case MouseLeave:
		b.hover = false
	case MouseMove:
		b.hover = m.Pt().In(b.rect)
	case MousePress:
		if m.Button != ButtonLeft || b.Disabled {
			return false
		}
		b.down = true
	case MouseRelease:
		// The press counts when the button comes up over it.
		was := b.down
		b.down = false
		if was && m.Pt().In(b.rect) {
			b.press()
		}
	default:
		return false
	}
	return true
}
