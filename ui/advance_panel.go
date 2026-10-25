package ui

import (
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/vibe"
)

// advance_panel.go is the Advance panel's overlay primitive: the whole screen
// but the command bar (advance_view.go draws it). Like the People and
// Expeditions panels it leaves the keyboard with the command bar and takes
// only the keys that print nothing, and only while the bar is empty: Up and
// Down pick a pillar, Enter advances when all four stand. Esc closes it.
// Anything typed goes to the bar. Opening it, moving about it and closing it
// change nothing in the game; only Enter on a met gate does, through the
// call the advance command makes.

// advFlashFor is how long a line stays on the status row.
const advFlashFor = 8 * time.Second

type advancePanel struct {
	barPanel
	engine *game.GameEngine
	view   game.AdvanceView
	ui     advanceUI
	// flashAt is when the status row's line was set.
	flashAt time.Time

	// settings reads the glyph tier.
	settings func() mapSettings

	now func() time.Time
}

func newAdvancePanel() *advancePanel {
	p := &advancePanel{now: time.Now}
	p.barPanel = newBarPanel(p.routeKey)
	return p
}

// advancePanelName is the overlay's name, and the command that opens it.
const advancePanelName = "charter"

// registerAdvancePanel gives the panel the dashboard's engine, settings and
// command bar, and registers it as an overlay. Like the other panels it
// leaves the keyboard with the command bar. (It is a method of its own and
// not part of NewDashboard so that the panel's title does not move the
// numbers of NewDashboard's text rows, which the VibeMode script names.)
func (d *Dashboard) registerAdvancePanel() {
	d.advancePanel.engine = d.engine
	d.advancePanel.settings = d.mapSettings
	d.advancePanel.prompt = d.researchPanel.prompt
	d.advancePanel.toPrompt = d.researchPanel.toPrompt
	d.overlayMgr.RegisterWidget(advancePanelName, vibe.E("ui.advance_panel.registerAdvancePanel.01", "Advance"), d.advancePanel.open, d.advancePanel.update, true)
}

// open is the overlay's build: it runs on every Show, on the first pillar.
func (p *advancePanel) open(state game.GameState) tview.Primitive {
	p.ui = advanceUI{}
	p.update(state)
	return overBar(p)
}

// update takes a new snapshot. The calendar is held by the engine, so the year
// the panel shows does not go back within an age.
func (p *advancePanel) update(state game.GameState) {
	v := state.AdvanceView()
	if p.engine != nil {
		v = p.engine.HoldCalendar(v)
	}
	// A line on the foot is about the gate as it stood: when the gate has
	// since been met or has slipped, the status says so instead.
	if v.Ready != p.view.Ready || v.Age != p.view.Age {
		p.ui.flash = ""
	}
	p.view = v
	p.ui.sel = clamp(p.ui.sel, 0, max(0, len(v.Pillars)-1))
}

// Draw paints the panel.
func (p *advancePanel) Draw(scr tcell.Screen) {
	x, y, w, h := p.GetRect()
	if w <= 0 || h <= 0 {
		return
	}
	plain := false
	if p.settings != nil {
		plain = p.settings().Tier == mapmodel.TierASCII
	}
	u := p.ui
	if u.flash != "" && p.now().Sub(p.flashAt) > advFlashFor {
		p.ui.flash, u.flash = "", ""
	}
	grid := renderAdvance(p.view, u, w, h)
	if plain {
		grid.fold()
	}
	paintExpedition(scr, grid, x, y, newExpPalette())
}

// routeKey takes a key if it is one of the panel's and reports whether it
// did. With something typed in the prompt every key belongs to the prompt.
func (p *advancePanel) routeKey(ev *tcell.EventKey, prompt string) bool {
	switch ev.Key() {
	case tcell.KeyUp, tcell.KeyDown, tcell.KeyEnter:
		if prompt != "" {
			return false
		}
	default:
		return false
	}
	p.handleKey(ev.Key())
	return true
}

func (p *advancePanel) handleKey(key tcell.Key) {
	p.ui.flash = ""
	switch key {
	case tcell.KeyUp:
		p.ui.sel = max(0, p.ui.sel-1)
	case tcell.KeyDown:
		p.ui.sel = min(max(0, len(p.view.Pillars)-1), p.ui.sel+1)
	case tcell.KeyEnter:
		p.advance()
	}
}

// say puts a line on the status row.
func (p *advancePanel) say(text string, good bool) {
	p.ui.flash, p.ui.flashGood, p.flashAt = text, good, p.now()
}

// advance asks for the advance when the four pillars stand, through the call
// the advance command makes; the engine's own refusal (a doom in the way, a
// catastrophe pending) is the panel's line when it says no.
func (p *advancePanel) advance() {
	v := p.view
	switch {
	case v.Final:
		p.say(vibe.Line("ui.advance_panel.advance.01", "This is the final age. There is no age beyond it."), false)
		return
	case !v.AllStand:
		var open []string
		for _, pl := range v.Pillars {
			if !pl.Met {
				open = append(open, pl.Name)
			}
		}
		p.say(vibe.Line("ui.advance_panel.advance.02", "Not yet: ")+textfmt.List(open)+vibe.Line("ui.advance_panel.advance.03", " still to stand."), false)
		return
	case p.engine == nil:
		return
	}
	err := p.engine.AdvanceAge()
	p.update(p.engine.GetState())
	if err != nil {
		p.say(textfmt.Sentence(err.Error()), false)
	} else {
		p.say(vibe.Line("ui.advance_panel.advance.04", "Advanced to the ")+v.NextName+".", true)
	}
}
