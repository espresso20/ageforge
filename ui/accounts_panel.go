package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/vibe"
)

// accountsPage is the page name under which the Accounts panel is registered on the root
// Pages. Added on demand (from the splash menu) and removed on Esc/Back so the account list
// is always rebuilt fresh.
const accountsPage = "accounts"

// Page names for the Accounts panel's transient modals. Each is distinct so two flows can
// never collide on the page stack, and distinct from the splash's account-wipe pages (those
// target the ACTIVE slot; ours target the SELECTED slot by id).
const (
	accountsImportPage   = "accounts_import"
	accountsWipeConfirm1 = "accounts_wipe_confirm"
	accountsSavesConfirm = "accounts_saves_confirm"
	accountsWipeTypeGate = "accounts_wipe_type"
	accountsMessagePage  = "accounts_message"
)

// accountsPanel holds the widgets and live state for the Accounts panel. Everything the input
// handlers and refresh logic touch hangs off this struct so the list can be rebuilt cheaply
// and a sensible selection restored after switch/new/import/wipe — the same shape as
// loadGameBrowser (load_game.go).
type accountsPanel struct {
	app            *tview.Application
	pages          *tview.Pages
	engine         *game.GameEngine
	currentVersion string

	root     *tview.Flex
	title    *tview.TextView
	subtitle *tview.TextView
	list     *tview.List
	detail   *tview.TextView
	status   *tview.TextView // transient confirmation line ("Now playing as …")

	summaries []game.AccountSummary // current rows, 1:1 with list items

	// returnPage is the page to return to on Esc/Back (the splash menu).
	returnPage string
}

// CreateAccountsPage builds the full-screen Accounts panel and returns its root primitive. The
// caller registers it (pages.AddPage(accountsPage, …)) and focuses the returned primitive. The
// screen is self-contained: all key handling lives on the list.
//
// returnPage names the page to return to on Esc (the splash menu). It is modeled on the
// load-game browser: a List on the left, a detail pane on the right, a footer action bar, and a
// live detail refresh wired through SetChangedFunc. Per the tview deadlock rule (Trello
// TCGiSWYX) NO handler here calls QueueUpdateDraw — tview redraws automatically after each
// input event, so SetText / re-filling the list / applyAccountTheme all paint on the next frame.
func CreateAccountsPage(app *tview.Application, pages *tview.Pages, engine *game.GameEngine, currentVersion string, returnPage string) tview.Primitive {
	return newAccountsPanel(app, pages, engine, currentVersion, returnPage).root
}

// newAccountsPanel builds the panel; its root is the page.
func newAccountsPanel(app *tview.Application, pages *tview.Pages, engine *game.GameEngine, currentVersion string, returnPage string) *accountsPanel {
	p := &accountsPanel{
		app:            app,
		pages:          pages,
		engine:         engine,
		currentVersion: currentVersion,
		returnPage:     returnPage,
	}

	// ── Title ──────────────────────────────────────────────────────────────────
	p.title = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)
	title := p.title

	p.subtitle = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)
	subtitle := p.subtitle

	// ── Account list ───────────────────────────────────────────────────────────
	p.list = tview.NewList()
	p.list.SetBorder(false)
	// Selection backs light text → Selection role (dark slate under Forge), not the gold
	// Accent — white-on-gold is unreadable (theme/themes_forge.go), same as load_game.
	theme.Track(func() {
		p.list.SetSelectedBackgroundColor(theme.Color(theme.RoleSelection)).
			SetSelectedTextColor(theme.Color(theme.RoleText))
	})
	p.list.ShowSecondaryText(false)
	p.list.SetChangedFunc(func(index int, _ string, _ string, _ rune) {
		// Live detail refresh. Runs on the tview MAIN goroutine, where tview redraws
		// automatically after the input event — so we MUST NOT QueueUpdateDraw here (it
		// would deadlock the app, Trello TCGiSWYX). We only SetText; the auto-redraw paints.
		p.updateDetail(index)
	})

	// ── Detail pane ────────────────────────────────────────────────────────────
	p.detail = tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(true)
	p.detail.SetBorder(true)
	theme.Track(func() {
		p.detail.SetBorderColor(theme.Color(theme.RoleAccent)).
			SetTitleColor(theme.Color(theme.RoleAccent))
	})

	// ── Body: list (left) + detail (right) ─────────────────────────────────────
	body := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(p.list, 0, 1, true).
		AddItem(p.detail, 0, 1, false)

	// ── Transient status line (switch confirmation) ────────────────────────────
	p.status = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)

	// ── Footer ─────────────────────────────────────────────────────────────────
	// Two rows: nine actions do not fit on one line under 126 columns, so on a
	// narrower terminal they are set over two (keyBar).
	footer := newFitView(func(w, h int) string { return keyBar(accountsKeys, w, h) })
	footer.SetTextAlign(tview.AlignCenter)
	footer.changed()

	// ── Layout ─────────────────────────────────────────────────────────────────
	p.root = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(title, 1, 0, false).
		AddItem(subtitle, 1, 0, false).
		AddItem(body, 0, 1, true). // weighted — list+detail take remaining space
		AddItem(p.status, 1, 0, false).
		AddItem(footer, 2, 0, false)

	p.list.SetInputCapture(p.handleKey)

	p.refresh(0)
	return p
}

