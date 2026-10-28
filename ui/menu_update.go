package ui

import (
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/vibe"
)

// menu_update.go is the main menu's Check for updates: the check, the
// offer and the install, each a modal over the menu.

// ── Update flow ───────────────────────────────────────────────────────────────

const updateModalPage = "update_modal"

func showUpdateCheck(app *tview.Application, pages *tview.Pages, currentVersion string) {
	modal := tview.NewModal().SetText(vibe.Line("ui.menu_update.showUpdateCheck.01", "Checking for updates..."))
	pages.AddPage(updateModalPage, modal, true, true)

	go func() {
		result, err := game.CheckLatest(currentVersion)
		app.QueueUpdateDraw(func() {
			pages.RemovePage(updateModalPage)
			if err != nil {
				showUpdateMsg(app, pages, vibe.Line("ui.menu_update.showUpdateCheck.02", "Update check failed:\n\n")+err.Error())
				return
			}
			if !result.IsNewer {
				showUpdateMsg(app, pages, vibe.Line("ui.menu_update.showUpdateCheck.03", "You're up to date.\n\n")+currentVersion+vibe.Line("ui.menu_update.showUpdateCheck.04", " is the latest version."))
				return
			}
			showUpdateConfirm(app, pages, result)
		})
	}()
}

func showUpdateMsg(app *tview.Application, pages *tview.Pages, msg string) {
	showButtonModal(pages, updateModalPage, msg, []string{txtOK.String()}, func(int, string) {})
}

func showUpdateConfirm(app *tview.Application, pages *tview.Pages, result game.UpdateResult) {
	msg := vibe.Sprintf(
		"ui.menu_update.showUpdateConfirm.01", "  ✦  Update available  ✦\n\n  Latest:   %s\n  Current:  %s\n\nDownload and install now?",
		result.LatestVersion, result.CurrentVersion,
	)
	buttons := []string{vibe.Line("ui.menu_update.showUpdateConfirm.02", "Update now"), vibe.Line("ui.menu_update.showUpdateConfirm.03", "Later")}
	showButtonModal(pages, updateModalPage, msg, buttons, func(button int, _ string) {
		if button == 0 { // "Update now"; Later and Esc (-1) do not
			showUpdateInstall(app, pages, result)
		}
	})
}

func showUpdateInstall(app *tview.Application, pages *tview.Pages, result game.UpdateResult) {
	modal := tview.NewModal().
		SetText(vibe.Sprintf("ui.menu_update.showUpdateInstall.01", "Downloading %s...\n\n%s", result.LatestVersion, result.BinaryName))
	pages.AddPage(updateModalPage, modal, true, true)

	go func() {
		msg, err := game.DownloadAndInstall(result)
		app.QueueUpdateDraw(func() {
			pages.RemovePage(updateModalPage)
			if err != nil {
				showUpdateMsg(app, pages, vibe.Line("ui.menu_update.showUpdateInstall.02", "Update failed:\n\n")+err.Error())
				return
			}
			showUpdateMsg(app, pages, "  ✓  "+msg)
		})
	}()
}
