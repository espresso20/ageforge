package ui

import (
	"image"

	"github.com/espresso20/ageforge/window/kit"
)

// A nativePages is a stack of named pages, one over
// another, the visible ones drawn back to front and the top one taking the
// keyboard. The calls and what they do to the focus are the UI's. It is a kit
// container, so the pages the kit draws are one tree: a page painted later
// lies over the ones before it, and where it paints nothing they show.
type nativePages struct {
	*nativeHost
	kit.Base
	pages []*nativePage
	// setFocus is how the stack hands the keyboard on, once it has been
	// given it (Focus passes it on).
	setFocus func(p Primitive)
	// inset is the rows a page leaves free at the foot (nil for none): the
	// band a panel's own lesson shows on.
	inset func(name string) int
}

type nativePage struct {
	name    string
	item    Primitive
	w       kit.Widget // item in the tree
	resize  bool
	visible bool
}

func newNativePages() *nativePages {
	p := &nativePages{}
	p.nativeHost = newNativeHost()
	return p
}

// refocus hands the keyboard to the top page again after the stack changed,
// if the stack had it.
func (p *nativePages) refocus(had bool) *nativePages {
	if had {
		p.Focus(p.setFocus)
	}
	return p
}

func (p *nativePages) index(name string) int {
	for i, pg := range p.pages {
		if pg.name == name {
			return i
		}
	}
	return -1
}

// AddPage puts item on top under name, in place of a page of that name.
// With resize it fills the stack's rectangle.
func (p *nativePages) AddPage(name string, item Primitive, resize, visible bool) *nativePages {
	had := p.HasFocus()
	if i := p.index(name); i >= 0 {
		p.pages = append(p.pages[:i], p.pages[i+1:]...)
	}
	p.pages = append(p.pages, &nativePage{name: name, item: item, w: widgetOf(item), resize: resize, visible: visible})
	return p.refocus(had)
}

// RemovePage takes the page away. If it was showing and nothing else is, the
// last page left is shown.
func (p *nativePages) RemovePage(name string) *nativePages {
	had := p.HasFocus()
	if i := p.index(name); i >= 0 {
		shown := p.pages[i].visible
		p.pages = append(p.pages[:i], p.pages[i+1:]...)
		for j, pg := range p.pages {
			if !shown || pg.visible {
				break // it was hidden anyway, or a page still shows
			}
			if j == len(p.pages)-1 {
				pg.visible = true
			}
		}
	}
	return p.refocus(had)
}

func (p *nativePages) HasPage(name string) bool { return p.index(name) >= 0 }

// GetPage is the page of that name, nil when there is none.
func (p *nativePages) GetPage(name string) Primitive {
	if i := p.index(name); i >= 0 {
		return p.pages[i].item
	}
	return nil
}

func (p *nativePages) GetPageCount() int { return len(p.pages) }

// GetPageNames is the pages' names from the top down.
func (p *nativePages) GetPageNames(visibleOnly bool) []string {
	var names []string
	for i := len(p.pages) - 1; i >= 0; i-- {
		if !visibleOnly || p.pages[i].visible {
			names = append(names, p.pages[i].name)
		}
	}
	return names
}

func (p *nativePages) show(name string, on bool) *nativePages {
	if i := p.index(name); i >= 0 {
		p.pages[i].visible = on
	}
	return p.refocus(p.HasFocus())
}

func (p *nativePages) ShowPage(name string) *nativePages { return p.show(name, true) }
func (p *nativePages) HidePage(name string) *nativePages { return p.show(name, false) }

// SwitchToPage shows the page of that name and hides every other.
func (p *nativePages) SwitchToPage(name string) *nativePages {
	for _, pg := range p.pages {
		pg.visible = pg.name == name
	}
	return p.refocus(p.HasFocus())
}

// SendToFront puts the page on top of the others.
func (p *nativePages) SendToFront(name string) *nativePages {
	if i := p.index(name); i >= 0 && i < len(p.pages)-1 {
		pg := p.pages[i]
		p.pages = append(append(p.pages[:i], p.pages[i+1:]...), pg)
	}
	return p.refocus(p.HasFocus())
}

// GetFrontPage is the top page that shows.
func (p *nativePages) GetFrontPage() (name string, item Primitive) {
	for i := len(p.pages) - 1; i >= 0; i-- {
		if p.pages[i].visible {
			return p.pages[i].name, p.pages[i].item
		}
	}
	return "", nil
}

// Focus hands the keyboard to the top page that shows.
func (p *nativePages) Focus(delegate func(p Primitive)) {
	if delegate == nil {
		return // nowhere to hand it on to
	}
	p.setFocus = delegate
	if _, top := p.GetFrontPage(); top != nil {
		delegate(top)
		return
	}
	p.primBase.Focus(delegate)
}

// HasFocus reports whether any page holds the keyboard, shown or not.
func (p *nativePages) HasFocus() bool {
	for _, pg := range p.pages {
		if pg.item.HasFocus() {
			return true
		}
	}
	return p.primBase.HasFocus()
}

// Children are the pages that show, back to front.
func (p *nativePages) Children() []kit.Widget {
	out := make([]kit.Widget, 0, len(p.pages))
	for _, pg := range p.pages {
		if pg.visible {
			out = append(out, pg.w)
		}
	}
	return out
}

func (p *nativePages) Layout(r image.Rectangle) {
	p.Place(r)
	if p.nativeHost == nil {
		return // newNativeHost lays the tree out before the host is set
	}
	off := screenOffset(p, r)
	for _, pg := range p.pages {
		switch {
		case !pg.visible:
		case pg.resize:
			pr := r
			if p.inset != nil {
				pr.Max.Y = max(pr.Max.Y-p.inset(pg.name), pr.Min.Y)
			}
			place(pg.w, pr, off)
		default:
			// A page that keeps its own size is where it put itself.
			x, y, w, h := pg.item.GetRect()
			pg.w.Layout(image.Rect(x, y, x+w, y+h).Sub(off))
		}
	}
}

func (p *nativePages) Paint(kp *kit.Painter) {
	for _, pg := range p.pages {
		if pg.visible {
			pg.w.Paint(kp)
		}
	}
}

// own: the stack clears its whole rectangle to the page's ground, under
// whatever its pages paint.
func (p *nativePages) own() []image.Rectangle { return []image.Rectangle{p.Bounds()} }
