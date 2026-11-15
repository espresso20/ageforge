package ui

import (
	"image"
	"sync/atomic"
	"time"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/vibe"
	"github.com/espresso20/ageforge/window/kit"
)

// expedition_panel.go is the Expeditions panel's overlay primitive: the
// Ready Room and the Expedition Report over the whole screen but the
// command bar (expedition_view.go draws them). Like the Map, Research and
// badge panels it leaves the keyboard with the command bar and takes only
// the keys that print nothing, and only while the bar is empty: the arrows
// hop between pins (or pick a party on the Report), Tab switches views,
// Enter asks to send the highlighted offer and Enter again sends it. Esc
// takes back the question, then closes the panel. Anything typed goes to
// the bar, where `expedition <name>` sends an offer by its name.

// expeditionKeys are the panel's keys as the Help panel lists them.
var expeditionKeys = []keyHelp{
	{"Arrows", vibe.E("ui.expedition_panel.expeditionKeys.02", "Hop from pin to pin on the Ready Room; pick a party on the Report")},
	{"Tab", vibe.E("ui.expedition_panel.expeditionKeys.04", "Switch between the Ready Room and the Expedition Report")},
	{"Enter", vibe.E("ui.expedition_panel.expeditionKeys.06", "Ask to send the highlighted offer, and send it (with something typed, Tab and Enter act on the prompt instead)")},
	{"Esc", vibe.E("ui.expedition_panel.expeditionKeys.08", "Take back the question, then close the panel")},
}

// expFlashFor is how long a line stays on the key bar.
const expFlashFor = 8 * time.Second

type expeditionPanel struct {
	barPanel
	kit.Base
	// host paints the panel's kit cells and takes the mouse; the keys stay
	// with barPanel.
	grid   *xgrid // what the panel last laid out, and where its pins are
	engine *game.GameEngine
	state  game.GameState
	view   expView
	// selIdx is where the highlighted offer stood among the offers, so a
	// pin that leaves the board hands the highlight to the one that takes
	// its place in the list.
	selIdx int
	// accounted is how many missions had come home when the Report last
	// counted them.
	accounted int
	flashAt   time.Time
	w, h      int // the last size drawn at

	// settings reads the glyph tier and the motion setting.
	settings func() mapSettings

	now   func() time.Time
	start time.Time
	// moving is set by a draw that drew something that moves, so the
	// dashboard's redraw loop runs at the animation rate only while there
	// is a reason to.
	moving atomic.Bool
}

func newExpeditionPanel() *expeditionPanel {
	p := &expeditionPanel{w: 100, h: 30, now: time.Now, start: time.Now()}
	p.barPanel = newBarPanel(p.routeKey)
	return p
}

// open is the overlay's build: it runs on every Show. The panel opens on
// the Report while a party is out or one has come home since it was last
// looked at, and on the Ready Room otherwise.
func (p *expeditionPanel) open(state game.GameState) Primitive {
	mil := state.Military
	p.state = state
	p.view = expView{}
	p.selIdx, p.accounted = 0, mil.ReturnsMade
	if len(mil.Active) > 0 || mil.ReturnsMade > mil.ReturnsSeen {
		p.setReport(true)
	}
	p.update(state)
	return overBar(p)
}

// scene is the map as it stands, on the panel's size.
func (p *expeditionPanel) scene() *expScene {
	l := layoutFor(p.w, p.h)
	if p.view.report {
		l = reportLayoutFor(p.w, p.h)
	}
	if l.small {
		return buildScene(p.state, 1, 1)
	}
	return buildScene(p.state, l.mapW-2, l.mainH-2)
}

