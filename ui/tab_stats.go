package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/rivo/tview"

	"github.com/user/ageforge/game"
)

// StatsTab displays game statistics, milestones, active events, and prestige
type StatsTab struct {
	root       *tview.Flex
	statsTV    *tview.TextView
	milestoTV  *tview.TextView
	eventsTV   *tview.TextView
	prestigeTV *tview.TextView
}

// NewStatsTab creates the stats tab
func NewStatsTab() *StatsTab {
	t := &StatsTab{}

	t.statsTV = tview.NewTextView().SetDynamicColors(true).SetScrollable(true)
	t.statsTV.SetBorder(true).SetTitle(" Statistics ").SetTitleColor(ColorTitle)

	t.milestoTV = tview.NewTextView().SetDynamicColors(true).SetScrollable(true)
	t.milestoTV.SetBorder(true).SetTitle(" Milestones ").SetTitleColor(ColorTitle)

	t.eventsTV = tview.NewTextView().SetDynamicColors(true).SetScrollable(true)
	t.eventsTV.SetBorder(true).SetTitle(" Active Events ").SetTitleColor(ColorTitle)

	t.prestigeTV = tview.NewTextView().SetDynamicColors(true).SetScrollable(true)
	t.prestigeTV.SetBorder(true).SetTitle(" Prestige ").SetTitleColor(ColorTitle)

	// Left: stats, Right: events + prestige + milestones
	rightPanel := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(t.eventsTV, 6, 0, false).
		AddItem(t.prestigeTV, 12, 0, false).
		AddItem(t.milestoTV, 0, 1, false)

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
	t.refreshMilestones(state)
	t.refreshEvents(state)
}

func (t *StatsTab) refreshStats(state game.GameState) {
	var sb strings.Builder
	s := state.Stats

	fmt.Fprintf(&sb, " [gold]Play Time:[-]         %s\n", s.PlayTime.Truncate(1e9))
	fmt.Fprintf(&sb, " [gold]Total Ticks:[-]       %d\n", state.Tick)
	fmt.Fprintf(&sb, " [gold]Buildings Built:[-]    %d\n", s.TotalBuilt)
	fmt.Fprintf(&sb, " [gold]Villagers Recruited:[-] %d\n", s.TotalRecruited)
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

func (t *StatsTab) refreshMilestones(state game.GameState) {
	var sb strings.Builder
	ms := state.Milestones

	// Header: progress + title
	fmt.Fprintf(&sb, " [gold]Progress:[-] %d/%d", ms.CompletedCount, ms.TotalCount)
	if ms.CurrentTitle != "" {
		fmt.Fprintf(&sb, "  [yellow]\"%s\"[-]", ms.CurrentTitle)
	}
	sb.WriteString("\n\n")

	// Build chain lookup by category
	chainByCategory := make(map[string]game.ChainInfo)
	for _, chain := range ms.Chains {
		chainByCategory[chain.Category] = chain
	}

	// Group milestones by category
	categoryMilestones := make(map[string][]game.MilestoneInfo)
	categoryKeys := make(map[string][]string)
	for key, m := range ms.Milestones {
		categoryMilestones[m.Category] = append(categoryMilestones[m.Category], m)
		categoryKeys[m.Category] = append(categoryKeys[m.Category], key)
	}

	// Display categories in order
	catOrder := []string{"settlement", "builder", "scholar", "military", "ages"}
	catNames := map[string]string{
		"settlement": "Settlement",
		"builder":    "Builder",
		"scholar":    "Scholar",
		"military":   "Military",
		"ages":       "Ages",
	}

	for _, cat := range catOrder {
		milestones := categoryMilestones[cat]
		if len(milestones) == 0 {
			continue
		}

		catName := catNames[cat]

		// Category header with chain progress
		if chain, ok := chainByCategory[cat]; ok {
			chainBar := ProgressBar(float64(chain.CompletedCount), float64(chain.TotalCount), 8)
			if chain.Complete {
				fmt.Fprintf(&sb, " [green]★ %s[-] [%d/%d %s] [green]✓ %s[-]",
					catName, chain.CompletedCount, chain.TotalCount, chainBar, chain.Title)
				if chain.BoostActive {
					sb.WriteString(" [cyan]⚡BOOST[-]")
				}
			} else {
				fmt.Fprintf(&sb, " [gold]◆ %s[-] [%d/%d %s]",
					catName, chain.CompletedCount, chain.TotalCount, chainBar)
			}
		} else {
			fmt.Fprintf(&sb, " [gold]◆ %s[-]", catName)
		}
		sb.WriteString("\n")

		// Sort: completed first, then by name
		sort.Slice(milestones, func(i, j int) bool {
			if milestones[i].Completed != milestones[j].Completed {
				return milestones[i].Completed
			}
			return milestones[i].Name < milestones[j].Name
		})

		hiddenCount := 0
		for _, m := range milestones {
			if !m.Visible {
				hiddenCount++
				continue
			}

			if m.Completed {
				fmt.Fprintf(&sb, "   [green]✓ %s[-]", m.Name)
				if m.RewardText != "" {
					fmt.Fprintf(&sb, "  [cyan]%s[-]", m.RewardText)
				}
				sb.WriteString("\n")
			} else {
				fmt.Fprintf(&sb, "   [gray]○[-] [white]%s[-]\n", m.Name)
				fmt.Fprintf(&sb, "     [gray]%s[-]\n", m.Description)
				// Per-condition progress bars
				for _, p := range m.Progress {
					if p.Met {
						fmt.Fprintf(&sb, "     [green]✓ %s[-]\n", p.Label)
					} else {
						bar := ProgressBar(p.Current, p.Target, 10)
						fmt.Fprintf(&sb, "     [yellow]%.0f/%.0f %s %s[-]\n",
							p.Current, p.Target, bar, p.Label)
					}
				}
				if m.RewardText != "" {
					fmt.Fprintf(&sb, "     [gray]Reward: %s[-]\n", m.RewardText)
				}
			}
		}

		if hiddenCount > 0 {
			fmt.Fprintf(&sb, "   [gray]+ %d hidden milestone(s)[-]\n", hiddenCount)
		}
		sb.WriteString("\n")
	}

	t.milestoTV.SetText(sb.String())
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
