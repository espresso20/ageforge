package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/vibe"
)

// tradeProvider renders the Trade routes panel: the routes, and the allied
// bonuses that change trade income.
func tradeProvider(state game.GameState, _ int) string {
	var sb strings.Builder
	trade := state.Trade

	sb.WriteString(vibe.Line("ui.overlay_trade.tradeProvider.01", " [gold]═══ Trade routes ═══[-]\n\n"))
	sb.WriteString(lockNotes(state, config.FeatureTradeRoutes, config.FeatureRouteRailFreight, config.FeatureRouteWarpCommerce, config.FeatureBlackMarket))

	// Disruption banner: a war or an embargo blockades these resources, and
	// any route importing one is suspended until the conflict ends.
	if len(trade.DisruptedResources) > 0 {
		names := make([]string, len(trade.DisruptedResources))
		for i, r := range trade.DisruptedResources {
			names[i] = game.ResourceName(r)
		}
		fmt.Fprint(&sb, vibe.Sprintf("ui.overlay_trade.tradeProvider.02", " [red]⚠ Routes importing %s are suspended (a war, or your embargo).[-]\n", textfmt.List(names)))
		sb.WriteString(vibe.Line("ui.overlay_trade.tradeProvider.03", " [gray]Lift your embargo with: diplomacy neutral <civ>[-]\n\n"))
	}

	if len(trade.ActiveRoutes) > 0 {
		sb.WriteString(vibe.Line("ui.overlay_trade.tradeProvider.04", " [gold]Active routes:[-]\n\n"))
		// The snapshot builds this list from a map; sort it so the panel does
		// not reshuffle on every refresh.
		active := append([]game.ActiveRouteInfo(nil), trade.ActiveRoutes...)
		sort.Slice(active, func(i, j int) bool { return active[i].Key < active[j].Key })
		for _, route := range active {
			marker := "[green]▸[-]"
			if route.Disrupted {
				marker = "[red]✖[-]"
			}
			fmt.Fprintf(&sb, " %s [cyan]%s[-]\n", marker, route.Name)
			fmt.Fprint(&sb, vibe.Sprintf("ui.overlay_trade.tradeProvider.05", "   Export: %s\n", game.Amounts(route.Export)))
			fmt.Fprint(&sb, vibe.Sprintf("ui.overlay_trade.tradeProvider.06", "   Import: %s\n", game.Amounts(route.Import)))
			if route.Disrupted {
				fmt.Fprint(&sb, vibe.Sprintf("ui.overlay_trade.tradeProvider.07", "   [red]Suspended: %s imports are blockaded[-]\n\n", game.ResourceName(route.DisruptedBy)))
			} else {
				fmt.Fprint(&sb, vibe.Sprintf("ui.overlay_trade.tradeProvider.08", "   %s left in this cycle  [gray](%s done)[-]\n\n",
					formatTicks(route.TicksLeft, state), textfmt.Count(route.CyclesDone, txtCycle.String(), txtCycles.String())))
			}
		}
	}

	if len(trade.AvailableRoutes) > 0 {
		sb.WriteString(vibe.Line("ui.overlay_trade.tradeProvider.11", " [gold]Available routes:[-]\n\n"))
		for _, route := range trade.AvailableRoutes {
			statusIcon := "[red]✗[-]"
			if route.CanStart {
				statusIcon = "[green]✓[-]"
			}
			fmt.Fprintf(&sb, " %s [cyan]%s[-]\n", statusIcon, route.Name)
			fmt.Fprintf(&sb, "   [gray]%s[-]\n", route.Description)
			fmt.Fprint(&sb, vibe.Sprintf("ui.overlay_trade.tradeProvider.12", "   Export: %s → Import: %s\n",
				game.Amounts(route.Export), game.Amounts(route.Import)))
			if route.CanStart {
				fmt.Fprint(&sb, vibe.Sprintf("ui.overlay_trade.tradeProvider.13", "   [green]trade route start %s[-]\n", route.Key))
			} else {
				fmt.Fprint(&sb, vibe.Sprintf("ui.overlay_trade.tradeProvider.14", "   [red]needs %s (have %s)[-]\n",
					game.BuildingCount(route.MinCount, route.RequiredBld),
					textfmt.Int(max(state.Buildings[route.RequiredBld].Count, state.Buildings[route.RequiredBld].HeldWithLater))))
			}
			sb.WriteString("\n")
		}
	}

	if len(trade.ActiveRoutes) == 0 && len(trade.AvailableRoutes) == 0 {
		sb.WriteString(vibe.Line("ui.overlay_trade.tradeProvider.15", " [gray]No trade routes in this age yet.[-]\n"))
	}

	writeAllyTradeBonuses(&sb, state.Diplomacy)

	sb.WriteString(vibe.Line("ui.overlay_trade.tradeProvider.16", " [gray]Commands: trade route start/stop <key>[-]\n"))

	// Civilization opinion lives on the Factions panel only. This panel used
	// to carry its own copy, which went stale.
	sb.WriteString(vibe.Line("ui.overlay_trade.tradeProvider.17", "\n [gray]Opinion and deals with other civilizations: type factions[-]\n"))

	return sb.String()
}

// writeAllyTradeBonuses lists the allies currently boosting your income, the
// one piece of civilization state that belongs on the trade screen. It
// mirrors game.DiplomacyManager.GetTradeBonus: an allied civ that is not at
// war adds its TradeBonus to its specialty resource, both to route imports
// and to that resource's production rate. Writes nothing when no ally
// applies.
func writeAllyTradeBonuses(sb *strings.Builder, dip game.DiplomacyState) {
	keys := keysWhere(dip.Factions, func(_ string, f game.FactionInfo) bool {
		return f.Discovered && f.Status == "allied" && !f.AtWar && f.TradeBonus > 0
	})
	if len(keys) == 0 {
		return
	}
	sb.WriteString(vibe.Line("ui.overlay_trade.writeAllyTradeBonuses.01", " [gold]Allied bonuses:[-]\n"))
	for _, k := range keys {
		f := dip.Factions[k]
		fmt.Fprint(sb, vibe.Sprintf("ui.overlay_trade.writeAllyTradeBonuses.02", "   %s: [green]%s %s[-] [gray](route imports and production)[-]\n",
			f.Name, textfmt.SignedPercent(f.TradeBonus), game.ResourceName(f.Specialty)))
	}
	sb.WriteString("\n")
}
