package ui

import (
	"image"

	"github.com/espresso20/ageforge/game"
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
	scene *pScene
	state game.GameState
	view  peopleView
	w, h  int // the last size drawn at

	// hist reads the run's recorder.
	hist func() *game.RunHistory
	// buy is the amount the Build buttons use, which Recruit takes as well.
	buy func() buyChoice
}

// wire makes the panel's keys press: its own (a link to a
// tab) are taken by the panel, the rest go to the app.
func (p *peoplePanel) wire(app *appLoop) {
	p.buttons.wire(app)
	press := p.buttons.press
	p.buttons.press = func(k kit.Key) {
		if p.take(k) {
			return
		}
		press(k)
	}
}

// take does what a private key of the panel stands for and reports whether k
// was one.
func (p *peoplePanel) take(k kit.Key) bool {
	if t, ok := goOf(k); ok {
		p.setTab(t)
		return true
	}
	return false
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
	p.view.sub = max(0, p.view.sub)
}

// Layout takes the panel's rectangle and works the panel out in it.
func (p *peoplePanel) Layout(r image.Rectangle) {
	p.Place(r)
	if r.Empty() {
		return
	}
	p.w, p.h = r.Dx(), r.Dy()
	var hist *game.RunHistory
	if p.hist != nil {
		hist = p.hist()
	}
	var buy buyChoice
	if p.buy != nil {
		buy = p.buy()
	}
	p.grid, p.scene = renderPeople(p.state, p.view, buy, hist, p.w, p.h)
}

// Paint puts the panel on the kit's painter in theme colors.
func (p *peoplePanel) Paint(p2 *kit.Painter) {
	r := p.Bounds()
	if p.scene == nil || p.grid == nil || r.Empty() {
		p.paintGrid(p2, p.grid, r)
		return
	}
	pal := newExpPalette()
	paintExpeditionKit(p2, p.grid, r.Min, pal)
	plain := pictureCell == nil
	p.scene.paint(p2, r.Min, plain)
	if p.view.tab != tabOverview && !plain {
		p.scene.lift(p2, p.grid, r.Min, pal)
	}
	p.buttons.paint(p2, p.grid, r.Min, pal)
}

// Mouse: a click on a tab switches to it, as Left and Right do; a click on a
// card, a work area or a building picks it, as Up and Down do; the wheel
// moves the pick a row a notch, as Up and Down do. (Enter opens or closes
// a work area on the Work tab.)
func (p *peoplePanel) Mouse(m kit.Mouse) bool {
	switch {
	case p.buttons.mouse(p.grid, p.Bounds().Min, m):
		return true
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
	if t, ok := p.grid.tabAt(pt); ok {
		p.setTab(t)
		return true
	}
	if n, ok := p.grid.hitAt(hitSub, pt); ok {
		p.view.sub = n
		return true
	}
	if n, ok := p.grid.hitAt(hitRow, pt); ok {
		p.pick(n)
		return true
	}
	return false
}

// PointerAt: the hand over a tab and over a row a click would pick.
func (p *peoplePanel) PointerAt(at image.Point) kit.Pointer {
	if p.grid == nil {
		return kit.PointerArrow
	}
	pt := at.Sub(p.Bounds().Min)
	_, tab := p.grid.tabAt(pt)
	_, row := p.grid.hitAt(hitRow, pt)
	_, sub := p.grid.hitAt(hitSub, pt)
	_, key := p.grid.keyAt(pt)
	return handIf(tab || row || sub || key)
}

// routeKey takes a key if it is one of the panel's and reports whether it
// did. With something typed in the prompt every key belongs to the prompt.
func (p *peoplePanel) routeKey(ev *kit.Key, prompt string) bool {
	switch ev.Code {
	case kit.KeyUp, kit.KeyDown, kit.KeyLeft, kit.KeyRight:
		if prompt != "" {
			return false
		}
	case kit.KeyEnter:
		if prompt != "" || p.view.tab != tabWork {
			return false
		}
		p.view.shut = !p.view.shut
		return true
	case kit.KeyRune:
		return p.take(*ev) || p.stepKey(ev.Rune, prompt)
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
		p.pick(p.view.row - 1)
	case kit.KeyDown:
		p.pick(p.view.row + 1)
	}
}

// pick picks a row of the tab, within it; the picked building of an area goes
// back to the first.
func (p *peoplePanel) pick(row int) {
	row = clamp(row, 0, max(0, peopleRows(p.state, p.view.tab)-1))
	if row != p.view.row {
		p.view.sub = 0
	}
	p.view.row = row
	p.view.shut = false
}

// stepKey takes A (all areas auto), B (buildings), G (growth on or off), and
// + and - on the Work tab with an empty prompt: they put a
// worker in the picked building and take one out, as its chips do.
func (p *peoplePanel) stepKey(r rune, prompt string) bool {
	if prompt != "" || p.buttons.run == nil {
		return false
	}
	if r == 't' || r == 'T' {
		p.buttons.run("research")
		return true
	}
	if p.view.tab != tabWork {
		return false
	}
	switch r {
	case 'a', 'A':
		p.buttons.run(command(verbRoster, "auto"))
		return true
	case 'b', 'B':
		p.buttons.run("buildings")
		return true
	case 'g', 'G':
		to := "off"
		if game.RecruitStatus(p.state) == game.RecruitOff {
			to = "on"
		}
		p.buttons.run(command("workers", "auto-recruit", to))
		return true
	}
	b, ok := pickedBuilding(p.state, p.view)
	if !ok {
		return false
	}
	minus, plus := buildingStep(b.key, b.BuildingState, p.state.Workers.TotalIdle)
	side := minus
	switch r {
	case '+', '=':
		side = plus
	case '-':
	default:
		return false
	}
	if !side.off {
		p.buttons.run(side.cmd)
	}
	return true
}

// setTab switches to a tab, on its first row.
func (p *peoplePanel) setTab(t int) {
	p.view.tab, p.view.row, p.view.sub = t, 0, 0
}
