package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// formatTechEffect converts a config.Effect into a human-readable display string.
func formatTechEffect(eff config.Effect) string {
	switch eff.Type {
	case "bonus":
		return fmt.Sprintf("+%.0f%% %s", eff.Value*100, formatBonusName(eff.Target))
	case "production":
		if eff.Value >= 0 {
			return fmt.Sprintf("+%.1f %s/tick", eff.Value, eff.Target)
		}
		return fmt.Sprintf("%.1f %s/tick", eff.Value, eff.Target)
	case "storage":
		return fmt.Sprintf("+%.0f %s storage", eff.Value, eff.Target)
	case "unlock":
		return fmt.Sprintf("unlock: %s", eff.Target)
	case "capacity":
		return fmt.Sprintf("+%.0f %s capacity", eff.Value, eff.Target)
	default:
		return fmt.Sprintf("%s: %s +%.1f", eff.Type, eff.Target, eff.Value)
	}
}

// researchProvider generates the full research overlay text. It renders three
// sections: (1) currently-in-progress tech with a tick progress bar,
// (2) active research bonuses, and (3) the full tech tree grouped by age.
// Tech visibility rules: researched=green ✓, in-progress=yellow ⟳,
// available=cyan ○, prereqs-met-but-age-locked=gray ○ (age locked),
// locked=gray • with prerequisite names shown.
func researchProvider(state game.GameState, _ int) string {
	var sb strings.Builder

	allTechs := config.TechByKey()
	techsByAge := config.TechsByAge()
	ageOrder := config.AgeOrder()
	ages := config.AgeByKey()

	// === Header ===
	fmt.Fprint(&sb, " [blue]research <key>  ·  research cancel  ·  research list[-]\n")
	fmt.Fprintf(&sb, " [gold]Progress: %d / %d techs researched[-]\n\n", state.Research.TotalResearched, len(state.Research.Techs))

	// === Currently Researching ===
	fmt.Fprintf(&sb, " [gold]═══ Currently Researching ═══[-]\n\n")
	if state.Research.CurrentTech != "" {
		done := state.Research.TotalTicks - state.Research.TicksLeft
		total := state.Research.TotalTicks
		var pct int
		if total > 0 {
			pct = done * 100 / total
		}
		bar := ProgressBar(float64(done), float64(total), 30)
		fmt.Fprintf(&sb, "  [yellow]⟳[-] %s\n", state.Research.CurrentTechName)
		// Bar + percentage already state the fraction; spend the remaining slot
		// on time-to-finish rather than restating it in ticks.
		fmt.Fprintf(&sb, "  %s %s left  (%d%%)\n", bar, formatTicks(state.Research.TicksLeft, state), pct)
	} else {
		sb.WriteString("  [gray]No research in progress — use: research <key>[-]\n")
	}

	// === Active Research Bonuses ===
	sb.WriteString("\n [gold]═══ Active Research Bonuses ═══[-]\n\n")
	if len(state.Research.Bonuses) == 0 {
		sb.WriteString("  [gray]No research bonuses yet[-]\n")
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
				fmt.Fprintf(&sb, "  [green]+%.0f%%[-]  %s\n\n", value*100, name)
			} else {
				fmt.Fprintf(&sb, "  [red]%.0f%%[-]  %s\n\n", value*100, name)
			}
		}
	}

	// === Available Now ===
	sb.WriteString(" [gold]═══ Available Now ═══[-]\n")

	knowledgeAmt := 0.0
	if rs, ok := state.Resources["knowledge"]; ok {
		knowledgeAmt = rs.Amount
	}

	for _, ageKey := range ageOrder {
		ageTechs, ok := techsByAge[ageKey]
		if !ok {
			continue
		}

		// Collect available techs for this age (not currently researching, not researched, available)
		var availNow []config.TechDef
		for _, tech := range ageTechs {
			ts, ok := state.Research.Techs[tech.Key]
			if !ok {
				continue
			}
			if ts.Available && !ts.Researched && state.Research.CurrentTech != tech.Key {
				availNow = append(availNow, tech)
			}
		}
		if len(availNow) == 0 {
			continue
		}

		sort.Slice(availNow, func(i, j int) bool {
			return availNow[i].Name < availNow[j].Name
		})

		ageName := ages[ageKey].Name
		fmt.Fprintf(&sb, "\n  [gold]── %s ──[-]\n", ageName)

		for _, tech := range availNow {
			ts := state.Research.Techs[tech.Key]
			def := allTechs[tech.Key]

			var affordStr string
			if knowledgeAmt < ts.Cost {
				need := ts.Cost - knowledgeAmt
				affordStr = fmt.Sprintf("  [red](need %.0f more)[-]", need)
			}

			fmt.Fprintf(&sb, "  [cyan]○[-]  %-24s [gray]%.0f knowledge · %s[-]%s\n",
				ts.Name, ts.Cost, formatTicks(def.ResearchTicks, state), affordStr)

			if ts.Description != "" {
				fmt.Fprintf(&sb, "     [gray]%s[-]\n", ts.Description)
			}

			if len(def.Effects) > 0 {
				var effStrs []string
				for _, eff := range def.Effects {
					effStrs = append(effStrs, formatTechEffect(eff))
				}
				fmt.Fprintf(&sb, "     [gray]Effects: %s[-]\n", strings.Join(effStrs, ", "))
			}
		}
	}

	// === Tech Tree ===
	sb.WriteString("\n [gold]═══ Tech Tree ═══[-]\n")

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

		if !hasVisible {
			// Check if all techs are locked and none visible — show gray locked header or skip entirely
			anyKnown := false
			for _, tech := range ageTechs {
				if _, ok := state.Research.Techs[tech.Key]; ok {
					anyKnown = true
					break
				}
			}
			if !anyKnown {
				continue
			}
			fmt.Fprintf(&sb, "\n  [gray]── %s (locked) ──[-]\n", ageName)
			continue
		}

		fmt.Fprintf(&sb, "\n  [gold]── %s ──[-]\n", ageName)

		// Sort techs by name for stable display
		sort.Slice(ageTechs, func(i, j int) bool {
			return ageTechs[i].Name < ageTechs[j].Name
		})

		for _, tech := range ageTechs {
			ts, ok := state.Research.Techs[tech.Key]
			if !ok {
				continue
			}

			def := allTechs[tech.Key]

			if ts.Researched {
				// Compact: show effects
				var effStrs []string
				for _, eff := range def.Effects {
					effStrs = append(effStrs, formatTechEffect(eff))
				}
				effStr := ""
				if len(effStrs) > 0 {
					effStr = "  [gray]" + strings.Join(effStrs, ", ") + "[-]"
				}
				fmt.Fprintf(&sb, "  [green]✓[-]  [green]%-24s[-]%s\n", ts.Name, effStr)

			} else if state.Research.CurrentTech == tech.Key {
				fmt.Fprintf(&sb, "  [yellow]⟳[-]  [yellow]%-24s[-]  [gray](in progress)[-]\n", ts.Name)

			} else if ts.Available {
				fmt.Fprintf(&sb, "  [cyan]○[-]  [cyan]%-24s[-]  [gray]%.0f knowledge — %s[-]", ts.Name, ts.Cost, formatTicks(def.ResearchTicks, state))
				// Show prereqs if any
				if len(ts.Prerequisites) > 0 {
					var prereqNames []string
					for _, prereq := range ts.Prerequisites {
						if p, ok := allTechs[prereq]; ok {
							prereqNames = append(prereqNames, p.Name)
						}
					}
					if len(prereqNames) > 0 {
						fmt.Fprintf(&sb, "  [gray]needs: %s[-]", strings.Join(prereqNames, ", "))
					}
				}
				sb.WriteString("\n")

			} else if ts.PrereqsMet {
				// Age-locked (prereqs met but age not yet reached)
				fmt.Fprintf(&sb, "  [gray]○  %-24s  (age locked)[-]\n", ts.Name)

			} else {
				// Locked — prereqs not met
				var prereqNames []string
				for _, prereq := range ts.Prerequisites {
					if p, ok := allTechs[prereq]; ok {
						prereqNames = append(prereqNames, p.Name)
					}
				}
				if len(prereqNames) > 0 {
					fmt.Fprintf(&sb, "  [gray]•  %-24s  needs: %s[-]\n", ts.Name, strings.Join(prereqNames, ", "))
				} else {
					fmt.Fprintf(&sb, "  [gray]•  %s[-]\n", ts.Name)
				}
			}
		}
	}

	// === Footer ===

	return sb.String()
}

// formatBonusName converts a research/milestone bonus key to a human-readable
// display name. Falls back to title-casing the key with spaces if no explicit
// mapping is defined (handles dynamically-generated resource rate keys like
// "iron_rate" → "Iron Rate").
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

// capitalize upper-cases the first letter of s.
func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
