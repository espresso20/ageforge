package ui

import (
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
)

// App manages the tview application and page routing
type App struct {
	tviewApp  *tview.Application
	pages     *tview.Pages
	engine    *game.GameEngine
	dashboard *Dashboard
	version   string
}

// NewApp creates the UI application
func NewApp(engine *game.GameEngine, version string) *App {
	// Assert the active theme before any widget is built so the name-remap and
	// tview.Styles chrome defaults are in place for the first Draw, and every
	// theme.Track closure in setup() applies against a known palette. The theme
	// package init() already seeds Forge defensively; this makes the boot explicit
	// and survives a future import reshuffle. Phase 2 swaps DefaultKey for the
	// account's stored active theme (theming.md §6). Ignoring the error is fine —
	// DefaultKey is a registered built-in.
	_ = theme.SetActive(theme.DefaultKey)

	a := &App{
		tviewApp: tview.NewApplication(),
		pages:    tview.NewPages(),
		engine:   engine,
		version:  version,
	}
	a.setup()
	return a
}

func (a *App) setup() {
	// Apply the account's persisted active theme (theming.md §6) before any widget
	// is built, so the splash and every theme.Track closure in the page constructors
	// below render in the chosen theme on the very first Draw. applyAccountTheme is
	// fully defensive: no engine/account, an empty stored key, or an unknown/locked
	// one all fall back to the default (Forge). NewApp already pinned Forge as a
	// floor; this layers the account's choice on top once the engine is in place.
	applyAccountTheme(a.engine)

	a.dashboard = NewDashboard(a.tviewApp, a.engine, a.pages)

	splash := CreateSplashPage(a.tviewApp, a.pages, a.engine, a.version)

	a.pages.AddPage("splash", splash, true, true)
	a.pages.AddPage("dashboard", a.dashboard.Root(), true, false)

	a.tviewApp.SetRoot(a.pages, true)

	// First run: if no established (named) account is wired, prompt for the account
	// name as the very first thing the player sees — over the splash menu. The name
	// derives the identity (game.CreateNamedAccount); the modal's Esc-accepts policy
	// guarantees we always come away with an account. Once confirmed, install it and
	// reveal the splash beneath.
	if acct := a.engine.Account(); acct == nil || !acct.Established() {
		showAccountNameModal(a.tviewApp, a.pages, splash, func(name string) {
			newAcct, err := game.CreateNamedAccount(name)
			if err != nil {
				// Account is non-critical; on the rare Save failure, fall through to
				// the splash accountless rather than block play.
				return
			}
			a.engine.SetAccount(newAcct)
			// A brand-new account starts on its own theme (ActiveTheme "" → Forge)
			// rather than inheriting whatever was previewed before it was installed
			// (theming.md §6). Re-resolve from the freshly-wired account.
			applyAccountTheme(a.engine)
		})
	}
}

// Run starts the tview application (blocks until exit)
func (a *App) Run() error {
	a.dashboard.StartUpdates()
	defer a.dashboard.StopUpdates()
	return a.tviewApp.Run()
}

// Stop halts the tview application
func (a *App) Stop() {
	a.tviewApp.Stop()
}