// retext writes the panel's own words in the script now in force. Switching to
// an account can switch VibeMode, so refresh does it again.
func (p *accountsPanel) retext() {
	p.title.SetText(vibe.Line("ui.accounts_panel.newAccountsPanel.01", "[gold]═══ Accounts ═══[-]"))
	p.subtitle.SetText(vibe.Line("ui.accounts_panel.newAccountsPanel.02", "[gray]Each account is its own slot. Switch between them, or back one up to a file.[-]"))
	p.detail.SetTitle(vibe.Line("ui.accounts_panel.newAccountsPanel.03", " Details "))
}

// refresh re-reads the account listing, rebuilds the list rows, and restores a clamped
// selection. wantIdx is the index to select after the rebuild (clamped to range); pass the
// current index to keep position. Called after switch/new/import/wipe.
func (p *accountsPanel) refresh(wantIdx int) {
	p.retext()
	p.summaries = p.engine.ListAccounts()

	p.list.Clear()
	if len(p.summaries) == 0 {
		// Empty state — the action keys guard on len(p.summaries) and become no-ops.
		p.detail.SetText(vibe.Line("ui.accounts_panel.refresh.01", "[gray]No accounts found.\n\nCreate one with 'n', or restore a backup with 'i'.[-]"))
		return
	}

	for _, s := range p.summaries {
		p.list.AddItem(accountRowLabel(s), "", 0, nil)
	}

	if wantIdx < 0 {
		wantIdx = 0
	}
	if wantIdx >= len(p.summaries) {
		wantIdx = len(p.summaries) - 1
	}
	p.list.SetCurrentItem(wantIdx)
	// SetCurrentItem fires SetChangedFunc only when the index actually changes; force the
	// detail pane to reflect the (possibly unchanged) selection.
	p.updateDetail(wantIdx)
}

// updateDetail renders the detail pane for the account at index. Out-of-range selections render
// nothing harmful. The recovery code is fetched by id (read-only; never changes the active
// account); a lookup error degrades gracefully to a dash.
func (p *accountsPanel) updateDetail(index int) {
	if index < 0 || index >= len(p.summaries) {
		return
	}
	s := p.summaries[index]
	recovery, err := p.engine.RecoveryCodeForID(s.AccountID)
	if err != nil {
		recovery = "unavailable"
	}
	p.detail.SetText(accountDetailText(s, recovery))
}

// selected returns the currently selected AccountSummary and true, or a zero value and false
// when the list is empty / selection is out of range.
func (p *accountsPanel) selected() (game.AccountSummary, bool) {
	idx := p.list.GetCurrentItem()
	if idx < 0 || idx >= len(p.summaries) {
		return game.AccountSummary{}, false
	}
	return p.summaries[idx], true
}

// handleKey routes the action keys. tview.List handles ↑/↓ natively; we intercept the rest.
func (p *accountsPanel) handleKey(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyEnter:
		p.doSwitch()
		return nil
	case tcell.KeyEsc:
		p.back()
		return nil
	case tcell.KeyRune:
		switch event.Rune() {
		case 'n', 'N':
			p.doNew()
			return nil
		case 'e', 'E':
			p.doExport()
			return nil
		case 'b', 'B':
			p.doBackup()
			return nil
		case 'i', 'I':
			p.doImport()
			return nil
		case 'r', 'R':
			p.doRecovery()
			return nil
		case 'w', 'W':
			p.doWipe()
			return nil
		case 'x', 'X':
			p.doDeleteSaves()
			return nil
		case 'q', 'Q':
			p.back()
			return nil
		}
	}
	return event
}

// back removes the accounts page and returns to the page it was opened from (the splash menu).
func (p *accountsPanel) back() {
	p.pages.RemovePage(accountsPage)
	p.pages.SwitchToPage(p.returnPage)
}

