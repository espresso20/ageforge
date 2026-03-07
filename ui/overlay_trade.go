package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/espresso20/ageforge/game"
)

// tradeProvider generates the trade overlay text from the current game state.
// It mirrors the logic from TradeTab.Refresh — same data, formatted as plain text.
func tradeProvider(state game.GameState) string {
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

	if len(trade.ActiveRoutes) > 0 {
		sb.WriteString(" [gold]Active Routes:[-]\n\n")
		for _, route := range trade.ActiveRoutes {
			fmt.Fprintf(&sb, " [green]▸[-] [cyan]%s[-]\n", route.Name)
			fmt.Fprintf(&sb, "   Export: %s\n", formatResMap(route.Export))
			fmt.Fprintf(&sb, "   Import: %s\n", formatResMap(route.Import))
			fmt.Fprintf(&sb, "   %d ticks remaining  [gray](%d cycles done)[-]\n\n",
				route.TicksLeft, route.CyclesDone)
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

	sb.WriteString(" [gray]Commands: trade route start/stop <key>[-]\n")

	// === Diplomacy ===
	sb.WriteString("\n [gold]═══ Diplomacy ═══[-]\n\n")
	dip := state.Diplomacy

	if len(dip.Factions) == 0 {
		sb.WriteString(" [gray]No factions discovered yet[-]\n")
		sb.WriteString(" [gray]Reach Colonial Age to discover factions[-]\n")
	} else {
		factionKeys := make([]string, 0, len(dip.Factions))
		for k := range dip.Factions {
			factionKeys = append(factionKeys, k)
		}
		sort.Strings(factionKeys)

		for _, key := range factionKeys {
			faction := dip.Factions[key]
			if !faction.Discovered {
				fmt.Fprintf(&sb, " [gray]??? %s [Undiscovered][-]\n", faction.Name)
				continue
			}

			statusColor := "white"
			switch faction.Status {
			case "allied":
				statusColor = "green"
			case "friendly":
				statusColor = "cyan"
			case "rival":
				statusColor = "red"
			case "embargo":
				statusColor = "yellow"
			}

			opinionColor := "white"
			if faction.Opinion >= 50 {
				opinionColor = "green"
			} else if faction.Opinion >= 25 {
				opinionColor = "cyan"
			} else if faction.Opinion < 0 {
				opinionColor = "red"
			}

			bonusStr := ""
			if faction.Status == "allied" && faction.TradeBonus > 0 {
				bonusStr = fmt.Sprintf("  [green]+%.0f%% %s[-]", faction.TradeBonus*100, faction.Specialty)
			}

			fmt.Fprintf(&sb, " %-20s [%s][%s][-]  Op: [%s]%d[-]%s  [gray](%d trades)[-]\n",
				faction.Name, statusColor, faction.Status, opinionColor, faction.Opinion,
				bonusStr, faction.TradeCount)
		}
	}

	sb.WriteString("\n [gray]Commands: diplomacy ally/rival/embargo/gift/neutral <faction>[-]\n")

	return sb.String()
}

// formatResMap formats a resource map for display.
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
