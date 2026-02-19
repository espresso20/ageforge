package ui

import (
	"github.com/rivo/tview"

	"github.com/user/ageforge/game"
)

// App manages the tview application and page routing
type App struct {
	tviewApp  *tview.Application
	pages     *tview.Pages
	engine    *game.GameEngine
	dashboard *Dashboard
}

// NewApp creates the UI application
func NewApp(engine *game.GameEngine) *App {
	a := &App{
		tviewApp: tview.NewApplication(),
		pages:    tview.NewPages(),
		engine:   engine,
	}
	a.setup()
	return a
}

func (a *App) setup() {
	// Create pages
	splash := CreateSplashPage(a.tviewApp, a.pages, a.engine)
	a.dashboard = NewDashboard(a.tviewApp, a.engine, a.pages)

	a.pages.AddPage("splash", splash, true, true)
	a.pages.AddPage("dashboard", a.dashboard.Root(), true, false)

	a.tviewApp.SetRoot(a.pages, true)
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
