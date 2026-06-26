package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/espresso20/ageforge/game"
)

// militaryProvider generates the military overlay text from the current game state.
// It mirrors the logic from MilitaryTab.Refresh — same data, formatted as plain text.
func militaryProvider(state game.GameState, _ int) string {
	var sb strings.Builder
	mil := state.Military

	// === Military Overview ===
	fmt.Fprintf(&sb, " [gold]═══ Military Overview ═══[-]\n\n")
	fmt.Fprintf(&sb, " [gold]Soldiers:[-]  %d / %d\n", mil.SoldierCount, mil.SoldierCap)
	fmt.Fprintf(&sb, " [gold]Training:[-]  %s/tick\n", FormatRate(mil.SoldierRate))
	fmt.Fprintf(&sb, " [gold]Defense:[-]   %.1f\n", mil.DefenseRating)

	if mil.MilitaryBonus > 0 {
		fmt.Fprintf(&sb, " [green]Military Bonus: +%.0f%%[-]\n", mil.MilitaryBonus*100)
	}
	if mil.ExpeditionBonus > 0 {
		fmt.Fprintf(&sb, " [green]Expedition Bonus: +%.0f%%[-]\n", mil.ExpeditionBonus*100)
	}

	// Military worker domain assignments (if unlocked)
	if domainState, ok := state.Workers.Types["military"]; ok && domainState.Unlocked {
		sb.WriteString("\n")
		fmt.Fprintf(&sb, " [gold]Military Workers:[-] %d total, %d idle\n",
			domainState.Count, domainState.IdleCount)
		if len(domainState.Assignments) > 0 {
			bldKeys := make([]string, 0, len(domainState.Assignments))
			for k := range domainState.Assignments {
				bldKeys = append(bldKeys, k)
			}
			sort.Strings(bldKeys)
			for _, bk := range bldKeys {
				count := domainState.Assignments[bk]
				bldName := bk
				if bs, ok := state.Buildings[bk]; ok {
					bldName = bs.Name
				}
				fmt.Fprintf(&sb, "   %-20s %d assigned\n", bldName, count)
			}
		}
	}

	// === Active Expedition ===
	sb.WriteString("\n [gold]═══ Active Expedition ═══[-]\n\n")
	if mil.ActiveExpedition != nil {
		exp := mil.ActiveExpedition
		fmt.Fprintf(&sb, " [yellow]Name:[-]     %s\n", exp.Name)
		fmt.Fprintf(&sb, " [yellow]Soldiers:[-] %d deployed\n", exp.Soldiers)
		fmt.Fprintf(&sb, " [yellow]Remaining:[-] %d ticks\n", exp.TicksLeft)
	} else {
		sb.WriteString(" [gray]No active expedition[-]\n")
	}
	fmt.Fprintf(&sb, "\n [gray]Completed: %d expedition(s)[-]\n", mil.CompletedCount)

	// === Available Expeditions ===
	sb.WriteString("\n [gold]═══ Available Expeditions ═══[-]\n\n")
	if len(mil.Expeditions) == 0 {
		sb.WriteString(" [gray]No expeditions available yet[-]\n")
		sb.WriteString(" [gray]Reach Bronze Age and recruit soldiers[-]\n")
		sb.WriteString(" [gray]to unlock expeditions.[-]\n")
	} else {
		for _, exp := range mil.Expeditions {
			statusIcon := "[red]▸[-]"
			if exp.CanLaunch {
				statusIcon = "[green]▸[-]"
			}

			diffColor := "green"
			if exp.Difficulty > 0.5 {
				diffColor = "red"
			} else if exp.Difficulty > 0.3 {
				diffColor = "yellow"
			}

			fmt.Fprintf(&sb, " %s [cyan]%s[-]\n", statusIcon, exp.Name)
			fmt.Fprintf(&sb, "   [gray]%s[-]\n", exp.Description)
			fmt.Fprintf(&sb, "   Soldiers: %d  Duration: %d ticks  Difficulty: [%s]%.0f%%[-]\n",
				exp.SoldiersNeeded, exp.Duration, diffColor, exp.Difficulty*100)
			if cost := formatExpeditionCost(exp.Cost); cost != "" {
				fmt.Fprintf(&sb, "   Cost: %s\n", cost)
			}

			if exp.CanLaunch {
				fmt.Fprintf(&sb, "   [green]expedition %s[-]\n", exp.Key)
			} else {
				fmt.Fprintf(&sb, "   [red]%s[-]\n", exp.LaunchBlockReason)
			}
			sb.WriteString("\n")
		}
	}

	// === Loot History ===
	sb.WriteString(" [gold]═══ Loot History ═══[-]\n\n")
	if len(mil.TotalLoot) == 0 {
		sb.WriteString(" [gray]No loot collected yet[-]\n")
		sb.WriteString(" [gray]Complete expeditions to earn rewards![-]\n")
	} else {
		sb.WriteString(" [gold]Total Loot Collected:[-]\n\n")
		lootKeys := make([]string, 0, len(mil.TotalLoot))
		for k := range mil.TotalLoot {
			lootKeys = append(lootKeys, k)
		}
		sort.Strings(lootKeys)
		for _, key := range lootKeys {
			amount := mil.TotalLoot[key]
			fmt.Fprintf(&sb, " %-12s [green]%.0f[-]\n", key, amount)
		}
	}

	sb.WriteString("\n [gray]Commands: expedition <key>[-]\n")

	return sb.String()
}

// formatExpeditionCost renders an expedition's resource cost in a readable,
// player-facing form (e.g. "30 food, 30 wood"), with keys sorted for stable
// output. Returns "" for a free (empty) cost so callers can omit the line.
func formatExpeditionCost(cost map[string]float64) string {
	if len(cost) == 0 {
		return ""
	}
	keys := make([]string, 0, len(cost))
	for k := range cost {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%.0f %s", cost[k], k))
	}
	return strings.Join(parts, ", ")
}
