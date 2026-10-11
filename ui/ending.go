package ui

import (
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
	"github.com/espresso20/ageforge/ui/mapstyle/all"
)

// ending.go runs the film of a run's ending: the overlay that plays when a
// run ends in a prestige or in a fall to a catastrophe (ending_page.go
// draws it). The dashboard shows it from the record the engine publishes
// (game.RunEnding), with the last snapshot of the old run for the town
// that burns and the new run's state for the fire that is lit.
//
// It plays through by itself and then closes. Any key skips to the next
// beat, and on the last beat closes the film. With the motion setting off
// each beat is one still frame, moved through by key or by a timer.
//
// The game is never held: the new run is already ticking when the film
// starts, and the last beat draws it as it stands. As on the arrival
// screen, nothing here queues an update from the event loop (keys are
// handled on the loop itself), and the timers and the redraw clock queue
// only from goroutines of their own.

// endingPageName is the film's name on the page stack.
const endingPageName = "run_ending"

// ending is the film.
type ending struct {
	*tview.Box
	om *OverlayManager

	view endView
	fed  bool

	// The town that was and the town that is, in the player's map style.
	reg              *mapstyle.Registry
	set              mapSettings
	oldTown, newTown *menuTown

	// The scene and the clock it runs on: the film is at frame base plus
	// what the clock has counted since start.
	sc     *endScene
	scKey  [3]int
	pal    *menuPalette
	now    func() time.Time
	start  time.Time
	base   int
	closed bool
	// quiet: keys before this moment are let go by. keys counts every key
	// the film was given, heeded or not (the freeze harness reads it).
	quiet   time.Time
	keys    int
	stop    chan struct{}
	cancels []func()
	after   func(d time.Duration, fn func()) (cancel func())
}

// ShowRunEnding puts up the film of a run's ending. It is called on the UI
// goroutine by the dashboard's refresh, before the overlay manager is given
// the new run's state: what the manager last saw is the run that ended.
func ShowRunEnding(om *OverlayManager, end game.RunEnding) {
	e := &ending{Box: tview.NewBox(), om: om, reg: all.Registry(), now: time.Now}
	e.set = defaultMapSettings(e.reg)
	e.pal = newMenuPalette(theme.Active())
	e.view = endViewFor(end, e.pal)
	e.settings()
	e.after = om.after
	if e.after == nil {
		e.after = func(d time.Duration, fn func()) func() {
			t := time.AfterFunc(d, func() { om.app.QueueUpdateDraw(fn) })
			return func() { t.Stop() }
		}
	}
	// With the engine at hand the film has the new run from its first
	// frame; without, the next refresh brings it (OverlayManager.Refresh).
	if om.engine != nil {
		cur := om.engine.GetState()
		e.feed(om.seen, &cur)
	}
	om.dropArrival() // an arrival or a film still up gives way to this one
	if om.active != "" {
		om.pages.RemovePage(om.active)
	}
	om.pages.AddPage(endingPageName, e, true, true)
	om.active, om.film, om.focus = endingPageName, e, e
	om.app.SetFocus(e)
	e.start = e.now()
	e.arm()
}

// settings reads the display settings into the view.
func (e *ending) settings() {
	e.view.plain = e.set.Tier == mapmodel.TierASCII
	e.view.motion = e.set.Motion
}

// feed gives the film the game's state. The first time, prev is the last
// snapshot of the run that ended (nil when there is none): the two towns
// are drawn from it and from cur. After that it keeps the new town's
// picture up with the new run.
func (e *ending) feed(prev, cur *game.GameState) {
	if cur == nil || e.closed {
		return
	}
	if e.fed {
		e.newTown.update(cur)
		return
	}
	e.fed = true
	if e.om != nil && e.om.engine != nil {
		e.set = resolveMapSettings(e.om.engine.Account(), e.reg)
		e.settings()
	}
	// One builder where the two runs are on one land.
	builder := mapmodel.NewBuilder(nil)
	if prev != nil && prev.Age == e.view.end.Age {
		e.oldTown = newMenuTownOn(builder, prev, e.reg, e.set.Style, false)
	}
	e.newTown = newMenuTownOn(builder, cur, e.reg, e.set.Style, true)
	e.view.captions = menuCaptions(e.newTown, cur)
	e.sc = nil
	if !e.start.IsZero() {
		e.arm() // the settings may have changed what the beats are
	}
}

// spans is the film's beats, from the scene when there is one.
func (e *ending) spans() []endSpan {
	if e.sc != nil {
		return e.sc.spans
	}
	return endSpans(e.view.prestige, e.view.motion, endReckFor(&e.view, e.pal, 76).frames)
}

// pos is the frame the film is at by its clock.
func (e *ending) pos() int {
	return e.base + int(e.now().Sub(e.start)/mapAnimStep)
}

// beat is the beat the film is in.
func (e *ending) beat() endSpan {
	spans, at := e.spans(), e.pos()
	for _, s := range spans {
		if at < s.end() {
			return s
		}
	}
	return spans[len(spans)-1]
}

