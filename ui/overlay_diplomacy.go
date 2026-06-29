package ui

import (
	"fmt"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// diplomacyProvider generates the full-panel diplomacy overlay from the current
// game state. It renders all six canonical factions (in BaseFactions order so
// undiscovered factions still appear as locked teasers), showing opinion, status,
// the active trade-rate bonus, the distance to the next status tier, and the
// commands available for each. Read-only over state.Diplomacy — safe for empty
// or zero state.
func diplomacyProvider(state game.GameState, _ int) string {
	var sb strings.Builder
	factions := state.Diplomacy.Factions // may be nil; indexing nil maps is safe

	// Count how many factions the player has actually discovered so we can show
	// the onboarding hint when diplomacy hasn't started yet.
	discovered := 0
	for _, f := range factions {
		if f.Discovered {
			discovered++
		}
	}

	fmt.Fprintf(&sb, " [gold]═══ Diplomacy — Factions ═══[-]\n\n")
	sb.WriteString(" [gray]Tip: assign workers to an Embassy (Colonial) or Grand Embassy (Industrial) to passively raise opinion.[-]\n\n")

	if discovered == 0 {
		sb.WriteString(" [gray]Reach the Colonial Age and build an Embassy to begin diplomacy.[-]\n\n")
	}

	ages := config.AgeByKey()

	for _, def := range config.BaseFactions() {
		// Name + personality + specialty resource header line.
		fmt.Fprintf(&sb, " [cyan]%s[-]  [%s]%s[-]  [gray](%s)[-]\n",
			def.Name, personalityColor(def.Personality), def.Personality, def.Specialty)

		f, ok := factions[def.Key]
		if !ok || !f.Discovered {
			// Locked teaser: dim line with the age the faction is reachable.
			ageName := def.MinAge
			if a, found := ages[def.MinAge]; found {
				ageName = a.Name
			}
			fmt.Fprintf(&sb, "   [gray]??? [Undiscovered — reach %s][-]\n\n", ageName)
			continue
		}

		// Backstory snippet (truncated to keep the panel tidy).
		if def.Backstory != "" {
			fmt.Fprintf(&sb, "   [gray]%s[-]\n", truncate(def.Backstory, 88))
		}

		// War banner takes precedence — it's the headline state when active.
		if f.AtWar {
			sb.WriteString("   [red]⚔ AT WAR — raids incoming. 'diplomacy tribute " + def.Key + "' to sue for peace.[-]\n")
		}

		// Opinion bar across the -100..100 range. Clamp first to stay panic-free.
		opinion := f.Opinion
		if opinion > 100 {
			opinion = 100
		} else if opinion < -100 {
			opinion = -100
		}
		const barCells = 20
		filled := (opinion + 100) * barCells / 200
		if filled < 0 {
			filled = 0
		} else if filled > barCells {
			filled = barCells
		}
		opinionColor := "white"
		switch {
		case f.Opinion >= 50:
			opinionColor = "green"
		case f.Opinion >= 25:
			opinionColor = "cyan"
		case f.Opinion < 0:
			opinionColor = "red"
		}
		bar := strings.Repeat("█", filled) + strings.Repeat("░", barCells-filled)
		fmt.Fprintf(&sb, "   Opinion: [%s]%4d[-]  [%s]%s[-]\n", opinionColor, f.Opinion, opinionColor, bar)

		// Status label, color-coded.
		statusColor := "white"
		switch f.Status {
		case "allied":
			statusColor = "green"
		case "friendly":
			statusColor = "cyan"
		case "rival":
			statusColor = "red"
		case "embargo":
			statusColor = "yellow"
		}
		// Active bonus + trade-rate modifier (only allied specialty trades get it).
		bonus := "[gray]no active bonus[-]"
		if f.Status == "allied" && f.TradeBonus > 0 {
			bonus = fmt.Sprintf("[green]+%.0f%% %s trades[-]", f.TradeBonus*100, f.Specialty)
		}
		fmt.Fprintf(&sb, "   Status:  [%s][%s][-]  %s  [gray](%d trades)[-]\n",
			statusColor, f.Status, bonus, f.TradeCount)

		// Threshold indicator — distance to the next status tier.
		fmt.Fprintf(&sb, "   %s\n", diplomacyThreshold(f.Status, f.Opinion))

		// Lent-worker status, if this civ has workers on loan with you.
		if f.LentWorkers > 0 {
			if f.LentPerm {
				fmt.Fprintf(&sb, "   [green]↳ %d workers on loan (permanent)[-]\n", f.LentWorkers)
			} else {
				fmt.Fprintf(&sb, "   [green]↳ %d workers on loan (temporary)[-]\n", f.LentWorkers)
			}
		}

		// Action hint: the diplomacy commands available given current status.
		switch {
		case f.AtWar:
			fmt.Fprintf(&sb, "   [gray]diplomacy tribute %s (sue for peace) · or wait them out[-]\n\n", def.Key)
		case f.Status == "allied":
			fmt.Fprintf(&sb, "   [gray]diplomacy rival/embargo/neutral %s[-]\n\n", def.Key)
		default:
			fmt.Fprintf(&sb, "   [gray]diplomacy gift %s (200g, +15) · ally/rival/embargo/neutral %s[-]\n\n",
				def.Key, def.Key)
		}
	}

	sb.WriteString(" [gray]Commands: diplomacy ally/rival/embargo/gift/neutral/tribute/raid <civ>[-]\n")

	return sb.String()
}

// personalityColor maps a civilization personality to a tview color for the
// overlay header. Unknown personalities render gray (panic-safe default).
func personalityColor(personality string) string {
	switch personality {
	case "aggressive":
		return "red"
	case "peaceful":
		return "green"
	case "mercantile":
		return "yellow"
	case "isolationist":
		return "lightblue"
	default:
		return "gray"
	}
}

// truncate shortens s to at most max runes, appending an ellipsis when cut.
// Operates on runes so multibyte backstory text is never split mid-character.
func truncate(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	if max <= 1 {
		return string(r[:max])
	}
	return string(r[:max-1]) + "…"
}

// diplomacyThreshold renders the distance-to-next-tier indicator for a faction
// given its current status and opinion. Hostile statuses (rival/embargo) decay
// toward neutral, so they report "decaying" rather than a climb target.
func diplomacyThreshold(status string, opinion int) string {
	switch status {
	case "allied":
		return "[gray](maxed)[-]"
	case "rival", "embargo":
		return "[yellow](opinion decaying)[-]"
	}
	// neutral / friendly: climbing toward the next eligibility gate.
	switch {
	case opinion < 25:
		return fmt.Sprintf("[gray](+%d to friendly)[-]", 25-opinion)
	case opinion < 50:
		return fmt.Sprintf("[gray](+%d to ally-eligible)[-]", 50-opinion)
	default:
		return "[gray](ally-eligible — 500g)[-]"
	}
}
