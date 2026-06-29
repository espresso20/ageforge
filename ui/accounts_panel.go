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

	root   *tview.Flex
	list   *tview.List
	detail *tview.TextView
	status *tview.TextView // transient confirmation line ("Now playing as …")

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
	p := &accountsPanel{
		app:            app,
		pages:          pages,
		engine:         engine,
		currentVersion: currentVersion,
		returnPage:     returnPage,
	}

	// ── Title ──────────────────────────────────────────────────────────────────
	title := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[gold]═══ Accounts ═══[-]")

	subtitle := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[gray]Each account is its own slot — switch between them, or back one up to a file.[-]")

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
	p.detail.SetBorder(true).
		SetTitle(" Details ")
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
	footer := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText(accountsFooterBar())

	// ── Layout ─────────────────────────────────────────────────────────────────
	p.root = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(title, 1, 0, false).
		AddItem(subtitle, 1, 0, false).
		AddItem(body, 0, 1, true). // weighted — list+detail take remaining space
		AddItem(p.status, 1, 0, false).
		AddItem(footer, 1, 0, false)

	p.list.SetInputCapture(p.handleKey)

	p.refresh(0)
	return p.root
}

// refresh re-reads the account listing, rebuilds the list rows, and restores a clamped
// selection. wantIdx is the index to select after the rebuild (clamped to range); pass the
// current index to keep position. Called after switch/new/import/wipe.
func (p *accountsPanel) refresh(wantIdx int) {
	p.summaries = p.engine.ListAccounts()

	p.list.Clear()
	if len(p.summaries) == 0 {
		// Empty state — the action keys guard on len(p.summaries) and become no-ops.
		p.detail.SetText("[gray]No accounts found.\n\nCreate one with 'n', or restore a backup with 'i'.[-]")
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
		recovery = "—"
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
	if err := p.engine.SwitchAccount(s.AccountID); err != nil {
		p.status.SetText(fmt.Sprintf("[red]Switch failed: %v[-]", err))
		return
	}
	// Re-apply the now-active account's persisted theme so the whole UI retints (theming.md
	// §6). applyAccountTheme runs SetActive + Restyle; no QueueUpdateDraw — we're on the main
	// goroutine and tview redraws after this input event.
	applyAccountTheme(p.engine)
	// Rebuild so the ● (current) marker moves to the now-active row, then keep the selection
	// on it. ListAccounts sorts active-first, so the active row is index 0 after the rebuild.
	p.refresh(0)
	p.selectByID(s.AccountID)
	name := displayNameOr(s)
	p.status.SetText(fmt.Sprintf("[green]● Now playing as %s[-]", name))
}

// doNew opens the name-entry modal (Esc cancels — escAccepts=false), creates the named account
// with NO carry-over, re-applies its (default) theme, then rebuilds the list selecting the new
// account.
func (p *accountsPanel) doNew() {
	showSaveNameModalOpts(p.app, p.pages, " Name New Account ", accountNamePage, p.list, "", func(name string) {
		acct, err := p.engine.CreateAccount(name)
		if err != nil {
			p.showMessage("New Account Failed", fmt.Sprintf("[red]%v[-]", err))
			return
		}
		// A fresh (or reopened same-name) account resolves its own theme — a brand-new one
		// starts on Forge rather than inheriting the previous account's theme (theming.md §6).
		applyAccountTheme(p.engine)
		p.refresh(0)
		p.selectByID(acct.AccountID)
		p.status.SetText(fmt.Sprintf("[green]● Created %s[-]", displayNameOrName(acct.DisplayName)))
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
		p.showMessage("Export Failed", fmt.Sprintf("[red]%v[-]", err))
		return
	}
	path := game.AccountExportPath(s.AccountID)
	if dir := filepath.Dir(path); dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			p.showMessage("Export Failed", fmt.Sprintf("[red]could not create %s: %v[-]", dir, err))
			return
		}
	}
	if err := os.WriteFile(path, blob, 0644); err != nil {
		p.showMessage("Export Failed", fmt.Sprintf("[red]%v[-]", err))
		return
	}
	// Also take a FULL slot snapshot (account.json + saves/) alongside the progress blob. A
	// backup failure doesn't fail the export — we just omit the line. Read by id, so the active
	// account is untouched.
	backupLine := ""
	if backupPath, bErr := p.engine.BackupAccount(s.AccountID); bErr == nil {
		backupLine = fmt.Sprintf("\n\n[gold]Full backup (account.json + saves):[-]\n%s", backupPath)
	}
	msg := fmt.Sprintf(
		"[gold]Backed up %s[-]\n\n%s\n\nThis file carries this account's progress (unlocks, stats,\nachievements). Restore it with Import (i) on any machine.%s",
		displayNameOr(s), path, backupLine,
	)
	p.showMessage("Account Exported", msg)
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
		p.showMessage("Backup Failed", fmt.Sprintf("[red]%v[-]", err))
		return
	}
	msg := fmt.Sprintf(
		"[gold]Backed up %s[-]\n\n%s\n\nThis is a FULL snapshot: account.json plus every save in this\naccount's slot. Restore by copying the folder's contents back\ninto data/accounts/<id>/. Only the 10 most recent are kept.",
		displayNameOr(s), backupPath,
	)
	p.showMessage("Account Backed Up", msg)
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
	errTV.SetBackgroundColor(theme.Color(theme.RoleBackground))

	input := tview.NewInputField().
		SetLabel("File path: ").
		SetText(defaultPath).
		SetFieldWidth(48).
		SetFieldBackgroundColor(theme.Color(theme.RoleSelection)).
		SetFieldTextColor(theme.Color(theme.RoleText))

	hintTV := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[gray]Enter: import (merge)  ·  Esc: cancel[-]")
	hintTV.SetBackgroundColor(theme.Color(theme.RoleBackground))

	close := func() {
		p.pages.RemovePage(accountsImportPage)
		p.app.SetFocus(p.list)
	}

	submit := func() {
		path := strings.TrimSpace(input.GetText())
		if path == "" {
			errTV.SetText("[red]Enter a file path.[-]")
			p.app.SetFocus(input)
			return
		}
		blob, err := os.ReadFile(path)
		if err != nil {
			errTV.SetText(fmt.Sprintf("[red]cannot read %s: %v[-]", path, err))
			p.app.SetFocus(input)
			return
		}
		// merge=true: fold the backup into its own slot, restoring without dropping any newer
		// local unlock. ImportAccountExport keys the target slot off the blob's id and does NOT
		// auto-switch, so the active account stays put; the list refresh surfaces the new slot.
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
		p.status.SetText(fmt.Sprintf("[green]Imported %s — press Enter to switch to it[-]", displayNameOrName(imported.DisplayName)))
	}

	input.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			submit()
		}
	})

	spacer := func() *tview.Box { return tview.NewBox().SetBackgroundColor(theme.Color(theme.RoleBackground)) }
	inner := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(input, 1, 0, true).
		AddItem(spacer(), 1, 0, false).
		AddItem(errTV, 1, 0, false).
		AddItem(spacer(), 1, 0, false).
		AddItem(hintTV, 1, 0, false)
	inner.SetBorder(true).
		SetTitle(" Import Account Backup ").
		SetTitleColor(theme.Color(theme.RoleAccent)).
		SetBorderColor(theme.Color(theme.RoleAccent))
	inner.SetBackgroundColor(theme.Color(theme.RoleBackground))

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
		p.showMessage("Recovery Code", fmt.Sprintf("[red]%v[-]", err))
		return
	}
	msg := fmt.Sprintf(
		"[gold]%s[-]\n\n[white::b]%s[-]\n\nThis code restores your IDENTITY (your account ID) across\nmachines and reinstalls — NOT your earned progress. It is not\na password; it only proves which account you are.\n\nWrite it down. Restore with:  account recover <code>\nProgress (unlocks, stats) is backed up separately via Export.",
		displayNameOr(s), code,
	)
	p.showMessage("Recovery Code", msg)
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
		SetText(fmt.Sprintf(
			"⚠  WIPE ACCOUNT  ⚠\n\nThis permanently deletes the account \"%s\":\nall theme unlocks, lifetime stats, and achievements.\n\nThis CANNOT be undone. Game saves are NOT affected.",
			name,
		)).
		AddButtons([]string{"Keep it", "Wipe it"}).
		SetDoneFunc(func(_ int, label string) {
			p.pages.RemovePage(accountsWipeConfirm1)
			if label == "Wipe it" {
				p.showWipeTypeGate(s.AccountID, name, wasActive)
				return
			}
			p.app.SetFocus(p.list)
		})
	step1.SetBackgroundColor(theme.Color(theme.RoleNegative))
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
		SetLabel("Account name: ").
		SetFieldWidth(40).
		SetFieldBackgroundColor(theme.Color(theme.RoleSelection)).
		SetFieldTextColor(theme.Color(theme.RoleText))

	promptTV := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText(fmt.Sprintf(
			"[white]Type this name exactly to confirm:[-]\n[yellow::b]%s[-]\n[red]This permanently deletes the account.[-]\n[gray]A full backup is saved to data/backups/ first, so a copy is recoverable.[-]",
			expectedName,
		))

	hintTV := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[red]Enter: confirm  ·  Esc: cancel[-]")

	abort := func() {
		p.pages.RemovePage(accountsWipeTypeGate)
		p.app.SetFocus(p.list)
	}

	confirm := func() {
		if strings.TrimSpace(input.GetText()) != expectedName {
			errTV.SetText("[yellow]Name doesn't match — account NOT wiped.[-]")
			p.app.SetFocus(input)
			return
		}
		backupPath, err := p.engine.WipeAccountByID(id)
		if err != nil {
			errTV.SetText(fmt.Sprintf("[red]Wipe failed: %v[-]", err))
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
		SetTitle(" Confirm Account Wipe ").
		SetTitleColor(theme.Color(theme.RoleText)).
		SetBorderColor(theme.Color(theme.RoleText))
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
		p.status.SetText(fmt.Sprintf("[gray]Account wiped — backup saved to %s[-]", backupPath))
	} else {
		p.status.SetText("[gray]Account wiped.[-]")
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
	tv.SetBackgroundColor(theme.Color(theme.RoleBackground))

	hint := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[gray]Enter / Esc: close[-]")
	hint.SetBackgroundColor(theme.Color(theme.RoleBackground))

	spacer := func() *tview.Box { return tview.NewBox().SetBackgroundColor(theme.Color(theme.RoleBackground)) }
	inner := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tv, 0, 1, false).
		AddItem(spacer(), 1, 0, false).
		AddItem(hint, 1, 0, false)
	inner.SetBorder(true).
		SetTitle(fmt.Sprintf(" %s ", title)).
		SetTitleColor(theme.Color(theme.RoleAccent)).
		SetBorderColor(theme.Color(theme.RoleAccent))
	inner.SetBackgroundColor(theme.Color(theme.RoleBackground))

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
		label = "[aqua]● " + name + "[-]   [gray]" + shortAccountID(s.AccountID) + "[-]   [aqua](current)[-]"
	}
	if s.Tampered {
		label += "   [red]⚠ modified[-]"
	}
	return label
}

