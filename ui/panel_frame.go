package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/vibe"
)

// panel_frame.go is the frame the text panels are drawn in, the one the
// People, Advance and Expeditions panels have: the whole screen but the
// command bar, a box with the panel's name in its top edge, the text inside
// it, and a line of keys along the foot. The text is a tview TextView laid
// inside the box, so it scrolls, wraps and takes colour tags as it always did;
// only the frame around it is new. Like the other full-screen panels it
// leaves the command bar showing under it (promptRows).

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

// panelFrame is the primitive: a box with a title, a TextView inside, and the
// keys. It hands the keyboard to the TextView.
type panelFrame struct {
	*tview.Box
	title vibe.Entry
	tv    *tview.TextView
	// keys is the foot's keys, read on every draw (the Wonders panel's change
	// with its view); hint is what the foot says may be typed, "" when the
	// panel keeps the keyboard for itself.
	keys func() [][2]string
	hint func() string
	// plain reports the ASCII glyph tier, when the frame folds to it.
	plain func() bool
	// iw and ih are the text area's size as last drawn (0 before the first).
	iw, ih int
}

func newPanelFrame(title vibe.Entry, tv *tview.TextView) *panelFrame {
	return &panelFrame{Box: tview.NewBox(), title: title, tv: tv}
}

// renderPanelFrame draws the frame at w by h: the box, the name in its top
// edge and the keys along the foot.
func renderPanelFrame(w, h int, title string, keys [][2]string, hint string) *xgrid {
	g := newXGrid(w, h)
	g.box(0, 0, w, h, xBorder, title, xTitle, false)
	if h >= panelMinRows {
		expKeyBar(g, expLayout{w: w, h: h, footY: h - 2}, keys, hint)
	}
	return g
}

// panelTextSize is the size of the text in a frame w by h: inside the edges
// with a column of air each side, and above the key line.
func panelTextSize(w, h int) (tw, th int) {
	return max(1, w-2*panelPadX), max(1, h-panelChromeRows)
}

// Draw paints the frame, then the text inside it.
func (f *panelFrame) Draw(scr tcell.Screen) {
	x, y, w, h := f.GetRect()
	if w <= 0 || h <= 0 {
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
	g := renderPanelFrame(w, h, f.title.String(), keys, hint)
	if f.plain != nil && f.plain() {
		g.fold()
	}
	paintExpedition(scr, g, x, y, newExpPalette())
	tw, th := panelTextSize(w, h)
	f.iw, f.ih = tw, th
	f.tv.SetRect(x+panelPadX, y+1, tw, th)
	f.tv.Draw(scr)
}

// Focus gives the keyboard to the text.
func (f *panelFrame) Focus(delegate func(p tview.Primitive)) { delegate(f.tv) }

// HasFocus reports whether the text holds the keyboard.
func (f *panelFrame) HasFocus() bool { return f.tv.HasFocus() }

// InputHandler is the text's.
func (f *panelFrame) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return f.tv.InputHandler()
}
