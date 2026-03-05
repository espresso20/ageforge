package ui

import (
	"fmt"
	"math/rand"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/game"
)

// eliteMessages are shown on the splash to players who have achieved the elite badge.
// Each message uses tview color tags: gold outer decorators, cyan inner text.
var eliteMessages = []string{
	"[gold]⚡[-][cyan] MASTER FORGER [-][gold]⚡[-]",
	"[gold]{[-][cyan] TOUCHED BY THE SOURCE [-][gold]}[-]",
	"[gold]✦[-][cyan] ARCHITECT OF THE FORGE [-][gold]✦[-]",
	"[gold][[-][cyan] REALITY.EXE PATCHED [-][gold]][-]",
	"[gold]⚙[-][cyan] THE FIRST MAKER [-][gold]⚙[-]",
}

// CreateSplashPage creates the main menu splash screen.
// wikiServer is started and opened in the browser when the player selects Wiki.
func CreateSplashPage(app *tview.Application, pages *tview.Pages, engine *game.GameEngine, wikiServer *game.WikiServer, currentVersion string) tview.Primitive {
	saveExists := game.SaveExists("autosave")
	_, eliteBadge := game.PeekSaveBadges("autosave")
	prestigeLevel := engine.Prestige.GetLevel()

	// Animated starfield + title canvas (takes the upper portion of the screen)
	canvas := newSplashCanvas(prestigeLevel)
	canvas.animate(app)

	// nav wraps any navigation action: halt animation then run it.
	nav := func(fn func()) {
		canvas.halt()
		fn()
	}

	// ── Single action list ───────────────────────────────────────────────────
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
			engine.Reset()
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
	mainList.AddItem("  ✗  Quit", "", 'q', func() {
		canvas.halt()
		app.Stop()
	})
	mainList.AddItem("  ✗  Wipe Save", "", 'x', func() {
		showWipeConfirmation(app, pages, engine, wikiServer, canvas.halt, currentVersion)
	})
	mainList.AddItem("  ↑  Check for Update", "", 'u', func() {
		showUpdateCheck(app, pages, currentVersion)
	})

	// Default selection
	if !saveExists {
		mainList.SetCurrentItem(1)
	}

	// Version row — always shows current version; update badge appears here async
	versionTV := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText(fmt.Sprintf("[gold]%s[-]", currentVersion))

	// Background update check — no-op on dev builds or network errors
	if currentVersion != "dev" {
		go func() {
			result, err := game.CheckLatest(currentVersion)
			if err != nil || !result.IsNewer {
				return
			}
			app.QueueUpdateDraw(func() {
				versionTV.SetText(fmt.Sprintf(
					"[gold]%s  ✦ new update available! (u)[-]",
					currentVersion,
				))
			})
		}()
	}

	// Footer hint
	footerTV := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[#8b949e]Arrow keys · Enter[-]")

	// ── Menu panel ────────────────────────────────────────────────────────────
	menuPanel := tview.NewFlex().SetDirection(tview.FlexRow)
	if eliteBadge {
		msg := eliteMessages[rand.Intn(len(eliteMessages))]
		eliteBadgeTV := tview.NewTextView().
			SetDynamicColors(true).
			SetTextAlign(tview.AlignCenter).
			SetText(msg)
		menuPanel.AddItem(eliteBadgeTV, 1, 0, false)
	}
	menuPanel.
		AddItem(mainList, 8, 0, true).
		AddItem(versionTV, 1, 0, false).
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
	// Elite badge adds 1 extra line to the menu panel height.
	menuHeight := 12
	if eliteBadge {
		menuHeight = 13
	}
	outer := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(canvas, 0, 1, false).
		AddItem(menuRow, menuHeight, 0, true)

	return outer
}

// showWipeConfirmation shows the "are you sure?" modal before wiping data.
// cleanup is called (to halt the canvas animation) before recreating the splash.
func showWipeConfirmation(app *tview.Application, pages *tview.Pages, engine *game.GameEngine, wikiServer *game.WikiServer, cleanup func(), currentVersion string) {
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
				newSplash := CreateSplashPage(app, pages, engine, wikiServer, currentVersion)
				pages.AddPage("splash", newSplash, true, true)
			}
		})
	modal.SetBackgroundColor(tcell.ColorDarkRed)
	pages.AddPage("wipe_confirm", modal, true, true)
}

// ── Update flow ───────────────────────────────────────────────────────────────

const updateModalPage = "update_modal"

func showUpdateCheck(app *tview.Application, pages *tview.Pages, currentVersion string) {
	modal := tview.NewModal().SetText("Checking for updates...")
	pages.AddPage(updateModalPage, modal, true, true)

	go func() {
		result, err := game.CheckLatest(currentVersion)
		app.QueueUpdateDraw(func() {
			pages.RemovePage(updateModalPage)
			if err != nil {
				showUpdateMsg(app, pages, "Update check failed:\n\n"+err.Error())
				return
			}
			if !result.IsNewer {
				showUpdateMsg(app, pages, "You're up to date!\n\n"+currentVersion+" is the latest version.")
				return
			}
			showUpdateConfirm(app, pages, result)
		})
	}()
}

func showUpdateMsg(app *tview.Application, pages *tview.Pages, msg string) {
	modal := tview.NewModal().
		SetText(msg).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(_ int, _ string) {
			pages.RemovePage(updateModalPage)
		})
	pages.AddPage(updateModalPage, modal, true, true)
}

func showUpdateConfirm(app *tview.Application, pages *tview.Pages, result game.UpdateResult) {
	msg := fmt.Sprintf(
		"  ✦  Update Available  ✦\n\n  Latest:   %s\n  Current:  %s\n\nDownload and install now?",
		result.LatestVersion, result.CurrentVersion,
	)
	modal := tview.NewModal().
		SetText(msg).
		AddButtons([]string{"Update Now", "Later"}).
		SetDoneFunc(func(_ int, label string) {
			pages.RemovePage(updateModalPage)
			if label == "Update Now" {
				showUpdateInstall(app, pages, result)
			}
		})
	pages.AddPage(updateModalPage, modal, true, true)
}

func showUpdateInstall(app *tview.Application, pages *tview.Pages, result game.UpdateResult) {
	modal := tview.NewModal().
		SetText(fmt.Sprintf("Downloading %s...\n\n%s", result.LatestVersion, result.BinaryName))
	pages.AddPage(updateModalPage, modal, true, true)

	go func() {
		msg, err := game.DownloadAndInstall(result)
		app.QueueUpdateDraw(func() {
			pages.RemovePage(updateModalPage)
			if err != nil {
				showUpdateMsg(app, pages, "Update failed:\n\n"+err.Error())
				return
			}
			showUpdateMsg(app, pages, "  ✓  "+msg)
		})
	}()
}
