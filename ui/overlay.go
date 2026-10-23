package ui

import (
	"time"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/vibe"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// OverlayProvider is a function that generates the full text content for a
// floating panel from the current game state. Called every 500 ms by Refresh.
type OverlayProvider func(state game.GameState, w int) string

// overlayEntry holds the tview widgets and content provider for a text-based overlay.
// The root Flex uses nested nil spacers to centre the TextView at ~85%×85% of the terminal.
type overlayEntry struct {
	title   vibe.Entry
	tv      *tview.TextView
	root    tview.Primitive // centered flex wrapper — added/removed from pages.Pages
	provide OverlayProvider
	// keys, when set, sees every key but Esc while the overlay has focus and
	// returns nil to consume it (SetKeyHandler). Runs on the tview goroutine,
	// never under the engine lock.
	keys func(event *tcell.EventKey) *tcell.EventKey
	// back, when set, is offered Esc before it closes the overlay: it returns
	// true when it used the key (a second view going back to the first), so
	// the overlay stays up (SetBackHandler). open, when set, runs each time
	// the overlay is shown, before its text is made (SetOpenHandler).
	back func() bool
	open func()
}

// widgetEntry backs an overlay whose content is any tview.Primitive (e.g. a custom
// map canvas). When refreshFn is set, Refresh calls it in-place instead of
// rebuilding and replacing the page (used for panels that manage their own state).
type widgetEntry struct {
	title      vibe.Entry
	build      func(state game.GameState) tview.Primitive
	refreshFn  func(state game.GameState) // optional; nil means full rebuild on each Refresh
	fullScreen bool                       // true = fills whole terminal; false = 85×85% modal
}

// OverlayManager manages the set of named floating panels shown over the Dashboard.
// Only one overlay is visible at a time. Panels are added/removed from d.pages.Pages
// rather than toggled via ShowPage/HidePage so that the underlying dashboard stays live.
type OverlayManager struct {
	entries       map[string]*overlayEntry
	widgetEntries map[string]*widgetEntry
	active        string // name of currently visible overlay; "" if none
	// focus is what holds the keyboard while the active overlay shows (its
	// TextView, or a widget overlay's root); nil with none.
	focus   tview.Primitive
	pages   *tview.Pages
	app     *tview.Application
	onClose func() // called after hide to restore focus to the command input field
	screenW int    // updated every frame via SetBeforeDrawFunc
	screenH int

	// seen is the last snapshot Refresh was given: for the arrival screen
	// it is the game as it stood before the advance (arrival.go). arrival is
	// that screen while it is up. engine, when the app has given it one,
	// lets the screen read the game's state and the player's display
	// settings at once; without it the next Refresh brings the state.
	seen    *game.GameState
	arrival *arrival
	// film is the film of a run's ending while it plays (ending.go).
	film   *ending
	engine *game.GameEngine
	// after, when set, is how the arrival screen sets its timers (tests hold
	// them); nil means the clock.
	after func(d time.Duration, fn func()) (cancel func())
}

// NewOverlayManager creates an OverlayManager. onClose is called whenever an
// overlay is closed so the caller can return focus to the command input field.
func NewOverlayManager(pages *tview.Pages, app *tview.Application, onClose func()) *OverlayManager {
	om := &OverlayManager{
		entries:       make(map[string]*overlayEntry),
		widgetEntries: make(map[string]*widgetEntry),
		pages:         pages,
		app:           app,
		onClose:       onClose,
		screenW:       120, // sensible default before first draw
		screenH:       40,
	}
	app.SetBeforeDrawFunc(func(screen tcell.Screen) bool {
		om.screenW, om.screenH = screen.Size()
		return false
	})
	return om
}

// Register adds a named text overlay backed by a TextView. Must be called before Show.
// The overlay is rendered as a centred 85%×85% modal box using nested Flex nil-spacers
// (tview's idiomatic way to achieve percentage-based positioning in a fixed terminal).
// NOTE: The ratio 17/20 = 85% — if you change this, update both inner and root Flex items.
func (om *OverlayManager) Register(name string, title vibe.Entry, provide OverlayProvider) {
	tv := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(true)
	tv.SetBorder(true).
		SetTitle(overlayTitle(title))
	// Persistent overlay chrome (built once, reused on every Show): enroll so a live
	// theme switch restyles the border/title/background.
	theme.Track(func() {
		tv.SetBorderColor(theme.Color(theme.RoleAccent)).
			SetTitleColor(theme.Color(theme.RoleAccent)).
			SetBackgroundColor(theme.Color(theme.RoleSurface))
	})
	entry := &overlayEntry{title: title, tv: tv, provide: provide}
	tv.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			if entry.back != nil && entry.back() {
				tv.SetText(safeTags(entry.provide(*om.seenState(), om.screenW)))
				tv.ScrollToBeginning()
				return nil
			}
			om.Hide()
			return nil
		}
		if entry.keys != nil {
			return entry.keys(event)
		}
		return event
	})

	// Centered 85%×85% box via nested Flex spacers (nil items act as flexible spacers)
	inner := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(tv, 0, 17, true). // 17/20 = 85% of height
		AddItem(nil, 0, 1, false)

	root := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(inner, 0, 17, true). // 17/20 = 85% of width
		AddItem(nil, 0, 1, false)

	entry.root = root
	om.entries[name] = entry
}

