package ui

import (
	"sort"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/vibe"
	"github.com/espresso20/ageforge/window/kit"
)

// tradePage is the Trade panel: the routes running, the routes that could
// be, and the allied bonuses that change trade income.
func tradePage(state game.GameState) deckPage {
	trade := state.Trade
	pg := deckPage{
		Title: vibe.Line("ui.overlay_trade.tradePage.01", "Trade routes"),
		Hint:  vibe.Line("ui.overlay_trade.tradePage.02", "trade route start/stop <key>"),
	}
	pg.Notes = append(pg.Notes, noteLinesOf(lockNotes(state, config.FeatureTradeRoutes, config.FeatureRouteRailFreight, config.FeatureRouteWarpCommerce, config.FeatureBlackMarket), xGold)...)

	// Disruption banner: a war or an embargo blockades these resources, and
	// any route importing one is suspended until the conflict ends.
	if len(trade.DisruptedResources) > 0 {
		names := make([]string, len(trade.DisruptedResources))
		for i, r := range trade.DisruptedResources {
			names[i] = game.ResourceName(r)
		}
		pg.Notes = append(pg.Notes,
			deckLine{Value: vibe.Sprintf("ui.overlay_trade.tradePage.03", "⚠ Routes importing %s are suspended (a war, or your embargo).", textfmt.List(names)), K: xBad},
			deckLine{Value: vibe.Line("ui.overlay_trade.tradePage.04", "Lift your embargo with: diplomacy neutral <civ>")})
	}

	// With trade routes themselves still locked, no route can start: its
	// card says why instead of offering a button the game would refuse.
	refusal, routesLocked := featureRefusal(state, config.FeatureTradeRoutes)
	open, locked := 0, 0
	for _, r := range trade.AvailableRoutes {
		if r.CanStart && !routesLocked {
			open++
		} else {
			locked++
		}
	}
	pg.Head = []xseg{
		{vibe.Line("ui.overlay_trade.tradePage.05", "Routes "), xDim},
		{vibe.Sprintf("ui.overlay_trade.tradePage.06", "%d running", len(trade.ActiveRoutes)), xGood}, {" · ", xFaint},
		{vibe.Sprintf("ui.overlay_trade.tradePage.07", "%d open", open), xText}, {" · ", xFaint},
		{vibe.Sprintf("ui.overlay_trade.tradePage.08", "%d locked", locked), xDim},
	}
	pg.Notes = append(pg.Notes, allyTradeNotes(state.Diplomacy)...)

	if len(trade.ActiveRoutes) > 0 {
		// The snapshot builds this list from a map; sort it so the panel does
		// not reshuffle on every refresh.
		active := append([]game.ActiveRouteInfo(nil), trade.ActiveRoutes...)
		sort.Slice(active, func(i, j int) bool { return active[i].Key < active[j].Key })
		sec := deckSection{Title: vibe.Line("ui.overlay_trade.tradePage.09", "Running"), MinW: 50}
		for _, route := range active {
			c := deckCard{
				Title: route.Name, Chip: vibe.Line("ui.overlay_trade.tradePage.10", "running"), ChipK: xChipShort,
				Lines:  []deckLine{{vibe.Line("ui.overlay_trade.tradePage.11", "Export"), game.Amounts(route.Export), 0}, {vibe.Line("ui.overlay_trade.tradePage.12", "Import"), game.Amounts(route.Import), xGood}},
				Action: "stop", Command: "trade route stop " + route.Key,
			}
			if route.Disrupted {
				c.Chip, c.ChipK = vibe.Line("ui.overlay_trade.tradePage.14", "suspended"), xChipMil
				c.Lines[1].K = xBad
				c.Why = vibe.Sprintf("ui.overlay_trade.tradePage.15", "Suspended: %s imports are blockaded", game.ResourceName(route.DisruptedBy))
			} else {
				c.Meter = true
				if route.CycleTicks > 0 {
					c.Share = max(0, min(1, 1-float64(route.TicksLeft)/float64(route.CycleTicks)))
				}
				c.MeterText = vibe.Sprintf("ui.overlay_trade.tradePage.16", "%s left in this cycle (%s done)",
					formatTicks(route.TicksLeft, state), textfmt.Count(route.CyclesDone, txtCycle.String(), txtCycles.String()))
			}
			sec.Cards = append(sec.Cards, c)
		}
		pg.Sections = append(pg.Sections, sec)
	}

	if len(trade.AvailableRoutes) > 0 {
		sec := deckSection{Title: vibe.Line("ui.overlay_trade.tradePage.17", "Open routes"), MinW: 46}
		for _, route := range trade.AvailableRoutes {
			c := deckCard{
				Title: route.Name, Desc: route.Description,
				Lines: []deckLine{{vibe.Line("ui.overlay_trade.tradePage.18", "Export"), game.Amounts(route.Export), 0}, {vibe.Line("ui.overlay_trade.tradePage.19", "Import"), game.Amounts(route.Import), xGood}},
			}
			if route.CanStart && !routesLocked {
				c.Chip, c.ChipK = vibe.Line("ui.overlay_trade.tradePage.20", "ready"), xChipCivic
				c.Action, c.Command = "start", "trade route start "+route.Key
			} else {
				c.Chip, c.ChipK, c.Quiet = vibe.Line("ui.overlay_trade.tradePage.22", "locked"), xKeyName, true
				c.Why = refusal
				if !route.CanStart {
					c.Why = vibe.Sprintf("ui.overlay_trade.tradePage.23", "needs %s (have %s)",
						game.BuildingCount(route.MinCount, route.RequiredBld),
						textfmt.Int(max(state.Buildings[route.RequiredBld].Count, state.Buildings[route.RequiredBld].HeldWithLater)))
				}
			}
			sec.Cards = append(sec.Cards, c)
		}
		pg.Sections = append(pg.Sections, sec)
	}

	if len(trade.ActiveRoutes) == 0 && len(trade.AvailableRoutes) == 0 {
		pg.Notes = append(pg.Notes, deckLine{Value: vibe.Line("ui.overlay_trade.tradePage.24", "No trade routes in this age yet.")})
	}

	// Civilization opinion lives on the Factions panel only. This panel used
	// to carry its own copy, which went stale.
	pg.Sections = append(pg.Sections, deckSection{Title: vibe.Line("ui.overlay_trade.tradePage.25", "Factions"), Cards: []deckCard{{
		Title: vibe.Line("ui.overlay_trade.tradePage.26", "Factions"), Desc: vibe.Line("ui.overlay_trade.tradePage.27", "Opinion and deals with other civilizations: type factions"),
		Action: vibe.Line("ui.overlay_trade.tradePage.28", "open"), Command: "factions",
	}}})
	return pg
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
			head = vibe.Line("ui.overlay_trade.allyTradeNotes.01", "Allied bonuses: ")
		}
		out = append(out, deckLine{Value: head + vibe.Sprintf("ui.overlay_trade.allyTradeNotes.02", "%s: %s %s (route imports and production)",
			f.Name, textfmt.SignedPercent(f.TradeBonus), game.ResourceName(f.Specialty)), K: xGood})
	}
	return out
}
