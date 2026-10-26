package ui

import (
	"strings"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/vibe"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const prestigeCuePage = "prestige_cue"

func prestigeCueText(state game.GameState, width int) string {
	cue := state.PrestigeCue
	iv := tickInterval(state)
	lines := []string{
		vibe.Line("ui.prestige_cue_notice.prestigeCueText.01", "Time to prestige?"),
		vibe.Sprintf("ui.prestige_cue_notice.prestigeCueText.02", "Finishing the %s takes about %s at today's income.", game.AgeName(state.Age), game.DurationText(cue.FinishTicks, iv)),
		vibe.Sprintf("ui.prestige_cue_notice.prestigeCueText.03", "A new run with x%.2f production would be back here and past it in about %s.", cue.NextMultiplier, game.DurationText(cue.FreshTicks, iv)),
		vibe.Line("ui.prestige_cue_notice.prestigeCueText.04", "Type prestige to see what you keep."),
	}
	max := width*17/20 - 4 // the notice uses the overlay's 85% width, less its frame
	if max < 20 {
		max = 20
	}
	var wrapped []string
	for _, line := range lines {
		wrapped = append(wrapped, wrapWords(line, max)...)
	}
	return strings.Join(wrapped, "\n")
}

func shouldShowPrestigeCue(state game.GameState, blocked bool) bool {
	return state.PrestigeCue.Ready && !state.PrestigeCue.Seen && !blocked
}

func (d *Dashboard) maybeShowPrestigeCue(state game.GameState) {
	blocked := d.overlayMgr.HasActive() || d.overlayMgr.FullScreenUp() ||
		d.pages.HasPage(catastrophePage) || d.pages.HasPage(ancientMemoryPage) ||
		d.pages.HasPage(prestigeCuePage) || d.pages.HasPage(arrivalPageName) ||
		// Any key closes the notice, so over a half-typed command it would
		// swallow a keystroke and leave a command nobody typed. It waits
		// for the prompt to be empty.
		d.inputField.GetText() != ""
	if !shouldShowPrestigeCue(state, blocked) {
		return
	}

	width := d.overlayMgr.screenW
	tv := tview.NewTextView().SetText(prestigeCueText(state, width)).SetWrap(false)
	tv.SetBorder(true)
	theme.Track(func() {
		tv.SetBorderColor(theme.Color(theme.RoleAccent)).
			SetTitleColor(theme.Color(theme.RoleAccent)).
			SetBackgroundColor(theme.Color(theme.RoleSurface))
	})
	inner := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(tv, 8, 0, true).
		AddItem(nil, 0, 1, false)
	root := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(inner, 0, 17, true).
		AddItem(nil, 0, 1, false)
	root.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		d.pages.RemovePage(prestigeCuePage)
		d.app.SetFocus(d.inputField)
		return nil
	})
	d.pages.AddPage(prestigeCuePage, root, true, true)
	d.app.SetFocus(root)
	// This runs in Dashboard.refresh on the tview goroutine, outside engine callbacks.
	d.engine.MarkPrestigeCueSeen()
}
