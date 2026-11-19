package ui

import (
	"fmt"
	"strings"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/rowid"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/window/kit"
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
// modal. "Create" calls the caller's onConfirm (which runs game.CreateAccount +
// SetAccount); "Re-type" reopens the entry modal pre-filled with the typed name so the
// player can fix a typo. onConfirm receives the cleaned name only once the player confirms.
func showAccountNameModal(app *appLoop, pages *nativePages, restoreFocus Primitive, onConfirm func(name string)) {
	// openEntry forward-declared so the confirm modal's "Re-type" can reopen it.
	var openEntry func(prefill string)

	openEntry = func(prefill string) {
		showSaveNameModalOpts(app, pages, rowid.Line("ui.newgame_modal.showAccountNameModal.01", " Name your AgeForge account "), accountNamePage, restoreFocus, prefill, func(name string) {
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
func showAccountNameConfirmModal(app *appLoop, pages *nativePages, name string, onCreate func(name string), onRetype func()) {
	text := rowid.Sprintf("ui.newgame_modal.showAccountNameConfirmModal.01", "Create your AgeForge account as %q?\nThis name is your identity. To restore your account on another device, enter it exactly as written here.", name)
	buttons := []string{rowid.Line("ui.newgame_modal.showAccountNameConfirmModal.02", "Create"), rowid.Line("ui.newgame_modal.showAccountNameConfirmModal.03", "Re-type")}
	modal := showButtonModal(app, pages, accountNameConfirmPage, text, buttons, func(button int, _ string) {
		if button == 0 { // "Create"
			onCreate(name)
			return
		}
		// "Re-type" or any close (Esc) — never abort into a no-account state.
		onRetype()
	})
	// Opaque background so the modal doesn't bleed the page beneath it, matching the
	// other button-modals (and the earlier modal-opacity fix).
	modal.SetBackgroundColor(theme.Of(theme.RoleSurface))
	app.SetFocus(modal)
}

// showNewGameNameModal pops the New Game name prompt. It is a thin wrapper over
// showSaveNameModal so the splash menu keeps its existing behavior unchanged.
//
// restoreFocus is the splash menu primitive to refocus on cancel.
func showNewGameNameModal(app *appLoop, pages *nativePages, restoreFocus Primitive, onConfirm func(name string)) {
	showSaveNameModal(app, pages, rowid.Line("ui.newgame_modal.showNewGameNameModal.01", " Name your civilization "), newGameNamePage, restoreFocus, onConfirm)
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
func showSaveNameModal(app *appLoop, pages *nativePages, title, pageName string, focusReturn Primitive, onConfirm func(name string)) {
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
func showSaveNameModalOpts(app *appLoop, pages *nativePages, title, pageName string, focusReturn Primitive, prefill string, onConfirm func(name string), escAccepts bool) {
	errTV := newNativeText().
		SetDynamicColors(true).
		SetTextAlign(kit.AlignCenter)

	initial := prefill
	if initial == "" {
		initial = game.GenerateSaveName()
	}
	input := newNativeInput().
		SetLabel(rowid.Line("ui.newgame_modal.showSaveNameModalOpts.01", "Name: ")).
		SetText(initial).
		SetFieldWidth(40).
		// Dark slate field with white text — the default light field
		// background renders the white name near-invisible on the dark modal.
		SetFieldBackgroundColor(theme.Of(theme.RoleSelection)).
		SetFieldTextColor(theme.Of(theme.RoleSelectionText))
	// Every child of the modal panel paints Surface so the panel reads as one
	// piece on themes where Surface differs from the canvas (Daylight).
	errTV.SetBackgroundColor(theme.Of(theme.RoleSurface))
	input.SetBackgroundColor(theme.Of(theme.RoleSurface))

	escHint := rowid.Line("ui.newgame_modal.showSaveNameModalOpts.02", "Esc: cancel")
	if escAccepts {
		escHint = rowid.Line("ui.newgame_modal.showSaveNameModalOpts.03", "Esc: accept")
	}
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
	input.SetDoneFunc(func(key kit.KeyCode) {
		if key == kit.KeyEnter {
			submit()
		}
	})

	escape := closeAndRestore
	if escAccepts {
		// First-run account naming can't be cancelled into a no-account state:
		// Esc accepts the current/generated value via the normal submit path.
		escape = submit
	}

	// The keys the box answers to, as buttons.
	hints, _ := keyHints(rowid.Line("ui.newgame_modal.showSaveNameModalOpts.04", "[gray]Enter: confirm  ·  Tab: reroll name  ·  ") + escHint + "[-]")
	var choices []choice
	for _, h := range hints {
		c := choice{label: h[1], key: h[0], kind: keyPlain}
		switch h[0] {
		case "Enter":
			c.kind, c.sign, c.press = keyPrimary, signTick, submit
		case "Esc":
			c.sign, c.press = signCross, escape
		case "Tab":
			c.press = reroll
		}
		choices = append(choices, c)
	}
	buttons, _ := choiceRow(choices, 1, -1)

	// Opaque spacers between the fields: transparent ones would let the page
	// beneath (the dashboard, for a Branch save) bleed through.
	inner := newNativeFlex().SetDirection(flexRow).
		AddItem(input, 1, 0, true).
		AddItem(surfaceBox(), 1, 0, false).
		AddItem(errTV, 1, 0, false).
		AddItem(surfaceBox(), 1, 0, false).
		AddItem(buttons, actionRows, 0, false)
	framed(inner, title, theme.RoleAccent).SetBackgroundColor(theme.Of(theme.RoleSurface))

	// Esc cancels; Tab rerolls a fresh suggestion. Enter is handled by the field's
	// DoneFunc above so plain typing keys still reach the input.
	//
	// Height is sized to the exact content (input, gap, error, gap, buttons + 2 border). An oversized
	// box would leave unallocated interior rows that no primitive paints, and
	// those rows would show the dashboard through them.
	modal := centeredModal(inner, 66, 4+actionRows+2)
	takeKeys(modal, map[kit.KeyCode]func(){kit.KeyEsc: escape, kit.KeyTab: reroll, kit.KeyBacktab: reroll})

	pages.AddPage(pageName, modal, true, true)
	app.SetFocus(input)
}
