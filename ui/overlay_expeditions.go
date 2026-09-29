package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
)

// expeditionsProvider generates the Expeditions panel text: the civilian
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
		fmt.Fprintf(&sb, "\n [green]Expedition rewards %s[-]\n", textfmt.SignedPercent(mil.ExpeditionBonus))
	}

	// Standing orders. This panel is the scouting surface, so it has to say
	// whether scouting is being done for you: a built Geographic Society was
	// otherwise invisible here. Full status lives in the Factions panel.
	if auto := mil.AutoExpedition; auto.Active {
		switch {
		case auto.Starved:
			sb.WriteString("\n [yellow]Geographic Society: a party is due, but you cannot pay the expedition cost.[-]\n")
		default:
			fmt.Fprintf(&sb, "\n [cyan]Geographic Society:[-] next party goes out in %s [gray](details: factions)[-]\n",
				formatTicks(auto.TicksLeft, state))
		}
	} else {
		sb.WriteString("\n [gray]Build a Geographic Society (Industrial Age) to scout automatically.[-]\n")
	}

	// === Active expedition ===
	sb.WriteString("\n [gold]═══ Active expedition ═══[-]\n\n")
	if mil.ActiveScout == nil {
		sb.WriteString(" [gray]No expedition out right now.[-]\n")
	} else {
		writeActiveExpedition(&sb, "Scouting", mil.ActiveScout, state)
	}

	// === Available expeditions ===
	sb.WriteString("\n [gold]═══ Available expeditions ═══[-]\n\n")
	if !hasCategory(mil.Expeditions, game.ExpeditionScouting) {
		sb.WriteString(" [gray]No expeditions available yet.[-]\n")
		sb.WriteString(" [gray]Reach the Bronze Age to unlock more scouting.[-]\n")
	} else {
		writeExpeditionGroup(&sb, "Scouting", mil.Expeditions, game.ExpeditionScouting, state)
	}

	// === Loot history ===
	// Total loot lives with Expeditions (the scouting/loot surface), not the Army
	// panel: loot accrues from resolving expeditions and campaigns alike.
	sb.WriteString("\n [gold]═══ Loot history ═══[-]\n\n")
	if len(mil.TotalLoot) == 0 {
		sb.WriteString(" [gray]No loot collected yet.[-]\n")
		sb.WriteString(" [gray]Send an expedition: expedition <key>[-]\n")
	} else {
		sb.WriteString(" [gold]Total loot collected:[-]\n\n")
		lootKeys := make([]string, 0, len(mil.TotalLoot))
		for k := range mil.TotalLoot {
			lootKeys = append(lootKeys, k)
		}
		sort.Strings(lootKeys)
		for _, key := range lootKeys {
			amount := mil.TotalLoot[key]
			fmt.Fprintf(&sb, " %-14s [green]%s[-]\n", textfmt.Capitalize(game.ResourceName(key)), FormatNumber(amount))
		}
	}

	sb.WriteString("\n [gray]Commands: expedition <key>[-]\n")

	return sb.String()
}
