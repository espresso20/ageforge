package ui

import (
	"image"
	"strings"

	"github.com/espresso20/ageforge/window/kit"
)

// text_buttons.go is what a text overlay needs to carry buttons: a grid laid
// over its text, the mouse for it and the way a click runs a command. The
// Wonders, Factions and Harbinger panels share it.

// textButtons holds the buttons laid over a text overlay at the last paint.
type textButtons struct {
	grid *xgrid
	org  image.Point
	// pr makes the buttons pressable.
	pr keyPresser
}

// over returns the painter of the overlay named name: it lays the buttons
// that layer gives for the text in view and draws them on it.
func (tb *textButtons) over(layer func(text image.Rectangle, top int) *xgrid) func(*kit.Painter, image.Rectangle, int) {
	return func(p *kit.Painter, text image.Rectangle, top int) {
		tb.grid, tb.org = layer(text, top), text.Min
		if tb.grid == nil {
			return
		}
		pal := newExpPalette()
		// Only the cells a button covers: the rest is the text underneath.
		for _, k := range tb.grid.keys {
			for y := k.r.Min.Y; y < k.r.Max.Y; y++ {
				for x := k.r.Min.X; x < k.r.Max.X; x++ {
					c := tb.grid.c[y*tb.grid.w+x]
					p.Put(text.Min.X+x, text.Min.Y+y, c.r, pal.style(c))
				}
			}
		}
		tb.pr.paint(p, tb.grid, text.Min, pal)
	}
}

// mouse gives a press or a move to the buttons.
func (tb *textButtons) mouse(m kit.Mouse) bool {
	if tb.grid == nil {
		return false
	}
	return tb.pr.mouse(tb.grid, tb.org, m)
}

// pointer is the hand over a button a click would press.
func (tb *textButtons) pointer(at image.Point) kit.Pointer {
	if tb.grid == nil {
		return kit.PointerArrow
	}
	return tb.pr.pointer(tb.grid, tb.org, at)
}

// attach makes the overlay's buttons work: they paint over its text, take its
// mouse and run their commands through the app.
func (tb *textButtons) attach(om *OverlayManager, name string, layer func(text image.Rectangle, top int) *xgrid) {
	tb.pr.wire(om.app)
	om.SetOver(name, tb.over(layer))
	om.SetButtons(name, tb.mouse, tb.pointer)
}

// textRow is the row the text at byte off starts on, in a view w cells wide.
func textRow(text string, off, w int) int {
	t := &kit.Text{Wrap: kit.WrapChar}
	t.SetText(safeTags(text[:off]) + "X")
	return len(t.Lines(w)) - 1
}

// btnSlot is a place in a text kept for buttons: the byte offset of its first
// reserved row, the column its buttons start at, whether they are one-row
// chips, and the buttons themselves.
type btnSlot struct {
	off, x int
	chips  bool
	btns   []actionSpec
}

// btnRec collects the slots of a text as it is written. A nil one records
// nothing and reserves nothing, so the same writer serves a text that has no
// buttons.
type btnRec struct{ slots []btnSlot }

// flowRows is the rows of buttons that btns take in a line width cells wide,
// breaking before a button that would run past it.
func flowRows(btns []actionSpec, width int) int {
	rows, x := 1, 0
	for _, b := range btns {
		w := actionW(b.label, b.key, signNone)
		if x > 0 && x+w > width {
			rows, x = rows+1, 0
		}
		x += w + 1
	}
	return rows
}

// reserve writes the empty rows btns will be laid on, in a text width cells
// wide, and records them. x is the column the buttons start at.
func (r *btnRec) reserve(sb *strings.Builder, x, width int, btns []actionSpec) {
	if r == nil {
		return
	}
	r.slots = append(r.slots, btnSlot{off: sb.Len(), x: x, btns: btns})
	sb.WriteString(strings.Repeat("\n", flowRows(btns, width-x)*actionRows))
}

// mark records a one-row chip on the line starting at the writer's position,
// x cells in. The line itself is the caller's to write.
func (r *btnRec) mark(sb *strings.Builder, x int, b actionSpec) {
	if r == nil {
		return
	}
	r.slots = append(r.slots, btnSlot{off: sb.Len(), x: x, chips: true, btns: []actionSpec{b}})
}

// lay puts the recorded slots on a grid for the part of text in view: the
// text's rectangle and its first row in view.
func (r *btnRec) lay(text string, rect image.Rectangle, top int) *xgrid {
	if r == nil || len(r.slots) == 0 || rect.Dx() < 1 {
		return nil
	}
	g := newXGrid(rect.Dx(), rect.Dy())
	for _, s := range r.slots {
		row := textRow(text, s.off, rect.Dx()) - top
		x := s.x
		for _, b := range s.btns {
			if s.chips {
				if row >= 0 && row < rect.Dy() && x+runeLen(b.label)+2 <= rect.Dx() {
					g.chip(x, row, b.label, b.kind, b.cmd, b.ask, b.yes, b.off)
				}
				continue
			}
			w := actionW(b.label, b.key, signNone)
			if x > s.x && x+w > rect.Dx() {
				x, row = s.x, row+actionRows
			}
			if row >= 0 && row+actionRows <= rect.Dy() && x+w <= rect.Dx() {
				g.actionButton(x, row, b)
			}
			x += w + 1
		}
	}
	return g
}
