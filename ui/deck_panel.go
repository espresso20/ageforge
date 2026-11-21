package ui

import (
	"image"
	"sync/atomic"
	"time"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/rowid"
	"github.com/espresso20/ageforge/window/kit"
)

// deck_panel.go is the overlay primitive of a deck (deck.go): the whole
// screen but the command bar. Like the People panel it leaves the keyboard
// with the command bar and takes only the keys that print nothing, and only
// while the bar is empty: the arrows move the pick from card to card, Enter
// runs the picked card's command, PgUp and PgDn turn a page. Esc closes it.
// A click picks a card; a click on a card's button (it carries the command, so
// the buttons' own path runs it), or a double click on the card, runs its
// command. Every command goes through the command bar's own
// path, so it is echoed and answered like one that was typed.

type deckPanel struct {
	barPanel
	kit.Base
	// page works the panel's content out from a snapshot.
	page func(game.GameState) deckPage
	// draw draws the page; nil draws it as a deck of cards. A panel with a
	// scene of its own (the Trade panel's map of roads) sets it.
	draw func(pg *deckPage, st game.GameState, sel, scroll, w, h int, said ...xseg) (*xgrid, deckLayout)
	// settings reads the motion setting.
	settings func() mapSettings
	// held says the game is held; the lanes stop with it. now is the frame
	// clock, clock the seconds run, moving whether a draw has something moving.
	held   func() bool
	now    func() time.Time
	last   time.Time
	clock  float64
	moving atomic.Bool

	state  game.GameState
	pg     deckPage
	grid   *xgrid
	lay    deckLayout
	sel    int
	scroll int
	// said is the last command run over the panel and what came of it,
	// shown over the keys: the log is behind the panel.
	said []xseg
}

func newDeckPanel(page func(game.GameState) deckPage) *deckPanel {
	p := &deckPanel{page: page, now: time.Now}
	p.barPanel = newBarPanel(p.routeKey)
	return p
}

// registerDeck makes a panel of cards from page and registers it as the
// overlay name, with its keyboard left to the command bar and its commands
// run as the command bar runs them (the buttons' own path, key_bar.go).
func (d *Dashboard) registerDeck(name string, title rowid.Entry, page func(game.GameState) deckPage) *deckPanel {
	p := newDeckPanel(page)
	p.settings = d.mapSettings
	p.held = d.engine.Held
	p.prompt = func() string { return d.inputField.GetText() }
	p.toPrompt = d.handToPrompt
	p.buttons.wire(d.app)
	d.decks[name] = p
	d.overlayMgr.RegisterWidget(name, title, p.open, p.update)
	return p
}

// open is the overlay's build: it runs on every Show, with the first card
// that can do something picked.
func (p *deckPanel) open(state game.GameState) Primitive {
	p.state, p.sel, p.scroll, p.said = state, 0, 0, nil
	pg := p.page(state)
	for n := 0; n < pg.count(); n++ {
		if c, _ := pg.card(n); c.Command != "" {
			p.sel = n
			break
		}
	}
	return overBar(p)
}

// update takes a new snapshot.
func (p *deckPanel) update(state game.GameState) { p.state = state }

// reply shows what a command run over the panel did: the command, then the
// first line of its answer, in the bad colour when it failed.
func (p *deckPanel) reply(cmd string, res CommandResult) {
	msg, bad := replyLine(res)
	k := xGood
	if bad {
		k = xBad
	}
	if msg == "" {
		msg = "✓" // the command's answer went to the log, behind the panel
	}
	p.said = []xseg{{"› " + cmd + "  ", xDim}, {msg, k}}
}