// update takes a new snapshot: the highlight follows its offer, or stays
// at its place when the offer is gone, and the Report counts what came
// home while it was open.
func (p *expeditionPanel) update(state game.GameState) {
	p.state = state
	sc := p.scene()
	if n := len(sc.offers); n == 0 {
		p.view.sel, p.selIdx = "", 0
	} else {
		i := p.selIdx
		for j, it := range sc.offers {
			if it.key == p.view.sel {
				i = j
			}
		}
		p.selIdx = clamp(i, 0, n-1)
		p.view.sel = sc.offers[p.selIdx].key
	}
	if n := len(sc.acts); n == 0 {
		p.view.party = 0
	} else {
		p.view.party = clamp(p.view.party, 0, n-1)
	}
	if mil := state.Military; p.view.report && mil.ReturnsMade > p.accounted {
		p.view.marked += mil.ReturnsMade - p.accounted
		p.accounted = mil.ReturnsMade
		p.markSeen()
	}
	p.view.marked = min(p.view.marked, len(state.Military.Returns))
}

// markSeen tells the game the Report was viewed.
func (p *expeditionPanel) markSeen() {
	if p.engine != nil {
		p.engine.MarkReturnsSeen()
	}
	p.state.Military.ReturnsSeen = p.state.Military.ReturnsMade
}

// setReport switches views. Viewing the Report clears the new count (the
// returns that were new stay marked while it is open).
func (p *expeditionPanel) setReport(report bool) {
	p.view.confirm, p.view.flash = false, ""
	p.view.report = report
	p.view.marked = 0
	if report {
		mil := p.state.Military
		p.view.marked = min(max(0, mil.ReturnsMade-mil.ReturnsSeen), len(mil.Returns))
		p.accounted = mil.ReturnsMade
		p.markSeen()
	}
}

// frame is the animation frame: the clock's, or atRest when the motion
// setting is off.
func (p *expeditionPanel) frame(motion bool) int {
	if !motion {
		return atRest
	}
	return int(p.now().Sub(p.start) / (4 * mapAnimStep))
}

// Layout takes the panel's rectangle and works the panel out in it.
func (p *expeditionPanel) Layout(r image.Rectangle) {
	p.Place(r)
	if r.Empty() {
		return
	}
	p.w, p.h = r.Dx(), r.Dy()
	motion, plain := true, false
	if p.settings != nil {
		set := p.settings()
		motion, plain = set.Motion, set.Tier == mapmodel.TierASCII
	}
	v := p.view
	v.frame = p.frame(motion)
	if v.flash != "" && p.now().Sub(p.flashAt) > expFlashFor {
		p.view.flash, v.flash = "", ""
	}
	p.grid = renderExpedition(p.state, v, p.w, p.h)
	if plain {
		p.grid.fold()
	}
	sc := p.scene()
	p.moving.Store(motion && (len(sc.acts) > 0 && v.report || len(sc.offers) > 0 && !v.report))
}

// Paint puts the panel on the kit's painter in theme colors.
func (p *expeditionPanel) Paint(p2 *kit.Painter) { p.paintGrid(p2, p.grid, p.Bounds()) }

