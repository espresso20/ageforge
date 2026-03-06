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

// OverlayManager manages named floating panels shown on top of d.pages.
type OverlayManager struct {
	entries map[string]*overlayEntry
	active  string // name of currently visible overlay, "" if none
	pages   *tview.Pages
	app     *tview.Application
	onClose func() // called after hide (to restore focus)
}

func NewOverlayManager(pages *tview.Pages, app *tview.Application, onClose func()) *OverlayManager {
	return &OverlayManager{
		entries: make(map[string]*overlayEntry),
		pages:   pages,
		app:     app,
		onClose: onClose,
	}
}

// Register adds a named overlay. Must be called before Show.
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

// Show displays the named overlay with current state. Returns false if name unknown.
func (om *OverlayManager) Show(name string, state game.GameState) bool {
	e, ok := om.entries[name]
	if !ok {
		return false
	}
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
func (om *OverlayManager) Refresh(state game.GameState) {
	if om.active == "" {
		return
	}
	e, ok := om.entries[om.active]
	if !ok {
		return
	}
	e.tv.SetText(e.provide(state))
}

// HasActive returns true when an overlay is currently visible.
func (om *OverlayManager) HasActive() bool {
	return om.active != ""
}
