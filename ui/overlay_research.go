package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// researchProvider generates the research overlay text from the current game state.
// It mirrors the logic from ResearchTab.Refresh — same data, formatted as plain text.
func researchProvider(state game.GameState) string {
	var sb strings.Builder

	// === Current Research ===
	fmt.Fprintf(&sb, " [gold]═══ Current Research ═══[-]\n\n")
	if state.Research.CurrentTech != "" {
		fmt.Fprintf(&sb, " [yellow]Researching:[-] %s\n", state.Research.CurrentTechName)
		done := state.Research.TotalTicks - state.Research.TicksLeft
		bar := ProgressBar(float64(done), float64(state.Research.TotalTicks), 25)
		fmt.Fprintf(&sb, " %s %d/%d ticks\n", bar, done, state.Research.TotalTicks)
	} else {
		sb.WriteString(" [gray]No research in progress[-]\n")
		sb.WriteString(" [gray]Use 'research <key>' to start[-]\n")
	}
	fmt.Fprintf(&sb, "\n [gold]Total Researched:[-] %d techs\n", state.Research.TotalResearched)

	// === Active Bonuses ===
	sb.WriteString("\n [gold]═══ Active Bonuses ═══[-]\n\n")
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

	// === Tech Tree ===
	sb.WriteString("\n [gold]═══ Tech Tree ═══[-]\n")

	techsByAge := config.TechsByAge()
	ageOrder := config.AgeOrder()
	ages := config.AgeByKey()

	for _, ageKey := range ageOrder {
		ageTechs, ok := techsByAge[ageKey]
		if !ok {
			continue
		}

		ageName := ages[ageKey].Name

		// Check if any tech in this age is visible
		hasVisible := false
		for _, tech := range ageTechs {
			ts, ok := state.Research.Techs[tech.Key]
			if ok && (ts.Researched || ts.Available || ts.PrereqsMet) {
				hasVisible = true
				break
			}
		}

		if hasVisible {
			fmt.Fprintf(&sb, "\n [gold]── %s ──[-]\n", ageName)
		} else {
			fmt.Fprintf(&sb, "\n [gray]── %s (locked) ──[-]\n", ageName)
			continue
		}

		// Sort techs by name for stable display
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
				allTechs := config.TechByKey()
				var prereqNames []string
				for _, prereq := range ts.Prerequisites {
					if p, ok := allTechs[prereq]; ok {
						prereqNames = append(prereqNames, p.Name)
					}
				}
				if len(prereqNames) > 0 {
					fmt.Fprintf(&sb, " [red]needs: %s[-]", strings.Join(prereqNames, ", "))
				}
			}

			sb.WriteString("\n")

			// Show description for available or currently-researching techs
			if ts.Available || state.Research.CurrentTech == tech.Key {
				fmt.Fprintf(&sb, "   [gray]%s[-]\n", ts.Description)
			}
		}
	}

	sb.WriteString("\n [gray]Commands: research <key> | research cancel | research list[-]\n")

	return sb.String()
}
