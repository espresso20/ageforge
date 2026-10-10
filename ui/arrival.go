package ui

import (
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/ui/mapstyle"
	"github.com/espresso20/ageforge/ui/mapstyle/all"
)

// arrival.go runs the arrival screen: the overlay that greets a new age
// (arrival_page.go draws it). It takes the place of the old age splash and
// keeps its name on the page stack, "age_splash", which the dashboard and
// the freeze harness know it by.
//
// It plays in two stages. The celebration is a few seconds on a fixed
// script and moves on by itself; any key moves on at once. The information
// stays until a key closes it, or until the screen has been up twenty
// seconds, as the old splash did. With the motion setting off the
// celebration is one still frame and nothing waits on an animation.
//
// The game is never held: the screen is a page over the dashboard, the
// engine ticks on its own goroutine, and nothing here takes the engine's
// lock but a snapshot. The screen has frozen the game before, each time
// through the event loop: so no handler here queues an update (keys are
// handled on the loop itself), and the timers and the redraw clock only
// ever queue from goroutines of their own.

// arrivalPageName is the screen's name on the page stack.
const arrivalPageName = "age_splash"

const (
	// arrivalFxLead sets a theme's ambient effect ahead of the script, so
	// that the one that comes in bursts (the glitch, every 64 frames from
	// its 57th) bursts as the era's name gives way to the age's.
	arrivalFxLead = 57 - arrEraFrames
	// arrivalHold is how long the screen stays up when no key is pressed.
	arrivalHold = 20 * time.Second
	// arrivalKeyGap is the least time between two keys that each count: a
	// key held down, or pressed twice by one tap, does not close both
	// stages at once.
	arrivalKeyGap = 200 * time.Millisecond
)

// The arrival's stages.
const (
	arrCelebrating = iota
	arrInforming
	arrClosed
)

// arrival is the arrival screen.
type arrival struct {
	*tview.Box
	om *OverlayManager

	// What the dashboard said arrived.
	oldAge, newAge string
	summary        game.AgeAdvanceSummary
	epoch          bool
	event          game.EpochEventRecord

	// What the screen says (view), from that and from the game's state
	// once it has one (feed).
	view arrivalView
	fed  bool

	// The town, before and after, in the player's map style.
	reg              *mapstyle.Registry
	set              mapSettings
	oldTown, newTown *menuTown

	// The scene and the clock it runs on.
	sc      *arrivalScene
	scKey   [4]int
	pal     *menuPalette
	info    infoLayout
	infoKey [3]int
	now     func() time.Time
	start   time.Time
	base    int
	stage   int
	page    int
	lastKey time.Time
	stop    chan struct{}
	cancels []func()
	// after runs fn on the event loop once d has passed, and returns how to
	// call it off (tests run it by hand).
	after func(d time.Duration, fn func()) (cancel func())
}

// newArrival makes the screen for an advance into newAge.
func newArrival(om *OverlayManager, oldAge, newAge string, summary game.AgeAdvanceSummary,
	epochChanged bool, epochEvent game.EpochEventRecord) *arrival {
	a := &arrival{
		Box: tview.NewBox(), om: om, oldAge: oldAge, newAge: newAge, summary: summary,
		epoch: epochChanged, event: epochEvent, reg: all.Registry(), now: time.Now,
	}
	a.set = defaultMapSettings(a.reg)
	a.after = om.after
	if a.after == nil {
		a.after = func(d time.Duration, fn func()) func() {
			t := time.AfterFunc(d, func() { om.app.QueueUpdateDraw(fn) })
			return func() { t.Stop() }
		}
	}
	a.say(nil)
	return a
}

// say works out what the screen says. cur is the game's state, nil when
// the screen has none (the names then come from the rules alone).
func (a *arrival) say(cur *game.GameState) {
	v := arrivalView{age: game.AgeName(a.newAge), era: a.event.EpochName, epoch: a.epoch}
	if cur != nil {
		set := cur.Ruleset()
		if def, ok := set.Age(a.newAge); ok {
			v.age = def.Name
		}
		if era, ok := set.Era(set.EraOf(a.newAge)); ok {
			v.era, v.eraIcon = era.Name, era.Icon
		}
	}
	v.lines = ageSplashLines(a.newAge, a.summary, a.epoch, a.event)
	plain := func(kind splashKind, nth int) string {
		for _, l := range v.lines {
			if l.kind != kind {
				continue
			}
			if nth == 0 {
				return strings.TrimSpace(plainOf(parseSplashLine(l.text, newMenuPalette(theme.Active()), 0)))
			}
			nth--
		}
		return ""
	}
	v.desc, v.quip, v.prompt = plain(skDesc, 0), plain(skQuip, 0), plain(skPrompt, 0)
	if v.epoch {
		// The old splash's own heading for a new epoch, without its rules.
		v.epochHeading = strings.TrimSpace(strings.Trim(plain(skEpoch, 0), "═ "))
		if v.era == "" {
			v.epoch = false // an epoch with no name: an age like any other
		}
	}
	v.passed = a.view.passed
	a.view = v
	a.settings()
}

