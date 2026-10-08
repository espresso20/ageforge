package ui

import (
	"strings"
	"sync/atomic"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// badge_panel.go is the badge case's panel: the case over the whole screen
// but the command bar (badge_case.go draws it). Like the Map and Research
// panels it leaves the keyboard with the command bar and takes only the
// keys that print nothing: the arrows move from badge to badge, Tab and
// Shift-Tab step through the families, PgUp and PgDn move a page, Home and
// End go to the first and the last badge, Enter opens the selected badge's
// detail. Esc closes the detail, then the panel.
//
// What moves in the case (a platinum badge's glints, a legendary badge's
// ring, the three hand-drawn badges) is drawn from a frame number the
// panel reads off the clock: nothing is stored, and the game's motion
// setting turns it off.

// badgeKeys are the badge case's keys as the Help panel lists them: the
// key, then what it does.
var badgeKeys = [][2]string{
	{"Arrows", "Move from badge to badge; the grid follows"},
	{"Tab", "The next family's tab (Shift-Tab: the one before)"},
	{"PgUp/PgDn", "A page up and down"},
	{"Home/End", "The first badge and the last"},
	{"Enter", "Open the selected badge's detail, and close it (with something typed, Tab and Enter act on the prompt instead)"},
	{"Esc", "Close the detail, then the panel"},
}

// badgeRequest is how a command asks the case to open: on a tab, or on a
// badge with its detail open. The zero value leaves the case as it was.
type badgeRequest struct {
	tab    string
	setTab bool
	sel    string
	card   bool
}

// badgePanel is the badge case's overlay primitive.
type badgePanel struct {
	*tview.Box
	views   []game.BadgeView
	sum     game.BadgeSummary
	account bool
	view    caseView
	model   *caseModel
	w, h    int // the last size drawn at

	// settings reads the glyph tier and the motion setting.
	settings func() mapSettings
	motion   bool
	// prompt reads the command bar ("" when it is empty).
	prompt func() string
	// toPrompt hands a key the panel does not take to the command bar.
	toPrompt func(ev *tcell.EventKey)

	// now and start are the animation clock.
	now   func() time.Time
	start time.Time
	// moving is set by a draw that drew something that moves, so the
	// dashboard's refresh loop redraws the case at the animation rate only
	// while there is a reason to. Read off the UI goroutine.
	moving atomic.Bool
}

func newBadgePanel() *badgePanel {
	return &badgePanel{Box: tview.NewBox(), w: 120, h: 37, now: time.Now, start: time.Now(), motion: true}
}

// open is the overlay's build: it runs on every Show.
func (p *badgePanel) open(state game.GameState) tview.Primitive {
	p.update(state)
	return tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(p, 0, 1, true).
		AddItem(nil, promptRows, 0, false)
}

// request sets what the next open shows.
func (p *badgePanel) request(r badgeRequest) {
	p.view.card, p.view.note = false, ""
	if r.setTab {
		p.view.tab, p.view.sel, p.view.vy = r.tab, "", 0
	}
	if r.sel != "" {
		// A badge by name: on All, where every badge is.
		p.view.tab, p.view.sel, p.view.card = caseTabAll, r.sel, r.card
	}
}

// update takes a new snapshot and lays the case out again.
func (p *badgePanel) update(state game.GameState) {
	p.account = state.AccountStats != nil
	p.views, p.sum = nil, game.BadgeSummary{}
	if p.account {
		p.views, p.sum = state.AccountStats.Badges, state.AccountStats.BadgeSummary
	}
	if p.settings != nil {
		s := p.settings()
		p.view.tier, p.motion = s.Tier, s.Motion
	}
	p.layout()
}

// layout rebuilds the model for the current size and tab, keeps the
// selection on something the tab lists, and scrolls it into view.
func (p *badgePanel) layout() {
	p.model = buildCase(p.views, p.sum, p.account, p.view.tab, caseLayoutFor(p.w, p.h, p.view.tab))
	known := false
	for _, t := range p.model.tabs {
		known = known || t.key == p.view.tab
	}
	if !known {
		p.view.tab = caseTabAll
	}
	if p.model.selected(p.view) == nil {
		p.view.sel, p.view.card = p.model.home(), false
	}
	p.model.follow(&p.view)
}

// frame is the animation frame: the clock's, or atRest when the motion
// setting is off.
func (p *badgePanel) frame() int {
	if !p.motion {
		return atRest
	}
	return int(p.now().Sub(p.start) / mapAnimStep)
}

// moves reports whether the case as it stands draws anything that moves:
// the selected badge's art, or a legendary badge on the page.
func (p *badgePanel) moves() bool {
	if !p.motion || p.model == nil || !p.account {
		return false
	}
	if it := p.model.selected(p.view); it != nil && medalOf(it.v, p.view.tier).animated() {
		return true
	}
	for _, ln := range p.model.lines {
		if ln.kind != lineBadges || !p.model.visible(ln, p.view.vy) {
			continue
		}
		for _, i := range ln.items {
			if md := medalOf(p.model.items[i].v, p.view.tier); md.state == medalEarned && md.tier == config.BadgeLegendary {
				return true
			}
		}
	}
	return false
}

// Draw paints the panel.
func (p *badgePanel) Draw(scr tcell.Screen) {
	x, y, w, h := p.GetRect()
	if w <= 0 || h <= 0 {
		return
	}
	if w != p.w || h != p.h || p.model == nil {
		p.w, p.h = w, h
		p.layout()
	}
	p.view.prompt = p.prompt != nil && p.prompt() != ""
	grid := renderCase(p.model, p.view, p.frame())
	if p.view.plain() {
		foldBadges(grid)
	}
	paintGrid(scr, grid, x, y, newGridPalette(0, nil))
	p.moving.Store(p.moves())
}

// InputHandler takes the panel's keys when it has the focus itself and
// hands the rest to the command bar.
func (p *badgePanel) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
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
func (p *badgePanel) routeKey(ev *tcell.EventKey, prompt string) bool {
	switch ev.Key() {
	case tcell.KeyUp, tcell.KeyDown, tcell.KeyLeft, tcell.KeyRight, tcell.KeyPgUp, tcell.KeyPgDn, tcell.KeyHome, tcell.KeyEnd:
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

// closeCard closes the open detail and reports whether there was one: Esc
// closes the detail first, then the panel.
func (p *badgePanel) closeCard() bool {
	if !p.view.card {
		return false
	}
	p.view.card = false
	return true
}

func (p *badgePanel) handleKey(key tcell.Key) {
	if p.model == nil {
		p.layout()
	}
	p.view.note = ""
	switch key {
	case tcell.KeyUp:
		p.moveLine(-1)
	case tcell.KeyDown:
		p.moveLine(1)
	case tcell.KeyLeft:
		p.step(-1)
	case tcell.KeyRight:
		p.step(1)
	case tcell.KeyPgUp:
		p.page(-1)
	case tcell.KeyPgDn:
		p.page(1)
	case tcell.KeyHome:
		if len(p.model.items) > 0 {
			p.view.sel = p.model.items[0].id
		}
	case tcell.KeyEnd:
		if n := len(p.model.items); n > 0 {
			p.view.sel = p.model.items[n-1].id
		}
	case tcell.KeyTab:
		p.tab(1)
	case tcell.KeyBacktab:
		p.tab(-1)
	case tcell.KeyEnter:
		p.view.card = !p.view.card && p.model.selected(p.view) != nil
	}
	p.layout()
}

// step selects the badge before or after the selected one, in order.
func (p *badgePanel) step(dir int) {
	i, ok := p.model.by[p.view.sel]
	if !ok {
		return
	}
	if j := i + dir; j >= 0 && j < len(p.model.items) {
		p.view.sel = p.model.items[j].id
	}
}

// moveLine selects the badge on the line above or below, in the nearest
// column.
func (p *badgePanel) moveLine(dir int) {
	p.jump(dir, 1)
}

// page moves the selection up or down by as many lines as the grid shows.
func (p *badgePanel) page(dir int) {
	lines := 1
	if p.model.lay.list {
		lines = max(p.model.lay.gh-1, 1)
	} else {
		lines = max(p.model.lay.gh/casePitchY, 1)
	}
	p.jump(dir, lines)
}

// jump moves the selection by count lines that hold badges, stopping at
// the first and the last, and keeps its column.
func (p *badgePanel) jump(dir, count int) {
	it := p.model.selected(p.view)
	if it == nil {
		return
	}
	line, col := it.line, it.col
	for l := line + dir; l >= 0 && l < len(p.model.lines) && count > 0; l += dir {
		if len(p.model.lines[l].items) > 0 {
			line = l
			count--
		}
	}
	items := p.model.lines[line].items
	if len(items) == 0 {
		return
	}
	p.view.sel = p.model.items[items[min(col, len(items)-1)]].id
}

// tab opens the tab after or before the open one, and wraps round.
func (p *badgePanel) tab(dir int) {
	tabs := p.model.tabs
	at := 0
	for i, t := range tabs {
		if t.key == p.view.tab {
			at = i
		}
	}
	p.view.tab = tabs[((at+dir)%len(tabs)+len(tabs))%len(tabs)].key
	p.view.sel, p.view.vy, p.view.card = "", 0, false
}

// findBadgeTab reads a word as a tab: a family's key or its heading, in
// any case ("ages", "age", "Lineages").
func findBadgeTab(views []game.BadgeView, word string) (string, bool) {
	word = strings.ToLower(strings.TrimSpace(word))
	for _, t := range caseTabs(views) {
		if t.key == caseTabAll || t.key == caseTabNext {
			continue
		}
		if word == strings.ToLower(t.key) || word == strings.ToLower(t.title) {
			return t.key, true
		}
	}
	return "", false
}

// findBadge reads a word as a badge in sight: its name, the start of it,
// or a part of it, in that order. A withheld badge has no name to find.
func findBadge(views []game.BadgeView, word string) (string, bool) {
	word = strings.ToLower(strings.TrimSpace(word))
	if word == "" {
		return "", false
	}
	best, rank := "", 3
	for _, v := range views {
		if v.Hidden {
			continue
		}
		name := strings.ToLower(v.Name)
		r := 3
		switch {
		case name == word:
			r = 0
		case strings.HasPrefix(name, word):
			r = 1
		case strings.Contains(name, word):
			r = 2
		}
		if r < rank {
			best, rank = v.Key, r
		}
	}
	return best, best != ""
}