// doSwitch makes the SELECTED account active, re-applies its theme, rebuilds the list (the
// active marker moves), and shows a transient confirmation. Switching to the already-active
// account is harmless (a no-op switch that just reaffirms it). We stay on the panel so the
// player sees the marker move.
func (p *accountsPanel) doSwitch() {
	s, ok := p.selected()
	if !ok {
		return
	}
	// The panel lives on the main menu, where no game is running (Esc saved and stopped
	// it), so the switch never has a live run to end.
	if _, err := p.engine.SwitchAccount(s.AccountID); err != nil {
		p.status.SetText(vibe.Sprintf("ui.accounts_panel.doSwitch.01", "[red]Switch failed: %v[-]", err))
		return
	}
	// Re-apply the now-active account's persisted theme so the whole UI retints (the theming design
	// §6). applyAccountTheme runs SetActive + Restyle; no QueueUpdateDraw — we're on the main
	// goroutine and tview redraws after this input event.
	applyAccountTheme(p.engine)
	// Rebuild so the ● (current) marker moves to the now-active row, then keep the selection
	// on it. ListAccounts sorts active-first, so the active row is index 0 after the rebuild.
	p.refresh(0)
	p.selectByID(s.AccountID)
	name := displayNameOr(s)
	p.status.SetText(vibe.Sprintf("ui.accounts_panel.doSwitch.02", "[green]● Now playing as %s[-]", name))
}

// doNew opens the name-entry modal (Esc cancels — escAccepts=false), creates the named account
// with NO carry-over, re-applies its (default) theme, then rebuilds the list selecting the new
// account.
func (p *accountsPanel) doNew() {
	showSaveNameModalOpts(p.app, p.pages, vibe.Line("ui.accounts_panel.doNew.01", " Name new account "), accountNamePage, p.list, "", func(name string) {
		acct, err := p.engine.CreateAccount(name)
		if err != nil {
			p.showMessage(vibe.Line("ui.accounts_panel.doNew.02", "New account failed"), fmt.Sprintf("[red]%v[-]", err))
			return
		}
		// A fresh (or reopened same-name) account resolves its own theme — a brand-new one
		// starts on Forge rather than inheriting the previous account's theme (the theming design §6).
		applyAccountTheme(p.engine)
		p.refresh(0)
		p.selectByID(acct.AccountID)
		p.status.SetText(vibe.Sprintf("ui.accounts_panel.doNew.03", "[green]● Created %s[-]", displayNameOrName(acct.DisplayName)))
	}, false)
}

// doExport writes the SELECTED account's progress backup to that account's OWN slot dir (never
// the active DataDir — wrong for a non-active selection) and shows the written path. The blob is
// read by id, so the active account is untouched.
func (p *accountsPanel) doExport() {
	s, ok := p.selected()
	if !ok {
		return
	}
	blob, err := p.engine.ExportAccountByID(s.AccountID)
	if err != nil {
		p.showMessage(vibe.Line("ui.accounts_panel.doExport.01", "Export failed"), fmt.Sprintf("[red]%v[-]", err))
		return
	}
	path := game.AccountExportPath(s.AccountID)
	if dir := filepath.Dir(path); dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			p.showMessage(vibe.Line("ui.accounts_panel.doExport.02", "Export failed"), vibe.Sprintf("ui.accounts_panel.doExport.03", "[red]Could not create %s: %v[-]", dir, err))
			return
		}
	}
	if err := os.WriteFile(path, blob, 0644); err != nil {
		p.showMessage(vibe.Line("ui.accounts_panel.doExport.04", "Export failed"), fmt.Sprintf("[red]%v[-]", err))
		return
	}
	// Also take a FULL slot snapshot (account.json + saves/) alongside the progress blob. A
	// backup failure doesn't fail the export — we just omit the line. Read by id, so the active
	// account is untouched.
	backupLine := ""
	if backupPath, bErr := p.engine.BackupAccount(s.AccountID); bErr == nil {
		backupLine = vibe.Sprintf("ui.accounts_panel.doExport.05", "\n\n[gold]Full backup (account files + saves):[-]\n%s", backupPath)
	}
	msg := vibe.Sprintf(
		"ui.accounts_panel.doExport.06", "[gold]Backed up %s[-]\n\n%s\n\nThis file carries this account's progress (unlocks, stats,\nbadges). Restore it with Import (i) on any machine.%s",
		displayNameOr(s), path, backupLine,
	)
	p.showMessage(vibe.Line("ui.accounts_panel.doExport.07", "Account exported"), msg)
	p.engine.NoteAccountExported(s.AccountID)
}