// arm sets the film's timers from where it is: a redraw as each beat ends
// (a still frame has no clock to draw the next one), and the close at the
// end of the last.
func (e *ending) arm() {
	for _, cancel := range e.cancels {
		cancel()
	}
	e.cancels = nil
	spans, at := e.spans(), e.pos()
	for i, s := range spans {
		if s.end() <= at {
			continue
		}
		wait := time.Duration(s.end()-at) * mapAnimStep
		if i == len(spans)-1 {
			e.cancels = append(e.cancels, e.after(wait, e.close))
		} else {
			e.cancels = append(e.cancels, e.after(wait, func() {}))
		}
	}
}

// close takes the film down and gives the keyboard back to the prompt. It
// is safe to call twice.
func (e *ending) close() {
	if e.closed {
		return
	}
	e.closed = true
	e.stopClock()
	for _, cancel := range e.cancels {
		cancel()
	}
	e.cancels = nil
	e.oldTown.close()
	e.newTown.close()
	if e.om == nil {
		return
	}
	// Only if the film is still the manager's (see arrival.close).
	if e.om.film == e {
		e.om.film = nil
		if e.om.active == endingPageName {
			e.om.Hide()
		}
	}
}

// startClock starts the redraw clock (see arrival.startClock).
func (e *ending) startClock() {
	if e.stop != nil || e.om == nil || e.om.app == nil {
		return
	}
	stop := make(chan struct{})
	e.stop = stop
	app := e.om.app
	go func() {
		t := time.NewTimer(mapAnimStep)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				asked := time.Now()
				app.QueueUpdateDraw(func() {})
				t.Reset(max(mapAnimStep-time.Since(asked), arrivalBreath))
			case <-stop:
				return
			}
		}
	}()
}

func (e *ending) stopClock() {
	if e.stop != nil {
		close(e.stop)
		e.stop = nil
	}
}

// InputHandler: any key skips to the next beat, and on the last beat
// closes the film.
func (e *ending) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return e.WrapInputHandler(func(_ *tcell.EventKey, _ func(tview.Primitive)) { e.key() })
}

// key is a key press, whichever key it was.
func (e *ending) key() {
	e.keys++
	now := e.now()
	if e.closed || now.Before(e.quiet) {
		return
	}
	e.quiet = now.Add(arrivalKeyGap)
	spans, cur := e.spans(), e.beat()
	if cur.beat == spans[len(spans)-1].beat {
		e.close()
		return
	}
	// The next beat starts now.
	e.base, e.start = cur.end(), now
	if e.sc != nil {
		e.sc.seek(e.base)
	}
	e.arm()
}

// Draw draws the film at the clock's frame.
func (e *ending) Draw(scr tcell.Screen) {
	if e.closed {
		return
	}
	x, y, w, h := e.GetRect()
	if w <= 0 || h <= 0 {
		return
	}
	grid := e.frameGrid(w, h)
	grid.flush(scr, x, y, e.pal, e.view.plain)
	// The theme's own weather, at full, from the heavy blow to the film's
	// end: twice over, in the cells the film leaves empty.
	if s, f := e.sc.span(e.sc.frame); e.view.motion && e.view.prestige &&
		(s.beat == beatStrike && f >= endBlow || s.beat == beatBeginning) {
		if effect := theme.Active().Effect; effect != "" {
			// The effect that comes in bursts (the glitch, every 64
			// frames from its 57th) bursts as the new land comes out of
			// the light.
			dawn := e.sc.spans[len(e.sc.spans)-1].from + 9
			n := e.sc.frame + 57 - dawn + 64
			drawThemeEffect(scr, x, y, w, h, effect, n, e.view.plain)
			drawThemeEffect(scr, x, y, w, h, effect, n+29, e.view.plain)
		}
	}
	if e.view.motion {
		e.startClock()
	} else {
		e.stopClock()
	}
}

// frameGrid brings the scene to the clock's frame and draws the film at w
// by h.
func (e *ending) frameGrid(w, h int) *mGrid {
	if th := theme.Active(); e.pal == nil || e.pal.key != th.Key {
		e.pal = newMenuPalette(th)
		e.sc = nil
	}
	v := &e.view
	motion := 0
	if v.motion {
		motion = 1
	}
	at := e.pos()
	if key := [3]int{w, h, motion}; e.sc == nil || key != e.scKey {
		// A new scene (the first draw, a new size, a new theme): it is
		// played from the start of the beat the film is in.
		e.sc, e.scKey = newEndScene(endLayoutFor(w, h, v), v, e.pal), key
		e.sc.seek(min(at, e.sc.total()-1))
	}
	sc := e.sc
	if !v.motion {
		s, _ := sc.span(min(at, sc.total()-1))
		sc.pose(s)
	} else {
		for sc.frame < min(at, sc.total()-1) {
			sc.step()
		}
	}
	mf := mapFrame(nil, e.set, e.now().Sub(e.start))
	return sc.render(e.pal, v, e.oldTown, e.newTown, mf, !v.motion)
}