// Mouse: a click on an offer's pin highlights it, as the arrows do, and a
// double click asks to send it, as Enter does; on the Report a click on a
// party picks it. A click outside the question before a party is sent takes
// it back, as Esc does. The wheel moves the highlight a notch, as Up and Down
// do.
func (p *expeditionPanel) Mouse(m kit.Mouse) bool {
	switch {
	case !p.view.confirm && p.buttons.mouse(p.grid, p.Bounds().Min, m):
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
	if p.view.confirm {
		_, in := p.grid.hitAt(hitConfirm, pt)
		return !in && p.closeConfirm()
	}
	if p.view.report {
		n, ok := p.grid.hitAt(hitParty, pt)
		if ok {
			p.view.flash, p.view.party = "", n
		}
		return ok
	}
	n, ok := p.grid.hitAt(hitOffer, pt)
	if !ok {
		return false
	}
	p.view.flash = ""
	if sc := p.scene(); n < len(sc.offers) {
		p.selIdx, p.view.sel = n, sc.offers[n].key
	}
	if m.Clicks == 2 {
		p.handleKey(kit.KeyEnter)
	}
	return true
}

// PointerAt: the hand over an offer's pin on the board and over a party on
// the report; with the confirm up, nothing behind it takes a click.
func (p *expeditionPanel) PointerAt(at image.Point) kit.Pointer {
	if p.grid == nil || p.view.confirm {
		return kit.PointerArrow
	}
	kind := hitOffer
	if p.view.report {
		kind = hitParty
	}
	pt := at.Sub(p.Bounds().Min)
	_, ok := p.grid.hitAt(kind, pt)
	_, key := p.grid.keyAt(pt)
	return handIf(ok || key)
}

// routeKey takes a key if it is one of the panel's and reports whether it
// did. With something typed in the prompt every key belongs to the prompt.
func (p *expeditionPanel) routeKey(ev *kit.Key, prompt string) bool {
	switch ev.Code {
	case kit.KeyUp, kit.KeyDown, kit.KeyLeft, kit.KeyRight, kit.KeyTab, kit.KeyBacktab, kit.KeyEnter:
		if prompt != "" {
			return false
		}
	default:
		return false
	}
	p.handleKey(ev.Code)
	return true
}

// closeConfirm takes back the question before a party is sent, and
// reports whether there was one: Esc does it first, then closes the panel.
func (p *expeditionPanel) closeConfirm() bool {
	if !p.view.confirm {
		return false
	}
	p.view.confirm = false
	return true
}

func (p *expeditionPanel) handleKey(key kit.KeyCode) {
	p.view.flash = ""
	if key == kit.KeyTab || key == kit.KeyBacktab {
		p.setReport(!p.view.report)
		return
	}
	if p.view.confirm {
		if key == kit.KeyEnter {
			p.send()
		}
		return
	}
	if p.view.report {
		switch key {
		case kit.KeyUp:
			p.view.party = max(0, p.view.party-1)
		case kit.KeyDown:
			p.view.party = min(max(0, len(p.scene().acts)-1), p.view.party+1)
		}
		return
	}
	sc := p.scene()
	if len(sc.offers) == 0 {
		return
	}
	switch key {
	case kit.KeyUp, kit.KeyDown, kit.KeyLeft, kit.KeyRight:
		dx, dy := 0, 0
		switch key {
		case kit.KeyUp:
			dy = -1
		case kit.KeyDown:
			dy = 1
		case kit.KeyLeft:
			dx = -1
		default:
			dx = 1
		}
		if to := hopTo(sc.pinCells(), sc.selected(p.view.sel), dx, dy); to >= 0 {
			p.selIdx, p.view.sel = to, sc.offers[to].key
		}
	case kit.KeyEnter:
		o := sc.offers[sc.selected(p.view.sel)].offer
		if o.CanLaunch {
			p.view.confirm = true
		} else {
			p.say(refusalText(*o), false)
		}
	}
}

// say puts a line on the key bar.
func (p *expeditionPanel) say(text string, good bool) {
	p.view.flash, p.view.flashGood, p.flashAt = text, good, p.now()
}

// send sends the highlighted offer through the same call as the typed
// command.
func (p *expeditionPanel) send() {
	p.view.confirm = false
	sc := p.scene()
	if len(sc.offers) == 0 {
		return
	}
	o := *sc.offers[sc.selected(p.view.sel)].offer
	if p.engine == nil {
		return
	}
	if err := p.engine.LaunchExpedition(o.Number); err != nil {
		p.say(textfmt.Sentence(err.Error()), false)
	} else {
		p.say(vibe.Line("ui.expedition_panel.send.01", "Sent: ")+goingTo(o.Label, o.Place)+vibe.Line("ui.expedition_panel.send.02", ". Back in ")+expDur(o.DurationMin, p.state)+vibe.Line("ui.expedition_panel.send.03", ". Tab to follow it."), true)
	}
	p.update(p.engine.GetState())
}