// doBackup takes a FULL snapshot of the SELECTED account's slot (account.json + saves/) into
// <root>/backups/ and shows the path. Read by id, so the active account is never touched. This
// is the standalone counterpart to the implicit backups that fire on export and wipe.
func (p *accountsPanel) doBackup() {
	s, ok := p.selected()
	if !ok {
		return
	}
	backupPath, err := p.engine.BackupAccount(s.AccountID)
	if err != nil {
		p.showMessage(vibe.Line("ui.accounts_panel.doBackup.01", "Backup failed"), fmt.Sprintf("[red]%v[-]", err))
		return
	}
	msg := vibe.Sprintf(
		"ui.accounts_panel.doBackup.02", "[gold]Backed up %s[-]\n\n%s\n\nThis is a full snapshot: the account's files plus every save in this\naccount's slot. Restore by copying the folder's contents back\ninto data/accounts/<id>/. Only the 10 most recent are kept.",
		displayNameOr(s), backupPath,
	)
	p.showMessage(vibe.Line("ui.accounts_panel.doBackup.03", "Account backed up"), msg)
}

// doImport opens a path-entry modal (defaulting to the selected account's export path), reads
// the file, and restores it into the blob's OWN slot via ImportAccountExport (merge=true). It
// does NOT auto-switch — the refreshed list shows the imported slot so the player can pick it.
// Bad path / corrupt / missing-id errors surface in the modal.
func (p *accountsPanel) doImport() {
	// Default the path to the selected account's export location (a sensible starting point);
	// the player can edit it to point anywhere.
	defaultPath := ""
	if s, ok := p.selected(); ok {
		defaultPath = game.AccountExportPath(s.AccountID)
	}

	errTV := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)
	errTV.SetBackgroundColor(theme.Color(theme.RoleSurface))

	input := tview.NewInputField().
		SetLabel(vibe.Line("ui.accounts_panel.doImport.01", "File path: ")).
		SetText(defaultPath).
		SetFieldWidth(48).
		SetFieldBackgroundColor(theme.Color(theme.RoleSelection)).
		SetFieldTextColor(theme.Color(theme.RoleSelectionText))
	input.SetBackgroundColor(theme.Color(theme.RoleSurface))

	hintTV := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText(vibe.Line("ui.accounts_panel.doImport.02", "[gray]Enter: import (merge)  ·  Esc: cancel[-]"))
	hintTV.SetBackgroundColor(theme.Color(theme.RoleSurface))

	close := func() {
		p.pages.RemovePage(accountsImportPage)
		p.app.SetFocus(p.list)
	}

	submit := func() {
		path := strings.TrimSpace(input.GetText())
		if path == "" {
			errTV.SetText(vibe.Line("ui.accounts_panel.doImport.03", "[red]Enter a file path.[-]"))
			p.app.SetFocus(input)
			return
		}
		blob, err := os.ReadFile(path)
		if err != nil {
			errTV.SetText(vibe.Sprintf("ui.accounts_panel.doImport.04", "[red]Cannot read %s: %v[-]", path, err))
			p.app.SetFocus(input)
			return
		}
		// merge=true: fold the backup into its own slot, restoring without dropping any newer
		// local unlock. ImportAccountExport keys the target slot off the blob's id (a backup
		// of the active account folds into the live account) and does NOT auto-switch, so the
		// active account stays put; the list refresh surfaces the new slot.
		imported, err := p.engine.ImportAccountExport(blob, true)
		if err != nil {
			errTV.SetText(fmt.Sprintf("[red]%v[-]", err))
			p.app.SetFocus(input)
			return
		}
		p.pages.RemovePage(accountsImportPage)
		p.app.SetFocus(p.list)
		p.refresh(0)
		p.selectByID(imported.AccountID)
		if live := p.engine.Account(); live != nil && live.AccountID == imported.AccountID {
			// A backup of the active account can bring a theme choice with it.
			applyAccountTheme(p.engine)
			p.status.SetText(vibe.Sprintf("ui.accounts_panel.doImport.05", "[green]Imported the backup into %s, the active account.[-]", displayNameOrName(imported.Name())))
			return
		}
		p.status.SetText(vibe.Sprintf("ui.accounts_panel.doImport.06", "[green]Imported %s. Press Enter to switch to it.[-]", displayNameOrName(imported.DisplayName)))
	}

	input.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			submit()
		}
	})

	spacer := func() *tview.Box { return tview.NewBox().SetBackgroundColor(theme.Color(theme.RoleSurface)) }
	inner := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(input, 1, 0, true).
		AddItem(spacer(), 1, 0, false).
		AddItem(errTV, 1, 0, false).
		AddItem(spacer(), 1, 0, false).
		AddItem(hintTV, 1, 0, false)
	inner.SetBorder(true).
		SetTitle(vibe.Line("ui.accounts_panel.doImport.07", " Import account backup ")).
		SetTitleColor(theme.Color(theme.RoleAccent)).
		SetBorderColor(theme.Color(theme.RoleAccent))
	inner.SetBackgroundColor(theme.Color(theme.RoleSurface))

	modal := centeredModal(inner, 68, 7)
	modal.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		if ev.Key() == tcell.KeyEsc {
			close()
			return nil
		}
		return ev
	})

	p.pages.AddPage(accountsImportPage, modal, true, true)
	p.app.SetFocus(input)
}

