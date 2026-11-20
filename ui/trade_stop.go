package ui

import (
	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/rowid"
)

// trade_stop.go is the smugglers' stop on the Trade page: a card for the
// black market with what a run costs, the odds, whether one can be made now
// and a chip for each resource a win would pay in.

// blackMarketSection is the page's section for the stop; it has none before
// the black market opens.
func blackMarketSection(state game.GameState) (deckSection, bool) {
	bm := state.BlackMarket
	if !bm.Available {
		return deckSection{}, false
	}
	c := deckCard{
		Title: rowid.Line("ui.trade_stop.blackMarketSection.02", "Black market"), Stop: true,
		Lines: []deckLine{
			{rowid.Line("ui.trade_stop.blackMarketSection.03", "Cost"), game.Amount(bm.Cost, "culture"), 0},
			{textfmt.Capitalize(game.ResourceName("culture")), textfmt.Number(bm.Culture), xGood},
			{rowid.Line("ui.trade_stop.blackMarketSection.04", "Odds"), rowid.Line("ui.trade_stop.blackMarketSection.05", "five in six"), 0},
		},
	}
	sec := deckSection{Title: rowid.Line("ui.trade_stop.blackMarketSection.01", "Smugglers"), MinW: 46}
	if refusal, locked := featureRefusal(state, config.FeatureBlackMarket); locked {
		c.Quiet, c.Why, c.Lines = true, refusal, nil
		sec.Cards = []deckCard{c}
		return sec, true
	}
	short := bm.Culture < bm.Cost
	switch {
	case !bm.Ready:
		c.Chip, c.ChipK = rowid.Line("ui.trade_stop.blackMarketSection.06", "lying low"), xChipMil
		c.Meter = true
		if bm.CooldownTicks > 0 {
			c.Share = max(0, min(1, 1-float64(bm.CooldownLeft)/float64(bm.CooldownTicks)))
		}
		c.MeterText = rowid.Sprintf("ui.trade_stop.blackMarketSection.07", "%s left", formatTicks(bm.CooldownLeft, state))
	case short:
		c.Chip, c.ChipK = rowid.Line("ui.trade_stop.blackMarketSection.08", "not enough culture"), xChipMil
		c.Lines[1].K = xBad
	default:
		c.Chip, c.ChipK = rowid.Line("ui.trade_stop.blackMarketSection.09", "ready"), xChipCivic
	}
	for _, p := range bm.Payouts {
		c.Chips = append(c.Chips, deckChip{Word: game.Amount(p.Amount, p.Resource), Command: "blackmarket " + p.Resource, Off: !bm.Ready || short})
	}
	sec.Cards = []deckCard{c}
	return sec, true
}
