package ui

import (
	"github.com/espresso20/ageforge/vibe"
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
		SetText(vibe.Sprintf("ui.save_modal.showSaveChoiceModal.01", "Save %q?", active)).
		AddButtons([]string{vibe.Line("ui.save_modal.showSaveChoiceModal.02", "Overwrite"), vibe.Line("ui.save_modal.showSaveChoiceModal.03", "Branch new"), vibe.Line("ui.save_modal.showSaveChoiceModal.04", "Cancel")}).
		SetDoneFunc(func(button int, _ string) {
			d.pages.RemovePage(saveChoicePage)
			switch button {
			case 0: // "Overwrite"
				name := d.engine.ActiveSaveName()
				if err := d.engine.SaveGame(name); err != nil {
					d.engine.AddLog("error", err.Error())
				} else {
					d.engine.AddLog("info", vibe.Line("ui.save_modal.showSaveChoiceModal.05", "[positive]Saved → ")+name)
				}
				d.app.SetFocus(d.inputField)
			case 1: // "Branch new"
				showSaveNameModal(d.app, d.pages, vibe.Line("ui.save_modal.showSaveChoiceModal.06", " Branch a new save "), branchNamePage, d.inputField, func(name string) {
					// showSaveNameModal validates the name format; BranchSave guards
					// against an existing name and logs an error if it's taken.
					if err := d.engine.BranchSave(name); err != nil {
						d.engine.AddLog("error", err.Error())
					} else {
						d.engine.AddLog("info", vibe.Line("ui.save_modal.showSaveChoiceModal.07", "[positive]Branched → ")+name+vibe.Line("ui.save_modal.showSaveChoiceModal.08", " (autosave now follows it)"))
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
