package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
)

// people_panel.go is the People panel's overlay primitive: the whole screen
// but the command bar (people_view.go draws it). Like the Expeditions panel
// it leaves the keyboard with the command bar and takes only the keys that
// print nothing, and only while the bar is empty: Left and Right switch
// tabs, Up and Down pick a row. Esc closes it. Anything typed goes to the
// bar, where `roster <area> <percent>` sets a share.

type peoplePanel struct {
	*tview.Box
	state game.GameState
	view  peopleView
	w, h  int // the last size drawn at

	// settings reads the glyph tier.
	settings func() mapSettings
	// toPrompt hands a key the panel does not take to the command bar.
	toPrompt func(ev *tcell.EventKey)
}

func newPeoplePanel() *peoplePanel {
	return &peoplePanel{Box: tview.NewBox(), w: 100, h: 30}
}

// open is the overlay's build: it runs on every Show, on the Overview.
func (p *peoplePanel) open(state game.GameState) tview.Primitive {
	p.state = state
	p.view = peopleView{}
	return tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(p, 0, 1, true).
		AddItem(nil, promptRows, 0, false)
}

// setSize tells the panel how big it is drawn, for the tests; a draw does it
// by itself.
func (p *peoplePanel) setSize(w, h int) { p.w, p.h = w, h }

// update takes a new snapshot; the picked row stays in range.
func (p *peoplePanel) update(state game.GameState) {
	p.state = state
	p.view.row = clamp(p.view.row, 0, peopleRows(state, p.view.tab)-1)
}

// Draw paints the panel.
func (p *peoplePanel) Draw(scr tcell.Screen) {
	x, y, w, h := p.GetRect()
	if w <= 0 || h <= 0 {
		return
	}
	p.w, p.h = w, h
	plain := false
	if p.settings != nil {
		plain = p.settings().Tier == mapmodel.TierASCII
	}
	grid := renderPeople(p.state, p.view, w, h)
	if plain {
		grid.fold()
	}
	paintExpedition(scr, grid, x, y, newExpPalette())
}

// InputHandler takes the panel's keys when it has the focus itself and
// hands the rest to the command bar.
func (p *peoplePanel) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return p.WrapInputHandler(func(ev *tcell.EventKey, _ func(tview.Primitive)) {
		if !p.routeKey(ev, "") && p.toPrompt != nil {
			p.toPrompt(ev)
		}
	})
}

// routeKey takes a key if it is one of the panel's and reports whether it
// did. With something typed in the prompt every key belongs to the prompt.
func (p *peoplePanel) routeKey(ev *tcell.EventKey, prompt string) bool {
	switch ev.Key() {
	case tcell.KeyUp, tcell.KeyDown, tcell.KeyLeft, tcell.KeyRight:
		if prompt != "" {
			return false
		}
	default:
		return false
	}
	p.handleKey(ev.Key())
	return true
}

func (p *peoplePanel) handleKey(key tcell.Key) {
	switch key {
	case tcell.KeyLeft:
		p.view.tab = (p.view.tab + len(peopleTabs) - 1) % len(peopleTabs)
		p.view.row = 0
	case tcell.KeyRight:
		p.view.tab = (p.view.tab + 1) % len(peopleTabs)
		p.view.row = 0
	case tcell.KeyUp:
		p.view.row = max(0, p.view.row-1)
	case tcell.KeyDown:
		p.view.row = min(max(0, peopleRows(p.state, p.view.tab)-1), p.view.row+1)
	}
}
