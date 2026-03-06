package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/espresso20/ageforge/game"
)

// milestonesProvider generates the milestones overlay text from the current game state.
// This is lifted from StatsTab.refreshMilestones — identical logic, returning a string
// instead of calling SetText on a TextView.
func milestonesProvider(state game.GameState) string {
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
	for _, m := range ms.Milestones {
		categoryMilestones[m.Category] = append(categoryMilestones[m.Category], m)
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

	return sb.String()
}
