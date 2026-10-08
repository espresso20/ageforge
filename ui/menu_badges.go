package ui

import (
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/ui/mapstyle"
)

// menu_badges.go opens the badge case from the main menu. The case is the
// same panel the game opens (badge_panel.go): it draws from the account's
// badges, which need no game loaded, so here it has the whole screen and
// the keyboard to itself. Esc closes an open badge, then the case.

// menuBadgesPage is the page the case is shown on from the menu.
const menuBadgesPage = "menu_badges"

// menuBadges is the badge case as a page of its own.
type menuBadges struct {
	*badgePanel
	close func()
}

// InputHandler: the case's own keys, and Esc (or q) to close.
func (b *menuBadges) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return b.WrapInputHandler(func(ev *tcell.EventKey, _ func(tview.Primitive)) {
		switch {
		case ev.Key() == tcell.KeyEsc, ev.Key() == tcell.KeyRune && (ev.Rune() == 'q' || ev.Rune() == 'Q'):
			if !b.closeCard() {
				b.close()
			}
		default:
			b.routeKey(ev, "")
		}
	})
}

// openMenuBadges shows the badge case over the menu and returns to
// returnPage when it is closed.
func openMenuBadges(app *tview.Application, pages *tview.Pages, engine *game.GameEngine, reg *mapstyle.Registry, returnPage string) {
	p := newBadgePanel()
	p.settings = func() mapSettings { return resolveMapSettings(engine.Account(), reg) }
	// The case reads a snapshot's account part and nothing else of it.
	st := game.GameState{}
	if engine.Account() != nil {
		views, sum := engine.Badges()
		st.AccountStats = &game.AccountStatsView{Badges: views, BadgeSummary: sum}
	}
	p.update(st)

	// A badge that moves is redrawn at the map's rate while the case is open.
	stop := make(chan struct{})
	b := &menuBadges{badgePanel: p}
	b.close = func() {
		select {
		case <-stop:
		default:
			close(stop)
		}
		pages.RemovePage(menuBadgesPage)
		pages.SwitchToPage(returnPage)
	}
	if p.motion {
		go func() {
			tk := time.NewTicker(mapAnimStep)
			defer tk.Stop()
			for {
				select {
				case <-tk.C:
					if p.moving.Load() {
						app.QueueUpdateDraw(func() {})
					}
				case <-stop:
					return
				}
			}
		}()
	}
	pages.AddPage(menuBadgesPage, b, true, true)
	app.SetFocus(b)
}
