package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/rivo/tview"

	"github.com/user/ageforge/game"
)

// TradeTab displays trade exchange, routes, and diplomacy
type TradeTab struct {
	root       *tview.Flex
	exchangeTV *tview.TextView
	routesTV   *tview.TextView
	diplomacyTV *tview.TextView
}

// NewTradeTab creates the trade tab
func NewTradeTab() *TradeTab {
	t := &TradeTab{}

	t.exchangeTV = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true)
	t.exchangeTV.SetBorder(true).SetTitle(" Exchange Rates ").SetTitleColor(ColorTitle)

	t.routesTV = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true)
	t.routesTV.SetBorder(true).SetTitle(" Trade Routes ").SetTitleColor(ColorTitle)

	t.diplomacyTV = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true)
	t.diplomacyTV.SetBorder(true).SetTitle(" Diplomacy ").SetTitleColor(ColorTitle)

	// Top: exchange (left) + routes (right), Bottom: diplomacy
	topPanel := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(t.exchangeTV, 0, 1, false).
		AddItem(t.routesTV, 0, 1, false)

	t.root = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(topPanel, 0, 2, false).
		AddItem(t.diplomacyTV, 0, 1, false)

	return t
}

// Root returns the root primitive
func (t *TradeTab) Root() tview.Primitive {
	return t.root
}

// Refresh updates the trade tab with current state
func (t *TradeTab) Refresh(state game.GameState) {
	t.refreshExchange(state)
	t.refreshRoutes(state)
	t.refreshDiplomacy(state)
}

func (t *TradeTab) refreshExchange(state game.GameState) {
	var sb strings.Builder
	trade := state.Trade

	if len(trade.ExchangeRates) == 0 {
		sb.WriteString(" [gray]No exchange rates available yet[-]\n")
		sb.WriteString(" [gray]Build a market to unlock trading[-]\n")
	} else {
		// Sort keys for stable display
		keys := make([]string, 0, len(trade.ExchangeRates))
		for k := range trade.ExchangeRates {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, key := range keys {
			info := trade.ExchangeRates[key]
			pressureColor := "gray"
			pressureStr := ""
			if info.Pressure > 0.1 {
				pressureColor = "red"
				pressureStr = fmt.Sprintf(" [%s]↓%.0f%%[-]", pressureColor, info.Pressure*30)
			} else if info.Pressure < -0.1 {
				pressureColor = "green"
				pressureStr = fmt.Sprintf(" [%s]↑%.0f%%[-]", pressureColor, -info.Pressure*30)
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

	// Stats
	if len(trade.TotalExchanged) > 0 {
		sb.WriteString("\n [gold]Total Exchanged:[-]\n")
		for res, amount := range trade.TotalExchanged {
			fmt.Fprintf(&sb, "   %s: %.0f\n", res, amount)
		}
	}

	t.exchangeTV.SetText(sb.String())
}

func (t *TradeTab) refreshRoutes(state game.GameState) {
	var sb strings.Builder
	trade := state.Trade

	// Active routes
	if len(trade.ActiveRoutes) > 0 {
		sb.WriteString(" [gold]Active Routes:[-]\n\n")
		for _, route := range trade.ActiveRoutes {
			fmt.Fprintf(&sb, " [green]▸[-] [cyan]%s[-]\n", route.Name)
			fmt.Fprintf(&sb, "   Export: %s\n", formatResMap(route.Export))
			fmt.Fprintf(&sb, "   Import: %s\n", formatResMap(route.Import))
			bar := ProgressBar(float64(route.TicksLeft), float64(route.TicksLeft+1), 15)
			fmt.Fprintf(&sb, "   %s %d ticks  [gray](%d cycles)[-]\n\n", bar, route.TicksLeft, route.CyclesDone)
		}
	}

	// Available routes
	if len(trade.AvailableRoutes) > 0 {
		sb.WriteString(" [gold]Available Routes:[-]\n\n")
		for _, route := range trade.AvailableRoutes {
			statusIcon := "[red]✗[-]"
			if route.CanStart {
				statusIcon = "[green]✓[-]"
			}
			fmt.Fprintf(&sb, " %s [cyan]%s[-]\n", statusIcon, route.Name)
			fmt.Fprintf(&sb, "   [gray]%s[-]\n", route.Description)
			fmt.Fprintf(&sb, "   Export: %s → Import: %s\n", formatResMap(route.Export), formatResMap(route.Import))
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

	t.routesTV.SetText(sb.String())
}

func (t *TradeTab) refreshDiplomacy(state game.GameState) {
	var sb strings.Builder
	dip := state.Diplomacy

	if len(dip.Factions) == 0 {
		sb.WriteString(" [gray]No factions discovered yet[-]\n")
		sb.WriteString(" [gray]Reach Colonial Age to discover factions[-]\n")
	} else {
		// Sort by key for stable display
		keys := make([]string, 0, len(dip.Factions))
		for k := range dip.Factions {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, key := range keys {
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

	t.diplomacyTV.SetText(sb.String())
}

// formatResMap formats a resource map for display
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
