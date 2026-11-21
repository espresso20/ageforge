package ui

import (
	"image"

	"github.com/espresso20/ageforge/rowid"
	"github.com/espresso20/ageforge/window/kit"
)

// panel_frame.go is the frame the text panels are drawn in, the one the
// People, Advance and Expeditions panels have: the whole screen but the
// command bar, a box with the panel's name in its top edge, the text inside
// it, and a line of keys along the foot. Like the other full-screen panels it
// leaves the command bar showing under it (barRows).
//
// The frame and the text are one kit tree on one host: the frame is the
// widget, and the text view's kit box is its child. This is the shape every
// container takes: adopt the children's kit widgets, paint
// them in one tree, and let the mouse find them there.

const (
	// panelPadX is the columns from the box's outer edge to the text, each
	// side: the edge itself and a column of air.
	panelPadX = 2
	// panelChromeRows is the rows the frame takes from the panel: the top
	// edge, the key line and the bottom edge.
	panelChromeRows = 3
	// panelMinRows is the least height at which the key line is drawn.
	panelMinRows = 6
)

// panelFrame is the primitive: a box with a title, the text inside, and the
// keys. It hands the keyboard to the text view, which still takes its keys as
// a Primitive (its input capture is where a panel's own keys live), so
// only the painting and the mouse go through the frame's host.
type panelFrame struct {
	*nativeHost
	kit.Base
	title rowid.Entry
	tv    *nativeText
	// keys is the foot's keys, read on every draw (the Wonders panel's change
	// with its view); hint is what the foot says may be typed, "" when the
	// panel keeps the keyboard for itself.
	keys func() [][2]string
	hint func() string
	// notice, when set and not empty, takes the key line's place for a while
	// (a confirmation, drawn in the good colour).
	notice func() string
	// iw and ih are the text area's size as last drawn (0 before the first).
	iw, ih int
	// buttons makes the foot's keys and the close button pressable, on the
	// frame as last drawn (grid).
	buttons keyPresser
	grid    *xgrid
}

func newPanelFrame(title rowid.Entry, tv *nativeText) *panelFrame {
	f := &panelFrame{title: title, tv: tv}
	f.nativeHost = newNativeHost()
	return f
}

// renderPanelFrame draws the frame at w by h: the box, the name in its top
// edge and the keys along the foot.
func renderPanelFrame(w, h int, title string, keys [][2]string, hint, notice string) *xgrid {
	g := newXGrid(w, h)
	g.shell(title, 0)
	if h >= panelMinRows && notice != "" {
		g.segsFit(2, h-2, w-4, xseg{notice, xGood})
	} else if h >= panelMinRows {
		g.footWords(keys, hint)
	}
	return g
}

// panelTextSize is the size of the text in a frame w by h: inside the edges
// with a column of air each side, and above the key line.
func panelTextSize(w, h int) (tw, th int) {
	return max(1, w-2*panelPadX), max(1, h-panelChromeRows)
}

// textRect is where the text goes in a frame at r.
func panelTextRect(r image.Rectangle) image.Rectangle {
	tw, th := panelTextSize(r.Dx(), r.Dy())
	return image.Rect(r.Min.X+panelPadX, r.Min.Y+1, r.Min.X+panelPadX+tw, r.Min.Y+1+th)
}

// Children: the text view is the frame's child.
func (f *panelFrame) Children() []kit.Widget { return []kit.Widget{f.tv} }

// Layout places the text inside the frame.
func (f *panelFrame) Layout(r image.Rectangle) {
	f.Place(r)
	if f.nativeHost == nil || r.Empty() {
		return // not on the screen yet
	}
	t := panelTextRect(r)
	f.iw, f.ih = t.Dx(), t.Dy()
	place(f.tv, t, screenOffset(f, r))
}

// Paint draws the frame, then the text inside it.
func (f *panelFrame) Paint(p *kit.Painter) {
	r := f.Bounds()
	if r.Empty() {
		return
	}
	var keys [][2]string
	if f.keys != nil {
		keys = f.keys()
	}
	hint := ""
	if f.hint != nil {
		hint = f.hint()
	}
	notice := ""
	if f.notice != nil {
		notice = f.notice()
	}
	g := renderPanelFrame(r.Dx(), r.Dy(), f.title.String(), keys, hint, notice)
	pal := newExpPalette()
	paintExpeditionKit(p, g, r.Min, pal)
	f.buttons.paint(p, g, r.Min, pal)
	f.grid = g
	f.tv.Paint(p)
}

// Mouse: a click on a key along the foot, or on the close button, presses
// that key.
func (f *panelFrame) Mouse(m kit.Mouse) bool {
	return f.buttons.mouse(f.grid, f.Bounds().Min, m)
}

// PointerAt: the hand over a key a click would press.
func (f *panelFrame) PointerAt(at image.Point) kit.Pointer {
	return f.buttons.pointer(f.grid, f.Bounds().Min, at)
}

// Focus gives the keyboard to the text.
func (f *panelFrame) Focus(delegate func(p Primitive)) { delegate(f.tv) }

// HasFocus reports whether the text holds the keyboard.
func (f *panelFrame) HasFocus() bool { return f.tv.HasFocus() }
