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
// The root Flex is the whole screen: the panel's frame (panel_frame.go) over the
// command bar's rows.
type overlayEntry struct {
	title   vibe.Entry
	tv      *tview.TextView
	frame   *panelFrame     // the box, name and keys the TextView is drawn in
	root    tview.Primitive // the frame over the command bar — added/removed from pages.Pages
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
	// footer, when set, is the keys on the panel's foot in place of the
	// scrolling ones (SetFooter).
	footer func() [][2]string
}

// widgetEntry backs an overlay whose content is any tview.Primitive (e.g. a custom
// map canvas). When refreshFn is set, Refresh calls it in-place instead of
// rebuilding and replacing the page (used for panels that manage their own state).
type widgetEntry struct {
	title     vibe.Entry
	build     func(state game.GameState) tview.Primitive
	refreshFn func(state game.GameState) // optional; nil means full rebuild on each Refresh
}

// OverlayManager manages the set of named panels shown over the Dashboard, each the
// whole screen but the command bar. Only one overlay is visible at a time. Panels are added/removed from d.pages.Pages
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
	// plain, when set, reports the ASCII glyph tier, which the text panels'
	// frames fold to.
	plain func() bool

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
// The overlay is the whole screen but the command bar: a framed panel (panel_frame.go)
// with the TextView inside it, over the rows the bar keeps.
func (om *OverlayManager) Register(name string, title vibe.Entry, provide OverlayProvider) {
	tv := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(true)
	// Persistent overlay chrome (built once, reused on every Show): enroll so a live
	// theme switch restyles the background.
	theme.Track(func() {
		tv.SetBackgroundColor(theme.Color(theme.RoleBackground))
	})
	frame := newPanelFrame(title, tv)
	frame.plain = func() bool { return om.plain != nil && om.plain() }
	entry := &overlayEntry{title: title, tv: tv, frame: frame, provide: provide}
	frame.keys = func() [][2]string { return om.footerKeys(entry) }
	frame.hint = func() string { return om.typeHint(entry) }
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

	// The frame is the whole screen but the command bar's rows, which stay
	// the dashboard's.
	entry.root = overBar(frame)
	om.entries[name] = entry
}

// keepsBar says whether the named text overlay leaves the keyboard with the
// command bar, as the People and Expeditions panels do: it takes the keys
// that print nothing (scrolling, see scrollKey) and the bar gets the rest. A
// panel whose own keys are letters (SetKeyHandler) keeps the keyboard.
func (om *OverlayManager) keepsBar(name string) bool {
	e, ok := om.entries[name]
	return ok && e.keys == nil
}

// typeHint is what the foot of a panel says may be typed: for a panel that
// leaves the keyboard with the command bar, a command; for one that keeps it,
// nothing.
func (om *OverlayManager) typeHint(e *overlayEntry) string {
	if e.keys != nil {
		return ""
	}
	return vibe.Line("ui.overlay.typeHint.01", "a command")
}

// footerKeys are the keys on a text panel's foot: what the panel set for
// itself, or the keys that scroll it.
func (om *OverlayManager) footerKeys(e *overlayEntry) [][2]string {
	if e.footer != nil {
		return e.footer()
	}
	return [][2]string{{"↑↓", "scroll"}, {"PgUp PgDn", vibe.Line("ui.overlay.footerKeys.03", "page")}, {"Esc", "close"}}
}

// scrollKey gives a key that prints nothing to the open text panel when it
// leaves the keyboard with the command bar and the bar is empty, and reports
// whether it took it. The caller is the dashboard's input capture, which sees
// the keys the bar has before the bar does.
func (om *OverlayManager) scrollKey(ev *tcell.EventKey, prompt string) bool {
	e, ok := om.entries[om.active]
	if !ok || e.keys != nil || prompt != "" {
		return false
	}
	switch ev.Key() {
	case tcell.KeyUp, tcell.KeyDown, tcell.KeyPgUp, tcell.KeyPgDn, tcell.KeyHome, tcell.KeyEnd:
		if h := e.tv.InputHandler(); h != nil {
			h(ev, func(tview.Primitive) {})
		}
		return true
	}
	return false
}

// SetFooter gives the named text overlay the keys on its foot, in place of the
// scrolling ones. Unknown names are ignored.
func (om *OverlayManager) SetFooter(name string, keys func() [][2]string) {
	if e, ok := om.entries[name]; ok {
		e.footer = keys
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

// textSize is the size of the text area of the named text overlay: what it was
// drawn at, or before its first draw what the screen makes it.
func (om *OverlayManager) textSize(name string) (w, h int) {
	if e, ok := om.entries[name]; ok && e.frame.iw > 0 && e.frame.ih > 0 {
		return e.frame.iw, e.frame.ih
	}
	return panelTextSize(max(om.screenW, 24), max(om.screenH-promptRows, panelMinRows+2))
}

// RegisterWidget registers an overlay backed by any tview.Primitive instead of a TextView.
// The build function receives the current GameState and returns a ready-to-display primitive,
// which fills the entire screen (the panel leaves the command bar's rows to the dashboard).
func (om *OverlayManager) RegisterWidget(name string, title vibe.Entry, build func(state game.GameState) tview.Primitive, refresh func(state game.GameState)) {
	om.widgetEntries[name] = &widgetEntry{
		title:     title,
		build:     build,
		refreshFn: refresh,
	}
}

// buildWidgetRoot creates the display tree for a widget overlay given a freshly-built primitive.
func (om *OverlayManager) buildWidgetRoot(_ *widgetEntry, prim tview.Primitive) tview.Primitive {
	// Wrap in a Flex so we can set ESC capture without touching the inner primitive's handler.
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
