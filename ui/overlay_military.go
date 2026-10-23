package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/vibe"
)

// militaryProvider generates the military overlay text from the current game state.
// It mirrors the logic from MilitaryTab.Refresh — same data, formatted as plain text.
func militaryProvider(state game.GameState, _ int) string {
	var sb strings.Builder
	mil := state.Military

	// === Army Overview ===
	fmt.Fprint(&sb, vibe.Sprintf("ui.overlay_military.militaryProvider.01", " [gold]═══ Army Overview ═══[-]\n\n"))
	fmt.Fprint(&sb, vibe.Sprintf("ui.overlay_military.militaryProvider.02", " [gold]Soldiers:[-]  %s posted of %s\n", FormatNumber(float64(mil.SoldierCount)), FormatNumber(float64(mil.SoldierCap))))
	if len(mil.Posts) == 0 {
		// Soldiers begin with the first military building: until one is built
		// there is nobody at a post, and the panel says so.
		fmt.Fprint(&sb, vibe.Sprintf("ui.overlay_military.militaryProvider.03", " [gold]Posts:[-]     %s\n", theme.Paint(theme.RoleDim, vibe.Line("ui.overlay_military.militaryProvider.04", "none yet. A military building comes with its soldiers."))))
	} else {
		for _, p := range mil.Posts {
			fmt.Fprint(&sb, vibe.Sprintf("ui.overlay_military.militaryProvider.05", " [gold]Posts:[-]     %s x%d: %s/%s\n", p.Name, p.Count, FormatNumber(float64(p.Posted)), FormatNumber(float64(p.Posts))))
		}
	}
	fmt.Fprint(&sb, vibe.Sprintf("ui.overlay_military.militaryProvider.06", " [gold]Defense:[-]   %s\n", FormatNumber(mil.DefenseRating)))
	writeGarrison(&sb, state)

	if mil.MilitaryBonus > 0 {
		fmt.Fprint(&sb, vibe.Sprintf("ui.overlay_military.militaryProvider.07", " [green]Military Bonus: +%.0f%%[-]\n", mil.MilitaryBonus*100))
	}
	if mil.ExpeditionBonus > 0 {
		fmt.Fprint(&sb, vibe.Sprintf("ui.overlay_military.militaryProvider.08", " [green]Expedition Bonus: +%.0f%%[-]\n", mil.ExpeditionBonus*100))
	}

	// === Active Campaign ===
	// Only the military campaign lives here; the active scout shows in the
	// Expeditions panel (the two run concurrently, one per category).
	sb.WriteString(vibe.Line("ui.overlay_military.militaryProvider.09", "\n [gold]═══ Active Campaign ═══[-]\n\n"))
	if mil.ActiveMilitary == nil {
		sb.WriteString(vibe.Line("ui.overlay_military.militaryProvider.10", " [gray]No active campaign[-]\n"))
	} else {
		writeActiveExpedition(&sb, vibe.Line("ui.overlay_military.militaryProvider.11", "Campaign"), mil.ActiveMilitary, state)
	}
	fmt.Fprint(&sb, vibe.Sprintf("ui.overlay_military.militaryProvider.12", "\n [gray]Completed: %d (expeditions and campaigns)[-]\n", mil.CompletedCount))

	// === Campaigns ===
	// The Army panel lists only military campaigns (they cost soldiers).
	sb.WriteString(vibe.Line("ui.overlay_military.militaryProvider.13", "\n [gold]═══ Campaigns ═══[-]\n\n"))
	if !hasCategory(mil.Expeditions, game.ExpeditionMilitary) {
		sb.WriteString(vibe.Line("ui.overlay_military.militaryProvider.14", " [gray]No campaigns available yet.[-]\n"))
		sb.WriteString(theme.Paint(theme.RoleDim, vibe.Line("ui.overlay_military.militaryProvider.15", " Campaigns are grand missions and cost soldiers,")) + "\n")
		sb.WriteString(theme.Paint(theme.RoleDim, vibe.Line("ui.overlay_military.militaryProvider.16", " which come with military buildings. They open as your civilization grows.")) + "\n")
	} else {
		writeExpeditionGroup(&sb, vibe.Line("ui.overlay_military.militaryProvider.17", "Campaigns"), mil.Expeditions, game.ExpeditionMilitary, state)
	}

	sb.WriteString(vibe.Line("ui.overlay_military.militaryProvider.18", "\n [gray]Commands: expedition opens the Expeditions panel; expedition <name> sends an offer[-]\n"))

	return sb.String()
}

