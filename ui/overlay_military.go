package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// militaryProvider generates the military overlay text from the current game state.
// It mirrors the logic from MilitaryTab.Refresh — same data, formatted as plain text.
func militaryProvider(state game.GameState, _ int) string {
	var sb strings.Builder
	mil := state.Military

	// === Army Overview ===
	fmt.Fprintf(&sb, " [gold]═══ Army Overview ═══[-]\n\n")
	fmt.Fprintf(&sb, " [gold]Soldiers:[-]  %s / %s\n", FormatNumber(float64(mil.SoldierCount)), FormatNumber(float64(mil.SoldierCap)))
	fmt.Fprintf(&sb, " [gold]Training:[-]  %s/tick\n", FormatRate(mil.SoldierRate))
	fmt.Fprintf(&sb, " [gold]Defense:[-]   %s\n", FormatNumber(mil.DefenseRating))
	writeGarrison(&sb, state)

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
		writeActiveExpedition(&sb, "Campaign", mil.ActiveMilitary, state)
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
		writeExpeditionGroup(&sb, "Campaigns", mil.Expeditions, game.ExpeditionMilitary, state)
	}

	// Loot totals now live in the Expeditions panel (see overlay_expeditions.go),
	// alongside the scouting/loot surface — not here in the Army panel.

	sb.WriteString("\n [gray]Commands: campaign <key>[-]\n")

	return sb.String()
}

// writeActiveExpedition renders one active expedition under a kind label
// (e.g. "Scouting"). Nothing is written when exp is nil, so the section only
// lists kinds that are actually running.
// (state is threaded through only to convert the countdown to wall-clock.)
func writeActiveExpedition(sb *strings.Builder, label string, exp *game.ExpeditionSnapshot, state game.GameState) {
	if exp == nil {
		return
	}
	fmt.Fprintf(sb, " [yellow]%s:[-] %s", label, exp.Name)
	if exp.Soldiers > 0 {
		fmt.Fprintf(sb, " — %d deployed", exp.Soldiers)
	}
	fmt.Fprintf(sb, " (%s left)\n", formatTicks(exp.TicksLeft, state))
}

// writeExpeditionGroup renders the subset of exps matching category under a
// labeled header (e.g. "Scouting"). If no expedition matches, nothing is
// written — the header is omitted for empty subsections.
func writeExpeditionGroup(sb *strings.Builder, label string, exps []game.ExpeditionInfo, category string, state game.GameState) {
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
		// Duration is rolled per launch, so preview the def's range rather than a
		// single number.
		durationStr := formatTickRange(exp.DurationMin, exp.DurationMax, state)
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

// writeGarrison renders what the army does for the player: how much of a raid
// the garrison would blunt against the current age's threat, and what it has
// saved this run. Worded from the player's side.
func writeGarrison(sb *strings.Builder, state game.GameState) {
	mil := state.Military
	capPct := config.DefenseMitigationCap * 100
	ageName := state.AgeName
	if ageName == "" {
		ageName = state.Age
	}
	fmt.Fprintf(sb, " [gold]Threat:[-]    %s (raids in the %s)\n", FormatNumber(mil.Threat), ageName)
	if mil.Mitigation <= 0 {
		sb.WriteString(" [yellow]You have no garrison: raids hit you with full force.[-]\n")
		fmt.Fprintf(sb, "   [gray]Soldiers blunt raids, war raids and what an Endure takes (up to %.0f%%).[-]\n", capPct)
	} else {
		fmt.Fprintf(sb, " [green]Your garrison would blunt about %.0f%% of a raid.[-]\n", mil.Mitigation*100)
		sb.WriteString("   [gray]Raids, war raids and an Endure's losses all hit you that much softer.[-]\n")
		twice := config.DefenseMitigation(mil.DefenseRating*2, mil.Threat)
		fmt.Fprintf(sb, "   [gray]Twice the garrison: about %.0f%%. No army blunts more than %.0f%%.[-]\n", twice*100, capPct)
	}
	if line := garrisonSavedSummary(mil.Saved); line != "" {
		fmt.Fprintf(sb, " [gold]Saved this run:[-] %s\n", line)
	}
}

// garrisonSavedSummary is a one-line summary of what the army has saved this
// run, largest resource amounts first (at most four), or "" when nothing.
func garrisonSavedSummary(t *game.DefenseTally) string {
	if t == nil {
		return ""
	}
	var parts []string
	if t.Buildings > 0 {
		parts = append(parts, fmt.Sprintf("%d %s", t.Buildings, pluralize("building", t.Buildings)))
	}
	if t.Workers > 0 {
		parts = append(parts, fmt.Sprintf("%d %s", t.Workers, pluralize("worker", t.Workers)))
	}
	keys := make([]string, 0, len(t.Resources))
	for k, v := range t.Resources {
		if v >= 1 {
			keys = append(keys, k)
		}
	}
	sort.Slice(keys, func(i, j int) bool {
		if t.Resources[keys[i]] != t.Resources[keys[j]] {
			return t.Resources[keys[i]] > t.Resources[keys[j]]
		}
		return keys[i] < keys[j]
	})
	if len(keys) > 4 {
		keys = keys[:4]
	}
	for _, k := range keys {
		parts = append(parts, FormatNumber(t.Resources[k])+" "+k)
	}
	if len(parts) == 0 {
		return ""
	}
	out := strings.Join(parts, ", ")
	if t.Raids > 0 {
		out += fmt.Sprintf(" [gray](%d %s blunted)[-]", t.Raids, pluralize("raid", t.Raids))
	}
	return out
}
