package ui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/game"
)

// newGameNamePage is the unique page name for the New Game name-entry modal.
const newGameNamePage = "newgame_name"

// accountNamePage is the unique page name for the first-run account-name modal.
const accountNamePage = "account_name"

// accountNameConfirmPage is the unique page name for the first-run account-name
// confirmation modal that sits between name entry and account creation.
const accountNameConfirmPage = "account_name_confirm"

// showAccountNameModal pops the first-run "name your account" prompt over the splash.
// It reuses the name-entry modal with the account variant's Esc policy (Esc accepts the
// current/generated value — naming can't be cancelled into an accountless state), so the
// first run ALWAYS yields a named account.
//
// Because the account identity is name-derived (a typo mints a different, empty account),
// the entry modal does NOT create immediately: submitting the name opens a confirmation
// modal. "Create" calls the caller's onConfirm (which runs game.CreateNamedAccount +
// SetAccount); "Re-type" reopens the entry modal pre-filled with the typed name so the
// player can fix a typo. onConfirm receives the cleaned name only once the player confirms.
func showAccountNameModal(app *tview.Application, pages *tview.Pages, restoreFocus tview.Primitive, onConfirm func(name string)) {
	// openEntry forward-declared so the confirm modal's "Re-type" can reopen it.
	var openEntry func(prefill string)

	openEntry = func(prefill string) {
		showSaveNameModalOpts(app, pages, " Name Your AgeForge Account ", accountNamePage, restoreFocus, prefill, func(name string) {
			// Don't create yet — confirm the name first, since it IS the identity.
			showAccountNameConfirmModal(app, pages, name, onConfirm, func() {
				// Re-type: reopen the entry modal pre-filled with what they typed.
				openEntry(name)
			})
		}, true)
	}

	openEntry("")
}

// showAccountNameConfirmModal pops the confirmation step between name entry and account
// creation. onCreate fires the caller's account-creation callback with the chosen name;
// onRetype reopens the entry modal so a mistyped name can be fixed. The modal cannot be
// cancelled into a no-account state: its own Esc/close routes to onRetype (re-enter),
// consistent with the entry modal's Esc-accepts policy.
func showAccountNameConfirmModal(app *tview.Application, pages *tview.Pages, name string, onCreate func(name string), onRetype func()) {
	modal := tview.NewModal().
		SetText(fmt.Sprintf("Create your AgeForge account as %q?\nThis name is your identity — re-enter it exactly to restore your account on another device.", name)).
		AddButtons([]string{"Create", "Re-type"}).
		SetDoneFunc(func(_ int, label string) {
			pages.RemovePage(accountNameConfirmPage)
			switch label {
			case "Create":
				onCreate(name)
			default:
				// "Re-type" or any close (Esc) — never abort into a no-account state.
				onRetype()
			}
		})
	// Opaque background so the modal doesn't bleed the page beneath it, matching the
	// other button-modals (and the earlier modal-opacity fix).
	modal.SetBackgroundColor(tcell.ColorBlack)

	pages.AddPage(accountNameConfirmPage, modal, true, true)
	app.SetFocus(modal)
}

// showNewGameNameModal pops the New Game name prompt. It is a thin wrapper over
// showSaveNameModal so the splash menu keeps its existing behavior unchanged.
//
// restoreFocus is the splash menu primitive to refocus on cancel.
func showNewGameNameModal(app *tview.Application, pages *tview.Pages, restoreFocus tview.Primitive, onConfirm func(name string)) {
	showSaveNameModal(app, pages, " Name Your Civilization ", newGameNamePage, restoreFocus, onConfirm)
}

// showSaveNameModal pops a centered, bordered name-entry prompt. The input is
// pre-filled with a procedural suggestion (game.GenerateSaveName) that Tab
// rerolls. Enter validates the name with the same rule the rename/save paths use
// (game.ValidateSaveName) and, on success, removes the page and calls onConfirm
// with the cleaned name. Esc cancels: it removes the page and restores focus to
// focusReturn without invoking onConfirm.
//
// title customizes the heading; pageName must be distinct per caller so two
// concurrent modals (e.g. New Game vs. Branch) can't collide on the page stack.
func showSaveNameModal(app *tview.Application, pages *tview.Pages, title, pageName string, focusReturn tview.Primitive, onConfirm func(name string)) {
	showSaveNameModalOpts(app, pages, title, pageName, focusReturn, "", onConfirm, false)
}

