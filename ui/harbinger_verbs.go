package ui

import (
	"image"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/vibe"
)

// harbinger_verbs.go is what the Harbinger panel does by click: Appease and
// Brace sit under their cost lines, Invite under its description. Each button
// carries the command a player would type. One the engine would refuse is drawn
// off, and Invite asks first because it cannot be undone.

const (
	// harbingerCmd is the command every answer starts with.
	harbingerCmd = "harbinger "
	// harbingerVerbIndent is the column the buttons start at, under the text.
	harbingerVerbIndent = 5
)

// The words the buttons say.
var (
	txtVerbAppease = vibe.E("ui.harbinger_verbs.appease.01", "Appease")
	txtVerbBrace   = vibe.E("ui.harbinger_verbs.brace.01", "Brace")
	txtVerbInvite  = vibe.E("ui.harbinger_verbs.invite.01", "Invite")
)

// harbingerBuyable is whether state holds every resource in cost.
func harbingerBuyable(state game.GameState, cost map[string]float64) bool {
	if len(cost) == 0 {
		return false
	}
	for res, need := range cost {
		if state.Resources[res].Amount < need {
			return false
		}
	}
	return true
}

// harbingerPayBtn is the button for Appease or Brace.
func harbingerPayBtn(state game.GameState, word, sub, blocked string, cost map[string]float64) []slotBtn {
	return []slotBtn{{word: word, kind: keyPlain, cmd: harbingerCmd + sub,
		off: blocked != "" || !harbingerBuyable(state, cost)}}
}

// harbingerInviteBtn is the button for Invite, which asks first.
func harbingerInviteBtn(h *game.HarbingerView) []slotBtn {
	return []slotBtn{{word: txtVerbInvite.String(), kind: keyDanger, cmd: harbingerCmd + "invite",
		ask: vibe.Sprintf("ui.harbinger_verbs.inviteAsk.01", "Invite %s?\n\nThis cannot be undone.", h.Name),
		yes: txtVerbInvite.String(), off: h.InviteBlocked != ""}}
}

// harbingerView is the panel's buttons: the slots its text last reserved for
// them, and the text they sit on.
type harbingerView struct {
	textButtons
	rec  *btnRec
	text string
}

// layer is the buttons laid over the text in view.
func (hv *harbingerView) layer(text image.Rectangle, top int) *xgrid {
	return hv.rec.lay(hv.text, text, top)
}
