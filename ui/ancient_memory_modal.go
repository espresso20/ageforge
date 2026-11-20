package ui

import (
	"fmt"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/rowid"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/window/kit"
)

// ancientMemoryPage is the page name for the Ancient Memory offer overlay.
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

	header := newNativeText().
		SetDynamicColors(true).
		SetTextAlign(kit.AlignCenter).
		SetText(rowid.Line("ui.ancient_memory_modal.showAncientMemoryModal.01", "[gray]You have discovered an old cache. It appears to contain\n"+
			"memories of a now-extinct civilization.[-]"))

	body := newNativeText().
		SetDynamicColors(true).
		SetText(rowid.Sprintf("ui.ancient_memory_modal.showAncientMemoryModal.02", "[white]── Recovered memory: [-][cyan]%s[-]\n", techName) +
			fmt.Sprintf("  [gray]%s[-]\n\n", def.Description) +
			rowid.Line("ui.ancient_memory_modal.showAncientMemoryModal.03", "  [green]✓ Research it now, free of prerequisites[-]\n") +
			rowid.Line("ui.ancient_memory_modal.showAncientMemoryModal.04", "  [red]• It takes twice as long to research[-]\n") +
			rowid.Line("ui.ancient_memory_modal.showAncientMemoryModal.05", "  [gray]Only one such memory surfaces per civilization.[-]"))

	acceptKey, acceptLabel := splitKeyed(rowid.Line("ui.ancient_memory_modal.showAncientMemoryModal.06", "[A[] Accept"))
	declineKey, declineLabel := splitKeyed(rowid.Line("ui.ancient_memory_modal.showAncientMemoryModal.08", "[D[] Decline"))
	btnRow, btns := choiceRow([]choice{
		{label: acceptLabel, key: acceptKey, kind: keyPrimary, sign: signTick, press: func() {
			if err := d.engine.AcceptAncientMemory(); err != nil {
				d.engine.AddLog(game.OfficeGame, game.LevelEverything, "error", rowid.Line("ui.ancient_memory_modal.showAncientMemoryModal.07", "Could not recover the memory: ")+err.Error())
			}
			d.closeAncientMemoryModal()
		}},
		{label: declineLabel, key: declineKey, kind: keyPlain, sign: signCross, press: func() {
			d.engine.DeclineAncientMemory()
			d.closeAncientMemoryModal()
		}},
	}, 2, 0)
	btnAccept, btnDecline := btns[0], btns[1]

	inner := newNativeFlex().SetDirection(flexRow).
		AddItem(header, 3, 0, false).
		AddItem(newNativeBox(), 1, 0, false).
		AddItem(body, 7, 0, false).
		AddItem(newNativeBox(), 1, 0, false).
		AddItem(btnRow, actionRows, 0, true)
	framed(inner, rowid.Line("ui.ancient_memory_modal.showAncientMemoryModal.09", " ✦ Ancient Memory "), theme.RoleAccent)

	modal := centeredModal(inner, 64, 14+actionRows)

	// Tab/arrows cycle the buttons; a/d are shortcuts for Accept/Decline.
	ring := newButtonRing(d.app, func(p Primitive) { d.memFocus = p })
	ring.add(btnAccept, 'a')
	ring.add(btnDecline, 'd')
	modal.SetInputCapture(func(event *kit.Key) *kit.Key {
		if ring.key(event) {
			return nil
		}
		return event
	})

	d.pages.AddPage(ancientMemoryPage, modal, true, true)
	ring.focus(0)
}

// closeAncientMemoryModal removes the Ancient Memory overlay page and hands
// the keyboard back (returnFocus). The engine clears PendingMemoryTech on
// Accept/Decline, so the refresh loop will reset memoryModalShown on the next
// tick and won't re-pop it.
func (d *Dashboard) closeAncientMemoryModal() {
	d.pages.RemovePage(ancientMemoryPage)
	d.memFocus = nil
	d.returnFocus()
}
