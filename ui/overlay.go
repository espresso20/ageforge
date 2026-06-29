package ui

import (
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// OverlayProvider is a function that generates the full text content for a
// floating panel from the current game state. Called every 500 ms by Refresh.
type OverlayProvider func(state game.GameState, w int) string

// overlayEntry holds the tview widgets and content provider for a text-based overlay.
// The root Flex uses nested nil spacers to centre the TextView at ~85%×85% of the terminal.
type overlayEntry struct {
	title   string
	tv      *tview.TextView
	root    tview.Primitive // centered flex wrapper — added/removed from pages.Pages
	provide OverlayProvider
}

// widgetEntry backs an overlay whose content is any tview.Primitive (e.g. a custom
// map canvas). When refreshFn is set, Refresh calls it in-place instead of
// rebuilding and replacing the page (used for panels that manage their own state).
type widgetEntry struct {
	title      string
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
	pages         *tview.Pages
	app           *tview.Application
	onClose       func() // called after hide to restore focus to the command input field
	screenW       int    // updated every frame via SetBeforeDrawFunc
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
	}
	app.SetBeforeDrawFunc(func(screen tcell.Screen) bool {
		om.screenW, _ = screen.Size()
		return false
	})
	return om
}

// Register adds a named text overlay backed by a TextView. Must be called before Show.
// The overlay is rendered as a centred 85%×85% modal box using nested Flex nil-spacers
// (tview's idiomatic way to achieve percentage-based positioning in a fixed terminal).
// NOTE: The ratio 17/20 = 85% — if you change this, update both inner and root Flex items.
func (om *OverlayManager) Register(name, title string, provide OverlayProvider) {
	tv := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(true)
	tv.SetBorder(true).
		SetTitle(" " + title + " — ESC to close ")
	// Persistent overlay chrome (built once, reused on every Show): enroll so a live
	// theme switch restyles the border/title/background.
	theme.Track(func() {
		tv.SetBorderColor(theme.Color(theme.RoleAccent)).
			SetTitleColor(theme.Color(theme.RoleAccent)).
			SetBackgroundColor(theme.Color(theme.RoleBackground))
	})
	tv.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			om.Hide()
			return nil
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

	om.entries[name] = &overlayEntry{title: title, tv: tv, root: root, provide: provide}
}

// RegisterWidget registers an overlay backed by any tview.Primitive instead of a TextView.
// The build function receives the current GameState and returns a ready-to-display primitive.
// When fullScreen is false the primitive is wrapped in the same centered 85%×85% modal box
// (gold border, title, ESC to close). When fullScreen is true the primitive fills the entire
// screen — useful for the Map which needs every pixel.
func (om *OverlayManager) RegisterWidget(name, title string, build func(state game.GameState) tview.Primitive, refresh func(state game.GameState), fullScreen bool) {
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
		SetTitle(" " + we.title + " — ESC to close ").
		SetTitleColor(theme.Color(theme.RoleAccent)).
		SetBackgroundColor(theme.Color(theme.RoleBackground))
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
		e.tv.SetText(e.provide(state, om.screenW))
		e.tv.ScrollToBeginning()
		if om.active != name {
			if om.active != "" {
				om.pages.RemovePage(om.active)
			}
			om.pages.AddPage(name, e.root, true, true)
			om.active = name
		}
		om.app.SetFocus(e.tv)
		return true
	}

	// Try widget overlays.
	if we, ok := om.widgetEntries[name]; ok {
		prim := we.build(state)
		root := om.buildWidgetRoot(we, prim)
		// Always rebuild and replace — widget overlays don't cache their root.
		if om.active != "" {
			om.pages.RemovePage(om.active)
		}
		om.pages.AddPage(name, root, true, true)
		om.active = name
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
	om.pages.RemovePage(om.active)
	om.active = ""
	if om.onClose != nil {
		om.onClose()
	}
}

// Refresh updates the active overlay content with fresh state.
// For text overlays the content is updated in-place.
// For widget overlays the primitive is rebuilt and the page is replaced.
func (om *OverlayManager) Refresh(state game.GameState) {
	if om.active == "" {
		return
	}
	// Text overlay path.
	if e, ok := om.entries[om.active]; ok {
		e.tv.SetText(e.provide(state, om.screenW))
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
