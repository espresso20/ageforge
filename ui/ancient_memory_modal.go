package ui

import (
	"fmt"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/theme"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ancientMemoryPage is the tview page name for the Ancient Memory offer overlay.
const ancientMemoryPage = "ancient_memory"

// showAncientMemoryModal presents the Ancient Civilization Memory offer (Trello
// yn98pTQw): early in a new prestige run the player has discovered a cache holding
// a memory of their extinct previous civilization, offering one tech free of
// prerequisites but at half research speed. The player may Accept (begin the
// slowed, prereq-bypassing research) or Decline (the cache crumbles, no effect).
//
// Mirrors showCatastropheModal: a centered Flex overlay with two callback buttons.
// Must be called from the UI goroutine. The button callbacks call engine methods
// directly — safe here because they run on the UI goroutine, not under the engine
// lock (same pattern as the catastrophe Endure/Succumb buttons).
func (d *Dashboard) showAncientMemoryModal(techKey, techName string) {
	def := config.TechByKey()[techKey]
	if techName == "" {
		techName = def.Name
	}

	header := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[gray]You have discovered an old cache. It appears to contain\n" +
			"memories of a now-extinct civilization.[-]")

	body := tview.NewTextView().
		SetDynamicColors(true).
		SetText(fmt.Sprintf("[white]── Recovered Memory: [-][cyan]%s[-]\n", techName) +
			fmt.Sprintf("  [gray]%s[-]\n\n", def.Description) +
			"  [green]✓ Research it now, free of prerequisites[-]\n" +
			"  [red]• It returns slowly — half research speed (2× ticks)[-]\n" +
			"  [gray]Only one such memory surfaces per civilization.[-]")

	btnAccept := tview.NewButton("[ACCEPT]").
		SetSelectedFunc(func() {
			if err := d.engine.AcceptAncientMemory(); err != nil {
				d.engine.AddLog("error", "Could not recover the memory: "+err.Error())
			}
			d.closeAncientMemoryModal()
		})
	btnAccept.SetBackgroundColor(theme.Color(theme.RolePositive))
	btnAccept.SetLabelColor(theme.Color(theme.RoleText))

	btnDecline := tview.NewButton("[Decline]").
		SetSelectedFunc(func() {
			d.engine.DeclineAncientMemory()
			d.closeAncientMemoryModal()
		})

	btnRow := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(btnAccept, 12, 0, true).
		AddItem(tview.NewBox(), 2, 0, false).
		AddItem(btnDecline, 12, 0, false)

	inner := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header, 3, 0, false).
		AddItem(tview.NewBox(), 1, 0, false).
		AddItem(body, 7, 0, false).
		AddItem(tview.NewBox(), 1, 0, false).
		AddItem(btnRow, 1, 0, true)
	inner.SetBorder(true).
		SetTitle(" ✦ Ancient Memory ").
		SetTitleColor(theme.Color(theme.RoleAccent)).
		SetBorderColor(theme.Color(theme.RoleAccent))

	modal := tview.NewFlex().
		AddItem(tview.NewBox(), 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(tview.NewBox(), 0, 1, false).
			AddItem(inner, 17, 0, true).
			AddItem(tview.NewBox(), 0, 1, false),
			64, 0, true).
		AddItem(tview.NewBox(), 0, 1, false)

	// Tab/arrows cycle the buttons; a/d are shortcuts for Accept/Decline.
	focusOrder := []tview.Primitive{btnAccept, btnDecline}
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
			case 'a', 'A':
				d.app.SetFocus(btnAccept)
				btnAccept.InputHandler()(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), nil)
				return nil
			case 'd', 'D':
				d.app.SetFocus(btnDecline)
				btnDecline.InputHandler()(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), nil)
				return nil
			}
		}
		return event
	})

	d.pages.AddPage(ancientMemoryPage, modal, true, true)
	d.app.SetFocus(btnAccept)
}

// closeAncientMemoryModal removes the Ancient Memory overlay page. The engine
// clears PendingMemoryTech on Accept/Decline, so the refresh loop will reset
// memoryModalShown on the next tick and won't re-pop it.
func (d *Dashboard) closeAncientMemoryModal() {
	d.pages.RemovePage(ancientMemoryPage)
}