// Layout takes the panel's rectangle and works the panel out in it.
func (p *deckPanel) Layout(r image.Rectangle) {
	p.Place(r)
	if r.Empty() {
		return
	}
	motion := p.settings == nil || p.settings().Motion
	if t := p.now(); !p.last.IsZero() && motion && (p.held == nil || !p.held()) {
		p.clock += t.Sub(p.last).Seconds()
	}
	p.last = p.now()
	p.pg = p.page(p.state)
	p.pg.clock = p.clock
	p.sel = clamp(p.sel, 0, max(0, p.pg.count()-1))
	if p.draw != nil {
		p.grid, p.lay = p.draw(&p.pg, p.state, p.sel, p.scroll, r.Dx(), r.Dy(), p.said...)
	} else {
		p.grid, p.lay = renderDeck(&p.pg, p.sel, p.scroll, r.Dx(), r.Dy(), p.said...)
	}
	p.scroll = clamp(p.scroll, 0, max(0, p.lay.total-p.lay.viewH))
	p.moving.Store(p.pg.animate && motion && (p.held == nil || !p.held()))
}

// Paint puts the panel on the kit's painter in theme colors.
func (p *deckPanel) Paint(p2 *kit.Painter) { p.paintGrid(p2, p.grid, p.Bounds()) }

// pick moves the pick to card n and scrolls it into view.
func (p *deckPanel) pick(n int) {
	p.sel = clamp(n, 0, max(0, len(p.lay.cards)-1))
	p.scroll = p.lay.reveal(p.sel, p.scroll)
}

// act runs the picked card's command and reports whether it has one.
func (p *deckPanel) act() bool {
	c, ok := p.pg.card(p.sel)
	if !ok || c.Command == "" || p.buttons.run == nil {
		return false
	}
	p.buttons.fire(xkey{cmd: c.Command})
	return true
}

// Mouse: the foot's buttons first; the wheel scrolls; a click picks a card,
// and on its button, or twice on the card, runs its command.
func (p *deckPanel) Mouse(m kit.Mouse) bool {
	switch {
	case p.grid == nil:
		return false
	case p.buttons.mouse(p.grid, p.Bounds().Min, m):
		return true
	case m.Kind == kit.MouseWheel:
		was := p.scroll
		p.scroll = clamp(p.scroll+3*m.Wheel, 0, max(0, p.lay.total-p.lay.viewH))
		return p.scroll != was
	case m.Kind != kit.MousePress || m.Button != kit.ButtonLeft:
		return false
	}
	pt := m.Pt().Sub(p.Bounds().Min)
	n, ok := p.grid.hitAt(hitDeckCard, pt)
	if !ok {
		return false
	}
	p.pick(n)
	if m.Clicks == 2 {
		p.act()
	}
	return true
}

// PointerAt: the hand over a card, its button and the foot's keys.
func (p *deckPanel) PointerAt(at image.Point) kit.Pointer {
	if p.grid == nil {
		return kit.PointerArrow
	}
	pt := at.Sub(p.Bounds().Min)
	_, card := p.grid.hitAt(hitDeckCard, pt)
	_, key := p.grid.keyAt(pt)
	return handIf(card || key)
}

// routeKey takes a key if it is one of the panel's and reports whether it
// did. With something typed in the prompt every key belongs to the prompt,
// and Enter is the panel's only when the picked card has a command.
func (p *deckPanel) routeKey(ev *kit.Key, prompt string) bool {
	if prompt != "" {
		return false
	}
	page := max(1, p.lay.viewH-2)
	switch ev.Code {
	case kit.KeyLeft:
		p.pick(p.lay.step(p.sel, -1, 0))
	case kit.KeyRight:
		p.pick(p.lay.step(p.sel, 1, 0))
	case kit.KeyUp:
		p.pick(p.lay.step(p.sel, 0, -1))
	case kit.KeyDown:
		p.pick(p.lay.step(p.sel, 0, 1))
	case kit.KeyPgUp:
		p.scroll = max(0, p.scroll-page)
	case kit.KeyPgDn:
		p.scroll = clamp(p.scroll+page, 0, max(0, p.lay.total-p.lay.viewH))
	case kit.KeyHome:
		p.pick(0)
	case kit.KeyEnd:
		p.pick(len(p.lay.cards) - 1)
	case kit.KeyEnter:
		return p.act()
	default:
		return false
	}
	return true
}
