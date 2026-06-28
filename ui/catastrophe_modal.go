package ui

import (
	"fmt"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/theme"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// showCatastropheModal displays a full-screen overlay letting the player choose
// Endure, Succumb, or Defer for the pending catastrophe.
// Must be called from the UI goroutine.
func (d *Dashboard) showCatastropheModal(epochKey string) {
	catName, catFlavor := config.CatastropheInfo(epochKey)
	ep := config.EpochByKey()[epochKey]

	// --- build text sections ---

	header := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText(fmt.Sprintf("[red]☄ %s[-]\n[gray]%s[-]", catName, catFlavor))

	endureText := tview.NewTextView().
		SetDynamicColors(true).
		SetText("[white]── ENDURE — Weather the catastrophe ──[-]\n" +
			"  [red]• 20% buildings randomly destroyed[-]\n" +
			"  [red]• All resources reduced to 15%[-]\n" +
			"  [red]• 25% of workers lost[-]\n" +
			"  [red]• Production -10% for 216 ticks (reconstruction)[-]\n" +
			"  [green]✓ Research & wonders preserved[-]\n" +
			"  [green]✓ Survived marker on epoch badge[-]\n" +
			"  [green]✓ Catastrophe recorded in civilization history[-]")

	succumbText := tview.NewTextView().
		SetDynamicColors(true).
		SetText("[white]── SUCCUMB — Let civilization fall ──[-]\n" +
			"  [red]• Full reset: buildings, resources, workers[-]\n" +
			"  [red]• Age resets to Primitive[-]\n" +
			"  [green]✓ 8 Ruins carry into next run (50% production)[-]\n" +
			"  [green]✓ Ancient Knowledge: +25% research speed (permanent)[-]\n" +
			fmt.Sprintf("  [gold]✓ %s Legacy Bonus: %s[-]\n", ep.Name, legacyBonusText(epochKey)) +
			"  [green]✓ Catastrophe recorded in civilization history[-]")

	// --- buttons ---
	btnEndure := tview.NewButton("[ENDURE]").
		SetSelectedFunc(func() {
			if err := d.engine.Endure(); err != nil {
				d.engine.AddLog("error", "Endure failed: "+err.Error())
			}
			d.closeCatastropheModal()
		})
	btnEndure.SetBackgroundColor(theme.Color(theme.RoleNegative))
	btnEndure.SetLabelColor(theme.Color(theme.RoleText))

	btnSuccumb := tview.NewButton("[SUCCUMB]").
		SetSelectedFunc(func() {
			if err := d.engine.Succumb(); err != nil {
				d.engine.AddLog("error", "Succumb failed: "+err.Error())
			}
			d.closeCatastropheModal()
		})
	btnSuccumb.SetBackgroundColor(theme.Color(theme.RoleNegative))
	btnSuccumb.SetLabelColor(theme.Color(theme.RoleHighlight))

	btnDefer := tview.NewButton("[Defer — Decide Later]").
		SetSelectedFunc(func() {
			d.closeCatastropheModal()
		})

	// --- button row ---
	btnRow := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(btnEndure, 12, 0, true).
		AddItem(tview.NewBox(), 2, 0, false).
		AddItem(btnSuccumb, 12, 0, false).
		AddItem(tview.NewBox(), 2, 0, false).
		AddItem(btnDefer, 22, 0, false)

	// --- inner box ---
	inner := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header, 3, 0, false).
		AddItem(tview.NewBox(), 1, 0, false).
		AddItem(endureText, 8, 0, false).
		AddItem(tview.NewBox(), 1, 0, false).
		AddItem(succumbText, 8, 0, false).
		AddItem(tview.NewBox(), 1, 0, false).
		AddItem(btnRow, 1, 0, true)
	inner.SetBorder(true).
		SetTitle(fmt.Sprintf(" ☄ %s Catastrophe ", ep.Name)).
		SetTitleColor(theme.Color(theme.RoleNegative)).
		SetBorderColor(theme.Color(theme.RoleNegative))

	// --- centered overlay ---
	modal := tview.NewFlex().
		AddItem(tview.NewBox(), 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(tview.NewBox(), 0, 1, false).
			AddItem(inner, 28, 0, true).
			AddItem(tview.NewBox(), 0, 1, false),
			70, 0, true).
		AddItem(tview.NewBox(), 0, 1, false)

	// Tab cycles through buttons; keyboard shortcuts
	focusOrder := []tview.Primitive{btnEndure, btnSuccumb, btnDefer}
	focusIdx := 0
	modal.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab, tcell.KeyRight:
			focusIdx = (focusIdx + 1) % len(focusOrder)
			d.app.SetFocus(focusOrder[focusIdx])
			return nil
		case tcell.KeyBacktab, tcell.KeyLeft:
			focusIdx = (focusIdx - 1 + len(focusOrder)) % len(focusOrder)
			d.app.SetFocus(focusOrder[focusIdx])
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'e', 'E':
				d.app.SetFocus(btnEndure)
				btnEndure.InputHandler()(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), nil)
				return nil
			case 's', 'S':
				d.app.SetFocus(btnSuccumb)
				btnSuccumb.InputHandler()(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), nil)
				return nil
			case 'd', 'D':
				d.closeCatastropheModal()
				return nil
			}
		}
		return event
	})

	d.pages.AddPage("catastrophe", modal, true, true)
	d.app.SetFocus(btnEndure)
}

// closeCatastropheModal removes the catastrophe overlay page.
func (d *Dashboard) closeCatastropheModal() {
	d.pages.RemovePage("catastrophe")
	// catModalShown stays set (prevents immediate re-show after Defer)
}

// legacyBonusText returns a short summary of the epoch legacy bonus.
func legacyBonusText(epochKey string) string {
	bonuses := config.LegacyBonusForEpoch(epochKey)
	if len(bonuses) == 0 {
		return "none"
	}
	first := true
	result := ""
	for res, mult := range bonuses {
		if !first {
			result += ", "
		}
		result += fmt.Sprintf("%s +%.0f%%", res, mult*100)
		first = false
	}
	return result
}

