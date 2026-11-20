package ui

import (
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/rowid"
	"github.com/espresso20/ageforge/window/kit"
)

// tradePage is the Trade panel: the routes running, the routes that could
// be, and the allied bonuses that change trade income.
func tradePage(state game.GameState) deckPage {
	trade := state.Trade
	pg := deckPage{
		Title: rowid.Line("ui.overlay_trade.tradePage.01", "Trade routes"),
		Hint:  rowid.Line("ui.overlay_trade.tradePage.02", "trade route start/stop <key>"),
	}
	// The black market's lock is said on its own stop, not here.
	pg.Notes = append(pg.Notes, noteLinesOf(lockNotes(state, config.FeatureTradeRoutes, config.FeatureRouteRailFreight, config.FeatureRouteWarpCommerce), xGold)...)

	// Disruption banner: a war or an embargo blockades these resources, and
	// any route importing one is suspended until the conflict ends.
	if len(trade.DisruptedResources) > 0 {
		names := make([]string, len(trade.DisruptedResources))
		for i, r := range trade.DisruptedResources {
			names[i] = game.ResourceName(r)
		}
		pg.Notes = append(pg.Notes,
			deckLine{Value: rowid.Sprintf("ui.overlay_trade.tradePage.03", "⚠ Routes importing %s are suspended (a war, or your embargo).", textfmt.List(names)), K: xBad},
			deckLine{Value: rowid.Line("ui.overlay_trade.tradePage.04", "Lift your embargo with: diplomacy neutral <civ>")})
	}

	routes := tradeRoutes(state)
	count := map[string]int{}
	for _, r := range routes {
		count[r.state]++
	}
	pg.Head = []xseg{
		{rowid.Line("ui.overlay_trade.tradePage.05", "Routes "), xDim},
		{rowid.Sprintf("ui.overlay_trade.tradePage.06", "%d running", count["running"]+count["short"]+count["suspended"]), xGood}, {" · ", xFaint},
		{rowid.Sprintf("ui.overlay_trade.tradePage.07", "%d open", count["ready"]), xText}, {" · ", xFaint},
		{rowid.Sprintf("ui.overlay_trade.tradePage.08", "%d locked", count["locked"]), xDim},
	}
	pg.Notes = append(pg.Notes, allyTradeNotes(state.Diplomacy)...)

	// The cards are the routes in the order the board lists them, then the
	// black market: the board draws them from the same numbers.
	sec := deckSection{Title: rowid.Line("ui.overlay_trade.tradePage.17", "Open routes"), MinW: 46}
	for _, r := range routes {
		word, _, plate := tradeLook(r.state)
		c := deckCard{Title: r.name, Chip: word, ChipK: plate,
			Lines: []deckLine{{rowid.Line("ui.overlay_trade.tradePage.18", "Export"), tradeJoin(r.out, false, -1, state.TickIntervalMs), 0}, {rowid.Line("ui.overlay_trade.tradePage.19", "Import"), tradeJoin(r.in, false, 1, state.TickIntervalMs), xGood}},
			Why:   r.why, Quiet: r.state == "locked"}
		if r.cmd != "" {
			c.Command, c.Action = r.cmd, rowid.Line("ui.trade_board.tradePage.01", "stop")
			c.Danger = r.state != "ready" && r.state != "paused"
			if r.state == "ready" || r.state == "paused" {
				c.Action = rowid.Line("ui.trade_board.tradePage.02", "start")
			}
		}
		if r.moving() {
			c.Meter, c.Share = true, max(0, min(1, r.run))
			c.MeterText = rowid.Sprintf("ui.overlay_trade.tradePage.16", "running at %.0f%% of its flow", r.run*100)
		}
		sec.Cards = append(sec.Cards, c)
	}
	if len(sec.Cards) > 0 {
		pg.Sections = append(pg.Sections, sec)
	} else {
		pg.Notes = append(pg.Notes, deckLine{Value: rowid.Line("ui.overlay_trade.tradePage.24", "No trade routes in this age yet.")})
	}
	if sec, ok := blackMarketSection(state); ok {
		pg.Sections = append(pg.Sections, sec)
	}
	if sec, ok := tradeOffersSection(state, routes); ok {
		pg.Sections = append(pg.Sections, sec)
	}
	return pg
}

// tradeOffersSection is the offers as cards, one for each, in the order the
// board lists them.
func tradeOffersSection(state game.GameState, routes []tradeRoute) (deckSection, bool) {
	civs := tradeCivs(state, routes)
	open, pay := tradeOfferCounts(state, civs)
	sec := deckSection{Title: rowid.Line("ui.overlay_trade.tradeOffersSection.01", "Offers"), MinW: 46,
		Meta: rowid.Sprintf("ui.overlay_trade.tradeOffersSection.02", "%d open · %d you can pay now", open, pay)}
	for _, c := range civs {
		for _, d := range c.offers {
			sec.Cards = append(sec.Cards, deckCard{
				Title:  c.name + " · " + game.DealKindLabel(d.Kind),
				Lines:  []deckLine{{rowid.Line("ui.overlay_trade.tradeOffersSection.03", "You give"), tradeOfferGive(d), 0}, {rowid.Line("ui.overlay_trade.tradeOffersSection.04", "You get"), tradeOfferGet(d), xGood}},
				Action: rowid.Line("ui.overlay_trade.tradeOffersSection.05", "accept"), Command: tradeOfferCmd(c.key, d),
				Quiet: !tradePayable(state, d)})
		}
	}
	return sec, len(sec.Cards) > 0
}

// noteLinesOf is a block of tagged text, a line to a note, in colour k:
// the tags and the blank lines are dropped.
func noteLinesOf(tagged string, k xk) []deckLine {
	var out []deckLine
	for _, l := range strings.Split(kit.Strip(tagged), "\n") {
		if l = strings.TrimSpace(l); l != "" {
			out = append(out, deckLine{Value: l, K: k})
		}
	}
	return out
}

// allyTradeNotes lists the allies currently boosting your income, the one
// piece of civilization state that belongs on the trade screen. It mirrors
// game.DiplomacyManager.GetTradeBonus: an allied civ that is not at war adds
// its TradeBonus to its specialty resource, both to route imports and to
// that resource's production rate. It is empty when no ally applies.
func allyTradeNotes(dip game.DiplomacyState) []deckLine {
	keys := keysWhere(dip.Factions, func(_ string, f game.FactionInfo) bool {
		return f.Discovered && f.Status == "allied" && !f.AtWar && f.TradeBonus > 0
	})
	var out []deckLine
	for i, k := range keys {
		f := dip.Factions[k]
		head := ""
		if i == 0 {
			head = rowid.Line("ui.overlay_trade.allyTradeNotes.01", "Allied bonuses: ")
		}
		out = append(out, deckLine{Value: head + rowid.Sprintf("ui.overlay_trade.allyTradeNotes.02", "%s: %s %s (route imports and production)",
			f.Name, textfmt.SignedPercent(f.TradeBonus), game.ResourceName(f.Specialty)), K: xGood})
	}
	return out
}
