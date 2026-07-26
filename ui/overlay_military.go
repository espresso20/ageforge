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

	// === Army Overview ===
	fmt.Fprintf(&sb, " [gold]═══ Army Overview ═══[-]\n\n")
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

	// === Active Campaign ===
	// Only the military campaign lives here; the active scout shows in the
	// Expeditions panel (the two run concurrently, one per category).
	sb.WriteString("\n [gold]═══ Active Campaign ═══[-]\n\n")
	if mil.ActiveMilitary == nil {
		sb.WriteString(" [gray]No active campaign[-]\n")
	} else {
		writeActiveExpedition(&sb, "Campaign", mil.ActiveMilitary)
	}
	fmt.Fprintf(&sb, "\n [gray]Completed: %d expedition(s)[-]\n", mil.CompletedCount)

	// === Campaigns ===
	// The Army panel lists only military campaigns (they cost soldiers).
	sb.WriteString("\n [gold]═══ Campaigns ═══[-]\n\n")
	if !hasCategory(mil.Expeditions, game.ExpeditionMilitary) {
		sb.WriteString(" [gray]No campaigns available yet[-]\n")
		sb.WriteString(" [gray]Reach Bronze Age and recruit soldiers[-]\n")
		sb.WriteString(" [gray]to unlock campaigns.[-]\n")
	} else {
		writeExpeditionGroup(&sb, "Campaigns", mil.Expeditions, game.ExpeditionMilitary)
	}

	// Loot totals now live in the Expeditions panel (see overlay_expeditions.go),
	// alongside the scouting/loot surface — not here in the Army panel.

	sb.WriteString("\n [gray]Commands: campaign <key>[-]\n")

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
		// Duration is a rolled range at launch; show "min-max ticks" when the def
		// defines one, else fall back to the legacy fixed value (never "0-0").
		durationStr := fmt.Sprintf("%d ticks", exp.Duration)
		if exp.DurationMax > exp.DurationMin {
			durationStr = fmt.Sprintf("%d-%d ticks", exp.DurationMin, exp.DurationMax)
		}
		fmt.Fprintf(sb, "   Soldiers: %d  Duration: %s  Difficulty: [%s]%.0f%%[-]\n",
			exp.SoldiersNeeded, durationStr, diffColor, exp.Difficulty*100)
		if cost := formatExpeditionCost(exp.Cost); cost != "" {
			fmt.Fprintf(sb, "   Cost: %s\n", cost)
		}

		// Launch hint uses the command that owns this category: scouting →
		// `expedition`, military → `campaign`.
		launchCmd := "expedition"
		if exp.Category == game.ExpeditionMilitary {
			launchCmd = "campaign"
		}
		if exp.CanLaunch {
			fmt.Fprintf(sb, "   [green]%s %s[-]\n", launchCmd, exp.Key)
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
