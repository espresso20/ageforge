package ui

import (
	"image"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/rowid"
	"github.com/espresso20/ageforge/window/kit"
)

// tree_panel.go is the tree panel: the tech tree as a map over the
// whole screen but the command bar (research_tree.go draws it). Like the
// Map panel it leaves the keyboard with the command bar and takes only the
// keys that print nothing: the arrows move from tech to tech and the view
// follows, Tab and Shift-Tab step through the techs that can start, PgUp
// and PgDn zoom out and in, Home jumps to the current age, Enter opens the
// selected tech's card and, on the card, does what the card says. Esc
// closes the card, then the panel.

// treeKeys are the tree panel's keys as the Help panel lists them:
// the key, then what it does.
var treeKeys = []keyHelp{
	{"Arrows", rowid.E("ui.research_panel.treeKeys.02", "Move to the nearest tech that way; the view follows")},
	{"Tab", rowid.E("ui.research_panel.treeKeys.04", "Next tech you can start (Shift-Tab: the one before)")},
	{"PgUp/PgDn", rowid.E("ui.research_panel.treeKeys.05", "Zoom out to a tech a line, and back in")},
	{"Home", rowid.E("ui.research_panel.treeKeys.07", "Back to your current age")},
	{"Enter", rowid.E("ui.research_panel.treeKeys.09", "Open the selected tech's card; on the card, do what its last line says (with something typed, Tab and Enter act on the prompt instead)")},
	{"Esc", rowid.E("ui.research_panel.treeKeys.11", "Close the card, then the panel")},
}

// treePanel is the Research overlay primitive.
type treePanel struct {
	barPanel
	kit.Base
	// host paints the panel's kit cells and takes the mouse; the keys stay
	// with barPanel.
	state game.GameState
	view  treeView
	model *treeModel
	w, h  int // the last size drawn at

	// engine starts research and adds to the plan from the card.
	engine *game.GameEngine
	// tier reads the glyph tier (the map glyphs setting).
	tier func() mapmodel.GlyphTier
	// back returns to the Research board: Esc does it once the card is closed.
	back func()
}

func newTreePanel() *treePanel {
	p := &treePanel{w: 120, h: 37}
	p.barPanel = newBarPanel(p.routeKey)
	return p
}

// open is the overlay's build: it runs on every Show.
func (p *treePanel) open(state game.GameState) Primitive {
	p.update(state)
	return overBar(p)
}

// request sets what the next open shows: a zoom ("close", "far" or "" to
// keep it) and a tech whose card to open ("" for none).
func (p *treePanel) request(zoom, card string) {
	switch zoom {
	case "close":
		p.view.far = false
	case "far":
		p.view.far = true
	}
	p.view.card, p.view.note = false, ""
	if card != "" {
		p.view.sel, p.view.card = card, true
	}
}

// show sets the tech the next open is on, with no card.
func (p *treePanel) show(key string) {
	p.view.card, p.view.note = false, ""
	if key != "" {
		p.view.sel = key
	}
}

// closeBack closes the card, and with none open goes back to the board:
// Esc does this before it would close the panel.
func (p *treePanel) closeBack() bool {
	if p.closeCard() {
		return true
	}
	if p.back == nil {
		return false
	}
	p.back()
	return true
}

// update takes a new snapshot and lays the tree out again.
func (p *treePanel) update(state game.GameState) {
	p.state = state
	p.view.ascii = p.tier != nil && p.tier() == mapmodel.TierASCII
	p.layout()
}

// layout rebuilds the model for the current size and zoom, keeps the
// selection on a tech that is on the map, and scrolls it into view.
func (p *treePanel) layout() {
	geom := treeGeomFor(p.w, p.view.far)
	p.model = buildTree(p.state, geom, p.view.ascii, p.view.sel)
	if p.model.by[p.view.sel] == nil {
		p.view.sel, p.view.card = p.model.home(), false
		p.model = buildTree(p.state, geom, p.view.ascii, p.view.sel)
	}
	p.model.follow(&p.view, p.w, p.h)
}

// home is the tech the panel opens on: the one being researched, else the
// first that can start in the current age, else the current age's first.
func (m *treeModel) home() string {
	first, ready := "", ""
	for _, n := range m.nodes {
		b := m.bandOf(n)
		switch {
		case n.mark == markRunning:
			return n.key
		case !b.here:
		case ready == "" && n.mark == markReady:
			ready = n.key
		case first == "":
			first = n.key
		}
	}
	switch {
	case ready != "":
		return ready
	case first != "":
		return first
	case len(m.nodes) > 0:
		return m.nodes[0].key
	}
	return ""
}

// Layout takes the panel's rectangle and lays the tree out again when its
// size has changed.
func (p *treePanel) Layout(r image.Rectangle) {
	p.Place(r)
	if w, h := r.Dx(), r.Dy(); w > 0 && h > 0 && (w != p.w || h != p.h) {
		p.w, p.h = w, h
		if p.model != nil {
			p.layout()
		}
	}
}

// Paint puts the tree on the kit's painter in theme colors (the shared grid
// painter, cell_grid.go). Lane hues go through the contrast rule for art;
// text roles are the theme's own.
func (p *treePanel) Paint(p2 *kit.Painter) {
	r := p.Bounds()
	if r.Empty() || p.model == nil {
		return
	}
	p.view.prompt = p.prompt != nil && p.prompt() != ""
	grid := renderTree(p.state, p.model, p.view, r.Dx(), r.Dy())
	if p.view.ascii {
		foldPlain(grid)
	}
	lanes := p.model.lanes
	paintGridKit(p2, grid, r.Min, newGridPalette(len(lanes), func(i int) string { return lanes[i].Hue }))
}