// doRecovery shows the SELECTED account's recovery code in a centered modal, with the same
// honest "identity-only" copy as the `account` command's recovery display.
func (p *accountsPanel) doRecovery() {
	s, ok := p.selected()
	if !ok {
		return
	}
	code, err := p.engine.RecoveryCodeForID(s.AccountID)
	if err != nil {
		p.showMessage(vibe.Line("ui.accounts_panel.doRecovery.01", "Recovery code"), fmt.Sprintf("[red]%v[-]", err))
		return
	}
	msg := vibe.Sprintf(
		"ui.accounts_panel.doRecovery.02", "[gold]%s[-]\n\n[white::b]%s[-]\n\nThis code restores your account ID on another machine or after\na reinstall. It does not restore progress (unlocks, stats,\nbadges); back those up with Export (e). It is not a\npassword: it only proves which account you are.\n\nWrite it down. Restore with:  account recover <code>",
		displayNameOr(s), code,
	)
	p.showMessage(vibe.Line("ui.accounts_panel.doRecovery.03", "Recovery code"), msg)
	p.engine.NoteRecoveryShown(s.AccountID)
}

// wipeSavesLabel is the destructive button on the delete-all-saves dialog.
var wipeSavesLabel = vibe.E("ui.accounts_panel.wipeSavesLabel.01", "Delete all saves")

// showWipeSavesConfirmation asks before every save of an account is deleted. It is the
// dialog the main menu's "Delete all saves" always showed; the entry now lives here, away
// from Quit, and names the account it acts on. confirmed runs on the destructive button;
// done runs after either answer, once the dialog is gone.
func showWipeSavesConfirmation(pages *tview.Pages, account string, confirmed, done func()) {
	modal := tview.NewModal().
		SetText(vibe.Sprintf("ui.accounts_panel.showWipeSavesConfirmation.01", "⚠  Delete all saves?\n\nEvery save of the account \"%s\" is deleted.\nThose runs and their prestige are lost.\nThe account (themes, lifetime stats, badges) is kept.\n\nThis cannot be undone.", account)).
		AddButtons([]string{vibe.Line("ui.accounts_panel.showWipeSavesConfirmation.02", "Cancel"), wipeSavesLabel.String()}).
		SetDoneFunc(func(_ int, buttonLabel string) {
			pages.RemovePage(accountsSavesConfirm)
			if buttonLabel == wipeSavesLabel.String() {
				confirmed()
			}
			done()
		})
	styleDangerModal(modal)
	pages.AddPage(accountsSavesConfirm, modal, true, true)
}

// doDeleteSaves deletes every save of the SELECTED account after the confirmation. The
// account stays. The main menu reads its saves again when the panel closes, so an account
// left with no game goes back to the first page.
func (p *accountsPanel) doDeleteSaves() {
	s, ok := p.selected()
	if !ok {
		return
	}
	idx := p.list.GetCurrentItem()
	showWipeSavesConfirmation(p.pages, displayNameOr(s), func() {
		n, err := p.engine.WipeSavesByID(s.AccountID)
		if err != nil {
			p.status.SetText(vibe.Sprintf("ui.accounts_panel.doDeleteSaves.01", "[red]Could not delete the saves: %v[-]", err))
			return
		}
		p.refresh(idx)
		switch n {
		case 0:
			p.status.SetText(vibe.Sprintf("ui.accounts_panel.doDeleteSaves.02", "[gray]%s has no saves to delete.[-]", displayNameOr(s)))
		case 1:
			p.status.SetText(vibe.Sprintf("ui.accounts_panel.doDeleteSaves.03", "[gray]Deleted the 1 save of %s.[-]", displayNameOr(s)))
		default:
			p.status.SetText(vibe.Sprintf("ui.accounts_panel.doDeleteSaves.04", "[gray]Deleted the %d saves of %s.[-]", n, displayNameOr(s)))
		}
	}, func() { p.app.SetFocus(p.list) })
}

