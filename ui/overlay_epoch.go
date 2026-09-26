package ui

import (
	"fmt"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
)

// epochProvider generates the epoch overlay text from the current game state.
// Shows current epoch status, epoch history, legacy bonuses, and civilisation log.
func epochProvider(state game.GameState, _ int) string {
	var sb strings.Builder

	sb.WriteString("[gold]═══ Epoch ═══[-]\n\n")

	epochProviderCurrentEpoch(&sb, state)
	sb.WriteString("\n")
	epochProviderHistory(&sb, state)
	sb.WriteString("\n")
	epochProviderLegacyBonuses(&sb, state)
	sb.WriteString("\n")
	epochProviderCivilizationLog(&sb, state)

	return sb.String()
}

// epochProviderCurrentEpoch renders the current epoch status section.
func epochProviderCurrentEpoch(sb *strings.Builder, state game.GameState) {
	sb.WriteString("[gold]── Current Epoch ──[-]\n\n")

	if state.EpochKey == "" {
		sb.WriteString(" [gray]No epoch data yet.[-]\n")
		return
	}

	epochs := config.EpochByKey()
	ep, ok := epochs[state.EpochKey]
	if !ok {
		fmt.Fprintf(sb, " [gold]%s %s[-]\n", state.EpochIcon, state.EpochName)
	} else {
		ageNames := make([]string, 0, len(ep.Ages))
		for _, a := range ep.Ages {
			ageNames = append(ageNames, epochOverlayFormatAgeKey(a))
		}
		fmt.Fprintf(sb, " %s%s %s[-]   Ages: %s\n",
			theme.NameTag(ep.Color), ep.Icon, ep.Name,
			strings.Join(ageNames, " · "))
		fmt.Fprintf(sb, " [gray]Primary resource: %s   Energy: %s[-]\n",
			ep.PrimaryResource, ep.EnergyResource)
	}

	sb.WriteString("\n")

	// Current epoch event
	currentEvent := findEpochEvent(state.EpochEventHistory, state.EpochKey)
	if currentEvent == nil {
		sb.WriteString(" Epoch event: [gray]No event this epoch[-]\n")
	} else {
		evColor := epochEventColor(currentEvent.EventType)
		if currentEvent.EventType == "catastrophe" {
			fmt.Fprintf(sb, " Epoch event: [%s]%s[-]   [red]— catastrophe[-]\n",
				evColor, currentEvent.EventName)
		} else {
			evDefs := config.EpochEventByKey()
			flavorStr := ""
			durationStr := ""
			if evDef, found := evDefs[currentEvent.EventKey]; found {
				if evDef.FlavorText != "" {
					flavorStr = evDef.FlavorText
				}
				if evDef.Duration > 0 {
					durationStr = fmt.Sprintf("   %s duration", formatTicks(evDef.Duration, state))
				}
			}
			fmt.Fprintf(sb, " Epoch event: [%s]%s[-]%s\n", evColor, currentEvent.EventName, durationStr)
			if flavorStr != "" {
				fmt.Fprintf(sb, " [gray]  %s[-]\n", flavorStr)
			}
		}
	}

	sb.WriteString("\n")

	// Catastrophe status
	switch {
	case state.PendingCatastrophe != "":
		sb.WriteString(" Catastrophe: [red]PENDING — type 'catastrophe' to choose Endure or Succumb[-]\n")
		sb.WriteString(" [gray]  Advancing and prestige are blocked until you decide.[-]\n")
	case !config.CatastropheAllowed(state.EpochKey):
		gate := config.EpochByKey()[config.CatastropheGateEpoch].Name
		fmt.Fprintf(sb, " Catastrophe: [gray]none before the %s[-]\n", gate)
	default:
		if r := latestCatastrophe(state.EpochEventHistory, state.EpochKey); r != nil {
			fmt.Fprintf(sb, " Catastrophe: %s\n", catastropheOutcomeLabel(r.Outcome))
		} else {
			sb.WriteString(" Catastrophe: [gray]not yet triggered[-]\n")
		}
	}

	// Odds at the next transition (same numbers the `catastrophe` command shows).
	if o := state.CatastropheOutlook; o.Possible {
		fmt.Fprintf(sb, " Next transition (%s): [yellow]%.0f%% catastrophe chance[-] [gray](%s; more faith, lower odds)[-]\n",
			config.EpochByKey()[o.NextEpochKey].Name, o.Probability*100, o.Tier)
	}

	sb.WriteString("\n")
	sb.WriteString(" To invoke voluntarily: [gray]catastrophe invoke[-]\n")
}

