package ui

import (
	"image"
	"strconv"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/vibe"
)

// faction_verbs.go is what the Factions panel does by click: each civilization
// met carries its diplomacy verbs as buttons under its card, and each trade
// offer carries Accept. Every button is an xkey that carries the command a
// player would type. A verb the engine would refuse is drawn off.

const (
	// diplomacyCmd is the command every verb starts with.
	diplomacyCmd = "diplomacy "
	// factionVerbIndent is the column a card's buttons start at, under its text.
	factionVerbIndent = 3
)

// The words the buttons say.
var (
	txtVerbGift    = vibe.E("ui.faction_verbs.gift.01", "Gift")
	txtVerbAlly    = vibe.E("ui.faction_verbs.ally.01", "Ally")
	txtVerbRival   = vibe.E("ui.faction_verbs.rival.01", "Rival")
	txtVerbEmbargo = vibe.E("ui.faction_verbs.embargo.01", "Embargo")
	txtVerbNeutral = vibe.E("ui.faction_verbs.neutral.01", "Neutral")
	txtVerbTribute = vibe.E("ui.faction_verbs.tribute.01", "Tribute")
	txtVerbRaid    = vibe.E("ui.faction_verbs.raid.01", "Raid")
	txtVerbAccept  = vibe.E("ui.faction_verbs.accept.01", "Accept")
)

// factionsView is the panel's buttons: the slots its text last reserved for
// them, and the text they sit on.
type factionsView struct {
	textButtons
	rec  *btnRec
	text string
}

// provider is the panel's text, with the rows its buttons will sit on kept.
func (fv *factionsView) provider(state game.GameState, w int) string {
	fv.rec = &btnRec{}
	fv.text = factionsProvider(state, w, fv.rec)
	return fv.text
}

// layer is the buttons laid over the text in view.
func (fv *factionsView) layer(text image.Rectangle, top int) *xgrid {
	return fv.rec.lay(fv.text, text, top)
}

// wire makes the panel's buttons work.
func (fv *factionsView) wire(om *OverlayManager) {
	fv.attach(om, "factions", fv.layer)
}

// raidAsk is the question a Raid asks, naming the civilization.
func raidAsk(name string) string {
	return vibe.Sprintf("ui.faction_verbs.raidAsk.01", "Raid the %s trade route?\n\nTheir opinion falls at once, and it may start a war.", name)
}

// factionVerbs is a civilization's buttons: Gift, the most used, leads, and
// Tribute takes its place while the civ is at war. A verb the engine would
// refuse is off.
func factionVerbs(state game.GameState, def config.FactionDef, f game.FactionInfo) []actionSpec {
	locked := false
	if _, l := featureRefusal(state, config.FeatureDiplomacy); l {
		locked = true
	}
	gold := state.Resources["gold"].Amount
	culture := state.Resources["culture"].Amount
	tributeGold, tributeCulture := game.TributePrice(def)
	dip := state.Diplomacy
	verb := func(label, sub string, kind keyKind, off bool) actionSpec {
		return actionSpec{label: label, kind: kind, cmd: diplomacyCmd + sub + " " + def.Key, off: off}
	}
	gift, tribute := keyPrimary, keyPlain
	if f.AtWar {
		gift, tribute = keyPlain, keyPrimary
	}
	raid := verb(txtVerbRaid.String(), "raid", keyDanger, false)
	raid.ask, raid.yes = raidAsk(def.Name), raid.label
	return []actionSpec{
		verb(txtVerbGift.String(), "gift", gift, locked || gold < dip.GiftCost),
		verb(txtVerbAlly.String(), "ally", keyPlain, locked || f.Status == "allied" || f.Opinion < game.AllyOpinion || gold < dip.AllyCost),
		verb(txtVerbRival.String(), "rival", keyPlain, locked || f.Status == "rival"),
		verb(txtVerbEmbargo.String(), "embargo", keyPlain, locked || f.Status == "embargo"),
		verb(txtVerbNeutral.String(), "neutral", keyPlain, f.Status == "neutral"),
		verb(txtVerbTribute.String(), "tribute", tribute, !f.AtWar || gold < tributeGold || culture < tributeCulture),
		raid,
	}
}

// dealButton is Accept for one offer of the civilization key, off when it is
// taken, cannot be paid or the feature is still locked.
func dealButton(state game.GameState, key string, d game.DealInfo) actionSpec {
	_, locked := featureRefusal(state, config.FeatureDiplomacy)
	off := locked || d.Taken || state.Resources[d.Give].Amount < d.GiveAmt
	cmd := strings.Join([]string{diplomacyCmd + "accept", key, strconv.Itoa(d.Num)}, " ")
	return actionSpec{label: txtVerbAccept.String(), kind: keySteel, cmd: cmd, off: off}
}
