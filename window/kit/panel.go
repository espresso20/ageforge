package kit

import (
	"image"

	"github.com/espresso20/ageforge/theme"
)

// Panel is a box around one child: a border with a title on it, a foot on
// its bottom edge, a ground and some padding. Every part is optional.
type Panel struct {
	Base
	Child Widget
	// Border is the line the box is drawn with, in BorderStyle (unset: the
	// theme's border color).
	Border      Border
	BorderStyle Style
	// TitleStyle is the title's style (unset: the accent color).
	TitleStyle Style
	// Title sits on the top edge and Foot on the bottom one, both tagged
	// text, placed by their Align (the title starts centered).
	Title, Foot           string
	TitleAlign, FootAlign Align
	// Ground fills the box; unset leaves what is under it.
	Ground Color
	// PadX and PadY are blank cells inside the border.
	PadX, PadY int
}

// NewPanel is a box with a line border and a centered title around child.
func NewPanel(title string, child Widget) *Panel {
	return &Panel{Child: child, Border: BorderLine, Title: title, TitleAlign: AlignCenter, FootAlign: AlignCenter}
}

func (b *Panel) Children() []Widget {
	if b.Child == nil {
		return nil
	}
	return []Widget{b.Child}
}

// edge is the cells the border and the padding take on each side.
func (b *Panel) edge() (x, y int) {
	if b.Border != BorderNone {
		return b.PadX + 1, b.PadY + 1
	}
	return b.PadX, b.PadY
}

// Inner is the rectangle the child gets.
func (b *Panel) Inner() image.Rectangle {
	x, y := b.edge()
	r := image.Rect(b.rect.Min.X+x, b.rect.Min.Y+y, b.rect.Max.X-x, b.rect.Max.Y-y)
	if r.Dx() < 0 || r.Dy() < 0 {
		return image.Rectangle{}
	}
	return r
}

func (b *Panel) Layout(r image.Rectangle) {
	b.Place(r)
	if b.Child != nil {
		b.Child.Layout(b.Inner())
	}
}

func (b *Panel) Size(most image.Point) image.Point {
	x, y := b.edge()
	in := image.Pt(max(most.X-2*x, 0), max(most.Y-2*y, 0))
	if s, ok := b.Child.(Sizer); ok {
		in = s.Size(in)
	}
	return image.Pt(min(in.X+2*x, most.X), min(in.Y+2*y, most.Y))
}

func (b *Panel) Paint(p *Painter) {
	r := b.rect
	p = p.Clip(r)
	p.Fill(r, b.Ground)
	st := b.BorderStyle
	if !st.Fg.IsSet() {
		st.Fg = Role(theme.RoleBorder)
	}
	p.Box(r, b.Border, st)
	if w := r.Dx() - 2; w > 0 {
		title := b.TitleStyle
		if !title.Fg.IsSet() {
			title.Fg = Role(theme.RoleAccent)
		}
		if b.Title != "" {
			p.PrintAligned(r.Min.X+1, r.Min.Y, w, b.Title, title, b.TitleAlign)
		}
		if b.Foot != "" && r.Dy() > 1 {
			p.PrintAligned(r.Min.X+1, r.Max.Y-1, w, b.Foot, Fg(theme.RoleDim), b.FootAlign)
		}
	}
	if b.Child != nil && !hidden(b.Child) {
		b.Child.Paint(p.Clip(b.Inner()))
	}
}
