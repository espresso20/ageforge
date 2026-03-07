package ui

import (
	"fmt"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// epochProvider generates the epoch overlay text from the current game state.
// Shows current epoch status, epoch history, legacy bonuses, and civilisation log.
func epochProvider(state game.GameState) string {
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
		fmt.Fprintf(sb, " [%s]%s %s[-]   Ages: %s\n",
			ep.Color, ep.Icon, ep.Name,
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
					durationStr = fmt.Sprintf("   %d ticks duration", evDef.Duration)
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
	hasCatastrophe := epochHasCatastrophe(state.EpochEventHistory, state.EpochKey)
	if state.PendingCatastrophe != "" {
		sb.WriteString(" Catastrophe: [red]PENDING — choose Endure / Succumb / Defer[-]\n")
	} else if hasCatastrophe {
		if state.LegacyBonuses[state.EpochKey] {
			sb.WriteString(" Catastrophe: [yellow]Succumbed — legacy bonus gained[-]\n")
		} else {
			sb.WriteString(" Catastrophe: [green]✓ Survived[-]\n")
		}
	} else {
		sb.WriteString(" Catastrophe: [gray]not yet triggered[-]\n")
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
		fmt.Fprintf(&line, "   [%s]%s %s[-]", ep.Color, ep.Icon, ep.Name)

		if isCurrent {
			line.WriteString("   [gray][current][-]")
		} else {
			if record != nil {
				evColor := epochEventColor(record.EventType)
				fmt.Fprintf(&line, "   [%s]%s[-]", evColor, record.EventName)
			} else {
				line.WriteString("   [gray]no event[-]")
			}

			hasCat := epochHasCatastrophe(state.EpochEventHistory, ep.Key)
			if hasCat {
				if state.LegacyBonuses[ep.Key] {
					line.WriteString("   [yellow]Succumbed[-]")
				} else {
					line.WriteString("   [green]✓ Survived[-]")
				}
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
		fmt.Fprintf(sb, "   [%s]%s %s:[-]  %s\n",
			ep.Color, ep.Icon, ep.Name,
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

// epochHasCatastrophe returns true if a catastrophe event was recorded for the given epoch.
func epochHasCatastrophe(history []game.EpochEventRecord, epochKey string) bool {
	for _, r := range history {
		if r.EpochKey == epochKey && r.EventType == "catastrophe" {
			return true
		}
	}
	return false
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
