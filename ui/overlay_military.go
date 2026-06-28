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

	// === Active Expeditions ===
	// A scouting and a military expedition can run concurrently (one per kind).
	sb.WriteString("\n [gold]═══ Active Expeditions ═══[-]\n\n")
	if mil.ActiveScout == nil && mil.ActiveMilitary == nil {
		sb.WriteString(" [gray]No active expeditions[-]\n")
	} else {
		writeActiveExpedition(&sb, "Scouting", mil.ActiveScout)
		writeActiveExpedition(&sb, "Military", mil.ActiveMilitary)
	}
	fmt.Fprintf(&sb, "\n [gray]Completed: %d expedition(s)[-]\n", mil.CompletedCount)

	// === Available Expeditions ===
	sb.WriteString("\n [gold]═══ Available Expeditions ═══[-]\n\n")
	if len(mil.Expeditions) == 0 {
		sb.WriteString(" [gray]No expeditions available yet[-]\n")
		sb.WriteString(" [gray]Reach Bronze Age and recruit soldiers[-]\n")
		sb.WriteString(" [gray]to unlock expeditions.[-]\n")
	} else {
		// Group available expeditions by category: Scouting first (resource cost,
		// no soldiers, available early), then Military Campaigns (cost soldiers).
		// A subsection header is omitted when it has no available entries.
		writeExpeditionGroup(&sb, "Scouting", mil.Expeditions, game.ExpeditionScouting)
		writeExpeditionGroup(&sb, "Military Campaigns", mil.Expeditions, game.ExpeditionMilitary)
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

// writeActiveExpedition renders one active expedition under a kind label
// (e.g. "Scouting"). Nothing is written when exp is nil, so the section only
// lists kinds that are actually running.
func writeActiveExpedition(sb *strings.Builder, label string, exp *game.ExpeditionSnapshot) {
	if exp == nil {
		return
	}
	fmt.Fprintf(sb, " [yellow]%s:[-] %s", label, exp.Name)
	if exp.Soldiers > 0 {
		fmt.Fprintf(sb, " — %d deployed", exp.Soldiers)
	}
	fmt.Fprintf(sb, " (%d ticks left)\n", exp.TicksLeft)
}

// writeExpeditionGroup renders the subset of exps matching category under a
// labeled header (e.g. "Scouting"). If no expedition matches, nothing is
// written — the header is omitted for empty subsections.
func writeExpeditionGroup(sb *strings.Builder, label string, exps []game.ExpeditionInfo, category string) {
	first := true
	for _, exp := range exps {
		if exp.Category != category {
			continue
		}
		if first {
			fmt.Fprintf(sb, " [yellow]── %s ──[-]\n\n", label)
			first = false
		}

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

		fmt.Fprintf(sb, " %s [cyan]%s[-]\n", statusIcon, exp.Name)
		fmt.Fprintf(sb, "   [gray]%s[-]\n", exp.Description)
		fmt.Fprintf(sb, "   Soldiers: %d  Duration: %d ticks  Difficulty: [%s]%.0f%%[-]\n",
			exp.SoldiersNeeded, exp.Duration, diffColor, exp.Difficulty*100)
		if cost := formatExpeditionCost(exp.Cost); cost != "" {
			fmt.Fprintf(sb, "   Cost: %s\n", cost)
		}

		if exp.CanLaunch {
			fmt.Fprintf(sb, "   [green]expedition %s[-]\n", exp.Key)
		} else {
			fmt.Fprintf(sb, "   [red]%s[-]\n", exp.LaunchBlockReason)
		}
		sb.WriteString("\n")
	}
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
