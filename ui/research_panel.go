package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
)

// research_panel.go is the Research panel: the tech tree as a map over the
// whole screen but the command bar (research_tree.go draws it). Like the
// Map panel it leaves the keyboard with the command bar and takes only the
// keys that print nothing: the arrows move from tech to tech and the view
// follows, Tab and Shift-Tab step through the techs that can start, PgUp
// and PgDn zoom out and in, Home jumps to the current age, Enter opens the
// selected tech's card and, on the card, does what the card says. Esc
// closes the card, then the panel.

// researchPanel is the Research overlay primitive.
type researchPanel struct {
	*tview.Box
	state game.GameState
	view  treeView
	model *treeModel
	w, h  int // the last size drawn at

	// engine starts research and adds to the plan from the card.
	engine *game.GameEngine
	// tier reads the glyph tier (the map glyphs setting).
	tier func() mapmodel.GlyphTier
	// prompt reads the command bar ("" when it is empty).
	prompt func() string
	// toPrompt hands a key the panel does not take to the command bar.
	toPrompt func(ev *tcell.EventKey)
}

func newResearchPanel() *researchPanel {
	return &researchPanel{Box: tview.NewBox(), w: 120, h: 37}
}

// open is the overlay's build: it runs on every Show.
func (p *researchPanel) open(state game.GameState) tview.Primitive {
	p.update(state)
	return tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(p, 0, 1, true).
		AddItem(nil, promptRows, 0, false)
}

