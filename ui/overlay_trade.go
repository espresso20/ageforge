package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/espresso20/ageforge/game"
)

// tradeProvider generates the trade overlay text from the current game state.
// It mirrors the logic from TradeTab.Refresh — same data, formatted as plain text.
func tradeProvider(state game.GameState, _ int) string {
	var sb strings.Builder
	trade := state.Trade

	// === Exchange Rates ===
	fmt.Fprintf(&sb, " [gold]═══ Exchange Rates ═══[-]\n\n")
	if len(trade.ExchangeRates) == 0 {
		sb.WriteString(" [gray]No exchange rates available yet[-]\n")
		sb.WriteString(" [gray]Build a market to unlock trading[-]\n")
	} else {
		keys := make([]string, 0, len(trade.ExchangeRates))
		for k := range trade.ExchangeRates {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, key := range keys {
			info := trade.ExchangeRates[key]
			// Market pressure is a short-term penalty/bonus from repeated trades.
			// Positive pressure = price moved against you (recently sold a lot).
			// Negative pressure = price moved in your favour (recovery period).
			pressureStr := ""
			if info.Pressure > 0.1 {
				pressureStr = fmt.Sprintf(" [red]↓%.0f%%[-]", info.Pressure*30)
			} else if info.Pressure < -0.1 {
				pressureStr = fmt.Sprintf(" [green]↑%.0f%%[-]", -info.Pressure*30)
			}

			rateColor := "white"
			if info.Rate > info.BaseRate {
				rateColor = "green"
			} else if info.Rate < info.BaseRate*0.9 {
				rateColor = "yellow"
			}

			fmt.Fprintf(&sb, " %s → %s: [%s]%.2f[-]%s\n",
				info.From, info.To, rateColor, info.Rate, pressureStr)
		}
	}

	sb.WriteString("\n [gray]Commands: trade <from> <to> <amount>[-]\n")
	sb.WriteString(" [gray]Example: trade food wood 50[-]\n")

	if len(trade.TotalExchanged) > 0 {
		sb.WriteString("\n [gold]Total Exchanged:[-]\n")
		exchKeys := make([]string, 0, len(trade.TotalExchanged))
		for k := range trade.TotalExchanged {
			exchKeys = append(exchKeys, k)
		}
		sort.Strings(exchKeys)
		for _, res := range exchKeys {
			fmt.Fprintf(&sb, "   %s: %.0f\n", res, trade.TotalExchanged[res])
		}
	}

	// === Trade Routes ===
	sb.WriteString("\n [gold]═══ Trade Routes ═══[-]\n\n")

	// Disruption banner: war/embargo blockades suspend any route importing the
	// listed resources until the conflict ends.
	if len(trade.DisruptedResources) > 0 {
		fmt.Fprintf(&sb, " [red]⚠ Trade disrupted by hostile powers:[-] %s\n",
			strings.Join(trade.DisruptedResources, ", "))
		sb.WriteString(" [gray]Routes importing these are suspended until peace (end the war/embargo).[-]\n\n")
	}

	if len(trade.ActiveRoutes) > 0 {
		sb.WriteString(" [gold]Active Routes:[-]\n\n")
		for _, route := range trade.ActiveRoutes {
			marker := "[green]▸[-]"
			if route.Disrupted {
				marker = "[red]✖[-]"
			}
			fmt.Fprintf(&sb, " %s [cyan]%s[-]\n", marker, route.Name)
			fmt.Fprintf(&sb, "   Export: %s\n", formatResMap(route.Export))
			fmt.Fprintf(&sb, "   Import: %s\n", formatResMap(route.Import))
			if route.Disrupted {
				fmt.Fprintf(&sb, "   [red]DISRUPTED — %s shipments blockaded[-]\n\n", route.DisruptedBy)
			} else {
				fmt.Fprintf(&sb, "   %s remaining  [gray](%d cycles done)[-]\n\n",
					formatTicks(route.TicksLeft, state), route.CyclesDone)
			}
		}
	}

	if len(trade.AvailableRoutes) > 0 {
		sb.WriteString(" [gold]Available Routes:[-]\n\n")
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
				fmt.Fprintf(&sb, "   [red]need %d %s[-]\n", route.MinCount, route.RequiredBld)
			}
			sb.WriteString("\n")
		}
	}

	if len(trade.ActiveRoutes) == 0 && len(trade.AvailableRoutes) == 0 {
		sb.WriteString(" [gray]No trade routes available yet[-]\n")
		sb.WriteString(" [gray]Build a market to unlock trade routes[-]\n")
	}

	writeAllyTradeBonuses(&sb, state.Diplomacy)

	sb.WriteString(" [gray]Commands: trade route start/stop <key>[-]\n")

	// Faction standing lives on the Factions panel only. This overlay used to
	// carry its own copy, which went stale (it still said factions were found
	// by reaching the Colonial Age; first contact comes from expeditions).
	sb.WriteString("\n [gray]Faction standing and deals: type factions[-]\n")

	return sb.String()
}

// writeAllyTradeBonuses lists the allies currently boosting your income, the
// one piece of faction state that belongs on the trade screen. It mirrors
// game.DiplomacyManager.GetTradeBonus: an allied civ that is not at war adds
// its TradeBonus to its specialty resource, both to route imports and to that
// resource's production rate. Writes nothing when no ally applies.
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
	sb.WriteString(" [gold]Allied Bonuses:[-]\n")
	for _, k := range keys {
		f := dip.Factions[k]
		fmt.Fprintf(sb, "   %s: [green]+%.0f%% %s[-] [gray](route imports and production)[-]\n",
			f.Name, f.TradeBonus*100, f.Specialty)
	}
	sb.WriteString("\n")
}

// formatResMap formats a resource→amount map into a sorted, comma-separated
// string (e.g. "50 food, 20 wood"). Returns "none" for empty maps.
// Keys are sorted for stable output across Go map iteration.
func formatResMap(m map[string]float64) string {
	if len(m) == 0 {
		return "none"
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%.0f %s", m[k], k))
	}
	return strings.Join(parts, ", ")
}
