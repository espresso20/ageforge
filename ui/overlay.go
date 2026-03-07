package ui

import (
	"github.com/espresso20/ageforge/game"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// OverlayProvider generates text content for a floating panel.
type OverlayProvider func(state game.GameState) string

type overlayEntry struct {
	title   string
	tv      *tview.TextView
	root    tview.Primitive // centered flex wrapper
	provide OverlayProvider
}

// widgetEntry backs an overlay whose content is an arbitrary tview.Primitive.
type widgetEntry struct {
	title      string
	build      func(state game.GameState) tview.Primitive
	refreshFn  func(state game.GameState) // optional: in-place update, skips page rebuild
	fullScreen bool
}

// OverlayManager manages named floating panels shown on top of d.pages.
type OverlayManager struct {
	entries       map[string]*overlayEntry
	widgetEntries map[string]*widgetEntry
	active        string // name of currently visible overlay, "" if none
	pages         *tview.Pages
	app           *tview.Application
	onClose       func() // called after hide (to restore focus)
}

func NewOverlayManager(pages *tview.Pages, app *tview.Application, onClose func()) *OverlayManager {
	return &OverlayManager{
		entries:       make(map[string]*overlayEntry),
		widgetEntries: make(map[string]*widgetEntry),
		pages:         pages,
		app:           app,
		onClose:       onClose,
	}
}

// Register adds a named text overlay. Must be called before Show.
func (om *OverlayManager) Register(name, title string, provide OverlayProvider) {
	tv := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(true)
	tv.SetBorder(true).
		SetBorderColor(tcell.ColorGold).
		SetTitle(" " + title + " — ESC to close ").
		SetTitleColor(tcell.ColorGold).
		SetBackgroundColor(tcell.ColorBlack)
	tv.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			om.Hide()
			return nil
		}
		return event
	})

	// Centered 85%×85% box via nested Flex spacers
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
	frame.SetBorder(true).
		SetBorderColor(tcell.ColorGold).
		SetTitle(" " + we.title + " — ESC to close ").
		SetTitleColor(tcell.ColorGold).
		SetBackgroundColor(tcell.ColorBlack)
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
		e.tv.SetText(e.provide(state))
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
		e.tv.SetText(e.provide(state))
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