// doWipe runs the two-step gate for permanently deleting the SELECTED account (identity + theme
// unlocks + lifetime stats + achievements), targeting it by id via WipeAccountByID — NOT the
// active-only WipeAccount. Step 1 is a dark-red yes/no; step 2 is a type-the-exact-name confirm.
// After the wipe: if the wiped account was active the engine clears ge.account; if no accounts
// remain we route to the first-run name flow; otherwise we just refresh the list.
func (p *accountsPanel) doWipe() {
	s, ok := p.selected()
	if !ok {
		return
	}
	name := displayNameOr(s)
	wasActive := s.Active

	// STEP 1 — the are-you-sure gate.
	step1 := tview.NewModal().
		SetText(vibe.Sprintf(
			"ui.accounts_panel.doWipe.01", "⚠  Wipe account?\n\nThis deletes the account \"%s\": its theme unlocks,\nlifetime stats, badges and every save in its slot.\n\nA backup goes to data/backups/ first; restoring it is manual.",
			name,
		)).
		AddButtons([]string{vibe.Line("ui.accounts_panel.doWipe.02", "Keep it"), vibe.Line("ui.accounts_panel.doWipe.03", "Wipe it")}).
		SetDoneFunc(func(button int, _ string) {
			p.pages.RemovePage(accountsWipeConfirm1)
			if button == 1 { // "Wipe it"; the other button and Esc (-1) keep the account
				p.showWipeTypeGate(s.AccountID, name, wasActive)
				return
			}
			p.app.SetFocus(p.list)
		})
	styleDangerModal(step1)
	p.pages.AddPage(accountsWipeConfirm1, step1, true, true)
}

// showWipeTypeGate is STEP 2: only an EXACT match of expectedName proceeds. A non-match stays
// with an inline error; Esc aborts with no wipe. On a match it wipes the SELECTED slot by id,
// then resolves the post-wipe state.
func (p *accountsPanel) showWipeTypeGate(id, expectedName string, wasActive bool) {
	errTV := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)

	input := tview.NewInputField().
		SetLabel(vibe.Line("ui.accounts_panel.showWipeTypeGate.01", "Account name: ")).
		SetLabelColor(theme.Color(theme.RoleOnNegative)).
		SetFieldWidth(40).
		SetFieldBackgroundColor(theme.Color(theme.RoleSelection)).
		SetFieldTextColor(theme.Color(theme.RoleText))

	promptTV := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText(fmt.Sprintf(
			dangerTag+vibe.Line("ui.accounts_panel.showWipeTypeGate.02", "Type this name exactly to confirm:[-]\n")+dangerTagBold+"%s[-:-:-]\n"+dangerTagBold+vibe.Line("ui.accounts_panel.showWipeTypeGate.03", "This permanently deletes the account.[-:-:-]\n")+dangerTag+vibe.Line("ui.accounts_panel.showWipeTypeGate.04", "A backup (account and saves) goes to data/backups/ first; restoring it is manual.[-]"),
			expectedName,
		))

	hintTV := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText(dangerTag + vibe.Line("ui.accounts_panel.showWipeTypeGate.05", "Enter: confirm  ·  Esc: cancel[-]"))

	abort := func() {
		p.pages.RemovePage(accountsWipeTypeGate)
		p.app.SetFocus(p.list)
	}

	confirm := func() {
		if strings.TrimSpace(input.GetText()) != expectedName {
			errTV.SetText(dangerTagBold + vibe.Line("ui.accounts_panel.showWipeTypeGate.06", "Name doesn't match. The account was not wiped.[-:-:-]"))
			p.app.SetFocus(input)
			return
		}
		backupPath, err := p.engine.WipeAccountByID(id)
		if err != nil {
			errTV.SetText(fmt.Sprintf(dangerTagBold+vibe.Line("ui.accounts_panel.showWipeTypeGate.07", "Wipe failed: %v[-:-:-]"), err))
			p.app.SetFocus(input)
			return
		}
		p.pages.RemovePage(accountsWipeTypeGate)
		p.resolveAfterWipe(wasActive, backupPath)
	}

	input.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			confirm()
		}
	})

	// Paint every interior primitive dark-red so the modal reads as one solid danger panel.
	promptTV.SetBackgroundColor(theme.Color(theme.RoleNegative))
	errTV.SetBackgroundColor(theme.Color(theme.RoleNegative))
	hintTV.SetBackgroundColor(theme.Color(theme.RoleNegative))
	input.SetBackgroundColor(theme.Color(theme.RoleNegative))
	spacer := func() *tview.Box { return tview.NewBox().SetBackgroundColor(theme.Color(theme.RoleNegative)) }
	inner := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(promptTV, 4, 0, false).
		AddItem(spacer(), 1, 0, false).
		AddItem(input, 1, 0, true).
		AddItem(spacer(), 1, 0, false).
		AddItem(errTV, 1, 0, false).
		AddItem(spacer(), 1, 0, false).
		AddItem(hintTV, 1, 0, false)
	inner.SetBorder(true).
		SetTitle(vibe.Line("ui.accounts_panel.showWipeTypeGate.08", " Confirm account wipe ")).
		SetTitleColor(theme.Color(theme.RoleOnNegative)).
		SetBorderColor(theme.Color(theme.RoleOnNegative))
	inner.SetBackgroundColor(theme.Color(theme.RoleNegative))

	modal := centeredModal(inner, 72, 13)
	modal.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		if ev.Key() == tcell.KeyEsc {
			abort()
			return nil
		}
		return ev
	})

	p.pages.AddPage(accountsWipeTypeGate, modal, true, true)
	p.app.SetFocus(input)
}