// epochProviderHistory renders the epoch history section.
func epochProviderHistory(sb *strings.Builder, state game.GameState) {
	sb.WriteString(" [yellow]── Epoch History ──[-]\n")

	allEpochs := config.Epochs()
	currentEpochOrder := -1
	epochByKey := config.EpochByKey()
	if ep, ok := epochByKey[state.EpochKey]; ok {
		currentEpochOrder = ep.Order
	}

	for _, ep := range allEpochs {
		if ep.Order > currentEpochOrder {
			break
		}

		isCurrent := ep.Key == state.EpochKey
		record := findEpochEvent(state.EpochEventHistory, ep.Key)

		var line strings.Builder
		fmt.Fprintf(&line, "   %s%s %s[-]", theme.NameTag(ep.Color), ep.Icon, ep.Name)

		if isCurrent {
			line.WriteString("   [gray][current][-]")
		} else {
			if record != nil {
				evColor := epochEventColor(record.EventType)
				fmt.Fprintf(&line, "   [%s]%s[-]", evColor, record.EventName)
			} else {
				line.WriteString("   [gray]no event[-]")
			}

			if r := latestCatastrophe(state.EpochEventHistory, ep.Key); r != nil {
				line.WriteString("   " + catastropheOutcomeLabel(r.Outcome))
			}
		}

		sb.WriteString(line.String())
		sb.WriteString("\n")
	}
}

// epochProviderLegacyBonuses renders the legacy bonuses section.
func epochProviderLegacyBonuses(sb *strings.Builder, state game.GameState) {
	sb.WriteString(" [yellow]── Legacy Bonuses (from Succumb) ──[-]\n")

	if len(state.LegacyBonuses) == 0 {
		sb.WriteString("   [gray]None yet[-]\n")
		return
	}

	allEpochs := config.Epochs()
	anyShown := false
	for _, ep := range allEpochs {
		if !state.LegacyBonuses[ep.Key] {
			continue
		}
		bonuses := config.LegacyBonusForEpoch(ep.Key)
		if len(bonuses) == 0 {
			continue
		}
		anyShown = true

		var parts []string
		for res, pct := range bonuses {
			parts = append(parts, fmt.Sprintf("%s +%.0f%%", res, pct*100))
		}
		fmt.Fprintf(sb, "   %s%s %s:[-]  %s\n",
			theme.NameTag(ep.Color), ep.Icon, ep.Name,
			strings.Join(parts, ", "))
	}

	if !anyShown {
		sb.WriteString("   [gray]None yet[-]\n")
	}
}

// epochProviderCivilizationLog renders the civilisation catastrophe log section.
func epochProviderCivilizationLog(sb *strings.Builder, state game.GameState) {
	sb.WriteString(" [yellow]── Civilization Log ──[-]\n")

	if len(state.CatastropheHistory) == 0 {
		sb.WriteString("   [gray]No catastrophes yet[-]\n")
		return
	}

	for _, entry := range state.CatastropheHistory {
		fmt.Fprintf(sb, "   · %s\n", entry)
	}
}

// findEpochEvent returns the most recent EpochEventRecord for the given epoch key, or nil.
func findEpochEvent(history []game.EpochEventRecord, epochKey string) *game.EpochEventRecord {
	var last *game.EpochEventRecord
	for i := range history {
		if history[i].EpochKey == epochKey {
			last = &history[i]
		}
	}
	return last
}

// latestCatastrophe returns the most recent catastrophe record for epochKey, or nil.
func latestCatastrophe(history []game.EpochEventRecord, epochKey string) *game.EpochEventRecord {
	for i := len(history) - 1; i >= 0; i-- {
		if history[i].EpochKey == epochKey && history[i].EventType == "catastrophe" {
			return &history[i]
		}
	}
	return nil
}

// catastropheOutcomeLabel renders a catastrophe record's outcome. Only an
// actual Endure reads as Survived: a pending catastrophe says so, and a record
// whose outcome was never stored (older saves) says it is unknown rather than
// claiming survival.
func catastropheOutcomeLabel(outcome string) string {
	switch outcome {
	case game.CatastropheEndured:
		return "[green]✓ Survived[-]"
	case game.CatastropheSuccumbed:
		return "[yellow]Succumbed — legacy bonus[-]"
	case game.CatastrophePending:
		return "[red]Pending[-]"
	}
	return "[gray]catastrophe (outcome not recorded)[-]"
}

// epochEventColor returns a tview color tag name for an epoch event type.
func epochEventColor(eventType string) string {
	switch eventType {
	case "good_legendary":
		return "gold"
	case "good_major":
		return "green"
	case "good_minor":
		return "cyan"
	case "bad_challenging":
		return "red"
	case "catastrophe":
		return "red"
	}
	return "white"
}

// epochOverlayFormatAgeKey converts a snake_case age key to a display-friendly title.
// e.g. "colonial_age" -> "Colonial Age"
func epochOverlayFormatAgeKey(key string) string {
	words := strings.Split(key, "_")
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}