// accountDetailText renders the detail pane for an account summary: identity, recovery code,
// meta-progression, and integrity/selection status.
func accountDetailText(s game.AccountSummary, recovery string) string {
	var lines []string
	lines = append(lines, fmt.Sprintf("[gold]%s[-]", displayNameOr(s)))
	lines = append(lines, fmt.Sprintf("[gray]ID[-] [white]%s[-] [gray](%s)[-]", s.AccountID, shortAccountID(s.AccountID)))
	lines = append(lines, fmt.Sprintf("[gray]Recovery code[-] [white]%s[-]", recovery))
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("[gray]Highest age[-] [white]%s[-]", ageDisplay(s.HighestAge)))
	lines = append(lines, fmt.Sprintf("[gray]Total prestiges[-] [white]%d[-]", s.TotalPrestiges))
	lines = append(lines, fmt.Sprintf("[gray]Achievements[-] [white]%d[-]", s.Achievements))
	lines = append(lines, "")

	var status []string
	if s.Active {
		status = append(status, "[aqua]● current account[-]")
	}
	if s.Tampered {
		status = append(status, "[red]⚠ modified outside the game[-]")
	}
	if len(status) == 0 {
		status = append(status, "[gray]idle[-]")
	}
	lines = append(lines, "[gray]Status[-] "+strings.Join(status, "   "))
	return strings.Join(lines, "\n")
}

// accountsFooterBar is the Accounts panel action bar: a keycap button per action (footerButton
// lives in load_game.go).
func accountsFooterBar() string {
	return "  " + strings.Join([]string{
		footerButton("Enter", "Switch"),
		footerButton("n", "New"),
		footerButton("e", "Export"),
		footerButton("b", "Backup"),
		footerButton("i", "Import"),
		footerButton("r", "Recovery"),
		footerButton("w", "Wipe"),
		footerButton("Esc", "Back"),
	}, "  ") + "  "
}

// displayNameOr returns the summary's DisplayName, or "(unnamed)" when empty.
func displayNameOr(s game.AccountSummary) string {
	return displayNameOrName(s.DisplayName)
}

// displayNameOrName returns name, or "(unnamed)" when empty.
func displayNameOrName(name string) string {
	if strings.TrimSpace(name) == "" {
		return "(unnamed)"
	}
	return name
}