// request sets what the next open shows: a zoom ("close", "far" or "" to
// keep it) and a tech whose card to open ("" for none).
func (p *researchPanel) request(zoom, card string) {
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

// update takes a new snapshot and lays the tree out again.
func (p *researchPanel) update(state game.GameState) {
	p.state = state
	p.view.ascii = p.tier != nil && p.tier() == mapmodel.TierASCII
	p.layout()
}

// layout rebuilds the model for the current size and zoom, keeps the
// selection on a tech that is on the map, and scrolls it into view.
func (p *researchPanel) layout() {
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

// Draw paints the panel.
func (p *researchPanel) Draw(scr tcell.Screen) {
	x, y, w, h := p.GetRect()
	if w <= 0 || h <= 0 {
		return
	}
	if w != p.w || h != p.h {
		p.w, p.h = w, h
		p.layout()
	}
	p.view.prompt = p.prompt != nil && p.prompt() != ""
	grid := renderTree(p.state, p.model, p.view, w, h)
	if p.view.ascii {
		foldPlain(grid)
	}
	paintTree(scr, grid, x, y, p.model.lanes)
}

// laneColor is a lane's identity colour.
func laneColor(hue string) tcell.Color {
	var r, g, b int32
	if _, err := fmt.Sscanf(hue, "#%02x%02x%02x", &r, &g, &b); err != nil {
		return theme.Color(theme.RoleAccent)
	}
	return tcell.NewRGBColor(r, g, b)
}

// treeStyles maps the tree's cell styles to the active theme. Lane hues go
// through the contrast rule for art; text roles are the theme's own.
func treeStyles(lanes int, hue func(i int) string) (styles [tsBorder + 1]tcell.Style, lane []tcell.Style) {
	bg := theme.Color(theme.RoleBackground)
	chip := theme.Color(theme.RoleChip)
	on := func(role theme.Role, back tcell.Color, ratio float64) tcell.Style {
		return tcell.StyleDefault.Background(back).Foreground(theme.Legible(theme.Color(role), back, ratio))
	}
	styles[tsText] = on(theme.RoleText, bg, 4.5)
	styles[tsDim] = on(theme.RoleDim, bg, 3)
	styles[tsBright] = on(theme.RoleBright, bg, 4.5).Bold(true)
	styles[tsHi] = on(theme.RoleHighlight, bg, 4.5).Bold(true)
	styles[tsLane] = on(theme.RoleAccent, bg, 3)
	styles[tsGold] = on(theme.RoleHighlight, bg, 3).Bold(true)
	styles[tsGoldDim] = on(theme.RoleWarning, bg, 3)
	styles[tsGood] = on(theme.RolePositive, bg, 4.5)
	sel := theme.Color(theme.RoleSelection)
	styles[tsSel] = tcell.StyleDefault.Background(sel).Foreground(theme.Legible(theme.Color(theme.RoleSelectionText), sel, 4.5)).Bold(true)
	styles[tsChip] = on(theme.RoleText, chip, 4.5)
	styles[tsChipKey] = on(theme.RoleAccent, chip, 3).Bold(true)
	styles[tsChipDim] = on(theme.RoleDim, chip, 3)
	styles[tsBorder] = on(theme.RoleBorder, bg, 3)
	for i := 0; i < lanes; i++ {
		lane = append(lane, tcell.StyleDefault.Background(bg).Foreground(theme.Legible(laneColor(hue(i)), bg, 3)))
	}
	return styles, lane
}

// paintTree puts a rendered grid on the screen in theme colours.
func paintTree(scr tcell.Screen, g *tGrid, x0, y0 int, lanes []config.TechLaneDef) {
	styles, lane := treeStyles(len(lanes), func(i int) string { return lanes[i].Hue })
	for y := 0; y < g.h; y++ {
		for x := 0; x < g.w; x++ {
			c := g.c[y*g.w+x]
			st := styles[c.st]
			if c.st == tsLane && int(c.lane) >= 0 && int(c.lane) < len(lane) {
				st = lane[c.lane]
			}
			scr.SetContent(x0+x, y0+y, c.r, nil, st)
		}
	}
}

// InputHandler takes the panel's keys when it has the focus itself and
// hands the rest to the command bar.
func (p *researchPanel) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return p.WrapInputHandler(func(ev *tcell.EventKey, _ func(tview.Primitive)) {
		prompt := ""
		if p.prompt != nil {
			prompt = p.prompt()
		}
		if !p.routeKey(ev, prompt) && p.toPrompt != nil {
			p.toPrompt(ev)
		}
	})
}

// routeKey takes a key if it is one of the panel's and reports whether it
// did. With something typed in the prompt, Tab and Enter belong to the
// prompt.
func (p *researchPanel) routeKey(ev *tcell.EventKey, prompt string) bool {
	switch ev.Key() {
	case tcell.KeyUp, tcell.KeyDown, tcell.KeyLeft, tcell.KeyRight, tcell.KeyPgUp, tcell.KeyPgDn, tcell.KeyHome:
	case tcell.KeyTab, tcell.KeyBacktab, tcell.KeyEnter:
		if prompt != "" {
			return false
		}
	default:
		return false
	}
	p.handleKey(ev.Key())
	return true
}

// closeCard closes the open card and reports whether there was one: Esc
// closes the card first, then the panel.
func (p *researchPanel) closeCard() bool {
	if !p.view.card {
		return false
	}
	p.view.card, p.view.note = false, ""
	return true
}

func (p *researchPanel) handleKey(key tcell.Key) {
	if p.model == nil {
		p.layout()
	}
	if p.view.card {
		if key == tcell.KeyEnter {
			p.act()
		}
		return
	}
	switch key {
	case tcell.KeyUp:
		p.move(0, -1)
	case tcell.KeyDown:
		p.move(0, 1)
	case tcell.KeyLeft:
		p.move(-1, 0)
	case tcell.KeyRight:
		p.move(1, 0)
	case tcell.KeyTab:
		p.step(1)
	case tcell.KeyBacktab:
		p.step(-1)
	case tcell.KeyPgUp:
		p.view.far = true
	case tcell.KeyPgDn:
		p.view.far = false
	case tcell.KeyHome:
		p.view.sel = ""
	case tcell.KeyEnter:
		p.view.card, p.view.note = p.model.by[p.view.sel] != nil, ""
	}
	p.layout()
}

// move selects the nearest tech in a direction: along the row first for
// left and right, down or up the lane first for up and down.
func (p *researchPanel) move(dx, dy int) {
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
		case dy != 0 && dr*dy > 0:
			s = abs(dr)*100 + abs(dl)
		case dx != 0 && dl*dx > 0:
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
func (p *researchPanel) step(dir int) {
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
func (p *researchPanel) act() {
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
