package ui

import (
	"image"
	"image/color"

	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/window/kit"
)

// tabs.go is the tab: a small plate on a row of them. The open tab is bright
// and joins the body below it; the others sit back and dim, and one under the
// pointer comes forward. Panels record their tabs on the keySet as they draw,
// so one painter and one hit test serve them all. Where pictures are not
// shown the cells keep the text the panel drew (a chip for the open tab).

const (
	// tabBack is how far a tab at rest sits back, in art pixels; a tab under
	// the pointer sits one back.
	tabBack, tabLift = 3, 1
)

// tabPlate is a tab: the cells it takes, its label, and whether it is open.
type tabPlate struct {
	r     image.Rectangle
	label string
	// n is the panel's own number for the tab.
	n    int
	open bool
}

// tab records a tab on the set.
func (s *keySet) tab(r image.Rectangle, label string, n int, open bool) {
	s.tabs = append(s.tabs, tabPlate{r, label, n, open})
}

// tabAt is the number of the tab the cell pt is in.
func (s *keySet) tabAt(pt image.Point) (int, bool) {
	if s != nil {
		for _, t := range s.tabs {
			if pt.In(t.r) {
				return t.n, true
			}
		}
	}
	return 0, false
}

// tabOver is the tab under the pointer, -1 for none.
func (k *keyPresser) tabOver(s *keySet) int {
	if s == nil || k.press == nil || !k.in {
		return -1
	}
	for i, t := range s.tabs {
		if k.over.In(t.r) {
			return i
		}
	}
	return -1
}

// tabKey is what a tab plate was drawn for.
type tabKey struct {
	open, hover  bool
	cols, cw, ch int
	edge, face   color.RGBA
}

// tab is the picture of a tab cols cells wide.
func (s *plateSet) tab(pk tabKey) *image.RGBA {
	s.fit(pk.cw, pk.ch)
	img, ok := s.tabs[pk]
	if !ok {
		if s.tabs == nil {
			s.tabs = map[tabKey]*image.RGBA{}
		}
		img = drawTab(pk)
		s.tabs[pk] = img
		s.bytes += len(img.Pix)
	}
	return img
}

// drawTab draws a tab over its cells: an outline with rounded top corners and
// no bottom edge when it is open.
func drawTab(pk tabKey) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, pk.cols*pk.cw, pk.ch))
	k := max(1, pk.ch/21)
	a := art{img, k}
	W, H := (img.Rect.Dx()+k-1)/k, (img.Rect.Dy()+k-1)/k
	top := tabBack
	switch {
	case pk.open:
		top = 0
	case pk.hover:
		top = tabLift
	}
	x0, x1 := 1, W-1
	a.fill(image.Rect(x0, top, x1, H), 1, pk.edge)
	bottom := H
	if !pk.open {
		bottom = H - 1
	}
	a.fill(image.Rect(x0+1, top+1, x1-1, bottom), 0, pk.face)
	return img
}

// paintTabs paints the tabs of s over the grid already painted at org.
func (k *keyPresser) paintTabs(p *kit.Painter, s *keySet, org image.Point, pal *expPalette) {
	if pictureCell == nil || s.plain || len(s.tabs) == 0 {
		return
	}
	cw, ch := pictureCell()
	if cw <= 0 || ch <= 0 {
		return
	}
	ground, text, dim, accent := pal.tone(xText, true), pal.tone(xText, false), pal.tone(xDim, false), pal.tone(xKeycap, true)
	over := k.tabOver(s)
	for i, t := range s.tabs {
		pk := tabKey{open: t.open, hover: !t.open && i == over, cols: t.r.Dx(), cw: cw, ch: ch}
		ink := dim
		switch {
		case t.open:
			pk.edge, pk.face, ink = accent, toward(ground, accent, .16), text
		case pk.hover:
			pk.edge, pk.face, ink = toward(accent, ground, .4), toward(ground, accent, .08), text
		default:
			pk.edge, pk.face = toward(dim, ground, .45), ground
		}
		r := t.r.Add(org)
		p.Fill(r, theme.RGB(ground.R, ground.G, ground.B))
		p.Picture(r, k.plates.tab(pk))
		label := []rune(t.label)
		st := kit.Style{}.Foreground(theme.RGB(ink.R, ink.G, ink.B)).Bold(t.open)
		p.Label(r.Min.X+max((r.Dx()-len(label))/2, 0), r.Min.Y, kit.Escape(t.label), st)
	}
}
