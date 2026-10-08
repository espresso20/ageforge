package ui

import (
	"fmt"

	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/game"
)

// menu_update.go is the main menu's Check for updates: the check, the
// offer and the install, each a modal over the menu.

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
				showUpdateMsg(app, pages, "You're up to date.\n\n"+currentVersion+" is the latest version.")
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
		"  ✦  Update available  ✦\n\n  Latest:   %s\n  Current:  %s\n\nDownload and install now?",
		result.LatestVersion, result.CurrentVersion,
	)
	modal := tview.NewModal().
		SetText(msg).
		AddButtons([]string{"Update now", "Later"}).
		SetDoneFunc(func(_ int, label string) {
			pages.RemovePage(updateModalPage)
			if label == "Update now" {
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
