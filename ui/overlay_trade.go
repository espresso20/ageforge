package ui

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
)

// tradeProvider renders the Trade panel: market rates, trade routes and the
// allied bonuses that change trade income.
func tradeProvider(state game.GameState, _ int) string {
	var sb strings.Builder
	trade := state.Trade

	// === Market rates ===
	fmt.Fprintf(&sb, " [gold]═══ Market rates ═══[-]\n\n")
	if trade.TradeBuildings < 1 {
		sb.WriteString(" [yellow]Trading needs a Market: build market[-]\n\n")
	}
	if len(trade.ExchangeRates) == 0 {
		sb.WriteString(" [gray]The market has no rates in this age.[-]\n")
	} else {
		keys := make([]string, 0, len(trade.ExchangeRates))
		for k := range trade.ExchangeRates {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, key := range keys {
			sb.WriteString(" " + exchangeRateLine(trade.ExchangeRates[key]) + "\n")
		}
	}

	sb.WriteString("\n [gray]Commands: trade <give> <get> <amount to give>[-]\n")
	sb.WriteString(" [gray]Example: trade food wood 50[-]\n")

	if len(trade.TotalSold) > 0 || len(trade.TotalBought) > 0 {
		sb.WriteString("\n [gold]Traded at the market:[-]\n")
		if len(trade.TotalSold) > 0 {
			fmt.Fprintf(&sb, "   Sold: %s\n", game.Amounts(trade.TotalSold))
		}
		if len(trade.TotalBought) > 0 {
			fmt.Fprintf(&sb, "   Bought: %s\n", game.Amounts(trade.TotalBought))
		}
	}

	// === Trade routes ===
	sb.WriteString("\n [gold]═══ Trade routes ═══[-]\n\n")

	// Disruption banner: a war or an embargo blockades these resources, and
	// any route importing one is suspended until the conflict ends.
	if len(trade.DisruptedResources) > 0 {
		names := make([]string, len(trade.DisruptedResources))
		for i, r := range trade.DisruptedResources {
			names[i] = game.ResourceName(r)
		}
		fmt.Fprintf(&sb, " [red]⚠ Routes importing %s are suspended (a war, or your embargo).[-]\n", textfmt.List(names))
		sb.WriteString(" [gray]Lift your embargo with: diplomacy neutral <civ>[-]\n\n")
	}

	if len(trade.ActiveRoutes) > 0 {
		sb.WriteString(" [gold]Active routes:[-]\n\n")
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
			fmt.Fprintf(&sb, "   Export: %s\n", formatResMap(route.Export))
			fmt.Fprintf(&sb, "   Import: %s\n", formatResMap(route.Import))
			if route.Disrupted {
				fmt.Fprintf(&sb, "   [red]Suspended: %s imports are blockaded[-]\n\n", game.ResourceName(route.DisruptedBy))
			} else {
				fmt.Fprintf(&sb, "   %s left in this cycle  [gray](%s done)[-]\n\n",
					formatTicks(route.TicksLeft, state), textfmt.Count(route.CyclesDone, "cycle", "cycles"))
			}
		}
	}

	if len(trade.AvailableRoutes) > 0 {
		sb.WriteString(" [gold]Available routes:[-]\n\n")
		for _, route := range trade.AvailableRoutes {
			statusIcon := "[red]✗[-]"
			if route.CanStart {
				statusIcon = "[green]✓[-]"
			}
			fmt.Fprintf(&sb, " %s [cyan]%s[-]\n", statusIcon, route.Name)
			fmt.Fprintf(&sb, "   [gray]%s[-]\n", route.Description)
			fmt.Fprintf(&sb, "   Export: %s → Import: %s\n",
				formatResMap(route.Export), formatResMap(route.Import))
			if route.CanStart {
				fmt.Fprintf(&sb, "   [green]trade route start %s[-]\n", route.Key)
			} else {
				fmt.Fprintf(&sb, "   [red]needs %s (have %s)[-]\n",
					game.BuildingCount(route.MinCount, route.RequiredBld),
					textfmt.Int(state.Buildings[route.RequiredBld].Count))
			}
			sb.WriteString("\n")
		}
	}

	if len(trade.ActiveRoutes) == 0 && len(trade.AvailableRoutes) == 0 {
		sb.WriteString(" [gray]No trade routes in this age yet.[-]\n")
	}

	writeAllyTradeBonuses(&sb, state.Diplomacy)

	sb.WriteString(" [gray]Commands: trade route start/stop <key>[-]\n")

	// Civilization opinion lives on the Factions panel only. This panel used
	// to carry its own copy, which went stale.
	sb.WriteString("\n [gray]Opinion and deals with other civilizations: type factions[-]\n")

	return sb.String()
}

// exchangeRateLine renders one market rate from the player's side:
// "1 food → 0.50 wood (-30% from recent selling)". The rate shown is the one
// the market pays: never below the 50% floor game.TradeManager applies, so
// the percentage never shows a drop larger than 50%.
func exchangeRateLine(info game.ExchangeRateInfo) string {
	rate := info.Rate
	if floor := info.BaseRate * 0.5; rate < floor {
		rate = floor
	}
	rateColor := "white"
	if rate > info.BaseRate {
		rateColor = "green"
	} else if rate < info.BaseRate*0.9 {
		rateColor = "yellow"
	}
	shift := ""
	if info.BaseRate > 0 {
		pct := math.Round((rate/info.BaseRate - 1) * 100)
		switch {
		case pct <= -1:
			shift = fmt.Sprintf(" [red](%.0f%% from recent selling)[-]", pct)
		case pct >= 1:
			shift = fmt.Sprintf(" [green](+%.0f%% while the market recovers)[-]", pct)
		}
	}
	return fmt.Sprintf("1 %s → [%s]%s %s[-]%s",
		game.ResourceName(info.From), rateColor, marketRate(rate), game.ResourceName(info.To), shift)
}

// marketRate prints a per-unit rate: two decimals below 100 (0.50, 1.25),
// the shared number format above that.
func marketRate(r float64) string {
	if r < 100 {
		return fmt.Sprintf("%.2f", r)
	}
	return textfmt.Number(r)
}

// writeAllyTradeBonuses lists the allies currently boosting your income, the
// one piece of civilization state that belongs on the trade screen. It
// mirrors game.DiplomacyManager.GetTradeBonus: an allied civ that is not at
// war adds its TradeBonus to its specialty resource, both to route imports
// and to that resource's production rate. Writes nothing when no ally
// applies.
func writeAllyTradeBonuses(sb *strings.Builder, dip game.DiplomacyState) {
	keys := make([]string, 0, len(dip.Factions))
	for k, f := range dip.Factions {
		if f.Discovered && f.Status == "allied" && !f.AtWar && f.TradeBonus > 0 {
			keys = append(keys, k)
		}
	}
	if len(keys) == 0 {
		return
	}
	sort.Strings(keys)
	sb.WriteString(" [gold]Allied bonuses:[-]\n")
	for _, k := range keys {
		f := dip.Factions[k]
		fmt.Fprintf(sb, "   %s: [green]%s %s[-] [gray](route imports and production)[-]\n",
			f.Name, textfmt.SignedPercent(f.TradeBonus), game.ResourceName(f.Specialty))
	}
	sb.WriteString("\n")
}

// formatResMap formats a resource→amount map as "50 food, 20 wood", sorted by
// key so the order is stable. Returns "none" for an empty map.
func formatResMap(m map[string]float64) string {
	if len(m) == 0 {
		return "none"
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, game.Amount(m[k], k))
	}
	return strings.Join(parts, ", ")
}