// resolveAfterWipe handles the post-wipe state. The engine cleared ge.account if the wiped slot
// was active (so Account() is nil now). If NO accounts remain, route to the first-run name flow
// so the player mints a fresh account. Otherwise: when the active account was wiped, re-theme to
// the default (the live account is gone); always refresh the list. backupPath is the snapshot
// the wipe took before deletion (empty if the backup failed) — surfaced so the player knows a
// recoverable copy survives the wipe.
func (p *accountsPanel) resolveAfterWipe(wasActive bool, backupPath string) {
	remaining := p.engine.ListAccounts()
	if len(remaining) == 0 {
		// Nothing left — mint a new account via the first-run name flow (mirrors splash's
		// post-wipe path). We leave the panel up behind the modal as the focus-return target.
		showAccountNameModal(p.app, p.pages, p.list, func(name string) {
			newAcct, err := game.CreateNamedAccount(name)
			if err != nil {
				return
			}
			p.engine.SetAccount(newAcct)
			applyAccountTheme(p.engine)
			p.refresh(0)
		})
		return
	}
	if wasActive {
		// The live account is gone; drop to the default theme rather than keep the wiped
		// account's. The player can Enter on a row to switch and re-theme.
		applyDefaultTheme()
	}
	p.refresh(0)
	p.app.SetFocus(p.list)
	if backupPath != "" {
		p.status.SetText(vibe.Sprintf("ui.accounts_panel.resolveAfterWipe.01", "[gray]Account wiped. Backup saved to %s[-]", backupPath))
	} else {
		p.status.SetText(vibe.Line("ui.accounts_panel.resolveAfterWipe.02", "[gray]Account wiped.[-]"))
	}
}

// showMessage pops a simple OK modal with a title + body. Opaque background so nothing bleeds.
// Focus returns to the list on dismiss.
func (p *accountsPanel) showMessage(title, body string) {
	tv := tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(true).
		SetTextAlign(tview.AlignCenter).
		SetText(body)
	tv.SetBackgroundColor(theme.Color(theme.RoleSurface))

	hint := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText(vibe.Line("ui.accounts_panel.showMessage.01", "[gray]Enter / Esc: close[-]"))
	hint.SetBackgroundColor(theme.Color(theme.RoleSurface))

	spacer := func() *tview.Box { return tview.NewBox().SetBackgroundColor(theme.Color(theme.RoleSurface)) }
	inner := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tv, 0, 1, false).
		AddItem(spacer(), 1, 0, false).
		AddItem(hint, 1, 0, false)
	inner.SetBorder(true).
		SetTitle(fmt.Sprintf(" %s ", title)).
		SetTitleColor(theme.Color(theme.RoleAccent)).
		SetBorderColor(theme.Color(theme.RoleAccent))
	inner.SetBackgroundColor(theme.Color(theme.RoleSurface))

	close := func() {
		p.pages.RemovePage(accountsMessagePage)
		p.app.SetFocus(p.list)
	}
	modal := centeredModal(inner, 72, 13)
	modal.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		if ev.Key() == tcell.KeyEsc || ev.Key() == tcell.KeyEnter {
			close()
			return nil
		}
		return ev
	})
	p.pages.AddPage(accountsMessagePage, modal, true, true)
	p.app.SetFocus(modal)
}

