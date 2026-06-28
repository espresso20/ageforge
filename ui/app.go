package ui

import (
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/game"
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