// settings reads the display settings into the view.
func (a *arrival) settings() {
	a.view.plain = a.set.Tier == mapmodel.TierASCII
	a.view.motion = a.set.Motion
}

// sameRun reports whether two snapshots are of one game in one sitting:
// the same seed, loaded at the same moment.
func sameRun(a, b *game.GameState) bool {
	key := func(s *game.GameState) (int, int64) {
		if s.SessionStart == nil {
			return -1, 0
		}
		return s.SessionStart.Tick, s.SessionStart.SavedAt.UnixNano()
	}
	at, as := key(a)
	bt, bs := key(b)
	return a.Seed == b.Seed && at == bt && as == bs
}

// stateAt is a snapshot as it would read in another age of the same game:
// the town as it stands, in that age's name and era. The picture of the
// age left behind is drawn from it when no snapshot of that age is at hand.
func stateAt(st game.GameState, age string) game.GameState {
	set := st.Ruleset()
	def, ok := set.Age(age)
	if !ok {
		return st
	}
	st.Age, st.AgeName = age, def.Name
	st.NextAge = set.Next(age)
	if next, ok := set.Age(st.NextAge); ok {
		st.NextAgeName = next.Name
	}
	st.EpochKey = set.EraOf(age)
	if era, ok := set.Era(st.EpochKey); ok {
		st.EpochName = era.Name
	}
	return st
}

// feed gives the screen the game's state. The first time, with prev (the
// snapshot before this one, nil when there is none) it works out which age
// was left behind, what was passed on the way and whether an epoch was
// entered, and draws the two towns. After that it keeps the new town's
// picture up with the game.
func (a *arrival) feed(prev, cur *game.GameState) {
	if cur == nil || a.stage == arrClosed {
		return
	}
	if a.fed {
		a.newTown.update(cur)
		return
	}
	a.fed = true
	set := cur.Ruleset()
	keys, at := set.AgeKeys(), set.Indexes()
	ni, known := at[a.newAge]
	if !known {
		a.say(cur)
		return
	}
	// The age left behind: the one the last snapshot was in when it is of
	// this same sitting; else the one the save was in when it was loaded
	// (the advance happened while the player was away); else the age
	// before. Never an age that is not behind this one.
	from := a.oldAge
	switch {
	case prev != nil && sameRun(prev, cur):
		from = prev.Age
	case cur.SessionStart != nil:
		from = cur.SessionStart.Age
	}
	fi, ok := at[from]
	if !ok || fi >= ni {
		fi = ni - 1
	}
	var passed []string
	for i := fi + 1; i < ni; i++ {
		if def, ok := set.Age(keys[i]); ok {
			passed = append(passed, def.Name)
		}
	}
	if fi >= 0 {
		from = keys[fi]
		if set.EraOf(from) != set.EraOf(a.newAge) {
			a.epoch = true
		}
	}
	if a.epoch && a.event.EpochKey == "" {
		// The dashboard did not see the epoch change (it happened over
		// several ages): the event it brought is the last on record.
		for i := len(cur.EpochEventHistory) - 1; i >= 0; i-- {
			if rec := cur.EpochEventHistory[i]; rec.EpochKey == set.EraOf(a.newAge) {
				a.event = rec
				break
			}
		}
	}
	a.view.passed = ""
	if len(passed) > 0 {
		a.view.passed = "Ages passed: " + strings.Join(passed, ", ")
	}
	a.say(cur)

	// The two towns, in the player's style.
	if a.om != nil && a.om.engine != nil {
		a.set = resolveMapSettings(a.om.engine.Account(), a.reg)
		a.settings()
	}
	before := stateAt(*cur, from)
	if prev != nil && sameRun(prev, cur) && prev.Age == from {
		before = *prev
	}
	if fi >= 0 {
		a.oldTown = newMenuTown(&before, a.reg, a.set.Style)
	}
	a.newTown = newMenuTown(cur, a.reg, a.set.Style)
	a.sc, a.info.pages = nil, nil
}

// ---- its life ----

// frames is how long the celebration is, in frames.
func (a *arrival) frames() int { return arrivalFrames(a.view.epoch) }

// begin starts the screen's timers: the celebration moves on to the
// information by itself, and the screen closes by itself.
func (a *arrival) begin() {
	a.start = a.now()
	a.cancels = append(a.cancels,
		a.after(time.Duration(a.frames())*mapAnimStep, func() { a.inform() }),
		a.after(arrivalHold, a.close))
}

// inform moves on from the celebration to the information.
func (a *arrival) inform() {
	if a.stage != arrCelebrating {
		return
	}
	a.stage, a.page = arrInforming, 0
	if a.sc != nil {
		a.sc.settle()
	}
	a.rebase()
}