// overlayTitle is the title on the border of an overlay: its name, and how to
// close it.
func overlayTitle(title vibe.Entry) string {
	return " " + title.String() + vibe.Line("ui.overlay.Register.01", " (Esc to close) ")
}

// RetitleAll writes the titles of the text overlays again, for when VibeMode
// has been switched. The widget overlays are built each time they open, so they
// read the title then.
func (om *OverlayManager) RetitleAll() {
	for _, e := range om.entries {
		e.tv.SetTitle(overlayTitle(e.title))
	}
}

// SetKeyHandler gives the named text overlay a key handler (see
// overlayEntry.keys). Unknown names are ignored.
func (om *OverlayManager) SetKeyHandler(name string, keys func(event *tcell.EventKey) *tcell.EventKey) {
	if e, ok := om.entries[name]; ok {
		e.keys = keys
	}
}

// SetBackHandler gives the named text overlay a say in Esc: back runs first
// and returns true when the overlay has gone back a step rather than closed
// (the text is made again from the last state the manager saw). Unknown names
// are ignored.
func (om *OverlayManager) SetBackHandler(name string, back func() bool) {
	if e, ok := om.entries[name]; ok {
		e.back = back
	}
}

// SetOpenHandler has the named text overlay run open each time it is shown,
// before its text is made. Unknown names are ignored.
func (om *OverlayManager) SetOpenHandler(name string, open func()) {
	if e, ok := om.entries[name]; ok {
		e.open = open
	}
}

// seenState is the last state Refresh was given (the zero state before any).
func (om *OverlayManager) seenState() *game.GameState {
	if om.seen == nil {
		return &game.GameState{}
	}
	return om.seen
}

// textSize is the inner size of the named text overlay's box: what it was
// drawn at, or before its first draw what the screen makes it (the box is
// 17/19 of the screen each way, less its border).
func (om *OverlayManager) textSize(name string) (w, h int) {
	if e, ok := om.entries[name]; ok {
		// (a box that has not been laid out yet is tview's default 15 by 10)
		if _, _, iw, ih := e.tv.GetInnerRect(); iw > 0 && ih > 0 && (iw != 13 || ih != 8) {
			return iw, ih
		}
	}
	return max(20, om.screenW*17/19-2), max(8, om.screenH*17/19-2)
}

// RegisterWidget registers an overlay backed by any tview.Primitive instead of a TextView.
// The build function receives the current GameState and returns a ready-to-display primitive.
// When fullScreen is false the primitive is wrapped in the same centered 85%×85% modal box
// (gold border, title, ESC to close). When fullScreen is true the primitive fills the entire
// screen — useful for the Map which needs every pixel.
func (om *OverlayManager) RegisterWidget(name string, title vibe.Entry, build func(state game.GameState) tview.Primitive, refresh func(state game.GameState), fullScreen bool) {
	om.widgetEntries[name] = &widgetEntry{
		title:      title,
		build:      build,
		refreshFn:  refresh,
		fullScreen: fullScreen,
	}
}

// buildWidgetRoot creates the display tree for a widget overlay given a freshly-built primitive.
func (om *OverlayManager) buildWidgetRoot(we *widgetEntry, prim tview.Primitive) tview.Primitive {
	if we.fullScreen {
		// Wrap in a Box so we can set ESC capture without touching the inner primitive's handler.
		wrapper := tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(prim, 0, 1, true)
		wrapper.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyEsc {
				om.Hide()
				return nil
			}
			return event
		})
		return wrapper
	}

	// Centered 85%×85% modal, same look as text overlays.
	// Use tview.Frame to add the gold border and title around any arbitrary primitive.
	frame := tview.NewFrame(prim).
		SetBorders(0, 0, 0, 0, 0, 0)
	// Rebuilt per Show, so construction-read theme.Color (no Track needed): a later
	// theme switch re-runs buildWidgetRoot on next open.
	frame.SetBorder(true).
		SetBorderColor(theme.Color(theme.RoleAccent)).
		SetTitle(" " + we.title.String() + vibe.Line("ui.overlay.buildWidgetRoot.01", " (Esc to close) ")).
		SetTitleColor(theme.Color(theme.RoleAccent)).
		SetBackgroundColor(theme.Color(theme.RoleSurface))
	frame.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			om.Hide()
			return nil
		}
		return event
	})

	inner := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(frame, 0, 17, true).
		AddItem(nil, 0, 1, false)

	root := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(inner, 0, 17, true).
		AddItem(nil, 0, 1, false)

	return root
}

