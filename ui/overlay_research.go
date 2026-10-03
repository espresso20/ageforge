package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/theme"
)

// formatTechEffect names a config.Effect in glossary words: "+10% all
// production", "+0.5 food/tick", "+500 storage for every resource",
// "+10 housing", "unlocks Lumber Mill".
func formatTechEffect(eff config.Effect) string {
	switch eff.Type {
	case "bonus":
		return textfmt.SignedPercent(eff.Value) + " " + game.EffectTargetName(eff.Target)
	case "production":
		return rateNumber(eff.Value) + " " + game.ResourceName(eff.Target) + "/tick"
	case "storage":
		if eff.Target == "all" {
			return textfmt.Signed(eff.Value) + " storage for every resource"
		}
		return textfmt.Signed(eff.Value) + " " + game.ResourceName(eff.Target) + " storage"
	case "unlock":
		return "unlocks " + game.BuildingName(eff.Target)
	case "capacity":
		switch eff.Target {
		case "population":
			return textfmt.Signed(eff.Value) + " housing"
		case "military":
			return textfmt.Signed(eff.Value) + " soldier storage"
		}
		return textfmt.Signed(eff.Value) + " " + game.EffectTargetName(eff.Target)
	default:
		return game.EffectTargetName(eff.Target) + " " + textfmt.Signed(eff.Value)
	}
}

// rateNumber prints a signed per-tick amount with the decimals a small rate
// needs ("+0.05", "+0.5", "+2.5", "-3"): textfmt.RateValue without its
// trailing zeros. textfmt.Signed would round 0.05 up to "0.1".
func rateNumber(v float64) string {
	s := textfmt.RateValue(v)
	if strings.Contains(s, ".") && !strings.ContainsAny(s, "KMBTQ") {
		s = strings.TrimRight(strings.TrimRight(s, "0"), ".")
	}
	return s
}

// techLabel is a tech's name with the key the research command takes:
// "Bronze Working (bronze_working)".
func techLabel(name, key string) string {
	return name + " (" + key + ")"
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
	fmt.Fprint(&sb, " [label]research <key>  ·  research cancel  ·  research list[-]\n")
	fmt.Fprintf(&sb, " [gold]Progress: %d / %d techs researched[-]\n\n", state.Research.TotalResearched, len(state.Research.Techs))

	// === Currently Researching ===
	fmt.Fprintf(&sb, " [gold]═══ Researching now ═══[-]\n\n")
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
		sb.WriteString("  [gray]No research in progress. Start one with: research <key>[-]\n")
	}

	// === Research bonuses ===
	sb.WriteString("\n [gold]═══ Research bonuses ═══[-]\n\n")
	if len(state.Research.Bonuses) == 0 {
		sb.WriteString("  [gray]No research bonuses yet.[-]\n")
	} else {
		keys := make([]string, 0, len(state.Research.Bonuses))
		for k := range state.Research.Bonuses {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, key := range keys {
			value := state.Research.Bonuses[key]
			name := formatBonusName(key)
			color := "green"
			if value <= 0 {
				color = "red"
			}
			fmt.Fprintf(&sb, "  [%s]%s[-]  %s\n\n", color, textfmt.SignedPercent(value), name)
		}
	}

	// === Available now ===
	sb.WriteString(" [gold]═══ Available now ═══[-]\n")

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
				affordStr = fmt.Sprintf("  [red](need %s more knowledge)[-]", FormatNumber(need))
			}

			fmt.Fprintf(&sb, "  [cyan]○[-]  %-40s [gray]%s knowledge · %s[-]%s\n",
				techLabel(ts.Name, tech.Key), FormatNumber(ts.Cost), formatTicks(masteryTicks(def.ResearchTicks, state), state), affordStr)

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

	// === Tech tree ===
	sb.WriteString("\n [gold]═══ Tech tree ═══[-]\n")

	// Ages the player cannot see named yet are not listed, only counted: no
	// age names and no tech names from past the next age (spoilers.go).
	sight := game.SightOf(&state)
	laterTechs := 0
	for _, ageKey := range ageOrder {
		ageTechs, ok := techsByAge[ageKey]
		if !ok {
			continue
		}
		if !sight.Age(ageKey) {
			laterTechs += len(ageTechs)
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
				fmt.Fprintf(&sb, "  [cyan]○[-]  [cyan]%-40s[-]  [gray]%s knowledge · %s[-]", techLabel(ts.Name, tech.Key), FormatNumber(ts.Cost), formatTicks(masteryTicks(def.ResearchTicks, state), state))
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

	if laterTechs > 0 {
		sb.WriteString("\n  " + theme.Paint(theme.RoleDim, fmt.Sprintf("── Later ages (locked): %s ──", textfmt.Count(laterTechs, "more tech", "more techs"))) + "\n")
	}

	// === Footer ===

	return sb.String()
}

// formatBonusName names a research/milestone bonus key in glossary words, as
// a label: "All production", "Worker output", "Iron production", "Housing".
func formatBonusName(key string) string {
	return textfmt.Capitalize(game.EffectTargetName(key))
}

// capitalize upper-cases the first letter of s.
func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
