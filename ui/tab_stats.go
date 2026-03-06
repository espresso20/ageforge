package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/rivo/tview"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// StatsTab displays game statistics, active events, and prestige
type StatsTab struct {
	root       *tview.Flex
	statsTV    *tview.TextView
	eventsTV   *tview.TextView
	prestigeTV *tview.TextView
}

// NewStatsTab creates the stats tab
func NewStatsTab() *StatsTab {
	t := &StatsTab{}

	t.statsTV = tview.NewTextView().SetDynamicColors(true).SetScrollable(true)
	t.statsTV.SetBorder(true).SetTitle(" Statistics ").SetTitleColor(ColorTitle)

	t.eventsTV = tview.NewTextView().SetDynamicColors(true).SetScrollable(true)
	t.eventsTV.SetBorder(true).SetTitle(" Active Events ").SetTitleColor(ColorTitle)

	t.prestigeTV = tview.NewTextView().SetDynamicColors(true).SetScrollable(true)
	t.prestigeTV.SetBorder(true).SetTitle(" Prestige ").SetTitleColor(ColorTitle)

	// Left: stats, Right: events + prestige
	rightPanel := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(t.eventsTV, 6, 0, false).
		AddItem(t.prestigeTV, 0, 1, false)

	t.root = tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(t.statsTV, 0, 1, false).
		AddItem(rightPanel, 0, 1, false)

	return t
}

// Root returns the root primitive
func (t *StatsTab) Root() tview.Primitive {
	return t.root
}

// Refresh updates the stats tab
func (t *StatsTab) Refresh(state game.GameState) {
	t.refreshStats(state)
	t.refreshPrestige(state)
	t.refreshEvents(state)
}

func (t *StatsTab) refreshStats(state game.GameState) {
	var sb strings.Builder
	s := state.Stats

	fmt.Fprintf(&sb, " [gold]Play Time:[-]         %s\n", s.PlayTime.Truncate(1e9))
	fmt.Fprintf(&sb, " [gold]Total Ticks:[-]       %d\n", state.Tick)
	fmt.Fprintf(&sb, " [gold]Buildings Built:[-]    %d\n", s.TotalBuilt)
	fmt.Fprintf(&sb, " [gold]Workers Recruited:[-]   %d\n", s.TotalRecruited)
	fmt.Fprintf(&sb, " [gold]Techs Researched:[-]  %d\n", state.Research.TotalResearched)
	fmt.Fprintf(&sb, " [gold]Expeditions Done:[-]  %d\n", state.Military.CompletedCount)

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
	if epochDisplay == " " {
		epochDisplay = "—"
	}
	fmt.Fprintf(&sb, " %-20s [cyan]%s[-]\n", "Current Epoch:", epochDisplay)

	// Epochs survived: EpochSurvived is a bool for the current epoch;
	// count of fully-survived past epochs comes from LegacyBonuses entries (each true = one succumb = one epoch done)
	// plus EpochSurvived bool (current epoch endured catastrophe).
	// The spec says count truthy entries in EpochSurvived — but since it's a bool, we count it as 0 or 1.
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
		// Collect epoch keys that have active legacy bonuses, sort for stable output
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

				// Format bonus pairs: "resource +X%, resource +X%"
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

	t.statsTV.SetText(sb.String())
}

func (t *StatsTab) refreshPrestige(state game.GameState) {
	var sb strings.Builder
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

	// Show purchased upgrades
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

	t.prestigeTV.SetText(sb.String())
}


func (t *StatsTab) refreshEvents(state game.GameState) {
	var sb strings.Builder

	if len(state.ActiveEvents) == 0 {
		sb.WriteString(" [gray]No active events[-]\n")
	} else {
		for _, evt := range state.ActiveEvents {
			fmt.Fprintf(&sb, " [yellow]⚡[-] [yellow]%s[-] (%d ticks left)\n", evt.Name, evt.TicksLeft)
		}
	}

	t.eventsTV.SetText(sb.String())
}