// showSaveNameModalOpts is showSaveNameModal with an explicit Esc policy. When
// escAccepts is false (the default for New Game / Branch), Esc cancels — removing
// the page and restoring focus without calling onConfirm. When escAccepts is true
// (first-run account naming), Esc instead ACCEPTS the current/generated value: the
// account naming step must always produce an account, so it cannot be cancelled into
// an accountless state. The hint line adjusts to match.
//
// prefill seeds the input. When empty, a procedural suggestion (game.GenerateSaveName)
// is used — the normal first-open behavior. A non-empty prefill is used verbatim, so the
// account "Re-type" path can reopen with the name the player just typed for editing.
func showSaveNameModalOpts(app *tview.Application, pages *tview.Pages, title, pageName string, focusReturn tview.Primitive, prefill string, onConfirm func(name string), escAccepts bool) {
	errTV := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)

	initial := prefill
	if initial == "" {
		initial = game.GenerateSaveName()
	}
	input := tview.NewInputField().
		SetLabel("Name: ").
		SetText(initial).
		SetFieldWidth(40).
		// Dark slate field with white text — tview's default light field
		// background renders the white name near-invisible on the dark modal.
		SetFieldBackgroundColor(tcell.NewRGBColor(48, 54, 61)).
		SetFieldTextColor(tcell.ColorWhite)

	escHint := "Esc: cancel"
	if escAccepts {
		escHint = "Esc: accept"
	}
	hintTV := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[#8b949e]Enter: confirm  ·  Tab: reroll name  ·  " + escHint + "[-]")

	// closeAndRestore removes the modal page and restores focus to focusReturn.
	// Used on cancel (Esc).
	closeAndRestore := func() {
		pages.RemovePage(pageName)
		if focusReturn != nil {
			app.SetFocus(focusReturn)
		}
	}

	submit := func() {
		raw := strings.TrimSpace(input.GetText())
		if raw == "" {
			// Empty → fall back to a fresh suggestion rather than erroring out.
			raw = game.GenerateSaveName()
		}
		name, err := game.ValidateSaveName(raw)
		if err != nil {
			errTV.SetText(fmt.Sprintf("[red]%v[-]", err))
			app.SetFocus(input)
			return
		}
		pages.RemovePage(pageName)
		onConfirm(name)
	}

	reroll := func() {
		input.SetText(game.GenerateSaveName())
		errTV.SetText("")
		app.SetFocus(input)
	}

	// Enter in the field submits.
	input.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			submit()
		}
	})

	// Opaque spacers between the fields. tview only clears cells a primitive
	// actually draws, so transparent spacers would let the page beneath (the
	// dashboard, for a Branch save) bleed through; an explicit background paints
	// them.
	spacer := func() *tview.Box { return tview.NewBox().SetBackgroundColor(tcell.ColorBlack) }
	inner := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(input, 1, 0, true).
		AddItem(spacer(), 1, 0, false).
		AddItem(errTV, 1, 0, false).
		AddItem(spacer(), 1, 0, false).
		AddItem(hintTV, 1, 0, false)
	inner.SetBorder(true).
		SetTitle(title).
		SetTitleColor(tcell.ColorGold).
		SetBorderColor(tcell.ColorGold)
	inner.SetBackgroundColor(tcell.ColorBlack)

	// Esc cancels; Tab rerolls a fresh suggestion. Enter is handled by the field's
	// DoneFunc above so plain typing keys still reach the input.
	//
	// Height is sized to the exact content (5 rows + 2 border = 7). An oversized
	// box would leave unallocated interior rows that no primitive paints, and
	// those rows would show the dashboard through them.
	modal := centeredModal(inner, 60, 7)
	modal.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		switch ev.Key() {
		case tcell.KeyEsc:
			if escAccepts {
				// First-run account naming can't be cancelled into a no-account state:
				// Esc accepts the current/generated value via the normal submit path.
				submit()
			} else {
				closeAndRestore()
			}
			return nil
		case tcell.KeyTab, tcell.KeyBacktab:
			reroll()
			return nil
		}
		return ev
	})

	pages.AddPage(pageName, modal, true, true)
	app.SetFocus(input)
}
