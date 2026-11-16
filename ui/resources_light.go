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
	// is outside the box. plain says whether to draw in ASCII.
	at    image.Point
	in    bool
	plain func() bool
}

func newResourceView(render func(w, h int) string, plain func() bool) *resourceView {
	v := &resourceView{fitView: newFitView(render), plain: plain}
	v.over = v.light
	return v
}

// Mouse notes where the pointer is, asks for a repaint when the lit line may
// have changed, and leaves the rest to the box.
func (v *resourceView) Mouse(m kit.Mouse) bool {
	in, at := m.Kind != kit.MouseLeave, m.Pt()
	changed := in != v.in || (in && at != v.at)
	v.in, v.at = in, at
	return v.fitView.Mouse(m) || changed
}

// light lights the line of text under the pointer, if it has text.
func (v *resourceView) light(p *kit.Painter, r image.Rectangle, top int) {
	if !v.in || !v.at.In(r) {
		return
	}
	lines := strings.Split(v.raw, "\n")
	i := top + v.at.Y - r.Min.Y
	if i < 0 || i >= len(lines) || strings.TrimSpace(kit.Strip(lines[i])) == "" {
		return
	}
	row := image.Rect(r.Min.X, v.at.Y, r.Max.X, v.at.Y+1)
	lightRow(p, row, false, true, r.Max.X-1, v.plain != nil && v.plain())
}

// Children: the box takes the pointer itself, so that it is told when the
// pointer leaves.
func (v *resourceView) Children() []kit.Widget { return nil }
