package ui

import (
	"image"
	"time"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/vibe"
	"github.com/espresso20/ageforge/window/kit"
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
	kit.Base
	// host paints the panel's kit cells and takes the mouse; the keys stay
	// with barPanel.
	grid   *xgrid // what the panel last laid out, and where its pillars are
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
	d.advancePanel.toPrompt = d.handToPrompt
	d.advancePanel.buttons.wire(d.app)
	d.overlayMgr.RegisterWidget(advancePanelName, vibe.E("ui.advance_panel.registerAdvancePanel.01", "Advance"), d.advancePanel.open, d.advancePanel.update)
}

// open is the overlay's build: it runs on every Show, on the first pillar.
func (p *advancePanel) open(state game.GameState) Primitive {
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

// Layout takes the panel's rectangle and works the panel out in it.
func (p *advancePanel) Layout(r image.Rectangle) {
	p.Place(r)
	if r.Empty() {
		return
	}
	u := p.ui
	if u.flash != "" && p.now().Sub(p.flashAt) > advFlashFor {
		p.ui.flash, u.flash = "", ""
	}
	p.grid = renderAdvance(p.view, u, r.Dx(), r.Dy())
	if p.settings != nil && p.settings().Tier == mapmodel.TierASCII {
		p.grid.fold()
	}
}

// Paint puts the panel on the kit's painter in theme colors.
func (p *advancePanel) Paint(p2 *kit.Painter) { p.paintGrid(p2, p.grid, p.Bounds()) }

// Mouse: a click on a pillar picks it and a double click advances when all
// four stand, as Enter does; the wheel picks a pillar a notch, as Up and Down
// do.
func (p *advancePanel) Mouse(m kit.Mouse) bool {
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
	n, ok := p.grid.hitAt(hitPillar, m.Pt().Sub(p.Bounds().Min))
	if !ok {
		return false
	}
	// A click only picks. Advancing the age stays on Enter and the advance
	// command: it is too large a thing to hang on a double click.
	p.ui.flash, p.ui.sel = "", n
	return true
}

// PointerAt: the hand over a pillar.
func (p *advancePanel) PointerAt(at image.Point) kit.Pointer {
	if p.grid == nil {
		return kit.PointerArrow
	}
	pt := at.Sub(p.Bounds().Min)
	_, pillar := p.grid.hitAt(hitPillar, pt)
	_, key := p.grid.keyAt(pt)
	return handIf(pillar || key)
}

// routeKey takes a key if it is one of the panel's and reports whether it
// did. With something typed in the prompt every key belongs to the prompt.
func (p *advancePanel) routeKey(ev *kit.Key, prompt string) bool {
	switch ev.Code {
	case kit.KeyUp, kit.KeyDown, kit.KeyEnter:
		if prompt != "" {
			return false
		}
	default:
		return false
	}
	p.handleKey(ev.Code)
	return true
}

func (p *advancePanel) handleKey(key kit.KeyCode) {
	p.ui.flash = ""
	switch key {
	case kit.KeyUp:
		p.ui.sel = max(0, p.ui.sel-1)
	case kit.KeyDown:
		p.ui.sel = min(max(0, len(p.view.Pillars)-1), p.ui.sel+1)
	case kit.KeyEnter:
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
