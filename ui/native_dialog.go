package ui

import (
	"image"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/window/kit"
)

// A nativeDialog stands where a tview.Modal stood: text over a row of
// buttons, in a box in the middle of the page, with the same calls, the same
// size and the same keys (Tab and the arrows move between the buttons, Enter
// presses one, Esc is button -1). The kit draws it, so a click presses a
// button too. Around the box the page beneath shows.
type nativeDialog struct {
	*nativeHost
	kit.Base
	app     *tview.Application
	text    string
	labels  []string
	buttons []*nativeButton
	at      int // the button the keyboard is on
	done    func(button int, label string)
	ground  kit.Color
	ink     kit.Color

	box   image.Rectangle // where the dialog is, in the page
	lines []string        // the text, wrapped for it
}

// newNativeDialog is an empty dialog as tview makes one: on the contrast
// ground, in the theme's text color.
func newNativeDialog(app *tview.Application) *nativeDialog {
	d := &nativeDialog{app: app, ground: kit.Role(theme.RoleSelection), ink: kit.Role(theme.RoleText)}
	d.nativeHost = newNativeHost(d)
	return d
}

func (d *nativeDialog) kitDrawn() {}

func (d *nativeDialog) SetText(s string) *nativeDialog {
	d.text = s
	return d
}

// AddButtons adds a button for each label, pressed in the theme's text on
// the page's ground and inverted while the keyboard is on it.
func (d *nativeDialog) AddButtons(labels []string) *nativeDialog {
	for _, label := range labels {
		i, b := len(d.buttons), newNativeButton(label)
		b.btn.Style = kit.Style{Fg: kit.Role(theme.RoleText), Bg: kit.Role(theme.RoleBackground)}
		b.btn.FocusStyle = kit.Style{Fg: kit.Role(theme.RoleBackground), Bg: kit.Role(theme.RoleText)}
		b.SetSelectedFunc(func() {
			if d.done != nil {
				d.done(i, label)
			}
		})
		d.labels, d.buttons = append(d.labels, label), append(d.buttons, b)
	}
	return d
}

// SetDoneFunc sets what a button calls, with its index and label; Esc calls
// it with -1 and no label.
func (d *nativeDialog) SetDoneFunc(fn func(button int, label string)) *nativeDialog {
	d.done = fn
	return d
}

func (d *nativeDialog) SetBackgroundColor(c tcell.Color) *nativeDialog {
	d.ground = kitColor(c)
	return d
}

func (d *nativeDialog) SetTextColor(c tcell.Color) *nativeDialog {
	d.ink = kitColor(c)
	return d
}

func (d *nativeDialog) Children() []kit.Widget {
	out := make([]kit.Widget, len(d.buttons))
	for i, b := range d.buttons {
		out[i] = b
	}
	return out
}

// Layout sizes the dialog as tview's Modal does: a third of the page wide,
// or as wide as its buttons need, and as tall as its wrapped text asks.
func (d *nativeDialog) Layout(r image.Rectangle) {
	d.Place(r)
	if d.nativeHost == nil {
		return // newNativeHost lays the tree out before the host is set
	}
	off := screenOffset(d, r)
	need, row := -2, -1
	for _, l := range d.labels {
		need += kit.Width(l) + 4 + 2
		row += kit.Width(l) + 4 + 1
	}
	width := max(r.Dx()/3, need)
	d.lines = kit.WordWrap(d.text, width)
	w, h := width+4, len(d.lines)+6
	x, y := r.Min.X+(r.Dx()-w)/2, r.Min.Y+(r.Dy()-h)/2
	d.box = image.Rect(x, y, x+w, y+h)

	bx, right, by := x+2, x+2+width, y+3+len(d.lines)
	if bx+row < right {
		bx = (bx + right - row) / 2
	}
	for i, b := range d.buttons {
		bw := min(kit.Width(d.labels[i])+4, right-bx)
		if right-bx < 1 {
			bw = 0 // no room for this one
		}
		placeAt(b, bx, by, bw, 1, off)
		bx += bw + 1
	}
}

// Paint draws the box, its border doubled while the keyboard is in it, the
// text centered, and the buttons.
func (d *nativeDialog) Paint(p *kit.Painter) {
	if d.box.Empty() {
		return
	}
	// The border keeps the contrast ground whatever the dialog's own is, as
	// the Modal's did: only the inside takes SetBackgroundColor.
	p.Fill(d.box, kit.Role(theme.RoleSelection))
	p.Fill(d.box.Inset(1), d.ground)
	p.Box(d.box, focusBorder(true, d.HasFocus()), kit.Style{Fg: kit.Role(theme.RoleBorder)})
	for i, line := range d.lines {
		p.PrintAligned(d.box.Min.X+2, d.box.Min.Y+2+i, d.box.Dx()-4, line, kit.Style{Fg: d.ink}, kit.AlignCenter)
	}
	for _, b := range d.buttons {
		if !b.Bounds().Empty() {
			b.Paint(p)
		}
	}
}

// Covers: the pointer is the dialog's inside its box and nowhere else.
func (d *nativeDialog) Covers(pt image.Point) bool { return pt.In(d.box) }

// covered: the dialog paints its box and nothing round it.
func (d *nativeDialog) covered() []image.Rectangle { return []image.Rectangle{d.box} }

// Key takes what the button with the keyboard left: Tab and the arrows move
// to the next button or the one before, round the ends, and Esc is done.
func (d *nativeDialog) Key(k kit.Key) bool {
	switch k.Code {
	case kit.KeyTab, kit.KeyDown, kit.KeyRight:
		d.move(1)
	case kit.KeyBacktab, kit.KeyUp, kit.KeyLeft:
		d.move(-1)
	case kit.KeyEsc:
		if d.done != nil {
			d.done(-1, "")
		}
	}
	return true
}

func (d *nativeDialog) move(by int) {
	if n := len(d.buttons); n > 0 {
		d.at = ((d.at+by)%n + n) % n
		d.app.SetFocus(d.buttons[d.at])
	}
}

// Focus gives the keyboard to the button it was last on.
func (d *nativeDialog) Focus(delegate func(p tview.Primitive)) {
	if len(d.buttons) == 0 {
		d.Box.Focus(delegate)
		return
	}
	delegate(d.buttons[d.at])
}

// HasFocus reports whether the dialog or one of its buttons has the keyboard.
func (d *nativeDialog) HasFocus() bool {
	for _, b := range d.buttons {
		if b.HasFocus() {
			return true
		}
	}
	return d.Box.HasFocus()
}
