package ui

import (
	"fmt"

	"github.com/rivo/tview"
)

// saveChoicePage is the unique page name for the Overwrite/Branch save prompt.
const saveChoicePage = "save_choice"

// branchNamePage is the page name for the Branch name-entry modal. It is
// distinct from newGameNamePage so the two name modals can't collide.
const branchNamePage = "branch_name"

// showSaveChoiceModal pops the Overwrite-vs-Branch prompt a bare `save` triggers.
// Overwrite writes the current run to its active slot. Branch forks a new save
// (suggested name, editable) whose parent is the current save; autosave then
// follows the new branch, leaving the old save frozen at the branch point.
// Cancel dismisses without saving. Every exit path refocuses the input field.
func (d *Dashboard) showSaveChoiceModal() {
	active := d.engine.ActiveSaveName()

	modal := tview.NewModal().
		SetText(fmt.Sprintf("Save %q?", active)).
		AddButtons([]string{"Overwrite", "Branch new", "Cancel"}).
		SetDoneFunc(func(_ int, label string) {
			d.pages.RemovePage(saveChoicePage)
			switch label {
			case "Overwrite":
				name := d.engine.ActiveSaveName()
				if err := d.engine.SaveGame(name); err != nil {
					d.engine.AddLog("error", err.Error())
				} else {
					d.engine.AddLog("info", "[lime]Saved → "+name)
				}
				d.app.SetFocus(d.inputField)
			case "Branch new":
				showSaveNameModal(d.app, d.pages, " Branch a New Save ", branchNamePage, d.inputField, func(name string) {
					// showSaveNameModal validates the name format; BranchSave guards
					// against an existing name and logs an error if it's taken.
					if err := d.engine.BranchSave(name); err != nil {
						d.engine.AddLog("error", err.Error())
					} else {
						d.engine.AddLog("info", "[lime]Branched → "+name+" (autosave now follows it)")
					}
					d.app.SetFocus(d.inputField)
				})
			default: // Cancel (or Esc)
				d.app.SetFocus(d.inputField)
			}
		})

	d.pages.AddPage(saveChoicePage, modal, true, true)
	d.app.SetFocus(modal)
}
