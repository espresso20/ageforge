package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/game"
)

// CreateSplashPage creates the main menu splash screen.
// wikiServer is started and opened in the browser when the player selects Wiki.
func CreateSplashPage(app *tview.Application, pages *tview.Pages, engine *game.GameEngine, wikiServer *game.WikiServer) tview.Primitive {
	saveExists := game.SaveExists("autosave")
	prestigeLevel := engine.Prestige.GetLevel()

	// Animated starfield + title canvas (takes the upper portion of the screen)
	canvas := newSplashCanvas(prestigeLevel)
	canvas.animate(app)

	// nav wraps any navigation action: halt animation then run it.
	nav := func(fn func()) {
		canvas.halt()
		fn()
	}

	// ── Primary action list ───────────────────────────────────────────────────
	mainList := tview.NewList()
	mainList.SetBorder(false)
	mainList.SetSelectedBackgroundColor(tcell.ColorGold)
	mainList.SetSelectedTextColor(tcell.ColorBlack)
	mainList.ShowSecondaryText(false)

	loadLabel := "  ⚔  Load Game"
	if !saveExists {
		loadLabel = "  ⚔  Load Game  [no save]"
	}
	mainList.AddItem(loadLabel, "", 'l', func() {
		nav(func() {
			if saveExists {
				if err := engine.LoadGame("autosave"); err != nil {
					engine.AddLog("error", fmt.Sprintf("Load failed: %v", err))
				} else {
					engine.AddLog("success", "Game loaded!")
				}
			}
			pages.SwitchToPage("dashboard")
			go engine.Start()
		})
	})
	mainList.AddItem("  ✦  New Game", "", 'n', func() {
		nav(func() {
			pages.SwitchToPage("dashboard")
			go engine.Start()
		})
	})
	mainList.AddItem("  📖  Wiki", "", 'w', func() {
		if wikiServer != nil {
			if err := wikiServer.Start(); err == nil {
				wikiServer.OpenBrowser()
			}
		}
	})

	// Default selection
	if !saveExists {
		mainList.SetCurrentItem(1)
	}

	// Separator
	sepTV := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[gold]━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━[-]")

	// ── Danger action list ────────────────────────────────────────────────────
	dangerList := tview.NewList()
	dangerList.SetBorder(false)
	dangerList.SetSelectedBackgroundColor(tcell.ColorDarkRed)
	dangerList.SetSelectedTextColor(tcell.ColorWhite)
	dangerList.ShowSecondaryText(false)

	dangerList.AddItem("  ✗  Wipe Save", "", 'x', func() {
		showWipeConfirmation(app, pages, engine, wikiServer, canvas.halt)
	})
	dangerList.AddItem("  ✗  Quit", "", 'q', func() {
		canvas.halt()
		app.Stop()
	})

	// Footer hint
	footerTV := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[#8b949e]Arrow keys · Enter · Tab to switch list[-]")

	// ── Menu panel ────────────────────────────────────────────────────────────
	menuPanel := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(mainList, 5, 0, true).
		AddItem(sepTV, 1, 0, false).
		AddItem(dangerList, 4, 0, false).
		AddItem(footerTV, 1, 0, false)
	menuPanel.
		SetBorder(true).
		SetBorderColor(tcell.ColorGold).
		SetTitle(" AgeForge ").
		SetTitleColor(tcell.ColorGold)

	// Centre the menu horizontally (fixed width)
	menuRow := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(nil, 0, 1, false).
		AddItem(menuPanel, 50, 0, true).
		AddItem(nil, 0, 1, false)

	// ── Outer layout ─────────────────────────────────────────────────────────
	// Canvas fills top space; compact menu anchored at bottom.
	outer := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(canvas, 0, 1, false).
		AddItem(menuRow, 13, 0, true)

	// Tab cycles between the two lists
	outer.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			if mainList.HasFocus() {
				app.SetFocus(dangerList)
			} else {
				app.SetFocus(mainList)
			}
			return nil
		case tcell.KeyBacktab:
			if dangerList.HasFocus() {
				app.SetFocus(mainList)
			} else {
				app.SetFocus(dangerList)
			}
			return nil
		}
		return event
	})

	return outer
}

// showWipeConfirmation shows the "are you sure?" modal before wiping data.
// cleanup is called (to halt the canvas animation) before recreating the splash.
func showWipeConfirmation(app *tview.Application, pages *tview.Pages, engine *game.GameEngine, wikiServer *game.WikiServer, cleanup func()) {
	modal := tview.NewModal().
		SetText("⚠  WIPE ALL DATA  ⚠\n\nThis will permanently delete ALL save files\nand reset the game to zero.\n\nPrestige, upgrades, progress — everything gone.\n\nAre you REALLY sure?").
		AddButtons([]string{"I'm Kidding!", "NUKE IT ALL"}).
		SetDoneFunc(func(_ int, buttonLabel string) {
			pages.RemovePage("wipe_confirm")
			if buttonLabel == "NUKE IT ALL" {
				cleanup() // stop old canvas goroutine
				game.WipeAllSaves()
				engine.Reset()
				pages.RemovePage("splash")
				newSplash := CreateSplashPage(app, pages, engine, wikiServer)
				pages.AddPage("splash", newSplash, true, true)
			}
		})
	modal.SetBackgroundColor(tcell.ColorDarkRed)
	pages.AddPage("wipe_confirm", modal, true, true)
}
