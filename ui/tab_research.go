package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/rivo/tview"

	"github.com/user/ageforge/config"
	"github.com/user/ageforge/game"
)

// ResearchTab displays the tech tree and research progress
type ResearchTab struct {
	root      *tview.Flex
	treeTV    *tview.TextView
	detailTV  *tview.TextView
	bonusesTV *tview.TextView
}

// NewResearchTab creates the research tab
func NewResearchTab() *ResearchTab {
	t := &ResearchTab{}

	t.treeTV = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true)
	t.treeTV.SetBorder(true).SetTitle(" Tech Tree ").SetTitleColor(ColorTitle)

	t.detailTV = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true)
	t.detailTV.SetBorder(true).SetTitle(" Current Research ").SetTitleColor(ColorTitle)

	t.bonusesTV = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true)
	t.bonusesTV.SetBorder(true).SetTitle(" Active Bonuses ").SetTitleColor(ColorTitle)

	// Right panel: current research + bonuses
	rightPanel := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(t.detailTV, 8, 0, false).
		AddItem(t.bonusesTV, 0, 1, false)

	t.root = tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(t.treeTV, 0, 2, false).
		AddItem(rightPanel, 0, 1, false)

	return t
}

// Root returns the root primitive
func (t *ResearchTab) Root() tview.Primitive {
	return t.root
}

// Refresh updates the research tab with current state
func (t *ResearchTab) Refresh(state game.GameState) {
	t.refreshTree(state)
	t.refreshDetail(state)
	t.refreshBonuses(state)
}

func (t *ResearchTab) refreshTree(state game.GameState) {
	var sb strings.Builder

	// Group techs by age
	techsByAge := config.TechsByAge()
	ageOrder := config.AgeOrder()
	ages := config.AgeByKey()

	for _, ageKey := range ageOrder {
		ageTechs, ok := techsByAge[ageKey]
		if !ok {
			continue
		}

		ageName := ages[ageKey].Name

		// Check if any techs are visible
		hasVisible := false
		for _, tech := range ageTechs {
			ts, ok := state.Research.Techs[tech.Key]
			if ok && (ts.Researched || ts.Available || ts.PrereqsMet) {
				hasVisible = true
				break
			}
		}

		if hasVisible {
			fmt.Fprintf(&sb, "\n [gold]═══ %s ═══[-]\n", ageName)
		} else {
			fmt.Fprintf(&sb, "\n [gray]═══ %s (locked) ═══[-]\n", ageName)
			continue
		}

		// Sort techs by name for stability
		sort.Slice(ageTechs, func(i, j int) bool {
			return ageTechs[i].Name < ageTechs[j].Name
		})

		for _, tech := range ageTechs {
			ts, ok := state.Research.Techs[tech.Key]
			if !ok {
				continue
			}

			var icon, color string
			if ts.Researched {
				icon = "[green]✓[-]"
				color = "green"
			} else if state.Research.CurrentTech == tech.Key {
				icon = "[yellow]⟳[-]"
				color = "yellow"
			} else if ts.Available {
				icon = "[cyan]○[-]"
				color = "cyan"
			} else if ts.PrereqsMet {
				icon = "[gray]○[-]"
				color = "gray"
			} else {
				icon = "[gray]•[-]"
				color = "gray"
			}

			costStr := ""
			if !ts.Researched {
				costStr = fmt.Sprintf(" [gray](%.0f knowledge)[-]", ts.Cost)
			}

			fmt.Fprintf(&sb, " %s [%s]%-22s[-]%s", icon, color, ts.Name, costStr)

			// Show prerequisites if not met
			if !ts.PrereqsMet && !ts.Researched {
				var prereqNames []string
				allTechs := config.TechByKey()
				for _, prereq := range ts.Prerequisites {
					if p, ok := allTechs[prereq]; ok {
						prereqNames = append(prereqNames, p.Name)
					}
				}
				fmt.Fprintf(&sb, " [red]needs: %s[-]", strings.Join(prereqNames, ", "))
			}

			sb.WriteString("\n")

			// Show description for available/researching techs
			if ts.Available || state.Research.CurrentTech == tech.Key {
				fmt.Fprintf(&sb, "   [gray]%s[-]\n", ts.Description)
			}
		}
	}

	sb.WriteString("\n [gray]Commands: research <key> | research cancel | research list[-]\n")

	t.treeTV.SetText(sb.String())
}

func (t *ResearchTab) refreshDetail(state game.GameState) {
	var sb strings.Builder

	if state.Research.CurrentTech != "" {
		fmt.Fprintf(&sb, " [yellow]Researching:[-] %s\n", state.Research.CurrentTechName)
		done := state.Research.TotalTicks - state.Research.TicksLeft
		bar := ProgressBar(float64(done), float64(state.Research.TotalTicks), 25)
		fmt.Fprintf(&sb, " %s %d/%d ticks\n", bar, done, state.Research.TotalTicks)
	} else {
		sb.WriteString(" [gray]No research in progress[-]\n")
		sb.WriteString(" [gray]Use 'research <key>' to start[-]\n")
	}

	fmt.Fprintf(&sb, "\n [gold]Total Researched:[-] %d techs", state.Research.TotalResearched)

	t.detailTV.SetText(sb.String())
}

func (t *ResearchTab) refreshBonuses(state game.GameState) {
	var sb strings.Builder

	if len(state.Research.Bonuses) == 0 {
		sb.WriteString(" [gray]No research bonuses yet[-]\n")
		sb.WriteString(" [gray]Research techs to earn bonuses![-]\n")
	} else {
		keys := make([]string, 0, len(state.Research.Bonuses))
		for k := range state.Research.Bonuses {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, key := range keys {
			value := state.Research.Bonuses[key]
			name := formatBonusName(key)
			if value > 0 {
				fmt.Fprintf(&sb, " [green]+%.0f%%[-] %s\n", value*100, name)
			} else {
				fmt.Fprintf(&sb, " [red]%.0f%%[-] %s\n", value*100, name)
			}
		}
	}

	t.bonusesTV.SetText(sb.String())
}

// formatBonusName converts a bonus key to a display name
func formatBonusName(key string) string {
	switch key {
	case "gather_rate":
		return "Gather Rate"
	case "production_all":
		return "All Production"
	case "military_power":
		return "Military Power"
	case "expedition_reward":
		return "Expedition Rewards"
	case "research_speed":
		return "Research Speed"
	case "build_cost":
		return "Build Cost"
	case "population":
		return "Population Cap"
	}
	// Convert key_rate pattern
	if strings.HasSuffix(key, "_rate") {
		res := strings.TrimSuffix(key, "_rate")
		return capitalize(res) + " Rate"
	}
	parts := strings.Split(key, "_")
	for i, p := range parts {
		parts[i] = capitalize(p)
	}
	return strings.Join(parts, " ")
}

func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
