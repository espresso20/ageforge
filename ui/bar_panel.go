package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// bar_panel.go is what the overlay panels that leave the keyboard with the
// command bar have in common (Advance, Badges, Expeditions, Map, People and
// Research): the box they draw in, the two calls that reach the command bar,
// the InputHandler that tries the panel's own keys first, and the layout that
// keeps the bar's rows free under the panel.

// barPanel is embedded in each of those panels in place of *tview.Box, so its
// InputHandler is the one the panel has.
type barPanel struct {
	*tview.Box
	// prompt reads the command bar ("" when it is empty).
	prompt func() string
	// toPrompt hands a key the panel does not take to the command bar, and
	// the keyboard with it, when the panel itself has the focus.
	toPrompt func(ev *tcell.EventKey)
	// route is the panel's routeKey: it takes a key if it is one of the
	// panel's, given the command bar's text, and reports whether it did.
	route func(ev *tcell.EventKey, prompt string) bool
}

// newBarPanel is a barPanel whose keys are decided by route.
func newBarPanel(route func(ev *tcell.EventKey, prompt string) bool) barPanel {
	return barPanel{Box: tview.NewBox(), route: route}
}

// InputHandler is for when the panel itself has the focus (the command bar
// normally keeps it): the panel's own keys work, and anything else goes to
// the command bar, which takes the keyboard back.
func (b *barPanel) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return b.WrapInputHandler(func(ev *tcell.EventKey, _ func(tview.Primitive)) {
		prompt := ""
		if b.prompt != nil {
			prompt = b.prompt()
		}
		if !b.route(ev, prompt) && b.toPrompt != nil {
			b.toPrompt(ev)
		}
	})
}

// overBar is an overlay's build result: the panel over the whole screen but
// the command bar's rows, which it leaves to the dashboard underneath (a Flex
// does not clear the cells it leaves empty), so the prompt stays in sight,
// and in use, while the panel is open.
func overBar(p tview.Primitive) tview.Primitive {
	return tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(p, 0, 1, true).
		AddItem(nil, promptRows, 0, false)
}
