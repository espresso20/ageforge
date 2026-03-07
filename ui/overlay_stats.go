package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// statsProvider generates the stats overlay text from the current game state.
// Covers: game statistics, active epoch events, and prestige upgrades/points.
func statsProvider(state game.GameState) string {
	var sb strings.Builder

	// ─── Statistics ───
	sb.WriteString("[gold]═══ Statistics ═══[-]\n\n")
	s := state.Stats

	fmt.Fprintf(&sb, " [gold]Play Time:[-]          %s\n", s.PlayTime.Truncate(1e9))
	fmt.Fprintf(&sb, " [gold]Total Ticks:[-]        %d\n", state.Tick)
	fmt.Fprintf(&sb, " [gold]Buildings Built:[-]    %d\n", s.TotalBuilt)
	fmt.Fprintf(&sb, " [gold]Workers Recruited:[-]  %d\n", s.TotalRecruited)
	fmt.Fprintf(&sb, " [gold]Techs Researched:[-]   %d\n", state.Research.TotalResearched)
	fmt.Fprintf(&sb, " [gold]Expeditions Done:[-]   %d\n", state.Military.CompletedCount)

	sb.WriteString("\n [gold]Ages Reached:[-]\n")
	for _, age := range s.AgesReached {
		fmt.Fprintf(&sb, "   [green]✓[-] %s\n", age)
	}

	sb.WriteString("\n [gold]Total Gathered:[-]\n")
	gKeys := make([]string, 0, len(s.TotalGathered))
	for k := range s.TotalGathered {
		gKeys = append(gKeys, k)
	}
	sort.Strings(gKeys)
	for _, k := range gKeys {
		fmt.Fprintf(&sb, "   %-12s %s\n", k, FormatNumber(s.TotalGathered[k]))
	}

	// Epoch & Legacy section
	sb.WriteString("\n [yellow]── Epoch & Legacy ──[-]\n")

	epochDisplay := state.EpochIcon + " " + state.EpochName
	if strings.TrimSpace(epochDisplay) == "" {
		epochDisplay = "—"
	}
	fmt.Fprintf(&sb, " %-20s [cyan]%s[-]\n", "Current Epoch:", epochDisplay)

	epochsSurvivedCount := len(state.CatastropheHistory)
	fmt.Fprintf(&sb, " %-20s %d\n", "Epochs Survived:", epochsSurvivedCount)

	succumbedCount := 0
	for _, active := range state.LegacyBonuses {
		if active {
			succumbedCount++
		}
	}
	totalCatastrophes := len(state.CatastropheHistory)
	enduredCount := totalCatastrophes - succumbedCount
	fmt.Fprintf(&sb, " %-20s %d  (Endured: %d  Succumbed: %d)\n",
		"Catastrophes:", totalCatastrophes, enduredCount, succumbedCount)

	// Legacy Bonuses
	sb.WriteString("\n [gold]Legacy Bonuses:[-]\n")
	if len(state.LegacyBonuses) == 0 {
		sb.WriteString("  [gray]None[-]\n")
	} else {
		activeEpochs := make([]string, 0, len(state.LegacyBonuses))
		for epochKey, active := range state.LegacyBonuses {
			if active {
				activeEpochs = append(activeEpochs, epochKey)
			}
		}
		sort.Strings(activeEpochs)

		if len(activeEpochs) == 0 {
			sb.WriteString("  [gray]None[-]\n")
		} else {
			epochByKey := config.EpochByKey()
			for _, epochKey := range activeEpochs {
				bonuses := config.LegacyBonusForEpoch(epochKey)
				epochDef, hasEpoch := epochByKey[epochKey]
				epochLabel := epochKey
				if hasEpoch {
					epochLabel = epochDef.Name
				}

				bonusKeys := make([]string, 0, len(bonuses))
				for k := range bonuses {
					bonusKeys = append(bonusKeys, k)
				}
				sort.Strings(bonusKeys)
				parts := make([]string, 0, len(bonusKeys))
				for _, k := range bonusKeys {
					parts = append(parts, fmt.Sprintf("%s +%.0f%%", k, bonuses[k]*100))
				}
				fmt.Fprintf(&sb, "  %-16s %s\n", epochLabel+":", strings.Join(parts, ",  "))
			}
		}
	}

	// Milestone summary hint
	ms := state.Milestones
	fmt.Fprintf(&sb, "\n [gray]Milestones: %d/%d — type [white]milestones[-][gray] to view[-]\n",
		ms.CompletedCount, ms.TotalCount)

	// ─── Active Events ───
	sb.WriteString("\n[gold]═══ Active Events ═══[-]\n\n")
	if len(state.ActiveEvents) == 0 {
		sb.WriteString(" [gray]No active events[-]\n")
	} else {
		for _, evt := range state.ActiveEvents {
			fmt.Fprintf(&sb, " [yellow]⚡[-] [yellow]%s[-] (%d ticks left)\n", evt.Name, evt.TicksLeft)
		}
	}

	// ─── Prestige ───
	sb.WriteString("\n[gold]═══ Prestige ═══[-]\n\n")
	p := state.Prestige

	fmt.Fprintf(&sb, " [gold]Level:[-] [cyan]%d[-]", p.Level)
	if p.PassiveBonus > 0 {
		fmt.Fprintf(&sb, "  [green]+%.0f%% production[-]", p.PassiveBonus*100)
	}
	sb.WriteString("\n")
	fmt.Fprintf(&sb, " [gold]Points:[-] [cyan]%d[-] available / %d total\n", p.Available, p.TotalEarned)

	if p.CanPrestige {
		fmt.Fprintf(&sb, " [green]Can prestige for %d pts![-]\n", p.PendingPoints)
	} else if p.Level == 0 {
		sb.WriteString(" [gray]Reach Medieval Age to prestige[-]\n")
	} else {
		sb.WriteString(" [yellow]Reach Medieval Age to prestige again[-]\n")
	}

	upgradeKeys := []string{
		"gather_boost", "storage_bonus", "research_speed", "military_power",
		"starting_food", "starting_wood", "population_cap", "expedition_loot",
	}
	hasPurchased := false
	for _, key := range upgradeKeys {
		u, ok := p.Upgrades[key]
		if !ok || u.Tier == 0 {
			continue
		}
		if !hasPurchased {
			sb.WriteString("\n [gold]Upgrades:[-]\n")
			hasPurchased = true
		}
		bar := ProgressBar(float64(u.Tier), float64(u.MaxTier), 5)
		fmt.Fprintf(&sb, "  %s %s [green]%s[-]\n", u.Name, bar, u.Effect)
	}

	if !hasPurchased && p.Level > 0 {
		sb.WriteString("\n [gray]No upgrades purchased yet[-]\n")
		sb.WriteString(" [gray]Type 'prestige shop' to browse[-]\n")
	}

	return sb.String()
}
