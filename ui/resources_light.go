package ui

import (
	"image"
	"strings"

	"github.com/espresso20/ageforge/window/kit"
)

// resourceView is the Resources box: a fitView whose line under the pointer
// is lit. The light has no click action.
type resourceView struct {
	*fitView
	// at is the cell the pointer is on, in screen cells; in is false when it
	// is outside the box.
	at image.Point
	in bool
	// tipFor says what the tooltip on text line i is; ok is false for a line
	// with none. The box and the first line it showed at the last draw are
	// kept so a pointer move can be traced to a line.
	tip       *kit.Tip
	tipFor    func(line int) (id any, text kit.TipText, ok bool)
	box       image.Rectangle
	firstLine int
}

func newResourceView(render func(w, h int) string) *resourceView {
	v := &resourceView{fitView: newFitView(render)}
	v.over = v.light
	return v
}

// Mouse notes where the pointer is, asks for a repaint when the lit line may
// have changed, and leaves the rest to the box.
func (v *resourceView) Mouse(m kit.Mouse) bool {
	in, at := m.Kind != kit.MouseLeave, m.Pt()
	changed := in != v.in || (in && at != v.at)
	v.in, v.at = in, at
	if m.Kind == kit.MouseMove && in && v.tip != nil && v.tipFor != nil && at.In(v.box) {
		if id, text, ok := v.tipFor(v.firstLine + at.Y - v.box.Min.Y); ok {
			v.tip.Aim(id, image.Rect(v.box.Min.X, at.Y, v.box.Max.X, at.Y+1), text, false)
		}
	}
	return v.fitView.Mouse(m) || changed
}

// light lights the line of text under the pointer, if it has text.
func (v *resourceView) light(p *kit.Painter, r image.Rectangle, top int) {
	v.box, v.firstLine = r, top
	if !v.in || !v.at.In(r) {
		return
	}
	lines := strings.Split(v.raw, "\n")
	i := top + v.at.Y - r.Min.Y
	if i < 0 || i >= len(lines) || strings.TrimSpace(kit.Strip(lines[i])) == "" {
		return
	}
	row := image.Rect(r.Min.X, v.at.Y, r.Max.X, v.at.Y+1)
	lightRow(p, row, false, true, r.Max.X-1)
}

// Children: the box takes the pointer itself, so that it is told when the
// pointer leaves.
func (v *resourceView) Children() []kit.Widget { return nil }
