package ui

import (
	"image"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/window/kit"
)

// bar_panel.go is what the overlay panels that leave the keyboard with the
// command bar have in common (Advance, Badges, Expeditions, Map, People and
// Research): the box they draw in, the two calls that reach the command bar,
// the InputHandler that tries the panel's own keys first, and the layout that
// keeps the bar's rows free under the panel.

// barPanel is embedded in each of those panels in place of a bare box, so its
// InputHandler is the one the panel has.
type barPanel struct {
	*primBase
	// prompt reads the command bar ("" when it is empty).
	prompt func() string
	// toPrompt hands a key the panel does not take to the command bar, and
	// the keyboard with it, when the panel itself has the focus.
	toPrompt func(ev *kit.Key)
	// route is the panel's routeKey: it takes a key if it is one of the
	// panel's, given the command bar's text, and reports whether it did.
	route func(ev *kit.Key, prompt string) bool
	// buttons makes the keys the panel draws pressable (key_bar.go).
	buttons keyPresser
}

// newBarPanel is a barPanel whose keys are decided by route.
func newBarPanel(route func(ev *kit.Key, prompt string) bool) barPanel {
	return barPanel{primBase: newPrimBase(), route: route}
}

// InputHandler is for when the panel itself has the focus (the command bar
// normally keeps it): the panel's own keys work, and anything else goes to
// the command bar, which takes the keyboard back.
func (b *barPanel) Key(k kit.Key) bool {
	prompt := ""
	if b.prompt != nil {
		prompt = b.prompt()
	}
	if !b.route(&k, prompt) && b.toPrompt != nil {
		b.toPrompt(&k)
	}
	return true
}

// paintGrid puts the grid a panel drew on the kit's painter at r in theme
// colours, with the key under the pointer lit.
func (b *barPanel) paintGrid(p *kit.Painter, g *xgrid, r image.Rectangle) {
	if !r.Empty() && g != nil {
		pal := newExpPalette()
		paintExpeditionKit(p, g, r.Min, pal)
		b.buttons.paint(p, g, r.Min, pal)
	}
}

// overBar is an overlay's build result: the panel over the whole screen but
// the command bar's rows, which it leaves to the dashboard underneath (a Flex
// does not clear the cells it leaves empty), so the prompt stays in sight,
// and in use, while the panel is open.
func overBar(p Primitive) Primitive {
	return newNativeFlex().SetDirection(flexRow).
		AddItem(p, 0, 1, true).
		AddItem(nil, promptRows, 0, false)
}

// barOverlay is how the dashboard drives one of those panels while it is the
// open overlay.
type barOverlay struct {
	// routeKey is the panel's routeKey: it takes a key that is the panel's
	// and reports whether it did.
	routeKey func(ev *kit.Key, prompt string) bool
	// closeFirst takes back what Esc closes before it closes the panel (a
	// card, a question) and reports whether there was one. Nil when there
	// is nothing.
	closeFirst func() bool
	// reply, when set, puts what a command typed over the panel did on the
	// panel's key bar.
	reply func(cmd string, res CommandResult)
	// update shows what a command typed over the panel changed now rather
	// than at the next refresh.
	update func(state game.GameState)
}

// newBarOverlays is the panels that leave the keyboard with the command bar,
// by overlay name. Only the Map has a key bar of its own to say what a command
// typed over it did, because the log is behind it.
func (d *Dashboard) newBarOverlays() map[string]barOverlay {
	overlays := map[string]barOverlay{
		"map": {
			routeKey: d.mapPanel.routeKey,
			reply:    d.mapPanel.reply,
			update:   d.mapPanel.update,
		},
		"techs": {
			routeKey:   d.researchPanel.routeKey,
			closeFirst: d.researchPanel.closeCard,
			update:     d.researchPanel.update,
		},
		"badges": {
			routeKey: d.cabinet.routeKey,
			update:   d.cabinet.update,
		},
		"expedition": {
			routeKey:   d.expPanel.routeKey,
			closeFirst: d.expPanel.closeConfirm,
			update:     d.expPanel.update,
		},
		"stats": {
			routeKey: d.statsPanel.routeKey,
			update:   d.statsPanel.update,
		},
		"history": {
			routeKey: d.historyPanel.routeKey,
			update:   d.historyPanel.update,
		},
		"people": {
			routeKey: d.peoplePanel.routeKey,
			update:   d.peoplePanel.update,
		},
		"charter": {
			routeKey: d.advancePanel.routeKey,
			update:   d.advancePanel.update,
		},
	}
	for name, p := range d.decks {
		overlays[name] = barOverlay{routeKey: p.routeKey, reply: p.reply, update: p.update}
	}
	return overlays
}

// handToPrompt gives the keyboard back to the command bar, starting with ev:
// a panel that had it for itself passes on a key it does not take.
func (d *Dashboard) handToPrompt(ev *kit.Key) {
	d.overlayMgr.FocusOn(d.inputField)
	keyTo(d.inputField, *ev)
}