// Mouse: a click picks the tech under the pointer and opens its card, as
// Enter on it does; a click outside an open card closes it, as Esc does;
// the wheel moves the selection a tech a notch, as the arrows do.
func (p *treePanel) Mouse(m kit.Mouse) bool {
	if p.model == nil {
		return false
	}
	switch {
	case m.Kind == kit.MouseWheel && !p.view.card:
		key := kit.KeyDown
		if m.Wheel < 0 {
			key = kit.KeyUp
		}
		for i := max(m.Wheel, -m.Wheel); i > 0; i-- {
			p.handleKey(key)
		}
		return m.Wheel != 0
	case m.Kind != kit.MousePress || m.Button != kit.ButtonLeft:
		return false
	case p.view.card:
		return !m.Pt().Sub(p.Bounds().Min).In(p.model.cardAt) && p.closeCard()
	}
	key := p.model.techAt(m.Pt().Sub(p.Bounds().Min), p.view, p.w, p.h)
	if key == "" {
		return false
	}
	// A click picks the tech and opens its card, as Enter on it does: a
	// click that only moved the pick looked like nothing had happened.
	p.view.sel = key
	p.layout()
	p.handleKey(kit.KeyEnter)
	return true
}

// PointerAt: the hand over a tech.
func (p *treePanel) PointerAt(at image.Point) kit.Pointer {
	return handIf(p.model != nil && !p.view.card && p.model.techAt(at.Sub(p.Bounds().Min), p.view, p.w, p.h) != "")
}

// routeKey takes a key if it is one of the panel's and reports whether it
// did. With something typed in the prompt, Tab and Enter belong to the
// prompt.
func (p *treePanel) routeKey(ev *kit.Key, prompt string) bool {
	switch ev.Code {
	case kit.KeyUp, kit.KeyDown, kit.KeyLeft, kit.KeyRight, kit.KeyPgUp, kit.KeyPgDn, kit.KeyHome:
	case kit.KeyTab, kit.KeyBacktab, kit.KeyEnter:
		if prompt != "" {
			return false
		}
	default:
		return false
	}
	p.handleKey(ev.Code)
	return true
}

// closeCard closes the open card and reports whether there was one: Esc
// closes the card first, then the panel.
func (p *treePanel) closeCard() bool {
	if !p.view.card {
		return false
	}
	p.view.card, p.view.note = false, ""
	return true
}

func (p *treePanel) handleKey(key kit.KeyCode) {
	if p.model == nil {
		p.layout()
	}
	if p.view.card {
		if key == kit.KeyEnter {
			p.act()
		}
		return
	}
	switch key {
	case kit.KeyUp:
		p.move(0, -1)
	case kit.KeyDown:
		p.move(0, 1)
	case kit.KeyLeft:
		p.move(-1, 0)
	case kit.KeyRight:
		p.move(1, 0)
	case kit.KeyTab:
		p.step(1)
	case kit.KeyBacktab:
		p.step(-1)
	case kit.KeyPgUp:
		p.view.far = true
	case kit.KeyPgDn:
		p.view.far = false
	case kit.KeyHome:
		p.view.sel = ""
	case kit.KeyEnter:
		p.view.card, p.view.note = p.model.by[p.view.sel] != nil, ""
	}
	p.layout()
}

// move selects the nearest tech in a direction: along the row first for
// left and right, down or up the lane first for up and down.
func (p *treePanel) move(dx, dy int) {
	cur := p.model.by[p.view.sel]
	if cur == nil {
		return
	}
	abs := func(v int) int {
		if v < 0 {
			return -v
		}
		return v
	}
	best, score := "", 1<<30
	for _, n := range p.model.nodes {
		dr, dl := n.row-cur.row, n.lane-cur.lane
		s := 0
		switch {
		case dy != 0 && dr*dy > 0, dx != 0 && dl*dx > 0:
			s = abs(dr)*100 + abs(dl)
		default:
			continue
		}
		if s < score {
			best, score = n.key, s
		}
	}
	if best != "" {
		p.view.sel = best
	}
}

// step selects the next (or the previous) tech that can start, in reading
// order, and wraps round.
func (p *treePanel) step(dir int) {
	var ready []string
	at := -1
	for _, n := range p.model.nodes {
		if n.mark == markReady || n.key == p.view.sel {
			if n.key == p.view.sel {
				at = len(ready)
			}
			ready = append(ready, n.key)
		}
	}
	if len(ready) == 0 {
		return
	}
	for i := 1; i <= len(ready); i++ {
		k := ready[((at+dir*i)%len(ready)+len(ready))%len(ready)]
		if p.model.by[k].mark == markReady {
			p.view.sel = k
			return
		}
	}
}

// act does what the open card says: it starts the tech, or adds it to the
// build plan. Research spends knowledge for good, so this is the second
// Enter, never the first. What the game answers goes on the card.
func (p *treePanel) act() {
	n := p.model.by[p.view.sel]
	if n == nil || p.engine == nil {
		return
	}
	var err error
	switch cardActionFor(p.state, n) {
	case actStart:
		err = p.engine.StartResearch(n.key)
	case actPlan:
		err = p.engine.PlanAddResearch(n.key)
	default:
		return
	}
	p.view.note = ""
	if err != nil {
		p.view.note = err.Error()
	}
	p.state = p.engine.GetState()
	p.layout()
}