// close takes the screen down and gives the keyboard back to the prompt.
// It is safe to call twice (a key and a timer may both reach it).
func (a *arrival) close() {
	if a.stage == arrClosed {
		return
	}
	a.stage = arrClosed
	a.stopClock()
	for _, cancel := range a.cancels {
		cancel()
	}
	a.cancels = nil
	a.oldTown.close()
	a.newTown.close()
	if a.om == nil {
		return
	}
	// Only if the screen is still the manager's: when another overlay took
	// its place, or the manager is the one closing it, the page in front is
	// not this screen's to take down.
	if a.om.arrival == a {
		a.om.arrival = nil
		if a.om.active == arrivalPageName {
			a.om.Hide()
		}
	}
}

// rebase restarts the clock at the scene's present frame.
func (a *arrival) rebase() {
	a.start, a.base = a.now(), 0
	if a.sc != nil {
		a.base = a.sc.frame
	}
}

// startClock starts the redraw clock, at the map's rate. Draw calls it.
func (a *arrival) startClock() {
	if a.stop != nil || a.om == nil || a.om.app == nil {
		return
	}
	stop := make(chan struct{})
	a.stop = stop
	app := a.om.app
	go func() {
		tk := time.NewTicker(mapAnimStep)
		defer tk.Stop()
		for {
			select {
			case <-tk.C:
				app.QueueUpdateDraw(func() {})
			case <-stop:
				return
			}
		}
	}()
}

func (a *arrival) stopClock() {
	if a.stop != nil {
		close(a.stop)
		a.stop = nil
	}
}

// InputHandler: any key moves on. From the celebration it goes to the
// information; from the information to its next page, or, on the last,
// it closes the screen.
func (a *arrival) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return a.WrapInputHandler(func(_ *tcell.EventKey, _ func(tview.Primitive)) { a.key() })
}

// key is a key press, whichever key it was.
func (a *arrival) key() {
	now := a.now()
	if !a.lastKey.IsZero() && now.Sub(a.lastKey) < arrivalKeyGap {
		return
	}
	a.lastKey = now
	switch a.stage {
	case arrCelebrating:
		a.inform()
	case arrInforming:
		if a.page+1 < len(a.info.pages) {
			a.page++
			return
		}
		a.close()
	}
}

// Draw draws the screen at the clock's frame.
func (a *arrival) Draw(scr tcell.Screen) {
	if a.stage == arrClosed {
		return
	}
	x, y, w, h := a.GetRect()
	if w <= 0 || h <= 0 {
		return
	}
	grid := a.frameGrid(w, h)
	grid.flush(scr, x, y, a.pal, a.view.plain)
	// An epoch's own weather: the theme's ambient effect, in the cells the
	// screen leaves empty, once the heavy blow has landed.
	if a.view.motion && a.view.epoch && a.stage == arrCelebrating && a.sc.frame >= arrEraStrike {
		if effect := theme.Active().Effect; effect != "" {
			drawThemeEffect(scr, x, y, w, h, effect, a.sc.frame+arrivalFxLead, a.view.plain)
		}
	}
	if a.view.motion {
		a.startClock()
	} else {
		a.stopClock()
	}
}

// frameGrid brings the scene to the clock's frame and draws the screen at
// w by h.
func (a *arrival) frameGrid(w, h int) *mGrid {
	if th := theme.Active(); a.pal == nil || a.pal.key != th.Key {
		a.pal = newMenuPalette(th)
		a.info.pages = nil
	}
	v := &a.view
	epoch := 0
	if v.epoch {
		epoch = 1
	}
	if key := [4]int{w, h, epoch, len(v.underLines(w - 6))}; a.sc == nil || key != a.scKey {
		// A new size: the scene starts over and is played up to where the
		// old one was, so a resize does not replay the celebration.
		at := 0
		if a.sc != nil {
			at = a.sc.frame
		}
		a.sc, a.scKey = newArrivalScene(arrivalLayoutFor(w, h, v), v.epoch), key
		a.info.pages = nil
		switch {
		case a.stage == arrInforming:
			a.sc.settle()
		case !v.motion:
			a.sc.rest()
		default:
			for a.sc.frame < at {
				a.sc.step()
			}
		}
		a.rebase()
	}
	since := time.Duration(0)
	if v.motion {
		since = a.now().Sub(a.start)
		target := a.base + int(since/mapAnimStep)
		if a.stage == arrCelebrating {
			// Every frame of the script is played, however late the draw.
			for a.sc.frame < min(target, a.frames()) {
				a.sc.step()
			}
			if a.sc.frame >= a.frames() {
				a.inform()
			}
		} else {
			for i := 0; a.sc.frame < target && i < menuCatchUp; i++ {
				a.sc.step()
			}
			a.sc.frame = max(a.sc.frame, target)
		}
	}
	mf := mapFrame(nil, a.set, since)
	if a.stage == arrCelebrating {
		return a.sc.renderMoment(a.pal, v, a.oldTown, a.newTown, mf)
	}
	if key := [3]int{w, h, len(v.lines)}; a.info.pages == nil || key != a.infoKey {
		a.info, a.infoKey = arrivalInfoFor(a.sc.L, v, a.pal), key
		a.page = min(a.page, max(len(a.info.pages)-1, 0))
	}
	return a.sc.renderInfo(a.pal, v, a.info, a.page, a.newTown, mf)
}
