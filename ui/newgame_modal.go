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

// showNewGameNameModal pops a centered, bordered prompt asking the player to name
// their civilization. The input is pre-filled with a procedural suggestion that
// Tab rerolls. Enter validates the name with the same rule the rename/save paths
// use (game.ValidateSaveName) and, on success, removes the page and calls
// onConfirm with the cleaned name. Esc cancels: it removes the page and restores
// focus to the splash menu without resetting or starting a game.
//
// restoreFocus is the splash menu primitive to refocus on cancel.
func showNewGameNameModal(app *tview.Application, pages *tview.Pages, restoreFocus tview.Primitive, onConfirm func(name string)) {
	errTV := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)

	input := tview.NewInputField().
		SetLabel("Name: ").
		SetText(game.GenerateSaveName()).
		SetFieldWidth(40)

	hintTV := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[#8b949e]Enter: begin  ·  Tab: reroll name  ·  Esc: cancel[-]")

	// close removes the modal page and restores focus to the splash menu. Used on
	// both cancel (Esc) and on a successful confirm (the page is gone before the
	// game starts in either path).
	closeAndRestore := func() {
		pages.RemovePage(newGameNamePage)
		if restoreFocus != nil {
			app.SetFocus(restoreFocus)
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
		pages.RemovePage(newGameNamePage)
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

	inner := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(input, 1, 0, true).
		AddItem(tview.NewBox(), 1, 0, false).
		AddItem(errTV, 1, 0, false).
		AddItem(tview.NewBox(), 1, 0, false).
		AddItem(hintTV, 1, 0, false)
	inner.SetBorder(true).
		SetTitle(" Name Your Civilization ").
		SetTitleColor(tcell.ColorGold).
		SetBorderColor(tcell.ColorGold)

	// Esc cancels; Tab rerolls a fresh suggestion. Enter is handled by the field's
	// DoneFunc above so plain typing keys still reach the input.
	modal := centeredModal(inner, 50, 9)
	modal.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		switch ev.Key() {
		case tcell.KeyEsc:
			closeAndRestore()
			return nil
		case tcell.KeyTab, tcell.KeyBacktab:
			reroll()
			return nil
		}
		return ev
	})

	pages.AddPage(newGameNamePage, modal, true, true)
	app.SetFocus(input)
}
