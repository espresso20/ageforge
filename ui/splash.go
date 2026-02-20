package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/user/ageforge/game"
)

// CreateSplashPage creates the main menu splash screen
func CreateSplashPage(app *tview.Application, pages *tview.Pages, engine *game.GameEngine) tview.Primitive {
	// ASCII art title with optional prestige level
	tagline := SplashTagline
	prestigeLevel := engine.Prestige.GetLevel()
	if prestigeLevel > 0 {
		tagline = fmt.Sprintf("%s\n[cyan]Prestige Level %d[-]", SplashTagline, prestigeLevel)
	}

	title := tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetDynamicColors(true).
		SetText(fmt.Sprintf("[gold]%s[-]\n\n[white]%s[-]", SplashArt, tagline))

	// Menu buttons
	newGameBtn := tview.NewButton("New Game").
		SetSelectedFunc(func() {
			pages.SwitchToPage("dashboard")
			go engine.Start()
		})
	newGameBtn.SetBackgroundColor(tcell.ColorDarkGreen)

	loadBtn := tview.NewButton("Load Game").
		SetSelectedFunc(func() {
			if err := engine.LoadGame("autosave"); err != nil {
				engine.AddLog("error", fmt.Sprintf("Load failed: %v", err))
			} else {
				engine.AddLog("success", "Game loaded!")
			}
			pages.SwitchToPage("dashboard")
			go engine.Start()
		})

	wipeBtn := tview.NewButton("Wipe Save").
		SetSelectedFunc(func() {
			showWipeConfirmation(app, pages, engine)
		})
	wipeBtn.SetBackgroundColor(tcell.ColorDarkRed)

	quitBtn := tview.NewButton("Quit").
		SetSelectedFunc(func() {
			app.Stop()
		})
	quitBtn.SetBackgroundColor(tcell.ColorDarkRed)

	// Button layout
	buttons := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(nil, 0, 1, false).
		AddItem(newGameBtn, 14, 0, true).
		AddItem(nil, 2, 0, false).
		AddItem(loadBtn, 14, 0, false).
		AddItem(nil, 2, 0, false).
		AddItem(wipeBtn, 14, 0, false).
		AddItem(nil, 2, 0, false).
		AddItem(quitBtn, 14, 0, false).
		AddItem(nil, 0, 1, false)

	// Assemble splash
	splash := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(title, 10, 0, false).
		AddItem(nil, 1, 0, false).
		AddItem(buttons, 3, 0, true).
		AddItem(nil, 0, 1, false)

	// Tab between buttons
	splash.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			if newGameBtn.HasFocus() {
				app.SetFocus(loadBtn)
			} else if loadBtn.HasFocus() {
				app.SetFocus(wipeBtn)
			} else if wipeBtn.HasFocus() {
				app.SetFocus(quitBtn)
			} else {
				app.SetFocus(newGameBtn)
			}
			return nil
		}
		return event
	})

	return splash
}

// showWipeConfirmation displays a modal warning before wiping all data
func showWipeConfirmation(app *tview.Application, pages *tview.Pages, engine *game.GameEngine) {
	modal := tview.NewModal().
		SetText("⚠ WIPE ALL DATA ⚠\n\nThis will permanently delete ALL save files\nand reset the game to zero.\n\nPrestige, upgrades, progress — everything gone.\n\nAre you sure?").
		AddButtons([]string{"Cancel", "WIPE EVERYTHING"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			pages.RemovePage("wipe_confirm")
			if buttonLabel == "WIPE EVERYTHING" {
				game.WipeAllSaves()
				engine.Reset()
				// Rebuild splash to reflect cleared state
				pages.RemovePage("splash")
				newSplash := CreateSplashPage(app, pages, engine)
				pages.AddPage("splash", newSplash, true, true)
			}
		})
	modal.SetBackgroundColor(tcell.ColorDarkRed)

	pages.AddPage("wipe_confirm", modal, true, true)
}