// hasCategory reports whether any offer in exps matches the given category.
func hasCategory(exps []game.ExpeditionInfo, category string) bool {
	for _, exp := range exps {
		if exp.Category == category {
			return true
		}
	}
	return false
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
		fmt.Fprint(sb, vibe.Sprintf("ui.overlay_military.writeActiveExpedition.01", ", %d soldiers deployed", exp.Soldiers))
	}
	fmt.Fprint(sb, vibe.Sprintf("ui.overlay_military.writeActiveExpedition.02", " (%s left)\n", formatTicks(exp.TicksLeft, state)))
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
		soldiers := ""
		if exp.SoldiersNeeded > 0 {
			soldiers = vibe.Sprintf("ui.overlay_military.writeExpeditionGroup.01", "Soldiers: %d  ", exp.SoldiersNeeded)
		}
		// The listed difficulty, and what the army leaves of it today.
		note := vibe.Line("ui.overlay_military.writeExpeditionGroup.02", "(chance of failure before bonuses)")
		if exp.Chance > 0 && exp.Chance < exp.Difficulty-0.0005 {
			note = vibe.Sprintf("ui.overlay_military.writeExpeditionGroup.03", "(chance of failure; %.0f%% with your army)", exp.Chance*100)
		}
		fmt.Fprint(sb, vibe.Sprintf("ui.overlay_military.writeExpeditionGroup.04", "   %sDuration: %s  Difficulty: [%s]%.0f%%[-] %s\n",
			soldiers, durationStr, diffColor, exp.Difficulty*100, note))
		if cost := formatExpeditionCost(exp.Cost, state); cost != "" {
			fmt.Fprint(sb, vibe.Sprintf("ui.overlay_military.writeExpeditionGroup.05", "   Cost: %s\n", cost))
		}

		if exp.CanLaunch {
			sb.WriteString(vibe.Line("ui.overlay_military.writeExpeditionGroup.06", "   [green]Ready: send it from the Expeditions panel[-]\n"))
		} else {
			fmt.Fprintf(sb, "   [red]%s[-]\n", exp.LaunchBlockReason)
		}
		sb.WriteString("\n")
	}
}

// formatExpeditionCost renders an expedition's resource cost in a readable,
// player-facing form (e.g. "30 food, 30 wood"), with keys sorted for stable
// output. Returns "" for a free (empty) cost so callers can omit the line.
func formatExpeditionCost(cost map[string]float64, state game.GameState) string {
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
		parts = append(parts, FormatNumber(cost[k])+" "+resourceDisplayName(state, k))
	}
	return strings.Join(parts, ", ")
}

func expeditionHaul(exp game.ExpeditionInfo, state game.GameState) string {
	if haul := formatExpeditionCost(exp.ExpectedHaulMin, state); haul != "" {
		return haul
	}
	return vibe.Sprintf("ui.overlay_military.expeditionHaul.01", "about %.0f min of income", exp.HaulMinutes)
}

func expeditionOfferCost(exp game.ExpeditionInfo, state game.GameState) string {
	if cost := formatExpeditionCost(exp.Cost, state); cost != "" {
		return cost
	}
	if exp.Tier == game.TierGrand && exp.Grand == game.GrandCivil {
		return "materials"
	}
	if exp.Tier == game.TierGrand && exp.Grand == game.GrandMilitary {
		return "soldiers"
	}
	return "free"
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
	fmt.Fprint(sb, vibe.Sprintf("ui.overlay_military.writeGarrison.01", " [gold]Threat:[-]    %s (raids in the %s)\n", FormatNumber(mil.Threat), ageName))
	if mil.Mitigation <= 0 {
		sb.WriteString(vibe.Line("ui.overlay_military.writeGarrison.02", " [yellow]You have no garrison: raids hit you with full force.[-]\n"))
		fmt.Fprint(sb, vibe.Sprintf("ui.overlay_military.writeGarrison.03", "   [gray]Soldiers blunt raids, war raids and what an Endure takes (up to %.0f%%).[-]\n", capPct))
	} else {
		fmt.Fprint(sb, vibe.Sprintf("ui.overlay_military.writeGarrison.04", " [green]Your garrison would blunt about %.0f%% of a raid.[-]\n", mil.Mitigation*100))
		sb.WriteString(vibe.Line("ui.overlay_military.writeGarrison.05", "   [gray]Raids, war raids and an Endure's losses all hit you that much softer.[-]\n"))
		twice := config.DefenseMitigation(mil.DefenseRating*2, mil.Threat)
		fmt.Fprint(sb, vibe.Sprintf("ui.overlay_military.writeGarrison.06", "   [gray]Twice the garrison: about %.0f%%. No army blunts more than %.0f%%.[-]\n", twice*100, capPct))
	}
	if line := garrisonSavedSummary(mil.Saved); line != "" {
		fmt.Fprint(sb, vibe.Sprintf("ui.overlay_military.writeGarrison.07", " [gold]Saved this run:[-] %s\n", line))
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
		parts = append(parts, fmt.Sprintf("%d %s", t.Buildings, pluralize(vibe.Line("ui.overlay_military.garrisonSavedSummary.01", "building"), t.Buildings)))
	}
	if t.Workers > 0 {
		parts = append(parts, fmt.Sprintf("%d %s", t.Workers, pluralize(vibe.Line("ui.overlay_military.garrisonSavedSummary.02", "worker"), t.Workers)))
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
		out += vibe.Sprintf("ui.overlay_military.garrisonSavedSummary.03", " [gray](%d %s blunted)[-]", t.Raids, pluralize(vibe.Line("ui.overlay_military.garrisonSavedSummary.04", "raid"), t.Raids))
	}
	return out
}

// resourceDisplayName is a resource's player-facing name in lower case ("dark
// matter"), from the snapshot, or the key with spaces when it has none.
func resourceDisplayName(state game.GameState, key string) string {
	if r, ok := state.Resources[key]; ok && r.Name != "" {
		return strings.ToLower(r.Name)
	}
	return strings.ReplaceAll(key, "_", " ")
}
