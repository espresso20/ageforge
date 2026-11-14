package ui

import (
	"image"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/window/kit"
)

// people_panel.go is the People panel's overlay primitive: the whole screen
// but the command bar (people_view.go draws it). Like the Expeditions panel
// it leaves the keyboard with the command bar and takes only the keys that
// print nothing, and only while the bar is empty: Left and Right switch
// tabs, Up and Down pick a row. Esc closes it. Anything typed goes to the
// bar, where `roster <area> <percent>` sets a share.

type peoplePanel struct {
	barPanel
	kit.Base
	// host paints the panel's kit cells and takes the mouse; the keys stay
	// with barPanel.
	grid  *xgrid // what the panel last laid out, and where its rows are
	state game.GameState
	view  peopleView
	w, h  int // the last size drawn at

	// settings reads the glyph tier.
	settings func() mapSettings
}

func newPeoplePanel() *peoplePanel {
	p := &peoplePanel{w: 100, h: 30}
	p.barPanel = newBarPanel(p.routeKey)
	return p
}

// open is the overlay's build: it runs on every Show, on the Overview.
func (p *peoplePanel) open(state game.GameState) Primitive {
	p.state = state
	p.view = peopleView{}
	return overBar(p)
}

// update takes a new snapshot; the picked row stays in range.
func (p *peoplePanel) update(state game.GameState) {
	p.state = state
	p.view.row = clamp(p.view.row, 0, peopleRows(state, p.view.tab)-1)
}

// Layout takes the panel's rectangle and works the panel out in it.
func (p *peoplePanel) Layout(r image.Rectangle) {
	p.Place(r)
	if r.Empty() {
		return
	}
	p.w, p.h = r.Dx(), r.Dy()
	p.grid = renderPeople(p.state, p.view, p.w, p.h)
	if p.settings != nil && p.settings().Tier == mapmodel.TierASCII {
		p.grid.fold()
	}
}

// Paint puts the panel on the kit's painter in theme colors.
func (p *peoplePanel) Paint(p2 *kit.Painter) {
	if r := p.Bounds(); !r.Empty() && p.grid != nil {
		paintExpeditionKit(p2, p.grid, r.Min, newExpPalette())
	}
}

// Mouse: a click on a tab switches to it, as Left and Right do; a click on a
// card, a work area or a building picks it, as Up and Down do; the wheel
// moves the pick a row a notch, as Up and Down do. (Enter does nothing on
// this panel, so a double click is a click.)
func (p *peoplePanel) Mouse(m kit.Mouse) bool {
	switch {
	case m.Kind == kit.MouseWheel:
		key := kit.KeyDown
		if m.Wheel < 0 {
			key = kit.KeyUp
		}
		for i := max(m.Wheel, -m.Wheel); i > 0; i-- {
			p.handleKey(key)
		}
		return m.Wheel != 0
	case m.Kind != kit.MousePress || m.Button != kit.ButtonLeft || p.grid == nil:
		return false
	}
	pt := m.Pt().Sub(p.Bounds().Min)
	if t, ok := p.grid.hitAt(hitTab, pt); ok {
		p.setTab(t)
		return true
	}
	if n, ok := p.grid.hitAt(hitRow, pt); ok {
		p.view.row = clamp(n, 0, peopleRows(p.state, p.view.tab)-1)
		return true
	}
	return false
}

// routeKey takes a key if it is one of the panel's and reports whether it
// did. With something typed in the prompt every key belongs to the prompt.
func (p *peoplePanel) routeKey(ev *kit.Key, prompt string) bool {
	switch ev.Code {
	case kit.KeyUp, kit.KeyDown, kit.KeyLeft, kit.KeyRight:
		if prompt != "" {
			return false
		}
	default:
		return false
	}
	p.handleKey(ev.Code)
	return true
}

func (p *peoplePanel) handleKey(key kit.KeyCode) {
	switch key {
	case kit.KeyLeft:
		p.setTab((p.view.tab + len(peopleTabs) - 1) % len(peopleTabs))
	case kit.KeyRight:
		p.setTab((p.view.tab + 1) % len(peopleTabs))
	case kit.KeyUp:
		p.view.row = max(0, p.view.row-1)
	case kit.KeyDown:
		p.view.row = min(max(0, peopleRows(p.state, p.view.tab)-1), p.view.row+1)
	}
}

// setTab switches to a tab, on its first row.
func (p *peoplePanel) setTab(t int) {
	p.view.tab, p.view.row = t, 0
}
