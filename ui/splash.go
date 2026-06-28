package ui

import (
	"fmt"
	"math/rand"
	"strings"

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
func CreateSplashPage(app *tview.Application, pages *tview.Pages, engine *game.GameEngine, currentVersion string) tview.Primitive {
	saveExists := game.SaveExists(game.AutosaveName)
	_, eliteBadge := game.PeekSaveBadges(game.AutosaveName)
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

	// The Load Game browser lists every save (player-named + autosave) and handles
	// its own empty state, so the menu item is always enabled — gating it on the
	// autosave alone would hide the browser even when player-named saves exist.
	mainList.AddItem("  ⚔  Load Game", "", 'l', func() {
		// Build the page fresh each time so the save list is always current,
		// then add + focus it (mirrors the modal add/remove lifecycle). We do NOT
		// halt the canvas here: the browser is an opaque full-screen page, so the
		// animation runs harmlessly underneath and is still live when the player
		// returns via Back (no need to rebuild the splash).
		page := CreateLoadGamePage(app, pages, engine, "splash", true)
		pages.AddPage(loadGamePage, page, true, true)
		app.SetFocus(page)
	})
	mainList.AddItem("  ✦  New Game", "", 'n', func() {
		// Prompt for a civilization name first; only start the game on confirm.
		// Cancel returns to the splash with the canvas still animating.
		showNewGameNameModal(app, pages, mainList, func(name string) {
			nav(func() {
				engine.StartNewNamedGame(name)
				pages.SwitchToPage("dashboard")
				go engine.Start()
			})
		})
	})
	mainList.AddItem("  ✗  Quit", "", 'q', func() {
		canvas.halt()
		app.Stop()
	})
	mainList.AddItem("  ✗  Wipe Save", "", 'x', func() {
		showWipeConfirmation(app, pages, engine, canvas.halt, currentVersion)
	})
	mainList.AddItem("  ✗  Wipe Account", "", 'a', func() {
		showAccountWipeConfirmation(app, pages, engine, currentVersion)
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
		AddItem(mainList, 9, 0, true).
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
	menuHeight := 13
	if eliteBadge {
		menuHeight = 14
	}
	outer := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(canvas, 0, 1, false).
		AddItem(menuRow, menuHeight, 0, true)

	return outer
}

// showWipeConfirmation shows the "are you sure?" modal before wiping data.
// cleanup is called (to halt the canvas animation) before recreating the splash.
func showWipeConfirmation(app *tview.Application, pages *tview.Pages, engine *game.GameEngine, cleanup func(), currentVersion string) {
	modal := tview.NewModal().
		SetText("⚠  WIPE ALL DATA  ⚠\n\nThis will permanently delete ALL save files\nand reset the game to zero.\n\nPrestige, upgrades, progress — everything gone.\n\nAre you REALLY sure?").
		AddButtons([]string{"I'm Kidding!", "NUKE IT ALL"}).
		SetDoneFunc(func(_ int, buttonLabel string) {
			pages.RemovePage("wipe_confirm")
			if buttonLabel == "NUKE IT ALL" {
				cleanup() // stop old canvas goroutine
				game.WipeAllSaves()
				engine.Reset()
				// AddPage with an existing name replaces atomically — no RemovePage
				// needed, and avoids a flash of the dashboard underneath.
				newSplash := CreateSplashPage(app, pages, engine, currentVersion)
				pages.AddPage("splash", newSplash, true, true)
			}
		})
	modal.SetBackgroundColor(tcell.ColorDarkRed)
	pages.AddPage("wipe_confirm", modal, true, true)
}

// Page names for the two-step account-wipe flow. Distinct from "wipe_confirm"
// (the save wipe) so the two destructive flows never collide on the page stack.
const (
	accountWipeConfirmPage = "account_wipe_confirm"
	accountWipeTypePage    = "account_wipe_type"
)

// showAccountWipeConfirmation runs the two-step gate for permanently deleting the
// player's account (identity + theme unlocks + lifetime stats + achievements). It is
// the account analogue of showWipeConfirmation but with a far stronger second gate:
// a type-your-exact-name confirm, since the wipe is irreversible and re-derivable only
// by re-naming.
//
//   STEP 1 — a dark-red yes/no modal making the stakes plain (and reassuring the
//            player their game saves are untouched).
//   STEP 2 — a type-to-confirm gate: only an EXACT match of the account's DisplayName
//            proceeds. On match it calls game.WipeAccount(), clears the engine account,
//            and immediately re-runs the first-run name flow so the player starts over
//            with a fresh, freshly-named account.
//
// At the splash an established account always exists (app.setup's first-run prompt
// guarantees it), so engine.Account() is non-nil here.
func showAccountWipeConfirmation(app *tview.Application, pages *tview.Pages, engine *game.GameEngine, currentVersion string) {
	acct := engine.Account()
	if acct == nil {
		// Defensive: no account to wipe. Nothing to do.
		return
	}
	name := acct.Name()

	// STEP 1 — the are-you-sure gate.
	step1 := tview.NewModal().
		SetText(fmt.Sprintf(
			"⚠  WIPE ACCOUNT  ⚠\n\nThis permanently deletes your account \"%s\":\nall theme unlocks, lifetime stats, and achievements.\n\nThis CANNOT be undone. Your game saves are NOT affected.",
			name,
		)).
		AddButtons([]string{"Keep my account", "Wipe it"}).
		SetDoneFunc(func(_ int, label string) {
			pages.RemovePage(accountWipeConfirmPage)
			if label == "Wipe it" {
				showAccountWipeTypeGate(app, pages, engine, name, currentVersion)
			}
		})
	step1.SetBackgroundColor(tcell.ColorDarkRed)
	pages.AddPage(accountWipeConfirmPage, step1, true, true)
}

// showAccountWipeTypeGate is STEP 2: the type-your-exact-name confirm. Only an exact
// match of expectedName proceeds; any other input shows an inline "name doesn't match"
// error and stays. Esc / Cancel aborts with no wipe. On an exact match it wipes, clears
// the engine account, removes the wipe pages, and re-runs the first-run name flow.
func showAccountWipeTypeGate(app *tview.Application, pages *tview.Pages, engine *game.GameEngine, expectedName, currentVersion string) {
	errTV := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)

	input := tview.NewInputField().
		SetLabel("Account name: ").
		SetFieldWidth(40).
		// Dark slate field so white text stays legible on the dark-red modal.
		SetFieldBackgroundColor(tcell.NewRGBColor(48, 54, 61)).
		SetFieldTextColor(tcell.ColorWhite)

	promptTV := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[white]Type your account name to confirm:[-]\n[#f0a0a0]This permanently deletes your account. It CANNOT be undone.[-]")

	hintTV := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[#f0a0a0]Enter: confirm  ·  Esc: cancel[-]")

	// abort removes the type gate and returns to the splash without wiping.
	abort := func() {
		pages.RemovePage(accountWipeTypePage)
	}

	// confirm proceeds only on an EXACT match. A non-match stays put with an error.
	confirm := func() {
		if strings.TrimSpace(input.GetText()) != expectedName {
			errTV.SetText("[red]Name doesn't match — account NOT wiped.[-]")
			app.SetFocus(input)
			return
		}
		// Exact match → wipe, detach, and re-run the first-run name flow so the
		// player starts over (mirrors app.setup's name modal → CreateNamedAccount →
		// SetAccount wiring). We rebuild the splash up front so there is a primitive
		// to restore focus to behind the name modal.
		_ = game.WipeAccount()
		engine.SetAccount(nil)
		pages.RemovePage(accountWipeTypePage)

		newSplash := CreateSplashPage(app, pages, engine, currentVersion)
		pages.AddPage("splash", newSplash, true, true)

		showAccountNameModal(app, pages, newSplash, func(name string) {
			newAcct, err := game.CreateNamedAccount(name)
			if err != nil {
				// Account is non-critical: on the rare Save failure fall through to
				// the splash accountless rather than block the player.
				return
			}
			engine.SetAccount(newAcct)
		})
	}

	input.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			confirm()
		}
	})

	// Opaque spacers so nothing bleeds through the modal interior.
	spacer := func() *tview.Box { return tview.NewBox().SetBackgroundColor(tcell.ColorDarkRed) }
	inner := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(promptTV, 2, 0, false).
		AddItem(spacer(), 1, 0, false).
		AddItem(input, 1, 0, true).
		AddItem(spacer(), 1, 0, false).
		AddItem(errTV, 1, 0, false).
		AddItem(spacer(), 1, 0, false).
		AddItem(hintTV, 1, 0, false)
	inner.SetBorder(true).
		SetTitle(" Confirm Account Wipe ").
		SetTitleColor(tcell.ColorWhite).
		SetBorderColor(tcell.ColorWhite)
	inner.SetBackgroundColor(tcell.ColorDarkRed)

	// Height = 7 content rows + 2 border = 9.
	modal := centeredModal(inner, 60, 9)
	modal.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		if ev.Key() == tcell.KeyEsc {
			abort()
			return nil
		}
		return ev
	})

	pages.AddPage(accountWipeTypePage, modal, true, true)
	app.SetFocus(input)
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
