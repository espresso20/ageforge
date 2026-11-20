package ui

import (
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/window/kit"
)

// menu_badges.go opens the Ranks screen from the main menu. It is the same
// panel the game opens (ranks_panel.go): it draws from the account's badges,
// which need no game loaded, so here it has the whole screen and the keyboard
// to itself. Esc steps back to the Ranks tab, then closes.

// menuBadgesPage is the page the screen is shown on from the menu.
const menuBadgesPage = "menu_badges"

// menuBadges is the Ranks screen as a page of its own.
type menuBadges struct {
	*ranksPanel
	close func()
}

// Key: the screen's own keys, and Esc (or q) to close.
func (b *menuBadges) Key(k kit.Key) bool {
	switch {
	case k.Code == kit.KeyEsc, k.Code == kit.KeyRune && (k.Rune == 'q' || k.Rune == 'Q'):
		if !b.closeBack() {
			b.close()
		}
	default:
		b.routeKey(&k, "")
	}
	return true
}

// openMenuBadges shows the Ranks screen over the menu and returns to
// returnPage when it is closed.
func openMenuBadges(app *appLoop, pages *nativePages, engine *game.GameEngine, returnPage string) {
	p := newRanksPanel()
	// The screen reads a snapshot's account part and nothing else of it.
	st := game.GameState{}
	if engine.Account() != nil {
		views, sum := engine.Badges()
		st.AccountStats = &game.AccountStatsView{Badges: views, BadgeSummary: sum}
	}
	p.update(st)

	b := &menuBadges{ranksPanel: p}
	b.close = func() {
		pages.RemovePage(menuBadgesPage)
		pages.SwitchToPage(returnPage)
	}
	pages.AddPage(menuBadgesPage, b, true, true)
	app.SetFocus(b)
}
