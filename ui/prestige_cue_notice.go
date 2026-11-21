package ui

import (
	"strings"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/rowid"
)

// prestigeCueCard is the card shown once in an age when a prestige would
// pass the age's gate sooner than playing on: the two times, what a prestige
// gives, takes and keeps, and a button for the full account.
func prestigeCueCard(state game.GameState) momentCard {
	cue := state.PrestigeCue
	iv := tickInterval(state)
	lines := []string{
		rowid.Sprintf("ui.prestige_cue_notice.prestigeCueText.02", "Finishing the %s takes about %s at today's income.", game.AgeName(state.Age), game.DurationText(cue.FinishTicks, iv)),
		rowid.Sprintf("ui.prestige_cue_notice.prestigeCueText.03", "A new run with x%.2f production would be back here and past it in about %s.", cue.NextMultiplier, game.DurationText(cue.FreshTicks, iv)),
		"",
	}
	lines = append(lines, prestigeTradeLines(state)...)
	lines = append(lines, "", rowid.Line("ui.prestige_cue_notice.prestigeCueText.04", "Type prestige to see what you keep."))
	return momentCard{
		title: rowid.Line("ui.prestige_cue_notice.prestigeCueText.01", "Time to prestige?"),
		body:  strings.Join(lines, "\n"),
		btns: []momentBtn{
			{label: rowid.Line("ui.prestige_cue_notice.prestigeCueCard.01", "Show me"), line: "prestige", kind: keyAccent},
			{label: rowid.Line("ui.prestige_cue_notice.prestigeCueCard.02", "Later"), kind: keySteel},
		},
	}
}

// maybeShowPrestigeCue opens the prestige cue, once in an age.
func (d *Dashboard) maybeShowPrestigeCue(state game.GameState) {
	if !state.PrestigeCue.Ready || state.PrestigeCue.Seen || !d.momentClear() {
		return
	}
	d.showMoment(prestigeCueCard(state))
	// This runs in Dashboard.refresh on the UI goroutine, outside engine callbacks.
	d.engine.MarkPrestigeCueSeen()
}
