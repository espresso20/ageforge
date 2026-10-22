package ui

import (
	"sync/atomic"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/pkg/textfmt"
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
var expeditionKeys = [][2]string{
	{"Arrows", "Hop from pin to pin on the Ready Room; pick a party on the Report"},
	{"Tab", "Switch between the Ready Room and the Expedition Report"},
	{"Enter", "Ask to send the highlighted offer, and send it (with something typed, Tab and Enter act on the prompt instead)"},
	{"Esc", "Take back the question, then close the panel"},
}

// expFlashFor is how long a line stays on the key bar.
const expFlashFor = 8 * time.Second

type expeditionPanel struct {
	*tview.Box
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
	// prompt reads the command bar ("" when it is empty).
	prompt func() string
	// toPrompt hands a key the panel does not take to the command bar.
	toPrompt func(ev *tcell.EventKey)

	now   func() time.Time
	start time.Time
	// moving is set by a draw that drew something that moves, so the
	// dashboard's redraw loop runs at the animation rate only while there
	// is a reason to.
	moving atomic.Bool
}

func newExpeditionPanel() *expeditionPanel {
	return &expeditionPanel{Box: tview.NewBox(), w: 100, h: 30, now: time.Now, start: time.Now()}
}

// open is the overlay's build: it runs on every Show. The panel opens on
// the Report while a party is out or one has come home since it was last
// looked at, and on the Ready Room otherwise.
func (p *expeditionPanel) open(state game.GameState) tview.Primitive {
	mil := state.Military
	p.state = state
	p.view = expView{}
	p.selIdx, p.accounted = 0, mil.ReturnsMade
	if mil.ActiveShort != nil || mil.ActiveLong != nil || mil.ActiveGrand != nil || mil.ReturnsMade > mil.ReturnsSeen {
		p.setReport(true)
	}
	p.update(state)
	return tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(p, 0, 1, true).
		AddItem(nil, promptRows, 0, false)
}

// setSize tells the panel how big it is drawn, for the tests; a draw does
// it by itself.
func (p *expeditionPanel) setSize(w, h int) { p.w, p.h = w, h }

// scene is the map as it stands, on the panel's size.
func (p *expeditionPanel) scene() *expScene {
	l := layoutFor(p.w, p.h)
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

// motion reports whether the game's motion setting is on.
func (p *expeditionPanel) motion() bool { return p.settings == nil || p.settings().Motion }

// frame is the animation frame: the clock's, or atRest when the motion
// setting is off.
func (p *expeditionPanel) frame() int {
	if !p.motion() {
		return atRest
	}
	return int(p.now().Sub(p.start) / (4 * mapAnimStep))
}

// Draw paints the panel.
func (p *expeditionPanel) Draw(scr tcell.Screen) {
	x, y, w, h := p.GetRect()
	if w <= 0 || h <= 0 {
		return
	}
	p.w, p.h = w, h
	v := p.view
	v.frame = p.frame()
	if v.flash != "" && p.now().Sub(p.flashAt) > expFlashFor {
		p.view.flash, v.flash = "", ""
	}
	grid := renderExpedition(p.state, v, w, h)
	if p.settings != nil && p.settings().Tier == mapmodel.TierASCII {
		grid.fold()
	}
	paintExpedition(scr, grid, x, y, newExpPalette())
	sc := p.scene()
	p.moving.Store(p.motion() && (len(sc.acts) > 0 && v.report || len(sc.offers) > 0 && !v.report))
}

// InputHandler takes the panel's keys when it has the focus itself and
// hands the rest to the command bar.
func (p *expeditionPanel) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
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
// did. With something typed in the prompt every key belongs to the prompt.
func (p *expeditionPanel) routeKey(ev *tcell.EventKey, prompt string) bool {
	switch ev.Key() {
	case tcell.KeyUp, tcell.KeyDown, tcell.KeyLeft, tcell.KeyRight, tcell.KeyTab, tcell.KeyBacktab, tcell.KeyEnter:
		if prompt != "" {
			return false
		}
	default:
		return false
	}
	p.handleKey(ev.Key())
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

func (p *expeditionPanel) handleKey(key tcell.Key) {
	p.view.flash = ""
	if key == tcell.KeyTab || key == tcell.KeyBacktab {
		p.setReport(!p.view.report)
		return
	}
	if p.view.confirm {
		if key == tcell.KeyEnter {
			p.send()
		}
		return
	}
	if p.view.report {
		switch key {
		case tcell.KeyUp:
			p.view.party = max(0, p.view.party-1)
		case tcell.KeyDown:
			p.view.party = min(max(0, len(p.scene().acts)-1), p.view.party+1)
		}
		return
	}
	sc := p.scene()
	if len(sc.offers) == 0 {
		return
	}
	switch key {
	case tcell.KeyUp, tcell.KeyDown, tcell.KeyLeft, tcell.KeyRight:
		dx, dy := 0, 0
		switch key {
		case tcell.KeyUp:
			dy = -1
		case tcell.KeyDown:
			dy = 1
		case tcell.KeyLeft:
			dx = -1
		default:
			dx = 1
		}
		if to := hopTo(sc.pinCells(), sc.selected(p.view.sel), dx, dy); to >= 0 {
			p.selIdx, p.view.sel = to, sc.offers[to].key
		}
	case tcell.KeyEnter:
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
		p.say("Sent: "+goingTo(o.Label, o.Place)+". Back in "+expDur(o.DurationMin, p.state)+". Tab to follow it.", true)
	}
	p.update(p.engine.GetState())
}
