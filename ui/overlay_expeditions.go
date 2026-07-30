package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/espresso20/ageforge/game"
)

// expeditionsProvider generates the Expeditions overlay text — the civilian
// SCOUTING surface. It shows ONLY scouting: the active scouting expedition and
// the available scouting expeditions (resource cost / rewards / age). No
// soldiers, no defense, no military bonuses live here — those belong to the
// Army panel. Formatting mirrors the militaryProvider sections.
func expeditionsProvider(state game.GameState, _ int) string {
	var sb strings.Builder
	mil := state.Military

	fmt.Fprintf(&sb, " [gold]═══ Expeditions ═══[-]\n\n")
	sb.WriteString(" [gray]Send scouts to explore. Expeditions cost resources,[-]\n")
	sb.WriteString(" [gray]not soldiers, and run alongside Army campaigns.[-]\n")

	if mil.ExpeditionBonus > 0 {
		fmt.Fprintf(&sb, "\n [green]Expedition Bonus: +%.0f%%[-]\n", mil.ExpeditionBonus*100)
	}

	// === Active Expedition ===
	sb.WriteString("\n [gold]═══ Active Expedition ═══[-]\n\n")
	if mil.ActiveScout == nil {
		sb.WriteString(" [gray]No active expedition[-]\n")
	} else {
		writeActiveExpedition(&sb, "Scouting", mil.ActiveScout)
	}

	// === Available Expeditions ===
	sb.WriteString("\n [gold]═══ Available Expeditions ═══[-]\n\n")
	if !hasCategory(mil.Expeditions, game.ExpeditionScouting) {
		sb.WriteString(" [gray]No expeditions available yet[-]\n")
		sb.WriteString(" [gray]Reach Bronze Age to unlock more scouting.[-]\n")
	} else {
		writeExpeditionGroup(&sb, "Scouting", mil.Expeditions, game.ExpeditionScouting)
	}

	// === Loot History ===
	// Total loot lives with Expeditions (the scouting/loot surface), not the Army
	// panel — loot accrues from resolving expeditions and campaigns alike.
	sb.WriteString("\n [gold]═══ Loot History ═══[-]\n\n")
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