// selectByID moves the selection to the row whose account id matches, if present.
func (p *accountsPanel) selectByID(id string) {
	for i, s := range p.summaries {
		if s.AccountID == id {
			p.list.SetCurrentItem(i)
			p.updateDetail(i)
			return
		}
	}
}

// ── Pure helpers (formatting) ───────────────────────────────────────────────

// accountRowLabel renders one list row: DisplayName + short id, a ● "(current)" marker on the
// Active account, and a ⚠ marker when the slot's signature is stale (Tampered).
func accountRowLabel(s game.AccountSummary) string {
	name := displayNameOr(s)
	label := fmt.Sprintf("%s   [gray]%s[-]", name, shortAccountID(s.AccountID))
	if s.Active {
		label = "[label]● " + name + "[-]   [gray]" + shortAccountID(s.AccountID) + vibe.Line("ui.accounts_panel.accountRowLabel.01", "[-]   [label](current)[-]")
	}
	if s.Tampered {
		label += vibe.Line("ui.accounts_panel.accountRowLabel.02", "   [red]⚠ modified[-]")
	}
	return label
}

// accountDetailText renders the detail pane for an account summary: identity, recovery code,
// meta-progression, and integrity/selection status.
func accountDetailText(s game.AccountSummary, recovery string) string {
	var lines []string
	lines = append(lines, fmt.Sprintf("[gold]%s[-]", displayNameOr(s)))
	lines = append(lines, vibe.Sprintf("ui.accounts_panel.accountDetailText.01", "[gray]ID[-] [white]%s[-] [gray](%s)[-]", s.AccountID, shortAccountID(s.AccountID)))
	lines = append(lines, vibe.Sprintf("ui.accounts_panel.accountDetailText.02", "[gray]Recovery code[-] [white]%s[-]", recovery))
	lines = append(lines, "")
	lines = append(lines, vibe.Sprintf("ui.accounts_panel.accountDetailText.03", "[gray]Highest age[-] [white]%s[-]", ageDisplay(s.HighestAge)))
	lines = append(lines, vibe.Sprintf("ui.accounts_panel.accountDetailText.04", "[gray]Total prestiges[-] [white]%d[-]", s.TotalPrestiges))
	lines = append(lines, vibe.Sprintf("ui.accounts_panel.accountDetailText.05", "[gray]Badges[-] [white]%d[-]", s.Badges))
	lines = append(lines, "")

	var status []string
	if s.Active {
		status = append(status, vibe.Line("ui.accounts_panel.accountDetailText.06", "[label]● current account[-]"))
	}
	if s.Tampered {
		status = append(status, vibe.Line("ui.accounts_panel.accountDetailText.07", "[red]⚠ modified outside the game[-]"))
	}
	if len(status) == 0 {
		status = append(status, vibe.Line("ui.accounts_panel.accountDetailText.08", "[gray]idle[-]"))
	}
	lines = append(lines, vibe.Line("ui.accounts_panel.accountDetailText.09", "[gray]Status[-] ")+strings.Join(status, "   "))
	return strings.Join(lines, "\n")
}

// accountsKeys are the Accounts panel's actions. Each acts on the selected account.
var accountsKeys = []footerKey{
	{"Enter", vibe.E("ui.accounts_panel.accountsKeys.02", "Switch"), vibe.Entry{}},
	{"n", vibe.E("ui.accounts_panel.accountsKeys.03", "New"), vibe.Entry{}},
	{"e", vibe.E("ui.accounts_panel.accountsKeys.04", "Export"), vibe.Entry{}},
	{"b", vibe.E("ui.accounts_panel.accountsKeys.05", "Backup"), vibe.Entry{}},
	{"i", vibe.E("ui.accounts_panel.accountsKeys.06", "Import"), vibe.Entry{}},
	{"r", vibe.E("ui.accounts_panel.accountsKeys.07", "Recovery"), vibe.Entry{}},
	{"x", vibe.E("ui.accounts_panel.accountsKeys.08", "Delete saves"), vibe.Entry{}},
	{"w", vibe.E("ui.accounts_panel.accountsKeys.09", "Wipe"), vibe.Entry{}},
	{"Esc", vibe.E("ui.accounts_panel.accountsKeys.11", "Back"), vibe.Entry{}},
}

// displayNameOr returns the summary's DisplayName, or "(unnamed)" when empty.
func displayNameOr(s game.AccountSummary) string {
	return displayNameOrName(s.DisplayName)
}

// displayNameOrName returns name, or "(unnamed)" when empty.
func displayNameOrName(name string) string {
	if strings.TrimSpace(name) == "" {
		return vibe.Line("ui.accounts_panel.displayNameOrName.01", "(unnamed)")
	}
	return name
}