// Show displays the named overlay with current state. Returns false if name unknown.
func (om *OverlayManager) Show(name string, state game.GameState) bool {
	// Try text overlays first.
	if e, ok := om.entries[name]; ok {
		if e.open != nil {
			e.open()
		}
		e.tv.SetText(safeTags(e.provide(state, om.screenW)))
		e.tv.ScrollToBeginning()
		if om.active != name {
			if om.active != "" {
				om.dropArrival()
				om.pages.RemovePage(om.active)
			}
			om.pages.AddPage(name, e.root, true, true)
			om.active = name
		}
		om.focus = e.tv
		om.app.SetFocus(e.tv)
		return true
	}

	// Try widget overlays.
	if we, ok := om.widgetEntries[name]; ok {
		prim := we.build(state)
		root := om.buildWidgetRoot(we, prim)
		// Always rebuild and replace — widget overlays don't cache their root.
		if om.active != "" {
			om.dropArrival()
			om.pages.RemovePage(om.active)
		}
		om.pages.AddPage(name, root, true, true)
		om.active = name
		om.focus = root
		om.app.SetFocus(root)
		return true
	}

	return false
}

// Hide closes the active overlay and calls onClose (to restore input focus).
func (om *OverlayManager) Hide() {
	if om.active == "" {
		return
	}
	om.dropArrival()
	om.pages.RemovePage(om.active)
	om.active = ""
	om.focus = nil
	if om.onClose != nil {
		om.onClose()
	}
}

// dropArrival lets go of the arrival screen, if it is up: its clock and its
// timers stop. Taking its page down is the caller's.
func (om *OverlayManager) dropArrival() {
	if a := om.arrival; a != nil {
		om.arrival = nil
		a.close()
	}
	if e := om.film; e != nil {
		om.film = nil
		e.close()
	}
}

// FullScreenUp reports whether the arrival screen or the film of a run's
// ending has the whole screen: the dashboard holds its own windows back
// until it is gone.
func (om *OverlayManager) FullScreenUp() bool {
	return om.active == arrivalPageName || om.active == endingPageName
}

// arrivalKey gives a key press to the arrival screen when it is the overlay
// in front, and reports whether it took it. It is how Esc reaches that
// screen as a key like any other (see the app's input capture): the
// dashboard would answer Esc with Hide before the screen saw it.
func (om *OverlayManager) arrivalKey() bool {
	switch {
	case om == nil:
	case om.arrival != nil && om.active == arrivalPageName:
		om.arrival.key()
		return true
	case om.film != nil && om.active == endingPageName:
		om.film.key()
		return true
	}
	return false
}

// FocusOn makes p hold the keyboard while the active overlay shows, in place
// of the overlay itself: the Map panel leaves it with the command bar, so a
// window closing over the map gives the prompt its keys back (Focus).
func (om *OverlayManager) FocusOn(p tview.Primitive) {
	if om.active == "" {
		return
	}
	om.focus = p
	om.app.SetFocus(p)
}

// Focus gives the keyboard back to the active overlay, when a window that
// opened over it closes. Reports whether an overlay was active.
func (om *OverlayManager) Focus() bool {
	if om.active == "" || om.focus == nil {
		return false
	}
	om.app.SetFocus(om.focus)
	return true
}

// Refresh updates the active overlay content with fresh state.
// For text overlays the content is updated in-place.
// For widget overlays the primitive is rebuilt and the page is replaced.
func (om *OverlayManager) Refresh(state game.GameState) {
	prev := om.seen
	om.seen = &state
	if om.arrival != nil {
		om.arrival.feed(prev, &state)
	}
	if om.film != nil {
		om.film.feed(prev, &state)
	}
	if om.active == "" {
		return
	}
	// Text overlay path.
	if e, ok := om.entries[om.active]; ok {
		e.tv.SetText(safeTags(e.provide(state, om.screenW)))
		return
	}
	// Widget overlay path — use in-place refresh if available, else rebuild.
	if we, ok := om.widgetEntries[om.active]; ok {
		if we.refreshFn != nil {
			we.refreshFn(state)
			return
		}
		prim := we.build(state)
		root := om.buildWidgetRoot(we, prim)
		om.pages.RemovePage(om.active)
		om.pages.AddPage(om.active, root, true, true)
		om.focus = root
		om.app.SetFocus(root)
	}
}

// HasActive returns true when an overlay is currently visible.
func (om *OverlayManager) HasActive() bool {
	return om.active != ""
}

// ActiveName returns the name of the currently visible overlay, or "" if none.
func (om *OverlayManager) ActiveName() string {
	return om.active
}
